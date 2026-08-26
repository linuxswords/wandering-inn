package ui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/internal/models"
)

var (
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("120")).
			Bold(false)

	bookHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true)
)

// bookPattern matches chapter titles that start with a book number like "7.6" or "10.62 NY".
var bookPattern = regexp.MustCompile(`^(\d+)\.`)

// computeBooks assigns a book number to every chapter. Chapters whose titles start with
// "N." (e.g. "7.6 R") get that N as their book. Interludes and other specials inherit
// the last seen book number.
// Returns a per-chapter book slice and a map from book number → first chapter index (0-based).
func computeBooks(chapters []models.Chapter) ([]int, map[int]int) {
	books := make([]int, len(chapters))
	bookMap := make(map[int]int)
	lastBook := 1
	for i, ch := range chapters {
		if m := bookPattern.FindStringSubmatch(ch.Title); m != nil {
			n, _ := strconv.Atoi(m[1])
			lastBook = n
		}
		books[i] = lastBook
		if _, exists := bookMap[lastBook]; !exists {
			bookMap[lastBook] = i
		}
	}
	return books, bookMap
}

type CLI struct {
	reader *bufio.Reader
}

type chapterSelectorModel struct {
	chapters    []models.Chapter
	cursor      int
	selected    int
	total       int
	quit        bool
	books       []int
	bookMap     map[int]int
	inputBuffer string
}

type endChapterSelectorModel struct {
	chapters     []models.Chapter
	cursor       int
	selected     int
	total        int
	startChapter int
	quit         bool
	books        []int
	bookMap      map[int]int
	inputBuffer  string
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
	books, bookMap := computeBooks(chapters)
	m := chapterSelectorModel{
		chapters: chapters,
		cursor:   len(chapters) - 1,
		total:    len(chapters),
		books:    books,
		bookMap:  bookMap,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
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
			m.inputBuffer = ""
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.inputBuffer = ""
			if m.cursor < m.total-1 {
				m.cursor++
			}
		case "g":
			m.inputBuffer = ""
			m.cursor = 0
		case "G":
			m.inputBuffer = ""
			m.cursor = m.total - 1
		case "[":
			// Jump to start of current book; if already there, jump to start of previous book.
			m.inputBuffer = ""
			if len(m.books) > 0 {
				cur := m.books[m.cursor]
				if m.bookMap[cur] < m.cursor {
					m.cursor = m.bookMap[cur]
				} else if cur > 1 {
					if prev, ok := m.bookMap[cur-1]; ok {
						m.cursor = prev
					}
				}
			}
		case "]":
			// Jump to the start of the next book.
			m.inputBuffer = ""
			if len(m.books) > 0 {
				cur := m.books[m.cursor]
				if next, ok := m.bookMap[cur+1]; ok {
					m.cursor = next
				} else {
					m.cursor = m.total - 1
				}
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.inputBuffer += msg.String()
		case "backspace":
			if len(m.inputBuffer) > 0 {
				m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
			}
		case "enter":
			if m.inputBuffer != "" {
				n, err := strconv.Atoi(m.inputBuffer)
				m.inputBuffer = ""
				if err == nil && n >= 1 && n <= m.total {
					m.cursor = n - 1
				}
			} else {
				m.selected = m.cursor + 1
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m chapterSelectorModel) View() string {
	currentBook := 0
	if len(m.books) > 0 && m.cursor < len(m.books) {
		currentBook = m.books[m.cursor]
	}

	s := "Wandering Inn EPUB Creator\n"
	s += "==========================\n"
	s += fmt.Sprintf("Select starting chapter (1-%d", m.total)
	if currentBook > 0 {
		s += fmt.Sprintf(", Book %d", currentBook)
	}
	s += "):\n"
	s += "↑/↓ j/k navigate  [ ] jump books  g/G first/last  type number+Enter to jump\n\n"

	windowSize := 10
	start := max(0, m.cursor-windowSize/2)
	end := min(m.total-1, start+windowSize-1)
	if end == m.total-1 {
		start = max(0, m.total-windowSize)
	}

	if start > 0 {
		s += fmt.Sprintf("  ... (%d more above)\n", start)
	}

	prevBook := -1
	for i := start; i <= end; i++ {
		if len(m.books) > 0 {
			book := m.books[i]
			if book != prevBook {
				s += bookHeaderStyle.Render(fmt.Sprintf("── Book %d ──────────────", book)) + "\n"
				prevBook = book
			}
		}

		chapterTitle := fmt.Sprintf("Chapter %d", i+1)
		if m.chapters != nil && i < len(m.chapters) {
			chapterTitle = m.chapters[i].Title
		}

		if m.cursor == i {
			s += cursorStyle.Render(fmt.Sprintf("> %d. %s", i+1, chapterTitle)) + "\n"
		} else {
			s += fmt.Sprintf("  %d. %s\n", i+1, chapterTitle)
		}
	}

	if end < m.total-1 {
		s += fmt.Sprintf("  ... (%d more below)\n", m.total-1-end)
	}

	if m.inputBuffer != "" {
		s += "\n" + inputStyle.Render(fmt.Sprintf("Jump to chapter: %s▌", m.inputBuffer))
	}

	return s
}

func (m endChapterSelectorModel) Init() tea.Cmd {
	return nil
}

func (m endChapterSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	minCursor := m.startChapter - 1
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			m.inputBuffer = ""
			if m.cursor > minCursor {
				m.cursor--
			}
		case "down", "j":
			m.inputBuffer = ""
			if m.cursor < m.total-1 {
				m.cursor++
			}
		case "G":
			m.inputBuffer = ""
			m.cursor = m.total - 1
		case "[":
			// Jump to start of current book (not below startChapter).
			m.inputBuffer = ""
			if len(m.books) > 0 {
				cur := m.books[m.cursor]
				target := max(minCursor, m.bookMap[cur])
				m.cursor = target
			}
		case "]":
			// Jump to last chapter of the current book.
			m.inputBuffer = ""
			if len(m.books) > 0 {
				cur := m.books[m.cursor]
				if next, ok := m.bookMap[cur+1]; ok {
					m.cursor = next - 1
				} else {
					m.cursor = m.total - 1
				}
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.inputBuffer += msg.String()
		case "backspace":
			if len(m.inputBuffer) > 0 {
				m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
			}
		case "enter":
			if m.inputBuffer != "" {
				n, err := strconv.Atoi(m.inputBuffer)
				m.inputBuffer = ""
				if err == nil && n >= m.startChapter && n <= m.total {
					m.cursor = n - 1
				}
			} else {
				m.selected = m.cursor + 1
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m endChapterSelectorModel) View() string {
	currentBook := 0
	if len(m.books) > 0 && m.cursor < len(m.books) {
		currentBook = m.books[m.cursor]
	}

	s := "Wandering Inn EPUB Creator\n"
	s += "==========================\n"
	s += fmt.Sprintf("Select ending chapter (%d-%d, default: %d", m.startChapter, m.total, m.total)
	if currentBook > 0 {
		s += fmt.Sprintf(", Book %d", currentBook)
	}
	s += "):\n"
	s += "↑/↓ j/k navigate  [ start of book  ] end of book  G last  type number+Enter to jump\n\n"

	windowSize := 10
	start := max(m.startChapter-1, max(0, m.cursor-windowSize/2))
	end := min(m.total-1, start+windowSize-1)
	if end == m.total-1 {
		start = max(m.startChapter-1, m.total-windowSize)
	}

	if start > m.startChapter-1 {
		s += fmt.Sprintf("  ... (%d more above)\n", start-(m.startChapter-1))
	}

	prevBook := -1
	for i := start; i <= end; i++ {
		if len(m.books) > 0 {
			book := m.books[i]
			if book != prevBook {
				s += bookHeaderStyle.Render(fmt.Sprintf("── Book %d ──────────────", book)) + "\n"
				prevBook = book
			}
		}

		chapterTitle := fmt.Sprintf("Chapter %d", i+1)
		if m.chapters != nil && i < len(m.chapters) {
			chapterTitle = m.chapters[i].Title
		}

		inRange := i >= m.startChapter-1 && i <= m.cursor

		if m.cursor == i {
			s += cursorStyle.Render(fmt.Sprintf("> %d. %s", i+1, chapterTitle)) + "\n"
		} else if inRange {
			s += selectedStyle.Render(fmt.Sprintf("  %d. %s", i+1, chapterTitle)) + "\n"
		} else {
			s += fmt.Sprintf("  %d. %s\n", i+1, chapterTitle)
		}
	}

	if end < m.total-1 {
		s += fmt.Sprintf("  ... (%d more below)\n", m.total-1-end)
	}

	if m.inputBuffer != "" {
		s += "\n" + inputStyle.Render(fmt.Sprintf("Jump to chapter: %s▌", m.inputBuffer))
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
			return totalChapters
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
	books, bookMap := computeBooks(chapters)
	m := endChapterSelectorModel{
		chapters:     chapters,
		cursor:       startChapter - 1,
		total:        len(chapters),
		startChapter: startChapter,
		books:        books,
		bookMap:      bookMap,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
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
