package api

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type APIRoutes struct {
	logger *zap.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *APIRoutes) Setup() *APIRoutes {
	r.router.
		Use(middleware.JSONMiddleware()).
		Mount("/auth", NewAuthAPIRoutes(r.logger, r.deps).Router())
	return r
}

func (r *APIRoutes) Router() *fiber.App {
	return r.router
}

func NewAPIRoutes(logger *zap.Logger, _ config.HttpConfig, deps *config.Dependencies) *APIRoutes {
	routes := APIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
	routes.Setup()
	return &routes
}
