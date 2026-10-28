package ui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/StarkDroid/Bento-Terminal/internal/syntax"
)

// previewRow is one rendered source row and the source line it shows.
type previewRow struct {
	line int
	text string
}

type codeSpan struct {
	text   string
	fg     lipgloss.TerminalColor
	bold   bool
	italic bool
}

// buildCode renders the current file preview and returns its rows plus the
// source lines that are visible after folding.
func (m Model) buildCode() ([]previewRow, []int) {
	return renderCode(m.sourceLines(), m.previewSpans, m.previewBlocks, m.previewFolds, m.previewCursor, m.previewCol, m.preview.Width, m.focus == PaneMain, m.findMatches, m.findIndex)
}

// refreshSpans tokenises the preview once, so resizing and folding only redo
// layout.
func (m Model) refreshSpans() Model {
	m.previewSpans = nil
	if !m.previewNumbered || !m.previewSyntax || !m.previewLang.Detected() {
		return m
	}
	raw, ok := syntax.Highlight(m.previewLang.Name, m.sourceLines())
	if !ok {
		return m
	}
	m.previewSpans = make([][]codeSpan, len(raw))
	for i, line := range raw {
		for _, span := range line {
			fg := lipgloss.TerminalColor(colorSand)
			if span.Colour != "" {
				fg = lipgloss.Color(span.Colour)
			}
			m.previewSpans[i] = append(m.previewSpans[i], codeSpan{
				text:   span.Text,
				fg:     fg,
				bold:   span.Bold,
				italic: span.Italic,
			})
		}
	}
	return m
}

func renderCode(source []string, spans [][]codeSpan, blocks []syntax.Block, folds map[int]bool, cursor, col, width int, active bool, matches []findMatch, current int) ([]previewRow, []int) {
	if width < 1 || len(source) == 0 {
		return nil, nil
	}
	byStart := make(map[int]syntax.Block, len(blocks))
	for _, block := range blocks {
		byStart[block.Start] = block
	}
	gutterW := len(strconv.Itoa(len(source)))
	textW := width - gutterW - 3 - 2
	if textW < 1 {
		return renderCodeNarrow(source, spans, blocks, folds, cursor, width, active, matches, current)
	}
	rows := make([]previewRow, 0, len(source))
	visible := make([]int, 0, len(source))
	for i := 0; i < len(source); i++ {
		bg := lipgloss.TerminalColor(colorPanel)
		cursorCell := -1
		if active && i == cursor {
			bg = colorSelectDim
			cursorCell = cellOffset(source[i], col)
		}
		block, foldable := byStart[i]
		collapsed := foldable && folds[i]

		var b strings.Builder
		gutter := lipgloss.NewStyle().
			Foreground(colorMoss).
			Background(bg).
			Width(gutterW).
			Align(lipgloss.Right).
			Render(strconv.Itoa(i + 1))
		b.WriteString(gutter)
		b.WriteString(lipgloss.NewStyle().Foreground(colorMoss).Background(bg).Render(" │ "))
		marker, markerFG := "  ", colorMoss
		if foldable {
			if collapsed {
				marker, markerFG = "▸ ", colorSun
			} else {
				marker, markerFG = "▾ ", colorLagoon
			}
		} else if active && i == cursor {
			marker, markerFG = "▌ ", colorLagoonBright
		}
		b.WriteString(lipgloss.NewStyle().Foreground(markerFG).Background(bg).Render(marker))
		lineSpans := spanFor(source, spans, i)
		code, used := renderSpans(lineSpans, textW, bg, cursorCell, lineHighlights(matches, current, i, source[i]))
		b.WriteString(code)
		rowWidth := gutterW + 3 + 2 + used
		if collapsed {
			summary := fmt.Sprintf("  … %d lines", block.End-block.Start)
			avail := width - rowWidth
			if avail > 0 {
				summary = truncateCells(summary, avail)
				b.WriteString(lipgloss.NewStyle().Foreground(colorMoss).Background(bg).Render(summary))
				rowWidth += lipgloss.Width(summary)
			}
		}
		if rowWidth < width {
			b.WriteString(lipgloss.NewStyle().Background(bg).Width(width - rowWidth).Render(""))
		}
		rows = append(rows, previewRow{line: i, text: b.String()})
		visible = append(visible, i)
		if collapsed {
			i = block.End
		}
	}
	return rows, visible
}

// renderCodeNarrow is the very narrow fallback: line numbers and text only.
func renderCodeNarrow(source []string, spans [][]codeSpan, blocks []syntax.Block, folds map[int]bool, cursor, width int, active bool, matches []findMatch, current int) ([]previewRow, []int) {
	rows := make([]previewRow, 0, len(source))
	visible := make([]int, 0, len(source))
	byStart := make(map[int]syntax.Block, len(blocks))
	for _, block := range blocks {
		byStart[block.Start] = block
	}
	for i := 0; i < len(source); i++ {
		bg := lipgloss.TerminalColor(colorPanel)
		if active && i == cursor {
			bg = colorSelectDim
		}
		lineSpans := spanFor(source, spans, i)
		code, used := renderSpans(lineSpans, width, bg, -1, lineHighlights(matches, current, i, source[i]))
		if used < width {
			code += lipgloss.NewStyle().Background(bg).Width(width - used).Render("")
		}
		rows = append(rows, previewRow{line: i, text: code})
		visible = append(visible, i)
		if block, ok := byStart[i]; ok && folds[i] {
			i = block.End
		}
	}
	return rows, visible
}

func spanFor(source []string, spans [][]codeSpan, i int) []codeSpan {
	if i < len(spans) && len(spans[i]) > 0 {
		return spans[i]
	}
	return []codeSpan{{text: source[i], fg: colorSand}}
}

// codeHighlight is a cell range on one rendered line. current marks the match
// the find bar is sitting on.
type codeHighlight struct {
	start   int
	end     int
	current bool
}

// lineHighlights converts this line's find matches into cell ranges.
func lineHighlights(matches []findMatch, current, line int, text string) []codeHighlight {
	if len(matches) == 0 {
		return nil
	}
	var out []codeHighlight
	for i, match := range matches {
		if match.line != line {
			continue
		}
		start := cellOffset(text, match.start)
		end := cellOffset(text, match.end)
		if end <= start {
			continue
		}
		out = append(out, codeHighlight{start: start, end: end, current: i == current})
	}
	return out
}

type spanStyle struct {
	fg     lipgloss.TerminalColor
	bg     lipgloss.TerminalColor
	bold   bool
	italic bool
}

// renderSpans draws styled spans, expanding tabs and cutting at limit cells.
// A non-negative cursor cell is drawn as a block caret; find highlights sit
// under the caret.
func renderSpans(spans []codeSpan, limit int, bg lipgloss.TerminalColor, cursor int, highlights []codeHighlight) (string, int) {
	if limit < 1 {
		return "", 0
	}
	type styledRune struct {
		text   string
		fg     lipgloss.TerminalColor
		bold   bool
		italic bool
	}
	var runes []styledRune
	for _, span := range spans {
		text := strings.ReplaceAll(span.text, "\t", "    ")
		for _, r := range text {
			runes = append(runes, styledRune{text: string(r), fg: span.fg, bold: span.bold, italic: span.italic})
		}
	}
	var b strings.Builder
	used := 0
	caretDone := cursor < 0
	var pending strings.Builder
	var key spanStyle
	flush := func() {
		if pending.Len() == 0 {
			return
		}
		style := lipgloss.NewStyle().Foreground(key.fg).Background(key.bg)
		if key.bold {
			style = style.Bold(true)
		}
		if key.italic {
			style = style.Italic(true)
		}
		b.WriteString(style.Render(pending.String()))
		pending.Reset()
	}
	add := func(text string, next spanStyle) {
		if pending.Len() > 0 && next != key {
			flush()
		}
		key = next
		pending.WriteString(text)
	}
	for _, r := range runes {
		cells := lipgloss.Width(r.text)
		if cells < 1 {
			cells = 1
		}
		if used+cells > limit {
			break
		}
		next := spanStyle{fg: r.fg, bg: bg, bold: r.bold, italic: r.italic}
		for _, h := range highlights {
			if used < h.start || used >= h.end {
				continue
			}
			next.fg = colorOcean
			next.bold = h.current
			next.italic = false
			if h.current {
				next.bg = colorLagoonBright
			} else {
				next.bg = colorSun
			}
		}
		if !caretDone && used == cursor {
			next.fg = colorOcean
			next.bg = colorLagoonBright
			next.bold = true
			next.italic = false
			caretDone = true
		}
		add(r.text, next)
		used += cells
	}
	flush()
	if !caretDone && cursor >= used && cursor < limit {
		b.WriteString(lipgloss.NewStyle().Foreground(colorOcean).Background(colorLagoonBright).Bold(true).Render(" "))
		used++
	}
	return b.String(), used
}

// writeStyled draws as much of text as fits and returns what is left.
func writeStyled(b *strings.Builder, style lipgloss.Style, text string, limit, used int) (int, string) {
	if text == "" || used >= limit {
		return used, ""
	}
	head, rest := cutWidth(text, limit-used)
	if head == "" {
		return used, rest
	}
	b.WriteString(style.Render(head))
	return used + lipgloss.Width(head), rest
}

// cellOffset returns the screen cell of a rune column, with tabs expanded.
func cellOffset(line string, col int) int {
	runes := []rune(line)
	col = clamp(col, 0, len(runes))
	used := 0
	for _, r := range runes[:col] {
		used += runeCells(r)
	}
	return used
}

func runeCells(r rune) int {
	if r == '\t' {
		return 4
	}
	w := lipgloss.Width(string(r))
	if w < 1 {
		return 1
	}
	return w
}

// splitAtCell cuts text at the rune covering target, splitting it into the
// text before the rune, the rune itself, and the text after it.
func splitAtCell(text string, target int) (head, mid, tail string) {
	used := 0
	for i, r := range text {
		w := runeCells(r)
		if target < used+w {
			_, size := utf8.DecodeRuneInString(text[i:])
			return text[:i], text[i : i+size], text[i+size:]
		}
		used += w
	}
	return text, "", ""
}
