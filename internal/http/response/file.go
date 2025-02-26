package response

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/go-shiori/shiori/internal/model"
)

// SendFileOptions contains options for sending files
type SendFileOptions struct {
	Headers []http.Header
}

// SendFile sends a file from storage to the response writer
func SendFile(c model.WebContext, storage model.StorageDomain, path string, options *SendFileOptions) error {
	if !storage.FileExists(path) {
		return SendError(c, http.StatusNotFound, "File not found", nil)
	}

	file, err := storage.FS().Open(path)
	if err != nil {
		return SendInternalServerError(c)
	}
	defer file.Close()

	// First try to get content type from extension
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		// If no extension or unknown, try to detect from content
		// Only the first 512 bytes are used to sniff the content type
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file header: %w", err)
		}
		contentType = http.DetectContentType(buffer[:n])

		// Seek back to start since we read some bytes
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
	}

	// Set content type
	c.ResponseWriter().Header().Set("Content-Type", contentType)

	// Set additional headers if provided
	if options != nil {
		for _, header := range options.Headers {
			for key, values := range header {
				for _, value := range values {
					c.ResponseWriter().Header().Add(key, value)
				}
			}
		}
	}

	// Copy file to response writer
	_, err = io.Copy(c.ResponseWriter(), file)
	if err != nil {
		return fmt.Errorf("failed to send file: %w", err)
	}

	return nil
}
