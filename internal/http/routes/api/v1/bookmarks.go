package api_v1

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type BookmarksAPIRoutes struct {
	logger *logrus.Logger
	deps   *config.Dependencies
}

func (r *BookmarksAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.GET("/", r.listHandler)
	g.POST("/", r.createHandler)
	g.DELETE("/:id", r.deleteHandler)
	return r
}

func (r *BookmarksAPIRoutes) listHandler(c *gin.Context) {
	bookmarks, err := r.deps.Database.GetBookmarks(c, database.GetBookmarksOptions{})
	if err != nil {
		r.logger.WithError(err).Error("error getting bookmakrs")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, 200, bookmarks)
}

type apiCreateBookmarkPayload struct {
	URL           string      `json:"url"`
	Title         string      `json:"title"`
	Excerpt       string      `json:"excerpt"`
	Tags          []model.Tag `json:"tags"`
	CreateArchive bool        `json:"createArchive"`
	MakePublic    int         `json:"public"`
	Async         bool        `json:"async"`
}

func (payload *apiCreateBookmarkPayload) ToBookmark() (*model.Bookmark, error) {
	bookmark := &model.Bookmark{
		URL:           payload.URL,
		Title:         payload.Title,
		Excerpt:       payload.Excerpt,
		Tags:          payload.Tags,
		Public:        payload.MakePublic,
		CreateArchive: payload.CreateArchive,
	}

	log.Println(bookmark.URL)

	var err error
	bookmark.URL, err = core.RemoveUTMParams(bookmark.URL)
	if err != nil {
		return nil, err
	}

	// Ensure title is not empty
	if bookmark.Title == "" {
		bookmark.Title = bookmark.URL
	}

	return bookmark, nil
}

func newAPICreateBookmarkPayload() *apiCreateBookmarkPayload {
	return &apiCreateBookmarkPayload{
		CreateArchive: false,
		Async:         true,
	}
}

func (r *BookmarksAPIRoutes) createHandler(c *gin.Context) {
	payload := newAPICreateBookmarkPayload()
	if err := c.ShouldBindJSON(&payload); err != nil {
		r.logger.WithError(err).Error("Error parsing payload")
		response.SendError(c, 400, "Couldn't understand request")
		return
	}

	bookmark, err := payload.ToBookmark()
	if err != nil {
		r.logger.WithError(err).Error("Error creating bookmark from request")
		response.SendError(c, 400, "Couldn't understand request parameters")
		return
	}

	results, err := r.deps.Database.SaveBookmarks(c, true, *bookmark)
	if err != nil || len(results) == 0 {
		r.logger.WithError(err).WithField("payload", payload).Error("Error creating bookmark")
		response.SendInternalServerError(c)
		return
	}

	book := results[0]

	if payload.Async {
		go func() {
			bookmark, err := r.deps.Domains.Archiver.DownloadBookmarkArchive(book)
			if err != nil {
				r.logger.WithError(err).Error("Error downloading bookmark")
				return
			}
			if _, err := r.deps.Database.SaveBookmarks(c, false, *bookmark); err != nil {
				r.logger.WithError(err).Error("Error saving bookmark")
			}
		}()
	} else {
		// Workaround. Download content after saving the bookmark so we have the proper database
		// id already set in the object regardless of the database engine.
		book, err := r.deps.Domains.Archiver.DownloadBookmarkArchive(book)
		if err != nil {
			r.logger.WithError(err).Error("Error downloading bookmark")
		} else if _, err := r.deps.Database.SaveBookmarks(c, false, *book); err != nil {
			r.logger.WithError(err).Error("Error saving bookmark")
		}
	}

	response.Send(c, 201, book)
}

func (r *BookmarksAPIRoutes) deleteHandler(c *gin.Context) {
	bookmarkIDParam, exists := c.Params.Get("id")
	if !exists {
		response.SendError(c, 400, "Incorrect bookmark ID")
		return
	}

	bookmarkID, err := strconv.Atoi(bookmarkIDParam)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	_, found, err := r.deps.Database.GetBookmark(c, bookmarkID, "")
	if err != nil {
		response.SendError(c, 400, "Incorrect bookmark ID")
		return
	}

	if !found {
		response.SendError(c, 404, "Bookmark not found")
		return
	}

	if err := r.deps.Database.DeleteBookmarks(c, bookmarkID); err != nil {
		r.logger.WithError(err).Error("Error deleting bookmark")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, 200, "Bookmark deleted")
}

func NewBookmarksPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *BookmarksAPIRoutes {
	return &BookmarksAPIRoutes{
		logger: logger,
		deps:   deps,
	}
}
