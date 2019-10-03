package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// PGDatabase is implementation of Database interface
// for connecting to PostgreSQL database.
type PGDatabase struct {
	sqlx.DB
}

// OpenPGDatabase creates and opens connection to a PostgreSQL Database.
func OpenPGDatabase(connString string) (pgDB *PGDatabase, err error) {
	// Open database and start transaction
	db := sqlx.MustConnect("postgres", connString)
	db.SetMaxOpenConns(100)

	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			pgDB = nil
			err = panicErr
		}
	}()

	// Create tables
	tx.MustExec(`CREATE TABLE IF NOT EXISTS account(
		id       SERIAL,
		username VARCHAR(250) NOT NULL,
		password BYTEA    NOT NULL,
		owner    BOOLEAN  NOT NULL DEFAULT FALSE,
		PRIMARY KEY (id),
		CONSTRAINT account_username_UNIQUE UNIQUE (username))`)

	tx.MustExec(`CREATE TABLE IF NOT EXISTS bookmark(
		id       SERIAL,
		url      TEXT       NOT NULL,
		title    TEXT       NOT NULL,
		excerpt  TEXT       NOT NULL DEFAULT '',
		author   TEXT       NOT NULL DEFAULT '',
		public   SMALLINT   NOT NULL DEFAULT 0,
		content  TEXT       NOT NULL DEFAULT '',
		html     TEXT       NOT NULL DEFAULT '',
		modified TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(id),
		CONSTRAINT bookmark_url_UNIQUE UNIQUE (url))`)

	tx.MustExec(`CREATE TABLE IF NOT EXISTS tag(
		id   SERIAL,
		name VARCHAR(250) NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT tag_name_UNIQUE UNIQUE (name))`)

	tx.MustExec(`CREATE TABLE IF NOT EXISTS bookmark_tag(
		bookmark_id INT      NOT NULL,
		tag_id      INT      NOT NULL,
		PRIMARY KEY(bookmark_id, tag_id),
		CONSTRAINT bookmark_tag_bookmark_id_FK FOREIGN KEY (bookmark_id) REFERENCES bookmark (id),
		CONSTRAINT bookmark_tag_tag_id_FK FOREIGN KEY (tag_id) REFERENCES tag (id))`)

	// Create indices
	tx.MustExec(`CREATE INDEX IF NOT EXISTS bookmark_tag_bookmark_id_FK ON bookmark_tag (bookmark_id)`)
	tx.MustExec(`CREATE INDEX IF NOT EXISTS bookmark_tag_tag_id_FK ON bookmark_tag (tag_id)`)

	err = tx.Commit()
	checkError(err)

	pgDB = &PGDatabase{*db}
	return pgDB, err
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *PGDatabase) SaveBookmarks(bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
	// Prepare transaction
	tx, err := db.Beginx()
	if err != nil {
		return []model.Bookmark{}, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			result = []model.Bookmark{}
			err = panicErr
		}
	}()

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
		modified = $8`)
	checkError(err)

	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = $1`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES ($1) RETURNING id`)
	checkError(err)

	stmtInsertBookTag, err := tx.Preparex(`INSERT INTO bookmark_tag
		(tag_id, bookmark_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	checkError(err)

	stmtDeleteBookTag, err := tx.Preparex(`DELETE FROM bookmark_tag
		WHERE bookmark_id = $1 AND tag_id = $2`)
	checkError(err)

	// Prepare modified time
	modifiedTime := time.Now().UTC().Format("2006-01-02 15:04:05")

	// Execute statements
	result = []model.Bookmark{}
	for _, book := range bookmarks {
		// Check ID, URL and title
		if book.ID == 0 {
			panic(fmt.Errorf("ID must not be empty"))
		}

		if book.URL == "" {
			panic(fmt.Errorf("URL must not be empty"))
		}

		if book.Title == "" {
			panic(fmt.Errorf("title must not be empty"))
		}

		// Set modified time
		book.Modified = modifiedTime

		// Save bookmark
		stmtInsertBook.MustExec(
			book.URL, book.Title, book.Excerpt, book.Author,
			book.Public, book.Content, book.HTML, book.Modified)

		// Save book tags
		newTags := []model.Tag{}
		for _, tag := range book.Tags {
			// If it's deleted tag, delete and continue
			if tag.Deleted {
				stmtDeleteBookTag.MustExec(book.ID, tag.ID)
				continue
			}

			// Normalize tag name
			tagName := strings.ToLower(tag.Name)
			tagName = strings.Join(strings.Fields(tagName), " ")

			// If tag doesn't have any ID, fetch it from database
			if tag.ID == 0 {
				err = stmtGetTag.Get(&tag.ID, tagName)
				checkError(err)

				// If tag doesn't exist in database, save it
				if tag.ID == 0 {
					var tagID64 int64
					err = stmtInsertTag.Get(&tagID64, tagName)
					checkError(err)

					tag.ID = int(tagID64)
				}

				stmtInsertBookTag.Exec(tag.ID, book.ID)
			}

			newTags = append(newTags, tag)
		}

		book.Tags = newTags
		result = append(result, book)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return result, err
}

// GetBookmarks fetch list of bookmarks based on submitted options.
func (db *PGDatabase) GetBookmarks(opts GetBookmarksOptions) ([]model.Bookmark, error) {
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
			MATCH(title, excerpt, content) AGAINST (:kw IN BOOLEAN MODE)
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
	query, args, err := sqlx.Named(query, arg)
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to expand query: %v", err)
	}
	query = db.Rebind(query)

	// Fetch bookmarks
	bookmarks := []model.Bookmark{}
	err = db.Select(&bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}

	// Fetch tags for each bookmarks
	stmtGetTags, err := db.Preparex(`SELECT t.id, t.name 
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
		err = stmtGetTags.Select(&book.Tags, book.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch tags: %v", err)
		}

		bookmarks[i] = book
	}

	return bookmarks, nil
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *PGDatabase) GetBookmarksCount(opts GetBookmarksOptions) (int, error) {
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
			MATCH(title, excerpt, content) AGAINST (:kw IN BOOLEAN MODE)
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
	query, args, err := sqlx.Named(query, arg)
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to expand query: %v", err)
	}
	query = db.Rebind(query)

	// Fetch count
	var nBookmarks int
	err = db.Get(&nBookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to fetch count: %v", err)
	}

	return nBookmarks, nil
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *PGDatabase) DeleteBookmarks(ids ...int) (err error) {
	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			err = panicErr
		}
	}()

	// Prepare queries
	delBookmark := `DELETE FROM bookmark`
	delBookmarkTag := `DELETE FROM bookmark_tag`

	// Delete bookmark(s)
	if len(ids) == 0 {
		tx.MustExec(delBookmarkTag)
		tx.MustExec(delBookmark)
	} else {
		delBookmark += ` WHERE id = $1`
		delBookmarkTag += ` WHERE bookmark_id = $1`

		stmtDelBookmark, _ := tx.Preparex(delBookmark)
		stmtDelBookmarkTag, _ := tx.Preparex(delBookmarkTag)

		for _, id := range ids {
			stmtDelBookmarkTag.MustExec(id)
			stmtDelBookmark.MustExec(id)
		}
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return err
}

// GetBookmark fetchs bookmark based on its ID or URL.
// Returns the bookmark and boolean whether it's exist or not.
func (db *PGDatabase) GetBookmark(id int, url string) (model.Bookmark, bool) {
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
	db.Get(&book, query, args...)

	return book, book.ID != 0
}

// SaveAccount saves new account to database. Returns error if any happened.
func (db *PGDatabase) SaveAccount(account model.Account) (err error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), 10)
	if err != nil {
		return err
	}

	// Insert account to database
	_, err = db.Exec(`INSERT INTO account
		(username, password, owner) VALUES ($1, $2, $3)
		ON CONFLICT(username) DO UPDATE SET
		password = $2,
		owner = $3`,
		account.Username, hashedPassword, account.Owner)

	return err
}

// GetAccounts fetch list of account (without its password) based on submitted options.
func (db *PGDatabase) GetAccounts(opts GetAccountsOptions) ([]model.Account, error) {
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
	err := db.Select(&accounts, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	return accounts, nil
}

// GetAccount fetch account with matching username.
// Returns the account and boolean whether it's exist or not.
func (db *PGDatabase) GetAccount(username string) (model.Account, bool) {
	account := model.Account{}
	db.Get(&account, `SELECT 
		id, username, password, owner FROM account WHERE username = $1`,
		username)

	return account, account.ID != 0
}

// DeleteAccounts removes all record with matching usernames.
func (db *PGDatabase) DeleteAccounts(usernames ...string) (err error) {
	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			err = panicErr
		}
	}()

	// Delete account
	stmtDelete, _ := tx.Preparex(`DELETE FROM account WHERE username = $1`)
	for _, username := range usernames {
		stmtDelete.MustExec(username)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return err
}

// GetTags fetch list of tags and their frequency.
func (db *PGDatabase) GetTags() ([]model.Tag, error) {
	tags := []model.Tag{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) n_bookmarks 
		FROM bookmark_tag bt 
		LEFT JOIN tag t ON bt.tag_id = t.id
		GROUP BY bt.tag_id, t.name ORDER BY t.name`

	err := db.Select(&tags, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch tags: %v", err)
	}

	return tags, nil
}

// RenameTag change the name of a tag.
func (db *PGDatabase) RenameTag(id int, newName string) error {
	_, err := db.Exec(`UPDATE tag SET name = $1 WHERE id = $2`, newName, id)
	return err
}

// CreateNewID creates new ID for specified table
func (db *PGDatabase) CreateNewID(table string) (int, error) {
	var tableID int
	query := fmt.Sprintf(`SELECT last_value from %s_id_seq;`, table)

	err := db.Get(&tableID, query)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	return tableID, nil
}
