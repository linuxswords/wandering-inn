package main

import (
	"log"

	"github.com/linuxswords/wandering-inn/internal/epub"
	"github.com/linuxswords/wandering-inn/internal/scraper"
	"github.com/linuxswords/wandering-inn/internal/ui"
)

func main() {
	cli := ui.NewCLI()
	cli.PrintWelcome()

	scraperImpl := scraper.NewWanderingInnScraper()
	epubCreator := epub.NewEPUBCreator()

	chapters, err := scraperImpl.FetchTableOfContents()
	if err != nil {
		log.Fatalf("Error fetching table of contents: %v", err)
	}

	cli.PrintChapterInfo(chapters)

	startIndex := cli.GetStartChapterInteractive(chapters)
	endIndex := cli.GetEndChapterInteractive(chapters, startIndex)

	selectedChapters := chapters[startIndex-1 : endIndex]

	cli.PrintCreationInfo(len(selectedChapters), startIndex, endIndex)

	epubCreator.SetProgressCallback(cli.PrintDownloadProgress)

	err = epubCreator.CreateEPUB(selectedChapters, scraperImpl)
	if err != nil {
		log.Fatalf("Error creating EPUB: %v", err)
	}
}
