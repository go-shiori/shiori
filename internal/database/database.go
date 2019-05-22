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
}

// DB is interface for accessing and manipulating data in database.
type DB interface {
	// InsertBookmark inserts new bookmark to database.
	InsertBookmark(bookmark model.Bookmark) (int, error)

	// GetBookmarks fetch list of bookmarks based on submitted options.
	GetBookmarks(opts GetBookmarksOptions) ([]model.Bookmark, error)

	// DeleteBookmarks removes all record with matching ids from database.
	DeleteBookmarks(ids ...int) error

	// CreateNewID creates new id for specified table.
	CreateNewID(table string) (int, error)
}

func checkError(err error) {
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
}
