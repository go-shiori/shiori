package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

func TestSaveAccountSettingsMySql(t *testing.T) {
	ctx := context.TODO()

	db, err := mysqlTestDatabaseFactory(ctx)
	assert.NoError(t, err)

	// Mock data
	account := model.Account{
		Username: "testuser",
		Config:   model.UserConfig{},
	}

	// Successful case
	err = db.SaveAccountSettings(ctx, account)
	assert.NoError(t, err)

	// Initialize not correct database
	ctx = context.TODO()
	factory := func(ctx context.Context) (DB, error) {
		return OpenSQLiteDatabase(ctx, filepath.Join(os.TempDir(), "shiori_test.db"))
	}
	db, err = factory(ctx)
	assert.Nil(t, err)
	account = model.Account{
		Username: "testuser",
		Config:   model.UserConfig{},
	}
	err = db.SaveAccountSettings(ctx, account)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SQL logic error: no such table: account (1)")
}

func TestGetAccountsMySql(t *testing.T) {
	ctx := context.TODO()

	db, err := mysqlTestDatabaseFactory(ctx)
	assert.NoError(t, err)

	// Insert test accounts
	testAccounts := []model.Account{
		{Username: "foo", Password: "bar", Owner: false},
		{Username: "hello", Password: "world", Owner: false},
		{Username: "foo_bar", Password: "foobar", Owner: true},
	}
	for _, acc := range testAccounts {
		err := db.SaveAccount(ctx, acc)
		assert.Nil(t, err)
	}

	// Successful case
	// without opt
	accounts, err := db.GetAccounts(ctx, GetAccountsOptions{})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(accounts))
	// with owner
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Owner: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(accounts))
	// with opt
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Keyword: "foo"})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(accounts))
	// with opt and owner
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Keyword: "hello", Owner: false})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(accounts))
	// with not result
	accounts, err = db.GetAccounts(ctx, GetAccountsOptions{Keyword: "shiori"})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(accounts))

	// Initialize not correct database
	ctx = context.TODO()
	factory := func(ctx context.Context) (DB, error) {
		return OpenSQLiteDatabase(ctx, filepath.Join(os.TempDir(), "shiori_test.db"))
	}
	db, err = factory(ctx)
	assert.Nil(t, err)
	// with invalid query
	opts := GetAccountsOptions{Keyword: "foo", Owner: true}
	_, err = db.GetAccounts(ctx, opts)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SQL logic error: no such table: account (1)")
}

func TestGetAccountMySql(t *testing.T) {
	ctx := context.TODO()

	db, err := mysqlTestDatabaseFactory(ctx)
	assert.NoError(t, err)

	// Insert test accounts
	testAccounts := []model.Account{
		{Username: "foo", Password: "bar", Owner: false},
		{Username: "hello", Password: "world", Owner: false},
		{Username: "foo_bar", Password: "foobar", Owner: true},
	}
	for _, acc := range testAccounts {
		err := db.SaveAccount(ctx, acc)
		assert.Nil(t, err)

		// Successful case
		account, exists, err := db.GetAccount(ctx, acc.Username)
		assert.Nil(t, err)
		assert.True(t, exists, "Expected account to exist")
		assert.Equal(t, acc.Username, account.Username)
	}
	// Falid case
	account, exists, err := db.GetAccount(ctx, "foobar")
	assert.NotNil(t, err)
	assert.False(t, exists, "Expected account to exist")
	assert.Empty(t, account.Username)
}
