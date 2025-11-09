# Wandering Inn EPUB Creator

[![Tests](https://github.com/linuxswords/wandering-inn/actions/workflows/test.yml/badge.svg)](https://github.com/linuxswords/wandering-inn/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/linuxswords/wandering-inn)](https://goreportcard.com/report/github.com/linuxswords/wandering-inn)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go tool that creates EPUB files from The Wandering Inn webserial by pirateaba.

## Features

- Downloads table of contents from wanderinginn.com
- **Interactive chapter selection UI** with arrow keys/vim bindings
  - Choose which chapter to start from
  - Choose which chapter to end at
  - **Color highlighting** shows your current selection and selected range
- Downloads chosen chapters in correct order
- Creates a properly formatted EPUB file

## Installation

1. Make sure you have Go installed (version 1.21 or higher)
2. Clone this repository:

   ```bash
   git clone https://github.com/linuxswords/wandering-inn.git
   cd wandering-inn
   ```

3. Install dependencies:

   ```bash
   go mod download
   ```

## Usage

1. Run the tool:

   ```bash
   go run ./cmd/wandering-inn
   ```

   Or build and run:

   ```bash
   go build ./cmd/wandering-inn
   ./wandering-inn
   ```

2. The tool will:
   - Fetch the table of contents from wanderinginn.com
   - Show an **interactive chapter selector** (use ↑/↓ arrow keys or j/k vim keys)
   - Ask you which chapter to **start** from
   - Ask you which chapter to **end** at (with color highlighting of your selection)
   - Download all selected chapters
   - Create an EPUB file in the current directory (e.g., `wandering_inn_chapter1-100.epub`)

## Example

```bash
$ go run ./cmd/wandering-inn
Wandering Inn EPUB Creator
==========================
Select starting chapter (1-450):
Use ↑/↓ arrow keys or j/k (vim keys) to navigate, Enter to select, 'q' to quit

  ...
  > 100. 2.00
    101. 2.01
  ...

[After selecting start chapter]

Select ending chapter (100-450, default: 450):
Use ↑/↓ arrow keys or j/k (vim keys) to navigate, Enter to select, 'q' to quit

  ...
    100. 2.00     [highlighted in green - start of range]
    101. 2.01     [highlighted in green - in range]
  > 150. 2.51     [highlighted in purple - cursor position]
    151. 2.52
  ...

Creating EPUB with 51 chapters from chapter 100 to 150...
Downloading chapter 1/51: 2.00
Downloading chapter 2/51: 2.01
...
EPUB created successfully: wandering_inn_2.00-2.51.epub
```

## Dependencies

- [go-epub](https://github.com/go-shiori/go-epub) - For EPUB creation
- [bubbletea](https://github.com/charmbracelet/bubbletea) - For interactive terminal UI
- [lipgloss](https://github.com/charmbracelet/lipgloss) - For terminal styling and colors
- [golang.org/x/net](https://pkg.go.dev/golang.org/x/net) - For HTML parsing

## Notes

- The tool ensures chapters are downloaded in the correct order as they appear in the table of contents
- If a chapter fails to download, the tool will show a warning and continue with the next chapter
- The resulting EPUB file will be named based on the selected chapters (e.g., `wandering_inn_2.00-2.51.epub`)
- You can quit the interactive selectors at any time by pressing 'q' or ESC

## License

This project is licensed under the AGPLv3 License - see the [LICENSE](LICENSE) file for details.
