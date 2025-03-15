package domains_test

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookmarkTagOperations(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// Setup using the test configuration and dependencies
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	bookmarksDomain := deps.Domains().Bookmarks()
	tagsDomain := deps.Domains().Tags()
	db := deps.Database()

	// Create a test bookmark
	bookmark := model.BookmarkDTO{
		URL:   "https://example.com/bookmark-tags-test",
		Title: "Bookmark Tags Test",
	}
	savedBookmarks, err := db.SaveBookmarks(ctx, true, bookmark)
	require.NoError(t, err)
	require.Len(t, savedBookmarks, 1)
	bookmarkID := savedBookmarks[0].ID

	// Create a test tag
	tagDTO := model.TagDTO{
		Tag: model.Tag{
			Name: "test-tag",
		},
	}
	createdTag, err := tagsDomain.CreateTag(ctx, tagDTO)
	require.NoError(t, err)
	tagID := createdTag.ID

	// Test BookmarkExists
	t.Run("BookmarkExists", func(t *testing.T) {
		// Test with existing bookmark
		exists, err := bookmarksDomain.BookmarkExists(ctx, bookmarkID)
		require.NoError(t, err)
		assert.True(t, exists, "Bookmark should exist")

		// Test with non-existent bookmark
		exists, err = bookmarksDomain.BookmarkExists(ctx, 9999)
		require.NoError(t, err)
		assert.False(t, exists, "Non-existent bookmark should not exist")
	})

	// Test TagExists
	t.Run("TagExists", func(t *testing.T) {
		// Test with existing tag
		exists, err := tagsDomain.TagExists(ctx, tagID)
		require.NoError(t, err)
		assert.True(t, exists, "Tag should exist")

		// Test with non-existent tag
		exists, err = tagsDomain.TagExists(ctx, 9999)
		require.NoError(t, err)
		assert.False(t, exists, "Non-existent tag should not exist")
	})

	// Test AddTagToBookmark
	t.Run("AddTagToBookmark", func(t *testing.T) {
		// Add tag to bookmark
		err := bookmarksDomain.AddTagToBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Verify tag was added by listing tags for the bookmark
		tags, err := tagsDomain.ListTags(ctx, model.ListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 1, "Should have exactly one tag")
		assert.Equal(t, tagID, tags[0].ID, "Tag ID should match")
		assert.Equal(t, "test-tag", tags[0].Name, "Tag name should match")

		// Test adding the same tag again (should not error)
		err = bookmarksDomain.AddTagToBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err, "Adding the same tag again should not error")

		// Test adding tag to non-existent bookmark
		err = bookmarksDomain.AddTagToBookmark(ctx, 9999, tagID)
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrBookmarkNotFound, "Should return bookmark not found error")

		// Test adding non-existent tag to bookmark
		err = bookmarksDomain.AddTagToBookmark(ctx, bookmarkID, 9999)
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrTagNotFound, "Should return tag not found error")
	})

	// Test RemoveTagFromBookmark
	t.Run("RemoveTagFromBookmark", func(t *testing.T) {
		// Remove tag from bookmark
		err := bookmarksDomain.RemoveTagFromBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Verify tag was removed by listing tags for the bookmark
		tags, err := tagsDomain.ListTags(ctx, model.ListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 0, "Should have no tags after removal")

		// Test removing a tag that's not associated with the bookmark (should not error)
		err = bookmarksDomain.RemoveTagFromBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err, "Removing a tag that's not associated should not error")

		// Test removing tag from non-existent bookmark
		err = bookmarksDomain.RemoveTagFromBookmark(ctx, 9999, tagID)
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrBookmarkNotFound, "Should return bookmark not found error")

		// Test removing non-existent tag from bookmark
		err = bookmarksDomain.RemoveTagFromBookmark(ctx, bookmarkID, 9999)
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrTagNotFound, "Should return tag not found error")
	})
}
