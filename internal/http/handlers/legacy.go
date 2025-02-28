package handlers

import (
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/webserver"
	"github.com/gofrs/uuid/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

type LegacyHandler struct {
	legacyHandler *webserver.Handler
}

func NewLegacyHandler(deps model.Dependencies) *LegacyHandler {
	handler := webserver.GetLegacyHandler(webserver.Config{
		DB:       deps.Database(),
		DataDir:  deps.Config().Storage.DataDir,
		RootPath: deps.Config().Http.RootPath,
		Log:      false, // Already handled by middleware
	}, deps)
	handler.PrepareSessionCache()

	return &LegacyHandler{
		legacyHandler: handler,
	}
}

// convertParams converts standard URL parameters to httprouter.Params
func (h *LegacyHandler) convertParams(r *http.Request) httprouter.Params {
	routerParams := httprouter.Params{}
	for key, value := range r.URL.Query() {
		routerParams = append(routerParams, httprouter.Param{
			Key:   key,
			Value: value[0],
		})
	}

	return routerParams
}

// HandleLogin handles the legacy login endpoint
func (h *LegacyHandler) HandleLogin(account *model.AccountDTO, expTime time.Duration) (string, error) {
	sessionID, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(err, "failed to create session ID")
	}

	strSessionID := sessionID.String()

	return strSessionID, nil
}

// HandleLogout handles the legacy logout endpoint
func (h *LegacyHandler) HandleLogout(deps model.Dependencies, c model.WebContext) {
	// TODO: Leave cookie handling to API consumer or middleware?
	// Remove token cookie
	c.Request().AddCookie(&http.Cookie{
		Name:  "token",
		Value: "",
	})
}

// HandleGetTags handles GET /api/tags
func (h *LegacyHandler) HandleGetTags(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiGetTags(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleRenameTag handles PUT /api/tags
func (h *LegacyHandler) HandleRenameTag(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiRenameTag(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleGetBookmarks handles GET /api/bookmarks
func (h *LegacyHandler) HandleGetBookmarks(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiGetBookmarks(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleInsertBookmark handles POST /api/bookmarks
func (h *LegacyHandler) HandleInsertBookmark(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiInsertBookmark(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleDeleteBookmark handles DELETE /api/bookmarks
func (h *LegacyHandler) HandleDeleteBookmark(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiDeleteBookmark(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleUpdateBookmark handles PUT /api/bookmarks
func (h *LegacyHandler) HandleUpdateBookmark(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiUpdateBookmark(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleUpdateBookmarkTags handles PUT /api/bookmarks/tags
func (h *LegacyHandler) HandleUpdateBookmarkTags(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiUpdateBookmarkTags(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleInsertViaExtension handles POST /api/bookmarks/ext
func (h *LegacyHandler) HandleInsertViaExtension(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiInsertViaExtension(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}

// HandleDeleteViaExtension handles DELETE /api/bookmarks/ext
func (h *LegacyHandler) HandleDeleteViaExtension(deps model.Dependencies, c model.WebContext) {
	h.legacyHandler.ApiDeleteViaExtension(c.ResponseWriter(), c.Request(), h.convertParams(c.Request()))
}
