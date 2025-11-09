package epub

import (
	"fmt"

	"github.com/go-shiori/go-epub"
	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/internal/models"
	"github.com/linuxswords/wandering-inn/pkg/utils"
)

type Creator interface {
	CreateEPUB(chapters []models.Chapter, scraper ChapterContentFetcher) error
}

type ChapterContentFetcher interface {
	FetchChapterContent(url, title string) (string, error)
}

type EPUBCreator struct {
	progressCallback func(current, total int, title string)
}

func NewEPUBCreator() *EPUBCreator {
	return &EPUBCreator{}
}

func (c *EPUBCreator) SetProgressCallback(callback func(current, total int, title string)) {
	c.progressCallback = callback
}

func (c *EPUBCreator) CreateEPUB(chapters []models.Chapter, scraper ChapterContentFetcher) error {
	e, err := epub.NewEpub(config.EpubTitle)
	if err != nil {
		return err
	}

	e.SetAuthor(config.EpubAuthor)
	e.SetDescription(config.EpubDescription)

	for i, chapter := range chapters {
		if c.progressCallback != nil {
			c.progressCallback(i+1, len(chapters), chapter.Title)
		}

		content, err := scraper.FetchChapterContent(chapter.URL, chapter.Title)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch chapter %s: %v\n", chapter.Title, err)
			continue
		}

		_, err = e.AddSection(content, chapter.Title, "", "")
		if err != nil {
			return err
		}
	}

	filename := utils.GenerateFilename(chapters)
	err = e.Write(filename)
	if err != nil {
		return err
	}

	fmt.Printf("EPUB created successfully: %s\n", filename)
	return nil
}
