package http

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/routes"
	"github.com/go-shiori/shiori/internal/http/routes/api"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/sirupsen/logrus"
)

type HttpServer struct {
	http   *fiber.App
	addr   string
	logger *logrus.Logger
}

func (s *HttpServer) Setup(cfg config.HttpConfig, deps *config.Dependencies) *HttpServer {
	s.http.
		Use(requestid.New(requestid.Config{
			Generator: utils.UUIDv4,
		})).
		Use(recover.New(recover.Config{
			EnableStackTrace: true,
			StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
				s.logger.WithError(e.(error)).Error("server error")
			},
		})).
		Use(middleware.NewLogrusMiddleware(middleware.LogrusMiddlewareConfig{
			Logger:      s.logger,
			CacheHeader: cache.ConfigDefault.CacheHeader,
		})).
		Mount(cfg.Routes.System.Path, routes.NewSystemRoutes(s.logger, cfg).Setup().Router()).
		Mount(cfg.Routes.API.Path, api.NewAPIRoutes(s.logger, cfg, deps).Setup().Router()).
		Mount(cfg.Routes.Bookmark.Path, routes.NewBookmarkRoutes(s.logger, deps).Setup().Router()).
		Mount(cfg.Routes.Frontend.Path, routes.NewFrontendRoutes(s.logger, cfg).Setup().Router())

	return s
}

func (s *HttpServer) Start(_ context.Context) error {
	s.logger.WithField("addr", s.addr).Info("starting http server")
	return s.http.Listen(s.addr)
}

func (s *HttpServer) Stop(ctx context.Context) error {
	s.logger.WithField("addr", s.addr).Info("stoppping http server")
	return s.http.Shutdown()
}

func (s *HttpServer) WaitStop(ctx context.Context) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	s.logger.WithField("signal", sig.String()).Info("signal received, shutting down")

	if err := s.Stop(ctx); err != nil {
		s.logger.WithError(err).Error("error stopping server")
	}
}

func NewHttpServer(logger *logrus.Logger, cfg config.HttpConfig, dependencies *config.Dependencies) *HttpServer {
	return &HttpServer{
		logger: logger,
		addr:   fmt.Sprintf("%s%d", cfg.Address, cfg.Port),
		http: fiber.New(fiber.Config{
			AppName:                      "shiori",
			PassLocalsToViews:            true,
			BodyLimit:                    cfg.BodyLimit,
			ReadTimeout:                  cfg.ReadTimeout,
			WriteTimeout:                 cfg.WriteTimeout,
			IdleTimeout:                  cfg.IDLETimeout,
			DisableKeepalive:             cfg.DisableKeepAlive,
			DisablePreParseMultipartForm: cfg.DisablePreParseMultipartForm,
		}),
	}
}
