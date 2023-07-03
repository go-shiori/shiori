package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type DebugAPIRoutes struct {
	logger *logrus.Logger
	deps   *config.Dependencies
}

func (r *DebugAPIRoutes) Setup(group *gin.RouterGroup) model.Routes {
	group.GET("/create_user", r.createUserHandler)
	return r
}

func (r *DebugAPIRoutes) createUserHandler(c *gin.Context) {
	account := model.Account{
		Username: "shiori",
		Password: "gopher",
		Owner:    true,
	}

	if err := r.deps.Database.SaveAccount(c, account); err != nil {
		response.SendError(c, 500, err.Error())
		return
	}

	response.Send(c, 201, account)
}

func NewDebugPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *DebugAPIRoutes {
	return &DebugAPIRoutes{
		logger: logger,
		deps:   deps,
	}
}
