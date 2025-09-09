package api_v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
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
