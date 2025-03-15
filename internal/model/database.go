package model

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type DBID int

// DB is interface for accessing and manipulating data in database.
type DB interface {
	// WriterDB is the underlying sqlx.DB
	WriterDB() *sqlx.DB

	// ReaderDB is the underlying sqlx.DB
	ReaderDB() *sqlx.DB

	// Flavor is the flavor of the database
	// Flavor() sqlbuilder.Flavor

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

	// SaveBookmark saves a single bookmark to database without handling tags.
	// It only updates the bookmark data in the database.
	SaveBookmark(ctx context.Context, bookmark Bookmark) error

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
	CreateTags(ctx context.Context, tags ...Tag) ([]Tag, error)

	// CreateTag creates a new tag in database.
	CreateTag(ctx context.Context, tag Tag) (Tag, error)

	// GetTags fetch list of tags and its frequency from database.
	GetTags(ctx context.Context, opts DBListTagsOptions) ([]TagDTO, error)

	// RenameTag change the name of a tag.
	RenameTag(ctx context.Context, id int, newName string) error

	// GetTag fetch a tag by its ID.
	GetTag(ctx context.Context, id int) (TagDTO, bool, error)

	// UpdateTag updates a tag in the database.
	UpdateTag(ctx context.Context, tag Tag) error

	// DeleteTag removes a tag from the database.
	DeleteTag(ctx context.Context, id int) error

	// BulkUpdateBookmarkTags updates tags for multiple bookmarks.
	// It ensures that all bookmarks and tags exist before proceeding.
	BulkUpdateBookmarkTags(ctx context.Context, bookmarkIDs []int, tagIDs []int) error

	// AddTagToBookmark adds a tag to a bookmark
	AddTagToBookmark(ctx context.Context, bookmarkID int, tagID int) error

	// RemoveTagFromBookmark removes a tag from a bookmark
	RemoveTagFromBookmark(ctx context.Context, bookmarkID int, tagID int) error

	// TagExists checks if a tag with the given ID exists in the database
	TagExists(ctx context.Context, tagID int) (bool, error)

	// BookmarkExists checks if a bookmark with the given ID exists in the database
	BookmarkExists(ctx context.Context, bookmarkID int) (bool, error)
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

type DBTagOrderBy string

const (
	DBTagOrderByTagName DBTagOrderBy = "name"
)

// DBListTagsOptions is options for fetching tags from database.
type DBListTagsOptions struct {
	BookmarkID        int
	WithBookmarkCount bool
	OrderBy           DBTagOrderBy
	Search            string
}
