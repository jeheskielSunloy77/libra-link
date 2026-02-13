package app

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/repo"
)

const (
	pageJumpSize = 20
)

var supportedBookFormats = map[string]struct{}{
	"txt":  {},
	"pdf":  {},
	"epub": {},
}

type Screen string

const (
	ScreenAuth      Screen = "auth"
	ScreenLibrary   Screen = "library"
	ScreenReader    Screen = "reader"
	ScreenCommunity Screen = "community"
	ScreenSettings  Screen = "settings"
)

type authMode string

const (
	authModeSignIn authMode = "sign_in"
	authModeSignUp authMode = "sign_up"
)

type UISettings struct {
	GutterPreset string
}

type ActionButton struct {
	ID       string
	Label    string
	Disabled bool
}

type Selectable struct {
	ID       string
	Label    string
	Disabled bool
}

type PaletteCommand struct {
	ID          string
	Group       string
	Icon        string
	Title       string
	Description string
}

type LoadingState struct {
	Active  bool
	Message string
	count   int
}

type paletteEntry struct {
	Command PaletteCommand
	Score   int
	Enabled bool
}

type paletteState struct {
	Active  bool
	Input   textinput.Model
	Entries []paletteEntry
	Index   int
}

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
	source string
	err    error
	prefs  *api.Preferences
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

type uiSettingsMsg struct {
	settings *repo.UISettings
	err      error
}

type uiSettingsSavedMsg struct {
	err error
}

type splashAnimTickMsg struct{}

type splashMinDurationMsg struct{}

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
