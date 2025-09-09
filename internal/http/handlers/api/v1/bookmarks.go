package api_v1

import (
	"encoding/json"
	"errors"
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
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	bookmark, err := deps.Domains().Bookmarks().GetBookmark(c.Request().Context(), model.DBID(bookmarkID))
	if err != nil {
		response.SendError(c, http.StatusNotFound, "Bookmark not found")
		return
	}

	response.SendJSON(c, http.StatusOK, readableResponseMessage{
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
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	// Parse request payload
	var payload updateCachePayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get bookmarks from database
	bookmarks, err := deps.Domains().Bookmarks().GetBookmarks(c.Request().Context(), payload.Ids)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to get bookmarks")
		return
	}

	if len(bookmarks) == 0 {
		response.SendError(c, http.StatusNotFound, "No bookmarks found")
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

	response.SendJSON(c, http.StatusOK, bookmarks)
}

type bulkUpdateBookmarkTagsPayload struct {
	BookmarkIDs []int `json:"bookmark_ids" validate:"required"`
	TagIDs      []int `json:"tag_ids" validate:"required"`
}

func (p *bulkUpdateBookmarkTagsPayload) IsValid() error {
	if len(p.BookmarkIDs) == 0 {
		return fmt.Errorf("bookmark_ids should not be empty")
	}
	if len(p.TagIDs) == 0 {
		return fmt.Errorf("tag_ids should not be empty")
	}
	return nil
}

// HandleGetBookmarkTags gets the tags for a bookmark
//
//	@Summary					Get tags for a bookmark.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Produce					json
//	@Param						id	path		int	true	"Bookmark ID"
//	@Success					200	{array}		model.TagDTO
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					404	{object}	nil	"Bookmark not found"
//	@Router						/api/v1/bookmarks/{id}/tags [get]
func HandleGetBookmarkTags(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	// Check if bookmark exists
	exists, err := deps.Domains().Bookmarks().BookmarkExists(c.Request().Context(), bookmarkID)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to check if bookmark exists")
		return
	}
	if !exists {
		response.SendError(c, http.StatusNotFound, "Bookmark not found")
		return
	}

	// Get bookmark to retrieve its tags
	tags, err := deps.Domains().Tags().ListTags(c.Request().Context(), model.ListTagsOptions{
		BookmarkID: bookmarkID,
	})
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to get bookmark tags")
		return
	}

	response.SendJSON(c, http.StatusOK, tags)
}

// bookmarkTagPayload is used for both adding and removing tags from bookmarks
type bookmarkTagPayload struct {
	TagID int `json:"tag_id" validate:"required"`
}

func (p *bookmarkTagPayload) IsValid() error {
	if p.TagID <= 0 {
		return fmt.Errorf("tag_id should be a positive integer")
	}
	return nil
}

// HandleAddTagToBookmark adds a tag to a bookmark
//
//	@Summary					Add a tag to a bookmark.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						id		path	int					true	"Bookmark ID"
//	@Param						payload	body	bookmarkTagPayload	true	"Add Tag Payload"
//	@Produce					json
//	@Success					200	{object}	nil
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					404	{object}	nil	"Bookmark or tag not found"
//	@Router						/api/v1/bookmarks/{id}/tags [post]
func HandleAddTagToBookmark(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	// Parse request payload
	var payload bookmarkTagPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Add tag to bookmark
	err = deps.Domains().Bookmarks().AddTagToBookmark(c.Request().Context(), bookmarkID, payload.TagID)
	if err != nil {
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		if errors.Is(err, model.ErrTagNotFound) {
			response.SendError(c, http.StatusNotFound, "Tag not found")
			return
		}
		response.SendError(c, http.StatusInternalServerError, "Failed to add tag to bookmark")
		return
	}

	response.SendJSON(c, http.StatusCreated, nil)
}

// HandleRemoveTagFromBookmark removes a tag from a bookmark
//
//	@Summary					Remove a tag from a bookmark.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						id		path	int					true	"Bookmark ID"
//	@Param						payload	body	bookmarkTagPayload	true	"Remove Tag Payload"
//	@Produce					json
//	@Success					200	{object}	nil
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					404	{object}	nil	"Bookmark not found"
//	@Router						/api/v1/bookmarks/{id}/tags [delete]
func HandleRemoveTagFromBookmark(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	// Parse request payload
	var payload bookmarkTagPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Remove tag from bookmark
	err = deps.Domains().Bookmarks().RemoveTagFromBookmark(c.Request().Context(), bookmarkID, payload.TagID)
	if err != nil {
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		if errors.Is(err, model.ErrTagNotFound) {
			response.SendError(c, http.StatusNotFound, "Tag not found")
			return
		}
		response.SendError(c, http.StatusInternalServerError, "Failed to remove tag from bookmark")
		return
	}

	response.SendJSON(c, http.StatusOK, nil)
}

// HandleBulkUpdateBookmarkTags updates the tags for multiple bookmarks
//
//	@Summary					Bulk update tags for multiple bookmarks.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						payload	body	bulkUpdateBookmarkTagsPayload	true	"Bulk Update Bookmark Tags Payload"
//	@Produce					json
//	@Success					200	{object}	[]model.BookmarkDTO
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					400	{object}	nil	"Invalid request payload"
//	@Failure					404	{object}	nil	"No bookmarks found"
//	@Router						/api/v1/bookmarks/bulk/tags [put]
func HandleBulkUpdateBookmarkTags(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	// Parse request payload
	var payload bulkUpdateBookmarkTagsPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Use the domain method to update bookmark tags
	err := deps.Domains().Bookmarks().BulkUpdateBookmarkTags(c.Request().Context(), payload.BookmarkIDs, payload.TagIDs)
	if err != nil {
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "No bookmarks found")
			return
		}
		response.SendError(c, http.StatusInternalServerError, "Failed to update bookmarks")
		return
	}

	response.SendJSON(c, http.StatusOK, nil)
}
