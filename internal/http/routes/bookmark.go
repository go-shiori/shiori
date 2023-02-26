package routes

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type BookmarkRoutes struct {
	logger *logrus.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *BookmarkRoutes) Setup() *BookmarkRoutes {
	r.router.Get("/:id/archive", r.bookmarkArchiveHandler)
	r.router.Get("/:id/content", r.bookmarkArchiveHandler)
	return r
}

func (r *BookmarkRoutes) Router() *fiber.App {
	return r.router
}

func (r *BookmarkRoutes) bookmarkArchiveHandler(c *fiber.Ctx) error {
	bookmarkID, err := c.ParamsInt("id")
	if err != nil || bookmarkID == 0 {
		return response.SendError(c, 400, "Invalid bookmark ID")
	}

	ctx := c.Context()
	bookmark, found, err := r.deps.Database.GetBookmark(ctx, bookmarkID, "")
	if err != nil || !found {
		return response.SendError(c, 404, nil)
	}

	return response.Send(c, 200, bookmark)
}

func NewBookmarkRoutes(logger *logrus.Logger, deps *config.Dependencies) *BookmarkRoutes {
	return &BookmarkRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
}
