package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func numberedLines(content string, width int) []string {
	raw := fileLines(content)
	if width < 1 {
		return nil
	}
	gutterW := len(strconv.Itoa(len(raw)))
	textW := width - gutterW - 3
	gutterSt := lipgloss.NewStyle().Foreground(colorSandDim).Background(colorPanel)
	textSt := lipgloss.NewStyle().Foreground(colorSand).Background(colorPanel)
	if textW < 1 {
		st := textSt.Width(width)
		out := make([]string, len(raw))
		for i, line := range raw {
			out[i] = st.Render(truncateCells(fmt.Sprintf("%d %s", i+1, line), width))
		}
		return out
	}
	out := make([]string, len(raw))
	for i, line := range raw {
		row := gutterSt.Render(fmt.Sprintf("%*d", gutterW, i+1)) +
			gutterSt.Render(" │ ") +
			textSt.Render(truncateCells(line, textW))
		if pad := width - lipgloss.Width(row); pad > 0 {
			row += textSt.Width(pad).Render("")
		}
		out[i] = row
	}
	return out
}

func wrappedLines(content string, width int) []string {
	if width < 1 {
		return nil
	}
	st := lipgloss.NewStyle().Foreground(colorSand).Background(colorPanel).Width(width)
	if strings.TrimSpace(content) == "" {
		return []string{st.Render("")}
	}
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if line == "" {
			out = append(out, st.Render(""))
			continue
		}
		for line != "" {
			head, rest := cutWidth(line, width)
			if head == "" {
				break
			}
			out = append(out, st.Render(head))
			if rest == "" || rest == line {
				break
			}
			line = rest
		}
	}
	if len(out) == 0 {
		out = []string{st.Render("")}
	}
	return out
}
