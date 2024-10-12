package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"strconv"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/julienschmidt/httprouter"
)

// ApiInsertViaExtension is handler for POST /api/bookmarks/ext
func (h *Handler) ApiInsertViaExtension(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := model.BookmarkDTO{}
	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Clean up bookmark URL
	request.URL, err = core.RemoveUTMParams(request.URL)
	if err != nil {
		panic(fmt.Errorf("failed to clean URL: %v", err))
	}

	// Check if bookmark already exists.
	book, exist, err := h.DB.GetBookmark(ctx, 0, request.URL)
	if err != nil {
		panic(fmt.Errorf("failed to get bookmark, URL: %v", err))
	}

	// If it already exists, we need to set ID and tags.
	if exist {
		book.HTML = request.HTML

		mapOldTags := map[string]model.Tag{}
		for _, oldTag := range book.Tags {
			mapOldTags[oldTag.Name] = oldTag
		}

		for _, newTag := range request.Tags {
			if _, tagExist := mapOldTags[newTag.Name]; !tagExist {
				book.Tags = append(book.Tags, newTag)
			}
		}
	} else if request.Title == "" {
		request.Title = request.URL
	}

	// Save the bookmark with whatever we already have downloaded
	// since we need the ID in order to download the archive
	// Only when old bookmark is not exists.
	if !exist {
		books, err := h.DB.SaveBookmarks(ctx, true, request)
		if err != nil {
			log.Printf("error saving bookmark before downloading content: %s", err)
			return
		}
		book = books[0]
	} else {
		books, err := h.DB.SaveBookmarks(ctx, false, book)
		if err != nil {
			log.Printf("error saving bookmark before downloading content: %s", err)
			return
		}
		book = books[0]
	}

	// At this point the web page already downloaded.
	// Time to process it.
	var result *model.BookmarkDTO
	var errArchiver error
	if request.HTML != "" {
		archiverReq := model.NewArchiverRequest(book, "text/html; charset=UTF-8", []byte(request.HTML))
		result, errArchiver = h.dependencies.Domains.Archiver.ProcessBookmarkArchive(archiverReq)
	} else {
		result, errArchiver = h.dependencies.Domains.Archiver.GenerateBookmarkArchive(book)
	}
	if errArchiver != nil {
		log.Printf("error downloading bookmark cache: %s", errArchiver)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Save the bookmark with whatever we already have downloaded
	// since we need the ID in order to download the archive
	books, err := h.DB.SaveBookmarks(ctx, request.ID == 0, *result)
	if err != nil {
		log.Printf("error saving bookmark from extension downloading content: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	book = books[0]

	// Return the new bookmark
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&result)
	checkError(err)
}

// ApiDeleteViaExtension is handler for DELETE /api/bookmark/ext
func (h *Handler) ApiDeleteViaExtension(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := model.BookmarkDTO{}
	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Check if bookmark already exists.
	book, exist, err := h.DB.GetBookmark(ctx, 0, request.URL)
	checkError(err)

	if exist {
		// Delete bookmarks
		err = h.DB.DeleteBookmarks(ctx, book.ID)
		checkError(err)

		// Delete thumbnail image and archives from local disk
		strID := strconv.Itoa(book.ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		archivePath := fp.Join(h.DataDir, "archive", strID)

		os.Remove(imgPath)
		os.Remove(archivePath)
	}

	fmt.Fprint(w, 1)
}
