package database

import (
	"database/sql"

	"github.com/go-shiori/shiori/internal/model"
)

// OrderMethod is the order method for getting bookmarks
type OrderMethod int

const (
	// DefaultOrder is oldest to newest.
	DefaultOrder OrderMethod = iota
	// ByLastAdded is from newest addition to the oldest.
	ByLastAdded
	// ByLastModified is from latest modified to the oldest.
	ByLastModified
)

// GetBookmarksOptions is options for fetching bookmarks from database.
type GetBookmarksOptions struct {
	IDs          []int
	Tags         []string
	ExcludedTags []string
	Keyword      string
	WithContent  bool
	OrderMethod  OrderMethod
	Limit        int
	Offset       int
}

// GetAccountsOptions is options for fetching accounts from database.
type GetAccountsOptions struct {
	Keyword string
	Owner   bool
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

	// SaveAccount saves new account in database
	SaveAccount(model.Account) error

	// GetAccounts fetch list of account (without its password) with matching keyword.
	GetAccounts(opts GetAccountsOptions) ([]model.Account, error)

	// GetAccount fetch account with matching username.
	GetAccount(username string) (model.Account, bool)

	// DeleteAccounts removes all record with matching usernames
	DeleteAccounts(usernames ...string) error

	// GetTags fetch list of tags and its frequency from database.
	GetTags() ([]model.Tag, error)

	// RenameTag change the name of a tag.
	RenameTag(id int, newName string) error

	// CreateNewID creates new id for specified table.
	CreateNewID(table string) (int, error)
}

func checkError(err error) {
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
}
