package domains

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
)

type BuiltInArchiver struct {
	deps *dependencies.Dependencies
}

func (d *BuiltInArchiver) ArchiveBookmark(book *model.BookmarkDTO, logEnabled bool) error {
	tmpFile, err := os.CreateTemp("", "archive")
	if err != nil {
		return fmt.Errorf("failed to create temp archive: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	archivalRequest := warc.ArchivalRequest{
		URL:         book.URL,
		UserAgent:   model.UserAgent,
		LogEnabled:  logEnabled,
	}

	err = warc.NewArchive(archivalRequest, tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}

	// Construct the destination path using the BookmarkID
	dstPath := model.GetArchivePath(book)

	err = d.deps.Domains().Storage().WriteFile(dstPath, tmpFile)
	if err != nil {
		return fmt.Errorf("failed to move archive to destination: %v", err)
	}

	return nil
}

func (d *BuiltInArchiver) HasArchive(book *model.BookmarkDTO) bool {
	archivePath := model.GetArchivePath(book)
	return d.deps.Domains().Storage().FileExists(archivePath)
}

func (d *BuiltInArchiver) GetBookmarkArchive(book *model.BookmarkDTO) (model.Archive, error) {
	archivePath := model.GetArchivePath(book)

	if !d.deps.Domains().Storage().FileExists(archivePath) {
		return nil, fmt.Errorf("archive for bookmark %d doesn't exist", book.ID)
	}

	// FIXME: This only works in local filesystem
	return warc.Open(filepath.Join(d.deps.Config().Storage.DataDir, archivePath))
}

func (d *BuiltInArchiver) DeleteArchive(book *model.BookmarkDTO) error {
	archivePath := model.GetArchivePath(book)
	if !d.deps.Domains().Storage().FileExists(archivePath) {
		return nil // No archive to delete
	}

	// Delete the archive file using the storage filesystem
	err := d.deps.Domains().Storage().FS().Remove(archivePath)
	if err != nil {
		return fmt.Errorf("failed to delete archive: %v", err)
	}

	return nil
}

func NewBuiltInArchiver(deps *dependencies.Dependencies) *BuiltInArchiver {
	return &BuiltInArchiver{
		deps: deps,
	}
}
