package serve

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	dt "github.com/RadhiFadlillah/shiori/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewServeCmd creates new command for serving web page
func NewServeCmd(db dt.Database, dataDir string) *cobra.Command {
	// Create handler
	hdl, err := newWebHandler(db, dataDir)
	checkError(err)

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve web app for managing bookmarks",
		Long: "Run a simple annd performant web server which serves the site for managing bookmarks." +
			"If --port flag is not used, it will use port 8080 by default.",
		Run: func(cmd *cobra.Command, args []string) {
			// Parse flags
			port, _ := cmd.Flags().GetInt("port")

			// Create router
			router := httprouter.New()

			router.GET("/js/*filepath", hdl.serveFiles)
			router.GET("/res/*filepath", hdl.serveFiles)
			router.GET("/css/*filepath", hdl.serveFiles)
			router.GET("/webfonts/*filepath", hdl.serveFiles)

			router.GET("/", hdl.serveIndexPage)
			router.GET("/login", hdl.serveLoginPage)
			router.GET("/bookmark/:id", hdl.serveBookmarkCache)
			router.GET("/thumb/:id", hdl.serveThumbnailImage)

			router.POST("/api/login", hdl.apiLogin)
			router.GET("/api/bookmarks", hdl.apiGetBookmarks)
			router.GET("/api/tags", hdl.apiGetTags)
			router.POST("/api/bookmarks", hdl.apiInsertBookmark)
			router.PUT("/api/cache", hdl.apiUpdateCache)
			router.PUT("/api/bookmarks", hdl.apiUpdateBookmark)
			router.PUT("/api/bookmarks/tags", hdl.apiUpdateBookmarkTags)
			router.DELETE("/api/bookmarks", hdl.apiDeleteBookmark)

			// Route for panic
			router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
				http.Error(w, fmt.Sprint(arg), 500)
			}

			// Create server
			url := fmt.Sprintf(":%d", port)
			svr := &http.Server{
				Addr:         url,
				Handler:      router,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 20 * time.Second,
			}

			// Serve app
			logrus.Infoln("Serve shiori in", url)
			logrus.Fatalln(svr.ListenAndServe())
		},
	}

	// Set flags for root command
	rootCmd.Flags().IntP("port", "p", 8080, "Port that used by server")

	return rootCmd
}

func checkError(err error) {
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
}
