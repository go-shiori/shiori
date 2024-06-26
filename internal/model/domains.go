package model

import (
	"context"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/spf13/afero"
)

type BookmarksDomain interface {
	HasEbook(b *BookmarkDTO) bool
	HasArchive(b *BookmarkDTO) bool
	HasThumbnail(b *BookmarkDTO) bool
	GetBookmark(ctx context.Context, id DBID) (*BookmarkDTO, error)
}

type AccountsDomain interface {
	CheckToken(ctx context.Context, userJWT string) (*Account, error)
	GetAccountFromCredentials(ctx context.Context, username, password string) (*Account, error)
	CreateTokenForAccount(account *Account, expiration time.Time) (string, error)
}

type ArchiverDomain interface {
	GenerateBookmarkArchive(book BookmarkDTO) (*BookmarkDTO, error)
	GenerateBookmarkEbook(book EbookProcessRequest) error
	ProcessBookmarkArchive(content io.ReadCloser, contentType string, book BookmarkDTO) (*BookmarkDTO, error)
	GetBookmarkArchiveFile(book *BookmarkDTO, archivePath string) (*ArchiveFile, error)
}

type StorageDomain interface {
	// Open(name string) (os.File, error)
	Stat(name string) (fs.FileInfo, error)
	FS() afero.Fs
	FileExists(path string) bool
	DirExists(path string) bool
	WriteData(dst string, data []byte) error
	WriteFile(dst string, src *os.File) error
	WriteReader(dst string, src io.Reader) error
}
