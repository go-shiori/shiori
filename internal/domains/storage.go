package domains

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/afero"
)

type StorageDomain struct {
	deps model.Dependencies
	fs   afero.Fs
}

func NewStorageDomain(deps model.Dependencies, fs afero.Fs) *StorageDomain {
	return &StorageDomain{
		deps: deps,
		fs:   fs,
	}
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

// WriteData writes bytes data to a file in storage.
// CAUTION: This function will overwrite existing file.
func (d *StorageDomain) WriteData(dst string, data []byte) error {
	// Create directory if not exist
	dir := filepath.Dir(dst)
	if !d.DirExists(dir) {
		err := d.fs.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create file
	file, err := d.fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data
	_, err = file.Write(data)
	return err
}

// WriteFile writes a file to storage.
func (d *StorageDomain) WriteFile(dst string, tmpFile *os.File) error {
	if dst != "" && !d.DirExists(dst) {
		err := d.fs.MkdirAll(filepath.Dir(dst), model.DataDirPerm)
		if err != nil {
			return fmt.Errorf("failed to create destination dir: %v", err)
		}
	}

	dstFile, err := d.fs.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to rewind temporary file: %v", err)
	}

	_, err = io.Copy(dstFile, tmpFile)
	if err != nil {
		return fmt.Errorf("failed to copy file to the destination")
	}

	return nil
}
