package api_v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type TagsAPIRoutes struct {
	logger *logrus.Logger
	deps   *dependencies.Dependencies
}

func (r *TagsAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.Use(middleware.AuthenticationRequired())
	g.GET("/", r.listHandler)
	g.POST("/", r.createHandler)
	return r
}

// @Summary					List tags
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Success					200	{object}	model.Tag	"List of tags"
// @Failure					403	{object}	nil			"Token not provided/invalid"
// @Router						/api/v1/tags [get]
func (r *TagsAPIRoutes) listHandler(c *gin.Context) {
	tags, err := r.deps.Database.GetTags(c)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusOK, tags)
}

// @Summary					Create tag
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Success					200	{object}	model.Tag	"Created tag"
// @Failure					400	{object}	nil			"Token not provided/invalid"
// @Failure					403	{object}	nil			"Token not provided/invalid"
// @Router						/api/v1/tags [post]
func (r *TagsAPIRoutes) createHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if !ctx.GetAccount().Owner {
		response.SendError(c, http.StatusForbidden, nil)
		return
	}

	var tag model.Tag
	if err := c.BindJSON(&tag); err != nil {
		response.SendError(c, http.StatusBadRequest, nil)
		return
	}

	err := r.deps.Database.CreateTags(c, tag)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusCreated, nil)
}

func NewTagsPIRoutes(logger *logrus.Logger, deps *dependencies.Dependencies) *TagsAPIRoutes {
	return &TagsAPIRoutes{
		logger: logger,
		deps:   deps,
	}
}
