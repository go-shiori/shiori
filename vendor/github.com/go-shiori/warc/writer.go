package warc

import (
	"fmt"
	"io"
	nurl "net/url"
	"os"
	fp "path/filepath"

	"github.com/go-shiori/warc/internal/archiver"
	"go.etcd.io/bbolt"
)

// ArchivalRequest is request for archiving a web page,
// either from URL or from an io.Reader.
type ArchivalRequest struct {
	URL         string
	Reader      io.Reader
	ContentType string
	UserAgent   string
	LogEnabled  bool
}

// NewArchive creates new archive based on submitted request,
// then save it to specified path.
func NewArchive(req ArchivalRequest, dstPath string) error {
	// Make sure URL is valid
	parsedURL, err := nurl.ParseRequestURI(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Hostname() == "" {
		return fmt.Errorf("url \"%s\" is not valid", req.URL)
	}

	// Create database for archive
	os.MkdirAll(fp.Dir(dstPath), os.ModePerm)

	db, err := bbolt.Open(dstPath, os.ModePerm, nil)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}
	defer db.Close()

	// Start archival
	arc := archiver.Archiver{
		DB:         db,
		UserAgent:  req.UserAgent,
		LogEnabled: req.LogEnabled,
	}

	arcRequest := archiver.Request{
		URL:         req.URL,
		Reader:      req.Reader,
		ContentType: req.ContentType,
	}

	err = arc.Start(arcRequest)
	if err != nil {
		return fmt.Errorf("archival failed: %v", err)
	}

	return nil
}
