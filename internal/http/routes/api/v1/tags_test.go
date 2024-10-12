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

func TestTagList(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	g := gin.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	g.Use(middleware.AuthMiddleware(deps))

	account := testutil.GetValidAccount()
	account.Owner = true
	require.NoError(t, deps.Database.SaveAccount(ctx, *account))
	token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Minute))
	require.NoError(t, err)

	bookmark := testutil.GetValidBookmark()
	bookmark.Tags = []model.Tag{
		{Name: "test"},
	}
	_, err = deps.Database.SaveBookmarks(ctx, true, *bookmark)
	require.NoError(t, err)

	router := NewTagsPIRoutes(logger, deps)
	router.Setup(g.Group("/"))

	t.Run("require authentication", func(t *testing.T) {
		w := testutil.PerformRequest(g, "GET", "/")
		require.Equal(t, http.StatusUnauthorized, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})

	t.Run("return tags", func(t *testing.T) {
		w := testutil.PerformRequest(g, "GET", "/", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertOk(t)
		response.AssertMessageIsListLength(t, 1)
	})
}

func TestTagCreate(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	g := gin.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	g.Use(middleware.AuthMiddleware(deps))

	account := testutil.GetValidAccount()
	account.Owner = true
	require.NoError(t, deps.Database.SaveAccount(ctx, *account))
	// token, err := deps.Domains.Auth.CreateTokenForAccount(&account, time.Now().Add(time.Minute))
	// require.NoError(t, err)

	router := NewTagsPIRoutes(logger, deps)
	router.Setup(g.Group("/"))

	t.Run("require authentication", func(t *testing.T) {
		w := testutil.PerformRequest(g, "POST", "/")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create tag", func(t *testing.T) {
		// TODO: Implement this test
		// Tags require a bookmark to be created, so we need to create a bookmark first
		// but I'm not sure if we should enforce this.
	})
}
