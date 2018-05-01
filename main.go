//go:generate go run assets-generator.go

package main

import (
	"github.com/RadhiFadlillah/shiori/cmd"
	dt "github.com/RadhiFadlillah/shiori/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var dbPath = "shiori.db"

func main() {
	// Open database
	sqliteDB, err := dt.OpenSQLiteDatabase(dbPath)
	checkError(err)

	// Start cmd
	shioriCmd := cmd.NewShioriCmd(sqliteDB)
	if err := shioriCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
