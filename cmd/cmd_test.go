package cmd

import (
	"fmt"
	"os"
	"testing"

	db "github.com/RadhiFadlillah/shiori/database"
	_ "github.com/mattn/go-sqlite3"
)

func TestMain(m *testing.M) {
	testDBFile := "shiori_test.db"
	sqliteDB, err := db.OpenSQLiteDatabase(testDBFile)
	if err != nil {
		fmt.Printf("failed to create tests DB: %v", err)
		os.Exit(1)
	}
	DB = sqliteDB

	code := m.Run()

	if err := os.Remove(testDBFile); err != nil {
		fmt.Printf("failed to delete tests DB: %v", err)
	}
	os.Exit(code)

}
