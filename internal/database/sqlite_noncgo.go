//go:build linux || windows || darwin || freebsd
// +build linux windows darwin freebsd

package database

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	_ "modernc.org/sqlite"
)

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenSQLiteDatabase(ctx context.Context, databasePath string) (sqliteDB *SQLiteDatabase, err error) {
	// Open database
	db, err := sqlx.ConnectContext(ctx, "sqlite", databasePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqliteDB = &SQLiteDatabase{dbbase: dbbase{db}}
	return sqliteDB, nil
}
