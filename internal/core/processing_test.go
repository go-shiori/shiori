package core_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	fp "path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMoveFileToDestination(t *testing.T) {
	t.Run("create  fails", func(t *testing.T) {
		t.Run("directory create fails", func(t *testing.T) {
			// test if create dir fails
			tmpFile, err := os.CreateTemp("", "image")

			assert.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			err = core.MoveFileToDestination("/destination/test", tmpFile)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to create destination dir")
		})
		t.Run("file create fails", func(t *testing.T) {
			// if create file failed
			tmpFile, err := os.CreateTemp("", "image")
			assert.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			// Create a destination directory
			dstDir := t.TempDir()
			assert.NoError(t, err)
			defer os.Remove(dstDir)

			// Set destination path to an invalid file name to force os.Create to fail
			dstPath := fp.Join(dstDir, "\000invalid\000")

			err = core.MoveFileToDestination(dstPath, tmpFile)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to create destination file")
		})
	})
}
func TestDownloadBookImage(t *testing.T) {
	t.Run("Download Images", func(t *testing.T) {
		t.Run("fails", func(t *testing.T) {
			// images is too small with unsupported format with a valid URL
			imageURL := "https://github.com/go-shiori/shiori/blob/master/internal/view/assets/res/apple-touch-icon-152x152.png"
			tempDir := t.TempDir()
			dstPath := fp.Join(tempDir, "1")
			defer os.Remove(dstPath)

			// Act
			err := core.DownloadBookImage(imageURL, dstPath)

			// Assert
			assert.EqualError(t, err, "unsupported image type")
			assert.NoFileExists(t, dstPath)
		})
		t.Run("sucssesful downlosd image", func(t *testing.T) {
			// Arrange
			imageURL := "https://raw.githubusercontent.com/go-shiori/shiori/master/docs/readme/cover.png"
			tempDir := t.TempDir()
			dstPath := fp.Join(tempDir, "1")
			defer os.Remove(dstPath)

			// Act
			err := core.DownloadBookImage(imageURL, dstPath)

			// Assert
			assert.NoError(t, err)
			assert.FileExists(t, dstPath)
		})
		t.Run("sucssesful downlosd medium size image", func(t *testing.T) {
			// create a file server handler for the 'testdata' directory
			fs := http.FileServer(http.Dir("../../testdata/"))

			// start a test server with the file server handler
			server := httptest.NewServer(fs)
			defer server.Close()

			// Arrange
			imageURL := server.URL + "/medium_image.png"
			tempDir := t.TempDir()
			dstPath := fp.Join(tempDir, "1")
			defer os.Remove(dstPath)

			// Act
			err := core.DownloadBookImage(imageURL, dstPath)

			// Assert
			assert.NoError(t, err)
			assert.FileExists(t, dstPath)

		})
	})
}

func TestProcessBookmark(t *testing.T) {
	t.Run("ProcessRequest with sucssesful result", func(t *testing.T) {
		t.Run("Normal without image", func(t *testing.T) {
			bookmark := model.Bookmark{
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
				DataDir:     "/tmp",
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(request)

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
			bookmark := model.Bookmark{
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
				DataDir:     "/tmp",
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(request)

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
			// create a file server handler for the 'testdata' directory
			fs := http.FileServer(http.Dir("../../testdata/"))

			// start a test server with the file server handler
			server := httptest.NewServer(fs)
			defer server.Close()

			html := `html<html>
  			<head>
    		<meta property="og:image" content="http://example.com/image1.jpg">
    		<meta property="og:image" content="` + server.URL + `/big_image.png">
    		<link rel="icon" type="image/svg" href="` + server.URL + `/favicon.svg">
  			</head>
  			<body>
    			<p>This is an example article</p>
  			</body>
			</html>`
			bookmark := model.Bookmark{
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
				DataDir:     "/tmp",
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(request)

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
			bookmark := model.Bookmark{
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
				DataDir:     "/tmp",
				KeepTitle:   true,
				KeepExcerpt: true,
			}
			expected, _, _ := core.ProcessBookmark(request)

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
			bookmark := model.Bookmark{
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
				DataDir:     "/tmp",
				KeepTitle:   true,
				KeepExcerpt: false,
			}
			expected, _, _ := core.ProcessBookmark(request)

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
			t.Run("ProcessRequest with ID zero", func(t *testing.T) {

				bookmark := model.Bookmark{
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
					DataDir:     "/tmp",
					KeepTitle:   true,
					KeepExcerpt: true,
				}
				_, isFatal, err := core.ProcessBookmark(request)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "bookmark ID is not valid")
				assert.True(t, isFatal)
			})

			t.Run("ProcessRequest that content type not zero", func(t *testing.T) {

				bookmark := model.Bookmark{
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
					DataDir:     "/tmp",
					KeepTitle:   true,
					KeepExcerpt: true,
				}
				_, _, err := core.ProcessBookmark(request)
				assert.NoError(t, err)
			})
		})
	})
}
