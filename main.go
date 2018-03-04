//go:generate fileb0x filebox.json
package main

import (
	"./cmd"
	db "./database"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	sqliteDB, err := db.OpenSQLiteDatabase()
	checkError(err)

	cmd.DB = sqliteDB
	cmd.Execute()
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
