package webserver

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/warc"
	"github.com/julienschmidt/httprouter"

	"github.com/go-shiori/shiori/internal/model"
)

// ServeBookmarkContent is handler for GET /bookmark/:id/content
func (h *Handler) ServeBookmarkContent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Get bookmark ID from URL
	strID := ps.ByName("id")
	id, err := strconv.Atoi(strID)
	checkError(err)

	// Get bookmark in database
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

	// Check if it has ebook.
	ebookPath := fp.Join(h.DataDir, "ebook", strID+".epub")
	if FileExists(ebookPath) {
		bookmark.HasEbook = true
	}
	archivePath := fp.Join(h.DataDir, "archive", strID)
	if FileExists(archivePath) {
		bookmark.HasArchive = true

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

		// Find all image and convert its source to use the archive URL.
		createArchivalURL := func(archivalName string) string {
			archivalURL := *r.URL
			archivalURL.Path = path.Join(h.RootPath, "bookmark", strID, "archive", archivalName)
			return archivalURL.String()
		}

		buffer := strings.NewReader(bookmark.HTML)
		doc, err := goquery.NewDocumentFromReader(buffer)
		checkError(err)

		doc.Find("img, picture, figure, source").Each(func(_ int, node *goquery.Selection) {
			// Get the needed attributes
			src, _ := node.Attr("src")
			strSrcSets, _ := node.Attr("srcset")

			// Convert `src` attributes
			if src != "" {
				archivalName := getArchivalName(src)
				if archivalName != "" && archive.HasResource(archivalName) {
					node.SetAttr("src", createArchivalURL(archivalName))
				}
			}

			// Split srcset by comma, then process it like any URLs
			srcSets := strings.Split(strSrcSets, ",")
			for i, srcSet := range srcSets {
				srcSet = strings.TrimSpace(srcSet)
				parts := strings.SplitN(srcSet, " ", 2)
				if parts[0] == "" {
					continue
				}

				archivalName := getArchivalName(parts[0])
				if archivalName != "" && archive.HasResource(archivalName) {
					archivalURL := createArchivalURL(archivalName)
					srcSets[i] = strings.Replace(srcSets[i], parts[0], archivalURL, 1)
				}
			}

			if len(srcSets) > 0 {
				node.SetAttr("srcset", strings.Join(srcSets, ","))
			}
		})

		bookmark.HTML, err = goquery.OuterHtml(doc.Selection)
		checkError(err)
	}

	// Execute template
	if developmentMode {
		if err := h.PrepareTemplates(); err != nil {
			log.Printf("error during template preparation: %s", err)
		}
	}

	tplData := struct {
		RootPath string
		Book     model.Bookmark
	}{h.RootPath, bookmark}

	err = h.templates["content"].Execute(w, &tplData)
	checkError(err)
}

// ServeThumbnailImage is handler for GET /bookmark/:id/thumb
func (h *Handler) ServeThumbnailImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get bookmark ID from URL
	ctx := r.Context()

	// Get bookmark ID from URL
	strID := ps.ByName("id")

	// Get bookmark from database
	id, err := strconv.Atoi(strID)
	checkError(err)
	bookmark, exist, err := h.DB.GetBookmark(ctx, id, "")
	checkError(err)

	if !exist {
		log.Println("error: bookmark not found")
		return
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
	// Open image
	imgPath := fp.Join(h.DataDir, "thumb", strID)
	img, err := os.Open(imgPath)
	checkError(err)
	defer img.Close()

	// Get image type from its 512 first bytes
	buffer := make([]byte, 512)
	_, err = img.Read(buffer)
	checkError(err)

	mimeType := http.DetectContentType(buffer)
	w.Header().Set("Content-Type", mimeType)

	// Set cache value
	info, err := img.Stat()
	checkError(err)

	etag := fmt.Sprintf(`W/"%x-%x"`, info.ModTime().Unix(), info.Size())
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "max-age=86400")

	// Serve image
	if _, err := img.Seek(0, 0); err != nil {
		log.Printf("error during image seek: %s", err)
	}
	_, err = io.Copy(w, img)
	checkError(err)
}

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

// serveEbook is handler for GET /bookmark/:id/ebook
func (h *Handler) ServeBookmarkEbook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Get bookmark ID from URL
	strID := ps.ByName("id")
	id, err := strconv.Atoi(strID)
	checkError(err)
	// Get bookmark in database
	bookmark, exist, err := h.DB.GetBookmark(ctx, id, "")
	checkError(err)

	if !exist {
		http.Error(w, "bookmark not found", http.StatusNotFound)
		return
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

	// Check if it has ebook.
	ebookPath := fp.Join(h.DataDir, "ebook", strID+".epub")
	if !FileExists(ebookPath) {
		http.Error(w, "ebook not found", http.StatusNotFound)
		return
	}

	epub, err := os.Open(ebookPath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusNotFound)
		return
	}
	defer epub.Close()
	// Set content type
	w.Header().Set("Content-Type", "application/epub+zip")
	// Set cache value
	info, err := epub.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	etag := fmt.Sprintf(`W/"%x-%x"`, info.ModTime().Unix(), info.Size())
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "max-age=86400")

	// Serve epub file
	if _, err := epub.Seek(0, 0); err != nil {
		log.Printf("error during epub seek: %s", err)
	}
	_, err = io.Copy(w, epub)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
