package api_v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func noopLegacyLoginHandler(_ model.Account, _ time.Duration) (string, error) {
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
		w := httptest.NewRecorder()
		body := []byte(`{"username": "gopher"}`)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		g.ServeHTTP(w, req)

		require.Equal(t, 400, w.Code)
	})

	t.Run("login incorrect", func(t *testing.T) {
		g := testutil.NewGin()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))
		w := httptest.NewRecorder()
		body := []byte(`{"username": "gopher", "password": "shiori"}`)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		g.ServeHTTP(w, req)

		require.Equal(t, 400, w.Code)
	})

	t.Run("login correct", func(t *testing.T) {
		g := testutil.NewGin()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))

		// Create an account manually to test
		account := model.Account{
			Username: "shiori",
			Password: "gopher",
			Owner:    true,
		}
		require.NoError(t, deps.Database.SaveAccount(ctx, account))

		w := httptest.NewRecorder()
		body := []byte(`{"username": "shiori", "password": "gopher"}`)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		g.ServeHTTP(w, req)

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
		require.NoError(t, deps.Database.SaveAccount(ctx, account))

		token, err := deps.Domains.Auth.CreateTokenForAccount(&account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Add("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		require.Equal(t, 200, w.Code)
	})

	t.Run("check /me (incorrect token)", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		g := testutil.NewGin()
		g.Use(middleware.AuthMiddleware(deps))

		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		router.Setup(g.Group("/"))

		req := httptest.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		require.Equal(t, 403, w.Code)
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
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("token invalid", func(t *testing.T) {
		w := testutil.PerformRequest(g, "POST", "/refresh")
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("token valid", func(t *testing.T) {
		token, err := deps.Domains.Auth.CreateTokenForAccount(&model.Account{
			Username: "shiori",
		}, time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := testutil.PerformRequest(g, "POST", "/refresh", testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))

		require.Equal(t, http.StatusAccepted, w.Code)
	})
}

func TestSettingsHandler(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()
	g := testutil.NewGin()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
	g.Use(middleware.AuthMiddleware(deps))
	router.Setup(g.Group("/"))

	t.Run("token valid", func(t *testing.T) {
		token, err := deps.Domains.Auth.CreateTokenForAccount(&model.Account{
			Username: "shiori",
		}, time.Now().Add(time.Minute))
		require.NoError(t, err)

		type settingRequestPayload struct {
			Config model.UserConfig `json:"config"`
		}
		payload := settingRequestPayload{
			Config: model.UserConfig{
				// add your configuration data here
			},
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			logrus.Printf("problem")
		}

		w := testutil.PerformRequest(g, "PATCH", "/account", testutil.WithBody(string(payloadJSON)), testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))

		require.Equal(t, http.StatusOK, w.Code)

	})

	t.Run("config not valid", func(t *testing.T) {
		token, err := deps.Domains.Auth.CreateTokenForAccount(&model.Account{
			Username: "shiori",
		}, time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := testutil.PerformRequest(g, "PATCH", "/account", testutil.WithBody("notValidConfig"), testutil.WithHeader(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token))

		require.Equal(t, http.StatusInternalServerError, w.Code)

	})
	t.Run("Test configure change in database", func(t *testing.T) {
		// Create a tmp database
		g := testutil.NewGin()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAuthAPIRoutes(logger, deps, noopLegacyLoginHandler)
		g.Use(middleware.AuthMiddleware(deps))
		router.Setup(g.Group("/"))

		// Create an account manually to test
		account := model.Account{
			Username: "shiori",
			Password: "gopher",
			Owner:    true,
			Config: model.UserConfig{
				ShowId:        true,
				ListMode:      true,
				HideThumbnail: true,
				HideExcerpt:   true,
				NightMode:     true,
				KeepMetadata:  true,
				UseArchive:    true,
				CreateEbook:   true,
				MakePublic:    true,
			},
		}
		require.NoError(t, deps.Database.SaveAccount(ctx, account))

		// Get current user config
		user, _, err := deps.Database.GetAccount(ctx, "shiori")
		require.NoError(t, err)
		require.Equal(t, user.Config, account.Config)

		// Send Request to update config for user
		token, err := deps.Domains.Auth.CreateTokenForAccount(&user, time.Now().Add(time.Minute))
		require.NoError(t, err)

		payloadJSON := []byte(`{
			"config": {
			"ShowId": false,
			"ListMode": false,
			"HideThumbnail": false,
			"HideExcerpt": false,
			"NightMode": false,
			"KeepMetadata": false,
			"UseArchive": false,
			"CreateEbook": false,
			"MakePublic": false
			  }
			}`)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/account", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)
		g.ServeHTTP(w, req)

		require.Equal(t, 200, w.Code)
		user, _, err = deps.Database.GetAccount(ctx, "shiori")

		require.NoError(t, err)
		require.NotEqual(t, user.Config, account.Config)

	})
}
