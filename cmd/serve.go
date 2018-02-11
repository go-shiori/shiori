package cmd

import (
	"encoding/json"
	"fmt"
	db "github.com/RadhiFadlillah/shiori/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	fp "path/filepath"
	"strconv"
	"strings"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve web app for managing bookmarks.",
		Long: "Run a simple annd performant web server which serves the site for managing bookmarks." +
			"If --port flag is not used, it will use port 8080 by default.",
		Run: func(cmd *cobra.Command, args []string) {
			router := httprouter.New()

			router.GET("/", serveFiles)
			router.GET("/js/*filepath", serveFiles)
			router.GET("/css/*filepath", serveFiles)
			router.GET("/webfonts/*filepath", serveFiles)
			router.GET("/api/bookmarks", apiGetBookmarks)

			url := fmt.Sprintf(":%d", 8080)
			logrus.Infoln("Serve shiori in", url)
			logrus.Fatalln(http.ListenAndServe(url, router))
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serveFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := r.URL.Path
	filepath = strings.TrimPrefix(filepath, "/")
	filepath = fp.Join("view", filepath)
	fmt.Println(filepath)
	http.ServeFile(w, r, filepath)
}

func apiGetBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	queries := r.URL.Query()
	strLimit := queries.Get("limit")
	strOffset := queries.Get("offset")

	limit, _ := strconv.Atoi(strLimit)
	if limit <= 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(strOffset)
	if offset <= 0 {
		offset = 0
	}

	bookmarks, err := DB.GetBookmarks(db.GetBookmarksOptions{
		Limit:       limit,
		Offset:      offset,
		OrderLatest: true})
	checkError(err)

	err = json.NewEncoder(w).Encode(&bookmarks)
	checkError(err)
}
