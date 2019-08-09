package cmd

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"os"
	fp "path/filepath"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	apppaths "github.com/muesli/go-app-paths"
	"github.com/spf13/cobra"
)

var (
	db              database.DB
	dataDir         string
	httpClient      *http.Client
	developmentMode bool
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

	rootCmd.PersistentPreRun = preRunRootHandler
	rootCmd.PersistentFlags().Bool("portable", false, "run shiori in portable mode")
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

func preRunRootHandler(cmd *cobra.Command, args []string) {
	// Read flag
	var err error
	portableMode, _ := cmd.Flags().GetBool("portable")

	// Get and create data dir
	dataDir, err = getDataDir(portableMode)
	if err != nil {
		cError.Printf("Failed to get data dir: %v\n", err)
		os.Exit(1)
	}

	err = os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		cError.Printf("Failed to create data dir: %v\n", err)
		os.Exit(1)
	}

	// Open database
	dbPath := fp.Join(dataDir, "shiori.db")
	db, err = database.OpenSQLiteDatabase(dbPath)
	if err != nil {
		cError.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
}

func getDataDir(portableMode bool) (string, error) {
	// If in portable mode, uses directory of executable
	if portableMode {
		exePath, err := os.Executable()
		if err != nil {
			return "", err
		}

		exeDir := fp.Dir(exePath)
		return fp.Join(exeDir, "shiori-data"), nil
	}

	if developmentMode {
		return "dev-data", nil
	}

	// Try to look at environment variables
	dataDir, found := os.LookupEnv("SHIORI_DIR")
	if found {
		return dataDir, nil
	}

	// Try to use platform specific app path
	userScope := apppaths.NewScope(apppaths.User, "shiori", "shiori")
	dataDir, err := userScope.DataDir()
	if err == nil {
		return dataDir, nil
	}

	// When all fail, use current working directory
	return ".", nil
}
