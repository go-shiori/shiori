package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTag_ToDTO(t *testing.T) {
	t.Run("Complete conversion", func(t *testing.T) {
		tag := Tag{
			ID:   42,
			Name: "test-tag",
		}

		dto := tag.ToDTO()

		assert.Equal(t, 42, dto.ID)
		assert.Equal(t, "test-tag", dto.Name)
		assert.Equal(t, int64(0), dto.BookmarkCount) // Default value
		assert.False(t, dto.Deleted)                 // Default value
	})
}

func TestTagDTO_ToTag(t *testing.T) {
	t.Run("Complete conversion", func(t *testing.T) {
		dto := TagDTO{
			Tag: Tag{
				ID:   123,
				Name: "conversion-test",
			},
			BookmarkCount: 42,
			Deleted:       true,
		}

		tag := dto.ToTag()

		assert.Equal(t, 123, tag.ID)
		assert.Equal(t, "conversion-test", tag.Name)
	})
}

func TestTagDTO_DefaultValues(t *testing.T) {
	t.Run("Default values are correct", func(t *testing.T) {
		var dto TagDTO

		assert.Equal(t, 0, dto.ID)
		assert.Equal(t, "", dto.Name)
		assert.Equal(t, int64(0), dto.BookmarkCount)
		assert.False(t, dto.Deleted)
	})
}

func TestBookmarkTag_Structure(t *testing.T) {
	t.Run("BookmarkTag has correct fields", func(t *testing.T) {
		bt := BookmarkTag{
			BookmarkID: 123,
			TagID:      456,
		}

		assert.Equal(t, 123, bt.BookmarkID)
		assert.Equal(t, 456, bt.TagID)
	})

	t.Run("BookmarkTag zero values", func(t *testing.T) {
		var bt BookmarkTag

		assert.Equal(t, 0, bt.BookmarkID)
		assert.Equal(t, 0, bt.TagID)
	})
}

func TestListTagsOptions_IsValid(t *testing.T) {
	t.Run("Valid empty options", func(t *testing.T) {
		opts := ListTagsOptions{}

		err := opts.IsValid()

		assert.NoError(t, err)
	})

	t.Run("Valid with search only", func(t *testing.T) {
		opts := ListTagsOptions{
			Search: "search-term",
		}

		err := opts.IsValid()

		assert.NoError(t, err)
	})

	t.Run("Valid with bookmark ID only", func(t *testing.T) {
		opts := ListTagsOptions{
			BookmarkID: 123,
		}

		err := opts.IsValid()

		assert.NoError(t, err)
	})

	t.Run("Invalid with both search and bookmark ID", func(t *testing.T) {
		opts := ListTagsOptions{
			Search:     "search-term",
			BookmarkID: 123,
		}

		err := opts.IsValid()

		assert.Error(t, err)
		assert.Equal(t, "search and bookmark ID filtering cannot be used together", err.Error())
	})

	t.Run("Valid with other fields", func(t *testing.T) {
		opts := ListTagsOptions{
			WithBookmarkCount: true,
			OrderBy:           DBTagOrderByTagName,
		}

		err := opts.IsValid()

		assert.NoError(t, err)
	})

	t.Run("Zero bookmark ID with search should be valid", func(t *testing.T) {
		opts := ListTagsOptions{
			Search:     "search-term",
			BookmarkID: 0, // Zero is same as not set
		}

		err := opts.IsValid()

		assert.NoError(t, err)
	})
}
