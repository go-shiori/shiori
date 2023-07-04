package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type APIRoutes struct {
	logger *logrus.Logger
	deps   *config.Dependencies
}

func (r *APIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	if r.deps.Config.Development {
		r.handle(g, "/debug", NewDebugPIRoutes(r.logger, r.deps))
	}

	// Account API handles authentication in each route
	r.handle(g, "/auth", NewAuthAPIRoutes(r.logger, r.deps))

	// From here on, all routes require authentication
	g.Use(middleware.AuthenticationRequired())
	r.handle(g, "/bookmarks", NewBookmarksPIRoutes(r.logger, r.deps))
	r.handle(g, "/tags", NewTagsPIRoutes(r.logger, r.deps))

	return r
}

func (s *APIRoutes) handle(g *gin.RouterGroup, path string, routes model.Routes) {
	group := g.Group(path)
	routes.Setup(group)
}

func NewAPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *APIRoutes {
	return &APIRoutes{
		logger: logger,
		deps:   deps,
	}
}
