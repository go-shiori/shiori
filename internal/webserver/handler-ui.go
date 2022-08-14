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

// serveFile is handler for general file request
func (h *handler) serveFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rootPath := strings.Trim(h.RootPath, "/")
	urlPath := strings.Trim(r.URL.Path, "/")
	filePath := strings.TrimPrefix(urlPath, rootPath)
	filePath = strings.Trim(filePath, "/")

	err := serveFile(w, filePath, true)
	checkError(err)
}

// serveJsFile is handler for GET /js/*filepath
func (h *handler) serveJsFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsFilePath := ps.ByName("filepath")
	jsFilePath = path.Join("js", jsFilePath)
	jsDir, jsName := path.Split(jsFilePath)

	if developmentMode && fp.Ext(jsName) == ".js" && strings.HasSuffix(jsName, ".min.js") {
		jsName = strings.TrimSuffix(jsName, ".min.js") + ".js"
		tmpPath := path.Join(jsDir, jsName)
		if assetExists(tmpPath) {
			jsFilePath = tmpPath
		}
	}

	err := serveFile(w, jsFilePath, true)
	checkError(err)
}

// serveIndexPage is handler for GET /
func (h *handler) serveIndexPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	if err != nil {
		newPath := path.Join(h.RootPath, "/login")
		redirectURL := createRedirectURL(newPath, r.URL.String())
		redirectPage(w, r, redirectURL)
		return
	}

	if developmentMode {
		if err := h.prepareTemplates(); err != nil {
			log.Printf("error during template preparation: %s", err)
		}
	}

	err = h.templates["index"].Execute(w, h.RootPath)
	checkError(err)
}

// serveLoginPage is handler for GET /login
func (h *handler) serveLoginPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session is not valid
	err := h.validateSession(r)
	if err == nil {
		redirectURL := path.Join(h.RootPath, "/")
		redirectPage(w, r, redirectURL)
		return
	}

	if developmentMode {
		if err := h.prepareTemplates(); err != nil {
			log.Printf("error during template preparation: %s", err)
		}
	}

	err = h.templates["login"].Execute(w, h.RootPath)
	checkError(err)
}

// serveBookmarkContent is handler for GET /bookmark/:id/content
func (h *handler) serveBookmarkContent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Check if it has archive.
	archivePath := fp.Join(h.DataDir, "archive", strID)
	if fileExists(archivePath) {
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
		if err := h.prepareTemplates(); err != nil {
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

// serveBookmarkArchive is handler for GET /bookmark/:id/archive/*filepath
func (h *handler) serveBookmarkArchive(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

		archiveCSSPath := path.Join(h.RootPath, "/css/archive.css")
		sourceSansProCSSPath := path.Join(h.RootPath, "/css/source-sans-pro.min.css")

		docHead := doc.Find("head")
		docHead.PrependHtml(`<meta charset="UTF-8">`)
		docHead.AppendHtml(`<link href="` + archiveCSSPath + `" rel="stylesheet">`)
		docHead.AppendHtml(`<link href="` + sourceSansProCSSPath + `" rel="stylesheet">`)
		doc.Find("body").PrependHtml(tplOutput.String())

		// Revert back to HTML
		outerHTML, err := goquery.OuterHtml(doc.Selection)
		checkError(err)

		// Gzip it again and send to response writer
		gzipWriter := gzip.NewWriter(w)
		if _, err := gzipWriter.Write([]byte(outerHTML)); err != nil {
			log.Printf("error writting gzip file: %s", err)
		}
		gzipWriter.Flush()
		return
	}

	// Serve content
	if _, err := w.Write(content); err != nil {
		log.Printf("error writting response: %s", err)
	}
}
