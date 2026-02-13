package keymap

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit key.Binding
	Help key.Binding
	Next key.Binding
	Prev key.Binding

	ToggleZen key.Binding
	ReaderUp  key.Binding
	ReaderDown key.Binding
	ReaderPrevPage key.Binding
	ReaderNextPage key.Binding
}

func Default() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev view"),
		),
		ToggleZen: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "toggle zen"),
		),
		ReaderUp: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "up"),
		),
		ReaderDown: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "down"),
		),
		ReaderPrevPage: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "prev page"),
		),
		ReaderNextPage: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "next page"),
		),
	}
}
