package routes

import (
	"embed"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/frontend"
	"github.com/sirupsen/logrus"
)

type frontendFS struct {
	http.FileSystem
}

func (fs frontendFS) Exists(prefix string, path string) bool {
	_, err := fs.Open(path)
	if err != nil {
		return false
	}
	return true
}

func NewFrontendFS(fs embed.FS) static.ServeFileSystem {
	return frontendFS{
		FileSystem: http.FS(fs),
	}
}

type FrontendRoutes struct {
	logger *logrus.Logger
	maxAge time.Duration
}

func (r *FrontendRoutes) Setup(e *gin.Engine) {
	e.Use(gzip.Gzip(gzip.DefaultCompression))
	e.Use(static.Serve("/", NewFrontendFS(frontend.Assets)))
}

func NewFrontendRoutes(logger *logrus.Logger, cfg config.HttpConfig) *FrontendRoutes {
	return &FrontendRoutes{
		logger: logger,
		maxAge: cfg.Routes.Frontend.MaxAge,
	}
}
