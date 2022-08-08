package api

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetTags Get all tags
func (s ShioriServer) GetTags(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// CreateTag Create new tag
func (s ShioriServer) CreateTag(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// DeleteTag Delete tag
func (s ShioriServer) DeleteTag(ctx echo.Context, tagID uint32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// GetTag Get single tag
func (s ShioriServer) GetTag(ctx echo.Context, tagID uint32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// EditTag Modify tag
func (s ShioriServer) EditTag(ctx echo.Context, tagID uint32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}
