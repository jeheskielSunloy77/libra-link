package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/config"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/keymap"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/reader"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/session"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/repo"
	syncer "github.com/jeheskielSunloy77/libra-link/apps/tui/internal/sync"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/theme"
)

const (
	pageJumpSize = 20
)

var supportedBookFormats = map[string]struct{}{
	"txt":  {},
	"pdf":  {},
	"epub": {},
}

type authMode string

const (
	authModeSignIn authMode = "sign_in"
	authModeSignUp authMode = "sign_up"
)

type bootstrapMsg struct {
	user *api.User
	err  error
}

type loginMsg struct {
	user *api.User
	err  error
}

type signupMsg struct {
	user *api.User
	err  error
}

type ebooksMsg struct {
	ebooks []repo.EbookCache
	err    error
}

type sharesMsg struct {
	shares []repo.ShareCache
	err    error
}

type prefsMsg struct {
	prefs *api.Preferences
	err   error
}

type stateMsg struct {
	state *api.ReaderState
	err   error
}

type patchResultMsg struct {
	err   error
	prefs *api.Preferences
}

type googleStartMsg struct {
	start *api.GoogleDeviceStart
	err   error
}

type googlePollMsg struct {
	result *api.GoogleDevicePoll
	err    error
}

type googlePollTickMsg struct{}

type addBookPrepared struct {
	SourcePath   string
	Title        string
	Description  string
	LanguageCode string
	Format       string
	Checksum     string
	FileSize     int64
	ImportedAt   time.Time
	BaseDestPath string
	Ext          string
}

type addBookResultMsg struct {
	created   *api.Ebook
	err       error
	duplicate bool
	prepared  *addBookPrepared
}

type Model struct {
	cfg         *config.Config
	apiClient   *api.Client
	repo        *repo.Repository
	sessionFile *session.Store
	worker      *syncer.Worker
	syncCancel  context.CancelFunc

	keys keymap.KeyMap
	tabs []string
	view int

	width  int
	height int

	showHelp bool
	loggedIn bool
	user     *api.User

	loginIDInput       textinput.Model
	loginPWInput       textinput.Model
	signupEmailInput   textinput.Model
	signupUserInput    textinput.Model
	signupPWInput      textinput.Model
	signupConfirmInput textinput.Model
	loginFocusIdx      int
	authMode           authMode
	googleAuthURL      string
	googleCode         string
	googleExpires      time.Time
	googlePollEvery    time.Duration

	ebooks       []repo.EbookCache
	ebookIndex   int
	searchQuery  string
	searchActive bool
	searchInput  textinput.Model
	addActive    bool
	addFocusIdx  int
	addSource    textinput.Model
	addTitle     textinput.Model
	addDesc      textinput.Model
	addLanguage  textinput.Model
	addFormat    textinput.Model
	addFormatSet bool

	addConfirmDuplicate bool
	pendingAdd          *addBookPrepared

	shares     []repo.ShareCache
	shareIndex int

	document    *reader.Document
	readerLine  int
	readingMode string

	prefs api.Preferences

	status string
	errMsg string
}

func New(cfg *config.Config, apiClient *api.Client, store *repo.Repository, sessionStore *session.Store, worker *syncer.Worker) *Model {
	idInput := textinput.New()
	idInput.Placeholder = "email or username"
	idInput.Prompt = "identifier: "
	idInput.Focus()

	pwInput := textinput.New()
	pwInput.Placeholder = "password"
	pwInput.Prompt = "password: "
	pwInput.EchoMode = textinput.EchoPassword
	pwInput.EchoCharacter = '•'

	signupEmailInput := textinput.New()
	signupEmailInput.Placeholder = "email"
	signupEmailInput.Prompt = "email: "

	signupUserInput := textinput.New()
	signupUserInput.Placeholder = "username"
	signupUserInput.Prompt = "username: "

	signupPWInput := textinput.New()
	signupPWInput.Placeholder = "password"
	signupPWInput.Prompt = "password: "
	signupPWInput.EchoMode = textinput.EchoPassword
	signupPWInput.EchoCharacter = '•'

	signupConfirmInput := textinput.New()
	signupConfirmInput.Placeholder = "confirm password"
	signupConfirmInput.Prompt = "confirm: "
	signupConfirmInput.EchoMode = textinput.EchoPassword
	signupConfirmInput.EchoCharacter = '•'

	searchInput := textinput.New()
	searchInput.Placeholder = "search title/author"
	searchInput.Prompt = "search: "

	addSourceInput := textinput.New()
	addSourceInput.Placeholder = "/path/to/book.txt"
	addSourceInput.Prompt = "source: "

	addTitleInput := textinput.New()
	addTitleInput.Placeholder = "book title"
	addTitleInput.Prompt = "title: "

	addDescInput := textinput.New()
	addDescInput.Placeholder = "description (optional)"
	addDescInput.Prompt = "description: "

	addLanguageInput := textinput.New()
	addLanguageInput.Placeholder = "language code (optional)"
	addLanguageInput.Prompt = "language: "

	addFormatInput := textinput.New()
	addFormatInput.Placeholder = "txt|pdf|epub"
	addFormatInput.Prompt = "format: "
	addFormatInput.SetValue("txt")

	syncCtx, cancel := context.WithCancel(context.Background())
	if worker != nil {
		go worker.Run(syncCtx, cfg.SyncInterval)
	}

	return &Model{
		cfg:                cfg,
		apiClient:          apiClient,
		repo:               store,
		sessionFile:        sessionStore,
		worker:             worker,
		syncCancel:         cancel,
		keys:               keymap.Default(),
		tabs:               []string{"Login", "Library", "Reader", "Community", "Settings"},
		loginIDInput:       idInput,
		loginPWInput:       pwInput,
		signupEmailInput:   signupEmailInput,
		signupUserInput:    signupUserInput,
		signupPWInput:      signupPWInput,
		signupConfirmInput: signupConfirmInput,
		authMode:           authModeSignIn,
		searchInput:        searchInput,
		addSource:          addSourceInput,
		addTitle:           addTitleInput,
		addDesc:            addDescInput,
		addLanguage:        addLanguageInput,
		addFormat:          addFormatInput,
		readingMode:        "normal",
		googlePollEvery:    2 * time.Second,
		prefs: api.Preferences{
			ReadingMode:       "normal",
			ZenRestoreOnOpen:  true,
			ThemeMode:         "dark",
			ThemeOverrides:    map[string]string{},
			TypographyProfile: "comfortable",
		},
	}
}

func (m *Model) Init() tea.Cmd {
	return m.bootstrapCmd()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(typed)
	case bootstrapMsg:
		if typed.err != nil {
			m.status = "No active session"
			m.errMsg = typed.err.Error()
			m.loggedIn = false
			m.view = 0
			return m, nil
		}
		if typed.user != nil {
			m.loggedIn = true
			m.user = typed.user
			m.view = 1
			m.status = fmt.Sprintf("Welcome back, %s", typed.user.Username)
			m.errMsg = ""
			return m, tea.Batch(m.fetchEbooksCmd(), m.fetchSharesCmd(), m.fetchPrefsCmd(), m.fetchReaderStateCmd())
		}
		return m, nil
	case loginMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Login failed"
			return m, nil
		}
		m.loggedIn = true
		m.user = typed.user
		m.view = 1
		m.status = fmt.Sprintf("Signed in as %s", typed.user.Username)
		m.errMsg = ""
		return m, tea.Batch(m.fetchEbooksCmd(), m.fetchSharesCmd(), m.fetchPrefsCmd(), m.fetchReaderStateCmd())
	case signupMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Sign up failed"
			return m, nil
		}
		m.loggedIn = true
		m.user = typed.user
		m.view = 1
		m.status = fmt.Sprintf("Account created as %s", typed.user.Username)
		m.errMsg = ""
		return m, tea.Batch(m.fetchEbooksCmd(), m.fetchSharesCmd(), m.fetchPrefsCmd(), m.fetchReaderStateCmd())
	case googleStartMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Google auth start failed"
			return m, nil
		}
		if typed.start == nil {
			return m, nil
		}
		m.googleCode = typed.start.DeviceCode
		m.googleAuthURL = typed.start.AuthURL
		m.googleExpires = typed.start.ExpiresAt
		if typed.start.IntervalSeconds > 0 {
			m.googlePollEvery = time.Duration(typed.start.IntervalSeconds) * time.Second
		}
		m.status = "Complete Google sign-in in browser, then waiting for approval..."
		m.errMsg = ""
		return m, m.pollGoogleDeviceCmd()
	case googlePollTickMsg:
		if m.googleCode == "" {
			return m, nil
		}
		return m, m.pollGoogleDeviceCmd()
	case googlePollMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Google auth failed"
			return m, nil
		}
		if typed.result == nil {
			return m, nil
		}
		switch strings.ToLower(typed.result.Status) {
		case "approved":
			if typed.result.User == nil {
				m.errMsg = "Google auth approved without user payload"
				return m, nil
			}
			m.loggedIn = true
			m.user = typed.result.User
			m.view = 1
			m.status = fmt.Sprintf("Signed in as %s via Google", typed.result.User.Username)
			m.errMsg = ""
			m.googleCode = ""
			m.googleAuthURL = ""
			_ = m.persistSession(typed.result.User.ID)
			return m, tea.Batch(m.fetchEbooksCmd(), m.fetchSharesCmd(), m.fetchPrefsCmd(), m.fetchReaderStateCmd())
		case "expired":
			m.errMsg = "Google device code expired. Press g to try again."
			m.status = "Google auth expired"
			m.googleCode = ""
			return m, nil
		case "failed":
			m.errMsg = "Google auth rejected. Press g to retry."
			m.status = "Google auth failed"
			m.googleCode = ""
			return m, nil
		default:
			return m, tea.Tick(m.googlePollEvery, func(time.Time) tea.Msg { return googlePollTickMsg{} })
		}
	case ebooksMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Failed to load library"
			return m, nil
		}
		m.ebooks = typed.ebooks
		if m.ebookIndex >= len(m.ebooks) {
			m.ebookIndex = len(m.ebooks) - 1
		}
		if m.ebookIndex < 0 {
			m.ebookIndex = 0
		}
		m.status = fmt.Sprintf("Library synced: %d books", len(m.ebooks))
		m.errMsg = ""
		return m, nil
	case sharesMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Failed to load community"
			return m, nil
		}
		m.shares = typed.shares
		if m.shareIndex >= len(m.shares) {
			m.shareIndex = len(m.shares) - 1
		}
		if m.shareIndex < 0 {
			m.shareIndex = 0
		}
		m.status = fmt.Sprintf("Community synced: %d shares", len(m.shares))
		m.errMsg = ""
		return m, nil
	case prefsMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			return m, nil
		}
		if typed.prefs != nil {
			m.prefs = *typed.prefs
			m.readingMode = typed.prefs.ReadingMode
		}
		return m, nil
	case stateMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			return m, nil
		}
		if typed.state != nil {
			if typed.state.ReadingMode != "" {
				m.readingMode = typed.state.ReadingMode
			}
		}
		return m, nil
	case patchResultMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Remote update queued for sync"
		} else {
			if typed.prefs != nil {
				m.prefs = *typed.prefs
			}
			m.errMsg = ""
			m.status = "Updated"
		}
		return m, nil
	case addBookResultMsg:
		if typed.err != nil {
			m.addConfirmDuplicate = false
			m.pendingAdd = nil
			m.errMsg = typed.err.Error()
			m.status = "Add book failed"
			return m, nil
		}
		if typed.duplicate {
			m.pendingAdd = typed.prepared
			m.addConfirmDuplicate = true
			m.status = "Duplicate detected. Press y to import anyway, n to cancel."
			m.errMsg = ""
			return m, nil
		}
		if typed.created != nil {
			title := fallback(typed.created.Title, "book")
			m.clearAddMode()
			m.status = "Book added: " + title
			m.errMsg = ""
			return m, m.fetchEbooksCmd()
		}
		return m, nil
	}

	if m.view == 0 && !m.loggedIn {
		var cmd tea.Cmd
		switch m.authMode {
		case authModeSignUp:
			switch m.loginFocusIdx {
			case 0:
				m.signupEmailInput, cmd = m.signupEmailInput.Update(msg)
			case 1:
				m.signupUserInput, cmd = m.signupUserInput.Update(msg)
			case 2:
				m.signupPWInput, cmd = m.signupPWInput.Update(msg)
			default:
				m.signupConfirmInput, cmd = m.signupConfirmInput.Update(msg)
			}
		default:
			if m.loginFocusIdx == 0 {
				m.loginIDInput, cmd = m.loginIDInput.Update(msg)
			} else {
				m.loginPWInput, cmd = m.loginPWInput.Update(msg)
			}
		}
		return m, cmd
	}

	if m.searchActive {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	styles := m.styles()
	var body string

	switch m.view {
	case 0:
		body = m.renderLogin(styles)
	case 1:
		body = m.renderLibrary(styles)
	case 2:
		body = m.renderReader(styles)
	case 3:
		body = m.renderCommunity(styles)
	case 4:
		body = m.renderSettings(styles)
	default:
		body = "Unknown view"
	}

	header := m.renderHeader(styles)
	footer := m.renderFooter(styles)
	helper := ""
	if m.showHelp {
		helper = "\n" + styles.help.Render("keys: tab/shift+tab switch view | q quit | ? help | j/k move | z zen mode")
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer) + helper
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" || msg.String() == "q" {
		if m.syncCancel != nil {
			m.syncCancel()
		}
		return m, tea.Quit
	}

	if msg.String() == "?" {
		m.showHelp = !m.showHelp
		return m, nil
	}

	if !m.loggedIn {
		switch msg.String() {
		case "ctrl+n":
			m.toggleAuthMode()
			return m, nil
		case "tab", "shift+tab":
			fieldCount := 2
			if m.authMode == authModeSignUp {
				fieldCount = 4
			}
			if msg.String() == "shift+tab" {
				m.loginFocusIdx = (m.loginFocusIdx - 1 + fieldCount) % fieldCount
			} else {
				m.loginFocusIdx = (m.loginFocusIdx + 1) % fieldCount
			}
			m.applyAuthFocus()
			return m, nil
		case "enter":
			if m.authMode == authModeSignUp {
				email := strings.TrimSpace(m.signupEmailInput.Value())
				username := strings.TrimSpace(m.signupUserInput.Value())
				password := strings.TrimSpace(m.signupPWInput.Value())
				confirm := strings.TrimSpace(m.signupConfirmInput.Value())
				if email == "" || username == "" || password == "" || confirm == "" {
					m.errMsg = "email, username, password, and confirm password are required"
					return m, nil
				}
				if _, err := mail.ParseAddress(email); err != nil {
					m.errMsg = "valid email is required"
					return m, nil
				}
				if password != confirm {
					m.errMsg = "password and confirm password must match"
					return m, nil
				}
				return m, m.signupCmd(email, username, password)
			}
			identifier := strings.TrimSpace(m.loginIDInput.Value())
			password := strings.TrimSpace(m.loginPWInput.Value())
			if identifier == "" || password == "" {
				m.errMsg = "identifier and password are required"
				return m, nil
			}
			return m, m.loginCmd(identifier, password)
		case "g":
			return m, m.startGoogleDeviceCmd()
		}
		var cmd tea.Cmd
		switch m.authMode {
		case authModeSignUp:
			switch m.loginFocusIdx {
			case 0:
				m.signupEmailInput, cmd = m.signupEmailInput.Update(msg)
			case 1:
				m.signupUserInput, cmd = m.signupUserInput.Update(msg)
			case 2:
				m.signupPWInput, cmd = m.signupPWInput.Update(msg)
			default:
				m.signupConfirmInput, cmd = m.signupConfirmInput.Update(msg)
			}
		default:
			if m.loginFocusIdx == 0 {
				m.loginIDInput, cmd = m.loginIDInput.Update(msg)
			} else {
				m.loginPWInput, cmd = m.loginPWInput.Update(msg)
			}
		}
		return m, cmd
	}

	if msg.String() == "tab" && !(m.view == 1 && m.addActive) {
		m.view = (m.view + 1) % len(m.tabs)
		return m, nil
	}
	if msg.String() == "shift+tab" && !(m.view == 1 && m.addActive) {
		m.view = (m.view - 1 + len(m.tabs)) % len(m.tabs)
		return m, nil
	}

	switch m.view {
	case 1:
		return m.handleLibraryKeys(msg)
	case 2:
		return m.handleReaderKeys(msg)
	case 3:
		return m.handleCommunityKeys(msg)
	case 4:
		return m.handleSettingsKeys(msg)
	default:
		return m, nil
	}
}

func (m *Model) handleLibraryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addActive {
		return m.handleAddBookKeys(msg)
	}

	if m.searchActive {
		switch msg.String() {
		case "esc":
			m.searchActive = false
			m.searchInput.Blur()
			m.searchQuery = ""
			return m, m.loadLocalEbooksCmd("")
		case "ctrl+s", "enter":
			m.searchQuery = strings.TrimSpace(m.searchInput.Value())
			m.searchActive = false
			m.searchInput.Blur()
			return m, m.loadLocalEbooksCmd(m.searchQuery)
		}
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "j", "down":
		if m.ebookIndex < len(m.ebooks)-1 {
			m.ebookIndex++
		}
	case "k", "up":
		if m.ebookIndex > 0 {
			m.ebookIndex--
		}
	case "r":
		return m, m.fetchEbooksCmd()
	case "a":
		m.enterAddMode()
		return m, nil
	case "/":
		m.searchActive = true
		m.searchInput.Focus()
		return m, nil
	case "enter":
		if len(m.ebooks) == 0 {
			return m, nil
		}
		selected := m.ebooks[m.ebookIndex]
		m.view = 2
		m.openBook(selected)
		return m, m.patchReaderStateCmd()
	}
	return m, nil
}

func (m *Model) handleAddBookKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addConfirmDuplicate {
		switch msg.String() {
		case "y":
			return m, m.confirmDuplicateAddBookCmd()
		case "n", "esc":
			m.addConfirmDuplicate = false
			m.pendingAdd = nil
			m.status = "Duplicate import canceled"
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "esc":
		m.clearAddMode()
		m.status = "Add book canceled"
		m.errMsg = ""
		return m, nil
	case "tab":
		m.addFocusIdx = (m.addFocusIdx + 1) % 5
		m.applyAddFocus()
		return m, nil
	case "shift+tab":
		m.addFocusIdx = (m.addFocusIdx - 1 + 5) % 5
		m.applyAddFocus()
		return m, nil
	case "ctrl+s":
		return m, m.prepareAddBookCmd()
	}

	var cmd tea.Cmd
	switch m.addFocusIdx {
	case 0:
		m.addSource, cmd = m.addSource.Update(msg)
		if inferred := inferFormatFromPath(m.addSource.Value()); inferred != "" && !m.addFormatSet {
			m.addFormat.SetValue(inferred)
		}
	case 1:
		m.addTitle, cmd = m.addTitle.Update(msg)
	case 2:
		m.addDesc, cmd = m.addDesc.Update(msg)
	case 3:
		m.addLanguage, cmd = m.addLanguage.Update(msg)
	default:
		before := m.addFormat.Value()
		m.addFormat, cmd = m.addFormat.Update(msg)
		if strings.TrimSpace(m.addFormat.Value()) != strings.TrimSpace(before) {
			m.addFormatSet = true
		}
	}
	return m, cmd
}

func (m *Model) handleReaderKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.document == nil {
		return m, nil
	}

	switch msg.String() {
	case "j", "down":
		m.readerLine = reader.ClampLocation(m.document, m.readerLine+1)
		return m, m.patchReaderStateCmd()
	case "k", "up":
		m.readerLine = reader.ClampLocation(m.document, m.readerLine-1)
		return m, m.patchReaderStateCmd()
	case "h":
		m.readerLine = reader.ClampLocation(m.document, m.readerLine-pageJumpSize)
		return m, m.patchReaderStateCmd()
	case "l":
		m.readerLine = reader.ClampLocation(m.document, m.readerLine+pageJumpSize)
		return m, m.patchReaderStateCmd()
	case "g":
		m.readerLine = 0
		return m, m.patchReaderStateCmd()
	case "G":
		m.readerLine = reader.ClampLocation(m.document, len(m.document.Lines)-1)
		return m, m.patchReaderStateCmd()
	case "z":
		if m.readingMode == "normal" {
			m.readingMode = "zen"
		} else {
			m.readingMode = "normal"
		}
		m.prefs.ReadingMode = m.readingMode
		return m, tea.Batch(m.patchPrefsCmd(), m.patchReaderStateCmd())
	}
	return m, nil
}

func (m *Model) handleCommunityKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.shareIndex < len(m.shares)-1 {
			m.shareIndex++
		}
	case "k", "up":
		if m.shareIndex > 0 {
			m.shareIndex--
		}
	case "r":
		return m, m.fetchSharesCmd()
	case "b":
		if len(m.shares) == 0 {
			return m, nil
		}
		selected := m.shares[m.shareIndex]
		return m, m.borrowShareCmd(selected.ID)
	}
	return m, nil
}

func (m *Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "t":
		m.prefs.ThemeMode = nextThemeMode(m.prefs.ThemeMode)
		return m, m.patchPrefsCmd()
	case "p":
		m.prefs.TypographyProfile = nextTypographyProfile(m.prefs.TypographyProfile)
		return m, m.patchPrefsCmd()
	case "x":
		m.prefs.ThemeOverrides = map[string]string{}
		return m, m.patchPrefsCmd()
	case "o":
		if m.prefs.ThemeOverrides == nil {
			m.prefs.ThemeOverrides = map[string]string{}
		}
		m.prefs.ThemeOverrides["accent"] = "#ff7f50"
		return m, m.patchPrefsCmd()
	}
	return m, nil
}

func (m *Model) renderHeader(styles viewStyles) string {
	parts := make([]string, 0, len(m.tabs))
	for idx, tab := range m.tabs {
		if idx == m.view {
			parts = append(parts, styles.tabActive.Render(" "+tab+" "))
		} else {
			parts = append(parts, styles.tab.Render(" "+tab+" "))
		}
	}
	user := "guest"
	if m.user != nil {
		user = m.user.Username
	}
	line := lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(parts, " "), styles.user.Render(" user: "+user+" "))
	return styles.border.Render(line)
}

func (m *Model) renderFooter(styles viewStyles) string {
	status := m.status
	if status == "" {
		status = "Ready"
	}
	if m.errMsg != "" {
		status = status + " | error: " + m.errMsg
	}
	return styles.border.Render(styles.status.Render(status))
}

func (m *Model) renderLogin(styles viewStyles) string {
	googleLine := "press g to start Google OAuth device flow"
	if m.googleAuthURL != "" {
		googleLine = fmt.Sprintf("google url: %s", m.googleAuthURL)
	}

	deviceLine := ""
	if m.googleCode != "" {
		deviceLine = fmt.Sprintf("device code: %s (expires %s)", m.googleCode[:min(len(m.googleCode), 12)], m.googleExpires.Format(time.Kitchen))
	}

	body := []string{}
	switch m.authMode {
	case authModeSignUp:
		body = append(body,
			styles.sectionTitle.Render("Sign up"),
			m.signupEmailInput.View(),
			m.signupUserInput.View(),
			m.signupPWInput.View(),
			m.signupConfirmInput.View(),
			styles.subtle.Render("press tab to switch field, enter to create account, ctrl+n to switch to sign in"),
		)
	default:
		body = append(body,
			styles.sectionTitle.Render("Sign in"),
			m.loginIDInput.View(),
			m.loginPWInput.View(),
			styles.subtle.Render("press tab to switch field, enter to sign in, ctrl+n to switch to sign up"),
		)
	}
	body = append(body,
		styles.subtle.Render(googleLine),
		styles.subtle.Render(deviceLine),
	)
	return styles.panel.Render(strings.Join(body, "\n"))
}

func (m *Model) renderLibrary(styles viewStyles) string {
	if m.addActive {
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
			rows = append(rows, styles.subtle.Render("Duplicate detected. Press y to import anyway, n to cancel."))
		} else {
			rows = append(rows, styles.subtle.Render("ctrl+s submit  esc cancel  tab/shift+tab switch field"))
		}
		return styles.panel.Render(strings.Join(rows, "\n"))
	}

	if len(m.ebooks) == 0 {
		return styles.panel.Render("Library is empty. press a to add a book or r to sync from API.")
	}
	rows := make([]string, 0, len(m.ebooks)+2)
	rows = append(rows, styles.sectionTitle.Render("Library"))
	if m.searchActive {
		rows = append(rows, m.searchInput.View())
	}
	for idx, book := range m.ebooks {
		prefix := "  "
		rowStyle := styles.row
		if idx == m.ebookIndex {
			prefix = "> "
			rowStyle = styles.rowActive
		}
		rows = append(rows, rowStyle.Render(fmt.Sprintf("%s%s [%s]", prefix, book.Title, fallback(book.Format, "unknown"))))
	}
	rows = append(rows, styles.subtle.Render("enter=open  a=add  r=refresh  /=search  ctrl+s=apply search"))
	return styles.panel.Render(strings.Join(rows, "\n"))
}

func (m *Model) renderReader(styles viewStyles) string {
	if m.document == nil {
		return styles.panel.Render("No book open. choose one in Library and press enter.")
	}

	start := m.readerLine
	end := min(start+max(5, m.height-8), len(m.document.Lines))
	visible := m.document.Lines[start:end]

	contentStyle := styles.reader
	if m.readingMode == "zen" {
		contentStyle = styles.readerZen
	}
	content := contentStyle.Render(strings.Join(visible, "\n"))

	if m.readingMode == "zen" {
		return styles.panel.Render(content)
	}

	header := styles.sectionTitle.Render(m.document.Title)
	meta := styles.subtle.Render(fmt.Sprintf("line %d/%d | mode=%s", m.readerLine+1, len(m.document.Lines), m.readingMode))
	footer := styles.subtle.Render("j/k scroll  h/l page  g/G top/bottom  z zen")
	return styles.panel.Render(strings.Join([]string{header, meta, content, footer}, "\n"))
}

func (m *Model) renderCommunity(styles viewStyles) string {
	if len(m.shares) == 0 {
		return styles.panel.Render("No community shares cached. press r to sync.")
	}
	rows := make([]string, 0, len(m.shares)+2)
	rows = append(rows, styles.sectionTitle.Render("Community Shares"))
	for idx, share := range m.shares {
		prefix := "  "
		rowStyle := styles.row
		if idx == m.shareIndex {
			prefix = "> "
			rowStyle = styles.rowActive
		}
		title := fallback(share.Title, share.ID)
		rows = append(rows, rowStyle.Render(fmt.Sprintf("%s%s (%s)", prefix, title, share.Status)))
	}
	rows = append(rows, styles.subtle.Render("r refresh | b borrow selected"))
	return styles.panel.Render(strings.Join(rows, "\n"))
}

func (m *Model) renderSettings(styles viewStyles) string {
	baseTokens := theme.DefaultTokens(theme.Mode(m.prefs.ThemeMode))
	resolved, err := theme.ApplyOverrides(baseTokens, m.prefs.ThemeOverrides)
	validation := "theme overrides valid"
	if err != nil {
		validation = err.Error()
	}

	lines := []string{
		styles.sectionTitle.Render("Reader Settings"),
		fmt.Sprintf("readingMode: %s", m.prefs.ReadingMode),
		fmt.Sprintf("themeMode: %s", m.prefs.ThemeMode),
		fmt.Sprintf("typographyProfile: %s", m.prefs.TypographyProfile),
		fmt.Sprintf("zenRestoreOnOpen: %v", m.prefs.ZenRestoreOnOpen),
		fmt.Sprintf("tokens: bg=%s text=%s accent=%s progress=%s", resolved.Background, resolved.Text, resolved.Accent, resolved.Progress),
		styles.subtle.Render(validation),
		styles.subtle.Render("t cycle theme | p cycle typography | o sample accent override | x clear overrides"),
	}
	return styles.panel.Render(strings.Join(lines, "\n"))
}

func (m *Model) styles() viewStyles {
	mode := theme.Mode(m.prefs.ThemeMode)
	base := theme.DefaultTokens(mode)
	resolved, err := theme.ApplyOverrides(base, m.prefs.ThemeOverrides)
	if err != nil {
		resolved = base
	}

	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(resolved.Background)).
		Foreground(lipgloss.Color(resolved.Text))

	return viewStyles{
		border:       baseStyle.Copy().Padding(0, 1),
		panel:        baseStyle.Copy().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(resolved.Accent)).Padding(1),
		tab:          baseStyle.Copy().Foreground(lipgloss.Color(resolved.Text)),
		tabActive:    baseStyle.Copy().Bold(true).Foreground(lipgloss.Color(resolved.Accent)),
		sectionTitle: baseStyle.Copy().Bold(true).Foreground(lipgloss.Color(resolved.Accent)),
		user:         baseStyle.Copy().Foreground(lipgloss.Color(resolved.Progress)),
		status:       baseStyle.Copy(),
		help:         baseStyle.Copy().Foreground(lipgloss.Color(resolved.Progress)),
		subtle:       baseStyle.Copy().Foreground(lipgloss.Color(resolved.Progress)),
		row:          baseStyle.Copy(),
		rowActive:    baseStyle.Copy().Bold(true).Foreground(lipgloss.Color(resolved.Accent)),
		reader:       baseStyle.Copy().Padding(0, 1),
		readerZen:    baseStyle.Copy().Bold(true).Padding(0, 2),
	}
}

func (m *Model) openBook(book repo.EbookCache) {
	path := strings.TrimSpace(book.FilePath)
	if path == "" {
		path = filepath.Join(m.cfg.DataDir, "books", book.ID+".txt")
	}

	doc, err := reader.Load(path)
	if err != nil {
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte("This is a local placeholder text for "+book.Title+".\nAdd real files to read full content."), 0o644)
		doc, _ = reader.Load(path)
	}
	m.document = doc
	m.readerLine = 0
	m.status = "Opened " + book.Title
}

func (m *Model) enterAddMode() {
	m.addActive = true
	m.addConfirmDuplicate = false
	m.pendingAdd = nil
	m.searchActive = false
	m.searchInput.Blur()
	m.addFocusIdx = 0
	m.addFormatSet = false
	m.addSource.SetValue("")
	m.addTitle.SetValue("")
	m.addDesc.SetValue("")
	m.addLanguage.SetValue("")
	m.addFormat.SetValue("txt")
	m.applyAddFocus()
	m.errMsg = ""
	m.status = "Add new book"
}

func (m *Model) clearAddMode() {
	m.addActive = false
	m.addConfirmDuplicate = false
	m.pendingAdd = nil
	m.addFocusIdx = 0
	m.addFormatSet = false
	m.addSource.SetValue("")
	m.addTitle.SetValue("")
	m.addDesc.SetValue("")
	m.addLanguage.SetValue("")
	m.addFormat.SetValue("txt")
	m.addSource.Blur()
	m.addTitle.Blur()
	m.addDesc.Blur()
	m.addLanguage.Blur()
	m.addFormat.Blur()
}

func (m *Model) applyAddFocus() {
	m.addSource.Blur()
	m.addTitle.Blur()
	m.addDesc.Blur()
	m.addLanguage.Blur()
	m.addFormat.Blur()

	switch m.addFocusIdx {
	case 0:
		m.addSource.Focus()
	case 1:
		m.addTitle.Focus()
	case 2:
		m.addDesc.Focus()
	case 3:
		m.addLanguage.Focus()
	default:
		m.addFormat.Focus()
	}
}

func (m *Model) addStorageKeyPreview() string {
	if m.pendingAdd != nil && strings.TrimSpace(m.pendingAdd.BaseDestPath) != "" {
		return m.pendingAdd.BaseDestPath
	}
	source := strings.TrimSpace(m.addSource.Value())
	if source == "" {
		return "(computed on submit)"
	}

	ext := strings.ToLower(filepath.Ext(source))
	if _, ok := supportedBookFormats[strings.TrimPrefix(ext, ".")]; !ok {
		return "(unsupported source extension)"
	}

	dataDir := ""
	if m.cfg != nil {
		dataDir = m.cfg.DataDir
	}
	return filepath.Join(dataDir, "books", "<sha256>"+ext)
}

func (m *Model) prepareAddBookCmd() tea.Cmd {
	sourcePath := strings.TrimSpace(m.addSource.Value())
	title := strings.TrimSpace(m.addTitle.Value())
	description := strings.TrimSpace(m.addDesc.Value())
	languageCode := strings.TrimSpace(m.addLanguage.Value())
	format := strings.ToLower(strings.TrimSpace(m.addFormat.Value()))

	return func() tea.Msg {
		if sourcePath == "" {
			return addBookResultMsg{err: fmt.Errorf("source file path is required")}
		}
		if title == "" {
			return addBookResultMsg{err: fmt.Errorf("title is required")}
		}
		if format == "" {
			inferred := inferFormatFromPath(sourcePath)
			if inferred == "" {
				return addBookResultMsg{err: fmt.Errorf("format is required (txt, pdf, epub)")}
			}
			format = inferred
		}
		if _, ok := supportedBookFormats[format]; !ok {
			return addBookResultMsg{err: fmt.Errorf("unsupported format %q (allowed: txt, pdf, epub)", format)}
		}

		absSourcePath, err := filepath.Abs(sourcePath)
		if err != nil {
			return addBookResultMsg{err: err}
		}
		info, err := os.Stat(absSourcePath)
		if err != nil {
			return addBookResultMsg{err: err}
		}
		if !info.Mode().IsRegular() {
			return addBookResultMsg{err: fmt.Errorf("source path must be a regular file")}
		}

		ext := strings.ToLower(filepath.Ext(absSourcePath))
		if _, ok := supportedBookFormats[strings.TrimPrefix(ext, ".")]; !ok {
			return addBookResultMsg{err: fmt.Errorf("unsupported source extension %q (allowed: .txt, .pdf, .epub)", ext)}
		}

		checksum, fileSize, err := computeFileChecksum(absSourcePath)
		if err != nil {
			return addBookResultMsg{err: err}
		}

		dataDir := ""
		timeout := 15 * time.Second
		if m.cfg != nil {
			dataDir = m.cfg.DataDir
			if m.cfg.HTTPTimeout > 0 {
				timeout = m.cfg.HTTPTimeout
			}
		}
		if strings.TrimSpace(dataDir) == "" {
			return addBookResultMsg{err: fmt.Errorf("data directory is not configured")}
		}

		prepared := &addBookPrepared{
			SourcePath:   absSourcePath,
			Title:        title,
			Description:  description,
			LanguageCode: languageCode,
			Format:       format,
			Checksum:     checksum,
			FileSize:     fileSize,
			ImportedAt:   time.Now().UTC(),
			BaseDestPath: filepath.Join(dataDir, "books", checksum+ext),
			Ext:          ext,
		}

		if _, err := os.Stat(prepared.BaseDestPath); err == nil {
			return addBookResultMsg{duplicate: true, prepared: prepared}
		} else if err != nil && !os.IsNotExist(err) {
			return addBookResultMsg{err: err}
		}

		if m.apiClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			ebooks, err := m.apiClient.ListEbooks(ctx, 200)
			if err == nil {
				for _, item := range ebooks {
					if strings.EqualFold(strings.TrimSpace(item.ChecksumSHA256), checksum) {
						return addBookResultMsg{duplicate: true, prepared: prepared}
					}
				}
			}
		}

		return m.createBookFromPrepared(prepared, false)
	}
}

func (m *Model) confirmDuplicateAddBookCmd() tea.Cmd {
	if m.pendingAdd == nil {
		return nil
	}
	prepared := *m.pendingAdd
	return func() tea.Msg {
		return m.createBookFromPrepared(&prepared, true)
	}
}

func (m *Model) createBookFromPrepared(prepared *addBookPrepared, allowDuplicate bool) tea.Msg {
	if prepared == nil {
		return addBookResultMsg{err: fmt.Errorf("missing prepared book payload")}
	}
	if m.apiClient == nil {
		return addBookResultMsg{err: fmt.Errorf("api client is not available")}
	}

	destPath := prepared.BaseDestPath
	if allowDuplicate {
		destPath = filepath.Join(filepath.Dir(prepared.BaseDestPath), fmt.Sprintf("%s-%d%s", prepared.Checksum, time.Now().UTC().UnixNano(), prepared.Ext))
	}

	if err := copyFile(prepared.SourcePath, destPath); err != nil {
		return addBookResultMsg{err: err}
	}

	maxInt := int64(^uint(0) >> 1)
	if prepared.FileSize > maxInt {
		_ = os.Remove(destPath)
		return addBookResultMsg{err: fmt.Errorf("file is too large to import")}
	}

	timeout := 15 * time.Second
	if m.cfg != nil && m.cfg.HTTPTimeout > 0 {
		timeout = m.cfg.HTTPTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	importedAt := prepared.ImportedAt
	created, err := m.apiClient.CreateEbook(ctx, api.CreateEbookInput{
		Title:          prepared.Title,
		Description:    prepared.Description,
		Format:         prepared.Format,
		LanguageCode:   prepared.LanguageCode,
		StorageKey:     destPath,
		FileSizeBytes:  int(prepared.FileSize),
		ChecksumSHA256: prepared.Checksum,
		ImportedAt:     &importedAt,
	})
	if err != nil {
		_ = os.Remove(destPath)
		return addBookResultMsg{err: err}
	}
	return addBookResultMsg{created: created}
}

func computeFileChecksum(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), size, nil
}

func copyFile(sourcePath, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	return destination.Sync()
}

func (m *Model) bootstrapCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		sessionState, err := m.sessionFile.Load()
		if err != nil {
			return bootstrapMsg{err: err}
		}
		if sessionState == nil {
			return bootstrapMsg{}
		}

		m.apiClient.SetSession(sessionState.AccessToken, sessionState.RefreshToken, sessionState.UserID)
		user, err := m.apiClient.Refresh(ctx)
		if err != nil {
			return bootstrapMsg{err: err}
		}
		_ = m.persistSession(user.ID)
		return bootstrapMsg{user: user}
	}
}

func (m *Model) loginCmd(identifier, password string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		user, err := m.apiClient.Login(ctx, identifier, password)
		if err != nil {
			return loginMsg{err: err}
		}
		_ = m.persistSession(user.ID)
		return loginMsg{user: user}
	}
}

func (m *Model) signupCmd(email, username, password string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		user, err := m.apiClient.Register(ctx, email, username, password)
		if err != nil {
			return signupMsg{err: err}
		}
		_ = m.persistSession(user.ID)
		return signupMsg{user: user}
	}
}

func (m *Model) startGoogleDeviceCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		start, err := m.apiClient.StartGoogleDeviceAuth(ctx)
		if err != nil {
			return googleStartMsg{err: err}
		}

		go func(url string) {
			if strings.TrimSpace(url) == "" {
				return
			}
			_ = exec.Command("xdg-open", url).Start()
		}(start.AuthURL)

		return googleStartMsg{start: start}
	}
}

func (m *Model) pollGoogleDeviceCmd() tea.Cmd {
	code := strings.TrimSpace(m.googleCode)
	return func() tea.Msg {
		if code == "" {
			return googlePollMsg{}
		}
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		result, err := m.apiClient.PollGoogleDeviceAuth(ctx, code)
		if err != nil {
			return googlePollMsg{err: err}
		}
		return googlePollMsg{result: result}
	}
}

func (m *Model) fetchEbooksCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		remote, err := m.apiClient.ListEbooks(ctx, 50)
		if err != nil {
			local, localErr := m.repo.ListEbooks(context.Background(), m.searchQuery)
			if localErr != nil {
				return ebooksMsg{err: err}
			}
			return ebooksMsg{ebooks: local, err: nil}
		}

		_ = m.repo.UpsertEbooksFromRemote(context.Background(), remote)
		local, localErr := m.repo.ListEbooks(context.Background(), m.searchQuery)
		if localErr != nil {
			return ebooksMsg{err: localErr}
		}
		return ebooksMsg{ebooks: local}
	}
}

func (m *Model) loadLocalEbooksCmd(query string) tea.Cmd {
	return func() tea.Msg {
		items, err := m.repo.ListEbooks(context.Background(), query)
		return ebooksMsg{ebooks: items, err: err}
	}
}

func (m *Model) fetchSharesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		remote, err := m.apiClient.ListShares(ctx, 50)
		if err != nil {
			local, localErr := m.repo.ListShares(context.Background())
			if localErr != nil {
				return sharesMsg{err: err}
			}
			return sharesMsg{shares: local}
		}

		_ = m.repo.UpsertSharesFromRemote(context.Background(), remote)
		local, localErr := m.repo.ListShares(context.Background())
		if localErr != nil {
			return sharesMsg{err: localErr}
		}
		return sharesMsg{shares: local}
	}
}

func (m *Model) fetchPrefsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		prefs, err := m.apiClient.GetPreferences(ctx)
		if err != nil {
			cached, cacheErr := m.repo.GetPreferences(context.Background(), m.currentUserID())
			if cacheErr == nil && cached != nil {
				return prefsMsg{prefs: &api.Preferences{
					UserID:            cached.UserID,
					ReadingMode:       cached.ReadingMode,
					ZenRestoreOnOpen:  cached.ZenRestoreOnOpen,
					ThemeMode:         cached.ThemeMode,
					ThemeOverrides:    cached.ThemeOverrides,
					TypographyProfile: cached.TypographyProfile,
					RowVersion:        int(cached.RowVersion),
				}}
			}
			return prefsMsg{err: err}
		}

		_ = m.repo.UpsertPreferences(context.Background(), repo.PreferencesCache{
			UserID:            prefs.UserID,
			ReadingMode:       prefs.ReadingMode,
			ZenRestoreOnOpen:  prefs.ZenRestoreOnOpen,
			ThemeMode:         prefs.ThemeMode,
			ThemeOverrides:    prefs.ThemeOverrides,
			TypographyProfile: prefs.TypographyProfile,
			RowVersion:        int64(prefs.RowVersion),
			UpdatedAt:         time.Now().UTC(),
		})
		return prefsMsg{prefs: prefs}
	}
}

func (m *Model) fetchReaderStateCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()
		state, err := m.apiClient.GetReaderState(ctx)
		if err != nil {
			cached, cacheErr := m.repo.GetReaderState(context.Background(), m.currentUserID())
			if cacheErr == nil && cached != nil {
				local := &api.ReaderState{
					UserID:          cached.UserID,
					CurrentEbookID:  cached.CurrentEbookID,
					CurrentLocation: cached.CurrentLocation,
					ReadingMode:     cached.ReadingMode,
					RowVersion:      int(cached.RowVersion),
					LastOpenedAt:    cached.LastOpenedAt,
				}
				return stateMsg{state: local}
			}
			return stateMsg{err: err}
		}

		_ = m.repo.UpsertReaderState(context.Background(), repo.ReaderStateCache{
			UserID:          state.UserID,
			CurrentEbookID:  state.CurrentEbookID,
			CurrentLocation: state.CurrentLocation,
			ReadingMode:     state.ReadingMode,
			RowVersion:      int64(state.RowVersion),
			LastOpenedAt:    state.LastOpenedAt,
			UpdatedAt:       time.Now().UTC(),
		})
		return stateMsg{state: state}
	}
}

func (m *Model) patchPrefsCmd() tea.Cmd {
	local := m.prefs
	local.UserID = m.currentUserID()

	_ = m.repo.UpsertPreferences(context.Background(), repo.PreferencesCache{
		UserID:            local.UserID,
		ReadingMode:       local.ReadingMode,
		ZenRestoreOnOpen:  local.ZenRestoreOnOpen,
		ThemeMode:         local.ThemeMode,
		ThemeOverrides:    local.ThemeOverrides,
		TypographyProfile: local.TypographyProfile,
		RowVersion:        int64(max(1, local.RowVersion)),
		UpdatedAt:         time.Now().UTC(),
	})

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		patched, err := m.apiClient.PatchPreferences(ctx, local)
		if err != nil {
			_ = m.repo.EnqueueOutbox(context.Background(), repo.OutboxEvent{
				EntityType: "preference",
				EntityID:   local.UserID,
				Operation:  "upsert",
				Payload: map[string]any{
					"readingMode":       local.ReadingMode,
					"zenRestoreOnOpen":  local.ZenRestoreOnOpen,
					"themeMode":         local.ThemeMode,
					"themeOverrides":    local.ThemeOverrides,
					"typographyProfile": local.TypographyProfile,
				},
				IdempotencyKey: uuid.NewString(),
			})
			return patchResultMsg{err: err}
		}
		_ = m.repo.UpsertPreferences(context.Background(), repo.PreferencesCache{
			UserID:            patched.UserID,
			ReadingMode:       patched.ReadingMode,
			ZenRestoreOnOpen:  patched.ZenRestoreOnOpen,
			ThemeMode:         patched.ThemeMode,
			ThemeOverrides:    patched.ThemeOverrides,
			TypographyProfile: patched.TypographyProfile,
			RowVersion:        int64(patched.RowVersion),
			UpdatedAt:         time.Now().UTC(),
		})
		return patchResultMsg{prefs: patched}
	}
}

func (m *Model) patchReaderStateCmd() tea.Cmd {
	state := api.ReaderState{
		UserID:          m.currentUserID(),
		ReadingMode:     m.readingMode,
		CurrentLocation: fmt.Sprintf("line:%d", m.readerLine),
		LastOpenedAt:    ptr(time.Now().UTC()),
	}

	if len(m.ebooks) > 0 && m.ebookIndex < len(m.ebooks) {
		state.CurrentEbookID = m.ebooks[m.ebookIndex].ID
	}

	_ = m.repo.UpsertReaderState(context.Background(), repo.ReaderStateCache{
		UserID:          state.UserID,
		CurrentEbookID:  state.CurrentEbookID,
		CurrentLocation: state.CurrentLocation,
		ReadingMode:     state.ReadingMode,
		RowVersion:      1,
		LastOpenedAt:    state.LastOpenedAt,
		UpdatedAt:       time.Now().UTC(),
	})

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		_, err := m.apiClient.PatchReaderState(ctx, state)
		if err != nil {
			_ = m.repo.EnqueueOutbox(context.Background(), repo.OutboxEvent{
				EntityType: "reader_state",
				EntityID:   state.UserID,
				Operation:  "upsert",
				Payload: map[string]any{
					"currentEbookId":  state.CurrentEbookID,
					"currentLocation": state.CurrentLocation,
					"readingMode":     state.ReadingMode,
					"lastOpenedAt":    state.LastOpenedAt,
				},
				IdempotencyKey: uuid.NewString(),
			})
			return patchResultMsg{err: err}
		}
		return patchResultMsg{}
	}
}

func (m *Model) borrowShareCmd(shareID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()
		if err := m.apiClient.BorrowShare(ctx, shareID); err != nil {
			return patchResultMsg{err: err}
		}
		return patchResultMsg{}
	}
}

func (m *Model) persistSession(userID string) error {
	access, refresh, _ := m.apiClient.Session()
	state := &session.State{AccessToken: access, RefreshToken: refresh, UserID: userID}
	if err := m.sessionFile.Save(state); err != nil {
		return err
	}
	return m.repo.UpsertSessionState(context.Background(), repo.SessionState{
		AccessToken:  access,
		RefreshToken: refresh,
		UserID:       userID,
		UpdatedAt:    time.Now().UTC(),
	})
}

func (m *Model) currentUserID() string {
	if m.user == nil {
		_, _, userID := m.apiClient.Session()
		return userID
	}
	return m.user.ID
}

type viewStyles struct {
	border       lipgloss.Style
	panel        lipgloss.Style
	tab          lipgloss.Style
	tabActive    lipgloss.Style
	sectionTitle lipgloss.Style
	user         lipgloss.Style
	status       lipgloss.Style
	help         lipgloss.Style
	subtle       lipgloss.Style
	row          lipgloss.Style
	rowActive    lipgloss.Style
	reader       lipgloss.Style
	readerZen    lipgloss.Style
}

func nextThemeMode(current string) string {
	modes := []string{"light", "dark", "sepia", "high_contrast"}
	for i, mode := range modes {
		if mode == current {
			return modes[(i+1)%len(modes)]
		}
	}
	return modes[0]
}

func nextTypographyProfile(current string) string {
	profiles := []string{"compact", "comfortable", "large"}
	for i, profile := range profiles {
		if profile == current {
			return profiles[(i+1)%len(profiles)]
		}
	}
	return profiles[1]
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func inferFormatFromPath(path string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(path))), ".")
	if _, ok := supportedBookFormats[ext]; ok {
		return ext
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ptr[T any](value T) *T {
	return &value
}

func (m *Model) toggleAuthMode() {
	if m.authMode == authModeSignUp {
		m.authMode = authModeSignIn
	} else {
		m.authMode = authModeSignUp
	}
	m.loginFocusIdx = 0
	m.errMsg = ""
	m.status = ""
	m.applyAuthFocus()
}

func (m *Model) applyAuthFocus() {
	switch m.authMode {
	case authModeSignUp:
		m.signupEmailInput.Blur()
		m.signupUserInput.Blur()
		m.signupPWInput.Blur()
		m.signupConfirmInput.Blur()
		switch m.loginFocusIdx {
		case 0:
			m.signupEmailInput.Focus()
		case 1:
			m.signupUserInput.Focus()
		case 2:
			m.signupPWInput.Focus()
		default:
			m.signupConfirmInput.Focus()
		}
	default:
		m.loginIDInput.Blur()
		m.loginPWInput.Blur()
		if m.loginFocusIdx == 0 {
			m.loginIDInput.Focus()
			return
		}
		m.loginPWInput.Focus()
	}
}
