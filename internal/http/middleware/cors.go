package middleware

import (
	"strings"

	"github.com/go-shiori/shiori/internal/model"
)

type CORSMiddleware struct {
	allowedOrigins []string
}

func (m *CORSMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	c.ResponseWriter().Header().Set("Access-Control-Allow-Origin", strings.Join(m.allowedOrigins, ", "))
	c.ResponseWriter().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.ResponseWriter().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Shiori-Response-Format")
	return nil
}

func (m *CORSMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	c.ResponseWriter().Header().Set("Access-Control-Allow-Origin", strings.Join(m.allowedOrigins, ", "))
	c.ResponseWriter().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.ResponseWriter().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Shiori-Response-Format")
	return nil
}

func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{allowedOrigins: allowedOrigins}
}
