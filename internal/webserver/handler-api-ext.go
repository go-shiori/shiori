package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/pkg/warc"
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

	// Clean up URL by removing its fragment and UTM parameters
	tmp, err := nurl.Parse(request.URL)
	if err != nil || tmp.Scheme == "" || tmp.Hostname() == "" {
		panic(fmt.Errorf("URL is not valid"))
	}

	tmp.Fragment = ""
	clearUTMParams(tmp)
	request.URL = tmp.String()

	// Check if bookmark already exists.
	book, exist := h.DB.GetBookmark(0, request.URL)

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
	} else {
		book = request
		book.ID, err = h.DB.CreateNewID("bookmark")
		if err != nil {
			panic(fmt.Errorf("failed to create ID: %v", err))
		}
	}

	// Since we are using extension, the extension might send the HTML content
	// so no need to download it again here. However, if it's empty, it might be not HTML file
	// so we download it here.
	contentType := "text/html; charset=UTF-8"
	contentBuffer := bytes.NewBufferString(book.HTML)
	if book.HTML == "" {
		func() {
			// Prepare download request
			req, err := http.NewRequest("GET", book.URL, nil)
			if err != nil {
				return
			}

			// Send download request
			req.Header.Set("User-Agent", "Shiori/2.0.0 (+https://github.com/go-shiori/shiori)")
			resp, err := httpClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			// Save response for later use
			contentType = resp.Header.Get("Content-Type")

			contentBuffer.Reset()
			_, err = io.Copy(contentBuffer, resp.Body)
			if err != nil {
				return
			}
		}()
	}

	// At this point the web page already downloaded.
	// Time to process it.
	func() {
		// Split response so it can be processed several times
		archivalInput := bytes.NewBuffer(nil)
		readabilityInput := bytes.NewBuffer(nil)
		readabilityCheckInput := bytes.NewBuffer(nil)
		multiWriter := io.MultiWriter(archivalInput, readabilityInput, readabilityCheckInput)

		_, err = io.Copy(multiWriter, contentBuffer)
		if err != nil {
			return
		}

		// If it's HTML, parse the readable content.
		if strings.Contains(contentType, "text/html") {
			isReadable := readability.IsReadable(readabilityCheckInput)

			article, err := readability.FromReader(readabilityInput, book.URL)
			if err != nil {
				return
			}

			book.Author = article.Byline
			book.Content = article.TextContent
			book.HTML = article.Content

			if book.Title == "" {
				if article.Title == "" {
					book.Title = book.URL
				} else {
					book.Title = article.Title
				}
			}

			if book.Excerpt == "" {
				book.Excerpt = article.Excerpt
			}

			if !isReadable {
				book.Content = ""
			}

			book.HasContent = book.Content != ""

			// Get image for thumbnail and save it to local disk
			var imageURLs []string
			if article.Image != "" {
				imageURLs = append(imageURLs, article.Image)
			}

			if article.Favicon != "" {
				imageURLs = append(imageURLs, article.Favicon)
			}

			// Save article image to local disk
			strID := strconv.Itoa(book.ID)
			imgPath := fp.Join(h.DataDir, "thumb", strID)
			for _, imageURL := range imageURLs {
				err = downloadBookImage(imageURL, imgPath, time.Minute)
				if err == nil {
					book.ImageURL = path.Join("/", "bookmark", strID, "thumb")
					break
				}
			}
		}

		// Create offline archive as well
		archivePath := fp.Join(h.DataDir, "archive", fmt.Sprintf("%d", book.ID))
		os.Remove(archivePath)

		archivalRequest := warc.ArchivalRequest{
			URL:         book.URL,
			Reader:      archivalInput,
			ContentType: contentType,
		}

		err = warc.NewArchive(archivalRequest, archivePath)
		if err != nil {
			return
		}

		book.HasArchive = true
	}()

	// Save bookmark to database
	results, err := h.DB.SaveBookmarks(book)
	if err != nil || len(results) == 0 {
		panic(fmt.Errorf("failed to save bookmark: %v", err))
	}
	book = results[0]

	// Return the new bookmark
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&book)
	checkError(err)
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
