package response_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func newMockStorage(deps model.Dependencies, fs afero.Fs) model.StorageDomain {
	return domains.NewStorageDomain(deps, fs)
}

func TestSendFile(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	storage := newMockStorage(deps, afero.NewMemMapFs())

	t.Run("sends file with correct content type from extension", func(t *testing.T) {
		// Create test file
		content := []byte("body { color: red; }")
		err := storage.WriteData("test.css", content)
		require.NoError(t, err)

		c, w := testutil.NewTestWebContext()
		err = response.SendFile(c, storage, "test.css", nil)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
		require.Equal(t, content, w.Body.Bytes())
	})

	t.Run("sends file with detected content type", func(t *testing.T) {
		// Create test file without extension
		content := []byte("<html><body>Hello</body></html>")
		err := storage.WriteData("test", content)
		require.NoError(t, err)

		c, w := testutil.NewTestWebContext()
		err = response.SendFile(c, storage, "test", nil)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
		require.Equal(t, content, w.Body.Bytes())
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		_ = response.SendFile(c, storage, "nonexistent.txt", nil)
		require.Equal(t, http.StatusNotFound, w.Code)
		require.Contains(t, w.Body.String(), "File not found")
	})

	t.Run("sets custom headers", func(t *testing.T) {
		// Create test file
		content := []byte("test content")
		err := storage.WriteData("test.txt", content)
		require.NoError(t, err)

		options := &response.SendFileOptions{
			Headers: []http.Header{
				{"Cache-Control": {"no-cache"}},
				{"X-Custom": {"value1", "value2"}},
			},
		}

		c, w := testutil.NewTestWebContext()
		err = response.SendFile(c, storage, "test.txt", options)
		require.NoError(t, err)

		require.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
		require.Equal(t, []string{"value1", "value2"}, w.Header().Values("X-Custom"))
	})

	t.Run("handles large files", func(t *testing.T) {
		// Create large test file (>512 bytes to test content type detection)
		binaryData := bytes.Repeat([]byte{0xFF, 0x00}, 1024*1024)
		err := storage.WriteData("large.bin", binaryData)
		require.NoError(t, err)

		c, w := testutil.NewTestWebContext()
		err = response.SendFile(c, storage, "large.bin", nil)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
		require.Equal(t, binaryData, w.Body.Bytes())
	})

	t.Run("handles empty files", func(t *testing.T) {
		err := storage.WriteData("empty.txt", []byte{})
		require.NoError(t, err)

		c, w := testutil.NewTestWebContext()
		err = response.SendFile(c, storage, "empty.txt", nil)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
		require.Empty(t, w.Body.Bytes())
	})

	t.Run("handles read errors", func(t *testing.T) {
		// Create mock file that returns error on read
		errorFs := &errorFs{
			Fs:  afero.NewMemMapFs(),
			err: io.ErrClosedPipe,
		}
		storage := newMockStorage(deps, errorFs)
		err := storage.WriteData("test.txt", []byte("test"))
		require.NoError(t, err)

		c, w := testutil.NewTestWebContext()
		_ = response.SendFile(c, storage, "test.txt", nil)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// errorFs is a mock filesystem that returns errors
type errorFs struct {
	afero.Fs
	err error
}

func (e *errorFs) Open(name string) (afero.File, error) {
	return nil, e.err
}
