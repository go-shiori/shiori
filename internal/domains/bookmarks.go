package domains

import (
	"context"
	"fmt"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
)

type BookmarksDomain struct {
	deps model.Dependencies
}

func (d *BookmarksDomain) HasEbook(b *model.BookmarkDTO) bool {
	ebookPath := model.GetEbookPath(b)
	return d.deps.Domains().Storage().FileExists(ebookPath)
}

func (d *BookmarksDomain) HasArchive(b *model.BookmarkDTO) bool {
	archivePath := model.GetArchivePath(b)
	return d.deps.Domains().Storage().FileExists(archivePath)
}

func (d *BookmarksDomain) HasThumbnail(b *model.BookmarkDTO) bool {
	thumbnailPath := model.GetThumbnailPath(b)
	return d.deps.Domains().Storage().FileExists(thumbnailPath)
}

func (d *BookmarksDomain) GetBookmark(ctx context.Context, id model.DBID) (*model.BookmarkDTO, error) {
	bookmark, exists, err := d.deps.Database().GetBookmark(ctx, int(id), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	if !exists {
		return nil, model.ErrBookmarkNotFound
	}

	// Check if it has ebook and archive.
	bookmark.HasEbook = d.HasEbook(&bookmark)
	bookmark.HasArchive = d.HasArchive(&bookmark)

	return &bookmark, nil
}

func (d *BookmarksDomain) GetBookmarks(ctx context.Context, ids []int) ([]model.BookmarkDTO, error) {
	var bookmarks []model.BookmarkDTO
	for _, id := range ids {
		bookmark, exists, err := d.deps.Database().GetBookmark(ctx, id, "")
		if err != nil {
			return nil, fmt.Errorf("failed to get bookmark %d: %w", id, err)
		}
		if !exists {
			continue
		}

		// Check if it has ebook and archive
		bookmark.HasEbook = d.HasEbook(&bookmark)
		bookmark.HasArchive = d.HasArchive(&bookmark)
		bookmarks = append(bookmarks, bookmark)
	}
	return bookmarks, nil
}

func (d *BookmarksDomain) UpdateBookmarkCache(ctx context.Context, bookmark model.BookmarkDTO, keepMetadata bool, skipExist bool) (*model.BookmarkDTO, error) {
	// Download data from internet
	content, contentType, err := core.DownloadBookmark(bookmark.Bookmark.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download bookmark: %w", err)
	}
	defer content.Close()

	// Check if we should skip existing ebook
	if skipExist && bookmark.CreateEbook {
		ebookPath := model.GetEbookPath(&bookmark)
		if d.deps.Domains().Storage().FileExists(ebookPath) {
			bookmark.CreateEbook = false
			bookmark.HasEbook = true
		}
	}

	// Process the bookmark
	request := core.ProcessRequest{
		DataDir:     d.deps.Config().Storage.DataDir,
		Bookmark:    bookmark,
		Content:     content,
		ContentType: contentType,
		KeepTitle:   keepMetadata,
		KeepExcerpt: keepMetadata,
	}

	processedBookmark, _, err := core.ProcessBookmark(d.deps, request)
	if err != nil {
		return nil, fmt.Errorf("failed to process bookmark: %w", err)
	}

	return &processedBookmark, nil
}

// BulkUpdateBookmarkTags updates tags for multiple bookmarks using tag IDs
func (d *BookmarksDomain) BulkUpdateBookmarkTags(ctx context.Context, bookmarkIDs []int, tagIDs []int) error {
	if len(bookmarkIDs) == 0 {
		return nil
	}

	// Call the database method directly
	err := d.deps.Database().BulkUpdateBookmarkTags(ctx, bookmarkIDs, tagIDs)
	if err != nil {
		return fmt.Errorf("failed to update bookmark tags: %w", err)
	}

	return nil
}

// AddTagToBookmark adds a tag to a bookmark
func (d *BookmarksDomain) AddTagToBookmark(ctx context.Context, bookmarkID int, tagID int) error {
	// Check if bookmark exists
	exists, err := d.BookmarkExists(ctx, bookmarkID)
	if err != nil {
		return err
	}
	if !exists {
		return model.ErrBookmarkNotFound
	}

	// Check if tag exists
	exists, err = d.deps.Domains().Tags().TagExists(ctx, tagID)
	if err != nil {
		return err
	}
	if !exists {
		return model.ErrTagNotFound
	}

	// Add tag to bookmark
	return d.deps.Database().AddTagToBookmark(ctx, bookmarkID, tagID)
}

// RemoveTagFromBookmark removes a tag from a bookmark
func (d *BookmarksDomain) RemoveTagFromBookmark(ctx context.Context, bookmarkID int, tagID int) error {
	// Check if bookmark exists
	exists, err := d.BookmarkExists(ctx, bookmarkID)
	if err != nil {
		return err
	}
	if !exists {
		return model.ErrBookmarkNotFound
	}

	// Check if tag exists
	exists, err = d.deps.Domains().Tags().TagExists(ctx, tagID)
	if err != nil {
		return err
	}
	if !exists {
		return model.ErrTagNotFound
	}

	// Remove tag from bookmark
	return d.deps.Database().RemoveTagFromBookmark(ctx, bookmarkID, tagID)
}

// BookmarkExists checks if a bookmark with the given ID exists
func (d *BookmarksDomain) BookmarkExists(ctx context.Context, id int) (bool, error) {
	return d.deps.Database().BookmarkExists(ctx, id)
}

// CreateBookmark creates a new bookmark
func (d *BookmarksDomain) CreateBookmark(ctx context.Context, bookmark model.Bookmark) (*model.BookmarkDTO, error) {
	// Convert to DTO for database operations
	dto := bookmark.ToDTO()
	
	// Save bookmark to database
	savedBookmarks, err := d.deps.Database().SaveBookmarks(ctx, true, dto)
	if err != nil {
		return nil, fmt.Errorf("failed to save bookmark: %w", err)
	}

	if len(savedBookmarks) == 0 {
		return nil, fmt.Errorf("no bookmark was saved")
	}

	savedBookmark := savedBookmarks[0]

	// Set additional properties
	savedBookmark.HasEbook = d.HasEbook(&savedBookmark)
	savedBookmark.HasArchive = d.HasArchive(&savedBookmark)
	
	return &savedBookmark, nil
}

// UpdateBookmark updates an existing bookmark
func (d *BookmarksDomain) UpdateBookmark(ctx context.Context, bookmark model.Bookmark) (*model.BookmarkDTO, error) {
	// Check if bookmark exists first
	exists, err := d.BookmarkExists(ctx, bookmark.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check bookmark existence: %w", err)
	}
	if !exists {
		return nil, model.ErrBookmarkNotFound
	}

	// Convert to DTO for database operations
	dto := bookmark.ToDTO()

	// Update bookmark in database
	savedBookmarks, err := d.deps.Database().SaveBookmarks(ctx, false, dto)
	if err != nil {
		return nil, fmt.Errorf("failed to update bookmark: %w", err)
	}

	if len(savedBookmarks) == 0 {
		return nil, fmt.Errorf("no bookmark was updated")
	}

	savedBookmark := savedBookmarks[0]

	// Set additional properties
	savedBookmark.HasEbook = d.HasEbook(&savedBookmark)
	savedBookmark.HasArchive = d.HasArchive(&savedBookmark)
	
	return &savedBookmark, nil
}

// DeleteBookmarks deletes multiple bookmarks by their IDs
func (d *BookmarksDomain) DeleteBookmarks(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	err := d.deps.Database().DeleteBookmarks(ctx, ids...)
	if err != nil {
		return fmt.Errorf("failed to delete bookmarks: %w", err)
	}

	return nil
}

func NewBookmarksDomain(deps model.Dependencies) *BookmarksDomain {
	return &BookmarksDomain{
		deps: deps,
	}
}
