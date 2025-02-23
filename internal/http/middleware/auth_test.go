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

func TestRequireAuthMiddleware(t *testing.T) {
	t.Run("test unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		logger := logrus.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

		middleware := NewRequireAuthMiddleware()
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, w.Code)

		resp, err := testutil.NewTestResponseFromRecorder(w)
		require.NoError(t, err)
		require.True(t, resp.Response.IsError())
	})

	t.Run("test authorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		logger := logrus.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

		c.SetAccount(&model.AccountDTO{Username: "test"})

		middleware := NewRequireAuthMiddleware()
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)
	})
}

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
}

func TestRequireAdminMiddleware(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("test unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)

		middleware := NewRequireAdminMiddleware()
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("test user but not admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)
		c.SetAccount(&model.AccountDTO{Owner: model.Ptr(false)})

		middleware := NewRequireAdminMiddleware()
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("test authorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := webcontext.NewWebContext(w, r)
		c.SetAccount(&model.AccountDTO{Owner: model.Ptr(true)})

		middleware := NewRequireAdminMiddleware()
		err := middleware.OnRequest(deps, c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)
	})
}
