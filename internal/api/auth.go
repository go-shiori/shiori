package api

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// Login Authenticates user
func (s ShioriServer) Login(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Logout Logs out user
func (s ShioriServer) Logout(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}
