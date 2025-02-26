package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestNewHttpServer(t *testing.T) {
	logger := logrus.New()
	server := NewHttpServer(logger)
	require.NotNil(t, server)
	require.Equal(t, logger, server.logger)
}

func TestHttpServer_Setup(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("successful setup", func(t *testing.T) {
		cfg, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		server := NewHttpServer(logger)

		s, err := server.Setup(cfg, deps)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.NotNil(t, s.mux)
		require.NotNil(t, s.server)
		require.Equal(t, fmt.Sprintf("%s%d", cfg.Http.Address, cfg.Http.Port), s.server.Addr)
	})

	t.Run("routes are registered correctly", func(t *testing.T) {
		cfg, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		server := NewHttpServer(logger)

		s, err := server.Setup(cfg, deps)
		require.NoError(t, err)

		// Test some key routes
		routes := []struct {
			method string
			path   string
			want   int
		}{
			{"GET", "/system/liveness", http.StatusOK},
			{"GET", "/api/v1/system/info", http.StatusUnauthorized}, // Requires auth
			{"GET", "/api/v1/accounts", http.StatusUnauthorized},    // Requires auth
			{"POST", "/api/v1/auth/login", http.StatusBadRequest},   // Bad request because no body
		}

		for _, tt := range routes {
			t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
				req := httptest.NewRequest(tt.method, tt.path, nil)
				w := httptest.NewRecorder()
				s.mux.ServeHTTP(w, req)
				require.Equal(t, tt.want, w.Code)
			})
		}
	})

	t.Run("swagger routes when enabled", func(t *testing.T) {
		cfg, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		cfg.Http.ServeSwagger = true
		server := NewHttpServer(logger)

		s, err := server.Setup(cfg, deps)
		require.NoError(t, err)

		// Test swagger doc endpoint
		req := httptest.NewRequest("GET", "/swagger/doc.json", nil)
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Test swagger UI endpoint (should redirect)
		req = httptest.NewRequest("GET", "/swagger/", nil)
		w = httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
		require.Equal(t, http.StatusMovedPermanently, w.Code)
		require.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
	})

	t.Run("web UI routes when enabled", func(t *testing.T) {
		cfg, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		cfg.Http.ServeWebUI = true
		server := NewHttpServer(logger)

		s, err := server.Setup(cfg, deps)
		require.NoError(t, err)

		routes := []struct {
			path string
			want int
		}{
			{"/", http.StatusOK},
			{"/assets/style.css", http.StatusNotFound}, // 404 because no actual assets in test
		}

		for _, tt := range routes {
			t.Run(tt.path, func(t *testing.T) {
				req := httptest.NewRequest("GET", tt.path, nil)
				w := httptest.NewRecorder()
				s.mux.ServeHTTP(w, req)
				require.Equal(t, tt.want, w.Code)
			})
		}
	})
}

func TestHttpServer_StartStop(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()
	cfg, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

	// Use a random port to avoid conflicts
	cfg.Http.Port = 0

	server := NewHttpServer(logger)
	s, err := server.Setup(cfg, deps)
	require.NoError(t, err)

	// Start the server
	err = s.Start(ctx)
	require.NoError(t, err)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the server
	err = s.Stop(ctx)
	require.NoError(t, err)
}

func TestHttpServer_Middleware(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()
	cfg, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	server := NewHttpServer(logger)

	s, err := server.Setup(cfg, deps)
	require.NoError(t, err)

	t.Run("logging middleware", func(t *testing.T) {
		// Capture log output
		var logBuf strings.Builder
		logger.SetOutput(&logBuf)
		logger.SetLevel(logrus.InfoLevel)

		req := httptest.NewRequest("GET", "/system/liveness", nil)
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)

		// Verify log contains request info
		logOutput := logBuf.String()
		require.Contains(t, logOutput, "request completed")
		require.Contains(t, logOutput, "path=/system/liveness")
	})

	t.Run("auth middleware", func(t *testing.T) {
		protectedRoutes := []struct {
			method string
			path   string
			want   int
			auth   bool
		}{
			{"GET", "/api/v1/accounts", http.StatusUnauthorized, false},
			{"GET", "/api/v1/auth/me", http.StatusUnauthorized, false},
			{"PUT", "/api/v1/bookmarks/cache", http.StatusForbidden, true}, // Requires admin access
		}

		for _, route := range protectedRoutes {
			t.Run(route.path, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)

				if route.auth {
					// Create a non-admin user token
					account := testutil.GetValidAccount()
					account.Owner = false // Ensure not admin
					accountDTO := account.ToDTO()
					token, err := deps.Domains().Auth().CreateTokenForAccount(&accountDTO, time.Now().Add(time.Hour))
					require.NoError(t, err)
					req.Header.Set(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token)
				}

				w := httptest.NewRecorder()
				s.mux.ServeHTTP(w, req)
				require.Equal(t, route.want, w.Code)
			})
		}
	})
}

func TestHttpServer_APIEndpoints(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()
	cfg, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	server := NewHttpServer(logger)

	s, err := server.Setup(cfg, deps)
	require.NoError(t, err)

	t.Run("login endpoint", func(t *testing.T) {
		body := strings.NewReader(`{"username": "test", "password": "test"}`)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		respBody, _ := io.ReadAll(w.Body)
		require.Contains(t, string(respBody), "username or password do not match")
	})

	t.Run("system info endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/system/info", nil)
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		respBody, _ := io.ReadAll(w.Body)
		require.Contains(t, string(respBody), "Authentication required")
	})
}
