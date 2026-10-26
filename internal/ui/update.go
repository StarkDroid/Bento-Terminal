package ui

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
	"github.com/StarkDroid/Bento-Terminal/internal/files"
	"github.com/StarkDroid/Bento-Terminal/internal/git"
	"github.com/StarkDroid/Bento-Terminal/internal/syntax"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.reflowPreview(false), nil
	case tea.KeyMsg:
		return m.onKey(msg)
	case tea.MouseMsg:
		return m.onMouse(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.running {
			return m, cmd
		}
		return m, nil
	case treeLoadedMsg:
		m.tree = msg.tree
		m.treeReady = true
		if msg.err != nil {
			m.treeErr = msg.err.Error()
		}
		m.fileCursor = 0
		m.fileScroll = 0
		return m, nil
	case gitLoadedMsg:
		return m.applyGit(msg.snap), nil
	case buildLoadedMsg:
		m.autoCmds = msg.result.Auto
		m.manualCmds = msg.result.Manual
		m.systems = msg.result.Systems
		m.buildReady = true
		return m.clampCmd(), nil
	case runStartedMsg:
		return m.onStarted(msg)
	case runLineMsg:
		if msg.id != m.runID {
			return m, nil
		}
		return m.appendOut(msg.line), listenCmd(msg.id, msg.events)
	case runDoneMsg:
		return m.onDone(msg), nil
	case refreshMsg:
		return m, tea.Batch(loadGitCmd(m.root), refreshCmd())
	default:
		return m, nil
	}
}

func (m Model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m.stopRun(), tea.Quit
	}
	if m.paletteOpen {
		return m.onPaletteKey(msg)
	}
	if key.Matches(msg, m.keys.Help) {
		return m.openPalette(), nil
	}
	for _, a := range m.actions() {
		if key.Matches(msg, a.binding) {
			return m.runAction(a.id)
		}
	}
	return m, nil
}

// onMouse focuses a pane on click and lets the pointer move the selection, fold
// a block, or scroll a list.
func (m Model) onMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.paletteOpen {
		return m, nil
	}
	lay := computeLayout(m.width, m.height)
	if !lay.OK {
		return m, nil
	}
	pane, ok := paneAt(lay, msg.X, msg.Y)
	if !ok {
		return m, nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if msg.Action == tea.MouseActionPress {
			return m.wheel(pane, -3), nil
		}
	case tea.MouseButtonWheelDown:
		if msg.Action == tea.MouseActionPress {
			return m.wheel(pane, 3), nil
		}
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			return m.click(pane, msg.Y, lay)
		}
	}
	return m, nil
}

func (m Model) click(pane Pane, y int, lay layout) (tea.Model, tea.Cmd) {
	m.focus = pane
	m = m.reflowPreview(false)
	switch pane {
	case PaneMain:
		if m.previewKind != previewFile {
			return m, nil
		}
		row := y - lay.MainBodyTop() + m.preview.YOffset
		if row < 0 || row >= len(m.previewRows) {
			return m, nil
		}
		line := m.previewRows[row].line
		m.previewCursor = line
		m.previewCol = 0
		if block, ok := syntax.BlockAt(m.previewBlocks, line); ok && block.Start == line {
			m.previewFolds[line] = !m.previewFolds[line]
		}
		return m.reflowPreview(false), nil
	case PaneFiles:
		rows := m.tree.Rows()
		row := y - lay.FilesBodyTop() + m.fileScroll
		if row < 0 || row >= len(rows) {
			return m, nil
		}
		m.fileCursor = row
		if rows[row].IsDir {
			m.tree.Toggle(row)
			return m, nil
		}
		return m.openSelection()
	case PaneHistory:
		row := y - lay.HistoryBodyTop() + m.commitScroll
		if row < 0 || row >= len(m.gitSnap.Commits) {
			return m, nil
		}
		m.commitCursor = row
		return m.openCommit()
	case PaneBuild:
		row := y - lay.BuildBodyTop() + m.cmdScroll
		if row < 0 || row >= len(m.commands()) {
			return m, nil
		}
		m.cmdCursor = row
	}
	return m, nil
}

func (m Model) wheel(pane Pane, delta int) Model {
	m.focus = pane
	m = m.reflowPreview(false)
	switch pane {
	case PaneMain:
		m = m.syncSizes()
		if delta > 0 {
			m.preview.ScrollDown(delta)
		} else {
			m.preview.ScrollUp(-delta)
		}
	case PaneFiles:
		if m.tree != nil {
			m = m.moveFile(delta)
		}
	case PaneHistory:
		m = m.moveCommit(delta)
	case PaneBuild:
		m = m.scrollOutput(delta)
	}
	return m
}

func (m Model) openSelection() (tea.Model, tea.Cmd) {
	if m.tree == nil {
		return m, nil
	}
	rows := m.tree.Rows()
	if m.fileCursor < 0 || m.fileCursor >= len(rows) {
		return m, nil
	}
	row := rows[m.fileCursor]
	if row.IsDir {
		m.tree.Toggle(m.fileCursor)
		return m, nil
	}
	body, err := files.Preview(filepath.Join(m.root, filepath.FromSlash(row.Rel)))
	m.previewKind = previewFile
	m.previewTitle = row.Name
	m.previewNumbered = false
	m.previewLang = syntax.Language{}
	m.previewSyntax = false
	m.previewCursor = 0
	m.previewCol = 0
	m.previewFolds = map[int]bool{}
	m.previewBlocks = nil
	switch {
	case errors.Is(err, files.ErrBinary):
		m.previewRaw = "Binary file"
	case err != nil:
		m.previewRaw = err.Error()
	case strings.TrimSpace(body) == "":
		m.previewRaw = "Empty file"
	default:
		m.previewRaw = body
		m.previewNumbered = true
		m.previewLang = syntax.Detect(row.Name, body)
		m.previewSyntax = m.previewLang.Detected()
	}
	m = m.prepareBlocks()
	m = m.refreshSpans()
	m.preview.GotoTop()
	return m.reflowPreview(true), nil
}

func (m Model) openCommit() (tea.Model, tea.Cmd) {
	if !m.gitSnap.Repo || m.commitCursor < 0 || m.commitCursor >= len(m.gitSnap.Commits) {
		return m, nil
	}
	c := m.gitSnap.Commits[m.commitCursor]
	body, err := git.Show(m.root, c.Hash)
	m.previewKind = previewCommit
	m.previewTitle = c.Hash
	m.previewNumbered = false
	m.previewLang = syntax.Language{}
	m.previewSyntax = false
	m.previewBlocks = nil
	m.previewSpans = nil
	m.previewRows = nil
	m.previewVisible = nil
	if err != nil {
		m.previewRaw = err.Error()
	} else if strings.TrimSpace(body) == "" {
		m.previewRaw = c.Subject
	} else {
		m.previewRaw = body
	}
	m.preview.GotoTop()
	return m.reflowPreview(true), nil
}

func (m Model) prepareBlocks() Model {
	m.previewBlocks = nil
	if !m.previewNumbered {
		return m
	}
	m.previewBlocks = syntax.Outline(m.previewLang, fileLines(m.previewRaw))
	return m
}

func (m Model) movePreview(delta int) Model {
	if m.previewKind == previewCommit {
		m = m.syncSizes()
		if delta > 0 {
			m.preview.ScrollDown(delta)
		} else {
			m.preview.ScrollUp(-delta)
		}
		return m
	}
	if m.previewKind != previewFile || len(m.previewVisible) == 0 {
		return m
	}
	pos := m.previewVisiblePos()
	next := clamp(pos+delta, 0, len(m.previewVisible)-1)
	if next == pos {
		return m
	}
	m.previewCursor = m.previewVisible[next]
	m = m.clampPreviewCol()
	m = m.reflowPreview(false)
	return m.scrollPreview()
}

// movePreviewCol moves the caret one character, wrapping at the ends of
// visible lines so arrow keys feel like a code editor.
func (m Model) movePreviewCol(delta int) Model {
	if m.previewKind != previewFile || len(m.previewVisible) == 0 {
		return m
	}
	lines := fileLines(m.previewRaw)
	if m.previewCursor < 0 || m.previewCursor >= len(lines) {
		return m
	}
	pos := m.previewVisiblePos()
	n := len([]rune(lines[m.previewCursor]))
	col := clamp(m.previewCol, 0, n)
	switch {
	case delta < 0 && col == 0:
		if pos == 0 {
			return m
		}
		pos--
		m.previewCursor = m.previewVisible[pos]
		m.previewCol = len([]rune(lines[m.previewCursor]))
	case delta > 0 && col == n:
		if pos >= len(m.previewVisible)-1 {
			return m
		}
		pos++
		m.previewCursor = m.previewVisible[pos]
		m.previewCol = 0
	default:
		m.previewCol = clamp(col+delta, 0, n)
	}
	m = m.reflowPreview(false)
	return m.scrollPreview()
}

func (m Model) previewVisiblePos() int {
	pos := 0
	for i, line := range m.previewVisible {
		if line <= m.previewCursor {
			pos = i
		}
	}
	return pos
}

func (m Model) clampPreviewCol() Model {
	lines := fileLines(m.previewRaw)
	if m.previewCursor < 0 || m.previewCursor >= len(lines) {
		m.previewCol = 0
		return m
	}
	m.previewCol = clamp(m.previewCol, 0, len([]rune(lines[m.previewCursor])))
	return m
}

func (m Model) scrollPreview() Model {
	m = m.syncSizes()
	row := -1
	for i, r := range m.previewRows {
		if r.line == m.previewCursor {
			row = i
			break
		}
	}
	if row < 0 {
		return m
	}
	m.preview.SetYOffset(ensure(row, m.preview.YOffset, m.preview.Height))
	return m
}

func (m Model) setFold(collapse bool) Model {
	if m.previewKind != previewFile {
		return m
	}
	block, ok := syntax.BlockAt(m.previewBlocks, m.previewCursor)
	if !ok {
		return m
	}
	if m.previewFolds == nil {
		m.previewFolds = map[int]bool{}
	}
	if collapse {
		m.previewFolds[block.Start] = true
	} else {
		delete(m.previewFolds, block.Start)
	}
	m.previewCursor = block.Start
	m.previewCol = 0
	m = m.reflowPreview(false)
	return m.scrollPreview()
}

func (m Model) toggleFold() Model {
	block, ok := syntax.BlockAt(m.previewBlocks, m.previewCursor)
	if !ok {
		return m
	}
	return m.setFold(!m.previewFolds[block.Start])
}

func (m Model) toggleSyntax() Model {
	if m.previewKind != previewFile || !m.previewNumbered || !m.previewLang.Detected() {
		return m
	}
	m.previewSyntax = !m.previewSyntax
	m = m.refreshSpans()
	return m.reflowPreview(false)
}

func (m Model) moveFile(delta int) Model {
	n := 0
	if m.tree != nil {
		n = len(m.tree.Rows())
	}
	if n == 0 {
		m.fileCursor = 0
		return m
	}
	m.fileCursor = clamp(m.fileCursor+delta, 0, n-1)
	m.fileScroll = ensure(m.fileCursor, m.fileScroll, m.filePage())
	return m
}

func (m Model) moveCommit(delta int) Model {
	n := len(m.gitSnap.Commits)
	if n == 0 {
		m.commitCursor = 0
		return m
	}
	m.commitCursor = clamp(m.commitCursor+delta, 0, n-1)
	m.commitScroll = ensure(m.commitCursor, m.commitScroll, m.historyPage())
	return m
}

func (m Model) moveCmd(delta int) Model {
	n := len(m.commands())
	if n == 0 {
		m.cmdCursor = 0
		return m
	}
	m.cmdCursor = clamp(m.cmdCursor+delta, 0, n-1)
	m.cmdScroll = ensure(m.cmdCursor, m.cmdScroll, m.commandPage())
	return m
}

func (m Model) commands() []build.Command {
	if m.mode == build.ModeManual {
		return m.manualCmds
	}
	return m.autoCmds
}

func (m Model) toggleMode() Model {
	if m.mode == build.ModeAuto {
		m.mode = build.ModeManual
	} else {
		m.mode = build.ModeAuto
	}
	m.cmdCursor = 0
	m.cmdScroll = 0
	return m.clampCmd()
}

func (m Model) clampCmd() Model {
	n := len(m.commands())
	if n == 0 {
		m.cmdCursor = 0
		return m
	}
	m.cmdCursor = clamp(m.cmdCursor, 0, n-1)
	m.cmdScroll = ensure(m.cmdCursor, m.cmdScroll, m.commandPage())
	return m
}

func (m Model) launch(c build.Command) (tea.Model, tea.Cmd) {
	m = m.stopRun()
	m.cancelRequested = false
	m.runID++
	id := m.runID
	m.running = true
	m.buildState = stateRunning
	m.follow = true
	copied := c
	m.last = &copied
	m.outLines = nil
	m = m.appendOut("$ " + c.Name)
	m = m.syncSizes()
	return m, tea.Batch(startCmd(id, m.root, c), m.spinner.Tick)
}

func (m Model) rerun() (tea.Model, tea.Cmd) {
	if m.last == nil {
		return m, nil
	}
	return m.launch(*m.last)
}

func (m Model) stopRun() Model {
	if m.cancel == nil {
		return m
	}
	m.cancelRequested = true
	cancel := m.cancel
	m.cancel = nil
	cancel()
	return m
}

func (m Model) onStarted(msg runStartedMsg) (tea.Model, tea.Cmd) {
	if msg.id != m.runID {
		if msg.cancel != nil {
			msg.cancel()
		}
		return m, nil
	}
	if msg.err != nil {
		m.running = false
		m.buildState = stateFailed
		m.cancel = nil
		return m.appendOut("✗ " + msg.err.Error()), nil
	}
	m.cancel = msg.cancel
	m.running = true
	if msg.events == nil {
		m.running = false
		m.buildState = stateFailed
		return m, nil
	}
	return m, listenCmd(msg.id, msg.events)
}

func (m Model) onDone(msg runDoneMsg) Model {
	if msg.id != m.runID {
		return m
	}
	m.running = false
	m.cancel = nil
	switch {
	case m.cancelRequested:
		m.buildState = stateCanceled
		m = m.appendOut("✗ cancelled")
	case msg.err == nil:
		m.buildState = stateOK
		m = m.appendOut("✓ done")
	default:
		m.buildState = stateFailed
		m = m.appendOut("✗ " + msg.err.Error())
	}
	return m
}

func (m Model) appendOut(line string) Model {
	m.outLines = append(m.outLines, line)
	if len(m.outLines) > 2000 {
		m.outLines = m.outLines[len(m.outLines)-2000:]
	}
	m.output.SetContent(strings.Join(m.outLines, "\n"))
	if m.follow {
		m.output.GotoBottom()
	}
	return m
}

func (m Model) applyGit(snap git.Snapshot) Model {
	var keep string
	if m.commitCursor >= 0 && m.commitCursor < len(m.gitSnap.Commits) {
		keep = m.gitSnap.Commits[m.commitCursor].Hash
	}
	m.gitSnap = snap
	m.gitReady = true
	m.commitCursor = 0
	if keep != "" {
		for i, c := range snap.Commits {
			if c.Hash == keep {
				m.commitCursor = i
				break
			}
		}
	}
	if n := len(snap.Commits); n > 0 {
		m.commitCursor = clamp(m.commitCursor, 0, n-1)
	}
	m.commitScroll = ensure(m.commitCursor, m.commitScroll, m.historyPage())
	return m
}

func (m Model) syncSizes() Model {
	lay := computeLayout(m.width, m.height)
	if !lay.OK {
		return m
	}
	m.preview.Width = lay.Main.W - 2
	m.preview.Height = lay.Main.H - 4
	if m.preview.Width < 1 {
		m.preview.Width = 1
	}
	if m.preview.Height < 1 {
		m.preview.Height = 1
	}
	_, outH := buildParts(lay.Build.H)
	m.output.Width = lay.Build.W - 2
	m.output.Height = outH
	if m.output.Width < 1 {
		m.output.Width = 1
	}
	if m.output.Height < 1 {
		m.output.Height = 1
	}
	return m
}

func (m Model) reflowPreview(reset bool) Model {
	m = m.syncSizes()
	y := m.preview.YOffset
	switch {
	case m.previewKind == previewNone:
		m.previewLines = nil
		m.previewRows = nil
		m.previewVisible = nil
		m.preview.SetContent("")
	case m.previewNumbered:
		m.previewRows, m.previewVisible = m.buildCode()
		m.previewLines = make([]string, len(m.previewRows))
		for i, row := range m.previewRows {
			m.previewLines[i] = row.text
		}
		m.preview.SetContent(strings.Join(m.previewLines, "\n"))
	default:
		m.previewLines = wrappedLines(m.previewRaw, m.preview.Width)
		m.previewRows = nil
		m.previewVisible = nil
		m.preview.SetContent(strings.Join(m.previewLines, "\n"))
	}
	if reset {
		m.preview.GotoTop()
	} else {
		m.preview.SetYOffset(y)
	}
	return m
}

func (m Model) filePage() int {
	lay := computeLayout(m.width, m.height)
	h := lay.Files.H - 4
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) historyPage() int {
	lay := computeLayout(m.width, m.height)
	h := lay.History.H - 4
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) commandPage() int {
	lay := computeLayout(m.width, m.height)
	listH, _ := buildParts(lay.Build.H)
	if listH < 1 {
		return 1
	}
	return listH
}

func (m Model) previewPage() int {
	lay := computeLayout(m.width, m.height)
	h := lay.Main.H - 4
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) outputPage() int {
	lay := computeLayout(m.width, m.height)
	_, outH := buildParts(lay.Build.H)
	if outH < 1 {
		return 1
	}
	return outH
}

func (m Model) scrollOutput(delta int) Model {
	m = m.syncSizes()
	if delta > 0 {
		m.output.ScrollDown(delta)
	} else {
		m.output.ScrollUp(-delta)
	}
	m.follow = m.output.AtBottom()
	return m
}
