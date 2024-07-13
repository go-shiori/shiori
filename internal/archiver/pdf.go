package archiver

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

type PDFArchiver struct {
	deps *dependencies.Dependencies
}

func (a *PDFArchiver) Matches(contentType string) bool {
	return strings.Contains(contentType, "application/pdf")
}

func (a *PDFArchiver) Archive(content io.ReadCloser, contentType string, bookmark model.BookmarkDTO) (*model.BookmarkDTO, error) {
	if err := a.deps.Domains.Storage.WriteReader(model.GetArchivePath(&bookmark), content); err != nil {
		return nil, fmt.Errorf("error saving pdf archive: %v", err)
	}

	bookmark.ArchivePath = model.GetArchivePath(&bookmark)
	bookmark.HasArchive = true
	bookmark.Archiver = model.ArchiverPDF

	return &bookmark, nil
}

func (a *PDFArchiver) GetArchiveFile(bookmark model.BookmarkDTO, resourcePath string) (*model.ArchiveFile, error) {
	archivePath := model.GetArchivePath(&bookmark)

	if !a.deps.Domains.Storage.FileExists(archivePath) {
		return nil, fmt.Errorf("archive for bookmark %d doesn't exist", bookmark.ID)
	}

	archiveFile, err := a.deps.Domains.Storage.FS().Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("error opening pdf archive: %w", err)
	}

	info, err := archiveFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting pdf archive info: %w", err)
	}

	return model.NewArchiveFile(archiveFile, "application/pdf", "", info.Size()), nil
}

func NewPDFArchiver(deps *dependencies.Dependencies) *PDFArchiver {
	return &PDFArchiver{
		deps: deps,
	}
}
