package routes

import (
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/config"
	views "github.com/go-shiori/shiori/internal/view"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"go.uber.org/zap"
)

type FrontendRoutes struct {
	logger *zap.Logger
	router *fiber.App
	maxAge time.Duration
}

func (r *FrontendRoutes) Setup() *FrontendRoutes {
	r.router.
		Use(compress.New()).
		Use("/", filesystem.New(filesystem.Config{
			Browse:       false,
			MaxAge:       int(r.maxAge.Seconds()),
			Root:         http.FS(views.Assets),
			NotFoundFile: "404.html",
		}))
	return r
}

func (r *FrontendRoutes) Router() *fiber.App {
	return r.router
}

func NewFrontendRoutes(logger *zap.Logger, cfg config.HttpConfig) *FrontendRoutes {
	routes := FrontendRoutes{
		logger: logger,
		router: fiber.New(),
		maxAge: cfg.Routes.Frontend.MaxAge,
	}
	routes.Setup()
	return &routes
}
