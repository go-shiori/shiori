package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStripWebrootPrefixMiddleware(t *testing.T) {
	prefix := "/prefix"
	// Create a new router with the middleware
	r := gin.New()
	r.Use(StripWebrootPrefixMiddleware(prefix))
	r.Use(func(ctx *gin.Context) {
		fmt.Println("asd", ctx.Request.URL.Path)
	})

	// Define a test route
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test")
	})

	// Create a test request
	req, err := http.NewRequest("GET", prefix+"/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Perform the request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "Test" {
		t.Errorf("Expected body %q but got %q", "Test", w.Body.String())
	}
}
