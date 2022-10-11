package database

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func init() {
	connString := os.Getenv("SHIORI_TEST_MYSQL_URL")
	if connString == "" {
		log.Fatal("mysql tests can't run without a MysQL database, set SHIORI_TEST_MYSQL_URL environment variable")
	}
}

func mysqlTestDatabaseFactory(ctx context.Context) (DB, error) {
	connString := os.Getenv("SHIORI_TEST_MYSQL_URL")
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

	if err = db.Migrate(); err != nil && !errors.Is(migrate.ErrNoChange, err) {
		return nil, err
	}

	return db, err
}

func TestMysqlsDatabase(t *testing.T) {
	testDatabase(t, mysqlTestDatabaseFactory)
}
