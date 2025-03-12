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

	// Test 6: Get tags for a non-existent bookmark
	t.Run("GetTagsForNonExistentBookmark", func(t *testing.T) {
		fetchedTags, err := db.GetTags(ctx, model.DBListTagsOptions{
			BookmarkID: 9999, // Non-existent ID
		})
		require.NoError(t, err)

		// Should return empty result
		assert.Empty(t, fetchedTags)
	})

	// Test 7: Get tags for a bookmark with no tags
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

	// Test 8: Get tags with combined options (order + count)
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
