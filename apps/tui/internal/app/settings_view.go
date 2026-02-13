package app

import (
	"fmt"
	"strings"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/theme"
)

func (m *Model) renderSettings(styles viewStyles) string {
	baseTokens := theme.DefaultTokens(theme.Mode(m.prefs.ThemeMode))
	resolved, err := theme.ApplyOverrides(baseTokens, m.prefs.ThemeOverrides)
	validation := "theme overrides valid"
	if err != nil {
		validation = err.Error()
	}

	rows := []string{
		styles.sectionTitle.Render("Settings"),
		fmt.Sprintf("readingMode: %s", m.prefs.ReadingMode),
		fmt.Sprintf("themeMode: %s", m.prefs.ThemeMode),
		fmt.Sprintf("typographyProfile: %s", m.prefs.TypographyProfile),
		fmt.Sprintf("gutterPreset: %s", m.uiSettings.GutterPreset),
		fmt.Sprintf("tokens: bg=%s text=%s accent=%s progress=%s", resolved.Background, resolved.Text, resolved.Accent, resolved.Progress),
		styles.subtle.Render(validation),
	}

	return styles.panel.Render(strings.Join(rows, "\n"))
}
