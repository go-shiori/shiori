package domains_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAccountDomainsListAccounts(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("empty", func(t *testing.T) {
		accounts, err := deps.Domains.Accounts.ListAccounts(context.Background())
		require.NoError(t, err)
		require.Empty(t, accounts)
	})

	t.Run("some accounts", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			_, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
				Username: fmt.Sprintf("user%d", i),
				Password: fmt.Sprintf("password%d", i),
			})
			require.NoError(t, err)
		}

		accounts, err := deps.Domains.Accounts.ListAccounts(context.Background())
		require.NoError(t, err)
		require.Len(t, accounts, 3)
		require.Equal(t, "", accounts[0].Password)
	})
}

func TestAccountDomainCreateAccount(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("create account", func(t *testing.T) {
		acc, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
			Username: "user",
			Password: "password",
			Owner:    model.Ptr(true),
			Config: &model.UserConfig{
				Theme: "dark",
			},
		})
		require.NoError(t, err)
		require.NotZero(t, acc.ID)
		require.Equal(t, "user", acc.Username)
		require.Equal(t, "dark", acc.Config.Theme)
	})

	t.Run("create account with empty username", func(t *testing.T) {
		_, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
			Username: "",
			Password: "password",
		})
		require.Error(t, err)
		_, isValidationErr := err.(model.ValidationError)
		require.True(t, isValidationErr)
	})

	t.Run("create account with empty password", func(t *testing.T) {
		_, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
			Username: "user",
			Password: "",
		})
		require.Error(t, err)
		_, isValidationErr := err.(model.ValidationError)
		require.True(t, isValidationErr)
	})
}

func TestAccountDomainUpdateAccount(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("update account", func(t *testing.T) {
		acc, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
			Username: "user",
			Password: "password",
		})
		require.NoError(t, err)

		acc, err = deps.Domains.Accounts.UpdateAccount(context.TODO(), model.AccountDTO{
			ID:       acc.ID,
			Username: "user2",
			Password: "password2",
			Owner:    model.Ptr(true),
			Config: &model.UserConfig{
				Theme: "light",
			},
		})
		require.NoError(t, err)
		require.Equal(t, "user2", acc.Username)
		require.Equal(t, "light", acc.Config.Theme)
	})

	t.Run("update non-existing account", func(t *testing.T) {
		_, err := deps.Domains.Accounts.UpdateAccount(context.TODO(), model.AccountDTO{
			ID:       999,
			Username: "user",
			Password: "password",
		})
		require.Error(t, err)
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("try to update with no changes", func(t *testing.T) {
		acc, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
			Username: "user",
			Password: "password",
		})
		require.NoError(t, err)

		_, err = deps.Domains.Accounts.UpdateAccount(context.TODO(), model.AccountDTO{
			ID: acc.ID,
		})
		require.Error(t, err)
		_, isValidationErr := err.(model.ValidationError)
		require.True(t, isValidationErr)
	})
}

func TestAccountsDomainCheckToken(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAuthDomain(deps)

	t.Run("valid token", func(t *testing.T) {
		// Create a valid token
		account := testutil.GetValidAccount().ToDTO()
		token, err := domain.CreateTokenForAccount(
			&account,
			time.Now().Add(time.Hour*1),
		)
		require.NoError(t, err)

		acc, err := domain.CheckToken(ctx, token)
		require.NoError(t, err)
		require.NotNil(t, acc)
		require.Equal(t, model.DBID(99), acc.ID)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create an expired token
		account := testutil.GetValidAccount().ToDTO()
		token, err := domain.CreateTokenForAccount(
			&account,
			time.Now().Add(time.Hour*-1),
		)
		require.NoError(t, err)

		acc, err := domain.CheckToken(ctx, token)
		require.Error(t, err)
		require.Nil(t, acc)
	})

	t.Run("invalid token", func(t *testing.T) {
		claims, err := domain.CheckToken(ctx, "invalid-token")
		require.Error(t, err)
		require.Nil(t, claims)
	})
}

func TestAccountsDomainGetAccountFromCredentials(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAuthDomain(deps)

	_, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
		Username: "test",
		Password: "test",
	})
	require.NoError(t, err)

	t.Run("valid credentials", func(t *testing.T) {
		acc, err := domain.GetAccountFromCredentials(ctx, "test", "test")
		require.NoError(t, err)
		require.NotNil(t, acc)
		require.Equal(t, "test", acc.Username)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		acc, err := domain.GetAccountFromCredentials(ctx, "test", "invalid")
		require.Error(t, err)
		require.Nil(t, acc)
	})

	t.Run("invalid username", func(t *testing.T) {
		acc, err := domain.GetAccountFromCredentials(ctx, "nope", "invalid")
		require.Error(t, err)
		require.Nil(t, acc)
	})

}

func TestAccountDomainDeleteAccount(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("delete account", func(t *testing.T) {
		acc, err := deps.Domains.Accounts.CreateAccount(context.TODO(), model.AccountDTO{
			Username: "user",
			Password: "password",
		})
		require.NoError(t, err)

		err = deps.Domains.Accounts.DeleteAccount(context.TODO(), int(acc.ID))
		require.NoError(t, err)

		accounts, err := deps.Domains.Accounts.ListAccounts(context.Background())
		require.NoError(t, err)
		require.Empty(t, accounts)
	})

	t.Run("delete non-existing account", func(t *testing.T) {
		err := deps.Domains.Accounts.DeleteAccount(context.TODO(), 999)
		require.Error(t, err)
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("valid account", func(t *testing.T) {
		account := testutil.GetValidAccount().ToDTO()
		token, err := deps.Domains.Auth.CreateTokenForAccount(
			&account,
			time.Now().Add(time.Hour*1),
		)
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})

	t.Run("nil account", func(t *testing.T) {
		token, err := deps.Domains.Auth.CreateTokenForAccount(
			nil,
			time.Now().Add(time.Hour*1),
		)
		require.Error(t, err)
		require.Empty(t, token)
	})

	t.Run("token expiration is valid", func(t *testing.T) {
		ctx := context.TODO()
		account := testutil.GetValidAccount().ToDTO()
		expiration := time.Now().Add(time.Hour * 9)
		token, err := deps.Domains.Auth.CreateTokenForAccount(
			&account,
			expiration,
		)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		tokenAccount, err := deps.Domains.Auth.CheckToken(ctx, token)
		require.NoError(t, err)
		require.NotNil(t, tokenAccount)
	})
}
