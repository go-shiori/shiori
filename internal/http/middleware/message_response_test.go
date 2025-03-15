package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageResponseMiddleware(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("wraps JSON response with success status", func(t *testing.T) {
		// Create test handler that returns JSON
		handler := func(deps model.Dependencies, c model.WebContext) {
			response := map[string]string{"data": "test"}
			c.ResponseWriter().Header().Set("Content-Type", "application/json")
			c.ResponseWriter().WriteHeader(http.StatusOK)
			json.NewEncoder(c.ResponseWriter()).Encode(response)
		}

		// Create test context
		c, w := testutil.NewTestWebContext()

		// Create and apply middleware
		middleware := NewMessageResponseMiddleware(deps)
		require.NoError(t, middleware.OnRequest(deps, c))

		// Execute handler
		handler(deps, c)

		// Apply response middleware
		require.NoError(t, middleware.OnResponse(deps, c))

		// Verify response
		var response responseMiddlewareBody
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, map[string]any{"data": "test"}, response.Message)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("wraps JSON response with error status", func(t *testing.T) {
		// Create test handler that returns JSON error
		handler := func(deps model.Dependencies, c model.WebContext) {
			response := map[string]string{"error": "test error"}
			c.ResponseWriter().Header().Set("Content-Type", "application/json")
			c.ResponseWriter().WriteHeader(http.StatusBadRequest)
			json.NewEncoder(c.ResponseWriter()).Encode(response)
		}

		// Create test context
		c, w := testutil.NewTestWebContext()

		// Create and apply middleware
		middleware := NewMessageResponseMiddleware(deps)
		require.NoError(t, middleware.OnRequest(deps, c))

		// Execute handler
		handler(deps, c)

		// Apply response middleware
		require.NoError(t, middleware.OnResponse(deps, c))

		// Verify response
		var response responseMiddlewareBody
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.False(t, response.Ok)
		assert.Equal(t, map[string]any{"error": "test error"}, response.Message)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("does not modify non-JSON response", func(t *testing.T) {
		// Create test handler that returns plain text
		handler := func(deps model.Dependencies, c model.WebContext) {
			c.ResponseWriter().Header().Set("Content-Type", "text/plain")
			c.ResponseWriter().WriteHeader(http.StatusOK)
			c.ResponseWriter().Write([]byte("test message"))
		}

		// Create test context
		c, w := testutil.NewTestWebContext()

		// Create and apply middleware
		middleware := NewMessageResponseMiddleware(deps)
		require.NoError(t, middleware.OnRequest(deps, c))

		// Execute handler
		handler(deps, c)

		// Apply response middleware
		require.NoError(t, middleware.OnResponse(deps, c))

		// Verify response is unchanged
		assert.Equal(t, "test message", w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	})

	t.Run("handles empty JSON response", func(t *testing.T) {
		// Create test handler that returns empty JSON
		handler := func(deps model.Dependencies, c model.WebContext) {
			c.ResponseWriter().Header().Set("Content-Type", "application/json")
			c.ResponseWriter().WriteHeader(http.StatusOK)
			c.ResponseWriter().Write([]byte("{}"))
		}

		// Create test context
		c, w := testutil.NewTestWebContext()

		// Create and apply middleware
		middleware := NewMessageResponseMiddleware(deps)
		require.NoError(t, middleware.OnRequest(deps, c))

		// Execute handler
		handler(deps, c)

		// Apply response middleware
		require.NoError(t, middleware.OnResponse(deps, c))

		// Verify response
		var response responseMiddlewareBody
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, map[string]any{}, response.Message)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("preserves custom headers", func(t *testing.T) {
		// Create test handler that sets custom headers
		handler := func(deps model.Dependencies, c model.WebContext) {
			c.ResponseWriter().Header().Set("Content-Type", "application/json")
			c.ResponseWriter().Header().Set("X-Custom-Header", "test-value")
			c.ResponseWriter().WriteHeader(http.StatusOK)
			json.NewEncoder(c.ResponseWriter()).Encode(map[string]string{"data": "test"})
		}

		// Create test context
		c, w := testutil.NewTestWebContext()

		// Create and apply middleware
		middleware := NewMessageResponseMiddleware(deps)
		require.NoError(t, middleware.OnRequest(deps, c))

		// Execute handler
		handler(deps, c)

		// Apply response middleware
		require.NoError(t, middleware.OnResponse(deps, c))

		// Verify headers are preserved
		assert.Equal(t, "test-value", w.Header().Get("X-Custom-Header"))
	})
}
