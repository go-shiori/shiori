package api

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetBookmarks Get bookmarks
func (s ShioriServer) GetBookmarks(ctx echo.Context, params GetBookmarksParams) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// CreateBookmark Create bookmark
func (s ShioriServer) CreateBookmark(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// ProbeBookmark Probe for bookmark existence
func (s ShioriServer) ProbeBookmark(ctx echo.Context, params ProbeBookmarkParams) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// DeleteBookmark Delete bookmark
func (s ShioriServer) DeleteBookmark(ctx echo.Context, bookmarkID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// GetBookmark Get single bookmark
func (s ShioriServer) GetBookmark(ctx echo.Context, bookmarkID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// EditBookmark Modify bookmark
func (s ShioriServer) EditBookmark(ctx echo.Context, bookmarkID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// RefreshBookmark Refresh bookmark
func (s ShioriServer) RefreshBookmark(ctx echo.Context, bookmarkID int32) error {
	return ctx.NoContent(http.StatusNotImplemented)
}
