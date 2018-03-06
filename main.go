//go:generate fileb0x filebox.json
package main

import (
	"os"

	"github.com/RadhiFadlillah/shiori/cmd"
	db "github.com/RadhiFadlillah/shiori/database"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	databasePath := "shiori.db"
	if value, found := os.LookupEnv("ENV_SHIORI_DB"); found {
		databasePath = value + "/" + databasePath
	}

	sqliteDB, err := db.OpenSQLiteDatabase(databasePath)
	checkError(err)

	cmd.DB = sqliteDB
	cmd.Execute()
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
