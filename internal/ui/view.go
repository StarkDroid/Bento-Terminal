package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
	"github.com/StarkDroid/Bento-Terminal/internal/files"
	"github.com/StarkDroid/Bento-Terminal/internal/git"
)

func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "Bento terminal"
	}
	if m.paletteOpen {
		return m.paletteView()
	}
	lay := computeLayout(m.width, m.height)
	if !lay.OK {
		return m.tooSmall()
	}
	main := m.renderMain(lay.Main.W, lay.Main.H)
	filesPane := m.renderFiles(lay.Files.W, lay.Files.H)
	history := m.renderHistory(lay.History.W, lay.History.H)
	buildPane := m.renderBuild(lay.Build.W, lay.Build.H)
	status := m.renderStatus(lay.Status.W)
	right := lipgloss.JoinVertical(lipgloss.Left, filesPane, gapRow(lay.Files.W), history)
	top := lipgloss.JoinHorizontal(lipgloss.Top, main, gapCol(lay.Main.H), right)
	return lipgloss.JoinVertical(lipgloss.Left, top, gapRow(m.width), buildPane, status)
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
	title := " " + filepath.Base(m.root)
	var body string
	switch m.previewKind {
	case previewFile:
		title = " ≡ " + m.previewTitle
		if m.previewLang.Detected() {
			lang := m.previewLang.Name
			if !m.previewSyntax {
				lang += " off"
			}
			title += "  ·  " + lang
		}
		body = windowLines(m.previewLines, m.preview.YOffset, innerW, bodyH)
	case previewCommit:
		title = " ⎇ " + m.previewTitle
		body = windowLines(m.previewLines, m.preview.YOffset, innerW, bodyH)
	default:
		body = m.homeBody(innerW, bodyH)
	}
	if bodyH > 0 {
		body += "\n"
	}
	body += m.footer(PaneMain, innerW)
	return renderPanel(paintTitle(title, innerW, focused), body, width, height, focused)
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
	return panel("⎇ History", strings.Join(lines, "\n"), width, height, focused)
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
	listH, outH := buildParts(height)
	var body []string
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
		scroll := ensure(m.cmdCursor, m.cmdScroll, listH)
		end := scroll + listH
		if end > len(cmds) {
			end = len(cmds)
		}
		for i := scroll; i < end; i++ {
			body = append(body, m.commandLine(cmds[i], innerW, i == m.cmdCursor))
		}
	}
	for len(body) < listH {
		body = append(body, noteLine("", innerW))
	}
	if len(body) > listH && listH > 0 {
		body = body[:listH]
	}
	if outH > 0 {
		body = append(body, sepLine(innerW))
		block := m.outputBlock(innerW, outH)
		if block != "" {
			body = append(body, strings.Split(block, "\n")...)
		}
	}
	body = append(body, m.footer(PaneBuild, innerW))
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
	mode := "auto"
	if m.mode == build.ModeManual {
		mode = "manual"
	}
	titleFG := colorSandDim
	if m.focus == PaneBuild {
		titleFG = colorLagoonBright
	}
	var parts []piece
	switch m.buildState {
	case stateRunning:
		spin := m.spinner.View()
		parts = append(parts,
			pieceText(" ", colorSand, colorPanel),
			piece{view: spin, width: lipgloss.Width(spin)},
			pieceText(" Build  ", titleFG, colorPanel),
		)
	case stateOK:
		parts = append(parts, pieceText(" ✓ Build  ", colorPalm, colorPanel))
	case stateFailed:
		parts = append(parts, pieceText(" ✗ Build  ", colorCoral, colorPanel))
	case stateCanceled:
		parts = append(parts, pieceText(" ✗ Build  ", colorCoral, colorPanel))
	default:
		parts = append(parts, pieceText(" ▶ Build  ", titleFG, colorPanel))
	}
	parts = append(parts, pieceText(mode, colorLagoon, colorPanel))
	if m.last != nil {
		parts = append(parts, pieceText("  ·  "+m.last.Name, colorSandDim, colorPanel))
	}
	return joinPieces(width, colorPanel, parts)
}

func (m Model) outputBlock(width, height int) string {
	if height <= 0 {
		return ""
	}
	if len(m.outLines) == 0 {
		return windowLines([]string{noteLine(" Idle. Select a command and press enter.", width)}, 0, width, height)
	}
	y := m.output.YOffset
	if y < 0 {
		y = 0
	}
	return windowLines(stylePlain(m.outLines, width), y, width, height)
}

func (m Model) footer(p Pane, width int) string {
	fg := colorSandDim
	if m.focus == p {
		fg = colorLagoon
	}
	text := " " + strings.Join(m.footerHints(p), " · ")
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(colorPanel).
		Width(width).
		Render(truncateCells(text, width))
}

func (m Model) footerHints(p Pane) []string {
	switch p {
	case PaneMain:
		switch m.previewKind {
		case previewFile:
			if m.previewLang.Detected() {
				return []string{"↑↓←→ move", "[ ] fold", "? keys"}
			}
			return []string{"↑↓←→ move", "? keys"}
		case previewCommit:
			return []string{"↑↓ scroll", "? keys"}
		}
		return []string{"? keys", "tab panes"}
	case PaneFiles:
		return []string{"↑↓", "enter open", "? keys"}
	case PaneHistory:
		return []string{"↑↓", "enter show", "? keys"}
	case PaneBuild:
		return []string{"↑↓ pick", "enter run", "a mode", "r rerun", "c cancel", "? keys"}
	}
	return []string{"? keys"}
}

func (m Model) renderStatus(width int) string {
	line := m.statusLine(width)
	if lipgloss.Width(line) != width {
		line = lipgloss.NewStyle().Background(colorStatus).Foreground(colorSand).Width(width).Render(" bento")
	}
	return line
}

func (m Model) statusLine(width int) string {
	brand := pieceText("bento", colorLagoon, colorStatus)
	left := m.statusLeft()
	used := 0
	for _, p := range left {
		used += p.width
	}
	gap := width - used - brand.width
	if gap < 1 {
		return joinPieces(width, colorStatus, left)
	}
	parts := append(append([]piece{}, left...), piece{
		view:  lipgloss.NewStyle().Background(colorStatus).Width(gap).Render(""),
		width: gap,
	}, brand)
	return joinPieces(width, colorStatus, parts)
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
