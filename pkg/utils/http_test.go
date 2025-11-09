package utils

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestGetAttr(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []html.Attribute
		key      string
		expected string
	}{
		{
			name: "attribute exists",
			attrs: []html.Attribute{
				{Key: "class", Val: "test-class"},
				{Key: "id", Val: "test-id"},
			},
			key:      "class",
			expected: "test-class",
		},
		{
			name: "attribute does not exist",
			attrs: []html.Attribute{
				{Key: "class", Val: "test-class"},
			},
			key:      "id",
			expected: "",
		},
		{
			name:     "no attributes",
			attrs:    []html.Attribute{},
			key:      "class",
			expected: "",
		},
		{
			name: "empty attribute value",
			attrs: []html.Attribute{
				{Key: "class", Val: ""},
			},
			key:      "class",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: tt.attrs,
			}
			result := GetAttr(node, tt.key)
			if result != tt.expected {
				t.Errorf("GetAttr() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple text node",
			html:     "Hello World",
			expected: "Hello World",
		},
		{
			name:     "nested elements",
			html:     "<div><p>Hello</p><p>World</p></div>",
			expected: "HelloWorld",
		},
		{
			name:     "mixed content",
			html:     "<div>Hello <strong>bold</strong> text</div>",
			expected: "Hello bold text",
		},
		{
			name:     "empty element",
			html:     "<div></div>",
			expected: "",
		},
		{
			name:     "deeply nested",
			html:     "<div><p><span><strong>Nested</strong></span> text</p></div>",
			expected: "Nested text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Find the body element or first element
			var targetNode *html.Node
			var findNode func(*html.Node)
			findNode = func(n *html.Node) {
				if targetNode != nil {
					return
				}
				if n.Type == html.ElementNode && (n.Data == "body" || n.Data == "div" || n.Data == "p") {
					targetNode = n
					return
				}
				if n.Type == html.TextNode && strings.TrimSpace(n.Data) != "" {
					targetNode = n
					return
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					findNode(c)
				}
			}
			findNode(doc)

			if targetNode == nil {
				t.Fatalf("Could not find target node in HTML")
			}

			result := ExtractText(targetNode)
			if result != tt.expected {
				t.Errorf("ExtractText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFetchAndParse(t *testing.T) {
	// Note: This test requires network access and may be flaky
	// In a real production environment, you might want to use a mock HTTP server
	t.Run("invalid URL", func(t *testing.T) {
		_, err := FetchAndParse("invalid-url")
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
	})

	// Test with a non-existent domain to ensure proper error handling
	t.Run("non-existent domain", func(t *testing.T) {
		_, err := FetchAndParse("http://this-domain-should-not-exist-12345.com")
		if err == nil {
			t.Error("Expected error for non-existent domain, got nil")
		}
	})
}

