package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Next     key.Binding
	Prev     key.Binding
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Enter    key.Binding
	Mode     key.Binding
	Rerun    key.Binding
	Cancel   key.Binding
	Syntax   key.Binding
	Fold     key.Binding
	Unfold   key.Binding
	Settings key.Binding
	Help     key.Binding
	Close    key.Binding
	Quit     key.Binding
}

func newKeys() keyMap {
	return keyMap{
		Next:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
		Prev:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "previous pane")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", "page down")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open or run")),
		Mode:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "auto or manual")),
		Rerun:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rerun")),
		Cancel:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cancel run")),
		Syntax:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "highlight")),
		Fold:     key.NewBinding(key.WithKeys("["), key.WithHelp("[", "fold")),
		Unfold:   key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "unfold")),
		Settings: key.NewBinding(key.WithKeys(","), key.WithHelp(",", "settings")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "commands")),
		Close:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close commands")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
