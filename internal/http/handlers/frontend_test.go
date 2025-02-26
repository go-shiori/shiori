package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/http/templates"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleFrontend(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	err := templates.SetupTemplates()
	require.NoError(t, err)

	t.Run("serves index page", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleFrontend(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	})
}

func TestHandleAssets(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("serves css file", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/assets/css/style.css")
		HandleAssets(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Type"), "text/css")
	})

	t.Run("returns 404 for missing file", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/assets/not-found.txt")
		HandleAssets(deps, c)
		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
