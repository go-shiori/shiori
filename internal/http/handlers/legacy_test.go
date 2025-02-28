package handlers

import (
	"context"
	"fmt"
	"net/http"
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
		testutil.SetFakeUser(c)
		SetFakeAuthorizationHeader(t, deps, c)
		handler.HandleGetTags(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("HandleGetBookmarks", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		SetFakeAuthorizationHeader(t, deps, c)
		handler.HandleGetBookmarks(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
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
