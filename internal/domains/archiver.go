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

// GenerateBookmarkEbook implements the interface method
func (d *ArchiverDomain) GenerateBookmarkEbook(request model.EbookProcessRequest) error {
	// For now, just return nil - this can be implemented later
	return nil
}

// ProcessBookmarkArchive implements the interface method
func (d *ArchiverDomain) ProcessBookmarkArchive(archiverReq *model.ArchiverRequest) (*model.BookmarkDTO, error) {
	// Use the archiver system to process the request
	for _, archiver := range d.archivers {
		if archiver.Matches(archiverReq) {
			return archiver.Archive(archiverReq)
		}
	}
	return nil, fmt.Errorf("no suitable archiver found for request")
}

// GetBookmarkArchiveFile implements the interface method
func (d *ArchiverDomain) GetBookmarkArchiveFile(book *model.BookmarkDTO, resourcePath string) (*model.ArchiveFile, error) {
	// Try to find an appropriate archiver for this bookmark
	for _, archiver := range d.archivers {
		// For now, try all archivers - this could be improved with better detection
		if archiveFile, err := archiver.GetArchiveFile(*book, resourcePath); err == nil {
			return archiveFile, nil
		}
	}
	return nil, fmt.Errorf("no archive file found for bookmark %d at path %s", book.ID, resourcePath)
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
