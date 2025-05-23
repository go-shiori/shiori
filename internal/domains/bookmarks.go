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
	content, contentType, err := core.DownloadBookmark(bookmark.URL)
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

func NewBookmarksDomain(deps model.Dependencies) *BookmarksDomain {
	return &BookmarksDomain{
		deps: deps,
	}
}
