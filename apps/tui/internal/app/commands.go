package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/reader"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/session"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/repo"
)

func (m *Model) openBook(book repo.EbookCache) bool {
	if m.cfg == nil {
		m.errMsg = "config is not available"
		return false
	}

	path := strings.TrimSpace(book.FilePath)
	if path == "" {
		path = filepath.Join(m.cfg.DataDir, "books", book.ID+".txt")
	}

	doc, err := reader.Load(path)
	if err != nil {
		m.errMsg = err.Error()
		m.status = "Failed to open book"
		return false
	}
	m.document = doc
	m.readerLine = 0
	if m.readerState != nil && strings.TrimSpace(m.readerState.CurrentEbookID) == strings.TrimSpace(book.ID) {
		if line, ok := reader.DecodeLocation(doc, m.readerState.CurrentLocation); ok {
			m.readerLine = line
		}
	}
	m.errMsg = ""
	m.status = "Opened " + book.Title
	return true
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
		if m.apiClient == nil || m.sessionFile == nil || m.cfg == nil {
			return bootstrapMsg{}
		}

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
		if m.apiClient == nil || m.cfg == nil {
			return loginMsg{err: fmt.Errorf("api client is not available")}
		}

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
		if m.apiClient == nil || m.cfg == nil {
			return signupMsg{err: fmt.Errorf("api client is not available")}
		}

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
		if m.apiClient == nil || m.cfg == nil {
			return googleStartMsg{err: fmt.Errorf("api client is not available")}
		}

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
		if m.apiClient == nil || m.cfg == nil {
			return googlePollMsg{err: fmt.Errorf("api client is not available")}
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
		if m.repo == nil {
			return ebooksMsg{err: fmt.Errorf("repository is not available")}
		}
		if m.apiClient == nil || m.cfg == nil {
			local, err := m.repo.ListEbooks(context.Background(), m.searchQuery)
			return ebooksMsg{ebooks: local, err: err}
		}

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
		if m.repo == nil {
			return ebooksMsg{err: fmt.Errorf("repository is not available")}
		}
		items, err := m.repo.ListEbooks(context.Background(), query)
		return ebooksMsg{ebooks: items, err: err}
	}
}

func (m *Model) fetchSharesCmd() tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			return sharesMsg{err: fmt.Errorf("repository is not available")}
		}
		if m.apiClient == nil || m.cfg == nil {
			local, err := m.repo.ListShares(context.Background())
			return sharesMsg{shares: local, err: err}
		}

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
		if m.repo == nil {
			return prefsMsg{}
		}
		if m.apiClient == nil || m.cfg == nil {
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
			return prefsMsg{}
		}

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
		if m.repo == nil {
			return stateMsg{}
		}
		if m.apiClient == nil || m.cfg == nil {
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
			return stateMsg{}
		}

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

	if m.repo != nil {
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
	}

	return func() tea.Msg {
		if m.apiClient == nil || m.cfg == nil {
			return patchResultMsg{source: "prefs"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		patched, err := m.apiClient.PatchPreferences(ctx, local)
		if err != nil {
			if m.repo != nil {
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
			}
			return patchResultMsg{source: "prefs", err: err}
		}
		if m.repo != nil {
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
		}
		return patchResultMsg{source: "prefs", prefs: patched}
	}
}

func (m *Model) patchReaderStateCmd() tea.Cmd {
	location := reader.EncodeLocation(m.document, m.readerLine)
	state := api.ReaderState{
		UserID:          m.currentUserID(),
		ReadingMode:     m.readingMode,
		CurrentLocation: location,
		LastOpenedAt:    ptr(time.Now().UTC()),
	}

	if len(m.ebooks) > 0 && m.ebookIndex < len(m.ebooks) {
		state.CurrentEbookID = m.ebooks[m.ebookIndex].ID
	}

	if m.repo != nil {
		_ = m.repo.UpsertReaderState(context.Background(), repo.ReaderStateCache{
			UserID:          state.UserID,
			CurrentEbookID:  state.CurrentEbookID,
			CurrentLocation: state.CurrentLocation,
			ReadingMode:     state.ReadingMode,
			RowVersion:      1,
			LastOpenedAt:    state.LastOpenedAt,
			UpdatedAt:       time.Now().UTC(),
		})
	}

	return func() tea.Msg {
		if m.apiClient == nil || m.cfg == nil {
			return patchResultMsg{source: "reader_state"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()

		_, err := m.apiClient.PatchReaderState(ctx, state)
		if err != nil {
			if m.repo != nil {
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
			}
			return patchResultMsg{source: "reader_state", err: err}
		}
		return patchResultMsg{source: "reader_state"}
	}
}

func (m *Model) borrowShareCmd(shareID string) tea.Cmd {
	return func() tea.Msg {
		if m.apiClient == nil || m.cfg == nil {
			return patchResultMsg{source: "borrow", err: fmt.Errorf("api client is not available")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTPTimeout)
		defer cancel()
		if err := m.apiClient.BorrowShare(ctx, shareID); err != nil {
			return patchResultMsg{source: "borrow", err: err}
		}
		return patchResultMsg{source: "borrow"}
	}
}

func (m *Model) loadUISettingsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			return uiSettingsMsg{}
		}
		settings, err := m.repo.GetUISettings(context.Background())
		return uiSettingsMsg{settings: settings, err: err}
	}
}

func (m *Model) persistUISettingsCmd() tea.Cmd {
	preset := strings.TrimSpace(m.uiSettings.GutterPreset)
	if preset == "" {
		preset = "comfortable"
	}
	return func() tea.Msg {
		if m.repo == nil {
			return uiSettingsSavedMsg{}
		}
		err := m.repo.UpsertUISettings(context.Background(), repo.UISettings{
			GutterPreset: preset,
			UpdatedAt:    time.Now().UTC(),
		})
		return uiSettingsSavedMsg{err: err}
	}
}

func (m *Model) persistSession(userID string) error {
	if m.apiClient == nil {
		return nil
	}
	access, refresh, _ := m.apiClient.Session()
	state := &session.State{AccessToken: access, RefreshToken: refresh, UserID: userID}
	if m.sessionFile != nil {
		if err := m.sessionFile.Save(state); err != nil {
			return err
		}
	}
	if m.repo == nil {
		return nil
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
		if m.apiClient == nil {
			return ""
		}
		_, _, userID := m.apiClient.Session()
		return userID
	}
	return m.user.ID
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

func nextGutterPreset(current string) string {
	presets := []string{"none", "narrow", "comfortable", "wide"}
	for i, preset := range presets {
		if preset == current {
			return presets[(i+1)%len(presets)]
		}
	}
	return "comfortable"
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
