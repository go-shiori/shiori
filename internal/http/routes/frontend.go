package routes

import (
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/frontend"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/sirupsen/logrus"
)

type FrontendRoutes struct {
	logger *logrus.Logger
	router *fiber.App
	maxAge time.Duration
}

func (r *FrontendRoutes) Setup() *FrontendRoutes {
	r.router.
		Use(compress.New()).
		Use("/", filesystem.New(filesystem.Config{
			Browse:       false,
			MaxAge:       int(r.maxAge.Seconds()),
			Root:         http.FS(frontend.Assets),
			NotFoundFile: "404.html",
		}))
	return r
}

func (r *FrontendRoutes) Router() *fiber.App {
	return r.router
}

func NewFrontendRoutes(logger *logrus.Logger, cfg config.HttpConfig) *FrontendRoutes {
	return &FrontendRoutes{
		logger: logger,
		router: fiber.New(),
		maxAge: cfg.Routes.Frontend.MaxAge,
	}
}
