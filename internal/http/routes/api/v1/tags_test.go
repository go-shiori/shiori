package api_v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTagList(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("empty tag list", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewTagsPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "GET", "/")
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertMessageIsEmptyList(t)
	})

	t.Run("return tags", func(t *testing.T) {
		ctx := context.TODO()

		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		router := NewTagsPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "GET", "/")
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertMessageIsEmptyList(t)
	})
}
