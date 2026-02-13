package app

import (
	"fmt"
	"strconv"
	"strings"
)

func (m *Model) rebuildFocus() {
	prevID := ""
	if len(m.selectables) > 0 && m.focusIdx >= 0 && m.focusIdx < len(m.selectables) {
		prevID = m.selectables[m.focusIdx].ID
	}

	m.selectables = m.buildSelectables()
	if len(m.selectables) == 0 {
		m.focusIdx = 0
		m.blurAllInputs()
		return
	}

	if prevID != "" {
		for i, item := range m.selectables {
			if item.ID == prevID && !item.Disabled {
				m.focusIdx = i
				m.applyInputFocus()
				m.syncSelectionFromFocus()
				return
			}
		}
	}

	m.focusIdx = 0
	for i, item := range m.selectables {
		if !item.Disabled {
			m.focusIdx = i
			break
		}
	}
	m.applyInputFocus()
	m.syncSelectionFromFocus()
}

func (m *Model) buildSelectables() []Selectable {
	if m.palette.Active {
		return nil
	}
	if m.showHelp {
		return []Selectable{{ID: "help.close", Label: "Close Help"}}
	}
	if !m.loggedIn {
		return m.buildAuthSelectables()
	}

	switch m.screen {
	case ScreenLibrary:
		return m.buildLibrarySelectables()
	case ScreenReader:
		return m.buildReaderSelectables()
	case ScreenCommunity:
		return m.buildCommunitySelectables()
	case ScreenSettings:
		return m.buildSettingsSelectables()
	default:
		return m.buildAuthSelectables()
	}
}

func (m *Model) buildAuthSelectables() []Selectable {
	items := []Selectable{}
	if m.authMode == authModeSignUp {
		items = append(items,
			Selectable{ID: "auth.field.email", Label: "Email"},
			Selectable{ID: "auth.field.username", Label: "Username"},
			Selectable{ID: "auth.field.password", Label: "Password"},
			Selectable{ID: "auth.field.confirm", Label: "Confirm Password"},
		)
	} else {
		items = append(items,
			Selectable{ID: "auth.field.identifier", Label: "Identifier"},
			Selectable{ID: "auth.field.password", Label: "Password"},
		)
	}
	items = append(items,
		Selectable{ID: "auth.action.submit", Label: "Submit"},
		Selectable{ID: "auth.action.switch_mode", Label: "Switch Mode"},
		Selectable{ID: "auth.action.google", Label: "Sign in with Google"},
	)
	return items
}

func (m *Model) buildLibrarySelectables() []Selectable {
	if m.addActive {
		if m.addConfirmDuplicate {
			return []Selectable{
				{ID: "library.add.duplicate_yes", Label: "Import Anyway"},
				{ID: "library.add.duplicate_no", Label: "Cancel"},
			}
		}
		return []Selectable{
			{ID: "library.add.field.source", Label: "Source Path"},
			{ID: "library.add.field.title", Label: "Title"},
			{ID: "library.add.field.description", Label: "Description"},
			{ID: "library.add.field.language", Label: "Language"},
			{ID: "library.add.field.format", Label: "Format"},
			{ID: "library.add.submit", Label: "Add Book"},
			{ID: "library.add.cancel", Label: "Cancel"},
		}
	}

	if m.searchActive {
		return []Selectable{
			{ID: "library.search.field.query", Label: "Search Query"},
			{ID: "library.search.submit", Label: "Search"},
			{ID: "library.search.clear", Label: "Clear"},
		}
	}

	items := make([]Selectable, 0, len(m.ebooks)+4)
	for i, book := range m.ebooks {
		items = append(items, Selectable{
			ID:    fmt.Sprintf("library.book.%d", i),
			Label: fallback(book.Title, "Untitled"),
		})
	}
	items = append(items,
		Selectable{ID: "library.action.search", Label: "Search"},
		Selectable{ID: "library.action.refresh", Label: "Refresh Library"},
		Selectable{ID: "library.action.add", Label: "Add New Book"},
		Selectable{ID: "library.action.open", Label: "Open Selected", Disabled: len(m.ebooks) == 0},
	)
	return items
}

func (m *Model) buildReaderSelectables() []Selectable {
	return []Selectable{
		{ID: "reader.content", Label: "Reader Content"},
		{ID: "reader.action.toggle_mode", Label: "Toggle Reading Mode"},
	}
}

func (m *Model) buildCommunitySelectables() []Selectable {
	items := make([]Selectable, 0, len(m.shares)+2)
	for i, share := range m.shares {
		items = append(items, Selectable{
			ID:    fmt.Sprintf("community.share.%d", i),
			Label: fallback(share.Title, share.ID),
		})
	}
	items = append(items,
		Selectable{ID: "community.action.refresh", Label: "Refresh Community"},
		Selectable{ID: "community.action.borrow", Label: "Borrow Selected", Disabled: len(m.shares) == 0},
	)
	return items
}

func (m *Model) buildSettingsSelectables() []Selectable {
	return []Selectable{
		{ID: "settings.action.theme", Label: "Next Theme"},
		{ID: "settings.action.typography", Label: "Next Typography"},
		{ID: "settings.action.accent", Label: "Apply Accent Override"},
		{ID: "settings.action.clear_overrides", Label: "Clear Theme Overrides"},
		{ID: "settings.action.reading_mode", Label: "Toggle Reading Mode"},
		{ID: "settings.action.gutter", Label: "Cycle Gutter Preset"},
	}
}

func (m *Model) moveFocus(delta int) {
	if len(m.selectables) == 0 {
		return
	}
	if delta == 0 {
		return
	}
	idx := m.focusIdx
	for i := 0; i < len(m.selectables); i++ {
		idx = (idx + delta + len(m.selectables)) % len(m.selectables)
		if !m.selectables[idx].Disabled {
			m.focusIdx = idx
			m.applyInputFocus()
			m.syncSelectionFromFocus()
			return
		}
	}
}

func (m *Model) focusByID(id string) {
	for i, item := range m.selectables {
		if item.ID == id && !item.Disabled {
			m.focusIdx = i
			m.applyInputFocus()
			m.syncSelectionFromFocus()
			return
		}
	}
}

func (m *Model) focusedSelectable() (Selectable, bool) {
	if len(m.selectables) == 0 {
		return Selectable{}, false
	}
	if m.focusIdx < 0 || m.focusIdx >= len(m.selectables) {
		return Selectable{}, false
	}
	return m.selectables[m.focusIdx], true
}

func (m *Model) focusedID() string {
	item, ok := m.focusedSelectable()
	if !ok {
		return ""
	}
	return item.ID
}

func (m *Model) syncSelectionFromFocus() {
	id := m.focusedID()
	if strings.HasPrefix(id, "library.book.") {
		idx, err := strconv.Atoi(strings.TrimPrefix(id, "library.book."))
		if err == nil && idx >= 0 && idx < len(m.ebooks) {
			m.ebookIndex = idx
		}
	}
	if strings.HasPrefix(id, "community.share.") {
		idx, err := strconv.Atoi(strings.TrimPrefix(id, "community.share."))
		if err == nil && idx >= 0 && idx < len(m.shares) {
			m.shareIndex = idx
		}
	}
}

func (m *Model) applyInputFocus() {
	m.blurAllInputs()

	id := m.focusedID()
	switch id {
	case "auth.field.identifier":
		m.loginIDInput.Focus()
	case "auth.field.password":
		if m.authMode == authModeSignUp {
			m.signupPWInput.Focus()
		} else {
			m.loginPWInput.Focus()
		}
	case "auth.field.email":
		m.signupEmailInput.Focus()
	case "auth.field.username":
		m.signupUserInput.Focus()
	case "auth.field.confirm":
		m.signupConfirmInput.Focus()
	case "library.search.field.query":
		m.searchInput.Focus()
	case "library.add.field.source":
		m.addSource.Focus()
	case "library.add.field.title":
		m.addTitle.Focus()
	case "library.add.field.description":
		m.addDesc.Focus()
	case "library.add.field.language":
		m.addLanguage.Focus()
	case "library.add.field.format":
		m.addFormat.Focus()
	}
}

func (m *Model) blurAllInputs() {
	m.loginIDInput.Blur()
	m.loginPWInput.Blur()
	m.signupEmailInput.Blur()
	m.signupUserInput.Blur()
	m.signupPWInput.Blur()
	m.signupConfirmInput.Blur()
	m.searchInput.Blur()
	m.addSource.Blur()
	m.addTitle.Blur()
	m.addDesc.Blur()
	m.addLanguage.Blur()
	m.addFormat.Blur()
	m.palette.Input.Blur()
}
