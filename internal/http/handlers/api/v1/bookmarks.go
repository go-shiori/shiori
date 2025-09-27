package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
//	@Tags						Bookmarks
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
//	@Tags						Bookmarks
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
				chProblem <- book.Bookmark.ID
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
//	@Tags						Bookmarks
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
//	@Tags						Bookmarks
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
//	@Tags						Bookmarks
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

// Bookmark CRUD operations

type createBookmarkPayload struct {
	URL         string   `json:"url" validate:"required"`
	Title       string   `json:"title"`
	Excerpt     string   `json:"excerpt"`
	Tags        []string `json:"tags"`
	CreateEbook bool     `json:"create_ebook"`
	Public      int      `json:"public"`
}

func (p *createBookmarkPayload) IsValid() error {
	if strings.TrimSpace(p.URL) == "" {
		return fmt.Errorf("url should not be empty")
	}
	return nil
}

// HandleCreateBookmark creates a new bookmark
//
//	@Summary			Create a new bookmark.
//	@Tags				Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param				payload	body		createBookmarkPayload	true	"Create Bookmark Payload"
//	@Produce			json
//	@Success			201	{object}	model.BookmarkDTO
//	@Failure			403	{object}	nil	"Token not provided/invalid"
//	@Failure			400	{object}	nil	"Invalid request payload"
//	@Router				/api/v1/bookmarks [post]
func HandleCreateBookmark(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	// Parse request payload
	var payload createBookmarkPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create bookmark for domain operations
	bookmark := model.Bookmark{
		URL:     payload.URL,
		Title:   payload.Title,
		Excerpt: payload.Excerpt,
		Public:  payload.Public,
	}

	// Create the bookmark
	createdBookmark, err := deps.Domains().Bookmarks().CreateBookmark(c.Request().Context(), bookmark)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to create bookmark")
		return
	}

	// Handle tags if provided
	if len(payload.Tags) > 0 {
		var addedTags []model.TagDTO
		for _, tagName := range payload.Tags {
			// Create or get tag
			tag, err := deps.Database().CreateTag(c.Request().Context(), model.Tag{Name: tagName})
			if err != nil {
				// Try to find existing tag if creation failed
				existingTags, getErr := deps.Database().GetTags(c.Request().Context(), model.DBListTagsOptions{
					Search: tagName,
				})
				if getErr != nil || len(existingTags) == 0 || existingTags[0].Name != tagName {
					response.SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create or find tag %s", tagName))
					return
				}
				tag = model.Tag{ID: existingTags[0].ID, Name: existingTags[0].Name}
			}

			// Add tag to bookmark
			err = deps.Domains().Bookmarks().AddTagToBookmark(c.Request().Context(), createdBookmark.ID, tag.ID)
			if err != nil {
				response.SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to add tag %s to bookmark", tagName))
				return
			}

			// Add to response tags
			addedTags = append(addedTags, model.TagDTO{
				Tag: tag,
			})
		}

		// Add tags to response
		createdBookmark.Tags = addedTags
	}

	response.SendJSON(c, http.StatusCreated, createdBookmark)
}

// HandleListBookmarks lists bookmarks with optional filtering
//
//	@Summary			List bookmarks with optional filtering.
//	@Tags				Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param				keyword	query		string	false	"Search keyword"
//	@Param				tags	query		string	false	"Comma-separated list of tags to include"
//	@Param				exclude	query		string	false	"Comma-separated list of tags to exclude"
//	@Param				page	query		int		false	"Page number (default: 1)"
//	@Param				limit	query		int		false	"Items per page (default: 30, max: 100)"
//	@Produce			json
//	@Success			200	{array}		model.BookmarkDTO
//	@Failure			403	{object}	nil	"Token not provided/invalid"
//	@Router				/api/v1/bookmarks [get]
func HandleListBookmarks(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	// Parse query parameters
	query := c.Request().URL.Query()
	keyword := query.Get("keyword")
	tagsStr := query.Get("tags")
	excludedTagsStr := query.Get("exclude")
	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	var excludedTags []string
	if excludedTagsStr != "" {
		excludedTags = strings.Split(excludedTagsStr, ",")
	}

	// Parse page and limit
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Prepare search options
	searchOptions := model.DBGetBookmarksOptions{
		Tags:         tags,
		ExcludedTags: excludedTags,
		Keyword:      keyword,
		Limit:        limit,
		Offset:       (page - 1) * limit,
		OrderMethod:  model.ByLastAdded,
	}

	// Get bookmarks
	bookmarks, err := deps.Database().GetBookmarks(c.Request().Context(), searchOptions)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to get bookmarks")
		return
	}

	response.SendJSON(c, http.StatusOK, bookmarks)
}

// HandleGetBookmark gets a single bookmark by ID
//
//	@Summary			Get a bookmark by ID.
//	@Tags				Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param				id	path		int	true	"Bookmark ID"
//	@Produce			json
//	@Success			200	{object}	model.BookmarkDTO
//	@Failure			403	{object}	nil	"Token not provided/invalid"
//	@Failure			404	{object}	nil	"Bookmark not found"
//	@Router				/api/v1/bookmarks/{id} [get]
func HandleGetBookmark(deps model.Dependencies, c model.WebContext) {
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

	response.SendJSON(c, http.StatusOK, bookmark)
}

type updateBookmarkPayload struct {
	URL         *string  `json:"url"`
	Title       *string  `json:"title"`
	Excerpt     *string  `json:"excerpt"`
	Tags        []string `json:"tags"`
	CreateEbook *bool    `json:"create_ebook"`
	Public      *int     `json:"public"`
}

// HandleUpdateBookmark updates an existing bookmark
//
//	@Summary			Update an existing bookmark.
//	@Tags				Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param				id		path		int						true	"Bookmark ID"
//	@Param				payload	body		updateBookmarkPayload	true	"Update Bookmark Payload"
//	@Produce			json
//	@Success			200	{object}	model.BookmarkDTO
//	@Failure			403	{object}	nil	"Token not provided/invalid"
//	@Failure			404	{object}	nil	"Bookmark not found"
//	@Failure			400	{object}	nil	"Invalid request payload"
//	@Router				/api/v1/bookmarks/{id} [put]
func HandleUpdateBookmark(deps model.Dependencies, c model.WebContext) {
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
	var payload updateBookmarkPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get existing bookmark
	existingBookmark, err := deps.Domains().Bookmarks().GetBookmark(c.Request().Context(), model.DBID(bookmarkID))
	if err != nil {
		response.SendError(c, http.StatusNotFound, "Bookmark not found")
		return
	}

	// Convert to Bookmark for domain operations
	bookmark := existingBookmark.ToBookmark()

	// Update fields if provided
	if payload.URL != nil {
		bookmark.URL = *payload.URL
	}
	if payload.Title != nil {
		bookmark.Title = *payload.Title
	}
	if payload.Excerpt != nil {
		bookmark.Excerpt = *payload.Excerpt
	}
	if payload.Public != nil {
		bookmark.Public = *payload.Public
	}

	// Update the bookmark
	updatedBookmark, err := deps.Domains().Bookmarks().UpdateBookmark(c.Request().Context(), bookmark)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to update bookmark")
		return
	}

	// Handle tags if provided
	if payload.Tags != nil {
		// Clear existing tags if empty array provided
		if len(payload.Tags) == 0 {
			for _, tag := range existingBookmark.Tags {
				err = deps.Domains().Bookmarks().RemoveTagFromBookmark(c.Request().Context(), bookmarkID, tag.ID)
				if err != nil {
					response.SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to remove tag %s", tag.Name))
					return
				}
			}
		} else {
			// Clear existing tags first
			for _, tag := range existingBookmark.Tags {
				err = deps.Domains().Bookmarks().RemoveTagFromBookmark(c.Request().Context(), bookmarkID, tag.ID)
				if err != nil {
					response.SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to remove existing tag %s", tag.Name))
					return
				}
			}

			// Add new tags
			for _, tagName := range payload.Tags {
				// Create or get tag
				tag, err := deps.Database().CreateTag(c.Request().Context(), model.Tag{Name: tagName})
				if err != nil {
					// Try to find existing tag if creation failed
					existingTags, getErr := deps.Database().GetTags(c.Request().Context(), model.DBListTagsOptions{
						Search: tagName,
					})
					if getErr != nil || len(existingTags) == 0 || existingTags[0].Name != tagName {
						response.SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create or find tag %s", tagName))
						return
					}
					tag = model.Tag{ID: existingTags[0].ID, Name: existingTags[0].Name}
				}

				// Add tag to bookmark
				err = deps.Domains().Bookmarks().AddTagToBookmark(c.Request().Context(), bookmarkID, tag.ID)
				if err != nil {
					response.SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to add tag %s to bookmark", tagName))
					return
				}
			}
		}

		// Add tags to response if any were added
		if len(payload.Tags) > 0 {
			var addedTags []model.TagDTO
			for _, tagName := range payload.Tags {
				// Find the tag (it should exist since we just added it)
				existingTags, err := deps.Database().GetTags(c.Request().Context(), model.DBListTagsOptions{
					Search: tagName,
				})
				if err == nil && len(existingTags) > 0 && existingTags[0].Name == tagName {
					addedTags = append(addedTags, model.TagDTO{
						Tag: model.Tag{
							ID:   existingTags[0].ID,
							Name: existingTags[0].Name,
						},
					})
				}
			}
			updatedBookmark.Tags = addedTags
		} else {
			// Clear tags in response
			updatedBookmark.Tags = []model.TagDTO{}
		}
	}

	response.SendJSON(c, http.StatusOK, updatedBookmark)
}

type deleteBookmarksPayload struct {
	IDs []int `json:"ids" validate:"required"`
}

func (p *deleteBookmarksPayload) IsValid() error {
	if len(p.IDs) == 0 {
		return fmt.Errorf("ids should not be empty")
	}
	for _, id := range p.IDs {
		if id <= 0 {
			return fmt.Errorf("id should not be 0 or negative")
		}
	}
	return nil
}

// HandleDeleteBookmarks deletes one or more bookmarks
//
//	@Summary			Delete one or more bookmarks.
//	@Tags				Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param				payload	body		deleteBookmarksPayload	true	"Delete Bookmarks Payload"
//	@Produce			json
//	@Success			200	{object}	nil
//	@Failure			403	{object}	nil	"Token not provided/invalid"
//	@Failure			400	{object}	nil	"Invalid request payload"
//	@Router				/api/v1/bookmarks [delete]
func HandleDeleteBookmarks(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	// Parse request payload
	var payload deleteBookmarksPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Delete bookmarks
	err := deps.Domains().Bookmarks().DeleteBookmarks(c.Request().Context(), payload.IDs)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to delete bookmarks")
		return
	}

	response.SendJSON(c, http.StatusOK, nil)
}

// HandleBulkUpdateBookmarkTags updates the tags for multiple bookmarks
//
//	@Summary					Bulk update tags for multiple bookmarks.
//	@Tags						Bookmarks
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
