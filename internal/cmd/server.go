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
		Short: "Run the Shiori webserver [alpha]",
		Long:  "Runs the new Shiori webserver with new API definitions. [alpha]",
		Run:   newServerCommandHandler(logger),
	}

	cmd.Flags().IntP("port", "p", 8080, "Port used by the server")
	cmd.Flags().StringP("address", "a", "", "Address the server listens to")
	cmd.Flags().StringP("webroot", "r", "/", "Root path that used by server")
	cmd.Flags().Bool("log", true, "Print out a non-standard access log")

	return cmd
}

func newServerCommandHandler(logger *logrus.Logger) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		database, err := openDatabase(ctx)
		if err != nil {
			logger.WithError(err).Fatal("error opening database")
		}

		cfg := config.ParseServerConfiguration(ctx, logger)

		if cfg.Development {
			logger.Warn("Development mode is ENABLED, this will enable some helpers for local development, unsuitable for production environments")
		}

		dependencies := config.NewDependencies(logger, database, cfg)
		dependencies.Domains.Auth = domains.NewAccountsDomain(logger, cfg.Http.SecretKey, database)
		dependencies.Domains.Archiver = domains.NewArchiverDomain(logger, cfg.Http.Storage.DataDir)

		// Get flags value
		port, _ := cmd.Flags().GetInt("port")
		address, _ := cmd.Flags().GetString("address")
		rootPath, _ := cmd.Flags().GetString("webroot")
		accessLog, _ := cmd.Flags().GetBool("log")

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

		server := http.NewHttpServer(logger, cfg.Http, dependencies).Setup(cfg.Http, dependencies)

		if err := server.Start(ctx); err != nil {
			logger.WithError(err).Fatal("error starting server")
		}
		logger.WithField("addr", address).Debug("started http server")

		server.WaitStop(ctx)
	}
}
