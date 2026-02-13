package reader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	Title string
	Lines []string
}

type Location struct {
	Line int
}

func Load(path string) (*Document, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", "":
		return loadTXT(path)
	case ".epub", ".pdf":
		return loadPlaceholder(path, ext)
	default:
		return loadTXT(path)
	}
}

func loadTXT(path string) (*Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, 2048)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		lines = append(lines, "")
	}

	return &Document{Title: filepath.Base(path), Lines: lines}, nil
}

func loadPlaceholder(path, format string) (*Document, error) {
	lines := []string{
		fmt.Sprintf("%s preview adapter is not fully implemented yet.", strings.ToUpper(strings.TrimPrefix(format, "."))),
		"",
		"You can still track reading mode, preferences, sync queue, and session state.",
		fmt.Sprintf("File: %s", path),
	}
	return &Document{Title: filepath.Base(path), Lines: lines}, nil
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
