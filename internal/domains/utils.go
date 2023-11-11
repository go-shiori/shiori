package domains

import (
	"net/url"
	"os"
	"regexp"
	"strings"
)

var rxRepeatedStrip = regexp.MustCompile(`(?i)-+`)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// getArchiveFileBasename converts an URL into an archival name.
func getArchiveFileBasename(src string) string {
	archivalURL := src

	// Some URL have its query or path escaped, e.g. Wikipedia and Dev.to.
	// For example, Wikipedia's stylesheet looks like this :
	//   load.php?lang=en&modules=ext.3d.styles%7Cext.cite.styles%7Cext.uls.interlanguage
	// However, when browser download it, it will be registered as unescaped query :
	//   load.php?lang=en&modules=ext.3d.styles|ext.cite.styles|ext.uls.interlanguage
	// So, for archival URL, we need to unescape the query and path first.
	tmp, err := url.Parse(src)
	if err == nil {
		unescapedQuery, _ := url.QueryUnescape(tmp.RawQuery)
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
