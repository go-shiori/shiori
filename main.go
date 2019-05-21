package main

import (
	"os"
	fp "path/filepath"

	"github.com/go-shiori/shiori/internal/cmd"
	"github.com/go-shiori/shiori/internal/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var dataDir = "dev-data"

func main() {
	// Make sure data dir exists
	os.MkdirAll(dataDir, os.ModePerm)

	// Open database
	dbPath := fp.Join(dataDir, "shiori.db")
	sqliteDB, err := database.OpenSQLiteDatabase(dbPath)
	if err != nil {
		logrus.Fatalln(err)
	}

	// Execute cmd
	cmd.DB = sqliteDB
	cmd.DataDir = dataDir
	if err := cmd.ShioriCmd().Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
