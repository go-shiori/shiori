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
		"testGetTags":    testGetTags,
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
	_, err := db.CreateTags(ctx, tag)
	assert.NoError(t, err, "Save tag must not fail")
}

func testCreateTags(t *testing.T, db DB) {
	ctx := context.TODO()

	t.Run("create multiple tags", func(t *testing.T) {
		tags := []model.Tag{
			{Name: "tag1"},
			{Name: "tag2"},
			{Name: "tag3"},
		}

		createdTags, err := db.CreateTags(ctx, tags...)
		assert.NoError(t, err, "Creating tags must not fail")
		assert.Len(t, createdTags, len(tags), "Should create all tags")

		for i, tag := range createdTags {
			assert.NotZero(t, tag.ID, "Created tag should have non-zero ID")
			assert.Equal(t, tags[i].Name, tag.Name, "Created tag should have correct name")
		}
	})

	t.Run("create empty tags slice", func(t *testing.T) {
		createdTags, err := db.CreateTags(ctx)
		assert.NoError(t, err, "Creating empty tags slice should not fail")
		assert.Empty(t, createdTags, "Should return empty slice for empty input")
	})

	t.Run("create duplicate tags", func(t *testing.T) {
		tag := model.Tag{Name: "duplicate"}

		// Create first tag
		tags1, err := db.CreateTags(ctx, tag)
		assert.NoError(t, err, "First tag creation should succeed")
		assert.Len(t, tags1, 1)

		// Try to create duplicate
		_, err = db.CreateTags(ctx, tag)
		assert.Error(t, err, "Duplicate tag creation should fail")
	})
}

func testGetTags(t *testing.T, db DB) {
	ctx := context.TODO()

	t.Run("get tags from empty database", func(t *testing.T) {
		tags, err := db.GetTags(ctx)
		assert.NoError(t, err, "Getting tags should not fail")
		assert.Empty(t, tags, "Should return empty slice when no tags exist")
	})

	t.Run("get existing tags", func(t *testing.T) {
		// Create some test tags first
		testTags := []model.Tag{
			{Name: "test1"},
			{Name: "test2"},
			{Name: "test3"},
		}
		createdTags, err := db.CreateTags(ctx, testTags...)
		assert.NoError(t, err, "Creating test tags should not fail")

		// Create some bookmarks with these tags
		book1 := model.BookmarkDTO{
			URL:   "https://example1.com",
			Title: "Example 1",
			Tags:  []model.Tag{createdTags[0], createdTags[1]}, // test1, test2
		}
		book2 := model.BookmarkDTO{
			URL:   "https://example2.com",
			Title: "Example 2",
			Tags:  []model.Tag{createdTags[1], createdTags[2]}, // test2, test3
		}

		_, err = db.SaveBookmarks(ctx, true, book1, book2)
		assert.NoError(t, err, "Creating bookmarks should not fail")

		// Get all tags
		tags, err := db.GetTags(ctx)
		assert.NoError(t, err, "Getting tags should not fail")
		assert.Len(t, tags, len(testTags), "Should return all created tags")

		// Verify returned tags
		tagMap := make(map[string]model.Tag)
		for _, tag := range tags {
			tagMap[tag.Name] = tag
			assert.NotZero(t, tag.ID, "Tag should have non-zero ID")
			assert.NotEmpty(t, tag.Name, "Tag should have non-empty name")
		}

		// Verify bookmark counts
		assert.Equal(t, 1, tagMap["test1"].BookmarkCount, "test1 should have 1 bookmark")
		assert.Equal(t, 2, tagMap["test2"].BookmarkCount, "test2 should have 2 bookmarks")
		assert.Equal(t, 1, tagMap["test3"].BookmarkCount, "test3 should have 1 bookmark")
	})

	t.Run("get tags after bookmark deletion", func(t *testing.T) {
		// Create a tag
		testTag := model.Tag{Name: "delete-test"}
		createdTags, err := db.CreateTags(ctx, testTag)
		assert.NoError(t, err, "Creating tag should not fail")
		assert.Len(t, createdTags, 1)

		// Create a bookmark with this tag
		book := model.BookmarkDTO{
			URL:   "https://delete-example.com",
			Title: "Delete Example",
			Tags:  createdTags,
		}
		savedBooks, err := db.SaveBookmarks(ctx, true, book)
		assert.NoError(t, err, "Creating bookmark should not fail")
		assert.Len(t, savedBooks, 1)

		// Verify tag exists with count 1
		tags, err := db.GetTags(ctx)
		assert.NoError(t, err, "Getting tags should not fail")
		var found bool
		for _, tag := range tags {
			if tag.Name == testTag.Name {
				assert.Equal(t, 1, tag.BookmarkCount, "Tag should have 1 bookmark")
				found = true
				break
			}
		}
		assert.True(t, found, "Should find the test tag")

		// Delete the bookmark
		err = db.DeleteBookmarks(ctx, savedBooks[0].ID)
		assert.NoError(t, err, "Deleting bookmark should not fail")

		// Verify tag is no longer returned (since it has no bookmarks)
		tags, err = db.GetTags(ctx)
		assert.NoError(t, err, "Getting tags after delete should not fail")
		found = false
		for _, tag := range tags {
			if tag.Name == testTag.Name {
				found = true
				break
			}
		}
		assert.False(t, found, "Tag should not be returned after its bookmark was deleted")
	})
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
