package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSwaggerRoutes(t *testing.T) {
	logger := logrus.New()

	g := gin.Default()

	router := NewSwaggerAPIRoutes(logger)
	router.Setup(g.Group("/swagger"))

	t.Run("/swagger/doc.json", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/swagger/doc.json", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

	t.Run("/swagger/ redirects", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/swagger/", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 302, w.Code)
		require.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
	})

	t.Run("/swagger redirects", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/swagger/", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 302, w.Code)
		require.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
	})
}
