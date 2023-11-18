package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var sqliteDatabaseTestPath string

func init() {
	sqliteDatabaseTestPath = filepath.Join(os.TempDir(), "shiori.db")
}

func sqliteTestDatabaseFactory(ctx context.Context) (DB, error) {
	os.Remove(sqliteDatabaseTestPath)

	db, err := OpenSQLiteDatabase(ctx, sqliteDatabaseTestPath)
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(); err != nil && !errors.Is(migrate.ErrNoChange, err) {
		return nil, err
	}

	return db, nil
}

func TestSqliteDatabase(t *testing.T) {
	testDatabase(t, sqliteTestDatabaseFactory)
	testSqliteGetBookmarksWithDash(t)
}

// testSqliteGetBookmarksWithDash ad-hoc test for SQLite that checks that a match search against
// the FTS5 engine does not fail by using dashes, making sqlite think that we are trying to avoid
// matching a column name. This works in a fun way and it seems that it depends on the tokens
// already scanned by the database, since trying to match for `go-shiori` with no bookmarks or only
// the shiori bookmark does not fail, but it fails if we add any other bookmark to the database, hence
// this test.
func testSqliteGetBookmarksWithDash(t *testing.T) {
	ctx := context.TODO()

	db, err := sqliteTestDatabaseFactory(ctx)
	assert.NoError(t, err)

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	_, err = db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	book = model.Bookmark{
		URL:   "https://github.com/jamiehannaford/what-happens-when-k8s",
		Title: "what-happens-when-k8s",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")
	savedBookmark := result[0]

	results, err := db.GetBookmarks(ctx, GetBookmarksOptions{
		Keyword: "what-happens-when",
	})

	assert.NoError(t, err, "Get bookmarks should not fail")
	assert.Len(t, results, 1, "results should contain one item")
	assert.Equal(t, savedBookmark.ID, results[0].ID, "bookmark should be the one saved")

}
func TestSQLiteDatabase_SaveAccount(t *testing.T) {

	ctx := context.TODO()

	// Initialize not correct database
	factory := func(ctx context.Context) (DB, error) {
		return OpenSQLiteDatabase(ctx, filepath.Join(os.TempDir(), "shiori_test.db"))
	}
	db, err := factory(ctx)
	assert.Nil(t, err)

	// Test falid database
	acc := model.Account{}
	err = db.SaveAccount(ctx, acc)
	assert.Contains(t, err.Error(), "SQL logic error: no such table: account (1)")

}

func TestSaveAccountSettings(t *testing.T) {
	ctx := context.TODO()

	db, err := sqliteTestDatabaseFactory(ctx)
	assert.NoError(t, err)

	// Mock data
	account := model.Account{
		Username: "testuser",
		Config:   model.UserConfig{},
	}

	// Successful case
	err = db.SaveAccountSettings(ctx, account)
	assert.NoError(t, err)

	// Initialize not correct database
	ctx = context.TODO()
	factory := func(ctx context.Context) (DB, error) {
		return OpenSQLiteDatabase(ctx, filepath.Join(os.TempDir(), "shiori_test.db"))
	}
	db, err = factory(ctx)
	assert.Nil(t, err)
	account = model.Account{
		Username: "testuser",
		Config:   model.UserConfig{},
	}
	err = db.SaveAccountSettings(ctx, account)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SQL logic error: no such table: account (1)")
}

func TestGetAccounts(t *testing.T) {
	ctx := context.TODO()

	db, err := sqliteTestDatabaseFactory(ctx)
	assert.NoError(t, err)

	// Insert test accounts
	testAccounts := []model.Account{
		{Username: "foo", Password: "bar", Owner: false},
		{Username: "hello", Password: "world", Owner: false},
		{Username: "foo_bar", Password: "foobar", Owner: true},
	}
	for _, acc := range testAccounts {
		err := db.SaveAccount(ctx, acc)
		assert.Nil(t, err)
	}

	// Successful case
	// without opt
	accounts, err := db.GetAccounts(ctx, GetAccountsOptions{})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(accounts))
	// with owner
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Owner: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(accounts))
	// with opt
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Keyword: "foo"})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(accounts))
	// with opt and owner
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Keyword: "hello", Owner: false})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(accounts))
	// with not result
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Keyword: "shiori"})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(accounts))

	// Initialize not correct database
	ctx = context.TODO()
	factory := func(ctx context.Context) (DB, error) {
		return OpenSQLiteDatabase(ctx, filepath.Join(os.TempDir(), "shiori_test.db"))
	}
	db, err = factory(ctx)
	assert.Nil(t, err)
	// with invalid query
	opts := GetAccountsOptions{Keyword: "foo", Owner: true}
	_, err = db.GetAccounts(ctx, opts)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SQL logic error: no such table: account (1)")
}
