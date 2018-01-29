package database

import (
	"database/sql"
	"fmt"
	"github.com/RadhiFadlillah/go-readability"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/jmoiron/sqlx"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SQLiteDatabase struct {
	sqlx.DB
}

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
		language TEXT NOT NULL DEFAULT "",
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

	tx.MustExec(`CREATE VIRTUAL TABLE IF NOT EXISTS bookmark_content USING fts4(title, content)`)

	err = tx.Commit()
	checkError(err)

	return &SQLiteDatabase{*db}, err
}

func (db *SQLiteDatabase) SaveBookmark(article readability.Article, tags ...string) (bookmark model.Bookmark, err error) {
	// Prepare transaction
	tx, err := db.Beginx()
	if err != nil {
		return model.Bookmark{}, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			bookmark = model.Bookmark{}
			err = panicErr
		}
	}()

	// Save article to database
	res := tx.MustExec(`INSERT INTO bookmark (
		url, title, image_url, excerpt, author, 
		language, min_read_time, max_read_time) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		article.URL,
		article.Meta.Title,
		article.Meta.Image,
		article.Meta.Excerpt,
		article.Meta.Author,
		article.Meta.Language,
		article.Meta.MinReadTime,
		article.Meta.MaxReadTime)

	// Get last inserted ID
	bookmarkID, err := res.LastInsertId()
	checkError(err)

	// Save bookmark content
	tx.MustExec(`INSERT INTO bookmark_content 
		(docid, title, content) VALUES (?, ?, ?)`,
		bookmarkID, article.Meta.Title, article.Content)

	// Save tags
	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
	checkError(err)

	stmtInsertBookmarkTag, err := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag (tag_id, bookmark_id) VALUES (?, ?)`)
	checkError(err)

	bookmarkTags := []model.Tag{}
	for _, tag := range tags {
		tag = strings.ToLower(tag)
		tag = strings.TrimSpace(tag)

		tagID := int64(-1)
		err = stmtGetTag.Get(&tagID, tag)
		checkError(err)

		if tagID == -1 {
			res := stmtInsertTag.MustExec(tag)
			tagID, err = res.LastInsertId()
			checkError(err)
		}

		stmtInsertBookmarkTag.Exec(tagID, bookmarkID)
		bookmarkTags = append(bookmarkTags, model.Tag{
			ID:   tagID,
			Name: tag,
		})
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	// Return result
	bookmark = model.Bookmark{
		ID:          bookmarkID,
		URL:         article.URL,
		Title:       article.Meta.Title,
		ImageURL:    article.Meta.Image,
		Excerpt:     article.Meta.Excerpt,
		Author:      article.Meta.Author,
		Language:    article.Meta.Language,
		MinReadTime: article.Meta.MinReadTime,
		MaxReadTime: article.Meta.MaxReadTime,
		Modified:    time.Now().Format("2006-01-02 15:04:05"),
		Tags:        bookmarkTags,
	}

	return bookmark, err
}

func (db *SQLiteDatabase) GetBookmarks(indices ...string) ([]model.Bookmark, error) {
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
		language, min_read_time, max_read_time, modified
		FROM bookmark` + whereClause

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
	res := tx.MustExec("DELETE FROM bookmark "+whereClause, args...)

	nAffected, err := res.RowsAffected()
	checkError(err)

	whereTagClause := strings.Replace(whereClause, "id", "bookmark_id", 1)
	tx.MustExec("DELETE FROM bookmark_tag "+whereTagClause, args...)

	whereContentClause := strings.Replace(whereClause, "id", "docid", 1)
	tx.MustExec("DELETE FROM bookmark_content "+whereContentClause, args...)

	// Get largest index
	oldIndices = []int{}
	err = tx.Select(&oldIndices, "SELECT id FROM bookmark ORDER BY id DESC LIMIT ?", nAffected)
	checkError(err)
	sort.Ints(oldIndices)

	// Update index
	newIndices = listIndex[:len(oldIndices)]
	stmtUpdateBookmark, err := tx.Preparex(`UPDATE bookmark SET id = ? WHERE id = ?`)
	checkError(err)

	stmtUpdateBookmarkTag, err := tx.Preparex(`UPDATE bookmark_tag SET bookmark_id = ? WHERE bookmark_id = ?`)
	checkError(err)

	stmtUpdateBookmarkContent, err := tx.Preparex(`UPDATE bookmark_content SET docid = ? WHERE docid = ?`)
	checkError(err)

	for i, oldIndex := range oldIndices {
		newIndex := newIndices[i]

		stmtUpdateBookmark.MustExec(newIndex, oldIndex)
		stmtUpdateBookmarkTag.MustExec(newIndex, oldIndex)
		stmtUpdateBookmarkContent.MustExec(newIndex, oldIndex)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return oldIndices, newIndices, err
}

func (db *SQLiteDatabase) SearchBookmarks(keyword string, tags ...string) ([]model.Bookmark, error) {
	// Create initial variable
	keyword = strings.TrimSpace(keyword)
	whereClause := "WHERE 1"
	args := []interface{}{}

	// Create where clause for keyword
	if keyword != "" {
		whereClause += ` AND id IN (
			SELECT docid id FROM bookmark_content 
			WHERE bookmark_content MATCH ?)`
		args = append(args, keyword)
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
		language, min_read_time, max_read_time, modified
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
