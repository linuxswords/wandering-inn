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
			Foreground(lipgloss.Color("15")). // White
			Background(lipgloss.Color("63")). // Purple
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).   // Black
			Background(lipgloss.Color("120")). // Light green
			Bold(false)

	bookHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")). // Yellow
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")). // Cyan
			Bold(true)
)

// bookPattern matches chapter titles starting with a book number e.g. "7.6 R" or "10.62 NY".
var bookPattern = regexp.MustCompile(`^(\d+)\.`)

// computeBooks assigns a book number to every chapter. Chapters whose titles start with
// "N." get that N as their book; interludes and specials inherit the last seen book.
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

// bookEntry holds the computed chapter range for one book.
type bookEntry struct {
	bookNum      int
	firstChapter int // 1-indexed
	lastChapter  int // 1-indexed
	firstTitle   string
	lastTitle    string
}

func buildBookEntries(chapters []models.Chapter, books []int, bookMap map[int]int) []bookEntry {
	seen := make(map[int]bool)
	var bookNums []int
	for _, b := range books {
		if !seen[b] {
			seen[b] = true
			bookNums = append(bookNums, b)
		}
	}

	entries := make([]bookEntry, len(bookNums))
	for i, bn := range bookNums {
		first := bookMap[bn]
		var last int
		if i+1 < len(bookNums) {
			last = bookMap[bookNums[i+1]] - 1
		} else {
			last = len(chapters) - 1
		}
		entries[i] = bookEntry{
			bookNum:      bn,
			firstChapter: first + 1,
			lastChapter:  last + 1,
			firstTitle:   chapters[first].Title,
			lastTitle:    chapters[last].Title,
		}
	}
	return entries
}

// ── Book selector ─────────────────────────────────────────────────────────────

type bookSelectorModel struct {
	entries     []bookEntry
	cursor      int
	total       int
	quit        bool
	wholeBook   bool // true = Enter pressed (whole book), false = c pressed (custom)
	inputBuffer string
}

func (m bookSelectorModel) Init() tea.Cmd { return nil }

func (m bookSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if err == nil {
					for i, e := range m.entries {
						if e.bookNum == n {
							m.cursor = i
							break
						}
					}
				}
			} else {
				m.wholeBook = true
				return m, tea.Quit
			}
		case "c":
			m.inputBuffer = ""
			m.wholeBook = false
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m bookSelectorModel) View() string {
	s := "Wandering Inn EPUB Creator\n"
	s += "==========================\n"
	s += "Select a book:\n"
	s += "↑/↓ j/k navigate  g/G first/last  type book number+Enter to jump\n"
	s += "Enter = whole book  c = pick chapters within book\n\n"

	windowSize := 12
	start := max(0, m.cursor-windowSize/2)
	end := min(m.total-1, start+windowSize-1)
	if end == m.total-1 {
		start = max(0, m.total-windowSize)
	}

	if start > 0 {
		s += fmt.Sprintf("  ... (%d more above)\n", start)
	}

	for i := start; i <= end; i++ {
		e := m.entries[i]
		count := e.lastChapter - e.firstChapter + 1
		line := fmt.Sprintf("Book %-3d  %3d chapters  (ch %d–%d)  %s → %s",
			e.bookNum, count, e.firstChapter, e.lastChapter, e.firstTitle, e.lastTitle)
		if m.cursor == i {
			s += cursorStyle.Render("> "+line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	if end < m.total-1 {
		s += fmt.Sprintf("  ... (%d more below)\n", m.total-1-end)
	}

	if m.inputBuffer != "" {
		s += "\n" + inputStyle.Render(fmt.Sprintf("Jump to book: %s▌", m.inputBuffer))
	}

	return s
}

// ── Chapter selectors ──────────────────────────────────────────────────────────

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

func (m chapterSelectorModel) Init() tea.Cmd { return nil }

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

func (m endChapterSelectorModel) Init() tea.Cmd { return nil }

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
		case "g":
			m.inputBuffer = ""
			m.cursor = minCursor
		case "G":
			m.inputBuffer = ""
			m.cursor = m.total - 1
		case "[":
			m.inputBuffer = ""
			if len(m.books) > 0 {
				cur := m.books[m.cursor]
				if bookStart, ok := m.bookMap[cur]; ok {
					m.cursor = max(minCursor, bookStart)
				}
			}
		case "]":
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
	s += fmt.Sprintf("Select ending chapter (%d-%d, default: %d", m.startChapter, m.total, m.cursor+1)
	if currentBook > 0 {
		s += fmt.Sprintf(", Book %d", currentBook)
	}
	s += "):\n"
	s += "↑/↓ j/k navigate  [ start of book  ] end of book  g/G first/last  type number+Enter to jump\n\n"

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

// ── CLI methods ────────────────────────────────────────────────────────────────

type CLI struct {
	reader *bufio.Reader
}

func NewCLI() *CLI {
	return &CLI{reader: bufio.NewReader(os.Stdin)}
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

// GetChapterRange shows the book selector first, then either returns the whole
// book immediately (Enter) or drops into chapter selectors (c).
func (cli *CLI) GetChapterRange(chapters []models.Chapter) (int, int) {
	if len(chapters) == 0 {
		fmt.Println("No chapters found.")
		os.Exit(1)
	}

	books, bookMap := computeBooks(chapters)
	entries := buildBookEntries(chapters, books, bookMap)

	bm := bookSelectorModel{
		entries: entries,
		cursor:  len(entries) - 1,
		total:   len(entries),
	}

	p := tea.NewProgram(bm, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		start := cli.runStartSelector(chapters, books, bookMap, len(chapters)-1)
		end := cli.runEndSelector(chapters, books, bookMap, start)
		return start, end
	}

	result := finalModel.(bookSelectorModel)
	if result.quit {
		fmt.Println("Exiting...")
		os.Exit(0)
	}

	selected := result.entries[result.cursor]

	if result.wholeBook {
		return selected.firstChapter, selected.lastChapter
	}

	// Custom: chapter selectors starting at the book's first chapter.
	startChapter := cli.runStartSelector(chapters, books, bookMap, selected.firstChapter-1)
	endChapter := cli.runEndSelector(chapters, books, bookMap, startChapter)
	return startChapter, endChapter
}

func (cli *CLI) runStartSelector(chapters []models.Chapter, books []int, bookMap map[int]int, defaultCursor int) int {
	m := chapterSelectorModel{
		chapters: chapters,
		cursor:   defaultCursor,
		total:    len(chapters),
		books:    books,
		bookMap:  bookMap,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return cli.getStartChapterTextInput(len(chapters))
	}

	result := finalModel.(chapterSelectorModel)
	if result.quit {
		fmt.Println("Exiting...")
		os.Exit(0)
	}
	return result.selected
}

func (cli *CLI) runEndSelector(chapters []models.Chapter, books []int, bookMap map[int]int, startChapter int) int {
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

	result := finalModel.(endChapterSelectorModel)
	if result.quit {
		fmt.Println("Exiting...")
		os.Exit(0)
	}
	return result.selected
}

// GetStartChapterInteractive is kept for compatibility.
func (cli *CLI) GetStartChapterInteractive(chapters []models.Chapter) int {
	books, bookMap := computeBooks(chapters)
	return cli.runStartSelector(chapters, books, bookMap, len(chapters)-1)
}

// GetEndChapterInteractive is kept for compatibility.
func (cli *CLI) GetEndChapterInteractive(chapters []models.Chapter, startChapter int) int {
	books, bookMap := computeBooks(chapters)
	return cli.runEndSelector(chapters, books, bookMap, startChapter)
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
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > totalChapters {
			fmt.Printf("Please enter a number between 1 and %d.\n", totalChapters)
			continue
		}
		return n
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
		n, err := strconv.Atoi(input)
		if err != nil || n < startChapter || n > totalChapters {
			fmt.Printf("Please enter a number between %d and %d.\n", startChapter, totalChapters)
			continue
		}
		return n
	}
}

func (cli *CLI) PrintCreationInfo(numChapters, startIndex, endIndex int) {
	if startIndex == endIndex {
		fmt.Printf("Creating EPUB with 1 chapter: chapter %d...\n", startIndex)
	} else {
		fmt.Printf("Creating EPUB with %d chapters from chapter %d to %d...\n", numChapters, startIndex, endIndex)
	}
}

func (cli *CLI) PrintDownloadProgress(current, total int, title string) {
	fmt.Printf("Downloading chapter %d/%d: %s\n", current, total, title)
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
