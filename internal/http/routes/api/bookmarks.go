package api

import (
	"encoding/json"
	"log"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type BookmarksAPIRoutes struct {
	logger *logrus.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *BookmarksAPIRoutes) Setup() *BookmarksAPIRoutes {
	r.router.Get("/", r.listHandler)
	r.router.Post("/", r.createHandler)
	r.router.Delete("/:id", r.deleteHandler)
	return r
}

func (r *BookmarksAPIRoutes) Router() *fiber.App {
	return r.router
}

func (r *BookmarksAPIRoutes) listHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	bookmarks, err := r.deps.Database.GetBookmarks(ctx, database.GetBookmarksOptions{})
	if err != nil {
		return errors.Wrap(err, "error getting bookmakrs")
	}
	return response.Send(c, 200, bookmarks)
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

func (r *BookmarksAPIRoutes) createHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	payload := newAPICreateBookmarkPayload()
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		r.logger.WithError(err).Error("Error parsing payload")
		return response.SendError(c, 400, "Couldn't understand request")
	}

	bookmark, err := payload.ToBookmark()
	if err != nil {
		r.logger.WithError(err).Error("Error creating bookmark from request")
		return response.SendError(c, 400, "Couldn't understand request parameters")
	}

	results, err := r.deps.Database.SaveBookmarks(ctx, true, *bookmark)
	if err != nil || len(results) == 0 {
		r.logger.WithError(err).WithField("payload", payload).Error("Error creating bookmark")
		return response.SendInternalServerError(c)
	}

	book := results[0]

	if payload.Async {
		go func() {
			bookmark, err := r.deps.Domains.Archiver.DownloadBookmarkArchive(book)
			if err != nil {
				r.logger.WithError(err).Error("Error downloading bookmark")
				return
			}
			if _, err := r.deps.Database.SaveBookmarks(ctx, false, *bookmark); err != nil {
				r.logger.WithError(err).Error("Error saving bookmark")
			}
		}()
	} else {
		// Workaround. Download content after saving the bookmark so we have the proper database
		// id already set in the object regardless of the database engine.
		book, err := r.deps.Domains.Archiver.DownloadBookmarkArchive(book)
		if err != nil {
			r.logger.WithError(err).Error("Error downloading bookmark")
		} else if _, err := r.deps.Database.SaveBookmarks(ctx, false, *book); err != nil {
			r.logger.WithError(err).Error("Error saving bookmark")
		}
	}

	return response.Send(c, 201, book)
}

func (r *BookmarksAPIRoutes) deleteHandler(c *fiber.Ctx) error {
	ctx := c.Context()
	bookmarkID, err := c.ParamsInt("id")
	if err != nil {
		return response.SendError(c, 400, "Incorrect bookmark ID")
	}

	_, found, err := r.deps.Database.GetBookmark(ctx, bookmarkID, "")
	if err != nil {
		return response.SendError(c, 400, "Incorrect bookmark ID")
	}

	if !found {
		return response.SendError(c, 404, "Bookmark not found")
	}

	if err := r.deps.Database.DeleteBookmarks(ctx, bookmarkID); err != nil {
		r.logger.WithError(err).Error("Error deleting bookmark")
		return response.SendInternalServerError(c)
	}

	return response.Send(c, 200, "Bookmark deleted")
}

func NewBookmarksPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *BookmarksAPIRoutes {
	return &BookmarksAPIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
}
