package processor

import (
	"bytes"
	"fmt"
	"io"
	nurl "net/url"
	"regexp"
	"strings"

	"github.com/tdewolff/parse/css"
)

var (
	rxStyleURL = regexp.MustCompile(`(?i)^url\((.+)\)$`)
)

// ProcessCSSFile process CSS file.
func ProcessCSSFile(req Request) (Resource, []Resource, error) {
	// Parse URL, then use it to extract CSS rules
	parsedURL, err := nurl.ParseRequestURI(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Hostname() == "" {
		return Resource{}, nil, fmt.Errorf("url %s is not valid", req.URL)
	}

	cssRules, subResources := processCSS(req.Reader, parsedURL)
	resource, err := createResource([]byte(cssRules), req.URL, nil)

	return resource, subResources, err
}

// processCSSRules extract resource URLs from the specified CSS input.
// Returns the new rules with all CSS URLs updated to the archival link.
func processCSS(input io.Reader, baseURL *nurl.URL) (string, []Resource) {
	// Prepare buffers
	buffer := bytes.NewBuffer(nil)

	// Scan CSS file and process the resource's URL
	lexer := css.NewLexer(input)
	subResources := []Resource{}

	for {
		token, bt := lexer.Next()

		// Check for error
		if token == css.ErrorToken {
			break
		}

		// If it's not an URL, just write it to buffer as it is
		if token != css.URLToken {
			buffer.Write(bt)
			continue
		}

		// Sanitize the URL by removing `url()`, quotation mark and trailing slash
		cssURL := string(bt)
		cssURL = rxStyleURL.ReplaceAllString(cssURL, "$1")
		cssURL = strings.TrimSpace(cssURL)
		cssURL = strings.Trim(cssURL, `'`)
		cssURL = strings.Trim(cssURL, `"`)

		// Create subresource from CSS URL
		subResource, err := createResource(nil, cssURL, baseURL)
		if err != nil {
			buffer.Write(bt)
			continue
		}

		// Write resource name instead of CSS URL
		buffer.WriteString(`url("` + subResource.Name + `")`)

		// Save sub resource
		subResources = append(subResources, subResource)
	}

	// Return the new rule after all URL has been processed
	return buffer.String(), subResources
}
