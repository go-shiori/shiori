package model

import (
	"path/filepath"
	"strconv"
)

// Bookmark is the database representation of a bookmark
type Bookmark struct {
	ID         int    `db:"id"`
	URL        string `db:"url"`
	Title      string `db:"title"`
	Excerpt    string `db:"excerpt"`
	Author     string `db:"author"`
	Public     int    `db:"public"`
	CreatedAt  string `db:"created_at"`
	ModifiedAt string `db:"modified_at"`
	Content    string `db:"content" json:"-"`
	HTML       string `db:"html"    json:",omitempty"`
	HasContent bool   `db:"has_content"`
}

// BookmarkDTO is the data transfer object for bookmarks sent to/from clients
// It embeds the Bookmark struct and adds additional fields for API responses
type BookmarkDTO struct {
	Bookmark      `db:",inline"`
	Tags          []TagDTO
	ImageURL      string
	HasArchive    bool
	HasEbook      bool
	HasThumbnail  bool
	CreateArchive bool
	CreateEbook   bool
}

// PopulateImageURL sets the ImageURL field based on HasThumbnail
func (dto *BookmarkDTO) PopulateImageURL() {
	if dto.HasThumbnail {
		dto.ImageURL = GetThumbnailURL(dto)
	}
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

// GetThumbnailURL returns the URL path to the thumbnail of a bookmark
func GetThumbnailURL(bookmark *BookmarkDTO) string {
	return "/bookmark/" + strconv.Itoa(bookmark.ID) + "/thumb"
}

// GetEbookURL returns the URL path to the ebook of a bookmark
func GetEbookURL(bookmark *BookmarkDTO) string {
	return "/bookmark/" + strconv.Itoa(bookmark.ID) + "/ebook"
}

// GetArchiveURL returns the URL path to the archive of a bookmark
func GetArchiveURL(bookmark *BookmarkDTO) string {
	return "/bookmark/" + strconv.Itoa(bookmark.ID) + "/archive"
}
