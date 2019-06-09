package cmd

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/spf13/cobra"
)

var (
	// DB is database that used by cmd
	DB database.DB

	// DataDir is directory for downloaded data
	DataDir string

	httpClient *http.Client
)

func init() {
	jar, _ := cookiejar.New(nil)
	httpClient = &http.Client{
		Timeout: time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Jar: jar,
	}
}

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
	)

	return rootCmd
}
