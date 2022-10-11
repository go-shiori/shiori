package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
)

var sqliteDatabaseTestPath string

func init() {
	sqliteDatabaseTestPath = filepath.Join(os.TempDir(), "shiori.db")
}

func sqliteTestDatabaseFactory(ctx context.Context) (DB, error) {
	os.Remove(sqliteDatabaseTestPath)

	db, err := OpenSQLiteDatabase(ctx, sqliteDatabaseTestPath)
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(); err != nil && !errors.Is(migrate.ErrNoChange, err) {
		return nil, err
	}

	return db, nil
}

func TestSqliteDatabase(t *testing.T) {
	testDatabase(t, sqliteTestDatabaseFactory)
}
