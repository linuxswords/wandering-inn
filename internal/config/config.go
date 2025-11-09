package config

import "regexp"

const (
	TOCUrl = "https://wanderinginn.com/table-of-contents/"

	EpubTitle       = "The Wandering Inn"
	EpubAuthor      = "pirateaba"
	EpubDescription = "The Wandering Inn web serial"

	DefaultFilename = "wandering_inn.epub"
	MaxFilenameLen  = 50

	LatestChaptersCount = 20
)

var (
	ChapterPattern = regexp.MustCompile(`(?i)(chapter|prologue|epilogue|interlude|\d+\.\d+)`)

	NavigationTerms = []string{
		"previous chapter",
		"next chapter",
		"← previous",
		"next →",
		"table of contents",
		"toc",
		"chapter index",
		"first chapter",
		"last chapter",
	}

	NavigationClasses = []string{
		"navigation",
		"nav",
		"chapter-nav",
		"post-nav",
		"entry-nav",
		"pagination",
		"prev-next",
		"chapter-links",
	}

	NavigationSymbolPattern = regexp.MustCompile(`^(←|→|«|»|\|)+$`)

	ColorClassMap = map[string]string{
		"has-red-color":     "red",
		"has-blue-color":    "blue",
		"has-green-color":   "green",
		"has-purple-color":  "purple",
		"has-orange-color":  "orange",
		"has-yellow-color":  "yellow",
		"has-brown-color":   "brown",
		"has-pink-color":    "pink",
		"has-cyan-color":    "cyan",
		"has-gray-color":    "gray",
		"has-grey-color":    "gray",
		"has-gold-color":    "gold",
		"has-silver-color":  "silver",
		"has-crimson-color": "crimson",
		"has-maroon-color":  "maroon",
		"has-navy-color":    "navy",
		"has-teal-color":    "teal",
	}
)

