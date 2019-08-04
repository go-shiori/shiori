package webserver

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	nurl "net/url"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"

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
		}

		return
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

	// Create template
	funcMap := template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"hostname": func(s string) string {
			parsed, err := nurl.ParseRequestURI(s)
			if err != nil || len(parsed.Scheme) == 0 {
				return s
			}
			return parsed.Hostname()
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
