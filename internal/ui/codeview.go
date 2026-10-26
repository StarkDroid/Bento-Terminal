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
	source := fileLines(m.previewRaw)
	return renderCode(source, m.previewSpans, m.previewBlocks, m.previewFolds, m.previewCursor, m.previewCol, m.preview.Width, m.focus == PaneMain)
}

// refreshSpans tokenises the preview once, so resizing and folding only redo
// layout.
func (m Model) refreshSpans() Model {
	m.previewSpans = nil
	if !m.previewNumbered || !m.previewSyntax || !m.previewLang.Detected() {
		return m
	}
	raw, ok := syntax.Highlight(m.previewLang.Name, fileLines(m.previewRaw))
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

func renderCode(source []string, spans [][]codeSpan, blocks []syntax.Block, folds map[int]bool, cursor, col, width int, active bool) ([]previewRow, []int) {
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
		return renderCodeNarrow(source, spans, blocks, folds, cursor, width, active)
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
		code, used := renderSpans(lineSpans, textW, bg, cursorCell)
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
func renderCodeNarrow(source []string, spans [][]codeSpan, blocks []syntax.Block, folds map[int]bool, cursor, width int, active bool) ([]previewRow, []int) {
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
		code, used := renderSpans(lineSpans, width, bg, -1)
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

// renderSpans draws styled spans, expanding tabs and cutting at limit cells.
// A non-negative cursor cell is drawn as a block caret.
func renderSpans(spans []codeSpan, limit int, bg lipgloss.TerminalColor, cursor int) (string, int) {
	if limit < 1 {
		return "", 0
	}
	var b strings.Builder
	used := 0
	caretDone := cursor < 0
	caretStyle := lipgloss.NewStyle().Foreground(colorOcean).Background(colorLagoonBright).Bold(true)
	for _, span := range spans {
		if used >= limit {
			break
		}
		text := strings.ReplaceAll(span.text, "\t", "    ")
		if text == "" {
			continue
		}
		style := lipgloss.NewStyle().Foreground(span.fg).Background(bg)
		if span.bold {
			style = style.Bold(true)
		}
		if span.italic {
			style = style.Italic(true)
		}
		for text != "" && used < limit {
			off := cursor - used
			if caretDone || off < 0 || off >= lipgloss.Width(text) {
				used, text = writeStyled(&b, style, text, limit, used)
				break
			}
			head, mid, tail := splitAtCell(text, off)
			if head != "" {
				b.WriteString(style.Render(head))
				used += lipgloss.Width(head)
			}
			if mid != "" && used+lipgloss.Width(mid) <= limit {
				b.WriteString(caretStyle.Render(mid))
				used += lipgloss.Width(mid)
				caretDone = true
			}
			text = tail
		}
	}
	if !caretDone && cursor >= used && cursor < limit {
		b.WriteString(caretStyle.Render(" "))
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
