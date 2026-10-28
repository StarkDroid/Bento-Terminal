package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// findMatch is one occurrence of the find text, in rune columns of a line.
type findMatch struct {
	line  int
	start int
	end   int
}

// findInLines returns every case-insensitive, non-overlapping occurrence.
func findInLines(lines []string, query string) []findMatch {
	q := []rune(query)
	if len(q) == 0 {
		return nil
	}
	var out []findMatch
	for i, line := range lines {
		runes := []rune(line)
		for j := 0; j+len(q) <= len(runes); {
			if strings.EqualFold(string(runes[j:j+len(q)]), query) {
				out = append(out, findMatch{line: i, start: j, end: j + len(q)})
				j += len(q)
				continue
			}
			j++
		}
	}
	return out
}

// openFind shows the find bar over a text file.
func (m Model) openFind() Model {
	if m.previewKind != previewFile || !m.previewNumbered {
		return m
	}
	m.findOpen = true
	m.findInput = ""
	m.findCursor = 0
	m.findMatches = nil
	m.findIndex = 0
	m.focus = PaneMain
	return m.reflowPreview(false)
}

// closeFind hides the bar and drops the match highlights, leaving the caret
// where the current match was.
func (m Model) closeFind() Model {
	m.findOpen = false
	m.findInput = ""
	m.findCursor = 0
	m.findMatches = nil
	m.findIndex = 0
	return m.reflowPreview(false)
}

func (m Model) onFindKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		return m.closeFind(), nil
	case tea.KeyEnter, tea.KeyDown:
		return m.findStep(1), nil
	case tea.KeyUp:
		return m.findStep(-1), nil
	case tea.KeyLeft:
		m.findCaret(-1)
		return m, nil
	case tea.KeyRight:
		m.findCaret(1)
		return m, nil
	case tea.KeyHome:
		m.findCaret(-len([]rune(m.findInput)))
		return m, nil
	case tea.KeyEnd:
		m.findCaret(len([]rune(m.findInput)))
		return m, nil
	case tea.KeyBackspace:
		text, caret := deleteBefore(m.findInput, m.findCursor)
		return m.updateFindInput(text, caret), nil
	case tea.KeyDelete:
		text, caret := deleteAt(m.findInput, m.findCursor)
		return m.updateFindInput(text, caret), nil
	case tea.KeySpace:
		text, caret := insertText(m.findInput, m.findCursor, " ")
		return m.updateFindInput(text, caret), nil
	case tea.KeyRunes:
		text, caret := insertText(m.findInput, m.findCursor, string(msg.Runes))
		return m.updateFindInput(text, caret), nil
	}
	return m, nil
}

func (m *Model) findCaret(delta int) {
	m.findCursor = clamp(m.findCursor+delta, 0, len([]rune(m.findInput)))
}

// updateFindInput recomputes matches from the current preview buffer and jumps
// to the first one.
func (m Model) updateFindInput(text string, caret int) Model {
	m.findInput = text
	m.findCursor = caret
	m.findMatches = findInLines(m.sourceLines(), m.findInput)
	m.findIndex = 0
	if len(m.findMatches) > 0 {
		return m.jumpToFind(0)
	}
	return m.reflowPreview(false)
}

// findStep moves through matches with wrap-around.
func (m Model) findStep(delta int) Model {
	if len(m.findMatches) == 0 {
		return m
	}
	return m.jumpToFind(m.findIndex + delta)
}

// jumpToFind moves the caret to a match and unfolds the block hiding it.
func (m Model) jumpToFind(index int) Model {
	n := len(m.findMatches)
	if n == 0 {
		return m
	}
	index = ((index % n) + n) % n
	match := m.findMatches[index]
	m.findIndex = index
	m.previewCursor = match.line
	m.previewCol = match.start
	for _, block := range m.previewBlocks {
		if match.line > block.Start && match.line <= block.End {
			delete(m.previewFolds, block.Start)
		}
	}
	m = m.reflowPreview(false)
	return m.scrollPreview()
}

// findRow is the one-line find prompt drawn above the code.
func (m Model) findRow(width int) string {
	if width < 1 {
		return ""
	}
	count := "0"
	if len(m.findMatches) > 0 {
		count = fmt.Sprintf("%d/%d", m.findIndex+1, len(m.findMatches))
	}
	return searchInputWithRight(" > ", m.findInput, m.findCursor, width, colorSelectDim, count)
}

// searchInputWithRight draws a prompt and caret, then a right-aligned label.
func searchInputWithRight(prefix, text string, cursor, width int, bg lipgloss.TerminalColor, right string) string {
	if width < 1 {
		return ""
	}
	rightW := 0
	if right != "" {
		rightW = lipgloss.Width(right) + 2
	}
	runes := []rune(text)
	cursor = clamp(cursor, 0, len(runes))
	prefixW := lipgloss.Width(prefix)
	avail := width - prefixW - rightW - 1
	if avail < 1 {
		return lipgloss.NewStyle().Background(bg).Width(width).Render("")
	}
	start := 0
	if cursor > avail-1 {
		start = cursor - avail + 1
	}
	end := min(start+avail, len(runes))
	before := runes[start:cursor]
	tail := runes[cursor:end]
	caret := " "
	after := tail
	if len(tail) > 0 {
		caret = string(tail[0])
		after = tail[1:]
	}
	row := lipgloss.NewStyle().Foreground(colorLagoonBright).Background(bg).Render(prefix) +
		lipgloss.NewStyle().Foreground(colorSand).Background(bg).Render(string(before)) +
		lipgloss.NewStyle().Foreground(colorOcean).Background(colorLagoonBright).Render(caret) +
		lipgloss.NewStyle().Foreground(colorSand).Background(bg).Render(string(after))
	if right != "" {
		used := lipgloss.Width(row)
		if used+rightW < width {
			row += lipgloss.NewStyle().Background(bg).Width(width - used - rightW).Render("")
		}
		row += lipgloss.NewStyle().Foreground(colorLagoonBright).Background(bg).Render(right + " ")
	}
	if used := lipgloss.Width(row); used < width {
		row += lipgloss.NewStyle().Background(bg).Width(width - used).Render("")
	}
	return row
}
