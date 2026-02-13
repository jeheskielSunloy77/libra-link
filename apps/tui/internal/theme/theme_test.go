package theme

import "testing"

func TestApplyOverrides_SucceedsForValidTokens(t *testing.T) {
	base := DefaultTokens(ModeDark)
	resolved, err := ApplyOverrides(base, map[string]string{
		"background": "#101010",
		"text":       "#f5f5f5",
		"accent":     "#ff8800",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved.Background != "#101010" {
		t.Fatalf("unexpected background: %s", resolved.Background)
	}
	if resolved.Text != "#f5f5f5" {
		t.Fatalf("unexpected text: %s", resolved.Text)
	}
	if resolved.Accent != "#ff8800" {
		t.Fatalf("unexpected accent: %s", resolved.Accent)
	}
}

func TestApplyOverrides_RejectsUnknownToken(t *testing.T) {
	_, err := ApplyOverrides(DefaultTokens(ModeDark), map[string]string{"unknown": "#ffffff"})
	if err == nil {
		t.Fatal("expected error for unknown token")
	}
}

func TestApplyOverrides_RejectsLowContrast(t *testing.T) {
	_, err := ApplyOverrides(DefaultTokens(ModeDark), map[string]string{
		"background": "#222222",
		"text":       "#252525",
	})
	if err == nil {
		t.Fatal("expected contrast validation error")
	}
}
