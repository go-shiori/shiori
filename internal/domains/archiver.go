package domains

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
	"github.com/sirupsen/logrus"
)

type ArchiverDomain struct {
	dataDir string
	logger  *logrus.Logger
}

func (d *ArchiverDomain) DownloadBookmarkArchive(book model.BookmarkDTO) (*model.BookmarkDTO, error) {
	content, contentType, err := core.DownloadBookmark(book.URL)
	if err != nil {
		return nil, fmt.Errorf("error downloading url: %s", err)
	}

	processRequest := core.ProcessRequest{
		DataDir:     d.dataDir,
		Bookmark:    book,
		Content:     content,
		ContentType: contentType,
	}

	result, isFatalErr, err := core.ProcessBookmark(processRequest)
	content.Close()

	if err != nil && isFatalErr {
		return nil, fmt.Errorf("failed to process: %v", err)
	}

	return &result, nil
}

func (d *ArchiverDomain) GetBookmarkArchive(book *model.BookmarkDTO) (*warc.Archive, error) {
	archivePath := filepath.Join(d.dataDir, "archive", strconv.Itoa(book.ID))

	if !FileExists(archivePath) {
		return nil, fmt.Errorf("archive not found")
	}

	return warc.Open(archivePath)
}

func NewArchiverDomain(logger *logrus.Logger, dataDir string) ArchiverDomain {
	return ArchiverDomain{
		dataDir: dataDir,
		logger:  logger,
	}
}
