package theme

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Mode string

const (
	ModeLight        Mode = "light"
	ModeDark         Mode = "dark"
	ModeSepia        Mode = "sepia"
	ModeHighContrast Mode = "high_contrast"
)

type Tokens struct {
	Background string
	Text       string
	Accent     string
	Progress   string
}

var (
	hexColorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{6})$`)
	allowedKeys   = map[string]struct{}{
		"background": {},
		"text":       {},
		"accent":     {},
		"progress":   {},
	}
)

func DefaultTokens(mode Mode) Tokens {
	switch mode {
	case ModeLight:
		return Tokens{Background: "#f5f5f2", Text: "#202022", Accent: "#2a6fdb", Progress: "#5f87ff"}
	case ModeSepia:
		return Tokens{Background: "#f4ecd8", Text: "#4e3f2f", Accent: "#9c5a16", Progress: "#b88c3a"}
	case ModeHighContrast:
		return Tokens{Background: "#000000", Text: "#ffffff", Accent: "#00ffff", Progress: "#ffff00"}
	case ModeDark:
		fallthrough
	default:
		return Tokens{Background: "#111318", Text: "#f5f7fa", Accent: "#64b5f6", Progress: "#82b1ff"}
	}
}

func ApplyOverrides(base Tokens, overrides map[string]string) (Tokens, error) {
	if len(overrides) == 0 {
		return base, nil
	}

	normalized := map[string]string{}
	for rawKey, rawValue := range overrides {
		key := strings.ToLower(strings.TrimSpace(rawKey))
		value := strings.TrimSpace(rawValue)
		if _, ok := allowedKeys[key]; !ok {
			return base, fmt.Errorf("unsupported token %q", rawKey)
		}
		if !hexColorRegex.MatchString(value) {
			return base, fmt.Errorf("invalid hex color for %q", rawKey)
		}
		normalized[key] = strings.ToLower(value)
	}

	if value, ok := normalized["background"]; ok {
		base.Background = value
	}
	if value, ok := normalized["text"]; ok {
		base.Text = value
	}
	if value, ok := normalized["accent"]; ok {
		base.Accent = value
	}
	if value, ok := normalized["progress"]; ok {
		base.Progress = value
	}

	if contrastRatio(base.Background, base.Text) < 4.5 {
		return base, fmt.Errorf("background/text contrast is below WCAG AA minimum")
	}

	return base, nil
}

func contrastRatio(bgHex, textHex string) float64 {
	bgLum := relativeLuminance(bgHex)
	textLum := relativeLuminance(textHex)
	lighter := math.Max(bgLum, textLum)
	darker := math.Min(bgLum, textLum)
	return (lighter + 0.05) / (darker + 0.05)
}

func relativeLuminance(hex string) float64 {
	r, g, b := hexToRGB(hex)
	rn := normalizeChannel(r)
	gn := normalizeChannel(g)
	bn := normalizeChannel(b)
	return 0.2126*rn + 0.7152*gn + 0.0722*bn
}

func hexToRGB(hex string) (int64, int64, int64) {
	clean := strings.TrimPrefix(hex, "#")
	if len(clean) != 6 {
		return 0, 0, 0
	}
	r, errR := strconv.ParseInt(clean[0:2], 16, 64)
	g, errG := strconv.ParseInt(clean[2:4], 16, 64)
	b, errB := strconv.ParseInt(clean[4:6], 16, 64)
	if errR != nil || errG != nil || errB != nil {
		return 0, 0, 0
	}
	return r, g, b
}

func normalizeChannel(v int64) float64 {
	c := float64(v) / 255
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}
