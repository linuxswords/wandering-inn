package utils

import (
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

func GetAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func ExtractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += ExtractText(c)
	}
	return text
}

func FetchAndParse(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Error pages are valid HTML, so without this check they parse into an
	// empty chapter instead of failing.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: %s", url, resp.Status)
	}

	return html.Parse(resp.Body)
}
