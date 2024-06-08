package api_v1

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type SystemAPIRoutes struct {
	logger *logrus.Logger
	deps   *dependencies.Dependencies
}

func (r *SystemAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.Use(middleware.AuthenticationRequired())
	g.GET("/info", r.infoHandler)
	return r
}

type infoResponse struct {
	Version struct {
		Tag    string `json:"tag"`
		Commit string `json:"commit"`
		Date   string `json:"date"`
	} `json:"version"`
	Database string `json:"database"`
	OS       string `json:"os"`
}

// System info API endpoint godoc
//
//	@Summary					Get general system information
//	@Description				Get general system information like Shiori version, database, and OS
//	@Tags						system
//	@Produce					json
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Success					200	{object}	infoResponse
//	@Failure					403	{object}	nil	"Only owners can access this endpoint"
//	@Router						/api/v1/system/info [get]
func (r *SystemAPIRoutes) infoHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if !ctx.GetAccount().Owner {
		response.SendError(c, http.StatusForbidden, "Only owners can access this endpoint")
		return
	}

	response.Send(c, 200, infoResponse{
		Version: struct {
			Tag    string `json:"tag"`
			Commit string `json:"commit"`
			Date   string `json:"date"`
		}{
			Tag:    model.BuildVersion,
			Commit: model.BuildCommit,
			Date:   model.BuildDate,
		},
		Database: r.deps.Database.DBx().DriverName(),
		OS:       runtime.GOOS + " (" + runtime.GOARCH + ")",
	})
}

func NewSystemAPIRoutes(logger *logrus.Logger, deps *dependencies.Dependencies) *SystemAPIRoutes {
	return &SystemAPIRoutes{
		logger: logger,
		deps:   deps,
	}
}
