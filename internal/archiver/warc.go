package archiver

import (
	"fmt"
	"io"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

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

func NewWARCArchiver(deps *dependencies.Dependencies) *WARCArchiver {
	return &WARCArchiver{
		deps: deps,
	}
}
