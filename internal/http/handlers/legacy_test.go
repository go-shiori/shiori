package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// SetFakeAuthorizationHeader sets a fake authorization header for the request in order to have
// a valid session. If we don't set this the `validateSession` function will return an error.
func SetFakeAuthorizationHeader(t *testing.T, deps model.Dependencies, c model.WebContext) {
	token, err := deps.Domains().Auth().CreateTokenForAccount(c.GetAccount(), time.Now().Add(time.Hour))
	require.NoError(t, err)
	c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func TestLegacyHandler(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)
	handler := NewLegacyHandler(deps)

	t.Run("HandleLogin", func(t *testing.T) {
		account := &model.AccountDTO{
			ID:       1,
			Username: "test",
			Owner:    model.Ptr(false),
		}

		sessionID, err := handler.HandleLogin(account, time.Hour)
		require.NoError(t, err)
		require.NotEmpty(t, sessionID)
	})

	t.Run("HandleGetTags", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()

		// Create a real account in the database
		account, err := deps.Domains().Accounts().CreateAccount(context.Background(), model.AccountDTO{
			Username: "testuser",
			Password: "testpass",
			Owner:    model.Ptr(false),
		})
		require.NoError(t, err)

		// Set the real account in the context
		c.SetAccount(account)
		SetFakeAuthorizationHeader(t, deps, c)
		handler.HandleGetTags(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("HandleGetBookmarks", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()

		// Create a real account in the database
		account, err := deps.Domains().Accounts().CreateAccount(context.Background(), model.AccountDTO{
			Username: "testuser2",
			Password: "testpass2",
			Owner:    model.Ptr(false),
		})
		require.NoError(t, err)

		// Set the real account in the context
		c.SetAccount(account)
		SetFakeAuthorizationHeader(t, deps, c)
		handler.HandleGetBookmarks(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("HandleGetBookmarks unauthenticated", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		handler.HandleGetBookmarks(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Equal(t, "Authentication required", resp["error"])
	})

	t.Run("HandleGetTags unauthenticated", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		handler.HandleGetTags(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Equal(t, "Authentication required", resp["error"])
	})

	t.Run("HandleOtherEndpoints unauthenticated", func(t *testing.T) {
		endpoints := []struct {
			name string
			fn   func(model.Dependencies, model.WebContext)
		}{
			{"HandleRenameTag", handler.HandleRenameTag},
			{"HandleInsertBookmark", handler.HandleInsertBookmark},
			{"HandleDeleteBookmark", handler.HandleDeleteBookmark},
			{"HandleUpdateBookmark", handler.HandleUpdateBookmark},
			{"HandleUpdateBookmarkTags", handler.HandleUpdateBookmarkTags},
			{"HandleInsertViaExtension", handler.HandleInsertViaExtension},
			{"HandleDeleteViaExtension", handler.HandleDeleteViaExtension},
		}

		for _, ep := range endpoints {
			t.Run(ep.name, func(t *testing.T) {
				c, w := testutil.NewTestWebContext()
				ep.fn(deps, c)
				require.Equal(t, http.StatusUnauthorized, w.Code)
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Equal(t, "Authentication required", resp["error"])
			})
		}
	})

	t.Run("Direct ApiGetBookmarks unauthenticated does not panic", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/api/bookmarks", http.NoBody)
		require.NoError(t, err)

		require.NotPanics(t, func() {
			handler.legacyHandler.ApiGetBookmarks(w, r, nil)
		})
		require.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.NotEmpty(t, resp["error"])
	})

	t.Run("convertParams", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/api/bookmarks?page=1&tags=test,dev", http.NoBody)
		params := handler.convertParams(r)

		require.Len(t, params, 2)

		// Create a map to check for parameters regardless of order
		paramMap := make(map[string]string)
		for _, param := range params {
			paramMap[param.Key] = param.Value
		}

		// Check that both parameters exist with the correct values
		require.Contains(t, paramMap, "page")
		require.Equal(t, "1", paramMap["page"])
		require.Contains(t, paramMap, "tags")
		require.Equal(t, "test,dev", paramMap["tags"])
	})
}
