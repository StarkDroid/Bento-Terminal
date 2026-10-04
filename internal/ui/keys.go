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
		Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "collapse")),
		Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "expand")),
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", "page down")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open or run")),
		Mode:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "auto or manual")),
		Rerun:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rerun")),
		Cancel:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cancel run")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Close:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close help")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Up, k.Enter, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Up, k.Down, k.Left, k.Right},
		{k.PageUp, k.PageDown, k.Enter},
		{k.Mode, k.Rerun, k.Cancel},
		{k.Help, k.Close, k.Quit},
	}
}
