package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
)

func GetTestConfigurationAndDependencies(t *testing.T, ctx context.Context, logger *logrus.Logger) (*config.Config, *dependencies.Dependencies) {
	t.Helper()

	tmp, err := os.CreateTemp("", "")
	require.NoError(t, err)

	cfg := config.ParseServerConfiguration(ctx, logger)
	cfg.Http.SecretKey = []byte("test")

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	db, err := database.OpenSQLiteDatabase(ctx, tmp.Name())
	require.NoError(t, err)
	require.NoError(t, db.Migrate())

	cfg.Storage.DataDir = tempDir

	deps := dependencies.NewDependencies(logger, db, cfg)
	deps.Database = db
	deps.Domains.Auth = domains.NewAccountsDomain(deps)
	deps.Domains.Archiver = domains.NewArchiverDomain(deps)
	deps.Domains.Bookmarks = domains.NewBookmarksDomain(deps)
	deps.Domains.Storage = domains.NewStorageDomain(deps, cfg.Storage.DataDir)

	return cfg, deps
}

func GetValidBookmark() *model.BookmarkDTO {
	uuidV4, _ := uuid.NewV4()
	return &model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori#" + uuidV4.String(),
		Title: "Shiori repository",
	}
}
