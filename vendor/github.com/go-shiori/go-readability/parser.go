package readability

import (
	"fmt"
	shtml "html"
	"io"
	"math"
	nurl "net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

// All of the regular expressions in use within readability.
// Defined up here so we don't instantiate them repeatedly in loops *.
var (
	rxUnlikelyCandidates   = regexp.MustCompile(`(?i)-ad-|ai2html|banner|breadcrumbs|combx|comment|community|cover-wrap|disqus|extra|footer|gdpr|header|legends|menu|related|remark|replies|rss|shoutbox|sidebar|skyscraper|social|sponsor|supplemental|ad-break|agegate|pagination|pager|popup|yom-remote`)
	rxOkMaybeItsACandidate = regexp.MustCompile(`(?i)and|article|body|column|content|main|shadow`)
	rxPositive             = regexp.MustCompile(`(?i)article|body|content|entry|hentry|h-entry|main|page|pagination|post|text|blog|story`)
	rxNegative             = regexp.MustCompile(`(?i)hidden|^hid$| hid$| hid |^hid |banner|combx|comment|com-|contact|foot|footer|footnote|gdpr|masthead|media|meta|outbrain|promo|related|scroll|share|shoutbox|sidebar|skyscraper|sponsor|shopping|tags|tool|widget`)
	rxExtraneous           = regexp.MustCompile(`(?i)print|archive|comment|discuss|e[\-]?mail|share|reply|all|login|sign|single|utility`)
	rxByline               = regexp.MustCompile(`(?i)byline|author|dateline|writtenby|p-author`)
	rxReplaceFonts         = regexp.MustCompile(`(?i)<(/?)font[^>]*>`)
	rxNormalize            = regexp.MustCompile(`(?i)\s{2,}`)
	rxVideos               = regexp.MustCompile(`(?i)//(www\.)?((dailymotion|youtube|youtube-nocookie|player\.vimeo|v\.qq)\.com|(archive|upload\.wikimedia)\.org|player\.twitch\.tv)`)
	rxNextLink             = regexp.MustCompile(`(?i)(next|weiter|continue|>([^\|]|$)|»([^\|]|$))`)
	rxPrevLink             = regexp.MustCompile(`(?i)(prev|earl|old|new|<|«)`)
	rxWhitespace           = regexp.MustCompile(`(?i)^\s*$`)
	rxHasContent           = regexp.MustCompile(`(?i)\S$`)
	rxPropertyPattern      = regexp.MustCompile(`(?i)\s*(dc|dcterm|og|twitter)\s*:\s*(author|creator|description|title|site_name|image\S*)\s*`)
	rxNamePattern          = regexp.MustCompile(`(?i)^\s*(?:(dc|dcterm|og|twitter|weibo:(article|webpage))\s*[\.:]\s*)?(author|creator|description|title|site_name|image)\s*$`)
	rxTitleSeparator       = regexp.MustCompile(`(?i) [\|\-\\/>»] `)
	rxTitleHierarchySep    = regexp.MustCompile(`(?i) [\\/>»] `)
	rxTitleRemoveFinalPart = regexp.MustCompile(`(?i)(.*)[\|\-\\/>»] .*`)
	rxTitleRemove1stPart   = regexp.MustCompile(`(?i)[^\|\-\\/>»]*[\|\-\\/>»](.*)`)
	rxTitleAnySeparator    = regexp.MustCompile(`(?i)[\|\-\\/>»]+`)
	rxDisplayNone          = regexp.MustCompile(`(?i)display\s*:\s*none`)
	rxSentencePeriod       = regexp.MustCompile(`(?i)\.( |$)`)
	rxShareElements        = regexp.MustCompile(`(?i)(\b|_)(share|sharedaddy)(\b|_)`)
	rxFaviconSize          = regexp.MustCompile(`(?i)(\d+)x(\d+)`)
	rxLazyImageSrcset      = regexp.MustCompile(`(?i)\.(jpg|jpeg|png|webp)\s+\d`)
	rxLazyImageSrc         = regexp.MustCompile(`(?i)^\s*\S+\.(jpg|jpeg|png|webp)\S*\s*$`)
	rxImgExtensions        = regexp.MustCompile(`(?i)\.(jpg|jpeg|png|webp)`)
	rxSrcsetURL            = regexp.MustCompile(`(?i)(\S+)(\s+[\d.]+[xw])?(\s*(?:,|$))`)
	rxB64DataURL           = regexp.MustCompile(`(?i)^data:\s*([^\s;,]+)\s*;\s*base64\s*`)
)

// Constants that used by readability.
var (
	divToPElems                  = []string{"a", "blockquote", "dl", "div", "img", "ol", "p", "pre", "table", "ul", "select"}
	alterToDivExceptions         = []string{"div", "article", "section", "p"}
	presentationalAttributes     = []string{"align", "background", "bgcolor", "border", "cellpadding", "cellspacing", "frame", "hspace", "rules", "style", "valign", "vspace"}
	deprecatedSizeAttributeElems = []string{"table", "th", "td", "hr", "pre"}
	phrasingElems                = []string{
		"abbr", "audio", "b", "bdo", "br", "button", "cite", "code", "data",
		"datalist", "dfn", "em", "embed", "i", "img", "input", "kbd", "label",
		"mark", "math", "meter", "noscript", "object", "output", "progress", "q",
		"ruby", "samp", "script", "select", "small", "span", "strong", "sub",
		"sup", "textarea", "time", "var", "wbr"}
)

// flags is flags that used by parser.
type flags struct {
	stripUnlikelys     bool
	useWeightClasses   bool
	cleanConditionally bool
}

// parseAttempt is container for the result of previous parse attempts.
type parseAttempt struct {
	articleContent *html.Node
	textLength     int
}

// Article is the final readable content.
type Article struct {
	Title       string
	Byline      string
	Node        *html.Node
	Content     string
	TextContent string
	Length      int
	Excerpt     string
	SiteName    string
	Image       string
	Favicon     string
}

// Parser is the parser that parses the page to get the readable content.
type Parser struct {
	// MaxElemsToParse is the max number of nodes supported by this
	// parser. Default: 0 (no limit)
	MaxElemsToParse int
	// NTopCandidates is the number of top candidates to consider when
	// analysing how tight the competition is among candidates.
	NTopCandidates int
	// CharThresholds is the default number of chars an article must
	// have in order to return a result
	CharThresholds int
	// ClassesToPreserve are the classes that readability sets itself.
	ClassesToPreserve []string
	// KeepClasses specify whether the classes should be stripped or not.
	KeepClasses bool
	// TagsToScore is element tags to score by default.
	TagsToScore []string
	// Debug determines if the log should be printed or not. Default: false.
	Debug bool

	doc             *html.Node
	documentURI     *nurl.URL
	articleTitle    string
	articleByline   string
	articleDir      string
	articleSiteName string
	attempts        []parseAttempt
	flags           flags
}

// NewParser returns new Parser which set up with default value.
func NewParser() Parser {
	return Parser{
		MaxElemsToParse:   0,
		NTopCandidates:    5,
		CharThresholds:    500,
		ClassesToPreserve: []string{"page"},
		KeepClasses:       false,
		TagsToScore:       []string{"section", "h2", "h3", "h4", "h5", "h6", "p", "td", "pre"},
		Debug:             false,
	}
}

// postProcessContent runs any post-process modifications to article
// content as necessary.
func (ps *Parser) postProcessContent(articleContent *html.Node) {
	// Readability cannot open relative uris so we convert them to absolute uris.
	ps.fixRelativeURIs(articleContent)

	// Remove classes.
	if !ps.KeepClasses {
		ps.cleanClasses(articleContent)
	}

	// Remove readability attributes.
	ps.clearReadabilityAttr(articleContent)
}

// removeNodes iterates over a NodeList, calls `filterFn` for each node
// and removes node if function returned `true`. If function is not
// passed, removes all the nodes in node list.
func (ps *Parser) removeNodes(nodeList []*html.Node, filterFn func(*html.Node) bool) {
	for i := len(nodeList) - 1; i >= 0; i-- {
		node := nodeList[i]
		parentNode := node.Parent
		if parentNode != nil && (filterFn == nil || filterFn(node)) {
			parentNode.RemoveChild(node)
		}
	}
}

// replaceNodeTags iterates over a NodeList, and calls setNodeTag for
// each node.
func (ps *Parser) replaceNodeTags(nodeList []*html.Node, newTagName string) {
	for i := len(nodeList) - 1; i >= 0; i-- {
		node := nodeList[i]
		ps.setNodeTag(node, newTagName)
	}
}

// forEachNode iterates over a NodeList and runs fn on each node.
func (ps *Parser) forEachNode(nodeList []*html.Node, fn func(*html.Node, int)) {
	for i := 0; i < len(nodeList); i++ {
		fn(nodeList[i], i)
	}
}

// someNode iterates over a NodeList, return true if any of the
// provided iterate function calls returns true, false otherwise.
func (ps *Parser) someNode(nodeList []*html.Node, fn func(*html.Node) bool) bool {
	for i := 0; i < len(nodeList); i++ {
		if fn(nodeList[i]) {
			return true
		}
	}
	return false
}

// everyNode iterates over a NodeList, return true if all of the
// provided iterate function calls returns true, false otherwise.
func (ps *Parser) everyNode(nodeList []*html.Node, fn func(*html.Node) bool) bool {
	for i := 0; i < len(nodeList); i++ {
		if !fn(nodeList[i]) {
			return false
		}
	}
	return true
}

// concatNodeLists concats all nodelists passed as arguments.
func (ps *Parser) concatNodeLists(nodeLists ...[]*html.Node) []*html.Node {
	var result []*html.Node
	for i := 0; i < len(nodeLists); i++ {
		result = append(result, nodeLists[i]...)
	}
	return result
}

// getAllNodesWithTag returns all nodes that has tag inside tagNames.
func (ps *Parser) getAllNodesWithTag(node *html.Node, tagNames ...string) []*html.Node {
	var result []*html.Node
	for i := 0; i < len(tagNames); i++ {
		result = append(result, dom.GetElementsByTagName(node, tagNames[i])...)
	}
	return result
}

// cleanClasses removes the class="" attribute from every element in the
// given subtree, except those that match CLASSES_TO_PRESERVE and the
// classesToPreserve array from the options object.
func (ps *Parser) cleanClasses(node *html.Node) {
	nodeClassName := dom.ClassName(node)
	preservedClassName := []string{}
	for _, class := range strings.Fields(nodeClassName) {
		if indexOf(ps.ClassesToPreserve, class) != -1 {
			preservedClassName = append(preservedClassName, class)
		}
	}

	if len(preservedClassName) > 0 {
		dom.SetAttribute(node, "class", strings.Join(preservedClassName, " "))
	} else {
		dom.RemoveAttribute(node, "class")
	}

	for child := dom.FirstElementChild(node); child != nil; child = dom.NextElementSibling(child) {
		ps.cleanClasses(child)
	}
}

// fixRelativeURIs converts each <a> and <img> uri in the given element
// to an absolute URI, ignoring #ref URIs.
func (ps *Parser) fixRelativeURIs(articleContent *html.Node) {
	links := ps.getAllNodesWithTag(articleContent, "a")
	ps.forEachNode(links, func(link *html.Node, _ int) {
		href := dom.GetAttribute(link, "href")
		if href == "" {
			return
		}

		// Remove links with javascript: URIs, since they won't
		// work after scripts have been removed from the page.
		if strings.HasPrefix(href, "javascript:") {
			linkChilds := dom.ChildNodes(link)

			if len(linkChilds) == 1 && linkChilds[0].Type == html.TextNode {
				// If the link only contains simple text content,
				// it can be converted to a text node
				text := dom.CreateTextNode(dom.TextContent(link))
				dom.ReplaceChild(link.Parent, text, link)
			} else {
				// If the link has multiple children, they should
				// all be preserved
				container := dom.CreateElement("span")
				for _, child := range linkChilds {
					container.AppendChild(dom.CloneNode(child))
				}
				dom.ReplaceChild(link.Parent, container, link)
			}
		} else {
			newHref := toAbsoluteURI(href, ps.documentURI)
			if newHref == "" {
				dom.RemoveAttribute(link, "href")
			} else {
				dom.SetAttribute(link, "href", newHref)
			}
		}
	})

	medias := ps.getAllNodesWithTag(articleContent, "img", "picture", "figure", "video", "audio", "source")
	ps.forEachNode(medias, func(media *html.Node, _ int) {
		src := dom.GetAttribute(media, "src")
		poster := dom.GetAttribute(media, "poster")
		srcset := dom.GetAttribute(media, "srcset")

		if src != "" {
			newSrc := toAbsoluteURI(src, ps.documentURI)
			dom.SetAttribute(media, "src", newSrc)
		}

		if poster != "" {
			newPoster := toAbsoluteURI(poster, ps.documentURI)
			dom.SetAttribute(media, "poster", newPoster)
		}

		if srcset != "" {
			newSrcset := rxSrcsetURL.ReplaceAllStringFunc(srcset, func(s string) string {
				p := rxSrcsetURL.FindStringSubmatch(s)
				return toAbsoluteURI(p[1], ps.documentURI) + p[2] + p[3]
			})

			dom.SetAttribute(media, "srcset", newSrcset)
		}
	})
}

// getArticleTitle attempts to get the article title.
func (ps *Parser) getArticleTitle() string {
	doc := ps.doc
	curTitle := ""
	origTitle := ""
	titleHadHierarchicalSeparators := false

	// If they had an element with tag "title" in their HTML
	if nodes := dom.GetElementsByTagName(doc, "title"); len(nodes) > 0 {
		origTitle = ps.getInnerText(nodes[0], true)
		curTitle = origTitle
	}

	// If there's a separator in the title, first remove the final part
	if rxTitleSeparator.MatchString(curTitle) {
		titleHadHierarchicalSeparators = rxTitleHierarchySep.MatchString(curTitle)
		curTitle = rxTitleRemoveFinalPart.ReplaceAllString(origTitle, "$1")

		// If the resulting title is too short (3 words or fewer), remove
		// the first part instead:
		if wordCount(curTitle) < 3 {
			curTitle = rxTitleRemove1stPart.ReplaceAllString(origTitle, "$1")
		}
	} else if strings.Index(curTitle, ": ") != -1 {
		// Check if we have an heading containing this exact string, so
		// we could assume it's the full title.
		headings := ps.concatNodeLists(
			dom.GetElementsByTagName(doc, "h1"),
			dom.GetElementsByTagName(doc, "h2"),
		)

		trimmedTitle := strings.TrimSpace(curTitle)
		match := ps.someNode(headings, func(heading *html.Node) bool {
			return strings.TrimSpace(dom.TextContent(heading)) == trimmedTitle
		})

		// If we don't, let's extract the title out of the original
		// title string.
		if !match {
			curTitle = origTitle[strings.LastIndex(origTitle, ":")+1:]

			// If the title is now too short, try the first colon instead:
			if wordCount(curTitle) < 3 {
				curTitle = origTitle[strings.Index(origTitle, ":")+1:]
				// But if we have too many words before the colon there's
				// something weird with the titles and the H tags so let's
				// just use the original title instead
			} else if wordCount(origTitle[:strings.Index(origTitle, ":")]) > 5 {
				curTitle = origTitle
			}
		}
	} else if charCount(curTitle) > 150 || charCount(curTitle) < 15 {
		if hOnes := dom.GetElementsByTagName(doc, "h1"); len(hOnes) == 1 {
			curTitle = ps.getInnerText(hOnes[0], true)
		}
	}

	curTitle = strings.TrimSpace(curTitle)
	curTitle = rxNormalize.ReplaceAllString(curTitle, " ")
	// If we now have 4 words or fewer as our title, and either no
	// 'hierarchical' separators (\, /, > or ») were found in the original
	// title or we decreased the number of words by more than 1 word, use
	// the original title.
	curTitleWordCount := wordCount(curTitle)
	tmpOrigTitle := rxTitleAnySeparator.ReplaceAllString(origTitle, "")

	if curTitleWordCount <= 4 &&
		(!titleHadHierarchicalSeparators ||
			curTitleWordCount != wordCount(tmpOrigTitle)-1) {
		curTitle = origTitle
	}

	return curTitle
}

// prepDocument prepares the HTML document for readability to scrape it.
// This includes things like stripping javascript, CSS, and handling
// terrible markup.
func (ps *Parser) prepDocument() {
	doc := ps.doc

	// ADDITIONAL, not exist in readability.js:
	// Remove all comments,
	ps.removeComments(doc)

	// Remove all style tags in head
	ps.removeNodes(dom.GetElementsByTagName(doc, "style"), nil)

	if nodes := dom.GetElementsByTagName(doc, "body"); len(nodes) > 0 && nodes[0] != nil {
		ps.replaceBrs(nodes[0])
	}

	ps.replaceNodeTags(dom.GetElementsByTagName(doc, "font"), "span")
}

// nextElement finds the next element, starting from the given node, and
// ignoring whitespace in between. If the given node is an element, the
// same node is returned.
func (ps *Parser) nextElement(node *html.Node) *html.Node {
	next := node
	for next != nil && next.Type != html.ElementNode && rxWhitespace.MatchString(dom.TextContent(next)) {
		next = next.NextSibling
	}
	return next
}

// replaceBrs replaces 2 or more successive <br> with a single <p>.
// Whitespace between <br> elements are ignored. For example:
//   <div>foo<br>bar<br> <br><br>abc</div>
// will become:
//   <div>foo<br>bar<p>abc</p></div>
func (ps *Parser) replaceBrs(elem *html.Node) {
	ps.forEachNode(ps.getAllNodesWithTag(elem, "br"), func(br *html.Node, _ int) {
		next := br.NextSibling

		// Whether 2 or more <br> elements have been found and replaced
		// with a <p> block.
		replaced := false

		// If we find a <br> chain, remove the <br>s until we hit another
		// element or non-whitespace. This leaves behind the first <br>
		// in the chain (which will be replaced with a <p> later).
		for {
			next = ps.nextElement(next)
			if next == nil || dom.TagName(next) != "br" {
				break
			}

			replaced = true
			brSibling := next.NextSibling
			next.Parent.RemoveChild(next)
			next = brSibling
		}

		// If we removed a <br> chain, replace the remaining <br> with a <p>. Add
		// all sibling nodes as children of the <p> until we hit another <br>
		// chain.
		if replaced {
			p := dom.CreateElement("p")
			dom.ReplaceChild(br.Parent, p, br)

			next = p.NextSibling
			for next != nil {
				// If we've hit another <br><br>, we're done adding children to this <p>.
				if dom.TagName(next) == "br" {
					nextElem := ps.nextElement(next.NextSibling)
					if nextElem != nil && dom.TagName(nextElem) == "br" {
						break
					}
				}

				if !ps.isPhrasingContent(next) {
					break
				}

				// Otherwise, make this node a child of the new <p>.
				sibling := next.NextSibling
				dom.AppendChild(p, next)
				next = sibling
			}

			for p.LastChild != nil && ps.isWhitespace(p.LastChild) {
				p.RemoveChild(p.LastChild)
			}

			if dom.TagName(p.Parent) == "p" {
				ps.setNodeTag(p.Parent, "div")
			}
		}
	})
}

// setNodeTag changes tag of the node to newTagName.
func (ps *Parser) setNodeTag(node *html.Node, newTagName string) {
	if node.Type == html.ElementNode {
		node.Data = newTagName
	}
}

// prepArticle prepares the article node for display. Clean out any
// inline styles, iframes, forms, strip extraneous <p> tags, etc.
func (ps *Parser) prepArticle(articleContent *html.Node) {
	ps.cleanStyles(articleContent)

	// Check for data tables before we continue, to avoid removing
	// items in those tables, which will often be isolated even
	// though they're visually linked to other content-ful elements
	// (text, images, etc.).
	ps.markDataTables(articleContent)

	ps.fixLazyImages(articleContent)

	// Clean out junk from the article content
	ps.cleanConditionally(articleContent, "form")
	ps.cleanConditionally(articleContent, "fieldset")
	ps.clean(articleContent, "object")
	ps.clean(articleContent, "embed")
	ps.clean(articleContent, "h1")
	ps.clean(articleContent, "footer")
	ps.clean(articleContent, "link")
	ps.clean(articleContent, "aside")

	// Clean out elements have "share" in their id/class combinations
	// from final top candidates, which means we don't remove the top
	// candidates even they have "share".
	shareElementThreshold := ps.CharThresholds

	ps.forEachNode(dom.Children(articleContent), func(topCandidate *html.Node, _ int) {
		ps.cleanMatchedNodes(topCandidate, func(node *html.Node, nodeClassID string) bool {
			return rxShareElements.MatchString(nodeClassID) && charCount(dom.TextContent(node)) < shareElementThreshold
		})
	})

	// If there is only one h2 and its text content substantially
	// equals article title, they are probably using it as a header
	// and not a subheader, so remove it since we already extract
	// the title separately.
	if h2s := dom.GetElementsByTagName(articleContent, "h2"); len(h2s) == 1 {
		h2 := h2s[0]
		h2Text := dom.TextContent(h2)
		lengthSimilarRate := float64(charCount(h2Text)-charCount(ps.articleTitle)) / float64(charCount(ps.articleTitle))
		if math.Abs(lengthSimilarRate) < 0.5 {
			titlesMatch := false
			if lengthSimilarRate > 0 {
				titlesMatch = strings.Contains(h2Text, ps.articleTitle)
			} else {
				titlesMatch = strings.Contains(ps.articleTitle, h2Text)
			}
			if titlesMatch {
				ps.clean(articleContent, "h2")
			}
		}
	}

	ps.clean(articleContent, "iframe")
	ps.clean(articleContent, "input")
	ps.clean(articleContent, "textarea")
	ps.clean(articleContent, "select")
	ps.clean(articleContent, "button")
	ps.cleanHeaders(articleContent)

	// Do these last as the previous stuff may have removed junk
	// that will affect these
	ps.cleanConditionally(articleContent, "table")
	ps.cleanConditionally(articleContent, "ul")
	ps.cleanConditionally(articleContent, "div")

	// Remove extra paragraphs
	ps.removeNodes(dom.GetElementsByTagName(articleContent, "p"), func(p *html.Node) bool {
		imgCount := len(dom.GetElementsByTagName(p, "img"))
		embedCount := len(dom.GetElementsByTagName(p, "embed"))
		objectCount := len(dom.GetElementsByTagName(p, "object"))
		// At this point, nasty iframes have been removed, only
		// remain embedded video ones.
		iframeCount := len(dom.GetElementsByTagName(p, "iframe"))
		totalCount := imgCount + embedCount + objectCount + iframeCount

		return totalCount == 0 && ps.getInnerText(p, false) == ""
	})

	ps.forEachNode(dom.GetElementsByTagName(articleContent, "br"), func(br *html.Node, _ int) {
		next := ps.nextElement(br.NextSibling)
		if next != nil && dom.TagName(next) == "p" {
			br.Parent.RemoveChild(br)
		}
	})

	// Remove single-cell tables
	ps.forEachNode(dom.GetElementsByTagName(articleContent, "table"), func(table *html.Node, _ int) {
		tbody := table
		if ps.hasSingleTagInsideElement(table, "tbody") {
			tbody = dom.FirstElementChild(table)
		}

		if ps.hasSingleTagInsideElement(tbody, "tr") {
			row := dom.FirstElementChild(tbody)
			if ps.hasSingleTagInsideElement(row, "td") {
				cell := dom.FirstElementChild(row)

				newTag := "div"
				if ps.everyNode(dom.ChildNodes(cell), ps.isPhrasingContent) {
					newTag = "p"
				}

				ps.setNodeTag(cell, newTag)
				dom.ReplaceChild(table.Parent, cell, table)
			}
		}
	})
}

// initializeNode initializes a node with the readability score.
// Also checks the className/id for special names to add to its score.
func (ps *Parser) initializeNode(node *html.Node) {
	contentScore := float64(ps.getClassWeight(node))
	switch dom.TagName(node) {
	case "div":
		contentScore += 5
	case "pre", "td", "blockquote":
		contentScore += 3
	case "address", "ol", "ul", "dl", "dd", "dt", "li", "form":
		contentScore -= 3
	case "h1", "h2", "h3", "h4", "h5", "h6", "th":
		contentScore -= 5
	}

	ps.setContentScore(node, contentScore)
}

// removeAndGetNext remove node and returns its next node.
func (ps *Parser) removeAndGetNext(node *html.Node) *html.Node {
	nextNode := ps.getNextNode(node, true)
	if node.Parent != nil {
		node.Parent.RemoveChild(node)
	}
	return nextNode
}

// getNextNode traverses the DOM from node to node, starting at the
// node passed in. Pass true for the second parameter to indicate
// this node itself (and its kids) are going away, and we want the
// next node over. Calling this in a loop will traverse the DOM
// depth-first.
// In Readability.js, ignoreSelfAndKids default to false.
func (ps *Parser) getNextNode(node *html.Node, ignoreSelfAndKids bool) *html.Node {
	// First check for kids if those aren't being ignored
	if firstChild := dom.FirstElementChild(node); !ignoreSelfAndKids && firstChild != nil {
		return firstChild
	}

	// Then for siblings...
	if sibling := dom.NextElementSibling(node); sibling != nil {
		return sibling
	}

	// And finally, move up the parent chain *and* find a sibling
	// (because this is depth-first traversal, we will have already
	// seen the parent nodes themselves).
	for {
		node = node.Parent
		if node == nil || dom.NextElementSibling(node) != nil {
			break
		}
	}

	if node != nil {
		return dom.NextElementSibling(node)
	}

	return nil
}

// checkByline determines if a node is used as byline.
func (ps *Parser) checkByline(node *html.Node, matchString string) bool {
	if ps.articleByline != "" {
		return false
	}

	rel := dom.GetAttribute(node, "rel")
	itemprop := dom.GetAttribute(node, "itemprop")
	nodeText := dom.TextContent(node)
	if (rel == "author" || strings.Contains(itemprop, "author") || rxByline.MatchString(matchString)) &&
		ps.isValidByline(nodeText) {
		nodeText = strings.TrimSpace(nodeText)
		nodeText = strings.Join(strings.Fields(nodeText), " ")
		ps.articleByline = nodeText
		return true
	}

	return false
}

// getNodeAncestors gets the node's direct parent and grandparents.
// In Readability.js, maxDepth default to 0.
func (ps *Parser) getNodeAncestors(node *html.Node, maxDepth int) []*html.Node {
	i := 0
	var ancestors []*html.Node

	for node.Parent != nil {
		i++
		ancestors = append(ancestors, node.Parent)
		if maxDepth > 0 && i == maxDepth {
			break
		}
		node = node.Parent
	}
	return ancestors
}

// grabArticle uses a variety of metrics (content score, classname,
// element types), find the content that is most likely to be the
// stuff a user wants to read. Then return it wrapped up in a div.
func (ps *Parser) grabArticle() *html.Node {
	for {
		doc := dom.CloneNode(ps.doc)

		var page *html.Node
		if nodes := dom.GetElementsByTagName(doc, "body"); len(nodes) > 0 {
			page = nodes[0]
		}

		// We can't grab an article if we don't have a page!
		if page == nil {
			return nil
		}

		// First, node prepping. Trash nodes that look cruddy (like ones
		// with the class name "comment", etc), and turn divs into P
		// tags where they have been used inappropriately (as in, where
		// they contain no other block level elements.)
		var elementsToScore []*html.Node
		var node = dom.DocumentElement(doc)

		for node != nil {
			matchString := dom.ClassName(node) + " " + dom.ID(node)

			if !ps.isProbablyVisible(node) {
				node = ps.removeAndGetNext(node)
				continue
			}

			// Check to see if this node is a byline, and remove it if
			// it is true.
			if ps.checkByline(node, matchString) {
				node = ps.removeAndGetNext(node)
				continue
			}

			// Remove unlikely candidates
			nodeTagName := dom.TagName(node)
			if ps.flags.stripUnlikelys {
				if rxUnlikelyCandidates.MatchString(matchString) &&
					!rxOkMaybeItsACandidate.MatchString(matchString) &&
					!ps.hasAncestorTag(node, "table", 3, nil) &&
					nodeTagName != "body" && nodeTagName != "a" {
					node = ps.removeAndGetNext(node)
					continue
				}

				if dom.GetAttribute(node, "role") == "complementary" {
					node = ps.removeAndGetNext(node)
					continue
				}
			}

			// Remove DIV, SECTION, and HEADER nodes without any
			// content(e.g. text, image, video, or iframe).
			switch nodeTagName {
			case "div", "section", "header",
				"h1", "h2", "h3", "h4", "h5", "h6":
				if ps.isElementWithoutContent(node) {
					node = ps.removeAndGetNext(node)
					continue
				}
			}

			if indexOf(ps.TagsToScore, nodeTagName) != -1 {
				elementsToScore = append(elementsToScore, node)
			}

			// Turn all divs that don't have children block level
			// elements into p's
			if nodeTagName == "div" {
				// Put phrasing content into paragraphs.
				var p *html.Node
				childNode := node.FirstChild
				for childNode != nil {
					nextSibling := childNode.NextSibling
					if ps.isPhrasingContent(childNode) {
						if p != nil {
							dom.AppendChild(p, childNode)
						} else if !ps.isWhitespace(childNode) {
							p = dom.CreateElement("p")
							dom.AppendChild(p, dom.CloneNode(childNode))
							dom.ReplaceChild(node, p, childNode)
						}
					} else if p != nil {
						for p.LastChild != nil && ps.isWhitespace(p.LastChild) {
							p.RemoveChild(p.LastChild)
						}
						p = nil
					}
					childNode = nextSibling
				}

				// Sites like http://mobile.slate.com encloses each
				// paragraph with a DIV element. DIVs with only a P
				// element inside and no text content can be safely
				// converted into plain P elements to avoid confusing
				// the scoring algorithm with DIVs with are, in
				// practice, paragraphs.
				if ps.hasSingleTagInsideElement(node, "p") && ps.getLinkDensity(node) < 0.25 {
					newNode := dom.Children(node)[0]
					newNode, _ = dom.ReplaceChild(node.Parent, newNode, node)
					node = newNode
					elementsToScore = append(elementsToScore, node)
				} else if !ps.hasChildBlockElement(node) {
					ps.setNodeTag(node, "p")
					elementsToScore = append(elementsToScore, node)
				}
			}
			node = ps.getNextNode(node, false)
		}

		// Loop through all paragraphs, and assign a score to them based
		// on how content-y they look. Then add their score to their
		// parent node. A score is determined by things like number of
		// commas, class names, etc. Maybe eventually link density.
		var candidates []*html.Node
		ps.forEachNode(elementsToScore, func(elementToScore *html.Node, _ int) {
			if elementToScore.Parent == nil || dom.TagName(elementToScore.Parent) == "" {
				return
			}

			// If this paragraph is less than 25 characters, don't even count it.
			innerText := ps.getInnerText(elementToScore, true)
			if charCount(innerText) < 25 {
				return
			}

			// Exclude nodes with no ancestor.
			ancestors := ps.getNodeAncestors(elementToScore, 3)
			if len(ancestors) == 0 {
				return
			}

			// Add a point for the paragraph itself as a base.
			contentScore := 1

			// Add points for any commas within this paragraph.
			contentScore += strings.Count(innerText, ",")

			// For every 100 characters in this paragraph, add another point. Up to 3 points.
			contentScore += int(math.Min(math.Floor(float64(charCount(innerText))/100.0), 3.0))

			// Initialize and score ancestors.
			ps.forEachNode(ancestors, func(ancestor *html.Node, level int) {
				if dom.TagName(ancestor) == "" || ancestor.Parent == nil || ancestor.Parent.Type != html.ElementNode {
					return
				}

				if !ps.hasContentScore(ancestor) {
					ps.initializeNode(ancestor)
					candidates = append(candidates, ancestor)
				}

				// Node score divider:
				// - parent:             1 (no division)
				// - grandparent:        2
				// - great grandparent+: ancestor level * 3
				scoreDivider := 1
				switch level {
				case 0:
					scoreDivider = 1
				case 1:
					scoreDivider = 2
				default:
					scoreDivider = level * 3
				}

				ancestorScore := ps.getContentScore(ancestor)
				ancestorScore += float64(contentScore) / float64(scoreDivider)
				ps.setContentScore(ancestor, ancestorScore)
			})
		})

		// These lines are a bit different compared to Readability.js.
		// In Readability.js, they fetch NTopCandidates utilising array
		// method like `splice` and `pop`. In Go, array method like that
		// is not as simple, especially since we are working with pointer.
		// So, here we simply sort top candidates, and limit it to
		// max NTopCandidates.

		// Scale the final candidates score based on link density. Good
		// content should have a relatively small link density (5% or
		// less) and be mostly unaffected by this operation.
		for i := 0; i < len(candidates); i++ {
			candidate := candidates[i]
			candidateScore := ps.getContentScore(candidate) * (1 - ps.getLinkDensity(candidate))
			ps.setContentScore(candidate, candidateScore)
		}

		// After we've calculated scores, sort through all of the possible
		// candidate nodes we found and find the one with the highest score.
		sort.Slice(candidates, func(i int, j int) bool {
			return ps.getContentScore(candidates[i]) > ps.getContentScore(candidates[j])
		})

		var topCandidates []*html.Node
		if len(candidates) > ps.NTopCandidates {
			topCandidates = candidates[:ps.NTopCandidates]
		} else {
			topCandidates = candidates
		}

		var topCandidate, parentOfTopCandidate *html.Node
		neededToCreateTopCandidate := false
		if len(topCandidates) > 0 {
			topCandidate = topCandidates[0]
		}

		// If we still have no top candidate, just use the body as a last
		// resort. We also have to copy the body node so it is something
		// we can modify.
		if topCandidate == nil || dom.TagName(topCandidate) == "body" {
			// Move all of the page's children into topCandidate
			topCandidate = dom.CreateElement("div")
			neededToCreateTopCandidate = true
			// Move everything (not just elements, also text nodes etc.)
			// into the container so we even include text directly in the body:
			kids := dom.ChildNodes(page)
			for i := 0; i < len(kids); i++ {
				dom.AppendChild(topCandidate, kids[i])
			}

			dom.AppendChild(page, topCandidate)
			ps.initializeNode(topCandidate)
		} else if topCandidate != nil {
			// Find a better top candidate node if it contains (at least three)
			// nodes which belong to `topCandidates` array and whose scores are
			// quite closed with current `topCandidate` node.
			topCandidateScore := ps.getContentScore(topCandidate)
			var alternativeCandidateAncestors [][]*html.Node
			for i := 1; i < len(topCandidates); i++ {
				if ps.getContentScore(topCandidates[i])/topCandidateScore >= 0.75 {
					topCandidateAncestors := ps.getNodeAncestors(topCandidates[i], 0)
					alternativeCandidateAncestors = append(alternativeCandidateAncestors, topCandidateAncestors)
				}
			}

			minimumTopCandidates := 3
			if len(alternativeCandidateAncestors) >= minimumTopCandidates {
				parentOfTopCandidate = topCandidate.Parent
				for parentOfTopCandidate != nil && dom.TagName(parentOfTopCandidate) != "body" {
					listContainingThisAncestor := 0
					for ancestorIndex := 0; ancestorIndex < len(alternativeCandidateAncestors) && listContainingThisAncestor < minimumTopCandidates; ancestorIndex++ {
						if dom.IncludeNode(alternativeCandidateAncestors[ancestorIndex], parentOfTopCandidate) {
							listContainingThisAncestor++
						}
					}

					if listContainingThisAncestor >= minimumTopCandidates {
						topCandidate = parentOfTopCandidate
						break
					}

					parentOfTopCandidate = parentOfTopCandidate.Parent
				}
			}

			if !ps.hasContentScore(topCandidate) {
				ps.initializeNode(topCandidate)
			}

			// Because of our bonus system, parents of candidates might
			// have scores themselves. They get half of the node. There
			// won't be nodes with higher scores than our topCandidate,
			// but if we see the score going *up* in the first few steps *
			// up the tree, that's a decent sign that there might be more
			// content lurking in other places that we want to unify in.
			// The sibling stuff below does some of that - but only if
			// we've looked high enough up the DOM tree.
			parentOfTopCandidate = topCandidate.Parent
			lastScore := ps.getContentScore(topCandidate)
			// The scores shouldn't get too lops.
			scoreThreshold := lastScore / 3.0
			for parentOfTopCandidate != nil && dom.TagName(parentOfTopCandidate) != "body" {
				if !ps.hasContentScore(parentOfTopCandidate) {
					parentOfTopCandidate = parentOfTopCandidate.Parent
					continue
				}

				parentScore := ps.getContentScore(parentOfTopCandidate)
				if parentScore < scoreThreshold {
					break
				}

				if parentScore > lastScore {
					// Alright! We found a better parent to use.
					topCandidate = parentOfTopCandidate
					break
				}

				lastScore = parentScore
				parentOfTopCandidate = parentOfTopCandidate.Parent
			}

			// If the top candidate is the only child, use parent
			// instead. This will help sibling joining logic when
			// adjacent content is actually located in parent's
			// sibling node.
			parentOfTopCandidate = topCandidate.Parent
			for parentOfTopCandidate != nil && dom.TagName(parentOfTopCandidate) != "body" && len(dom.Children(parentOfTopCandidate)) == 1 {
				topCandidate = parentOfTopCandidate
				parentOfTopCandidate = topCandidate.Parent
			}

			if !ps.hasContentScore(topCandidate) {
				ps.initializeNode(topCandidate)
			}
		}

		// Now that we have the top candidate, look through its siblings
		// for content that might also be related. Things like preambles,
		// content split by ads that we removed, etc.
		articleContent := dom.CreateElement("div")
		siblingScoreThreshold := math.Max(10, ps.getContentScore(topCandidate)*0.2)

		// Keep potential top candidate's parent node to try to get text direction of it later.
		topCandidateScore := ps.getContentScore(topCandidate)
		topCandidateClassName := dom.ClassName(topCandidate)

		parentOfTopCandidate = topCandidate.Parent
		siblings := dom.Children(parentOfTopCandidate)
		for s := 0; s < len(siblings); s++ {
			sibling := siblings[s]
			appendNode := false

			if sibling == topCandidate {
				appendNode = true
			} else {
				contentBonus := float64(0)

				// Give a bonus if sibling nodes and top candidates have the example same classname
				if dom.ClassName(sibling) == topCandidateClassName && topCandidateClassName != "" {
					contentBonus += topCandidateScore * 0.2
				}

				if ps.hasContentScore(sibling) && ps.getContentScore(sibling)+contentBonus >= siblingScoreThreshold {
					appendNode = true
				} else if dom.TagName(sibling) == "p" {
					linkDensity := ps.getLinkDensity(sibling)
					nodeContent := ps.getInnerText(sibling, true)
					nodeLength := charCount(nodeContent)

					if nodeLength > 80 && linkDensity < 0.25 {
						appendNode = true
					} else if nodeLength < 80 && nodeLength > 0 && linkDensity == 0 &&
						rxSentencePeriod.MatchString(nodeContent) {
						appendNode = true
					}
				}
			}

			if appendNode {
				// We have a node that isn't a common block level
				// element, like a form or td tag. Turn it into a div
				// so it doesn't get filtered out later by accident.
				if indexOf(alterToDivExceptions, dom.TagName(sibling)) == -1 {
					ps.setNodeTag(sibling, "div")
				}

				dom.AppendChild(articleContent, sibling)
			}
		}

		// So we have all of the content that we need. Now we clean
		// it up for presentation.
		ps.prepArticle(articleContent)

		if neededToCreateTopCandidate {
			// We already created a fake div thing, and there wouldn't
			// have been any siblings left for the previous loop, so
			// there's no point trying to create a new div, and then
			// move all the children over. Just assign IDs and class
			// names here. No need to append because that already
			// happened anyway.
			//
			// By the way, this line is different with Readability.js.
			// In Readability.js, when using `appendChild`, the node is
			// still referenced. Meanwhile here, our `appendChild` will
			// clone the node, put it in the new place, then delete
			// the original.
			firstChild := dom.FirstElementChild(articleContent)
			if firstChild != nil && dom.TagName(firstChild) == "div" {
				dom.SetAttribute(firstChild, "id", "readability-page-1")
				dom.SetAttribute(firstChild, "class", "page")
			}
		} else {
			div := dom.CreateElement("div")
			dom.SetAttribute(div, "id", "readability-page-1")
			dom.SetAttribute(div, "class", "page")
			childs := dom.ChildNodes(articleContent)
			for i := 0; i < len(childs); i++ {
				dom.AppendChild(div, childs[i])
			}
			dom.AppendChild(articleContent, div)
		}

		parseSuccessful := true

		// Now that we've gone through the full algorithm, check to
		// see if we got any meaningful content. If we didn't, we may
		// need to re-run grabArticle with different flags set. This
		// gives us a higher likelihood of finding the content, and
		// the sieve approach gives us a higher likelihood of
		// finding the -right- content.
		textLength := charCount(ps.getInnerText(articleContent, true))
		if textLength < ps.CharThresholds {
			parseSuccessful = false

			if ps.flags.stripUnlikelys {
				ps.flags.stripUnlikelys = false
				ps.attempts = append(ps.attempts, parseAttempt{
					articleContent: articleContent,
					textLength:     textLength,
				})
			} else if ps.flags.useWeightClasses {
				ps.flags.useWeightClasses = false
				ps.attempts = append(ps.attempts, parseAttempt{
					articleContent: articleContent,
					textLength:     textLength,
				})
			} else if ps.flags.cleanConditionally {
				ps.flags.cleanConditionally = false
				ps.attempts = append(ps.attempts, parseAttempt{
					articleContent: articleContent,
					textLength:     textLength,
				})
			} else {
				ps.attempts = append(ps.attempts, parseAttempt{
					articleContent: articleContent,
					textLength:     textLength,
				})

				// No luck after removing flags, just return the
				// longest text we found during the different loops *
				sort.Slice(ps.attempts, func(i, j int) bool {
					return ps.attempts[i].textLength > ps.attempts[j].textLength
				})

				// But first check if we actually have something
				if ps.attempts[0].textLength == 0 {
					return nil
				}

				articleContent = ps.attempts[0].articleContent
				parseSuccessful = true
			}
		}

		if parseSuccessful {
			return articleContent
		}
	}
}

// isValidByline checks whether the input string could be a byline.
// This verifies that the input is a string, and that the length
// is less than 100 chars.
func (ps *Parser) isValidByline(byline string) bool {
	byline = strings.TrimSpace(byline)
	nChar := charCount(byline)
	return nChar > 0 && nChar < 100
}

// getArticleMetadata attempts to get excerpt and byline
// metadata for the article.
func (ps *Parser) getArticleMetadata() map[string]string {
	values := make(map[string]string)
	metaElements := dom.GetElementsByTagName(ps.doc, "meta")

	// Find description tags.
	ps.forEachNode(metaElements, func(element *html.Node, _ int) {
		elementName := dom.GetAttribute(element, "name")
		elementProperty := dom.GetAttribute(element, "property")
		content := dom.GetAttribute(element, "content")
		if content == "" {
			return
		}
		matches := []string{}
		name := ""

		if elementProperty != "" {
			matches = rxPropertyPattern.FindAllString(elementProperty, -1)
			for i := len(matches) - 1; i >= 0; i-- {
				// Convert to lowercase, and remove any whitespace
				// so we can match belops.
				name = strings.ToLower(matches[i])
				name = strings.Join(strings.Fields(name), "")
				// multiple authors
				values[name] = strings.TrimSpace(content)
			}
		}

		if len(matches) == 0 && elementName != "" && rxNamePattern.MatchString(elementName) {
			// Convert to lowercase, remove any whitespace, and convert
			// dots to colons so we can match belops.
			name = strings.ToLower(elementName)
			name = strings.Join(strings.Fields(name), "")
			name = strings.Replace(name, ".", ":", -1)
			values[name] = strings.TrimSpace(content)
		}
	})

	// get title
	metadataTitle := ""
	possibleAttrNames := []string{
		"dc:title", "dcterm:title", "og:title", "weibo:article:title",
		"weibo:webpage:title", "title", "twitter:title"}
	for _, name := range possibleAttrNames {
		if value, ok := values[name]; ok {
			metadataTitle = value
			break
		}
	}

	if metadataTitle == "" {
		metadataTitle = ps.getArticleTitle()
	}

	// get author
	metadataByline := ""
	possibleAttrNames = []string{"dc:creator", "dcterm:creator", "author"}
	for _, name := range possibleAttrNames {
		if value, ok := values[name]; ok {
			metadataByline = value
			break
		}
	}

	// get description
	metadataExcerpt := ""
	possibleAttrNames = []string{
		"dc:description", "dcterm:description", "og:description",
		"weibo:article:description", "weibo:webpage:description",
		"description", "twitter:description"}
	for _, name := range possibleAttrNames {
		if value, ok := values[name]; ok {
			metadataExcerpt = value
			break
		}
	}

	// get site name
	metadataSiteName := values["og:site_name"]

	// get image thumbnail
	metadataImage := ""
	possibleAttrNames = []string{"og:image", "image", "twitter:image"}
	for _, name := range possibleAttrNames {
		if value, ok := values[name]; ok {
			metadataImage = toAbsoluteURI(value, ps.documentURI)
			break
		}
	}

	// get favicon
	metadataFavicon := ps.getArticleFavicon()

	// in many sites the meta value is escaped with HTML entities,
	// so here we need to unescape it
	metadataTitle = shtml.UnescapeString(metadataTitle)
	metadataByline = shtml.UnescapeString(metadataByline)
	metadataExcerpt = shtml.UnescapeString(metadataExcerpt)
	metadataSiteName = shtml.UnescapeString(metadataSiteName)

	return map[string]string{
		"title":    metadataTitle,
		"byline":   metadataByline,
		"excerpt":  metadataExcerpt,
		"siteName": metadataSiteName,
		"image":    metadataImage,
		"favicon":  metadataFavicon,
	}
}

// isSingleImage checks if node is image, or if node contains exactly
// only one image whether as a direct child or as its descendants.
func (ps *Parser) isSingleImage(node *html.Node) bool {
	if dom.TagName(node) == "img" {
		return true
	}

	children := dom.Children(node)
	textContent := dom.TextContent(node)
	if len(children) != 1 || strings.TrimSpace(textContent) != "" {
		return false
	}

	return ps.isSingleImage(children[0])
}

// unwrapNoscriptImages finds all <noscript> that are located after <img> nodes,
// and which contain only one <img> element. Replace the first image with
// the image from inside the <noscript> tag, and remove the <noscript> tag.
// This improves the quality of the images we use on some sites (e.g. Medium).
func (ps *Parser) unwrapNoscriptImages(doc *html.Node) {
	// Find img without source or attributes that might contains image, and
	// remove it. This is done to prevent a placeholder img is replaced by
	// img from noscript in next step.
	imgs := dom.GetElementsByTagName(doc, "img")
	ps.forEachNode(imgs, func(img *html.Node, _ int) {
		for _, attr := range img.Attr {
			switch attr.Key {
			case "src", "data-src", "srcset", "data-srcset":
				return
			}

			if rxImgExtensions.MatchString(attr.Val) {
				return
			}
		}

		img.Parent.RemoveChild(img)
	})

	// Next find noscript and try to extract its image
	noscripts := dom.GetElementsByTagName(doc, "noscript")
	ps.forEachNode(noscripts, func(noscript *html.Node, _ int) {
		// Parse content of noscript and make sure it only contains image
		noscriptContent := dom.TextContent(noscript)
		tmpDoc, err := html.Parse(strings.NewReader(noscriptContent))
		if err != nil {
			return
		}

		tmpBody := dom.GetElementsByTagName(tmpDoc, "body")[0]
		if !ps.isSingleImage(tmpBody) {
			return
		}

		// If noscript has previous sibling and it only contains image,
		// replace it with noscript content. However we also keep old
		// attributes that might contains image.
		prevElement := dom.PreviousElementSibling(noscript)
		if prevElement != nil && ps.isSingleImage(prevElement) {
			prevImg := prevElement
			if dom.TagName(prevImg) != "img" {
				prevImg = dom.GetElementsByTagName(prevElement, "img")[0]
			}

			newImg := dom.GetElementsByTagName(tmpBody, "img")[0]
			for _, attr := range prevImg.Attr {
				if attr.Val == "" {
					continue
				}

				if attr.Key == "src" || attr.Key == "srcset" || rxImgExtensions.MatchString(attr.Val) {
					if dom.GetAttribute(newImg, attr.Key) == attr.Val {
						continue
					}

					attrName := attr.Key
					if dom.HasAttribute(newImg, attrName) {
						attrName = "data-old-" + attrName
					}

					dom.SetAttribute(newImg, attrName, attr.Val)
				}
			}

			dom.ReplaceChild(noscript.Parent, dom.FirstElementChild(tmpBody), prevElement)
		}
	})
}

// removeScripts removes script tags from the document.
func (ps *Parser) removeScripts(doc *html.Node) {
	scripts := dom.GetElementsByTagName(doc, "script")
	noScripts := dom.GetElementsByTagName(doc, "noscript")
	ps.removeNodes(scripts, nil)
	ps.removeNodes(noScripts, nil)
}

// hasSingleTagInsideElement check if this node has only whitespace
// and a single element with given tag. Returns false if the DIV node
// contains non-empty text nodes or if it contains no element with
// given tag or more than 1 element.
func (ps *Parser) hasSingleTagInsideElement(element *html.Node, tag string) bool {
	// There should be exactly 1 element child with given tag
	if childs := dom.Children(element); len(childs) != 1 || dom.TagName(childs[0]) != tag {
		return false
	}

	// And there should be no text nodes with real content
	return !ps.someNode(dom.ChildNodes(element), func(node *html.Node) bool {
		return node.Type == html.TextNode && rxHasContent.MatchString(dom.TextContent(node))
	})
}

// isElementWithoutContent determines if node is empty
// or only fille with <br> and <hr>.
func (ps *Parser) isElementWithoutContent(node *html.Node) bool {
	brs := dom.GetElementsByTagName(node, "br")
	hrs := dom.GetElementsByTagName(node, "hr")
	childs := dom.Children(node)

	return node.Type == html.ElementNode &&
		strings.TrimSpace(dom.TextContent(node)) == "" &&
		(len(childs) == 0 || len(childs) == len(brs)+len(hrs))
}

// hasChildBlockElement determines whether element has any children
// block level elements.
func (ps *Parser) hasChildBlockElement(element *html.Node) bool {
	return ps.someNode(dom.ChildNodes(element), func(node *html.Node) bool {
		return indexOf(divToPElems, dom.TagName(node)) != -1 ||
			ps.hasChildBlockElement(node)
	})
}

// isPhrasingContent determines if a node qualifies as phrasing content.
func (ps *Parser) isPhrasingContent(node *html.Node) bool {
	nodeTagName := dom.TagName(node)
	return node.Type == html.TextNode || indexOf(phrasingElems, nodeTagName) != -1 ||
		((nodeTagName == "a" || nodeTagName == "del" || nodeTagName == "ins") &&
			ps.everyNode(dom.ChildNodes(node), ps.isPhrasingContent))
}

// isWhitespace determines if a node only used as whitespace.
func (ps *Parser) isWhitespace(node *html.Node) bool {
	return (node.Type == html.TextNode && strings.TrimSpace(dom.TextContent(node)) == "") ||
		(node.Type == html.ElementNode && dom.TagName(node) == "br")
}

// getInnerText gets the inner text of a node.
// This also strips * out any excess whitespace to be found.
// In Readability.js, normalizeSpaces default to true.
func (ps *Parser) getInnerText(node *html.Node, normalizeSpaces bool) string {
	textContent := strings.TrimSpace(dom.TextContent(node))
	if normalizeSpaces {
		textContent = rxNormalize.ReplaceAllString(textContent, " ")
	}
	return textContent
}

// getCharCount returns the number of times a string s
// appears in the node.
func (ps *Parser) getCharCount(node *html.Node, s string) int {
	innerText := ps.getInnerText(node, true)
	return strings.Count(innerText, s)
}

// cleanStyles removes the style attribute on every node and under.
func (ps *Parser) cleanStyles(node *html.Node) {
	nodeTagName := dom.TagName(node)
	if node == nil || nodeTagName == "svg" {
		return
	}

	// Remove `style` and deprecated presentational attributes
	for i := 0; i < len(presentationalAttributes); i++ {
		dom.RemoveAttribute(node, presentationalAttributes[i])
	}

	if indexOf(deprecatedSizeAttributeElems, nodeTagName) != -1 {
		dom.RemoveAttribute(node, "width")
		dom.RemoveAttribute(node, "height")
	}

	for child := dom.FirstElementChild(node); child != nil; child = dom.NextElementSibling(child) {
		ps.cleanStyles(child)
	}
}

// getLinkDensity gets the density of links as a percentage of the
// content. This is the amount of text that is inside a link divided
// by the total text in the node.
func (ps *Parser) getLinkDensity(element *html.Node) float64 {
	textLength := charCount(ps.getInnerText(element, true))
	if textLength == 0 {
		return 0
	}

	linkLength := 0
	ps.forEachNode(dom.GetElementsByTagName(element, "a"), func(linkNode *html.Node, _ int) {
		linkLength += charCount(ps.getInnerText(linkNode, true))
	})

	return float64(linkLength) / float64(textLength)
}

// getClassWeight gets an elements class/id weight. Uses regular
// expressions to tell if this element looks good or bad.
func (ps *Parser) getClassWeight(node *html.Node) int {
	if !ps.flags.useWeightClasses {
		return 0
	}

	weight := 0

	// Look for a special classname
	if nodeClassName := dom.ClassName(node); nodeClassName != "" {
		if rxNegative.MatchString(nodeClassName) {
			weight -= 25
		}

		if rxPositive.MatchString(nodeClassName) {
			weight += 25
		}
	}

	// Look for a special ID
	if nodeID := dom.ID(node); nodeID != "" {
		if rxNegative.MatchString(nodeID) {
			weight -= 25
		}

		if rxPositive.MatchString(nodeID) {
			weight += 25
		}
	}

	return weight
}

// clean cleans a node of all elements of type "tag".
// (Unless it's a youtube/vimeo video. People love movies.)
func (ps *Parser) clean(node *html.Node, tag string) {
	isEmbed := indexOf([]string{"object", "embed", "iframe"}, tag) != -1

	ps.removeNodes(dom.GetElementsByTagName(node, tag), func(element *html.Node) bool {
		// Allow youtube and vimeo videos through as people usually want to see those.
		if isEmbed {
			// First, check the elements attributes to see if any of them contain
			// youtube or vimeo
			for _, attr := range element.Attr {
				if rxVideos.MatchString(attr.Val) {
					return false
				}
			}

			// For embed with <object> tag, check inner HTML as well.
			if dom.TagName(element) == "object" && rxVideos.MatchString(dom.InnerHTML(element)) {
				return false
			}
		}
		return true
	})
}

// hasAncestorTag checks if a given node has one of its ancestor tag
// name matching the provided one. In Readability.js, default value
// for maxDepth is 3.
func (ps *Parser) hasAncestorTag(node *html.Node, tag string, maxDepth int, filterFn func(*html.Node) bool) bool {
	depth := 0
	for node.Parent != nil {
		if maxDepth > 0 && depth > maxDepth {
			return false
		}

		if dom.TagName(node.Parent) == tag && (filterFn == nil || filterFn(node.Parent)) {
			return true
		}

		node = node.Parent
		depth++
	}
	return false
}

// getRowAndColumnCount returns how many rows and columns this table has.
func (ps *Parser) getRowAndColumnCount(table *html.Node) (int, int) {
	rows := 0
	columns := 0
	trs := dom.GetElementsByTagName(table, "tr")
	for i := 0; i < len(trs); i++ {
		strRowSpan := dom.GetAttribute(trs[i], "rowspan")
		rowSpan, _ := strconv.Atoi(strRowSpan)
		if rowSpan == 0 {
			rowSpan = 1
		}
		rows += rowSpan

		// Now look for column-related info
		columnsInThisRow := 0
		cells := dom.GetElementsByTagName(trs[i], "td")
		for j := 0; j < len(cells); j++ {
			strColSpan := dom.GetAttribute(cells[j], "colspan")
			colSpan, _ := strconv.Atoi(strColSpan)
			if colSpan == 0 {
				colSpan = 1
			}
			columnsInThisRow += colSpan
		}

		if columnsInThisRow > columns {
			columns = columnsInThisRow
		}
	}

	return rows, columns
}

// markDataTables looks for 'data' (as opposed to 'layout') tables
// and mark it.
func (ps *Parser) markDataTables(root *html.Node) {
	tables := dom.GetElementsByTagName(root, "table")
	for i := 0; i < len(tables); i++ {
		table := tables[i]

		role := dom.GetAttribute(table, "role")
		if role == "presentation" {
			ps.setReadabilityDataTable(table, false)
			continue
		}

		datatable := dom.GetAttribute(table, "datatable")
		if datatable == "0" {
			ps.setReadabilityDataTable(table, false)
			continue
		}

		if dom.HasAttribute(table, "summary") {
			ps.setReadabilityDataTable(table, true)
			continue
		}

		if captions := dom.GetElementsByTagName(table, "caption"); len(captions) > 0 {
			if caption := captions[0]; caption != nil && len(dom.ChildNodes(caption)) > 0 {
				ps.setReadabilityDataTable(table, true)
				continue
			}
		}

		// If the table has a descendant with any of these tags, consider a data table:
		hasDataTableDescendantTags := false
		for _, descendantTag := range []string{"col", "colgroup", "tfoot", "thead", "th"} {
			descendants := dom.GetElementsByTagName(table, descendantTag)
			if len(descendants) > 0 && descendants[0] != nil {
				hasDataTableDescendantTags = true
				break
			}
		}

		if hasDataTableDescendantTags {
			ps.setReadabilityDataTable(table, true)
			continue
		}

		// Nested tables indicates a layout table:
		if len(dom.GetElementsByTagName(table, "table")) > 0 {
			ps.setReadabilityDataTable(table, false)
			continue
		}

		rows, columns := ps.getRowAndColumnCount(table)
		if rows >= 10 || columns > 4 {
			ps.setReadabilityDataTable(table, true)
			continue
		}

		// Now just go by size entirely:
		if rows*columns > 10 {
			ps.setReadabilityDataTable(table, true)
		}
	}
}

// fixLazyImages convert images and figures that have properties like data-src into
// images that can be loaded without JS.
func (ps *Parser) fixLazyImages(root *html.Node) {
	imageNodes := ps.getAllNodesWithTag(root, "img", "picture", "figure")
	ps.forEachNode(imageNodes, func(elem *html.Node, _ int) {
		src := dom.GetAttribute(elem, "src")
		srcset := dom.GetAttribute(elem, "srcset")
		nodeTag := dom.TagName(elem)
		nodeClass := dom.ClassName(elem)

		// In some sites (e.g. Kotaku), they put 1px square image as data uri in
		// the src attribute. So, here we check if the data uri is too short,
		// just might as well remove it.
		if src != "" && rxB64DataURL.MatchString(src) {
			// Make sure it's not SVG, because SVG can have a meaningful image
			// in under 133 bytes.
			parts := rxB64DataURL.FindStringSubmatch(src)
			if parts[1] == "image/svg+xml" {
				return
			}

			// Make sure this element has other attributes which contains
			// image. If it doesn't, then this src is important and
			// shouldn't be removed.
			srcCouldBeRemoved := false
			for _, attr := range elem.Attr {
				if attr.Key == "src" {
					continue
				}

				if rxImgExtensions.MatchString(attr.Val) && isValidURL(attr.Val) {
					srcCouldBeRemoved = true
					break
				}
			}

			// Here we assume if image is less than 100 bytes (or 133B
			// after encoded to base64) it will be too small, therefore
			// it might be placeholder image.
			if srcCouldBeRemoved {
				b64starts := strings.Index(src, "base64") + 7
				b64length := len(src) - b64starts
				if b64length < 133 {
					src = ""
					dom.RemoveAttribute(elem, "src")
				}
			}
		}

		if (src != "" || srcset != "") && !strings.Contains(strings.ToLower(nodeClass), "lazy") {
			return
		}

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

			if copyTo == "" || !isValidURL(attr.Val) {
				continue
			}

			if nodeTag == "img" || nodeTag == "picture" {
				// if this is an img or picture, set the attribute directly
				dom.SetAttribute(elem, copyTo, attr.Val)
			} else if nodeTag == "figure" && len(ps.getAllNodesWithTag(elem, "img", "picture")) == 0 {
				// if the item is a <figure> that does not contain an image or picture,
				// create one and place it inside the figure see the nytimes-3
				// testcase for an example
				img := dom.CreateElement("img")
				dom.SetAttribute(img, copyTo, attr.Val)
				dom.AppendChild(elem, img)
			}
		}
	})
}

// cleanConditionally cleans an element of all tags of type "tag" if
// they look fishy. "Fishy" is an algorithm based on content length,
// classnames, link density, number of images & embeds, etc.
func (ps *Parser) cleanConditionally(element *html.Node, tag string) {
	if !ps.flags.cleanConditionally {
		return
	}

	isList := tag == "ul" || tag == "ol"

	// Gather counts for other typical elements embedded within.
	// Traverse backwards so we can remove nodes at the same time
	// without effecting the traversal.
	ps.removeNodes(dom.GetElementsByTagName(element, tag), func(node *html.Node) bool {
		if tag == "table" && ps.isReadabilityDataTable(node) {
			return false
		}

		if ps.hasAncestorTag(node, "table", -1, ps.isReadabilityDataTable) {
			return false
		}

		weight := ps.getClassWeight(node)
		if weight < 0 {
			return true
		}

		if ps.getCharCount(node, ",") < 10 {
			// If there are not very many commas, and the number of
			// non-paragraph elements is more than paragraphs or other
			// ominous signs, remove the element.
			p := float64(len(dom.GetElementsByTagName(node, "p")))
			img := float64(len(dom.GetElementsByTagName(node, "img")))
			li := float64(len(dom.GetElementsByTagName(node, "li")) - 100)
			input := float64(len(dom.GetElementsByTagName(node, "input")))

			embedCount := 0
			embeds := ps.getAllNodesWithTag(node, "object", "embed", "iframe")

			for _, embed := range embeds {
				// If this embed has attribute that matches video regex,
				// don't delete it.
				for _, attr := range embed.Attr {
					if rxVideos.MatchString(attr.Val) {
						return false
					}
				}

				// For embed with <object> tag, check inner HTML as well.
				if dom.TagName(embed) == "object" && rxVideos.MatchString(dom.InnerHTML(embed)) {
					return false
				}

				embedCount++
			}

			linkDensity := ps.getLinkDensity(node)
			contentLength := charCount(ps.getInnerText(node, true))

			return (img > 1 && p/img < 0.5 && !ps.hasAncestorTag(node, "figure", 3, nil)) ||
				(!isList && li > p) ||
				(input > math.Floor(p/3)) ||
				(!isList && contentLength < 25 && (img == 0 || img > 2) && !ps.hasAncestorTag(node, "figure", 3, nil)) ||
				(!isList && weight < 25 && linkDensity > 0.2) ||
				(weight >= 25 && linkDensity > 0.5) ||
				((embedCount == 1 && contentLength < 75) || embedCount > 1)
		}

		return false
	})
}

// cleanMatchedNodes cleans out elements whose id/class
// combinations match specific string.
func (ps *Parser) cleanMatchedNodes(e *html.Node, filter func(*html.Node, string) bool) {
	endOfSearchMarkerNode := ps.getNextNode(e, true)
	next := ps.getNextNode(e, false)
	for next != nil && next != endOfSearchMarkerNode {
		if filter != nil && filter(next, dom.ClassName(next)+" "+dom.ID(next)) {
			next = ps.removeAndGetNext(next)
		} else {
			next = ps.getNextNode(next, false)
		}
	}
}

// cleanHeaders cleans out spurious headers from an Element.
// Checks things like classnames and link density.
func (ps *Parser) cleanHeaders(e *html.Node) {
	for headerIndex := 1; headerIndex < 3; headerIndex++ {
		headerTag := fmt.Sprintf("h%d", headerIndex)
		ps.removeNodes(dom.GetElementsByTagName(e, headerTag), func(header *html.Node) bool {
			return ps.getClassWeight(header) < 0
		})
	}
}

// isProbablyVisible determines if a node is visible.
func (ps *Parser) isProbablyVisible(node *html.Node) bool {
	nodeStyle := dom.GetAttribute(node, "style")
	nodeAriaHidden := dom.GetAttribute(node, "aria-hidden")
	className := dom.GetAttribute(node, "class")

	// Have to null-check node.style and node.className.indexOf to deal
	// with SVG and MathML nodes. Also check for "fallback-image" so that
	// Wikimedia Math images are displayed
	return (nodeStyle == "" || !rxDisplayNone.MatchString(nodeStyle)) &&
		!dom.HasAttribute(node, "hidden") &&
		(nodeAriaHidden == "" || nodeAriaHidden != "true" || strings.Contains(className, "fallback-image"))
}

// Parse parses input and find the main readable content.
func (ps *Parser) Parse(input io.Reader, pageURL string) (Article, error) {
	// Reset parser data
	ps.articleTitle = ""
	ps.articleByline = ""
	ps.articleDir = ""
	ps.articleSiteName = ""
	ps.attempts = []parseAttempt{}
	ps.flags = flags{
		stripUnlikelys:     true,
		useWeightClasses:   true,
		cleanConditionally: true,
	}

	// Parse page url
	var err error
	ps.documentURI, err = nurl.ParseRequestURI(pageURL)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Parse input
	ps.doc, err = html.Parse(input)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse input: %v", err)
	}

	// Avoid parsing too large documents, as per configuration option
	if ps.MaxElemsToParse > 0 {
		numTags := len(dom.GetElementsByTagName(ps.doc, "*"))
		if numTags > ps.MaxElemsToParse {
			return Article{}, fmt.Errorf("documents too large: %d elements", numTags)
		}
	}

	// ADDITIONAL, not exist in readability.js
	// Unwrap image from noscript
	ps.unwrapNoscriptImages(ps.doc)

	// Remove script tags from the document.
	ps.removeScripts(ps.doc)

	// Prepares the HTML document
	ps.prepDocument()

	// Fetch metadata
	metadata := ps.getArticleMetadata()
	ps.articleTitle = metadata["title"]

	// Try to grab article content
	finalHTMLContent := ""
	finalTextContent := ""
	articleContent := ps.grabArticle()
	var readableNode *html.Node

	if articleContent != nil {
		ps.postProcessContent(articleContent)

		// If we haven't found an excerpt in the article's metadata,
		// use the article's first paragraph as the excerpt. This is used
		// for displaying a preview of the article's content.
		if metadata["excerpt"] == "" {
			paragraphs := dom.GetElementsByTagName(articleContent, "p")
			if len(paragraphs) > 0 {
				metadata["excerpt"] = strings.TrimSpace(dom.TextContent(paragraphs[0]))
			}
		}

		readableNode = dom.FirstElementChild(articleContent)
		finalHTMLContent = dom.InnerHTML(articleContent)
		finalTextContent = dom.TextContent(articleContent)
		finalTextContent = strings.TrimSpace(finalTextContent)
	}

	finalByline := metadata["byline"]
	if finalByline == "" {
		finalByline = ps.articleByline
	}

	// Excerpt is an supposed to be short and concise,
	// so it shouldn't have any new line
	excerpt := strings.TrimSpace(metadata["excerpt"])
	excerpt = strings.Join(strings.Fields(excerpt), " ")

	// go-readability special:
	// Internet is dangerous and weird, and sometimes we will find
	// metadata isn't encoded using a valid Utf-8, so here we check it.
	validTitle := strings.ToValidUTF8(ps.articleTitle, pageURL)
	validByline := strings.ToValidUTF8(finalByline, "")
	validExcerpt := strings.ToValidUTF8(excerpt, "")

	return Article{
		Title:       validTitle,
		Byline:      validByline,
		Node:        readableNode,
		Content:     finalHTMLContent,
		TextContent: finalTextContent,
		Length:      charCount(finalTextContent),
		Excerpt:     validExcerpt,
		SiteName:    metadata["siteName"],
		Image:       metadata["image"],
		Favicon:     metadata["favicon"],
	}, nil
}

// IsReadable decides whether or not the document is reader-able
// without parsing the whole thing. In `mozilla/readability`,
// this method is located in `Readability-readable.js`.
func (ps *Parser) IsReadable(input io.Reader) bool {
	// Parse input
	doc, err := html.Parse(input)
	if err != nil {
		return false
	}

	// Get <p> and <pre> nodes.
	// Also get <div> nodes which have <br> node(s) and append
	// them into the `nodes` variable.
	// Some articles' DOM structures might look like :
	//
	// <div>
	//     Sentences<br>
	//     <br>
	//     Sentences<br>
	// </div>
	//
	// So we need to make sure only fetch the div once.
	// To do so, we will use map as dictionary.
	nodeList := make([]*html.Node, 0)
	nodeDict := make(map[*html.Node]struct{})
	var finder func(*html.Node)

	finder = func(node *html.Node) {
		if node.Type == html.ElementNode {
			tag := dom.TagName(node)
			if tag == "p" || tag == "pre" {
				if _, exist := nodeDict[node]; !exist {
					nodeList = append(nodeList, node)
					nodeDict[node] = struct{}{}
				}
			} else if tag == "br" && node.Parent != nil && dom.TagName(node.Parent) == "div" {
				if _, exist := nodeDict[node.Parent]; !exist {
					nodeList = append(nodeList, node.Parent)
					nodeDict[node.Parent] = struct{}{}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}

	finder(doc)

	// This is a little cheeky, we use the accumulator 'score'
	// to decide what to return from this callback.
	score := float64(0)
	return ps.someNode(nodeList, func(node *html.Node) bool {
		if !ps.isProbablyVisible(node) {
			return false
		}

		matchString := dom.ClassName(node) + " " + dom.ID(node)
		if rxUnlikelyCandidates.MatchString(matchString) &&
			!rxOkMaybeItsACandidate.MatchString(matchString) {
			return false
		}

		if dom.TagName(node) == "p" && ps.hasAncestorTag(node, "li", -1, nil) {
			return false
		}

		nodeText := strings.TrimSpace(dom.TextContent(node))
		nodeTextLength := charCount(nodeText)
		if nodeTextLength < 140 {
			return false
		}

		score += math.Sqrt(float64(nodeTextLength - 140))
		if score > 20 {
			return true
		}

		return false
	})
}

// ====================== INFORMATION ======================
// Methods below these point are not exist in Readability.js.
// They are only used as workaround since Readability.js is
// written in JS which is a dynamic language, while this
// package is written in Go, which is static.
// =========================================================

// getArticleFavicon attempts to get high quality favicon
// that used in article. It will only pick favicon in PNG
// format, so small favicon that uses ico file won't be picked.
// Using algorithm by philippe_b.
func (ps *Parser) getArticleFavicon() string {
	favicon := ""
	faviconSize := -1
	linkElements := dom.GetElementsByTagName(ps.doc, "link")

	ps.forEachNode(linkElements, func(link *html.Node, _ int) {
		linkRel := strings.TrimSpace(dom.GetAttribute(link, "rel"))
		linkType := strings.TrimSpace(dom.GetAttribute(link, "type"))
		linkHref := strings.TrimSpace(dom.GetAttribute(link, "href"))
		linkSizes := strings.TrimSpace(dom.GetAttribute(link, "sizes"))

		if linkHref == "" || !strings.Contains(linkRel, "icon") {
			return
		}

		if linkType != "image/png" && !strings.Contains(linkHref, ".png") {
			return
		}

		size := 0
		for _, sizesLocation := range []string{linkSizes, linkHref} {
			sizeParts := rxFaviconSize.FindStringSubmatch(sizesLocation)
			if len(sizeParts) != 3 || sizeParts[1] != sizeParts[2] {
				continue
			}

			size, _ = strconv.Atoi(sizeParts[1])
			break
		}

		if size > faviconSize {
			faviconSize = size
			favicon = linkHref
		}
	})

	return toAbsoluteURI(favicon, ps.documentURI)
}

// removeComments find all comments in document then remove it.
func (ps *Parser) removeComments(doc *html.Node) {
	// Find all comments
	var comments []*html.Node
	var finder func(*html.Node)

	finder = func(node *html.Node) {
		if node.Type == html.CommentNode {
			comments = append(comments, node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}

	for child := doc.FirstChild; child != nil; child = child.NextSibling {
		finder(child)
	}

	// Remove it
	ps.removeNodes(comments, nil)
}

// In dynamic language like JavaScript, we can easily add new
// property to an existing object by simply writing :
//
//   obj.newProperty = newValue
//
// This is extensively used in Readability.js to save readability
// content score; and to mark whether a table is data container or
// only used for layout.
//
// However, since Go is static typed, we can't do it that way.
// As workaround, we just saved those data as attribute in the
// HTML nodes. Hence why these methods exists.

// setReadabilityDataTable marks whether a Node is data table or not.
func (ps *Parser) setReadabilityDataTable(node *html.Node, isDataTable bool) {
	if isDataTable {
		dom.SetAttribute(node, "data-readability-table", "true")
	} else {
		dom.RemoveAttribute(node, "data-readability-table")
	}
}

// isReadabilityDataTable determines if node is data table.
func (ps *Parser) isReadabilityDataTable(node *html.Node) bool {
	return dom.HasAttribute(node, "data-readability-table")
}

// setContentScore sets the readability score for a node.
func (ps *Parser) setContentScore(node *html.Node, score float64) {
	dom.SetAttribute(node, "data-readability-score", fmt.Sprintf("%.4f", score))
}

// hasContentScore checks if node has readability score.
func (ps *Parser) hasContentScore(node *html.Node) bool {
	return dom.HasAttribute(node, "data-readability-score")
}

// getContentScore gets the readability score of a node.
func (ps *Parser) getContentScore(node *html.Node) float64 {
	strScore := dom.GetAttribute(node, "data-readability-score")
	strScore = strings.TrimSpace(strScore)
	if strScore == "" {
		return 0
	}

	score, _ := strconv.ParseFloat(strScore, 64)
	return score
}

// clearReadabilityAttr removes Readability attribute that
// created by this package. Used in `postProcessContent`.
func (ps *Parser) clearReadabilityAttr(node *html.Node) {
	dom.RemoveAttribute(node, "data-readability-score")
	dom.RemoveAttribute(node, "data-readability-table")

	for child := dom.FirstElementChild(node); child != nil; child = dom.NextElementSibling(child) {
		ps.clearReadabilityAttr(child)
	}
}
