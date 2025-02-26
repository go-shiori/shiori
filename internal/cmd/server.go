package cmd

import (
	"context"
	"strings"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Starts the Shiori webserver",
		Long:  "Serves the Shiori web interface and API.",
		Run:   newServerCommandHandler(),
	}

	cmd.Flags().IntP("port", "p", 8080, "Port used by the server")
	cmd.Flags().StringP("address", "a", "", "Address the server listens to")
	cmd.Flags().StringP("webroot", "r", "/", "Root path that used by server")
	cmd.Flags().Bool("access-log", false, "Print out a non-standard access log")
	cmd.Flags().Bool("serve-web-ui", true, "Serve static files from the webroot path")
	cmd.Flags().String("secret-key", "", "Secret key used for encrypting session data")

	return cmd
}

func setIfFlagChanged(flagName string, flags *pflag.FlagSet, cfg *config.Config, fn func(cfg *config.Config)) {
	if flags.Changed(flagName) {
		fn(cfg)
	}
}

func newServerCommandHandler() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get flags values
		port, _ := cmd.Flags().GetInt("port")
		address, _ := cmd.Flags().GetString("address")
		rootPath, _ := cmd.Flags().GetString("webroot")
		accessLog, _ := cmd.Flags().GetBool("access-log")
		serveWebUI, _ := cmd.Flags().GetBool("serve-web-ui")
		secretKey, _ := cmd.Flags().GetBytesHex("secret-key")

		cfg, dependencies := initShiori(ctx, cmd)

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

		// Override configuration from flags if needed
		setIfFlagChanged("port", cmd.Flags(), cfg, func(cfg *config.Config) {
			cfg.Http.Port = port
		})
		setIfFlagChanged("address", cmd.Flags(), cfg, func(cfg *config.Config) {
			cfg.Http.Address = address + ":"
		})
		setIfFlagChanged("webroot", cmd.Flags(), cfg, func(cfg *config.Config) {
			cfg.Http.RootPath = rootPath
		})
		setIfFlagChanged("access-log", cmd.Flags(), cfg, func(cfg *config.Config) {
			cfg.Http.AccessLog = accessLog
		})
		setIfFlagChanged("serve-web-ui", cmd.Flags(), cfg, func(cfg *config.Config) {
			cfg.Http.ServeWebUI = serveWebUI
		})
		setIfFlagChanged("secret-key", cmd.Flags(), cfg, func(cfg *config.Config) {
			cfg.Http.SecretKey = secretKey
		})

		dependencies.Logger().Infof("Starting Shiori v%s", model.BuildVersion)

		server, err := http.NewHttpServer(dependencies.Logger()).Setup(cfg, dependencies)
		if err != nil {
			dependencies.Logger().WithError(err).Fatal("error setting up server")
		}

		if err := server.Start(ctx); err != nil {
			dependencies.Logger().WithError(err).Fatal("error starting server")
		}
		dependencies.Logger().Debug("started http server")

		server.WaitStop(ctx)
	}
}
