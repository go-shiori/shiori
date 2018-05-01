//go:generate go run assets-generator.go

package main

import (
	"os"
	fp "path/filepath"

	"github.com/RadhiFadlillah/shiori/cmd"
	dt "github.com/RadhiFadlillah/shiori/database"
	_ "github.com/mattn/go-sqlite3"
	apppaths "github.com/muesli/go-app-paths"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create database path
	dbPath := createDatabasePath()

	// Make sure directory exist
	os.MkdirAll(fp.Dir(dbPath), os.ModePerm)

	// Open database
	sqliteDB, err := dt.OpenSQLiteDatabase(dbPath)
	checkError(err)

	// Start cmd
	shioriCmd := cmd.NewShioriCmd(sqliteDB)
	if err := shioriCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
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

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
