package database

import (
	"github.com/jmoiron/sqlx"
)

// SQLiteDatabase is implementation of Database interface
// for connecting to SQLite3 database.
type SQLiteDatabase struct {
	sqlx.DB
}
