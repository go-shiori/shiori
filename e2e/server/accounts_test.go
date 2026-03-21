package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/stretchr/testify/require"
)

// apiLogin logs in as the given user and returns the bearer token.
func apiLogin(t *testing.T, baseURL, username, password string) string {
	t.Helper()
	resp, err := http.Post(
		baseURL+"/api/v1/auth/login",
		"application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{"Username":"%s","Password":"%s"}`, username, password))),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Token string `json:"Token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	resp.Body.Close()
	require.NotEmpty(t, result.Token)
	return result.Token
}

// apiCreateAccount creates a new account and returns its ID.
func apiCreateAccount(t *testing.T, baseURL, token, username, password string, owner bool) int {
	t.Helper()
	body := fmt.Sprintf(`{"Username":"%s","Password":"%s","Owner":%t}`, username, password, owner)
	req, err := http.NewRequest("POST", baseURL+"/api/v1/accounts", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result struct {
		ID int `json:"ID"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	resp.Body.Close()
	require.NotZero(t, result.ID)
	return result.ID
}

func TestAccountRoleChange(t *testing.T) {
	container := e2eutil.NewShioriContainer(t, "")
	baseURL := "http://localhost:" + container.GetPort()

	// Login as default admin
	adminToken := apiLogin(t, baseURL, "shiori", "gopher")

	// Create a regular user
	userID := apiCreateAccount(t, baseURL, adminToken, "testuser", "testpass", false)

	t.Run("promote user to admin", func(t *testing.T) {
		body := `{"Owner": true}`
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/accounts/%d", baseURL, userID), bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Owner bool `json:"Owner"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		resp.Body.Close()
		require.True(t, result.Owner, "User should now be an admin/owner")
	})

	t.Run("verify promoted user can access admin endpoints", func(t *testing.T) {
		// Login as the promoted user
		userToken := apiLogin(t, baseURL, "testuser", "testpass")

		// Access admin-only endpoint: list accounts
		req, err := http.NewRequest("GET", baseURL+"/api/v1/accounts", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Promoted user should access admin endpoints")
		resp.Body.Close()
	})

	t.Run("demote admin back to regular user", func(t *testing.T) {
		body := `{"Owner": false}`
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/accounts/%d", baseURL, userID), bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Owner bool `json:"Owner"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		resp.Body.Close()
		require.False(t, result.Owner, "User should no longer be an admin/owner")
	})

	t.Run("verify demoted user cannot access admin endpoints", func(t *testing.T) {
		// Login as the demoted user (need fresh token reflecting new role)
		userToken := apiLogin(t, baseURL, "testuser", "testpass")

		// Try to access admin-only endpoint: list accounts
		req, err := http.NewRequest("GET", baseURL+"/api/v1/accounts", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode, "Demoted user should not access admin endpoints")
		resp.Body.Close()
	})
}

func TestAdminCannotDeleteOwnAccount(t *testing.T) {
	container := e2eutil.NewShioriContainer(t, "")
	baseURL := "http://localhost:" + container.GetPort()

	// Login as default admin
	adminToken := apiLogin(t, baseURL, "shiori", "gopher")

	// Get the admin's own account ID via /api/v1/auth/me
	req, err := http.NewRequest("GET", baseURL+"/api/v1/auth/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var me struct {
		ID int `json:"ID"`
	}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	require.NoError(t, json.Unmarshal(body, &me))
	require.NotZero(t, me.ID)

	t.Run("deleting own account should be rejected", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/accounts/%d", baseURL, me.ID), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		// The API should reject self-deletion with a 4xx status code
		require.GreaterOrEqual(t, resp.StatusCode, 400, "Admin should not be able to delete their own account")
		resp.Body.Close()
	})
}
