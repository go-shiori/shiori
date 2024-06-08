//go:build openbsd
// +build openbsd

package database

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	_ "git.sr.ht/~emersion/go-sqlite3-fts5"
	_ "github.com/mattn/go-sqlite3"
)

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenSQLiteDatabase(ctx context.Context, databasePath string) (sqliteDB *SQLiteDatabase, err error) {
	// Open database
	db, err := sqlx.ConnectContext(ctx, "sqlite3", databasePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqliteDB = &SQLiteDatabase{dbbase: dbbase{db}}
	return sqliteDB, nil
}
