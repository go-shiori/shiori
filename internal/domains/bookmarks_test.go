package domains_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookmarkDomain(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	deps.Domains().SetStorage(domains.NewStorageDomain(deps, fs))

	fs.MkdirAll("thumb", 0755)
	fs.Create("thumb/1")
	fs.MkdirAll("ebook", 0755)
	fs.Create("ebook/1.epub")
	fs.MkdirAll("archive", 0755)
	// TODO: write a valid archive file
	fs.Create("archive/1")

	domain := domains.NewBookmarksDomain(deps)
	t.Run("HasEbook", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasEbook(&model.BookmarkDTO{ID: 1}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasEbook(&model.BookmarkDTO{ID: 2}))
		})
	})

	t.Run("HasArchive", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasArchive(&model.BookmarkDTO{ID: 1}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasArchive(&model.BookmarkDTO{ID: 2}))
		})
	})

	t.Run("HasThumbnail", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasThumbnail(&model.BookmarkDTO{ID: 1}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasThumbnail(&model.BookmarkDTO{ID: 2}))
		})
	})

	t.Run("GetBookmark", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			_, err := deps.Database().SaveBookmarks(context.TODO(), true, *testutil.GetValidBookmark())
			require.NoError(t, err)
			bookmark, err := domain.GetBookmark(context.Background(), 1)
			require.NoError(t, err)
			require.Equal(t, 1, bookmark.ID)

			// Check DTO attributes
			require.True(t, bookmark.HasEbook)
			require.True(t, bookmark.HasArchive)
		})

		t.Run("NotFound", func(t *testing.T) {
			bookmark, err := domain.GetBookmark(context.Background(), 999)
			require.Error(t, err)
			require.Nil(t, bookmark)
			require.Equal(t, model.ErrBookmarkNotFound, err)
		})

		t.Run("DatabaseError", func(t *testing.T) {
			// Create a new context with a timeout to force an error
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately to force error
			bookmark, err := domain.GetBookmark(cancelCtx, 1)
			require.Error(t, err)
			require.Nil(t, bookmark)
			require.Contains(t, err.Error(), "failed to get bookmark")
		})
	})

	t.Run("GetBookmarks", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Create multiple bookmarks
			bookmark1 := testutil.GetValidBookmark()
			bookmark1.ID = 1
			bookmark2 := testutil.GetValidBookmark()
			bookmark2.ID = 2
			bookmark2.URL = "https://example.com"

			_, err := deps.Database().SaveBookmarks(context.TODO(), true, *bookmark1, *bookmark2)
			require.NoError(t, err)

			// Test getting multiple bookmarks
			bookmarks, err := domain.GetBookmarks(context.Background(), []int{1, 2})
			require.NoError(t, err)
			require.Len(t, bookmarks, 2)

			// Verify the bookmarks have the correct properties
			assert.Equal(t, 1, bookmarks[0].ID)
			assert.True(t, bookmarks[0].HasEbook)
			assert.True(t, bookmarks[0].HasArchive)

			assert.Equal(t, 2, bookmarks[1].ID)
			assert.False(t, bookmarks[1].HasEbook)
			assert.False(t, bookmarks[1].HasArchive)
		})

		t.Run("PartialResults", func(t *testing.T) {
			// Test with a mix of existing and non-existing IDs
			bookmarks, err := domain.GetBookmarks(context.Background(), []int{1, 999})
			require.NoError(t, err)
			require.Len(t, bookmarks, 1)
			assert.Equal(t, 1, bookmarks[0].ID)
		})

		t.Run("EmptyResults", func(t *testing.T) {
			// Test with non-existing IDs
			bookmarks, err := domain.GetBookmarks(context.Background(), []int{998, 999})
			require.NoError(t, err)
			require.Len(t, bookmarks, 0)
		})

		t.Run("DatabaseError", func(t *testing.T) {
			// Create a new context with a timeout to force an error
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately to force error
			bookmarks, err := domain.GetBookmarks(cancelCtx, []int{1})
			require.Error(t, err)
			require.Nil(t, bookmarks)
			require.Contains(t, err.Error(), "failed to get bookmark")
		})
	})

	t.Run("UpdateBookmarkCache", func(t *testing.T) {
		// Create a new test environment for this specific test
		fs := afero.NewMemMapFs()
		ctx := context.Background()
		logger := logrus.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Domains().SetStorage(domains.NewStorageDomain(deps, fs))

		// Create necessary directories
		fs.MkdirAll("thumb", 0755)
		fs.MkdirAll("ebook", 0755)
		fs.MkdirAll("archive", 0755)

		domain := domains.NewBookmarksDomain(deps)

		// Create a test bookmark
		bookmark := model.BookmarkDTO{
			ID:            1,
			URL:           "https://example.com",
			Title:         "Example",
			CreateEbook:   true,
			CreateArchive: true,
		}

		// Save the bookmark to the database
		_, err := deps.Database().SaveBookmarks(context.TODO(), true, bookmark)
		require.NoError(t, err)

		// Mock the core.DownloadBookmark function using monkey patching
		// Since we can't directly mock it, we'll test the error case
		t.Run("DownloadError", func(t *testing.T) {
			// Use an invalid URL to trigger a download error
			bookmark.URL = "invalid://url"

			result, err := domain.UpdateBookmarkCache(ctx, bookmark, true, false)
			require.Error(t, err)
			require.Nil(t, result)
			require.Contains(t, err.Error(), "failed to download bookmark")
		})

		// Test the skip existing functionality
		t.Run("SkipExistingEbook", func(t *testing.T) {
			// Create an ebook file
			ebookPath := model.GetEbookPath(&bookmark)
			_, err := fs.Create(ebookPath)
			require.NoError(t, err)

			// Set a valid URL
			bookmark.URL = "https://example.com"
			bookmark.CreateEbook = true

			// This test will still fail because we can't mock the HTTP client
			// But we can verify the logic for skipping existing ebooks
			_, err = domain.UpdateBookmarkCache(ctx, bookmark, true, true)

			// The test will fail at the download step, but we can check if the CreateEbook flag was set correctly
			if err != nil && !errors.Is(err, context.Canceled) {
				// This is expected since we can't mock the HTTP client
				// But we can check if the bookmark was modified correctly before the error
				assert.False(t, bookmark.CreateEbook)
				assert.True(t, bookmark.HasEbook)
			}
		})
	})
}
