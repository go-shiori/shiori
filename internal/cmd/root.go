package cmd

import (
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/spf13/cobra"
)

var (
	// DB is database that used by cmd
	DB database.DB

	// DataDir is directory for downloaded data
	DataDir string

	httpClient = &http.Client{Timeout: time.Minute}
)

// ShioriCmd returns the root command for shiori
func ShioriCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "shiori",
		Short: "Simple command-line bookmark manager built with Go",
	}

	rootCmd.AddCommand(
		addCmd(),
		printCmd(),
		updateCmd(),
		deleteCmd(),
		openCmd(),
		importCmd(),
		exportCmd(),
		pocketCmd(),
		serveCmd(),
		accountCmd(),
	)

	return rootCmd
}
