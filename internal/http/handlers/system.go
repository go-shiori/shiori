package handlers

import (
	"net/http"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

// HandleLiveness handles the liveness check endpoint
func HandleLiveness(deps model.Dependencies, c model.WebContext) {
	response.SendJSON(c, http.StatusOK, struct {
		Version string
		Commit  string
		Date    string
	}{
		Version: model.BuildVersion,
		Commit:  model.BuildCommit,
		Date:    model.BuildDate,
	})
}
