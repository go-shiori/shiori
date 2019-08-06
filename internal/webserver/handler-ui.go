package webserver

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/shiori/pkg/warc"
	"github.com/julienschmidt/httprouter"
)

// serveFile is handler for general file request
func (h *handler) serveFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := serveFile(w, r.URL.Path, true)
	checkError(err)
}

// serveJsFile is handler for GET /js/*filepath
func (h *handler) serveJsFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filePath := r.URL.Path
	fileName := path.Base(filePath)
	fileDir := path.Dir(filePath)

	if developmentMode && fp.Ext(fileName) == ".js" && strings.HasSuffix(fileName, ".min.js") {
		fileName = strings.TrimSuffix(fileName, ".min.js") + ".js"
		filePath = path.Join(fileDir, fileName)
		if assetExists(filePath) {
			redirectPage(w, r, filePath)
			return
		}
	}

	err := serveFile(w, r.URL.Path, true)
	checkError(err)
}

// serveIndexPage is handler for GET /
func (h *handler) serveIndexPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	if err != nil {
		redirectPage(w, r, "/login")
		return
	}

	err = serveFile(w, "index.html", false)
	checkError(err)
}

// serveLoginPage is handler for GET /login
func (h *handler) serveLoginPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session is not valid
	err := h.validateSession(r)
	if err == nil {
		redirectPage(w, r, "/")
		return
	}

	err = serveFile(w, "login.html", false)
	checkError(err)
}

// serveBookmarkContent is handler for GET /bookmark/:id/content
func (h *handler) serveBookmarkContent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get bookmark ID from URL
	strID := ps.ByName("id")
	id, err := strconv.Atoi(strID)
	checkError(err)

	// Get bookmark in database
	bookmark, exist := h.DB.GetBookmark(id, "")
	if !exist {
		panic(fmt.Errorf("Bookmark not found"))
	}

	// Check if it has archive
	archivePath := fp.Join(h.DataDir, "archive", strID)
	if fileExists(archivePath) {
		bookmark.HasArchive = true
	}

	// Create template
	funcMap := template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	tplCache, err := createTemplate("content.html", funcMap)
	checkError(err)

	// Execute template
	err = tplCache.Execute(w, &bookmark)
	checkError(err)
}

// serveThumbnailImage is handler for GET /bookmark/:id/thumb
func (h *handler) serveThumbnailImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get bookmark ID from URL
	id := ps.ByName("id")

	// Open image
	imgPath := fp.Join(h.DataDir, "thumb", id)
	img, err := os.Open(imgPath)
	checkError(err)
	defer img.Close()

	// Get image type from its 512 first bytes
	buffer := make([]byte, 512)
	_, err = img.Read(buffer)
	checkError(err)

	mimeType := http.DetectContentType(buffer)
	w.Header().Set("Content-Type", mimeType)

	// Serve image
	img.Seek(0, 0)
	_, err = io.Copy(w, img)
	checkError(err)
}

// serveBookmarkArchive is handler for GET /bookmark/:id/archive/*filepath
func (h *handler) serveBookmarkArchive(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get parameter from URL
	strID := ps.ByName("id")
	resourcePath := ps.ByName("filepath")
	resourcePath = strings.TrimPrefix(resourcePath, "/")

	// Get bookmark from database
	id, err := strconv.Atoi(strID)
	checkError(err)

	bookmark, exist := h.DB.GetBookmark(id, "")
	if !exist {
		panic(fmt.Errorf("Bookmark not found"))
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
		tpl, err := template.New("archive").Parse(
			`<div id="shiori-archive-header">
			<p id="shiori-logo"><span>栞</span>shiori</p>
			<div class="spacer"></div>
			<a href="{{.URL}}" target="_blank">View Original</a>
			{{if .HasContent}}
			<a href="/bookmark/{{.ID}}/content">View Readable</a>
			{{end}}
			</div>`)
		checkError(err)

		tplOutput := bytes.NewBuffer(nil)
		err = tpl.Execute(tplOutput, &bookmark)
		checkError(err)

		doc.Find("head").AppendHtml(`<link href="/css/source-sans-pro.min.css" rel="stylesheet">`)
		doc.Find("head").AppendHtml(`<link href="/css/archive.css" rel="stylesheet">`)
		doc.Find("body").PrependHtml(tplOutput.String())

		// Revert back to HTML
		outerHTML, err := goquery.OuterHtml(doc.Selection)
		checkError(err)

		// Gzip it again and send to response writer
		gzipWriter := gzip.NewWriter(w)
		gzipWriter.Write([]byte(outerHTML))
		gzipWriter.Flush()
		return
	}

	// Serve content
	w.Write(content)
}
