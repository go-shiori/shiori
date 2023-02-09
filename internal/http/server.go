package http

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/http/routes"
	"github.com/go-shiori/shiori/internal/http/routes/api"
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

type HttpServer struct {
	http   *fiber.App
	addr   string
	logger *zap.Logger
}

func (s *HttpServer) Setup(cfg config.HttpConfig, deps *config.Dependencies) *HttpServer {
	fiberzapConfig := fiberzap.ConfigDefault
	fiberzapConfig.Logger = s.logger

	s.http.
		Use(requestid.New(requestid.Config{
			Generator: utils.UUIDv4,
		})).
		Use(fiberzap.New(fiberzapConfig)).
		Use(recover.New()).
		Mount(cfg.Routes.System.Path, routes.NewSystemRoutes(s.logger, cfg).Setup().Router()).
		Mount(cfg.Routes.API.Path, api.NewAPIRoutes(s.logger, cfg, deps).Setup().Router()).
		Mount(cfg.Routes.Frontend.Path, routes.NewFrontendRoutes(s.logger, cfg).Setup().Router())

	return s
}

func (s *HttpServer) Start(_ context.Context) error {
	s.logger.Info("starting http server", zap.String("addr", s.addr))
	return s.http.Listen(s.addr)
}

func (s *HttpServer) Stop(ctx context.Context) error {
	s.logger.Info("stoppping http server", zap.String("address", s.addr))
	return s.http.Shutdown()
}

func (s *HttpServer) WaitStop(ctx context.Context) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	s.logger.Info("signal received, shutting down", zap.String("signal", sig.String()))

	if err := s.Stop(ctx); err != nil {
		s.logger.Error("error stopping server", zap.Error(err))
	}
}

func NewHttpServer(logger *zap.Logger, cfg config.HttpConfig, dependencies *config.Dependencies) *HttpServer {
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
	}
}
