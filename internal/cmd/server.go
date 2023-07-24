package cmd

import (
	"context"
	"strings"

	"github.com/go-shiori/shiori/internal/http"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
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

func newServerCommandHandler() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get flags values
		port, _ := cmd.Flags().GetInt("port")
		address, _ := cmd.Flags().GetString("address")
		rootPath, _ := cmd.Flags().GetString("webroot")
		accessLog, _ := cmd.Flags().GetBool("access-log")
		serveWebUI, _ := cmd.Flags().GetBool("serve-web-ui")
		secretKey, _ := cmd.Flags().GetString("secret-key")

		cfg, dependencies := initShiori(ctx, cmd)

		cfg.Http.SetDefaults(dependencies.Log)

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

		dependencies.Log.Infof("Starting Shiori v%s", model.BuildVersion)

		server := http.NewHttpServer(dependencies.Log).Setup(cfg, dependencies)

		if err := server.Start(ctx); err != nil {
			dependencies.Log.WithError(err).Fatal("error starting server")
		}
		dependencies.Log.WithField("addr", address).Debug("started http server")

		server.WaitStop(ctx)
	}
}
