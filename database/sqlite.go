package database

import (
	"database/sql"
	"fmt"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"sort"
	"strconv"
	"strings"
)

// SQLiteDatabase is implementation of Database interface for connecting to SQLite3 database.
type SQLiteDatabase struct {
	sqlx.DB
}

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenSQLiteDatabase() (*SQLiteDatabase, error) {
	// Open database and start transaction
	var err error
	db := sqlx.MustConnect("sqlite3", "shiori.db")
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
		return -1, fmt.Errorf("URL must not empty")
	}

	if bookmark.Title == "" {
		return -1, fmt.Errorf("Title must not empty")
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
	res := tx.MustExec(`INSERT INTO bookmark (
		url, title, image_url, excerpt, author, 
		min_read_time, max_read_time) 
		VALUES(?, ?, ?, ?, ?, ?, ?)`,
		bookmark.URL,
		bookmark.Title,
		bookmark.ImageURL,
		bookmark.Excerpt,
		bookmark.Author,
		bookmark.MinReadTime,
		bookmark.MaxReadTime)

	// Get last inserted ID
	bookmarkID, err = res.LastInsertId()
	checkError(err)

	// Save bookmark content
	tx.MustExec(`INSERT INTO bookmark_content 
		(docid, title, content, html) VALUES (?, ?, ?, ?)`,
		bookmarkID, bookmark.Title, bookmark.Content, bookmark.HTML)

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

		stmtInsertBookmarkTag.Exec(tagID, bookmarkID)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return bookmarkID, err
}

// GetBookmarks fetch list of bookmarks based on submitted indices.
func (db *SQLiteDatabase) GetBookmarks(options GetBookmarksOptions, indices ...string) ([]model.Bookmark, error) {
	// Convert list of index to int
	listIndex := []int{}
	errInvalidIndex := fmt.Errorf("Index is not valid")

	for _, strIndex := range indices {
		if strings.Contains(strIndex, "-") {
			parts := strings.Split(strIndex, "-")
			if len(parts) != 2 {
				return nil, errInvalidIndex
			}

			minIndex, errMin := strconv.Atoi(parts[0])
			maxIndex, errMax := strconv.Atoi(parts[1])
			if errMin != nil || errMax != nil || minIndex < 1 || minIndex > maxIndex {
				return nil, errInvalidIndex
			}

			for i := minIndex; i <= maxIndex; i++ {
				listIndex = append(listIndex, i)
			}
		} else {
			index, err := strconv.Atoi(strIndex)
			if err != nil || index < 1 {
				return nil, errInvalidIndex
			}

			listIndex = append(listIndex, index)
		}
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

	if options.OrderLatest {
		query += ` ORDER BY modified DESC`
	}

	bookmarks := []model.Bookmark{}
	err := db.Select(&bookmarks, query, args...)
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

		if options.WithContents {
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
func (db *SQLiteDatabase) DeleteBookmarks(indices ...string) (oldIndices, newIndices []int, err error) {
	// Convert list of index to int
	listIndex := []int{}
	errInvalidIndex := fmt.Errorf("Index is not valid")

	for _, strIndex := range indices {
		if strings.Contains(strIndex, "-") {
			parts := strings.Split(strIndex, "-")
			if len(parts) != 2 {
				return nil, nil, errInvalidIndex
			}

			minIndex, errMin := strconv.Atoi(parts[0])
			maxIndex, errMax := strconv.Atoi(parts[1])
			if errMin != nil || errMax != nil || minIndex < 1 || minIndex > maxIndex {
				return nil, nil, errInvalidIndex
			}

			for i := minIndex; i <= maxIndex; i++ {
				listIndex = append(listIndex, i)
			}
		} else {
			index, err := strconv.Atoi(strIndex)
			if err != nil || index < 1 {
				return nil, nil, errInvalidIndex
			}

			listIndex = append(listIndex, index)
		}
	}

	// Sort the index
	sort.Ints(listIndex)

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
		return nil, nil, errInvalidIndex
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			oldIndices = nil
			newIndices = nil
			err = panicErr
		}
	}()

	// Delete bookmarks
	whereTagClause := strings.Replace(whereClause, "id", "bookmark_id", 1)
	whereContentClause := strings.Replace(whereClause, "id", "docid", 1)

	tx.MustExec("DELETE FROM bookmark "+whereClause, args...)
	tx.MustExec("DELETE FROM bookmark_tag "+whereTagClause, args...)
	tx.MustExec("DELETE FROM bookmark_content "+whereContentClause, args...)

	// Prepare statement for updating index
	stmtGetMaxID, err := tx.Preparex(`SELECT IFNULL(MAX(id), 0) FROM bookmark`)
	checkError(err)

	stmtUpdateBookmark, err := tx.Preparex(`UPDATE bookmark SET id = ? WHERE id = ?`)
	checkError(err)

	stmtUpdateBookmarkTag, err := tx.Preparex(`UPDATE bookmark_tag SET bookmark_id = ? WHERE bookmark_id = ?`)
	checkError(err)

	stmtUpdateBookmarkContent, err := tx.Preparex(`UPDATE bookmark_content SET docid = ? WHERE docid = ?`)
	checkError(err)

	// Get list of removed indices
	maxIndex := 0
	err = stmtGetMaxID.Get(&maxIndex)
	checkError(err)

	removedIndices := []int{}
	err = tx.Select(&removedIndices,
		`WITH cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt LIMIT ?)
		SELECT x FROM cnt WHERE x NOT IN (SELECT id FROM bookmark)`,
		maxIndex)
	checkError(err)

	// Fill removed indices
	newIndices = []int{}
	oldIndices = []int{}
	for _, removedIndex := range removedIndices {
		oldIndex := 0
		err = stmtGetMaxID.Get(&oldIndex)
		checkError(err)

		if oldIndex <= removedIndex {
			break
		}

		stmtUpdateBookmark.MustExec(removedIndex, oldIndex)
		stmtUpdateBookmarkTag.MustExec(removedIndex, oldIndex)
		stmtUpdateBookmarkContent.MustExec(removedIndex, oldIndex)

		newIndices = append(newIndices, removedIndex)
		oldIndices = append(oldIndices, oldIndex)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return oldIndices, newIndices, err
}

// SearchBookmarks search bookmarks by the keyword or tags.
func (db *SQLiteDatabase) SearchBookmarks(keyword string, tags ...string) ([]model.Bookmark, error) {
	// Create initial variable
	keyword = strings.TrimSpace(keyword)
	whereClause := "WHERE 1"
	args := []interface{}{}

	// Create where clause for keyword
	if keyword != "" {
		whereClause += ` AND id IN (
			SELECT docid id FROM bookmark_content 
			WHERE title MATCH ? OR content MATCH ?)`
		args = append(args, keyword, keyword)
	}

	// Create where clause for tags
	if len(tags) > 0 {
		whereTagClause := ` AND id IN (
			SELECT DISTINCT bookmark_id FROM bookmark_tag 
			WHERE tag_id IN (SELECT id FROM tag WHERE name IN (`

		for _, tag := range tags {
			args = append(args, tag)
			whereTagClause += "?,"
		}

		whereTagClause = whereTagClause[:len(whereTagClause)-1]
		whereTagClause += ")))"

		whereClause += whereTagClause
	}

	// Search bookmarks
	query := `SELECT id, 
		url, title, image_url, excerpt, author, 
		min_read_time, max_read_time, modified
		FROM bookmark ` + whereClause

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
func (db *SQLiteDatabase) UpdateBookmarks(bookmarks []model.Bookmark) (err error) {
	// Prepare transaction
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

	// Prepare statement
	stmtUpdateBookmark, err := db.Preparex(`UPDATE bookmark SET
		url = ?, title = ?, image_url = ?, excerpt = ?, author = ?,
		min_read_time = ?, max_read_time = ? WHERE id = ?`)
	checkError(err)

	stmtUpdateBookmarkContent, err := db.Preparex(`UPDATE bookmark_content SET
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

	for _, book := range bookmarks {
		stmtUpdateBookmark.MustExec(
			book.URL,
			book.Title,
			book.ImageURL,
			book.Excerpt,
			book.Author,
			book.MinReadTime,
			book.MaxReadTime,
			book.ID)

		stmtUpdateBookmarkContent.MustExec(
			book.Title,
			book.Content,
			book.HTML,
			book.ID)

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
		}
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return err
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
	if err != nil {
		return err
	}

	return nil
}

// GetAccounts fetch list of accounts in database
func (db *SQLiteDatabase) GetAccounts(keyword string) ([]model.Account, error) {
	query := `SELECT id, username, password FROM account`
	args := []interface{}{}
	if keyword != "" {
		query += ` WHERE username LIKE ?`
		args = append(args, "%"+keyword+"%")
	}
	query += ` ORDER BY username`

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
	_,err := db.Exec(`DELETE FROM account `+whereClause, args...)
	return err
}
