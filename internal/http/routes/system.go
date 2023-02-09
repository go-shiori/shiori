package routes

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type SystemRoutes struct {
	logger *zap.Logger
	router *fiber.App
}

func (r *SystemRoutes) Setup() *SystemRoutes {
	r.router.
		Get("/liveness", r.livenessHandler)
	return r
}

func (r *SystemRoutes) Router() *fiber.App {
	return r.router
}

func (r *SystemRoutes) livenessHandler(c *fiber.Ctx) error {
	return response.Send(c, 200, "ok")
}

func NewSystemRoutes(logger *zap.Logger, _ config.HttpConfig) *SystemRoutes {
	return &SystemRoutes{
		logger: logger,
		router: fiber.New(),
	}
}
