package handlers

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
)

func TestHandleLiveness(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("returns build info", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleLiveness(deps, c)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONContains(t, `{"version":"dev","commit":"none","date":"unknown"}`)
	})

	t.Run("handles without auth", func(t *testing.T) {
		// Test that liveness check works without authentication
		c, w := testutil.NewTestWebContext()
		HandleLiveness(deps, c)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
	})
}
