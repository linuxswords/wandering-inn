package scraper

import (
	"sort"
	"strings"

	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/internal/models"
	"github.com/linuxswords/wandering-inn/pkg/utils"
	"golang.org/x/net/html"
)

type Scraper interface {
	FetchTableOfContents() ([]models.Chapter, error)
	FetchChapterContent(url, title string) (string, error)
}

type WanderingInnScraper struct{}

func NewWanderingInnScraper() *WanderingInnScraper {
	return &WanderingInnScraper{}
}

func (s *WanderingInnScraper) FetchTableOfContents() ([]models.Chapter, error) {
	doc, err := utils.FetchAndParse(config.TOCUrl)
	if err != nil {
		return nil, err
	}

	var chapters []models.Chapter
	chapterIndex := 0

	var findChapters func(*html.Node)
	findChapters = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := utils.GetAttr(n, "href")
			if href != "" && strings.Contains(href, "wanderinginn.com") &&
				!strings.Contains(href, "table-of-contents") {
				title := utils.ExtractText(n)
				if title != "" && s.isChapterLink(title, href) {
					chapters = append(chapters, models.Chapter{
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

func (s *WanderingInnScraper) FetchChapterContent(url, title string) (string, error) {
	doc, err := utils.FetchAndParse(url)
	if err != nil {
		return "", err
	}

	parser := NewHTMLParser()
	content := parser.ExtractChapterHTML(doc, title)
	return content, nil
}

func (s *WanderingInnScraper) isChapterLink(title, href string) bool {
	return config.ChapterPattern.MatchString(title) && !strings.Contains(strings.ToLower(title), "table of contents")
}
