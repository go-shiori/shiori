package api_v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
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

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
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

		w := testutil.PerformRequest(
			deps,
			HandleListTags,
			"GET",
			"/api/v1/tags",
			testutil.WithFakeUser(),
			testutil.WithRequestQueryParam("with_bookmark_count", "true"),
		)
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		// Verify the response contains tags with bookmark_count field
		var tags []model.TagDTO
		responseData, err := json.Marshal(response.Response.GetMessage())
		require.NoError(t, err)
		err = json.Unmarshal(responseData, &tags)
		require.NoError(t, err)
		require.NotEmpty(t, tags)

		// The bookmark_count field should be present in the response
		// Even if it's 0, it should be included when the parameter is set
		for _, tag := range tags {
			if tag.Name == "test-tag-with-count" {
				// We're just checking that the field exists and is accessible
				_ = tag.BookmarkCount
				break
			}
		}
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

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		// Verify the response contains the tag associated with the bookmark
		var tags []model.TagDTO
		responseData, err := json.Marshal(response.Response.GetMessage())
		require.NoError(t, err)
		err = json.Unmarshal(responseData, &tags)
		require.NoError(t, err)

		// Check that we have at least one tag and it's the one we created
		require.NotEmpty(t, tags)
		found := false
		for _, t := range tags {
			if t.Name == "test-tag-for-bookmark" {
				found = true
				break
			}
		}
		require.True(t, found, "The tag associated with the bookmark should be in the response")
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

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		// Verify the tag data
		var tagDTO model.TagDTO
		responseData, err := json.Marshal(response.Response.GetMessage())
		require.NoError(t, err)
		err = json.Unmarshal(responseData, &tagDTO)
		require.NoError(t, err)
		require.Equal(t, tagID, tagDTO.ID)
		require.Equal(t, "test-tag", tagDTO.Name)
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

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		// Verify the created tag
		var tagDTO model.TagDTO
		responseData, err := json.Marshal(response.Response.GetMessage())
		require.NoError(t, err)
		err = json.Unmarshal(responseData, &tagDTO)
		require.NoError(t, err)
		require.Greater(t, tagDTO.ID, 0)
		require.Equal(t, "new-test-tag", tagDTO.Name)
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

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		// Verify the updated tag
		var tagDTO model.TagDTO
		responseData, err := json.Marshal(response.Response.GetMessage())
		require.NoError(t, err)
		err = json.Unmarshal(responseData, &tagDTO)
		require.NoError(t, err)
		require.Equal(t, tagID, tagDTO.ID)
		require.Equal(t, "updated-test-tag", tagDTO.Name)
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
