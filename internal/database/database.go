package database

import (
	"database/sql"

	"github.com/go-shiori/shiori/internal/model"
)

// GetBookmarksOptions is options for fetching bookmarks from database.
type GetBookmarksOptions struct {
	IDs         []int
	Tags        []string
	Keyword     string
	WithContent bool
	OrderLatest bool
	Limit       int
	Offset      int
}

// DB is interface for accessing and manipulating data in database.
type DB interface {
	// SaveBookmarks saves bookmarks data to database.
	SaveBookmarks(bookmarks ...model.Bookmark) ([]model.Bookmark, error)

	// GetBookmarks fetch list of bookmarks based on submitted options.
	GetBookmarks(opts GetBookmarksOptions) ([]model.Bookmark, error)

	// GetBookmarksCount get count of bookmarks in database.
	GetBookmarksCount(opts GetBookmarksOptions) (int, error)

	// DeleteBookmarks removes all record with matching ids from database.
	DeleteBookmarks(ids ...int) error

	// GetBookmark fetchs bookmark based on its ID or URL.
	GetBookmark(id int, url string) (model.Bookmark, bool)

	// GetAccounts fetch list of accounts with matching keyword.
	GetAccounts(keyword string) ([]model.Account, error)

	// GetAccount fetch account with matching username.
	GetAccount(username string) (model.Account, bool)

	// GetTags fetch list of tags and its frequency from database.
	GetTags() ([]model.Tag, error)

	// CreateNewID creates new id for specified table.
	CreateNewID(table string) (int, error)
}

func checkError(err error) {
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
}
