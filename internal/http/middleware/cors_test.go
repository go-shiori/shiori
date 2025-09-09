package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-shiori/shiori/internal/http/webcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware(t *testing.T) {
	t.Run("test single origin", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:8080"}
		middleware := NewCORSMiddleware(allowedOrigins)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		err := middleware.OnRequest(nil, c)
		require.NoError(t, err)

		headers := w.Header()
		assert.Equal(t, "http://localhost:8080", headers.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", headers.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-Shiori-Response-Format", headers.Get("Access-Control-Allow-Headers"))
	})

	t.Run("test multiple origins", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:8080", "http://example.com"}
		middleware := NewCORSMiddleware(allowedOrigins)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		err := middleware.OnRequest(nil, c)
		require.NoError(t, err)

		headers := w.Header()
		assert.Equal(t, strings.Join(allowedOrigins, ", "), headers.Get("Access-Control-Allow-Origin"))
	})

	t.Run("test empty origins", func(t *testing.T) {
		middleware := NewCORSMiddleware([]string{})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		err := middleware.OnRequest(nil, c)
		require.NoError(t, err)

		headers := w.Header()
		assert.Equal(t, "", headers.Get("Access-Control-Allow-Origin"))
	})

	t.Run("test OnResponse headers", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:8080"}
		middleware := NewCORSMiddleware(allowedOrigins)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		err := middleware.OnResponse(nil, c)
		require.NoError(t, err)

		headers := w.Header()
		assert.Equal(t, "http://localhost:8080", headers.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", headers.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-Shiori-Response-Format", headers.Get("Access-Control-Allow-Headers"))
	})
}
