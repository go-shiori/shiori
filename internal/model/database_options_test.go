package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBGetBookmarksOptions(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		options := DBGetBookmarksOptions{}

		assert.Nil(t, options.IDs)
		assert.Nil(t, options.Tags)
		assert.Nil(t, options.ExcludedTags)
		assert.Equal(t, "", options.Keyword)
		assert.False(t, options.WithContent)
		assert.Equal(t, DefaultOrder, options.OrderMethod)
		assert.Equal(t, 0, options.Limit)
		assert.Equal(t, 0, options.Offset)
	})
}

func TestDBListAccountsOptions(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		options := DBListAccountsOptions{}

		assert.Equal(t, "", options.Keyword)
		assert.Equal(t, "", options.Username)
		assert.False(t, options.Owner)
		assert.False(t, options.WithPassword)
	})
}

func TestDBListTagsOptions(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		options := DBListTagsOptions{}

		assert.Equal(t, 0, options.BookmarkID)
		assert.False(t, options.WithBookmarkCount)
		assert.Equal(t, DBTagOrderBy(""), options.OrderBy)
		assert.Equal(t, "", options.Search)
	})
}
