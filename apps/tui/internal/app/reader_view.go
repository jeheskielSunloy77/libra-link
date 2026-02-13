package app

import (
	"fmt"
	"strings"
)

func (m *Model) renderReader(styles viewStyles) string {
	if m.document == nil {
		rows := []string{
			styles.sectionTitle.Render("Reader"),
			"No book open.",
		}
		return styles.panel.Render(strings.Join(rows, "\n"))
	}

	start := m.readerLine
	available := max(5, m.height-12)
	end := min(start+available, len(m.document.Lines))
	visible := m.document.Lines[start:end]

	rows := []string{
		styles.sectionTitle.Render(m.document.Title),
		styles.subtle.Render(fmt.Sprintf("line %d/%d | mode=%s", m.readerLine+1, len(m.document.Lines), m.readingMode)),
		strings.Join(visible, "\n"),
	}
	return styles.panel.Render(strings.Join(rows, "\n"))
}
