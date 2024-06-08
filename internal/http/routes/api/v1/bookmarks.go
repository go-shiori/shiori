package api_v1

import (
	"fmt"
	"net/http"
	"os"
	fp "path/filepath"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type BookmarksAPIRoutes struct {
	logger *logrus.Logger
	deps   *dependencies.Dependencies
}

func (r *BookmarksAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.Use(middleware.AuthenticationRequired())
	g.PUT("/cache", r.updateCache)
	g.GET("/:id/readable", r.bookmarkReadable)
	return r
}

func NewBookmarksAPIRoutes(logger *logrus.Logger, deps *dependencies.Dependencies) *BookmarksAPIRoutes {
	return &BookmarksAPIRoutes{
		logger: logger,
		deps:   deps,
	}
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

func (r *BookmarksAPIRoutes) getBookmark(c *context.Context) (*model.BookmarkDTO, error) {
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

	return bookmark, nil
}

type readableResponseMessage struct {
	Content string `json:"content"`
	Html    string `json:"html"`
}

// Bookmark Readable godoc
//
//	@Summary					Get readable version of bookmark.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Produce					json
//	@Success					200	{object}	readableResponseMessage
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/bookmarks/id/readable [get]
func (r *BookmarksAPIRoutes) bookmarkReadable(c *gin.Context) {
	ctx := context.NewContextFromGin(c)

	bookmark, err := r.getBookmark(ctx)
	if err != nil {
		return
	}

	response.Send(c, 200, readableResponseMessage{
		Content: bookmark.Content,
		Html:    bookmark.HTML,
	})
}

// updateCache godoc
//
//	@Summary					Update Cache and Ebook on server.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						payload	body	updateCachePayload	true	"Update Cache Payload"`
//	@Produce					json
//	@Success					200	{object}	model.BookmarkDTO
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/bookmarks/cache [put]
func (r *BookmarksAPIRoutes) updateCache(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if !ctx.GetAccount().Owner {
		response.SendError(c, http.StatusForbidden, nil)
		return
	}

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

		go func(i int, book model.BookmarkDTO, keep_metadata bool) {
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
				DataDir:     r.deps.Config.Storage.DataDir,
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

			book, _, err = core.ProcessBookmark(r.deps, request)
			content.Close()

			if err != nil {
				r.logger.WithFields(logrus.Fields{
					"bookmark_id": book.ID,
					"url":         book.URL,
					"error":       err,
				}).Error("error downloading bookmark cache")
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
	_, err = r.deps.Database.SaveBookmarks(c, false, bookmarks...)
	if err != nil {
		r.logger.WithError(err).Error("error update bookmakrs on deatabas")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, 200, bookmarks)
}
