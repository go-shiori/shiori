package webserver

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/warc"
	"github.com/julienschmidt/httprouter"
)

// ServeBookmarkArchive is handler for GET /bookmark/:id/archive/*filepath
func (h *Handler) ServeBookmarkArchive(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Get parameter from URL
	strID := ps.ByName("id")
	resourcePath := ps.ByName("filepath")
	resourcePath = strings.TrimPrefix(resourcePath, "/")

	// Get bookmark from database
	id, err := strconv.Atoi(strID)
	checkError(err)

	bookmark, exist, err := h.DB.GetBookmark(ctx, id, "")
	checkError(err)

	if !exist {
		panic(fmt.Errorf("bookmark not found"))
	}

	// If it's not public, make sure session still valid
	if bookmark.Public != 1 {
		err = h.validateSession(r)
		if err != nil {
			newPath := path.Join(h.RootPath, "/login")
			redirectURL := createRedirectURL(newPath, r.URL.String())
			redirectPage(w, r, redirectURL)
			return
		}
	}

	// Open archive, look in cache first
	var archive *warc.Archive
	cacheData, found := h.ArchiveCache.Get(strID)

	if found {
		archive = cacheData.(*warc.Archive)
	} else {
		archivePath := fp.Join(h.DataDir, "archive", strID)
		archive, err = warc.Open(archivePath)
		checkError(err)

		h.ArchiveCache.Set(strID, archive, 0)
	}

	content, contentType, err := archive.Read(resourcePath)
	checkError(err)

	// Set response header
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", contentType)

	// If this is HTML and root, inject shiori header
	if strings.Contains(strings.ToLower(contentType), "text/html") && resourcePath == "" {
		// Extract gzip
		buffer := bytes.NewBuffer(content)
		gzipReader, err := gzip.NewReader(buffer)
		checkError(err)

		// Parse gzipped content
		doc, err := goquery.NewDocumentFromReader(gzipReader)
		checkError(err)

		// Add Shiori overlay
		tplOutput := bytes.NewBuffer(nil)
		err = h.templates["archive"].Execute(tplOutput, &bookmark)
		checkError(err)

		archiveCSSPath := path.Join(h.RootPath, "/assets/css/archive.css")

		docHead := doc.Find("head")
		docHead.PrependHtml(`<meta charset="UTF-8">`)
		docHead.AppendHtml(`<link href="` + archiveCSSPath + `" rel="stylesheet">`)
		doc.Find("body").PrependHtml(tplOutput.String())
		doc.Find("body").AddClass("shiori-archive-content")

		// Revert back to HTML
		outerHTML, err := goquery.OuterHtml(doc.Selection)
		checkError(err)

		// Gzip it again and send to response writer
		gzipWriter := gzip.NewWriter(w)
		if _, err := gzipWriter.Write([]byte(outerHTML)); err != nil {
			log.Printf("error writing gzip file: %s", err)
		}
		gzipWriter.Flush()
		return
	}

	// Serve content
	if _, err := w.Write(content); err != nil {
		log.Printf("error writing response: %s", err)
	}
}
