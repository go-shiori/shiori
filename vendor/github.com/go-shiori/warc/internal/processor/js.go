package processor

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	nurl "net/url"
	"path"
	"regexp"
	"strings"

	"github.com/tdewolff/parse/js"
)

var (
	rxJSContentType = regexp.MustCompile(`(?i)(text|application)/(java|ecma)script`)
)

// processJavascript extract resource URLs from the specified JS input.
// Returns the new rules with all URLs updated to the archival link.
func processJS(input io.Reader, baseURL *nurl.URL) (string, []Resource) {
	// Prepare buffers
	buffer := bytes.NewBuffer(nil)

	// Scan JS file and process the resource's URL
	lexer := js.NewLexer(input)
	subResources := []Resource{}

	for {
		token, bt := lexer.Next()

		// Check for error
		if token == js.ErrorToken {
			break
		}

		// If it's not a string, just write it to buffer as it is
		if token != js.StringToken {
			buffer.Write(bt)
			continue
		}

		// Process the string.
		// Unlike CSS, JS doesn't have it's own URL token. So, we can only guess whether
		// a string is URL or not. There are several criteria to decide if it's URL :
		// - It surrounded by `url()` just like CSS
		// - It started with http(s):// for absolute URL
		// - It started with slash (/) for relative URL
		// -
		// If it doesn't fulfill any of criteria above, just write it as it is.

		text := string(bt)
		text = strings.TrimSpace(text)
		text = strings.Trim(text, `'`)
		text = strings.Trim(text, `"`)

		var err error
		newURL := text
		subRes := Resource{}

		if strings.HasPrefix(text, "url(") {
			cssURL := rxStyleURL.ReplaceAllString(text, "$1")
			cssURL = strings.TrimSpace(cssURL)
			cssURL = strings.Trim(cssURL, `'`)
			cssURL = strings.Trim(cssURL, `"`)

			subRes, err = createResource(nil, cssURL, baseURL)
			if err != nil {
				buffer.Write(bt)
				continue
			}

			newURL = fmt.Sprintf("\"url('%s')\"", subRes.Name)
		} else if strings.HasPrefix(text, "/") || rxHTTPScheme.MatchString(text) {
			subRes, err = createResource(nil, text, baseURL)
			if err != nil {
				buffer.Write(bt)
				continue
			}

			tmp, err := nurl.Parse(subRes.URL)
			if err != nil {
				buffer.Write(bt)
				continue
			}

			ext := path.Ext(tmp.Path)
			cType := mime.TypeByExtension(ext)

			switch {
			case rxJSContentType.MatchString(cType),
				strings.Contains(cType, "text/css"),
				strings.Contains(cType, "image/"),
				strings.Contains(cType, "audio/"),
				strings.Contains(cType, "video/"):
			default:
				buffer.Write(bt)
				continue
			}

			newURL = fmt.Sprintf("\"%s\"", subRes.Name)
		} else {
			buffer.Write(bt)
			continue
		}

		buffer.WriteString(newURL)
		subResources = append(subResources, subRes)
	}

	// Return the new rule after all URL has been processed
	return buffer.String(), subResources
}
