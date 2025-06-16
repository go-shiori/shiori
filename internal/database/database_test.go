package database

import (
	"context"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type databaseTestCase func(t *testing.T, db model.DB)
type testDatabaseFactory func(t *testing.T, ctx context.Context) (model.DB, error)

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
		"testGetBookmarksWithTags":              testGetBookmarksWithTags,
		"testGetBookmarksWithSQLCharacters":     testGetBookmarksWithSQLCharacters,
		"testGetBookmarksCount":                 testGetBookmarksCount,
		"testSaveBookmark":                      testSaveBookmark,
		"testBulkUpdateBookmarkTags":            testBulkUpdateBookmarkTags,
		"testBookmarkExists":                    testBookmarkExists,
		// Tags
		"testCreateTag":             testCreateTag,
		"testCreateTags":            testCreateTags,
		"testTagExists":             testTagExists,
		"testGetTags":               testGetTags,
		"testGetTagsFunction":       testGetTagsFunction,
		"testGetTag":                testGetTag,
		"testGetTagNotExistent":     testGetTagNotExistent,
		"testUpdateTag":             testUpdateTag,
		"testRenameTag":             testRenameTag,
		"testDeleteTag":             testDeleteTag,
		"testDeleteTagNotExistent":  testDeleteTagNotExistent,
		"testAddTagToBookmark":      testAddTagToBookmark,
		"testRemoveTagFromBookmark": testRemoveTagFromBookmark,
		"testTagBookmarkEdgeCases":  testTagBookmarkEdgeCases,
		"testTagBookmarkOperations": testTagBookmarkOperations,
		// Accounts
		"testCreateAccount":              testCreateAccount,
		"testCreateDuplicateAccount":     testCreateDuplicateAccount,
		"testDeleteAccount":              testDeleteAccount,
		"testDeleteNonExistantAccount":   testDeleteNonExistantAccount,
		"testUpdateAccount":              testUpdateAccount,
		"testUpdateAccountDuplicateUser": testUpdateAccountDuplicateUser,
		"testGetAccount":                 testGetAccount,
		"testListAccounts":               testListAccounts,
		"testListAccountsWithPassword":   testListAccountsWithPassword,
	}

	for testName, testCase := range tests {
		t.Run(testName, func(tInner *testing.T) {
			ctx := context.TODO()
			db, err := dbFactory(t, ctx)
			require.NoError(tInner, err, "Error recreating database")
			testCase(tInner, db)
		})
	}
}

func testBookmarkAutoIncrement(t *testing.T, db model.DB) {
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

func testCreateBookmark(t *testing.T, db model.DB) {
	ctx := context.TODO()

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)

	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, 1, result[0].ID, "Saved bookmark must have an ID set")
}

func testCreateBookmarkWithContent(t *testing.T, db model.DB) {
	ctx := context.TODO()

	book := model.BookmarkDTO{
		URL:     "https://github.com/go-shiori/obelisk",
		Title:   "shiori",
		Content: "Some content",
		HTML:    "Some HTML content",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	books, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
		IDs:         []int{result[0].ID},
		WithContent: true,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	assert.Len(t, books, 1)

	assert.Equal(t, 1, books[0].ID, "Saved bookmark must have an ID set")
	assert.Equal(t, book.Content, books[0].Content, "Saved bookmark must have content")
	assert.Equal(t, book.HTML, books[0].HTML, "Saved bookmark must have HTML")
}

func testCreateBookmarkWithTag(t *testing.T, db model.DB) {
	ctx := context.TODO()

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "shiori",
		Tags: []model.TagDTO{
			{
				Tag: model.Tag{
					Name: "test-tag",
				},
			},
		},
	}

	result, err := db.SaveBookmarks(ctx, true, book)

	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, book.URL, result[0].URL)
	assert.Equal(t, book.Tags[0].Name, result[0].Tags[0].Name)
}

func testCreateBookmarkTwice(t *testing.T, db model.DB) {
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

func testCreateTwoDifferentBookmarks(t *testing.T, db model.DB) {
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

func testUpdateBookmark(t *testing.T, db model.DB) {
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

func testUpdateBookmarkWithContent(t *testing.T, db model.DB) {
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

	books, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
		IDs:         []int{result[0].ID},
		WithContent: true,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	assert.Len(t, books, 1)

	assert.Equal(t, 1, books[0].ID, "Saved bookmark must have an ID set")
	assert.Equal(t, updatedBook.Content, books[0].Content, "Saved bookmark must have updated content")
	assert.Equal(t, updatedBook.HTML, books[0].HTML, "Saved bookmark must have updated HTML")
}

func testGetBookmark(t *testing.T, db model.DB) {
	ctx := context.TODO()

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	result, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	savedBookmark, exists, err := db.GetBookmark(ctx, result[0].ID, "")
	assert.NoError(t, err, "Get bookmark should not fail")
	assert.True(t, exists, "Bookmark should exist")
	assert.Equal(t, result[0].ID, savedBookmark.ID, "Retrieved bookmark should be the same")
	assert.Equal(t, book.URL, savedBookmark.URL, "Retrieved bookmark should be the same")
}

func testGetBookmarkNotExistent(t *testing.T, db model.DB) {
	ctx := context.TODO()

	savedBookmark, exists, err := db.GetBookmark(ctx, 1, "")
	assert.NoError(t, err, "Get bookmark should not fail")
	assert.False(t, exists, "Bookmark should not exist")
	assert.Equal(t, model.BookmarkDTO{}, savedBookmark)
}

func testGetBookmarks(t *testing.T, db model.DB) {
	ctx := context.TODO()

	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	bookmarks, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	savedBookmark := bookmarks[0]

	results, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
		Keyword: "go-shiori",
	})

	assert.NoError(t, err, "Get bookmarks should not fail")
	assert.Len(t, results, 1, "results should contain one item")
	assert.Equal(t, savedBookmark.ID, results[0].ID, "bookmark should be the one saved")
}

func testGetBookmarksWithSQLCharacters(t *testing.T, db model.DB) {
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
			_, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
				Keyword: char,
			})
			assert.NoError(t, err, "Get bookmarks should not fail")
		})

		t.Run("GetBookmarksCount/"+char, func(t *testing.T) {
			_, err := db.GetBookmarksCount(ctx, model.DBGetBookmarksOptions{
				Keyword: char,
			})
			assert.NoError(t, err, "Get bookmarks count should not fail")
		})
	}
}

func testGetBookmarksWithTags(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create test tags
	tags := []model.Tag{
		{Name: "programming"},
		{Name: "golang"},
		{Name: "database"},
		{Name: "testing"},
	}
	createdTags, err := db.CreateTags(ctx, tags...)
	require.NoError(t, err)
	require.Len(t, createdTags, 4)

	// Create bookmarks with different tag combinations
	bookmarks := []model.BookmarkDTO{
		{
			URL:   "https://golang.org",
			Title: "Go Language",
			Tags: []model.TagDTO{
				{Tag: model.Tag{Name: "programming"}},
				{Tag: model.Tag{Name: "golang"}},
			},
		},
		{
			URL:   "https://postgresql.org",
			Title: "PostgreSQL",
			Tags: []model.TagDTO{
				{Tag: model.Tag{Name: "programming"}},
				{Tag: model.Tag{Name: "database"}},
			},
		},
		{
			URL:   "https://sqlite.org",
			Title: "SQLite",
			Tags: []model.TagDTO{
				{Tag: model.Tag{Name: "database"}},
			},
		},
		{
			URL:   "https://example.com",
			Title: "No Tags Example",
		},
	}

	// Save all bookmarks
	for _, bookmark := range bookmarks {
		results, err := db.SaveBookmarks(ctx, true, bookmark)
		require.NoError(t, err)
		require.Len(t, results, 1)
	}

	tests := []struct {
		name           string
		opts           model.DBGetBookmarksOptions
		expectedCount  int
		expectedTitles []string
	}{
		{
			name: "single tag - programming",
			opts: model.DBGetBookmarksOptions{
				Tags: []string{"programming"},
			},
			expectedCount:  2,
			expectedTitles: []string{"Go Language", "PostgreSQL"},
		},
		{
			name: "multiple tags - programming AND golang",
			opts: model.DBGetBookmarksOptions{
				Tags: []string{"programming", "golang"},
			},
			expectedCount:  1,
			expectedTitles: []string{"Go Language"},
		},
		{
			name: "all tags using *",
			opts: model.DBGetBookmarksOptions{
				Tags: []string{"*"},
			},
			expectedCount:  3,
			expectedTitles: []string{"Go Language", "PostgreSQL", "SQLite"},
		},
		{
			name: "exclude database tag",
			opts: model.DBGetBookmarksOptions{
				ExcludedTags: []string{"database"},
			},
			expectedCount:  2,
			expectedTitles: []string{"Go Language", "No Tags Example"},
		},
		{
			name: "no tags only",
			opts: model.DBGetBookmarksOptions{
				ExcludedTags: []string{"*"},
			},
			expectedCount:  1,
			expectedTitles: []string{"No Tags Example"},
		},
		{
			name: "non-existent tag",
			opts: model.DBGetBookmarksOptions{
				Tags: []string{"nonexistent"},
			},
			expectedCount:  0,
			expectedTitles: []string{},
		},
	}

	t.Run("ensure tags are present", func(t *testing.T) {
		tags, err := db.GetTags(ctx, model.DBListTagsOptions{})
		require.NoError(t, err)
		assert.Len(t, tags, 4)
	})

	t.Run("ensure test data is correct", func(t *testing.T) {
		results, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{})
		require.NoError(t, err)
		require.Len(t, results, 4)
		for _, book := range results {
			if book.Title == "No Tags Example" {
				assert.Empty(t, book.Tags)
			} else {
				assert.NotEmpty(t, book.Tags)
			}

			// Ensure tags contain their ID and name
			for _, tag := range book.Tags {
				assert.NotZero(t, tag.ID)
				assert.NotEmpty(t, tag.Name)
			}
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.GetBookmarks(ctx, tt.opts)
			require.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)

			// Check if all expected titles are present
			titles := make([]string, len(results))
			for i, result := range results {
				titles[i] = result.Title
			}
			assert.ElementsMatch(t, tt.expectedTitles, titles)
		})
	}
}

func testGetBookmarksCount(t *testing.T, db model.DB) {
	ctx := context.TODO()

	expectedCount := 1
	book := model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori",
		Title: "shiori",
	}

	_, err := db.SaveBookmarks(ctx, true, book)
	assert.NoError(t, err, "Save bookmarks must not fail")

	count, err := db.GetBookmarksCount(ctx, model.DBGetBookmarksOptions{
		Keyword: "go-shiori",
	})
	assert.NoError(t, err, "Get bookmarks count should not fail")
	assert.Equal(t, count, expectedCount, "count should be %d", expectedCount)
}

func testCreateTag(t *testing.T, db model.DB) {
	ctx := context.TODO()
	tag := model.Tag{Name: "shiori"}
	createdTags, err := db.CreateTags(ctx, tag)
	assert.NoError(t, err, "Save tag must not fail")
	assert.Len(t, createdTags, 1, "Should return one created tag")
	assert.Greater(t, createdTags[0].ID, 0, "Created tag should have a valid ID")
	assert.Equal(t, "shiori", createdTags[0].Name, "Created tag should have the correct name")
}

func testCreateTags(t *testing.T, db model.DB) {
	ctx := context.TODO()
	createdTags, err := db.CreateTags(ctx, model.Tag{Name: "shiori"}, model.Tag{Name: "shiori2"})
	assert.NoError(t, err, "Save tag must not fail")
	assert.Len(t, createdTags, 2, "Should return two created tags")
	assert.Greater(t, createdTags[0].ID, 0, "First created tag should have a valid ID")
	assert.Greater(t, createdTags[1].ID, 0, "Second created tag should have a valid ID")
	assert.Equal(t, "shiori", createdTags[0].Name, "First created tag should have the correct name")
	assert.Equal(t, "shiori2", createdTags[1].Name, "Second created tag should have the correct name")
}

// ----------------- ACCOUNTS -----------------
func testCreateAccount(t *testing.T, db model.DB) {
	ctx := context.TODO()

	acc := model.Account{
		Username: "testuser",
		Password: "testpass",
		Owner:    true,
	}
	insertedAccount, err := db.CreateAccount(ctx, acc)
	assert.NoError(t, err, "Save account must not fail")
	assert.Equal(t, acc.Username, insertedAccount.Username, "Saved account must have an username set")
	assert.Equal(t, acc.Password, insertedAccount.Password, "Saved account must have a password set")
	assert.Equal(t, acc.Owner, insertedAccount.Owner, "Saved account must have an owner set")
	assert.NotEmpty(t, insertedAccount.ID, "Saved account must have an ID set")
}

func testDeleteAccount(t *testing.T, db model.DB) {
	ctx := context.TODO()

	acc := model.Account{
		Username: "testuser",
		Password: "testpass",
		Owner:    true,
	}
	storedAccount, err := db.CreateAccount(ctx, acc)
	assert.NoError(t, err, "Save account must not fail")

	err = db.DeleteAccount(ctx, storedAccount.ID)
	assert.NoError(t, err, "Delete account must not fail")

	_, exists, err := db.GetAccount(ctx, storedAccount.ID)
	assert.False(t, exists, "Account must not exist")
	assert.ErrorIs(t, err, ErrNotFound, "Get account must return not found error")
}

func testDeleteNonExistantAccount(t *testing.T, db model.DB) {
	ctx := context.TODO()
	err := db.DeleteAccount(ctx, model.DBID(99))
	assert.ErrorIs(t, err, ErrNotFound, "Delete account must fail")
}

func testUpdateAccount(t *testing.T, db model.DB) {
	ctx := context.TODO()

	acc := model.Account{
		Username: "testuser",
		Password: "testpass",
		Owner:    true,
		Config: model.UserConfig{
			ShowId: true,
		},
	}

	account, err := db.CreateAccount(ctx, acc)
	require.Nil(t, err)
	require.NotNil(t, account)
	require.NotEmpty(t, account.ID)

	account, _, err = db.GetAccount(ctx, account.ID)
	require.Nil(t, err)

	t.Run("update", func(t *testing.T) {
		acc := model.Account{
			ID:       account.ID,
			Username: "asdlasd",
			Owner:    false,
			Password: "another",
			Config: model.UserConfig{
				ShowId: false,
			},
		}

		err := db.UpdateAccount(ctx, acc)
		require.Nil(t, err)

		updatedAccount, exists, err := db.GetAccount(ctx, account.ID)
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, acc.Username, updatedAccount.Username)
		require.Equal(t, acc.Owner, updatedAccount.Owner)
		require.Equal(t, acc.Config, updatedAccount.Config)
		require.NotEqual(t, acc.Password, account.Password)
	})
}

func testGetAccount(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Insert test accounts
	testAccounts := []model.Account{
		{Username: "foo", Password: "bar", Owner: false},
		{Username: "hello", Password: "world", Owner: false},
		{Username: "foo_bar", Password: "foobar", Owner: true},
	}

	for _, acc := range testAccounts {
		storedAcc, err := db.CreateAccount(ctx, acc)
		assert.Nil(t, err)

		// Successful case
		account, exists, err := db.GetAccount(ctx, storedAcc.ID)
		assert.Nil(t, err)
		assert.True(t, exists, "Expected account to exist")
		assert.Equal(t, storedAcc.Username, account.Username)
	}

	// Failed case
	account, exists, err := db.GetAccount(ctx, 99)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.False(t, exists, "Expected account to exist")
	assert.Empty(t, account.Username)
}

func testListAccounts(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// prepare database
	testAccounts := []model.Account{
		{Username: "foo", Password: "bar", Owner: false},
		{Username: "hello", Password: "world", Owner: false},
		{Username: "foo_bar", Password: "foobar", Owner: true},
	}
	for _, acc := range testAccounts {
		_, err := db.CreateAccount(ctx, acc)
		assert.Nil(t, err)
	}

	tests := []struct {
		name     string
		options  model.DBListAccountsOptions
		expected int
	}{
		{"default", model.DBListAccountsOptions{}, 3},
		{"with owner", model.DBListAccountsOptions{Owner: true}, 1},
		{"with keyword", model.DBListAccountsOptions{Keyword: "foo"}, 2},
		{"with keyword and owner", model.DBListAccountsOptions{Keyword: "hello", Owner: false}, 1},
		{"with no result", model.DBListAccountsOptions{Keyword: "shiori"}, 0},
		{"with username", model.DBListAccountsOptions{Username: "foo"}, 1},
		{"with non-existent username", model.DBListAccountsOptions{Username: "non-existant"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, err := db.ListAccounts(ctx, tt.options)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, len(accounts))
		})
	}
}

func testCreateDuplicateAccount(t *testing.T, db model.DB) {
	ctx := context.TODO()

	acc := model.Account{
		Username: "testuser",
		Password: "testpass",
		Owner:    false,
	}

	// Create first account
	_, err := db.CreateAccount(ctx, acc)
	assert.NoError(t, err, "First account creation must not fail")

	// Try to create account with same username
	_, err = db.CreateAccount(ctx, acc)
	assert.ErrorIs(t, err, ErrAlreadyExists, "Creating duplicate account must return ErrAlreadyExists")
}

func testUpdateAccountDuplicateUser(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create first account
	acc1 := model.Account{
		Username: "testuser1",
		Password: "testpass",
		Owner:    false,
	}
	storedAcc1, err := db.CreateAccount(ctx, acc1)
	assert.NoError(t, err, "First account creation must not fail")

	// Create second account
	acc2 := model.Account{
		Username: "testuser2",
		Password: "testpass",
		Owner:    false,
	}
	storedAcc2, err := db.CreateAccount(ctx, acc2)
	assert.NoError(t, err, "Second account creation must not fail")

	// Try to update second account to have same username as first
	storedAcc2.Username = storedAcc1.Username
	err = db.UpdateAccount(ctx, *storedAcc2)
	assert.ErrorIs(t, err, ErrAlreadyExists, "Updating to duplicate username must return ErrAlreadyExists")
}

func testListAccountsWithPassword(t *testing.T, db model.DB) {
	ctx := context.TODO()
	_, err := db.CreateAccount(ctx, model.Account{
		Username: "gopher",
		Password: "shiori",
	})
	assert.Nil(t, err)

	storedAccounts, err := db.ListAccounts(ctx, model.DBListAccountsOptions{
		WithPassword: true,
	})
	require.NoError(t, err)
	for _, acc := range storedAccounts {
		require.NotEmpty(t, acc.Password)
	}
}

// TODO: Consider using `t.Parallel()` once we have automated database tests spawning databases using testcontainers.
func testUpdateBookmarkUpdatesModifiedTime(t *testing.T, db model.DB) {
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
func testGetBoomarksWithTimeFilters(t *testing.T, db model.DB) {
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
	booksOrderByLastAdded, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
		IDs:         []int{resultUpdatedBook1[0].ID, resultUpdatedBook2[0].ID},
		OrderMethod: 1,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	booksOrderByLastModified, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
		IDs:         []int{resultUpdatedBook1[0].ID, resultUpdatedBook2[0].ID},
		OrderMethod: 2,
	})
	assert.NoError(t, err, "Get bookmarks must not fail")
	booksOrderById, err := db.GetBookmarks(ctx, model.DBGetBookmarksOptions{
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

// Additional tag test functions

func testGetTags(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create initial tag to ensure there's at least one tag
	initialTag := model.Tag{Name: "initial-test-tag"}
	_, err := db.CreateTags(ctx, initialTag)
	require.NoError(t, err)

	// Create additional tags
	tags := []model.Tag{
		{Name: "tag1"},
		{Name: "tag2"},
		{Name: "tag3"},
	}
	createdTags, err := db.CreateTags(ctx, tags...)
	require.NoError(t, err)
	require.Len(t, createdTags, 3)

	// Fetch all tags
	fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(fetchedTags), 4) // At least 3 new tags + 1 initial tag

	// Check that all expected tags are present
	tagNames := make(map[string]bool)
	for _, tag := range fetchedTags {
		tagNames[tag.Name] = true
	}

	assert.True(t, tagNames["tag1"], "Tag 'tag1' should be present")
	assert.True(t, tagNames["tag2"], "Tag 'tag2' should be present")
	assert.True(t, tagNames["tag3"], "Tag 'tag3' should be present")
	assert.True(t, tagNames["initial-test-tag"], "Tag 'initial-test-tag' should be present")
}

func testGetTag(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create a tag
	tag := model.Tag{Name: "get-tag-test"}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Get the tag
	fetchedTag, exists, err := db.GetTag(ctx, tagID)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, tagID, fetchedTag.ID)
	assert.Equal(t, tag.Name, fetchedTag.Name)
}

func testGetTagNotExistent(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Test non-existent tag
	nonExistentTag, exists, err := db.GetTag(ctx, 9999)
	require.NoError(t, err)
	require.False(t, exists)
	assert.Empty(t, nonExistentTag.Name)
}

func testUpdateTag(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create a tag
	tag := model.Tag{Name: "update-tag-test"}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)

	// Update the tag
	tagToUpdate := model.Tag{
		ID:   createdTags[0].ID,
		Name: "updated-tag",
	}
	err = db.UpdateTag(ctx, tagToUpdate)
	require.NoError(t, err)

	// Verify the tag was updated
	updatedTag, exists, err := db.GetTag(ctx, tagToUpdate.ID)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "updated-tag", updatedTag.Name)
}

func testRenameTag(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create a tag
	tag := model.Tag{Name: "rename-tag-test"}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Rename the tag
	err = db.RenameTag(ctx, tagID, "renamed-tag")
	require.NoError(t, err)

	// Verify the tag was renamed
	renamedTag, exists, err := db.GetTag(ctx, tagID)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "renamed-tag", renamedTag.Name)
}

func testDeleteTag(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create a tag
	tag := model.Tag{Name: "delete-tag-test"}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Delete the tag
	err = db.DeleteTag(ctx, tagID)
	require.NoError(t, err)

	// Verify the tag was deleted
	_, exists, err := db.GetTag(ctx, tagID)
	require.NoError(t, err)
	require.False(t, exists)
}

func testDeleteTagNotExistent(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Test deleting a non-existent tag
	err := db.DeleteTag(ctx, 9999)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound, "Error should be ErrNotFound")
}

func testSaveBookmark(t *testing.T, db model.DB) {
	ctx := context.TODO()

	t.Run("invalid_bookmark_id", func(t *testing.T) {
		bookmark := model.Bookmark{
			ID:    0, // Invalid ID
			URL:   "https://example.com",
			Title: "Example",
		}
		err := db.SaveBookmark(ctx, bookmark)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bookmark ID must be greater than 0")
	})

	t.Run("empty_url", func(t *testing.T) {
		bookmark := model.Bookmark{
			ID:    1,
			URL:   "", // Empty URL
			Title: "Example",
		}
		err := db.SaveBookmark(ctx, bookmark)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "URL must not be empty")
	})

	t.Run("empty_title", func(t *testing.T) {
		bookmark := model.Bookmark{
			ID:    1,
			URL:   "https://example.com",
			Title: "", // Empty title
		}
		err := db.SaveBookmark(ctx, bookmark)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title must not be empty")
	})

	t.Run("successful_update", func(t *testing.T) {
		// First create a bookmark
		bookmark := model.BookmarkDTO{
			URL:   "https://example.com",
			Title: "Example",
		}
		results, err := db.SaveBookmarks(ctx, true, bookmark)
		require.NoError(t, err)
		bookmarkID := results[0].ID

		// Now update it
		updatedBookmark := model.Bookmark{
			ID:      bookmarkID,
			URL:     "https://updated-example.com",
			Title:   "Updated Example",
			Excerpt: "Updated excerpt",
			Author:  "Updated Author",
			Public:  1, // Use 1 for SQLite, should work for other DBs too
		}

		err = db.SaveBookmark(ctx, updatedBookmark)
		require.NoError(t, err)

		// Verify the bookmark was updated
		retrievedBookmark, exists, err := db.GetBookmark(ctx, bookmarkID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Equal(t, updatedBookmark.URL, retrievedBookmark.URL)
		assert.Equal(t, updatedBookmark.Title, retrievedBookmark.Title)
		assert.Equal(t, updatedBookmark.Excerpt, retrievedBookmark.Excerpt)
		assert.Equal(t, updatedBookmark.Author, retrievedBookmark.Author)
		assert.Equal(t, updatedBookmark.Public, retrievedBookmark.Public)
	})
}

func testBulkUpdateBookmarkTags(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create test bookmarks
	bookmark1 := model.BookmarkDTO{
		URL:   "https://example1.com",
		Title: "Example 1",
	}
	bookmark2 := model.BookmarkDTO{
		URL:   "https://example2.com",
		Title: "Example 2",
	}
	bookmark3 := model.BookmarkDTO{
		URL:   "https://example3.com",
		Title: "Example 3",
	}

	results1, err := db.SaveBookmarks(ctx, true, bookmark1)
	require.NoError(t, err)
	bookmark1ID := results1[0].ID

	results2, err := db.SaveBookmarks(ctx, true, bookmark2)
	require.NoError(t, err)
	bookmark2ID := results2[0].ID

	results3, err := db.SaveBookmarks(ctx, true, bookmark3)
	require.NoError(t, err)
	bookmark3ID := results3[0].ID

	// Create test tags
	tag1 := model.Tag{Name: "tag1-bulk-test"}
	tag2 := model.Tag{Name: "tag2-bulk-test"}
	tag3 := model.Tag{Name: "tag3-bulk-test"}
	tag4 := model.Tag{Name: "tag4-bulk-test"}

	createdTags, err := db.CreateTags(ctx, tag1, tag2, tag3, tag4)
	require.NoError(t, err)
	require.Len(t, createdTags, 4)

	tag1ID := createdTags[0].ID
	tag2ID := createdTags[1].ID
	tag3ID := createdTags[2].ID
	tag4ID := createdTags[3].ID

	t.Run("empty_bookmark_ids", func(t *testing.T) {
		err := db.BulkUpdateBookmarkTags(ctx, []int{}, []int{tag1ID, tag2ID})
		require.NoError(t, err, "Empty bookmark IDs should not cause an error")
	})

	t.Run("empty_tag_ids", func(t *testing.T) {
		err := db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID, bookmark2ID}, []int{})
		require.NoError(t, err, "Empty tag IDs should not cause an error")

		// Verify tags were removed
		bookmark, exists, err := db.GetBookmark(ctx, bookmark1ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Empty(t, bookmark.Tags, "Tags should be empty after update with empty tag IDs")
	})

	t.Run("non_existent_bookmark", func(t *testing.T) {
		nonExistentID := 9999
		err := db.BulkUpdateBookmarkTags(ctx, []int{nonExistentID}, []int{tag1ID})
		require.Error(t, err, "Non-existent bookmark ID should cause an error")
		assert.Contains(t, err.Error(), "some bookmarks do not exist")
	})

	t.Run("non_existent_tag", func(t *testing.T) {
		nonExistentID := 9999
		err := db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID}, []int{nonExistentID})
		require.Error(t, err, "Non-existent tag ID should cause an error")
		assert.Contains(t, err.Error(), "some tags do not exist")
	})

	t.Run("multiple_non_existent_bookmarks", func(t *testing.T) {
		err := db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID, 9998, 9999}, []int{tag1ID})
		require.Error(t, err, "Multiple non-existent bookmark IDs should cause an error")
		assert.Contains(t, err.Error(), "some bookmarks do not exist")
	})

	t.Run("multiple_non_existent_tags", func(t *testing.T) {
		err := db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID}, []int{tag1ID, 9998, 9999})
		require.Error(t, err, "Multiple non-existent tag IDs should cause an error")
		assert.Contains(t, err.Error(), "some tags do not exist")
	})

	t.Run("successful_update", func(t *testing.T) {
		// Update both bookmarks with both tags
		err := db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID, bookmark2ID}, []int{tag1ID, tag2ID})
		require.NoError(t, err, "Bulk update should succeed")

		// Verify bookmark1 has both tags
		bookmark1, exists, err := db.GetBookmark(ctx, bookmark1ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark1.Tags, 2, "Bookmark 1 should have 2 tags")

		// Verify bookmark2 has both tags
		bookmark2, exists, err := db.GetBookmark(ctx, bookmark2ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark2.Tags, 2, "Bookmark 2 should have 2 tags")

		// Verify tag names
		tagNames := make(map[string]bool)
		for _, tag := range bookmark1.Tags {
			tagNames[tag.Name] = true
		}
		assert.True(t, tagNames[tag1.Name], "Bookmark 1 should have tag1")
		assert.True(t, tagNames[tag2.Name], "Bookmark 1 should have tag2")

		// Update with a single tag
		err = db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID}, []int{tag1ID})
		require.NoError(t, err, "Update with single tag should succeed")

		// Verify bookmark1 now has only one tag
		bookmark1, exists, err = db.GetBookmark(ctx, bookmark1ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark1.Tags, 1, "Bookmark 1 should have 1 tag after update")
		assert.Equal(t, tag1.Name, bookmark1.Tags[0].Name, "Bookmark 1 should have tag1")

		// Verify bookmark2 still has both tags
		bookmark2, exists, err = db.GetBookmark(ctx, bookmark2ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark2.Tags, 2, "Bookmark 2 should still have 2 tags")
	})

	t.Run("multiple_updates", func(t *testing.T) {
		// First update
		err := db.BulkUpdateBookmarkTags(ctx, []int{bookmark3ID}, []int{tag1ID, tag2ID})
		require.NoError(t, err, "First update should succeed")

		// Verify bookmark3 has both tags
		bookmark3, exists, err := db.GetBookmark(ctx, bookmark3ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark3.Tags, 2, "Bookmark 3 should have 2 tags after first update")

		// Second update with different tags
		err = db.BulkUpdateBookmarkTags(ctx, []int{bookmark3ID}, []int{tag3ID, tag4ID})
		require.NoError(t, err, "Second update should succeed")

		// Verify bookmark3 now has the new tags and not the old ones
		bookmark3, exists, err = db.GetBookmark(ctx, bookmark3ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark3.Tags, 2, "Bookmark 3 should have 2 tags after second update")

		// Check tag names
		tagNames := make(map[string]bool)
		for _, tag := range bookmark3.Tags {
			tagNames[tag.Name] = true
		}
		assert.False(t, tagNames[tag1.Name], "Bookmark 3 should not have tag1 after second update")
		assert.False(t, tagNames[tag2.Name], "Bookmark 3 should not have tag2 after second update")
		assert.True(t, tagNames[tag3.Name], "Bookmark 3 should have tag3 after second update")
		assert.True(t, tagNames[tag4.Name], "Bookmark 3 should have tag4 after second update")
	})

	t.Run("update_multiple_bookmarks_with_different_initial_tags", func(t *testing.T) {
		// Setup: bookmark1 has tag1, bookmark2 has tag1 and tag2
		err := db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID}, []int{tag1ID})
		require.NoError(t, err)

		err = db.BulkUpdateBookmarkTags(ctx, []int{bookmark2ID}, []int{tag1ID, tag2ID})
		require.NoError(t, err)

		// Verify initial state
		bookmark1, exists, err := db.GetBookmark(ctx, bookmark1ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark1.Tags, 1, "Bookmark 1 should have 1 tag initially")

		bookmark2, exists, err := db.GetBookmark(ctx, bookmark2ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark2.Tags, 2, "Bookmark 2 should have 2 tags initially")

		// Update both bookmarks with tag3 and tag4
		err = db.BulkUpdateBookmarkTags(ctx, []int{bookmark1ID, bookmark2ID}, []int{tag3ID, tag4ID})
		require.NoError(t, err, "Bulk update should succeed")

		// Verify both bookmarks now have tag3 and tag4 only
		bookmark1, exists, err = db.GetBookmark(ctx, bookmark1ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark1.Tags, 2, "Bookmark 1 should have 2 tags after update")

		bookmark2, exists, err = db.GetBookmark(ctx, bookmark2ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		assert.Len(t, bookmark2.Tags, 2, "Bookmark 2 should have 2 tags after update")

		// Check tag names for bookmark1
		tagNames1 := make(map[string]bool)
		for _, tag := range bookmark1.Tags {
			tagNames1[tag.Name] = true
		}
		assert.False(t, tagNames1[tag1.Name], "Bookmark 1 should not have tag1 after update")
		assert.False(t, tagNames1[tag2.Name], "Bookmark 1 should not have tag2 after update")
		assert.True(t, tagNames1[tag3.Name], "Bookmark 1 should have tag3 after update")
		assert.True(t, tagNames1[tag4.Name], "Bookmark 1 should have tag4 after update")

		// Check tag names for bookmark2
		tagNames2 := make(map[string]bool)
		for _, tag := range bookmark2.Tags {
			tagNames2[tag.Name] = true
		}
		assert.False(t, tagNames2[tag1.Name], "Bookmark 2 should not have tag1 after update")
		assert.False(t, tagNames2[tag2.Name], "Bookmark 2 should not have tag2 after update")
		assert.True(t, tagNames2[tag3.Name], "Bookmark 2 should have tag3 after update")
		assert.True(t, tagNames2[tag4.Name], "Bookmark 2 should have tag4 after update")
	})
}
