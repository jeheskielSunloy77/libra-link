package keymap

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit     key.Binding
	Palette  key.Binding
	Help     key.Binding
	MoveNext key.Binding
	MovePrev key.Binding
	Submit   key.Binding
}

func Default() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Palette: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "command palette"),
		),
		Help: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "help"),
		),
		MoveNext: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("down/tab", "next selectable"),
		),
		MovePrev: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("up/shift+tab", "prev selectable"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "activate/submit"),
		),
	}
}
