package warc

import (
	"fmt"
	"os"

	"go.etcd.io/bbolt"
)

// Archive is the storage for archiving the web page.
type Archive struct {
	db *bbolt.DB
}

// Open opens the archive from specified path.
func Open(path string) (*Archive, error) {
	// Make sure archive exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) || info.IsDir() {
		return nil, fmt.Errorf("archive doesn't exist")
	}

	// Open database
	options := &bbolt.Options{
		ReadOnly: true,
	}

	db, err := bbolt.Open(path, os.ModePerm, options)
	if err != nil {
		return nil, err
	}

	return &Archive{db: db}, nil
}

// Close closes the storage.
func (arc *Archive) Close() {
	arc.db.Close()
}

// Read fetch the resource with specified name from archive.
func (arc *Archive) Read(name string) ([]byte, string, error) {
	// Make sure name exists
	if name == "" {
		name = "archive-root"
	}

	var content []byte
	var strContentType string

	err := arc.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		if bucket == nil {
			return fmt.Errorf("%s doesn't exist", name)
		}

		contentType := bucket.Get([]byte("type"))
		if contentType == nil {
			return fmt.Errorf("%s doesn't exist", name)
		}
		strContentType = string(contentType)

		content = bucket.Get([]byte("content"))
		if content == nil {
			return fmt.Errorf("%s doesn't exist", name)
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return content, strContentType, nil
}

// HasResource checks if the resource exists in archive.
func (arc *Archive) HasResource(name string) bool {
	// Make sure name exists
	if name == "" {
		name = "archive-root"
	}

	var exists bool
	arc.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		exists = bucket != nil
		return nil
	})

	return exists
}
