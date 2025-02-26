package handlers

import (
	"net/http"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

// HandleLiveness handles the liveness check endpoint
func HandleLiveness(deps model.Dependencies, c model.WebContext) {
	response.Send(c, http.StatusOK, struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Date    string `json:"date"`
	}{
		Version: model.BuildVersion,
		Commit:  model.BuildCommit,
		Date:    model.BuildDate,
	})
}
