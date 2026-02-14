package reader

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Document struct {
	Title     string
	Format    string
	Lines     []string
	LineIndex []LineAnchor
}

type Location struct {
	Line int
}

type LineAnchor struct {
	Line   int
	Page   int
	Spine  int
	Offset int
}

func Load(path string) (*Document, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", "":
		return loadTXT(path)
	case ".pdf":
		return loadPDF(path)
	case ".epub":
		return loadEPUB(path)
	default:
		return nil, fmt.Errorf("unsupported document format %q", ext)
	}
}

func ClampLocation(doc *Document, line int) int {
	if doc == nil || len(doc.Lines) == 0 {
		return 0
	}
	if line < 0 {
		return 0
	}
	if line >= len(doc.Lines) {
		return len(doc.Lines) - 1
	}
	return line
}

func appendLine(doc *Document, text string, anchor LineAnchor) {
	anchor.Line = len(doc.Lines)
	doc.Lines = append(doc.Lines, text)
	doc.LineIndex = append(doc.LineIndex, anchor)
}
