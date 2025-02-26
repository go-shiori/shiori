package model

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// DB is interface for accessing and manipulating data in database.
type DB interface {
	// WriterDB is the underlying sqlx.DB
	WriterDB() *sqlx.DB

	// ReaderDB is the underlying sqlx.DB
	ReaderDB() *sqlx.DB

	// Init initializes the database
	Init(ctx context.Context) error

	// Migrate runs migrations for this database
	Migrate(ctx context.Context) error

	// GetDatabaseSchemaVersion gets the version of the database
	GetDatabaseSchemaVersion(ctx context.Context) (string, error)

	// SetDatabaseSchemaVersion sets the version of the database
	SetDatabaseSchemaVersion(ctx context.Context, version string) error

	// SaveBookmarks saves bookmarks data to database.
	SaveBookmarks(ctx context.Context, create bool, bookmarks ...BookmarkDTO) ([]BookmarkDTO, error)

	// GetBookmarks fetch list of bookmarks based on submitted options.
	GetBookmarks(ctx context.Context, opts DBGetBookmarksOptions) ([]BookmarkDTO, error)

	// GetBookmarksCount get count of bookmarks in database.
	GetBookmarksCount(ctx context.Context, opts DBGetBookmarksOptions) (int, error)

	// DeleteBookmarks removes all record with matching ids from database.
	DeleteBookmarks(ctx context.Context, ids ...int) error

	// GetBookmark fetches bookmark based on its ID or URL.
	GetBookmark(ctx context.Context, id int, url string) (BookmarkDTO, bool, error)

	// CreateAccount saves new account in database
	CreateAccount(ctx context.Context, a Account) (*Account, error)

	// UpdateAccount updates account in database
	UpdateAccount(ctx context.Context, a Account) error

	// ListAccounts fetch list of account (without its password) with matching keyword.
	ListAccounts(ctx context.Context, opts DBListAccountsOptions) ([]Account, error)

	// GetAccount fetch account with matching username.
	GetAccount(ctx context.Context, id DBID) (*Account, bool, error)

	// DeleteAccount removes account with matching id
	DeleteAccount(ctx context.Context, id DBID) error

	// CreateTags creates new tags in database.
	CreateTags(ctx context.Context, tags ...Tag) error

	// GetTags fetch list of tags and its frequency from database.
	GetTags(ctx context.Context) ([]TagDTO, error)

	// RenameTag change the name of a tag.
	RenameTag(ctx context.Context, id int, newName string) error
}

// DBOrderMethod is the order method for getting bookmarks
type DBOrderMethod int

const (
	// DefaultOrder is oldest to newest.
	DefaultOrder DBOrderMethod = iota
	// ByLastAdded is from newest addition to the oldest.
	ByLastAdded
	// ByLastModified is from latest modified to the oldest.
	ByLastModified
)

// DBGetBookmarksOptions is options for fetching bookmarks from database.
type DBGetBookmarksOptions struct {
	IDs          []int
	Tags         []string
	ExcludedTags []string
	Keyword      string
	WithContent  bool
	OrderMethod  DBOrderMethod
	Limit        int
	Offset       int
}

// DBListAccountsOptions is options for fetching accounts from database.
type DBListAccountsOptions struct {
	// Filter accounts by a keyword
	Keyword string
	// Filter accounts by exact useranme
	Username string
	// Return owner accounts only
	Owner bool
	// Retrieve password content
	WithPassword bool
}
