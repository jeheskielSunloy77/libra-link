package app

import (
	"context"
	"fmt"
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

type bootstrapMsg struct {
	user *api.User
	err  error
}

type loginMsg struct {
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

	loginIDInput    textinput.Model
	loginPWInput    textinput.Model
	loginFocusIdx   int
	googleAuthURL   string
	googleCode      string
	googleExpires   time.Time
	googlePollEvery time.Duration

	ebooks       []repo.EbookCache
	ebookIndex   int
	searchQuery  string
	searchActive bool
	searchInput  textinput.Model

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

	searchInput := textinput.New()
	searchInput.Placeholder = "search title/author"
	searchInput.Prompt = "search: "

	syncCtx, cancel := context.WithCancel(context.Background())
	if worker != nil {
		go worker.Run(syncCtx, cfg.SyncInterval)
	}

	return &Model{
		cfg:             cfg,
		apiClient:       apiClient,
		repo:            store,
		sessionFile:     sessionStore,
		worker:          worker,
		syncCancel:      cancel,
		keys:            keymap.Default(),
		tabs:            []string{"Login", "Library", "Reader", "Community", "Settings"},
		loginIDInput:    idInput,
		loginPWInput:    pwInput,
		searchInput:     searchInput,
		readingMode:     "normal",
		googlePollEvery: 2 * time.Second,
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
	}

	if m.view == 0 && !m.loggedIn {
		var cmd tea.Cmd
		if m.loginFocusIdx == 0 {
			m.loginIDInput, cmd = m.loginIDInput.Update(msg)
		} else {
			m.loginPWInput, cmd = m.loginPWInput.Update(msg)
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
		case "tab", "shift+tab":
			m.loginFocusIdx = (m.loginFocusIdx + 1) % 2
			if m.loginFocusIdx == 0 {
				m.loginIDInput.Focus()
				m.loginPWInput.Blur()
			} else {
				m.loginPWInput.Focus()
				m.loginIDInput.Blur()
			}
			return m, nil
		case "enter":
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
		return m, nil
	}

	if msg.String() == "tab" {
		m.view = (m.view + 1) % len(m.tabs)
		return m, nil
	}
	if msg.String() == "shift+tab" {
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

	body := []string{
		styles.sectionTitle.Render("Sign in"),
		m.loginIDInput.View(),
		m.loginPWInput.View(),
		styles.subtle.Render("press tab to switch field, enter to sign in"),
		styles.subtle.Render(googleLine),
		styles.subtle.Render(deviceLine),
	}
	return styles.panel.Render(strings.Join(body, "\n"))
}

func (m *Model) renderLibrary(styles viewStyles) string {
	if len(m.ebooks) == 0 {
		return styles.panel.Render("Library is empty. press r to sync from API.")
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
	rows = append(rows, styles.subtle.Render("enter=open  r=refresh  /=search  ctrl+s=apply search"))
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
