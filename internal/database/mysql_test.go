//go:build !test_sqlite_only
// +build !test_sqlite_only

package database

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
)

func init() {
	connString := os.Getenv("SHIORI_TEST_MYSQL_URL")
	if connString == "" {
		log.Fatal("mysql tests can't run without a MysQL database, set SHIORI_TEST_MYSQL_URL environment variable")
	}
}

func mysqlTestDatabaseFactory(envKey string) testDatabaseFactory {
	return func(_ *testing.T, ctx context.Context) (model.DB, error) {
		connString := os.Getenv(envKey)
		db, err := OpenMySQLDatabase(ctx, connString)
		if err != nil {
			return nil, err
		}

		var dbname string
		err = db.withTx(ctx, func(tx *sqlx.Tx) error {
			err := tx.QueryRow("SELECT DATABASE()").Scan(&dbname)
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, "DROP DATABASE IF EXISTS "+dbname)
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, "CREATE DATABASE "+dbname)
			return err
		})
		if err != nil {
			return nil, err
		}

		if _, err := db.Exec("USE " + dbname); err != nil {
			return nil, err
		}

		if err = db.Migrate(context.TODO()); err != nil {
			return nil, err
		}

		return db, err
	}
}

func TestMysqlsDatabase(t *testing.T) {
	testDatabase(t, mysqlTestDatabaseFactory("SHIORI_TEST_MYSQL_URL"))
}

func TestMariaDBDatabase(t *testing.T) {
	testDatabase(t, mysqlTestDatabaseFactory("SHIORI_TEST_MARIADB_URL"))
}
