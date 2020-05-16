package readability

import (
	nurl "net/url"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/go-shiori/dom"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// indexOf returns the position of the first occurrence of a
// specified  value in a string array. Returns -1 if the
// value to search for never occurs.
func indexOf(array []string, key string) int {
	for i := 0; i < len(array); i++ {
		if array[i] == key {
			return i
		}
	}
	return -1
}

// wordCount returns number of word in str.
func wordCount(str string) int {
	return len(strings.Fields(str))
}

// charCount returns number of char in str.
func charCount(str string) int {
	return utf8.RuneCountInString(str)
}

// isValidURL checks if URL is valid.
func isValidURL(s string) bool {
	_, err := nurl.ParseRequestURI(s)
	return err == nil
}

// toAbsoluteURI convert uri to absolute path based on base.
// However, if uri is prefixed with hash (#), the uri won't be changed.
func toAbsoluteURI(uri string, base *nurl.URL) string {
	if uri == "" || base == nil {
		return ""
	}

	// If it is hash tag, return as it is
	if strings.HasPrefix(uri, "#") {
		return uri
	}

	// If it is data URI, return as it is
	if strings.HasPrefix(uri, "data:") {
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

	return base.ResolveReference(tmp).String()
}

// renderToFile ender an element and save it to file.
// It will panic if it fails to create destination file.
func renderToFile(element *html.Node, filename string) {
	dstFile, err := os.Create(filename)
	if err != nil {
		logrus.Fatalln("failed to create file:", err)
	}
	defer dstFile.Close()
	html.Render(dstFile, element)
}

func parseHTMLString(str string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader(str))
	if err != nil {
		return nil, err
	}

	body := dom.GetElementsByTagName(doc, "body")[0]
	return body, nil
}
