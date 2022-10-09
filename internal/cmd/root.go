package cmd

import (
	"fmt"
	"os"
	fp "path/filepath"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	apppaths "github.com/muesli/go-app-paths"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	db              database.DB
	dataDir         string
	developmentMode bool
)

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
		checkCmd(),
		migrateCmd(),
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

	err = os.MkdirAll(dataDir, model.DataDirPerm)
	if err != nil {
		cError.Printf("Failed to create data dir: %v\n", err)
		os.Exit(1)
	}

	// Open database
	db, err = openDatabase(cmd.Context())
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
	userScope := apppaths.NewScope(apppaths.User, "shiori")
	dataDir, err := userScope.DataPath("")
	if err == nil {
		return dataDir, nil
	}

	// When all fail, use current working directory
	return ".", nil
}

func openDatabase(ctx context.Context) (database.DB, error) {
	switch dbms, _ := os.LookupEnv("SHIORI_DBMS"); dbms {
	case "mysql":
		return openMySQLDatabase(ctx)
	case "postgresql":
		return openPostgreSQLDatabase(ctx)
	default:
		return openSQLiteDatabase(ctx)
	}
}

func openSQLiteDatabase(ctx context.Context) (database.DB, error) {
	dbPath := fp.Join(dataDir, "shiori.db")
	return database.OpenSQLiteDatabase(ctx, dbPath)
}

func openMySQLDatabase(ctx context.Context) (database.DB, error) {
	user, _ := os.LookupEnv("SHIORI_MYSQL_USER")
	password, _ := os.LookupEnv("SHIORI_MYSQL_PASS")
	dbName, _ := os.LookupEnv("SHIORI_MYSQL_NAME")
	dbAddress, _ := os.LookupEnv("SHIORI_MYSQL_ADDRESS")

	connString := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4", user, password, dbAddress, dbName)
	return database.OpenMySQLDatabase(ctx, connString)
}

func openPostgreSQLDatabase(ctx context.Context) (database.DB, error) {
	host, _ := os.LookupEnv("SHIORI_PG_HOST")
	port, _ := os.LookupEnv("SHIORI_PG_PORT")
	user, _ := os.LookupEnv("SHIORI_PG_USER")
	password, _ := os.LookupEnv("SHIORI_PG_PASS")
	dbName, _ := os.LookupEnv("SHIORI_PG_NAME")

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	return database.OpenPGDatabase(ctx, connString)
}
