package cmd

import (
	"fmt"
	"os"
	fp "path/filepath"
	"time"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// ShioriCmd returns the root command for shiori
func ShioriCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "shiori",
		Short: "Simple command-line bookmark manager built with Go",
	}

	rootCmd.PersistentFlags().Bool("portable", false, "run shiori in portable mode")
	rootCmd.PersistentFlags().String("storage-directory", "", "path to store shiori data")
	rootCmd.MarkFlagsMutuallyExclusive("portable", "storage-directory")

	rootCmd.PersistentFlags().String("log-level", logrus.InfoLevel.String(), "set logrus loglevel")
	rootCmd.PersistentFlags().Bool("log-caller", false, "logrus report caller or not")

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
		newVersionCommand(),
		newServerCommand(),
	)

	return rootCmd
}

func initShiori(ctx context.Context, cmd *cobra.Command) (*config.Config, *config.Dependencies) {
	logger := logrus.New()

	portableMode, _ := cmd.Flags().GetBool("portable")
	logLevel, _ := cmd.Flags().GetString("log-level")
	logCaller, _ := cmd.Flags().GetBool("log-caller")
	storageDirectory, _ := cmd.Flags().GetString("storage-directory")

	logger.SetReportCaller(logCaller)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		TimestampFormat:  time.RFC3339,
		CallerPrettyfier: SFCallerPrettyfier,
	})

	if lvl, err := logrus.ParseLevel(logLevel); err != nil {
		logger.WithError(err).Panic("failed to set log level")
	} else {
		logger.SetLevel(lvl)
	}

	cfg := config.ParseServerConfiguration(ctx, logger)

	if storageDirectory != "" && cfg.Storage.DataDir != "" {
		logger.Warn("--storage-directory is set, overriding SHIORI_DIR.")
		cfg.Storage.DataDir = storageDirectory
	}

	cfg.SetDefaults(logger, portableMode)

	err := os.MkdirAll(cfg.Storage.DataDir, model.DataDirPerm)
	if err != nil {
		logger.WithError(err).Fatal("error creating data directory")
	}

	db, err := openDatabase(logger, ctx, cfg)
	if err != nil {
		logger.WithError(err).Fatal("error opening database")
	}

	// Migrate
	if err := db.Migrate(); err != nil {
		logger.WithError(err).Fatalf("Error running migration")
	}

	if cfg.Development {
		logger.Warn("Development mode is ENABLED, this will enable some helpers for local development, unsuitable for production environments")
	}

	dependencies := config.NewDependencies(logger, db, cfg)
	dependencies.Domains.Auth = domains.NewAccountsDomain(logger, cfg.Http.SecretKey, db)
	dependencies.Domains.Archiver = domains.NewArchiverDomain(logger, cfg.Storage.DataDir)

	// Workaround: Get accounts to make sure at least one is present in the database.
	// If there's no accounts in the database, create the shiori/gopher account the legacy api
	// hardcoded in the login handler.
	accounts, err := db.GetAccounts(cmd.Context(), database.GetAccountsOptions{})
	if err != nil {
		cError.Printf("Failed to get owner account: %v\n", err)
		os.Exit(1)
	}

	if len(accounts) == 0 {
		account := model.Account{
			Username: "shiori",
			Password: "gopher",
			Owner:    true,
		}

		if err := db.SaveAccount(cmd.Context(), account); err != nil {
			logger.WithError(err).Fatal("error ensuring owner account")
		}
	}

	return cfg, dependencies
}

func openDatabase(logger *logrus.Logger, ctx context.Context, cfg *config.Config) (database.DB, error) {
	if cfg.Database.URL != "" {
		return database.Connect(ctx, cfg.Database.URL)
	}

	if cfg.Database.DBMS != "" {
		logger.Warnf("The use of SHIORI_DBMS is deprecated and will be removed in the future. Please migrate to SHIORI_DATABASE_URL instead.")
	}

	// TODO remove this the moment DBMS is deprecated
	if cfg.Database.DBMS == "mysql" {
		return openMySQLDatabase(ctx)
	}
	if cfg.Database.DBMS == "postgresql" {
		return openPostgreSQLDatabase(ctx)
	}

	return database.OpenSQLiteDatabase(ctx, fp.Join(cfg.Storage.DataDir, "shiori.db"))
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
	sslmode, _ := os.LookupEnv("SHIORI_PG_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)
	return database.OpenPGDatabase(ctx, connString)
}
