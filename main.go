//go:generate fileb0x filebox.json
package main

import (
	"os"
	"os/user"
	fp "path/filepath"

	"github.com/RadhiFadlillah/shiori/cmd"
	db "github.com/RadhiFadlillah/shiori/database"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// check and use mysql if env values set
	if mysqlDBName := os.Getenv("SHIORI_MYSQL_DBNAME"); mysqlDBName != "" {
		mysqlDBUser := os.Getenv("SHIORI_MYSQL_USER")
		mysqlDBPass := os.Getenv("SHIORI_MYSQL_PASS")
		mysqlDB, err := db.OpenMySQLDatabase(mysqlDBUser, mysqlDBPass, mysqlDBName)
		checkError(err)
		cmd.DB = mysqlDB
		cmd.Execute()
		return
	}

	databasePath := fp.Join(getHomeDir(), ".shiori.db")
	if value, found := os.LookupEnv("ENV_SHIORI_DB"); found {
		// If ENV_SHIORI_DB is directory, append ".shiori.db" as filename
		if f1, err := os.Stat(value); err == nil && f1.IsDir() {
			value = fp.Join(value, ".shiori.db")
		}

		databasePath = value
	}

	sqliteDB, err := db.OpenSQLiteDatabase(databasePath)
	checkError(err)

	cmd.DB = sqliteDB
	cmd.Execute()
}

func getHomeDir() string {
	user, err := user.Current()
	if err != nil {
		return ""
	}

	return user.HomeDir
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
