package model

// Tag is the database representation of a tag object
type Tag struct {
	ID      int    `db:"id"          json:"id"`
	Name    string `db:"name"        json:"name"`
	Deleted bool   `db:"-"           json:"-"`
}

func (t *Tag) ToDTO() TagDTO {
	return TagDTO{
		ID:   t.ID,
		Name: t.Name,
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
