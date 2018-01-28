package database

import (
	"database/sql"
	"fmt"
	"github.com/RadhiFadlillah/go-readability"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/jmoiron/sqlx"
	"log"
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
			log.Println("Database error:", panicErr)
			tx.Rollback()

			db = nil
			err = panicErr
		}
	}()

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS account(
		id INTEGER NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		CONSTRAINT account_PK PRIMARY KEY(id),
		CONSTRAINT account_username_UNIQUE UNIQUE(username))`)
	checkError(err)

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS bookmark(
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
	checkError(err)

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS tag(
		id INTEGER NOT NULL,
		name TEXT NOT NULL,
		CONSTRAINT tag_PK PRIMARY KEY(id),
		CONSTRAINT tag_name_UNIQUE UNIQUE(name))`)
	checkError(err)

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS bookmark_tag(
		bookmark_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		CONSTRAINT bookmark_tag_PK PRIMARY KEY(bookmark_id, tag_id),
		CONSTRAINT bookmark_id_FK FOREIGN KEY(bookmark_id) REFERENCES bookmark(id),
		CONSTRAINT tag_id_FK FOREIGN KEY(tag_id) REFERENCES tag(id))`)
	checkError(err)

	_, err = tx.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS bookmark_content USING fts4(title, content)`)
	checkError(err)

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
	res, err := tx.Exec(`INSERT INTO bookmark (
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
	checkError(err)

	// Get last inserted ID
	bookmarkID, err := res.LastInsertId()
	checkError(err)

	// Save bookmark content
	_, err = tx.Exec(`INSERT INTO bookmark_content 
		(docid, title, content) VALUES (?, ?, ?)`,
		bookmarkID, article.Meta.Title, article.Content)
	checkError(err)

	// Save tags
	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
	checkError(err)

	stmtInsertBookmarkTag, err := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag 
		(tag_id, bookmark_id) VALUES (?, ?)`)
	checkError(err)

	bookmarkTags := []model.Tag{}
	for _, tag := range tags {
		tag = strings.ToLower(tag)
		tag = strings.TrimSpace(tag)

		tagID := int64(-1)
		err = stmtGetTag.Get(&tagID, tag)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}

		if tagID == -1 {
			res, err := stmtInsertTag.Exec(tag)
			checkError(err)

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
	// Prepare query
	query := `SELECT id, 
		url, title, image_url, excerpt, author, 
		language, min_read_time, max_read_time, modified
		FROM bookmark`
	args := []interface{}{}

	if len(indices) == 0 {
		query += " WHERE 1"
	} else {
		query += " WHERE 0"
	}

	// Add where clause
	for _, strIndex := range indices {
		if strings.Contains(strIndex, "-") {
			parts := strings.Split(strIndex, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("Index is not valid")
			}

			minIndex, errMin := strconv.Atoi(parts[0])
			maxIndex, errMax := strconv.Atoi(parts[1])
			if errMin != nil || errMax != nil {
				return nil, fmt.Errorf("Index is not valid")
			}

			query += ` OR (id BETWEEN ? AND ?)`
			args = append(args, minIndex, maxIndex)
		} else {
			index, err := strconv.Atoi(strIndex)
			if err != nil {
				return nil, fmt.Errorf("Index is not valid")
			}

			query += ` OR id = ?`
			args = append(args, index)
		}
	}

	// Fetch bookmarks
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
