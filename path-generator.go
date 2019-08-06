// +build !dev

package main

import (
	"os"

	apppaths "github.com/muesli/go-app-paths"
)

func init() {
	// Get data directory
	dataDir = getDataDirectory()

	// Make sure directory exist
	os.MkdirAll(dataDir, os.ModePerm)
}

func getDataDirectory() string {
	// Try to look at environment variables
	dataDir, found := os.LookupEnv("SHIORI_DIR")
	if found {
		return dataDir
	}

	// Try to use platform specific app path
	userScope := apppaths.NewScope(apppaths.User, "shiori", "shiori")
	dataDir, err := userScope.DataDir()
	if err == nil {
		return dataDir
	}

	// When all fail, use current working directory
	return "."
}
