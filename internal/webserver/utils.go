package webserver

import (
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"net"
	"net/http"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
)

var rxRepeatedStrip = regexp.MustCompile(`(?i)-+`)

func serveFile(w http.ResponseWriter, filePath string, cache bool) error {
	// Open file
	src, err := assets.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Cache this file if needed
	if cache {
		info, err := src.Stat()
		if err != nil {
			return err
		}

		etag := fmt.Sprintf(`W/"%x-%x"`, info.ModTime().Unix(), info.Size())
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "max-age=86400")
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	// Set content type
	ext := fp.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}

	// Serve file
	_, err = io.Copy(w, src)
	return err
}

func createRedirectURL(newPath, previousPath string) string {
	urlQueries := nurl.Values{}
	urlQueries.Set("dst", previousPath)

	redirectURL, _ := nurl.Parse(newPath)
	redirectURL.RawQuery = urlQueries.Encode()
	return redirectURL.String()
}

func redirectPage(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, url, 301)
}

func assetExists(filePath string) bool {
	f, err := assets.Open(filePath)
	if f != nil {
		f.Close()
	}
	return err == nil || !os.IsNotExist(err)
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return !os.IsNotExist(err) && !info.IsDir()
}

func clearUTMParams(url *nurl.URL) {
	queries := url.Query()

	for key := range queries {
		if strings.HasPrefix(key, "utm_") {
			queries.Del(key)
		}
	}

	url.RawQuery = queries.Encode()
}

func downloadBookImage(url, dstPath string, timeout time.Duration) error {
	// Fetch data from URL
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure it's JPG or PNG image
	cp := resp.Header.Get("Content-Type")
	if !strings.Contains(cp, "image/jpeg") && !strings.Contains(cp, "image/png") {
		return fmt.Errorf("%s is not a supported image", url)
	}

	// At this point, the download has finished successfully.
	// Prepare destination file.
	err = os.MkdirAll(fp.Dir(dstPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create image dir: %v", err)
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %v", err)
	}
	defer dstFile.Close()

	// Parse image and process it.
	// If image is smaller than 600x400 or its ratio is less than 4:3, resize.
	// Else, save it as it is.
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse image %s: %v", url, err)
	}

	imgRect := img.Bounds()
	imgWidth := imgRect.Dx()
	imgHeight := imgRect.Dy()
	imgRatio := float64(imgWidth) / float64(imgHeight)

	if imgWidth >= 600 && imgHeight >= 400 && imgRatio > 1.3 {
		err = jpeg.Encode(dstFile, img, nil)
	} else {
		// Create background
		bg := image.NewNRGBA(imgRect)
		draw.Draw(bg, imgRect, image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(bg, imgRect, img, image.Point{}, draw.Over)

		bg = imaging.Fill(bg, 600, 400, imaging.Center, imaging.Lanczos)
		bg = imaging.Blur(bg, 150)
		bg = imaging.AdjustBrightness(bg, 30)

		// Create foreground
		fg := imaging.Fit(img, 600, 400, imaging.Lanczos)

		// Merge foreground and background
		bgRect := bg.Bounds()
		fgRect := fg.Bounds()
		fgPosition := image.Point{
			X: bgRect.Min.X - int(math.Round(float64(bgRect.Dx()-fgRect.Dx())/2)),
			Y: bgRect.Min.Y - int(math.Round(float64(bgRect.Dy()-fgRect.Dy())/2)),
		}

		draw.Draw(bg, bgRect, fg, fgPosition, draw.Over)

		// Save to file
		err = jpeg.Encode(dstFile, bg, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to save image %s: %v", url, err)
	}

	return nil
}

func createTemplate(filename string, funcMap template.FuncMap) (*template.Template, error) {
	// Open file
	src, err := assets.Open(filename)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Read file content
	srcContent, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	// Create template
	return template.New(filename).Delims("$$", "$$").Funcs(funcMap).Parse(string(srcContent))
}

// getArchivalName converts an URL into an archival name.
func getArchivalName(src string) string {
	archivalURL := src

	// Some URL have its query or path escaped, e.g. Wikipedia and Dev.to.
	// For example, Wikipedia's stylesheet looks like this :
	//   load.php?lang=en&modules=ext.3d.styles%7Cext.cite.styles%7Cext.uls.interlanguage
	// However, when browser download it, it will be registered as unescaped query :
	//   load.php?lang=en&modules=ext.3d.styles|ext.cite.styles|ext.uls.interlanguage
	// So, for archival URL, we need to unescape the query and path first.
	tmp, err := nurl.Parse(src)
	if err == nil {
		unescapedQuery, _ := nurl.QueryUnescape(tmp.RawQuery)
		if unescapedQuery != "" {
			tmp.RawQuery = unescapedQuery
		}

		archivalURL = tmp.String()
		archivalURL = strings.Replace(archivalURL, tmp.EscapedPath(), tmp.Path, 1)
	}

	archivalURL = strings.ReplaceAll(archivalURL, "://", "/")
	archivalURL = strings.ReplaceAll(archivalURL, "?", "-")
	archivalURL = strings.ReplaceAll(archivalURL, "#", "-")
	archivalURL = strings.ReplaceAll(archivalURL, "/", "-")
	archivalURL = strings.ReplaceAll(archivalURL, " ", "-")
	archivalURL = rxRepeatedStrip.ReplaceAllString(archivalURL, "-")

	return archivalURL
}

func checkError(err error) {
	if err == nil {
		return
	}

	// Check for a broken connection, as it is not really a
	// condition that warrants a panic stack trace.
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if se.Err == syscall.EPIPE || se.Err == syscall.ECONNRESET {
				return
			}
		}
	}

	panic(err)
}
