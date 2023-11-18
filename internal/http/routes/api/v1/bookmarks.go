package api_v1

import (
	"fmt"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/http/context"
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
	g.PUT("/cache", r.updateCache)
	g.POST("/", r.createHandler)
	g.DELETE("/:id", r.deleteHandler)
	return r
}

type updateCachePayload struct {
	Ids           []int `json:"ids"    validate:"required"`
	KeepMetadata  bool  `json:"keep_metadata"`
	CreateArchive bool  `json:"create_archive"`
	CreateEbook   bool  `json:"create_ebook"`
	SkipExist     bool  `json:"skip_exist"`
}

func (p *updateCachePayload) IsValid() error {
	if len(p.Ids) == 0 {
		return fmt.Errorf("id should not be empty")
	}
	for _, id := range p.Ids {
		if id <= 0 {
			return fmt.Errorf("id should not be 0 or negative")
		}
	}
	return nil
}

func (r *BookmarksAPIRoutes) listHandler(c *gin.Context) {
	bookmarks, err := r.deps.Database.GetBookmarks(c, database.GetBookmarksOptions{})
	if err != nil {
		r.logger.WithError(err).Error("error getting bookmarks")
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
	CreateArchive bool        `json:"create_archive"`
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

// updateCache godoc
//
//	@Summary					Update Cache and Ebook on server.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						payload	body	updateCachePayload	true "Update Cache Payload"`
//	@Produce					json
//	@Success					200	{object}	model.Bookmark
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/bookmarks/cache [put]
func (r *BookmarksAPIRoutes) updateCache(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if !ctx.UserIsLogged() {
		response.SendError(c, http.StatusForbidden, nil)
		return
	}

	// Get server config
	logger := logrus.New()
	cfg := config.ParseServerConfiguration(ctx, logger)

	var payload updateCachePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.SendInternalServerError(c)
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// send request to database and get bookmarks
	filter := database.GetBookmarksOptions{
		IDs:         payload.Ids,
		WithContent: true,
	}

	bookmarks, err := r.deps.Database.GetBookmarks(c, filter)
	if len(bookmarks) == 0 {
		r.logger.WithError(err).Error("Bookmark not found")
		response.SendError(c, 404, "Bookmark not found")
		return
	}

	if err != nil {
		r.logger.WithError(err).Error("error getting bookmakrs")
		response.SendInternalServerError(c)
		return
	}
	// TODO: limit request to 20

	// Fetch data from internet
	mx := sync.RWMutex{}
	wg := sync.WaitGroup{}
	chDone := make(chan struct{})
	chProblem := make(chan int, 10)
	semaphore := make(chan struct{}, 10)

	for i, book := range bookmarks {
		wg.Add(1)

		// Mark whether book will be archived or ebook generate request
		book.CreateArchive = payload.CreateArchive
		book.CreateEbook = payload.CreateEbook

		go func(i int, book model.Bookmark, keep_metadata bool) {
			// Make sure to finish the WG
			defer wg.Done()

			// Register goroutine to semaphore
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			// Download data from internet
			content, contentType, err := core.DownloadBookmark(book.URL)
			if err != nil {
				chProblem <- book.ID
				return
			}

			request := core.ProcessRequest{
				DataDir:     cfg.Storage.DataDir,
				Bookmark:    book,
				Content:     content,
				ContentType: contentType,
				KeepTitle:   keep_metadata,
				KeepExcerpt: keep_metadata,
			}

			if payload.SkipExist && book.CreateEbook {
				strID := strconv.Itoa(book.ID)
				ebookPath := fp.Join(request.DataDir, "ebook", strID+".epub")
				_, err = os.Stat(ebookPath)
				if err == nil {
					request.Bookmark.CreateEbook = false
					request.Bookmark.HasEbook = true
				}
			}

			book, _, err = core.ProcessBookmark(request)
			content.Close()

			if err != nil {
				chProblem <- book.ID
				return
			}

			// Update list of bookmarks
			mx.Lock()
			bookmarks[i] = book
			mx.Unlock()
		}(i, book, payload.KeepMetadata)
	}
	// Receive all problematic bookmarks
	idWithProblems := []int{}
	go func() {
		for {
			select {
			case <-chDone:
				return
			case id := <-chProblem:
				idWithProblems = append(idWithProblems, id)
			}
		}
	}()

	// Wait until all download finished
	wg.Wait()
	close(chDone)

	// Update database
	_, err = r.deps.Database.SaveBookmarks(ctx, false, bookmarks...)
	if err != nil {
		r.logger.WithError(err).Error("error update bookmakrs on deatabas")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, 200, bookmarks)
}
