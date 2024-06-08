package api_v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSystemRoute(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("valid response", func(t *testing.T) {
		g := testutil.NewGin()
		g.Use(testutil.FakeAdminLoggedInMiddlewware)
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewSystemAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, http.MethodGet, "/info")
		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertOk(t)
	})

	t.Run("requires authentication", func(t *testing.T) {
		g := testutil.NewGin()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewSystemAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, http.MethodGet, "/info")
		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("requires admin", func(t *testing.T) {
		g := testutil.NewGin()
		g.Use(testutil.FakeUserLoggedInMiddlewware)
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewSystemAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, http.MethodGet, "/info")
		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
		require.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})
}
