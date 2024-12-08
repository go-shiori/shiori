package response

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/model"
)

type SendFileOptions struct {
	Headers []http.Header
}

// SendFile sends file to client with caching header
func SendFile(c *gin.Context, storageDomain model.StorageDomain, path string, options *SendFileOptions) {
	if !storageDomain.FileExists(path) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	info, err := storageDomain.Stat(path)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("ETag", fmt.Sprintf("W/%x-%x", info.ModTime().Unix(), info.Size()))

	if options != nil {
		for _, header := range options.Headers {
			for key, value := range header {
				c.Header(key, value[0])
			}
		}
	}

	// TODO: Find a better way to send the file to the client from the FS, probably making a
	// conversion between afero.Fs and http.FileSystem to use c.FileFromFS.
	fileContent, err := storageDomain.FS().Open(path)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(c.Writer, fileContent)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
