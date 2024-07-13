package archiver

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
)

// LEGACY WARNING
// This file contains legacy code that will be removed once we move on to Obelisk as
// general archiver.

type WARCArchiver struct {
	deps *dependencies.Dependencies
}

func (a *WARCArchiver) Matches(contentType string) bool {
	return true
}

func (a *WARCArchiver) Archive(content io.ReadCloser, contentType string, bookmark model.BookmarkDTO) (*model.BookmarkDTO, error) {
	processRequest := core.ProcessRequest{
		DataDir:     a.deps.Config.Storage.DataDir,
		Bookmark:    bookmark,
		Content:     content,
		ContentType: contentType,
	}

	result, isFatalErr, err := core.ProcessBookmark(a.deps, processRequest)
	content.Close()

	if err != nil && isFatalErr {
		return nil, fmt.Errorf("failed to process: %v", err)
	}

	return &result, nil
}

func (a *WARCArchiver) GetArchiveFile(bookmark model.BookmarkDTO, resourcePath string) (*model.ArchiveFile, error) {
	archivePath := model.GetArchivePath(&bookmark)

	if !a.deps.Domains.Storage.FileExists(archivePath) {
		return nil, fmt.Errorf("archive for bookmark %d doesn't exist", bookmark.ID)
	}

	warcFile, err := warc.Open(filepath.Join(a.deps.Config.Storage.DataDir, archivePath))
	if err != nil {
		return nil, fmt.Errorf("error opening warc file: %w", err)
	}

	defer warcFile.Close()

	if !warcFile.HasResource(resourcePath) {
		return nil, fmt.Errorf("resource %s doesn't exist in archive", resourcePath)
	}

	content, contentType, err := warcFile.Read(resourcePath)
	if err != nil {
		return nil, fmt.Errorf("error reading resource %s: %w", resourcePath, err)
	}

	// Note: Using this method to send the reader instead of `bytes.NewReader` because that
	// crashes the moment we try to retrieve it for some reason. Since this is a legacy archiver
	// I don't want to spend more time on this. (@fmartingr)
	return model.NewArchiveFile(strings.NewReader(string(content)), contentType, "gzip", int64(len(content))), nil
}

func NewWARCArchiver(deps *dependencies.Dependencies) *WARCArchiver {
	return &WARCArchiver{
		deps: deps,
	}
}
