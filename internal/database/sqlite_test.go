package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sqliteTestDatabaseFactory(t *testing.T, ctx context.Context) (model.DB, error) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	db, err := OpenSQLiteDatabase(ctx, filepath.Join(tmpDir, "shiori.db"))
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(context.TODO()); err != nil {
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

	db, err := sqliteTestDatabaseFactory(t, ctx)
	assert.NoError(t, err)

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	_, err = db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	book = model.BookmarkDTO{
		URL:   "https://github.com/jamiehannaford/what-happens-when-k8s",
		Title: "what-happens-when-k8s",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")
	savedBookmark := result[0]

	results, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
		Keyword: "what-happens-when",
	})

	assert.NoError(t, err, "Get bookmarks should not fail")
	assert.Len(t, results, 1, "results should contain one item")
	assert.Equal(t, savedBookmark.ID, results[0].ID, "bookmark should be the one saved")
}
