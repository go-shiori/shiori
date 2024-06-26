package database

import (
	"context"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type databaseTestCase func(t *testing.T, db DB)
type testDatabaseFactory func(t *testing.T, ctx context.Context) (DB, error)

func testDatabase(t *testing.T, dbFactory testDatabaseFactory) {
	tests := map[string]databaseTestCase{
		// Bookmarks
		"testBookmarkAutoIncrement":             testBookmarkAutoIncrement,
		"testCreateBookmark":                    testCreateBookmark,
		"testCreateBookmarkWithContent":         testCreateBookmarkWithContent,
		"testCreateBookmarkTwice":               testCreateBookmarkTwice,
		"testCreateBookmarkWithTag":             testCreateBookmarkWithTag,
		"testCreateTwoDifferentBookmarks":       testCreateTwoDifferentBookmarks,
		"testUpdateBookmark":                    testUpdateBookmark,
		"testUpdateBookmarkUpdatesModifiedTime": testUpdateBookmarkUpdatesModifiedTime,
		"testGetBoomarksWithTimeFilters":        testGetBoomarksWithTimeFilters,
		"testUpdateBookmarkWithContent":         testUpdateBookmarkWithContent,
		"testGetBookmark":                       testGetBookmark,
		"testGetBookmarkNotExistent":            testGetBookmarkNotExistent,
		"testGetBookmarks":                      testGetBookmarks,
		"testGetBookmarksWithSQLCharacters":     testGetBookmarksWithSQLCharacters,
		"testGetBookmarksCount":                 testGetBookmarksCount,
		// Tags
		"testCreateTag":  testCreateTag,
		"testCreateTags": testCreateTags,
		// Accounts
		"testSaveAccount":        testSaveAccount,
		"testSaveAccountSetting": testSaveAccountSettings,
		"testGetAccount":         testGetAccount,
		"testGetAccounts":        testGetAccounts,
	}

	for testName, testCase := range tests {
		t.Run(testName, func(tInner *testing.T) {
			ctx := context.TODO()
			db, err := dbFactory(t, ctx)
			assert.NoError(tInner, err, "Error recreating database")
			testCase(tInner, db)
		})
	}
}

func testBookmarkAutoIncrement(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testCreateBookmark(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)

	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 1, result[0].ID, "Saved bookmark must have an ID set")
}

func testCreateBookmarkWithContent(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testCreateBookmarkWithTag(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testCreateBookmarkTwice(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testCreateTwoDifferentBookmarks(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testUpdateBookmark(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testUpdateBookmarkWithContent(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testGetBookmark(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testGetBookmarkNotExistent(t *testing.T, db DB) {
	ctx := context.TODO()

	savedBookmark, exists, err := db.GetBookmark(ctx, 1, "")
	assert.NoError(t, err, "Get bookmark should not fail")
	assert.False(t, exists, "Bookmark should not exist")
	assert.Equal(t, model.BookmarkDTO{}, savedBookmark)
}

func testGetBookmarks(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testGetBookmarksWithSQLCharacters(t *testing.T, db DB) {
	ctx := context.TODO()

	// _ := 0
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

func testGetBookmarksCount(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testCreateTag(t *testing.T, db DB) {
	ctx := context.TODO()
	tag := model.Tag{Name: "shiori"}
	err := db.CreateTags(ctx, tag)
	assert.NoError(t, err, "Save tag must not fail")
}

func testCreateTags(t *testing.T, db DB) {
	ctx := context.TODO()
	err := db.CreateTags(ctx, model.Tag{Name: "shiori"}, model.Tag{Name: "shiori2"})
	assert.NoError(t, err, "Save tag must not fail")
}

func testSaveAccount(t *testing.T, db DB) {
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
		acc := model.Account{
			Username: "testuser",
			Config:   model.UserConfig{},
		}

		err := db.SaveAccount(ctx, acc)
		require.Nil(t, err)
	})
}

func testSaveAccountSettings(t *testing.T, db DB) {
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
		acc := model.Account{
			Username: "test",
			Config:   model.UserConfig{},
		}

		err := db.SaveAccountSettings(ctx, acc)
		require.Nil(t, err)
	})
}

func testGetAccount(t *testing.T, db DB) {
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
		// Insert test accounts
		testAccounts := []model.Account{
			{Username: "foo", Password: "bar", Owner: false},
			{Username: "hello", Password: "world", Owner: false},
			{Username: "foo_bar", Password: "foobar", Owner: true},
		}
		for _, acc := range testAccounts {
			err := db.SaveAccount(ctx, acc)
			assert.Nil(t, err)

			// Successful case
			account, exists, err := db.GetAccount(ctx, acc.Username)
			assert.Nil(t, err)
			assert.True(t, exists, "Expected account to exist")
			assert.Equal(t, acc.Username, account.Username)
		}
		// Falid case
		account, exists, err := db.GetAccount(ctx, "foobar")
		assert.NotNil(t, err)
		assert.False(t, exists, "Expected account to exist")
		assert.Empty(t, account.Username)
	})
}

func testGetAccounts(t *testing.T, db DB) {
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
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
	})
}

// TODO: Consider using `t.Parallel()` once we have automated database tests spawning databases using testcontainers.
func testUpdateBookmarkUpdatesModifiedTime(t *testing.T, db DB) {
	ctx := context.TODO()

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	resultBook, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	updatedBook := resultBook[0]
	updatedBook.Title = "modified"
	updatedBook.ModifiedAt = ""

	time.Sleep(1 * time.Second)
	resultUpdatedBooks, err := db.SaveBookmarks(ctx, false, updatedBook)
	assert.NoError(t, err, "Save bookmarks must not fail")

	assert.NotEqual(t, resultBook[0].ModifiedAt, resultUpdatedBooks[0].ModifiedAt)
	assert.Equal(t, resultBook[0].CreatedAt, resultUpdatedBooks[0].CreatedAt)
	assert.Equal(t, resultBook[0].CreatedAt, resultBook[0].ModifiedAt)
	assert.NoError(t, err, "Get bookmarks must not fail")

	assert.Equal(t, updatedBook.Title, resultUpdatedBooks[0].Title, "Saved bookmark must have updated Title")
}

// TODO: Consider using `t.Parallel()` once we have automated database tests spawning databases using testcontainers.
func testGetBoomarksWithTimeFilters(t *testing.T, db DB) {
	ctx := context.TODO()

	book1 := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori/one",
		Title: "Added First but Modified Last",
	}
	book2 := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori/second",
		Title: "Added Last but Modified First",
	}

	// create two new bookmark
	resultBook1, err := db.SaveBookmarks(ctx, true, book1)
	assert.NoError(t, err, "Save bookmarks must not fail")
	time.Sleep(1 * time.Second)
	resultBook2, err := db.SaveBookmarks(ctx, true, book2)
	assert.NoError(t, err, "Save bookmarks must not fail")

	// update those bookmarks
	updatedBook1 := resultBook1[0]
	updatedBook1.Title = "Added First but Modified Last Updated Title"
	updatedBook1.ModifiedAt = ""

	updatedBook2 := resultBook2[0]
	updatedBook2.Title = "Last Added but modified First Updated Title"
	updatedBook2.ModifiedAt = ""

	// modified bookmark2 first after one second modified bookmark1
	resultUpdatedBook2, err := db.SaveBookmarks(ctx, false, updatedBook2)
	assert.NoError(t, err, "Save bookmarks must not fail")
	time.Sleep(1 * time.Second)
	resultUpdatedBook1, err := db.SaveBookmarks(ctx, false, updatedBook1)
	assert.NoError(t, err, "Save bookmarks must not fail")

	// get diffrent filteter combination
	booksOrderByLastAdded, err := db.GetBookmarks(ctx, GetBookmarksOptions{
		IDs:         []int{resultUpdatedBook1[0].ID, resultUpdatedBook2[0].ID},
		OrderMethod: 1,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	booksOrderByLastModified, err := db.GetBookmarks(ctx, GetBookmarksOptions{
		IDs:         []int{resultUpdatedBook1[0].ID, resultUpdatedBook2[0].ID},
		OrderMethod: 2,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	booksOrderById, err := db.GetBookmarks(ctx, GetBookmarksOptions{
		IDs:         []int{resultUpdatedBook1[0].ID, resultUpdatedBook2[0].ID},
		OrderMethod: 0,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")

	// Check Last Added
	assert.Equal(t, booksOrderByLastAdded[0].Title, updatedBook2.Title)
	// Check Last Modified
	assert.Equal(t, booksOrderByLastModified[0].Title, updatedBook1.Title)
	// Second id should be 2 if order them by id
	assert.Equal(t, booksOrderById[1].ID, 2)
}
