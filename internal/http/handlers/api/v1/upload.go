package api_v1

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

const maxUploadSize = 32 << 20 // 32 MB

// HandleUploadBookmark creates or updates a bookmark from a SingleFile HTML upload.
// URL and title are extracted automatically from the HTML — no extra fields required.
//
//	@Summary					Create/update a bookmark from a SingleFile HTML export.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Accept						multipart/form-data
//	@Param						file	formData	file	true	"SingleFile HTML export (.html)"
//	@Produce					json
//	@Success					200	{object}	model.BookmarkDTO
//	@Failure					400	{object}	nil	"Missing file or no URL found in HTML"
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Failure					500	{object}	nil	"Storage or processing error"
//	@Router						/api/v1/bookmarks/upload [post]
func HandleUploadBookmark(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		response.SendError(c, http.StatusForbidden, err.Error())
		return
	}

	if err := c.Request().ParseMultipartForm(maxUploadSize); err != nil {
		response.SendError(c, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	f, _, err := c.Request().FormFile("file")
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Missing 'file' field")
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to read uploaded file")
		return
	}

	meta := core.ExtractSingleFileMetadata(content)
	if meta.URL == "" {
		response.SendError(c, http.StatusBadRequest, "No original URL found in uploaded HTML")
		return
	}

	cleanURL, err := core.RemoveUTMParams(meta.URL)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid URL in HTML: %v", err))
		return
	}
	meta.URL = cleanURL

	ctx := c.Request().Context()
	db := deps.Database()

	book, exists, err := db.GetBookmark(ctx, 0, meta.URL)
	if err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to look up bookmark")
		return
	}

	if !exists {
		title := meta.Title
		if title == "" {
			title = meta.URL
		}
		book = model.BookmarkDTO{URL: meta.URL, Title: title}
		saved, err := db.SaveBookmarks(ctx, true, book)
		if err != nil {
			response.SendError(c, http.StatusInternalServerError, "Failed to save bookmark")
			return
		}
		book = saved[0]
	}

	book.CreateArchive = true
	req := core.ProcessRequest{
		DataDir:     deps.Config().Storage.DataDir,
		Bookmark:    book,
		Content:     bytes.NewReader(content),
		ContentType: "text/html; charset=UTF-8",
	}

	processed, isFatal, err := core.ProcessBookmark(deps, req)
	if err != nil {
		if isFatal {
			response.SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to process bookmark: %v", err))
			return
		}
		log.Printf("non-fatal error processing upload for %s: %v", meta.URL, err)
	}

	if _, err := db.SaveBookmarks(ctx, false, processed); err != nil {
		response.SendError(c, http.StatusInternalServerError, "Failed to save processed bookmark")
		return
	}

	response.SendJSON(c, http.StatusOK, processed)
}
