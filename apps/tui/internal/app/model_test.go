package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/config"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/repo"
)

func TestGlobalCtrlPTogglesPalette(t *testing.T) {
	m := newModelForTest()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got := updated.(*Model)
	if !got.palette.Active {
		t.Fatal("expected command palette to open")
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got = updated.(*Model)
	if got.palette.Active {
		t.Fatal("expected command palette to close")
	}
}

func TestGlobalCtrlHTogglesHelp(t *testing.T) {
	m := newModelForTest()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	got := updated.(*Model)
	if !got.showHelp {
		t.Fatal("expected help overlay to open")
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	got = updated.(*Model)
	if got.showHelp {
		t.Fatal("expected help overlay to close")
	}
}

func TestAuthFocusNavigationNormalized(t *testing.T) {
	m := newModelForTest()
	if got := m.focusedID(); got != "auth.field.identifier" {
		t.Fatalf("expected initial auth focus on identifier, got %q", got)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	got := updated.(*Model)
	if got.focusedID() != "auth.field.password" {
		t.Fatalf("expected focus on password, got %q", got.focusedID())
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyTab})
	got = updated.(*Model)
	if got.focusedID() != "auth.action.submit" {
		t.Fatalf("expected focus on submit button, got %q", got.focusedID())
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyUp})
	got = updated.(*Model)
	if got.focusedID() != "auth.field.password" {
		t.Fatalf("expected focus back on password, got %q", got.focusedID())
	}
}

func TestAuthEnterValidationMissingFields(t *testing.T) {
	m := newModelForTest()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(*Model)
	if got.errMsg != "identifier and password are required" {
		t.Fatalf("unexpected error message: %q", got.errMsg)
	}
}

func TestAuthSwitchModeViaButton(t *testing.T) {
	m := newModelForTest()
	m.focusByID("auth.action.switch_mode")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(*Model)
	if got.authMode != authModeSignUp {
		t.Fatalf("expected signup mode, got %q", got.authMode)
	}
}

func TestNoTabScreenSwitchInLibrary(t *testing.T) {
	m := newModelForTest()
	m.loggedIn = true
	m.screen = ScreenLibrary
	m.rebuildFocus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	got := updated.(*Model)
	if got.screen != ScreenLibrary {
		t.Fatalf("expected screen to stay library, got %q", got.screen)
	}
}

func TestLibraryRefreshSetsLoading(t *testing.T) {
	m := newModelForTest()
	m.loggedIn = true
	m.screen = ScreenLibrary
	m.rebuildFocus()
	m.focusByID("library.action.refresh")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(*Model)
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
	if !got.loading.Active {
		t.Fatal("expected loading state to be active")
	}
}

func TestCommandPaletteExecutesNavigation(t *testing.T) {
	m := newModelForTest()
	m.loggedIn = true
	m.screen = ScreenLibrary
	m.rebuildFocus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got := updated.(*Model)
	if !got.palette.Active {
		t.Fatal("expected palette to open")
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("settings")})
	got = updated.(*Model)
	if len(got.palette.Entries) == 0 {
		t.Fatal("expected filtered palette entries")
	}
	targetIdx := -1
	for i, entry := range got.palette.Entries {
		if entry.Command.ID == "nav.settings" {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		t.Fatal("expected nav.settings command in palette results")
	}
	got.palette.Index = targetIdx

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = updated.(*Model)
	if got.screen != ScreenSettings {
		t.Fatalf("expected settings screen, got %q", got.screen)
	}
	if got.palette.Active {
		t.Fatal("expected palette to close after command execution")
	}
}

func TestGutterDefaultAndCycle(t *testing.T) {
	m := newModelForTest()
	if m.uiSettings.GutterPreset != "comfortable" {
		t.Fatalf("expected default gutter preset comfortable, got %q", m.uiSettings.GutterPreset)
	}

	m.loggedIn = true
	m.screen = ScreenSettings
	m.rebuildFocus()
	m.focusByID("settings.action.gutter")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(*Model)
	if got.uiSettings.GutterPreset != "wide" {
		t.Fatalf("expected gutter preset wide after cycle, got %q", got.uiSettings.GutterPreset)
	}
}

func TestGoogleSingleKeyStartsAction(t *testing.T) {
	m := newModelForTest()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	got := updated.(*Model)
	if cmd == nil {
		t.Fatal("expected command for plain g key")
	}
	if !got.loading.Active {
		t.Fatal("expected loading state for plain g key")
	}
}

func TestPaletteIncludesCommandAndBookGroups(t *testing.T) {
	m := newModelForTest()
	m.loggedIn = true
	m.screen = ScreenLibrary
	m.ebooks = []repo.EbookCache{
		{ID: "b1", Title: "Clean Code", Format: "pdf", Author: "Robert C. Martin"},
	}
	m.rebuildFocus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got := updated.(*Model)
	if !got.palette.Active {
		t.Fatal("expected palette to open")
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("clean")})
	got = updated.(*Model)

	hasCommands := false
	hasBooks := false
	for _, entry := range got.palette.Entries {
		if entry.Command.Group == "commands" {
			hasCommands = true
		}
		if entry.Command.Group == "books" {
			hasBooks = true
		}
	}
	if !hasCommands || !hasBooks {
		t.Fatalf("expected both commands and books groups, got commands=%v books=%v", hasCommands, hasBooks)
	}
}

func TestPaletteDisabledCommandDoesNotExecute(t *testing.T) {
	m := newModelForTest()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got := updated.(*Model)
	if !got.palette.Active {
		t.Fatal("expected palette to open")
	}

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("borrow selected share")})
	got = updated.(*Model)
	if len(got.palette.Entries) == 0 {
		t.Fatal("expected at least one palette entry")
	}
	targetIdx := -1
	for i, entry := range got.palette.Entries {
		if entry.Command.ID == "community.borrow" {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		t.Fatal("expected community.borrow command in palette results")
	}
	got.palette.Index = targetIdx

	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = updated.(*Model)
	if got.screen != ScreenAuth {
		t.Fatalf("expected screen unchanged on disabled command, got %q", got.screen)
	}
	if got.status != "Command unavailable in current state" {
		t.Fatalf("unexpected status: %q", got.status)
	}
	if !got.palette.Active {
		t.Fatal("expected palette to remain open when command disabled")
	}
}

func TestSplashStaysVisibleUntilMinDuration(t *testing.T) {
	m := newModelForTest()
	if !m.splashActive {
		t.Fatal("expected splash to be active on startup")
	}

	updated, _ := m.Update(bootstrapMsg{})
	got := updated.(*Model)
	if !got.splashReady {
		t.Fatal("expected splash ready after bootstrap completion")
	}
	if !got.splashActive {
		t.Fatal("expected splash to remain active before minimum duration")
	}

	updated, _ = got.Update(splashMinDurationMsg{})
	got = updated.(*Model)
	if got.splashActive {
		t.Fatal("expected splash to close after minimum duration elapsed")
	}
}

func TestOpenBookUnreadablePDFFailsExplicitly(t *testing.T) {
	m := newModelForTest()
	source := filepath.Join(t.TempDir(), "broken.pdf")
	if err := os.WriteFile(source, []byte("not a pdf"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	ok := m.openBook(repo.EbookCache{ID: "b1", Title: "Broken PDF", FilePath: source, Format: "pdf"})
	if ok {
		t.Fatal("expected openBook to fail")
	}
	if m.status != "Failed to open book" {
		t.Fatalf("expected failed status, got %q", m.status)
	}
	if m.errMsg == "" {
		t.Fatal("expected error message when opening broken pdf")
	}
}

func TestOpenBookRestoresLocationForSameEbook(t *testing.T) {
	m := newModelForTest()
	source := filepath.Join(t.TempDir(), "sample.txt")
	if err := os.WriteFile(source, []byte("one\ntwo\nthree\nfour\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	m.readerState = &api.ReaderState{
		CurrentEbookID:  "book-1",
		CurrentLocation: "fmt=txt;line=2",
	}

	ok := m.openBook(repo.EbookCache{ID: "book-1", Title: "Sample", FilePath: source, Format: "txt"})
	if !ok {
		t.Fatalf("expected openBook to succeed: %s", m.errMsg)
	}
	if m.readerLine != 2 {
		t.Fatalf("expected restored readerLine=2, got %d", m.readerLine)
	}
}

func TestOpenBookDifferentEbookStartsAtTop(t *testing.T) {
	m := newModelForTest()
	source := filepath.Join(t.TempDir(), "sample.txt")
	if err := os.WriteFile(source, []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	m.readerState = &api.ReaderState{
		CurrentEbookID:  "other-book",
		CurrentLocation: "fmt=txt;line=2",
	}

	ok := m.openBook(repo.EbookCache{ID: "book-1", Title: "Sample", FilePath: source, Format: "txt"})
	if !ok {
		t.Fatalf("expected openBook to succeed: %s", m.errMsg)
	}
	if m.readerLine != 0 {
		t.Fatalf("expected readerLine=0 for different book, got %d", m.readerLine)
	}
}

func newModelForTest() *Model {
	cfg := &config.Config{
		HTTPTimeout:  time.Second,
		SyncInterval: time.Second,
		DataDir:      "/tmp",
	}
	m := New(cfg, nil, nil, nil, nil)
	m.rebuildFocus()
	return m
}
