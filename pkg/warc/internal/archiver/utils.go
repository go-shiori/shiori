package archiver

import (
	nurl "net/url"
	"strings"
)

func clearUTMParams(url *nurl.URL) {
	queries := url.Query()

	for key := range queries {
		if strings.HasPrefix(key, "utm_") {
			queries.Del(key)
		}
	}

	url.RawQuery = queries.Encode()
}

// toAbsoluteURI convert uri to absolute path based on base.
// However, if uri is prefixed with hash (#), the uri won't be changed.
func toAbsoluteURI(uri string, base *nurl.URL) string {
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
		return uri
	}

	// Otherwise, resolve against base URI.
	tmp, err = nurl.Parse(uri)
	if err != nil {
		return uri
	}

	clearUTMParams(tmp)
	return base.ResolveReference(tmp).String()
}
