package database

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
)

type databaseTestCase func(t *testing.T, db DB)
type testDatabaseFactory func(ctx context.Context) (DB, error)

func testDatabase(t *testing.T, dbFactory testDatabaseFactory) {
	tests := map[string]databaseTestCase{
		"testBookmarkAutoIncrement":       testBookmarkAutoIncrement,
		"testCreateBookmark":              testCreateBookmark,
		"testCreateBookmarkWithContent":   testCreateBookmarkWithContent,
		"testCreateBookmarkTwice":         testCreateBookmarkTwice,
		"testCreateBookmarkWithTag":       testCreateBookmarkWithTag,
		"testCreateTwoDifferentBookmarks": testCreateTwoDifferentBookmarks,
		"testUpdateBookmark":              testUpdateBookmark,
		"testUpdateBookmarkWithContent":   testUpdateBookmarkWithContent,
		"testGetBookmark":                 testGetBookmark,
		"testGetBookmarkNotExistant":      testGetBookmarkNotExistant,
		"testGetBookmarks":                testGetBookmarks,
		"testGetBookmarksCount":           testGetBookmarksCount,
	}

	for testName, testCase := range tests {
		t.Run(testName, func(tInner *testing.T) {
			ctx := context.TODO()
			db, err := dbFactory(ctx)
			assert.NoError(tInner, err, "Error recreating database")
			testCase(tInner, db)
		})
	}
}

func testBookmarkAutoIncrement(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 1, result[0].ID, "Saved bookmark must have ID %d", 1)

	book = model.Bookmark{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "obelisk",
	}

	result, err = db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 2, result[0].ID, "Saved bookmark must have ID %d", 2)
}

func testCreateBookmark(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)

	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 1, result[0].ID, "Saved bookmark must have an ID set")
}

func testCreateBookmarkWithContent(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:     "https://github.com/go-shiori/obelisk",
		Title:   "shiori",
		Content: "Some content",
		HTML:    "Some HTML content",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	books, err := db.GetBookmarks(ctx, GetBookmarksOptions{
		IDs:         []int{result[0].ID},
		WithContent: true,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	assert.Len(t, books, 1)

	assert.Equal(t, 1, books[0].ID, "Saved bookmark must have an ID set")
	assert.Equal(t, book.Content, books[0].Content, "Saved bookmark must have content")
	assert.Equal(t, book.HTML, books[0].HTML, "Saved bookmark must have HTML")
}

func testCreateBookmarkWithTag(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "shiori",
		Tags: []model.Tag{
			{
				Name: "test-tag",
			},
		},
	}

	result, err := db.SaveBookmarks(ctx, true, book)

	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, book.URL, result[0].URL)
	assert.Equal(t, book.Tags[0].Name, result[0].Tags[0].Name)
}

func testCreateBookmarkTwice(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	savedBookmark := result[0]
	savedBookmark.Title = "modified"

	_, err = db.SaveBookmarks(ctx, true, savedBookmark)
	assert.Error(t, err, "Save bookmarks must fail")
}

func testCreateTwoDifferentBookmarks(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	_, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save first bookmark must not fail")

	book = model.Bookmark{
		URL:   "https://github.com/go-shiori/go-readability",
		Title: "go-readability",
	}
	_, err = db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save second bookmark must not fail")
}

func testUpdateBookmark(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	savedBookmark := result[0]
	savedBookmark.Title = "modified"

	result, err = db.SaveBookmarks(ctx, false, savedBookmark)
	assert.NoError(t, err, "Save bookmarks must not fail")

	assert.Equal(t, "modified", result[0].Title)
	assert.Equal(t, savedBookmark.ID, result[0].ID)
}

func testUpdateBookmarkWithContent(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:     "https://github.com/go-shiori/obelisk",
		Title:   "shiori",
		Content: "Some content",
		HTML:    "Some HTML content",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	updatedBook := result[0]
	updatedBook.Content = "Some updated content"
	updatedBook.HTML = "Some updated HTML content"

	_, err = db.SaveBookmarks(ctx, false, updatedBook)
	assert.NoError(t, err, "Save bookmarks must not fail")

	books, err := db.GetBookmarks(ctx, GetBookmarksOptions{
		IDs:         []int{result[0].ID},
		WithContent: true,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	assert.Len(t, books, 1)

	assert.Equal(t, 1, books[0].ID, "Saved bookmark must have an ID set")
	assert.Equal(t, updatedBook.Content, books[0].Content, "Saved bookmark must have updated content")
	assert.Equal(t, updatedBook.HTML, books[0].HTML, "Saved bookmark must have updated HTML")
}

func testGetBookmark(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	savedBookmark, exists, err := db.GetBookmark(ctx, result[0].ID, "")
	assert.True(t, exists, "Bookmark should exist")
	assert.NoError(t, err, "Get bookmark should not fail")
	assert.Equal(t, result[0].ID, savedBookmark.ID, "Retrieved bookmark should be the same")
	assert.Equal(t, book.URL, savedBookmark.URL, "Retrieved bookmark should be the same")
}

func testGetBookmarkNotExistant(t *testing.T, db DB) {
	ctx := context.TODO()

	savedBookmark, exists, err := db.GetBookmark(ctx, 1, "")
	assert.NoError(t, err, "Get bookmark should not fail")
	assert.False(t, exists, "Bookmark should not exist")
	assert.Equal(t, model.Bookmark{}, savedBookmark)
}

func testGetBookmarks(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	bookmarks, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	savedBookmark := bookmarks[0]

	results, err := db.GetBookmarks(ctx, GetBookmarksOptions{
		Keyword: "go-shiori",
	})

	assert.NoError(t, err, "Get bookmarks should not fail")
	assert.Len(t, results, 1, "results should contain one item")
	assert.Equal(t, savedBookmark.ID, results[0].ID, "bookmark should be the one saved")
}

func testGetBookmarksCount(t *testing.T, db DB) {
	ctx := context.TODO()

	expectedCount := 1
	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	_, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	count, err := db.GetBookmarksCount(ctx, GetBookmarksOptions{
		Keyword: "go-shiori",
	})
	assert.NoError(t, err, "Get bookmarks count should not fail")
	assert.Equal(t, count, expectedCount, "count should be %d", expectedCount)
}
