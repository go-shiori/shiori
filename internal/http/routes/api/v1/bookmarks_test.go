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
