package cmd

import (
	"encoding/json"
	"fmt"
	db "github.com/RadhiFadlillah/shiori/database"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	fp "path/filepath"
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
			router.POST("/api/bookmarks", apiInsertBookmarks)
			router.PUT("/api/bookmarks", apiUpdateBookmarks)
			router.DELETE("/api/bookmarks", apiDeleteBookmarks)

			// Route for panic
			router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
				http.Error(w, fmt.Sprint(arg), 500)
			}

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
	bookmarks, err := DB.GetBookmarks(db.GetBookmarksOptions{OrderLatest: true})
	checkError(err)

	err = json.NewEncoder(w).Encode(&bookmarks)
	checkError(err)
}

func apiInsertBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Decode request
	request := model.Bookmark{}
	err := json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Save bookmark
	tags := make([]string, len(request.Tags))
	for i, tag := range request.Tags {
		tags[i] = tag.Name
	}

	book, err := addBookmark(request.URL, request.Title, request.Excerpt, tags, false)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&book)
	checkError(err)
}

func apiUpdateBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Decode request
	request := model.Bookmark{}
	err := json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Convert tags and ID
	id := []string{fmt.Sprintf("%d", request.ID)}
	tags := make([]string, len(request.Tags))
	for i, tag := range request.Tags {
		tags[i] = tag.Name
	}

	// Update bookmark
	bookmarks, err := updateBookmarks(id, request.URL, request.Title, request.Excerpt, tags, false)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&bookmarks[0])
	checkError(err)
}

func apiDeleteBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Decode request
	request := []string{}
	err := json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Delete bookmarks
	_, _, err = DB.DeleteBookmarks(request...)
	checkError(err)

	fmt.Fprint(w, request)
}
