package handlers

import (
	"net/http"

	"github.com/go-shiori/shiori/internal/model"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/go-shiori/shiori/docs/swagger" // swagger docs
)

// HandleSwagger serves the swagger documentation UI
func HandleSwagger(deps model.Dependencies, c model.WebContext) {
	// Redirect /swagger to /swagger/
	path := c.Request().URL.Path
	if path == "/swagger" {
		http.Redirect(c.ResponseWriter(), c.Request(), "/swagger/index.html", http.StatusPermanentRedirect)
		return
	}

	// Strip /swagger prefix and serve swagger UI
	handler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // URL pointing to API definition
	)
	http.StripPrefix("/swagger", handler).ServeHTTP(c.ResponseWriter(), c.Request())
}
