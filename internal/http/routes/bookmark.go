package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type BookmarkRoutes struct {
	logger *logrus.Logger
	deps   *config.Dependencies
}

func (r *BookmarkRoutes) Setup(group *gin.RouterGroup) model.Routes {
	group.GET("/:id/archive", r.bookmarkArchiveHandler)
	group.GET("/:id/content", r.bookmarkContentHandler)
	return r
}

func (r *BookmarkRoutes) bookmarkContentHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	bookmarkIDParam, present := c.Params.Get("id")
	if !present {
		response.SendError(c, 400, "Invalid bookmark ID")
		return
	}

	bookmarkID, err := strconv.Atoi(bookmarkIDParam)
	if err != nil {
		r.logger.WithError(err).Error("error parsing bookmark ID parameter")
		response.SendInternalServerError(c)
		return
	}

	if bookmarkID == 0 {
		response.SendError(c, 404, nil)
		return
	}

	bookmark, found, err := r.deps.Database.GetBookmark(c, bookmarkID, "")
	if err != nil || !found {
		response.SendError(c, 404, nil)
		return
	}

	if bookmark.Public != 1 && !ctx.UserIsLogged() {
		response.SendError(c, http.StatusForbidden, nil)
		return
	}

	response.Send(c, 200, bookmark.Content)
}

func (r *BookmarkRoutes) bookmarkArchiveHandler(c *gin.Context) {}

func NewBookmarkRoutes(logger *logrus.Logger, deps *config.Dependencies) *BookmarkRoutes {
	return &BookmarkRoutes{
		logger: logger,
		deps:   deps,
	}
}
