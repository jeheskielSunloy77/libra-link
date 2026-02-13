package app

import (
	"fmt"
	"strings"
)

func (m *Model) renderAuth(styles viewStyles) string {
	rows := []string{}
	if m.authMode == authModeSignUp {
		rows = append(rows,
			styles.sectionTitle.Render("Create Account"),
			m.signupEmailInput.View(),
			m.signupUserInput.View(),
			m.signupPWInput.View(),
			m.signupConfirmInput.View(),
		)
	} else {
		rows = append(rows,
			styles.sectionTitle.Render("Sign In"),
			m.loginIDInput.View(),
			m.loginPWInput.View(),
		)
	}

	googleLine := "Google OAuth ready"
	if m.googleAuthURL != "" {
		googleLine = fmt.Sprintf("Open URL: %s", m.googleAuthURL)
	}
	rows = append(rows, styles.subtle.Render(googleLine))
	if m.googleCode != "" {
		expires := ""
		if !m.googleExpires.IsZero() {
			expires = m.googleExpires.Format("3:04PM")
		}
		rows = append(rows, styles.subtle.Render("Device code: "+m.googleCode+" (expires "+expires+")"))
	}

	return styles.panel.Render(strings.Join(rows, "\n"))
}
