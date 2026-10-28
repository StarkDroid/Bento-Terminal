package ui

import (
	"strings"

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

	var rows []string
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
	rows = append(rows, settingsRow("github.com/StarkDroid/Bento-Terminal", colorPalm, innerW, false))
	rows = append(rows, settingsBlank(innerW))
	rows = append(rows, settingsRow("?  commands     ,  settings     q  quit", colorSandDim, innerW, false))
	rows = append(rows, settingsRow("esc  close", colorMoss, innerW, false))

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

	title := paintTitle(" , Settings · About", innerW, true)
	box := renderPanel(title, strings.Join(rows, "\n"), innerW+2, boxH, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
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
