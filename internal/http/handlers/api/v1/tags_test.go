package api_v1

import (
	"context"
	"net/http"
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
		c, w := testutil.NewTestWebContext()
		HandleListTags(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns tags list", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test tag
		_, err := deps.Domains().Tags().CreateTag(ctx, model.TagDTO{
			Tag: model.Tag{Name: "test-tag"},
		})
		require.NoError(t, err)

		t.Log(deps.Database().GetTags(ctx))

		w := testutil.PerformRequest(deps, HandleListTags, "GET", "/api/v1/tags", testutil.WithFakeAccount(true))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)
		response.AssertMessageIsNotEmptyList(t)
	})
}
