package webserver

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/julienschmidt/httprouter"
	cch "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Config is parameter that used for starting web server
type Config struct {
	DB            database.DB
	DataDir       string
	ServerAddress string
	ServerPort    int
	RootPath      string
}

// ServeApp serves wb interface in specified port
func ServeApp(cfg Config) error {
	// Create handler
	hdl := handler{
		DB:           cfg.DB,
		DataDir:      cfg.DataDir,
		UserCache:    cch.New(time.Hour, 10*time.Minute),
		SessionCache: cch.New(time.Hour, 10*time.Minute),
		ArchiveCache: cch.New(time.Minute, 5*time.Minute),
		RootPath:     cfg.RootPath,
	}

	hdl.prepareSessionCache()
	hdl.prepareArchiveCache()

	err := hdl.prepareTemplates()
	if err != nil {
		return fmt.Errorf("failed to prepare templates: %v", err)
	}

	// Create router
	router := httprouter.New()

	// jp here means "join path", as in "join route with root path"
	jp := func(route string) string {
		return path.Join(cfg.RootPath, route)
	}

	router.GET(jp("/js/*filepath"), hdl.serveJsFile)
	router.GET(jp("/res/*filepath"), hdl.serveFile)
	router.GET(jp("/css/*filepath"), hdl.serveFile)
	router.GET(jp("/fonts/*filepath"), hdl.serveFile)

	router.GET(jp("/"), hdl.serveIndexPage)
	router.GET(jp("/login"), hdl.serveLoginPage)
	router.GET(jp("/bookmark/:id/thumb"), hdl.serveThumbnailImage)
	router.GET(jp("/bookmark/:id/content"), hdl.serveBookmarkContent)
	router.GET(jp("/bookmark/:id/archive/*filepath"), hdl.serveBookmarkArchive)

	router.POST(jp("/api/login"), hdl.apiLogin)
	router.POST(jp("/api/logout"), hdl.apiLogout)
	router.GET(jp("/api/bookmarks"), hdl.apiGetBookmarks)
	router.GET(jp("/api/tags"), hdl.apiGetTags)
	router.PUT(jp("/api/tag"), hdl.apiRenameTag)
	router.POST(jp("/api/bookmarks"), hdl.apiInsertBookmark)
	router.DELETE(jp("/api/bookmarks"), hdl.apiDeleteBookmark)
	router.PUT(jp("/api/bookmarks"), hdl.apiUpdateBookmark)
	router.PUT(jp("/api/cache"), hdl.apiUpdateCache)
	router.PUT(jp("/api/bookmarks/tags"), hdl.apiUpdateBookmarkTags)
	router.POST(jp("/api/bookmarks/ext"), hdl.apiInsertViaExtension)
	router.DELETE(jp("/api/bookmarks/ext"), hdl.apiDeleteViaExtension)
	router.GET(jp("/api/bookmarks/ext"), hdl.apiGetViaExtension)

	router.GET(jp("/api/accounts"), hdl.apiGetAccounts)
	router.PUT(jp("/api/accounts"), hdl.apiUpdateAccount)
	router.POST(jp("/api/accounts"), hdl.apiInsertAccount)
	router.DELETE(jp("/api/accounts"), hdl.apiDeleteAccount)

	// Route for panic
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
		http.Error(w, fmt.Sprint(arg), 500)
	}

	// Create server
	url := fmt.Sprintf("%s:%d", cfg.ServerAddress, cfg.ServerPort)
	svr := &http.Server{
		Addr:         url,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: time.Minute,
	}

	// Serve app
	logrus.Infoln("Serve shiori in", url)
	return svr.ListenAndServe()
}
