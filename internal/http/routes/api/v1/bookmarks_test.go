package api_v1

import (
	"context"
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

	account := testutil.GetValidAccount()
	require.NoError(t, deps.Database.SaveAccount(ctx, *account))
	token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Minute))
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

	account := testutil.GetValidAccount()
	require.NoError(t, deps.Database.SaveAccount(ctx, *account))
	token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Minute))
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
