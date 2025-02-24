package model

import (
	"context"
	"io/fs"
	"os"
	"time"

	"github.com/go-shiori/warc"
	"github.com/spf13/afero"
)

type BookmarksDomain interface {
	HasEbook(b *BookmarkDTO) bool
	HasArchive(b *BookmarkDTO) bool
	HasThumbnail(b *BookmarkDTO) bool
	GetBookmark(ctx context.Context, id DBID) (*BookmarkDTO, error)
	GetBookmarks(ctx context.Context, ids []int) ([]BookmarkDTO, error)
	UpdateBookmarkCache(ctx context.Context, bookmark BookmarkDTO, keepMetadata bool, skipExist bool) (*BookmarkDTO, error)
}

type AuthDomain interface {
	CheckToken(ctx context.Context, userJWT string) (*AccountDTO, error)
	GetAccountFromCredentials(ctx context.Context, username, password string) (*AccountDTO, error)
	CreateTokenForAccount(account *AccountDTO, expiration time.Time) (string, error)
}

type AccountsDomain interface {
	ListAccounts(ctx context.Context) ([]AccountDTO, error)
	CreateAccount(ctx context.Context, account AccountDTO) (*AccountDTO, error)
	UpdateAccount(ctx context.Context, account AccountDTO) (*AccountDTO, error)
	DeleteAccount(ctx context.Context, id int) error
}

type ArchiverDomain interface {
	DownloadBookmarkArchive(book BookmarkDTO) (*BookmarkDTO, error)
	GetBookmarkArchive(book *BookmarkDTO) (*warc.Archive, error)
}

type StorageDomain interface {
	Stat(name string) (fs.FileInfo, error)
	FS() afero.Fs
	FileExists(path string) bool
	DirExists(path string) bool
	WriteData(dst string, data []byte) error
	WriteFile(dst string, src *os.File) error
}

type TagsDomain interface {
	ListTags(ctx context.Context) ([]TagDTO, error)
	CreateTag(ctx context.Context, tag TagDTO) (TagDTO, error)
	// GetTag(ctx context.Context, id int64) (TagDTO, error)
	// UpdateTag(ctx context.Context, tag TagDTO) (TagDTO, error)
	// DeleteTag(ctx context.Context, id int64) error
}
