package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofrs/uuid/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func GetTestConfigurationAndDependencies(t *testing.T, ctx context.Context, logger *logrus.Logger) (*config.Config, *dependencies.Dependencies) {
	t.Helper()

	tmp, err := os.CreateTemp("", "")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(tmp.Name())
	})

	cfg := config.ParseServerConfiguration(ctx, logger)
	cfg.Http.SecretKey = []byte("test")

	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	db, err := database.OpenSQLiteDatabase(ctx, tmp.Name())
	require.NoError(t, err)
	require.NoError(t, db.Migrate(context.TODO()))

	cfg.Storage.DataDir = tmpDir

	deps := dependencies.NewDependencies(logger, db, cfg)
	deps.Domains().SetAccounts(domains.NewAccountsDomain(deps))
	deps.Domains().SetArchiver(domains.NewArchiverDomain(deps))
	deps.Domains().SetAuth(domains.NewAuthDomain(deps))
	deps.Domains().SetBookmarks(domains.NewBookmarksDomain(deps))
	deps.Domains().SetStorage(domains.NewStorageDomain(deps, afero.NewBasePathFs(afero.NewOsFs(), cfg.Storage.DataDir)))
	deps.Domains().SetTags(domains.NewTagsDomain(deps))

	return cfg, deps
}

func GetValidBookmark() *model.BookmarkDTO {
	uuidV4, _ := uuid.NewV4()
	return &model.BookmarkDTO{
		URL:   "https://github.com/go-shiori/shiori#" + uuidV4.String(),
		Title: "Shiori repository",
	}
}

// GetValidAccount returns a valid account for testing
// It includes an ID to properly use the account when testing authentication methods
// without interacting with the database.
func GetValidAccount() *model.Account {
	return &model.Account{
		ID:       99,
		Username: "test",
		Password: "test",
	}
}
