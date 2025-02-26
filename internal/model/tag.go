package model

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
