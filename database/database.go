package database

import (
	"database/sql"
	"github.com/RadhiFadlillah/shiori/model"
)

// Database is interface for manipulating data in database.
type Database interface {
	// SaveBookmark saves new bookmark to database.
	SaveBookmark(bookmark model.Bookmark) (int64, error)

	// GetBookmarks fetch list of bookmarks based on submitted indices.
	GetBookmarks(withContent bool, indices ...string) ([]model.Bookmark, error)

	// DeleteBookmarks removes all record with matching indices from database.
	DeleteBookmarks(indices ...string) ([]int, []int, error)

	// SearchBookmarks search bookmarks by the keyword or tags.
	SearchBookmarks(keyword string, tags ...string) ([]model.Bookmark, error)

	// UpdateBookmarks updates the saved bookmark in database.
	UpdateBookmarks(bookmarks []model.Bookmark) error
}

func checkError(err error) {
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
}
