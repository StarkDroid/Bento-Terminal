package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// bentoArt is the about wordmark. It uses block glyphs that Menlo and SF Mono
// can draw, so it needs no extra font.
var bentoArt = []string{
	"██████╗ ███████╗███╗   ██╗████████╗ ██████╗",
	"██╔══██╗██╔════╝████╗  ██║╚══██╔══╝██╔═══██╗",
	"██████╔╝█████╗  ██╔██╗ ██║   ██║   ██║   ██║",
	"██╔══██╗██╔══╝  ██║╚██╗██║   ██║   ██║   ██║",
	"██████╔╝███████╗██║ ╚████║   ██║   ╚██████╔╝",
	"╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝",
}

// About overlay button ids.
const (
	aboutGitHub   = "github"
	aboutCommands = "commands"
	aboutQuit     = "quit"
	aboutClose    = "close"
)

// aboutHit is the screen rectangle of one chip in the about overlay.
type aboutHit struct {
	id    string
	x     int
	y     int
	width int
}

// settingsFrame is the about overlay's layout: the rows to draw and the screen
// cells of its buttons. View and the mouse handler share it, so a click lands
// on the chip it covers.
type settingsFrame struct {
	innerW  int
	boxH    int
	rows    []string
	buttons []aboutHit
}

func (m Model) openSettings() Model {
	if m.settingsOpen {
		m.settingsOpen = false
		return m
	}
	m.settingsOpen = true
	m.paletteOpen = false
	return m
}

func (m Model) settingsView() string {
	frame := m.settingsFrame()
	title := paintTitle(" , Settings · About", frame.innerW, true)
	box := renderPanel(title, strings.Join(frame.rows, "\n"), frame.innerW+2, frame.boxH, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
}

func (m Model) settingsFrame() settingsFrame {
	innerW := m.width - 6
	if innerW > 68 {
		innerW = 68
	}
	if innerW < 24 {
		innerW = m.width - 2
	}
	if innerW < 8 {
		innerW = 8
	}

	rows := make([]string, 0, 16)
	for _, line := range bentoArt {
		rows = append(rows, lipgloss.NewStyle().
			Foreground(colorLagoonBright).
			Background(colorPanel).
			Width(innerW).
			Align(lipgloss.Center).
			Render(truncateCells(line, innerW)))
	}
	rows = append(rows, settingsBlank(innerW))
	rows = append(rows, settingsRow("terminal workspace", colorSand, innerW, true))
	rows = append(rows, settingsRow(versionLabel(m.version), colorLagoon, innerW, true))
	rows = append(rows, settingsBlank(innerW))
	rows = append(rows, settingsRow("A full-screen terminal workspace for the directory you open.", colorSand, innerW, false))
	rows = append(rows, settingsRow("Preview, fold, browse, and build without leaving the terminal.", colorSandDim, innerW, false))
	rows = append(rows, settingsBlank(innerW))

	buttonRow := len(rows)
	row, buttons := aboutRow(innerW)
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

	// renderPanel draws the title in the first row inside the border, so the
	// body starts two rows below the box's top edge.
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
	return settingsFrame{innerW: innerW, boxH: boxH, rows: rows, buttons: hits}
}

func settingsRow(text string, fg lipgloss.TerminalColor, width int, center bool) string {
	style := lipgloss.NewStyle().Foreground(fg).Background(colorPanel).Width(width)
	if center {
		return style.Align(lipgloss.Center).Render(truncateCells(text, width))
	}
	return style.Render(truncateCells(" "+text, width))
}

func settingsBlank(width int) string {
	return lipgloss.NewStyle().Background(colorPanel).Width(width).Render("")
}

func versionLabel(version string) string {
	version = strings.TrimSpace(version)
	switch {
	case version == "", version == "dev", version == "development":
		return "development build"
	case strings.HasPrefix(version, "v"):
		return version
	}
	return "v" + version
}

// aboutButton is one chip in the about row and its offset inside the body.
type aboutButton struct {
	id    string
	x     int
	piece piece
}

func aboutRow(innerW int) (string, []aboutButton) {
	specs := []struct {
		id    string
		key   string
		label string
		fg    lipgloss.TerminalColor
	}{
		{aboutGitHub, "", "github", colorPalm},
		{aboutCommands, "?", "commands", colorSand},
		{aboutQuit, "q", "quit", colorSand},
		{aboutClose, "esc", "close", colorSand},
	}

	bg := colorSelectDim
	buttons := make([]aboutButton, 0, len(specs))
	total := 0
	for i, s := range specs {
		parts := make([]piece, 0, 3)
		if s.key != "" {
			parts = append(parts, pieceText(s.key, colorLagoonBright, bg))
			parts = append(parts, pieceText(" ", colorSand, bg))
		}
		parts = append(parts, pieceText(s.label, s.fg, bg))
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

// settingsClick runs the about button under a left click. Other mouse events are
// ignored while the overlay is open.
func (m Model) settingsClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	for _, hit := range m.settingsFrame().buttons {
		if msg.Y != hit.y || msg.X < hit.x || msg.X >= hit.x+hit.width {
			continue
		}
		switch hit.id {
		case aboutGitHub:
			return m, openURL(projectURL)
		case aboutCommands:
			return m.openPalette(), nil
		case aboutQuit:
			return m.stopRun(), tea.Quit
		case aboutClose:
			m.settingsOpen = false
		}
		return m, nil
	}
	return m, nil
}
