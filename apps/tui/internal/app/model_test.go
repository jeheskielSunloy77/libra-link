package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/config"
)

func TestAuthToggleCtrlN(t *testing.T) {
	m := newAuthModelForTest()
	m.errMsg = "stale error"
	m.status = "stale status"

	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	got := updated.(*Model)
	if got.authMode != authModeSignUp {
		t.Fatalf("expected signup mode, got %q", got.authMode)
	}
	if got.loginFocusIdx != 0 {
		t.Fatalf("expected focus index 0, got %d", got.loginFocusIdx)
	}
	if !got.signupEmailInput.Focused() {
		t.Fatal("expected signup email field to be focused")
	}
	if got.errMsg != "" || got.status != "" {
		t.Fatalf("expected status/error to be cleared, got status=%q err=%q", got.status, got.errMsg)
	}

	updated, _ = got.handleKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	got = updated.(*Model)
	if got.authMode != authModeSignIn {
		t.Fatalf("expected signin mode, got %q", got.authMode)
	}
	if !got.loginIDInput.Focused() {
		t.Fatal("expected login identifier field to be focused")
	}
}

func TestSignupTabAndShiftTabCycleFourFields(t *testing.T) {
	m := newAuthModelForTest()
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	got := updated.(*Model)

	for i := 0; i < 4; i++ {
		updated, _ = got.handleKey(tea.KeyMsg{Type: tea.KeyTab})
		got = updated.(*Model)
	}
	if got.loginFocusIdx != 0 {
		t.Fatalf("expected tab wrap to index 0, got %d", got.loginFocusIdx)
	}
	if !got.signupEmailInput.Focused() {
		t.Fatal("expected email field focused after wrap")
	}

	updated, _ = got.handleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	got = updated.(*Model)
	if got.loginFocusIdx != 3 {
		t.Fatalf("expected shift+tab wrap to index 3, got %d", got.loginFocusIdx)
	}
	if !got.signupConfirmInput.Focused() {
		t.Fatal("expected confirm field focused after shift+tab wrap")
	}
}

func TestSignupEnterValidationMissingFields(t *testing.T) {
	m := newAuthModelForTest()
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	got := updated.(*Model)

	updated, cmd := got.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	got = updated.(*Model)
	if cmd != nil {
		t.Fatal("expected nil command when signup fields are missing")
	}
	if got.errMsg != "email, username, password, and confirm password are required" {
		t.Fatalf("unexpected error message: %q", got.errMsg)
	}
}

func TestSignupEnterValidationInvalidEmail(t *testing.T) {
	m := newAuthModelForTest()
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	got := updated.(*Model)
	got.signupEmailInput.SetValue("not-an-email")
	got.signupUserInput.SetValue("jay")
	got.signupPWInput.SetValue("secret123")
	got.signupConfirmInput.SetValue("secret123")

	updated, cmd := got.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	got = updated.(*Model)
	if cmd != nil {
		t.Fatal("expected nil command when email is invalid")
	}
	if got.errMsg != "valid email is required" {
		t.Fatalf("unexpected error message: %q", got.errMsg)
	}
}

func TestSignupEnterValidationPasswordMismatch(t *testing.T) {
	m := newAuthModelForTest()
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlN})
	got := updated.(*Model)
	got.signupEmailInput.SetValue("jay@example.com")
	got.signupUserInput.SetValue("jay")
	got.signupPWInput.SetValue("secret123")
	got.signupConfirmInput.SetValue("different")

	updated, cmd := got.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	got = updated.(*Model)
	if cmd != nil {
		t.Fatal("expected nil command when passwords do not match")
	}
	if got.errMsg != "password and confirm password must match" {
		t.Fatalf("unexpected error message: %q", got.errMsg)
	}
}

func TestUpdateSignupMsgSuccessTransitionsToLoggedIn(t *testing.T) {
	m := newAuthModelForTest()

	updated, _ := m.Update(signupMsg{
		user: &api.User{
			ID:       "user-1",
			Username: "jay",
		},
	})
	got := updated.(*Model)
	if !got.loggedIn {
		t.Fatal("expected loggedIn to be true")
	}
	if got.view != 1 {
		t.Fatalf("expected view 1, got %d", got.view)
	}
	if got.user == nil || got.user.ID != "user-1" {
		t.Fatalf("expected user to be assigned, got %#v", got.user)
	}
	if got.status != "Account created as jay" {
		t.Fatalf("unexpected status: %q", got.status)
	}
	if got.errMsg != "" {
		t.Fatalf("expected empty errMsg, got %q", got.errMsg)
	}
}

func TestUpdateSignupMsgErrorSetsStatus(t *testing.T) {
	m := newAuthModelForTest()

	updated, _ := m.Update(signupMsg{err: errors.New("register failed")})
	got := updated.(*Model)
	if got.status != "Sign up failed" {
		t.Fatalf("unexpected status: %q", got.status)
	}
	if got.errMsg != "register failed" {
		t.Fatalf("unexpected errMsg: %q", got.errMsg)
	}
}

func TestSigninRegressionMissingFields(t *testing.T) {
	m := newAuthModelForTest()

	updated, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(*Model)
	if cmd != nil {
		t.Fatal("expected nil command when signin fields are missing")
	}
	if got.errMsg != "identifier and password are required" {
		t.Fatalf("unexpected error message: %q", got.errMsg)
	}
}

func TestSigninRegressionGoogleKeyReturnsCommand(t *testing.T) {
	m := newAuthModelForTest()

	_, cmd := m.handleKey(tea.KeyMsg{Runes: []rune{'g'}, Type: tea.KeyRunes})
	if cmd == nil {
		t.Fatal("expected google auth command to be returned")
	}
}

func newAuthModelForTest() *Model {
	idInput := textinput.New()
	idInput.Focus()
	pwInput := textinput.New()
	pwInput.EchoMode = textinput.EchoPassword
	pwInput.EchoCharacter = '•'

	signupEmail := textinput.New()
	signupUser := textinput.New()
	signupPW := textinput.New()
	signupPW.EchoMode = textinput.EchoPassword
	signupPW.EchoCharacter = '•'
	signupConfirm := textinput.New()
	signupConfirm.EchoMode = textinput.EchoPassword
	signupConfirm.EchoCharacter = '•'

	return &Model{
		cfg: &config.Config{
			HTTPTimeout: time.Second,
		},
		authMode:           authModeSignIn,
		loginIDInput:       idInput,
		loginPWInput:       pwInput,
		signupEmailInput:   signupEmail,
		signupUserInput:    signupUser,
		signupPWInput:      signupPW,
		signupConfirmInput: signupConfirm,
		loginFocusIdx:      0,
		googlePollEvery:    2 * time.Second,
		prefs:              api.Preferences{ThemeOverrides: map[string]string{}},
	}
}

func TestLibraryAddKeyEntersAddModeAndClearsErrors(t *testing.T) {
	m := newLibraryModelForTest(t)
	m.errMsg = "stale error"
	m.status = "stale status"

	updated, cmd := m.handleLibraryKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	got := updated.(*Model)
	if cmd != nil {
		t.Fatal("expected nil command when entering add mode")
	}
	if !got.addActive {
		t.Fatal("expected add mode to be active")
	}
	if !got.addSource.Focused() {
		t.Fatal("expected source input to be focused")
	}
	if got.errMsg != "" {
		t.Fatalf("expected cleared error message, got %q", got.errMsg)
	}
}

func TestLibraryAddSubmitValidationMissingFields(t *testing.T) {
	m := newLibraryModelForTest(t)
	m.enterAddMode()

	updated, cmd := m.handleLibraryKeys(tea.KeyMsg{Type: tea.KeyCtrlS})
	got := updated.(*Model)
	if cmd == nil {
		t.Fatal("expected submit command")
	}

	msg := cmd()
	updated, _ = got.Update(msg)
	got = updated.(*Model)
	if got.errMsg != "source file path is required" {
		t.Fatalf("unexpected validation error: %q", got.errMsg)
	}
	if !got.addActive {
		t.Fatal("expected add mode to stay active on validation failure")
	}
}

func TestLibraryAddSubmitRejectsUnsupportedFormat(t *testing.T) {
	m := newLibraryModelForTest(t)
	m.enterAddMode()

	sourcePath := filepath.Join(t.TempDir(), "book.txt")
	if err := os.WriteFile(sourcePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write temp source file: %v", err)
	}

	m.addSource.SetValue(sourcePath)
	m.addTitle.SetValue("A Book")
	m.addFormat.SetValue("docx")
	m.addFormatSet = true

	updated, cmd := m.handleLibraryKeys(tea.KeyMsg{Type: tea.KeyCtrlS})
	got := updated.(*Model)
	if cmd == nil {
		t.Fatal("expected submit command")
	}

	msg := cmd()
	updated, _ = got.Update(msg)
	got = updated.(*Model)
	if !strings.Contains(got.errMsg, "unsupported format") {
		t.Fatalf("unexpected validation error: %q", got.errMsg)
	}
}

func TestLibraryAddDuplicateDetectionEntersConfirmState(t *testing.T) {
	m := newLibraryModelForTest(t)
	m.enterAddMode()

	sourcePath := filepath.Join(t.TempDir(), "book.txt")
	content := []byte("duplicate content")
	if err := os.WriteFile(sourcePath, content, 0o644); err != nil {
		t.Fatalf("write temp source file: %v", err)
	}

	checksum, _, err := computeFileChecksum(sourcePath)
	if err != nil {
		t.Fatalf("compute checksum: %v", err)
	}
	destPath := filepath.Join(m.cfg.DataDir, "books", checksum+".txt")
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		t.Fatalf("mkdir books dir: %v", err)
	}
	if err := os.WriteFile(destPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write duplicate destination: %v", err)
	}

	m.addSource.SetValue(sourcePath)
	m.addTitle.SetValue("Duplicate Book")
	m.addFormat.SetValue("txt")
	m.addFormatSet = true

	updated, cmd := m.handleLibraryKeys(tea.KeyMsg{Type: tea.KeyCtrlS})
	got := updated.(*Model)
	if cmd == nil {
		t.Fatal("expected submit command")
	}

	msg := cmd()
	updated, _ = got.Update(msg)
	got = updated.(*Model)
	if !got.addConfirmDuplicate {
		t.Fatal("expected duplicate confirmation state")
	}
	if got.pendingAdd == nil {
		t.Fatal("expected pending add payload")
	}
	if got.status != "Duplicate detected. Press y to import anyway, n to cancel." {
		t.Fatalf("unexpected duplicate status: %q", got.status)
	}
}

func TestLibraryDuplicateConfirmNReturnsEditableForm(t *testing.T) {
	m := newLibraryModelForTest(t)
	m.addActive = true
	m.addConfirmDuplicate = true
	m.pendingAdd = &addBookPrepared{Title: "Book"}

	updated, cmd := m.handleLibraryKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := updated.(*Model)
	if cmd != nil {
		t.Fatal("expected nil command when canceling duplicate import")
	}
	if got.addConfirmDuplicate {
		t.Fatal("expected duplicate confirm state to be cleared")
	}
	if !got.addActive {
		t.Fatal("expected add form to remain active")
	}
}

func TestLibraryDuplicateConfirmYReturnsCommand(t *testing.T) {
	m := newLibraryModelForTest(t)
	m.addActive = true
	m.addConfirmDuplicate = true
	m.pendingAdd = &addBookPrepared{
		SourcePath:   filepath.Join(t.TempDir(), "book.txt"),
		Title:        "Book",
		Format:       "txt",
		Checksum:     strings.Repeat("a", 64),
		FileSize:     5,
		ImportedAt:   time.Now().UTC(),
		BaseDestPath: filepath.Join(m.cfg.DataDir, "books", strings.Repeat("a", 64)+".txt"),
		Ext:          ".txt",
	}

	updated, cmd := m.handleLibraryKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	got := updated.(*Model)
	if !got.addConfirmDuplicate {
		t.Fatal("expected duplicate confirm state to remain until command result")
	}
	if cmd == nil {
		t.Fatal("expected command to continue duplicate import")
	}
}

func TestLibraryEscCancelsAddMode(t *testing.T) {
	m := newLibraryModelForTest(t)
	m.enterAddMode()

	updated, cmd := m.handleLibraryKeys(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(*Model)
	if cmd != nil {
		t.Fatal("expected nil command when canceling add mode")
	}
	if got.addActive {
		t.Fatal("expected add mode to be closed")
	}
}

func newLibraryModelForTest(t *testing.T) *Model {
	t.Helper()

	cfg := &config.Config{
		DataDir:      t.TempDir(),
		HTTPTimeout:  time.Second,
		SyncInterval: time.Second,
	}
	m := New(cfg, nil, nil, nil, nil)
	m.loggedIn = true
	m.view = 1
	m.prefs.ThemeOverrides = map[string]string{}
	return m
}
