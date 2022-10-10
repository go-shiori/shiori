package api

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type APIRoutes struct {
	logger *zap.Logger
	router *fiber.App
	deps   *config.Dependencies
	secret string
}

func (r *APIRoutes) Setup() *APIRoutes {
	r.router.
		Use(middleware.JSONMiddleware()).
		Mount("/auth", NewAuthAPIRoutes(r.logger, r.deps).Router()).
		Use(middleware.AuthMiddleware(r.secret)).
		Get("/private", func(c *fiber.Ctx) error {
			return response.Send(c, 200, c.Locals("account").(model.Account))
		})
	return r
}

func (r *APIRoutes) Router() *fiber.App {
	return r.router
}

func NewAPIRoutes(logger *zap.Logger, cfg config.HttpConfig, deps *config.Dependencies) *APIRoutes {
	routes := APIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
		secret: cfg.SecretKey,
	}
	routes.Setup()
	return &routes
}
