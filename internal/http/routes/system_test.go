package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSystemRoutes(t *testing.T) {
	logger := logrus.New()

	g := gin.Default()
	router := NewSystemRoutes(logger)
	router.Setup(g.Group("/system"))

	t.Run("/system/liveness", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/system/liveness", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

}
