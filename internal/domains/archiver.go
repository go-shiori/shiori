package domains

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
)

type ArchiverDomain struct {
	deps *dependencies.Dependencies
}

func (d *ArchiverDomain) ArchiveBookmark(req model.ArchivalRequest) error {
	tmpFile, err := os.CreateTemp("", "archive")
	if err != nil {
		return fmt.Errorf("failed to create temp archive: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Use default user agent if not provided
	userAgent := req.UserAgent
	if userAgent == "" {
		userAgent = "Shiori/1.0" // Default user agent
	}

	// The Reader field is left nil, so warc.NewArchive will download the content
	archivalRequest := warc.ArchivalRequest{
		URL:         req.Bookmark.URL,
		ContentType: req.ContentType,
		UserAgent:   userAgent,
		LogEnabled:  req.LogEnabled,
	}

	err = warc.NewArchive(archivalRequest, tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}

	// Construct the destination path using the BookmarkID
	dstPath := model.GetArchivePath(req.Bookmark)

	err = d.deps.Domains().Storage().WriteFile(dstPath, tmpFile)
	if err != nil {
		return fmt.Errorf("failed to move archive to destination: %v", err)
	}

	return nil
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
	return &ArchiverDomain{
		deps: deps,
	}
}
