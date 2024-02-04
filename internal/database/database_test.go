package database

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type databaseTestCase func(t *testing.T, dbFactory testDatabaseFactory)
type testDatabaseFactory func(ctx context.Context) (DB, error)

func testDatabase(t *testing.T, dbFactory testDatabaseFactory) {
	tests := map[string]databaseTestCase{
		// Bookmarks
		"testBookmarkAutoIncrement":         testBookmarkAutoIncrement,
		"testCreateBookmark":                testCreateBookmark,
		"testCreateBookmarkWithContent":     testCreateBookmarkWithContent,
		"testCreateBookmarkTwice":           testCreateBookmarkTwice,
		"testCreateBookmarkWithTag":         testCreateBookmarkWithTag,
		"testCreateTwoDifferentBookmarks":   testCreateTwoDifferentBookmarks,
		"testUpdateBookmark":                testUpdateBookmark,
		"testUpdateBookmarkWithContent":     testUpdateBookmarkWithContent,
		"testGetBookmark":                   testGetBookmark,
		"testGetBookmarkNotExistent":        testGetBookmarkNotExistent,
		"testGetBookmarks":                  testGetBookmarks,
		"testGetBookmarksWithSQLCharacters": testGetBookmarksWithSQLCharacters,
		"testGetBookmarksCount":             testGetBookmarksCount,
		// Tags
		"testCreateTag":  testCreateTag,
		"testCreateTags": testCreateTags,
		// Accoubnts
		"testCreateAccount": testCreateAccount,
		"testDeleteAccount": testDeleteAccount,
	}

	for testName, testCase := range tests {
		t.Run(testName, func(tInner *testing.T) {
			testCase(tInner, dbFactory)
		})
	}
}

func testBookmarkAutoIncrement(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 1, result[0].ID, "Saved bookmark must have ID %d", 1)

	book = model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "obelisk",
	}

	result, err = db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 2, result[0].ID, "Saved bookmark must have ID %d", 2)
}

func testCreateBookmark(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)

	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 1, result[0].ID, "Saved bookmark must have an ID set")
}

func testCreateBookmarkWithContent(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
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

func testCreateBookmarkWithTag(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
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

func testCreateBookmarkTwice(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
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

func testCreateTwoDifferentBookmarks(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	_, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save first bookmark must not fail")

	book = model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/go-readability",
		Title: "go-readability",
	}
	_, err = db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save second bookmark must not fail")
}

func testUpdateBookmark(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
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

func testUpdateBookmarkWithContent(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
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

func testGetBookmark(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
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

func testGetBookmarkNotExistent(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	savedBookmark, exists, err := db.GetBookmark(ctx, 1, "")
	assert.NoError(t, err, "Get bookmark should not fail")
	assert.False(t, exists, "Bookmark should not exist")
	assert.Equal(t, model.BookmarkDTO{}, savedBookmark)
}

func testGetBookmarks(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
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

func testGetBookmarksWithSQLCharacters(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}
	_, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	characters := []string{";", "%", "_", "\\", "\"", ":"}

	for _, char := range characters {
		t.Run("GetBookmarks/"+char, func(t *testing.T) {
			_, err := db.GetBookmarks(ctx, GetBookmarksOptions{
				Keyword: char,
			})
			assert.NoError(t, err, "Get bookmarks should not fail")
		})

		t.Run("GetBookmarksCount/"+char, func(t *testing.T) {
			_, err := db.GetBookmarksCount(ctx, GetBookmarksOptions{
				Keyword: char,
			})
			assert.NoError(t, err, "Get bookmarks count should not fail")
		})
	}
}

func testGetBookmarksCount(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	expectedCount := 1
	book := model.BookmarkDTO{
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

// ----------------- TAGS -----------------

func testCreateTag(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	tag := model.Tag{Name: "shiori"}
	err := db.CreateTags(ctx, tag)
	assert.NoError(t, err, "Save tag must not fail")
}

func testCreateTags(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	err := db.CreateTags(ctx, model.Tag{Name: "shiori"}, model.Tag{Name: "shiori2"})
	assert.NoError(t, err, "Save tag must not fail")
}

// ----------------- ACCOUNTS -----------------
func testCreateAccount(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()
	db, errDB := dbFactory(ctx)
	require.NoError(t, errDB)

	acc := model.Account{
		Username: "testuser",
		Password: "testpass",
		Owner:    true,
	}
	insertedAccount, err := db.SaveAccount(ctx, acc)
	assert.NoError(t, err, "Save account must not fail")
	assert.Equal(t, acc.Username, insertedAccount.Username, "Saved account must have an username set")
	assert.Equal(t, acc.Password, insertedAccount.Password, "Saved account must have a password set")
	assert.Equal(t, acc.Owner, insertedAccount.Owner, "Saved account must have an owner set")
	assert.NotEmpty(t, insertedAccount.ID, "Saved account must have an ID set")
}

func testDeleteAccount(t *testing.T, dbFactory testDatabaseFactory) {
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
		db, errDB := dbFactory(ctx)
		require.NoError(t, errDB)

		acc := model.Account{
			Username: "testuser",
			Password: "testpass",
			Owner:    true,
		}
		storedAccount, err := db.SaveAccount(ctx, acc)
		assert.NoError(t, err, "Save account must not fail")

		err = db.DeleteAccount(ctx, storedAccount.Username)
		assert.NoError(t, err, "Delete account must not fail")

		_, exists, err := db.GetAccount(ctx, storedAccount.Username)
		assert.NoError(t, err, "Get account must not fail")
		assert.False(t, exists, "Account must not exist")
	})

	t.Run("not existent", func(t *testing.T) {
		db, errDB := dbFactory(ctx)
		require.NoError(t, errDB)

		err := db.DeleteAccount(ctx, "notexistent")
		assert.ErrorIs(t, ErrNotFound, err, "Delete account must fail")
	})
}
