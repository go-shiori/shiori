package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	_ "github.com/go-sql-driver/mysql"
)

var mysqlMigrations = []migration{
	newFileMigration("0.0.0", "0.1.0", "mysql/0000_system_create"),
	newFileMigration("0.1.0", "0.2.0", "mysql/0000_system_insert"),
	newFileMigration("0.2.0", "0.3.0", "mysql/0001_initial_account"),
	newFileMigration("0.3.0", "0.4.0", "mysql/0002_initial_bookmark"),
	newFileMigration("0.4.0", "0.5.0", "mysql/0003_initial_tag"),
	newFileMigration("0.5.0", "0.6.0", "mysql/0004_initial_bookmark_tag"),
	newFuncMigration("0.6.0", "0.7.0", func(db *sql.DB) error {
		// Ensure that bookmark table has `has_content` column and account table has `config` column
		// for users upgrading from <1.5.4 directly into this version.
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.Exec(`ALTER TABLE bookmark ADD COLUMN has_content BOOLEAN DEFAULT 0`)
		if err != nil && strings.Contains(err.Error(), `Duplicate column name`) {
			tx.Rollback()
		} else if err != nil {
			return fmt.Errorf("failed to add has_content column to bookmark table: %w", err)
		} else if err == nil {
			if errCommit := tx.Commit(); errCommit != nil {
				return fmt.Errorf("failed to commit transaction: %w", errCommit)
			}
		}

		tx, err = db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.Exec(`ALTER TABLE account ADD COLUMN config JSON  NOT NULL DEFAULT ('{}')`)
		if err != nil && strings.Contains(err.Error(), `Duplicate column name`) {
			tx.Rollback()
		} else if err != nil {
			return fmt.Errorf("failed to add config column to account table: %w", err)
		} else if err == nil {
			if errCommit := tx.Commit(); errCommit != nil {
				return fmt.Errorf("failed to commit transaction: %w", errCommit)
			}
		}

		return nil
	}),
	newFileMigration("0.7.0", "0.8.0", "mysql/0005_rename_to_created_at"),
	newFileMigration("0.8.0", "0.8.1", "mysql/0006_change_created_at_settings"),
	newFileMigration("0.8.1", "0.8.2", "mysql/0007_add_modified_at"),
	newFileMigration("0.8.2", "0.8.3", "mysql/0008_set_modified_at_equal_created_at"),
	newFileMigration("0.8.3", "0.8.4", "mysql/0009_index_for_created_at"),
	newFileMigration("0.8.4", "0.8.5", "mysql/0010_index_for_modified_at"),
}

// MySQLDatabase is implementation of Database interface
// for connecting to MySQL or MariaDB database.
type MySQLDatabase struct {
	dbbase
}

// OpenMySQLDatabase creates and opens connection to a MySQL Database.
func OpenMySQLDatabase(ctx context.Context, connString string) (mysqlDB *MySQLDatabase, err error) {
	// Open database and start transaction
	db, err := sqlx.ConnectContext(ctx, "mysql", connString)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Second) // in case mysql client has longer timeout (driver issue #674)

	mysqlDB = &MySQLDatabase{dbbase: dbbase{db}}
	return mysqlDB, err
}

// WriterDB returns the underlying sqlx.DB object
func (db *MySQLDatabase) WriterDB() *sqlx.DB {
	return db.DB
}

// ReaderDB returns the underlying sqlx.DB object
func (db *MySQLDatabase) ReaderDB() *sqlx.DB {
	return db.DB
}

// Init initializes the database
func (db *MySQLDatabase) Init(ctx context.Context) error {
	return nil
}

// Migrate runs migrations for this database engine
func (db *MySQLDatabase) Migrate(ctx context.Context) error {
	if err := runMigrations(ctx, db, mysqlMigrations); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// GetDatabaseSchemaVersion fetches the current migrations version of the database
func (db *MySQLDatabase) GetDatabaseSchemaVersion(ctx context.Context) (string, error) {
	var version string

	err := db.GetContext(ctx, &version, "SELECT database_schema_version FROM shiori_system")
	if err != nil {
		return "", errors.WithStack(err)
	}

	return version, nil
}

// SetDatabaseSchemaVersion sets the current migrations version of the database
func (db *MySQLDatabase) SetDatabaseSchemaVersion(ctx context.Context, version string) error {
	tx := db.MustBegin()
	defer tx.Rollback()

	_, err := tx.Exec("UPDATE shiori_system SET database_schema_version = ?", version)
	if err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit()
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *MySQLDatabase) SaveBookmarks(ctx context.Context, create bool, bookmarks ...model.BookmarkDTO) ([]model.BookmarkDTO, error) {
	var result []model.BookmarkDTO

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Prepare statement
		stmtInsertBook, err := tx.Preparex(`INSERT INTO bookmark
			(url, title, excerpt, author, public, content, html, modified_at, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtUpdateBook, err := tx.Preparex(`UPDATE bookmark
		SET url      = ?,
			title    = ?,
			excerpt  = ?,
			author   = ?,
			public   = ?,
			content  = ?,
			html     = ?,
			modified_at = ?
		WHERE id = ?`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtInsertBookTag, err := tx.Preparex(`INSERT IGNORE INTO bookmark_tag
			(tag_id, bookmark_id) VALUES (?, ?)`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtDeleteBookTag, err := tx.Preparex(`DELETE FROM bookmark_tag
			WHERE bookmark_id = ? AND tag_id = ?`)
		if err != nil {
			return errors.WithStack(err)
		}

		// Prepare modified time
		modifiedTime := time.Now().UTC().Format(model.DatabaseDateFormat)

		// Execute statements

		for _, book := range bookmarks {
			// Check URL and title
			if book.URL == "" {
				return errors.New("URL must not be empty")
			}

			if book.Title == "" {
				return errors.New("title must not be empty")
			}

			// Set modified time
			if book.ModifiedAt == "" {
				book.ModifiedAt = modifiedTime
			}

			// Save bookmark
			var err error
			if create {
				book.CreatedAt = modifiedTime
				var res sql.Result
				res, err = stmtInsertBook.ExecContext(ctx,
					book.URL, book.Title, book.Excerpt, book.Author,
					book.Public, book.Content, book.HTML, book.ModifiedAt, book.CreatedAt)
				if err != nil {
					return errors.WithStack(err)
				}
				bookID, err := res.LastInsertId()
				if err != nil {
					return errors.WithStack(err)
				}
				book.ID = int(bookID)
			} else {
				_, err = stmtUpdateBook.ExecContext(ctx,
					book.URL, book.Title, book.Excerpt, book.Author,
					book.Public, book.Content, book.HTML, book.ModifiedAt, book.ID)
			}
			if err != nil {
				return errors.WithStack(err)
			}

			// Save book tags
			newTags := []model.TagDTO{}
			for _, tag := range book.Tags {
				t := tag.ToDTO()
				// If it's deleted tag, delete and continue
				if t.Deleted {
					_, err = stmtDeleteBookTag.ExecContext(ctx, book.ID, t.ID)
					if err != nil {
						return errors.WithStack(err)
					}

					continue
				}

				// Normalize tag name
				tagName := strings.ToLower(tag.Name)
				tagName = strings.Join(strings.Fields(tagName), " ")

				// If tag doesn't have any ID, fetch it from database
				if tag.ID == 0 {
					if err := stmtGetTag.GetContext(ctx, &tag.ID, tagName); err != nil && err != sql.ErrNoRows {
						return errors.WithStack(err)
					}

					// If tag doesn't exist in database, save it
					if tag.ID == 0 {
						res, err := stmtInsertTag.ExecContext(ctx, tagName)
						if err != nil {
							return errors.WithStack(err)
						}

						tagID64, err := res.LastInsertId()
						if err != nil {
							return errors.WithStack(err)
						}

						tag.ID = int(tagID64)
					}

					if _, err := stmtInsertBookTag.ExecContext(ctx, t.ID, book.ID); err != nil {
						return errors.WithStack(err)
					}
				}

				newTags = append(newTags, t)
			}

			book.Tags = newTags
			result = append(result, book)
		}

		return nil
	}); err != nil {
		return result, errors.WithStack(err)
	}

	return result, nil
}

// GetBookmarks fetch list of bookmarks based on submitted options.
func (db *MySQLDatabase) GetBookmarks(ctx context.Context, opts model.DBGetBookmarksOptions) ([]model.BookmarkDTO, error) {
	// Create initial query
	columns := []string{
		`id`,
		`url`,
		`title`,
		`excerpt`,
		`author`,
		`public`,
		`created_at`,
		`modified_at`,
		`content <> "" has_content`}

	if opts.WithContent {
		columns = append(columns, `content`, `html`)
	}

	query := `SELECT ` + strings.Join(columns, ",") + `
		FROM bookmark WHERE 1`

	// Add where clause
	args := []interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND id IN (?)`
		args = append(args, opts.IDs)
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (
			url LIKE ? OR
			MATCH(title, excerpt, content) AGAINST (? IN BOOLEAN MODE)
		)`

		args = append(args, "%"+opts.Keyword+"%", opts.Keyword)
	}

	// Add where clause for tags.
	// First we check for * in excluded and included tags,
	// which means all tags will be excluded and included, respectively.
	excludeAllTags := false
	for _, excludedTag := range opts.ExcludedTags {
		if excludedTag == "*" {
			excludeAllTags = true
			opts.ExcludedTags = []string{}
			break
		}
	}

	includeAllTags := false
	for _, includedTag := range opts.Tags {
		if includedTag == "*" {
			includeAllTags = true
			opts.Tags = []string{}
			break
		}
	}

	// If all tags excluded, we will only show bookmark without tags.
	// In other hand, if all tags included, we will only show bookmark with tags.
	if excludeAllTags {
		query += ` AND id NOT IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	} else if includeAllTags {
		query += ` AND id IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	}

	// Now we only need to find the normal tags
	if len(opts.Tags) > 0 {
		query += ` AND id IN (
			SELECT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = ?)`

		args = append(args, opts.Tags, len(opts.Tags))
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?))`

		args = append(args, opts.ExcludedTags)
	}

	// Add order clause
	switch opts.OrderMethod {
	case model.ByLastAdded:
		query += ` ORDER BY id DESC`
	case model.ByLastModified:
		query += ` ORDER BY modified_at DESC`
	default:
		query += ` ORDER BY id`
	}

	if opts.Limit > 0 && opts.Offset >= 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, opts.Limit, opts.Offset)
	}

	// Expand query, because some of the args might be an array
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Fetch bookmarks
	bookmarks := []model.BookmarkDTO{}
	err = db.Select(&bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WithStack(err)
	}

	// Fetch tags for each bookmarks
	stmtGetTags, err := db.PreparexContext(ctx, `SELECT t.id, t.name
		FROM bookmark_tag bt
		LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id = ?
		ORDER BY t.name`)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer stmtGetTags.Close()

	for _, book := range bookmarks {
		book.Tags = []model.TagDTO{}
		err = stmtGetTags.SelectContext(ctx, &book.Tags, book.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.WithStack(err)
		}
	}

	return bookmarks, nil
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *MySQLDatabase) GetBookmarksCount(ctx context.Context, opts model.DBGetBookmarksOptions) (int, error) {
	// Create initial query
	query := `SELECT COUNT(id) FROM bookmark WHERE 1`

	// Add where clause
	args := []interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND id IN (?)`
		args = append(args, opts.IDs)
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (
			url LIKE ? OR
			MATCH(title, excerpt, content) AGAINST (? IN BOOLEAN MODE)
		)`

		args = append(args,
			"%"+opts.Keyword+"%",
			opts.Keyword)
	}

	// Add where clause for tags.
	// First we check for * in excluded and included tags,
	// which means all tags will be excluded and included, respectively.
	excludeAllTags := false
	for _, excludedTag := range opts.ExcludedTags {
		if excludedTag == "*" {
			excludeAllTags = true
			opts.ExcludedTags = []string{}
			break
		}
	}

	includeAllTags := false
	for _, includedTag := range opts.Tags {
		if includedTag == "*" {
			includeAllTags = true
			opts.Tags = []string{}
			break
		}
	}

	// If all tags excluded, we will only show bookmark without tags.
	// In other hand, if all tags included, we will only show bookmark with tags.
	if excludeAllTags {
		query += ` AND id NOT IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	} else if includeAllTags {
		query += ` AND id IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	}

	// Now we only need to find the normal tags
	if len(opts.Tags) > 0 {
		query += ` AND id IN (
			SELECT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = ?)`

		args = append(args, opts.Tags, len(opts.Tags))
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?))`

		args = append(args, opts.ExcludedTags)
	}

	// Expand query, because some of the args might be an array
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	// Fetch count
	var nBookmarks int
	err = db.GetContext(ctx, &nBookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, errors.WithStack(err)
	}

	return nBookmarks, nil
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *MySQLDatabase) DeleteBookmarks(ctx context.Context, ids ...int) (err error) {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Prepare queries
		delBookmark := `DELETE FROM bookmark`
		delBookmarkTag := `DELETE FROM bookmark_tag`

		// Delete bookmark(s)
		if len(ids) == 0 {
			_, err := tx.ExecContext(ctx, delBookmarkTag)
			if err != nil {
				return errors.WithStack(err)
			}

			_, err = tx.ExecContext(ctx, delBookmark)
			if err != nil {
				return errors.WithStack(err)
			}
		} else {
			delBookmark += ` WHERE id = ?`
			delBookmarkTag += ` WHERE bookmark_id = ?`

			stmtDelBookmark, _ := tx.Preparex(delBookmark)
			stmtDelBookmarkTag, _ := tx.Preparex(delBookmarkTag)

			for _, id := range ids {
				_, err := stmtDelBookmarkTag.ExecContext(ctx, id)
				if err != nil {
					return errors.WithStack(err)
				}

				_, err = stmtDelBookmark.ExecContext(ctx, id)
				if err != nil {
					return errors.WithStack(err)
				}
			}
		}

		return nil
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// GetBookmark fetches bookmark based on its ID or URL.
// Returns the bookmark and boolean whether it's exist or not.
func (db *MySQLDatabase) GetBookmark(ctx context.Context, id int, url string) (model.BookmarkDTO, bool, error) {
	args := []interface{}{id}
	query := `SELECT
		id, url, title, excerpt, author, public,
		content, html, modified_at, created_at, content <> '' has_content
		FROM bookmark WHERE id = ?`

	if url != "" {
		query += ` OR url = ?`
		args = append(args, url)
	}

	book := model.BookmarkDTO{}
	if err := db.GetContext(ctx, &book, query, args...); err != nil && err != sql.ErrNoRows {
		return book, false, errors.WithStack(err)
	}

	return book, book.ID != 0, nil
}

// CreateAccount saves new account to database. Returns error if any happened.
func (db *MySQLDatabase) CreateAccount(ctx context.Context, account model.Account) (*model.Account, error) {
	var accountID int64
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Check for existing username
		var exists bool
		err := tx.QueryRowContext(
			ctx, "SELECT EXISTS(SELECT 1 FROM account WHERE username = ?)",
			account.Username,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking username: %w", err)
		}
		if exists {
			return ErrAlreadyExists
		}

		// Create the account
		result, err := tx.ExecContext(ctx, `INSERT INTO account
			(username, password, owner, config) VALUES (?, ?, ?, ?)`,
			account.Username, account.Password, account.Owner, account.Config)
		if err != nil {
			return errors.WithStack(err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return errors.WithStack(err)
		}
		accountID = id
		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	account.ID = model.DBID(accountID)
	return &account, nil
}

// UpdateAccount update account in database
func (db *MySQLDatabase) UpdateAccount(ctx context.Context, account model.Account) error {
	if account.ID == 0 {
		return ErrNotFound
	}

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Check for existing username
		var exists bool
		err := tx.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM account WHERE username = ? AND id != ?)",
			account.Username, account.ID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking username: %w", err)
		}
		if exists {
			return ErrAlreadyExists
		}

		result, err := tx.ExecContext(ctx, `UPDATE account
			SET username = ?, password = ?, owner = ?, config = ?
			WHERE id = ?`,
			account.Username, account.Password, account.Owner, account.Config, account.ID)
		if err != nil {
			return errors.WithStack(err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return errors.WithStack(err)
		}
		if rows == 0 {
			return ErrNotFound
		}

		return nil
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// ListAccounts fetch list of account (without its password) based on submitted options.
func (db *MySQLDatabase) ListAccounts(ctx context.Context, opts model.DBListAccountsOptions) ([]model.Account, error) {
	// Create query
	args := []interface{}{}
	fields := []string{"id", "username", "owner", "config"}
	if opts.WithPassword {
		fields = append(fields, "password")
	}

	query := fmt.Sprintf(`SELECT %s FROM account WHERE 1`, strings.Join(fields, ", "))

	if opts.Keyword != "" {
		query += " AND username LIKE ?"
		args = append(args, "%"+opts.Keyword+"%")
	}

	if opts.Username != "" {
		query += " AND username = ?"
		args = append(args, opts.Username)
	}

	if opts.Owner {
		query += " AND owner = 1"
	}

	// Fetch list account
	accounts := []model.Account{}
	err := db.SelectContext(ctx, &accounts, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WithStack(err)
	}

	return accounts, nil
}

// GetAccount fetch account with matching ID.
// Returns the account and boolean whether it's exist or not.
func (db *MySQLDatabase) GetAccount(ctx context.Context, id model.DBID) (*model.Account, bool, error) {
	account := model.Account{}
	err := db.GetContext(ctx, &account, `SELECT
		id, username, password, owner, config FROM account WHERE id = ?`,
		id,
	)
	if err != nil && err != sql.ErrNoRows {
		return &account, false, errors.WithStack(err)
	}

	// Use custom not found error if that's the result of the query
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}

	return &account, account.ID != 0, err
}

// DeleteAccount removes record with matching ID.
func (db *MySQLDatabase) DeleteAccount(ctx context.Context, id model.DBID) error {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM account WHERE id = ?`, id)
		if err != nil {
			return errors.WithStack(fmt.Errorf("error deleting account: %v", err))
		}

		rows, err := result.RowsAffected()
		if err != nil && err != sql.ErrNoRows {
			return errors.WithStack(fmt.Errorf("error getting rows affected: %v", err))
		}

		if rows == 0 {
			return ErrNotFound
		}

		return nil
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// CreateTags creates new tags from submitted objects.
func (db *MySQLDatabase) CreateTags(ctx context.Context, tags ...model.Tag) error {
	query := `INSERT INTO tag (name) VALUES `
	values := []interface{}{}

	for _, t := range tags {
		query += "(?),"
		values = append(values, t.Name)
	}
	query = query[0 : len(query)-1]

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		stmt, err := tx.Preparex(query)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "error preparing query")
		}

		_, err = stmt.ExecContext(ctx, values...)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "error executing query")
		}

		return nil
	}); err != nil {
		return errors.Wrap(errors.WithStack(err), "error running transaction")
	}

	return nil
}

// GetTags fetch list of tags and their frequency.
func (db *MySQLDatabase) GetTags(ctx context.Context) ([]model.TagDTO, error) {
	tags := []model.TagDTO{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) bookmark_count
		FROM bookmark_tag bt
		LEFT JOIN tag t ON bt.tag_id = t.id
		GROUP BY bt.tag_id ORDER BY t.name`

	err := db.SelectContext(ctx, &tags, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WithStack(err)
	}

	return tags, nil
}

// RenameTag change the name of a tag.
func (db *MySQLDatabase) RenameTag(ctx context.Context, id int, newName string) error {
	err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := db.ExecContext(ctx, `UPDATE tag SET name = ? WHERE id = ?`, newName, id)
		return errors.WithStack(err)
	})

	return errors.WithStack(err)
}
