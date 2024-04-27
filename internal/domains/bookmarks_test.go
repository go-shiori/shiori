package domains_test

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestBookmarkDomain(t *testing.T) {
	fs := afero.NewMemMapFs()

	db, err := database.OpenSQLiteDatabase(context.TODO(), ":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(context.TODO()))

	deps := &dependencies.Dependencies{
		Database: db,
		Config:   config.ParseServerConfiguration(context.TODO(), logrus.New()),
		Log:      logrus.New(),
		Domains:  &dependencies.Domains{},
	}
	deps.Domains.Storage = domains.NewStorageDomain(deps, fs)

	fs.MkdirAll("thumb", 0755)
	fs.Create("thumb/1")
	fs.MkdirAll("ebook", 0755)
	fs.Create("ebook/1.epub")
	fs.MkdirAll("archive", 0755)
	// TODO: write a valid archive file
	fs.Create("archive/1")

	domain := domains.NewBookmarksDomain(deps)
	t.Run("HasEbook", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasEbook(&model.BookmarkDTO{ID: 1}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasEbook(&model.BookmarkDTO{ID: 2}))
		})
	})

	t.Run("HasArchive", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasArchive(&model.BookmarkDTO{ID: 1}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasArchive(&model.BookmarkDTO{ID: 2}))
		})
	})

	t.Run("HasThumbnail", func(t *testing.T) {
		t.Run("Yes", func(t *testing.T) {
			require.True(t, domain.HasThumbnail(&model.BookmarkDTO{ID: 1}))
		})
		t.Run("No", func(t *testing.T) {
			require.False(t, domain.HasThumbnail(&model.BookmarkDTO{ID: 2}))
		})
	})

	t.Run("GetBookmark", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			_, err := deps.Database.SaveBookmarks(context.TODO(), true, *testutil.GetValidBookmark())
			require.NoError(t, err)
			bookmark, err := domain.GetBookmark(context.Background(), 1)
			require.NoError(t, err)
			require.Equal(t, 1, bookmark.ID)

			// Check DTO attributes
			require.True(t, bookmark.HasEbook)
			require.True(t, bookmark.HasArchive)
		})
	})
}
