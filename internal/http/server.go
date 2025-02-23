package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/handlers"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/sirupsen/logrus"
)

type HttpServer struct {
	mux    *http.ServeMux
	server *http.Server
	logger *logrus.Logger
}

func (s *HttpServer) Setup(cfg *config.Config, deps *dependencies.Dependencies) (*HttpServer, error) {
	s.mux = http.NewServeMux()

	// Register routes using standard http handlers
	// if cfg.Http.ServeWebUI {
	// 	s.mux.Handle("/", NewFrontendHandler(s.logger, cfg))
	// }

	// API routes with auth
	// apiHandler := handlers.NewAPIHandler(s.logger, deps)
	// s.mux.Handle("/api/v1/", middleware.NewAuthMiddleware(deps).OnRequest(deps, c))

	// System routes with logging middleware
	s.mux.HandleFunc("/system/liveness", ToHTTPHandler(deps,
		handlers.HandleLiveness,
		middleware.NewLoggingMiddleware(),
	))

	// Bookmark routes
	// s.mux.Handle("/bookmark/", http.StripPrefix("/bookmark", NewBookmarkHandler(s.logger, deps)))

	// if cfg.Http.ServeSwagger {
	// 	s.mux.Handle("/swagger/", http.StripPrefix("/swagger", NewSwaggerHandler(s.logger)))
	// }

	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s%d", cfg.Http.Address, cfg.Http.Port),
		Handler: s.mux,
	}

	return s, nil
}

func (s *HttpServer) Start(_ context.Context) error {
	s.logger.WithField("addr", s.server.Addr).Info("starting http server")
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("listen and serve error: %s\n", err)
		}
	}()
	return nil
}

func (s *HttpServer) Stop(ctx context.Context) error {
	s.logger.WithField("addr", s.server.Addr).Info("stopping http server")
	return s.server.Shutdown(ctx)
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

func NewHttpServer(logger *logrus.Logger) *HttpServer {
	return &HttpServer{
		logger: logger,
	}
}
