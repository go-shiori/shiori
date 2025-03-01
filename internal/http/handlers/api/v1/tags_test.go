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
