package api_v1

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestUpdateBookmarkCache(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	g := gin.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	g.Use(middleware.AuthMiddleware(deps))

	router := NewBookmarksAPIRoutes(logger, deps)
	router.Setup(g.Group("/"))

	account := model.Account{
		Username: "test",
		Password: "test",
		Owner:    false,
	}
	require.NoError(t, deps.Database.SaveAccount(ctx, account))
	token, err := deps.Domains.Auth.CreateTokenForAccount(&account, time.Now().Add(time.Minute))
	require.NoError(t, err)

	t.Run("require authentication", func(t *testing.T) {
		w := testutil.PerformRequest(g, "PUT", "/cache")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("require owner", func(t *testing.T) {
		w := testutil.PerformRequest(g, "PUT", "/cache", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))
		require.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestReadableeBookmarkContent(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	g := gin.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	g.Use(middleware.AuthMiddleware(deps))

	router := NewBookmarksAPIRoutes(logger, deps)
	router.Setup(g.Group("/"))

	account := model.Account{
		Username: "test",
		Password: "test",
		Owner:    false,
	}
	require.NoError(t, deps.Database.SaveAccount(ctx, account))
	token, err := deps.Domains.Auth.CreateTokenForAccount(&account, time.Now().Add(time.Minute))
	require.NoError(t, err)

	bookmark := testutil.GetValidBookmark()
	_, err = deps.Database.SaveBookmarks(ctx, true, *bookmark)
	require.NoError(t, err)
	response := `{"ok":true,"message":{"content":"","html":""}}`

	t.Run("require authentication", func(t *testing.T) {
		w := testutil.PerformRequest(g, "GET", "/1/readable")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("get content but invalid id", func(t *testing.T) {
		w := testutil.PerformRequest(g, "GET", "/invalidId/readable", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
	t.Run("get content but 0 id", func(t *testing.T) {
		w := testutil.PerformRequest(g, "GET", "/0/readable", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))
		require.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("get content but not exist", func(t *testing.T) {
		w := testutil.PerformRequest(g, "GET", "/2/readable", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))
		require.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("get content", func(t *testing.T) {
		w := testutil.PerformRequest(g, "GET", "/1/readable", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))
		require.Equal(t, response, w.Body.String())
		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSync(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	g := gin.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	g.Use(middleware.AuthMiddleware(deps))

	router := NewBookmarksAPIRoutes(logger, deps)
	router.Setup(g.Group("/"))

	account := model.Account{
		Username: "test",
		Password: "test",
		Owner:    false,
	}
	require.NoError(t, deps.Database.SaveAccount(ctx, account))
	token, err := deps.Domains.Auth.CreateTokenForAccount(&account, time.Now().Add(time.Minute))
	require.NoError(t, err)

	// all payloads need
	payloadInvalidID := syncPayload{
		Ids:      []int{0, -1},
		LastSync: 0,
		Page:     1,
	}

	// Json format of payloads
	payloadJSONInvalidID, err := json.Marshal(payloadInvalidID)
	if err != nil {
		logrus.Printf("can't create a valid json")
	}

	// Add bookmarks to the database
	bookmarkFirst := testutil.GetValidBookmark()
	_, err = deps.Database.SaveBookmarks(ctx, true, *bookmarkFirst)
	require.NoError(t, err)

	bookmarkSecond := testutil.GetValidBookmark()
	bookmarkSecond.Title = "second bookmark"
	unixTimestampOneSecondLater := time.Now().UTC().Add(1 * time.Second).Unix()
	bookmarkSecond.ModifiedAt = time.Unix(unixTimestampOneSecondLater, 0).UTC().Format(model.DatabaseDateFormat)
	_, err = deps.Database.SaveBookmarks(ctx, true, *bookmarkSecond)
	require.NoError(t, err)

	payloadValid := syncPayload{
		Ids:      []int{},
		LastSync: unixTimestampOneSecondLater,
		Page:     1,
	}

	payloadJSONValid, err := json.Marshal(payloadValid)
	if err != nil {
		logrus.Printf("can't create a valid json")
	}
	t.Run("require authentication", func(t *testing.T) {
		w := testutil.PerformRequest(g, "POST", "/sync")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		w := testutil.PerformRequest(g, "POST", "/sync", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token), testutil.WithBody(string(payloadJSONInvalidID)))
		require.Equal(t, http.StatusBadRequest, w.Code)

		// Check the response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "failed to unmarshal response body")

		// Assert that the response message is as expected for 0 or negative id
		require.Equal(t, "id should not be 0 or negative", response["message"])
	})

	t.Run("retun both bookmark with sync api", func(t *testing.T) {
		w := testutil.PerformRequest(g, "POST", "/sync", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token), testutil.WithBody(string(payloadJSONValid)))
		require.Equal(t, http.StatusOK, w.Code)

		// Check the response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "failed to unmarshal response body")

		// Assert that the response message is as expected
		require.Equal(t, true, response["ok"])

		// Access the bookmarks
		message := response["message"].(map[string]interface{})
		modified := message["modified"].(map[string]interface{})
		bookmarks := modified["bookmarks"].([]interface{})

		// Check the IDs of the bookmarks
		var ids []int
		for _, bookmark := range bookmarks {
			bookmarkMap := bookmark.(map[string]interface{})
			id := int(bookmarkMap["id"].(float64))
			ids = append(ids, id)
		}

		// Assert that the IDs are as expected
		expectedIDs := []int{2}
		require.ElementsMatch(t, expectedIDs, ids, "bookmark IDs do not match")
	})
}
