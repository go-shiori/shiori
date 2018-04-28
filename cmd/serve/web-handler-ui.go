package serve

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	fp "path/filepath"

	"github.com/julienschmidt/httprouter"
)

// serveFiles serve files
func (h *webHandler) serveFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := serveFile(w, r.URL.Path)
	checkError(err)
}

// serveIndexPage is handler for GET /
func (h *webHandler) serveIndexPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkToken(r)
	if err != nil {
		redirectPage(w, r, "/login")
		return
	}

	err = serveFile(w, "index.html")
	checkError(err)
}

// serveLoginPage is handler for GET /login
func (h *webHandler) serveLoginPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkToken(r)
	if err == nil {
		redirectPage(w, r, "/")
		return
	}

	err = serveFile(w, "login.html")
	checkError(err)
}

// serveBookmarkCache is handler for GET /bookmark/:id
func (h *webHandler) serveBookmarkCache(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get bookmark ID from URL
	id := ps.ByName("id")

	// Get bookmarks in database
	bookmarks, err := h.db.GetBookmarks(true, id)
	checkError(err)

	if len(bookmarks) == 0 {
		panic(fmt.Errorf("No bookmark with matching index"))
	}

	// Execute template
	err = h.tplCache.Execute(w, &bookmarks[0])
	checkError(err)
}

func serveFile(w http.ResponseWriter, path string) error {
	// Open file
	src, err := assets.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	// Get content type
	ext := fp.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}

	// Serve file
	_, err = io.Copy(w, src)
	return err
}
