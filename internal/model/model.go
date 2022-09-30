package model

import "gorm.io/gorm"

// Tag is the tag for a bookmark.
type Tag struct {
	gorm.Model
	ID         int    `gorm:"id,primaryKey,index"          json:"id"`
	Name       string `gorm:"name"        json:"name"`
	NBookmarks int    `gorm:"n_bookmarks" json:"nBookmarks,omitempty"`
	Deleted    bool   `gorm:"-" json:"-"`
}

func (Tag) TableName() string {
	return "tag"
}

// Bookmark is the record for an URL.
type Bookmark struct {
	gorm.Model
	ID            int    `gorm:"id,primaryKey,index"            json:"id"`
	URL           string `gorm:"url"           json:"url"`
	Title         string `gorm:"title"         json:"title"`
	Excerpt       string `gorm:"excerpt"       json:"excerpt"`
	Author        string `gorm:"author"        json:"author"`
	Public        int    `gorm:"public"        json:"public"`
	Modified      string `gorm:"modified"      json:"modified"`
	Content       string `gorm:"content"       json:"-"`
	HTML          string `gorm:"html"          json:"html,omitempty"`
	ImageURL      string `gorm:"image_url"     json:"imageURL"`
	HasContent    bool   `gorm:"has_content"   json:"hasContent"`
	HasArchive    bool   `gorm:"-" json:"hasArchive"`
	Tags          []Tag  `gorm:"many2many:bookmark_tag;" json:"tags"`
	CreateArchive bool   `gorm:"-" json:"createArchive"`
}

func (Bookmark) TableName() string {
	return "bookmark"
}

// Account is person that allowed to access web interface.
type Account struct {
	gorm.Model
	ID       int    `gorm:"id,primaryKey,index" json:"id"`
	Username string `gorm:"username" json:"username"`
	Password string `gorm:"password" json:"password,omitempty"`
	Owner    bool   `gorm:"owner"    json:"owner"`
}

func (Account) TableName() string {
	return "account"
}
