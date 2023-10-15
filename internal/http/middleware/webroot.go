package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// StripWebrootPrefixiddleware is a middleware that strips prefix from request path.
func StripWebrootPrefixMiddleware(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If prefix does not start with slash, add it
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		fmt.Println(c.Request.URL.Path)
		fmt.Println(prefix)
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, prefix)
		fmt.Println(c.Request.URL.Path)
		c.Next()
	}
}
