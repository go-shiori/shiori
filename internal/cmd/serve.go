package cmd

import (
	"github.com/go-shiori/shiori/internal/webserver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve web interface for managing bookmarks",
		Long: "Run a simple and performant web server which " +
			"serves the site for managing bookmarks. If --port " +
			"flag is not used, it will use port 8080 by default.",
		Run: serveHandler,
	}

	cmd.Flags().IntP("port", "p", 8080, "Port used by the server")
	cmd.Flags().StringP("address", "a", "", "Address the server listens to")

	return cmd
}

func serveHandler(cmd *cobra.Command, args []string) {
	port, _ := cmd.Flags().GetInt("port")
	address, _ := cmd.Flags().GetString("address")

	err := webserver.ServeApp(db, dataDir, address, port)
	if err != nil {
		logrus.Fatalf("Server error: %v\n", err)
	}
}
