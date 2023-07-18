package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestFrontendRoutes(t *testing.T) {
	logger := logrus.New()

	g := gin.Default()
	router := NewFrontendRoutes(logger, &config.HttpConfig{})
	router.Setup(g)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

	t.Run("/login.html", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/login.html", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

	t.Run("/css/stylesheet.css", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/css/stylesheet.css", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})
}
