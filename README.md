# Wandering Inn EPUB Creator

A Go tool that creates EPUB files from The Wandering Inn webserial by pirateaba.

## Features

- Downloads table of contents from wanderinginn.com
- Interactive chapter selection - choose which chapter to start from
- Downloads all chapters from selected chapter to the end in correct order
- Creates a properly formatted EPUB file

## Installation

1. Make sure you have Go installed (version 1.21 or higher)
2. Clone or download this project
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

1. Run the tool:
   ```bash
   go run main.go
   ```

2. The tool will:
   - Fetch the table of contents from wanderinginn.com
   - Display the first few chapters found
   - Ask you which chapter to start downloading from
   - Download all chapters from that point to the end
   - Create `wandering_inn.epub` in the current directory

## Example

```bash
$ go run main.go
Wandering Inn EPUB Creator
==========================
Found 450 chapters
First few chapters:
1. Prologue
2. 1.00
3. 1.01
4. 1.02
5. 1.03
Enter starting chapter number (1-450): 100
Creating EPUB with 351 chapters starting from chapter 100...
Downloading chapter 1/351: 2.00
Downloading chapter 2/351: 2.01
...
EPUB created successfully: wandering_inn.epub
```

## Dependencies

- [go-epub](https://github.com/go-shiori/go-epub) - For EPUB creation
- Standard Go libraries for HTTP requests and HTML parsing

## Notes

- The tool ensures chapters are downloaded in the correct order as they appear in the table of contents
- If a chapter fails to download, the tool will show a warning and continue with the next chapter
- The resulting EPUB file will be named `wandering_inn_chaptertitle.epub`
