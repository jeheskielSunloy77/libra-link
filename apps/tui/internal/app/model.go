package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/config"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/keymap"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/reader"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/session"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/repo"
	syncer "github.com/jeheskielSunloy77/libra-link/apps/tui/internal/sync"
)

type Model struct {
	cfg         *config.Config
	apiClient   *api.Client
	repo        *repo.Repository
	sessionFile *session.Store
	worker      *syncer.Worker
	syncCancel  context.CancelFunc

	keys keymap.KeyMap

	screen Screen

	width  int
	height int

	showHelp bool
	loggedIn bool
	user     *api.User

	authMode           authMode
	loginIDInput       textinput.Model
	loginPWInput       textinput.Model
	signupEmailInput   textinput.Model
	signupUserInput    textinput.Model
	signupPWInput      textinput.Model
	signupConfirmInput textinput.Model
	googleAuthURL      string
	googleCode         string
	googleExpires      time.Time
	googlePollEvery    time.Duration

	ebooks       []repo.EbookCache
	ebookIndex   int
	searchQuery  string
	searchInput  textinput.Model
	searchActive bool

	addActive    bool
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

	prefs      api.Preferences
	uiSettings UISettings

	palette        paletteState
	loading        LoadingState
	spinner        spinner.Model
	selectables    []Selectable
	focusIdx       int
	splashActive   bool
	splashProgress int
	splashTarget   int
	splashMessage  string
	splashReady    bool
	splashMinDone  bool

	status string
	errMsg string
}

func New(cfg *config.Config, apiClient *api.Client, store *repo.Repository, sessionStore *session.Store, worker *syncer.Worker) *Model {
	idInput := textinput.New()
	idInput.Placeholder = "email or username"
	idInput.Prompt = "identifier: "

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

	paletteInput := textinput.New()
	paletteInput.Placeholder = "Type a command..."
	paletteInput.Prompt = "⌘ "

	spin := spinner.New()
	spin.Spinner = spinner.Line

	syncCtx, cancel := context.WithCancel(context.Background())
	if worker != nil && cfg != nil {
		go worker.Run(syncCtx, cfg.SyncInterval)
	}

	m := &Model{
		cfg:                cfg,
		apiClient:          apiClient,
		repo:               store,
		sessionFile:        sessionStore,
		worker:             worker,
		syncCancel:         cancel,
		keys:               keymap.Default(),
		screen:             ScreenAuth,
		authMode:           authModeSignIn,
		loginIDInput:       idInput,
		loginPWInput:       pwInput,
		signupEmailInput:   signupEmailInput,
		signupUserInput:    signupUserInput,
		signupPWInput:      signupPWInput,
		signupConfirmInput: signupConfirmInput,
		searchInput:        searchInput,
		addSource:          addSourceInput,
		addTitle:           addTitleInput,
		addDesc:            addDescInput,
		addLanguage:        addLanguageInput,
		addFormat:          addFormatInput,
		googlePollEvery:    2 * time.Second,
		readingMode:        "normal",
		prefs: api.Preferences{
			ReadingMode:       "normal",
			ZenRestoreOnOpen:  true,
			ThemeMode:         "dark",
			ThemeOverrides:    map[string]string{},
			TypographyProfile: "comfortable",
		},
		uiSettings: UISettings{GutterPreset: "comfortable"},
		palette: paletteState{
			Input: paletteInput,
		},
		spinner:        spin,
		splashActive:   true,
		splashProgress: 0,
		splashTarget:   12,
		splashMessage:  "Initializing...",
	}
	m.rebuildFocus()
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.runBlocking("Restoring session...", m.bootstrapCmd()),
		m.loadUISettingsCmd(),
		m.splashAnimTickCmd(),
		tea.Tick(3*time.Second, func(time.Time) tea.Msg { return splashMinDurationMsg{} }),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		return m.finalize(nil)
	case tea.KeyMsg:
		return m.finalize(m.handleKey(typed))
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(typed)
		if m.loading.Active {
			return m.finalize(cmd)
		}
		return m.finalize(nil)
	case splashAnimTickMsg:
		if !m.splashActive {
			return m.finalize(nil)
		}
		if m.splashProgress < m.splashTarget {
			m.splashProgress = min(m.splashTarget, m.splashProgress+2)
		} else if !m.splashReady && m.splashProgress < 92 {
			m.splashProgress++
		}
		return m.finalize(m.splashAnimTickCmd())
	case splashMinDurationMsg:
		m.splashMinDone = true
		m.tryCloseSplash()
		return m.finalize(nil)
	case bootstrapMsg:
		m.endLoading()
		m.splashTarget = 60
		m.splashMessage = "Checking session..."
		if typed.err != nil {
			m.status = "No active session"
			m.errMsg = typed.err.Error()
			m.loggedIn = false
			m.screen = ScreenAuth
			m.markSplashReady("Ready")
			return m.finalize(nil)
		}
		if typed.user != nil {
			m.loggedIn = true
			m.user = typed.user
			m.screen = ScreenLibrary
			m.status = fmt.Sprintf("Welcome back, %s", typed.user.Username)
			m.errMsg = ""
			m.splashMessage = "Syncing library..."
			m.splashTarget = 82
			return m.finalize(tea.Batch(
				m.runBlocking("Loading library...", m.fetchEbooksCmd()),
				m.fetchSharesCmd(),
				m.fetchPrefsCmd(),
				m.fetchReaderStateCmd(),
			))
		}
		m.markSplashReady("Ready")
		return m.finalize(nil)
	case loginMsg:
		m.endLoading()
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Login failed"
			return m.finalize(nil)
		}
		m.loggedIn = true
		m.user = typed.user
		m.screen = ScreenLibrary
		m.status = fmt.Sprintf("Signed in as %s", typed.user.Username)
		m.errMsg = ""
		return m.finalize(tea.Batch(
			m.runBlocking("Loading library...", m.fetchEbooksCmd()),
			m.fetchSharesCmd(),
			m.fetchPrefsCmd(),
			m.fetchReaderStateCmd(),
		))
	case signupMsg:
		m.endLoading()
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Sign up failed"
			return m.finalize(nil)
		}
		m.loggedIn = true
		m.user = typed.user
		m.screen = ScreenLibrary
		m.status = fmt.Sprintf("Account created as %s", typed.user.Username)
		m.errMsg = ""
		return m.finalize(tea.Batch(
			m.runBlocking("Loading library...", m.fetchEbooksCmd()),
			m.fetchSharesCmd(),
			m.fetchPrefsCmd(),
			m.fetchReaderStateCmd(),
		))
	case googleStartMsg:
		if typed.err != nil {
			m.endLoading()
			m.errMsg = typed.err.Error()
			m.status = "Google auth start failed"
			return m.finalize(nil)
		}
		if typed.start == nil {
			m.endLoading()
			return m.finalize(nil)
		}
		m.googleCode = typed.start.DeviceCode
		m.googleAuthURL = typed.start.AuthURL
		m.googleExpires = typed.start.ExpiresAt
		if typed.start.IntervalSeconds > 0 {
			m.googlePollEvery = time.Duration(typed.start.IntervalSeconds) * time.Second
		}
		m.status = "Complete Google sign-in in browser. Waiting for approval..."
		m.errMsg = ""
		m.loading.Active = true
		m.loading.Message = "Waiting for Google approval..."
		return m.finalize(tea.Batch(m.pollGoogleDeviceCmd(), m.spinner.Tick))
	case googlePollTickMsg:
		if m.googleCode == "" {
			return m.finalize(nil)
		}
		return m.finalize(tea.Batch(m.pollGoogleDeviceCmd(), m.spinner.Tick))
	case googlePollMsg:
		if typed.err != nil {
			m.endLoading()
			m.errMsg = typed.err.Error()
			m.status = "Google auth failed"
			return m.finalize(nil)
		}
		if typed.result == nil {
			return m.finalize(nil)
		}
		switch strings.ToLower(typed.result.Status) {
		case "approved":
			m.endLoading()
			if typed.result.User == nil {
				m.errMsg = "Google auth approved without user payload"
				return m.finalize(nil)
			}
			m.loggedIn = true
			m.user = typed.result.User
			m.screen = ScreenLibrary
			m.status = fmt.Sprintf("Signed in as %s via Google", typed.result.User.Username)
			m.errMsg = ""
			m.googleCode = ""
			m.googleAuthURL = ""
			_ = m.persistSession(typed.result.User.ID)
			return m.finalize(tea.Batch(
				m.runBlocking("Loading library...", m.fetchEbooksCmd()),
				m.fetchSharesCmd(),
				m.fetchPrefsCmd(),
				m.fetchReaderStateCmd(),
			))
		case "expired":
			m.endLoading()
			m.errMsg = "Google device code expired. Use command palette to retry."
			m.status = "Google auth expired"
			m.googleCode = ""
			return m.finalize(nil)
		case "failed":
			m.endLoading()
			m.errMsg = "Google auth rejected. Use command palette to retry."
			m.status = "Google auth failed"
			m.googleCode = ""
			return m.finalize(nil)
		default:
			return m.finalize(tea.Batch(
				tea.Tick(m.googlePollEvery, func(time.Time) tea.Msg { return googlePollTickMsg{} }),
				m.spinner.Tick,
			))
		}
	case ebooksMsg:
		m.endLoading()
		if m.splashActive {
			m.markSplashReady("Ready")
		}
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Failed to load library"
			return m.finalize(nil)
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
		return m.finalize(nil)
	case sharesMsg:
		m.endLoading()
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Failed to load community"
			return m.finalize(nil)
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
		return m.finalize(nil)
	case prefsMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			return m.finalize(nil)
		}
		if typed.prefs != nil {
			m.prefs = *typed.prefs
			m.readingMode = typed.prefs.ReadingMode
		}
		return m.finalize(nil)
	case stateMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			return m.finalize(nil)
		}
		if typed.state != nil {
			if typed.state.ReadingMode != "" {
				m.readingMode = typed.state.ReadingMode
			}
		}
		return m.finalize(nil)
	case patchResultMsg:
		if typed.err != nil {
			if typed.source == "borrow" {
				m.endLoading()
			}
			m.errMsg = typed.err.Error()
			if typed.source == "borrow" {
				m.status = "Borrow failed"
			} else if typed.source == "prefs" {
				m.status = "Preference update queued for sync"
			}
			return m.finalize(nil)
		}
		if typed.source == "borrow" {
			m.endLoading()
			m.status = "Borrow request completed"
		}
		if typed.prefs != nil {
			m.prefs = *typed.prefs
		}
		if typed.source == "prefs" {
			m.status = "Updated"
		}
		m.errMsg = ""
		return m.finalize(nil)
	case addBookResultMsg:
		m.endLoading()
		if typed.err != nil {
			m.addConfirmDuplicate = false
			m.pendingAdd = nil
			m.errMsg = typed.err.Error()
			m.status = "Add book failed"
			return m.finalize(nil)
		}
		if typed.duplicate {
			m.pendingAdd = typed.prepared
			m.addConfirmDuplicate = true
			m.status = "Duplicate detected. Choose Import Anyway or Cancel."
			m.errMsg = ""
			return m.finalize(nil)
		}
		if typed.created != nil {
			title := fallback(typed.created.Title, "book")
			m.clearAddMode()
			m.status = "Book added: " + title
			m.errMsg = ""
			return m.finalize(m.runBlocking("Loading library...", m.fetchEbooksCmd()))
		}
		return m.finalize(nil)
	case uiSettingsMsg:
		m.splashTarget = 35
		m.splashMessage = "Loading preferences..."
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			return m.finalize(nil)
		}
		if typed.settings != nil && strings.TrimSpace(typed.settings.GutterPreset) != "" {
			m.uiSettings.GutterPreset = strings.TrimSpace(typed.settings.GutterPreset)
		}
		return m.finalize(nil)
	case uiSettingsSavedMsg:
		if typed.err != nil {
			m.errMsg = typed.err.Error()
			m.status = "Failed to save UI settings"
			return m.finalize(nil)
		}
		m.errMsg = ""
		m.status = "UI settings saved"
		return m.finalize(nil)
	}

	return m.finalize(nil)
}

func (m *Model) finalize(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.rebuildFocus()
	return m, cmd
}

func (m *Model) runBlocking(message string, cmd tea.Cmd) tea.Cmd {
	m.beginLoading(message)
	if cmd == nil {
		return m.spinner.Tick
	}
	return tea.Batch(cmd, m.spinner.Tick)
}

func (m *Model) beginLoading(message string) {
	m.loading.count++
	m.loading.Active = true
	if strings.TrimSpace(message) != "" {
		m.loading.Message = message
	}
}

func (m *Model) endLoading() {
	if m.loading.count > 0 {
		m.loading.count--
	}
	if m.loading.count > 0 {
		return
	}
	m.loading.Active = false
	m.loading.Message = ""
	m.loading.count = 0
}

func (m *Model) splashAnimTickCmd() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg { return splashAnimTickMsg{} })
}

func (m *Model) markSplashReady(message string) {
	m.splashReady = true
	m.splashTarget = 100
	m.splashMessage = message
	m.tryCloseSplash()
}

func (m *Model) tryCloseSplash() {
	if !m.splashReady || !m.splashMinDone {
		return
	}
	m.splashProgress = 100
	m.splashActive = false
}
