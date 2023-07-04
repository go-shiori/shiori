package context

import (
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/require"
)

func TestUserIsLogged(t *testing.T) {
	t.Run("test user is logged", func(t *testing.T) {
		c := New()
		c.Set(model.ContextAccountKey, "test")
		require.True(t, c.UserIsLogged())
	})

	t.Run("test user is not logged", func(t *testing.T) {
		c := New()
		require.False(t, c.UserIsLogged())
	})
}

func TestGetAccount(t *testing.T) {
	t.Run("test get account (logged in)", func(t *testing.T) {
		account := model.Account{
			Username: "shiori",
		}
		c := New()
		c.Set(model.ContextAccountKey, &account)
		require.Equal(t, account, *c.GetAccount())
	})

	t.Run("test get account (not logged in)", func(t *testing.T) {
		c := New()
		require.Nil(t, c.GetAccount())
	})
}
