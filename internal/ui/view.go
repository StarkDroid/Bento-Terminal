package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
	"github.com/StarkDroid/Bento-Terminal/internal/files"
	"github.com/StarkDroid/Bento-Terminal/internal/git"
)

func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "Bento terminal"
	}
	if m.updateOpen {
		return m.updateView()
	}
	if m.newFileOpen {
		return m.newFileView()
	}
	if m.fileOpOpen {
		return m.fileOpView()
	}
	if m.finderOpen {
		return m.finderView()
	}
	if m.recentOpen {
		return m.recentView()
	}
	if m.paletteOpen {
		return m.paletteView()
	}
	if m.settingsOpen {
		return m.settingsView()
	}
	lay := computeLayout(m.width, m.height)
	if !lay.OK {
		return m.tooSmall()
	}
	previewW, aiW := lay.Main.W, 0
	if m.aiOpen {
		previewW, aiW = lay.MainSplit()
	}
	main := m.renderMain(previewW, lay.Main.H)
	right := lipgloss.JoinVertical(lipgloss.Left,
		m.renderFiles(lay.Files.W, lay.Files.H),
		gapRow(lay.Files.W),
		m.renderHistory(lay.History.W, lay.History.H),
	)
	topParts := []string{main}
	if m.aiOpen {
		topParts = append(topParts, gapCol(lay.Main.H), m.renderAI(aiW, lay.Main.H))
	}
	topParts = append(topParts, gapCol(lay.Main.H), right)
	top := lipgloss.JoinHorizontal(lipgloss.Top, topParts...)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderBuild(lay.Build.W, lay.Build.H),
		gapCol(lay.Build.H),
		m.renderTerminal(lay.Terminal.W, lay.Build.H),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, gapRow(m.width), bottom, m.renderStatus(lay.Status.W))
}

func (m Model) renderMain(width, height int) string {
	focused := m.focus == PaneMain
	innerW := width - 2
	bodyH := height - 4
	if innerW < 1 {
		innerW = 1
	}
	if bodyH < 0 {
		bodyH = 0
	}
	title := paintTitle(" "+filepath.Base(m.root), innerW, focused)
	var body string
	switch m.previewKind {
	case previewFile:
		title = m.fileTitle(innerW, focused)
		if m.findOpen && bodyH > 1 {
			body = m.findRow(innerW) + "\n" + windowLines(m.previewLines, m.preview.YOffset, innerW, bodyH-1)
		} else if m.findOpen && bodyH == 1 {
			body = m.findRow(innerW)
		} else {
			body = windowLines(m.previewLines, m.preview.YOffset, innerW, bodyH)
		}
	case previewCommit:
		title = paintTitle(" ⎇ "+m.previewTitle, innerW, focused)
		body = windowLines(m.previewLines, m.preview.YOffset, innerW, bodyH)
	case previewDiff:
		title = paintTitle(" ± "+m.previewTitle, innerW, focused)
		body = windowLines(m.previewLines, m.preview.YOffset, innerW, bodyH)
	default:
		body = m.homeBody(innerW, bodyH)
	}
	if bodyH > 0 {
		body += "\n"
	}
	body += m.footer(PaneMain, innerW)
	return renderPanel(title, body, width, height, focused)
}

// fileTab draws the open file as the active tab: icon, name, language badge,
// and the save marker. Its width does not depend on focus, so click zones can
// measure it.
func (m Model) fileTab(focused bool) piece {
	fg, bg := tabColors(true, focused)
	text := func(s string, color lipgloss.TerminalColor) piece {
		st := lipgloss.NewStyle().Foreground(color).Background(bg)
		if focused {
			st = st.Bold(true)
		}
		return piece{view: st.Render(s), width: lipgloss.Width(s)}
	}
	parts := []piece{
		text("≡", fg),
		text(" "+m.previewTitle, fg),
	}
	if m.previewLang.Detected() {
		lang := m.previewLang.Name
		if !m.previewSyntax {
			lang += " off"
		}
		parts = append(parts, text("  ·  "+lang, colorLagoonBright))
	}
	switch {
	case m.previewSaveErr:
		parts = append(parts, text("  save failed", colorCoral))
	case m.previewDirty:
		parts = append(parts, text("  ●", colorCoral))
	case m.previewSaved:
		parts = append(parts, text("  ✓", colorPalm))
	}
	if m.previewEditing {
		parts = append(parts, text("  edit", colorSun))
	}
	return chip(bg, parts)
}

// newFileTab draws the new-file button as an inactive tab sitting directly
// after the open file's tab, the way a browser keeps its new-tab button.
func (m Model) newFileTab(focused bool) piece {
	_, bg := tabColors(false, focused)
	label := lipgloss.TerminalColor(colorSandDim)
	if focused {
		label = colorSand
	}
	return chip(bg, []piece{
		pieceText("+", colorLagoonBright, bg),
		pieceText(" new file", label, bg),
	})
}

// fileTitle draws the open file's tab and the new-file tab beside it.
func (m Model) fileTitle(width int, focused bool) string {
	parts := []piece{m.fileTab(focused)}
	if !m.previewEditing {
		parts = append(parts, m.newFileTab(focused))
	}
	return joinPieces(width, colorPanel, parts)
}

func (m Model) homeBody(width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}
	name := filepath.Base(m.root)
	var branch string
	switch {
	case !m.gitReady:
		branch = "⎇ …"
	case !m.gitSnap.Repo && m.gitSnap.Err != "":
		branch = "⎇ " + m.gitSnap.Err
	case !m.gitSnap.Repo:
		branch = "⎇ not a git repository"
	default:
		b := m.gitSnap.Status.Branch
		if b == "" {
			b = "HEAD"
		}
		branch = "⎇ " + b
	}
	var systems string
	switch {
	case !m.buildReady:
		systems = "Detecting build…"
	case len(m.systems) == 0:
		systems = "No build system detected"
	default:
		systems = strings.Join(m.systems, " · ")
	}
	hint := "Select a file to preview it."
	lines := []string{
		lipgloss.NewStyle().Foreground(colorSand).Background(colorPanel).Bold(true).Render(truncateCells(name, width)),
		lipgloss.NewStyle().Foreground(colorLagoonBright).Background(colorPanel).Render(truncateCells(branch, width)),
		lipgloss.NewStyle().Foreground(colorPalm).Background(colorPanel).Render(truncateCells(systems, width)),
		lipgloss.NewStyle().Background(colorPanel).Width(1).Render(""),
		lipgloss.NewStyle().Foreground(colorSandDim).Background(colorPanel).Render(truncateCells(hint, width)),
	}
	block := lipgloss.JoinVertical(lipgloss.Center, lines...)
	placed := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorPanel),
	)
	return fitBlock(placed, width, height)
}

func (m Model) renderFiles(width, height int) string {
	focused := m.focus == PaneFiles
	bodyW := width - 2
	bodyH := height - 4
	if bodyW < 1 {
		bodyW = 1
	}
	if bodyH < 0 {
		bodyH = 0
	}
	var lines []string
	switch {
	case !m.treeReady:
		lines = []string{noteLine(" Reading directory…", bodyW)}
	case m.treeErr != "":
		lines = []string{noteLine(" "+m.treeErr, bodyW)}
	case m.tree == nil || len(m.tree.Rows()) == 0:
		lines = []string{noteLine(" Empty directory", bodyW)}
	default:
		rows := m.tree.Rows()
		scroll := ensure(m.fileCursor, m.fileScroll, bodyH)
		end := scroll + bodyH
		if end > len(rows) {
			end = len(rows)
		}
		for i := scroll; i < end; i++ {
			lines = append(lines, m.fileLine(rows[i], bodyW, i == m.fileCursor))
		}
	}
	for len(lines) < bodyH {
		lines = append(lines, noteLine("", bodyW))
	}
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	lines = append(lines, m.footer(PaneFiles, bodyW))
	return panel("≡ Files", strings.Join(lines, "\n"), width, height, focused)
}

func (m Model) fileLine(row files.Row, width int, selected bool) string {
	icon := "≡"
	if row.IsDir {
		if row.Expanded {
			icon = "▾"
		} else {
			icon = "▸"
		}
	}
	label := strings.Repeat("  ", row.Depth) + icon + " " + row.Name
	if m.pathDirty(row) {
		label += " ●"
	}
	return paintRow(" "+label, width, selected, m.focus == PaneFiles)
}

func (m Model) pathDirty(row files.Row) bool {
	dirty := m.gitSnap.Status.Dirty
	if len(dirty) == 0 {
		return false
	}
	if _, ok := dirty[row.Rel]; ok {
		return true
	}
	if !row.IsDir {
		return false
	}
	prefix := row.Rel + "/"
	for path := range dirty {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func (m Model) renderHistory(width, height int) string {
	focused := m.focus == PaneHistory
	bodyW := width - 2
	bodyH := height - 4
	if bodyW < 1 {
		bodyW = 1
	}
	if bodyH < 0 {
		bodyH = 0
	}
	var lines []string
	switch {
	case !m.gitReady:
		lines = []string{noteLine(" Reading history…", bodyW)}
	case !m.gitSnap.Repo && m.gitSnap.Err != "":
		lines = []string{noteLine(" "+m.gitSnap.Err, bodyW)}
	case !m.gitSnap.Repo:
		lines = []string{noteLine(" Not a git repository", bodyW)}
	case m.historyTab == histTabChanges:
		lines = m.changeLines(bodyW, bodyH)
	case len(m.gitSnap.Commits) == 0:
		lines = []string{noteLine(" No commits on this branch", bodyW)}
	default:
		scroll := ensure(m.commitCursor, m.commitScroll, bodyH)
		end := scroll + bodyH
		if end > len(m.gitSnap.Commits) {
			end = len(m.gitSnap.Commits)
		}
		for i := scroll; i < end; i++ {
			lines = append(lines, m.commitLine(m.gitSnap.Commits[i], bodyW, i == m.commitCursor))
		}
	}
	for len(lines) < bodyH {
		lines = append(lines, noteLine("", bodyW))
	}
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	lines = append(lines, m.footer(PaneHistory, bodyW))
	return renderPanel(m.historyTitle(bodyW), strings.Join(lines, "\n"), width, height, focused)
}

// historyTitle draws the History and Changes tabs, with a change count on the
// right while Changes is visible.
func (m Model) historyTitle(width int) string {
	focused := m.focus == PaneHistory
	left := []piece{
		pieceText(" ⎇ ", colorLagoonBright, colorPanel),
		tabPiece("History", m.historyTab == histTabCommits, focused),
		pieceText(" ", colorSand, colorPanel),
		tabPiece("Changes", m.historyTab == histTabChanges, focused),
	}
	var right []piece
	if n := len(m.gitSnap.Status.Files); n > 0 && m.gitSnap.Repo && m.historyTab == histTabChanges {
		right = append(right, pieceText(fmt.Sprintf(" %d ", n), colorCoral, colorPanel))
	}
	return joinLeftRight(width, colorPanel, left, right)
}

// changeLines lists the working-tree changes, or a note when the tree is clean.
func (m Model) changeLines(width, height int) []string {
	changes := m.gitSnap.Status.Files
	if len(changes) == 0 {
		return []string{noteLine(" Working tree clean", width)}
	}
	scroll := ensure(m.changeCursor, m.changeScroll, height)
	end := scroll + height
	if end > len(changes) {
		end = len(changes)
	}
	var lines []string
	for i := scroll; i < end; i++ {
		lines = append(lines, m.changeLine(changes[i], width, i == m.changeCursor))
	}
	return lines
}

// changeLine shows the porcelain code and path, with renames rendered as
// "old → new".
func (m Model) changeLine(f git.FileStatus, width int, selected bool) string {
	if width < 1 {
		return ""
	}
	bg := lipgloss.TerminalColor(colorPanel)
	row := lipgloss.NewStyle().Width(width)
	switch {
	case selected && m.focus == PaneHistory:
		bg = colorSelect
		row = row.Bold(true)
	case selected:
		bg = colorSelectDim
	}
	codeFG := colorAmber
	switch {
	case f.Untracked:
		codeFG = colorPalm
	case f.Unmerged:
		codeFG = colorCoral
	case f.Staged() && f.Unstaged():
		codeFG = colorCoral
	case f.Staged():
		codeFG = colorLagoonBright
	}
	code := lipgloss.NewStyle().Foreground(codeFG).Background(bg).Render(" " + f.Code() + " ")
	name := f.Path
	if f.OrigPath != "" {
		name = f.OrigPath + " → " + f.Path
	}
	textW := width - 4
	if textW < 1 {
		textW = 1
	}
	text := lipgloss.NewStyle().
		Foreground(colorSand).
		Background(bg).
		Width(textW).
		Render(truncateCells(name, textW))
	return row.Background(bg).Render(code + text)
}

func (m Model) commitLine(c git.Commit, width int, selected bool) string {
	hash := c.Hash
	if len(hash) > 7 {
		hash = hash[:7]
	}
	date := git.FormatWhen(c.When, time.Now())
	meta := lipgloss.Width(hash) + lipgloss.Width(date) + 4
	subjW := width - 1 - meta
	if subjW < 4 {
		return paintRow(" "+hash+" "+c.Subject, width, selected, m.focus == PaneHistory)
	}
	subject := truncateCells(c.Subject, subjW)
	gap := subjW - lipgloss.Width(subject)
	if gap < 0 {
		gap = 0
	}
	label := " " + hash + "  " + subject + strings.Repeat(" ", gap) + "  " + date
	return paintRow(label, width, selected, m.focus == PaneHistory)
}

func (m Model) renderBuild(width, height int) string {
	focused := m.focus == PaneBuild
	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}
	bodyH := height - 4
	if bodyH < 0 {
		bodyH = 0
	}
	var body []string
	if m.buildTab == buildTabOutput {
		if block := m.outputBlock(innerW, bodyH); block != "" {
			body = strings.Split(block, "\n")
		}
	} else {
		cmds := m.commands()
		switch {
		case !m.buildReady:
			body = append(body, noteLine(" Detecting commands…", innerW))
		case len(cmds) == 0:
			msg := " No commands detected"
			if m.mode == build.ModeManual {
				msg = " No run commands"
			}
			body = append(body, noteLine(msg, innerW))
			other := m.manualCmds
			if m.mode == build.ModeManual {
				other = m.autoCmds
			}
			if len(other) > 0 {
				body = append(body, noteLine(" Press a to switch mode", innerW))
			}
		default:
			scroll := ensure(m.cmdCursor, m.cmdScroll, bodyH)
			end := scroll + bodyH
			if end > len(cmds) {
				end = len(cmds)
			}
			for i := scroll; i < end; i++ {
				body = append(body, m.commandLine(cmds[i], innerW, i == m.cmdCursor))
			}
		}
	}
	for len(body) < bodyH {
		body = append(body, noteLine("", innerW))
	}
	if len(body) > bodyH {
		body = body[:bodyH]
	}
	if bodyH > 0 {
		body = append(body, m.footer(PaneBuild, innerW))
	}
	return renderPanel(m.buildTitle(innerW), strings.Join(body, "\n"), width, height, focused)
}

func (m Model) commandLine(c build.Command, width int, selected bool) string {
	label := " ▶ " + c.Name
	if c.Source != "" && lipgloss.Width(label)+2+lipgloss.Width(c.Source) <= width {
		gap := width - lipgloss.Width(label) - lipgloss.Width(c.Source)
		if gap < 1 {
			gap = 1
		}
		label = label + strings.Repeat(" ", gap) + c.Source
	}
	return paintRow(label, width, selected, m.focus == PaneBuild)
}

func (m Model) buildTitle(width int) string {
	focused := m.focus == PaneBuild
	left := []piece{
		m.runGlyph(m.buildState, focused),
		pieceText(" Build ", colorSandDim, colorPanel),
		tabPiece("Commands", m.buildTab == buildTabCommands, focused),
		pieceText(" ", colorSand, colorPanel),
		tabPiece("Output", m.buildTab == buildTabOutput, focused),
	}
	var right []piece
	mode := "auto"
	if m.mode == build.ModeManual {
		mode = "manual"
	}
	right = append(right, pieceText(" "+mode, colorLagoon, colorPanel))
	if m.last != nil {
		right = append(right, pieceText("  ·  "+m.last.Name, colorSandDim, colorPanel))
	}
	return joinLeftRight(width, colorPanel, left, right)
}

// buildTabAt maps an inner title cell to a build tab.
func (m Model) buildTabAt(rel int) (buildTab, bool) {
	start := m.runGlyph(m.buildState, false).width + lipgloss.Width(" Build ")
	if rel >= start && rel < start+lipgloss.Width(" Commands ") {
		return buildTabCommands, true
	}
	start += lipgloss.Width(" Commands ") + 1
	if rel >= start && rel < start+lipgloss.Width(" Output ") {
		return buildTabOutput, true
	}
	return 0, false
}

// runGlyph is the state marker in a panel title: a spinner while running, a
// tick or cross after a run, and the panel glyph when idle.
func (m Model) runGlyph(state runState, focused bool) piece {
	switch state {
	case stateRunning:
		spin := m.spinner.View()
		return piece{
			view:  lipgloss.NewStyle().Background(colorPanel).Render(" ") + spin + lipgloss.NewStyle().Background(colorPanel).Render(" "),
			width: 2 + lipgloss.Width(spin),
		}
	case stateOK:
		return pieceText(" ✓ ", colorPalm, colorPanel)
	case stateFailed, stateCanceled:
		return pieceText(" ✗ ", colorCoral, colorPanel)
	}
	fg := colorMoss
	if focused {
		fg = colorLagoonBright
	}
	return pieceText(" ▶ ", fg, colorPanel)
}

func (m Model) terminalTitle(width int) string {
	focused := m.focus == PaneTerminal
	left := []piece{
		m.runGlyph(m.termState, focused),
		pieceText(" Terminal ", colorSandDim, colorPanel),
	}
	for i := range m.termTabs {
		left = append(left, tabPiece(fmt.Sprintf("%d", i+1), i == m.termIndex, focused))
	}
	left = append(left, tabPiece("+", false, focused))

	var right []piece
	switch {
	case m.running && m.runTarget == targetTerminal && m.termRunIndex == m.termIndex:
		right = append(right, pieceText(" running ", colorPalm, colorPanel))
	case m.running && m.runTarget == targetTerminal:
		right = append(right, pieceText(fmt.Sprintf(" shell %d running ", m.termRunIndex+1), colorPalm, colorPanel))
	default:
		right = append(right, pieceText(" $SHELL ", colorSandDim, colorPanel))
	}
	return joinLeftRight(width, colorPanel, left, right)
}

// terminalTabAt maps an inner title cell to a shell tab. The second result is
// true for the + tab, which creates a new shell.
func (m Model) terminalTabAt(rel int) (int, bool, bool) {
	start := m.runGlyph(m.termState, false).width + lipgloss.Width(" Terminal ")
	for i := range m.termTabs {
		w := lipgloss.Width(fmt.Sprintf(" %d ", i+1))
		if rel >= start && rel < start+w {
			return i, false, true
		}
		start += w
	}
	if rel >= start && rel < start+3 {
		return 0, true, true
	}
	return 0, false, false
}

// renderTerminal draws the terminal scrollback, the command prompt, and the
// footer. The prompt always has the keyboard.
func (m Model) renderTerminal(width, height int) string {
	focused := m.focus == PaneTerminal
	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}
	outH := height - 5
	if outH < 0 {
		outH = 0
	}
	var lines []string
	if outH > 0 {
		if block := m.terminalBlock(innerW, outH); block != "" {
			lines = append(lines, strings.Split(block, "\n")...)
		}
	}
	lines = append(lines, m.terminalPrompt(innerW))
	lines = append(lines, m.footer(PaneTerminal, innerW))
	return renderPanel(m.terminalTitle(innerW), strings.Join(lines, "\n"), width, height, focused)
}

func (m Model) terminalBlock(width, height int) string {
	if height <= 0 {
		return ""
	}
	if len(m.termLines) == 0 {
		return windowLines([]string{noteLine(" Type a command and press enter.", width)}, 0, width, height)
	}
	y := m.termOutput.YOffset
	if y < 0 {
		y = 0
	}
	return windowLines(stylePlain(m.termLines, width), y, width, height)
}

// terminalPrompt draws "$" and the input with a block caret. The input scrolls
// horizontally so the caret stays visible in a long command.
func (m Model) terminalPrompt(width int) string {
	if width < 1 {
		return ""
	}
	avail := width - 3
	if avail < 1 {
		return noteLine("", width)
	}
	runes := []rune(m.termInput)
	col := clamp(m.termCursor, 0, len(runes))
	start := 0
	if col > avail-1 {
		start = col - avail + 1
	}
	end := start + avail
	if end > len(runes) {
		end = len(runes)
	}
	visible := runes[start:end]
	caret := col - start
	pad := lipgloss.NewStyle().Background(colorPanel)
	prefix := lipgloss.NewStyle().Foreground(colorLagoonBright).Background(colorPanel).Render(" $ ")
	text := lipgloss.NewStyle().Foreground(colorSand).Background(colorPanel).Render(string(visible[:caret]))
	block := lipgloss.NewStyle().Foreground(colorOcean).Background(colorLagoonBright).Render(" ")
	tail := visible[caret:]
	if len(tail) > 0 {
		block = lipgloss.NewStyle().Foreground(colorOcean).Background(colorLagoonBright).Render(string(tail[0]))
		tail = tail[1:]
	}
	row := prefix + text + block +
		lipgloss.NewStyle().Foreground(colorSand).Background(colorPanel).Render(string(tail))
	if w := lipgloss.Width(row); w < width {
		row += pad.Width(width - w).Render("")
	}
	return row
}

func (m Model) outputBlock(width, height int) string {
	if height <= 0 {
		return ""
	}
	if len(m.outLines) == 0 {
		return windowLines([]string{noteLine(" No output yet. Run a command from Commands.", width)}, 0, width, height)
	}
	y := m.output.YOffset
	if y < 0 {
		y = 0
	}
	return windowLines(stylePlain(m.outLines, width), y, width, height)
}

// footerHint is one shortcut in a pane footer: the keys to press and what they
// do. Keys and labels are painted separately so the binding reads first.
type footerHint struct {
	key   string
	label string
}

// footerColours returns the key and label colours for a footer. The focused
// pane is brighter, and the key is always a different colour from its label.
func footerColours(focused bool) (key, label lipgloss.TerminalColor) {
	if focused {
		return colorLagoonBright, colorSand
	}
	return colorSand, colorMoss
}

func (m Model) footer(p Pane, width int) string {
	if width < 1 {
		return ""
	}
	keyFG, labelFG := footerColours(m.focus == p)
	paint := func(text string, fg lipgloss.TerminalColor) string {
		return lipgloss.NewStyle().Foreground(fg).Background(colorPanel).Render(text)
	}
	var b strings.Builder
	b.WriteString(paint(" ", labelFG))
	for i, h := range m.footerHints(p) {
		if i > 0 {
			b.WriteString(paint(" · ", colorMoss))
		}
		if h.key != "" {
			b.WriteString(paint(h.key, keyFG))
		}
		if h.label != "" {
			if h.key != "" {
				b.WriteString(paint(" ", labelFG))
			}
			b.WriteString(paint(h.label, labelFG))
		}
	}
	line := ansi.Truncate(b.String(), width, "")
	if used := lipgloss.Width(line); used < width {
		line += lipgloss.NewStyle().Background(colorPanel).Width(width - used).Render("")
	}
	return line
}

func (m Model) footerHints(p Pane) []footerHint {
	switch p {
	case PaneMain:
		switch m.previewKind {
		case previewFile:
			if m.findOpen {
				return []footerHint{{"↑↓", "match"}, {"enter", "next"}, {"esc", "close"}}
			}
			if m.previewEditing {
				return []footerHint{{"↑↓←→", "move"}, {"ctrl+s", "save"}, {"esc", "done"}}
			}
			if m.previewLang.Detected() {
				return []footerHint{
					{"↑↓←→", "move"}, {"[ ]", "fold"}, {"e", "edit"},
					{"ctrl+f", "find"}, {"esc", "close"}, {"ctrl+a", "ai"}, {"?", "keys"},
				}
			}
			return []footerHint{
				{"↑↓←→", "move"}, {"e", "edit"}, {"ctrl+f", "find"},
				{"esc", "close"}, {"ctrl+a", "ai"}, {"?", "keys"},
			}
		case previewCommit, previewDiff:
			return []footerHint{{"↑↓", "scroll"}, {"esc", "close"}, {"ctrl+a", "ai"}, {"?", "keys"}}
		}
		return []footerHint{
			{"n", "new"}, {"ctrl+p", "find"}, {"ctrl+r", "recent"},
			{"ctrl+a", "ai"}, {"?", "keys"}, {"tab", "panes"},
		}
	case PaneFiles:
		return []footerHint{{"↑↓", ""}, {"enter", ""}, {"R d", "manage"}, {"ctrl+p", "find"}, {"?", "keys"}}
	case PaneHistory:
		return []footerHint{{"↑↓", ""}, {"←→", "tab"}, {"?", "keys"}}
	case PaneBuild:
		if m.buildTab == buildTabOutput {
			return []footerHint{
				{"↑↓", "scroll"}, {"←→", "tab"}, {"r", "rerun"},
				{"c", "stop"}, {"?", "keys"},
			}
		}
		return []footerHint{
			{"↑↓", "pick"}, {"enter", "run"}, {"a", "mode"},
			{"r", "rerun"}, {"←→", "tab"}, {"?", "keys"},
		}
	case PaneTerminal:
		return []footerHint{
			{"enter", "run"}, {"ctrl+c", "stop"}, {"ctrl+t", "new"},
			{"ctrl+w", "close"}, {"ctrl+n/p", "shell"}, {"?", "keys"},
		}
	case PaneAI:
		if m.aiSession != nil {
			return nil
		}
		if m.aiInstall != "" {
			if m.running && m.runTarget == targetAI {
				return []footerHint{{"esc", "cancel"}}
			}
			return []footerHint{{"esc", "close"}}
		}
		return []footerHint{
			{"↑↓", "pick"}, {"enter", ""}, {"ctrl+a", "close"},
			{"r", "rescan"}, {"?", "keys"},
		}
	}
	return []footerHint{{"?", "keys"}}
}

// newFileView is the modal prompt for a new project path.
func (m Model) newFileView() string {
	innerW := m.width - 6
	if innerW > 64 {
		innerW = 64
	}
	if innerW < 24 {
		innerW = m.width - 2
	}
	if innerW < 8 {
		innerW = 8
	}
	rows := []string{
		noteLine(" Path is relative to "+filepath.Base(m.root), innerW),
		m.newFileInput(innerW),
	}
	if m.newFileErr != "" {
		rows = append(rows, noteLine(" "+m.newFileErr, innerW))
	} else {
		rows = append(rows, noteLine(" enter create · esc cancel", innerW))
	}
	boxH := len(rows) + 3
	if boxH > m.height {
		boxH = m.height
	}
	if boxH < 5 {
		boxH = 5
	}
	box := renderPanel(paintTitle(" + New file", innerW, true), strings.Join(rows, "\n"), innerW+2, boxH, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
}

func (m Model) newFileInput(width int) string {
	text := truncateCells(" "+m.newFilePath+"▌", width)
	return lipgloss.NewStyle().
		Foreground(colorLagoonBright).
		Background(colorSelectDim).
		Width(width).
		Render(text)
}

func (m Model) renderStatus(width int) string {
	line := m.statusLine(width)
	if lipgloss.Width(line) != width {
		line = lipgloss.NewStyle().Background(colorStatus).Foreground(colorSand).Width(width).Render(" bento")
	}
	return line
}

func (m Model) statusLine(width int) string {
	left := m.statusLeft()
	var right []piece
	if badge := m.updateBadge(); len(badge) > 0 {
		right = append(right, badge...)
		right = append(right, pieceText(" ", colorSand, colorStatus))
	}
	if m.aiSession != nil {
		name := providerName(m.aiSessionID)
		right = append(right,
			pieceText("▾ ", colorPalm, colorStatus),
			pieceText(name, colorPalm, colorStatus),
		)
		if m.focus == PaneAI {
			right = append(right, pieceText("  ctrl+\\ detach ", colorSandDim, colorStatus))
		} else {
			right = append(right, pieceText("  running ", colorSandDim, colorStatus))
		}
	}
	right = append(right, pieceText("bento", colorLagoon, colorStatus))
	return joinLeftRight(width, colorStatus, left, right)
}

func (m Model) statusLeft() []piece {
	bg := colorStatus
	if !m.gitReady {
		return []piece{pieceText(" ⎇ …", colorSandDim, bg)}
	}
	if !m.gitSnap.Repo {
		text := " ⎇ not a git repository"
		if m.gitSnap.Err != "" {
			text = " ⎇ " + m.gitSnap.Err
		}
		return []piece{pieceText(text, colorSand, bg)}
	}
	st := m.gitSnap.Status
	branch := st.Branch
	if branch == "" {
		branch = "HEAD"
	}
	branchFG := colorSand
	if st.Staged+st.Unstaged > 0 {
		branchFG = colorCoral
	}
	ahead, behind := "↑—", "↓—"
	aheadFG, behindFG := colorSandDim, colorSandDim
	if st.Upstream != "" || st.Ahead > 0 || st.Behind > 0 {
		ahead = fmt.Sprintf("↑%d", st.Ahead)
		behind = fmt.Sprintf("↓%d", st.Behind)
		if st.Ahead > 0 {
			aheadFG = colorPalm
		}
		if st.Behind > 0 {
			behindFG = colorCoral
		}
	}
	stageFG := colorSandDim
	if st.Staged > 0 {
		stageFG = colorCoral
	}
	unstagedFG := colorSandDim
	if st.Unstaged > 0 {
		unstagedFG = colorCoral
	}
	stashFG := colorSandDim
	if st.Stash > 0 {
		stashFG = colorLagoonBright
	}
	return []piece{
		pieceText(" ⎇ ", colorLagoonBright, bg),
		pieceText(branch, branchFG, bg),
		pieceText("  ", colorSand, bg),
		pieceText(ahead, aheadFG, bg),
		pieceText(" ", colorSand, bg),
		pieceText(behind, behindFG, bg),
		pieceText("  ● staged ", stageFG, bg),
		pieceText(fmt.Sprintf("%d", st.Staged), stageFG, bg),
		pieceText("  ● unstaged ", unstagedFG, bg),
		pieceText(fmt.Sprintf("%d", st.Unstaged), unstagedFG, bg),
		pieceText("  stash ", stashFG, bg),
		pieceText(fmt.Sprintf("%d", st.Stash), stashFG, bg),
	}
}

func (m Model) paletteView() string {
	innerW := m.width - 6
	if innerW > 78 {
		innerW = 78
	}
	if innerW < 24 {
		innerW = m.width - 2
	}
	if innerW < 8 {
		innerW = 8
	}
	items := m.paletteItems()
	var rows []string
	section := ""
	for i, it := range items {
		if it.section != section {
			if section != "" {
				rows = append(rows, noteLine("", innerW))
			}
			section = it.section
			rows = append(rows, lipgloss.NewStyle().
				Foreground(colorMoss).
				Background(colorPanel).
				Bold(true).
				Width(innerW).
				Render(" "+strings.ToUpper(section)))
		}
		rows = append(rows, paletteRow(it, i == m.paletteCursor, innerW))
	}
	rows = append(rows, noteLine("", innerW))
	rows = append(rows, noteLine(" ↑↓ move · enter run · ? close", innerW))
	boxH := len(rows) + 3
	if boxH > m.height {
		boxH = m.height
	}
	if boxH < 5 {
		boxH = 5
	}
	title := "? Keys · " + m.paneTitle(m.palettePane)
	bodyH := boxH - 3
	if bodyH < 0 {
		bodyH = 0
	}
	if len(rows) > bodyH {
		rows = rows[:bodyH]
	}
	box := renderPanel(paintTitle(" "+title, innerW, true), strings.Join(rows, "\n"), innerW+2, boxH, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
}

func paletteRow(item paletteItem, selected bool, width int) string {
	bg := lipgloss.TerminalColor(colorPanel)
	if selected {
		bg = colorSelect
	}
	keyText := fmt.Sprintf(" %-12s", keyLabel(item.action.binding))
	keyStyle := lipgloss.NewStyle().Foreground(colorLagoonBright).Background(bg)
	if selected {
		keyStyle = keyStyle.Bold(true)
	}
	keyText = truncateCells(keyText, width)
	descW := width - lipgloss.Width(keyText)
	if descW < 1 {
		return keyStyle.Width(width).Render(keyText)
	}
	desc := truncateCells(item.action.desc, descW)
	return keyStyle.Render(keyText) +
		lipgloss.NewStyle().
			Foreground(colorSand).
			Background(bg).
			Width(descW).
			Render(desc)
}

func (m Model) tooSmall() string {
	msg := "Make the terminal larger to use Bento terminal.\nPress q to quit."
	box := lipgloss.NewStyle().
		Foreground(colorSand).
		Background(colorPanel).
		Padding(1, 2).
		Render(msg)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
}
