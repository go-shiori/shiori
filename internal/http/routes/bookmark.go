package routes

import (
	"net/http"
	"strconv"

	fp "path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	ws "github.com/go-shiori/shiori/internal/webserver"
	"github.com/sirupsen/logrus"
)

type BookmarkRoutes struct {
	logger *logrus.Logger
	deps   *config.Dependencies
}

func (r *BookmarkRoutes) Setup(group *gin.RouterGroup) model.Routes {
	//group.GET("/:id/archive", r.bookmarkArchiveHandler)
	//group.GET("/:id/content", r.bookmarkContentHandler)
	group.GET("/:id/ebook", r.bookmarkEbookHandler)
	return r
}

// func (r *BookmarkRoutes) bookmarkContentHandler(c *gin.Context) {
// 	ctx := context.NewContextFromGin(c)

// 	bookmarkIDParam, present := c.Params.Get("id")
// 	if !present {
// 		response.SendError(c, 400, "Invalid bookmark ID")
// 		return
// 	}

// 	bookmarkID, err := strconv.Atoi(bookmarkIDParam)
// 	if err != nil {
// 		r.logger.WithError(err).Error("error parsing bookmark ID parameter")
// 		response.SendInternalServerError(c)
// 		return
// 	}

// 	if bookmarkID == 0 {
// 		response.SendError(c, 404, nil)
// 		return
// 	}

// 	bookmark, found, err := r.deps.Database.GetBookmark(c, bookmarkID, "")
// 	if err != nil || !found {
// 		response.SendError(c, 404, nil)
// 		return
// 	}

// 	if bookmark.Public != 1 && !ctx.UserIsLogged() {
// 		response.SendError(c, http.StatusForbidden, nil)
// 		return
// 	}

// 	response.Send(c, 200, bookmark.Content)
// }

// func (r *BookmarkRoutes) bookmarkArchiveHandler(c *gin.Context) {}

func NewBookmarkRoutes(logger *logrus.Logger, deps *config.Dependencies) *BookmarkRoutes {
	return &BookmarkRoutes{
		logger: logger,
		deps:   deps,
	}
}

func (r *BookmarkRoutes) bookmarkEbookHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	// Get server config
	logger := logrus.New()
	cfg := config.ParseServerConfiguration(ctx, logger)
	DataDir := cfg.Storage.DataDir

	bookmarkIDParam, present := c.Params.Get("id")
	if !present {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	bookmarkID, err := strconv.Atoi(bookmarkIDParam)
	if err != nil {
		r.logger.WithError(err).Error("error parsing bookmark ID parameter")
		response.SendInternalServerError(c)
		return
	}

	if bookmarkID == 0 {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}

	bookmark, found, err := r.deps.Database.GetBookmark(c, bookmarkID, "")
	if err != nil || !found {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}

	if bookmark.Public != 1 && !ctx.UserIsLogged() {
		response.SendError(c, http.StatusUnauthorized, nil)
		return
	}

	ebookPath := fp.Join(DataDir, "ebook", bookmarkIDParam+".epub")
	if !ws.FileExists(ebookPath) {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}
	filename := bookmark.Title + ".epub"
	c.FileAttachment(ebookPath, filename)
}
