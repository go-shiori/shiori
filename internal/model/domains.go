package model

import (
	"context"
	"html/template"
	"time"

	"github.com/go-shiori/warc"
)

type BookmarksDomain interface {
	HasEbook(b *BookmarkDTO) bool
	HasArchive(b *BookmarkDTO) bool
	GetThumbnailPath(b *BookmarkDTO) string
	HasThumbnail(b *BookmarkDTO) bool
	GetBookmark(ctx context.Context, id DBID) (*BookmarkDTO, error)
	GetBookmarkContentsFromArchive(bookmark *BookmarkDTO) (template.HTML, error)
}

type AccountsDomain interface {
	CheckToken(ctx context.Context, userJWT string) (*Account, error)
	GetAccountFromCredentials(ctx context.Context, username, password string) (*Account, error)
	CreateTokenForAccount(account *Account, expiration time.Time) (string, error)
}

type ArchiverDomain interface {
	DownloadBookmarkArchive(book BookmarkDTO) (*BookmarkDTO, error)
	GetBookmarkArchive(book *BookmarkDTO) (*warc.Archive, error)
}

type StorageDomain interface {
	FileExists(path string) bool
	DirExists(path string) bool
}
