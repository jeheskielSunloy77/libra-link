package app

import (
	"fmt"
	"strings"
)

func (m *Model) renderLibrary(styles viewStyles) string {
	if m.addActive {
		return m.renderLibraryAdd(styles)
	}
	if m.searchActive {
		return m.renderLibrarySearch(styles)
	}

	rows := []string{styles.sectionTitle.Render("Library")}
	focused := m.focusedID()
	if len(m.ebooks) == 0 {
		rows = append(rows, styles.subtle.Render("Library is empty."))
	} else {
		for i, book := range m.ebooks {
			id := fmt.Sprintf("library.book.%d", i)
			prefix := "  "
			style := styles.row
			if focused == id {
				prefix = "> "
				style = styles.rowActive
			}
			rows = append(rows, style.Render(fmt.Sprintf("%s%s [%s]", prefix, fallback(book.Title, "Untitled"), fallback(book.Format, "unknown"))))
		}
	}

	return styles.panel.Render(strings.Join(rows, "\n"))
}

func (m *Model) renderLibrarySearch(styles viewStyles) string {
	rows := []string{
		styles.sectionTitle.Render("Search Library"),
		m.searchInput.View(),
	}
	return styles.panel.Render(strings.Join(rows, "\n"))
}

func (m *Model) renderLibraryAdd(styles viewStyles) string {
	rows := []string{
		styles.sectionTitle.Render("Add Book"),
		m.addSource.View(),
		m.addTitle.View(),
		m.addDesc.View(),
		m.addLanguage.View(),
		m.addFormat.View(),
		styles.subtle.Render("storageKey: " + m.addStorageKeyPreview()),
	}

	if m.addConfirmDuplicate {
		rows = append(rows, styles.subtle.Render("Duplicate detected. Import anyway?"))
	}

	return styles.panel.Render(strings.Join(rows, "\n"))
}
