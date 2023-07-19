package cmd

import (
	"context"
	"strings"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/http"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newServerCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Starts the Shiori webserver",
		Long:  "Serves the Shiori web interface and API.",
		Run:   newServerCommandHandler(logger),
	}

	cmd.Flags().IntP("port", "p", 8080, "Port used by the server")
	cmd.Flags().StringP("address", "a", "", "Address the server listens to")
	cmd.Flags().StringP("webroot", "r", "/", "Root path that used by server")
	cmd.Flags().Bool("access-log", true, "Print out a non-standard access log")
	cmd.Flags().Bool("serve-web-ui", true, "Serve static files from the webroot path")
	cmd.Flags().String("secret-key", "", "Secret key used for encrypting session data")

	return cmd
}

func newServerCommandHandler(logger *logrus.Logger) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		logger.Warn("This server is still in alpha, use it at your own risk. For more information check https://github.com/go-shiori/shiori/issues/640")

		ctx := context.Background()

		cfg := config.ParseServerConfiguration(ctx, logger)

		database, err := openDatabase(ctx, cfg.Database.DBMS, cfg.Database.URL)
		if err != nil {
			logger.WithError(err).Fatal("error opening database")
		}

		if cfg.Development {
			logger.Warn("Development mode is ENABLED, this will enable some helpers for local development, unsuitable for production environments")
		}

		dependencies := config.NewDependencies(logger, database, cfg)
		dependencies.Domains.Auth = domains.NewAccountsDomain(logger, cfg.Http.SecretKey, database)
		dependencies.Domains.Archiver = domains.NewArchiverDomain(logger, cfg.Storage.DataDir)

		// Get flags value
		port, _ := cmd.Flags().GetInt("port")
		address, _ := cmd.Flags().GetString("address")
		rootPath, _ := cmd.Flags().GetString("webroot")
		accessLog, _ := cmd.Flags().GetBool("access-log")
		serveWebUI, _ := cmd.Flags().GetBool("serve-web-ui")
		secretKey, _ := cmd.Flags().GetString("secret-key")
		portableMode, _ := cmd.Flags().GetBool("portable")

		// Validate root path
		if rootPath == "" {
			rootPath = "/"
		}

		if !strings.HasPrefix(rootPath, "/") {
			rootPath = "/" + rootPath
		}

		if !strings.HasSuffix(rootPath, "/") {
			rootPath += "/"
		}

		// Override configuration from flags
		cfg.Http.Port = port
		cfg.Http.Address = address + ":"
		cfg.Http.RootPath = rootPath
		cfg.Http.AccessLog = accessLog
		cfg.Http.ServeWebUI = serveWebUI
		cfg.Http.SecretKey = secretKey

		// Check configuration
		// For now it will just log to the console, but in the future it will be fatal. The only required
		// setting for now is the secret key.
		if errs, isValid := cfg.IsValid(); !isValid {
			logger.Error("Found some errors in configuration.For now server will start but this will be fatal in the future.")
			for _, err := range errs {
				logger.WithError(err).Error("found invalid configuration")
			}
		}

		if cfg.Storage.DataDir == "" {
			cfg.Storage.DataDir, err = getDataDir(portableMode)
			if err != nil {
				logger.WithError(err).Warn("error getting data directory, using default.")
			}
		}

		server := http.NewHttpServer(logger).Setup(cfg, dependencies)

		if err := server.Start(ctx); err != nil {
			logger.WithError(err).Fatal("error starting server")
		}
		logger.WithField("addr", address).Debug("started http server")

		server.WaitStop(ctx)
	}
}
