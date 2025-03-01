package domains_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the tagsDomain implementation
func TestTagsDomain(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// Setup using the test configuration and dependencies
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	tagsDomain := deps.Domains().Tags()
	db := deps.Database()

	// Test ListTags
	t.Run("ListTags", func(t *testing.T) {
		// Create some test tags first
		testTags := []model.Tag{
			{Name: "tag1"},
			{Name: "tag2"},
		}
		createdTags, err := db.CreateTags(ctx, testTags...)
		require.NoError(t, err)
		require.Len(t, createdTags, 2)

		// List the tags
		tags, err := tagsDomain.ListTags(ctx)
		require.NoError(t, err)
		require.Len(t, tags, 2)

		// Verify the tags
		assert.Equal(t, "tag1", tags[0].Name)
		assert.Equal(t, "tag2", tags[1].Name)
	})

	// Test CreateTag
	t.Run("CreateTag", func(t *testing.T) {
		// Create a new tag
		tagDTO := model.TagDTO{
			Tag: model.Tag{
				Name: "new-tag",
			},
		}

		createdTag, err := tagsDomain.CreateTag(ctx, tagDTO)
		require.NoError(t, err)
		assert.Equal(t, "new-tag", createdTag.Name)
		assert.Greater(t, createdTag.ID, 0, "The created tag should have a valid ID")

		// Verify the tag was created in the database
		allTags, err := db.GetTags(ctx)
		require.NoError(t, err)
		require.Len(t, allTags, 3) // 2 from previous test + 1 new

		// Find the created tag in the list
		var found bool
		for _, tag := range allTags {
			if tag.Name == "new-tag" {
				found = true
				assert.Greater(t, tag.ID, 0, "The tag in the database should have a valid ID")
				break
			}
		}
		assert.True(t, found, "The created tag should be found in the database")
	})

	// Test GetTag - Success
	t.Run("GetTag_Success", func(t *testing.T) {
		// Get all tags to find an ID
		allTags, err := db.GetTags(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, allTags)

		tagID := allTags[0].ID

		// Get the tag by ID
		tag, err := tagsDomain.GetTag(ctx, tagID)
		require.NoError(t, err)
		assert.Equal(t, tagID, tag.ID)
		assert.Equal(t, allTags[0].Name, tag.Name)
	})

	// Test GetTag - Not Found
	t.Run("GetTag_NotFound", func(t *testing.T) {
		// Try to get a non-existent tag
		_, err := tagsDomain.GetTag(ctx, 9999)
		require.Error(t, err)
		assert.Equal(t, model.ErrNotFound, err)
	})

	// Test UpdateTag
	t.Run("UpdateTag", func(t *testing.T) {
		// Get all tags to find an ID
		allTags, err := db.GetTags(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, allTags)

		tagID := allTags[0].ID

		// Update the tag
		tagDTO := model.TagDTO{
			Tag: model.Tag{
				ID:   tagID,
				Name: "updated-tag",
			},
		}

		updatedTag, err := tagsDomain.UpdateTag(ctx, tagDTO)
		require.NoError(t, err)
		assert.Equal(t, tagID, updatedTag.ID)
		assert.Equal(t, "updated-tag", updatedTag.Name)

		// Verify the tag was updated in the database
		dbTag, exists, err := db.GetTag(ctx, tagID)
		require.NoError(t, err)
		require.True(t, exists)
		assert.Equal(t, "updated-tag", dbTag.Name)
	})

	// Test DeleteTag
	t.Run("DeleteTag", func(t *testing.T) {
		// Get all tags to find an ID
		allTags, err := db.GetTags(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, allTags)

		tagID := allTags[1].ID

		// Delete the tag
		err = tagsDomain.DeleteTag(ctx, tagID)
		require.NoError(t, err)

		// Verify the tag was deleted from the database
		_, exists, err := db.GetTag(ctx, tagID)
		require.NoError(t, err)
		require.False(t, exists)
	})

	// Test DeleteTag - Not Found
	t.Run("DeleteTag_NotFound", func(t *testing.T) {
		// Try to delete a non-existent tag
		err := tagsDomain.DeleteTag(ctx, 9999)
		require.Error(t, err)
		// Use errors.Is to check if the error is or wraps model.ErrNotFound
		assert.True(t, errors.Is(err, model.ErrNotFound) || strings.Contains(err.Error(), "not found"),
			"Expected error to be or contain 'not found', got: %v", err)
	})
}
