package model

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// Archive represents an interface for accessing archived web pages
type Archive interface {
	// Close closes the archive
	Close()
	// HasResource checks if a resource exists in the archive
	HasResource(name string) bool
	// Read reads a resource from the archive
	Read(name string) ([]byte, string, error)
}

// SingleFileArchive is a simple implementation of Archive for external archivers
// that wraps around string content instead of reading from files
type SingleFileArchive struct {
	content []byte
	closed  bool
}

// NewSingleFileArchive creates a new SingleFileArchive instance
func NewSingleFileArchive(content []byte) *SingleFileArchive {
	return &SingleFileArchive{
		content: content,
		closed:  false,
	}
}

// Close implements the Archive interface
func (a *SingleFileArchive) Close() {
	a.closed = true
	a.content = nil
}

// HasResource implements the Archive interface
func (a *SingleFileArchive) HasResource(name string) bool {
	// If any other file but the default one is requested, return false
	return name == ""
}

// Read implements the Archive interface
func (a *SingleFileArchive) Read(name string) ([]byte, string, error) {
	if a.closed {
		return nil, "", fmt.Errorf("archive is closed")
	}
	
	// If any other file but the default one is requested, return an error
	if name != "" {
		return nil, "", io.EOF
	}

	if len(a.content) == 0 {
		return nil, "", fmt.Errorf("archive content is empty")
	}

	// Compress the content with gzip
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(a.content); err != nil {
		return nil, "", fmt.Errorf("failed to compress content: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes(), "text/html", nil
}


