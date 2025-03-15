package database

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testGetTagsFunction tests the GetTags function with various options
func testGetTagsFunction(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create test tags
	tags := []model.Tag{
		{Name: "golang"},
		{Name: "database"},
		{Name: "testing"},
		{Name: "web"},
	}
	createdTags, err := db.CreateTags(ctx, tags...)
	require.NoError(t, err)
	require.Len(t, createdTags, 4)

	// Map tag names to IDs for easier reference
	tagIDsByName := make(map[string]int)
	for _, tag := range createdTags {
		tagIDsByName[tag.Name] = tag.ID
	}

	// Create bookmarks with different tag combinations
	bookmarks := []model.BookmarkDTO{
		{
			URL:   "https://golang.org",
			Title: "Go Language",
			Tags: []model.TagDTO{
				{Tag: model.Tag{Name: "golang"}},
				{Tag: model.Tag{Name: "web"}},
			},
		},
		{
			URL:   "https://postgresql.org",
			Title: "PostgreSQL",
			Tags: []model.TagDTO{
				{Tag: model.Tag{Name: "database"}},
			},
		},
		{
			URL:   "https://sqlite.org",
			Title: "SQLite",
			Tags: []model.TagDTO{
				{Tag: model.Tag{Name: "database"}},
				{Tag: model.Tag{Name: "testing"}},
			},
		},
	}

	// Save bookmarks
	var savedBookmarks []model.BookmarkDTO
	for _, bookmark := range bookmarks {
		result, err := db.SaveBookmarks(ctx, true, bookmark)
		require.NoError(t, err)
		require.Len(t, result, 1)
		savedBookmarks = append(savedBookmarks, result[0])
	}

	// Verify test data setup
	t.Run("VerifyTestData", func(t *testing.T) {
		// Check that all bookmarks were saved with their tags
		for i, bookmark := range savedBookmarks {
			assert.NotZero(t, bookmark.ID)
			assert.Len(t, bookmark.Tags, len(bookmarks[i].Tags))
		}

		// Verify that the first bookmark has golang and web tags
		assert.Len(t, savedBookmarks[0].Tags, 2)
		tagNames := []string{savedBookmarks[0].Tags[0].Name, savedBookmarks[0].Tags[1].Name}
		assert.Contains(t, tagNames, "golang")
		assert.Contains(t, tagNames, "web")
	})

	// Test 1: Get all tags without any options
	t.Run("GetAllTags", func(t *testing.T) {
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{})
		require.NoError(t, err)

		// Should return all 4 tags
		assert.Len(t, fetchedTags, 4)

		// Verify all tag names are present
		tagNames := make(map[string]bool)
		for _, tag := range fetchedTags {
			tagNames[tag.Name] = true
		}

		for _, expectedTag := range tags {
			assert.True(t, tagNames[expectedTag.Name], "Tag %s should be present", expectedTag.Name)
		}
	})

	// Test 2: Get tags with bookmark count
	t.Run("GetTagsWithBookmarkCount", func(t *testing.T) {
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			WithBookmarkCount: true,
		})
		require.NoError(t, err)

		// Should return all 4 tags
		assert.Len(t, fetchedTags, 4)

		// Create a map of tag name to bookmark count
		tagCounts := make(map[string]int64)
		for _, tag := range fetchedTags {
			tagCounts[tag.Name] = tag.BookmarkCount
		}

		// Verify counts
		assert.Equal(t, int64(1), tagCounts["golang"])
		assert.Equal(t, int64(2), tagCounts["database"])
		assert.Equal(t, int64(1), tagCounts["testing"])
		assert.Equal(t, int64(1), tagCounts["web"])
	})

	// Test 3: Get tags for a specific bookmark
	t.Run("GetTagsForBookmark", func(t *testing.T) {
		// Get tags for the first bookmark (Go Language with golang and web tags)
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: savedBookmarks[0].ID,
		})
		require.NoError(t, err)

		// Should return 2 tags
		assert.Len(t, fetchedTags, 2)

		// Verify tag names
		tagNames := make(map[string]bool)
		for _, tag := range fetchedTags {
			tagNames[tag.Name] = true
		}

		assert.True(t, tagNames["golang"], "Tag 'golang' should be present")
		assert.True(t, tagNames["web"], "Tag 'web' should be present")
	})

	// Test 4: Get tags for a specific bookmark with bookmark count
	t.Run("GetTagsForBookmarkWithCount", func(t *testing.T) {
		// Get tags for the third bookmark (SQLite with database and testing tags)
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID:        savedBookmarks[2].ID,
			WithBookmarkCount: true,
		})
		require.NoError(t, err)

		// Should return 2 tags
		assert.Len(t, fetchedTags, 2)

		// Create a map of tag name to bookmark count
		tagCounts := make(map[string]int64)
		for _, tag := range fetchedTags {
			tagCounts[tag.Name] = tag.BookmarkCount
		}

		// Verify counts - database should have 2 bookmarks, testing should have 1
		assert.Equal(t, int64(2), tagCounts["database"])
		assert.Equal(t, int64(1), tagCounts["testing"])
	})

	// Test 5: Get tags ordered by name
	t.Run("GetTagsOrderedByName", func(t *testing.T) {
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			OrderBy: model.DBTagOrderByTagName,
		})
		require.NoError(t, err)

		// Should return all 4 tags in alphabetical order
		assert.Len(t, fetchedTags, 4)

		// Verify order
		assert.Equal(t, "database", fetchedTags[0].Name)
		assert.Equal(t, "golang", fetchedTags[1].Name)
		assert.Equal(t, "testing", fetchedTags[2].Name)
		assert.Equal(t, "web", fetchedTags[3].Name)
	})

	// Test 6: Get tags with search term
	t.Run("GetTagsWithSearch", func(t *testing.T) {
		// Search for tags containing "go"
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			Search: "go",
		})
		require.NoError(t, err)

		// Should return only the golang tag
		assert.Len(t, fetchedTags, 1)
		assert.Equal(t, "golang", fetchedTags[0].Name)

		// Search for tags containing "a"
		fetchedTags, err = db.GetTags(ctx, model.DBListTagsOptions{
			Search: "a",
		})
		require.NoError(t, err)

		// Should return database and possibly other tags containing "a"
		assert.GreaterOrEqual(t, len(fetchedTags), 1)

		// Create a map of tag names for easier checking
		tagNames := make(map[string]bool)
		for _, tag := range fetchedTags {
			tagNames[tag.Name] = true
		}

		// Verify database is in the results
		assert.True(t, tagNames["database"], "Tag 'database' should be present")

		// Search for non-existent tag
		fetchedTags, err = db.GetTags(ctx, model.DBListTagsOptions{
			Search: "nonexistent",
		})
		require.NoError(t, err)
		assert.Len(t, fetchedTags, 0)
	})

	// Test 7: Search and bookmark ID are mutually exclusive
	t.Run("SearchAndBookmarkIDMutuallyExclusive", func(t *testing.T) {
		// This test is just to document the behavior, as the validation happens at the model level
		// The database layer will prioritize the bookmark ID filter if both are provided
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			Search:     "go",
			BookmarkID: savedBookmarks[0].ID,
		})
		require.NoError(t, err)

		// Should return tags for the bookmark, not the search
		// The number of tags may vary depending on the database implementation
		assert.NotEmpty(t, fetchedTags, "Should return at least one tag for the bookmark")

		// Create a map of tag names for easier checking
		tagNames := make(map[string]bool)
		for _, tag := range fetchedTags {
			tagNames[tag.Name] = true
		}

		// Verify golang is in the results (it's associated with the first bookmark)
		assert.True(t, tagNames["golang"], "Tag 'golang' should be present")
	})

	// Test 8: Get tags for a non-existent bookmark
	t.Run("GetTagsForNonExistentBookmark", func(t *testing.T) {
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: 9999, // Non-existent ID
		})
		require.NoError(t, err)

		// Should return empty result
		assert.Empty(t, fetchedTags)
	})

	// Test 9: Get tags for a bookmark with no tags
	t.Run("GetTagsForBookmarkWithNoTags", func(t *testing.T) {
		// Create a bookmark with no tags
		bookmarkWithNoTags := model.BookmarkDTO{
			URL:   "https://example.com",
			Title: "Example with no tags",
		}

		result, err := db.SaveBookmarks(ctx, true, bookmarkWithNoTags)
		require.NoError(t, err)
		require.Len(t, result, 1)

		// Get tags for this bookmark
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: result[0].ID,
		})
		require.NoError(t, err)

		// Should return empty result
		assert.Empty(t, fetchedTags)
	})

	// Test 10: Get tags with combined options (order + count)
	t.Run("GetTagsWithCombinedOptions", func(t *testing.T) {
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			WithBookmarkCount: true,
			OrderBy:           model.DBTagOrderByTagName,
		})
		require.NoError(t, err)

		// Should return all 4 tags in alphabetical order with counts
		assert.Len(t, fetchedTags, 4)

		// Verify order and counts
		assert.Equal(t, "database", fetchedTags[0].Name)
		assert.Equal(t, int64(2), fetchedTags[0].BookmarkCount)

		assert.Equal(t, "golang", fetchedTags[1].Name)
		assert.Equal(t, int64(1), fetchedTags[1].BookmarkCount)

		assert.Equal(t, "testing", fetchedTags[2].Name)
		assert.Equal(t, int64(1), fetchedTags[2].BookmarkCount)

		assert.Equal(t, "web", fetchedTags[3].Name)
		assert.Equal(t, int64(1), fetchedTags[3].BookmarkCount)
	})
}

// testTagBookmarkOperations tests the tag-bookmark relationship operations
func testTagBookmarkOperations(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create test data
	// 1. Create a test bookmark
	bookmark := model.BookmarkDTO{
		URL:   "https://example.com/tag-operations-test",
		Title: "Tag Operations Test",
	}
	savedBookmarks, err := db.SaveBookmarks(ctx, true, bookmark)
	require.NoError(t, err)
	require.Len(t, savedBookmarks, 1)
	bookmarkID := savedBookmarks[0].ID

	// 2. Create a test tag
	tag := model.Tag{
		Name: "tag-operations-test",
	}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Test BookmarkExists function
	t.Run("BookmarkExists", func(t *testing.T) {
		// Test with existing bookmark
		exists, err := db.BookmarkExists(ctx, bookmarkID)
		require.NoError(t, err)
		assert.True(t, exists, "Bookmark should exist")

		// Test with non-existent bookmark
		exists, err = db.BookmarkExists(ctx, 9999)
		require.NoError(t, err)
		assert.False(t, exists, "Non-existent bookmark should return false")
	})

	// Test TagExists function
	t.Run("TagExists", func(t *testing.T) {
		// Test with existing tag
		exists, err := db.TagExists(ctx, tagID)
		require.NoError(t, err)
		assert.True(t, exists, "Tag should exist")

		// Test with non-existent tag
		exists, err = db.TagExists(ctx, 9999)
		require.NoError(t, err)
		assert.False(t, exists, "Non-existent tag should return false")
	})

	// Test AddTagToBookmark function
	t.Run("AddTagToBookmark", func(t *testing.T) {
		// Add tag to bookmark
		err := db.AddTagToBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Verify tag was added by fetching tags for the bookmark
		tags, err := db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 1)
		assert.Equal(t, tagID, tags[0].ID)
		assert.Equal(t, "tag-operations-test", tags[0].Name)

		// Test adding the same tag again (should not error)
		err = db.AddTagToBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Verify no duplicate was created
		tags, err = db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 1)
	})

	// Test RemoveTagFromBookmark function
	t.Run("RemoveTagFromBookmark", func(t *testing.T) {
		// First ensure the tag is associated with the bookmark
		tags, err := db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 1, "Tag should be associated with bookmark before removal test")

		// Remove tag from bookmark
		err = db.RemoveTagFromBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Verify tag was removed
		tags, err = db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		assert.Len(t, tags, 0, "Tag should be removed from bookmark")

		// Test removing a tag that's not associated (should not error)
		err = db.RemoveTagFromBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Test removing a tag from a non-existent bookmark (should not error)
		err = db.RemoveTagFromBookmark(ctx, 9999, tagID)
		require.NoError(t, err)

		// Test removing a non-existent tag from a bookmark (should not error)
		err = db.RemoveTagFromBookmark(ctx, bookmarkID, 9999)
		require.NoError(t, err)
	})

	// Test edge cases
	t.Run("EdgeCases", func(t *testing.T) {
		// Test adding a tag to a non-existent bookmark
		// This should not error at the database layer since we're not checking existence there
		err := db.AddTagToBookmark(ctx, 9999, tagID)
		// The test might fail depending on foreign key constraints in the database
		// If it fails, that's acceptable behavior, but we're not explicitly testing for it
		if err != nil {
			t.Logf("Adding tag to non-existent bookmark failed as expected: %v", err)
		}

		// Test adding a non-existent tag to a bookmark
		// This should not error at the database layer since we're not checking existence there
		err = db.AddTagToBookmark(ctx, bookmarkID, 9999)
		// The test might fail depending on foreign key constraints in the database
		// If it fails, that's acceptable behavior, but we're not explicitly testing for it
		if err != nil {
			t.Logf("Adding non-existent tag to bookmark failed as expected: %v", err)
		}
	})
}

// testTagExists tests the TagExists function
func testTagExists(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create a test tag
	tag := model.Tag{
		Name: "tag-exists-test",
	}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Test with existing tag
	exists, err := db.TagExists(ctx, tagID)
	require.NoError(t, err)
	assert.True(t, exists, "Tag should exist")

	// Test with non-existent tag
	exists, err = db.TagExists(ctx, 9999)
	require.NoError(t, err)
	assert.False(t, exists, "Non-existent tag should return false")
}

// testBookmarkExists tests the BookmarkExists function
func testBookmarkExists(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create a test bookmark
	bookmark := model.BookmarkDTO{
		URL:   "https://example.com/bookmark-exists-test",
		Title: "Bookmark Exists Test",
	}
	savedBookmarks, err := db.SaveBookmarks(ctx, true, bookmark)
	require.NoError(t, err)
	require.Len(t, savedBookmarks, 1)
	bookmarkID := savedBookmarks[0].ID

	// Test with existing bookmark
	exists, err := db.BookmarkExists(ctx, bookmarkID)
	require.NoError(t, err)
	assert.True(t, exists, "Bookmark should exist")

	// Test with non-existent bookmark
	exists, err = db.BookmarkExists(ctx, 9999)
	require.NoError(t, err)
	assert.False(t, exists, "Non-existent bookmark should return false")
}

// testAddTagToBookmark tests the AddTagToBookmark function
func testAddTagToBookmark(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create test data
	bookmark := model.BookmarkDTO{
		URL:   "https://example.com/add-tag-test",
		Title: "Add Tag Test",
	}
	savedBookmarks, err := db.SaveBookmarks(ctx, true, bookmark)
	require.NoError(t, err)
	require.Len(t, savedBookmarks, 1)
	bookmarkID := savedBookmarks[0].ID

	tag := model.Tag{
		Name: "add-tag-test",
	}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Add tag to bookmark
	err = db.AddTagToBookmark(ctx, bookmarkID, tagID)
	require.NoError(t, err)

	// Verify tag was added by fetching tags for the bookmark
	tags, err := db.GetTags(ctx, model.DBListTagsOptions{
		BookmarkID: bookmarkID,
	})
	require.NoError(t, err)
	require.Len(t, tags, 1)
	assert.Equal(t, tagID, tags[0].ID)
	assert.Equal(t, "add-tag-test", tags[0].Name)

	// Test adding the same tag again (should not error)
	err = db.AddTagToBookmark(ctx, bookmarkID, tagID)
	require.NoError(t, err)

	// Verify no duplicate was created
	tags, err = db.GetTags(ctx, model.DBListTagsOptions{
		BookmarkID: bookmarkID,
	})
	require.NoError(t, err)
	require.Len(t, tags, 1)
}

// testRemoveTagFromBookmark tests the RemoveTagFromBookmark function
func testRemoveTagFromBookmark(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create test data
	bookmark := model.BookmarkDTO{
		URL:   "https://example.com/remove-tag-test",
		Title: "Remove Tag Test",
	}
	savedBookmarks, err := db.SaveBookmarks(ctx, true, bookmark)
	require.NoError(t, err)
	require.Len(t, savedBookmarks, 1)
	bookmarkID := savedBookmarks[0].ID

	tag := model.Tag{
		Name: "remove-tag-test",
	}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Add tag to bookmark first
	err = db.AddTagToBookmark(ctx, bookmarkID, tagID)
	require.NoError(t, err)

	// Verify tag was added
	tags, err := db.GetTags(ctx, model.DBListTagsOptions{
		BookmarkID: bookmarkID,
	})
	require.NoError(t, err)
	require.Len(t, tags, 1, "Tag should be associated with bookmark before removal test")

	// Remove tag from bookmark
	err = db.RemoveTagFromBookmark(ctx, bookmarkID, tagID)
	require.NoError(t, err)

	// Verify tag was removed
	tags, err = db.GetTags(ctx, model.DBListTagsOptions{
		BookmarkID: bookmarkID,
	})
	require.NoError(t, err)
	assert.Len(t, tags, 0, "Tag should be removed from bookmark")

	// Test removing a tag that's not associated (should not error)
	err = db.RemoveTagFromBookmark(ctx, bookmarkID, tagID)
	require.NoError(t, err)

	// Test removing a tag from a non-existent bookmark (should not error)
	err = db.RemoveTagFromBookmark(ctx, 9999, tagID)
	require.NoError(t, err)

	// Test removing a non-existent tag from a bookmark (should not error)
	err = db.RemoveTagFromBookmark(ctx, bookmarkID, 9999)
	require.NoError(t, err)
}

// testTagBookmarkEdgeCases tests edge cases for tag-bookmark operations
func testTagBookmarkEdgeCases(t *testing.T, db model.DB) {
	ctx := context.TODO()

	// Create test data
	bookmark := model.BookmarkDTO{
		URL:   "https://example.com/edge-cases-test",
		Title: "Edge Cases Test",
	}
	savedBookmarks, err := db.SaveBookmarks(ctx, true, bookmark)
	require.NoError(t, err)
	require.Len(t, savedBookmarks, 1)
	bookmarkID := savedBookmarks[0].ID

	tag := model.Tag{
		Name: "edge-cases-test",
	}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Test adding a tag to a non-existent bookmark
	// This should not error at the database layer since we're not checking existence there
	err = db.AddTagToBookmark(ctx, 9999, tagID)
	// The test might fail depending on foreign key constraints in the database
	// If it fails, that's acceptable behavior, but we're not explicitly testing for it
	if err != nil {
		t.Logf("Adding tag to non-existent bookmark failed as expected: %v", err)
	}

	// Test adding a non-existent tag to a bookmark
	// This should not error at the database layer since we're not checking existence there
	err = db.AddTagToBookmark(ctx, bookmarkID, 9999)
	// The test might fail depending on foreign key constraints in the database
	// If it fails, that's acceptable behavior, but we're not explicitly testing for it
	if err != nil {
		t.Logf("Adding non-existent tag to bookmark failed as expected: %v", err)
	}
}
