package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/stretchr/testify/require"
)

func TestAuthLogin(t *testing.T) {
	container := e2eutil.NewShioriContainer(t, "")
	baseURL := "http://localhost:" + container.GetPort()

	t.Run("login ok", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"Username": "shiori", "Password": "gopher"}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("wrong credentials", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"Username": "wrong", "Password": "wrong"}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("wrong password", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"Username": "shiori", "Password": "wrong"}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty username", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"Username": "", "Password": "gopher"}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty password", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"Username": "shiori", "Password": ""}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("logout", func(t *testing.T) {
		token := apiLogin(t, baseURL, "shiori", "gopher")

		req, err := http.NewRequest("POST", baseURL+"/api/v1/auth/logout", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("remember me short", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"Username": "shiori", "Password": "gopher"}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Expiration int64 `json:"Expiration"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		resp.Body.Close()

		expiresIn := time.Until(time.Unix(result.Expiration, 0))
		require.Less(t, expiresIn, 2*time.Hour, "without RememberMe, expiration should be less than 2 hours")
	})

	t.Run("remember me long", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"Username": "shiori", "Password": "gopher", "RememberMe": true}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Expiration int64 `json:"Expiration"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		resp.Body.Close()

		expiresIn := time.Until(time.Unix(result.Expiration, 0))
		require.Greater(t, expiresIn, 29*24*time.Hour, "with RememberMe, expiration should be more than 29 days")
	})

	t.Run("unauthenticated 401", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURL+"/api/v1/auth/me", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
