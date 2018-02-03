package database

import (
	"database/sql"
	"github.com/RadhiFadlillah/shiori/model"
)

type Database interface {
	SaveBookmark(bookmark model.Bookmark) (int64, error)
	GetBookmarks(withContent bool, indices ...string) ([]model.Bookmark, error)
	DeleteBookmarks(indices ...string) ([]int, []int, error)
	SearchBookmarks(keyword string, tags ...string) ([]model.Bookmark, error)
	UpdateBookmarks(bookmarks []model.Bookmark) error
}

func checkError(err error) {
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
}
