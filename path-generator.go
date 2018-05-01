// +build !dev

package main

import (
	"os"
	fp "path/filepath"

	apppaths "github.com/muesli/go-app-paths"
)

func init() {
	// Set database path
	dbPath = createDatabasePath()

	// Make sure directory exist
	os.MkdirAll(fp.Dir(dbPath), os.ModePerm)
}

func createDatabasePath() string {
	// Try to look at environment variables
	dbPath, found := os.LookupEnv("ENV_SHIORI_DB")
	if found {
		// If ENV_SHIORI_DB is directory, append "shiori.db" as filename
		if f1, err := os.Stat(dbPath); err == nil && f1.IsDir() {
			dbPath = fp.Join(dbPath, "shiori.db")
		}

		return dbPath
	}

	// Try to use platform specific app path
	userScope := apppaths.NewScope(apppaths.User, "shiori", "shiori")
	dataDir, err := userScope.DataDir()
	if err == nil {
		return fp.Join(dataDir, "shiori.db")
	}

	// When all fail, create database in working directory
	return "shiori.db"
}
