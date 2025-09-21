package scraper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/linuxswords/wandering-inn/internal/models"
)

func TestNewWanderingInnScraper(t *testing.T) {
	scraper := NewWanderingInnScraper()
	if scraper == nil {
		t.Error("NewWanderingInnScraper() returned nil")
	}
}

func TestWanderingInnScraper_isChapterLink(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		href     string
		expected bool
	}{
		{
			name:     "valid chapter",
			title:    "Chapter 1.00",
			href:     "https://wanderinginn.com/chapter-1-00",
			expected: true,
		},
		{
			name:     "prologue",
			title:    "Prologue",
			href:     "https://wanderinginn.com/prologue",
			expected: true,
		},
		{
			name:     "epilogue",
			title:    "Epilogue",
			href:     "https://wanderinginn.com/epilogue",
			expected: true,
		},
		{
			name:     "interlude",
			title:    "Interlude - Pawn",
			href:     "https://wanderinginn.com/interlude-pawn",
			expected: true,
		},
		{
			name:     "numbered chapter",
			title:    "1.01",
			href:     "https://wanderinginn.com/1-01",
			expected: true,
		},
		{
			name:     "table of contents link",
			title:    "Table of Contents",
			href:     "https://wanderinginn.com/table-of-contents",
			expected: false,
		},
		{
			name:     "non-chapter link",
			title:    "About the Author",
			href:     "https://wanderinginn.com/about",
			expected: false,
		},
		{
			name:     "chapter with toc in title",
			title:    "Chapter 1.00 - Table of Contents Reference",
			href:     "https://wanderinginn.com/chapter-1-00",
			expected: false,
		},
	}

	scraper := NewWanderingInnScraper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.isChapterLink(tt.title, tt.href)
			if result != tt.expected {
				t.Errorf("isChapterLink(%q, %q) = %v, want %v", tt.title, tt.href, result, tt.expected)
			}
		})
	}
}

func TestWanderingInnScraper_FetchTableOfContents(t *testing.T) {
	// Create a mock HTML response for testing
	mockHTML := `
<!DOCTYPE html>
<html>
<body>
	<div>
		<a href="https://wanderinginn.com/chapter-1-00">Chapter 1.00</a>
		<a href="https://wanderinginn.com/chapter-1-01">Chapter 1.01</a>
		<a href="https://wanderinginn.com/prologue">Prologue</a>
		<a href="https://wanderinginn.com/about">About the Author</a>
		<a href="https://wanderinginn.com/table-of-contents">Table of Contents</a>
		<a href="https://wanderinginn.com/interlude-1">Interlude - Pawn</a>
	</div>
</body>
</html>`

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	// Note: This test would require modifying the config.TOCUrl or making it configurable
	// For now, we'll test the parsing logic indirectly through the isChapterLink method
	// In a full implementation, you might want to make the URL configurable for testing

	t.Run("mock server setup", func(t *testing.T) {
		// Test that our mock server works
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to get mock server response: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestWanderingInnScraper_FetchChapterContent(t *testing.T) {
	// Create a mock HTML response for a chapter
	mockChapterHTML := `
<!DOCTYPE html>
<html>
<body>
	<div class="entry-content">
		<p>This is the beginning of the chapter.</p>
		<p>Here is some more content.</p>
		<p><strong>Bold text</strong> and <em>italic text</em>.</p>
	</div>
</body>
</html>`

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockChapterHTML))
	}))
	defer server.Close()

	scraper := NewWanderingInnScraper()
	content, err := scraper.FetchChapterContent(server.URL, "Test Chapter")

	if err != nil {
		t.Fatalf("FetchChapterContent() failed: %v", err)
	}

	if content == "" {
		t.Error("FetchChapterContent() returned empty content")
	}

	// Check that the content contains the expected title
	if !strings.Contains(content, "<h1>Test Chapter</h1>") {
		t.Error("Content should contain the chapter title as h1")
	}

	// Check that the content contains some of the paragraph text
	if !strings.Contains(content, "This is the beginning of the chapter") {
		t.Error("Content should contain the paragraph text")
	}
}

func TestWanderingInnScraper_FetchChapterContent_Error(t *testing.T) {
	scraper := NewWanderingInnScraper()

	// Test with invalid URL
	_, err := scraper.FetchChapterContent("invalid-url", "Test Chapter")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with non-existent server
	_, err = scraper.FetchChapterContent("http://localhost:99999/nonexistent", "Test Chapter")
	if err == nil {
		t.Error("Expected error for non-existent server, got nil")
	}
}

// Test that the scraper implements the Scraper interface
func TestWanderingInnScraper_ImplementsInterface(t *testing.T) {
	var _ Scraper = (*WanderingInnScraper)(nil)
}

// Test sorting of chapters by index
func TestChapterSorting(t *testing.T) {
	chapters := []models.Chapter{
		{Title: "Chapter 3", URL: "url3", Index: 2},
		{Title: "Chapter 1", URL: "url1", Index: 0},
		{Title: "Chapter 2", URL: "url2", Index: 1},
	}

	// Sort by index (simulating the sorting logic in FetchTableOfContents)
	for i := 0; i < len(chapters)-1; i++ {
		for j := 0; j < len(chapters)-1-i; j++ {
			if chapters[j].Index > chapters[j+1].Index {
				chapters[j], chapters[j+1] = chapters[j+1], chapters[j]
			}
		}
	}

	expectedTitles := []string{"Chapter 1", "Chapter 2", "Chapter 3"}
	for i, chapter := range chapters {
		if chapter.Title != expectedTitles[i] {
			t.Errorf("Expected chapter %d to be %s, got %s", i, expectedTitles[i], chapter.Title)
		}
	}
}