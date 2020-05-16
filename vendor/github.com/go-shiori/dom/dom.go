package dom

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// GetElementsByTagName returns a collection of all elements in the document with
// the specified tag name, as an array of Node object.
// The special tag "*" will represents all elements.
func GetElementsByTagName(doc *html.Node, tagName string) []*html.Node {
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

// CreateElement creates a new ElementNode with specified tag.
func CreateElement(tagName string) *html.Node {
	return &html.Node{
		Type: html.ElementNode,
		Data: tagName,
	}
}

// CreateTextNode creates a new Text node.
func CreateTextNode(data string) *html.Node {
	return &html.Node{
		Type: html.TextNode,
		Data: data,
	}
}

// TagName returns the tag name of a Node.
// If it's not ElementNode, return empty string.
func TagName(node *html.Node) string {
	if node == nil {
		return ""
	}

	if node.Type != html.ElementNode {
		return ""
	}

	return node.Data
}

// GetAttribute returns the value of a specified attribute on
// the element. If the given attribute does not exist, the value
// returned will be an empty string.
func GetAttribute(node *html.Node, attrName string) string {
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == attrName {
			return node.Attr[i].Val
		}
	}
	return ""
}

// SetAttribute sets attribute for node. If attribute already exists,
// it will be replaced.
func SetAttribute(node *html.Node, attrName string, attrValue string) {
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

// RemoveAttribute removes attribute with given name.
func RemoveAttribute(node *html.Node, attrName string) {
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

// HasAttribute returns a Boolean value indicating whether the
// specified node has the specified attribute or not.
func HasAttribute(node *html.Node, attrName string) bool {
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == attrName {
			return true
		}
	}
	return false
}

// TextContent returns the text content of the specified node,
// and all its descendants.
func TextContent(node *html.Node) string {
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

// OuterHTML returns an HTML serialization of the element and its descendants.
// The returned HTML value is escaped.
func OuterHTML(node *html.Node) string {
	if node == nil {
		return ""
	}

	var buffer bytes.Buffer
	err := html.Render(&buffer, node)
	if err != nil {
		return ""
	}

	return buffer.String()
}

// InnerHTML returns the HTML content (inner HTML) of an element.
// The returned HTML value is escaped.
func InnerHTML(node *html.Node) string {
	var err error
	var buffer bytes.Buffer

	if node == nil {
		return ""
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		err = html.Render(&buffer, child)
		if err != nil {
			return ""
		}
	}

	return strings.TrimSpace(buffer.String())
}

// DocumentElement returns the Element that is the root element
// of the document. Since we are working with HTML document,
// the root will be <html> element for HTML documents).
func DocumentElement(doc *html.Node) *html.Node {
	if nodes := GetElementsByTagName(doc, "html"); len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

// ID returns the value of the id attribute of the specified element.
func ID(node *html.Node) string {
	id := GetAttribute(node, "id")
	id = strings.TrimSpace(id)
	return id
}

// ClassName returns the value of the class attribute of
// the specified element.
func ClassName(node *html.Node) string {
	className := GetAttribute(node, "class")
	className = strings.TrimSpace(className)
	className = strings.Join(strings.Fields(className), " ")
	return className
}

// Children returns an HTMLCollection of the direct child elements of Node.
func Children(node *html.Node) []*html.Node {
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

// ChildNodes returns list of a node's direct children.
func ChildNodes(node *html.Node) []*html.Node {
	var childNodes []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childNodes = append(childNodes, child)
	}
	return childNodes
}

// FirstElementChild returns the object's first child Element,
// or nil if there are no child elements.
func FirstElementChild(node *html.Node) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			return child
		}
	}
	return nil
}

// PreviousElementSibling returns the the Element immediately prior
// to the specified one in its parent's children list, or null if
// the specified element is the first one in the list.
func PreviousElementSibling(node *html.Node) *html.Node {
	for sibling := node.PrevSibling; sibling != nil; sibling = sibling.PrevSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}
	return nil
}

// NextElementSibling returns the Element immediately following
// the specified one in its parent's children list, or nil if the
// specified Element is the last one in the list.
func NextElementSibling(node *html.Node) *html.Node {
	for sibling := node.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}
	return nil
}

// AppendChild adds a node to the end of the list of children of a
// specified parent node. If the given child is a reference to an
// existing node in the document, AppendChild() moves it from its
// current position to the new position.
func AppendChild(node *html.Node, child *html.Node) {
	if child.Parent != nil {
		temp := CloneNode(child)
		node.AppendChild(temp)
		child.Parent.RemoveChild(child)
	} else {
		node.AppendChild(child)
	}
}

// PrependChild works like AppendChild() except it adds a node to the
// beginning of the list of children of a specified parent node.
func PrependChild(node *html.Node, child *html.Node) {
	if child.Parent != nil {
		temp := CloneNode(child)
		child.Parent.RemoveChild(child)
		child = temp
	}

	if node.FirstChild != nil {
		node.InsertBefore(child, node.FirstChild)
	} else {
		node.AppendChild(child)
	}
}

// ReplaceChild replaces a child node within the given (parent) node.
// If the new child is already exist in document, ReplaceChild() will move it
// from its current position to replace old child. Returns both the new and old child.
func ReplaceChild(parent *html.Node, newChild *html.Node, oldChild *html.Node) (*html.Node, *html.Node) {
	if parent == nil {
		return nil, nil
	}

	if oldChild.Parent != parent {
		return nil, nil
	}

	if newChild.Parent != nil {
		tmp := CloneNode(newChild)
		newChild.Parent.RemoveChild(newChild)
		newChild = tmp
	}

	newChild.PrevSibling = nil
	newChild.NextSibling = nil
	parent.InsertBefore(newChild, oldChild)
	parent.RemoveChild(oldChild)

	return newChild, oldChild
}

// IncludeNode determines if node is included inside nodeList.
func IncludeNode(nodeList []*html.Node, node *html.Node) bool {
	for i := 0; i < len(nodeList); i++ {
		if nodeList[i] == node {
			return true
		}
	}
	return false
}

// CloneNode returns a deep clone of the node and its children.
// However, it will be detached from the original's parents
// and siblings.
func CloneNode(src *html.Node) *html.Node {
	clone := &html.Node{
		Type:     src.Type,
		DataAtom: src.DataAtom,
		Data:     src.Data,
		Attr:     make([]html.Attribute, len(src.Attr)),
	}

	copy(clone.Attr, src.Attr)
	for child := src.FirstChild; child != nil; child = child.NextSibling {
		clone.AppendChild(CloneNode(child))
	}

	return clone
}

// GetAllNodesWithTag is wrapper for GetElementsByTagName()
// which allow to get several tags at once.
func GetAllNodesWithTag(node *html.Node, tagNames ...string) []*html.Node {
	var result []*html.Node
	for i := 0; i < len(tagNames); i++ {
		result = append(result, GetElementsByTagName(node, tagNames[i])...)
	}
	return result
}

// ForEachNode iterates over a NodeList and runs fn on each node.
func ForEachNode(nodeList []*html.Node, fn func(*html.Node, int)) {
	for i := 0; i < len(nodeList); i++ {
		fn(nodeList[i], i)
	}
}

// RemoveNodes iterates over a NodeList, calls `filterFn` for each node
// and removes node if function returned `true`. If function is not
// passed, removes all the nodes in node list.
func RemoveNodes(nodeList []*html.Node, filterFn func(*html.Node) bool) {
	for i := len(nodeList) - 1; i >= 0; i-- {
		node := nodeList[i]
		parentNode := node.Parent
		if parentNode != nil && (filterFn == nil || filterFn(node)) {
			parentNode.RemoveChild(node)
		}
	}
}

// SetTextContent sets the text content of the specified node.
func SetTextContent(node *html.Node, text string) {
	child := node.FirstChild
	for child != nil {
		nextSibling := child.NextSibling
		node.RemoveChild(child)
		child = nextSibling
	}

	node.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}
