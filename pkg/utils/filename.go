package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/linuxswords/wandering-inn/internal/config"
	"github.com/linuxswords/wandering-inn/internal/models"
)

func GenerateFilename(chapters []models.Chapter) string {
	if len(chapters) == 0 {
		return config.DefaultFilename
	}

	startChapter := SanitizeFilename(chapters[0].Title)
	return fmt.Sprintf("wandering_inn_%s.epub", startChapter)
}

func SanitizeFilename(title string) string {
	title = strings.TrimSpace(title)

	title = regexp.MustCompile(`[^\w\s\-\.]`).ReplaceAllString(title, "")

	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, "_")

	title = strings.ToLower(title)

	if len(title) > config.MaxFilenameLen {
		title = title[:config.MaxFilenameLen]
	}

	title = strings.Trim(title, "_-.")

	if title == "" {
		return "chapter"
	}

	return title
}

