package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appstate "github.com/StarkDroid/Bento-Terminal/internal/state"
)

// openRecent shows the projects remembered from previous sessions. The current
// project is omitted because reopening it would only reload the same model.
func (m Model) openRecent() (tea.Model, tea.Cmd) {
	if m.previewEditing {
		return m, nil
	}
	m.recentOpen = true
	m.recentCursor = 0
	m.recentScroll = 0
	m.recent = nil
	if m.store != nil {
		for _, p := range m.store.Recent() {
			if filepath.Clean(p.Path) == filepath.Clean(m.root) {
				continue
			}
			if info, err := os.Stat(p.Path); err == nil && info.IsDir() {
				m.recent = append(m.recent, p)
			}
		}
	}
	return m, nil
}

func (m Model) onRecentKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.recentOpen = false
		return m, nil
	case tea.KeyEnter:
		return m.openRecentSelection()
	case tea.KeyUp:
		return m.moveRecent(-1), nil
	case tea.KeyDown:
		return m.moveRecent(1), nil
	case tea.KeyPgUp:
		return m.moveRecent(-m.recentRows()), nil
	case tea.KeyPgDown:
		return m.moveRecent(m.recentRows()), nil
	}
	return m, nil
}

func (m Model) moveRecent(delta int) Model {
	if len(m.recent) == 0 {
		return m
	}
	m.recentCursor = clamp(m.recentCursor+delta, 0, len(m.recent)-1)
	m.recentScroll = ensure(m.recentCursor, m.recentScroll, m.recentRows())
	return m
}

// openRecentSelection saves this project and starts the selected one in the
// same terminal. Restoring the new project is handled by NewWithState.
func (m Model) openRecentSelection() (tea.Model, tea.Cmd) {
	if len(m.recent) == 0 {
		m.recentOpen = false
		return m, nil
	}
	p := m.recent[clamp(m.recentCursor, 0, len(m.recent)-1)]
	m.recentOpen = false
	if filepath.Clean(p.Path) == filepath.Clean(m.root) {
		return m, nil
	}
	m = m.stopRun()
	m = m.closeAISession()
	m.SaveState()
	next := NewWithState(p.Path, m.version, m.store)
	next.width = m.width
	next.height = m.height
	next = next.reflowPreview(false)
	return next, next.Init()
}

func (m Model) recentRows() int {
	rows := 8
	if m.height > 0 {
		if limit := m.height - 8; rows > limit {
			rows = limit
		}
	}
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m Model) recentLine(p appstate.Project, width int, selected bool) string {
	name := p.Name
	if name == "" {
		name = filepath.Base(p.Path)
	}
	label := " " + name + "  " + appstate.ShortPath(p.Path)
	return paintRow(label, width, selected, true)
}

// recentView is the recent-project picker.
func (m Model) recentView() string {
	innerW := m.width - 6
	if innerW > 76 {
		innerW = 76
	}
	if innerW < 24 {
		innerW = m.width - 2
	}
	if innerW < 8 {
		innerW = 8
	}
	rows := m.recentRows()
	list := make([]string, 0, rows)
	switch {
	case len(m.recent) == 0:
		list = append(list, noteLine(" No other recent projects", innerW))
	default:
		start := ensure(m.recentCursor, m.recentScroll, rows)
		end := min(start+rows, len(m.recent))
		for i := start; i < end; i++ {
			list = append(list, m.recentLine(m.recent[i], innerW, i == m.recentCursor))
		}
	}
	for len(list) < rows {
		list = append(list, noteLine("", innerW))
	}

	body := []string{
		noteLine(" Open a project and pick up where you left off.", innerW),
		sepLine(innerW),
	}
	body = append(body, list...)
	body = append(body, noteLine(" ↑↓ pick · enter open · esc close", innerW))

	count := "0"
	if len(m.recent) > 0 {
		count = fmt.Sprintf("%d/%d", m.recentCursor+1, len(m.recent))
	}
	title := joinLeftRight(innerW, colorPanel,
		[]piece{pieceText(" ▸ Recent projects", colorLagoonBright, colorPanel)},
		[]piece{pieceText(count+" ", colorLagoon, colorPanel)},
	)
	boxH := len(body) + 3
	if boxH > m.height {
		boxH = m.height
	}
	if boxH < 5 {
		boxH = 5
	}
	box := renderPanel(title, strings.Join(body, "\n"), innerW+2, boxH, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
}
