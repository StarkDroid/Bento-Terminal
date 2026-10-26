package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// paletteItem is one row in the per-pane command list.
type paletteItem struct {
	section string
	action  action
}

func (m Model) openPalette() Model {
	if m.paletteOpen {
		m.paletteOpen = false
		return m
	}
	m.paletteOpen = true
	m.settingsOpen = false
	m.palettePane = m.focus
	m.paletteCursor = 0
	return m
}

func (m Model) onPaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.paletteItems()
	switch {
	case key.Matches(msg, m.keys.Close, m.keys.Help):
		m.paletteOpen = false
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.paletteCursor = clamp(m.paletteCursor-1, 0, max(0, len(items)-1))
		return m, nil
	case key.Matches(msg, m.keys.Down):
		m.paletteCursor = clamp(m.paletteCursor+1, 0, max(0, len(items)-1))
		return m, nil
	case key.Matches(msg, m.keys.PageUp):
		m.paletteCursor = clamp(m.paletteCursor-8, 0, max(0, len(items)-1))
		return m, nil
	case key.Matches(msg, m.keys.PageDown):
		m.paletteCursor = clamp(m.paletteCursor+8, 0, max(0, len(items)-1))
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		if m.paletteCursor >= 0 && m.paletteCursor < len(items) {
			return m.runAction(items[m.paletteCursor].action.id)
		}
		return m, nil
	}
	// A listed shortcut works straight from the list too.
	if len(msg.Runes) == 1 {
		for _, it := range items {
			if key.Matches(msg, it.action.binding) {
				return m.runAction(it.action.id)
			}
		}
	}
	return m, nil
}
