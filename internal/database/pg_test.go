package database

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

func init() {
	connString := os.Getenv("SHIORI_TEST_PG_URL")
	if connString == "" {
		log.Fatal("psql tests can't run without a PSQL database, set SHIORI_TEST_PG_URL environment variable")
	}
}

func postgresqlTestDatabaseFactory(ctx context.Context) (DB, error) {
	db, err := OpenPGDatabase(ctx, os.Getenv("SHIORI_TEST_PG_URL"))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(); err != nil && !errors.Is(migrate.ErrNoChange, err) {
		return nil, err
	}

	return db, nil
}

func TestPostgresDatabase(t *testing.T) {
	testDatabase(t, postgresqlTestDatabaseFactory)
}
