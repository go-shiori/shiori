package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleSwagger(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("serves swagger doc.json", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/swagger/doc.json")
		HandleSwagger(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("redirects /swagger/ to index", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/swagger/")
		HandleSwagger(deps, c)
		require.Equal(t, 301, w.Code)
		require.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
	})

	t.Run("redirects /swagger to index", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/swagger")
		HandleSwagger(deps, c)
		require.Equal(t, http.StatusPermanentRedirect, w.Code)
		require.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
	})
}
