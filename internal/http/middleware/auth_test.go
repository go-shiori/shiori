package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/http/webcontext"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("test no authorization method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())
	})

	t.Run("test authorization header", func(t *testing.T) {
		account := testutil.GetValidAccount().ToDTO()
		token, err := deps.Domains().Auth().CreateTokenForAccount(&account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token)
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err = middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.NotNil(t, c.GetAccount())
	})

	t.Run("test authorization cookie", func(t *testing.T) {
		account := model.AccountDTO{Username: "shiori"}
		token, err := deps.Domains().Auth().CreateTokenForAccount(&account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{
			Name:   "token",
			Value:  token,
			MaxAge: int(time.Now().Add(time.Minute).Unix()),
		})
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err = middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.NotNil(t, c.GetAccount())
	})

	t.Run("test invalid token cookie is removed", func(t *testing.T) {
		// Create an invalid token
		invalidToken := "invalid-token"

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{
			Name:   "token",
			Value:  invalidToken,
			MaxAge: int(time.Now().Add(time.Minute).Unix()),
		})
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())

		// Check that the token cookie was removed in the response
		responseCookies := w.Result().Cookies()

		var tokenCookie *http.Cookie
		for _, cookie := range responseCookies {
			if cookie.Name == "token" {
				tokenCookie = cookie
				break
			}
		}

		require.NotNil(t, tokenCookie, "Token cookie should exist in response")
		require.Empty(t, tokenCookie.Value, "Token cookie value should be empty")
	})
}

func TestAuthMiddlewareWithSSO(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)
	deps.Config().Http.SSOEnable = true

	account, err := deps.Domains().Accounts().CreateAccount(context.TODO(), model.AccountDTO{
		ID: model.DBID(98),
		Username: "test_username",
		Password: "super_secure_password",
	})
	require.NoError(t, err)

	t.Run("test no authorization method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())
	})

	t.Run("test authorization header", func(t *testing.T) {
		account := testutil.GetValidAccount().ToDTO()
		token, err := deps.Domains().Auth().CreateTokenForAccount(&account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token)
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err = middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.NotNil(t, c.GetAccount())
	})

	t.Run("test authorization cookie", func(t *testing.T) {
		account := model.AccountDTO{Username: "shiori"}
		token, err := deps.Domains().Auth().CreateTokenForAccount(&account, time.Now().Add(time.Minute))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{
			Name:   "token",
			Value:  token,
			MaxAge: int(time.Now().Add(time.Minute).Unix()),
		})
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err = middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.NotNil(t, c.GetAccount())
	})

	t.Run("test invalid token cookie is removed", func(t *testing.T) {
		// Create an invalid token
		invalidToken := "invalid-token"

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{
			Name:   "token",
			Value:  invalidToken,
			MaxAge: int(time.Now().Add(time.Minute).Unix()),
		})
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())

		// Check that the token cookie was removed in the response
		responseCookies := w.Result().Cookies()

		var tokenCookie *http.Cookie
		for _, cookie := range responseCookies {
			if cookie.Name == "token" {
				tokenCookie = cookie
				break
			}
		}

		require.NotNil(t, tokenCookie, "Token cookie should exist in response")
		require.Empty(t, tokenCookie.Value, "Token cookie value should be empty")
	})
	t.Run("test untrusted ip", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "invalid-ip"
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())
	})

	t.Run("test empty header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.3"
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())
	})

	t.Run("test invalid sso username", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.3"
		r.Header.Add("Remote-User", "username")
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())
	})

	t.Run("test sso login", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.3"
		r.Header.Add("Remote-User", account.Username)
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.NotNil(t, c.GetAccount())
	})
}

func TestRequireLoggedInUser(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("returns error when user not logged in", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := RequireLoggedInUser(deps, c)
		require.Error(t, err)
		require.Equal(t, "authentication required", err.Error())
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("succeeds when user is logged in", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		err := RequireLoggedInUser(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("succeeds when admin is logged in", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		err := RequireLoggedInUser(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequireLoggedInAdmin(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("returns error when user not logged in", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		err := RequireLoggedInAdmin(deps, c)
		require.Error(t, err)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns error when non-admin user is logged in", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		err := RequireLoggedInAdmin(deps, c)
		require.Error(t, err)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("succeeds when admin is logged in", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		err := RequireLoggedInAdmin(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)
	})
}
