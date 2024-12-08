package api_v1

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type APIRoutes struct {
	logger       *logrus.Logger
	deps         *dependencies.Dependencies
	loginHandler model.LegacyLoginHandler
}

func (r *APIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	// Account API handles authentication in each route
	r.handle(g, "/auth", NewAuthAPIRoutes(r.logger, r.deps, r.loginHandler))
	r.handle(g, "/bookmarks", NewBookmarksAPIRoutes(r.logger, r.deps))
	r.handle(g, "/tags", NewTagsPIRoutes(r.logger, r.deps))
	r.handle(g, "/system", NewSystemAPIRoutes(r.logger, r.deps))

	return r
}

func (s *APIRoutes) handle(g *gin.RouterGroup, path string, routes model.Routes) {
	group := g.Group(path)
	routes.Setup(group)
}

func NewAPIRoutes(logger *logrus.Logger, deps *dependencies.Dependencies, loginHandler model.LegacyLoginHandler) *APIRoutes {
	return &APIRoutes{
		logger:       logger,
		deps:         deps,
		loginHandler: loginHandler,
	}
}
