package processor

import (
	"fmt"
	nurl "net/url"
	"regexp"
	"strings"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

var (
	rxLazyImageSrcset = regexp.MustCompile(`(?i)\.(jpg|jpeg|png|webp)\s+\d`)
	rxLazyImageSrc    = regexp.MustCompile(`(?i)^\s*\S+\.(jpg|jpeg|png|webp)\S*\s*$`)
	rxImageMeta       = regexp.MustCompile(`(?i)image|thumbnail`)
)

// ProcessHTMLFile process HTML file.
func ProcessHTMLFile(req Request) (Resource, []Resource, error) {
	// Parse URL
	pageURL, err := nurl.ParseRequestURI(req.URL)
	if err != nil || pageURL.Scheme == "" || pageURL.Hostname() == "" {
		return Resource{}, nil, fmt.Errorf("url %s is not valid", req.URL)
	}

	// Parse HTML document
	doc, err := html.Parse(req.Reader)
	if err != nil {
		return Resource{}, nil, fmt.Errorf("failed to parse HTML for %s: %v", req.URL, err)
	}

	// TODO: I'm still not really sure, but IMHO it's safer to
	// disable Javascript. Ideally, we only want to remove XHR request
	// using disableXHR(). Unfortunately, the result is not that good for now.
	dom.RemoveNodes(dom.GetElementsByTagName(doc, "script"), nil)

	// Convert lazy loaded image to normal
	fixLazyImages(doc)

	// Convert hyperlinks with relative URL
	fixRelativeURIs(doc, pageURL)

	// Extract subresources from each nodes
	subResources := []Resource{}
	for _, node := range dom.GetElementsByTagName(doc, "*") {
		// First extract resources from inline style
		cssResources := processInlineCSS(node, pageURL)
		subResources = append(subResources, cssResources...)

		// Next extract resources from tag's specific attribute
		nodeResources := []Resource{}
		switch dom.TagName(node) {
		case "style":
			nodeResources = processStyleTag(node, pageURL)
		case "script":
			nodeResources = processScriptTag(node, pageURL)
		case "meta":
			nodeResources = processMetaTag(node, pageURL)
		case "img", "picture", "figure", "video", "audio", "source":
			nodeResources = processMediaTag(node, pageURL)
		case "link":
			nodeResources = processGenericTag(node, "href", pageURL)
		case "iframe":
			nodeResources = processGenericTag(node, "src", pageURL)
		case "object":
			nodeResources = processGenericTag(node, "data", pageURL)
		default:
			continue
		}
		subResources = append(subResources, nodeResources...)
	}

	// Return outer HTML of the doc
	outerHTML := dom.OuterHTML(doc)
	resource, err := createResource([]byte(outerHTML), req.URL, nil)

	return resource, subResources, err
}

func disableXHR(doc *html.Node) {
	var head *html.Node
	heads := dom.GetElementsByTagName(doc, "head")
	if len(heads) > 0 {
		head = heads[0]
	} else {
		head = dom.CreateElement("head")
		dom.PrependChild(doc, head)
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

	script := dom.CreateElement("script")
	scriptContent := dom.CreateTextNode(xhrDisabler)
	dom.PrependChild(script, scriptContent)
	dom.PrependChild(head, script)
}

// fixRelativeURIs converts each <a> in the given element
// to an absolute URI, ignoring #ref URIs.
func fixRelativeURIs(doc *html.Node, pageURL *nurl.URL) {
	links := dom.GetAllNodesWithTag(doc, "a")
	dom.ForEachNode(links, func(link *html.Node, _ int) {
		href := dom.GetAttribute(link, "href")
		if href == "" {
			return
		}

		// Replace links with javascript: URIs with text content,
		// since they won't work after scripts have been removed
		// from the page.
		if strings.HasPrefix(href, "javascript:") {
			text := dom.CreateTextNode(dom.TextContent(link))
			dom.ReplaceChild(link.Parent, text, link)
		} else {
			newHref := createAbsoluteURL(href, pageURL)
			if newHref == "" {
				dom.RemoveAttribute(link, "href")
			} else {
				dom.SetAttribute(link, "href", newHref)
			}
		}
	})
}

// fixLazyImages convert images and figures that have properties like
// data-src into images that can be loaded without JS.
func fixLazyImages(root *html.Node) {
	imageNodes := dom.GetAllNodesWithTag(root, "img", "picture", "figure")
	dom.ForEachNode(imageNodes, func(elem *html.Node, _ int) {
		src := dom.GetAttribute(elem, "src")
		srcset := dom.GetAttribute(elem, "srcset")
		nodeTag := dom.TagName(elem)
		nodeClass := dom.ClassName(elem)

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
					dom.SetAttribute(elem, copyTo, attr.Val)
				} else if nodeTag == "figure" && len(dom.GetAllNodesWithTag(elem, "img", "picture")) == 0 {
					// if the item is a <figure> that does not contain an image or picture,
					// create one and place it inside the figure see the nytimes-3
					// testcase for an example
					img := dom.CreateElement("img")
					dom.SetAttribute(img, copyTo, attr.Val)
					dom.AppendChild(elem, img)
				}
			}
		}
	})
}

// processInlineCSS extract subresources from the CSS rules inside
// style attribute. Once finished, all CSS URLs in the style attribute
// will be updated to use the resource name.
func processInlineCSS(node *html.Node, pageURL *nurl.URL) []Resource {
	// Make sure this node has inline style
	styleAttr := dom.GetAttribute(node, "style")
	styleAttr = strings.TrimSpace(styleAttr)
	if styleAttr == "" {
		return nil
	}

	// Extract resource URLs from the inline style
	// and update the CSS rules accordingly.
	reader := strings.NewReader(styleAttr)
	newStyleAttr, subResources := processCSS(reader, pageURL)
	dom.SetAttribute(node, "style", newStyleAttr)

	return subResources
}

// processStyleTag extract subresources from inside a <style> tag.
// Once finished, all CSS URLs will be updated to use the resource name.
func processStyleTag(styleNode *html.Node, pageURL *nurl.URL) []Resource {
	// Extract CSS rules from <style>
	rules := dom.TextContent(styleNode)
	rules = strings.TrimSpace(rules)
	if rules == "" {
		return nil
	}

	// Extract resource URLs from the rules and update it accordingly.
	reader := strings.NewReader(rules)
	newRules, subResources := processCSS(reader, pageURL)
	dom.SetTextContent(styleNode, newRules)

	return subResources
}

// processScriptTag extract archive's resource from inside a <script> tag.
// Once finished, all URLs inside it will be updated to use the resource name.
func processScriptTag(node *html.Node, pageURL *nurl.URL) []Resource {
	// Also get the URL from `src` attribute
	subResources := processGenericTag(node, "src", pageURL)

	// Extract JS code from the <script> itself
	script := dom.TextContent(node)
	script = strings.TrimSpace(script)
	if script == "" {
		return subResources
	}

	reader := strings.NewReader(script)
	newScript, scriptResources := processJS(reader, pageURL)
	dom.SetTextContent(node, newScript)

	// Merge script resources
	subResources = append(subResources, scriptResources...)
	return subResources
}

// extractMetaTag extract archive's resource from inside a <meta>.
// Normally, <meta> doesn't have any resource URLs. However, as
// social media come and grow, a new metadata is added to contain
// the hero image for a web page, e.g. og:image, twitter:image, etc.
// Once finished, all URLs in <meta> for image will be updated
// to use the resource name.
func processMetaTag(node *html.Node, pageURL *nurl.URL) []Resource {
	// Get the needed attributes
	name := dom.GetAttribute(node, "name")
	property := dom.GetAttribute(node, "property")
	content := dom.GetAttribute(node, "content")

	// If this <meta> is not for image, don't process it
	if !rxImageMeta.MatchString(name + " " + property) {
		return nil
	}

	// If URL is not valid, skip
	tmp, err := nurl.ParseRequestURI(content)
	if err != nil || tmp.Scheme == "" || tmp.Hostname() == "" {
		return nil
	}

	// Create subresource and update the URL
	subResource, err := createResource(nil, content, pageURL)
	if err != nil {
		return nil
	}

	dom.SetAttribute(node, "content", subResource.Name)
	return []Resource{subResource}
}

// processMediaTag extract resource from inside a media tag e.g.
// <img>, <video>, <audio>, <source>. Once finished, all URLs will be
// updated to use the resource name.
func processMediaTag(node *html.Node, pageURL *nurl.URL) []Resource {
	// Create initial subresources
	subResources := []Resource{}

	// Save `src` and `poster` of media to subresources
	for _, attrName := range []string{"src", "poster"} {
		attrValue := dom.GetAttribute(node, attrName)
		if attrValue == "" {
			continue
		}

		subResource, err := createResource(nil, attrValue, pageURL)
		if err != nil {
			continue
		}

		dom.SetAttribute(node, attrName, subResource.Name)
		subResources = append(subResources, subResource)
	}

	// Get `srcset`, split it by comma, then process it like any URLs
	strSrcSets := dom.GetAttribute(node, "srcset")
	srcSets := strings.Split(strSrcSets, ",")
	for i, srcSet := range srcSets {
		srcSet = strings.TrimSpace(srcSet)
		parts := strings.SplitN(srcSet, " ", 2)
		if parts[0] == "" {
			continue
		}

		subResource, err := createResource(nil, parts[0], pageURL)
		if err != nil {
			continue
		}

		srcSets[i] = strings.Replace(srcSets[i], parts[0], subResource.Name, 1)
		subResources = append(subResources, subResource)
	}

	if len(srcSets) > 0 {
		dom.SetAttribute(node, "srcset", strings.Join(srcSets, ","))
	}

	return subResources
}

// processGenericTag extract resource from specified attribute.
// This method is used for tags where the URL is obviously exist in
// the tag, without any additional process needed to extract it.
// For example is <link> with its href, <object> with its data, etc.
// Once finished, the URL attribute will be updated to use the
// resource name.
func processGenericTag(node *html.Node, attrName string, pageURL *nurl.URL) []Resource {
	// Get the needed attributes
	attrValue := dom.GetAttribute(node, attrName)
	if attrValue == "" {
		return nil
	}

	subResource, err := createResource(nil, attrValue, pageURL)
	if err != nil {
		return nil
	}

	if dom.TagName(node) == "iframe" {
		subResource.IsEmbed = true
	}

	dom.SetAttribute(node, attrName, subResource.Name)
	return []Resource{subResource}
}
