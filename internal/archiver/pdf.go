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

func NewPDFArchiver(deps *dependencies.Dependencies) *PDFArchiver {
	return &PDFArchiver{
		deps: deps,
	}
}
