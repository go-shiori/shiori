package response

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SendFile sends file to client with caching header
func SendFile(c *gin.Context, path string) {
	c.Header("Cache-Control", "public, max-age=86400")

	info, err := os.Stat(path)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("ETag", fmt.Sprintf("W/%x-%x", info.ModTime().Unix(), info.Size()))
	c.File(path)
}
