package model

// Tag is the database representation of a tag object
type Tag struct {
	ID            int    `db:"id"`
	Name          string `db:"name"`
	Deleted       bool   `db:"-" `
	BookmarkCount int    `db:"bookmark_count"`
}

func (t *Tag) ToDTO() TagDTO {
	return TagDTO{
		ID:            t.ID,
		Name:          t.Name,
		BookmarkCount: t.BookmarkCount,
	}
}

// TagDTO is the data transfer object representation of a tag object
type TagDTO struct {
	ID            int
	Name          string
	BookmarkCount int
}

func (tdto *TagDTO) ToTag() Tag {
	return Tag{
		ID:   tdto.ID,
		Name: tdto.Name,
	}
}
