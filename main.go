//go:generate go run assets-generator.go

package main

import (
	"github.com/go-shiori/shiori/internal/cmd"
	"github.com/sirupsen/logrus"
	// Database driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	// Add this to prevent it removed by go mod tidy
	_ "github.com/shurcooL/vfsgen"
)

func main() {
	err := cmd.ShioriCmd().Execute()
	if err != nil {
		logrus.Fatalln(err)
	}
}
