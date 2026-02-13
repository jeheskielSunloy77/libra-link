package app

import (
	"fmt"
	"strings"
)

func (m *Model) renderCommunity(styles viewStyles) string {
	rows := []string{styles.sectionTitle.Render("Community Shares")}
	focused := m.focusedID()

	if len(m.shares) == 0 {
		rows = append(rows, styles.subtle.Render("No community shares cached."))
	} else {
		for i, share := range m.shares {
			id := fmt.Sprintf("community.share.%d", i)
			prefix := "  "
			rowStyle := styles.row
			if focused == id {
				prefix = "> "
				rowStyle = styles.rowActive
			}
			title := fallback(share.Title, share.ID)
			rows = append(rows, rowStyle.Render(fmt.Sprintf("%s%s (%s)", prefix, title, share.Status)))
		}
	}

	return styles.panel.Render(strings.Join(rows, "\n"))
}
