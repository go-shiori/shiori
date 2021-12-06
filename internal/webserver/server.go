package webserver

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-shiori/shiori/internal/database"
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
	Log           bool
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
		Log:          cfg.Log,
	}

	hdl.prepareSessionCache()
	hdl.prepareArchiveCache()

	err := hdl.prepareTemplates()
	if err != nil {
		return fmt.Errorf("failed to prepare templates: %v", err)
	}

	// Create router
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)

	// jp here means "join path", as in "join route with root path"
	jp := func(route string) string {
		return path.Join(cfg.RootPath, route)
	}

	r.Group(func(r chi.Router) {
		r.Get(jp("/js/*"), hdl.serveJsFile)
		r.Get(jp("/res/*"), hdl.serveFile)
		r.Get(jp("/css/*"), hdl.serveFile)
		r.Get(jp("/fonts/*"), hdl.serveFile)
		r.Post("/api/login", hdl.apiLogin)
		r.Get(jp("/login"), hdl.serveLoginPage)
	})

	r.Group(func(r chi.Router) {
		r.Use(hdl.sessionValidateRedirect)

		r.Get(jp("/"), hdl.serveIndexPage)
		r.Get(jp("/bookmark/{id}/thumb"), hdl.serveThumbnailImage)
		r.Get(jp("/bookmark/{id}/content"), hdl.serveBookmarkContent)
		r.Get(jp("/bookmark/{id}/content/*"), hdl.serveBookmarkContent)
		r.Get(jp("/bookmark/{id}/archive"), hdl.redirectSlashAppend)
		r.Get(jp("/bookmark/{id}/archive/*"), hdl.serveBookmarkArchive)

		r.Route("/api", func(r chi.Router) {
			r.Get("/bookmarks", hdl.apiGetBookmarks)
			r.Get("/tags", hdl.apiGetTags)
			r.Put("/tag", hdl.apiRenameTag)
			r.Post("/bookmarks", hdl.apiInsertBookmark)
			r.Delete("/bookmarks", hdl.apiDeleteBookmark)
			r.Put("/bookmarks", hdl.apiUpdateBookmark)
			r.Put("/cache", hdl.apiUpdateCache)
			r.Put("/bookmarks/tags", hdl.apiUpdateBookmarkTags)
			r.Post("/bookmarks/ext", hdl.apiInsertViaExtension)
			r.Delete("/bookmarks/ext", hdl.apiDeleteViaExtension)

			r.Post("/logout", hdl.apiLogout)
			r.Get("/accounts", hdl.apiGetAccounts)
			r.Put("/accounts", hdl.apiUpdateAccount)
			r.Post("/accounts", hdl.apiInsertAccount)
			r.Delete("/accounts", hdl.apiDeleteAccount)
		})

	})

	// Create server
	url := fmt.Sprintf("%s:%d", cfg.ServerAddress, cfg.ServerPort)
	svr := &http.Server{
		Addr:         url,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: time.Minute,
	}

	// Serve app
	logrus.Infoln("Serve shiori in", url, cfg.RootPath)
	return svr.ListenAndServe()
}
