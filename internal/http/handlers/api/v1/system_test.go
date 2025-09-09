package api_v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleSystemInfo(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("requires authentication", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleSystemInfo(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requires admin access", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		HandleSystemInfo(deps, c)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("returns system info for admin", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		HandleSystemInfo(deps, c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))

		response := testutil.NewTestResponseFromRecorder(w)

		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "version", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
		response.AssertMessageJSONKeyValue(t, "database", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
		response.AssertMessageJSONKeyValue(t, "os", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
	})
}
