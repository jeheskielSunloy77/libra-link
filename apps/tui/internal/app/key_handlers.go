package app

import (
	"net/mail"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/reader"
)

func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if key == "ctrl+c" {
		if m.syncCancel != nil {
			m.syncCancel()
		}
		return tea.Quit
	}

	if key == "ctrl+p" {
		if m.palette.Active {
			m.closePalette()
		} else {
			m.openPalette()
		}
		return nil
	}

	if key == "ctrl+h" {
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.closePalette()
		}
		return nil
	}

	if m.palette.Active {
		return m.handlePaletteKey(msg)
	}

	if m.showHelp {
		if key == "esc" || key == "enter" {
			m.showHelp = false
		}
		return nil
	}

	if m.loading.Active {
		return nil
	}

	if !m.loggedIn {
		return m.handleAuthKeys(msg)
	}

	switch m.screen {
	case ScreenLibrary:
		return m.handleLibraryKeys(msg)
	case ScreenReader:
		return m.handleReaderKeys(msg)
	case ScreenCommunity:
		return m.handleCommunityKeys(msg)
	case ScreenSettings:
		return m.handleSettingsKeys(msg)
	default:
		return m.handleAuthKeys(msg)
	}
}

func (m *Model) handleAuthKeys(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	focusedID := m.focusedID()

	if key == "ctrl+n" {
		return m.activateByID("auth.action.switch_mode")
	}
	if key == "g" {
		return m.activateByID("auth.action.google")
	}

	if isNextKey(key) {
		m.moveFocus(1)
		return nil
	}
	if isPrevKey(key) {
		m.moveFocus(-1)
		return nil
	}
	if key == "enter" {
		if strings.HasPrefix(focusedID, "auth.field.") {
			return m.activateByID("auth.action.submit")
		}
		return m.activateByID(focusedID)
	}

	switch focusedID {
	case "auth.field.identifier":
		var cmd tea.Cmd
		m.loginIDInput, cmd = m.loginIDInput.Update(msg)
		return cmd
	case "auth.field.password":
		var cmd tea.Cmd
		if m.authMode == authModeSignUp {
			m.signupPWInput, cmd = m.signupPWInput.Update(msg)
		} else {
			m.loginPWInput, cmd = m.loginPWInput.Update(msg)
		}
		return cmd
	case "auth.field.email":
		var cmd tea.Cmd
		m.signupEmailInput, cmd = m.signupEmailInput.Update(msg)
		return cmd
	case "auth.field.username":
		var cmd tea.Cmd
		m.signupUserInput, cmd = m.signupUserInput.Update(msg)
		return cmd
	case "auth.field.confirm":
		var cmd tea.Cmd
		m.signupConfirmInput, cmd = m.signupConfirmInput.Update(msg)
		return cmd
	}

	return nil
}

func (m *Model) handleLibraryKeys(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	focusedID := m.focusedID()

	if m.addActive && m.addConfirmDuplicate {
		if key == "y" {
			return m.activateByID("library.add.duplicate_yes")
		}
		if key == "n" {
			return m.activateByID("library.add.duplicate_no")
		}
	}

	if key == "esc" {
		if m.addActive {
			m.clearAddMode()
			m.status = "Add book canceled"
			m.errMsg = ""
			return nil
		}
		if m.searchActive {
			m.searchActive = false
			m.searchInput.SetValue("")
			m.searchQuery = ""
			return m.runBlocking("Loading library...", m.loadLocalEbooksCmd(""))
		}
	}

	if !m.addActive && !m.searchActive {
		switch key {
		case "down":
			if len(m.ebooks) > 0 && m.ebookIndex < len(m.ebooks)-1 {
				m.ebookIndex++
			}
			m.focusByID("library.book." + strconv.Itoa(m.ebookIndex))
			return nil
		case "up":
			if len(m.ebooks) > 0 && m.ebookIndex > 0 {
				m.ebookIndex--
			}
			m.focusByID("library.book." + strconv.Itoa(m.ebookIndex))
			return nil
		case "a":
			return m.activateByID("library.action.add")
		case "ctrl+f":
			return m.activateByID("library.action.search")
		case "ctrl+r":
			return m.activateByID("library.action.refresh")
		}
	}
	if m.searchActive && key == "ctrl+s" {
		return m.activateByID("library.search.submit")
	}
	if m.addActive && !m.addConfirmDuplicate && key == "ctrl+s" {
		return m.activateByID("library.add.submit")
	}

	if isNextKey(key) {
		m.moveFocus(1)
		return nil
	}
	if isPrevKey(key) {
		m.moveFocus(-1)
		return nil
	}
	if key == "enter" {
		switch focusedID {
		case "library.search.field.query":
			return m.activateByID("library.search.submit")
		case "library.add.field.source", "library.add.field.title", "library.add.field.description", "library.add.field.language", "library.add.field.format":
			return m.activateByID("library.add.submit")
		default:
			return m.activateByID(focusedID)
		}
	}

	switch focusedID {
	case "library.search.field.query":
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return cmd
	case "library.add.field.source":
		var cmd tea.Cmd
		before := m.addSource.Value()
		m.addSource, cmd = m.addSource.Update(msg)
		if inferred := inferFormatFromPath(m.addSource.Value()); inferred != "" && !m.addFormatSet && m.addSource.Value() != before {
			m.addFormat.SetValue(inferred)
		}
		return cmd
	case "library.add.field.title":
		var cmd tea.Cmd
		m.addTitle, cmd = m.addTitle.Update(msg)
		return cmd
	case "library.add.field.description":
		var cmd tea.Cmd
		m.addDesc, cmd = m.addDesc.Update(msg)
		return cmd
	case "library.add.field.language":
		var cmd tea.Cmd
		m.addLanguage, cmd = m.addLanguage.Update(msg)
		return cmd
	case "library.add.field.format":
		var cmd tea.Cmd
		before := strings.TrimSpace(m.addFormat.Value())
		m.addFormat, cmd = m.addFormat.Update(msg)
		after := strings.TrimSpace(m.addFormat.Value())
		if before != after {
			m.addFormatSet = true
		}
		return cmd
	}

	return nil
}

func (m *Model) handleReaderKeys(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	focusedID := m.focusedID()

	if key == "z" {
		return m.activateByID("reader.action.toggle_mode")
	}

	if focusedID == "reader.content" {
		switch key {
		case "down":
			if m.document == nil {
				return nil
			}
			m.readerLine = reader.ClampLocation(m.document, m.readerLine+1)
			return m.patchReaderStateCmd()
		case "up":
			if m.document == nil {
				return nil
			}
			m.readerLine = reader.ClampLocation(m.document, m.readerLine-1)
			return m.patchReaderStateCmd()
		case "h":
			if m.document == nil {
				return nil
			}
			m.readerLine = reader.ClampLocation(m.document, m.readerLine-pageJumpSize)
			return m.patchReaderStateCmd()
		case "l":
			if m.document == nil {
				return nil
			}
			m.readerLine = reader.ClampLocation(m.document, m.readerLine+pageJumpSize)
			return m.patchReaderStateCmd()
		case "g":
			if m.document == nil {
				return nil
			}
			m.readerLine = 0
			return m.patchReaderStateCmd()
		case "G":
			if m.document == nil {
				return nil
			}
			m.readerLine = reader.ClampLocation(m.document, len(m.document.Lines)-1)
			return m.patchReaderStateCmd()
		case "tab":
			m.moveFocus(1)
			return nil
		case "shift+tab":
			m.moveFocus(-1)
			return nil
		case "enter":
			return nil
		}
	}

	if isNextKey(key) {
		m.moveFocus(1)
		return nil
	}
	if isPrevKey(key) {
		m.moveFocus(-1)
		return nil
	}
	if key == "enter" {
		return m.activateByID(focusedID)
	}
	return nil
}

func (m *Model) handleCommunityKeys(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	focusedID := m.focusedID()

	switch key {
	case "down":
		if len(m.shares) > 0 && m.shareIndex < len(m.shares)-1 {
			m.shareIndex++
		}
		m.focusByID("community.share." + strconv.Itoa(m.shareIndex))
		return nil
	case "up":
		if len(m.shares) > 0 && m.shareIndex > 0 {
			m.shareIndex--
		}
		m.focusByID("community.share." + strconv.Itoa(m.shareIndex))
		return nil
	case "r":
		return m.activateByID("community.action.refresh")
	case "b":
		return m.activateByID("community.action.borrow")
	}

	if isNextKey(key) {
		m.moveFocus(1)
		return nil
	}
	if isPrevKey(key) {
		m.moveFocus(-1)
		return nil
	}
	if key == "enter" {
		return m.activateByID(focusedID)
	}
	return nil
}

func (m *Model) handleSettingsKeys(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	focusedID := m.focusedID()

	switch key {
	case "t":
		return m.activateByID("settings.action.theme")
	case "p":
		return m.activateByID("settings.action.typography")
	case "o":
		return m.activateByID("settings.action.accent")
	case "x":
		return m.activateByID("settings.action.clear_overrides")
	case "]":
		return m.activateByID("settings.action.gutter")
	}

	if isNextKey(key) {
		m.moveFocus(1)
		return nil
	}
	if isPrevKey(key) {
		m.moveFocus(-1)
		return nil
	}
	if key == "enter" {
		return m.activateByID(focusedID)
	}
	return nil
}

func (m *Model) activateByID(id string) tea.Cmd {
	if id == "" {
		return nil
	}
	for _, item := range m.selectables {
		if item.ID == id && item.Disabled {
			return nil
		}
	}

	switch id {
	case "help.close":
		m.showHelp = false
		return nil
	case "auth.action.submit":
		if m.authMode == authModeSignUp {
			email := strings.TrimSpace(m.signupEmailInput.Value())
			username := strings.TrimSpace(m.signupUserInput.Value())
			password := strings.TrimSpace(m.signupPWInput.Value())
			confirm := strings.TrimSpace(m.signupConfirmInput.Value())
			if email == "" || username == "" || password == "" || confirm == "" {
				m.errMsg = "email, username, password, and confirm password are required"
				return nil
			}
			if _, err := mail.ParseAddress(email); err != nil {
				m.errMsg = "valid email is required"
				return nil
			}
			if password != confirm {
				m.errMsg = "password and confirm password must match"
				return nil
			}
			m.errMsg = ""
			return m.runBlocking("Creating account...", m.signupCmd(email, username, password))
		}

		identifier := strings.TrimSpace(m.loginIDInput.Value())
		password := strings.TrimSpace(m.loginPWInput.Value())
		if identifier == "" || password == "" {
			m.errMsg = "identifier and password are required"
			return nil
		}
		m.errMsg = ""
		return m.runBlocking("Signing in...", m.loginCmd(identifier, password))
	case "auth.action.switch_mode":
		m.toggleAuthMode()
		return nil
	case "auth.action.google":
		return m.runBlocking("Starting Google sign-in...", m.startGoogleDeviceCmd())
	case "library.action.search":
		m.searchActive = true
		m.addActive = false
		m.searchInput.SetValue(m.searchQuery)
		return nil
	case "library.action.refresh":
		return m.runBlocking("Loading library...", m.fetchEbooksCmd())
	case "library.action.add":
		m.enterAddMode()
		return nil
	case "library.action.open":
		if len(m.ebooks) == 0 || m.ebookIndex < 0 || m.ebookIndex >= len(m.ebooks) {
			return nil
		}
		m.screen = ScreenReader
		m.openBook(m.ebooks[m.ebookIndex])
		return m.patchReaderStateCmd()
	case "library.search.submit":
		m.searchQuery = strings.TrimSpace(m.searchInput.Value())
		m.searchActive = false
		return m.runBlocking("Searching library...", m.loadLocalEbooksCmd(m.searchQuery))
	case "library.search.clear":
		m.searchQuery = ""
		m.searchInput.SetValue("")
		m.searchActive = false
		return m.runBlocking("Loading library...", m.loadLocalEbooksCmd(""))
	case "library.add.submit":
		return m.runBlocking("Importing book...", m.prepareAddBookCmd())
	case "library.add.cancel":
		m.clearAddMode()
		m.status = "Add book canceled"
		m.errMsg = ""
		return nil
	case "library.add.duplicate_yes":
		return m.runBlocking("Importing duplicate...", m.confirmDuplicateAddBookCmd())
	case "library.add.duplicate_no":
		m.addConfirmDuplicate = false
		m.pendingAdd = nil
		m.status = "Duplicate import canceled"
		return nil
	case "reader.action.toggle_mode":
		if m.readingMode == "normal" {
			m.readingMode = "zen"
		} else {
			m.readingMode = "normal"
		}
		m.prefs.ReadingMode = m.readingMode
		return tea.Batch(m.patchPrefsCmd(), m.patchReaderStateCmd())
	case "community.action.refresh":
		return m.runBlocking("Loading community...", m.fetchSharesCmd())
	case "community.action.borrow":
		if len(m.shares) == 0 || m.shareIndex < 0 || m.shareIndex >= len(m.shares) {
			return nil
		}
		selected := m.shares[m.shareIndex]
		return m.runBlocking("Borrowing share...", m.borrowShareCmd(selected.ID))
	case "settings.action.theme":
		m.prefs.ThemeMode = nextThemeMode(m.prefs.ThemeMode)
		return m.patchPrefsCmd()
	case "settings.action.typography":
		m.prefs.TypographyProfile = nextTypographyProfile(m.prefs.TypographyProfile)
		return m.patchPrefsCmd()
	case "settings.action.accent":
		if m.prefs.ThemeOverrides == nil {
			m.prefs.ThemeOverrides = map[string]string{}
		}
		m.prefs.ThemeOverrides["accent"] = "#ff7f50"
		return m.patchPrefsCmd()
	case "settings.action.clear_overrides":
		m.prefs.ThemeOverrides = map[string]string{}
		return m.patchPrefsCmd()
	case "settings.action.reading_mode":
		if m.prefs.ReadingMode == "normal" {
			m.prefs.ReadingMode = "zen"
		} else {
			m.prefs.ReadingMode = "normal"
		}
		m.readingMode = m.prefs.ReadingMode
		return tea.Batch(m.patchPrefsCmd(), m.patchReaderStateCmd())
	case "settings.action.gutter":
		m.uiSettings.GutterPreset = nextGutterPreset(m.uiSettings.GutterPreset)
		return m.persistUISettingsCmd()
	}

	if strings.HasPrefix(id, "library.book.") {
		idx, err := strconv.Atoi(strings.TrimPrefix(id, "library.book."))
		if err != nil || idx < 0 || idx >= len(m.ebooks) {
			return nil
		}
		m.ebookIndex = idx
		m.screen = ScreenReader
		m.openBook(m.ebooks[idx])
		return m.patchReaderStateCmd()
	}

	if strings.HasPrefix(id, "community.share.") {
		idx, err := strconv.Atoi(strings.TrimPrefix(id, "community.share."))
		if err != nil || idx < 0 || idx >= len(m.shares) {
			return nil
		}
		m.shareIndex = idx
		m.status = "Selected share: " + fallback(m.shares[idx].Title, m.shares[idx].ID)
		return nil
	}

	return nil
}

func (m *Model) toggleAuthMode() {
	if m.authMode == authModeSignUp {
		m.authMode = authModeSignIn
	} else {
		m.authMode = authModeSignUp
	}
	m.errMsg = ""
	m.status = ""
}

func (m *Model) enterAddMode() {
	m.addActive = true
	m.addConfirmDuplicate = false
	m.pendingAdd = nil
	m.searchActive = false
	m.searchInput.SetValue("")
	m.addFormatSet = false
	m.addSource.SetValue("")
	m.addTitle.SetValue("")
	m.addDesc.SetValue("")
	m.addLanguage.SetValue("")
	m.addFormat.SetValue("txt")
	m.errMsg = ""
	m.status = "Add new book"
}

func (m *Model) clearAddMode() {
	m.addActive = false
	m.addConfirmDuplicate = false
	m.pendingAdd = nil
	m.addFormatSet = false
	m.addSource.SetValue("")
	m.addTitle.SetValue("")
	m.addDesc.SetValue("")
	m.addLanguage.SetValue("")
	m.addFormat.SetValue("txt")
}

func isNextKey(key string) bool {
	switch key {
	case "down", "tab":
		return true
	default:
		return false
	}
}

func isPrevKey(key string) bool {
	switch key {
	case "up", "shift+tab":
		return true
	default:
		return false
	}
}
