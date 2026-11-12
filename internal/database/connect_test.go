package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnectDatabaseScheme verifies that Connect recognizes the database URL
// schemes it is supposed to support. In particular, the PostgreSQL driver
// (lib/pq) accepts both the "postgres://" and "postgresql://" URL prefixes, so
// Connect must route both to the PostgreSQL implementation instead of rejecting
// "postgresql://" as an unsupported scheme.
//
// See https://github.com/go-shiori/shiori/issues/1098
func TestConnectDatabaseScheme(t *testing.T) {
	// Point the connection at a loopback address with no listener so the
	// connection attempt fails fast and deterministically. We only care about
	// the scheme routing here, not about establishing a real connection.
	const unreachable = "127.0.0.1:1"

	t.Run("postgresql scheme is recognized", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := Connect(ctx, "postgresql://shiori:shiori@"+unreachable+"/shiori?sslmode=disable")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "unsupported database scheme",
			"postgresql:// should be routed to the PostgreSQL driver, not rejected as an unsupported scheme")
	})

	t.Run("postgres scheme is recognized", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := Connect(ctx, "postgres://shiori:shiori@"+unreachable+"/shiori?sslmode=disable")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "unsupported database scheme",
			"postgres:// should be routed to the PostgreSQL driver")
	})

	t.Run("unknown scheme is rejected", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := Connect(ctx, "mongodb://shiori:shiori@"+unreachable+"/shiori")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported database scheme",
			"unknown schemes must still be rejected")
	})
}
