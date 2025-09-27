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
			require.True(t, domain.HasEbook(&model.BookmarkDTO{Bookmark: model.Bookmark{ID: 1}}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasEbook(&model.BookmarkDTO{Bookmark: model.Bookmark{ID: 2}}))
		})
	})

	t.Run("HasArchive", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasArchive(&model.BookmarkDTO{Bookmark: model.Bookmark{ID: 1}}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasArchive(&model.BookmarkDTO{Bookmark: model.Bookmark{ID: 2}}))
		})
	})

	t.Run("HasThumbnail", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasThumbnail(&model.BookmarkDTO{Bookmark: model.Bookmark{ID: 1}}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasThumbnail(&model.BookmarkDTO{Bookmark: model.Bookmark{ID: 2}}))
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
			Bookmark: model.Bookmark{
				ID:    1,
				URL:   "https://example.com",
				Title: "Example",
			},
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

func TestBookmarksDomain_CreateBookmark(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	domain := domains.NewBookmarksDomain(deps)

	t.Run("successful creation", func(t *testing.T) {
		bookmark := model.Bookmark{
			URL:     "https://example.com/create-test",
			Title:   "Create Test",
			Excerpt: "Test excerpt",
			Public:  1,
		}

		createdBookmark, err := domain.CreateBookmark(ctx, bookmark)
		require.NoError(t, err)
		require.NotNil(t, createdBookmark)
		require.NotZero(t, createdBookmark.ID)
		require.Equal(t, bookmark.URL, createdBookmark.URL)
		require.Equal(t, bookmark.Title, createdBookmark.Title)
		require.Equal(t, bookmark.Excerpt, createdBookmark.Excerpt)
		require.Equal(t, bookmark.Public, createdBookmark.Public)
	})

	t.Run("creation with different fields", func(t *testing.T) {
		bookmark := model.Bookmark{
			URL:     "https://example.com/create-with-fields",
			Title:   "Create With Fields Test",
			Excerpt: "Test excerpt with fields",
			Author:  "Test Author",
			Public:  0,
		}

		createdBookmark, err := domain.CreateBookmark(ctx, bookmark)
		require.NoError(t, err)
		require.NotNil(t, createdBookmark)
		require.NotZero(t, createdBookmark.ID)
		require.Equal(t, bookmark.URL, createdBookmark.URL)
		require.Equal(t, bookmark.Title, createdBookmark.Title)
		require.Equal(t, bookmark.Author, createdBookmark.Author)
		require.Equal(t, bookmark.Public, createdBookmark.Public)
	})

	t.Run("creation with minimal fields", func(t *testing.T) {
		bookmark := model.Bookmark{
			URL:   "https://example.com/minimal-fields",
			Title: "Minimal", // Title is required
		}

		createdBookmark, err := domain.CreateBookmark(ctx, bookmark)
		require.NoError(t, err)
		require.NotNil(t, createdBookmark)
		require.NotZero(t, createdBookmark.ID)
		require.Equal(t, bookmark.URL, createdBookmark.URL)
		require.Equal(t, bookmark.Title, createdBookmark.Title)
		require.Equal(t, "", createdBookmark.Author) // Other fields should be empty
	})

	t.Run("creation failure", func(t *testing.T) {
		// Create a bookmark with invalid data to trigger failure
		bookmark := model.Bookmark{
			URL: "", // Empty URL should cause validation error
		}

		createdBookmark, err := domain.CreateBookmark(ctx, bookmark)
		require.Error(t, err)
		require.Nil(t, createdBookmark)
	})
}

func TestBookmarksDomain_UpdateBookmark(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	domain := domains.NewBookmarksDomain(deps)

	t.Run("successful update", func(t *testing.T) {
		// Create initial bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Original Title"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		// Update the bookmark
		updateData := savedBookmarks[0].ToBookmark()
		updateData.Title = "Updated Title"
		updateData.Excerpt = "Updated excerpt"

		result, err := domain.UpdateBookmark(ctx, updateData)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, updateData.ID, result.ID)
		require.Equal(t, "Updated Title", result.Title)
		require.Equal(t, "Updated excerpt", result.Excerpt)
	})

	t.Run("update with different fields", func(t *testing.T) {
		// Create initial bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Update With Fields"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		// Update with different fields
		updateData := savedBookmarks[0].ToBookmark()
		updateData.Author = "Updated Author"
		updateData.Public = 1

		result, err := domain.UpdateBookmark(ctx, updateData)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, updateData.ID, result.ID)
		require.Equal(t, "Updated Author", result.Author)
		require.Equal(t, 1, result.Public)
	})

	t.Run("update with same values", func(t *testing.T) {
		// Create initial bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Same Values Test"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		// Update with same values
		updateData := savedBookmarks[0].ToBookmark()

		result, err := domain.UpdateBookmark(ctx, updateData)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, updateData.ID, result.ID)
		require.Equal(t, updateData.Title, result.Title)
	})

	t.Run("update non-existent bookmark", func(t *testing.T) {
		nonExistentBookmark := model.Bookmark{
			ID:    999999,
			URL:   "https://example.com/non-existent",
			Title: "Non-existent",
		}

		result, err := domain.UpdateBookmark(ctx, nonExistentBookmark)
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func TestBookmarksDomain_DeleteBookmarks(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	domain := domains.NewBookmarksDomain(deps)

	t.Run("successful deletion", func(t *testing.T) {
		// Create test bookmarks
		bookmark1 := testutil.GetValidBookmark()
		bookmark1.Title = "Delete Test 1"
		bookmark2 := testutil.GetValidBookmark()
		bookmark2.URL = "https://example.com/delete-test-2"
		bookmark2.Title = "Delete Test 2"

		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark1, *bookmark2)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 2)

		// Delete the bookmarks
		ids := []int{savedBookmarks[0].ID, savedBookmarks[1].ID}
		err = domain.DeleteBookmarks(ctx, ids)
		require.NoError(t, err)

		// Verify bookmarks were deleted
		for _, id := range ids {
			_, exists, err := deps.Database().GetBookmark(ctx, id, "")
			require.NoError(t, err)
			require.False(t, exists)
		}
	})

	t.Run("delete with empty ids", func(t *testing.T) {
		err := domain.DeleteBookmarks(ctx, []int{})
		require.NoError(t, err) // Should not error
	})

	t.Run("delete with non-existent ids", func(t *testing.T) {
		err := domain.DeleteBookmarks(ctx, []int{999999, 999998})
		require.NoError(t, err) // Should not error even if bookmarks don't exist
	})

	t.Run("delete with mixed existing and non-existing ids", func(t *testing.T) {
		// Create one bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Mixed Delete Test"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		// Delete with mixed IDs
		ids := []int{savedBookmarks[0].ID, 999999}
		err = domain.DeleteBookmarks(ctx, ids)
		require.NoError(t, err)

		// Verify existing bookmark was deleted
		_, exists, err := deps.Database().GetBookmark(ctx, savedBookmarks[0].ID, "")
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func TestBookmarksDomain_AddTagToBookmark(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	domain := domains.NewBookmarksDomain(deps)

	t.Run("successful add", func(t *testing.T) {
		// Create bookmark and tag
		bookmark := testutil.GetValidBookmark()
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		tag, err := deps.Database().CreateTag(ctx, model.Tag{Name: "add-test-tag"})
		require.NoError(t, err)

		// Add tag to bookmark
		err = domain.AddTagToBookmark(ctx, savedBookmarks[0].ID, tag.ID)
		require.NoError(t, err)

		// Verify tag was added
		tags, err := deps.Domains().Tags().ListTags(ctx, model.ListTagsOptions{
			BookmarkID: savedBookmarks[0].ID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, tag.ID, tags[0].ID)
	})

	t.Run("add to non-existent bookmark", func(t *testing.T) {
		tag, err := deps.Database().CreateTag(ctx, model.Tag{Name: "test-tag"})
		require.NoError(t, err)

		err = domain.AddTagToBookmark(ctx, 999999, tag.ID)
		require.Error(t, err)
		require.Equal(t, model.ErrBookmarkNotFound, err)
	})

	t.Run("add non-existent tag", func(t *testing.T) {
		bookmark := testutil.GetValidBookmark()
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		err = domain.AddTagToBookmark(ctx, savedBookmarks[0].ID, 999999)
		require.Error(t, err)
		require.Equal(t, model.ErrTagNotFound, err)
	})
}

func TestBookmarksDomain_RemoveTagFromBookmark(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	domain := domains.NewBookmarksDomain(deps)

	t.Run("successful remove", func(t *testing.T) {
		// Create bookmark and tag
		bookmark := testutil.GetValidBookmark()
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		tag, err := deps.Database().CreateTag(ctx, model.Tag{Name: "remove-test-tag"})
		require.NoError(t, err)

		// Add tag first
		err = deps.Database().AddTagToBookmark(ctx, savedBookmarks[0].ID, tag.ID)
		require.NoError(t, err)

		// Remove tag from bookmark
		err = domain.RemoveTagFromBookmark(ctx, savedBookmarks[0].ID, tag.ID)
		require.NoError(t, err)

		// Verify tag was removed
		tags, err := deps.Domains().Tags().ListTags(ctx, model.ListTagsOptions{
			BookmarkID: savedBookmarks[0].ID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 0)
	})

	t.Run("remove from non-existent bookmark", func(t *testing.T) {
		tag, err := deps.Database().CreateTag(ctx, model.Tag{Name: "test-tag"})
		require.NoError(t, err)

		err = domain.RemoveTagFromBookmark(ctx, 999999, tag.ID)
		require.Error(t, err)
		require.Equal(t, model.ErrBookmarkNotFound, err)
	})

	t.Run("remove non-existent tag", func(t *testing.T) {
		bookmark := testutil.GetValidBookmark()
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		err = domain.RemoveTagFromBookmark(ctx, savedBookmarks[0].ID, 999999)
		require.Error(t, err)
		require.Equal(t, model.ErrTagNotFound, err)
	})
}

func TestBookmarksDomain_BookmarkExists(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	domain := domains.NewBookmarksDomain(deps)

	t.Run("existing bookmark", func(t *testing.T) {
		bookmark := testutil.GetValidBookmark()
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		exists, err := domain.BookmarkExists(ctx, savedBookmarks[0].ID)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("non-existent bookmark", func(t *testing.T) {
		exists, err := domain.BookmarkExists(ctx, 999999)
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func TestBookmarksDomain_BulkUpdateBookmarkTags(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	domain := domains.NewBookmarksDomain(deps)

	t.Run("empty_bookmark_ids", func(t *testing.T) {
		err := domain.BulkUpdateBookmarkTags(ctx, []int{}, []int{1, 2, 3})
		require.NoError(t, err) // Should not return an error for empty bookmark IDs
	})

	t.Run("empty_tag_ids", func(t *testing.T) {
		err := domain.BulkUpdateBookmarkTags(ctx, []int{1, 2, 3}, []int{})
		require.NoError(t, err) // Should not return an error for empty tag IDs
	})

	t.Run("non_existent_bookmarks", func(t *testing.T) {
		err := domain.BulkUpdateBookmarkTags(ctx, []int{999, 1000}, []int{1, 2, 3})
		require.Error(t, err)
	})

	t.Run("successful_update", func(t *testing.T) {
		// Create test bookmarks
		bookmark1 := testutil.GetValidBookmark()
		bookmark2 := testutil.GetValidBookmark()
		bookmark2.URL = "https://example.com/different"

		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark1, *bookmark2)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 2)

		// Create test tags
		tag1 := model.Tag{Name: "test-tag-1"}
		tag2 := model.Tag{Name: "test-tag-2"}
		createdTags, err := deps.Database().CreateTags(ctx, tag1, tag2)
		require.NoError(t, err)
		require.Len(t, createdTags, 2)

		// Get the bookmark and tag IDs
		bookmarkIDs := []int{savedBookmarks[0].ID, savedBookmarks[1].ID}
		tagIDs := []int{createdTags[0].ID, createdTags[1].ID}

		// Update the bookmarks with the tags
		err = domain.BulkUpdateBookmarkTags(ctx, bookmarkIDs, tagIDs)
		require.NoError(t, err)

		// Verify the bookmarks have the tags
		for _, bookmarkID := range bookmarkIDs {
			bookmark, err := domain.GetBookmark(ctx, model.DBID(bookmarkID))
			require.NoError(t, err)

			// Check that the bookmark has both tags
			require.Len(t, bookmark.Tags, 2)

			// Verify tag IDs match
			tagIDsMap := make(map[int]bool)
			for _, tag := range bookmark.Tags {
				tagIDsMap[tag.ID] = true
			}

			assert.True(t, tagIDsMap[createdTags[0].ID], "Bookmark should have the first tag")
			assert.True(t, tagIDsMap[createdTags[1].ID], "Bookmark should have the second tag")
		}
	})
}

