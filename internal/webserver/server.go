package webserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/pkg/warc"
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
		ArchiveCache: cch.New(time.Minute, 5*time.Minute),
	}

	hdl.ArchiveCache.OnEvicted(func(key string, data interface{}) {
		archive := data.(*warc.Archive)
		archive.Close()
	})

	// Create router
	router := httprouter.New()

	router.GET("/js/*filepath", hdl.serveJsFile)
	router.GET("/res/*filepath", hdl.serveFile)
	router.GET("/css/*filepath", hdl.serveFile)
	router.GET("/fonts/*filepath", hdl.serveFile)

	router.GET("/", hdl.serveIndexPage)
	router.GET("/login", hdl.serveLoginPage)
	router.GET("/bookmark/:id/thumb", hdl.serveThumbnailImage)
	router.GET("/bookmark/:id/content", hdl.serveBookmarkContent)
	router.GET("/bookmark/:id/archive/*filepath", hdl.serveBookmarkArchive)

	router.POST("/api/login", hdl.apiLogin)
	router.POST("/api/logout", hdl.apiLogout)
	router.GET("/api/bookmarks", hdl.apiGetBookmarks)
	router.GET("/api/tags", hdl.apiGetTags)
	router.PUT("/api/tag", hdl.apiRenameTag)
	router.POST("/api/bookmarks", hdl.apiInsertBookmark)
	router.DELETE("/api/bookmarks", hdl.apiDeleteBookmark)
	router.PUT("/api/bookmarks", hdl.apiUpdateBookmark)
	router.PUT("/api/cache", hdl.apiUpdateCache)
	router.PUT("/api/bookmarks/tags", hdl.apiUpdateBookmarkTags)

	router.GET("/api/accounts", hdl.apiGetAccounts)
	router.PUT("/api/accounts", hdl.apiUpdateAccount)
	router.POST("/api/accounts", hdl.apiInsertAccount)
	router.DELETE("/api/accounts", hdl.apiDeleteAccount)

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
