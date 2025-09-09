package domains

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/archiver"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
)

type ArchiverDomain struct {
	deps      *dependencies.Dependencies
	archivers map[string]model.Archiver
}

func (d *ArchiverDomain) GenerateBookmarkArchive(book model.BookmarkDTO) (*model.BookmarkDTO, error) {
	content, contentType, err := core.DownloadBookmark(book.URL)
	if err != nil {
		return nil, fmt.Errorf("error downloading url: %s", err)
	}

	processRequest := core.ProcessRequest{
		DataDir:     d.deps.Config().Storage.DataDir,
		Bookmark:    book,
		Content:     content,
		ContentType: contentType,
	}
	content.Close()

	processedBookmark, _, err := core.ProcessBookmark(d.deps, processRequest)
	if err != nil {
		return nil, fmt.Errorf("error processing bookmark archive: %w", err)
	}

	saved, err := d.deps.Database().SaveBookmarks(context.Background(), false, processedBookmark)
	if err != nil {
		return nil, fmt.Errorf("error saving bookmark: %w", err)
	}

	return &saved[0], nil
}

func (d *ArchiverDomain) GetBookmarkArchive(book *model.BookmarkDTO) (*warc.Archive, error) {
	archivePath := model.GetArchivePath(book)

	if !d.deps.Domains().Storage().FileExists(archivePath) {
		return nil, fmt.Errorf("archive for bookmark %d doesn't exist", book.ID)
	}

	// FIXME: This only works in local filesystem
	return warc.Open(filepath.Join(d.deps.Config().Storage.DataDir, archivePath))
}

func NewArchiverDomain(deps *dependencies.Dependencies) *ArchiverDomain {
	archivers := map[string]model.Archiver{
		model.ArchiverPDF:  archiver.NewPDFArchiver(deps),
		model.ArchiverWARC: archiver.NewWARCArchiver(deps),
	}
	return &ArchiverDomain{
		deps:      deps,
		archivers: archivers,
	}
}
