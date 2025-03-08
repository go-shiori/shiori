// Package database implements database operations and migrations
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path"

	"github.com/blang/semver"
	"github.com/go-shiori/shiori/internal/model"
)

//go:embed migrations/*
var migrationFiles embed.FS

// migration represents a database schema migration
type migration struct {
	fromVersion   semver.Version
	toVersion     semver.Version
	migrationFunc func(db *sql.DB) error
}

// txFn is a function that runs in a transaction.
type txFn func(tx *sql.Tx) error

// runInTransaction runs the given function in a transaction.
func runInTransaction(db *sql.DB, fn txFn) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return fmt.Errorf("failed to run transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// newFuncMigration creates a new migration from a function.
func newFuncMigration(fromVersion, toVersion string, migrationFunc func(db *sql.DB) error) migration {
	return migration{
		fromVersion:   semver.MustParse(fromVersion),
		toVersion:     semver.MustParse(toVersion),
		migrationFunc: migrationFunc,
	}
}

// newFileMigration creates a new migration from a file.
func newFileMigration(fromVersion, toVersion, filename string) migration {
	return newFuncMigration(fromVersion, toVersion, func(db *sql.DB) error {
		return runInTransaction(db, func(tx *sql.Tx) error {
			migrationSQL, err := migrationFiles.ReadFile(path.Join("migrations", filename+".up.sql"))
			if err != nil {
				return fmt.Errorf("failed to read migration file: %w", err)
			}

			if _, err := tx.Exec(string(migrationSQL)); err != nil {
				return fmt.Errorf("failed to execute migration %s to %s: %w", fromVersion, toVersion, err)
			}
			return nil
		})
	})
}

// runMigrations runs the given migrations.
func runMigrations(ctx context.Context, db model.DB, migrations []migration) error {
	currentVersion := semver.Version{}

	// Get current database version
	dbVersion, err := db.GetDatabaseSchemaVersion(ctx)
	if err == nil && dbVersion != "" {
		currentVersion = semver.MustParse(dbVersion)
	}

	for _, migration := range migrations {
		if !currentVersion.EQ(migration.fromVersion) {
			continue
		}

		if err := migration.migrationFunc(db.WriterDB().DB); err != nil {
			return fmt.Errorf("failed to run migration from %s to %s: %w", migration.fromVersion, migration.toVersion, err)
		}

		currentVersion = migration.toVersion

		if err := db.SetDatabaseSchemaVersion(ctx, currentVersion.String()); err != nil {
			return fmt.Errorf("failed to store database version %s from %s to %s: %w", currentVersion.String(), migration.fromVersion, migration.toVersion, err)
		}
	}

	return nil
}
