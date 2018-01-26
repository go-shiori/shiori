package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func openDatabase() (db *sqlx.DB, err error) {
	// Open database and start transaction
	db = sqlx.MustConnect("sqlite3", "shiori.db")
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
		account_id INTEGER DEFAULT NULL,
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
		CONSTRAINT bookmark_url_UNIQUE UNIQUE(url),
		CONSTRAINT bookmark_account_id_FK FOREIGN KEY(account_id) REFERENCES account(id))`)
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

	return db, err
}
