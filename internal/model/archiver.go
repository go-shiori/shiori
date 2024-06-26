package model

import (
	"io"
	"strconv"
)

const (
	ArchiverPDF  = "pdf"
	ArchiverWARC = "warc"
)

type ArchiveFile struct {
	reader      io.Reader
	contentType string
	size        int64 // bytes
	encoding    string
}

func (a *ArchiveFile) Reader() io.Reader {
	return a.reader
}

func (a *ArchiveFile) ContentType() string {
	return a.contentType
}

func (a *ArchiveFile) Size() int64 {
	return a.size
}

func (a *ArchiveFile) Encoding() string {
	return a.encoding
}

func (a *ArchiveFile) AsHTTPHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type": a.contentType,
	}

	if a.size > 0 {
		headers["Content-Length"] = strconv.FormatInt(a.size, 10)
	}

	if a.encoding != "" {
		headers["Content-Encoding"] = a.encoding
	}

	return headers
}

func NewArchiveFile(reader io.Reader, contentType, encoding string, size int64) *ArchiveFile {
	return &ArchiveFile{
		reader:      reader,
		contentType: contentType,
		encoding:    encoding,
		size:        size,
	}
}

type ArchiveProcessRequest struct {
	Bookmark     BookmarkDTO
	Content      io.Reader
	ContentType  string
	SkipExisting bool
}

type EbookProcessRequest struct {
	Bookmark     BookmarkDTO
	SkipExisting bool
}

type Archiver interface {
	Archive(content io.ReadCloser, contentType string, bookmark BookmarkDTO) (*BookmarkDTO, error)
	Matches(contentType string) bool
	GetArchiveFile(bookmark BookmarkDTO, resourcePath string) (*ArchiveFile, error)
}
