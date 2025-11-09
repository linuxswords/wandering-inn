package utils

import (
	"testing"

	"github.com/linuxswords/wandering-inn/internal/models"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		chapters []models.Chapter
		expected string
	}{
		{
			name:     "empty chapters",
			chapters: []models.Chapter{},
			expected: "wandering_inn.epub",
		},
		{
			name: "single chapter",
			chapters: []models.Chapter{
				{Title: "Chapter 1.00", URL: "test", Index: 0},
			},
			expected: "wandering_inn_chapter_1.00.epub",
		},
		{
			name: "multiple chapters",
			chapters: []models.Chapter{
				{Title: "Chapter 1.00", URL: "test", Index: 0},
				{Title: "Chapter 1.01", URL: "test", Index: 1},
			},
			expected: "wandering_inn_chapter_1.00.epub",
		},
		{
			name: "chapter with special characters",
			chapters: []models.Chapter{
				{Title: "Chapter 1.00 - The Last Hero", URL: "test", Index: 0},
			},
			expected: "wandering_inn_chapter_1.00_-_the_last_hero.epub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFilename(tt.chapters)
			if result != tt.expected {
				t.Errorf("GenerateFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple title",
			input:    "Chapter 1.00",
			expected: "chapter_1.00",
		},
		{
			name:     "title with spaces",
			input:    "Chapter 1.00 The Last Hero",
			expected: "chapter_1.00_the_last_hero",
		},
		{
			name:     "title with special characters",
			input:    "Chapter 1.00: The Last Hero!",
			expected: "chapter_1.00_the_last_hero",
		},
		{
			name:     "title with multiple spaces",
			input:    "Chapter   1.00    The Last Hero",
			expected: "chapter_1.00_the_last_hero",
		},
		{
			name:     "very long title",
			input:    "This is a very long chapter title that exceeds the maximum length limit",
			expected: "this_is_a_very_long_chapter_title_that_exceeds_the",
		},
		{
			name:     "title with leading/trailing characters",
			input:    "___Chapter 1.00___",
			expected: "chapter_1.00",
		},
		{
			name:     "empty title",
			input:    "",
			expected: "chapter",
		},
		{
			name:     "title with only special characters",
			input:    "!@#$%^&*()",
			expected: "chapter",
		},
		{
			name:     "title with unicode characters",
			input:    "Chapter 1.00 – The Last Hero",
			expected: "chapter_1.00_the_last_hero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
