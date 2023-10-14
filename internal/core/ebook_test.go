package core_test

import (
	"fmt"
	"os"
	fp "path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateEbook(t *testing.T) {
	t.Run("Successful ebook generate", func(t *testing.T) {
		t.Run("valid bookmarkId that return HasEbook true", func(t *testing.T) {
			// test cae
			tempDir := t.TempDir()
			dstDir := t.TempDir()

			mockRequest := core.ProcessRequest{
				Bookmark: model.Bookmark{
					ID:       1,
					Title:    "Example Bookmark",
					HTML:     "<html><body>Example HTML</body></html>",
					HasEbook: false,
				},
				DataDir:     dstDir,
				ContentType: "text/html",
			}

			bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))

			assert.True(t, bookmark.HasEbook)
			assert.NoError(t, err)
		})
		t.Run("ebook generate with valid BookmarkID EbookExist ImagePathExist ReturnWithHasEbookTrue", func(t *testing.T) {
			tempDir := t.TempDir()
			dstDir := t.TempDir()

			mockRequest := core.ProcessRequest{
				Bookmark: model.Bookmark{
					ID:       1,
					HasEbook: false,
				},
				DataDir:     dstDir,
				ContentType: "text/html",
			}
			// Create the image directory
			imageDir := fp.Join(mockRequest.DataDir, "thumb")
			err := os.MkdirAll(imageDir, os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}
			// Create the image file
			imagePath := fp.Join(mockRequest.DataDir, "thumb", fmt.Sprintf("%d", mockRequest.Bookmark.ID))
			file, err := os.Create(imagePath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))
			expectedimagePath := "/bookmark/1/thumb"
			if expectedimagePath != bookmark.ImageURL {
				t.Errorf("Expected imageURL %s, but got %s", bookmark.ImageURL, expectedimagePath)
			}
			assert.True(t, bookmark.HasEbook)
			assert.NoError(t, err)
		})
		t.Run("generate ebook valid BookmarkID EbookExist Returnh HasArchive True", func(t *testing.T) {

			tempDir := t.TempDir()
			dstDir := t.TempDir()

			mockRequest := core.ProcessRequest{
				Bookmark: model.Bookmark{
					ID:       1,
					HasEbook: false,
				},
				DataDir:     dstDir,
				ContentType: "text/html",
			}
			// Create the archive directory
			archiveDir := fp.Join(mockRequest.DataDir, "archive")
			err := os.MkdirAll(archiveDir, os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}
			// Create the archive file
			archivePath := fp.Join(mockRequest.DataDir, "archive", fmt.Sprintf("%d", mockRequest.Bookmark.ID))
			file, err := os.Create(archivePath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))
			assert.True(t, bookmark.HasArchive)
			assert.NoError(t, err)
		})
	})
	t.Run("specific ebook generate case", func(t *testing.T) {
		t.Run("invalid bookmarkId that return Error", func(t *testing.T) {
			tempDir := t.TempDir()
			mockRequest := core.ProcessRequest{
				Bookmark: model.Bookmark{
					ID:       0,
					HasEbook: false,
				},
				DataDir:     tempDir,
				ContentType: "text/html",
			}

			bookmark, err := core.GenerateEbook(mockRequest, tempDir)

			assert.Equal(t, model.Bookmark{
				ID:       0,
				HasEbook: false,
			}, bookmark)
			assert.Error(t, err)
		})
		t.Run("ebook exist return HasEbook true", func(t *testing.T) {
			tempDir := t.TempDir()
			dstDir := t.TempDir()

			mockRequest := core.ProcessRequest{
				Bookmark: model.Bookmark{
					ID:       1,
					HasEbook: false,
				},
				DataDir:     dstDir,
				ContentType: "text/html",
			}
			// Create the ebook directory
			ebookDir := fp.Join(mockRequest.DataDir, "ebook")
			err := os.MkdirAll(ebookDir, os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}
			// Create the ebook file
			ebookfile := fp.Join(mockRequest.DataDir, "ebook", fmt.Sprintf("%d.epub", mockRequest.Bookmark.ID))
			file, err := os.Create(ebookfile)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))

			assert.True(t, bookmark.HasEbook)
			assert.NoError(t, err)
		})
		t.Run("generate ebook valid BookmarkID RetuenError for PDF file", func(t *testing.T) {
			tempDir := t.TempDir()

			mockRequest := core.ProcessRequest{
				Bookmark: model.Bookmark{
					ID:       1,
					HasEbook: false,
				},
				DataDir:     tempDir,
				ContentType: "application/pdf",
			}

			bookmark, err := core.GenerateEbook(mockRequest, tempDir)

			assert.False(t, bookmark.HasEbook)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "can't create ebook for pdf")
		})
	})
}
