package archiver

import (
	nurl "net/url"
	"regexp"
	"strings"
)

var (
	rxHTTPScheme    = regexp.MustCompile(`(?i)^https?:\/{2}`)
	rxTrailingSlash = regexp.MustCompile(`(?i)/+$`)
	rxRepeatedStrip = regexp.MustCompile(`(?i)-+`)
)

// ResourceURL is strcut that contains URL for downloading
// and archiving a resource.
type ResourceURL struct {
	DownloadURL string
	ArchivalURL string
	Parent      string
	IsEmbedded  bool
}

// ToResourceURL generates an uri into a Resource URL.
func ToResourceURL(uri string, base *nurl.URL) ResourceURL {
	// Make sure URL has a valid scheme
	uri = strings.TrimSpace(uri)
	switch {
	case uri == "",
		strings.Contains(uri, ":") && !rxHTTPScheme.MatchString(uri):
		return ResourceURL{}
	}

	// Create download URL
	downloadURL := toAbsoluteURI(uri, base)
	downloadURL = rxTrailingSlash.ReplaceAllString(downloadURL, "")
	downloadURL = strings.ReplaceAll(downloadURL, " ", "+")

	// Create archival URL
	archivalURL := downloadURL

	// Some URL have its query escaped.
	// For example, Wikipedia's stylesheet looks like this :
	//   load.php?lang=en&modules=ext.3d.styles%7Cext.cite.styles%7Cext.uls.interlanguage
	// However, when browser download it, it will be registered as unescaped query :
	//   load.php?lang=en&modules=ext.3d.styles|ext.cite.styles|ext.uls.interlanguage
	// So, for archival URL, we need to unescape the query first.
	tmp, err := nurl.Parse(downloadURL)
	if err == nil {
		newQuery, _ := nurl.QueryUnescape(tmp.RawQuery)
		if newQuery != "" {
			tmp.RawQuery = newQuery
			archivalURL = tmp.String()
		}
	}

	archivalURL = strings.Replace(archivalURL, "://", "/", 1)
	archivalURL = strings.ReplaceAll(archivalURL, "?", "-")
	archivalURL = strings.ReplaceAll(archivalURL, "#", "-")
	archivalURL = strings.ReplaceAll(archivalURL, "/", "-")
	archivalURL = strings.ReplaceAll(archivalURL, " ", "-")
	archivalURL = rxRepeatedStrip.ReplaceAllString(archivalURL, "-")

	return ResourceURL{
		DownloadURL: downloadURL,
		ArchivalURL: archivalURL,
		Parent:      base.String(),
	}
}
