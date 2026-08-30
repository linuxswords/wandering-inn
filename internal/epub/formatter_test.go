package epub

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	if formatter == nil {
		t.Error("NewFormatter() returned nil")
	}
	if formatter.css != DefaultCSS {
		t.Error("NewFormatter() did not set default CSS")
	}
}

func TestFormatter_SetCSS(t *testing.T) {
	formatter := NewFormatter()
	customCSS := "body { font-family: Arial; }"

	formatter.SetCSS(customCSS)

	if formatter.css != customCSS {
		t.Errorf("SetCSS() = %v, want %v", formatter.css, customCSS)
	}
}

func TestFormatter_GetCSS(t *testing.T) {
	tests := []struct {
		name string
		css  string
	}{
		{
			name: "default CSS",
			css:  DefaultCSS,
		},
		{
			name: "custom CSS",
			css:  "body { color: red; }",
		},
		{
			name: "empty CSS",
			css:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter()
			formatter.SetCSS(tt.css)

			result := formatter.GetCSS()
			if result != tt.css {
				t.Errorf("GetCSS() = %v, want %v", result, tt.css)
			}
		})
	}
}

func TestFormatter_DataURL(t *testing.T) {
	formatter := NewFormatter()
	css := "body { color: red; }"
	formatter.SetCSS(css)

	dataURL := formatter.DataURL()

	const prefix = "data:text/css;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		t.Fatalf("DataURL() = %q, want prefix %q", dataURL, prefix)
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		t.Fatalf("DataURL() payload is not valid base64: %v", err)
	}
	if string(decoded) != css {
		t.Errorf("DataURL() decodes to %q, want %q", decoded, css)
	}
}

func TestDefaultCSS(t *testing.T) {
	// Test that DefaultCSS contains expected style definitions
	expectedStyles := []string{
		"body {",
		"font-family: Georgia, serif;",
		"line-height: 1.6;",
		"color: #333;",
		"h1 {",
		".red { color: #e74c3c; }",
		".blue { color: #3498db; }",
		".green { color: #27ae60; }",
		".purple { color: #9b59b6; }",
	}

	for _, style := range expectedStyles {
		if !strings.Contains(DefaultCSS, style) {
			t.Errorf("DefaultCSS missing expected style: %s", style)
		}
	}
}

func TestDefaultCSS_ColorStyles(t *testing.T) {
	// Test that all expected color classes are present
	expectedColors := []string{
		"red", "blue", "green", "purple", "orange", "yellow",
		"brown", "pink", "cyan", "gray", "gold", "silver",
		"crimson", "maroon", "navy", "teal",
	}

	for _, color := range expectedColors {
		className := "." + color + " {"
		if !strings.Contains(DefaultCSS, className) {
			t.Errorf("DefaultCSS missing color class: %s", color)
		}
	}
}
