package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

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

	pgDB = &PGDatabase{dbbase: dbbase{*db}}
	return pgDB, err
}

// Migrate runs migrations for this database engine
func (db *PGDatabase) Migrate() error {
	sourceDriver, err := iofs.New(migrations, "migrations/postgres")
	if err != nil {
		return errors.WithStack(err)
	}

	dbDriver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	if err != nil {
		return errors.WithStack(err)
	}

	migration, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"postgres",
		dbDriver,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return migration.Up()
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *PGDatabase) SaveBookmarks(ctx context.Context, bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
	result = []model.Bookmark{}
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Prepare statement
		stmtInsertBook, err := tx.Preparex(`INSERT INTO bookmark
			(url, title, excerpt, author, public, content, html, modified)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT(url) DO UPDATE SET
			url      = $1,
			title    = $2,
			excerpt  = $3,
			author   = $4,
			public   = $5,
			content  = $6,
			html     = $7,
			modified = $8
		RETURNING id`)
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
		result = []model.Bookmark{}
		for _, book := range bookmarks {
			// URL and title
			if book.URL == "" {
				return errors.New("URL must not be empty")
			}

			if book.Title == "" {
				return errors.New("title must not be empty")
			}

			// Set modified time
			if book.Modified == "" {
				book.Modified = modifiedTime
			}

			// Save bookmark
			err := stmtInsertBook.QueryRowContext(ctx,
				book.URL, book.Title, book.Excerpt, book.Author,
				book.Public, book.Content, book.HTML, book.Modified).Scan(&book.ID)
			if err != nil {
				return errors.WithStack(err)
			}

			// Save book tags
			newTags := []model.Tag{}
			for _, tag := range book.Tags {
				// If it's deleted tag, delete and continue
				if tag.Deleted {
					_, err = stmtDeleteBookTag.ExecContext(ctx, book.ID, tag.ID)
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

				newTags = append(newTags, tag)
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
func (db *PGDatabase) GetBookmarks(ctx context.Context, opts GetBookmarksOptions) ([]model.Bookmark, error) {
	// Create initial query
	columns := []string{
		`id`,
		`url`,
		`title`,
		`excerpt`,
		`author`,
		`public`,
		`modified`,
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
			url LIKE :lkw OR
			title LIKE :kw OR
			excerpt LIKE :kw OR
			content LIKE :kw
		)`

		arg["lkw"] = "%" + opts.Keyword + "%"
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
	case ByLastAdded:
		query += ` ORDER BY id DESC`
	case ByLastModified:
		query += ` ORDER BY modified DESC`
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
	bookmarks := []model.Bookmark{}
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
		book.Tags = []model.Tag{}
		err = stmtGetTags.SelectContext(ctx, &book.Tags, book.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch tags: %v", err)
		}

		bookmarks[i] = book
	}

	return bookmarks, nil
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *PGDatabase) GetBookmarksCount(ctx context.Context, opts GetBookmarksOptions) (int, error) {
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
			url LIKE :lurl OR
			title LIKE :kw OR
			excerpt LIKE :kw OR
			content LIKE :kw
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

// GetBookmark fetchs bookmark based on its ID or URL.
// Returns the bookmark and boolean whether it's exist or not.
func (db *PGDatabase) GetBookmark(ctx context.Context, id int, url string) (model.Bookmark, bool, error) {
	args := []interface{}{id}
	query := `SELECT
		id, url, title, excerpt, author, public,
		content, html, modified, content <> '' has_content
		FROM bookmark WHERE id = $1`

	if url != "" {
		query += ` OR url = $2`
		args = append(args, url)
	}

	book := model.Bookmark{}
	if err := db.GetContext(ctx, &book, query, args...); err != nil {
		return book, false, errors.WithStack(err)
	}

	return book, book.ID != 0, nil
}

// SaveAccount saves new account to database. Returns error if any happened.
func (db *PGDatabase) SaveAccount(ctx context.Context, account model.Account) (err error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), 10)
	if err != nil {
		return err
	}

	// Insert account to database
	_, err = db.ExecContext(ctx, `INSERT INTO account
		(username, password, owner) VALUES ($1, $2, $3)
		ON CONFLICT(username) DO UPDATE SET
		password = $2,
		owner = $3`,
		account.Username, hashedPassword, account.Owner)

	return errors.WithStack(err)
}

// GetAccounts fetch list of account (without its password) based on submitted options.
func (db *PGDatabase) GetAccounts(ctx context.Context, opts GetAccountsOptions) ([]model.Account, error) {
	// Create query
	args := []interface{}{}
	query := `SELECT id, username, owner FROM account WHERE TRUE`

	if opts.Keyword != "" {
		query += " AND username LIKE $1"
		args = append(args, "%"+opts.Keyword+"%")
	}

	if opts.Owner {
		query += " AND owner = TRUE"
	}

	query += ` ORDER BY username`

	// Fetch list account
	accounts := []model.Account{}
	err := db.SelectContext(ctx, &accounts, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WithStack(err)
	}

	return accounts, nil
}

// GetAccount fetch account with matching username.
// Returns the account and boolean whether it's exist or not.
func (db *PGDatabase) GetAccount(ctx context.Context, username string) (model.Account, bool, error) {
	account := model.Account{}
	if err := db.GetContext(ctx, &account, `SELECT
		id, username, password, owner FROM account WHERE username = $1`,
		username,
	); err != nil {
		return account, false, errors.WithStack(err)
	}

	return account, account.ID != 0, nil
}

// DeleteAccounts removes all record with matching usernames.
func (db *PGDatabase) DeleteAccounts(ctx context.Context, usernames ...string) (err error) {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Delete account
		stmtDelete, _ := tx.Preparex(`DELETE FROM account WHERE username = $1`)
		for _, username := range usernames {
			if _, err := stmtDelete.ExecContext(ctx, username); err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// GetTags fetch list of tags and their frequency.
func (db *PGDatabase) GetTags(ctx context.Context) ([]model.Tag, error) {
	tags := []model.Tag{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) n_bookmarks
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

// CreateNewID creates new ID for specified table
func (db *PGDatabase) CreateNewID(ctx context.Context, table string) (int, error) {
	var tableID int
	query := fmt.Sprintf(`SELECT last_value from %s_id_seq;`, table)

	err := db.GetContext(ctx, &tableID, query)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	return tableID, nil
}
