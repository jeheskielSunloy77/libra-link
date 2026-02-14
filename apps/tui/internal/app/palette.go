package app

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openPalette() {
	m.showHelp = false
	m.palette.Active = true
	m.palette.Input.SetValue("")
	m.palette.Index = 0
	m.palette.Input.Focus()
	m.filterPaletteCommands()
}

func (m *Model) closePalette() {
	m.palette.Active = false
	m.palette.Input.Blur()
	m.palette.Entries = nil
	m.palette.Index = 0
}

func (m *Model) handlePaletteKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	switch key {
	case "esc":
		m.closePalette()
		return nil
	case "down", "tab":
		if len(m.palette.Entries) > 0 {
			m.palette.Index = (m.palette.Index + 1) % len(m.palette.Entries)
		}
		return nil
	case "up", "shift+tab":
		if len(m.palette.Entries) > 0 {
			m.palette.Index = (m.palette.Index - 1 + len(m.palette.Entries)) % len(m.palette.Entries)
		}
		return nil
	case "enter":
		if len(m.palette.Entries) == 0 {
			return nil
		}
		entry := m.palette.Entries[m.palette.Index]
		if !entry.Enabled {
			m.status = "Command unavailable in current state"
			return nil
		}
		m.closePalette()
		return m.executePaletteCommand(entry.Command.ID)
	}

	var cmd tea.Cmd
	m.palette.Input, cmd = m.palette.Input.Update(msg)
	m.filterPaletteCommands()
	return cmd
}

func (m *Model) filterPaletteCommands() {
	query := strings.TrimSpace(m.palette.Input.Value())
	commands := m.paletteCommands()
	commandEntries := make([]paletteEntry, 0, len(commands))
	bookEntries := make([]paletteEntry, 0, len(m.ebooks))

	for _, command := range commands {
		target := command.Title + " " + command.Description + " " + command.ID
		score := fuzzyScore(query, target)
		if query != "" && score < 0 {
			continue
		}
		commandEntries = append(commandEntries, paletteEntry{
			Command: command,
			Score:   score,
			Enabled: m.isPaletteCommandEnabled(command.ID),
		})
	}

	for i, book := range m.ebooks {
		title := fallback(book.Title, "Untitled")
		desc := fmt.Sprintf("%s format • %s", fallback(book.Format, "unknown"), fallback(book.Author, "unknown author"))
		command := PaletteCommand{
			ID:          fmt.Sprintf("book.open.%d", i),
			Group:       "books",
			Icon:        "📘",
			Title:       title,
			Description: desc,
		}
		target := command.Title + " " + command.Description + " " + command.ID
		score := fuzzyScore(query, target)
		if query != "" && score < 0 {
			continue
		}
		bookEntries = append(bookEntries, paletteEntry{
			Command: command,
			Score:   score,
			Enabled: true,
		})
	}

	sort.SliceStable(commandEntries, func(i, j int) bool {
		if commandEntries[i].Score == commandEntries[j].Score {
			if commandEntries[i].Enabled != commandEntries[j].Enabled {
				return commandEntries[i].Enabled
			}
			return commandEntries[i].Command.Title < commandEntries[j].Command.Title
		}
		return commandEntries[i].Score > commandEntries[j].Score
	})
	sort.SliceStable(bookEntries, func(i, j int) bool {
		if bookEntries[i].Score == bookEntries[j].Score {
			return bookEntries[i].Command.Title < bookEntries[j].Command.Title
		}
		return bookEntries[i].Score > bookEntries[j].Score
	})

	entries := make([]paletteEntry, 0, len(commandEntries)+len(bookEntries))
	entries = append(entries, commandEntries...)
	entries = append(entries, bookEntries...)

	m.palette.Entries = entries
	if len(entries) == 0 {
		m.palette.Index = 0
		return
	}
	if m.palette.Index >= len(entries) {
		m.palette.Index = len(entries) - 1
	}
	if m.palette.Index < 0 {
		m.palette.Index = 0
	}
}

func (m *Model) paletteCommands() []PaletteCommand {
	return []PaletteCommand{
		{ID: "nav.auth", Group: "commands", Icon: "🔐", Title: "Go to Auth", Description: "Open sign in/sign up screen"},
		{ID: "nav.library", Group: "commands", Icon: "📚", Title: "Go to Library", Description: "Open your library"},
		{ID: "nav.reader", Group: "commands", Icon: "📖", Title: "Go to Reader", Description: "Open reader screen"},
		{ID: "nav.community", Group: "commands", Icon: "🌐", Title: "Go to Community", Description: "Open community shares"},
		{ID: "nav.settings", Group: "commands", Icon: "⚙", Title: "Go to Settings", Description: "Open settings screen"},
		{ID: "auth.submit", Group: "commands", Icon: "↵", Title: "Submit Auth Form", Description: "Submit sign in or sign up"},
		{ID: "auth.switch_mode", Group: "commands", Icon: "⇆", Title: "Switch Auth Mode", Description: "Switch sign in/sign up"},
		{ID: "auth.google", Group: "commands", Icon: "G", Title: "Start Google Sign-In", Description: "Begin device auth"},
		{ID: "library.refresh", Group: "commands", Icon: "⟳", Title: "Refresh Library", Description: "Sync library list"},
		{ID: "library.search", Group: "commands", Icon: "⌕", Title: "Open Library Search", Description: "Focus search form"},
		{ID: "library.search_apply", Group: "commands", Icon: "↵", Title: "Apply Search", Description: "Submit current search"},
		{ID: "library.search_clear", Group: "commands", Icon: "✕", Title: "Clear Search", Description: "Clear search filter"},
		{ID: "library.add", Group: "commands", Icon: "+", Title: "Add New Book", Description: "Open add book form"},
		{ID: "library.add_submit", Group: "commands", Icon: "↵", Title: "Submit Add Book", Description: "Import new book"},
		{ID: "library.add_cancel", Group: "commands", Icon: "✕", Title: "Cancel Add Book", Description: "Close add book form"},
		{ID: "library.open", Group: "commands", Icon: "→", Title: "Open Selected Book", Description: "Open highlighted book in reader"},
		{ID: "reader.toggle_mode", Group: "commands", Icon: "Z", Title: "Toggle Reading Mode", Description: "Switch normal and zen"},
		{ID: "community.refresh", Group: "commands", Icon: "⟳", Title: "Refresh Community", Description: "Sync shares"},
		{ID: "community.borrow", Group: "commands", Icon: "↓", Title: "Borrow Selected Share", Description: "Borrow highlighted share"},
		{ID: "settings.theme", Group: "commands", Icon: "T", Title: "Next Theme", Description: "Cycle theme mode"},
		{ID: "settings.typography", Group: "commands", Icon: "P", Title: "Next Typography", Description: "Cycle typography profile"},
		{ID: "settings.accent", Group: "commands", Icon: "O", Title: "Apply Accent Override", Description: "Set sample accent color"},
		{ID: "settings.clear_overrides", Group: "commands", Icon: "X", Title: "Clear Theme Overrides", Description: "Reset custom colors"},
		{ID: "settings.gutter", Group: "commands", Icon: "]", Title: "Cycle Gutter Preset", Description: "Change horizontal focus width"},
		{ID: "app.help", Group: "commands", Icon: "?", Title: "Toggle Help", Description: "Open keyboard help"},
		{ID: "app.quit", Group: "commands", Icon: "⎋", Title: "Quit Application", Description: "Exit TUI"},
	}
}

func (m *Model) isPaletteCommandEnabled(id string) bool {
	switch id {
	case "nav.auth", "app.help", "app.quit":
		return true
	case "auth.submit", "auth.switch_mode", "auth.google":
		return !m.loggedIn || m.screen == ScreenAuth
	case "nav.library", "nav.reader", "nav.community", "nav.settings",
		"library.refresh", "library.search", "library.add", "library.open",
		"library.search_apply", "library.search_clear", "library.add_submit", "library.add_cancel",
		"reader.toggle_mode", "community.refresh", "community.borrow",
		"settings.theme", "settings.typography", "settings.accent", "settings.clear_overrides", "settings.gutter":
		if !m.loggedIn {
			return false
		}
	}

	switch id {
	case "library.open":
		return len(m.ebooks) > 0
	case "library.search_apply":
		return m.searchActive
	case "library.search_clear":
		return m.searchActive || m.searchQuery != ""
	case "library.add_submit", "library.add_cancel":
		return m.addActive
	case "community.borrow":
		return len(m.shares) > 0
	case "reader.toggle_mode":
		return m.document != nil
	}

	return true
}

func (m *Model) executePaletteCommand(id string) tea.Cmd {
	switch id {
	case "nav.auth":
		m.screen = ScreenAuth
		m.status = "Opened Auth"
		return nil
	case "nav.library":
		if !m.loggedIn {
			return nil
		}
		m.screen = ScreenLibrary
		m.status = "Opened Library"
		return nil
	case "nav.reader":
		if !m.loggedIn {
			return nil
		}
		m.screen = ScreenReader
		m.status = "Opened Reader"
		return nil
	case "nav.community":
		if !m.loggedIn {
			return nil
		}
		m.screen = ScreenCommunity
		m.status = "Opened Community"
		return nil
	case "nav.settings":
		if !m.loggedIn {
			return nil
		}
		m.screen = ScreenSettings
		m.status = "Opened Settings"
		return nil
	case "auth.submit":
		return m.activateByID("auth.action.submit")
	case "auth.switch_mode":
		return m.activateByID("auth.action.switch_mode")
	case "auth.google":
		return m.activateByID("auth.action.google")
	case "library.refresh":
		m.screen = ScreenLibrary
		return m.activateByID("library.action.refresh")
	case "library.search":
		m.screen = ScreenLibrary
		return m.activateByID("library.action.search")
	case "library.search_apply":
		m.screen = ScreenLibrary
		return m.activateByID("library.search.submit")
	case "library.search_clear":
		m.screen = ScreenLibrary
		return m.activateByID("library.search.clear")
	case "library.add":
		m.screen = ScreenLibrary
		return m.activateByID("library.action.add")
	case "library.add_submit":
		m.screen = ScreenLibrary
		return m.activateByID("library.add.submit")
	case "library.add_cancel":
		m.screen = ScreenLibrary
		return m.activateByID("library.add.cancel")
	case "library.open":
		m.screen = ScreenLibrary
		return m.activateByID("library.action.open")
	case "reader.toggle_mode":
		m.screen = ScreenReader
		return m.activateByID("reader.action.toggle_mode")
	case "community.refresh":
		m.screen = ScreenCommunity
		return m.activateByID("community.action.refresh")
	case "community.borrow":
		m.screen = ScreenCommunity
		return m.activateByID("community.action.borrow")
	case "settings.theme":
		m.screen = ScreenSettings
		return m.activateByID("settings.action.theme")
	case "settings.typography":
		m.screen = ScreenSettings
		return m.activateByID("settings.action.typography")
	case "settings.accent":
		m.screen = ScreenSettings
		return m.activateByID("settings.action.accent")
	case "settings.clear_overrides":
		m.screen = ScreenSettings
		return m.activateByID("settings.action.clear_overrides")
	case "settings.gutter":
		m.screen = ScreenSettings
		return m.activateByID("settings.action.gutter")
	case "app.help":
		m.showHelp = !m.showHelp
		return nil
	case "app.quit":
		if m.syncCancel != nil {
			m.syncCancel()
		}
		return tea.Quit
	default:
		if strings.HasPrefix(id, "book.open.") {
			idx, err := strconv.Atoi(strings.TrimPrefix(id, "book.open."))
			if err != nil || idx < 0 || idx >= len(m.ebooks) {
				return nil
			}
			m.ebookIndex = idx
			if !m.openBook(m.ebooks[idx]) {
				return nil
			}
			m.screen = ScreenReader
			return m.patchReaderStateCmd()
		}
		return nil
	}
}

func fuzzyScore(query, target string) int {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return 0
	}
	target = strings.ToLower(target)
	qi := 0
	score := 0
	streak := 0

	for i := 0; i < len(target) && qi < len(query); i++ {
		if target[i] == query[qi] {
			score += 10
			if qi > 0 && i > 0 && query[qi-1] == target[i-1] {
				streak++
				score += streak * 3
			} else {
				streak = 0
			}
			qi++
		}
	}

	if qi < len(query) {
		return -1
	}
	return score - len(target)
}
