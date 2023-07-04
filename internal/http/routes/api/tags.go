package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type TagsAPIRoutes struct {
	logger *logrus.Logger
	deps   *config.Dependencies
}

func (r *TagsAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.GET("/", r.listHandler)
	return r
}

func (r *TagsAPIRoutes) listHandler(c *gin.Context) {
	tags, err := r.deps.Database.GetTags(c)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, 200, tags)
}

func NewTagsPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *TagsAPIRoutes {
	return &TagsAPIRoutes{
		logger: logger,
		deps:   deps,
	}
}
