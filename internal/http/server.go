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
	api_v1 "github.com/go-shiori/shiori/internal/http/handlers/api/v1"
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
		middleware.NewAuthMiddleware(deps),
		middleware.NewRequestIDMiddleware(deps),
	}

	if cfg.Http.AccessLog {
		globalMiddleware = append(globalMiddleware, middleware.NewLoggingMiddleware())
	}

	// System routes with logging middleware
	s.mux.HandleFunc("GET /system/liveness", ToHTTPHandler(deps,
		handlers.HandleLiveness,
		globalMiddleware...,
	))

	// Bookmark routes
	s.mux.HandleFunc("GET /bookmark/{id}/content", ToHTTPHandler(deps, handlers.HandleBookmarkContent, globalMiddleware...))
	s.mux.HandleFunc("GET /bookmark/{id}/archive", ToHTTPHandler(deps, handlers.HandleBookmarkArchive, globalMiddleware...))
	s.mux.HandleFunc("GET /bookmark/{id}/archive/file/{path...}", ToHTTPHandler(deps, handlers.HandleBookmarkArchiveFile, globalMiddleware...))
	s.mux.HandleFunc("GET /bookmark/{id}/thumb", ToHTTPHandler(deps, handlers.HandleBookmarkThumbnail, globalMiddleware...))
	s.mux.HandleFunc("GET /bookmark/{id}/ebook", ToHTTPHandler(deps, handlers.HandleBookmarkEbook, globalMiddleware...))

	// Add this inside Setup() where other routes are registered
	if cfg.Http.ServeSwagger {
		s.mux.HandleFunc("/swagger/", ToHTTPHandler(deps,
			handlers.HandleSwagger,
			globalMiddleware...,
		))
	}

	// API v1 routes
	s.mux.HandleFunc("GET /api/v1/system/info", ToHTTPHandler(deps,
		api_v1.HandleSystemInfo,
		globalMiddleware...,
	))

	// Legacy API routes
	// TODO: Remove this once the legacy API is removed
	legacyHandler := handlers.NewLegacyHandler(deps)

	s.mux.HandleFunc("GET /api/tags", ToHTTPHandler(deps, legacyHandler.HandleGetTags, globalMiddleware...))
	s.mux.HandleFunc("PUT /api/tags", ToHTTPHandler(deps, legacyHandler.HandleRenameTag, globalMiddleware...))
	s.mux.HandleFunc("GET /api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleGetBookmarks, globalMiddleware...))
	s.mux.HandleFunc("POST /api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleInsertBookmark, globalMiddleware...))
	s.mux.HandleFunc("DELETE /api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleDeleteBookmark, globalMiddleware...))
	s.mux.HandleFunc("PUT /api/bookmarks", ToHTTPHandler(deps, legacyHandler.HandleUpdateBookmark, globalMiddleware...))
	s.mux.HandleFunc("PUT /api/bookmarks/tags", ToHTTPHandler(deps, legacyHandler.HandleUpdateBookmarkTags, globalMiddleware...))
	s.mux.HandleFunc("POST /api/bookmarks/ext", ToHTTPHandler(deps, legacyHandler.HandleInsertViaExtension, globalMiddleware...))
	s.mux.HandleFunc("DELETE /api/bookmarks/ext", ToHTTPHandler(deps, legacyHandler.HandleDeleteViaExtension, globalMiddleware...))

	// Register routes using standard http handlers
	if cfg.Http.ServeWebUI {
		// Frontend routes
		s.mux.HandleFunc("/", ToHTTPHandler(deps,
			handlers.HandleFrontend,
			globalMiddleware...,
		))
		s.mux.HandleFunc("GET /assets/", ToHTTPHandler(deps,
			handlers.HandleAssets,
			globalMiddleware...,
		))
	}

	// API v1 routes
	// Auth
	s.mux.HandleFunc("POST /api/v1/auth/login", ToHTTPHandler(deps,
		api_v1.HandleLogin,
		globalMiddleware...,
	))
	s.mux.HandleFunc("POST /api/v1/auth/refresh", ToHTTPHandler(deps,
		api_v1.HandleRefreshToken,
		globalMiddleware...,
	))
	s.mux.HandleFunc("GET /api/v1/auth/me", ToHTTPHandler(deps,
		api_v1.HandleGetMe,
		globalMiddleware...,
	))
	s.mux.HandleFunc("PATCH /api/v1/auth/account", ToHTTPHandler(deps,
		api_v1.HandleUpdateLoggedAccount,
		globalMiddleware...,
	))
	s.mux.HandleFunc("POST /api/v1/auth/logout", ToHTTPHandler(deps,
		api_v1.HandleLogout,
		globalMiddleware...,
	))
	// Accounts
	s.mux.HandleFunc("GET /api/v1/accounts", ToHTTPHandler(deps,
		api_v1.HandleListAccounts,
		globalMiddleware...,
	))
	s.mux.HandleFunc("POST /api/v1/accounts", ToHTTPHandler(deps,
		api_v1.HandleCreateAccount,
		globalMiddleware...,
	))
	s.mux.HandleFunc("DELETE /api/v1/accounts/{id}", ToHTTPHandler(deps,
		api_v1.HandleDeleteAccount,
		globalMiddleware...,
	))
	s.mux.HandleFunc("PATCH /api/v1/accounts/{id}", ToHTTPHandler(deps,
		api_v1.HandleUpdateAccount,
		globalMiddleware...,
	))
	// Tags
	s.mux.HandleFunc("GET /api/v1/tags", ToHTTPHandler(deps,
		api_v1.HandleListTags,
		globalMiddleware...,
	))
	// Bookmarks
	s.mux.HandleFunc("PUT /api/v1/bookmarks/cache", ToHTTPHandler(deps,
		api_v1.HandleUpdateCache,
		globalMiddleware...,
	))
	s.mux.HandleFunc("GET /api/v1/bookmarks/{id}/readable", ToHTTPHandler(deps,
		api_v1.HandleBookmarkReadable,
		globalMiddleware...,
	))

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
