package domains

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/archiver"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
)

type ArchiverDomain struct {
	deps      *dependencies.Dependencies
	archivers []model.Archiver
}

func (d *ArchiverDomain) DownloadBookmarkArchive(book model.BookmarkDTO) (*model.BookmarkDTO, error) {
	content, contentType, err := core.DownloadBookmark(book.URL)
	if err != nil {
		return nil, fmt.Errorf("error downloading url: %s", err)
	}

	return d.ProcessBookmarkArchive(content, contentType, book)
}

func (d *ArchiverDomain) ProcessBookmarkArchive(content io.ReadCloser, contentType string, book model.BookmarkDTO) (*model.BookmarkDTO, error) {
	for _, archiver := range d.archivers {
		if archiver.Matches(contentType) {
			return archiver.Archive(content, contentType, book)
		}
	}

	return nil, fmt.Errorf("no archiver found for content type: %s", contentType)
}

func (d *ArchiverDomain) GetBookmarkArchive(book *model.BookmarkDTO) (*warc.Archive, error) {
	archivePath := model.GetArchivePath(book)

	if !d.deps.Domains.Storage.FileExists(archivePath) {
		return nil, fmt.Errorf("archive for bookmark %d doesn't exist", book.ID)
	}

	// FIXME: This only works in local filesystem
	return warc.Open(filepath.Join(d.deps.Config.Storage.DataDir, archivePath))
}

func NewArchiverDomain(deps *dependencies.Dependencies) *ArchiverDomain {
	archivers := []model.Archiver{
		archiver.NewPDFArchiver(deps),
		archiver.NewWARCArchiver(deps),
	}
	return &ArchiverDomain{
		deps:      deps,
		archivers: archivers,
	}
}
