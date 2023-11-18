package domains

import (
	"io/fs"

	"github.com/go-shiori/shiori/internal/dependencies"
)

type StorageDomain struct {
	deps *dependencies.Dependencies
	fs   fs.FS
}

// FileExists checks if a file exists in storage.
func (d *StorageDomain) FileExists(name string) bool {
	info, err := fs.Stat(d.fs, name)
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists in storage.
func (d *StorageDomain) DirExists(name string) bool {
	info, err := fs.Stat(d.fs, name)
	d.deps.Log.Info(info)
	return err == nil && info.IsDir()
}

func NewStorageDomain(deps *dependencies.Dependencies, fs fs.FS) *StorageDomain {
	return &StorageDomain{
		deps: deps,
		fs:   fs,
	}
}
