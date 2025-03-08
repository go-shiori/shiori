package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// ErrNotFound is error returned when record is not found in database.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is error returned when record already exists in database.
var ErrAlreadyExists = errors.New("already exists")

// Connect connects to database based on submitted database URL.
func Connect(ctx context.Context, dbURL string) (model.DB, error) {
	dbU, err := url.Parse(dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse database URL")
	}

	switch dbU.Scheme {
	case "mysql":
		urlNoSchema := strings.Split(dbURL, "://")[1]
		return OpenMySQLDatabase(ctx, urlNoSchema)
	case "postgres":
		return OpenPGDatabase(ctx, dbURL)
	case "sqlite":
		return OpenSQLiteDatabase(ctx, dbU.Path[1:])
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", dbU.Scheme)
}

type dbbase struct {
	*sqlx.DB
}

func (db *dbbase) withTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		if err := tx.Commit(); err != nil {
			log.Printf("error during commit: %s", err)
		}
	}()

	err = fn(tx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("error during rollback: %s", err)
		}
		return errors.WithStack(err)
	}

	return err
}
