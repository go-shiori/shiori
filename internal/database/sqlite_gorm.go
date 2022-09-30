package database

import (
	"context"
	"fmt"
	"log"

	"github.com/go-shiori/shiori/internal/model"
	"gorm.io/gorm"

	"gorm.io/driver/sqlite"
)

// GormDatabase is implementation of Database interface
// for connecting to Gorm generated databases.
type GormDatabase struct {
	dbbase

	gorm *gorm.DB
}

// OpenGORMDatabase creates and open connection
func OpenGORMDatabase(ctx context.Context, databasePath string) (gormdb *GormDatabase, err error) {
	db, _ := gorm.Open(sqlite.Open(databasePath))
	gormdb = &GormDatabase{
		gorm: db,
	}

	return gormdb, gormdb.Migrate()
}

// Migrate runs migrations for this database engine
func (db *GormDatabase) Migrate() error {
	if err := db.gorm.AutoMigrate(&model.Account{}); err != nil {
		log.Fatal(err)
		return fmt.Errorf("error migrating account table: %s", err)
	}

	// if err := db.gorm.AutoMigrate(&model.Bookmark{}); err != nil {
	// 	log.Fatal(err)
	// 	return fmt.Errorf("error migrating bookmark table: %s", err)
	// }

	// if err := db.gorm.AutoMigrate(&model.Tag{}); err != nil {
	// 	log.Fatal(err)
	// 	return fmt.Errorf("error migrating tag table: %s", err)
	// }

	return nil
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *GormDatabase) SaveBookmarks(ctx context.Context, bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
	return
}

// GetBookmarks fetch list of bookmarks based on submitted options.
func (db *GormDatabase) GetBookmarks(ctx context.Context, opts GetBookmarksOptions) (bookmarks []model.Bookmark, err error) {
	tx := db.gorm.Find(&bookmarks)
	return bookmarks, tx.Error
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *GormDatabase) GetBookmarksCount(ctx context.Context, opts GetBookmarksOptions) (c int, err error) {
	return
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *GormDatabase) DeleteBookmarks(ctx context.Context, ids ...int) error {
	return nil
}

// GetBookmark fetches bookmark based on its ID or URL.
// Returns the bookmark and boolean whether it's exist or not.
func (db *GormDatabase) GetBookmark(ctx context.Context, id int, url string) (bookmark model.Bookmark, exists bool, err error) {
	return
}

// SaveAccount saves new account to database. Returns error if any happened.
func (db *GormDatabase) SaveAccount(ctx context.Context, account model.Account) error {
	return nil
}

// GetAccounts fetch list of account (without its password) based on submitted options.
func (db *GormDatabase) GetAccounts(ctx context.Context, opts GetAccountsOptions) (accounts []model.Account, err error) {
	return
}

// GetAccount fetch account with matching username.
// Returns the account and boolean whether it's exist or not.
func (db *GormDatabase) GetAccount(ctx context.Context, username string) (account model.Account, exists bool, err error) {
	tx := db.gorm.First(&account, "username = ?", username)
	return account, account.ID != 0, tx.Error
}

// DeleteAccounts removes all record with matching usernames.
func (db *GormDatabase) DeleteAccounts(ctx context.Context, usernames ...string) error {

	return nil
}

// GetTags fetch list of tags and their frequency.
func (db *GormDatabase) GetTags(ctx context.Context) (tags []model.Tag, err error) {
	tx := db.gorm.Find(&tags)
	return tags, tx.Error
}

// RenameTag change the name of a tag.
func (db *GormDatabase) RenameTag(ctx context.Context, id int, newName string) error {
	return nil
}

// CreateNewID creates new ID for specified table
func (db *GormDatabase) CreateNewID(ctx context.Context, table string) (n int, err error) {
	return
}
