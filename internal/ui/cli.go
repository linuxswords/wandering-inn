package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/internal/models"
)

type CLI struct {
	reader *bufio.Reader
}

type chapterSelectorModel struct {
	chapters []models.Chapter
	cursor   int
	selected int
	total    int
	quit     bool
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
	return cli.GetStartChapterInteractive(nil, totalChapters)
}

func (cli *CLI) GetStartChapterInteractive(chapters []models.Chapter, totalChapters int) int {
	m := chapterSelectorModel{
		chapters: chapters,
		cursor:   totalChapters - 1, // Start from the last chapter (0-indexed)
		total:    totalChapters,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		// Fallback to text input if bubbletea fails
		return cli.getStartChapterTextInput(totalChapters)
	}

	if finalModel.(chapterSelectorModel).quit {
		fmt.Println("Exiting...")
		os.Exit(0)
	}

	return finalModel.(chapterSelectorModel).selected
}

func (m chapterSelectorModel) Init() tea.Cmd {
	return nil
}

func (m chapterSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.total-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor + 1 // Convert to 1-indexed
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m chapterSelectorModel) View() string {
	s := "Wandering Inn EPUB Creator\n"
	s += "==========================\n"
	s += fmt.Sprintf("Select starting chapter (1-%d):\n", m.total)
	s += "Use ↑/↓ arrow keys or j/k (vim keys) to navigate, Enter to select, 'q' to quit\n\n"

	// Show a window of chapters around the selected one
	windowSize := 10
	start := max(0, m.cursor-windowSize/2)
	end := min(m.total-1, start+windowSize-1)

	// Adjust start if we're near the end
	if end == m.total-1 {
		start = max(0, m.total-windowSize)
	}

	for i := start; i <= end; i++ {
		chapterTitle := fmt.Sprintf("Chapter %d", i+1)
		if m.chapters != nil && i < len(m.chapters) {
			chapterTitle = m.chapters[i].Title
		}

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %d. %s\n", cursor, i+1, chapterTitle)
	}

	if start > 0 {
		s += fmt.Sprintf("  ... (%d more above)\n", start)
	}
	if end < m.total-1 {
		s += fmt.Sprintf("  ... (%d more below)\n", m.total-1-end)
	}

	return s
}

func (cli *CLI) getStartChapterTextInput(totalChapters int) int {
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


func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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