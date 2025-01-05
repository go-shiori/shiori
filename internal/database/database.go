package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// ErrNotFound is error returned when record is not found in database.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is error returned when record already exists in database.
var ErrAlreadyExists = errors.New("already exists")

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

// ListAccountsOptions is options for fetching accounts from database.
type ListAccountsOptions struct {
	// Filter accounts by a keyword
	Keyword string
	// Filter accounts by exact useranme
	Username string
	// Return owner accounts only
	Owner bool
	// Retrieve password content
	WithPassword bool
}

// Connect connects to database based on submitted database URL.
func Connect(ctx context.Context, dbURL string) (DB, error) {
	dbU, err := url.Parse(dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse database URL")
	}

	switch dbU.Scheme {
	case "mysql":
		urlNoSchema := strings.Split(dbURL, "://")[1]
		return OpenMySQLDatabase(ctx, urlNoSchema)
	case "postgres":
		return OpenPGDatabase(ctx, dbURL)
	case "sqlite":
		return OpenSQLiteDatabase(ctx, dbU.Path[1:])
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", dbU.Scheme)
}

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
	SaveBookmarks(ctx context.Context, create bool, bookmarks ...model.BookmarkDTO) ([]model.BookmarkDTO, error)

	// GetBookmarks fetch list of bookmarks based on submitted options.
	GetBookmarks(ctx context.Context, opts GetBookmarksOptions) ([]model.BookmarkDTO, error)

	// GetBookmarksCount get count of bookmarks in database.
	GetBookmarksCount(ctx context.Context, opts GetBookmarksOptions) (int, error)

	// DeleteBookmarks removes all record with matching ids from database.
	DeleteBookmarks(ctx context.Context, ids ...int) error

	// GetBookmark fetches bookmark based on its ID or URL.
	GetBookmark(ctx context.Context, id int, url string) (model.BookmarkDTO, bool, error)

	// CreateAccount saves new account in database
	CreateAccount(ctx context.Context, a model.Account) (*model.Account, error)

	// UpdateAccount updates account in database
	UpdateAccount(ctx context.Context, a model.Account) error

	// ListAccounts fetch list of account (without its password) with matching keyword.
	ListAccounts(ctx context.Context, opts ListAccountsOptions) ([]model.Account, error)

	// GetAccount fetch account with matching username.
	GetAccount(ctx context.Context, id model.DBID) (*model.Account, bool, error)

	// DeleteAccount removes account with matching id
	DeleteAccount(ctx context.Context, id model.DBID) error

	// CreateTags creates new tags in database.
	CreateTags(ctx context.Context, tags ...model.Tag) error

	// GetTags fetch list of tags and its frequency from database.
	GetTags(ctx context.Context) ([]model.Tag, error)

	// RenameTag change the name of a tag.
	RenameTag(ctx context.Context, id int, newName string) error
}

type dbbase struct {
	*sqlx.DB
}

func (db *dbbase) withTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		if err := tx.Commit(); err != nil {
			log.Printf("error during commit: %s", err)
		}
	}()

	err = fn(tx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("error during rollback: %s", err)
		}
		return errors.WithStack(err)
	}

	return err
}
