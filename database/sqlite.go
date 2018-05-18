package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RadhiFadlillah/shiori/model"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// SQLiteDatabase is implementation of Database interface for connecting to SQLite3 database.
type SQLiteDatabase struct {
	sqlx.DB
}

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenSQLiteDatabase(databasePath string) (*SQLiteDatabase, error) {
	// Open database and start transaction
	var err error
	db := sqlx.MustConnect("sqlite3", databasePath)
	tx := db.MustBegin()

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			fmt.Println("Database error:", panicErr)
			tx.Rollback()

			db = nil
			err = panicErr
		}
	}()

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
		image_url TEXT NOT NULL DEFAULT "",
		excerpt TEXT NOT NULL DEFAULT "",
		author TEXT NOT NULL DEFAULT "",
		min_read_time INTEGER NOT NULL DEFAULT 0,
		max_read_time INTEGER NOT NULL DEFAULT 0,
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

// CreateBookmark saves new bookmark to database. Returns new ID and error if any happened.
func (db *SQLiteDatabase) CreateBookmark(bookmark model.Bookmark) (bookmarkID int64, err error) {
	// Check URL and title
	if bookmark.URL == "" {
		return -1, fmt.Errorf("URL must not be empty")
	}

	if bookmark.Title == "" {
		return -1, fmt.Errorf("Title must not be empty")
	}

	// Set default ID and modified time
	if bookmark.ID == 0 {
		bookmark.ID, err = db.GetNewID("bookmark")
		if err != nil {
			return -1, err
		}
	}

	if bookmark.Modified == "" {
		bookmark.Modified = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	// Prepare transaction
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
		id, url, title, image_url, excerpt, author, 
		min_read_time, max_read_time, modified) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		bookmark.ID,
		bookmark.URL,
		bookmark.Title,
		bookmark.ImageURL,
		bookmark.Excerpt,
		bookmark.Author,
		bookmark.MinReadTime,
		bookmark.MaxReadTime,
		bookmark.Modified)

	// Save bookmark content
	tx.MustExec(`INSERT INTO bookmark_content 
		(docid, title, content, html) VALUES (?, ?, ?, ?)`,
		bookmark.ID, bookmark.Title, bookmark.Content, bookmark.HTML)

	// Save tags
	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
	checkError(err)

	stmtInsertBookmarkTag, err := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag (tag_id, bookmark_id) VALUES (?, ?)`)
	checkError(err)

	for _, tag := range bookmark.Tags {
		tagName := strings.ToLower(tag.Name)
		tagName = strings.TrimSpace(tagName)

		tagID := int64(-1)
		err = stmtGetTag.Get(&tagID, tagName)
		checkError(err)

		if tagID == -1 {
			res := stmtInsertTag.MustExec(tagName)
			tagID, err = res.LastInsertId()
			checkError(err)
		}

		stmtInsertBookmarkTag.Exec(tagID, bookmark.ID)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	bookmarkID = bookmark.ID
	return bookmarkID, err
}

// GetBookmarks fetch list of bookmarks based on submitted indices.
func (db *SQLiteDatabase) GetBookmarks(withContent bool, indices ...string) ([]model.Bookmark, error) {
	// Get list of index
	listIndex, err := parseIndexList(indices)
	if err != nil {
		return nil, err
	}

	// Prepare where clause
	args := []interface{}{}
	whereClause := " WHERE 1"

	if len(listIndex) > 0 {
		whereClause = " WHERE id IN ("
		for _, idx := range listIndex {
			args = append(args, idx)
			whereClause += "?,"
		}

		whereClause = whereClause[:len(whereClause)-1]
		whereClause += ")"
	}

	// Fetch bookmarks
	query := `SELECT id, 
		url, title, image_url, excerpt, author, 
		min_read_time, max_read_time, modified
		FROM bookmark` + whereClause

	bookmarks := []model.Bookmark{}
	err = db.Select(&bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Fetch tags and contents for each bookmarks
	stmtGetTags, err := db.Preparex(`SELECT t.id, t.name 
		FROM bookmark_tag bt LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id = ? ORDER BY t.name`)
	if err != nil {
		return nil, err
	}

	stmtGetContent, err := db.Preparex(`SELECT title, content, html FROM bookmark_content WHERE docid = ?`)
	if err != nil {
		return nil, err
	}

	defer stmtGetTags.Close()
	defer stmtGetContent.Close()

	for i, book := range bookmarks {
		book.Tags = []model.Tag{}
		err = stmtGetTags.Select(&book.Tags, book.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if withContent {
			err = stmtGetContent.Get(&book, book.ID)
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}
		}

		bookmarks[i] = book
	}

	return bookmarks, nil
}

// DeleteBookmarks removes all record with matching indices from database.
func (db *SQLiteDatabase) DeleteBookmarks(indices ...string) (err error) {
	// Get list of index
	listIndex, err := parseIndexList(indices)
	if err != nil {
		return err
	}

	// Create args and where clause
	args := []interface{}{}
	whereClause := " WHERE 1"

	if len(listIndex) > 0 {
		whereClause = " WHERE id IN ("
		for _, idx := range listIndex {
			args = append(args, idx)
			whereClause += "?,"
		}

		whereClause = whereClause[:len(whereClause)-1]
		whereClause += ")"
	}

	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return ErrInvalidIndex
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			err = panicErr
		}
	}()

	// Delete bookmarks
	whereTagClause := strings.Replace(whereClause, "id", "bookmark_id", 1)
	whereContentClause := strings.Replace(whereClause, "id", "docid", 1)

	tx.MustExec("DELETE FROM bookmark "+whereClause, args...)
	tx.MustExec("DELETE FROM bookmark_tag "+whereTagClause, args...)
	tx.MustExec("DELETE FROM bookmark_content "+whereContentClause, args...)

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return err
}

// SearchBookmarks search bookmarks by the keyword or tags.
func (db *SQLiteDatabase) SearchBookmarks(orderLatest bool, keyword string, tags ...string) ([]model.Bookmark, error) {
	// Create initial variable
	keyword = strings.TrimSpace(keyword)
	whereClause := "WHERE 1"
	args := []interface{}{}

	// Create where clause for keyword
	if keyword != "" {
		whereClause += ` AND (url LIKE ? OR id IN (
			SELECT docid id FROM bookmark_content 
			WHERE title MATCH ? OR content MATCH ?))`
		args = append(args, "%"+keyword+"%", keyword, keyword)
	}

	// Create where clause for tags
	if len(tags) > 0 {
		whereTagClause := ` AND id IN (
			SELECT bookmark_id FROM bookmark_tag 
			WHERE tag_id IN (SELECT id FROM tag WHERE name IN (`

		for _, tag := range tags {
			args = append(args, tag)
			whereTagClause += "?,"
		}

		whereTagClause = whereTagClause[:len(whereTagClause)-1]
		whereTagClause += `)) GROUP BY bookmark_id HAVING COUNT(bookmark_id) >= ?)`
		args = append(args, len(tags))

		whereClause += whereTagClause
	}

	// Search bookmarks
	query := `SELECT id, 
		url, title, image_url, excerpt, author, 
		min_read_time, max_read_time, modified
		FROM bookmark ` + whereClause

	if orderLatest {
		query += ` ORDER BY id DESC`
	}

	bookmarks := []model.Bookmark{}
	err := db.Select(&bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Fetch tags for each bookmarks
	stmtGetTags, err := db.Preparex(`SELECT t.id, t.name 
		FROM bookmark_tag bt LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id = ? ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer stmtGetTags.Close()

	for i := range bookmarks {
		tags := []model.Tag{}
		err = stmtGetTags.Select(&tags, bookmarks[i].ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		bookmarks[i].Tags = tags
	}

	return bookmarks, nil
}

// UpdateBookmarks updates the saved bookmark in database.
func (db *SQLiteDatabase) UpdateBookmarks(bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
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
	stmtUpdateBookmark, err := tx.Preparex(`UPDATE bookmark SET
		url = ?, title = ?, image_url = ?, excerpt = ?, author = ?,
		min_read_time = ?, max_read_time = ?, modified = ? WHERE id = ?`)
	checkError(err)

	stmtUpdateBookmarkContent, err := tx.Preparex(`UPDATE bookmark_content SET
		title = ?, content = ?, html = ? WHERE docid = ?`)
	checkError(err)

	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
	checkError(err)

	stmtInsertBookmarkTag, err := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag (tag_id, bookmark_id) VALUES (?, ?)`)
	checkError(err)

	stmtDeleteBookmarkTag, err := tx.Preparex(`DELETE FROM bookmark_tag WHERE bookmark_id = ? AND tag_id = ?`)
	checkError(err)

	result = []model.Bookmark{}
	for _, book := range bookmarks {
		stmtUpdateBookmark.MustExec(
			book.URL,
			book.Title,
			book.ImageURL,
			book.Excerpt,
			book.Author,
			book.MinReadTime,
			book.MaxReadTime,
			book.Modified,
			book.ID)

		stmtUpdateBookmarkContent.MustExec(
			book.Title,
			book.Content,
			book.HTML,
			book.ID)

		newTags := []model.Tag{}
		for _, tag := range book.Tags {
			if tag.Deleted {
				stmtDeleteBookmarkTag.MustExec(book.ID, tag.ID)
				continue
			}

			if tag.ID == 0 {
				tagID := int64(-1)
				err = stmtGetTag.Get(&tagID, tag.Name)
				checkError(err)

				if tagID == -1 {
					res := stmtInsertTag.MustExec(tag.Name)
					tagID, err = res.LastInsertId()
					checkError(err)
				}

				stmtInsertBookmarkTag.Exec(tagID, book.ID)
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

// CreateAccount saves new account to database. Returns new ID and error if any happened.
func (db *SQLiteDatabase) CreateAccount(username, password string) (err error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}

	// Insert account to database
	_, err = db.Exec(`INSERT INTO account
		(username, password) VALUES (?, ?)`,
		username, hashedPassword)

	return err
}

// GetAccount fetch account with matching username
func (db *SQLiteDatabase) GetAccount(username string) (model.Account, error) {
	account := model.Account{}
	err := db.Get(&account,
		`SELECT id, username, password FROM account WHERE username = ?`,
		username)
	return account, err
}

// GetAccounts fetch list of accounts with matching keyword
func (db *SQLiteDatabase) GetAccounts(keyword string) ([]model.Account, error) {
	// Create query
	args := []interface{}{}
	query := `SELECT id, username, password FROM account`

	if keyword == "" {
		query += " WHERE 1"
	} else {
		query += " WHERE username LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	query += ` ORDER BY username`

	// Fetch list account
	accounts := []model.Account{}
	err := db.Select(&accounts, query, args...)
	return accounts, err
}

// DeleteAccounts removes all record with matching usernames
func (db *SQLiteDatabase) DeleteAccounts(usernames ...string) error {
	// Prepare where clause
	args := []interface{}{}
	whereClause := " WHERE 1"

	if len(usernames) > 0 {
		whereClause = " WHERE username IN ("
		for _, username := range usernames {
			args = append(args, username)
			whereClause += "?,"
		}

		whereClause = whereClause[:len(whereClause)-1]
		whereClause += ")"
	}

	// Delete usernames
	_, err := db.Exec(`DELETE FROM account `+whereClause, args...)
	return err
}

// GetTags fetch list of tags and their frequency
func (db *SQLiteDatabase) GetTags() ([]model.Tag, error) {
	tags := []model.Tag{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) n_bookmarks 
		FROM bookmark_tag bt 
		LEFT JOIN tag t ON bt.tag_id = t.id
		GROUP BY bt.tag_id ORDER BY t.name`

	err := db.Select(&tags, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return tags, nil
}

// GetNewID creates new ID for specified table
func (db *SQLiteDatabase) GetNewID(table string) (int64, error) {
	var tableID int64
	query := fmt.Sprintf(`SELECT IFNULL(MAX(id) + 1, 1) FROM %s`, table)

	err := db.Get(&tableID, query)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	return tableID, nil
}

// ErrInvalidIndex is returned is an index is not valid
var ErrInvalidIndex = errors.New("Index is not valid")

// parseIndexList converts a list of indices to their integer values
func parseIndexList(indices []string) ([]int, error) {
	var listIndex []int
	for _, strIndex := range indices {
		if !strings.Contains(strIndex, "-") {
			index, err := strconv.Atoi(strIndex)
			if err != nil || index < 1 {
				return nil, ErrInvalidIndex
			}

			listIndex = append(listIndex, index)
			continue
		}

		parts := strings.Split(strIndex, "-")
		if len(parts) != 2 {
			return nil, ErrInvalidIndex
		}

		minIndex, errMin := strconv.Atoi(parts[0])
		maxIndex, errMax := strconv.Atoi(parts[1])
		if errMin != nil || errMax != nil || minIndex < 1 || minIndex > maxIndex {
			return nil, ErrInvalidIndex
		}

		for i := minIndex; i <= maxIndex; i++ {
			listIndex = append(listIndex, i)
		}
	}
	return listIndex, nil
}
