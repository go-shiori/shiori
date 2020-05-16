package processor

import (
	"fmt"
	"io"
	nurl "net/url"
	"regexp"
	"strings"
)

var (
	rxHTTPScheme    = regexp.MustCompile(`(?i)^https?:\/{2}`)
	rxRepeatedStrip = regexp.MustCompile(`(?i)-+`)
	rxTrailingSlash = regexp.MustCompile(`(?i)/+$`)
)

// Request is struct that contains data that want to be processed.
type Request struct {
	Reader io.Reader
	URL    string
}

// Resource is struct that contains URL for downloading
// and archiving a resource.
type Resource struct {
	Name    string
	URL     string
	Content []byte
	IsEmbed bool
}

func createResource(content []byte, url string, baseURL *nurl.URL) (Resource, error) {
	// Make sure URL has a valid scheme
	url = strings.TrimSpace(url)
	if url == "" || strings.Contains(url, ":") && !rxHTTPScheme.MatchString(url) {
		return Resource{}, fmt.Errorf("invalid url")
	}

	// Convert URL to absolute URL
	if baseURL != nil {
		url = createAbsoluteURL(url, baseURL)
	}

	url = rxTrailingSlash.ReplaceAllString(url, "")
	url = strings.ReplaceAll(url, " ", "+")

	// Create resource name
	resourceName := url

	// Some URL have its query or path escaped, e.g. Wikipedia and Dev.to.
	// For example, Wikipedia's stylesheet looks like this :
	//   load.php?lang=en&modules=ext.3d.styles%7Cext.cite.styles%7Cext.uls.interlanguage
	// However, when browser download it, it will be registered as unescaped query :
	//   load.php?lang=en&modules=ext.3d.styles|ext.cite.styles|ext.uls.interlanguage
	// So, for archival URL, we need to unescape the query and path first.
	tmp, err := nurl.Parse(url)
	if err == nil {
		unescapedQuery, _ := nurl.QueryUnescape(tmp.RawQuery)
		if unescapedQuery != "" {
			tmp.RawQuery = unescapedQuery
		}

		resourceName = tmp.String()
		resourceName = strings.Replace(resourceName, tmp.EscapedPath(), tmp.Path, 1)
	}

	resourceName = strings.ReplaceAll(resourceName, "://", "/")
	resourceName = strings.ReplaceAll(resourceName, ":", "-")
	resourceName = strings.ReplaceAll(resourceName, "?", "-")
	resourceName = strings.ReplaceAll(resourceName, "#", "-")
	resourceName = strings.ReplaceAll(resourceName, "/", "-")
	resourceName = strings.ReplaceAll(resourceName, " ", "-")
	resourceName = rxRepeatedStrip.ReplaceAllString(resourceName, "-")

	return Resource{
		Name:    resourceName,
		URL:     url,
		Content: content,
	}, nil
}

// createAbsoluteURL convert url to absolute path based on base.
// However, if uri is prefixed with hash (#), the uri won't be changed.
func createAbsoluteURL(uri string, base *nurl.URL) string {
	if uri == "" || base == nil {
		return ""
	}

	// If it is hash tag, return as it is
	if uri[:1] == "#" {
		return uri
	}

	// If it is already an absolute URL, return as it is
	tmp, err := nurl.ParseRequestURI(uri)
	if err == nil && tmp.Scheme != "" && tmp.Hostname() != "" {
		cleanURL(tmp)
		return tmp.String()
	}

	// Otherwise, resolve against base URI.
	tmp, err = nurl.Parse(uri)
	if err != nil {
		return uri
	}

	cleanURL(tmp)
	return base.ResolveReference(tmp).String()
}

// cleanURL removes fragment (#fragment) and UTM queries from URL
func cleanURL(url *nurl.URL) {
	queries := url.Query()

	for key := range queries {
		if strings.HasPrefix(key, "utm_") {
			queries.Del(key)
		}
	}

	url.Fragment = ""
	url.RawQuery = queries.Encode()
}
