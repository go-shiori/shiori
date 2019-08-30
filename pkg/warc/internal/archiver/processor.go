package archiver

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	nurl "net/url"
	"path"
	"regexp"
	"strings"

	"github.com/tdewolff/parse/v2/css"
	"github.com/tdewolff/parse/v2/js"
	"golang.org/x/net/html"
)

// ProcessResult is the result from content processing.
type ProcessResult struct {
	Name        string
	ContentType string
	Content     []byte
}

var (
	rxImageMeta       = regexp.MustCompile(`(?i)image|thumbnail`)
	rxLazyImageSrcset = regexp.MustCompile(`(?i)\.(jpg|jpeg|png|webp)\s+\d`)
	rxLazyImageSrc    = regexp.MustCompile(`(?i)^\s*\S+\.(jpg|jpeg|png|webp)\S*\s*$`)
	rxStyleURL        = regexp.MustCompile(`(?i)^url\((.+)\)$`)
	rxJSContentType   = regexp.MustCompile(`(?i)(text|application)/(java|ecma)script`)
)

// ProcessHTMLFile process HTML file that submitted through the io.Reader.
func (arc *Archiver) ProcessHTMLFile(res ResourceURL, input io.Reader) (result ProcessResult, resources []ResourceURL, err error) {
	// Parse HTML document
	doc, err := html.Parse(input)
	if err != nil {
		return ProcessResult{}, nil, fmt.Errorf("failed to parse HTML for %s: %v", res.DownloadURL, err)
	}

	// Parse URL
	parsedURL, err := nurl.ParseRequestURI(res.DownloadURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Hostname() == "" {
		return ProcessResult{}, nil, fmt.Errorf("url %s is not valid", res.DownloadURL)
	}

	// TODO: I'm still not really sure, but IMHO it's safer to disable Javascript
	// Ideally, we only want to remove XHR request by using function disableXHR(doc).
	// Unfortunately, the result is not that good for now, so it's still not used.
	removeNodes(getElementsByTagName(doc, "script"), nil)

	// Convert lazy loaded image to normal
	fixLazyImages(doc)

	// Convert hyperlinks rith relative URL
	fixRelativeURIs(doc, parsedURL)

	// Extract resources from each nodes
	for _, node := range getElementsByTagName(doc, "*") {
		// First extract resources from inline style
		cssResources := extractInlineCSS(node, parsedURL)
		resources = append(resources, cssResources...)

		// Next extract resources from tag's specific attribute
		nodeResources := []ResourceURL{}
		switch tagName(node) {
		case "style":
			nodeResources = extractStyleTag(node, parsedURL)
		case "script":
			nodeResources = extractScriptTag(node, parsedURL)
		case "meta":
			nodeResources = extractMetaTag(node, parsedURL)
		case "img", "picture", "figure", "video", "audio", "source":
			nodeResources = extractMediaTag(node, parsedURL)
		case "link":
			nodeResources = extractGenericTag(node, "href", parsedURL)
		case "iframe":
			nodeResources = extractGenericTag(node, "src", parsedURL)
		case "object":
			nodeResources = extractGenericTag(node, "data", parsedURL)
		default:
			continue
		}
		resources = append(resources, nodeResources...)
	}

	// Get outer HTML of the doc
	result = ProcessResult{
		Name:    res.ArchivalURL,
		Content: outerHTML(doc),
	}

	return result, resources, nil
}

// ProcessCSSFile process CSS file that submitted through the io.Reader.
func (arc *Archiver) ProcessCSSFile(res ResourceURL, input io.Reader) (result ProcessResult, resources []ResourceURL, err error) {
	// Parse URL
	parsedURL, err := nurl.ParseRequestURI(res.DownloadURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Hostname() == "" {
		return ProcessResult{}, nil, fmt.Errorf("url %s is not valid", res.DownloadURL)
	}

	// Extract CSS rules
	rules, resources := processCSS(input, parsedURL)

	result = ProcessResult{
		Name:    res.ArchivalURL,
		Content: []byte(rules),
	}

	return result, resources, nil
}

// ProcessOtherFile process files that not HTML, JS or CSS that submitted through the io.Reader.
func (arc *Archiver) ProcessOtherFile(res ResourceURL, input io.Reader) (result ProcessResult, err error) {
	// Copy data to buffer
	buffer := bytes.NewBuffer(nil)

	_, err = io.Copy(buffer, input)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed to copy data: %v", err)
	}

	// Create result
	result = ProcessResult{
		Name:    res.ArchivalURL,
		Content: buffer.Bytes(),
	}

	return result, nil
}

func disableXHR(doc *html.Node) {
	var head *html.Node
	heads := getElementsByTagName(doc, "head")
	if len(heads) > 0 {
		head = heads[0]
	} else {
		head = createElement("head")
		prependChild(doc, head)
	}

	xhrDisabler := `
	fetch = new Promise();

	XMLHttpRequest = function() {};
	XMLHttpRequest.prototype = {
		open: function(){},
		send: function(){},
		abort: function(){},
		setRequestHeader: function(){},
		overrideMimeType: function(){},
		getResponseHeaders(): function(){},
		getAllResponseHeaders(): function(){},
	};`

	script := createElement("script")
	scriptContent := createTextNode(xhrDisabler)
	prependChild(script, scriptContent)
	prependChild(head, script)
}

// fixRelativeURIs converts each <a> in the given element
// to an absolute URI, ignoring #ref URIs.
func fixRelativeURIs(doc *html.Node, pageURL *nurl.URL) {
	links := getAllNodesWithTag(doc, "a")
	forEachNode(links, func(link *html.Node, _ int) {
		href := getAttribute(link, "href")
		if href == "" {
			return
		}

		// Replace links with javascript: URIs with text content,
		// since they won't work after scripts have been removed
		// from the page.
		if strings.HasPrefix(href, "javascript:") {
			text := createTextNode(textContent(link))
			replaceNode(link, text)
		} else {
			newHref := toAbsoluteURI(href, pageURL)
			if newHref == "" {
				removeAttribute(link, "href")
			} else {
				setAttribute(link, "href", newHref)
			}
		}
	})
}

// fixLazyImages convert images and figures that have properties like data-src into
// images that can be loaded without JS.
func fixLazyImages(root *html.Node) {
	imageNodes := getAllNodesWithTag(root, "img", "picture", "figure")
	forEachNode(imageNodes, func(elem *html.Node, _ int) {
		src := getAttribute(elem, "src")
		srcset := getAttribute(elem, "srcset")
		nodeTag := tagName(elem)
		nodeClass := className(elem)

		if (src == "" && srcset == "") || strings.Contains(strings.ToLower(nodeClass), "lazy") {
			for i := 0; i < len(elem.Attr); i++ {
				attr := elem.Attr[i]
				if attr.Key == "src" || attr.Key == "srcset" {
					continue
				}

				copyTo := ""
				if rxLazyImageSrcset.MatchString(attr.Val) {
					copyTo = "srcset"
				} else if rxLazyImageSrc.MatchString(attr.Val) {
					copyTo = "src"
				}

				if copyTo == "" {
					continue
				}

				if nodeTag == "img" || nodeTag == "picture" {
					// if this is an img or picture, set the attribute directly
					setAttribute(elem, copyTo, attr.Val)
				} else if nodeTag == "figure" && len(getAllNodesWithTag(elem, "img", "picture")) == 0 {
					// if the item is a <figure> that does not contain an image or picture,
					// create one and place it inside the figure see the nytimes-3
					// testcase for an example
					img := createElement("img")
					setAttribute(img, copyTo, attr.Val)
					appendChild(elem, img)
				}
			}
		}
	})
}

// extractInlineCSS extract archive's resource from the CSS rules inside
// style attribute. Once finished, all CSS URLs in the style attribute
// will be updated to use the archival URL.
func extractInlineCSS(node *html.Node, pageURL *nurl.URL) []ResourceURL {
	// Make sure this node has inline style
	styleAttr := getAttribute(node, "style")
	if styleAttr == "" {
		return nil
	}

	// Extract resource URLs from the inline style
	// and update the CSS rules accordingly.
	reader := strings.NewReader(styleAttr)
	newStyleAttr, resources := processCSS(reader, pageURL)
	setAttribute(node, "style", newStyleAttr)

	return resources
}

// extractStyleTag extract archive's resource from inside a <style> tag.
// Once finished, all CSS URLs will be updated to use the archival URL.
func extractStyleTag(node *html.Node, pageURL *nurl.URL) []ResourceURL {
	// Extract CSS rules from <style>
	rules := textContent(node)
	rules = strings.TrimSpace(rules)
	if rules == "" {
		return nil
	}

	// Extract resource URLs from the rules and update it accordingly.
	reader := strings.NewReader(rules)
	newRules, resources := processCSS(reader, pageURL)
	setTextContent(node, newRules)

	return resources
}

// extractScriptTag extract archive's resource from inside a <script> tag.
// Once finished, all URLs inside it will be updated to use the archival URL.
func extractScriptTag(node *html.Node, pageURL *nurl.URL) []ResourceURL {
	// Also get the URL from `src` attribute
	resources := extractGenericTag(node, "src", pageURL)

	// Extract JS code from the <script> itself
	script := textContent(node)
	script = strings.TrimSpace(script)
	if script == "" {
		return resources
	}

	reader := strings.NewReader(script)
	newScript, scriptResources := processJS(reader, pageURL)
	setTextContent(node, newScript)
	resources = append(resources, scriptResources...)

	return resources
}

// extractMetaTag extract archive's resource from inside a <meta>.
// Normally, <meta> doesn't have any resource URLs. However, as
// social media come and grow, a new metadata is added to contain
// the hero image for a web page, e.g. og:image, twitter:image, etc.
// Once finished, all URLs in <meta> for image will be updated
// to use the archival URL.
func extractMetaTag(node *html.Node, pageURL *nurl.URL) []ResourceURL {
	// Get the needed attributes
	name := getAttribute(node, "name")
	property := getAttribute(node, "property")
	content := getAttribute(node, "content")

	// If this <meta> is not for image, don't process it
	if !rxImageMeta.MatchString(name + " " + property) {
		return nil
	}

	// If URL is not valid, skip
	tmp, err := nurl.ParseRequestURI(content)
	if err != nil || tmp.Scheme == "" || tmp.Hostname() == "" {
		return nil
	}

	// Create archive resource and update the href URL
	res := ToResourceURL(content, pageURL)
	if res.ArchivalURL == "" {
		return nil
	}

	setAttribute(node, "content", res.ArchivalURL)
	return []ResourceURL{res}
}

// extractMediaTag extract resource from inside a media tag e.g.
// <img>, <video>, <audio>, <source>. Once finished, all URLs will be
// updated to use the archival URL.
func extractMediaTag(node *html.Node, pageURL *nurl.URL) []ResourceURL {
	// Get the needed attributes
	src := getAttribute(node, "src")
	poster := getAttribute(node, "poster")
	strSrcSets := getAttribute(node, "srcset")

	// Create initial resources
	resources := []ResourceURL{}

	// Save `src` and `poster` to resources
	if src != "" {
		res := ToResourceURL(src, pageURL)
		if res.ArchivalURL != "" {
			setAttribute(node, "src", res.ArchivalURL)
			resources = append(resources, res)
		}
	}

	if poster != "" {
		res := ToResourceURL(poster, pageURL)
		if res.ArchivalURL != "" {
			setAttribute(node, "poster", res.ArchivalURL)
			resources = append(resources, res)
		}
	}

	// Split srcset by comma, then process it like any URLs
	srcSets := strings.Split(strSrcSets, ",")
	for i, srcSet := range srcSets {
		srcSet = strings.TrimSpace(srcSet)
		parts := strings.SplitN(srcSet, " ", 2)
		if parts[0] == "" {
			continue
		}

		res := ToResourceURL(parts[0], pageURL)
		if res.ArchivalURL == "" {
			continue
		}

		srcSets[i] = strings.Replace(srcSets[i], parts[0], res.ArchivalURL, 1)
		resources = append(resources, res)
	}

	if len(srcSets) > 0 {
		setAttribute(node, "srcset", strings.Join(srcSets, ","))
	}

	return resources
}

// extractGenericTag extract resource from specified attribute.
// This method is used for tags where the URL is obviously exist in
// the tag, without any additional process needed to extract it.
// For example is <link> with its href, <object> with its data, etc.
// Once finished, the URL attribute will be updated to use the
// archival URL.
func extractGenericTag(node *html.Node, attrName string, pageURL *nurl.URL) []ResourceURL {
	// Get the needed attributes
	attrValue := getAttribute(node, attrName)
	if attrValue == "" {
		return nil
	}

	res := ToResourceURL(attrValue, pageURL)
	if res.ArchivalURL == "" {
		return nil
	}

	// If this node is iframe, mark it as embedded
	if tagName(node) == "iframe" {
		res.IsEmbedded = true
	}

	setAttribute(node, attrName, res.ArchivalURL)
	return []ResourceURL{res}
}

// processCSSRules extract resource URLs from the specified CSS input.
// Returns the new rules with all CSS URLs updated to the archival link.
func processCSS(input io.Reader, baseURL *nurl.URL) (string, []ResourceURL) {
	// Prepare buffers
	buffer := bytes.NewBuffer(nil)

	// Scan CSS file and process the resource's URL
	lexer := css.NewLexer(input)
	resources := []ResourceURL{}

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

		// Save the CSS URL and replace it with archival URL
		res := ToResourceURL(cssURL, baseURL)
		if res.ArchivalURL == "" {
			buffer.Write(bt)
			continue
		}

		cssURL = `url("` + res.ArchivalURL + `")`
		buffer.WriteString(cssURL)
		resources = append(resources, res)
	}

	// Return the new rule after all URL has been processed
	return buffer.String(), resources
}

// processJavascript extract resource URLs from the specified JS input.
// Returns the new rules with all URLs updated to the archival link.
func processJS(input io.Reader, baseURL *nurl.URL) (string, []ResourceURL) {
	// Prepare buffers
	buffer := bytes.NewBuffer(nil)

	// Scan JS file and process the resource's URL
	lexer := js.NewLexer(input)
	resources := []ResourceURL{}

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
		var res ResourceURL
		var newURL string

		text := string(bt)
		text = strings.TrimSpace(text)
		text = strings.Trim(text, `'`)
		text = strings.Trim(text, `"`)

		if strings.HasPrefix(text, "url(") {
			cssURL := rxStyleURL.ReplaceAllString(text, "$1")
			cssURL = strings.TrimSpace(cssURL)
			cssURL = strings.Trim(cssURL, `'`)
			cssURL = strings.Trim(cssURL, `"`)

			res = ToResourceURL(cssURL, baseURL)
			newURL = fmt.Sprintf("\"url('%s')\"", res.ArchivalURL)
		} else if strings.HasPrefix(text, "/") || rxHTTPScheme.MatchString(text) {
			res = ToResourceURL(text, baseURL)

			tmp, err := nurl.Parse(res.DownloadURL)
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

			newURL = fmt.Sprintf("\"%s\"", res.ArchivalURL)
		} else {
			buffer.Write(bt)
			continue
		}

		if res.ArchivalURL == "" {
			continue
		}

		buffer.WriteString(newURL)
		resources = append(resources, res)
	}

	// Return the new rule after all URL has been processed
	return buffer.String(), resources
}
