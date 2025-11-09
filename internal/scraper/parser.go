package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/pkg/utils"
	"golang.org/x/net/html"
)

type HTMLParser struct{}

func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

func (p *HTMLParser) ExtractChapterHTML(n *html.Node, title string) string {
	if n.Type == html.ElementNode && (n.Data == "div" || n.Data == "article") {
		class := utils.GetAttr(n, "class")
		if strings.Contains(class, "entry-content") || strings.Contains(class, "post-content") {
			content := p.extractHTMLContent(n)
			return fmt.Sprintf("<h1>%s</h1>\n%s", title, content)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if content := p.ExtractChapterHTML(c, title); content != "" {
			return content
		}
	}

	return ""
}

func (p *HTMLParser) extractHTMLContent(n *html.Node) string {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if p.isNavigationText(text) {
			return ""
		}
		return html.EscapeString(n.Data)
	}

	if n.Type == html.ElementNode {
		if p.isNavigationElement(n) {
			return ""
		}

		switch n.Data {
		case "p":
			return p.handleParagraph(n)
		case "br":
			return "<br/>\n"
		case "strong", "b":
			return p.handleStrongOrBold(n)
		case "em", "i":
			return p.handleEmphasisOrItalic(n)
		case "div":
			return p.handleDiv(n)
		case "span":
			return p.handleSpan(n)
		case "script", "style", "nav", "footer", "header":
			return ""
		case "h1", "h2", "h3", "h4", "h5", "h6":
			return p.handleHeading(n)
		case "a":
			return p.handleAnchor(n)
		}
	}

	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	return content
}

func (p *HTMLParser) handleParagraph(n *html.Node) string {
	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	content = strings.TrimSpace(content)
	if content == "" || p.isNavigationText(content) {
		return ""
	}
	style := utils.GetAttr(n, "style")
	class := utils.GetAttr(n, "class")
	if style != "" || class != "" {
		return fmt.Sprintf("<p%s>%s</p>\n", p.buildAttributes(style, class), content)
	}
	return fmt.Sprintf("<p>%s</p>\n", content)
}

func (p *HTMLParser) handleStrongOrBold(n *html.Node) string {
	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	if p.isNavigationText(content) {
		return ""
	}
	style := utils.GetAttr(n, "style")
	class := utils.GetAttr(n, "class")
	if style != "" || class != "" {
		return fmt.Sprintf("<strong%s>%s</strong>", p.buildAttributes(style, class), content)
	}
	return fmt.Sprintf("<strong>%s</strong>", content)
}

func (p *HTMLParser) handleEmphasisOrItalic(n *html.Node) string {
	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	if p.isNavigationText(content) {
		return ""
	}
	style := utils.GetAttr(n, "style")
	class := utils.GetAttr(n, "class")
	if style != "" || class != "" {
		return fmt.Sprintf("<em%s>%s</em>", p.buildAttributes(style, class), content)
	}
	return fmt.Sprintf("<em>%s</em>", content)
}

func (p *HTMLParser) handleDiv(n *html.Node) string {
	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	return content
}

func (p *HTMLParser) handleSpan(n *html.Node) string {
	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	if p.isNavigationText(content) {
		return ""
	}
	style := utils.GetAttr(n, "style")
	class := utils.GetAttr(n, "class")
	if style != "" || class != "" {
		return fmt.Sprintf("<span%s>%s</span>", p.buildAttributes(style, class), content)
	}
	return content
}

func (p *HTMLParser) handleHeading(n *html.Node) string {
	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	if p.isNavigationText(content) {
		return ""
	}
	return fmt.Sprintf("<%s>%s</%s>\n", n.Data, content, n.Data)
}

func (p *HTMLParser) handleAnchor(n *html.Node) string {
	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += p.extractHTMLContent(c)
	}
	if p.isNavigationText(content) {
		return ""
	}
	return content
}

func (p *HTMLParser) isNavigationText(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))

	// Check for exact navigation patterns first
	for _, term := range config.NavigationTerms {
		// Special handling for "toc" to avoid matching inside words like "restock"
		if term == "toc" {
			// Only match "toc" as a standalone word
			if p.containsWord(text, "toc") {
				return true
			}
		} else {
			if strings.Contains(text, term) {
				return true
			}
		}
	}

	if config.NavigationSymbolPattern.MatchString(text) {
		return true
	}

	// Check for standalone navigation words only if they appear to be navigation links
	// (short text, likely to be a standalone navigation element)
	if len(text) < 30 {
		lowerText := strings.ToLower(text)
		if lowerText == "previous" || lowerText == "next" {
			return true
		}
		// Check for common navigation patterns
		if strings.Contains(lowerText, "← ") || strings.Contains(lowerText, " →") {
			return true
		}
	}

	return false
}

// containsWord checks if a word appears as a standalone word (not as part of another word)
func (p *HTMLParser) containsWord(text, word string) bool {
	// Use regex word boundary \b to match whole words only
	pattern := `\b` + regexp.QuoteMeta(word) + `\b`
	matched, _ := regexp.MatchString(`(?i)`+pattern, text)
	return matched
}

func (p *HTMLParser) isNavigationElement(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	class := utils.GetAttr(n, "class")
	id := utils.GetAttr(n, "id")

	for _, navClass := range config.NavigationClasses {
		if strings.Contains(strings.ToLower(class), navClass) {
			return true
		}
		if strings.Contains(strings.ToLower(id), navClass) {
			return true
		}
	}

	return false
}

func (p *HTMLParser) buildAttributes(style, class string) string {
	var attrs []string

	if class != "" {
		class = p.mapColorClass(class)
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, html.EscapeString(class)))
	}

	if style != "" {
		style = p.sanitizeStyle(style)
		if style != "" {
			attrs = append(attrs, fmt.Sprintf(`style="%s"`, html.EscapeString(style)))
		}
	}

	if len(attrs) > 0 {
		return " " + strings.Join(attrs, " ")
	}
	return ""
}

func (p *HTMLParser) mapColorClass(class string) string {
	class = strings.ToLower(strings.TrimSpace(class))

	for original, mapped := range config.ColorClassMap {
		if strings.Contains(class, original) {
			return mapped
		}
	}

	return class
}

func (p *HTMLParser) sanitizeStyle(style string) string {
	style = strings.TrimSpace(style)
	if style == "" {
		return ""
	}

	if strings.Contains(style, "color:") {
		return style
	}

	return ""
}

