package readability

import (
	"bytes"
	"fmt"
	ghtml "html"
	"io"
	"math"
	"net/http"
	nurl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	wl "github.com/abadojack/whatlanggo"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	dataTableAttr          = "XXX-DATA-TABLE"
	rxSpaces               = regexp.MustCompile(`(?is)\s{2,}|\n+`)
	rxReplaceBrs           = regexp.MustCompile(`(?is)(<br[^>]*>[ \n\r\t]*){2,}`)
	rxByline               = regexp.MustCompile(`(?is)byline|author|dateline|writtenby|p-author`)
	rxUnlikelyCandidates   = regexp.MustCompile(`(?is)banner|breadcrumbs|combx|comment|community|cover-wrap|disqus|extra|foot|header|legends|menu|related|remark|replies|rss|shoutbox|sidebar|skyscraper|social|sponsor|supplemental|ad-break|agegate|pagination|pager|popup|yom-remote`)
	rxOkMaybeItsACandidate = regexp.MustCompile(`(?is)and|article|body|column|main|shadow`)
	rxUnlikelyElements     = regexp.MustCompile(`(?is)(input|time|button)`)
	rxDivToPElements       = regexp.MustCompile(`(?is)<(a|blockquote|dl|div|img|ol|p|pre|table|ul|select)`)
	rxPositive             = regexp.MustCompile(`(?is)article|body|content|entry|hentry|h-entry|main|page|pagination|post|text|blog|story`)
	rxNegative             = regexp.MustCompile(`(?is)hidden|^hid$| hid$| hid |^hid |banner|combx|comment|com-|contact|foot|footer|footnote|masthead|media|meta|outbrain|promo|related|scroll|share|shoutbox|sidebar|skyscraper|sponsor|shopping|tags|tool|widget`)
	rxPIsSentence          = regexp.MustCompile(`(?is)\.( |$)`)
	rxVideos               = regexp.MustCompile(`(?is)//(www\.)?(dailymotion|youtube|youtube-nocookie|player\.vimeo)\.com`)
	rxKillBreaks           = regexp.MustCompile(`(?is)(<br\s*/?>(\s|&nbsp;?)*)+`)
	rxComments             = regexp.MustCompile(`(?is)<!--[^>]+-->`)
)

type candidateItem struct {
	score float64
	node  *goquery.Selection
}

type readability struct {
	html       string
	url        *nurl.URL
	candidates map[string]candidateItem
}

// Metadata is metadata of an article
type Metadata struct {
	Title       string
	Image       string
	Excerpt     string
	Author      string
	MinReadTime int
	MaxReadTime int
}

// Article is the content of an URL
type Article struct {
	URL        string
	Meta       Metadata
	Content    string
	RawContent string
}

// removeScripts removes script tags from the document.
func removeScripts(doc *goquery.Document) {
	doc.Find("script").Remove()
	doc.Find("noscript").Remove()
}

// replaceBrs replaces 2 or more successive <br> elements with a single <p>.
// Whitespace between <br> elements are ignored. For example:
//   <div>foo<br>bar<br> <br><br>abc</div>
// will become:
//   <div>foo<br>bar<p>abc</p></div>
func replaceBrs(doc *goquery.Document) {
	// Remove BRs in body
	body := doc.Find("body")

	html, _ := body.Html()
	html = rxReplaceBrs.ReplaceAllString(html, "</p><p>")

	body.SetHtml(html)

	// Remove empty p
	body.Find("p").Each(func(_ int, p *goquery.Selection) {
		html, _ := p.Html()
		html = strings.TrimSpace(html)
		if html == "" {
			p.Remove()
		}
	})
}

// prepDocument prepares the HTML document for readability to scrape it.
// This includes things like stripping JS, CSS, and handling terrible markup.
func prepDocument(doc *goquery.Document) {
	// Remove all style tags in head
	doc.Find("style").Remove()

	// Replace all br
	replaceBrs(doc)

	// Replace font tags to span
	doc.Find("font").Each(func(_ int, font *goquery.Selection) {
		html, _ := font.Html()
		font.ReplaceWithHtml("<span>" + html + "</span>")
	})
}

// getArticleTitle fetchs the article title
func getArticleTitle(doc *goquery.Document) string {
	// Get title tag
	title := doc.Find("title").First().Text()
	title = normalizeText(title)
	originalTitle := title

	// Create list of separator
	separators := []string{`|`, `-`, `\`, `/`, `>`, `»`}
	hierarchialSeparators := []string{`\`, `/`, `>`, `»`}

	// If there's a separator in the title, first remove the final part
	titleHadHierarchicalSeparators := false
	if idx, sep := findSeparator(title, separators...); idx != -1 {
		titleHadHierarchicalSeparators = hasSeparator(title, hierarchialSeparators...)

		index := strings.LastIndex(originalTitle, sep)
		title = originalTitle[:index]

		// If the resulting title is too short (3 words or fewer), remove
		// the first part instead:
		if len(strings.Fields(title)) < 3 {
			index = strings.Index(originalTitle, sep)
			title = originalTitle[index+1:]
		}
	} else if strings.Contains(title, ": ") {
		// Check if we have an heading containing this exact string, so we
		// could assume it's the full title.
		existInHeading := false
		doc.Find("h1,h2").EachWithBreak(func(_ int, heading *goquery.Selection) bool {
			headingText := strings.TrimSpace(heading.Text())
			if headingText == title {
				existInHeading = true
				return false
			}

			return true
		})

		// If we don't, let's extract the title out of the original title string.
		if !existInHeading {
			index := strings.LastIndex(originalTitle, ":")
			title = originalTitle[index+1:]

			// If the title is now too short, try the first colon instead:
			if len(strings.Fields(title)) < 3 {
				index = strings.Index(originalTitle, ":")
				title = originalTitle[:index]
				// But if we have too many words before the colon there's something weird
				// with the titles and the H tags so let's just use the original title instead
			} else {
				index = strings.Index(originalTitle, ":")
				beforeColon := originalTitle[:index]
				if len(strings.Fields(beforeColon)) > 5 {
					title = originalTitle
				}
			}
		}
	} else if strLen(title) > 150 || strLen(title) < 15 {
		hOne := doc.Find("h1").First()
		if hOne != nil {
			title = hOne.Text()
		}
	}

	// If we now have 4 words or fewer as our title, and either no
	// 'hierarchical' separators (\, /, > or ») were found in the original
	// title or we decreased the number of words by more than 1 word, use
	// the original title.
	curTitleWordCount := len(strings.Fields(title))
	noSeparatorWordCount := len(strings.Fields(removeSeparator(originalTitle, separators...)))
	if curTitleWordCount <= 4 && (!titleHadHierarchicalSeparators || curTitleWordCount != noSeparatorWordCount-1) {
		title = originalTitle
	}

	return normalizeText(title)
}

// getArticleMetadata attempts to get excerpt and byline metadata for the article.
func getArticleMetadata(doc *goquery.Document) Metadata {
	metadata := Metadata{}
	mapAttribute := make(map[string]string)

	doc.Find("meta").Each(func(_ int, meta *goquery.Selection) {
		metaName, _ := meta.Attr("name")
		metaProperty, _ := meta.Attr("property")
		metaContent, _ := meta.Attr("content")

		metaName = strings.TrimSpace(metaName)
		metaProperty = strings.TrimSpace(metaProperty)
		metaContent = strings.TrimSpace(metaContent)

		// Fetch author name
		if strings.Contains(metaName+metaProperty, "author") {
			metadata.Author = metaContent
			return
		}

		// Fetch description and title
		if metaName == "title" ||
			metaName == "description" ||
			metaName == "twitter:title" ||
			metaName == "twitter:image" ||
			metaName == "twitter:description" {
			if _, exist := mapAttribute[metaName]; !exist {
				mapAttribute[metaName] = metaContent
			}
			return
		}

		if metaProperty == "og:description" ||
			metaProperty == "og:image" ||
			metaProperty == "og:title" {
			if _, exist := mapAttribute[metaProperty]; !exist {
				mapAttribute[metaProperty] = metaContent
			}
			return
		}
	})

	// Set final image
	if _, exist := mapAttribute["og:image"]; exist {
		metadata.Image = mapAttribute["og:image"]
	} else if _, exist := mapAttribute["twitter:image"]; exist {
		metadata.Image = mapAttribute["twitter:image"]
	}

	if metadata.Image != "" && strings.HasPrefix(metadata.Image, "//") {
		metadata.Image = "http:" + metadata.Image
	}

	// Set final excerpt
	if _, exist := mapAttribute["description"]; exist {
		metadata.Excerpt = mapAttribute["description"]
	} else if _, exist := mapAttribute["og:description"]; exist {
		metadata.Excerpt = mapAttribute["og:description"]
	} else if _, exist := mapAttribute["twitter:description"]; exist {
		metadata.Excerpt = mapAttribute["twitter:description"]
	}

	// Set final title
	metadata.Title = getArticleTitle(doc)
	if metadata.Title == "" {
		if _, exist := mapAttribute["og:title"]; exist {
			metadata.Title = mapAttribute["og:title"]
		} else if _, exist := mapAttribute["twitter:title"]; exist {
			metadata.Title = mapAttribute["twitter:title"]
		}
	}

	// Clean up the metadata
	metadata.Title = normalizeText(metadata.Title)
	metadata.Excerpt = normalizeText(metadata.Excerpt)

	return metadata
}

// isValidByline checks whether the input string could be a byline.
// This verifies that the input is a string, and that the length
// is less than 100 chars.
func isValidByline(str string) bool {
	return strLen(str) > 0 && strLen(str) < 100
}

func isElementWithoutContent(s *goquery.Selection) bool {
	if s == nil {
		return true
	}

	html, _ := s.Html()
	html = strings.TrimSpace(html)
	return html == ""
}

// hasSinglePInsideElement checks if this node has only whitespace and a single P element.
// Returns false if the DIV node contains non-empty text nodes
// or if it contains no P or more than 1 element.
func hasSinglePInsideElement(s *goquery.Selection) bool {
	// There should be exactly 1 element child which is a P
	return s.Children().Length() == 1 && s.Children().First().Is("p")
}

// hasChildBlockElement determines whether element has any children
// block level elements.
func hasChildBlockElement(s *goquery.Selection) bool {
	html, _ := s.Html()
	return rxDivToPElements.MatchString(html)
}

func setNodeTag(s *goquery.Selection, tag string) {
	html, _ := s.Html()
	newHTML := fmt.Sprintf("<%s>%s</%s>", tag, html, tag)
	s.ReplaceWithHtml(newHTML)
}

func getNodeAncestors(node *goquery.Selection, maxDepth int) []*goquery.Selection {
	ancestors := []*goquery.Selection{}
	parent := node

	for i := 0; i < maxDepth; i++ {
		parent = parent.Parent()
		if len(parent.Nodes) == 0 {
			return ancestors
		}

		ancestors = append(ancestors, parent)
	}

	return ancestors
}

func hasAncestorTag(node *goquery.Selection, tag string, maxDepth int) (*goquery.Selection, bool) {
	parent := node

	if maxDepth < 0 {
		maxDepth = 100
	}

	for i := 0; i < maxDepth; i++ {
		parent = parent.Parent()
		if len(parent.Nodes) == 0 {
			break
		}

		if parent.Is(tag) {
			return parent, true
		}
	}

	return nil, false
}

// initializeNodeScore initializes a node and checks the className/id
// for special names to add to its score.
func initializeNodeScore(node *goquery.Selection) candidateItem {
	contentScore := 0.0
	tagName := goquery.NodeName(node)
	switch strings.ToLower(tagName) {
	case "article":
		contentScore += 10
	case "section":
		contentScore += 8
	case "div":
		contentScore += 5
	case "pre", "blockquote", "td":
		contentScore += 3
	case "form", "ol", "ul", "dl", "dd", "dt", "li", "address":
		contentScore -= 3
	case "th", "h1", "h2", "h3", "h4", "h5", "h6":
		contentScore -= 5
	}

	contentScore += getClassWeight(node)
	return candidateItem{contentScore, node}
}

// getClassWeight gets an elements class/id weight.
// Uses regular expressions to tell if this element looks good or bad.
func getClassWeight(node *goquery.Selection) float64 {
	weight := 0.0
	if str, b := node.Attr("class"); b {
		if rxNegative.MatchString(str) {
			weight -= 25
		}

		if rxPositive.MatchString(str) {
			weight += 25
		}
	}

	if str, b := node.Attr("id"); b {
		if rxNegative.MatchString(str) {
			weight -= 25
		}

		if rxPositive.MatchString(str) {
			weight += 25
		}
	}

	return weight
}

// getLinkDensity gets the density of links as a percentage of the content
// This is the amount of text that is inside a link divided by the total text in the node.
func getLinkDensity(node *goquery.Selection) float64 {
	textLength := strLen(normalizeText(node.Text()))
	if textLength == 0 {
		return 0
	}

	linkLength := 0
	node.Find("a").Each(func(_ int, link *goquery.Selection) {
		linkLength += strLen(link.Text())
	})

	return float64(linkLength) / float64(textLength)
}

// Remove the style attribute on every e and under.
func cleanStyle(s *goquery.Selection) {
	s.Find("*").Each(func(i int, s1 *goquery.Selection) {
		tagName := goquery.NodeName(s1)
		if strings.ToLower(tagName) == "svg" {
			return
		}

		s1.RemoveAttr("align")
		s1.RemoveAttr("background")
		s1.RemoveAttr("bgcolor")
		s1.RemoveAttr("border")
		s1.RemoveAttr("cellpadding")
		s1.RemoveAttr("cellspacing")
		s1.RemoveAttr("frame")
		s1.RemoveAttr("hspace")
		s1.RemoveAttr("rules")
		s1.RemoveAttr("style")
		s1.RemoveAttr("valign")
		s1.RemoveAttr("vspace")
		s1.RemoveAttr("onclick")
		s1.RemoveAttr("onmouseover")
		s1.RemoveAttr("border")
		s1.RemoveAttr("style")

		if tagName != "table" && tagName != "th" && tagName != "td" &&
			tagName != "hr" && tagName != "pre" {
			s1.RemoveAttr("width")
			s1.RemoveAttr("height")
		}
	})
}

// Return an object indicating how many rows and columns this table has.
func getTableRowAndColumnCount(table *goquery.Selection) (int, int) {
	rows := 0
	columns := 0
	table.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		// Look for rows
		strRowSpan, _ := tr.Attr("rowspan")
		rowSpan, err := strconv.Atoi(strRowSpan)
		if err != nil {
			rowSpan = 1
		}
		rows += rowSpan

		// Now look for columns
		columnInThisRow := 0
		tr.Find("td").Each(func(_ int, td *goquery.Selection) {
			strColSpan, _ := tr.Attr("colspan")
			colSpan, err := strconv.Atoi(strColSpan)
			if err != nil {
				colSpan = 1
			}
			columnInThisRow += colSpan
		})

		if columnInThisRow > columns {
			columns = columnInThisRow
		}
	})

	return rows, columns
}

// Look for 'data' (as opposed to 'layout') tables
func markDataTables(s *goquery.Selection) {
	s.Find("table").Each(func(_ int, table *goquery.Selection) {
		role, _ := table.Attr("role")
		if role == "presentation" {
			return
		}

		datatable, _ := table.Attr("datatable")
		if datatable == "0" {
			return
		}

		_, summaryExist := table.Attr("summary")
		if summaryExist {
			table.SetAttr(dataTableAttr, "1")
			return
		}

		caption := table.Find("caption").First()
		if len(caption.Nodes) > 0 && caption.Children().Length() > 0 {
			table.SetAttr(dataTableAttr, "1")
			return
		}

		// If the table has a descendant with any of these tags, consider a data table:
		dataTableDescendants := []string{"col", "colgroup", "tfoot", "thead", "th"}
		for _, tag := range dataTableDescendants {
			if table.Find(tag).Length() > 0 {
				table.SetAttr(dataTableAttr, "1")
				return
			}
		}

		// Nested tables indicate a layout table:
		if table.Find("table").Length() > 0 {
			return
		}

		nRow, nColumn := getTableRowAndColumnCount(table)
		if nRow >= 10 || nColumn > 4 {
			table.SetAttr(dataTableAttr, "1")
			return
		}

		// Now just go by size entirely:
		if nRow*nColumn > 10 {
			table.SetAttr(dataTableAttr, "1")
			return
		}
	})
}

// Clean an element of all tags of type "tag" if they look fishy.
// "Fishy" is an algorithm based on content length, classnames, link density, number of images & embeds, etc.
func cleanConditionally(e *goquery.Selection, tag string) {
	isList := tag == "ul" || tag == "ol"

	e.Find(tag).Each(func(i int, node *goquery.Selection) {
		// First check if we're in a data table, in which case don't remove it
		if ancestor, hasTag := hasAncestorTag(node, "table", -1); hasTag {
			if attr, _ := ancestor.Attr(dataTableAttr); attr == "1" {
				return
			}
		}

		// If it is table, remove data table marker
		if tag == "table" {
			node.RemoveAttr(dataTableAttr)
		}

		contentScore := 0.0
		weight := getClassWeight(node)
		if weight+contentScore < 0 {
			node.Remove()
			return
		}

		// If there are not very many commas, and the number of
		// non-paragraph elements is more than paragraphs or other
		// ominous signs, remove the element.
		nodeText := normalizeText(node.Text())
		nCommas := strings.Count(nodeText, ",")
		nCommas += strings.Count(nodeText, "，")
		if nCommas < 10 {
			p := node.Find("p").Length()
			img := node.Find("img").Length()
			li := node.Find("li").Length() - 100
			input := node.Find("input").Length()

			embedCount := 0
			node.Find("embed").Each(func(i int, embed *goquery.Selection) {
				if !rxVideos.MatchString(embed.AttrOr("src", "")) {
					embedCount++
				}
			})

			contentLength := strLen(nodeText)
			linkDensity := getLinkDensity(node)
			_, hasFigureAncestor := hasAncestorTag(node, "figure", 3)

			haveToRemove := (!isList && li > p) ||
				(img > 1 && float64(p)/float64(img) < 0.5 && !hasFigureAncestor) ||
				(float64(input) > math.Floor(float64(p)/3)) ||
				(!isList && contentLength < 25 && (img == 0 || img > 2) && !hasFigureAncestor) ||
				(!isList && weight < 25 && linkDensity > 0.2) ||
				(weight >= 25 && linkDensity > 0.5) ||
				((embedCount == 1 && contentLength < 75) || embedCount > 1)

			if haveToRemove {
				node.Remove()
			}
		}
	})
}

// Clean a node of all elements of type "tag".
// (Unless it's a youtube/vimeo video. People love movies.)
func clean(s *goquery.Selection, tag string) {
	isEmbed := tag == "object" || tag == "embed" || tag == "iframe"

	s.Find(tag).Each(func(i int, target *goquery.Selection) {
		attributeValues := ""
		for _, attribute := range target.Nodes[0].Attr {
			attributeValues += " " + attribute.Val
		}

		if isEmbed && rxVideos.MatchString(attributeValues) {
			return
		}

		if isEmbed && rxVideos.MatchString(target.Text()) {
			return
		}

		target.Remove()
	})
}

// Clean out spurious headers from an Element. Checks things like classnames and link density.
func cleanHeaders(s *goquery.Selection) {
	s.Find("h1,h2,h3").Each(func(_ int, s1 *goquery.Selection) {
		if getClassWeight(s1) < 0 {
			s1.Remove()
		}
	})
}

// Prepare the article node for display. Clean out any inline styles,
// iframes, forms, strip extraneous <p> tags, etc.
func prepArticle(articleContent *goquery.Selection, articleTitle string) {
	if articleContent == nil {
		return
	}

	// Check for data tables before we continue, to avoid removing items in
	// those tables, which will often be isolated even though they're
	// visually linked to other content-ful elements (text, images, etc.).
	markDataTables(articleContent)

	// Remove style attribute
	cleanStyle(articleContent)

	// Clean out junk from the article content
	cleanConditionally(articleContent, "form")
	cleanConditionally(articleContent, "fieldset")
	clean(articleContent, "h1")
	clean(articleContent, "object")
	clean(articleContent, "embed")
	clean(articleContent, "footer")
	clean(articleContent, "link")

	// Clean out elements have "share" in their id/class combinations from final top candidates,
	// which means we don't remove the top candidates even they have "share".
	articleContent.Find("*").Each(func(_ int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		class, _ := s.Attr("class")
		matchString := class + " " + id
		if strings.Contains(matchString, "share") {
			s.Remove()
		}
	})

	// If there is only one h2 and its text content substantially equals article title,
	// they are probably using it as a header and not a subheader,
	// so remove it since we already extract the title separately.
	h2s := articleContent.Find("h2")
	if h2s.Length() == 1 {
		h2 := h2s.First()
		h2Text := normalizeText(h2.Text())
		lengthSimilarRate := float64(strLen(h2Text)-strLen(articleTitle)) /
			float64(strLen(articleTitle))

		if math.Abs(lengthSimilarRate) < 0.5 {
			titlesMatch := false
			if lengthSimilarRate > 0 {
				titlesMatch = strings.Contains(h2Text, articleTitle)
			} else {
				titlesMatch = strings.Contains(articleTitle, h2Text)
			}

			if titlesMatch {
				h2.Remove()
			}
		}
	}

	clean(articleContent, "iframe")
	clean(articleContent, "input")
	clean(articleContent, "textarea")
	clean(articleContent, "select")
	clean(articleContent, "button")
	cleanHeaders(articleContent)

	// Do these last as the previous stuff may have removed junk
	// that will affect these
	cleanConditionally(articleContent, "table")
	cleanConditionally(articleContent, "ul")
	cleanConditionally(articleContent, "div")

	// Remove extra paragraphs
	// At this point, nasty iframes have been removed, only remain embedded video ones.
	articleContent.Find("p").Each(func(_ int, p *goquery.Selection) {
		imgCount := p.Find("img").Length()
		embedCount := p.Find("embed").Length()
		objectCount := p.Find("object").Length()
		iframeCount := p.Find("iframe").Length()
		totalCount := imgCount + embedCount + objectCount + iframeCount

		pText := normalizeText(p.Text())
		if totalCount == 0 && strLen(pText) == 0 {
			p.Remove()
		}
	})

	articleContent.Find("br").Each(func(_ int, br *goquery.Selection) {
		if br.Next().Is("p") {
			br.Remove()
		}
	})
}

// grabArticle fetch the articles using a variety of metrics (content score, classname, element types),
// find the content that is most likely to be the stuff a user wants to read.
// Then return it wrapped up in a div.
func grabArticle(doc *goquery.Document, articleTitle string) (*goquery.Selection, string) {
	// Create initial variable
	author := ""
	elementsToScore := []*goquery.Selection{}

	// First, node prepping. Trash nodes that look cruddy (like ones with the
	// class name "comment", etc), and turn divs into P tags where they have been
	// used inappropriately (as in, where they contain no other block level elements.)
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		matchString := s.AttrOr("class", "") + " " + s.AttrOr("id", "")

		// If byline, remove this element
		if rel := s.AttrOr("rel", ""); rel == "author" || rxByline.MatchString(matchString) {
			text := s.Text()
			text = strings.TrimSpace(text)
			if isValidByline(text) {
				author = text
				s.Remove()
				return
			}
		}

		// Remove unlikely candidates
		if rxUnlikelyCandidates.MatchString(matchString) &&
			!rxOkMaybeItsACandidate.MatchString(matchString) &&
			!s.Is("body") && !s.Is("a") {
			s.Remove()
			return
		}

		if rxUnlikelyElements.MatchString(goquery.NodeName(s)) {
			s.Remove()
			return
		}

		// Remove DIV, SECTION, and HEADER nodes without any content(e.g. text, image, video, or iframe).
		if s.Is("div,section,header,h1,h2,h3,h4,h5,h6") && isElementWithoutContent(s) {
			s.Remove()
			return
		}

		if s.Is("section,h2,h3,h4,h5,h6,p,td,pre") {
			elementsToScore = append(elementsToScore, s)
		}

		// Turn all divs that don't have children block level elements into p's
		if s.Is("div") {
			// Sites like http://mobile.slate.com encloses each paragraph with a DIV
			// element. DIVs with only a P element inside and no text content can be
			// safely converted into plain P elements to avoid confusing the scoring
			// algorithm with DIVs with are, in practice, paragraphs.
			if hasSinglePInsideElement(s) {
				newNode := s.Children().First()
				s.ReplaceWithSelection(newNode)
				elementsToScore = append(elementsToScore, s)
			} else if !hasChildBlockElement(s) {
				setNodeTag(s, "p")
				elementsToScore = append(elementsToScore, s)
			}
		}
	})

	// Loop through all paragraphs, and assign a score to them based on how content-y they look.
	// Then add their score to their parent node.
	// A score is determined by things like number of commas, class names, etc. Maybe eventually link density.
	candidates := make(map[string]candidateItem)
	for _, s := range elementsToScore {
		// If this paragraph is less than 25 characters, don't even count it.
		innerText := normalizeText(s.Text())
		if strLen(innerText) < 25 {
			continue
		}

		// Exclude nodes with no ancestor.
		ancestors := getNodeAncestors(s, 3)
		if len(ancestors) == 0 {
			continue
		}

		// Calculate content score
		// Add a point for the paragraph itself as a base.
		contentScore := 1.0

		// Add points for any commas within this paragraph.
		contentScore += float64(strings.Count(innerText, ","))
		contentScore += float64(strings.Count(innerText, "，"))

		// For every 100 characters in this paragraph, add another point. Up to 3 points.
		contentScore += math.Min(math.Floor(float64(strLen(innerText)/100)), 3)

		// Initialize and score ancestors.
		for level, ancestor := range ancestors {
			// Node score divider:
			// - parent:             1 (no division)
			// - grandparent:        2
			// - great grandparent+: ancestor level * 3
			scoreDivider := 0
			if level == 0 {
				scoreDivider = 1
			} else if level == 1 {
				scoreDivider = 2
			} else {
				scoreDivider = level * 3
			}

			ancestorHash := hashNode(ancestor)
			if _, ok := candidates[ancestorHash]; !ok {
				candidates[ancestorHash] = initializeNodeScore(ancestor)
			}

			candidate := candidates[ancestorHash]
			candidate.score += contentScore / float64(scoreDivider)
			candidates[ancestorHash] = candidate
		}
	}

	// Scale the final candidates score based on link density. Good content
	// should have a relatively small link density (5% or less) and be mostly
	// unaffected by this operation.
	topCandidate := candidateItem{}
	for hash, candidate := range candidates {
		candidate.score = candidate.score * (1 - getLinkDensity(candidate.node))
		candidates[hash] = candidate

		if topCandidate.node == nil || candidate.score > topCandidate.score {
			topCandidate = candidate
		}
	}

	// If we still have no top candidate, use the body as a last resort.
	if topCandidate.node == nil {
		body := doc.Find("body").First()

		bodyHTML, _ := body.Html()
		newHTML := fmt.Sprintf(`<div id="xxx-readability-body">%s<div>`, bodyHTML)
		body.AppendHtml(newHTML)

		tempReadabilityBody := body.Find("div#xxx-readability-body").First()
		tempReadabilityBody.RemoveAttr("id")

		tempHash := hashNode(tempReadabilityBody)
		if _, ok := candidates[tempHash]; !ok {
			candidates[tempHash] = initializeNodeScore(tempReadabilityBody)
		}

		topCandidate = candidates[tempHash]
	}

	// Create new document to save the final article content.
	reader := strings.NewReader(`<div id="readability-content"></div>`)
	newDoc, _ := goquery.NewDocumentFromReader(reader)
	articleContent := newDoc.Find("div#readability-content").First()

	// Now that we have the top candidate, look through its siblings for content
	// that might also be related. Things like preambles, content split by ads
	// that we removed, etc.
	topCandidateClass, _ := topCandidate.node.Attr("class")
	siblingScoreThreshold := math.Max(10.0, topCandidate.score*0.2)
	topCandidate.node.Parent().Children().Each(func(_ int, sibling *goquery.Selection) {
		appendSibling := false

		if sibling.IsSelection(topCandidate.node) {
			appendSibling = true
		} else {
			contentBonus := 0.0
			siblingClass, _ := sibling.Attr("class")
			if siblingClass == topCandidateClass && topCandidateClass != "" {
				contentBonus += topCandidate.score * 0.2
			}

			siblingHash := hashNode(sibling)
			if item, ok := candidates[siblingHash]; ok && item.score > siblingScoreThreshold {
				appendSibling = true
			} else if sibling.Is("p") {
				linkDensity := getLinkDensity(sibling)
				nodeContent := normalizeText(sibling.Text())
				nodeLength := strLen(nodeContent)

				if nodeLength > 80 && linkDensity < 0.25 {
					appendSibling = true
				} else if nodeLength < 80 && nodeLength > 0 &&
					linkDensity == 0 && rxPIsSentence.MatchString(nodeContent) {
					appendSibling = true
				}
			}
		}

		if appendSibling {
			articleContent.AppendSelection(sibling)
		}
	})

	// So we have all of the content that we need.
	// Now we clean it up for presentation.
	prepArticle(articleContent, articleTitle)

	return articleContent, author
}

// Convert relative uri to absolute
func toAbsoluteURI(uri string, base *nurl.URL) string {
	if uri == "" || base == nil {
		return ""
	}

	// If it is hash tag, return as it is
	if uri[0:1] == "#" {
		return uri
	}

	// If it is already an absolute URL, return as it is
	tempURI, err := nurl.ParseRequestURI(uri)
	if err == nil && len(tempURI.Scheme) == 0 {
		return uri
	}

	// Otherwise, put it as path of base URL
	newURI := nurl.URL(*base)
	newURI.Path = uri

	return newURI.String()
}

// Converts each <a> and <img> uri in the given element to an absolute URI,
// ignoring #ref URIs.
func fixRelativeURIs(articleContent *goquery.Selection, base *nurl.URL) {
	articleContent.Find("a").Each(func(_ int, a *goquery.Selection) {
		if href, exist := a.Attr("href"); exist {
			// Replace links with javascript: URIs with text content, since
			// they won't work after scripts have been removed from the page.
			if strings.HasPrefix(href, "javascript:") {
				text := a.Text()
				a.ReplaceWithHtml(text)
			} else {
				a.SetAttr("href", toAbsoluteURI(href, base))
			}
		}
	})

	articleContent.Find("img").Each(func(_ int, img *goquery.Selection) {
		if src, exist := img.Attr("src"); exist {
			img.SetAttr("src", toAbsoluteURI(src, base))
		}
	})
}

func postProcessContent(articleContent *goquery.Selection, uri *nurl.URL) {
	// Readability cannot open relative uris so we convert them to absolute uris.
	fixRelativeURIs(articleContent, uri)

	// Last time, clean all empty tags and remove id and class name
	articleContent.Find("*").Each(func(_ int, s *goquery.Selection) {
		html, _ := s.Html()
		html = strings.TrimSpace(html)
		if html == "" {
			s.Remove()
		}

		s.RemoveAttr("class")
		s.RemoveAttr("id")
	})
}

// getHTMLContent fetch and cleans the raw html from article
func getHTMLContent(articleContent *goquery.Selection) string {
	html, err := articleContent.Html()
	if err != nil {
		return ""
	}

	html = ghtml.UnescapeString(html)
	html = rxComments.ReplaceAllString(html, "")
	html = rxKillBreaks.ReplaceAllString(html, "<br />")
	html = rxSpaces.ReplaceAllString(html, " ")
	return html
}

// getTextContent fetch and cleans the text from article
func getTextContent(articleContent *goquery.Selection) string {
	var buf bytes.Buffer

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			nodeText := normalizeText(n.Data)
			if nodeText != "" {
				buf.WriteString(nodeText)
			}
		} else if n.Parent != nil && n.Parent.DataAtom != atom.P {
			buf.WriteString("|X|")
		}

		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}

	for _, n := range articleContent.Nodes {
		f(n)
	}

	finalContent := ""
	paragraphs := strings.Split(buf.String(), "|X|")
	for _, paragraph := range paragraphs {
		if paragraph != "" {
			finalContent += paragraph + "\n\n"
		}
	}

	finalContent = strings.TrimSpace(finalContent)
	return finalContent
}

// Estimate read time based on the language number of character in contents.
// Using data from http://iovs.arvojournals.org/article.aspx?articleid=2166061
func estimateReadTime(articleContent *goquery.Selection) (int, int) {
	if articleContent == nil {
		return 0, 0
	}

	// Check the language
	contentText := normalizeText(articleContent.Text())
	lang := wl.LangToString(wl.DetectLang(contentText))

	// Get number of words and images
	nChar := strLen(contentText)
	nImg := articleContent.Find("img").Length()
	if nChar == 0 && nImg == 0 {
		return 0, 0
	}

	// Calculate character per minute by language
	// Fallback to english
	var cpm, sd float64
	switch lang {
	case "arb":
		sd = 88
		cpm = 612
	case "nld":
		sd = 143
		cpm = 978
	case "fin":
		sd = 121
		cpm = 1078
	case "fra":
		sd = 126
		cpm = 998
	case "deu":
		sd = 86
		cpm = 920
	case "heb":
		sd = 130
		cpm = 833
	case "ita":
		sd = 140
		cpm = 950
	case "jpn":
		sd = 56
		cpm = 357
	case "pol":
		sd = 126
		cpm = 916
	case "por":
		sd = 145
		cpm = 913
	case "rus":
		sd = 175
		cpm = 986
	case "slv":
		sd = 145
		cpm = 885
	case "spa":
		sd = 127
		cpm = 1025
	case "swe":
		sd = 156
		cpm = 917
	case "tur":
		sd = 156
		cpm = 1054
	default:
		sd = 188
		cpm = 987
	}

	// Calculate read time, assumed one image requires 12 second (0.2 minute)
	minReadTime := float64(nChar)/(cpm+sd) + float64(nImg)*0.2
	maxReadTime := float64(nChar)/(cpm-sd) + float64(nImg)*0.2

	// Round number
	minReadTime = math.Floor(minReadTime + 0.5)
	maxReadTime = math.Floor(maxReadTime + 0.5)

	return int(minReadTime), int(maxReadTime)
}

// FromURL get readable content from the specified URL
func FromURL(url *nurl.URL, timeout time.Duration) (Article, error) {
	// Fetch page from URL
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url.String())
	if err != nil {
		return Article{}, err
	}
	defer resp.Body.Close()

	// Check content type. If not HTML, stop process
	contentType := resp.Header.Get("Content-type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if !strings.HasPrefix(contentType, "text/html") {
		return Article{}, fmt.Errorf("URL must be a text/html, found %s", contentType)
	}

	// Parse response body
	return FromReader(resp.Body, url)
}

// FromReader get readable content from the specified io.Reader
func FromReader(reader io.Reader, url *nurl.URL) (Article, error) {
	// Create goquery document
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return Article{}, err
	}

	// Prepare document
	removeScripts(doc)
	prepDocument(doc)

	// Get metadata and article
	metadata := getArticleMetadata(doc)
	articleContent, author := grabArticle(doc, metadata.Title)
	if articleContent == nil {
		return Article{}, fmt.Errorf("No article body detected")
	}

	// Post process content
	postProcessContent(articleContent, url)

	// Estimate read time
	minTime, maxTime := estimateReadTime(articleContent)
	metadata.MinReadTime = minTime
	metadata.MaxReadTime = maxTime

	// Update author data in metadata
	if author != "" {
		metadata.Author = author
	}

	// If we haven't found an excerpt in the article's metadata, use the first paragraph
	if metadata.Excerpt == "" {
		p := articleContent.Find("p").First().Text()
		metadata.Excerpt = normalizeText(p)
	}

	// Get text and HTML from content
	textContent := getTextContent(articleContent)
	htmlContent := getHTMLContent(articleContent)

	article := Article{
		URL:        url.String(),
		Meta:       metadata,
		Content:    textContent,
		RawContent: htmlContent,
	}

	return article, nil
}
