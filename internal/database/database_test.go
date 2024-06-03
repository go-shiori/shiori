package database

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type databaseTestCase func(t *testing.T, db DB)
type testDatabaseFactory func(t *testing.T, ctx context.Context) (DB, error)

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
		// Accounts
		"testCreateAccount":            testCreateAccount,
		"testDeleteAccount":            testDeleteAccount,
		"testDeleteNonExistantAccount": testDeleteNonExistantAccount,
		"testSaveAccount":              testSaveAccount,
		"testSaveAccountSetting":       testSaveAccountSettings,
		"testGetAccount":               testGetAccount,
		"testGetAccounts":              testListAccounts,
		"testListAccountsWithPassword": testListAccountsWithPassword,
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

// ----------------- ACCOUNTS -----------------
func testCreateAccount(t *testing.T, db DB) {
	ctx := context.TODO()

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

func testDeleteAccount(t *testing.T, db DB) {
	ctx := context.TODO()

	acc := model.Account{
		Username: "testuser",
		Password: "testpass",
		Owner:    true,
	}
	storedAccount, err := db.SaveAccount(ctx, acc)
	assert.NoError(t, err, "Save account must not fail")

	err = db.DeleteAccount(ctx, storedAccount.ID)
	assert.NoError(t, err, "Delete account must not fail")

	_, exists, err := db.GetAccount(ctx, storedAccount.ID)
	assert.False(t, exists, "Account must not exist")
	assert.ErrorIs(t, err, ErrNotFound, "Get account must return not found error")
}

func testDeleteNonExistantAccount(t *testing.T, db DB) {
	ctx := context.TODO()
	err := db.DeleteAccount(ctx, model.DBID(99))
	assert.ErrorIs(t, err, ErrNotFound, "Delete account must fail")
}

func testSaveAccount(t *testing.T, db DB) {
	ctx := context.TODO()

	acc := model.Account{
		Username: "testuser",
		Config:   model.UserConfig{},
	}

	account, err := db.SaveAccount(ctx, acc)
	require.Nil(t, err)
	require.NotNil(t, account)
	require.NotEmpty(t, account.ID)
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

	// Insert test accounts
	testAccounts := []model.Account{
		{Username: "foo", Password: "bar", Owner: false},
		{Username: "hello", Password: "world", Owner: false},
		{Username: "foo_bar", Password: "foobar", Owner: true},
	}

	for _, acc := range testAccounts {
		storedAcc, err := db.SaveAccount(ctx, acc)
		assert.Nil(t, err)

		// Successful case
		account, exists, err := db.GetAccount(ctx, storedAcc.ID)
		assert.Nil(t, err)
		assert.True(t, exists, "Expected account to exist")
		assert.Equal(t, storedAcc.Username, account.Username)
	}

	// Failed case
	account, exists, err := db.GetAccount(ctx, 99)
	assert.NotNil(t, err)
	assert.False(t, exists, "Expected account to exist")
	assert.Empty(t, account.Username)
}

func testListAccounts(t *testing.T, db DB) {
	ctx := context.TODO()

	// prepare database
	testAccounts := []model.Account{
		{Username: "foo", Password: "bar", Owner: false},
		{Username: "hello", Password: "world", Owner: false},
		{Username: "foo_bar", Password: "foobar", Owner: true},
	}
	for _, acc := range testAccounts {
		_, err := db.SaveAccount(ctx, acc)
		assert.Nil(t, err)
	}

	tests := []struct {
		name     string
		options  ListAccountsOptions
		expected int
	}{
		{"default", ListAccountsOptions{}, 3},
		{"with owner", ListAccountsOptions{Owner: true}, 1},
		{"with keyword", ListAccountsOptions{Keyword: "foo"}, 2},
		{"with keyword and owner", ListAccountsOptions{Keyword: "hello", Owner: false}, 1},
		{"with no result", ListAccountsOptions{Keyword: "shiori"}, 0},
		{"with username", ListAccountsOptions{Username: "foo"}, 1},
		{"with non-existent username", ListAccountsOptions{Username: "non-existant"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, err := db.ListAccounts(ctx, tt.options)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, len(accounts))
		})
	}
}

func testListAccountsWithPassword(t *testing.T, db DB) {
	ctx := context.TODO()
	_, err := db.SaveAccount(ctx, model.Account{
		Username: "gopher",
		Password: "shiori",
	})
	assert.Nil(t, err)

	storedAccounts, err := db.ListAccounts(ctx, ListAccountsOptions{
		WithPassword: true,
	})
	for _, acc := range storedAccounts {
		require.NotEmpty(t, acc.Password)
	}
}
