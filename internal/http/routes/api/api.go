package api

import (
	"errors"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type APIRoutes struct {
	logger *logrus.Logger
	router *fiber.App
	deps   *config.Dependencies
	secret string
}

func (r *APIRoutes) Setup() *APIRoutes {
	r.router.
		Use(middleware.JSONMiddleware()).
		Use(middleware.AuthMiddleware(r.secret)).
		Mount("/account", NewAccountAPIRoutes(r.logger, r.deps).Setup().Router()).
		Mount("/bookmarks", NewBookmarksPIRoutes(r.logger, r.deps).Setup().Router()).
		Mount("/tags", NewTagsPIRoutes(r.logger, r.deps).Setup().Router())

	if r.deps.Config.Development {
		r.router.Mount("/debug", NewDebugPIRoutes(r.logger, r.deps).Setup().Router())
	}

	return r
}

func (r *APIRoutes) Router() *fiber.App {
	return r.router
}

func NewAPIRoutes(logger *logrus.Logger, cfg config.HttpConfig, deps *config.Dependencies) *APIRoutes {
	return &APIRoutes{
		logger: logger,
		router: fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				// Broken: https://github.com/gofiber/fiber/issues/2233
				code := fiber.StatusInternalServerError
				var e *fiber.Error
				if errors.As(err, &e) {
					code = e.Code
				}
				return response.SendError(c, code, "")
			},
		}),
		deps:   deps,
		secret: cfg.SecretKey,
	}
}
