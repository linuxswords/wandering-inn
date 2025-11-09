package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/internal/models"
)

var (
	// Style for the current cursor position
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // White
			Background(lipgloss.Color("63")). // Purple
			Bold(true)

	// Style for selected range (between start and cursor in end selector)
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).  // Black
			Background(lipgloss.Color("120")). // Light green
			Bold(false)
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

type endChapterSelectorModel struct {
	chapters     []models.Chapter
	cursor       int
	selected     int
	total        int
	startChapter int
	quit         bool
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

func (cli *CLI) GetStartChapter() int {
	return cli.GetStartChapterInteractive(nil)
}

func (cli *CLI) GetStartChapterInteractive(chapters []models.Chapter) int {
	m := chapterSelectorModel{
		chapters: chapters,
		cursor:   len(chapters) - 1, // Start from the last chapter (0-indexed)
		total:    len(chapters),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		// Fallback to text input if bubbletea fails
		return cli.getStartChapterTextInput(len(chapters))
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
		line := fmt.Sprintf("%s %d. %s", cursor, i+1, chapterTitle)

		if m.cursor == i {
			cursor = ">"
			line = cursorStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, chapterTitle))
		}

		s += line + "\n"
	}

	if start > 0 {
		s += fmt.Sprintf("  ... (%d more above)\n", start)
	}
	if end < m.total-1 {
		s += fmt.Sprintf("  ... (%d more below)\n", m.total-1-end)
	}

	return s
}

func (m endChapterSelectorModel) Init() tea.Cmd {
	return nil
}

func (m endChapterSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > m.startChapter-1 {
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

func (m endChapterSelectorModel) View() string {
	s := "Wandering Inn EPUB Creator\n"
	s += "==========================\n"
	s += fmt.Sprintf("Select ending chapter (%d-%d, default: %d):\n", m.startChapter, m.total, m.total)
	s += "Use ↑/↓ arrow keys or j/k (vim keys) to navigate, Enter to select, 'q' to quit\n\n"

	// Show a window of chapters around the selected one
	windowSize := 10
	start := max(m.startChapter-1, max(0, m.cursor-windowSize/2))
	end := min(m.total-1, start+windowSize-1)

	// Adjust start if we're near the end
	if end == m.total-1 {
		start = max(m.startChapter-1, m.total-windowSize)
	}

	for i := start; i <= end; i++ {
		chapterTitle := fmt.Sprintf("Chapter %d", i+1)
		if m.chapters != nil && i < len(m.chapters) {
			chapterTitle = m.chapters[i].Title
		}

		cursor := " "
		line := fmt.Sprintf("%s %d. %s", cursor, i+1, chapterTitle)

		// Determine if this chapter is in the selected range
		inSelectedRange := i >= m.startChapter-1 && i <= m.cursor

		if m.cursor == i {
			// Current cursor position - highlighted with cursor style
			cursor = ">"
			line = cursorStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, chapterTitle))
		} else if inSelectedRange {
			// In the selected range but not at cursor - show with selected style
			line = selectedStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, chapterTitle))
		}

		s += line + "\n"
	}

	if start > m.startChapter-1 {
		s += fmt.Sprintf("  ... (%d more above)\n", start-(m.startChapter-1))
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

func (cli *CLI) getEndChapterTextInput(totalChapters, startChapter int) int {
	for {
		fmt.Printf("Enter ending chapter number (%d-%d, default: %d): ", startChapter, totalChapters, totalChapters)
		input, err := cli.reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input, please try again.")
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			return totalChapters // Default to last chapter
		}

		endChapter, err := strconv.Atoi(input)

		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}

		if endChapter < startChapter || endChapter > totalChapters {
			fmt.Printf("Please enter a number between %d and %d.\n", startChapter, totalChapters)
			continue
		}

		return endChapter
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (cli *CLI) PrintCreationInfo(numChapters, startIndex, endIndex int) {
	if startIndex == endIndex {
		fmt.Printf("Creating EPUB with 1 chapter: chapter %d...\n", startIndex)
	} else {
		fmt.Printf("Creating EPUB with %d chapters from chapter %d to %d...\n", numChapters, startIndex, endIndex)
	}
}

func (cli *CLI) GetEndChapterInteractive(chapters []models.Chapter, startChapter int) int {
	m := endChapterSelectorModel{
		chapters:     chapters,
		cursor:       len(chapters) - 1, // Default to the last chapter (0-indexed)
		total:        len(chapters),
		startChapter: startChapter,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		// Fallback to text input if bubbletea fails
		return cli.getEndChapterTextInput(len(chapters), startChapter)
	}

	if finalModel.(endChapterSelectorModel).quit {
		fmt.Println("Exiting...")
		os.Exit(0)
	}

	return finalModel.(endChapterSelectorModel).selected
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

