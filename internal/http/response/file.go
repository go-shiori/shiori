package response

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-shiori/shiori/internal/model"
)

// SendFile sends file to client with caching header
func SendFile(c *gin.Context, storageDomain model.StorageDomain, path string) {
	c.Header("Cache-Control", "public, max-age=86400")

	if !storageDomain.FileExists(path) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	info, err := storageDomain.Stat(path)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("ETag", fmt.Sprintf("W/%x-%x", info.ModTime().Unix(), info.Size()))

	fileHandler, err := storageDomain.Open(path)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer fileHandler.Close()

	_, err = io.Copy(c.Writer, fileHandler)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
