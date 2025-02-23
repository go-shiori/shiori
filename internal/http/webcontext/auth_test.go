package webcontext

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/require"
)

func TestUserIsLogged(t *testing.T) {
	t.Run("test user is logged", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := NewWebContext(w, r)

		c.SetAccount(&model.AccountDTO{Username: "test"})

		require.True(t, c.UserIsLogged())
		account := c.GetAccount()
		require.NotNil(t, account)
		require.Equal(t, "test", account.Username)
	})

	t.Run("test user is not logged", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := NewWebContext(w, r)

		require.False(t, c.UserIsLogged())
		require.Nil(t, c.GetAccount())
	})
}

func TestGetAccount(t *testing.T) {
	t.Run("test get account (logged in)", func(t *testing.T) {
		account := model.AccountDTO{
			Username: "shiori",
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := NewWebContext(w, r)

		c.SetAccount(&account)
		gotAccount := c.GetAccount()

		require.NotNil(t, gotAccount)
		require.Equal(t, account, *gotAccount)
	})

	t.Run("test get account (not logged in)", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		c := NewWebContext(w, r)

		require.Nil(t, c.GetAccount())
	})
}

func TestWithAccount(t *testing.T) {
	account := &model.AccountDTO{
		Username: "shiori",
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := NewWebContext(w, r)

	c.SetAccount(account)
	gotAccount := c.GetAccount()

	require.Equal(t, account, gotAccount)
}
