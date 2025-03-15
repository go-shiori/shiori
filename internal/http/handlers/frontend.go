package handlers

import (
	"embed"
	"net/http"
	"path"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	views "github.com/go-shiori/shiori/internal/view"
	webapp "github.com/go-shiori/shiori/webapp"
)

type assetsFS struct {
	http.FileSystem
	serveWebUIV2 bool
}

func (fs assetsFS) Open(name string) (http.File, error) {
	pathJoin := "assets"
	if fs.serveWebUIV2 {
		pathJoin = "dist/assets"
	}

	return fs.FileSystem.Open(path.Join(pathJoin, name))
}

func newAssetsFS(fs embed.FS, serveWebUIV2 bool) http.FileSystem {
	return assetsFS{
		FileSystem:   http.FS(fs),
		serveWebUIV2: serveWebUIV2,
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
	fs := views.Assets
	if deps.Config().Http.ServeWebUIV2 {
		fs = webapp.Assets
	}
	http.StripPrefix("/assets/", http.FileServer(newAssetsFS(fs, deps.Config().Http.ServeWebUIV2))).ServeHTTP(c.ResponseWriter(), c.Request())
}
