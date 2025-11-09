package epub

import (
	"errors"
	"os"
	"testing"

	"github.com/linuxswords/wandering-inn/internal/models"
)

// Mock implementation of ChapterContentFetcher for testing
type mockChapterContentFetcher struct {
	chapters map[string]string
	errors   map[string]error
}

func (m *mockChapterContentFetcher) FetchChapterContent(url, title string) (string, error) {
	if err, exists := m.errors[url]; exists {
		return "", err
	}
	if content, exists := m.chapters[url]; exists {
		return content, nil
	}
	return "<p>Default chapter content for " + title + "</p>", nil
}

func TestNewEPUBCreator(t *testing.T) {
	creator := NewEPUBCreator()
	if creator == nil {
		t.Error("NewEPUBCreator() returned nil")
	}
}

func TestEPUBCreator_SetProgressCallback(t *testing.T) {
	creator := NewEPUBCreator()

	var callbackCalled bool
	var receivedCurrent, receivedTotal int
	var receivedTitle string

	callback := func(current, total int, title string) {
		callbackCalled = true
		receivedCurrent = current
		receivedTotal = total
		receivedTitle = title
	}

	creator.SetProgressCallback(callback)

	// Test that the callback is stored and can be called
	if creator.progressCallback != nil {
		creator.progressCallback(1, 10, "Test Chapter")
	}

	if !callbackCalled {
		t.Error("Progress callback was not called")
	}
	if receivedCurrent != 1 {
		t.Errorf("Expected current=1, got %d", receivedCurrent)
	}
	if receivedTotal != 10 {
		t.Errorf("Expected total=10, got %d", receivedTotal)
	}
	if receivedTitle != "Test Chapter" {
		t.Errorf("Expected title='Test Chapter', got %q", receivedTitle)
	}
}

func TestEPUBCreator_CreateEPUB(t *testing.T) {
	creator := NewEPUBCreator()

	// Create test chapters
	chapters := []models.Chapter{
		{Title: "Chapter 1", URL: "url1", Index: 0},
		{Title: "Chapter 2", URL: "url2", Index: 1},
	}

	// Create mock fetcher
	fetcher := &mockChapterContentFetcher{
		chapters: map[string]string{
			"url1": "<p>Content for chapter 1</p>",
			"url2": "<p>Content for chapter 2</p>",
		},
	}

	// Track progress callback calls
	var progressCalls []struct {
		current int
		total   int
		title   string
	}

	creator.SetProgressCallback(func(current, total int, title string) {
		progressCalls = append(progressCalls, struct {
			current int
			total   int
			title   string
		}{current, total, title})
	})

	// Create EPUB
	err := creator.CreateEPUB(chapters, fetcher)
	if err != nil {
		t.Fatalf("CreateEPUB() failed: %v", err)
	}

	// Check that progress callbacks were called
	if len(progressCalls) != 2 {
		t.Errorf("Expected 2 progress calls, got %d", len(progressCalls))
	}

	// Check first progress call
	if len(progressCalls) > 0 {
		call := progressCalls[0]
		if call.current != 1 || call.total != 2 || call.title != "Chapter 1" {
			t.Errorf("First progress call: expected (1, 2, 'Chapter 1'), got (%d, %d, %q)",
				call.current, call.total, call.title)
		}
	}

	// Check that EPUB file was created
	expectedFilename := "wandering_inn_chapter_1.epub"
	if _, err := os.Stat(expectedFilename); os.IsNotExist(err) {
		t.Errorf("Expected EPUB file %s was not created", expectedFilename)
	} else {
		// Clean up the test file
		os.Remove(expectedFilename)
	}
}

func TestEPUBCreator_CreateEPUB_EmptyChapters(t *testing.T) {
	creator := NewEPUBCreator()

	// Create empty chapters slice
	chapters := []models.Chapter{}

	// Create mock fetcher
	fetcher := &mockChapterContentFetcher{
		chapters: map[string]string{},
	}

	// Create EPUB
	err := creator.CreateEPUB(chapters, fetcher)
	if err != nil {
		t.Fatalf("CreateEPUB() with empty chapters failed: %v", err)
	}

	// Check that default EPUB file was created
	expectedFilename := "wandering_inn.epub"
	if _, err := os.Stat(expectedFilename); os.IsNotExist(err) {
		t.Errorf("Expected EPUB file %s was not created", expectedFilename)
	} else {
		// Clean up the test file
		os.Remove(expectedFilename)
	}
}

func TestEPUBCreator_CreateEPUB_FetchError(t *testing.T) {
	creator := NewEPUBCreator()

	// Create test chapters
	chapters := []models.Chapter{
		{Title: "Chapter 1", URL: "url1", Index: 0},
		{Title: "Chapter 2", URL: "url2", Index: 1},
	}

	// Create mock fetcher with error for one chapter
	fetcher := &mockChapterContentFetcher{
		chapters: map[string]string{
			"url1": "<p>Content for chapter 1</p>",
		},
		errors: map[string]error{
			"url2": errors.New("failed to fetch chapter 2"),
		},
	}

	// Create EPUB (should continue despite one chapter failing)
	err := creator.CreateEPUB(chapters, fetcher)
	if err != nil {
		t.Fatalf("CreateEPUB() failed: %v", err)
	}

	// Check that EPUB file was created despite the error
	expectedFilename := "wandering_inn_chapter_1.epub"
	if _, err := os.Stat(expectedFilename); os.IsNotExist(err) {
		t.Errorf("Expected EPUB file %s was not created", expectedFilename)
	} else {
		// Clean up the test file
		os.Remove(expectedFilename)
	}
}

func TestEPUBCreator_CreateEPUB_NoProgressCallback(t *testing.T) {
	creator := NewEPUBCreator()
	// Don't set a progress callback

	// Create test chapters
	chapters := []models.Chapter{
		{Title: "Chapter 1", URL: "url1", Index: 0},
	}

	// Create mock fetcher
	fetcher := &mockChapterContentFetcher{
		chapters: map[string]string{
			"url1": "<p>Content for chapter 1</p>",
		},
	}

	// Create EPUB (should work without progress callback)
	err := creator.CreateEPUB(chapters, fetcher)
	if err != nil {
		t.Fatalf("CreateEPUB() without progress callback failed: %v", err)
	}

	// Check that EPUB file was created
	expectedFilename := "wandering_inn_chapter_1.epub"
	if _, err := os.Stat(expectedFilename); os.IsNotExist(err) {
		t.Errorf("Expected EPUB file %s was not created", expectedFilename)
	} else {
		// Clean up the test file
		os.Remove(expectedFilename)
	}
}

// Test that EPUBCreator implements the Creator interface
func TestEPUBCreator_ImplementsInterface(t *testing.T) {
	var _ Creator = (*EPUBCreator)(nil)
}

// Test that mockChapterContentFetcher implements ChapterContentFetcher interface
func TestMockFetcher_ImplementsInterface(t *testing.T) {
	var _ ChapterContentFetcher = (*mockChapterContentFetcher)(nil)
}
