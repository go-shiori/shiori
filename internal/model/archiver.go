package model

import "io"

const (
	ArchiverPDF  = "pdf"
	ArchiverWARC = "warc"
)

type Archiver interface {
	Archive(content io.ReadCloser, contentType string, bookmark BookmarkDTO) (*BookmarkDTO, error)
	Matches(contentType string) bool
}
