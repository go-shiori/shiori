package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
)

// SQLiteDatabase is implementation of Database interface
// for connecting to SQLite3 database.
type SQLiteDatabase struct {
	sqlx.DB
}

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenSQLiteDatabase(databasePath string) (*SQLiteDatabase, error) {
	// Open database and start transaction
	var err error
	db := sqlx.MustConnect("sqlite3", databasePath)

	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			db = nil
			err = panicErr
		}
	}()

	// Create tables
	tx.MustExec(`CREATE TABLE IF NOT EXISTS account(
		id INTEGER NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		CONSTRAINT account_PK PRIMARY KEY(id),
		CONSTRAINT account_username_UNIQUE UNIQUE(username))`)

	tx.MustExec(`CREATE TABLE IF NOT EXISTS bookmark(
		id INTEGER NOT NULL,
		url TEXT NOT NULL,
		title TEXT NOT NULL,
		excerpt TEXT NOT NULL DEFAULT "",
		author TEXT NOT NULL DEFAULT "",
		modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT bookmark_PK PRIMARY KEY(id),
		CONSTRAINT bookmark_url_UNIQUE UNIQUE(url))`)

	tx.MustExec(`CREATE TABLE IF NOT EXISTS tag(
		id INTEGER NOT NULL,
		name TEXT NOT NULL,
		CONSTRAINT tag_PK PRIMARY KEY(id),
		CONSTRAINT tag_name_UNIQUE UNIQUE(name))`)

	tx.MustExec(`CREATE TABLE IF NOT EXISTS bookmark_tag(
		bookmark_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		CONSTRAINT bookmark_tag_PK PRIMARY KEY(bookmark_id, tag_id),
		CONSTRAINT bookmark_id_FK FOREIGN KEY(bookmark_id) REFERENCES bookmark(id),
		CONSTRAINT tag_id_FK FOREIGN KEY(tag_id) REFERENCES tag(id))`)

	tx.MustExec(`CREATE VIRTUAL TABLE IF NOT EXISTS bookmark_content USING fts4(title, content, html)`)

	err = tx.Commit()
	checkError(err)

	return &SQLiteDatabase{*db}, err
}

// InsertBookmark saves new bookmark to database.
// Returns new ID and error message if any happened.
func (db *SQLiteDatabase) InsertBookmark(bookmark model.Bookmark) (bookmarkID int, err error) {
	// Check URL and title
	if bookmark.URL == "" {
		return -1, fmt.Errorf("URL must not be empty")
	}

	if bookmark.Title == "" {
		return -1, fmt.Errorf("title must not be empty")
	}

	// Create ID (if needed) and modified time
	if bookmark.ID != 0 {
		bookmarkID = bookmark.ID
	} else {
		bookmarkID, err = db.CreateNewID("bookmark")
		if err != nil {
			return -1, err
		}
	}

	if bookmark.Modified == "" {
		bookmark.Modified = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return -1, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			bookmarkID = -1
			err = panicErr
		}
	}()

	// Save article to database
	tx.MustExec(`INSERT INTO bookmark (
		id, url, title, excerpt, author, modified) 
		VALUES(?, ?, ?, ?, ?, ?)`,
		bookmarkID,
		bookmark.URL,
		bookmark.Title,
		bookmark.Excerpt,
		bookmark.Author,
		bookmark.Modified)

	// Save bookmark content
	tx.MustExec(`INSERT INTO bookmark_content 
		(docid, title, content, html) VALUES (?, ?, ?, ?)`,
		bookmarkID,
		bookmark.Title,
		bookmark.Content,
		bookmark.HTML)

	// Save tags
	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
	checkError(err)

	stmtInsertBookmarkTag, err := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag 
		(tag_id, bookmark_id) VALUES (?, ?)`)
	checkError(err)

	for _, tag := range bookmark.Tags {
		tagName := strings.ToLower(tag.Name)
		tagName = strings.TrimSpace(tagName)

		tagID := -1
		err = stmtGetTag.Get(&tagID, tagName)
		checkError(err)

		if tagID == -1 {
			res := stmtInsertTag.MustExec(tagName)
			tagID64, err := res.LastInsertId()
			checkError(err)

			tagID = int(tagID64)
		}

		stmtInsertBookmarkTag.Exec(tagID, bookmarkID)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return bookmarkID, err
}

// GetBookmarks fetch list of bookmarks based on submitted ids.
func (db *SQLiteDatabase) GetBookmarks(opts GetBookmarksOptions) ([]model.Bookmark, error) {
	// Create initial query
	columns := []string{
		`b.id`,
		`b.url`,
		`b.title`,
		`b.excerpt`,
		`b.author`,
		`b.modified`,
		`bc.content <> "" has_content`}

	if opts.WithContent {
		columns = append(columns, `bc.content`, `bc.html`)
	}

	query := `SELECT ` + strings.Join(columns, ",") + `
		FROM bookmark b
		LEFT JOIN bookmark_content bc ON bc.docid = b.id
		WHERE 1`

	// Add where clause
	args := []interface{}{}

	if len(opts.IDs) > 0 {
		query += ` AND b.id IN (?)`
		args = append(args, opts.IDs)
	}

	if opts.Keyword != "" {
		query += ` AND (b.url LIKE ? OR b.id IN (
			SELECT docid id 
			FROM bookmark_content 
			WHERE title MATCH ? OR content MATCH ?))`

		args = append(args,
			"%"+opts.Keyword+"%",
			opts.Keyword,
			opts.Keyword)
	}

	if len(opts.Tags) > 0 {
		query += ` AND b.id IN (
			SELECT bookmark_id FROM bookmark_tag 
			WHERE tag_id IN (SELECT id FROM tag WHERE name IN (?)))`

		args = append(args, opts.Tags)
	}

	// Add order clause
	if opts.OrderLatest {
		query += ` ORDER BY b.modified DESC`
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

	// Fetch tags for each bookmarks
	stmtGetTags, err := db.Preparex(`SELECT t.id, t.name 
		FROM bookmark_tag bt 
		LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id = ? 
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
