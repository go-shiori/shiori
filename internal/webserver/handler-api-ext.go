package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	book, exist := h.DB.GetBookmark(0, request.URL)
	if !exist {
		// save the bookmark first and process it in a goroutine
		book = request
		if book.Title == "" {
			book.Title = "unprocessed"
		}

		book.ID, err = h.DB.CreateNewID("bookmark")
		if err != nil {
			panic(fmt.Errorf("failed to create ID: %v", err))
		}

		results, err := h.DB.SaveBookmarks(book)
		if err != nil || len(results) == 0 {
			panic(fmt.Errorf("failed to save bookmark: %v", err))
		}
	}

	go func() {
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
		}

		// Since we are using extension, the extension might send the HTML content
		// so no need to download it again here. However, if it's empty, it might be not HTML file
		// so we download it here.
		var contentType string
		var contentBuffer io.Reader

		if book.HTML == "" {
			contentBuffer, contentType, _ = core.DownloadBookmark(book.URL)
		} else {
			contentType = "text/html; charset=UTF-8"
			contentBuffer = bytes.NewBufferString(book.HTML)
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
			fmt.Printf("url: [%s] is processing...", book.URL)
			book, isFatalErr, err = core.ProcessBookmark(request)
			fmt.Printf("done\n")

			if tmp, ok := contentBuffer.(io.ReadCloser); ok {
				tmp.Close()
			}

			if err != nil && isFatalErr {
				panic(fmt.Errorf("failed to process bookmark: %v", err))
			}
		}

		// Save bookmark to database
		results, err := h.DB.SaveBookmarks(book)
		if err != nil || len(results) == 0 {
			panic(fmt.Errorf("failed to save bookmark: %v", err))
		}
	}()

	w.Header().Set("Content-Type", "application/json")
}

// apiDeleteViaExtension is handler for DELETE /api/bookmark/ext
func (h *handler) apiDeleteViaExtension(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := model.Bookmark{}
	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Check if bookmark already exists.
	book, exist := h.DB.GetBookmark(0, request.URL)
	if exist {
		// Delete bookmarks
		err = h.DB.DeleteBookmarks(book.ID)
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
