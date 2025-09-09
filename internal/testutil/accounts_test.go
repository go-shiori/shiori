package testutil

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestNewAdminUser(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	_, deps := GetTestConfigurationAndDependencies(t, ctx, logger)

	t.Run("successful admin user creation", func(t *testing.T) {
		account, token, err := NewAdminUser(deps)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		require.NotNil(t, account)
		require.Equal(t, "admin", account.Username)
		require.True(t, *account.Owner)

		// Verify the token works
		tokenAccount, err := deps.Domains().Auth().CheckToken(ctx, token)
		require.NoError(t, err)
		require.NotNil(t, tokenAccount)
		require.Equal(t, account.ID, tokenAccount.ID)
		require.Equal(t, account.Username, tokenAccount.Username)
		require.True(t, *tokenAccount.Owner)
	})

	t.Run("duplicate admin user creation", func(t *testing.T) {
		// Try to create another admin user
		account, token, err := NewAdminUser(deps)
		require.Error(t, err)
		require.Empty(t, token)
		require.Nil(t, account)
	})
}
