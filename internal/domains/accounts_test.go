package domains_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAccountsDomainParseToken(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAccountsDomain(deps)

	t.Run("valid token", func(t *testing.T) {
		// Create a valid token
		token, err := domain.CreateTokenForAccount(
			testutil.GetValidAccount(),
			time.Now().Add(time.Hour*1),
		)
		require.NoError(t, err)

		claims, err := domain.ParseToken(token)
		require.NoError(t, err)
		require.NotNil(t, claims)
		require.Equal(t, 99, claims.Account.ID)
	})

	t.Run("invalid token", func(t *testing.T) {
		claims, err := domain.ParseToken("invalid-token")
		require.Error(t, err)
		require.Nil(t, claims)
	})
}

func TestAccountsDomainCheckToken(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAccountsDomain(deps)

	t.Run("valid token", func(t *testing.T) {
		// Create a valid token
		token, err := domain.CreateTokenForAccount(
			testutil.GetValidAccount(),
			time.Now().Add(time.Hour*1),
		)
		require.NoError(t, err)

		acc, err := domain.CheckToken(ctx, token)
		require.NoError(t, err)
		require.NotNil(t, acc)
		require.Equal(t, 99, acc.ID)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create an expired token
		token, err := domain.CreateTokenForAccount(
			testutil.GetValidAccount(),
			time.Now().Add(time.Hour*-1),
		)
		require.NoError(t, err)

		acc, err := domain.CheckToken(ctx, token)
		require.Error(t, err)
		require.Nil(t, acc)
	})
}

func TestAccountsDomainGetAccountFromCredentials(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAccountsDomain(deps)

	require.NoError(t, deps.Database.SaveAccount(ctx, model.Account{
		Username: "test",
		Password: "test",
	}))

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

func TestAccountsDomainCreateTokenForAccount(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAccountsDomain(deps)

	t.Run("valid account", func(t *testing.T) {
		token, err := domain.CreateTokenForAccount(
			testutil.GetValidAccount(),
			time.Now().Add(time.Hour*1),
		)
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})

	t.Run("nil account", func(t *testing.T) {
		token, err := domain.CreateTokenForAccount(
			nil,
			time.Now().Add(time.Hour*1),
		)
		require.Error(t, err)
		require.Empty(t, token)
	})

	t.Run("token expiration is valid", func(t *testing.T) {
		expiration := time.Now().Add(time.Hour * 9)
		token, err := domain.CreateTokenForAccount(
			testutil.GetValidAccount(),
			expiration,
		)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		claims, err := domain.ParseToken(token)
		require.NoError(t, err)
		require.NotNil(t, claims)
		require.Equal(t, expiration.Unix(), claims.ExpiresAt.Time.Unix())
	})
}
