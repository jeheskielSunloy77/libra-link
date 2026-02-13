package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	styles := m.styles()
	contentWidth := m.contentWidth()
	bodyHeight := m.height - 2
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	body := m.renderScreen(styles, contentWidth)

	if m.splashActive {
		body = m.renderSplash(styles, contentWidth, bodyHeight)
	}
	if m.showHelp {
		body = m.renderHelpOverlay(styles, contentWidth)
	}
	if m.palette.Active {
		body = m.renderPaletteOverlay(styles, contentWidth)
	}
	if m.loading.Active {
		body = m.renderLoadingOverlay(styles, contentWidth)
	}

	bodyBlock := fitToHeight(body, bodyHeight)
	statusLine := m.renderStatusLine(styles, contentWidth)
	controlsLine := m.renderControlsLine(styles, contentWidth)

	ui := lipgloss.JoinVertical(lipgloss.Left, bodyBlock, statusLine, controlsLine)
	centered := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, lipgloss.NewStyle().Width(contentWidth).Render(ui))
	return fitToHeight(centered, m.height)
}

func (m *Model) renderScreen(styles viewStyles, width int) string {
	switch m.screen {
	case ScreenAuth:
		return m.renderAuth(styles)
	case ScreenLibrary:
		return m.renderLibrary(styles)
	case ScreenReader:
		return m.renderReader(styles)
	case ScreenCommunity:
		return m.renderCommunity(styles)
	case ScreenSettings:
		return m.renderSettings(styles)
	default:
		return styles.panel.Render("Unknown screen")
	}
}

func (m *Model) renderStatusLine(styles viewStyles, width int) string {
	message := m.status
	if strings.TrimSpace(message) == "" {
		message = "Ready"
	}
	if strings.TrimSpace(m.errMsg) != "" {
		message = message + " | error: " + m.errMsg
	}
	line := styles.status.Render(message)
	return lipgloss.NewStyle().Width(width).Render(line)
}

func (m *Model) renderControlsLine(styles viewStyles, width int) string {
	left := m.contextControls()
	right := styles.help.Render("Ctrl+P Command Palette")

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		gap = 2
	}
	line := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().Width(width).Render(line)
}

func (m *Model) contextControls() string {
	base := ""
	if !m.loggedIn {
		base = "Auth: enter submit | ctrl+n switch mode | g google oauth"
	} else {
		switch m.screen {
		case ScreenLibrary:
			if m.addActive {
				if m.addConfirmDuplicate {
					base = "y import anyway | n cancel duplicate | esc cancel"
				} else {
					base = "ctrl+s submit add | esc cancel add | tab/shift+tab next/prev field"
				}
			} else if m.searchActive {
				base = "ctrl+s or enter apply search | esc clear search"
			} else {
				base = "up/down move | a add | ctrl+f search | ctrl+r refresh"
			}
		case ScreenReader:
			base = "Reader: up/down scroll | h/l page | g/G jump | z zen toggle"
		case ScreenCommunity:
			base = "Community: up/down move | b borrow | r refresh"
		case ScreenSettings:
			base = "Settings: t theme | p typography | o accent | x clear | ] gutter"
		default:
			base = ""
		}
	}
	if m.palette.Active {
		base = "Palette: type to filter, Up/Down to select, Enter run, Esc close"
	}
	if m.loading.Active {
		base = "Loading... actions are temporarily blocked"
	}
	return base
}

func (m *Model) renderLoadingOverlay(styles viewStyles, width int) string {
	label := m.loading.Message
	if strings.TrimSpace(label) == "" {
		label = "Loading"
	}
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.sectionTitle.Render(m.spinner.View()+" "+label),
		styles.subtle.Render("Please wait..."),
	)
	block := styles.overlay.Width(min(width-4, 72)).Render(content)
	return lipgloss.Place(width, 10, lipgloss.Center, lipgloss.Center, block)
}

func (m *Model) renderHelpOverlay(styles viewStyles, width int) string {
	lines := []string{
		styles.sectionTitle.Render("Help"),
		"Navigation: Up/Down, Tab/Shift+Tab",
		"Activation: Enter",
		"Command Palette: Ctrl+P",
		"Help: Ctrl+H",
		"Quit: Ctrl+C",
		"",
		"View-specific commands are listed in the bottom-left controls.",
	}
	block := styles.overlay.Width(min(width-4, 90)).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, 12, lipgloss.Center, lipgloss.Center, block)
}

func (m *Model) renderPaletteOverlay(styles viewStyles, width int) string {
	rows := []string{styles.sectionTitle.Render("Command Palette"), m.palette.Input.View()}
	if len(m.palette.Entries) == 0 {
		rows = append(rows, styles.subtle.Render("No matching commands"))
	} else {
		maxRows := min(10, len(m.palette.Entries))
		lastGroup := ""
		for i := 0; i < maxRows; i++ {
			entry := m.palette.Entries[i]
			if entry.Command.Group != lastGroup {
				lastGroup = entry.Command.Group
				rows = append(rows, "")
				rows = append(rows, styles.sectionTitle.Render(strings.ToUpper(lastGroup)))
			}
			prefix := "  "
			rowStyle := styles.row
			if i == m.palette.Index {
				prefix = "> "
				rowStyle = styles.rowActive
			}
			disabled := ""
			if !entry.Enabled {
				disabled = " (disabled)"
			}
			label := prefix + entry.Command.Icon + " " + entry.Command.Title + disabled
			rows = append(rows, rowStyle.Render(label))
			rows = append(rows, styles.subtle.Render("   "+entry.Command.Description))
		}
	}
	rows = append(rows, styles.subtle.Render("Enter to run command, Esc to close"))
	block := styles.overlay.Width(min(width-4, 100)).Render(strings.Join(rows, "\n"))
	return lipgloss.Place(width, 16, lipgloss.Center, lipgloss.Center, block)
}

func (m *Model) renderSplash(styles viewStyles, width, height int) string {
	username := m.currentUserName()
	if !m.loggedIn {
		username = "guest"
	}
	progress := m.splashProgress
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	barWidth := min(40, max(20, width-20))
	filled := (barWidth * progress) / 100
	bar := "[" + strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled) + "]"

	logo := []string{
		" _     _ _                _     _       _    ",
		"| |   (_) |__  _ __ __ _  | |   (_)_ __ | | __",
		"| |   | | '_ \\| '__/ _` | | |   | | '_ \\| |/ /",
		"| |___| | |_) | | | (_| | | |___| | | | |   < ",
		"|_____|_|_.__/|_|  \\__,_| |_____|_|_| |_|_|\\_\\",
	}
	content := []string{
		styles.title.Render(strings.Join(logo, "\n")),
		"",
		styles.status.Render(bar + " " + strings.TrimSpace(m.splashMessage)),
		styles.subtle.Render(strings.TrimSpace(strings.Join([]string{"user:", username}, " "))),
	}
	block := lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(content, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}

func (m *Model) renderButtons(styles viewStyles, buttons []ActionButton) string {
	focused := m.focusedID()
	parts := make([]string, 0, len(buttons))
	for _, button := range buttons {
		label := " " + button.Label + " "
		s := styles.button
		if button.Disabled {
			s = styles.buttonDisabled
		} else if focused == button.ID {
			s = styles.buttonActive
		}
		parts = append(parts, s.Render(label))
	}
	return strings.Join(parts, " ")
}

func (m *Model) contentWidth() int {
	width := m.width
	if width <= 0 {
		return 80
	}
	gutter := m.gutterWidth()
	result := width - (gutter * 2)
	if result < 40 {
		return width
	}
	return result
}

func (m *Model) gutterWidth() int {
	if m.width <= 0 {
		return 0
	}
	switch strings.ToLower(strings.TrimSpace(m.uiSettings.GutterPreset)) {
	case "none":
		return 0
	case "narrow":
		return max(2, m.width/16)
	case "wide":
		return max(4, m.width/6)
	case "comfortable", "":
		fallthrough
	default:
		return max(3, m.width/8)
	}
}

func (m *Model) currentUserName() string {
	if m.user == nil || strings.TrimSpace(m.user.Username) == "" {
		return "guest"
	}
	return m.user.Username
}

func fitToHeight(text string, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
