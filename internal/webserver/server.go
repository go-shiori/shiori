package webserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/julienschmidt/httprouter"
	cch "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

var httpClient = &http.Client{Timeout: time.Minute}

// ServeApp serves wb interface in specified port
func ServeApp(DB database.DB, dataDir string, port int) error {
	// Create handler
	hdl := handler{
		DB:           DB,
		DataDir:      dataDir,
		UserCache:    cch.New(time.Hour, 10*time.Minute),
		SessionCache: cch.New(time.Hour, 10*time.Minute),
	}

	// Create router
	router := httprouter.New()

	router.GET("/js/*filepath", hdl.serveJsFile)
	router.GET("/res/*filepath", hdl.serveFile)
	router.GET("/css/*filepath", hdl.serveFile)
	router.GET("/fonts/*filepath", hdl.serveFile)

	router.GET("/", hdl.serveIndexPage)
	router.GET("/login", hdl.serveLoginPage)
	router.GET("/thumb/:id", hdl.serveThumbnailImage)

	router.POST("/api/login", hdl.apiLogin)
	router.POST("/api/logout", hdl.apiLogout)
	router.GET("/api/bookmarks", hdl.apiGetBookmarks)
	router.GET("/api/tags", hdl.apiGetTags)
	router.POST("/api/bookmarks", hdl.apiInsertBookmark)
	router.DELETE("/api/bookmarks", hdl.apiDeleteBookmark)
	router.PUT("/api/bookmarks", hdl.apiUpdateBookmark)
	router.PUT("/api/archive", hdl.apiUpdateArchive)
	// router.PUT("/api/bookmarks/tags", hdl.apiUpdateBookmarkTags)

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
	return svr.ListenAndServe()
}
