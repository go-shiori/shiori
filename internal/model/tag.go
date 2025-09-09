package model

import (
	"errors"
)

// BookmarkTag is the relationship between a bookmark and a tag.
type BookmarkTag struct {
	BookmarkID int `db:"bookmark_id"`
	TagID      int `db:"tag_id"`
}

// Tag is the tag for a bookmark.
type Tag struct {
	ID   int    `db:"id"          json:"id"`
	Name string `db:"name"        json:"name"`
}

// TagDTO represents a tag in the application
type TagDTO struct {
	Tag
	BookmarkCount int64 `db:"bookmark_count" json:"bookmark_count"` // Number of bookmarks with this tag
	Deleted       bool  `json:"deleted"`                            // Marks when a tag is deleted from a bookmark
}

func (t *Tag) ToDTO() TagDTO {
	return TagDTO{
		Tag: Tag{
			ID:   t.ID,
			Name: t.Name,
		},
	}
}

func (t *TagDTO) ToTag() Tag {
	return Tag{
		ID:   t.ID,
		Name: t.Name,
	}
}

// ListTagsOptions is options for fetching tags from database.
type ListTagsOptions struct {
	BookmarkID        int
	WithBookmarkCount bool
	OrderBy           DBTagOrderBy
	Search            string
}

// IsValid validates the ListTagsOptions.
// Returns an error if the options are invalid, nil otherwise.
// Currently, it checks that Search and BookmarkID are not used together.
func (o ListTagsOptions) IsValid() error {
	if o.Search != "" && o.BookmarkID > 0 {
		return errors.New("search and bookmark ID filtering cannot be used together")
	}
	return nil
}
