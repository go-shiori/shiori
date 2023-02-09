package api

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type TagsAPIRoutes struct {
	logger *zap.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *TagsAPIRoutes) Setup() *TagsAPIRoutes {
	r.router.Get("/", r.listHandler)
	return r
}

func (r *TagsAPIRoutes) Router() *fiber.App {
	return r.router
}

func (r *TagsAPIRoutes) listHandler(c *fiber.Ctx) error {
	return response.Send(c, 200, []string{})
}

func NewTagsPIRoutes(logger *zap.Logger, deps *config.Dependencies) *TagsAPIRoutes {
	return &TagsAPIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
}
