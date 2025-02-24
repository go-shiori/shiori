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
	"github.com/go-shiori/shiori/internal/http/templates"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type HttpServer struct {
	mux    *http.ServeMux
	server *http.Server
	logger *logrus.Logger
}

func (s *HttpServer) Setup(cfg *config.Config, deps *dependencies.Dependencies) (*HttpServer, error) {
	s.mux = http.NewServeMux()

	if err := templates.SetupTemplates(); err != nil {
		return nil, fmt.Errorf("failed to setup templates: %w", err)
	}

	globalMiddleware := []model.HttpMiddleware{
		middleware.NewLoggingMiddleware(),
		middleware.NewAuthMiddleware(deps),
	}

	// Register routes using standard http handlers
	if cfg.Http.ServeWebUI {
		// Frontend routes
		s.mux.HandleFunc("/", ToHTTPHandler(deps,
			handlers.HandleFrontend,
			globalMiddleware...,
		))
		s.mux.HandleFunc("/assets/", ToHTTPHandler(deps,
			handlers.HandleAssets,
			globalMiddleware...,
		))

	}

	// System routes with logging middleware
	s.mux.HandleFunc("/system/liveness", ToHTTPHandler(deps,
		handlers.HandleLiveness,
		globalMiddleware...,
	))

	// Bookmark routes
	// s.mux.Handle("/bookmark/", http.StripPrefix("/bookmark", NewBookmarkHandler(s.logger, deps)))

	// Add this inside Setup() where other routes are registered
	if cfg.Http.ServeSwagger {
		s.mux.HandleFunc("/swagger/", ToHTTPHandler(deps,
			handlers.HandleSwagger,
			globalMiddleware...,
		))
	}

	// Legacy API routes
	// TODO: Remove this once the legacy API is removed
	legacyHandler := handlers.NewLegacyHandler(deps)

	s.mux.HandleFunc("/api/tags", ToHTTPHandler(deps, legacyHandler.HandleGetTags, globalMiddleware...))
	s.mux.HandleFunc("/api/tags", ToHTTPHandler(deps, legacyHandler.HandleRenameTag, globalMiddleware...))
	s.mux.HandleFunc("/api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleGetBookmarks, globalMiddleware...))
	s.mux.HandleFunc("/api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleInsertBookmark, globalMiddleware...))
	s.mux.HandleFunc("/api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleDeleteBookmark, globalMiddleware...))
	s.mux.HandleFunc("/api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleUpdateBookmark, globalMiddleware...))
	s.mux.HandleFunc("/api/bookmarks/tags", ToHTTPHandler(deps, legacyHandler.HandleUpdateBookmarkTags, globalMiddleware...))
	s.mux.HandleFunc("/api/bookmarks/ext", ToHTTPHandler(deps, legacyHandler.HandleInsertViaExtension, globalMiddleware...))
	s.mux.HandleFunc("/api/bookmarks/ext", ToHTTPHandler(deps, legacyHandler.HandleDeleteViaExtension, globalMiddleware...))

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
