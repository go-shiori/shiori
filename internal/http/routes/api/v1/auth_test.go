package api_v1

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func noopLegacyLoginHandler(_ *model.AccountDTO, _ time.Duration) (string, error) {
	return "", nil
}

func TestAccountsRoute(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("login invalid", func(t *testing.T) {
		g := testutil.NewGin()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))
		body := `{"username": "gopher"}`
		w := testutil.PerformRequest(g, "POST", "/login", testutil.WithBody(body))

		require.Equal(t, 400, w.Code)
	})

	t.Run("login incorrect", func(t *testing.T) {
		g := testutil.NewGin()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))
		body := `{"username": "gopher", "password": "shiori"}`
		w := testutil.PerformRequest(g, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, 400, w.Code)
	})

	t.Run("login correct", func(t *testing.T) {
		g := testutil.NewGin()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))

		// Create an account manually to test
		account := model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
			Owner:    model.Ptr(true),
		}

		_, accountInsertErr := deps.Domains.Accounts.CreateAccount(ctx, account)
		require.NoError(t, accountInsertErr)

		w := testutil.PerformRequest(g, "POST", "/login", testutil.WithBody(`{"username": "shiori", "password": "gopher"}`))

		require.Equal(t, 200, w.Code)
	})

	t.Run("check /me (correct token)", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		g := testutil.NewGin()
		g.Use(middleware.AuthMiddleware(deps))

		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))

		// Create an account manually to test
		account := model.Account{
			Username: "shiori",
			Password: "gopher",
			Owner:    true,
		}
		_, accountInsertErr := deps.Database.CreateAccount(ctx, account)
		require.NoError(t, accountInsertErr)

		token, err := deps.Domains.Auth.CreateTokenForAccount(model.Ptr(account.ToDTO()), time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := testutil.PerformRequest(g, "GET", "/me", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("check /me (incorrect token)", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		g := testutil.NewGin()
		g.Use(middleware.AuthMiddleware(deps))

		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/refresh", testutil.WithAuthToken("nometokens"))

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestLoginRequestPayload(t *testing.T) {
	// Test empty payload
	t.Run("test empty payload", func(t *testing.T) {
		payload := loginRequestPayload{}
		err := payload.IsValid()
		require.Error(t, err)
	})

	// Test empty username
	t.Run("test empty username", func(t *testing.T) {
		payload := loginRequestPayload{
			Password: "gopher",
		}
		err := payload.IsValid()
		require.Error(t, err)
	})

	// Test empty password
	t.Run("test empty password", func(t *testing.T) {
		payload := loginRequestPayload{
			Username: "shiori",
		}
		err := payload.IsValid()
		require.Error(t, err)
	})

	// Test valid payload
	t.Run("test valid payload", func(t *testing.T) {
		payload := loginRequestPayload{
			Username: "shiori",
			Password: "gopher",
		}
		err := payload.IsValid()
		require.NoError(t, err)
	})
}

func TestRefreshHandler(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()
	g := testutil.NewGin()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
	g.Use(middleware.AuthMiddleware(deps)) // Requires AuthMiddleware to manipulate context
	router.Setup(g.Group("/"))

	t.Run("empty headers", func(t *testing.T) {
		w := testutil.PerformRequest(g, "POST", "/refresh")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token invalid", func(t *testing.T) {
		w := testutil.PerformRequest(g, "POST", "/refresh", testutil.WithAuthToken("nometokens"))
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token valid", func(t *testing.T) {
		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		w := testutil.PerformRequest(g, "POST", "/refresh", testutil.WithAuthToken(token))

		require.Equal(t, http.StatusAccepted, w.Code)
	})
}

func TestUpdateHandler(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()
	g := testutil.NewGin()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
	g.Use(middleware.AuthMiddleware(deps))
	router.Setup(g.Group("/"))

	account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
		Username: "shiori",
		Password: "gopher",
		Owner:    model.Ptr(true),
		Config: model.Ptr(model.UserConfig{
			ShowId:        true,
			ListMode:      true,
			HideThumbnail: true,
			HideExcerpt:   true,
			KeepMetadata:  true,
			UseArchive:    true,
			CreateEbook:   true,
			MakePublic:    true,
		}),
	})
	require.NoError(t, err)

	t.Run("require authentication", func(t *testing.T) {
		w := testutil.PerformRequest(g, "PATCH", "/account")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("config not valid", func(t *testing.T) {
		token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := testutil.PerformRequest(g, "PATCH", "/account", testutil.WithBody("notValidConfig"), testutil.WithAuthToken(token))

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("password update with invalid old password", func(t *testing.T) {
		token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		payloadJSON := `{
			"old_password": "wrongpassword",
			"new_password": "newpassword"
		}`

		w := testutil.PerformRequest(g, "PATCH", "/account", testutil.WithBody(payloadJSON), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("password update with correct old password", func(t *testing.T) {
		token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		payloadJSON := `{
			"old_password": "gopher",
			"new_password": "newpassword"
		}`

		w := testutil.PerformRequest(g, "PATCH", "/account", testutil.WithBody(payloadJSON), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusOK, w.Code)

		// Verify we can login with new password
		loginW := testutil.PerformRequest(g, "POST", "/login", testutil.WithBody(`{"username": "shiori", "password": "newpassword"}`))
		require.Equal(t, http.StatusOK, loginW.Code)
	})

	t.Run("Test configure change in database", func(t *testing.T) {
		// Get current user config
		user, _, err := deps.Database.GetAccount(ctx, account.ID)
		require.NoError(t, err)
		require.Equal(t, user.ToDTO().Config, account.Config)

		// Send Request to update config for user
		token, err := deps.Domains.Auth.CreateTokenForAccount(model.Ptr(user.ToDTO()), time.Now().Add(time.Minute))
		require.NoError(t, err)

		payloadJSON := `{
			"config": {
			"ShowId": false,
			"ListMode": false,
			"HideThumbnail": false,
			"HideExcerpt": false,
			"Theme": "follow",
			"KeepMetadata": false,
			"UseArchive": false,
			"CreateEbook": false,
			"MakePublic": false
			}}`

		w := testutil.PerformRequest(g, "PATCH", "/account", testutil.WithBody(payloadJSON), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusOK, w.Code)

		user, _, err = deps.Database.GetAccount(ctx, account.ID)
		require.NoError(t, err)

		require.NotEqualValues(t, user.ToDTO().Config, account.Config)
	})
}
