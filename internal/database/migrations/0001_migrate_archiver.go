package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	gap "github.com/muesli/go-app-paths"
)

// getPortableModeEnabled_171 checks if portable mode is enabled by naively checking the
// os.Args for the --portable flag. This is a workaround to use in this migration with the
// current state of the code as of 1.7.1.
func getPortableModeEnabled_171() bool {
	return slices.Contains(os.Args, "--portable")
}

// getStorageDirectory_170 returns the directory where shiori data is stored
// for the 1.7.1 version of shiori.
// This function is just a copy of the original as of 1.7.1.
func getStorageDirectory_171(portableMode bool) (string, error) {
	// If in portable mode, uses directory of executable
	if portableMode {
		exePath, err := os.Executable()
		if err != nil {
			return "", err
		}

		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, "shiori-data"), nil
	}

	// Try to use platform specific app path
	userScope := gap.NewScope(gap.User, "shiori")
	dataDir, err := userScope.DataPath("")
	if err == nil {
		return dataDir, nil
	}

	return "", fmt.Errorf("couldn't determine the data directory")
}

// getDataDir_171 returns the directory where shiori data is stored using the logic flow
// of the 1.7.1 version of shiori.
func getDataDir_171() (string, error) {
	dataDir := os.Getenv("SHIORI_DIR")
	if dataDir == "" {
		var err error
		dataDir, err = getStorageDirectory_171(getPortableModeEnabled_171())
		if err != nil {
			return "", fmt.Errorf("failed to get data directory: %w", err)
		}
	}
	return dataDir, nil
}

// MigrateArchiver adds new columns for the archiver and archiver_path
// This migration manually checks that the existing bookmarks have a file in the default archive path:
// SHIORI_DIR/archives/ID
// If the file exists, it will update the archiver=warc (the only one at this point) and archiver_path=path
// This migration is driver agnostic.
func MigrateArchiverMigration(sqlDB *sql.DB, driver string) error {
	var flavor sqlbuilder.Flavor
	switch driver {
	case "mysql":
		flavor = sqlbuilder.MySQL
	case "postgres":
		flavor = sqlbuilder.PostgreSQL
	case "sqlite":
		flavor = sqlbuilder.SQLite
	default:
		return fmt.Errorf("unsupported driver: %s", driver)
	}

	ctx := context.Background()
	sqlX := sqlx.NewDb(sqlDB, driver)

	tx, err := sqlX.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	perPage := 50
	page := 1

	for {
		var bookmarkIDs []int
		sb := sqlbuilder.NewSelectBuilder()
		sb.SetFlavor(flavor)
		sb.Select("id")
		sb.From("bookmark")
		sb.OrderBy("id ASC")
		sb.Where(sb.Equal("archiver", ""))
		sb.Limit(perPage)
		sb.Offset((page - 1) * perPage)

		sqlQuery, args := sb.Build()
		if err := sqlX.Select(&bookmarkIDs, sqlQuery, args...); err != nil {
			return fmt.Errorf("failed to get bookmarks: %w", err)
		}

		if len(bookmarkIDs) == 0 {
			break
		}

		dataDir, err := getDataDir_171()
		if err != nil {
			return fmt.Errorf("failed to get data directory: %w", err)
		}

		for _, bookID := range bookmarkIDs {
			archivePath := filepath.Join(dataDir, "archive", fmt.Sprintf("%d", bookID))

			// If the file exists, we assume it's a WARC file and update the row
			if _, err := os.Stat(archivePath); err == nil {
				sb := sqlbuilder.NewUpdateBuilder()
				sb.Update("bookmark")
				sb.Set(
					sb.Assign("archiver", "warc"),
					sb.Assign("archive_path", archivePath),
				)
				sb.Where(sb.Equal("id", bookID))

				sqlQuery, args := sb.Build()
				if _, err := tx.ExecContext(ctx, sqlQuery, args...); err != nil {
					return fmt.Errorf("failed to update bookmark %d: %w", bookID, err)
				}
			}
		}

		page++
	}

	return tx.Commit()
}
