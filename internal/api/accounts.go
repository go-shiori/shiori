package api

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetAccounts Get all accounts
func (s ShioriServer) GetAccounts(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// CreateAccount Create new account
func (s ShioriServer) CreateAccount(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// DeleteAccount Delete account
func (s ShioriServer) DeleteAccount(ctx echo.Context, accountID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// GetAccount Get single account
func (s ShioriServer) GetAccount(ctx echo.Context, accountID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// EditAccount Change account details
func (s ShioriServer) EditAccount(ctx echo.Context, accountID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// ChangePassword Change password for account
func (s ShioriServer) ChangePassword(ctx echo.Context, accountID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}
