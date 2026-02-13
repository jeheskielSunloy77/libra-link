package app

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/theme"
)

type viewStyles struct {
	title          lipgloss.Style
	panel          lipgloss.Style
	sectionTitle   lipgloss.Style
	subtle         lipgloss.Style
	status         lipgloss.Style
	errorText      lipgloss.Style
	button         lipgloss.Style
	buttonActive   lipgloss.Style
	buttonDisabled lipgloss.Style
	row            lipgloss.Style
	rowActive      lipgloss.Style
	overlay        lipgloss.Style
	inputLabel     lipgloss.Style
	help           lipgloss.Style
}

func (m *Model) styles() viewStyles {
	mode := theme.Mode(m.prefs.ThemeMode)
	base := theme.DefaultTokens(mode)
	resolved, err := theme.ApplyOverrides(base, m.prefs.ThemeOverrides)
	if err != nil {
		resolved = base
	}

	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(resolved.Text))

	return viewStyles{
		title:          baseStyle.Copy().Bold(true).Foreground(lipgloss.Color(resolved.Accent)),
		panel:          baseStyle.Copy().Padding(1, 0),
		sectionTitle:   baseStyle.Copy().Bold(true).Foreground(lipgloss.Color(resolved.Accent)),
		subtle:         baseStyle.Copy().Foreground(lipgloss.Color(resolved.Progress)),
		status:         baseStyle.Copy(),
		errorText:      baseStyle.Copy().Foreground(lipgloss.Color("#ff6b6b")),
		button:         baseStyle.Copy().Border(lipgloss.NormalBorder()).Padding(0, 1).Foreground(lipgloss.Color(resolved.Accent)),
		buttonActive:   baseStyle.Copy().Bold(true).Border(lipgloss.ThickBorder()).Padding(0, 1).Foreground(lipgloss.Color(resolved.Accent)),
		buttonDisabled: baseStyle.Copy().Border(lipgloss.NormalBorder()).Padding(0, 1).Foreground(lipgloss.Color(resolved.Progress)),
		row:            baseStyle.Copy(),
		rowActive:      baseStyle.Copy().Bold(true).Foreground(lipgloss.Color(resolved.Accent)),
		overlay:        baseStyle.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color(resolved.Accent)).Padding(1),
		inputLabel:     baseStyle.Copy().Bold(true),
		help:           baseStyle.Copy().Foreground(lipgloss.Color(resolved.Progress)),
	}
}
