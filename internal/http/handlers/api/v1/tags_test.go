package api_v1

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleListTags(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(deps, HandleListTags, "GET", "/api/v1/tags")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns tags list", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a test tag
		tag := model.Tag{Name: "test-tag"}
		createdTags, err := deps.Database().CreateTags(ctx, tag)
		require.NoError(t, err)
		require.Len(t, createdTags, 1)

		w := testutil.PerformRequest(deps, HandleListTags, "GET", "/api/v1/tags", testutil.WithFakeUser())
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageIsNotEmptyList(t)
	})

	t.Run("with_bookmark_count parameter", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a test tag
		tag := model.Tag{Name: "test-tag-with-count"}
		createdTags, err := deps.Database().CreateTags(ctx, tag)
		require.NoError(t, err)
		require.Len(t, createdTags, 1)

		// Create a bookmark with this tag
		bookmark := model.BookmarkDTO{
			URL:   "https://example.com/test",
			Title: "Test Bookmark",
			Tags:  []model.TagDTO{{Tag: model.Tag{Name: tag.Name}}},
		}
		_, err = deps.Database().SaveBookmarks(ctx, true, bookmark)
		require.NoError(t, err)

		w := testutil.PerformRequest(
			deps,
			HandleListTags,
			"GET",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithRequestQueryParam("with_bookmark_count", "true"),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		response.AssertMessageIsNotEmptyList(t)

		response.ForEach(t, func(item map[string]any) {
			t.Logf("item: %+v", item)
			if tag, ok := item["name"].(string); ok {
				if tag == "test-tag-with-count" {
					require.NotZero(t, item["bookmark_count"])
				}
			}
		})
	})

	t.Run("invalid bookmark_id parameter", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		w := testutil.PerformRequest(
			deps,
			HandleListTags,
			"GET",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithRequestQueryParam("bookmark_id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bookmark_id parameter", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, bookmarks, 1)
		bookmarkID := bookmarks[0].ID

		// Create a test tag
		tag := model.Tag{Name: "test-tag-for-bookmark"}
		createdTags, err := deps.Database().CreateTags(ctx, tag)
		require.NoError(t, err)
		require.Len(t, createdTags, 1)

		// Associate the tag with the bookmark
		err = deps.Database().BulkUpdateBookmarkTags(ctx, []int{bookmarkID}, []int{createdTags[0].ID})
		require.NoError(t, err)

		w := testutil.PerformRequest(
			deps,
			HandleListTags,
			"GET",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithRequestQueryParam("bookmark_id", strconv.Itoa(bookmarkID)),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		// Verify the response contains the tag associated with the bookmark
		found := false
		response.ForEach(t, func(item map[string]any) {
			if tag, ok := item["name"].(string); ok {
				if tag == "test-tag-for-bookmark" {
					found = true
				}
			}
		})
		require.True(t, found, "The tag associated with the bookmark should be in the response")
	})

	t.Run("search parameter", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test tags with different names
		tags := []model.Tag{
			{Name: "golang"},
			{Name: "python"},
			{Name: "javascript"},
		}
		createdTags, err := deps.Database().CreateTags(ctx, tags...)
		require.NoError(t, err)
		require.Len(t, createdTags, 3)

		// Test searching for "go"
		w := testutil.PerformRequest(
			deps,
			HandleListTags,
			"GET",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithRequestQueryParam("search", "go"),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		response.AssertMessageIsNotEmptyList(t)

		found := false
		response.ForEach(t, func(item map[string]any) {
			if tag, ok := item["name"].(string); ok {
				if tag == "golang" {
					found = true
				}
			}
		})
		require.True(t, found, "Tag 'golang' should be present")

		// Test searching for "on"
		w = testutil.PerformRequest(
			deps,
			HandleListTags,
			"GET",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithRequestQueryParam("search", "on"),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response = testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		response.AssertMessageIsNotEmptyList(t)

		found = false
		response.ForEach(t, func(item map[string]any) {
			if tag, ok := item["name"].(string); ok {
				if strings.Contains(tag, "python") {
					found = true
				}
			}
		})
		require.True(t, found, "Tag 'python' should be present")
	})

	t.Run("search and bookmark_id parameters together", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a test bookmark
		bookmark := testutil.GetValidBookmark()
		bookmarks, err := deps.Database().SaveBookmarks(ctx, true, *bookmark)
		require.NoError(t, err)
		require.Len(t, bookmarks, 1)
		bookmarkID := bookmarks[0].ID

		// Test using both search and bookmark_id parameters
		w := testutil.PerformRequest(
			deps,
			HandleListTags,
			"GET",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithRequestQueryParam("search", "go"),
			testutil.WithRequestQueryParam("bookmark_id", strconv.Itoa(bookmarkID)),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertNotOk(t)

		// Verify the error message
		response.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
			require.Equal(t, "search and bookmark ID filtering cannot be used together", value)
		})
	})
}

func TestHandleGetTag(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetTag,
			"GET",
			"/api/v1/tags/1",
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid tag id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetTag,
			"GET",
			"/api/v1/tags/invalid",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("tag not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleGetTag,
			"GET",
			"/api/v1/tags/999",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "999"),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a test tag
		tag := model.Tag{Name: "test-tag"}
		createdTags, err := deps.Database().CreateTags(ctx, tag)
		require.NoError(t, err)
		require.Len(t, createdTags, 1)

		tagID := createdTags[0].ID
		w := testutil.PerformRequest(
			deps,
			HandleGetTag,
			"GET",
			"/api/v1/tags/"+strconv.Itoa(tagID),
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(tagID)),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		// Verify the tag data
		response.AssertMessageJSONKeyValue(t, "id", func(t *testing.T, value any) {
			require.Equal(t, tagID, int(value.(float64))) // TODO: Float64??
		})
		response.AssertMessageJSONKeyValue(t, "name", func(t *testing.T, value any) {
			require.Equal(t, "test-tag", value)
		})
	})
}

func TestHandleCreateTag(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(deps, HandleCreateTag, "POST", "/api/v1/tags")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleCreateTag,
			"POST",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithBody("invalid json"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty tag name", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleCreateTag,
			"POST",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithBody(`{"name": ""}`),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("successful creation", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleCreateTag,
			"POST",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithBody(`{"name": "new-test-tag"}`),
		)
		require.Equal(t, http.StatusCreated, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		// Verify the created tag
		response.AssertMessageJSONKeyValue(t, "name", func(t *testing.T, value any) {
			require.Equal(t, "new-test-tag", value)
		})
		response.AssertMessageJSONKeyValue(t, "id", func(t *testing.T, value any) {
			require.Greater(t, value.(float64), float64(0)) // TODO: Float64??
		})
	})
}

func TestHandleUpdateTag(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateTag,
			"PUT",
			"/api/v1/tags/1",
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid tag id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateTag,
			"PUT",
			"/api/v1/tags/invalid",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateTag,
			"PUT",
			"/api/v1/tags/1",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "1"),
			testutil.WithBody("invalid json"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty tag name", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateTag,
			"PUT",
			"/api/v1/tags/1",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "1"),
			testutil.WithBody(`{"name": ""}`),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("tag not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleUpdateTag,
			"PUT",
			"/api/v1/tags/999",
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", "999"),
			testutil.WithBody(`{"name": "updated-tag"}`),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful update", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a test tag
		tag := model.Tag{Name: "test-tag-for-update"}
		createdTags, err := deps.Database().CreateTags(ctx, tag)
		require.NoError(t, err)
		require.Len(t, createdTags, 1)

		tagID := createdTags[0].ID
		w := testutil.PerformRequest(
			deps,
			HandleUpdateTag,
			"PUT",
			"/api/v1/tags/"+strconv.Itoa(tagID),
			testutil.WithFakeUser(),
			testutil.WithRequestPathValue("id", strconv.Itoa(tagID)),
			testutil.WithBody(`{"name": "updated-test-tag"}`),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)

		// Verify the updated tag
		response.AssertMessageJSONKeyValue(t, "name", func(t *testing.T, value any) {
			require.Equal(t, "updated-test-tag", value)
		})

		// Ensure database was updated
		updatedTag, exists, err := deps.Database().GetTag(ctx, tagID)
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, "updated-test-tag", updatedTag.Name)
	})
}

func TestHandleDeleteTag(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleDeleteTag,
			"DELETE",
			"/api/v1/tags/1",
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requires admin privileges", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleDeleteTag,
			"DELETE",
			"/api/v1/tags/1",
			testutil.WithFakeUser(), // Regular user, not admin
			testutil.WithRequestPathValue("id", "1"),
		)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid tag id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		w := testutil.PerformRequest(
			deps,
			HandleDeleteTag,
			"DELETE",
			"/api/v1/tags/invalid",
			testutil.WithFakeAdmin(),
			testutil.WithRequestPathValue("id", "invalid"),
		)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("tag not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		w := testutil.PerformRequest(
			deps,
			HandleDeleteTag,
			"DELETE",
			"/api/v1/tags/999",
			testutil.WithFakeAdmin(),
			testutil.WithRequestPathValue("id", "999"),
		)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful deletion", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create a test tag
		tag := model.Tag{Name: "test-tag-for-deletion"}
		createdTags, err := deps.Database().CreateTags(ctx, tag)
		require.NoError(t, err)
		require.Len(t, createdTags, 1)

		tagID := createdTags[0].ID
		w := testutil.PerformRequest(
			deps,
			HandleDeleteTag,
			"DELETE",
			"/api/v1/tags/"+strconv.Itoa(tagID),
			testutil.WithFakeAdmin(),
			testutil.WithRequestPathValue("id", strconv.Itoa(tagID)),
		)
		require.Equal(t, http.StatusNoContent, w.Code)

		// Verify the tag was deleted
		_, exists, err := deps.Database().GetTag(ctx, tagID)
		require.NoError(t, err)
		require.False(t, exists)
	})
}
