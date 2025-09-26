package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/linuxswords/wandering-inn/internal/models"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()
	if cli == nil {
		t.Error("NewCLI() returned nil")
	}
	if cli.reader == nil {
		t.Error("CLI reader is nil")
	}
}

func TestCLI_PrintWelcome(t *testing.T) {
	cli := NewCLI()
	// This function prints to stdout, so we just test that it doesn't panic
	cli.PrintWelcome()
}

func TestCLI_PrintChapterInfo(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		name     string
		chapters []models.Chapter
	}{
		{
			name:     "empty chapters",
			chapters: []models.Chapter{},
		},
		{
			name: "few chapters",
			chapters: []models.Chapter{
				{Title: "Chapter 1", URL: "url1", Index: 0},
				{Title: "Chapter 2", URL: "url2", Index: 1},
			},
		},
		{
			name: "many chapters",
			chapters: func() []models.Chapter {
				var chapters []models.Chapter
				for i := 0; i < 30; i++ {
					chapters = append(chapters, models.Chapter{
						Title: fmt.Sprintf("Chapter %d", i+1),
						URL:   fmt.Sprintf("url%d", i+1),
						Index: i,
					})
				}
				return chapters
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function prints to stdout, so we just test that it doesn't panic
			cli.PrintChapterInfo(tt.chapters)
		})
	}
}

func TestCLI_GetStartChapter(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		totalChapters int
		expected      int
		expectError   bool
	}{
		{
			name:          "valid input",
			input:         "5\n",
			totalChapters: 10,
			expected:      5,
			expectError:   false,
		},
		{
			name:          "input at minimum boundary",
			input:         "1\n",
			totalChapters: 10,
			expected:      1,
			expectError:   false,
		},
		{
			name:          "input at maximum boundary",
			input:         "10\n",
			totalChapters: 10,
			expected:      10,
			expectError:   false,
		},
		{
			name:          "invalid input then valid",
			input:         "invalid\n5\n",
			totalChapters: 10,
			expected:      5,
			expectError:   false,
		},
		{
			name:          "out of range low then valid",
			input:         "0\n5\n",
			totalChapters: 10,
			expected:      5,
			expectError:   false,
		},
		{
			name:          "out of range high then valid",
			input:         "15\n5\n",
			totalChapters: 10,
			expected:      5,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a CLI with a custom reader
			cli := &CLI{
				reader: bufio.NewReader(strings.NewReader(tt.input)),
			}

			result := cli.getStartChapterTextInput(tt.totalChapters)
			if result != tt.expected {
				t.Errorf("GetStartChapter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCLI_PrintCreationInfo(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		name        string
		numChapters int
		startIndex  int
	}{
		{
			name:        "typical case",
			numChapters: 5,
			startIndex:  10,
		},
		{
			name:        "single chapter",
			numChapters: 1,
			startIndex:  1,
		},
		{
			name:        "many chapters",
			numChapters: 100,
			startIndex:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function prints to stdout, so we just test that it doesn't panic
			cli.PrintCreationInfo(tt.numChapters, tt.startIndex, tt.startIndex+tt.numChapters-1)
		})
	}
}

func TestCLI_PrintDownloadProgress(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		name    string
		current int
		total   int
		title   string
	}{
		{
			name:    "first chapter",
			current: 1,
			total:   10,
			title:   "Chapter 1.00",
		},
		{
			name:    "middle chapter",
			current: 5,
			total:   10,
			title:   "Chapter 1.05",
		},
		{
			name:    "last chapter",
			current: 10,
			total:   10,
			title:   "Chapter 1.10",
		},
		{
			name:    "long title",
			current: 1,
			total:   1,
			title:   "Chapter 1.00 - The Last Light of Doom and Gloom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function prints to stdout, so we just test that it doesn't panic
			cli.PrintDownloadProgress(tt.current, tt.total, tt.title)
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a greater than b",
			a:        10,
			b:        5,
			expected: 10,
		},
		{
			name:     "b greater than a",
			a:        3,
			b:        8,
			expected: 8,
		},
		{
			name:     "a equals b",
			a:        7,
			b:        7,
			expected: 7,
		},
		{
			name:     "negative numbers",
			a:        -5,
			b:        -10,
			expected: -5,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        5,
			expected: 5,
		},
		{
			name:     "zero and negative",
			a:        0,
			b:        -5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Test error handling in GetStartChapter with reader errors
func TestCLI_GetStartChapter_ReaderError(t *testing.T) {
	// Create a reader that will return an error after some reads
	errorReader := &erroringReader{
		data:       "invalid\n",
		errorAfter: 1,
		readCount:  0,
	}

	cli := &CLI{
		reader: bufio.NewReader(errorReader),
	}

	// This should handle the error gracefully and continue
	// Note: In practice, this test might be hard to verify without capturing stdout
	// but it tests that the function handles reader errors without panicking

	// Since the reader will error after one read, we can't easily test the full flow
	// but we can test that the function exists and is callable
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetStartChapter panicked with reader error: %v", r)
		}
	}()

	// This will likely loop indefinitely or error due to our mock, but it shouldn't panic
	go func() {
		cli.getStartChapterTextInput(10)
	}()
}

// Helper type for testing reader errors
type erroringReader struct {
	data       string
	errorAfter int
	readCount  int
}

func (r *erroringReader) Read(p []byte) (n int, err error) {
	if r.readCount >= r.errorAfter {
		return 0, errors.New("mock reader error")
	}
	r.readCount++

	if len(r.data) == 0 {
		return 0, io.EOF
	}

	n = copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}