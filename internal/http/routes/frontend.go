package routes

import (
	"embed"
	"net/http"
	"path"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/model"
	views "github.com/go-shiori/shiori/internal/view"
	"github.com/sirupsen/logrus"
)

type assetsFS struct {
	http.FileSystem
	logger *logrus.Logger
}

func (fs assetsFS) Exists(prefix string, path string) bool {
	_, err := fs.Open(path)
	if err != nil {
		logrus.WithError(err).WithField("path", path).WithField("prefix", prefix).Error("requested frontend file not found")
	}
	return err == nil
}

func (fs assetsFS) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(path.Join("assets", name))
	if err != nil {
		logrus.WithError(err).WithField("path", name).Error("requested frontend file not found")
	}
	return f, err
}

func newAssetsFS(logger *logrus.Logger, fs embed.FS) static.ServeFileSystem {
	return assetsFS{
		logger:     logger,
		FileSystem: http.FS(fs),
	}
}

type FrontendRoutes struct {
	logger *logrus.Logger
	cfg    *config.Config
}

func (r *FrontendRoutes) Setup(e *gin.Engine) {
	group := e.Group("/")
	group.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"RootPath": r.cfg.Http.RootPath,
			"Version":  model.BuildVersion,
		})
	})
	e.StaticFS("/assets", newAssetsFS(r.logger, views.Assets))
}

func NewFrontendRoutes(logger *logrus.Logger, cfg *config.Config) *FrontendRoutes {
	return &FrontendRoutes{
		logger: logger,
		cfg:    cfg,
	}
}
