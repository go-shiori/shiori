package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestFrontendRoutes(t *testing.T) {
	logger := logrus.New()

	cfg, _ := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	g := gin.Default()
	router := NewFrontendRoutes(logger, cfg)
	router.Setup(g)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

	t.Run("/login", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/login", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

	t.Run("/css/stylesheet.css", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/assets/css/stylesheet.css", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})
}
