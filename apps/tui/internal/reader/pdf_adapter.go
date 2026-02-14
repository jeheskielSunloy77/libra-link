package reader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

func loadPDF(path string) (*Document, error) {
	file, docReader, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc := &Document{
		Title:     filepath.Base(path),
		Format:    "pdf",
		Lines:     make([]string, 0, 4096),
		LineIndex: make([]LineAnchor, 0, 4096),
	}

	numPages := docReader.NumPage()
	hasText := false
	for page := 1; page <= numPages; page++ {
		pdfPage := docReader.Page(page)
		if pdfPage.V.IsNull() {
			continue
		}
		content, err := pdfPage.GetPlainText(nil)
		if err != nil {
			return nil, err
		}

		appendLine(doc, fmt.Sprintf("--- Page %d ---", page), LineAnchor{Page: page, Spine: -1, Offset: 0})

		lines := splitNormalizedLines(content)
		offset := 1
		for _, line := range lines {
			appendLine(doc, line, LineAnchor{Page: page, Spine: -1, Offset: offset})
			offset++
			if strings.TrimSpace(line) != "" {
				hasText = true
			}
		}
	}

	if !hasText {
		return nil, fmt.Errorf("pdf has no extractable text (likely scanned/image-only)")
	}

	return doc, nil
}

func splitNormalizedLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	rawLines := strings.Split(normalized, "\n")
	out := make([]string, 0, len(rawLines))
	prevBlank := false
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !prevBlank && len(out) > 0 {
				out = append(out, "")
				prevBlank = true
			}
			continue
		}

		out = append(out, strings.Join(strings.Fields(trimmed), " "))
		prevBlank = false
	}

	if len(out) == 0 {
		return []string{""}
	}
	return out
}
