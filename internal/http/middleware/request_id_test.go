package middleware

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("adds request ID to context and headers", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		middleware := NewRequestIDMiddleware(deps)

		c, w := testutil.NewTestWebContext()
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)

		// Check that request ID was added to context
		requestID := c.GetRequestID()
		require.NotEmpty(t, requestID)

		// Check that request ID was added to headers
		headerRequestID := w.Header().Get(RequestIDHeader)
		require.Equal(t, requestID, headerRequestID)
	})
}
