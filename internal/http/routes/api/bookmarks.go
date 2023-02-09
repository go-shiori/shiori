package api

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type BookmarksAPIRoutes struct {
	logger *zap.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *BookmarksAPIRoutes) Setup() *BookmarksAPIRoutes {
	r.router.Get("/", r.listHandler)
	return r
}

func (r *BookmarksAPIRoutes) Router() *fiber.App {
	return r.router
}

func (r *BookmarksAPIRoutes) listHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	bookmarks, err := r.deps.Database.GetBookmarks(ctx, database.GetBookmarksOptions{})
	if err != nil {
		return errors.Wrap(err, "error getting bookmakrs")
	}
	return response.Send(c, 200, bookmarks)
}

func NewBookmarksPIRoutes(logger *zap.Logger, deps *config.Dependencies) *BookmarksAPIRoutes {
	return &BookmarksAPIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
}
