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

// FromReader create archive from the specified io.Reader.
func FromReader(input io.Reader, url, contentType, dstPath string) error {
	// Make sure URL is valid
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Hostname() == "" {
		return fmt.Errorf("url %s is not valid", url)
	}

	// Generate resource URL
	res := archiver.ToResourceURL(url, parsedURL)
	res.ArchivalURL = "archive-root"

	// Create database for archive
	os.MkdirAll(fp.Dir(dstPath), os.ModePerm)

	db, err := bbolt.Open(dstPath, os.ModePerm, nil)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}

	// Create archiver
	arc := &archiver.Archiver{
		DB:          db,
		ChDone:      make(chan struct{}),
		ChErrors:    make(chan error),
		ChWarnings:  make(chan error),
		ChRequest:   make(chan archiver.ResourceURL, 10),
		ResourceMap: make(map[string]struct{}),
		LogEnabled:  true,
	}
	defer arc.Close()

	// Process input depending on its type.
	// If it's HTML, we need to extract the sub resources that used by it, e.g some CSS or JS files.
	// If it's not HTML, we can just save it to archive.
	var result archiver.ProcessResult
	var subResources []archiver.ResourceURL

	if strings.Contains(contentType, "text/html") {
		result, subResources, err = arc.ProcessHTMLFile(res, input)
	} else {
		result, err = arc.ProcessOtherFile(res, input)
	}

	if err != nil {
		return fmt.Errorf("archival failed: %v", err)
	}

	// Add this url to resource map to mark it as processed
	arc.ResourceMap[res.DownloadURL] = struct{}{}

	// Save content to storage
	arc.Logf(0, "Downloaded %s", res.DownloadURL)

	result.ContentType = contentType
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

// FromURL create archive from the specified URL.
func FromURL(url, dstPath string) error {
	// Download URL
	resp, err := archiver.DownloadData(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	return FromReader(resp.Body, url, contentType, dstPath)
}
