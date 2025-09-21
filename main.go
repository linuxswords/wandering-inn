package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-shiori/go-epub"
	"golang.org/x/net/html"
)

type Chapter struct {
	Title string
	URL   string
	Index int
}

func main() {
	fmt.Println("Wandering Inn EPUB Creator")
	fmt.Println("==========================")

	chapters, err := fetchTableOfContents()
	if err != nil {
		log.Fatalf("Error fetching table of contents: %v", err)
	}

	fmt.Printf("Found %d chapters\n", len(chapters))
	fmt.Println("Latest 20 chapters:")
	start := max(0, len(chapters)-20)
	for i := start; i < len(chapters); i++ {
		fmt.Printf("%d. %s\n", i+1, chapters[i].Title)
	}

	startIndex := getStartChapter(len(chapters))
	selectedChapters := chapters[startIndex-1:]

	fmt.Printf("Creating EPUB with %d chapters starting from chapter %d...\n", len(selectedChapters), startIndex)

	err = createEPUB(selectedChapters)
	if err != nil {
		log.Fatalf("Error creating EPUB: %v", err)
	}
}

func fetchTableOfContents() ([]Chapter, error) {
	tocURL := "https://wanderinginn.com/table-of-contents/"

	resp, err := http.Get(tocURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var chapters []Chapter
	chapterIndex := 0

	var findChapters func(*html.Node)
	findChapters = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := getAttr(n, "href")
			if href != "" && strings.Contains(href, "wanderinginn.com") &&
				!strings.Contains(href, "table-of-contents") {
				title := extractText(n)
				if title != "" && isChapterLink(title, href) {
					chapters = append(chapters, Chapter{
						Title: strings.TrimSpace(title),
						URL:   href,
						Index: chapterIndex,
					})
					chapterIndex++
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findChapters(c)
		}
	}

	findChapters(doc)

	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Index < chapters[j].Index
	})

	return chapters, nil
}

func isChapterLink(title, href string) bool {
	chapterPattern := regexp.MustCompile(`(?i)(chapter|prologue|epilogue|interlude|\d+\.\d+)`)
	return chapterPattern.MatchString(title) && !strings.Contains(strings.ToLower(title), "table of contents")
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += extractText(c)
	}
	return text
}

func getStartChapter(totalChapters int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Enter starting chapter number (1-%d): ", totalChapters)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input, please try again.")
			continue
		}

		input = strings.TrimSpace(input)
		startChapter, err := strconv.Atoi(input)

		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}

		if startChapter < 1 || startChapter > totalChapters {
			fmt.Printf("Please enter a number between 1 and %d.\n", totalChapters)
			continue
		}

		return startChapter
	}
}

func createEPUB(chapters []Chapter) error {
	e, err := epub.NewEpub("The Wandering Inn")
	if err != nil {
		return err
	}

	e.SetAuthor("pirateaba")
	e.SetDescription("The Wandering Inn web serial")

	// css := "body { font-family: Georgia, serif; line-height: 1.6; color: #333; } " +
	// 	"h1 { color: #2c3e50; text-align: center; border-bottom: 2px solid #3498db; padding-bottom: 10px; } " +
	// 	"p { margin-bottom: 1em; text-align: justify; } " +
	// 	".red { color: #e74c3c; } " +
	// 	".blue { color: #3498db; } " +
	// 	".green { color: #27ae60; } " +
	// 	".purple { color: #9b59b6; } " +
	// 	".orange { color: #e67e22; } " +
	// 	".yellow { color: #f1c40f; } " +
	// 	".brown { color: #8b4513; } " +
	// 	".pink { color: #e91e63; } " +
	// 	".cyan { color: #1abc9c; } " +
	// 	".gray { color: #7f8c8d; } " +
	// 	".gold { color: #ffd700; } " +
	// 	".silver { color: #c0c0c0; } " +
	// 	".crimson { color: #dc143c; } " +
	// 	".maroon { color: #800000; } " +
	// 	".navy { color: #000080; } " +
	// 	".teal { color: #008080; }"
	//
	// _, err = e.AddCSS(css, "styles.css")
	// if err != nil {
	// return err
	// }

	for i, chapter := range chapters {
		fmt.Printf("Downloading chapter %d/%d: %s\n", i+1, len(chapters), chapter.Title)

		content, err := fetchChapterContent(chapter.URL, chapter.Title)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch chapter %s: %v\n", chapter.Title, err)
			continue
		}

		_, err = e.AddSection(content, chapter.Title, "", "")
		if err != nil {
			return err
		}
	}

	filename := generateFilename(chapters)
	err = e.Write(filename)
	if err != nil {
		return err
	}

	fmt.Printf("EPUB created successfully: %s\n", filename)
	return nil
}

func generateFilename(chapters []Chapter) string {
	if len(chapters) == 0 {
		return "wandering_inn.epub"
	}

	startChapter := sanitizeFilename(chapters[0].Title)
	return fmt.Sprintf("wandering_inn_%s.epub", startChapter)
}

func sanitizeFilename(title string) string {
	title = strings.TrimSpace(title)

	title = regexp.MustCompile(`[^\w\s\-\.]`).ReplaceAllString(title, "")

	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, "_")

	title = strings.ToLower(title)

	if len(title) > 50 {
		title = title[:50]
	}

	title = strings.Trim(title, "_-.")

	if title == "" {
		return "chapter"
	}

	return title
}

func fetchChapterContent(url, title string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	content := extractChapterHTML(doc, title)
	return content, nil
}

func extractChapterHTML(n *html.Node, title string) string {
	if n.Type == html.ElementNode && (n.Data == "div" || n.Data == "article") {
		class := getAttr(n, "class")
		if strings.Contains(class, "entry-content") || strings.Contains(class, "post-content") {
			content := extractHTMLContent(n)
			return fmt.Sprintf("<h1>%s</h1>\n%s", title, content)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if content := extractChapterHTML(c, title); content != "" {
			return content
		}
	}

	return ""
}

func extractHTMLContent(n *html.Node) string {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if isNavigationText(text) {
			return ""
		}
		return html.EscapeString(n.Data)
	}

	if n.Type == html.ElementNode {
		if isNavigationElement(n) {
			return ""
		}

		switch n.Data {
		case "p":
			var content string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				content += extractHTMLContent(c)
			}
			content = strings.TrimSpace(content)
			if content == "" || isNavigationText(content) {
				return ""
			}
			style := getAttr(n, "style")
			class := getAttr(n, "class")
			if style != "" || class != "" {
				return fmt.Sprintf("<p%s>%s</p>\n", buildAttributes(style, class), content)
			}
			return fmt.Sprintf("<p>%s</p>\n", content)
		case "br":
			return "<br/>\n"
		case "strong", "b":
			var content string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				content += extractHTMLContent(c)
			}
			if isNavigationText(content) {
				return ""
			}
			style := getAttr(n, "style")
			class := getAttr(n, "class")
			if style != "" || class != "" {
				return fmt.Sprintf("<strong%s>%s</strong>", buildAttributes(style, class), content)
			}
			return fmt.Sprintf("<strong>%s</strong>", content)
		case "em", "i":
			var content string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				content += extractHTMLContent(c)
			}
			if isNavigationText(content) {
				return ""
			}
			style := getAttr(n, "style")
			class := getAttr(n, "class")
			if style != "" || class != "" {
				return fmt.Sprintf("<em%s>%s</em>", buildAttributes(style, class), content)
			}
			return fmt.Sprintf("<em>%s</em>", content)
		case "div":
			var content string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				content += extractHTMLContent(c)
			}
			return content
		case "span":
			var content string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				content += extractHTMLContent(c)
			}
			if isNavigationText(content) {
				return ""
			}
			style := getAttr(n, "style")
			class := getAttr(n, "class")
			if style != "" || class != "" {
				return fmt.Sprintf("<span%s>%s</span>", buildAttributes(style, class), content)
			}
			return content
		case "script", "style", "nav", "footer", "header":
			return ""
		case "h1", "h2", "h3", "h4", "h5", "h6":
			var content string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				content += extractHTMLContent(c)
			}
			if isNavigationText(content) {
				return ""
			}
			return fmt.Sprintf("<%s>%s</%s>\n", n.Data, content, n.Data)
		case "a":
			var content string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				content += extractHTMLContent(c)
			}
			if isNavigationText(content) {
				return ""
			}
			return content
		}
	}

	var content string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content += extractHTMLContent(c)
	}
	return content
}

func isNavigationText(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	navigationTerms := []string{
		"previous chapter",
		"next chapter",
		"← previous",
		"next →",
		"previous",
		"next",
		"table of contents",
		"toc",
		"chapter index",
		"first chapter",
		"last chapter",
	}

	for _, term := range navigationTerms {
		if strings.Contains(text, term) {
			return true
		}
	}

	if regexp.MustCompile(`^(←|→|«|»|\|)+$`).MatchString(text) {
		return true
	}

	return false
}

func isNavigationElement(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	class := getAttr(n, "class")
	id := getAttr(n, "id")

	navClasses := []string{
		"navigation",
		"nav",
		"chapter-nav",
		"post-nav",
		"entry-nav",
		"pagination",
		"prev-next",
		"chapter-links",
	}

	for _, navClass := range navClasses {
		if strings.Contains(strings.ToLower(class), navClass) {
			return true
		}
		if strings.Contains(strings.ToLower(id), navClass) {
			return true
		}
	}

	return false
}

func buildAttributes(style, class string) string {
	var attrs []string

	if class != "" {
		class = mapColorClass(class)
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, html.EscapeString(class)))
	}

	if style != "" {
		style = sanitizeStyle(style)
		if style != "" {
			attrs = append(attrs, fmt.Sprintf(`style="%s"`, html.EscapeString(style)))
		}
	}

	if len(attrs) > 0 {
		return " " + strings.Join(attrs, " ")
	}
	return ""
}

func mapColorClass(class string) string {
	class = strings.ToLower(strings.TrimSpace(class))

	colorMap := map[string]string{
		"has-red-color":     "red",
		"has-blue-color":    "blue",
		"has-green-color":   "green",
		"has-purple-color":  "purple",
		"has-orange-color":  "orange",
		"has-yellow-color":  "yellow",
		"has-brown-color":   "brown",
		"has-pink-color":    "pink",
		"has-cyan-color":    "cyan",
		"has-gray-color":    "gray",
		"has-grey-color":    "gray",
		"has-gold-color":    "gold",
		"has-silver-color":  "silver",
		"has-crimson-color": "crimson",
		"has-maroon-color":  "maroon",
		"has-navy-color":    "navy",
		"has-teal-color":    "teal",
	}

	for original, mapped := range colorMap {
		if strings.Contains(class, original) {
			return mapped
		}
	}

	return class
}

func sanitizeStyle(style string) string {
	style = strings.TrimSpace(style)
	if style == "" {
		return ""
	}

	if strings.Contains(style, "color:") {
		return style
	}

	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

