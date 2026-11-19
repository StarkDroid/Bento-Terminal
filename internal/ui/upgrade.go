package ui

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	upd "github.com/StarkDroid/Bento-Terminal/internal/update"
)

const (
	updateButtonID = "update"
	laterButtonID  = "later"
)

type updateCheckedMsg struct {
	release upd.Release
	newer   bool
	err     error
}

func checkUpdateCmd(version string) tea.Cmd {
	if !upd.Versioned(version) || os.Getenv("BENTO_NO_UPDATE_CHECK") != "" {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		release, newer, err := upd.Check(ctx, version)
		return updateCheckedMsg{release: release, newer: newer, err: err}
	}
}

func (m Model) updateAvailable() bool {
	return m.updateRelease.Version != "" && !m.updateRunning && !m.updateDone
}

func (m Model) updateCommand() upd.Method {
	if m.updateMethod.Command != "" {
		return m.updateMethod
	}
	exe, err := os.Executable()
	if err != nil {
		exe = ""
	}
	return upd.ForExecutable(exe)
}

func (m Model) anyOverlayOpen() bool {
	return m.updateOpen || m.newFileOpen || m.finderOpen || m.findOpen ||
		m.recentOpen || m.paletteOpen || m.settingsOpen
}

func (m Model) onUpdateChecked(msg updateCheckedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil || !msg.newer {
		return m, nil
	}
	m.updateRelease = msg.release
	if !m.updateDismissed && !m.anyOverlayOpen() && m.focus != PaneTerminal && m.aiSession == nil {
		m.updateOpen = true
	}
	return m, nil
}

func (m Model) onUpdateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Update), key.Matches(msg, m.keys.Enter):
		return m.startUpdate()
	case key.Matches(msg, m.keys.Close):
		m.updateOpen = false
		m.updateDismissed = true
		return m, nil
	case msg.String() == "l":
		m.updateOpen = false
		m.updateDismissed = true
		return m, nil
	}
	return m, nil
}

// startUpdate runs the install-specific update command in the Terminal pane.
// The command is visible and cancellable like any other terminal command.
func (m Model) startUpdate() (tea.Model, tea.Cmd) {
	if !m.updateAvailable() {
		return m, nil
	}
	method := m.updateCommand()
	if strings.TrimSpace(method.Command) == "" {
		return m, nil
	}
	m.updateMethod = method
	m.updateOpen = false
	m.settingsOpen = false
	m.paletteOpen = false
	m.focus = PaneTerminal
	m.termInput = method.Command
	m.termCursor = len([]rune(method.Command))
	m.updateRunning = true
	next, cmd := m.runTerminal()
	mm := next.(Model)
	mm.updateRunning = true
	mm.updateMethod = method
	return mm, cmd
}

// updateBadge is the status-bar badge. The u key is part of the badge so it
// only appears while an update is waiting.
func (m Model) updateStartFailed() Model {
	if m.runTarget == targetTerminal && m.updateRunning {
		m.updateRunning = false
		m = m.appendTerm("✗ Update failed to start. Run the command manually in the terminal.")
	}
	return m
}

func (m Model) updateAboutLine() string {
	method := m.updateCommand()
	line := "press u to update via " + method.Name
	if len(method.Command) <= 36 {
		line += " · " + method.Command
	}
	return line
}

func (m Model) updateBadge() []piece {
	switch {
	case m.updateRunning:
		return []piece{pieceText(" updating… ", colorOcean, colorSun)}
	case m.updateDone:
		return []piece{pieceText(" restart Bento ", colorOcean, colorSun)}
	case m.updateAvailable():
		return []piece{pieceText(" update "+upd.Display(m.updateRelease.Version)+" · u ", colorOcean, colorSun)}
	}
	return nil
}

type updateFrame struct {
	innerW  int
	boxH    int
	rows    []string
	buttons []aboutHit
}

func (m Model) updateView() string {
	frame := m.updateFrame()
	title := paintTitle(" ↑ Update", frame.innerW, true)
	box := renderPanel(title, strings.Join(frame.rows, "\n"), frame.innerW+2, frame.boxH, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
}

func (m Model) updateFrame() updateFrame {
	innerW := m.width - 6
	if innerW > 62 {
		innerW = 62
	}
	if innerW < 24 {
		innerW = m.width - 2
	}
	if innerW < 8 {
		innerW = 8
	}

	rows := []string{
		settingsBlank(innerW),
		settingsRow("Update available", colorSun, innerW, true),
		settingsRow(upd.Display(m.version)+"  →  "+upd.Display(m.updateRelease.Version), colorSand, innerW, true),
		settingsRow("A new Bento release is ready.", colorSandDim, innerW, true),
		settingsRow("via "+m.updateCommand().Name, colorSandDim, innerW, true),
		settingsBlank(innerW),
		settingsRow("Later keeps the u shortcut in the status bar.", colorMoss, innerW, true),
	}

	buttonRow := len(rows)
	row, buttons := updateRow(innerW)
	rows = append(rows, row)

	boxH := len(rows) + 3
	if boxH > m.height-1 {
		boxH = m.height - 1
	}
	if boxH < 7 {
		boxH = 7
	}
	bodyH := boxH - 3
	if len(rows) > bodyH {
		rows = rows[:bodyH]
	}

	x0 := (m.width - (innerW + 2)) / 2
	y0 := (m.height - boxH) / 2

	hits := make([]aboutHit, 0, len(buttons))
	if buttonRow < bodyH {
		for _, b := range buttons {
			if b.x+b.piece.width > innerW {
				continue
			}
			hits = append(hits, aboutHit{
				id:    b.id,
				x:     x0 + 1 + b.x,
				y:     y0 + 2 + buttonRow,
				width: b.piece.width,
			})
		}
	}
	return updateFrame{innerW: innerW, boxH: boxH, rows: rows, buttons: hits}
}

func updateRow(innerW int) (string, []aboutButton) {
	specs := []struct {
		id    string
		key   string
		label string
		fg    lipgloss.TerminalColor
	}{
		{updateButtonID, "u", "update now", colorSun},
		{laterButtonID, "esc", "later", colorSand},
	}

	bg := colorSelectDim
	buttons := make([]aboutButton, 0, len(specs))
	total := 0
	for i, s := range specs {
		parts := []piece{
			pieceText(s.key, colorLagoonBright, bg),
			pieceText(" ", colorSand, bg),
			pieceText(s.label, s.fg, bg),
		}
		p := chip(bg, parts)
		if i > 0 {
			total++
		}
		total += p.width
		buttons = append(buttons, aboutButton{id: s.id, piece: p})
	}

	pad := (innerW - total) / 2
	if pad < 0 {
		pad = 0
	}
	parts := []piece{pieceText(strings.Repeat(" ", pad), colorSand, colorPanel)}
	x := pad
	for i := range buttons {
		if i > 0 {
			parts = append(parts, pieceText(" ", colorSand, colorPanel))
			x++
		}
		buttons[i].x = x
		parts = append(parts, buttons[i].piece)
		x += buttons[i].piece.width
	}
	return joinPieces(innerW, colorPanel, parts), buttons
}

func (m Model) onUpdateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	for _, hit := range m.updateFrame().buttons {
		if msg.Y != hit.y || msg.X < hit.x || msg.X >= hit.x+hit.width {
			continue
		}
		switch hit.id {
		case updateButtonID:
			return m.startUpdate()
		case laterButtonID:
			m.updateOpen = false
			m.updateDismissed = true
		}
		return m, nil
	}
	return m, nil
}
