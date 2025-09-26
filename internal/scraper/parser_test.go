package scraper

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestNewHTMLParser(t *testing.T) {
	parser := NewHTMLParser()
	if parser == nil {
		t.Error("NewHTMLParser() returned nil")
	}
}

func TestHTMLParser_ExtractChapterHTML(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		title    string
		expected string
	}{
		{
			name:     "entry-content div",
			htmlStr:  `<div class="entry-content"><p>Chapter content here</p></div>`,
			title:    "Test Chapter",
			expected: "<h1>Test Chapter</h1>\n<p>Chapter content here</p>\n",
		},
		{
			name:     "post-content div",
			htmlStr:  `<div class="post-content"><p>Chapter content here</p></div>`,
			title:    "Test Chapter",
			expected: "<h1>Test Chapter</h1>\n<p>Chapter content here</p>\n",
		},
		{
			name:     "no matching content",
			htmlStr:  `<div class="other"><p>Other content</p></div>`,
			title:    "Test Chapter",
			expected: "",
		},
		{
			name:     "article with entry-content",
			htmlStr:  `<article class="entry-content"><p>Article content</p></article>`,
			title:    "Test Chapter",
			expected: "<h1>Test Chapter</h1>\n<p>Article content</p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			parser := NewHTMLParser()
			result := parser.ExtractChapterHTML(doc, tt.title)
			if result != tt.expected {
				t.Errorf("ExtractChapterHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHTMLParser_isNavigationText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "previous chapter",
			text:     "Previous Chapter",
			expected: true,
		},
		{
			name:     "next chapter",
			text:     "Next Chapter",
			expected: true,
		},
		{
			name:     "table of contents",
			text:     "Table of Contents",
			expected: true,
		},
		{
			name:     "navigation symbols",
			text:     "←",
			expected: true,
		},
		{
			name:     "multiple navigation symbols",
			text:     "→→→",
			expected: true,
		},
		{
			name:     "regular content",
			text:     "This is regular chapter content",
			expected: false,
		},
		{
			name:     "case insensitive",
			text:     "PREVIOUS CHAPTER",
			expected: true,
		},
		{
			name:     "empty text",
			text:     "",
			expected: false,
		},
		{
			name:     "whitespace only",
			text:     "   ",
			expected: false,
		},
	}

	parser := NewHTMLParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isNavigationText(tt.text)
			if result != tt.expected {
				t.Errorf("isNavigationText(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestHTMLParser_isNavigationElement(t *testing.T) {
	tests := []struct {
		name     string
		node     *html.Node
		expected bool
	}{
		{
			name: "navigation class",
			node: &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{{Key: "class", Val: "navigation"}},
			},
			expected: true,
		},
		{
			name: "nav class",
			node: &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{{Key: "class", Val: "nav"}},
			},
			expected: true,
		},
		{
			name: "navigation id",
			node: &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{{Key: "id", Val: "navigation"}},
			},
			expected: true,
		},
		{
			name: "regular element",
			node: &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{{Key: "class", Val: "content"}},
			},
			expected: false,
		},
		{
			name: "text node",
			node: &html.Node{
				Type: html.TextNode,
				Data: "some text",
			},
			expected: false,
		},
	}

	parser := NewHTMLParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isNavigationElement(tt.node)
			if result != tt.expected {
				t.Errorf("isNavigationElement() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHTMLParser_mapColorClass(t *testing.T) {
	tests := []struct {
		name     string
		class    string
		expected string
	}{
		{
			name:     "red color class",
			class:    "has-red-color",
			expected: "red",
		},
		{
			name:     "blue color class",
			class:    "has-blue-color",
			expected: "blue",
		},
		{
			name:     "multiple classes with color",
			class:    "some-class has-green-color other-class",
			expected: "green",
		},
		{
			name:     "case insensitive",
			class:    "HAS-PURPLE-COLOR",
			expected: "purple",
		},
		{
			name:     "no color class",
			class:    "regular-class",
			expected: "regular-class",
		},
		{
			name:     "empty class",
			class:    "",
			expected: "",
		},
	}

	parser := NewHTMLParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.mapColorClass(tt.class)
			if result != tt.expected {
				t.Errorf("mapColorClass(%q) = %q, want %q", tt.class, result, tt.expected)
			}
		})
	}
}

func TestHTMLParser_sanitizeStyle(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		expected string
	}{
		{
			name:     "color style",
			style:    "color: red;",
			expected: "color: red;",
		},
		{
			name:     "color style with other properties",
			style:    "color: blue; font-size: 12px;",
			expected: "color: blue; font-size: 12px;",
		},
		{
			name:     "no color style",
			style:    "font-size: 12px; font-weight: bold;",
			expected: "",
		},
		{
			name:     "empty style",
			style:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			style:    "   ",
			expected: "",
		},
	}

	parser := NewHTMLParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.sanitizeStyle(tt.style)
			if result != tt.expected {
				t.Errorf("sanitizeStyle(%q) = %q, want %q", tt.style, result, tt.expected)
			}
		})
	}
}

func TestHTMLParser_buildAttributes(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		class    string
		expected string
	}{
		{
			name:     "both style and class",
			style:    "color: red;",
			class:    "test-class",
			expected: ` class="test-class" style="color: red;"`,
		},
		{
			name:     "class only",
			style:    "",
			class:    "test-class",
			expected: ` class="test-class"`,
		},
		{
			name:     "style only",
			style:    "color: red;",
			class:    "",
			expected: ` style="color: red;"`,
		},
		{
			name:     "neither style nor class",
			style:    "",
			class:    "",
			expected: "",
		},
		{
			name:     "sanitized style (no color)",
			style:    "font-size: 12px;",
			class:    "test-class",
			expected: ` class="test-class"`,
		},
		{
			name:     "color class mapping",
			style:    "",
			class:    "has-red-color",
			expected: ` class="red"`,
		},
	}

	parser := NewHTMLParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.buildAttributes(tt.style, tt.class)
			if result != tt.expected {
				t.Errorf("buildAttributes(%q, %q) = %q, want %q", tt.style, tt.class, result, tt.expected)
			}
		})
	}
}

func TestHTMLParser_handleParagraph(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		expected string
	}{
		{
			name:     "simple paragraph",
			htmlStr:  `<p>Simple paragraph content</p>`,
			expected: "<p>Simple paragraph content</p>\n",
		},
		{
			name:     "paragraph with navigation text",
			htmlStr:  `<p>Previous Chapter</p>`,
			expected: "",
		},
		{
			name:     "empty paragraph",
			htmlStr:  `<p></p>`,
			expected: "",
		},
		{
			name:     "paragraph with style",
			htmlStr:  `<p style="color: red;">Styled paragraph</p>`,
			expected: `<p style="color: red;">Styled paragraph</p>` + "\n",
		},
		{
			name:     "paragraph with class",
			htmlStr:  `<p class="has-blue-color">Blue paragraph</p>`,
			expected: `<p class="blue">Blue paragraph</p>` + "\n",
		},
		{
			name:     "paragraph with emphasis",
			htmlStr:  `<p>"I don't know. Get <em>me</em> something that brightens my day without turning me into a smiling loon, please? He's off his vacation next week; I'll live."</p>`,
			expected: `<p>&#34;I don&#39;t know. Get <em>me</em> something that brightens my day without turning me into a smiling loon, please? He&#39;s off his vacation next week; I&#39;ll live.&#34;</p>` + "\n",
		},
	}

	parser := NewHTMLParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Find the paragraph element
			var pNode *html.Node
			var findP func(*html.Node)
			findP = func(n *html.Node) {
				if n.Type == html.ElementNode && n.Data == "p" {
					pNode = n
					return
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					findP(c)
				}
			}
			findP(doc)

			if pNode == nil {
				t.Fatalf("Could not find paragraph element")
			}

			result := parser.handleParagraph(pNode)
			if result != tt.expected {
				t.Errorf("handleParagraph() = %q, want %q", result, tt.expected)
			}
		})
	}
}