package model

import "io"

type Archiver interface {
	Archive(content io.ReadCloser, contentType string, bookmark BookmarkDTO) (*BookmarkDTO, error)
	Matches(contentType string) bool
}
