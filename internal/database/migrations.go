package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/blang/semver"
)

//go:embed migrations/*
var migrationFiles embed.FS

type migration struct {
	fromVersion   semver.Version
	toVersion     semver.Version
	migrationFunc func(tx *sql.Tx) error
}

func newFuncMigration(fromVersion, toVersion string, migrationFunc func(tx *sql.Tx) error) migration {
	return migration{
		fromVersion:   semver.MustParse(fromVersion),
		toVersion:     semver.MustParse(toVersion),
		migrationFunc: migrationFunc,
	}
}

func newFileMigration(fromVersion, toVersion, filename string) migration {
	return newFuncMigration(fromVersion, toVersion, func(tx *sql.Tx) error {
		migrationSQL, err := migrationFiles.ReadFile(filepath.Join("migrations", filename+".up.sql"))
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		if _, err := tx.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration %s to %s: %w", fromVersion, toVersion, err)
		}

		return nil
	})
}

func runMigrations(ctx context.Context, db DB, migrations []migration) error {
	currentVersion := semver.Version{}

	// Get current database version
	dbVersion, err := db.GetDatabaseVersion(ctx)
	if err == nil && dbVersion != "" {
		currentVersion = semver.MustParse(dbVersion)
	}

	for _, migration := range migrations {
		if !currentVersion.EQ(migration.fromVersion) {
			continue
		}

		tx, err := db.DBx().BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start migration transaction from %s to %s: %w", migration.fromVersion, migration.toVersion, err)
		}
		defer tx.Rollback()

		if err := migration.migrationFunc(tx); err != nil {
			return fmt.Errorf("failed to run migration from %s to %s: %w", migration.fromVersion, migration.toVersion, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit migration from %s to %s: %w", migration.fromVersion, migration.toVersion, err)
		}

		currentVersion = migration.toVersion

		if err := db.SetDatabaseVersion(ctx, currentVersion.String()); err != nil {
			return fmt.Errorf("failed to store database version %s from %s to %s: %w", currentVersion.String(), migration.fromVersion, migration.toVersion, err)
		}
	}

	return nil
}
