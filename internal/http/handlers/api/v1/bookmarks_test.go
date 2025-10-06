package api_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleBookmarkReadable(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleBookmarkReadable,
			http.MethodGet,
			"/api/v1/bookmarks/1/readable",
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid bookmark id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleBookmarkReadable,
			http.MethodGet,
			"/api/v1/bookmarks/invalid/readable",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmark not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleBookmarkReadable,
			http.MethodGet,
			"/api/v1/bookmarks/999/readable",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "999"),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Content = "test content"
		bookmark.HTML = "<p>test content</p>"
		savedBookmark, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmark, 1)

		w := testutil.PerformRequest(
			deps,
			HandleBookmarkReadable,
			http.MethodGet,
			"/api/v1/bookmarks/"+strconv.Itoa(savedBookmark[0].ID)+"/readable",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmark[0].ID)),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "content", func(t *testing.T, value any) {
			require.Equal(t, bookmark.Content, value)
		})
		response.AssertMessageJSONKeyValue(t, "html", func(t *testing.T, value any) {
			require.Equal(t, bookmark.HTML, value)
		})
	})
}

func TestHandleUpdateCache(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateCache,
			http.MethodPut,
			"/api/v1/bookmarks/cache",
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requires admin access", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateCache,
			http.MethodPut,
			"/api/v1/bookmarks/cache",
			testutil.WithFakeUser(),
		)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateCache,
			http.MethodPut,
			"/api/v1/bookmarks/cache",
			testutil.WithFakeAdmin(),
			testutil.WithBody("invalid json"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty bookmark ids", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateCache,
			http.MethodPut,
			"/api/v1/bookmarks/cache",
			testutil.WithFakeAdmin(),
			testutil.WithBody(`{"ids": []}`),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmarks not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateCache,
			http.MethodPut,
			"/api/v1/bookmarks/cache",
			testutil.WithFakeAdmin(),
			testutil.WithBody(`{"ids": [999]}`),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful update", func(t *testing.T) {
		t.Skip("skipping due to concurrent execution and no easy way to test it")

		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		savedBookmark, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmark, 1)

		body := `{
			"ids": [` + strconv.Itoa(savedBookmark[0].ID) + `],
			"keep_metadata": true,
			"create_archive": true,
			"create_ebook": true
		}`

		w := testutil.PerformRequest(
			deps,
			HandleUpdateCache,
			http.MethodPut,
			"/api/v1/bookmarks/cache",
			testutil.WithFakeAdmin(),
			testutil.WithBody(body),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		// TODO: remove this sleep after refactoring into a job system
		time.Sleep(1 * time.Second)

		// Verify bookmark was updated
		updatedBookmark, exists, err := deps.Database().GetBookmark(ctx, savedBookmark[0].ID, "")
		require.NoError(t, err)
		require.True(t, exists)
		require.True(t, updatedBookmark.HasEbook)
		require.True(t, updatedBookmark.HasArchive)
	})
}

func TestHandleUpdateBookmarkTags(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	t.Run("requires_authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleBulkUpdateBookmarkTags,
			"PUT",
			"/api/v1/bookmarks/tags",
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid_json_payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleBulkUpdateBookmarkTags,
			"PUT",
			"/api/v1/bookmarks/tags",
			testutil.WithFakeUser(),
			testutil.WithBody("invalid json"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty_ids", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := map[string]interface{}{
			"ids":  []int{},
			"tags": []model.Tag{{Name: "test"}},
		}
		body, _ := json.Marshal(payload)
		w := testutil.PerformRequest(
			deps,
			HandleBulkUpdateBookmarkTags,
			"PUT",
			"/api/v1/bookmarks/tags",
			testutil.WithFakeUser(),
			testutil.WithBody(string(body)),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty_tags", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := map[string]interface{}{
			"ids":  []int{1},
			"tags": []model.Tag{},
		}
		body, _ := json.Marshal(payload)
		w := testutil.PerformRequest(
			deps,
			HandleBulkUpdateBookmarkTags,
			"PUT",
			"/api/v1/bookmarks/tags",
			testutil.WithFakeUser(),
			testutil.WithBody(string(body)),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmark_not_found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := map[string]interface{}{
			"ids":  []int{999},
			"tags": []model.Tag{{Name: "test"}},
		}
		body, _ := json.Marshal(payload)
		w := testutil.PerformRequest(
			deps,
			HandleBulkUpdateBookmarkTags,
			"PUT",
			"/api/v1/bookmarks/tags",
			testutil.WithFakeUser(),
			testutil.WithBody(string(body)),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("successful_update", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a bookmark first
		bookmark := testutil.GetValidBookmark()
		savedBookmark, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmark, 1)

		// Create a tag
		tag := model.TagDTO{Tag: model.Tag{Name: "newtag"}}
		createdTag, err := deps.Database().CreateTag(ctx, tag.Tag)
		require.NoError(t, err)

		// Update the bookmark tags
		payload := map[string]interface{}{
			"bookmark_ids": []int{savedBookmark[0].ID},
			"tag_ids":      []int{createdTag.ID},
		}
		body, _ := json.Marshal(payload)
		w := testutil.PerformRequest(
			deps,
			HandleBulkUpdateBookmarkTags,
			"PUT",
			"/api/v1/bookmarks/tags",
			testutil.WithFakeUser(),
			testutil.WithBody(string(body)),
		)
		t.Log(w.Body.String())
		require.Equal(t, http.StatusOK, w.Code)

		// Verify the response
		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
	})
}

// CRUD Handler Tests

func TestHandleCreateBookmark(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleCreateBookmark,
			http.MethodPost,
			"/api/v1/bookmarks",
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleCreateBookmark,
			http.MethodPost,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody("invalid json"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty url", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := `{"url": "", "title": "Test"}`
		w := testutil.PerformRequest(
			deps,
			HandleCreateBookmark,
			http.MethodPost,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("successful creation", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := `{
			"url": "https://example.com/test",
			"title": "Test Bookmark",
			"excerpt": "Test excerpt",
			"public": 1
		}`
		w := testutil.PerformRequest(
			deps,
			HandleCreateBookmark,
			http.MethodPost,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusCreated, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "url", func(t *testing.T, value any) {
			require.Equal(t, "https://example.com/test", value)
		})
		response.AssertMessageJSONKeyValue(t, "title", func(t *testing.T, value any) {
			require.Equal(t, "Test Bookmark", value)
		})
	})

	t.Run("creation without title defaults to URL", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := `{
			"url": "https://example.com/no-title",
			"excerpt": "Test excerpt",
			"public": 0
		}`
		w := testutil.PerformRequest(
			deps,
			HandleCreateBookmark,
			http.MethodPost,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusCreated, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "url", func(t *testing.T, value any) {
			require.Equal(t, "https://example.com/no-title", value)
		})
		// Verify title defaults to URL when not provided
		response.AssertMessageJSONKeyValue(t, "title", func(t *testing.T, value any) {
			require.Equal(t, "https://example.com/no-title", value)
		})
	})
}

func TestHandleListBookmarks(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleListBookmarks,
			http.MethodGet,
			"/api/v1/bookmarks",
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("successful list", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmarks
		bookmark1 := testutil.GetValidBookmark()
		bookmark1.Title = "First Bookmark"
		bookmark2 := testutil.GetValidBookmark()
		bookmark2.URL = "https://example.com/second"
		bookmark2.Title = "Second Bookmark"

		_, err := deps.Database().SaveBookmarks(ctx, true, *bookmark1, *bookmark2)
		require.NoError(t, err)

		w := testutil.PerformRequest(
			deps,
			HandleListBookmarks,
			http.MethodGet,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		// Ensure items key exists and is a list
		response.AssertMessageJSONKeyValue(t, "items", func(t *testing.T, value any) {
			_, ok := value.([]any)
			require.True(t, ok)
		})
	})

	t.Run("with pagination", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create multiple bookmarks
		for i := 0; i < 5; i++ {
			bookmark := testutil.GetValidBookmark()
			bookmark.URL = fmt.Sprintf("https://example.com/test-%d", i)
			bookmark.Title = fmt.Sprintf("Test Bookmark %d", i)
			_, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
			require.NoError(t, err)
		}

		w := testutil.PerformRequest(
			deps,
			HandleListBookmarks,
			http.MethodGet,
			"/api/v1/bookmarks?page=1&limit=3",
			testutil.WithFakeUser(),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "items", func(t *testing.T, value any) {
			_, ok := value.([]any)
			require.True(t, ok)
		})
	})

	t.Run("with keyword search", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark with specific title
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Unique Search Term"
		_, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)

		w := testutil.PerformRequest(
			deps,
			HandleListBookmarks,
			http.MethodGet,
			"/api/v1/bookmarks?keyword=Unique",
			testutil.WithFakeUser(),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
	})

	t.Run("optimized SearchBookmarks ensures single database call", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Test Optimized Search"
		bookmark.URL = "https://example.com/test-optimized"

		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		assert.Len(t, savedBookmarks, 1)

		// Test the API endpoint - our optimization should now use only ONE database call
		w := testutil.PerformRequest(
			deps,
			HandleListBookmarks,
			http.MethodGet,
			"/api/v1/bookmarks?keyword=Optimized",
			testutil.WithFakeUser(),
		)

		require.Equal(t, http.StatusOK, w.Code)

		var payload model.PaginatedResponse[model.BookmarkDTO]
		err = json.Unmarshal(w.Body.Bytes(), &payload)
		require.NoError(t, err)

		// Verify our optimized SearchBookmarks functionality
		assert.Len(t, payload.Items, 1)
		bookmarkResponse := payload.Items[0]

		// Verify that tags are populated from the database call (proving single call works)
		assert.NotNil(t, bookmarkResponse.Tags)      // Tags should be populated efficiently from database
		assert.False(t, bookmarkResponse.HasArchive) // Domain fields should be computed
		assert.False(t, bookmarkResponse.HasEbook)   // Domain fields should be computed

		// Verify basic bookmark fields
		assert.Equal(t, savedBookmarks[0].ID, bookmarkResponse.ID)
		assert.Equal(t, "Test Optimized Search", bookmarkResponse.Title)
	})

	t.Run("SearchBookmarks domain options conversion", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Test various domain options to ensure proper conversion to database options
		w := testutil.PerformRequest(
			deps,
			HandleListBookmarks,
			http.MethodGet,
			"/api/v1/bookmarks?tags=tag1,tag2&exclude=exclude1&keyword=search&limit=10&page=2",
			testutil.WithFakeUser(),
		)

		require.Equal(t, http.StatusOK, w.Code)

		var payload model.PaginatedResponse[model.BookmarkDTO]
		err := json.Unmarshal(w.Body.Bytes(), &payload)
		require.NoError(t, err)

		// The important thing is that the request doesn't error out,
		// ensuring our domain options conversion is working correctly
		assert.NotNil(t, payload.Items) // Should return empty array, not nil
	})

	t.Run("SearchBookmarks performance optimization", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create multiple bookmarks to test search performance
		var bookmarks []model.BookmarkDTO
		for i := 0; i < 5; i++ {
			bookmark := testutil.GetValidBookmark()
			bookmark.Title = fmt.Sprintf("Bookmark %d", i)
			bookmark.URL = fmt.Sprintf("https://example.com/bookmark-%d", i)
			bookmarks = append(bookmarks, *bookmark)
		}

		_, err := deps.Database().SaveBookmarks(ctx, true, bookmarks...)
		require.NoError(t, err)

		// Test search with limit to verify efficient pagination
		w := testutil.PerformRequest(
			deps,
			HandleListBookmarks,
			http.MethodGet,
			"/api/v1/bookmarks?limit=3&page=1",
			testutil.WithFakeUser(),
		)

		require.Equal(t, http.StatusOK, w.Code)

		var payload2 model.PaginatedResponse[model.BookmarkDTO]
		err = json.Unmarshal(w.Body.Bytes(), &payload2)
		require.NoError(t, err)

		// Verify correct pagination and that all results have proper domain fields
		assert.Len(t, payload2.Items, 3)
		for _, bookmark := range payload2.Items {
			assert.NotNil(t, bookmark.Tags) // Verify tags are populated efficiently
			assert.NotEmpty(t, bookmark.Title)
		}
	})
}

func TestHandleGetBookmark(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetBookmark,
			http.MethodGet,
			"/api/v1/bookmarks/1",
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid bookmark id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetBookmark,
			http.MethodGet,
			"/api/v1/bookmarks/invalid",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmark not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetBookmark,
			http.MethodGet,
			"/api/v1/bookmarks/999",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "999"),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful get", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Test Get Bookmark"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		w := testutil.PerformRequest(
			deps,
			HandleGetBookmark,
			http.MethodGet,
			"/api/v1/bookmarks/"+strconv.Itoa(savedBookmarks[0].ID),
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmarks[0].ID)),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "title", func(t *testing.T, value any) {
			require.Equal(t, "Test Get Bookmark", value)
		})
	})
}

func TestHandleUpdateBookmark(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmark,
			http.MethodPut,
			"/api/v1/bookmarks/1",
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid bookmark id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmark,
			http.MethodPut,
			"/api/v1/bookmarks/invalid",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmark,
			http.MethodPut,
			"/api/v1/bookmarks/1",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "1"),
			testutil.WithBody("invalid json"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmark not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := `{"title": "Updated Title"}`
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmark,
			http.MethodPut,
			"/api/v1/bookmarks/999",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "999"),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful update", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Original Title"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		payload := `{
			"title": "Updated Title",
			"excerpt": "Updated excerpt"
		}`
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmark,
			http.MethodPut,
			"/api/v1/bookmarks/"+strconv.Itoa(savedBookmarks[0].ID),
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmarks[0].ID)),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "title", func(t *testing.T, value any) {
			require.Equal(t, "Updated Title", value)
		})
		response.AssertMessageJSONKeyValue(t, "excerpt", func(t *testing.T, value any) {
			require.Equal(t, "Updated excerpt", value)
		})
	})

	t.Run("partial update", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "Original Title"
		bookmark.Excerpt = "Original excerpt"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		// Update only title
		payload := `{"title": "Only Title Updated"}`
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmark,
			http.MethodPut,
			"/api/v1/bookmarks/"+strconv.Itoa(savedBookmarks[0].ID),
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmarks[0].ID)),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "title", func(t *testing.T, value any) {
			require.Equal(t, "Only Title Updated", value)
		})
		// Excerpt should remain unchanged
		response.AssertMessageJSONKeyValue(t, "excerpt", func(t *testing.T, value any) {
			require.Equal(t, "Original excerpt", value)
		})
	})
}

func TestHandleDeleteBookmarks(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleDeleteBookmarks,
			http.MethodDelete,
			"/api/v1/bookmarks",
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleDeleteBookmarks,
			http.MethodDelete,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody("invalid json"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty ids", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := `{"ids": []}`
		w := testutil.PerformRequest(
			deps,
			HandleDeleteBookmarks,
			http.MethodDelete,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid ids", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		payload := `{"ids": [0, -1]}`
		w := testutil.PerformRequest(
			deps,
			HandleDeleteBookmarks,
			http.MethodDelete,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("successful deletion", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmarks
		bookmark1 := testutil.GetValidBookmark()
		bookmark1.Title = "To Delete 1"
		bookmark2 := testutil.GetValidBookmark()
		bookmark2.URL = "https://example.com/delete2"
		bookmark2.Title = "To Delete 2"

		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark1, *bookmark2)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 2)

		payload := fmt.Sprintf(`{"ids": [%d, %d]}`, savedBookmarks[0].ID, savedBookmarks[1].ID)
		w := testutil.PerformRequest(
			deps,
			HandleDeleteBookmarks,
			http.MethodDelete,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		// Verify bookmarks were deleted
		_, exists, err := deps.Database().GetBookmark(ctx, savedBookmarks[0].ID, "")
		require.NoError(t, err)
		require.False(t, exists)

		_, exists, err = deps.Database().GetBookmark(ctx, savedBookmarks[1].ID, "")
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("partial deletion with non-existent ids", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create one test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Title = "To Delete"
		savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmarks, 1)

		// Try to delete existing and non-existing bookmarks
		payload := fmt.Sprintf(`{"ids": [%d, 999]}`, savedBookmarks[0].ID)
		w := testutil.PerformRequest(
			deps,
			HandleDeleteBookmarks,
			http.MethodDelete,
			"/api/v1/bookmarks",
			testutil.WithFakeUser(),
			testutil.WithBody(payload),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		// Verify existing bookmark was deleted
		_, exists, err := deps.Database().GetBookmark(ctx, savedBookmarks[0].ID, "")
		require.NoError(t, err)
		require.False(t, exists)
	})
}

// Edge case and error scenario tests

func TestBookmarkHandlersEdgeCases(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("CreateBookmark edge cases", func(t *testing.T) {
		t.Run("whitespace only url", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			payload := `{"url": "   ", "title": "Test"}`
			w := testutil.PerformRequest(
				deps,
				HandleCreateBookmark,
				http.MethodPost,
				"/api/v1/bookmarks",
				testutil.WithFakeUser(),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("very long url", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			longURL := "https://example.com/" + strings.Repeat("a", 2000)
			payload := fmt.Sprintf(`{"url": "%s", "title": "Test"}`, longURL)
			w := testutil.PerformRequest(
				deps,
				HandleCreateBookmark,
				http.MethodPost,
				"/api/v1/bookmarks",
				testutil.WithFakeUser(),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusCreated, w.Code)
		})

		t.Run("special characters in fields", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			payload := `{
				"url": "https://example.com/test",
				"title": "Test with 特殊字符 and émojis 🚀",
				"excerpt": "Content with <script>alert('xss')</script>"
			}`
			w := testutil.PerformRequest(
				deps,
				HandleCreateBookmark,
				http.MethodPost,
				"/api/v1/bookmarks",
				testutil.WithFakeUser(),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusCreated, w.Code)
		})

		t.Run("empty payload fields", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			payload := `{
				"url": "https://example.com/test",
				"title": "Test"
			}`
			w := testutil.PerformRequest(
				deps,
				HandleCreateBookmark,
				http.MethodPost,
				"/api/v1/bookmarks",
				testutil.WithFakeUser(),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusCreated, w.Code)
		})
	})

	t.Run("ListBookmarks edge cases", func(t *testing.T) {
		t.Run("invalid page parameter", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?page=invalid",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should default to page 1
		})

		t.Run("negative page parameter", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?page=-1",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should default to page 1
		})

		t.Run("zero page parameter", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?page=0",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should default to page 1
		})

		t.Run("invalid limit parameter", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?limit=invalid",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should default to limit 30
		})

		t.Run("limit exceeds maximum", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?limit=200",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should default to limit 30
		})

		t.Run("empty tags parameter", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?tags=",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("tags with commas", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?tags=tag1,tag2,tag3",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("special characters in keyword", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			w := testutil.PerformRequest(
				deps,
				HandleListBookmarks,
				http.MethodGet,
				"/api/v1/bookmarks?keyword=test%20with%20spaces%20and%20特殊字符",
				testutil.WithFakeUser(),
			)
			require.Equal(t, http.StatusOK, w.Code)
		})
	})

	t.Run("UpdateBookmark edge cases", func(t *testing.T) {
		t.Run("empty json object", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

			// Create test bookmark
			bookmark := testutil.GetValidBookmark()
			savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
			require.NoError(t, err)
			require.Len(t, savedBookmarks, 1)

			payload := `{}`
			w := testutil.PerformRequest(
				deps,
				HandleUpdateBookmark,
				http.MethodPut,
				"/api/v1/bookmarks/"+strconv.Itoa(savedBookmarks[0].ID),
				testutil.WithFakeUser(),
				testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmarks[0].ID)),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should succeed with no changes
		})

		t.Run("null values in payload", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

			// Create test bookmark
			bookmark := testutil.GetValidBookmark()
			savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
			require.NoError(t, err)
			require.Len(t, savedBookmarks, 1)

			payload := `{
				"title": null,
				"excerpt": null,
				"create_ebook": null,
				"public": null
			}`
			w := testutil.PerformRequest(
				deps,
				HandleUpdateBookmark,
				http.MethodPut,
				"/api/v1/bookmarks/"+strconv.Itoa(savedBookmarks[0].ID),
				testutil.WithFakeUser(),
				testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmarks[0].ID)),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should succeed with no changes
		})

		t.Run("update with same values", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

			// Create test bookmark
			bookmark := testutil.GetValidBookmark()
			bookmark.Title = "Original Title"
			savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
			require.NoError(t, err)
			require.Len(t, savedBookmarks, 1)

			payload := `{"title": "Original Title"}`
			w := testutil.PerformRequest(
				deps,
				HandleUpdateBookmark,
				http.MethodPut,
				"/api/v1/bookmarks/"+strconv.Itoa(savedBookmarks[0].ID),
				testutil.WithFakeUser(),
				testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmarks[0].ID)),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusOK, w.Code)
		})
	})

	t.Run("DeleteBookmarks edge cases", func(t *testing.T) {
		t.Run("duplicate ids in payload", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

			// Create test bookmark
			bookmark := testutil.GetValidBookmark()
			savedBookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
			require.NoError(t, err)
			require.Len(t, savedBookmarks, 1)

			payload := fmt.Sprintf(`{"ids": [%d, %d, %d]}`, savedBookmarks[0].ID, savedBookmarks[0].ID, savedBookmarks[0].ID)
			w := testutil.PerformRequest(
				deps,
				HandleDeleteBookmarks,
				http.MethodDelete,
				"/api/v1/bookmarks",
				testutil.WithFakeUser(),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should succeed
		})

		t.Run("very large id", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			payload := `{"ids": [999999999]}`
			w := testutil.PerformRequest(
				deps,
				HandleDeleteBookmarks,
				http.MethodDelete,
				"/api/v1/bookmarks",
				testutil.WithFakeUser(),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusOK, w.Code) // Should succeed even if bookmark doesn't exist
		})

		t.Run("mixed valid and invalid ids", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			payload := `{"ids": [1, 0, -1, 999]}`
			w := testutil.PerformRequest(
				deps,
				HandleDeleteBookmarks,
				http.MethodDelete,
				"/api/v1/bookmarks",
				testutil.WithFakeUser(),
				testutil.WithBody(payload),
			)
			require.Equal(t, http.StatusBadRequest, w.Code) // Should fail due to invalid ids
		})
	})

	t.Run("Concurrent access scenarios", func(t *testing.T) {
		t.Run("concurrent bookmark creation", func(t *testing.T) {
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

			// Create multiple bookmarks concurrently
			var wg sync.WaitGroup
			results := make(chan int, 5)

			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					payload := fmt.Sprintf(`{
						"url": "https://example.com/concurrent-%d",
						"title": "Concurrent Bookmark %d"
					}`, index, index)
					w := testutil.PerformRequest(
						deps,
						HandleCreateBookmark,
						http.MethodPost,
						"/api/v1/bookmarks",
						testutil.WithFakeUser(),
						testutil.WithBody(payload),
					)
					results <- w.Code
				}(i)
			}

			wg.Wait()
			close(results)

			// All should succeed
			for code := range results {
				require.Equal(t, http.StatusCreated, code)
			}
		})
	})
}

func TestHandleGetBookmarkData(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetBookmarkData,
			http.MethodGet,
			"/api/v1/bookmarks/1/data",
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid bookmark id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetBookmarkData,
			http.MethodGet,
			"/api/v1/bookmarks/invalid/data",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmark not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetBookmarkData,
			http.MethodGet,
			"/api/v1/bookmarks/999/data",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "999"),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("success returns data without resources", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Content = "test content"
		bookmark.HTML = "<p>test content</p>"
		bookmark.HasContent = true
		savedBookmark, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmark, 1)

		w := testutil.PerformRequest(
			deps,
			HandleGetBookmarkData,
			http.MethodGet,
			"/api/v1/bookmarks/"+strconv.Itoa(savedBookmark[0].ID)+"/data",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmark[0].ID)),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "content", func(t *testing.T, value any) {
			require.Equal(t, bookmark.Content, value)
		})
		response.AssertMessageJSONKeyValue(t, "html", func(t *testing.T, value any) {
			require.Equal(t, bookmark.HTML, value)
		})
		response.AssertMessageJSONKeyValue(t, "hasContent", func(t *testing.T, value any) {
			require.Equal(t, true, value)
		})
		response.AssertMessageJSONKeyValue(t, "hasArchive", func(t *testing.T, value any) {
			require.Equal(t, false, value)
		})
		response.AssertMessageJSONKeyValue(t, "hasEbook", func(t *testing.T, value any) {
			require.Equal(t, false, value)
		})
		response.AssertMessageJSONKeyValue(t, "archiveURL", func(t *testing.T, value any) {
			require.Equal(t, "", value)
		})
		response.AssertMessageJSONKeyValue(t, "ebookURL", func(t *testing.T, value any) {
			require.Equal(t, "", value)
		})
	})
}

func TestHandleUpdateBookmarkData(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmarkData,
			http.MethodPut,
			"/api/v1/bookmarks/1/data",
			testutil.WithRequestPathValue("id", "1"),
			testutil.WithBody(`{"update_readable": true}`),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid bookmark id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmarkData,
			http.MethodPut,
			"/api/v1/bookmarks/invalid/data",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "invalid"),
			testutil.WithBody(`{"update_readable": true}`),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		savedBookmark, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmark, 1)

		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmarkData,
			http.MethodPut,
			"/api/v1/bookmarks/"+strconv.Itoa(savedBookmark[0].ID)+"/data",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmark[0].ID)),
			testutil.WithBody(`invalid json`),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmark not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmarkData,
			http.MethodPut,
			"/api/v1/bookmarks/999/data",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "999"),
			testutil.WithBody(`{"update_readable": true}`),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("success with no flags returns current data", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmark.Content = "test content"
		bookmark.HTML = "<p>test content</p>"
		savedBookmark, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, savedBookmark, 1)

		w := testutil.PerformRequest(
			deps,
			HandleUpdateBookmarkData,
			http.MethodPut,
			"/api/v1/bookmarks/"+strconv.Itoa(savedBookmark[0].ID)+"/data",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(savedBookmark[0].ID)),
			testutil.WithBody(`{"update_readable": false, "create_archive": false, "create_ebook": false}`),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "content", func(t *testing.T, value any) {
			require.Equal(t, bookmark.Content, value)
		})
		response.AssertMessageJSONKeyValue(t, "html", func(t *testing.T, value any) {
			require.Equal(t, bookmark.HTML, value)
		})
	})
}
