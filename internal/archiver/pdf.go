package archiver

import (
	"fmt"
	"io"
	"os"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

type PDFArchiver struct {
	deps *dependencies.Dependencies
}

func (a *PDFArchiver) Matches(contentType string) bool {
	return contentType == "application/pdf"
}

func (a *PDFArchiver) Archive(content io.ReadCloser, contentType string, bookmark model.BookmarkDTO) (*model.BookmarkDTO, error) {
	if err := a.deps.Domains.Storage.WriteFile(model.GetArchivePath(&bookmark), content.(*os.File)); err != nil {
		return nil, fmt.Errorf("error saving pdf archive: %v", err)
	}

	return nil, nil
}

func NewPDFArchiver(deps *dependencies.Dependencies) *PDFArchiver {
	return &PDFArchiver{
		deps: deps,
	}
}
