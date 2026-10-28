package ui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fileMatch is one ranked file search result. pos holds the rune indexes that
// matched the query, so the list can highlight them.
type fileMatch struct {
	path  string
	pos   []int
	score int
}

// rankFiles returns the files that fuzzy-match query, best first. An empty
// query lists the files in their natural order.
func rankFiles(files []string, query string, limit int) []fileMatch {
	if limit < 1 {
		return nil
	}
	if query == "" {
		out := make([]fileMatch, 0, min(limit, len(files)))
		for _, path := range files {
			if len(out) == limit {
				break
			}
			out = append(out, fileMatch{path: path})
		}
		return out
	}
	q := []rune(strings.ToLower(query))
	matches := make([]fileMatch, 0, len(files))
	for _, path := range files {
		if pos, score, ok := fuzzyPath(path, q); ok {
			matches = append(matches, fileMatch{path: path, pos: pos, score: score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		if len(matches[i].path) != len(matches[j].path) {
			return len(matches[i].path) < len(matches[j].path)
		}
		return matches[i].path < matches[j].path
	})
	if len(matches) == 0 {
		return nil
	}
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

// fuzzyPath matches query as a case-insensitive subsequence of path and scores
// the result by consecutive runs, word starts, and basename hits.
func fuzzyPath(path string, query []rune) ([]int, int, bool) {
	runes := []rune(path)
	if len(query) == 0 || len(query) > len(runes) {
		return nil, 0, false
	}
	lower := make([]rune, len(runes))
	base := 0
	for i, r := range runes {
		lower[i] = unicode.ToLower(r)
		if r == '/' {
			base = i + 1
		}
	}
	pos := make([]int, 0, len(query))
	qi := 0
	last := -2
	score := 0
	for i, r := range lower {
		if qi >= len(query) {
			break
		}
		if r != query[qi] {
			continue
		}
		score += 10
		if i == last+1 {
			score += 12
		}
		if i == 0 || lower[i-1] == '/' || lower[i-1] == '-' || lower[i-1] == '_' || lower[i-1] == '.' {
			score += 14
		}
		if i >= base {
			score += 6
		}
		pos = append(pos, i)
		last = i
		qi++
	}
	if qi < len(query) {
		return nil, 0, false
	}
	if strings.Contains(string(lower), string(query)) {
		score += 40
	}
	if len(query) <= len(lower)-base && strings.EqualFold(string(runes[base:base+len(query)]), string(query)) {
		score += 24
	}
	score -= len(runes) / 8
	return pos, score, true
}

// openFinder shows the project-wide file search, loading the file list once.
func (m Model) openFinder() (tea.Model, tea.Cmd) {
	if m.previewEditing {
		return m, nil
	}
	m.finderOpen = true
	m.finderInput = ""
	m.finderCursor = 0
	m.finderInputCursor = 0
	m.finderScroll = 0
	m.finderMatches = nil
	m.finderErr = ""
	if !m.finderReady {
		if m.finderLoading {
			return m, nil
		}
		m.finderLoading = true
		return m, loadFilesListCmd(m.root)
	}
	return m.rankFinder(), nil
}

// rankFinder recomputes the visible results for the current input.
func (m Model) rankFinder() Model {
	m.finderMatches = rankFiles(m.finderFiles, m.finderInput, 200)
	if len(m.finderMatches) == 0 {
		m.finderCursor = 0
		m.finderScroll = 0
		return m
	}
	m.finderCursor = clamp(m.finderCursor, 0, len(m.finderMatches)-1)
	m.finderScroll = ensure(m.finderCursor, m.finderScroll, m.finderRows())
	return m
}

// finderRows is the number of result rows the modal can show.
func (m Model) finderRows() int {
	rows := 10
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

func (m Model) moveFinder(delta int) Model {
	if len(m.finderMatches) == 0 {
		return m
	}
	m.finderCursor = clamp(m.finderCursor+delta, 0, len(m.finderMatches)-1)
	m.finderScroll = ensure(m.finderCursor, m.finderScroll, m.finderRows())
	return m
}

// updateFinderInput re-ranks from the top after an edit.
func (m Model) updateFinderInput(text string, caret int) Model {
	m.finderInput = text
	m.finderInputCursor = caret
	m.finderCursor = 0
	m.finderScroll = 0
	return m.rankFinder()
}

func (m Model) onFinderKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.finderOpen = false
		return m, nil
	case tea.KeyEnter:
		return m.openFinderSelection()
	case tea.KeyUp:
		return m.moveFinder(-1), nil
	case tea.KeyDown:
		return m.moveFinder(1), nil
	case tea.KeyPgUp:
		return m.moveFinder(-m.finderRows()), nil
	case tea.KeyPgDown:
		return m.moveFinder(m.finderRows()), nil
	case tea.KeyLeft:
		m.finderCaret(-1)
		return m, nil
	case tea.KeyRight:
		m.finderCaret(1)
		return m, nil
	case tea.KeyHome:
		m.finderCaret(-len([]rune(m.finderInput)))
		return m, nil
	case tea.KeyEnd:
		m.finderCaret(len([]rune(m.finderInput)))
		return m, nil
	case tea.KeyBackspace:
		text, caret := deleteBefore(m.finderInput, m.finderInputCursor)
		return m.updateFinderInput(text, caret), nil
	case tea.KeyDelete:
		text, caret := deleteAt(m.finderInput, m.finderInputCursor)
		return m.updateFinderInput(text, caret), nil
	case tea.KeySpace:
		text, caret := insertText(m.finderInput, m.finderInputCursor, " ")
		return m.updateFinderInput(text, caret), nil
	case tea.KeyRunes:
		text, caret := insertText(m.finderInput, m.finderInputCursor, string(msg.Runes))
		return m.updateFinderInput(text, caret), nil
	}
	return m, nil
}

// finderCaret moves the input caret. The result selection has its own cursor.
func (m *Model) finderCaret(delta int) {
	m.finderInputCursor = clamp(m.finderInputCursor+delta, 0, len([]rune(m.finderInput)))
}

func (m Model) openFinderSelection() (tea.Model, tea.Cmd) {
	if len(m.finderMatches) == 0 {
		return m, nil
	}
	rel := m.finderMatches[clamp(m.finderCursor, 0, len(m.finderMatches)-1)].path
	m.finderOpen = false
	m.focus = PaneMain
	return m.openFile(rel, filepath.Base(rel)), nil
}

// finderView is the project file search modal.
func (m Model) finderView() string {
	innerW := m.width - 6
	if innerW > 72 {
		innerW = 72
	}
	if innerW < 24 {
		innerW = m.width - 2
	}
	if innerW < 8 {
		innerW = 8
	}
	rows := m.finderRows()
	list := make([]string, 0, rows)
	switch {
	case m.finderErr != "":
		list = append(list, noteLine(" "+m.finderErr, innerW))
	case m.finderLoading:
		list = append(list, noteLine(" Reading the project…", innerW))
	case len(m.finderMatches) == 0:
		if len(m.finderFiles) == 0 {
			list = append(list, noteLine(" No files in this project", innerW))
		} else {
			list = append(list, noteLine(" No files match "+m.finderInput, innerW))
		}
	default:
		start := ensure(m.finderCursor, m.finderScroll, rows)
		end := min(start+rows, len(m.finderMatches))
		for i := start; i < end; i++ {
			list = append(list, m.finderLine(m.finderMatches[i], innerW, i == m.finderCursor))
		}
	}
	for len(list) < rows {
		list = append(list, noteLine("", innerW))
	}

	body := []string{m.finderInputRow(innerW), sepLine(innerW)}
	body = append(body, list...)
	body = append(body, noteLine(" ↑↓ pick · enter open · esc close", innerW))

	count := "…"
	if !m.finderLoading && m.finderErr == "" {
		if len(m.finderMatches) == 0 {
			count = "0"
		} else {
			count = fmt.Sprintf("%d/%d", m.finderCursor+1, len(m.finderMatches))
		}
	}
	title := joinLeftRight(innerW, colorPanel,
		[]piece{pieceText(" ▸ Find file", colorLagoonBright, colorPanel)},
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

func (m Model) finderInputRow(width int) string {
	return searchInputWithRight(" > ", m.finderInput, m.finderInputCursor, width, colorSelectDim, "")
}

func (m Model) finderLine(match fileMatch, width int, selected bool) string {
	bg := lipgloss.TerminalColor(colorPanel)
	marker := "  "
	markerFG := lipgloss.TerminalColor(colorMoss)
	if selected {
		bg = colorSelect
		marker = "▸ "
		markerFG = colorLagoonBright
	}
	textW := width - 2
	if textW < 1 {
		textW = 1
	}
	path := searchPath(match.path, match.pos, textW, bg)
	row := lipgloss.NewStyle().Foreground(markerFG).Background(bg).Render(marker) + path
	if used := lipgloss.Width(row); used < width {
		row += lipgloss.NewStyle().Background(bg).Width(width - used).Render("")
	}
	return row
}

// searchPath draws a path with matched runes in the sun colour. Directories
// stay dim so the basename reads first.
func searchPath(path string, positions []int, width int, bg lipgloss.TerminalColor) string {
	if width < 1 {
		return lipgloss.NewStyle().Background(bg).Width(width).Render("")
	}
	runes := []rune(path)
	start := 0
	truncated := false
	if len(runes) > width {
		start = len(runes) - width + 1
		truncated = true
	}
	base := 0
	for i, r := range runes {
		if r == '/' {
			base = i + 1
		}
	}
	matched := make(map[int]struct{}, len(positions))
	for _, p := range positions {
		matched[p] = struct{}{}
	}
	var b strings.Builder
	used := 0
	if truncated {
		b.WriteString(lipgloss.NewStyle().Foreground(colorMoss).Background(bg).Render("…"))
		used++
	}
	var pending strings.Builder
	var pendingFG lipgloss.TerminalColor
	pendingBold := false
	flush := func() {
		if pending.Len() == 0 {
			return
		}
		style := lipgloss.NewStyle().Foreground(pendingFG).Background(bg)
		if pendingBold {
			style = style.Bold(true)
		}
		b.WriteString(style.Render(pending.String()))
		pending.Reset()
	}
	for i := start; i < len(runes) && used < width; i++ {
		fg := colorSand
		bold := false
		if i < base {
			fg = colorSandDim
		}
		if _, ok := matched[i]; ok {
			fg = colorSun
			bold = true
		}
		if pending.Len() == 0 {
			pendingFG = fg
			pendingBold = bold
		} else if fg != pendingFG || bold != pendingBold {
			flush()
			pendingFG = fg
			pendingBold = bold
		}
		pending.WriteRune(runes[i])
		used++
	}
	flush()
	if used < width {
		b.WriteString(lipgloss.NewStyle().Background(bg).Width(width - used).Render(""))
	}
	return b.String()
}

// insertText puts text at cursor and returns the new value and caret.
func insertText(value string, cursor int, text string) (string, int) {
	runes := []rune(value)
	cursor = clamp(cursor, 0, len(runes))
	typed := []rune(text)
	next := make([]rune, 0, len(runes)+len(typed))
	next = append(next, runes[:cursor]...)
	next = append(next, typed...)
	next = append(next, runes[cursor:]...)
	return string(next), cursor + len(typed)
}

// deleteBefore removes the rune before cursor.
func deleteBefore(value string, cursor int) (string, int) {
	runes := []rune(value)
	cursor = clamp(cursor, 0, len(runes))
	if cursor == 0 {
		return value, cursor
	}
	next := make([]rune, 0, len(runes)-1)
	next = append(next, runes[:cursor-1]...)
	next = append(next, runes[cursor:]...)
	return string(next), cursor - 1
}

// deleteAt removes the rune at cursor.
func deleteAt(value string, cursor int) (string, int) {
	runes := []rune(value)
	cursor = clamp(cursor, 0, len(runes))
	if cursor >= len(runes) {
		return value, cursor
	}
	next := make([]rune, 0, len(runes)-1)
	next = append(next, runes[:cursor]...)
	next = append(next, runes[cursor+1:]...)
	return string(next), cursor
}
