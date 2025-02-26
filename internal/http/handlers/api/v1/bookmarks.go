package api_v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

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

type readableResponseMessage struct {
	Content string `json:"content"`
	HTML    string `json:"html"`
}

// HandleBookmarkReadable returns the readable version of a bookmark
//
//	@Summary					Get readable version of bookmark.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Produce					json
//	@Success					200	{object}	readableResponseMessage
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/bookmarks/id/readable [get]
func HandleBookmarkReadable(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error(), nil)
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID", nil)
		return
	}

	bookmark, err := deps.Domains().Bookmarks().GetBookmark(c.Request().Context(), model.DBID(bookmarkID))
	if err != nil {
		response.SendError(c, http.StatusNotFound, "Bookmark not found", nil)
		return
	}

	response.Send(c, http.StatusOK, readableResponseMessage{
		Content: bookmark.Content,
		HTML:    bookmark.HTML,
	})
}

// HandleUpdateCache updates the cache and ebook for bookmarks
//
//	@Summary					Update Cache and Ebook on server.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						payload	body	updateCachePayload	true	"Update Cache Payload"
//	@Produce					json
//	@Success					200	{object}	model.BookmarkDTO
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/bookmarks/cache [put]
func HandleUpdateCache(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error(), nil)
		return
	}

	// Parse request payload
	var payload updateCachePayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Get bookmarks from database
	bookmarks, err := deps.Domains().Bookmarks().GetBookmarks(c.Request().Context(), payload.Ids)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to get bookmarks", nil)
		return
	}

	if len(bookmarks) == 0 {
		response.SendError(c, http.StatusNotFound, "No bookmarks found", nil)
		return
	}

	// Process bookmarks concurrently
	mx := sync.RWMutex{}
	wg := sync.WaitGroup{}
	chDone := make(chan struct{})
	chProblem := make(chan int, 10)
	semaphore := make(chan struct{}, 10)

	for i, book := range bookmarks {
		wg.Add(1)

		book.CreateArchive = payload.CreateArchive
		book.CreateEbook = payload.CreateEbook

		go func(i int, book model.BookmarkDTO) {
			defer wg.Done()
			defer func() { <-semaphore }()
			semaphore <- struct{}{}

			// Download and process bookmark
			updatedBook, err := deps.Domains().Bookmarks().UpdateBookmarkCache(c.Request().Context(), book, payload.KeepMetadata, payload.SkipExist)
			if err != nil {
				deps.Logger().WithError(err).Error("error updating bookmark cache")
				chProblem <- book.ID
				return
			}

			mx.Lock()
			bookmarks[i] = *updatedBook
			mx.Unlock()
		}(i, book)
	}

	// Collect problematic bookmarks
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

	wg.Wait()
	close(chDone)

	response.Send(c, http.StatusOK, bookmarks)
}
