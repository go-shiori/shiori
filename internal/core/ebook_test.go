package core_test

import (
	"context"
	"os"
	fp "path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGenerateEbook(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("Successful ebook generate", func(t *testing.T) {
		t.Run("valid bookmarkId that return HasEbook true", func(t *testing.T) {
			// test cae
			tempDir := t.TempDir()
			dstDir := t.TempDir()

			deps.Domains.Storage = domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), dstDir))

			mockRequest := core.ProcessRequest{
				Bookmark: model.BookmarkDTO{
					ID:       1,
					Title:    "Example Bookmark",
					HTML:     "<html><body>Example HTML</body></html>",
					HasEbook: false,
				},
				DataDir:     dstDir,
				ContentType: "text/html",
			}

			bookmark, err := core.GenerateEbook(deps, mockRequest, fp.Join(tempDir, "1"))

			assert.True(t, bookmark.HasEbook)
			assert.NoError(t, err)
		})
		t.Run("ebook generate with valid BookmarkID EbookExist ImagePathExist ReturnWithHasEbookTrue", func(t *testing.T) {
			dstDir := t.TempDir()

			deps.Domains.Storage = domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), dstDir))

			bookmark := model.BookmarkDTO{
				ID:       2,
				HasEbook: false,
			}
			mockRequest := core.ProcessRequest{
				Bookmark:    bookmark,
				DataDir:     dstDir,
				ContentType: "text/html",
			}
			// Create the thumbnail file
			imagePath := model.GetThumbnailPath(&bookmark)
			deps.Domains.Storage.FS().MkdirAll(fp.Dir(imagePath), os.ModePerm)
			file, err := deps.Domains.Storage.FS().Create(imagePath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err = core.GenerateEbook(deps, mockRequest, model.GetEbookPath(&bookmark))
			expectedImagePath := string(fp.Separator) + fp.Join("bookmark", "2", "thumb")
			assert.NoError(t, err)
			assert.True(t, bookmark.HasEbook)
			assert.Equalf(t, expectedImagePath, bookmark.ImageURL, "Expected imageURL %s, but got %s", expectedImagePath, bookmark.ImageURL)
		})
		t.Run("generate ebook valid BookmarkID EbookExist ReturnHasArchiveTrue", func(t *testing.T) {
			tempDir := t.TempDir()
			dstDir := t.TempDir()

			deps.Domains.Storage = domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), dstDir))

			bookmark := model.BookmarkDTO{
				ID:       3,
				HasEbook: false,
			}
			mockRequest := core.ProcessRequest{
				Bookmark:    bookmark,
				DataDir:     dstDir,
				ContentType: "text/html",
			}
			// Create the archive file
			archivePath := model.GetArchivePath(&bookmark)
			deps.Domains.Storage.FS().MkdirAll(fp.Dir(archivePath), os.ModePerm)
			file, err := deps.Domains.Storage.FS().Create(archivePath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err = core.GenerateEbook(deps, mockRequest, fp.Join(tempDir, "1"))
			assert.True(t, bookmark.HasArchive)
			assert.NoError(t, err)
		})
	})
	t.Run("specific ebook generate case", func(t *testing.T) {
		t.Run("invalid bookmarkId that return Error", func(t *testing.T) {
			tempDir := t.TempDir()
			mockRequest := core.ProcessRequest{
				Bookmark: model.BookmarkDTO{
					ID:       0,
					HasEbook: false,
				},
				DataDir:     tempDir,
				ContentType: "text/html",
			}

			bookmark, err := core.GenerateEbook(deps, mockRequest, tempDir)

			assert.Equal(t, model.BookmarkDTO{
				ID:       0,
				HasEbook: false,
			}, bookmark)
			assert.Error(t, err)
		})
		t.Run("ebook exist return HasEbook true", func(t *testing.T) {
			dstDir := t.TempDir()

			deps.Domains.Storage = domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), dstDir))

			bookmark := model.BookmarkDTO{
				ID:       1,
				HasEbook: false,
			}
			mockRequest := core.ProcessRequest{
				Bookmark:    bookmark,
				DataDir:     dstDir,
				ContentType: "text/html",
			}
			// Create the ebook file
			ebookFilePath := model.GetEbookPath(&bookmark)
			deps.Domains.Storage.FS().MkdirAll(fp.Dir(ebookFilePath), os.ModePerm)
			file, err := deps.Domains.Storage.FS().Create(ebookFilePath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err = core.GenerateEbook(deps, mockRequest, ebookFilePath)

			assert.True(t, bookmark.HasEbook)
			assert.NoError(t, err)
		})
		t.Run("generate ebook valid BookmarkID RetuenError for PDF file", func(t *testing.T) {
			tempDir := t.TempDir()

			mockRequest := core.ProcessRequest{
				Bookmark: model.BookmarkDTO{
					ID:       1,
					HasEbook: false,
				},
				DataDir:     tempDir,
				ContentType: "application/pdf",
			}

			bookmark, err := core.GenerateEbook(deps, mockRequest, tempDir)

			assert.False(t, bookmark.HasEbook)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "can't create ebook for pdf")
		})
	})
}
