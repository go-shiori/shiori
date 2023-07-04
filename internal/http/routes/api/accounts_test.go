package api

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAccountsRoute(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("login incorrect", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := httptest.NewRecorder()
		body := []byte(`{"username": "gopher", "password": "shiori"}`)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		g.ServeHTTP(w, req)

		require.Equal(t, 400, w.Code)
	})

	t.Run("login correct", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		// Create an account manually to test
		account := model.Account{
			Username: "shiori",
			Password: "gopher",
			Owner:    true,
		}
		require.NoError(t, deps.Database.SaveAccount(ctx, account))

		w := httptest.NewRecorder()
		body := []byte(`{"username": "shiori", "password": "gopher"}`)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		g.ServeHTTP(w, req)

		require.Equal(t, 200, w.Code)
	})

	t.Run("check /me (correct token)", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		g := gin.New()
		g.Use(middleware.AuthMiddleware(deps))

		router := NewAccountAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		// Create an account manually to test
		account := model.Account{
			Username: "shiori",
			Password: "gopher",
			Owner:    true,
		}
		require.NoError(t, deps.Database.SaveAccount(ctx, account))

		token, err := deps.Domains.Auth.CreateTokenForAccount(&account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Add("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		require.Equal(t, 200, w.Code)
	})

	t.Run("check /me (incorrect token)", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		g := gin.New()
		g.Use(middleware.AuthMiddleware(deps))

		router := NewAccountAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		req := httptest.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		require.Equal(t, 403, w.Code)
	})
}
