package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/julienschmidt/httprouter"
)

func downloadBookmarkContent(deps model.Dependencies, book *model.BookmarkDTO, dataDir string, _ *http.Request, keepTitle, keepExcerpt bool) (*model.BookmarkDTO, error) {
	content, contentType, err := core.DownloadBookmark(book.URL)
	if err != nil {
		return nil, fmt.Errorf("error downloading url: %s", err)
	}

	processRequest := core.ProcessRequest{
		DataDir:     dataDir,
		Bookmark:    *book,
		Content:     content,
		ContentType: contentType,
		KeepTitle:   keepTitle,
		KeepExcerpt: keepExcerpt,
	}

	result, isFatalErr, err := core.ProcessBookmark(deps, processRequest)
	content.Close()

	if err != nil && isFatalErr {
		return nil, fmt.Errorf("failed to process: %v", err)
	}

	return &result, err
}

// ApiGetBookmarks is handler for GET /api/bookmarks
func (h *Handler) ApiGetBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Get URL queries
	keyword := r.URL.Query().Get("keyword")
	strPage := r.URL.Query().Get("page")
	strTags := r.URL.Query().Get("tags")
	strExcludedTags := r.URL.Query().Get("exclude")

	tags := strings.Split(strTags, ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}

	excludedTags := strings.Split(strExcludedTags, ",")
	if len(excludedTags) == 1 && excludedTags[0] == "" {
		excludedTags = []string{}
	}

	page, _ := strconv.Atoi(strPage)
	if page < 1 {
		page = 1
	}

	// Prepare filter for database
	searchOptions := model.DBGetBookmarksOptions{
		Tags:         tags,
		ExcludedTags: excludedTags,
		Keyword:      keyword,
		Limit:        30,
		Offset:       (page - 1) * 30,
		OrderMethod:  model.ByLastAdded,
	}

	// Calculate max page
	nBookmarks, err := h.DB.GetBookmarksCount(ctx, searchOptions)
	checkError(err)
	maxPage := int(math.Ceil(float64(nBookmarks) / 30))

	// Fetch all matching bookmarks
	bookmarks, err := h.DB.GetBookmarks(ctx, searchOptions)
	checkError(err)

	// Get image URL for each bookmark, and check if it has archive
	for i := range bookmarks {
		strID := strconv.Itoa(bookmarks[i].ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		archivePath := fp.Join(h.DataDir, "archive", strID)
		ebookPath := fp.Join(h.DataDir, "ebook", strID+".epub")

		if FileExists(imgPath) {
			bookmarks[i].ImageURL = path.Join(h.RootPath, "bookmark", strID, "thumb")
		}

		if FileExists(archivePath) {
			bookmarks[i].HasArchive = true
		}
		if FileExists(ebookPath) {
			bookmarks[i].HasEbook = true
		}
	}

	// Return JSON response
	resp := map[string]interface{}{
		"page":      page,
		"maxPage":   maxPage,
		"bookmarks": bookmarks,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&resp)
	checkError(err)
}

// ApiGetTags is handler for GET /api/tags
func (h *Handler) ApiGetTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Fetch all tags
	tags, err := h.DB.GetTags(ctx)
	checkError(err)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&tags)
	checkError(err)
}

// ApiRenameTag is handler for PUT /api/tag
func (h *Handler) ApiRenameTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	tag := model.Tag{}
	err = json.NewDecoder(r.Body).Decode(&tag)
	checkError(err)

	// Update name
	err = h.DB.RenameTag(ctx, tag.ID, tag.Name)
	checkError(err)

	fmt.Fprint(w, 1)
}

// Bookmark is the record for an URL.
type apiInsertBookmarkPayload struct {
	URL           string      `json:"url"`
	Title         string      `json:"title"`
	Excerpt       string      `json:"excerpt"`
	Tags          []model.Tag `json:"tags"`
	CreateArchive bool        `json:"create_archive"`
	CreateEbook   bool        `json:"create_ebook"`
	MakePublic    int         `json:"public"`
	Async         bool        `json:"async"`
}

// newApiInsertBookmarkPayload
// Returns the payload struct with its defaults
func newAPIInsertBookmarkPayload() *apiInsertBookmarkPayload {
	return &apiInsertBookmarkPayload{
		CreateArchive: false,
		CreateEbook:   false,
		Async:         true,
	}
}

// ApiInsertBookmark is handler for POST /api/bookmark
func (h *Handler) ApiInsertBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	payload := newAPIInsertBookmarkPayload()
	err = json.NewDecoder(r.Body).Decode(&payload)
	checkError(err)

	book := &model.BookmarkDTO{
		URL:           payload.URL,
		Title:         payload.Title,
		Excerpt:       payload.Excerpt,
		Tags:          make([]model.TagDTO, len(payload.Tags)),
		Public:        payload.MakePublic,
		CreateArchive: payload.CreateArchive,
		CreateEbook:   payload.CreateEbook,
	}

	for i, tag := range payload.Tags {
		book.Tags[i] = tag.ToDTO()
	}

	// Clean up bookmark URL
	book.URL, err = core.RemoveUTMParams(book.URL)
	if err != nil {
		panic(fmt.Errorf("failed to clean URL: %v", err))
	}

	userHasDefinedTitle := book.Title != ""
	// Make sure bookmark's title not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Save bookmark to database
	results, err := h.DB.SaveBookmarks(ctx, true, *book)
	if err != nil || len(results) == 0 {
		panic(fmt.Errorf("failed to save bookmark: %v", err))
	}

	book = &results[0]

	if payload.Async {
		go func() {
			bookmark, err := downloadBookmarkContent(h.dependencies, book, h.DataDir, r, userHasDefinedTitle, book.Excerpt != "")
			if err != nil {
				log.Printf("error downloading boorkmark: %s", err)
				return
			}
			if _, err := h.DB.SaveBookmarks(context.Background(), false, *bookmark); err != nil {
				log.Printf("failed to save bookmark: %s", err)
			}
		}()
	} else {
		// Workaround. Download content after saving the bookmark so we have the proper database
		// id already set in the object regardless of the database engine.
		book, err = downloadBookmarkContent(h.dependencies, book, h.DataDir, r, userHasDefinedTitle, book.Excerpt != "")
		if err != nil {
			log.Printf("error downloading boorkmark: %s", err)
		} else if _, err := h.DB.SaveBookmarks(ctx, false, *book); err != nil {
			log.Printf("failed to save bookmark: %s", err)
		}
	}

	// Return the new bookmark
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(results[0])
	checkError(err)
}

// ApiDeleteBookmarks is handler for DELETE /api/bookmark
func (h *Handler) ApiDeleteBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	ids := []int{}
	err = json.NewDecoder(r.Body).Decode(&ids)
	checkError(err)

	// Delete bookmarks
	err = h.DB.DeleteBookmarks(ctx, ids...)
	checkError(err)

	// Delete thumbnail image and archives from local disk
	for _, id := range ids {
		strID := strconv.Itoa(id)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		archivePath := fp.Join(h.DataDir, "archive", strID)
		ebookPath := fp.Join(h.DataDir, "ebook", strID+".epub")

		os.Remove(imgPath)
		os.Remove(archivePath)
		os.Remove(ebookPath)
	}

	fmt.Fprint(w, 1)
}

// ApiUpdateBookmark is handler for PUT /api/bookmarks
func (h *Handler) ApiUpdateBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := model.BookmarkDTO{}
	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Validate input
	if request.Title == "" {
		panic(fmt.Errorf("title must not empty"))
	}

	// Get existing bookmark from database
	filter := model.DBGetBookmarksOptions{
		IDs:         []int{request.ID},
		WithContent: true,
	}

	bookmarks, err := h.DB.GetBookmarks(ctx, filter)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("no bookmark with matching ids"))
	}

	// Set new bookmark data
	book := bookmarks[0]
	book.URL = request.URL
	book.Title = request.Title
	book.Excerpt = request.Excerpt
	book.Public = request.Public

	// Clean up bookmark URL
	book.URL, err = core.RemoveUTMParams(book.URL)
	if err != nil {
		panic(fmt.Errorf("failed to clean URL: %v", err))
	}

	// Set new tags
	for i := range book.Tags {
		book.Tags[i].Deleted = true
	}

	for _, newTag := range request.Tags {
		for i, oldTag := range book.Tags {
			if newTag.Name == oldTag.Name {
				newTag.ID = oldTag.ID
				book.Tags[i].Deleted = false
				break
			}
		}

		if newTag.ID == 0 {
			book.Tags = append(book.Tags, newTag)
		}
	}

	// Set bookmark modified
	book.ModifiedAt = ""

	// Update database
	res, err := h.DB.SaveBookmarks(ctx, false, book)
	checkError(err)

	// Add thumbnail image to the saved bookmarks again
	newBook := res[0]
	newBook.ImageURL = request.ImageURL
	newBook.HasArchive = request.HasArchive

	// Return new saved result
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&newBook)
	checkError(err)
}

// ApiUpdateBookmarkTags is handler for PUT /api/bookmarks/tags
func (h *Handler) ApiUpdateBookmarkTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := struct {
		IDs  []int       `json:"ids"`
		Tags []model.Tag `json:"tags"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Validate input
	if len(request.IDs) == 0 || len(request.Tags) == 0 {
		panic(fmt.Errorf("IDs and tags must not empty"))
	}

	// Get existing bookmark from database
	filter := model.DBGetBookmarksOptions{
		IDs:         request.IDs,
		WithContent: true,
	}

	bookmarks, err := h.DB.GetBookmarks(ctx, filter)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("no bookmark with matching ids"))
	}

	// Set new tags
	for i, book := range bookmarks {
		for _, newTag := range request.Tags {
			for _, oldTag := range book.Tags {
				if newTag.Name == oldTag.Name {
					newTag.ID = oldTag.ID
					break
				}
			}

			if newTag.ID == 0 {
				book.Tags = append(book.Tags, newTag.ToDTO())
			}
		}

		bookmarks[i] = book
	}

	// Update database
	bookmarks, err = h.DB.SaveBookmarks(ctx, false, bookmarks...)
	checkError(err)

	// Get image URL for each bookmark
	for i := range bookmarks {
		strID := strconv.Itoa(bookmarks[i].ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		imgURL := path.Join(h.RootPath, "bookmark", strID, "thumb")

		if FileExists(imgPath) {
			bookmarks[i].ImageURL = imgURL
		}
	}

	// Return new saved result
	err = json.NewEncoder(w).Encode(&bookmarks)
	checkError(err)
}
