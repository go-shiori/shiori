package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// SQLiteDatabase is implementation of Database interface
// for connecting to SQLite3 database.
type SQLiteDatabase struct {
	sqlx.DB
}

type BookmarkContent struct {
	ID      int    `db:"docid"`
	Content string `db:"content"`
	HTML    string `db:"html"`
}

type TagContent struct {
	ID int `db:"bookmark_id"`
	model.Tag
}

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenSQLiteDatabase(databasePath string) (sqliteDB *SQLiteDatabase, err error) {
	// Open database
	db := sqlx.MustConnect("sqlite", databasePath)
	sqliteDB = &SQLiteDatabase{*db}
	return sqliteDB, err
}

// Migrate runs migrations for this database engine
func (db *SQLiteDatabase) Migrate() error {
	sourceDriver, err := iofs.New(migrations, "migrations/sqlite")
	checkError(err)

	dbDriver, err := sqlite.WithInstance(db.DB.DB, &sqlite.Config{})
	checkError(err)

	migration, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"sqlite",
		dbDriver,
	)

	checkError(err)

	return migration.Up()
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *SQLiteDatabase) SaveBookmarks(bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
	// Prepare transaction
	tx, err := db.Beginx()
	if err != nil {
		return []model.Bookmark{}, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			if err := tx.Rollback(); err != nil {
				log.Printf("error during rollback: %s", err)
			}
			result = []model.Bookmark{}
			err = panicErr
		}
	}()

	// Prepare statement
	stmtInsertBook, _ := tx.Preparex(`INSERT INTO bookmark
		(id, url, title, excerpt, author, public, modified, has_content)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		url = ?, title = ?,	excerpt = ?, author = ?,
		public = ?, modified = ?, has_content = ?`)

	stmtInsertBookContent, _ := tx.Preparex(`INSERT OR REPLACE INTO bookmark_content
		(docid, title, content, html)
		VALUES (?, ?, ?, ?)`)

	stmtUpdateBookContent, _ := tx.Preparex(`UPDATE bookmark_content SET
		title = ?, content = ?, html = ?
		WHERE docid = ?`)

	stmtGetTag, _ := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)

	stmtInsertTag, _ := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)

	stmtInsertBookTag, _ := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag
		(tag_id, bookmark_id) VALUES (?, ?)`)

	stmtDeleteBookTag, _ := tx.Preparex(`DELETE FROM bookmark_tag
		WHERE bookmark_id = ? AND tag_id = ?`)

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
		hasContent := book.Content != ""
		stmtInsertBook.MustExec(book.ID,
			book.URL, book.Title, book.Excerpt, book.Author, book.Public, book.Modified, hasContent,
			book.URL, book.Title, book.Excerpt, book.Author, book.Public, book.Modified, hasContent)

		// Try to update it first to check for existence, we can't do an UPSERT here because
		// bookmark_content is a virtual table
		res := stmtUpdateBookContent.MustExec(book.Title, book.Content, book.HTML, book.ID)
		rows, _ := res.RowsAffected()
		if rows == 0 {
			stmtInsertBookContent.MustExec(book.ID, book.Title, book.Content, book.HTML)
		}

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
					res := stmtInsertTag.MustExec(tagName)
					tagID64, err := res.LastInsertId()
					checkError(err)

					tag.ID = int(tagID64)
				}

				if _, err := stmtInsertBookTag.Exec(tag.ID, book.ID); err != nil {
					log.Printf("error during insert: %s", err)
				}
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
func (db *SQLiteDatabase) GetBookmarks(opts GetBookmarksOptions) ([]model.Bookmark, error) {
	// Create initial query
	query := `SELECT 
		b.id,
		b.url,
		b.title,
		b.excerpt,
		b.author,
		b.public,
		b.modified,
		b.has_content
		FROM bookmark b
		WHERE 1`

	// Add where clause
	args := []interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND b.id IN (?)`
		args = append(args, opts.IDs)
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (b.url LIKE ? OR b.excerpt LIKE ? OR b.id IN (
			SELECT docid id
			FROM bookmark_content
			WHERE title MATCH ? OR content MATCH ?))`

		args = append(args,
			"%"+opts.Keyword+"%",
			"%"+opts.Keyword+"%",
			opts.Keyword,
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
		query += ` AND b.id NOT IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	} else if includeAllTags {
		query += ` AND b.id IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	}

	// Now we only need to find the normal tags
	if len(opts.Tags) > 0 {
		query += ` AND b.id IN (
			SELECT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = ?)`

		args = append(args, opts.Tags, len(opts.Tags))
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND b.id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?))`

		args = append(args, opts.ExcludedTags)
	}

	// Add order clause
	switch opts.OrderMethod {
	case ByLastAdded:
		query += ` ORDER BY b.id DESC`
	case ByLastModified:
		query += ` ORDER BY b.modified DESC`
	default:
		query += ` ORDER BY b.id`
	}

	if opts.Limit > 0 && opts.Offset >= 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, opts.Limit, opts.Offset)
	}

	// Expand query, because some of the args might be an array
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to expand query: %v", err)
	}

	// Fetch bookmarks
	bookmarks := []model.Bookmark{}
	err = db.Select(&bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}

	// store bookmark IDs for further enrichment
	var bookmarkIds = make([]int, 0, len(bookmarks))
	for _, book := range bookmarks {
		bookmarkIds = append(bookmarkIds, book.ID)
	}

	// If content needed, fetch it separately
	// It's faster than join with virtual table
	if opts.WithContent {
		contents := make([]BookmarkContent, 0, len(bookmarks))
		contentMap := make(map[int]BookmarkContent, len(bookmarks))

		contentQuery, args, err := sqlx.In(`SELECT docid, content, html FROM bookmark_content WHERE docid IN (?)`, bookmarkIds)
		contentQuery = db.Rebind(contentQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to expand query: %v", err)
		}

		err = db.Select(&contents, contentQuery, args...)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch content for bookmarks (%v): %v", bookmarkIds, err)
		}
		for _, content := range contents {
			contentMap[content.ID] = content
		}
		for i := range bookmarks[:] {
			book := &bookmarks[i]
			if bookmarkContent, found := contentMap[book.ID]; found {
				book.Content = bookmarkContent.Content
				book.HTML = bookmarkContent.HTML
			} else {
				log.Printf("not found content for bookmark %d, but it should be; check DB consistency", book.ID)
			}
		}

	}

	// Fetch tags for each bookmark
	tags := make([]TagContent, 0, len(bookmarks))
	tagsMap := make(map[int][]model.Tag, len(bookmarks))

	tagsQuery, tagArgs, err := sqlx.In(`SELECT bt.bookmark_id, t.id, t.name
		FROM bookmark_tag bt
		LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id IN (?)
		ORDER BY t.name`, bookmarkIds)
	tagsQuery = db.Rebind(tagsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to expand query: %v", err)
	}

	err = db.Select(&tags, tagsQuery, tagArgs...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch tags for bookmarks (%v): %v", bookmarkIds, err)
	}
	for _, fetchedTag := range tags {
		if tags, found := tagsMap[fetchedTag.ID]; found {
			tagsMap[fetchedTag.ID] = append(tags, fetchedTag.Tag)
		} else {
			tagsMap[fetchedTag.ID] = []model.Tag{fetchedTag.Tag}
		}
	}
	for i := range bookmarks[:] {
		book := &bookmarks[i]
		if tags, found := tagsMap[book.ID]; found {
			book.Tags = tags
		} else {
			book.Tags = []model.Tag{}
		}
	}

	return bookmarks, nil
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *SQLiteDatabase) GetBookmarksCount(opts GetBookmarksOptions) (int, error) {
	// Create initial query
	query := `SELECT COUNT(b.id)
		FROM bookmark b
		WHERE 1`

	// Add where clause
	args := []interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND b.id IN (?)`
		args = append(args, opts.IDs)
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (b.url LIKE ? OR b.excerpt LIKE ? OR b.id IN (
			SELECT docid id
			FROM bookmark_content
			WHERE title MATCH ? OR content MATCH ?))`

		args = append(args,
			"%"+opts.Keyword+"%",
			"%"+opts.Keyword+"%",
			opts.Keyword,
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
		query += ` AND b.id NOT IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	} else if includeAllTags {
		query += ` AND b.id IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	}

	// Now we only need to find the normal tags
	if len(opts.Tags) > 0 {
		query += ` AND b.id IN (
			SELECT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = ?)`

		args = append(args, opts.Tags, len(opts.Tags))
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND b.id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?))`

		args = append(args, opts.ExcludedTags)
	}

	// Expand query, because some of the args might be an array
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to expand query: %v", err)
	}

	// Fetch count
	var nBookmarks int
	err = db.Get(&nBookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to fetch count: %v", err)
	}

	return nBookmarks, nil
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *SQLiteDatabase) DeleteBookmarks(ids ...int) (err error) {
	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			if err := tx.Rollback(); err != nil {
				log.Printf("error during rollback: %s", err)
			}
			err = panicErr
		}
	}()

	// Prepare queries
	delBookmark := `DELETE FROM bookmark`
	delBookmarkTag := `DELETE FROM bookmark_tag`
	delBookmarkContent := `DELETE FROM bookmark_content`

	// Delete bookmark(s)
	if len(ids) == 0 {
		tx.MustExec(delBookmarkContent)
		tx.MustExec(delBookmarkTag)
		tx.MustExec(delBookmark)
	} else {
		delBookmark += ` WHERE id = ?`
		delBookmarkTag += ` WHERE bookmark_id = ?`
		delBookmarkContent += ` WHERE docid = ?`

		stmtDelBookmark, _ := tx.Preparex(delBookmark)
		stmtDelBookmarkTag, _ := tx.Preparex(delBookmarkTag)
		stmtDelBookmarkContent, _ := tx.Preparex(delBookmarkContent)

		for _, id := range ids {
			stmtDelBookmarkContent.MustExec(id)
			stmtDelBookmarkTag.MustExec(id)
			stmtDelBookmark.MustExec(id)
		}
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return err
}

// GetBookmark fetches bookmark based on its ID or URL.
// Returns the bookmark and boolean whether it's exist or not.
func (db *SQLiteDatabase) GetBookmark(id int, url string) (model.Bookmark, bool) {
	args := []interface{}{id}
	query := `SELECT
		b.id, b.url, b.title, b.excerpt, b.author, b.public, b.modified,
		bc.content, bc.html, b.has_content
		FROM bookmark b
		LEFT JOIN bookmark_content bc ON bc.docid = b.id
		WHERE b.id = ?`

	if url != "" {
		query += ` OR b.url = ?`
		args = append(args, url)
	}

	book := model.Bookmark{}
	if err := db.Get(&book, query, args...); err != nil {
		log.Printf("error during db.get: %s", err)
	}

	return book, book.ID != 0
}

// SaveAccount saves new account to database. Returns error if any happened.
func (db *SQLiteDatabase) SaveAccount(account model.Account) (err error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), 10)
	if err != nil {
		return err
	}

	// Insert account to database
	_, err = db.Exec(`INSERT INTO account
		(username, password, owner) VALUES (?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET
		password = ?, owner = ?`,
		account.Username, hashedPassword, account.Owner,
		hashedPassword, account.Owner)

	return err
}

// GetAccounts fetch list of account (without its password) based on submitted options.
func (db *SQLiteDatabase) GetAccounts(opts GetAccountsOptions) ([]model.Account, error) {
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
	err := db.Select(&accounts, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	return accounts, nil
}

// GetAccount fetch account with matching username.
// Returns the account and boolean whether it's exist or not.
func (db *SQLiteDatabase) GetAccount(username string) (model.Account, bool) {
	account := model.Account{}
	if err := db.Get(&account, `SELECT
		id, username, password, owner FROM account WHERE username = ?`,
		username,
	); err != nil {
		log.Printf("error during db.get: %s", err)
	}

	return account, account.ID != 0
}

// DeleteAccounts removes all record with matching usernames.
func (db *SQLiteDatabase) DeleteAccounts(usernames ...string) (err error) {
	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			if err := tx.Rollback(); err != nil {
				log.Printf("error during rollback: %s", err)
			}
			err = panicErr
		}
	}()

	// Delete account
	stmtDelete, _ := tx.Preparex(`DELETE FROM account WHERE username = ?`)
	for _, username := range usernames {
		stmtDelete.MustExec(username)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return err
}

// GetTags fetch list of tags and their frequency.
func (db *SQLiteDatabase) GetTags() ([]model.Tag, error) {
	tags := []model.Tag{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) n_bookmarks
		FROM bookmark_tag bt
		LEFT JOIN tag t ON bt.tag_id = t.id
		GROUP BY bt.tag_id ORDER BY t.name`

	err := db.Select(&tags, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch tags: %v", err)
	}

	return tags, nil
}

// RenameTag change the name of a tag.
func (db *SQLiteDatabase) RenameTag(id int, newName string) error {
	_, err := db.Exec(`UPDATE tag SET name = ? WHERE id = ?`, newName, id)
	return err
}

// CreateNewID creates new ID for specified table
func (db *SQLiteDatabase) CreateNewID(table string) (int, error) {
	var tableID int
	query := fmt.Sprintf(`SELECT IFNULL(MAX(id) + 1, 1) FROM %s`, table)

	err := db.Get(&tableID, query)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	return tableID, nil
}
