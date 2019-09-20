package core

import (
	"fmt"
	nurl "net/url"
	"strings"
)

// RemoveUTMParams removes the UTM parameters from URL.
func RemoveUTMParams(url string) (string, error) {
	// Parse string URL
	tmp, err := nurl.Parse(url)
	if err != nil || tmp.Scheme == "" || tmp.Hostname() == "" {
		return url, fmt.Errorf("URL is not valid")
	}

	// Remove UTM queries
	queries := tmp.Query()
	for key := range queries {
		if strings.HasPrefix(key, "utm_") {
			queries.Del(key)
		}
	}

	tmp.Fragment = ""
	tmp.RawQuery = queries.Encode()
	return tmp.String(), nil
}
