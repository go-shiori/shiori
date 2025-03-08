package http

import (
	"net/http"

	"github.com/go-shiori/shiori/internal/http/webcontext"
	"github.com/go-shiori/shiori/internal/model"
)

// ToHTTPHandler converts a model.HttpHandler to http.HandlerFunc with dependencies and middlewares
func ToHTTPHandler(deps model.Dependencies, h model.HttpHandler, middlewares ...model.HttpMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := webcontext.NewWebContext(w, r)

		// Execute OnRequest middlewares
		for _, m := range middlewares {
			if err := m.OnRequest(deps, c); err != nil {
				// Handle middleware error
				deps.Logger().WithError(err).Error("middleware error in request")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Execute handler
		h(deps, c)

		// Execute OnResponse middlewares
		for _, m := range middlewares {
			if err := m.OnResponse(deps, c); err != nil {
				deps.Logger().WithError(err).Error("middleware error in response")
				return
			}
		}
	}
}
