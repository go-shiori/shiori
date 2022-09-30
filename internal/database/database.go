package database

import (
	"context"
	"embed"
	"log"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

//go:embed migrations/*
var migrations embed.FS

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
	// Migrate runs migrations for this database
	Migrate() error

	// SaveBookmarks saves bookmarks data to database.
	SaveBookmarks(ctx context.Context, bookmarks ...model.Bookmark) ([]model.Bookmark, error)

	// GetBookmarks fetch list of bookmarks based on submitted options.
	GetBookmarks(ctx context.Context, opts GetBookmarksOptions) ([]model.Bookmark, error)

	// GetBookmarksCount get count of bookmarks in database.
	GetBookmarksCount(ctx context.Context, opts GetBookmarksOptions) (int, error)

	// DeleteBookmarks removes all record with matching ids from database.
	DeleteBookmarks(ctx context.Context, ids ...int) error

	// GetBookmark fetchs bookmark based on its ID or URL.
	GetBookmark(ctx context.Context, id int, url string) (model.Bookmark, bool, error)

	// SaveAccount saves new account in database
	SaveAccount(ctx context.Context, a model.Account) error

	// GetAccounts fetch list of account (without its password) with matching keyword.
	GetAccounts(ctx context.Context, opts GetAccountsOptions) ([]model.Account, error)

	// GetAccount fetch account with matching username.
	GetAccount(ctx context.Context, username string) (model.Account, bool, error)

	// DeleteAccounts removes all record with matching usernames
	DeleteAccounts(ctx context.Context, usernames ...string) error

	// GetTags fetch list of tags and its frequency from database.
	GetTags(ctx context.Context) ([]model.Tag, error)

	// RenameTag change the name of a tag.
	RenameTag(ctx context.Context, id int, newName string) error

	// CreateNewID creates new id for specified table.
	CreateNewID(ctx context.Context, table string) (int, error)
}

type dbbase struct {
	sqlx.DB
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
