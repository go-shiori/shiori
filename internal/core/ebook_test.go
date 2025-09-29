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
			tmpDir := t.TempDir()

			deps.Domains().SetStorage(domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), tmpDir)))

			mockRequest := core.ProcessRequest{
				Bookmark: model.BookmarkDTO{
					ID:       1,
					Title:    "Example Bookmark",
					HTML:     "<html><body>Example HTML</body></html>",
					HasEbook: false,
				},
			}
			bookmark, err := core.GenerateEbook(deps, mockRequest)

			assert.True(t, bookmark.HasEbook)
			assert.NoError(t, err)
		})
		t.Run("ebook generate with valid BookmarkID EbookExist ImagePathExist ReturnWithHasEbookTrue", func(t *testing.T) {
			tmpDir := t.TempDir()

			deps.Domains().SetStorage(domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), tmpDir)))

			bookmark := model.BookmarkDTO{
				ID:       2,
				HasEbook: false,
			}
			mockRequest := core.ProcessRequest{
				Bookmark: bookmark,
			}
			// Create the thumbnail file
			imagePath := model.GetThumbnailPath(&bookmark)
			imagedirPath := fp.Dir(imagePath)
			deps.Domains().Storage().FS().MkdirAll(imagedirPath, os.ModePerm)
			file, err := deps.Domains().Storage().FS().Create(imagePath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err = core.GenerateEbook(deps, mockRequest)
			expectedImagePath := string(fp.Separator) + fp.Join("bookmark", "2", "thumb")
			assert.NoError(t, err)
			assert.True(t, bookmark.HasEbook)
			assert.Equalf(t, expectedImagePath, bookmark.ImageURL, "Expected imageURL %s, but got %s", expectedImagePath, bookmark.ImageURL)
		})
		t.Run("generate ebook valid BookmarkID EbookExist", func(t *testing.T) {
			tmpDir := t.TempDir()

			deps.Domains().SetStorage(domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), tmpDir)))

			bookmark := model.BookmarkDTO{
				ID:       3,
				HasEbook: false,
			}
			mockRequest := core.ProcessRequest{
				Bookmark: bookmark,
			}
			// Create the archive file
			archivePath := model.GetArchivePath(&bookmark)
			archiveDirPath := fp.Dir(archivePath)
			deps.Domains().Storage().FS().MkdirAll(archiveDirPath, os.ModePerm)
			file, err := deps.Domains().Storage().FS().Create(archivePath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err = core.GenerateEbook(deps, mockRequest)
			assert.NoError(t, err)
		})
	})
	t.Run("specific ebook generate case", func(t *testing.T) {
		t.Run("invalid bookmarkId that return Error", func(t *testing.T) {
			mockRequest := core.ProcessRequest{
				Bookmark: model.BookmarkDTO{
					ID:       0,
					HasEbook: false,
				},
			}

			bookmark, err := core.GenerateEbook(deps, mockRequest)

			assert.Equal(t, model.BookmarkDTO{
				ID:       0,
				HasEbook: false,
			}, bookmark)
			assert.EqualError(t, err, "bookmark ID is not valid")
		})
		t.Run("ebook exist return HasEbook true", func(t *testing.T) {
			tmpDir := t.TempDir()

			deps.Domains().SetStorage(domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), tmpDir)))

			bookmark := model.BookmarkDTO{
				ID:       1,
				HasEbook: false,
			}
			mockRequest := core.ProcessRequest{
				Bookmark: bookmark,
			}
			// Create the ebook file
			ebookPath := model.GetEbookPath(&bookmark)
			ebookDirPath := fp.Dir(ebookPath)
			deps.Domains().Storage().FS().MkdirAll(ebookDirPath, os.ModePerm)
			file, err := deps.Domains().Storage().FS().Create(ebookPath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			bookmark, err = core.GenerateEbook(deps, mockRequest)

			assert.True(t, bookmark.HasEbook)
			assert.NoError(t, err)
		})
	})
}
