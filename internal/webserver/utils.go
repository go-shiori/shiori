package webserver

import (
	"html/template"
	"io"
	"net"
	"net/http"
	nurl "net/url"
	"os"
	"regexp"
	"strings"
	"syscall"
)

var rxRepeatedStrip = regexp.MustCompile(`(?i)-+`)

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
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

func createTemplate(filename string, funcMap template.FuncMap) (*template.Template, error) {
	// Open file
	src, err := assets.Open(filename)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Read file content
	srcContent, err := io.ReadAll(src)
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
