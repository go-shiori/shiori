package database

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
)

func init() {
	testPsqlURL := os.Getenv("SHIORI_TEST_PG_URL")
	if testPsqlURL == "" {
		log.Fatal("psql tests can't run without a PSQL database")
	}
}

func TestPsqlSaveBookmarkWithTag(t *testing.T) {
	ctx := context.TODO()
	pgDB, err := OpenPGDatabase(ctx, os.Getenv("SHIORI_TEST_PG_URL"))
	if err != nil {
		t.Error(err)
	}

	if err := pgDB.Migrate(); err != nil && !errors.Is(migrate.ErrNoChange, err) {
		t.Error(err)
	}

	book := model.Bookmark{
		URL:   "https://github.com/go-shiori/obelisk",
		Title: "shiori",
		Tags: []model.Tag{
			{
				Name: "test-tag",
			},
		},
	}

	result, err := pgDB.SaveBookmarks(ctx, book)

	assert.NoError(t, err, "Save bookmarks must not fail")
	assert.Equal(t, book.URL, result[0].URL)
	assert.Equal(t, book.Tags[0].Name, result[0].Tags[0].Name)
}
