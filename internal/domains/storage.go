package domains

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

type StorageDomain struct {
	deps *dependencies.Dependencies
	path string
}

func NewStorageDomain(deps *dependencies.Dependencies, path string) *StorageDomain {
	return &StorageDomain{
		deps: deps,
		path: path,
	}
}

func (d *StorageDomain) generateFullPath(name string) string {
	return filepath.Join(d.path, name)
}

// Stat returns the FileInfo structure describing file.
func (d *StorageDomain) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(d.generateFullPath(name))
}

// Create creates a file in storage.
func (d *StorageDomain) Create(name string) (fs.File, error) {
	return os.Create(d.generateFullPath(name))
}

// Open opens a file in storage.
func (d *StorageDomain) Open(name string) (fs.File, error) {
	return os.Open(d.generateFullPath(name))
}

// MkDirAll creates a directory in storage.
func (d *StorageDomain) MkDirAll(name string, mode os.FileMode) error {
	return os.MkdirAll(d.generateFullPath(name), mode)
}

// Remove removes a file in storage.
func (d *StorageDomain) Remove(name string) error {
	return os.Remove(d.generateFullPath(name))
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
	if dir != "" && !d.DirExists(dir) {
		if err := d.MkDirAll(dir, model.DataDirPerm); err != nil {
			return err
		}
	}

	// Create file
	file, err := os.Create(d.generateFullPath(dst))
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
	dir := filepath.Dir(dst)
	if dir != "" && !d.DirExists(dir) {
		err := d.MkDirAll(dir, model.DataDirPerm)
		if err != nil {
			return fmt.Errorf("failed to create destination dir: %v", err)
		}
	}

	dstFile, err := os.Create(d.generateFullPath(dst))
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
