package model

import (
	"path/filepath"
	"strconv"
)

// Bookmark is the database representation of a bookmark
type Bookmark struct {
	ID         int    `db:"id"         json:"id"`
	URL        string `db:"url"        json:"url"`
	Title      string `db:"title"      json:"title"`
	Excerpt    string `db:"excerpt"    json:"excerpt"`
	Author     string `db:"author"     json:"author"`
	Public     int    `db:"public"     json:"public"`
	CreatedAt  string `db:"created_at" json:"createdAt"`
	ModifiedAt string `db:"modified_at" json:"modifiedAt"`
	Content    string `db:"content"    json:"-"`
	HTML       string `db:"html"       json:"html,omitempty"`
	HasContent bool   `db:"has_content" json:"hasContent"`
}

// BookmarkDTO is the data transfer object for bookmarks sent to/from clients
// It embeds the Bookmark struct and adds additional fields for API responses
type BookmarkDTO struct {
	Bookmark      `db:",inline"`
	Tags          []TagDTO `json:"tags"`
	HasArchive    bool     `json:"hasArchive"`
	HasEbook      bool     `json:"hasEbook"`
	CreateArchive bool     `json:"create_archive"`
	CreateEbook   bool     `json:"create_ebook"`
	ImageURL      string   `json:"imageURL"`
}

// ToBookmark extracts the embedded Bookmark from BookmarkDTO
func (dto *BookmarkDTO) ToBookmark() Bookmark {
	return dto.Bookmark
}

// ToDTO converts a Bookmark to a BookmarkDTO
func (b *Bookmark) ToDTO() BookmarkDTO {
	return BookmarkDTO{
		Bookmark: *b,
		Tags:     []TagDTO{},
	}
}

// GetTumnbailPath returns the relative path to the thumbnail of a bookmark in the filesystem
func GetThumbnailPath(bookmark *BookmarkDTO) string {
	return filepath.Join("thumb", strconv.Itoa(bookmark.ID))
}

// GetEbookPath returns the relative path to the ebook of a bookmark in the filesystem
func GetEbookPath(bookmark *BookmarkDTO) string {
	return filepath.Join("ebook", strconv.Itoa(bookmark.ID)+".epub")
}

// GetArchivePath returns the relative path to the archive of a bookmark in the filesystem
func GetArchivePath(bookmark *BookmarkDTO) string {
	return filepath.Join("archive", strconv.Itoa(bookmark.ID))
}
