package tools

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func textOf(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var out strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		out.WriteString(textOf(c))
	}
	return out.String()
}

func attrOf(n *html.Node, name string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val, true
		}
	}
	return "", false
}

func matches(n *html.Node, selector string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch {
	case strings.HasPrefix(selector, "#"):
		v, ok := attrOf(n, "id")
		return ok && v == selector[1:]
	case strings.HasPrefix(selector, "."):
		v, ok := attrOf(n, "class")
		return ok && hasClass(v, selector[1:])
	case strings.HasPrefix(selector, "//"):
		return matchXPath(n, selector[2:])
	default:
		return n.Data == selector
	}
}

func hasClass(classes, want string) bool {
	for _, c := range strings.Fields(classes) {
		if c == want {
			return true
		}
	}
	return false
}

func matchXPath(n *html.Node, expr string) bool {
	open := strings.Index(expr, "[@")
	if open < 0 {
		return n.Data == expr
	}
	tag := expr[:open]
	if n.Data != tag {
		return false
	}
	rest := strings.TrimSuffix(expr[open+2:], "]")
	parts := strings.SplitN(rest, "=", 2)
	if len(parts) != 2 {
		return false
	}
	_, ok := attrOf(n, parts[0])
	if !ok {
		return false
	}
	want := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
	got, _ := attrOf(n, parts[0])
	return got == want
}

func selectNodes(document, selector string) ([]*html.Node, error) {
	doc, err := html.Parse(strings.NewReader(document))
	if err != nil {
		return nil, err
	}
	var found []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if matches(n, selector) {
			found = append(found, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return found, nil
}

// Select returns the text of every element matching selector.
func Select(document, selector string) ([]string, error) {
	nodes, err := selectNodes(document, selector)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, strings.TrimSpace(textOf(n)))
	}
	return out, nil
}

// Attr returns the named attribute of the first matching element.
func Attr(document, selector, name string) (string, error) {
	nodes, err := selectNodes(document, selector)
	if err != nil {
		return "", err
	}
	if len(nodes) == 0 {
		return "", fmt.Errorf("html: no element matches %q", selector)
	}
	value, ok := attrOf(nodes[0], name)
	if !ok {
		return "", fmt.Errorf("html: attribute %q not found", name)
	}
	return value, nil
}

// XPath returns text of elements matching //tag[@attr='value'].
func XPath(document, expr string) ([]string, error) {
	return Select(document, expr)
}
