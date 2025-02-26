package core_test

import (
	"bytes"
	"context"
	"os"
	fp "path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadBookImage(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("Download Images", func(t *testing.T) {
		t.Run("fails", func(t *testing.T) {
			// images is too small with unsupported format with a valid URL
			imageURL := "https://github.com/go-shiori/shiori/blob/master/internal/view/assets/res/apple-touch-icon-152x152.png"
			tmpDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			dstFile := fp.Join(tmpDir, "image.png")

			// Act
			err = core.DownloadBookImage(deps, imageURL, dstFile)

			// Assert
			assert.EqualError(t, err, "unsupported image type")
			assert.False(t, deps.Domains().Storage().FileExists(dstFile))
		})
		t.Run("successful download image", func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))
			// Arrange
			imageURL := "https://raw.githubusercontent.com/go-shiori/shiori/master/docs/assets/screenshots/cover.png"
			dstFile := "." + string(fp.Separator) + "cover.png"

			// Act
			err = core.DownloadBookImage(deps, imageURL, dstFile)

			// Assert
			assert.NoError(t, err)
			assert.True(t, deps.Domains().Storage().FileExists(dstFile))
		})
		t.Run("successful download medium size image", func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))

			// Arrange
			imageURL := "https://raw.githubusercontent.com/go-shiori/shiori/master/testdata/medium_image.png"
			dstFile := "." + string(fp.Separator) + "medium_image.png"

			// Act
			err = core.DownloadBookImage(deps, imageURL, dstFile)

			// Assert
			assert.NoError(t, err)
			assert.True(t, deps.Domains().Storage().FileExists(dstFile))
		})
	})
}

func TestProcessBookmark(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("ProcessRequest with sucssesful result", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Run("Normal without image", func(t *testing.T) {
			bookmark := model.BookmarkDTO{
				ID:            1,
				URL:           "https://example.com",
				Title:         "Example",
				Excerpt:       "This is an example article",
				CreateEbook:   true,
				CreateArchive: true,
			}
			content := bytes.NewBufferString("<html><head></head><body><p>This is an example article</p></body></html>")
			request := core.ProcessRequest{
				Bookmark:    bookmark,
				Content:     content,
				ContentType: "text/html",
				DataDir:     tmpDir,
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(deps, request)

			if expected.ID != bookmark.ID {
				t.Errorf("Unexpected ID: got %v, want %v", expected.ID, bookmark.ID)
			}
			if expected.URL != bookmark.URL {
				t.Errorf("Unexpected URL: got %v, want %v", expected.URL, bookmark.URL)
			}
			if expected.Title != bookmark.Title {
				t.Errorf("Unexpected Title: got %v, want %v", expected.Title, bookmark.Title)
			}
			if expected.Excerpt != bookmark.Excerpt {
				t.Errorf("Unexpected Excerpt: got %v, want %v", expected.Excerpt, bookmark.Excerpt)
			}
		})
		t.Run("Normal with multipleimage", func(t *testing.T) {
			tmpDir := t.TempDir()
			html := `html<html>
		  <head>
		    <meta property="og:image" content="http://example.com/image1.jpg">
		    <meta property="og:image" content="http://example.com/image2.jpg">
		    <link rel="icon" type="image/png" href="http://example.com/favicon.png">
		  </head>
		  <body>
		    <p>This is an example article</p>
		  </body>
		</html>`
			bookmark := model.BookmarkDTO{
				ID:            1,
				URL:           "https://example.com",
				Title:         "Example",
				Excerpt:       "This is an example article",
				CreateEbook:   true,
				CreateArchive: true,
			}
			content := bytes.NewBufferString(html)
			request := core.ProcessRequest{
				Bookmark:    bookmark,
				Content:     content,
				ContentType: "text/html",
				DataDir:     tmpDir,
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(deps, request)

			if expected.ID != bookmark.ID {
				t.Errorf("Unexpected ID: got %v, want %v", expected.ID, bookmark.ID)
			}
			if expected.URL != bookmark.URL {
				t.Errorf("Unexpected URL: got %v, want %v", expected.URL, bookmark.URL)
			}
			if expected.Title != bookmark.Title {
				t.Errorf("Unexpected Title: got %v, want %v", expected.Title, bookmark.Title)
			}
			if expected.Excerpt != bookmark.Excerpt {
				t.Errorf("Unexpected Excerpt: got %v, want %v", expected.Excerpt, bookmark.Excerpt)
			}
		})
		t.Run("ProcessRequest sucssesful with multipleimage included favicon and Thumbnail ", func(t *testing.T) {
			tmpDir := t.TempDir()
			html := `html<html>
  			<head>
    		<meta property="og:image" content="http://example.com/image1.jpg">
    		<meta property="og:image" content="https://raw.githubusercontent.com/go-shiori/shiori/master/testdata/big_image.png">
    		<link rel="icon" type="image/svg" href="https://raw.githubusercontent.com/go-shiori/shiori/master/testdata/favicon.svg">
  			</head>
  			<body>
    			<p>This is an example article</p>
  			</body>
			</html>`
			bookmark := model.BookmarkDTO{
				ID:            1,
				URL:           "https://example.com",
				Title:         "Example",
				Excerpt:       "This is an example article",
				CreateEbook:   true,
				CreateArchive: true,
			}
			content := bytes.NewBufferString(html)
			request := core.ProcessRequest{
				Bookmark:    bookmark,
				Content:     content,
				ContentType: "text/html",
				DataDir:     tmpDir,
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(deps, request)
			assert.True(t, deps.Domains().Storage().FileExists(fp.Join("thumb", "1")))
			if expected.ID != bookmark.ID {
				t.Errorf("Unexpected ID: got %v, want %v", expected.ID, bookmark.ID)
			}
			if expected.URL != bookmark.URL {
				t.Errorf("Unexpected URL: got %v, want %v", expected.URL, bookmark.URL)
			}
			if expected.Title != bookmark.Title {
				t.Errorf("Unexpected Title: got %v, want %v", expected.Title, bookmark.Title)
			}
			if expected.Excerpt != bookmark.Excerpt {
				t.Errorf("Unexpected Excerpt: got %v, want %v", expected.Excerpt, bookmark.Excerpt)
			}
		})
		t.Run("ProcessRequest sucssesful with empty title ", func(t *testing.T) {
			tmpDir := t.TempDir()
			bookmark := model.BookmarkDTO{
				ID:            1,
				URL:           "https://example.com",
				Title:         "",
				Excerpt:       "This is an example article",
				CreateEbook:   true,
				CreateArchive: true,
			}
			content := bytes.NewBufferString("<html><head></head><body><p>This is an example article</p></body></html>")
			request := core.ProcessRequest{
				Bookmark:    bookmark,
				Content:     content,
				ContentType: "text/html",
				DataDir:     tmpDir,
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(deps, request)

			if expected.ID != bookmark.ID {
				t.Errorf("Unexpected ID: got %v, want %v", expected.ID, bookmark.ID)
			}
			if expected.URL != bookmark.URL {
				t.Errorf("Unexpected URL: got %v, want %v", expected.URL, bookmark.URL)
			}
			if expected.Title != bookmark.URL {
				t.Errorf("Unexpected Title: got %v, want %v", expected.Title, bookmark.Title)
			}
			if expected.Excerpt != bookmark.Excerpt {
				t.Errorf("Unexpected Excerpt: got %v, want %v", expected.Excerpt, bookmark.Excerpt)
			}
		})
		t.Run("ProcessRequest sucssesful with empty Excerpt", func(t *testing.T) {
			tmpDir := t.TempDir()
			bookmark := model.BookmarkDTO{
				ID:            1,
				URL:           "https://example.com",
				Title:         "",
				Excerpt:       "This is an example article",
				CreateEbook:   true,
				CreateArchive: true,
			}
			content := bytes.NewBufferString("<html><head></head><body><p>This is an example article</p></body></html>")
			request := core.ProcessRequest{
				Bookmark:    bookmark,
				Content:     content,
				ContentType: "text/html",
				DataDir:     tmpDir,
				KeepTitle:   true,
				KeepExcerpt: false,
			}
			expected, _, _ := core.ProcessBookmark(deps, request)

			if expected.ID != bookmark.ID {
				t.Errorf("Unexpected ID: got %v, want %v", expected.ID, bookmark.ID)
			}
			if expected.URL != bookmark.URL {
				t.Errorf("Unexpected URL: got %v, want %v", expected.URL, bookmark.URL)
			}
			if expected.Title != bookmark.URL {
				t.Errorf("Unexpected Title: got %v, want %v", expected.Title, bookmark.Title)
			}
			if expected.Excerpt != bookmark.Excerpt {
				t.Errorf("Unexpected Excerpt: got %v, want %v", expected.Excerpt, bookmark.Excerpt)
			}
		})
		t.Run("Specific case", func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Run("ProcessRequest with ID zero", func(t *testing.T) {

				bookmark := model.BookmarkDTO{
					ID:            0,
					URL:           "https://example.com",
					Title:         "Example",
					Excerpt:       "This is an example article",
					CreateEbook:   true,
					CreateArchive: true,
				}
				content := bytes.NewBufferString("<html><head></head><body><p>This is an example article</p></body></html>")
				request := core.ProcessRequest{
					Bookmark:    bookmark,
					Content:     content,
					ContentType: "text/html",
					DataDir:     tmpDir,
					KeepTitle:   true,
					KeepExcerpt: true,
				}
				_, isFatal, err := core.ProcessBookmark(deps, request)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "bookmark ID is not valid")
				assert.True(t, isFatal)
			})

			t.Run("ProcessRequest that content type not zero", func(t *testing.T) {
				tmpDir := t.TempDir()
				bookmark := model.BookmarkDTO{
					ID:            1,
					URL:           "https://example.com",
					Title:         "Example",
					Excerpt:       "This is an example article",
					CreateEbook:   true,
					CreateArchive: true,
				}
				content := bytes.NewBufferString("<html><head></head><body><p>This is an example article</p></body></html>")
				request := core.ProcessRequest{
					Bookmark:    bookmark,
					Content:     content,
					ContentType: "application/pdf",
					DataDir:     tmpDir,
					KeepTitle:   true,
					KeepExcerpt: true,
				}
				_, _, err := core.ProcessBookmark(deps, request)
				assert.NoError(t, err)
			})
		})
	})
}
