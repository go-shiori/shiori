package handlers

import (
	"embed"
	"net/http"
	"path"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	views "github.com/go-shiori/shiori/internal/view"
)

type assetsFS struct {
	http.FileSystem
}

func (fs assetsFS) Open(name string) (http.File, error) {
	return fs.FileSystem.Open(path.Join("assets", name))
}

func newAssetsFS(fs embed.FS) http.FileSystem {
	return assetsFS{
		FileSystem: http.FS(fs),
	}
}

// HandleFrontend serves the main frontend page
func HandleFrontend(deps model.Dependencies, c model.WebContext) {
	data := map[string]any{
		"RootPath": deps.Config().Http.RootPath,
		"Version":  model.BuildVersion,
	}

	if err := response.SendTemplate(c, "index.html", data); err != nil {
		deps.Logger().WithError(err).Error("failed to render template")
	}
}

// HandleAssets serves static assets
func HandleAssets(deps model.Dependencies, c model.WebContext) {
	fs := newAssetsFS(views.Assets)
	http.StripPrefix("/assets/", http.FileServer(fs)).ServeHTTP(c.ResponseWriter(), c.Request())
}
