package main

import (
	"github.com/go-shiori/shiori/internal/cmd"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"

	// Add this to prevent it removed by go mod tidy
	_ "github.com/shurcooL/vfsgen"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	// Set globally
	model.BuildVersion = version
	model.BuildCommit = commit
	model.BuildDate = date
}

// @title						Shiori API
// @version					1.0
// @description				Shiori is a simple bookmarks manager. This is the documentation for its HTTP API.
// @BasePath					/
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and the JWT token.
func main() {
	err := cmd.ShioriCmd().Execute()
	if err != nil {
		logrus.Fatalln(err)
	}
}
