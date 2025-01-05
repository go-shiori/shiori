//go:build linux || windows || darwin || freebsd
// +build linux windows darwin freebsd

package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenSQLiteDatabase(ctx context.Context, databasePath string) (sqliteDB *SQLiteDatabase, err error) {
	// Open database
	rwDB, err := sqlx.ConnectContext(ctx, "sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("error opening writer database: %w", err)
	}

	rDB, err := sqlx.ConnectContext(ctx, "sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("error opening reader database: %w", err)
	}

	sqliteDB = &SQLiteDatabase{
		writer: &dbbase{rwDB},
		reader: &dbbase{rDB},
	}

	if err := sqliteDB.Init(ctx); err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	return sqliteDB, nil
}
