package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

type updateCachePayload struct {
	IDs           []int `validate:"required"`
	KeepMetadata  bool
	CreateArchive bool
	CreateEbook   bool
	SkipExist     bool
}

func (p *updateCachePayload) IsValid() error {
	if len(p.IDs) == 0 {
		return fmt.Errorf("id should not be empty")
	}
	for _, id := range p.IDs {
		if id <= 0 {
			return fmt.Errorf("id should not be 0 or negative")
		}
	}
	return nil
}

type readableResponseMessage struct {
	Content string
	HTML    string
}

type updateBookmarkDataPayload struct {
	UpdateReadable bool
	CreateArchive  bool
	CreateEbook    bool
	KeepMetadata   bool
	SkipExisting   bool
}

type bookmarkDataResponse struct {
	Content      string
	HTML         string
	ImageURL     string
	ArchiveURL   string
	EbookURL     string
	HasContent   bool
	HasArchive   bool
	HasEbook     bool
	HasThumbnail bool
}

// HandleBookmarkReadable returns the readable version of a bookmark
//
//	@Summary					Get readable version of bookmark.
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						id	path	int	true	"Bookmark ID"
//	@Produce					json
//	@Success					200	{object}	readableResponseMessage
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/bookmarks/{id}/readable [get]
func HandleBookmarkReadable(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	bookmark, err := deps.Domains().Bookmarks().GetBookmark(c.Request().Context(), model.DBID(bookmarkID))
	if err != nil {
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		deps.Logger().WithError(err).Error("failed to get bookmark")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, readableResponseMessage{
		Content: bookmark.Content,
		HTML:    bookmark.HTML,
	})
}

// HandleGetBookmarkData returns the data and resource URLs for a bookmark
//
//	@Summary					Get bookmark data and resource URLs.
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						id	path	int	true	"Bookmark ID"
//	@Produce					json
//	@Success					200	{object}	bookmarkDataResponse
//	@Failure					400	{object}	nil	"Invalid bookmark ID"
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					404	{object}	nil	"Bookmark not found"
//	@Router						/api/v1/bookmarks/{id}/data [get]
func HandleGetBookmarkData(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	bookmark, err := deps.Domains().Bookmarks().GetBookmark(c.Request().Context(), model.DBID(bookmarkID))
	if err != nil {
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		deps.Logger().WithError(err).Error("failed to get bookmark")
		response.SendInternalServerError(c)
		return
	}

	// Build resource URLs
	archiveURL := ""
	if bookmark.HasArchive {
		archiveURL = model.GetArchiveURL(bookmark)
	}

	ebookURL := ""
	if bookmark.HasEbook {
		ebookURL = model.GetEbookURL(bookmark)
	}

	imageURL := ""
	if bookmark.HasThumbnail {
		imageURL = model.GetThumbnailURL(bookmark)
	}

	response.SendJSON(c, http.StatusOK, bookmarkDataResponse{
		Content:      bookmark.Content,
		HTML:         bookmark.HTML,
		ImageURL:     imageURL,
		ArchiveURL:   archiveURL,
		EbookURL:     ebookURL,
		HasContent:   bookmark.HasContent,
		HasArchive:   bookmark.HasArchive,
		HasEbook:     bookmark.HasEbook,
		HasThumbnail: bookmark.HasThumbnail,
	})
}

// HandleUpdateBookmarkData updates the data for a bookmark (readable content, archive, ebook)
//
//	@Summary					Update bookmark data (content, archive, ebook).
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						id		path	int							true	"Bookmark ID"
//	@Param						payload	body	updateBookmarkDataPayload	true	"Update Bookmark Data Payload"
//	@Produce					json
//	@Success					200	{object}	bookmarkDataResponse
//	@Failure					400	{object}	nil	"Invalid bookmark ID or payload"
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					404	{object}	nil	"Bookmark not found"
//	@Failure					500	{object}	nil	"Failed to update bookmark data"
//	@Router						/api/v1/bookmarks/{id}/data [put]
func HandleUpdateBookmarkData(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	// Parse request payload
	var payload updateBookmarkDataPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get bookmark from database
	bookmark, err := deps.Domains().Bookmarks().GetBookmark(c.Request().Context(), model.DBID(bookmarkID))
	if err != nil {
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		deps.Logger().WithError(err).Error("failed to get bookmark")
		response.SendInternalServerError(c)
		return
	}

	// Only update if at least one of the flags is true
	if payload.UpdateReadable || payload.CreateArchive || payload.CreateEbook {
		// Set flags on bookmark
		bookmark.CreateArchive = payload.CreateArchive
		bookmark.CreateEbook = payload.CreateEbook

		// Update bookmark cache
		updatedBookmark, err := deps.Domains().Bookmarks().UpdateBookmarkCache(
			c.Request().Context(),
			*bookmark,
			payload.KeepMetadata,
			payload.SkipExisting,
		)
		if err != nil {
			deps.Logger().WithError(err).Error("failed to update bookmark data")
			response.SendError(c, http.StatusInternalServerError, "Failed to update bookmark data")
			return
		}

		// Save updated bookmark to database
		_, err = deps.Domains().Bookmarks().UpdateBookmark(c.Request().Context(), updatedBookmark.Bookmark)
		if err != nil {
			deps.Logger().WithError(err).Error("failed to save updated bookmark")
			response.SendError(c, http.StatusInternalServerError, "Failed to save bookmark")
			return
		}

		bookmark = updatedBookmark
	}

	// Build resource URLs
	archiveURL := ""
	if bookmark.HasArchive {
		archiveURL = model.GetArchiveURL(bookmark)
	}

	ebookURL := ""
	if bookmark.HasEbook {
		ebookURL = model.GetEbookURL(bookmark)
	}

	imageURL := ""
	if bookmark.HasThumbnail {
		imageURL = model.GetThumbnailURL(bookmark)
	}

	response.SendJSON(c, http.StatusOK, bookmarkDataResponse{
		Content:      bookmark.Content,
		HTML:         bookmark.HTML,
		ImageURL:     imageURL,
		ArchiveURL:   archiveURL,
		EbookURL:     ebookURL,
		HasContent:   bookmark.HasContent,
		HasArchive:   bookmark.HasArchive,
		HasEbook:     bookmark.HasEbook,
		HasThumbnail: bookmark.HasThumbnail,
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
	bookmarks, err := deps.Domains().Bookmarks().GetBookmarks(c.Request().Context(), payload.IDs)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to get bookmarks")
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
	semaphore := make(chan struct{}, 10)

	for i, book := range bookmarks {
		wg.Add(1)

		book.CreateArchive = payload.CreateArchive
		book.CreateEbook = payload.CreateEbook

		go func(i int, book model.BookmarkDTO) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Download and process bookmark
			updatedBook, err := deps.Domains().Bookmarks().UpdateBookmarkCache(c.Request().Context(), book, payload.KeepMetadata, payload.SkipExist)
			if err != nil {
				deps.Logger().WithError(err).Error("error updating bookmark cache")
				return
			}

			mx.Lock()
			bookmarks[i] = *updatedBook
			mx.Unlock()
		}(i, book)
	}

	wg.Wait()

	response.SendJSON(c, http.StatusOK, bookmarks)
}

type bulkUpdateBookmarkTagsPayload struct {
	BookmarkIDs []int `validate:"required"`
	TagIDs      []int `validate:"required"`
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
		deps.Logger().WithError(err).Error("failed to check bookmark existence")
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
		deps.Logger().WithError(err).Error("failed to get bookmark tags")
		response.SendError(c, http.StatusInternalServerError, "Failed to get bookmark tags")
		return
	}

	response.SendJSON(c, http.StatusOK, tags)
}

// bookmarkTagPayload is used for both adding and removing tags from bookmarks
type bookmarkTagPayload struct {
	TagID int `validate:"required"`
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
		deps.Logger().WithError(err).Error("failed to add tag to bookmark")
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
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
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
		deps.Logger().WithError(err).Error("failed to remove tag from bookmark")
		response.SendError(c, http.StatusInternalServerError, "Failed to remove tag from bookmark")
		return
	}

	response.SendJSON(c, http.StatusOK, nil)
}

// Bookmark CRUD operations

type createBookmarkOptions struct {
	CreateArchive bool
	CreateEbook   bool
}

type createBookmarkPayload struct {
	URL     string                 `validate:"required"`
	Title   string                 `json:",omitempty"`
	Excerpt string                 `json:",omitempty"`
	Public  *int                   `json:",omitempty"`
	Options *createBookmarkOptions `json:",omitempty"`
}

// IsQuickAdd returns true if only the URL was provided (quick add scenario)
func (p *createBookmarkPayload) IsQuickAdd() bool {
	hasTitle := strings.TrimSpace(p.Title) != ""
	hasExcerpt := strings.TrimSpace(p.Excerpt) != ""
	hasPublic := p.Public != nil
	hasOptions := p.Options != nil && (p.Options.CreateArchive || p.Options.CreateEbook)

	// Quick add if only URL is provided (no title, excerpt, public flag, or options)
	return !hasTitle && !hasExcerpt && !hasPublic && !hasOptions
}

func (p *createBookmarkPayload) IsValid() error {
	if strings.TrimSpace(p.URL) == "" {
		return fmt.Errorf("url should not be empty")
	}
	return nil
}

// HandleCreateBookmark creates a new bookmark
//
//	@Summary					Create a new bookmark.
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						payload	body	createBookmarkPayload	true	"Create Bookmark Payload"
//	@Produce					json
//	@Success					201	{object}	model.BookmarkDTO
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					400	{object}	nil	"Invalid request payload"
//	@Router						/api/v1/bookmarks [post]
func HandleCreateBookmark(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
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

	cleanURL, err := core.RemoveUTMParams(strings.TrimSpace(payload.URL))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid URL")
		return
	}
	payload.URL = cleanURL

	// Create bookmark for domain operations
	bookmark := model.Bookmark{
		URL:     payload.URL,
		Title:   payload.Title,
		Excerpt: payload.Excerpt,
	}

	// Set Public field if provided, otherwise default to 0 (internal)
	if payload.Public != nil {
		bookmark.Public = *payload.Public
	} else {
		bookmark.Public = 0
	}

	// If title is not provided, default to URL
	if strings.TrimSpace(bookmark.Title) == "" {
		bookmark.Title = bookmark.URL
	}

	// Create the bookmark
	createdBookmark, err := deps.Domains().Bookmarks().CreateBookmark(c.Request().Context(), bookmark)
	if err != nil {
		deps.Logger().Error("error creating bookmark: ", err)

		// Check for specific errors
		if errors.Is(err, model.ErrBookmarkAlreadyExists) {
			response.SendError(c, http.StatusConflict, "already_exists.bookmark")
			return
		}

		response.SendError(c, http.StatusInternalServerError, "Failed to create bookmark")
		return
	}

	// Process bookmark to create readable content and optionally archive/ebook
	// This matches the behavior of the legacy API
	userHasDefinedTitle := strings.TrimSpace(payload.Title) != ""
	userHasDefinedExcerpt := strings.TrimSpace(payload.Excerpt) != ""

	// Set flags for archive/ebook creation from options
	// Default behavior: always create archive unless explicitly disabled
	if payload.Options != nil {
		// Use explicit options if provided
		createdBookmark.CreateArchive = payload.Options.CreateArchive
		createdBookmark.CreateEbook = payload.Options.CreateEbook
	} else {
		// No options provided - default to creating archive (matches legacy API behavior)
		createdBookmark.CreateArchive = true
		createdBookmark.CreateEbook = false
	}

	// Update bookmark cache to process content
	// Keep metadata if user provided title or excerpt (UpdateBookmarkCache uses same flag for both)
	// For quick add, always extract metadata from the page
	keepMetadata := userHasDefinedTitle || userHasDefinedExcerpt
	processedBookmark, err := deps.Domains().Bookmarks().UpdateBookmarkCache(
		c.Request().Context(),
		*createdBookmark,
		keepMetadata,
		false, // Don't skip existing
	)
	if err != nil {
		// Log error but don't fail the request - bookmark was created successfully
		deps.Logger().WithError(err).Warn("failed to process bookmark content, bookmark was still created")
		// Return the created bookmark without processed content
		response.SendJSON(c, http.StatusCreated, createdBookmark)
		return
	}

	// Save the processed bookmark back to database
	updatedBookmark, err := deps.Domains().Bookmarks().UpdateBookmark(c.Request().Context(), processedBookmark.Bookmark)
	if err != nil {
		deps.Logger().WithError(err).Warn("failed to save processed bookmark, returning created bookmark")
		// Return the created bookmark even if saving processed version failed
		response.SendJSON(c, http.StatusCreated, createdBookmark)
		return
	}

	response.SendJSON(c, http.StatusCreated, updatedBookmark)
}

// HandleListBookmarks lists bookmarks with optional filtering
//
//	@Summary					List bookmarks with optional filtering and pagination.
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						keyword	query	string	false	"Search keyword"
//	@Param						tags	query	string	false	"Comma-separated list of tags to include"
//	@Param						exclude	query	string	false	"Comma-separated list of tags to exclude"
//	@Param						page	query	int		false	"Page number (default: 1)"
//	@Param						limit	query	int		false	"Items per page (default: 30, max: 100)"
//	@Produce					json
//	@Success					200	{object}	model.PaginatedResponse[model.BookmarkDTO]
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/bookmarks [get]
func HandleListBookmarks(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
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
		p, err := strconv.Atoi(pageStr)
		if err != nil || p <= 0 {
			response.SendError(c, http.StatusBadRequest, "Invalid page parameter: must be a positive integer")
			return
		}
		page = p
	}

	limit := 30
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 || l > 100 {
			response.SendError(c, http.StatusBadRequest, "Invalid limit parameter: must be an integer between 1 and 100")
			return
		}
		limit = l
	}

	// Prepare search options using domain-level options
	searchOptions := model.BookmarksSearchOptions{
		Tags:         tags,
		ExcludedTags: excludedTags,
		Keyword:      keyword,
		Limit:        limit,
		Offset:       (page - 1) * limit,
		OrderMethod:  model.ByLastAddedSearchOrder,
	}

	// Search bookmarks through domain layer
	bookmarks, err := deps.Domains().Bookmarks().SearchBookmarks(c.Request().Context(), searchOptions)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to search bookmarks")
		response.SendError(c, http.StatusInternalServerError, "Failed to search bookmarks")
		return
	}

	// Get total count
	countOptions := model.BookmarksSearchOptions{
		Tags:         tags,
		ExcludedTags: excludedTags,
		Keyword:      keyword,
	}
	totalCount, err := deps.Domains().Bookmarks().CountBookmarks(c.Request().Context(), countOptions)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to get bookmarks count")
		response.SendError(c, http.StatusInternalServerError, "Failed to get bookmarks count")
		return
	}

	// Return paginated response
	paginatedResponse := model.PaginatedResponse[model.BookmarkDTO]{
		Items: bookmarks,
		Total: totalCount,
	}

	response.SendJSON(c, http.StatusOK, paginatedResponse)
}

// HandleGetBookmark gets a single bookmark by ID
//
//	@Summary					Get a bookmark by ID.
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						id	path	int	true	"Bookmark ID"
//	@Produce					json
//	@Success					200	{object}	model.BookmarkDTO
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					404	{object}	nil	"Bookmark not found"
//	@Router						/api/v1/bookmarks/{id} [get]
func HandleGetBookmark(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	bookmarkID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	bookmark, err := deps.Domains().Bookmarks().GetBookmark(c.Request().Context(), model.DBID(bookmarkID))
	if err != nil {
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		deps.Logger().WithError(err).Error("failed to get bookmark")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, bookmark)
}

type updateBookmarkPayload struct {
	URL     *string
	Title   *string
	Excerpt *string
	Public  *int
}

// HandleUpdateBookmark updates an existing bookmark
//
//	@Summary					Update an existing bookmark.
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						id		path	int						true	"Bookmark ID"
//	@Param						payload	body	updateBookmarkPayload	true	"Update Bookmark Payload"
//	@Produce					json
//	@Success					200	{object}	model.BookmarkDTO
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					404	{object}	nil	"Bookmark not found"
//	@Failure					400	{object}	nil	"Invalid request payload"
//	@Router						/api/v1/bookmarks/{id} [put]
func HandleUpdateBookmark(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
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
		if errors.Is(err, model.ErrBookmarkNotFound) {
			response.SendError(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		deps.Logger().WithError(err).Error("failed to get bookmark")
		response.SendInternalServerError(c)
		return
	}

	// Convert to Bookmark for domain operations
	bookmark := existingBookmark.ToBookmark()

	// Validate URL if provided
	if payload.URL != nil && strings.TrimSpace(*payload.URL) == "" {
		response.SendError(c, http.StatusBadRequest, "URL should not be empty")
		return
	}

	// Update fields if provided
	if payload.URL != nil {
		cleanURL, err := core.RemoveUTMParams(strings.TrimSpace(*payload.URL))
		if err != nil {
			response.SendError(c, http.StatusBadRequest, "Invalid URL")
			return
		}
		bookmark.URL = cleanURL
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
	if errors.Is(err, model.ErrBookmarkAlreadyExists) {
		response.SendError(c, http.StatusConflict, "already_exists.bookmark")
		return
	}
	if err != nil {
		deps.Logger().WithError(err).Error("failed to update bookmark")
		response.SendError(c, http.StatusInternalServerError, "Failed to update bookmark")
		return
	}

	response.SendJSON(c, http.StatusOK, updatedBookmark)
}

type deleteBookmarksPayload struct {
	IDs []int `validate:"required"`
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
//	@Summary					Delete one or more bookmarks.
//	@Tags						Bookmarks
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						payload	body	deleteBookmarksPayload	true	"Delete Bookmarks Payload"
//	@Produce					json
//	@Success					200	{object}	nil
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					400	{object}	nil	"Invalid request payload"
//	@Router						/api/v1/bookmarks [delete]
func HandleDeleteBookmarks(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
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
		deps.Logger().WithError(err).Error("failed to delete bookmarks")
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
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
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
		deps.Logger().WithError(err).Error("failed to update bookmark tags")
		response.SendError(c, http.StatusInternalServerError, "Failed to update bookmarks")
		return
	}

	bookmarks, err := deps.Domains().Bookmarks().GetBookmarks(c.Request().Context(), payload.BookmarkIDs)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to get updated bookmarks")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, bookmarks)
}
