package domains

import (
	"context"
	"fmt"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

type BookmarksDomain struct {
	deps *dependencies.Dependencies
}

func (d *BookmarksDomain) HasEbook(b *model.BookmarkDTO) bool {
	ebookPath := model.GetEbookPath(b)
	return d.deps.Domains.Storage.FileExists(ebookPath)
}

func (d *BookmarksDomain) HasArchive(b *model.BookmarkDTO) bool {
	archivePath := model.GetArchivePath(b)
	return d.deps.Domains.Storage.FileExists(archivePath)
}

func (d *BookmarksDomain) HasThumbnail(b *model.BookmarkDTO) bool {
	thumbnailPath := model.GetThumbnailPath(b)
	return d.deps.Domains.Storage.FileExists(thumbnailPath)
}

func (d *BookmarksDomain) GetBookmark(ctx context.Context, id model.DBID) (*model.BookmarkDTO, error) {
	bookmark, exists, err := d.deps.Database.GetBookmark(ctx, int(id), "")
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

func NewBookmarksDomain(deps *dependencies.Dependencies) *BookmarksDomain {
	return &BookmarksDomain{
		deps: deps,
	}
}
