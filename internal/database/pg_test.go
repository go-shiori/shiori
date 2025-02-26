//go:build !test_sqlite_only
// +build !test_sqlite_only

package database

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
)

func init() {
	connString := os.Getenv("SHIORI_TEST_PG_URL")
	if connString == "" {
		log.Fatal("psql tests can't run without a PSQL database, set SHIORI_TEST_PG_URL environment variable")
	}
}

func postgresqlTestDatabaseFactory(_ *testing.T, ctx context.Context) (model.DB, error) {
	db, err := OpenPGDatabase(ctx, os.Getenv("SHIORI_TEST_PG_URL"))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(context.TODO()); err != nil {
		return nil, err
	}

	return db, nil
}

func TestPostgresDatabase(t *testing.T) {
	testDatabase(t, postgresqlTestDatabaseFactory)
}
