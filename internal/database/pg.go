package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/lib/pq"
)

var postgresMigrations = []migration{
	newFileMigration("0.0.0", "0.1.0", "postgres/0000_system"),
	newFileMigration("0.1.0", "0.2.0", "postgres/0001_initial"),
	newFuncMigration("0.2.0", "0.3.0", func(db *sql.DB) error {
		// Ensure that bookmark table has `has_content` column and account table has `config` column
		// for users upgrading from <1.5.4 directly into this version.
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		_, err = tx.Exec(`ALTER TABLE bookmark ADD COLUMN has_content BOOLEAN DEFAULT FALSE NOT NULL`)
		if err != nil {
			// Check if this is a "column already exists" error (PostgreSQL error code 42701)
			// If it's not, return error.
			// This is needed for users upgrading from >1.5.4 directly into this version.
			pqErr, ok := err.(*pq.Error)
			if ok && pqErr.Code == "42701" {
				tx.Rollback()
			} else {
				return fmt.Errorf("failed to add has_content column to bookmark table: %w", err)
			}
		} else {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}
		}

		tx, err = db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		_, err = tx.Exec(`ALTER TABLE account ADD COLUMN config JSONB NOT NULL DEFAULT '{}'`)
		if err != nil {
			// Check if this is a "column already exists" error (PostgreSQL error code 42701)
			// If it's not, return error
			// This is needed for users upgrading from >1.5.4 directly into this version.
			pqErr, ok := err.(*pq.Error)
			if ok && pqErr.Code == "42701" {
				tx.Rollback()
			} else {
				return fmt.Errorf("failed to add config column to account table: %w", err)
			}
		} else {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}
		}

		return nil
	}),
	newFileMigration("0.3.0", "0.4.0", "postgres/0002_created_time"),
}

// PGDatabase is implementation of Database interface
// for connecting to PostgreSQL database.
type PGDatabase struct {
	dbbase
}

// OpenPGDatabase creates and opens connection to a PostgreSQL Database.
func OpenPGDatabase(ctx context.Context, connString string) (pgDB *PGDatabase, err error) {
	// Open database and start transaction
	db, err := sqlx.ConnectContext(ctx, "postgres", connString)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Second)

	pgDB = &PGDatabase{dbbase: dbbase{db}}
	return pgDB, err
}

// WriterDB returns the underlying sqlx.DB object
func (db *PGDatabase) WriterDB() *sqlx.DB {
	return db.DB
}

// ReaderDB returns the underlying sqlx.DB object
func (db *PGDatabase) ReaderDB() *sqlx.DB {
	return db.DB
}

// Init initializes the database
func (db *PGDatabase) Init(ctx context.Context) error {
	return nil
}

// Migrate runs migrations for this database engine
func (db *PGDatabase) Migrate(ctx context.Context) error {
	if err := runMigrations(ctx, db, postgresMigrations); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// GetDatabaseSchemaVersion fetches the current migrations version of the database
func (db *PGDatabase) GetDatabaseSchemaVersion(ctx context.Context) (string, error) {
	var version string

	err := db.GetContext(ctx, &version, "SELECT database_schema_version FROM shiori_system")
	if err != nil {
		return "", errors.WithStack(err)
	}

	return version, nil
}

// SetDatabaseSchemaVersion sets the current migrations version of the database
func (db *PGDatabase) SetDatabaseSchemaVersion(ctx context.Context, version string) error {
	tx := db.MustBegin()
	defer tx.Rollback()

	return db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec("UPDATE shiori_system SET database_schema_version = $1", version)
		if err != nil {
			return errors.WithStack(err)
		}

		return tx.Commit()
	})
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *PGDatabase) SaveBookmarks(ctx context.Context, create bool, bookmarks ...model.BookmarkDTO) (result []model.BookmarkDTO, err error) {
	result = []model.BookmarkDTO{}
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Prepare statement
		stmtInsertBook, err := tx.Preparex(`INSERT INTO bookmark
			(url, title, excerpt, author, public, content, html, modified_at, created_at)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtUpdateBook, err := tx.Preparex(`UPDATE bookmark SET
			url      = $1,
			title    = $2,
			excerpt  = $3,
			author   = $4,
			public   = $5,
			content  = $6,
			html     = $7,
			modified_at = $8
			WHERE id = $9`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = $1`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES ($1) RETURNING id`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtInsertBookTag, err := tx.Preparex(`INSERT INTO bookmark_tag
			(tag_id, bookmark_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
		if err != nil {
			return errors.WithStack(err)
		}

		stmtDeleteBookTag, err := tx.Preparex(`DELETE FROM bookmark_tag
			WHERE bookmark_id = $1 AND tag_id = $2`)
		if err != nil {
			return errors.WithStack(err)
		}

		// Prepare modified time
		modifiedTime := time.Now().UTC().Format(model.DatabaseDateFormat)

		// Execute statements
		result = []model.BookmarkDTO{}
		for _, book := range bookmarks {
			// URL and title
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
				err = stmtInsertBook.QueryRowContext(ctx,
					book.URL, book.Title, book.Excerpt, book.Author,
					book.Public, book.Content, book.HTML, book.ModifiedAt, book.CreatedAt).Scan(&book.ID)
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
					err = stmtGetTag.GetContext(ctx, &tag.ID, tagName)
					if err != nil && !errors.Is(err, sql.ErrNoRows) {
						return errors.WithStack(err)
					}

					// If tag doesn't exist in database, save it
					if tag.ID == 0 {
						var tagID64 int64
						err = stmtInsertTag.GetContext(ctx, &tagID64, tagName)
						if err != nil {
							return errors.WithStack(err)
						}

						tag.ID = int(tagID64)
					}

					if _, err := stmtInsertBookTag.ExecContext(ctx, tag.ID, book.ID); err != nil {
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
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// GetBookmarks fetch list of bookmarks based on submitted options.
func (db *PGDatabase) GetBookmarks(ctx context.Context, opts model.DBGetBookmarksOptions) ([]model.BookmarkDTO, error) {
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
		`content <> '' has_content`}

	if opts.WithContent {
		columns = append(columns, `content`, `html`)
	}

	query := `SELECT ` + strings.Join(columns, ",") + `
		FROM bookmark WHERE TRUE`

	// Add where clause
	arg := map[string]interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND id IN (:ids)`
		arg["ids"] = opts.IDs
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (
			url LIKE '%' || :kw || '%' OR
			title LIKE '%' || :kw || '%' OR
			excerpt LIKE '%' || :kw || '%' OR
			content LIKE '%' || :kw || '%'
		)`

		arg["kw"] = opts.Keyword
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
			WHERE t.name IN(:tags)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = :ltags)`

		arg["tags"] = opts.Tags
		arg["ltags"] = len(opts.Tags)
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(:extags))`

		arg["extags"] = opts.ExcludedTags
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
		query += ` LIMIT :limit OFFSET :offset`
		arg["limit"] = opts.Limit
		arg["offset"] = opts.Offset
	}

	// Expand query, because some of the args might be an array
	var err error
	query, args, _ := sqlx.Named(query, arg)
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to expand query: %v", err)
	}
	query = db.Rebind(query)

	// Fetch bookmarks
	bookmarks := []model.BookmarkDTO{}
	err = db.SelectContext(ctx, &bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}

	// Fetch tags for each bookmarks
	stmtGetTags, err := db.PreparexContext(ctx, `SELECT t.id, t.name
		FROM bookmark_tag bt
		LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id = $1
		ORDER BY t.name`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tag query: %v", err)
	}
	defer stmtGetTags.Close()

	for i, book := range bookmarks {
		book.Tags = []model.TagDTO{}
		err = stmtGetTags.SelectContext(ctx, &book.Tags, book.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch tags: %v", err)
		}

		bookmarks[i] = book
	}

	return bookmarks, nil
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *PGDatabase) GetBookmarksCount(ctx context.Context, opts model.DBGetBookmarksOptions) (int, error) {
	// Create initial query
	query := `SELECT COUNT(id) FROM bookmark WHERE TRUE`

	arg := map[string]interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND id IN (:ids)`
		arg["ids"] = opts.IDs
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (
			url LIKE '%' || :kw || '%' OR
			title LIKE '%' || :kw || '%' OR
			excerpt LIKE '%' || :kw || '%' OR
			content LIKE '%' || :kw || '%'
		)`

		arg["lurl"] = "%" + opts.Keyword + "%"
		arg["kw"] = opts.Keyword
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
			WHERE t.name IN(:tags)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = :ltags)`

		arg["tags"] = opts.Tags
		arg["ltags"] = len(opts.Tags)
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(:etags))`

		arg["etags"] = opts.ExcludedTags
	}

	// Expand query, because some of the args might be an array
	var err error
	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	query = db.Rebind(query)

	// Fetch count
	var nBookmarks int
	err = db.GetContext(ctx, &nBookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, errors.WithStack(err)
	}

	return nBookmarks, nil
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *PGDatabase) DeleteBookmarks(ctx context.Context, ids ...int) (err error) {
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
			delBookmark += ` WHERE id = $1`
			delBookmarkTag += ` WHERE bookmark_id = $1`

			stmtDelBookmark, err := tx.Preparex(delBookmark)
			if err != nil {
				return errors.WithStack(err)
			}
			stmtDelBookmarkTag, err := tx.Preparex(delBookmarkTag)
			if err != nil {
				return errors.WithStack(err)
			}

			for _, id := range ids {
				_, err = stmtDelBookmarkTag.ExecContext(ctx, id)
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
func (db *PGDatabase) GetBookmark(ctx context.Context, id int, url string) (model.BookmarkDTO, bool, error) {
	args := []interface{}{id}
	query := `SELECT
		id, url, title, excerpt, author, public,
		content, html, modified_at, created_at, content <> '' has_content
		FROM bookmark WHERE id = $1`

	if url != "" {
		query += ` OR url = $2`
		args = append(args, url)
	}

	book := model.BookmarkDTO{}
	if err := db.GetContext(ctx, &book, query, args...); err != nil && err != sql.ErrNoRows {
		return book, false, errors.WithStack(err)
	}

	return book, book.ID != 0, nil
}

// CreateAccount saves new account to database. Returns error if any happened.
func (db *PGDatabase) CreateAccount(ctx context.Context, account model.Account) (*model.Account, error) {
	var accountID int64
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Check for existing username
		var exists bool
		err := tx.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM account WHERE username = $1)",
			account.Username).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking username: %w", err)
		}
		if exists {
			return ErrAlreadyExists
		}

		// Create the account
		query, err := tx.PrepareContext(ctx, `INSERT INTO account
			(username, password, owner, config) VALUES ($1, $2, $3, $4)
			RETURNING id`)
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = query.QueryRowContext(ctx,
			account.Username, account.Password, account.Owner, account.Config).Scan(&accountID)
		if err != nil {
			return fmt.Errorf("error executing query: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("error during transaction: %w", err)
	}

	account.ID = model.DBID(accountID)

	return &account, nil
}

// UpdateAccount updates account in database.
func (db *PGDatabase) UpdateAccount(ctx context.Context, account model.Account) error {
	if account.ID == 0 {
		return ErrNotFound
	}

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Check for existing username
		var exists bool
		err := tx.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM account WHERE username = $1 AND id != $2)",
			account.Username, account.ID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking username: %w", err)
		}
		if exists {
			return ErrAlreadyExists
		}

		result, err := tx.ExecContext(ctx, `UPDATE account
			SET username = $1, password = $2, owner = $3, config = $4
			WHERE id = $5`,
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
func (db *PGDatabase) ListAccounts(ctx context.Context, opts model.DBListAccountsOptions) ([]model.Account, error) {
	// Create query
	args := []interface{}{}
	fields := []string{"id", "username", "owner", "config"}
	if opts.WithPassword {
		fields = append(fields, "password")
	}

	query := fmt.Sprintf(`SELECT %s FROM account WHERE TRUE`, strings.Join(fields, ", "))

	if opts.Keyword != "" {
		query += " AND username LIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+opts.Keyword+"%")
	}

	if opts.Username != "" {
		query += " AND username = $" + strconv.Itoa(len(args)+1)
		args = append(args, opts.Username)
	}

	if opts.Owner {
		query += " AND owner = TRUE"
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
func (db *PGDatabase) GetAccount(ctx context.Context, id model.DBID) (*model.Account, bool, error) {
	account := model.Account{}
	err := db.GetContext(ctx, &account, `SELECT
		id, username, password, owner, config FROM account WHERE id = $1`,
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
func (db *PGDatabase) DeleteAccount(ctx context.Context, id model.DBID) error {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM account WHERE id = $1`, id)
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
func (db *PGDatabase) CreateTags(ctx context.Context, tags ...model.Tag) error {
	query := `INSERT INTO tag (name) VALUES (:name)`

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.NamedExec(query, tags); err != nil {
			return errors.WithStack(err)
		}

		return nil
	}); err != nil {
		return errors.Wrap(errors.WithStack(err), "error running transaction")
	}

	return nil
}

// GetTags fetch list of tags and their frequency.
func (db *PGDatabase) GetTags(ctx context.Context) ([]model.TagDTO, error) {
	tags := []model.TagDTO{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) bookmark_count
		FROM bookmark_tag bt
		LEFT JOIN tag t ON bt.tag_id = t.id
		GROUP BY bt.tag_id, t.name ORDER BY t.name`

	err := db.SelectContext(ctx, &tags, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WithStack(err)
	}

	return tags, nil
}

// RenameTag change the name of a tag.
func (db *PGDatabase) RenameTag(ctx context.Context, id int, newName string) error {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := db.Exec(`UPDATE tag SET name = $1 WHERE id = $2`, newName, id)
		return errors.WithStack(err)
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
