package keymap

import "testing"

func TestDefaultBindings(t *testing.T) {
	km := Default()

	if len(km.Quit.Keys()) == 0 || km.Quit.Keys()[0] != "q" {
		t.Fatalf("quit key binding mismatch: %#v", km.Quit.Keys())
	}
	if len(km.ToggleZen.Keys()) == 0 || km.ToggleZen.Keys()[0] != "z" {
		t.Fatalf("toggle zen key binding mismatch: %#v", km.ToggleZen.Keys())
	}
	if len(km.Next.Keys()) == 0 || km.Next.Keys()[0] != "tab" {
		t.Fatalf("next key binding mismatch: %#v", km.Next.Keys())
	}
}
