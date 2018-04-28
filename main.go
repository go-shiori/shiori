//go:generate go run assets-generator.go

package main

import (
	"os"
	"os/user"
	fp "path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/RadhiFadlillah/shiori/cmd"
	dt "github.com/RadhiFadlillah/shiori/database"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	databasePath := fp.Join(getHomeDir(), ".shiori.db")
	if value, found := os.LookupEnv("ENV_SHIORI_DB"); found {
		// If ENV_SHIORI_DB is directory, append ".shiori.db" as filename
		if f1, err := os.Stat(value); err == nil && f1.IsDir() {
			value = fp.Join(value, ".shiori.db")
		}

		databasePath = value
	}

	sqliteDB, err := dt.OpenSQLiteDatabase(databasePath)
	checkError(err)

	shioriCmd := cmd.NewShioriCmd(sqliteDB)
	if err := shioriCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
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
