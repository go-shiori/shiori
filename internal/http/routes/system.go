package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type SystemRoutes struct {
	logger *logrus.Logger
}

func (r *SystemRoutes) Setup(group *gin.RouterGroup) model.Routes {
	group.GET("/liveness", r.livenessHandler)
	return r
}

func (r *SystemRoutes) livenessHandler(c *gin.Context) {
	response.Send(c, 200, struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Date    string `json:"date"`
	}{
		Version: model.BuildVersion,
		Commit:  model.BuildCommit,
		Date:    model.BuildDate,
	})
}

func NewSystemRoutes(logger *logrus.Logger) *SystemRoutes {
	return &SystemRoutes{
		logger: logger,
	}
}
