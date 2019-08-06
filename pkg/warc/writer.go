package warc

import (
	"fmt"
	"io"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/go-shiori/shiori/pkg/warc/internal/archiver"
	"go.etcd.io/bbolt"
)

// ArchivalRequest is request for archiving a web page,
// either from URL or from an io.Reader.
type ArchivalRequest struct {
	URL         string
	Reader      io.Reader
	ContentType string
	LogEnabled  bool
}

// NewArchive creates new archive based on submitted request,
// then save it to specified path.
func NewArchive(req ArchivalRequest, dstPath string) error {
	// Make sure URL is valid
	parsedURL, err := nurl.ParseRequestURI(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Hostname() == "" {
		return fmt.Errorf("url %s is not valid", req.URL)
	}

	// Generate resource URL
	res := archiver.ToResourceURL(req.URL, parsedURL)
	res.ArchivalURL = "archive-root"

	// Download URL if needed
	if req.Reader == nil || req.ContentType == "" {
		resp, err := archiver.DownloadData(res.DownloadURL)
		if err != nil {
			return fmt.Errorf("failed to download %s: %v", req.URL, err)
		}
		defer resp.Body.Close()

		req.Reader = resp.Body
		req.ContentType = resp.Header.Get("Content-Type")
	}

	// Create database for archive
	os.MkdirAll(fp.Dir(dstPath), os.ModePerm)

	db, err := bbolt.Open(dstPath, os.ModePerm, nil)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}
	defer db.Close()

	// Create archiver
	arc := &archiver.Archiver{
		DB:          db,
		ChDone:      make(chan struct{}),
		ChErrors:    make(chan error),
		ChWarnings:  make(chan error),
		ChRequest:   make(chan archiver.ResourceURL, 10),
		ResourceMap: make(map[string]struct{}),
		LogEnabled:  req.LogEnabled,
	}
	defer arc.Close()

	// Process input depending on its type.
	// If it's HTML, we need to extract the sub resources that used by it, e.g some CSS or JS files.
	// If it's not HTML, we can just save it to archive.
	var result archiver.ProcessResult
	var subResources []archiver.ResourceURL

	if strings.Contains(req.ContentType, "text/html") {
		result, subResources, err = arc.ProcessHTMLFile(res, req.Reader)
	} else {
		result, err = arc.ProcessOtherFile(res, req.Reader)
	}

	if err != nil {
		return fmt.Errorf("archival failed: %v", err)
	}

	// Add this url to resource map to mark it as processed
	arc.ResourceMap[res.DownloadURL] = struct{}{}

	// Save content to storage
	arc.Logf(0, "Downloaded %s", res.DownloadURL)

	result.ContentType = req.ContentType
	err = arc.SaveToStorage(result)
	if err != nil {
		return fmt.Errorf("failed to save %s: %v", res.DownloadURL, err)
	}

	// If there are no sub resources found, our job is finished.
	if len(subResources) == 0 {
		return nil
	}

	// However, if there are, we need to run the archiver in background to
	// process the sub resources concurrently.
	go func() {
		for _, subRes := range subResources {
			arc.ChRequest <- subRes
		}
	}()

	time.Sleep(time.Second)
	arc.StartArchiver()
	return nil
}
