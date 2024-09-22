package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofrs/uuid/v5"
	"github.com/sirupsen/logrus"
)

type BookmarkRoutes struct {
	logger *logrus.Logger
	deps   *dependencies.Dependencies
}

func NewBookmarkRoutes(logger *logrus.Logger, deps *dependencies.Dependencies) *BookmarkRoutes {
	return &BookmarkRoutes{
		logger: logger,
		deps:   deps,
	}
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
		response.SendError(c.Context, http.StatusBadRequest, "Invalid bookmark ID")
		return nil, model.ErrBookmarkInvalidID
	}

	bookmarkID, err := strconv.Atoi(bookmarkIDParam)
	if err != nil {
		r.logger.WithError(err).Error("error parsing bookmark ID parameter")
		response.SendInternalServerError(c.Context)
		return nil, err
	}

	if bookmarkID == 0 {
		response.SendError(c.Context, http.StatusNotFound, nil)
		return nil, model.ErrBookmarkNotFound
	}

	bookmark, err := r.deps.Domains.Bookmarks.GetBookmark(c.Context, model.DBID(bookmarkID))
	if err != nil {
		response.SendError(c.Context, http.StatusNotFound, nil)
		return nil, model.ErrBookmarkNotFound
	}

	if bookmark.Public != 1 && !c.UserIsLogged() {
		response.RedirectToLogin(c.Context, c.Request.URL.String())
		return nil, model.ErrUnauthorized
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
		response.NotFound(c)
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
		response.NotFound(c)
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
		response.NotFound(c)
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
		response.NotFound(c)
		return
	}

	etag := "w/" + model.GetThumbnailPath(bookmark) + "-" + bookmark.ModifiedAt

	// Check if the client's ETag matches the current ETag
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}

	options := &response.SendFileOptions{
		Headers: []http.Header{
			{"Cache-Control": {"no-cache , must-revalidate"}},
			{"Last-Modified": {bookmark.ModifiedAt}},
			{"ETag": {etag}},
		},
	}

	response.SendFile(c, r.deps.Domains.Storage, model.GetThumbnailPath(bookmark), options)
}

func (r *BookmarkRoutes) bookmarkEbookHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	bookmark, err := r.getBookmark(ctx)
	if err != nil {
		return
	}

	ebookPath := model.GetEbookPath(bookmark)

	if !r.deps.Domains.Storage.FileExists(ebookPath) {
		response.SendError(c, http.StatusNotFound, nil)
		return
	}

	// TODO: Potentially improve this
	c.Header("Content-Disposition", `attachment; filename="`+bookmark.Title+`.epub"`)
	response.SendFile(c, r.deps.Domains.Storage, model.GetEbookPath(bookmark), nil)
}
