package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/routes"
	api_v1 "github.com/go-shiori/shiori/internal/http/routes/api/v1"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

type HttpServer struct {
	engine *gin.Engine
	http   *http.Server
	logger *logrus.Logger
}

func (s *HttpServer) Setup(cfg config.HttpConfig, deps *config.Dependencies) *HttpServer {
	s.engine.Use(
		requestid.New(),
		ginlogrus.Logger(deps.Log),
		middleware.AuthMiddleware(deps),
		gin.Recovery(),
	)

	if !deps.Config.Development {
		gin.SetMode(gin.ReleaseMode)
	}

	s.handle("/system", routes.NewSystemRoutes(s.logger))
	s.handle("/bookmark", routes.NewBookmarkRoutes(s.logger, deps))
	s.handle("/api/v1", api_v1.NewAPIRoutes(s.logger, deps))
	s.handle("/swagger", routes.NewSwaggerAPIRoutes(s.logger))
	routes.NewFrontendRoutes(s.logger, cfg).Setup(s.engine)

	s.http.Handler = s.engine
	s.http.Addr = fmt.Sprintf("%s%d", cfg.Address, cfg.Port)

	return s
}

func (s *HttpServer) handle(path string, routes model.Routes) {
	group := s.engine.Group(path)
	routes.Setup(group)
}

func (s *HttpServer) Start(_ context.Context) error {
	s.logger.WithField("addr", s.http.Addr).Info("starting http server")
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("listen and serve error: %s\n", err)
		}
	}()
	return nil
}

func (s *HttpServer) Stop(ctx context.Context) error {
	s.logger.WithField("addr", s.http.Addr).Info("stoppping http server")
	return s.http.Shutdown(ctx)
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
		http:   &http.Server{},
		engine: gin.New(),
	}
}
