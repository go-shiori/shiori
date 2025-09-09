package response_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates successful response", func(t *testing.T) {
		resp := response.New(http.StatusOK, "test data")
		assert.False(t, resp.IsError())
		assert.Equal(t, "test data", resp.GetData())
	})

	t.Run("creates error response", func(t *testing.T) {
		resp := response.New(http.StatusBadRequest, "error data")
		assert.True(t, resp.IsError())
		assert.Equal(t, "error data", resp.GetData())
	})
}

func TestSend(t *testing.T) {
	t.Run("sends successful response", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := response.Send(c, http.StatusOK, "success message", "text/plain")
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageIsBytes(t, []byte("success message"))
	})

	t.Run("sends error response for status >= 400", func(t *testing.T) {
		message := "error message"
		c, w := testutil.NewTestWebContext()
		err := response.Send(c, http.StatusBadRequest, message, "text/plain")
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := response.NewResponse(message, http.StatusBadRequest)
		assert.True(t, response.IsError())
		assert.Equal(t, message, response.GetData())
	})
}

func TestSendError(t *testing.T) {
	t.Run("sends error response without params", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := response.SendError(c, http.StatusBadRequest, "error message")
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		responseBody := struct {
			Error string `json:"error"`
		}{Error: "error message"}
		response := response.NewResponse(responseBody, http.StatusBadRequest)

		assert.True(t, response.IsError())
		assert.Equal(t, responseBody, response.GetData())
	})

	t.Run("sends error response with params", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := response.SendError(c, http.StatusBadRequest, "error message")
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		responseBody := struct {
			Error string `json:"error"`
		}{Error: "error message"}
		response := response.NewResponse(responseBody, http.StatusBadRequest)

		assert.True(t, response.IsError())
		assert.Equal(t, responseBody, response.GetData())
	})
}

func TestSendInternalServerError(t *testing.T) {
	c, w := testutil.NewTestWebContext()
	err := response.SendInternalServerError(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	responseBody := struct {
		Error string `json:"error"`
	}{Error: "Internal server error, please contact an administrator"}
	response := response.NewResponse(responseBody, http.StatusInternalServerError)

	assert.True(t, response.IsError())
	assert.Equal(t, responseBody, response.GetData())
}

func TestRedirectToLogin(t *testing.T) {
	t.Run("redirects to login without destination", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		response.RedirectToLogin(c, "/", "")

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/?dst=", w.Header().Get("Location"))
	})

	t.Run("redirects to login with destination", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		response.RedirectToLogin(c, "/", "/dashboard")

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/?dst=%2Fdashboard", w.Header().Get("Location"))
	})
}

func TestNotFound(t *testing.T) {
	c, w := testutil.NewTestWebContext()
	response.NotFound(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "404 page not found")
}

func TestSendJSON(t *testing.T) {
	t.Run("sends JSON response", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		data := map[string]string{"key": "value"}
		err := response.SendJSON(c, http.StatusOK, data)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var result map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, data, result)
	})

	t.Run("handles encoding error", func(t *testing.T) {
		c, _ := testutil.NewTestWebContext()
		// Create a value that can't be marshaled to JSON
		data := map[string]any{"fn": func() {}}
		err := response.SendJSON(c, http.StatusOK, data)
		assert.Error(t, err)
	})
}
