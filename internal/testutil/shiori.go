package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func GetTestConfigurationAndDependencies(t *testing.T, ctx context.Context, logger *logrus.Logger) (*config.Config, *config.Dependencies) {
	tmp, err := os.CreateTemp("", "")
	require.NoError(t, err)

	cfg := config.ParseServerConfiguration(ctx, logger)
	cfg.Http.SecretKey = "test"

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	db, err := database.OpenSQLiteDatabase(ctx, tmp.Name())
	require.NoError(t, err)
	require.NoError(t, db.Migrate())

	cfg.Storage.DataDir = tempDir

	deps := config.NewDependencies(logger, db, cfg)
	deps.Database = db
	deps.Domains.Auth = domains.NewAccountsDomain(logger, cfg.Http.SecretKey, db)
	deps.Domains.Archiver = domains.NewArchiverDomain(logger, cfg.Storage.DataDir)

	return cfg, deps
}
