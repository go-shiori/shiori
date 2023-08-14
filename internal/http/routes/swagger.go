package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/go-shiori/shiori/docs/swagger"
)

type SwaggerAPIRoutes struct {
	logger *logrus.Logger
}

func (r *SwaggerAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/swagger" || c.Request.URL.Path == "/swagger/" {
			c.Redirect(302, "/swagger/index.html")
			return
		}
	})
	g.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}

func NewSwaggerAPIRoutes(logger *logrus.Logger) *SwaggerAPIRoutes {
	return &SwaggerAPIRoutes{
		logger: logger,
	}
}
