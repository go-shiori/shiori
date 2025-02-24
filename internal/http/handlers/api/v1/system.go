package api_v1

import (
	"net/http"
	"runtime"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

type infoResponse struct {
	Version struct {
		Tag    string `json:"tag"`
		Commit string `json:"commit"`
		Date   string `json:"date"`
	} `json:"version"`
	Database string `json:"database"`
	OS       string `json:"os"`
}

// HandleSystemInfo serves the system info API endpoint
func HandleSystemInfo(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		return
	}

	response.Send(c, http.StatusOK, infoResponse{
		Version: struct {
			Tag    string `json:"tag"`
			Commit string `json:"commit"`
			Date   string `json:"date"`
		}{
			Tag:    model.BuildVersion,
			Commit: model.BuildCommit,
			Date:   model.BuildDate,
		},
		Database: deps.Database().ReaderDB().DriverName(),
		OS:       runtime.GOOS + " (" + runtime.GOARCH + ")",
	})
}
