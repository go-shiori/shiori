package domains

import (
	"context"
	"fmt"
	"io"

	"github.com/go-shiori/shiori/internal/archiver"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
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

	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("error reading content: %s", err)
	}
	content.Close()

	archiverReq := model.NewArchiverRequest(book, contentType, contentBytes)

	processedBookmark, err := d.ProcessBookmarkArchive(archiverReq)
	if err != nil {
		return nil, fmt.Errorf("error processing bookmark archive: %w", err)
	}

	saved, err := d.deps.Database.SaveBookmarks(context.Background(), false, *processedBookmark)
	if err != nil {
		return nil, fmt.Errorf("error saving bookmark: %w", err)
	}

	return &saved[0], nil
}

func (d *ArchiverDomain) GenerateBookmarkEbook(request model.EbookProcessRequest) error {
	_, err := core.GenerateEbook(d.deps, request)
	if err != nil {
		return fmt.Errorf("error generating ebook: %s", err)
	}

	return nil
}

func (d *ArchiverDomain) ProcessBookmarkArchive(archiverRequest *model.ArchiverRequest) (*model.BookmarkDTO, error) {
	for _, archiver := range d.archivers {
		if archiver.Matches(archiverRequest) {
			book, err := archiver.Archive(archiverRequest)
			if err != nil {
				d.deps.Log.Errorf("Error archiving bookmark with archviver: %s", err)
				continue
			}
			return book, nil
		}
	}

	return nil, fmt.Errorf("no archiver found for request: %s", archiverRequest.String())
}

func (d *ArchiverDomain) GetBookmarkArchiveFile(book *model.BookmarkDTO, resourcePath string) (*model.ArchiveFile, error) {
	archiver, err := d.GetArchiver(book.Archiver)
	if err != nil {
		return nil, err
	}

	archiveFile, err := archiver.GetArchiveFile(*book, resourcePath)
	if err != nil {
		return nil, fmt.Errorf("error getting archive file: %w", err)
	}

	return archiveFile, nil
}

func (d *ArchiverDomain) GetArchiver(name string) (model.Archiver, error) {
	archiver, ok := d.archivers[name]
	if !ok {
		return nil, fmt.Errorf("archiver %s not found", name)
	}
	return archiver, nil
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
