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
		resp := response.New(true, http.StatusOK, "test data")
		assert.True(t, resp.Ok)
		assert.Equal(t, "test data", resp.Message)
		assert.Nil(t, resp.ErrorParams)
	})

	t.Run("creates error response", func(t *testing.T) {
		errorParams := map[string]string{"field": "error message"}
		resp := response.NewResponse(false, "error data", errorParams, http.StatusBadRequest)
		assert.False(t, resp.Ok)
		assert.Equal(t, "error data", resp.Message)
		assert.Equal(t, errorParams, resp.ErrorParams)
	})
}

func TestSend(t *testing.T) {
	t.Run("sends successful response", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := response.Send(c, http.StatusOK, "success message")
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp.Ok)
		assert.Equal(t, "success message", resp.Message)
		assert.Nil(t, resp.ErrorParams)
	})

	t.Run("sends error response for status >= 400", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := response.Send(c, http.StatusBadRequest, "error message")
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp.Ok)
		assert.Equal(t, "error message", resp.Message)
		assert.Nil(t, resp.ErrorParams)
	})
}

func TestSendError(t *testing.T) {
	t.Run("sends error response without params", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := response.SendError(c, http.StatusBadRequest, "error message", nil)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp.Ok)
		assert.Equal(t, "error message", resp.Message)
		assert.Nil(t, resp.ErrorParams)
	})

	t.Run("sends error response with params", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		errorParams := map[string]string{"field": "validation error"}
		err := response.SendError(c, http.StatusBadRequest, "error message", errorParams)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp.Ok)
		assert.Equal(t, "error message", resp.Message)
		assert.Equal(t, errorParams, resp.ErrorParams)
	})
}

func TestSendErrorWithParams(t *testing.T) {
	c, w := testutil.NewTestWebContext()
	errorParams := map[string]string{"field": "validation error"}
	err := response.SendErrorWithParams(c, http.StatusBadRequest, "error message", errorParams)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Ok)
	assert.Equal(t, "error message", resp.Message)
	assert.Equal(t, errorParams, resp.ErrorParams)
}

func TestSendInternalServerError(t *testing.T) {
	c, w := testutil.NewTestWebContext()
	err := response.SendInternalServerError(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Ok)
	assert.Equal(t, "Internal server error, please contact an administrator", resp.Message)
	assert.Nil(t, resp.ErrorParams)
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
		data := map[string]interface{}{"fn": func() {}}
		err := response.SendJSON(c, http.StatusOK, data)
		assert.Error(t, err)
	})
}

func TestSendErrorJSON(t *testing.T) {
	c, w := testutil.NewTestWebContext()
	err := response.SendErrorJSON(c, http.StatusBadRequest, "error message")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Ok)
	assert.Equal(t, "error message", resp.Message)
	assert.Nil(t, resp.ErrorParams)
}
