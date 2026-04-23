package core

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// SingleFileMetadata holds the original URL and title extracted from a
// SingleFile-exported HTML document.
type SingleFileMetadata struct {
	URL   string
	Title string
}

var (
	// SingleFile primary marker: <!-- saved from url=(0073)https://... -->
	reSavedFromURL = regexp.MustCompile(`saved from url=\(\d+\)(\S+)`)
	// Alternate block-comment format: url: https://...
	reSFURLLine = regexp.MustCompile(`(?m)^\s*url:\s*(\S+)\s*$`)
)

// ExtractSingleFileMetadata extracts the original URL and page title from a
// SingleFile-exported HTML document. Tries in order:
//  1. <!-- saved from url=(NNNN)URL --> comment
//  2. url: URL  line in a comment block
//  3. <base href="URL">
//  4. <link rel="canonical" href="URL">
//  5. <meta property="og:url"> / <meta name="twitter:url">
//
// Title comes from <title>, falling back to <meta property="og:title">.
func ExtractSingleFileMetadata(content []byte) SingleFileMetadata {
	var meta SingleFileMetadata

	// Fast path: regex scan before full HTML parse
	if m := reSavedFromURL.FindSubmatch(content); m != nil {
		meta.URL = strings.TrimSpace(string(m[1]))
	} else if m := reSFURLLine.FindSubmatch(content); m != nil {
		meta.URL = strings.TrimSpace(string(m[1]))
	}

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return meta
	}

	var ogTitle string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch strings.ToLower(n.Data) {
			case "title":
				if n.FirstChild != nil && meta.Title == "" {
					meta.Title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "base":
				if meta.URL == "" {
					for _, a := range n.Attr {
						if a.Key == "href" && a.Val != "" {
							meta.URL = a.Val
						}
					}
				}
			case "link":
				if meta.URL == "" {
					var rel, href string
					for _, a := range n.Attr {
						switch a.Key {
						case "rel":
							rel = a.Val
						case "href":
							href = a.Val
						}
					}
					if rel == "canonical" && href != "" {
						meta.URL = href
					}
				}
			case "meta":
				var prop, name, content string
				for _, a := range n.Attr {
					switch a.Key {
					case "property":
						prop = a.Val
					case "name":
						name = a.Val
					case "content":
						content = a.Val
					}
				}
				if content != "" {
					if meta.URL == "" && (prop == "og:url" || name == "twitter:url") {
						meta.URL = content
					}
					if ogTitle == "" && prop == "og:title" {
						ogTitle = strings.TrimSpace(content)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if meta.Title == "" {
		meta.Title = ogTitle
	}

	return meta
}
