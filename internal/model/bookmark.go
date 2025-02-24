package model

import (
	"path/filepath"
	"strconv"
)

// BookmarkDTO is the bookmark object representation in database and the data transfer object
// at the same time, pending a refactor to two separate object to represent each role.
type BookmarkDTO struct {
	ID            int      `db:"id"            json:"id"`
	URL           string   `db:"url"           json:"url"`
	Title         string   `db:"title"         json:"title"`
	Excerpt       string   `db:"excerpt"       json:"excerpt"`
	Author        string   `db:"author"        json:"author"`
	Public        int      `db:"public"        json:"public"`
	CreatedAt     string   `db:"created_at"    json:"createdAt"`
	ModifiedAt    string   `db:"modified_at"   json:"modifiedAt"`
	Content       string   `db:"content"       json:"-"`
	HTML          string   `db:"html"          json:"html,omitempty"`
	ImageURL      string   `db:"image_url"     json:"imageURL"`
	HasContent    bool     `db:"has_content"   json:"hasContent"`
	Tags          []TagDTO `json:"tags"`
	HasArchive    bool     `json:"hasArchive"`
	HasEbook      bool     `json:"hasEbook"`
	CreateArchive bool     `json:"create_archive"` // TODO: migrate outside the DTO
	CreateEbook   bool     `json:"create_ebook"`   // TODO: migrate outside the DTO
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
