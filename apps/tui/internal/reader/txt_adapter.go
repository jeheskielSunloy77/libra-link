package reader

import (
	"bufio"
	"os"
	"path/filepath"
)

func loadTXT(path string) (*Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc := &Document{
		Title:     filepath.Base(path),
		Format:    "txt",
		Lines:     make([]string, 0, 2048),
		LineIndex: make([]LineAnchor, 0, 2048),
	}

	scanner := bufio.NewScanner(file)
	offset := 0
	for scanner.Scan() {
		appendLine(doc, scanner.Text(), LineAnchor{Page: -1, Spine: -1, Offset: offset})
		offset++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(doc.Lines) == 0 {
		appendLine(doc, "", LineAnchor{Page: -1, Spine: -1, Offset: 0})
	}

	return doc, nil
}
