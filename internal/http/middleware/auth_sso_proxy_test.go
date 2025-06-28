package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-shiori/shiori/internal/http/webcontext"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddlewareWithSSO(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)
	deps.Config().Http.SSOProxyAuth = true

	account, err := deps.Domains().Accounts().CreateAccount(context.TODO(), model.AccountDTO{
		ID:       model.DBID(98),
		Username: "test_username",
		Password: "super_secure_password",
	})
	require.NoError(t, err)

	t.Run("test no authorization method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthSSOProxyMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())
	})

	t.Run("test untrusted ip", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "invalid-ip"
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthSSOProxyMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Nil(t, c.GetAccount())
	})

	t.Run("test empty header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.3"
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthSSOProxyMiddleware(deps)
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

		middleware := NewAuthSSOProxyMiddleware(deps)
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

		middleware := NewAuthSSOProxyMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.NotNil(t, c.GetAccount())
	})

	t.Run("test sso login ip:port", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.3:65342"
		r.Header.Add("Remote-User", account.Username)
		c := webcontext.NewWebContext(w, r)

		middleware := NewAuthSSOProxyMiddleware(deps)
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.NotNil(t, c.GetAccount())
	})
}
