package database

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/huandu/go-sqlbuilder"
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

	mysqlDB = &MySQLDatabase{dbbase: NewDBBase(db, db, sqlbuilder.MySQL)}
	return mysqlDB, err
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
						t.ID = int(tagID64)
					}
				}

				// Always insert the tag-bookmark association
				if _, err := stmtInsertBookTag.ExecContext(ctx, tag.ID, book.ID); err != nil {
					return errors.WithStack(err)
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
		`content <> "" as has_content`}

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
	if slices.Contains(opts.ExcludedTags, "*") {
		excludeAllTags = true
		opts.ExcludedTags = []string{}
	}

	includeAllTags := false
	if slices.Contains(opts.Tags, "*") {
		includeAllTags = true
		opts.Tags = []string{}
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

	// Fetch tags for each bookmark
	for i, book := range bookmarks {
		tags, err := db.getTagsForBookmark(ctx, book.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags: %w", err)
		}
		bookmarks[i].Tags = tags
	}

	return bookmarks, nil
}

func (db *MySQLDatabase) getTagsForBookmark(ctx context.Context, bookmarkID int) ([]model.TagDTO, error) {
	sb := sqlbuilder.MySQL.NewSelectBuilder()
	sb.Select("t.id", "t.name")
	sb.From("bookmark_tag bt")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "tag t", "bt.tag_id = t.id")
	sb.Where(sb.Equal("bt.bookmark_id", bookmarkID))
	sb.OrderBy("t.name")

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	tags := []model.TagDTO{}
	err := db.ReaderDB().SelectContext(ctx, &tags, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
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
	if slices.Contains(opts.ExcludedTags, "*") {
		excludeAllTags = true
		opts.ExcludedTags = []string{}
	}

	includeAllTags := false
	if slices.Contains(opts.Tags, "*") {
		includeAllTags = true
		opts.Tags = []string{}
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
	// Create the main query builder for bookmark data
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"id", "url", "title", "excerpt", "author", `public`, "modified_at",
		"content", "html", "created_at", "has_content")
	sb.From("bookmark")

	// Add conditions
	if id != 0 {
		sb.Where(sb.Equal("id", id))
	} else if url != "" {
		sb.Where(sb.Equal("url", url))
	} else {
		return model.BookmarkDTO{}, false, fmt.Errorf("id or url is required")
	}

	// Build the query
	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)
	// Execute the query
	book := model.BookmarkDTO{}
	err := db.ReaderDB().GetContext(ctx, &book, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return book, false, nil
		}
		return book, false, fmt.Errorf("failed to get bookmark: %w", err)
	}

	// If bookmark exists, fetch its tags
	if book.ID != 0 {
		// Create query builder for tags
		tagSb := sqlbuilder.NewSelectBuilder()
		tagSb.Select("t.id", "t.name")
		tagSb.From("tag t")
		tagSb.JoinWithOption(sqlbuilder.InnerJoin, "bookmark_tag bt", "bt.tag_id = t.id")
		tagSb.Where(tagSb.Equal("bt.bookmark_id", book.ID))

		// Build the query
		tagQuery, tagArgs := tagSb.Build()
		tagQuery = db.ReaderDB().Rebind(tagQuery)
		// Execute the query
		tags := []model.TagDTO{}
		if err := db.ReaderDB().SelectContext(ctx, &tags, tagQuery, tagArgs...); err != nil && err != sql.ErrNoRows {
			return book, false, fmt.Errorf("failed to get tags: %w", err)
		}

		book.Tags = tags
	}

	return book, true, nil
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
			return fmt.Errorf("error executing query: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("error getting last insert id: %w", err)
		}
		accountID = id
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error running transaction: %w", err)
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
			return fmt.Errorf("error updating account: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if rows == 0 {
			return ErrNotFound
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error running transaction: %w", err)
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
	if err != nil {
		if err == sql.ErrNoRows {
			return &account, false, ErrNotFound
		}
		return &account, false, fmt.Errorf("error getting account: %w", err)
	}

	return &account, true, nil
}

// DeleteAccount removes record with matching ID.
func (db *MySQLDatabase) DeleteAccount(ctx context.Context, id model.DBID) error {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM account WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("error deleting account: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}

		if rows == 0 {
			return ErrNotFound
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error running transaction: %w", err)
	}

	return nil
}

// CreateTags creates new tags from submitted objects.
func (db *MySQLDatabase) CreateTags(ctx context.Context, tags ...model.Tag) ([]model.Tag, error) {
	if len(tags) == 0 {
		return []model.Tag{}, nil
	}

	// Create a slice to hold the created tags
	createdTags := make([]model.Tag, len(tags))
	copy(createdTags, tags)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// For MySQL, we need to insert tags one by one to get their IDs
		stmtInsertTag, err := tx.PrepareContext(ctx, "INSERT INTO tag (name) VALUES (?)")
		if err != nil {
			return fmt.Errorf("failed to prepare tag insertion statement: %w", err)
		}
		defer stmtInsertTag.Close()

		// Insert each tag and get its ID
		for i, tag := range createdTags {
			result, err := stmtInsertTag.ExecContext(ctx, tag.Name)
			if err != nil {
				return fmt.Errorf("failed to insert tag: %w", err)
			}

			// Get the last inserted ID
			tagID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get last insert ID: %w", err)
			}

			createdTags[i].ID = int(tagID)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to run tag creation transaction: %w", err)
	}

	return createdTags, nil
}

// CreateTag creates a new tag in database.
func (db *MySQLDatabase) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	// Use CreateTags to implement this method
	createdTags, err := db.CreateTags(ctx, tag)
	if err != nil {
		return model.Tag{}, err
	}

	if len(createdTags) == 0 {
		return model.Tag{}, fmt.Errorf("failed to create tag")
	}

	return createdTags[0], nil
}

// RenameTag change the name of a tag.
func (db *MySQLDatabase) RenameTag(ctx context.Context, id int, newName string) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("tag")
	sb.Set(sb.Assign("name", newName))
	sb.Where(sb.Equal("id", id))

	query, args := sb.Build()
	query = db.WriterDB().Rebind(query)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to rename tag: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// GetTag fetch a tag by its ID.
func (db *MySQLDatabase) GetTag(ctx context.Context, id int) (model.TagDTO, bool, error) {
	sb := sqlbuilder.MySQL.NewSelectBuilder()
	sb.Select("t.id", "t.name", "COUNT(bt.tag_id) bookmark_count")
	sb.From("tag t")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "bookmark_tag bt", "bt.tag_id = t.id")
	sb.Where(sb.Equal("t.id", id))
	sb.GroupBy("t.id")
	sb.OrderBy("t.name")

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	var tag model.TagDTO
	err := db.ReaderDB().GetContext(ctx, &tag, query, args...)
	if err == sql.ErrNoRows {
		return model.TagDTO{}, false, nil
	}
	if err != nil {
		return model.TagDTO{}, false, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, true, nil
}

// UpdateTag updates a tag in the database.
func (db *MySQLDatabase) UpdateTag(ctx context.Context, tag model.Tag) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("tag")
	sb.Set(sb.Assign("name", tag.Name))
	sb.Where(sb.Equal("id", tag.ID))

	query, args := sb.Build()
	query = db.WriterDB().Rebind(query)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update tag: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// DeleteTag removes a tag from the database.
func (db *MySQLDatabase) DeleteTag(ctx context.Context, id int) error {
	// First, check if the tag exists
	_, exists, err := db.GetTag(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if tag exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	// Delete all bookmark_tag associations
	deleteAssocSb := sqlbuilder.NewDeleteBuilder()
	deleteAssocSb.DeleteFrom("bookmark_tag")
	deleteAssocSb.Where(deleteAssocSb.Equal("tag_id", id))

	deleteAssocQuery, deleteAssocArgs := deleteAssocSb.Build()
	deleteAssocQuery = db.WriterDB().Rebind(deleteAssocQuery)

	// Then, delete the tag itself
	deleteTagSb := sqlbuilder.NewDeleteBuilder()
	deleteTagSb.DeleteFrom("tag")
	deleteTagSb.Where(deleteTagSb.Equal("id", id))

	deleteTagQuery, deleteTagArgs := deleteTagSb.Build()
	deleteTagQuery = db.WriterDB().Rebind(deleteTagQuery)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Delete bookmark_tag associations
		_, err := tx.ExecContext(ctx, deleteAssocQuery, deleteAssocArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete tag associations: %w", err)
		}

		// Delete the tag
		_, err = tx.ExecContext(ctx, deleteTagQuery, deleteTagArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete tag: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// SaveBookmark saves a single bookmark to database without handling tags.
// It only updates the bookmark data in the database.
func (db *MySQLDatabase) SaveBookmark(ctx context.Context, bookmark model.Bookmark) error {
	if bookmark.ID <= 0 {
		return fmt.Errorf("bookmark ID must be greater than 0")
	}

	// Prepare modified time if not set
	if bookmark.ModifiedAt == "" {
		bookmark.ModifiedAt = time.Now().UTC().Format(model.DatabaseDateFormat)
	}

	// Check URL and title
	if bookmark.URL == "" {
		return errors.New("URL must not be empty")
	}

	if bookmark.Title == "" {
		return errors.New("title must not be empty")
	}

	// Use sqlbuilder to build the update query
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("bookmark")
	sb.Set(
		sb.Assign("url", bookmark.URL),
		sb.Assign("title", bookmark.Title),
		sb.Assign("excerpt", bookmark.Excerpt),
		sb.Assign("author", bookmark.Author),
		sb.Assign("public", bookmark.Public),
		sb.Assign("modified_at", bookmark.ModifiedAt),
		sb.Assign("has_content", bookmark.HasContent),
	)
	sb.Where(sb.Equal("id", bookmark.ID))

	query, args := sb.Build()
	query = db.WriterDB().Rebind(query)

	return db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Update bookmark
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update bookmark: %w", err)
		}

		return nil
	})
}

func (db *MySQLDatabase) SaveBookmarkTags(ctx context.Context, bookmarkID int, tagIDs []int) error {
	return db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Prepare statements
		stmtDeleteAllBookmarkTags, err := tx.PreparexContext(ctx, `DELETE FROM bookmark_tag WHERE bookmark_id = ?`)
		if err != nil {
			return fmt.Errorf("failed to prepare delete all bookmark tags statement: %w", err)
		}

		stmtInsertBookTag, err := tx.PreparexContext(ctx, `INSERT IGNORE INTO bookmark_tag
			(tag_id, bookmark_id) VALUES (?, ?)`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert book tag statement: %w", err)
		}

		// Delete all existing tags for this bookmark
		_, err = stmtDeleteAllBookmarkTags.ExecContext(ctx, bookmarkID)
		if err != nil {
			return fmt.Errorf("failed to delete existing bookmark tags: %w", err)
		}

		// Insert new tags
		for _, tagID := range tagIDs {
			_, err := stmtInsertBookTag.ExecContext(ctx, tagID, bookmarkID)
			if err != nil {
				return fmt.Errorf("failed to insert bookmark tag: %w", err)
			}
		}

		return nil
	})
}

// BulkUpdateBookmarkTags updates tags for multiple bookmarks.
// It ensures that all bookmarks and tags exist before proceeding.
func (db *MySQLDatabase) BulkUpdateBookmarkTags(ctx context.Context, bookmarkIDs []int, tagIDs []int) error {
	if len(bookmarkIDs) == 0 || len(tagIDs) == 0 {
		return nil
	}

	// Convert int slices to interface slices for sqlbuilder
	bookmarkIDsIface := make([]interface{}, len(bookmarkIDs))
	for i, id := range bookmarkIDs {
		bookmarkIDsIface[i] = id
	}

	tagIDsIface := make([]interface{}, len(tagIDs))
	for i, id := range tagIDs {
		tagIDsIface[i] = id
	}

	// Verify all bookmarks exist
	bookmarkSb := sqlbuilder.NewSelectBuilder()
	bookmarkSb.Select("id")
	bookmarkSb.From("bookmark")
	bookmarkSb.Where(bookmarkSb.In("id", bookmarkIDsIface...))

	bookmarkQuery, bookmarkArgs := bookmarkSb.Build()
	bookmarkQuery = db.ReaderDB().Rebind(bookmarkQuery)

	var existingBookmarkIDs []int
	err := db.ReaderDB().SelectContext(ctx, &existingBookmarkIDs, bookmarkQuery, bookmarkArgs...)
	if err != nil {
		return fmt.Errorf("failed to check bookmarks: %w", err)
	}

	if len(existingBookmarkIDs) != len(bookmarkIDs) {
		// Find which bookmarks don't exist
		missingBookmarkIDs := model.SliceDifference(bookmarkIDs, existingBookmarkIDs)
		return fmt.Errorf("some bookmarks do not exist: %v", missingBookmarkIDs)
	}

	// Verify all tags exist
	tagSb := sqlbuilder.NewSelectBuilder()
	tagSb.Select("id")
	tagSb.From("tag")
	tagSb.Where(tagSb.In("id", tagIDsIface...))

	tagQuery, tagArgs := tagSb.Build()
	tagQuery = db.ReaderDB().Rebind(tagQuery)

	var existingTagIDs []int
	err = db.ReaderDB().SelectContext(ctx, &existingTagIDs, tagQuery, tagArgs...)
	if err != nil {
		return fmt.Errorf("failed to check tags: %w", err)
	}

	if len(existingTagIDs) != len(tagIDs) {
		// Find which tags don't exist
		missingTagIDs := model.SliceDifference(tagIDs, existingTagIDs)
		return fmt.Errorf("some tags do not exist: %v", missingTagIDs)
	}

	return db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Delete existing bookmark-tag associations
		deleteSb := sqlbuilder.NewDeleteBuilder()
		deleteSb.DeleteFrom("bookmark_tag")
		deleteSb.Where(deleteSb.In("bookmark_id", bookmarkIDsIface...))

		deleteQuery, deleteArgs := deleteSb.Build()
		deleteQuery = tx.Rebind(deleteQuery)

		_, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete existing bookmark tags: %w", err)
		}

		// Insert new bookmark-tag associations
		if len(tagIDs) > 0 {
			// Build values for bulk insert
			insertSb := sqlbuilder.NewInsertBuilder()
			insertSb.InsertInto("bookmark_tag")
			// Fix column order to match database schema
			insertSb.Cols("bookmark_id", "tag_id")

			for _, bookmarkID := range bookmarkIDs {
				for _, tagID := range tagIDs {
					// Match the column order in Values
					insertSb.Values(bookmarkID, tagID)
				}
			}

			insertQuery, insertArgs := insertSb.Build()
			// Add MySQL-specific INSERT IGNORE INTO syntax
			insertQuery = strings.Replace(insertQuery, "INSERT INTO", "INSERT IGNORE INTO", 1)
			insertQuery = tx.Rebind(insertQuery)

			_, err = tx.ExecContext(ctx, insertQuery, insertArgs...)
			if err != nil {
				return fmt.Errorf("failed to insert bookmark tags: %w", err)
			}
		}

		return nil
	})
}
