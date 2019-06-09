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

	// Create archive URL
	downloadURL := toAbsoluteURI(uri, base)
	downloadURL = rxTrailingSlash.ReplaceAllString(downloadURL, "")
	downloadURL = strings.ReplaceAll(downloadURL, " ", "+")

	archivalURL := strings.Replace(downloadURL, "://", "/", 1)
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
