package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	fp "path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	ws "github.com/go-shiori/shiori/internal/webserver"
	"github.com/gofrs/uuid/v5"
	"github.com/sirupsen/logrus"
)

type BookmarkRoutes struct {
	logger *logrus.Logger
	deps   *dependencies.Dependencies
}

func (r *BookmarkRoutes) Setup(group *gin.RouterGroup) model.Routes {
	group.GET("/:id/archive", r.bookmarkArchiveHandler)
	group.GET("/:id/archive/file/*filepath", r.bookmarkArchiveFileHandler)
	group.GET("/:id/content", r.bookmarkContentHandler)
	group.GET("/:id/thumb", r.bookmarkThumbnailHandler)
	group.GET("/:id/ebook", r.bookmarkEbookHandler)

	return r
}

func (r *BookmarkRoutes) getBookmark(c *context.Context) (*model.BookmarkDTO, error) {
	bookmarkIDParam, present := c.Params.Get("id")
	if !present {
		response.SendError(c.Context, 400, "Invalid bookmark ID")
	}

	bookmarkID, err := strconv.Atoi(bookmarkIDParam)
	if err != nil {
		r.logger.WithError(err).Error("error parsing bookmark ID parameter")
		response.SendInternalServerError(c.Context)
		return nil, err
	}

	if bookmarkID == 0 {
		response.SendError(c.Context, 404, nil)
		return nil, err
	}

	bookmark, err := r.deps.Domains.Bookmarks.GetBookmark(c.Context, model.DBID(bookmarkID))
	if err != nil {
		response.SendError(c.Context, 404, nil)
		return nil, err
	}

	if bookmark.Public != 1 && !c.UserIsLogged() {
		response.RedirectToLogin(c.Context, c.Request.URL.String())
		return nil, err
	}

	return bookmark, nil
}

func (r *BookmarkRoutes) bookmarkContentHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	bookmark, err := r.getBookmark(ctx)
	if err != nil {
		return
	}

	ctx.HTML(http.StatusOK, "content.html", gin.H{
		"RootPath": r.deps.Config.Http.RootPath,
		"Version":  model.BuildVersion,
		"Book":     bookmark,
		"HTML":     template.HTML(bookmark.HTML),
	})
}

func (r *BookmarkRoutes) bookmarkArchiveHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	bookmark, err := r.getBookmark(ctx)
	if err != nil {
		return
	}

	if !r.deps.Domains.Bookmarks.HasArchive(bookmark) {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}

	c.HTML(http.StatusOK, "archive.html", gin.H{
		"RootPath": r.deps.Config.Http.RootPath,
		"Version":  model.BuildVersion,
		"Book":     bookmark,
	})
}

func (r *BookmarkRoutes) bookmarkArchiveFileHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	bookmark, err := r.getBookmark(ctx)
	if err != nil {
		return
	}

	if !r.deps.Domains.Bookmarks.HasArchive(bookmark) {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}

	resourcePath, _ := c.Params.Get("filepath")
	resourcePath = strings.TrimPrefix(resourcePath, "/")

	archive, err := r.deps.Domains.Archiver.GetBookmarkArchive(bookmark)
	if err != nil {
		r.logger.WithError(err).Error("error opening archive")
		response.SendInternalServerError(c)
		return
	}
	defer archive.Close()

	if !archive.HasResource(resourcePath) {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}

	content, resourceContentType, err := archive.Read(resourcePath)
	if err != nil {
		r.logger.WithError(err).Error("error reading archive file")
		response.SendInternalServerError(c)
		return
	}

	// Generate weak ETAG
	shioriUUID := uuid.NewV5(uuid.NamespaceURL, model.ShioriURLNamespace)
	c.Header("Etag", fmt.Sprintf("W/%s", uuid.NewV5(shioriUUID, fmt.Sprintf("%x-%x-%x", bookmark.ID, resourcePath, len(content)))))
	c.Header("Cache-Control", "max-age=31536000")

	c.Header("Content-Encoding", "gzip")
	c.Data(http.StatusOK, resourceContentType, content)
}

func (r *BookmarkRoutes) bookmarkThumbnailHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	bookmark, err := r.getBookmark(ctx)
	if err != nil {
		return
	}

	if !r.deps.Domains.Bookmarks.HasThumbnail(bookmark) {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}

	response.SendFile(c, r.deps.Domains.Bookmarks.GetThumbnailPath(bookmark))
}

func NewBookmarkRoutes(logger *logrus.Logger, deps *dependencies.Dependencies) *BookmarkRoutes {
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
