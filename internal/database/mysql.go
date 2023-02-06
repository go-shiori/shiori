package database

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

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

	mysqlDB = &MySQLDatabase{dbbase: dbbase{*db}}
	return mysqlDB, err
}

// Migrate runs migrations for this database engine
func (db *MySQLDatabase) Migrate() error {
	sourceDriver, err := iofs.New(migrations, "migrations/mysql")
	if err != nil {
		return errors.WithStack(err)
	}

	dbDriver, err := mysql.WithInstance(db.DB.DB, &mysql.Config{})
	if err != nil {
		return errors.WithStack(err)
	}

	migration, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"mysql",
		dbDriver,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *MySQLDatabase) SaveBookmarks(ctx context.Context, create bool, bookmarks ...model.Bookmark) ([]model.Bookmark, error) {
	var result []model.Bookmark

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Prepare statement
		stmtInsertBook, err := tx.Preparex(`INSERT INTO bookmark
			(url, title, excerpt, author, public, content, html, modified)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?)`)
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
			modified = ?
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
			if book.Modified == "" {
				book.Modified = modifiedTime
			}

			// Save bookmark
			var err error
			if create {
				var res sql.Result
				res, err = stmtInsertBook.ExecContext(ctx,
					book.URL, book.Title, book.Excerpt, book.Author,
					book.Public, book.Content, book.HTML, book.Modified)
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
					book.Public, book.Content, book.HTML, book.Modified, book.ID)
			}
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
		return result, errors.WithStack(err)
	}

	return result, nil
}

// GetBookmarks fetch list of bookmarks based on submitted options.
func (db *MySQLDatabase) GetBookmarks(ctx context.Context, opts GetBookmarksOptions) ([]model.Bookmark, error) {
	// Create initial query
	columns := []string{
		`id`,
		`url`,
		`title`,
		`excerpt`,
		`author`,
		`public`,
		`modified`,
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
	case ByLastAdded:
		query += ` ORDER BY id DESC`
	case ByLastModified:
		query += ` ORDER BY modified DESC`
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
	bookmarks := []model.Bookmark{}
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

	for i, book := range bookmarks {
		book.Tags = []model.Tag{}
		err = stmtGetTags.SelectContext(ctx, &book.Tags, book.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.WithStack(err)
		}

		bookmarks[i] = book
	}

	return bookmarks, nil
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *MySQLDatabase) GetBookmarksCount(ctx context.Context, opts GetBookmarksOptions) (int, error) {
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

// GetBookmark fetchs bookmark based on its ID or URL.
// Returns the bookmark and boolean whether it's exist or not.
func (db *MySQLDatabase) GetBookmark(ctx context.Context, id int, url string) (model.Bookmark, bool, error) {
	args := []interface{}{id}
	query := `SELECT
		id, url, title, excerpt, author, public,
		content, html, modified, content <> '' has_content
		FROM bookmark WHERE id = ?`

	if url != "" {
		query += ` OR url = ?`
		args = append(args, url)
	}

	book := model.Bookmark{}
	if err := db.GetContext(ctx, &book, query, args...); err != nil && err != sql.ErrNoRows {
		return book, false, errors.WithStack(err)
	}

	return book, book.ID != 0, nil
}

// SaveAccount saves new account to database. Returns error if any happened.
func (db *MySQLDatabase) SaveAccount(ctx context.Context, account model.Account) (err error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), 10)
	if err != nil {
		return errors.WithStack(err)
	}

	// Insert account to database
	_, err = db.ExecContext(ctx, `INSERT INTO account
		(username, password, owner) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
		password = VALUES(password),
		owner = VALUES(owner)`,
		account.Username, hashedPassword, account.Owner)

	return errors.WithStack(err)
}

// GetAccounts fetch list of account (without its password) based on submitted options.
func (db *MySQLDatabase) GetAccounts(ctx context.Context, opts GetAccountsOptions) ([]model.Account, error) {
	// Create query
	args := []interface{}{}
	query := `SELECT id, username, owner FROM account WHERE 1`

	if opts.Keyword != "" {
		query += " AND username LIKE ?"
		args = append(args, "%"+opts.Keyword+"%")
	}

	if opts.Owner {
		query += " AND owner = 1"
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
func (db *MySQLDatabase) GetAccount(ctx context.Context, username string) (model.Account, bool, error) {
	account := model.Account{}
	if err := db.GetContext(ctx, &account, `SELECT
		id, username, password, owner FROM account WHERE username = ?`,
		username,
	); err != nil {
		return account, false, errors.WithStack(err)
	}

	return account, account.ID != 0, nil
}

// DeleteAccounts removes all record with matching usernames.
func (db *MySQLDatabase) DeleteAccounts(ctx context.Context, usernames ...string) error {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Delete account
		stmtDelete, _ := tx.Preparex(`DELETE FROM account WHERE username = ?`)
		for _, username := range usernames {
			_, err := stmtDelete.ExecContext(ctx, username)
			if err != nil {
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
func (db *MySQLDatabase) GetTags(ctx context.Context) ([]model.Tag, error) {
	tags := []model.Tag{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) n_bookmarks
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
