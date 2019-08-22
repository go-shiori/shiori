package archiver

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// getElementsByTagName returns a collection of all elements in the document with
// the specified tag name, as an array of Node object.
// The special tag "*" will represents all elements.
func getElementsByTagName(doc *html.Node, tagName string) []*html.Node {
	var results []*html.Node
	var finder func(*html.Node)

	finder = func(node *html.Node) {
		if node.Type == html.ElementNode && (tagName == "*" || node.Data == tagName) {
			results = append(results, node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}

	for child := doc.FirstChild; child != nil; child = child.NextSibling {
		finder(child)
	}

	return results
}

// createElement creates a new ElementNode with specified tag.
func createElement(tagName string) *html.Node {
	return &html.Node{
		Type: html.ElementNode,
		Data: tagName,
	}
}

// createTextNode creates a new Text node.
func createTextNode(data string) *html.Node {
	return &html.Node{
		Type: html.TextNode,
		Data: data,
	}
}

// tagName returns the tag name of a Node.
// If it's not ElementNode, return empty string.
func tagName(node *html.Node) string {
	if node.Type != html.ElementNode {
		return ""
	}
	return node.Data
}

// getAttribute returns the value of a specified attribute on
// the element. If the given attribute does not exist, the value
// returned will be an empty string.
func getAttribute(node *html.Node, attrName string) string {
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == attrName {
			return node.Attr[i].Val
		}
	}
	return ""
}

// setAttribute sets attribute for node. If attribute already exists,
// it will be replaced.
func setAttribute(node *html.Node, attrName string, attrValue string) {
	attrIdx := -1
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == attrName {
			attrIdx = i
			break
		}
	}

	if attrIdx >= 0 {
		node.Attr[attrIdx].Val = attrValue
	} else {
		node.Attr = append(node.Attr, html.Attribute{
			Key: attrName,
			Val: attrValue,
		})
	}
}

// removeAttribute removes attribute with given name.
func removeAttribute(node *html.Node, attrName string) {
	attrIdx := -1
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == attrName {
			attrIdx = i
			break
		}
	}

	if attrIdx >= 0 {
		a := node.Attr
		a = append(a[:attrIdx], a[attrIdx+1:]...)
		node.Attr = a
	}
}

// hasAttribute returns a Boolean value indicating whether the
// specified node has the specified attribute or not.
func hasAttribute(node *html.Node, attrName string) bool {
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == attrName {
			return true
		}
	}
	return false
}

// textContent returns the text content of the specified node,
// and all its descendants.
func textContent(node *html.Node) string {
	var buffer bytes.Buffer
	var finder func(*html.Node)

	finder = func(n *html.Node) {
		if n.Type == html.TextNode {
			buffer.WriteString(n.Data)
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}

	finder(node)
	return buffer.String()
}

// outerHTML returns an HTML serialization of the element and its descendants.
func outerHTML(node *html.Node) []byte {
	var buffer bytes.Buffer
	err := html.Render(&buffer, node)
	if err != nil {
		return []byte{}
	}
	return buffer.Bytes()
}

// innerHTML returns the HTML content (inner HTML) of an element.
func innerHTML(node *html.Node) string {
	var err error
	var buffer bytes.Buffer

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		err = html.Render(&buffer, child)
		if err != nil {
			return ""
		}
	}

	return strings.TrimSpace(buffer.String())
}

// documentElement returns the Element that is the root element
// of the document. Since we are working with HTML document,
// the root will be <html> element for HTML documents).
func documentElement(doc *html.Node) *html.Node {
	if nodes := getElementsByTagName(doc, "html"); len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

// id returns the value of the id attribute of the specified element.
func id(node *html.Node) string {
	id := getAttribute(node, "id")
	id = strings.TrimSpace(id)
	return id
}

// className returns the value of the class attribute of
// the specified element.
func className(node *html.Node) string {
	className := getAttribute(node, "class")
	className = strings.TrimSpace(className)
	className = strings.Join(strings.Fields(className), " ")
	return className
}

// children returns an HTMLCollection of the child elements of Node.
func children(node *html.Node) []*html.Node {
	var children []*html.Node
	if node == nil {
		return nil
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			children = append(children, child)
		}
	}
	return children
}

// childNodes returns list of a node's direct children.
func childNodes(node *html.Node) []*html.Node {
	var childNodes []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childNodes = append(childNodes, child)
	}
	return childNodes
}

// firstElementChild returns the object's first child Element,
// or nil if there are no child elements.
func firstElementChild(node *html.Node) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			return child
		}
	}
	return nil
}

// nextElementSibling returns the Element immediately following
// the specified one in its parent's children list, or nil if the
// specified Element is the last one in the list.
func nextElementSibling(node *html.Node) *html.Node {
	for sibling := node.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}
	return nil
}

// appendChild adds a node to the end of the list of children of a
// specified parent node. If the given child is a reference to an
// existing node in the document, appendChild() moves it from its
// current position to the new position.
func appendChild(node *html.Node, child *html.Node) {
	if child.Parent != nil {
		temp := cloneNode(child)
		node.AppendChild(temp)
		child.Parent.RemoveChild(child)
	} else {
		node.AppendChild(child)
	}
}

// prependChild works like appendChild, except it adds a node to the
// beginning of the list of children of a specified parent node.
func prependChild(node *html.Node, child *html.Node) {
	if child.Parent != nil {
		temp := cloneNode(child)
		child.Parent.RemoveChild(child)
		child = temp
	}

	if node.FirstChild != nil {
		node.InsertBefore(child, node.FirstChild)
	} else {
		node.AppendChild(child)
	}
}

// replaceNode replaces an OldNode with a NewNode.
func replaceNode(oldNode *html.Node, newNode *html.Node) {
	if oldNode.Parent == nil {
		return
	}

	newNode.Parent = nil
	newNode.PrevSibling = nil
	newNode.NextSibling = nil
	oldNode.Parent.InsertBefore(newNode, oldNode)
	oldNode.Parent.RemoveChild(oldNode)
}

// includeNode determines if node is included inside nodeList.
func includeNode(nodeList []*html.Node, node *html.Node) bool {
	for i := 0; i < len(nodeList); i++ {
		if nodeList[i] == node {
			return true
		}
	}
	return false
}

// cloneNode returns a deep clone of the node and its children.
// However, it will be detached from the original's parents
// and siblings.
func cloneNode(src *html.Node) *html.Node {
	clone := &html.Node{
		Type:     src.Type,
		DataAtom: src.DataAtom,
		Data:     src.Data,
		Attr:     make([]html.Attribute, len(src.Attr)),
	}

	copy(clone.Attr, src.Attr)
	for child := src.FirstChild; child != nil; child = child.NextSibling {
		clone.AppendChild(cloneNode(child))
	}

	return clone
}

func getAllNodesWithTag(node *html.Node, tagNames ...string) []*html.Node {
	var result []*html.Node
	for i := 0; i < len(tagNames); i++ {
		result = append(result, getElementsByTagName(node, tagNames[i])...)
	}
	return result
}

// forEachNode iterates over a NodeList and runs fn on each node.
func forEachNode(nodeList []*html.Node, fn func(*html.Node, int)) {
	for i := 0; i < len(nodeList); i++ {
		fn(nodeList[i], i)
	}
}

// removeNodes iterates over a NodeList, calls `filterFn` for each node
// and removes node if function returned `true`. If function is not
// passed, removes all the nodes in node list.
func removeNodes(nodeList []*html.Node, filterFn func(*html.Node) bool) {
	for i := len(nodeList) - 1; i >= 0; i-- {
		node := nodeList[i]
		parentNode := node.Parent
		if parentNode != nil && (filterFn == nil || filterFn(node)) {
			parentNode.RemoveChild(node)
		}
	}
}

// setTextContent sets the text content of the specified node.
func setTextContent(node *html.Node, text string) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Parent != nil {
			child.Parent.RemoveChild(child)
		}
	}

	node.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}
