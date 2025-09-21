package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/internal/models"
)

type CLI struct {
	reader *bufio.Reader
}

func NewCLI() *CLI {
	return &CLI{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (cli *CLI) PrintWelcome() {
	fmt.Println("Wandering Inn EPUB Creator")
	fmt.Println("==========================")
}

func (cli *CLI) PrintChapterInfo(chapters []models.Chapter) {
	fmt.Printf("Found %d chapters\n", len(chapters))
	fmt.Printf("Latest %d chapters:\n", config.LatestChaptersCount)
	start := max(0, len(chapters)-config.LatestChaptersCount)
	for i := start; i < len(chapters); i++ {
		fmt.Printf("%d. %s\n", i+1, chapters[i].Title)
	}
}

func (cli *CLI) GetStartChapter(totalChapters int) int {
	for {
		fmt.Printf("Enter starting chapter number (1-%d): ", totalChapters)
		input, err := cli.reader.ReadString('\n')
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

func (cli *CLI) PrintCreationInfo(numChapters, startIndex int) {
	fmt.Printf("Creating EPUB with %d chapters starting from chapter %d...\n", numChapters, startIndex)
}

func (cli *CLI) PrintDownloadProgress(current, total int, title string) {
	fmt.Printf("Downloading chapter %d/%d: %s\n", current, total, title)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}