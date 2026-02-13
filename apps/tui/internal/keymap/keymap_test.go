package keymap

import "testing"

func TestDefaultBindings(t *testing.T) {
	km := Default()

	if len(km.Quit.Keys()) == 0 || km.Quit.Keys()[0] != "ctrl+c" {
		t.Fatalf("quit key binding mismatch: %#v", km.Quit.Keys())
	}
	if len(km.Palette.Keys()) == 0 || km.Palette.Keys()[0] != "ctrl+p" {
		t.Fatalf("palette key binding mismatch: %#v", km.Palette.Keys())
	}
	if len(km.Help.Keys()) == 0 || km.Help.Keys()[0] != "ctrl+h" {
		t.Fatalf("help key binding mismatch: %#v", km.Help.Keys())
	}
	if len(km.MoveNext.Keys()) == 0 || km.MoveNext.Keys()[0] != "down" {
		t.Fatalf("move next key binding mismatch: %#v", km.MoveNext.Keys())
	}
}
