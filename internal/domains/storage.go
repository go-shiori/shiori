package domains

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/spf13/afero"
)

type StorageDomain struct {
	deps *dependencies.Dependencies
	fs   afero.Fs
}

// Stat returns the FileInfo structure describing file.
func (d *StorageDomain) Stat(name string) (fs.FileInfo, error) {
	return d.fs.Stat(name)
}

// FS returns the filesystem used by this domain.
func (d *StorageDomain) FS() afero.Fs {
	return d.fs
}

// FileExists checks if a file exists in storage.
func (d *StorageDomain) FileExists(name string) bool {
	info, err := d.Stat(name)
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists in storage.
func (d *StorageDomain) DirExists(name string) bool {
	info, err := d.Stat(name)
	return err == nil && info.IsDir()
}

// Write writes data to a file in storage.
// CAUTION: This function will overwrite existing file.
func (d *StorageDomain) Save(name string, data []byte) error {
	// Create directory if not exist
	dir := filepath.Dir(name)
	if !d.DirExists(dir) {
		err := d.fs.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create file
	file, err := d.fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data
	_, err = file.Write(data)
	return err
}

func NewStorageDomain(deps *dependencies.Dependencies, fs afero.Fs) *StorageDomain {
	return &StorageDomain{
		deps: deps,
		fs:   fs,
	}
}
