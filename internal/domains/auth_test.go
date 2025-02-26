package domains_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAuthDomainCheckToken(t *testing.T) {
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

	t.Run("nil account", func(t *testing.T) {
		token, err := domain.CreateTokenForAccount(nil, time.Now().Add(time.Hour))
		require.Error(t, err)
		require.Empty(t, token)
		require.Contains(t, err.Error(), "account is nil")
	})
}

func TestAuthDomainCheckTokenInvalidMethod(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAuthDomain(deps)

	// Create a token with an unsupported signing method
	account := testutil.GetValidAccount().ToDTO()
	claims := jwt.MapClaims{
		"account": account,
		"exp":     time.Now().Add(time.Hour).UTC().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Try to verify the token
	acc, err := domain.CheckToken(ctx, tokenString)
	require.Error(t, err)
	require.Nil(t, acc)
	require.Contains(t, err.Error(), "Unexpected signing method")
}

func TestAuthDomainGetAccountFromCredentials(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewAuthDomain(deps)

	_, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
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
