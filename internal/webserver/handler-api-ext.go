package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"strconv"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/julienschmidt/httprouter"
)

// apiInsertViaExtension is handler for POST /api/bookmarks/ext
func (h *handler) apiInsertViaExtension(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := model.Bookmark{}
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

	// Since we are using extension, the extension might send the HTML content
	// so no need to download it again here. However, if it's empty, it might be not HTML file
	// so we download it here.
	var contentType string
	var contentBuffer io.Reader

	if request.HTML == "" {
		contentBuffer, contentType, _ = core.DownloadBookmark(request.URL)
	} else {
		contentType = "text/html; charset=UTF-8"
		contentBuffer = bytes.NewBufferString(request.HTML)
	}

	// Save the bookmark with whatever we already have downloaded
	// since we need the ID in order to download the archive
	// Only when old bookmark is not exists.
	if (!exist) {
		books, err := h.DB.SaveBookmarks(ctx, true, request)
		if err != nil {
			log.Printf("error saving bookmark before downloading content: %s", err)
			return
		}
		book = books[0]
	}

	// At this point the web page already downloaded.
	// Time to process it.
	if contentBuffer != nil {
		book.CreateArchive = true
		request := core.ProcessRequest{
			DataDir:     h.DataDir,
			Bookmark:    book,
			Content:     contentBuffer,
			ContentType: contentType,
		}

		var isFatalErr bool
		book, isFatalErr, err = core.ProcessBookmark(request)

		if tmp, ok := contentBuffer.(io.ReadCloser); ok {
			tmp.Close()
		}

		// If we can't process or update the saved bookmark, just log it and continue on with the
		// request.
		if err != nil && isFatalErr {
			log.Printf("failed to process bookmark: %v", err)
		} else if _, err := h.DB.SaveBookmarks(ctx, false, book); err != nil {
			log.Printf("error saving bookmark after downloading content: %s", err)
		}
	}

	// Return the new bookmark
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&book)
	checkError(err)
}

// apiDeleteViaExtension is handler for DELETE /api/bookmark/ext
func (h *handler) apiDeleteViaExtension(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := model.Bookmark{}
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
