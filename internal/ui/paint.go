package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type piece struct {
	view  string
	width int
}

func pieceText(text string, fg, bg lipgloss.TerminalColor) piece {
	return piece{
		view:  lipgloss.NewStyle().Foreground(fg).Background(bg).Render(text),
		width: lipgloss.Width(text),
	}
}

func joinPieces(width int, bg lipgloss.TerminalColor, parts []piece) string {
	if width < 1 {
		return ""
	}
	var b strings.Builder
	used := 0
	for _, p := range parts {
		if p.width <= 0 || used >= width {
			continue
		}
		if used+p.width > width {
			remain := width - used
			b.WriteString(lipgloss.NewStyle().Background(bg).Width(remain).Render(""))
			used = width
			break
		}
		b.WriteString(p.view)
		used += p.width
	}
	if used < width {
		b.WriteString(lipgloss.NewStyle().Background(bg).Width(width - used).Render(""))
	}
	return b.String()
}

func borderStyle(focused bool) lipgloss.Style {
	color := colorBorderDim
	if focused {
		color = colorLagoonBright
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		BorderBackground(colorPanel).
		Background(colorPanel).
		Foreground(colorSand).
		Align(lipgloss.Left, lipgloss.Top)
}

func paintTitle(text string, width int, focused bool) string {
	fg := colorSandDim
	if focused {
		fg = colorLagoonBright
	}
	style := lipgloss.NewStyle().Foreground(fg).Background(colorPanel).Width(width)
	if focused {
		style = style.Bold(true)
	}
	return style.Render(truncateCells(text, width))
}

func paintRow(label string, width int, selected, focused bool) string {
	if width < 1 {
		return ""
	}
	fg := colorSand
	bg := colorPanel
	style := lipgloss.NewStyle().Width(width)
	switch {
	case selected && focused:
		bg = colorSelect
		style = style.Bold(true)
	case selected:
		bg = colorSelectDim
	}
	return style.Foreground(fg).Background(bg).Render(truncateCells(label, width))
}

func noteLine(text string, width int) string {
	if width < 1 {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(colorSandDim).
		Background(colorPanel).
		Width(width).
		Render(truncateCells(text, width))
}

func sepLine(width int) string {
	if width < 1 {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(colorBorderDim).
		Background(colorPanel).
		Width(width).
		Render(strings.Repeat("─", width))
}

func fitBlock(body string, width, height int) string {
	if height <= 0 {
		return ""
	}
	blank := lipgloss.NewStyle().Width(width).Background(colorPanel).Render("")
	var lines []string
	if body != "" {
		lines = strings.Split(body, "\n")
	}
	out := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(lines) && lines[i] != "" {
			out[i] = lines[i]
			continue
		}
		if i < len(lines) {
			out[i] = blank
			continue
		}
		out[i] = blank
	}
	return strings.Join(out, "\n")
}

func renderPanel(titleLine, body string, width, height int, focused bool) string {
	st := borderStyle(focused)
	innerW := width - st.GetHorizontalBorderSize()
	innerH := height - st.GetVerticalBorderSize()
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}
	if lipgloss.Width(titleLine) != innerW {
		titleLine = paintTitle(titleLine, innerW, focused)
	}
	bodyH := innerH - 1
	content := titleLine
	if bodyH > 0 {
		content += "\n" + fitBlock(body, innerW, bodyH)
	}
	return st.Width(innerW).Height(innerH).Render(content)
}

func panel(title, body string, width, height int, focused bool) string {
	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}
	return renderPanel(paintTitle(" "+title, innerW, focused), body, width, height, focused)
}

func gapRow(width int) string {
	if width < 1 {
		width = 1
	}
	return lipgloss.NewStyle().Background(colorOcean).Width(width).Render("")
}

func gapCol(height int) string {
	if height < 1 {
		height = 1
	}
	line := lipgloss.NewStyle().Background(colorOcean).Width(1).Render("")
	lines := make([]string, height)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func windowLines(lines []string, offset, width, height int) string {
	if height <= 0 {
		return ""
	}
	blank := lipgloss.NewStyle().Width(width).Background(colorPanel).Render("")
	out := make([]string, height)
	for i := 0; i < height; i++ {
		idx := offset + i
		if idx >= 0 && idx < len(lines) && lines[idx] != "" {
			out[i] = lines[idx]
			continue
		}
		out[i] = blank
	}
	return strings.Join(out, "\n")
}

func stylePlain(lines []string, width int) []string {
	st := lipgloss.NewStyle().Foreground(colorSand).Background(colorPanel).Width(width)
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = st.Render(truncateCells(line, width))
	}
	return out
}
