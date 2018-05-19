package database

import (
	"database/sql"

	"github.com/RadhiFadlillah/shiori/model"
)

// Database is interface for manipulating data in database.
type Database interface {
	// SaveBookmark saves new bookmark to database.
	CreateBookmark(bookmark model.Bookmark) (int, error)

	// GetBookmarks fetch list of bookmarks based on submitted ids.
	GetBookmarks(withContent bool, ids ...int) ([]model.Bookmark, error)

	// GetTags fetch list of tags and their frequency
	GetTags() ([]model.Tag, error)

	//GetNewID get new id for specified table
	GetNewID(table string) (int, error)

	// DeleteBookmarks removes all record with matching ids from database.
	DeleteBookmarks(ids ...int) error

	// SearchBookmarks search bookmarks by the keyword or tags.
	SearchBookmarks(orderLatest bool, keyword string, tags ...string) ([]model.Bookmark, error)

	// UpdateBookmarks updates the saved bookmark in database.
	UpdateBookmarks(bookmarks ...model.Bookmark) ([]model.Bookmark, error)

	// CreateAccount creates new account in database
	CreateAccount(username, password string) error

	// GetAccount fetch account with matching username
	GetAccount(username string) (model.Account, error)

	// GetAccounts fetch list of accounts with matching keyword
	GetAccounts(keyword string) ([]model.Account, error)

	// DeleteAccounts removes all record with matching usernames
	DeleteAccounts(usernames ...string) error
}

func checkError(err error) {
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
}
