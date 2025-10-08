package middleware

import (
	"net/http"

	"github.com/go-shiori/shiori/internal/model"
)

type CORSMiddleware struct {
	enabled          bool
	allowedOrigins   []string
	allowCredentials bool
}

func (m *CORSMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	if !m.enabled {
		return nil
	}

	origin := c.Request().Header.Get("Origin")
	allowOrigin := m.getAllowOrigin(origin)

	if allowOrigin != "" {
		c.ResponseWriter().Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.ResponseWriter().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.ResponseWriter().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Shiori-Response-Format")

		if m.allowCredentials {
			c.ResponseWriter().Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}

	// Handle preflight requests
	if c.Request().Method == http.MethodOptions {
		c.ResponseWriter().WriteHeader(http.StatusOK)
		return nil
	}

	return nil
}

func (m *CORSMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	if !m.enabled {
		return nil
	}

	origin := c.Request().Header.Get("Origin")
	allowOrigin := m.getAllowOrigin(origin)

	if allowOrigin != "" {
		c.ResponseWriter().Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.ResponseWriter().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.ResponseWriter().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Shiori-Response-Format")

		if m.allowCredentials {
			c.ResponseWriter().Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}

	return nil
}

func (m *CORSMiddleware) getAllowOrigin(origin string) string {
	// Allow all origins if "*" is specified and credentials are not enabled
	if len(m.allowedOrigins) == 1 && m.allowedOrigins[0] == "*" && !m.allowCredentials {
		return "*"
	}

	// If no origin is provided, return empty string (no CORS headers)
	if origin == "" {
		return ""
	}

	// Check if the origin is in the allowed list
	for _, allowedOrigin := range m.allowedOrigins {
		if allowedOrigin == "*" {
			return "*"
		}
		if allowedOrigin == origin {
			return origin
		}
	}

	return ""
}

func NewCORSMiddleware(enabled bool, allowedOrigins []string, allowCredentials bool) *CORSMiddleware {
	return &CORSMiddleware{
		enabled:          enabled,
		allowedOrigins:   allowedOrigins,
		allowCredentials: allowCredentials,
	}
}
