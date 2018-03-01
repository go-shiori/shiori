package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	db "github.com/RadhiFadlillah/shiori/database"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// DB is database that used by this cli
	DB db.Database

	rootCmd = &cobra.Command{
		Use:   "shiori",
		Short: "Simple command-line bookmark manager built with Go.",
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init () {
	databasePath := "shiori.db"
	if value, found := os.LookupEnv("ENV_SHIORI_DB"); found {
		databasePath = value
	}
	sqliteDB, err := db.OpenSQLiteDatabase(databasePath)
	checkError(err)

	DB = sqliteDB
}