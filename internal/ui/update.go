package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
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
		return m.appendRun(msg.line), listenCmd(msg.id, msg.events)
	case runDoneMsg:
		return m.onDone(msg), nil
	case refreshMsg:
		return m, tea.Batch(loadGitCmd(m.root), refreshCmd())
	default:
		return m, nil
	}
}

func (m Model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.newFileOpen {
		return m.onNewFileKey(msg)
	}
	if m.previewEditing {
		return m.onEditorKey(msg)
	}
	if !m.paletteOpen && !m.settingsOpen && m.focus == PaneBuild && m.buildTab == buildTabTerminal {
		return m.onTerminalKey(msg)
	}
	if key.Matches(msg, m.keys.Quit) {
		return m.stopRun(), tea.Quit
	}
	if m.settingsOpen {
		if key.Matches(msg, m.keys.Close, m.keys.Settings) {
			m.settingsOpen = false
		}
		return m, nil
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
	if m.settingsOpen {
		return m.settingsClick(msg)
	}
	if m.paletteOpen || m.newFileOpen {
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
			return m.click(pane, msg.X, msg.Y, lay)
		}
	}
	return m, nil
}

func (m Model) click(pane Pane, x, y int, lay layout) (tea.Model, tea.Cmd) {
	if y == 1 && m.previewKind == previewFile && !m.previewEditing && m.newFileTabAt(x-1, lay.Main.W-2) {
		return m.openNewFile(), nil
	}
	if y == lay.HistoryBodyTop()-1 {
		if tab, ok := historyTabAt(x - (lay.Main.W + 2)); ok {
			m.focus = PaneHistory
			m.historyTab = tab
			return m.reflowPreview(false), nil
		}
	}
	if y == lay.BuildBodyTop()-1 {
		if tab, ok := buildTabAt(x - 1); ok {
			m.focus = PaneBuild
			m.buildTab = tab
			return m.reflowPreview(false), nil
		}
	}
	m.focus = pane
	m = m.reflowPreview(false)
	switch pane {
	case PaneMain:
		if m.previewKind != previewFile || m.previewEditing {
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
		if m.historyTab == histTabChanges {
			row := y - lay.HistoryBodyTop() + m.changeScroll
			if row < 0 || row >= len(m.gitSnap.Status.Files) {
				return m, nil
			}
			m.changeCursor = row
			return m.openChange()
		}
		row := y - lay.HistoryBodyTop() + m.commitScroll
		if row < 0 || row >= len(m.gitSnap.Commits) {
			return m, nil
		}
		m.commitCursor = row
		return m.openCommit()
	case PaneBuild:
		if m.buildTab == buildTabTerminal {
			return m, nil
		}
		row := y - lay.BuildBodyTop() + m.cmdScroll
		if row < 0 || row >= len(m.commands()) {
			return m, nil
		}
		m.cmdCursor = row
	}
	return m, nil
}

// newFileTabAt reports whether an inner title cell lands on the new-file tab.
// The tab sits directly after the file tab, so its range depends on the name.
// A long name can push the tab past the right edge; the clipped part is dead.
func (m Model) newFileTabAt(rel, width int) bool {
	start := m.fileTab(false).width
	end := start + m.newFileTab(false).width
	if end > width {
		end = width
	}
	return rel >= start && rel < end
}

// historyTabAt maps an inner title cell to a history tab. Tabs are drawn as
// " ⎇ " then " History " and " Changes ".
func historyTabAt(rel int) (historyTab, bool) {
	switch {
	case rel >= 3 && rel < 12:
		return histTabCommits, true
	case rel >= 13 && rel < 22:
		return histTabChanges, true
	}
	return 0, false
}

// buildTabAt maps an inner title cell to a build tab. Tabs are drawn as a
// three-cell state glyph then " Build " and " Terminal ".
func buildTabAt(rel int) (buildTab, bool) {
	switch {
	case rel >= 3 && rel < 10:
		return buildTabCommands, true
	case rel >= 11 && rel < 21:
		return buildTabTerminal, true
	}
	return 0, false
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
		m = m.moveHistory(delta)
	case PaneBuild:
		if m.buildTab == buildTabTerminal {
			m = m.scrollTerminal(delta)
		} else {
			m = m.scrollOutput(delta)
		}
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
	return m.openFile(row.Rel, row.Name), nil
}

func (m Model) openCommit() (tea.Model, tea.Cmd) {
	if !m.gitSnap.Repo || m.commitCursor < 0 || m.commitCursor >= len(m.gitSnap.Commits) {
		return m, nil
	}
	c := m.gitSnap.Commits[m.commitCursor]
	body, err := git.Show(m.root, c.Hash)
	m.previewKind = previewCommit
	m.previewTitle = c.Hash
	m.previewPath = ""
	m = m.resetPreviewState()
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

// openChange previews the patch for the selected working-tree change.
func (m Model) openChange() (tea.Model, tea.Cmd) {
	files := m.gitSnap.Status.Files
	if !m.gitSnap.Repo || m.changeCursor < 0 || m.changeCursor >= len(files) {
		return m, nil
	}
	f := files[m.changeCursor]
	body, err := git.Diff(m.root, f)
	m.previewKind = previewDiff
	m.previewTitle = f.Path
	m.previewPath = ""
	m = m.resetPreviewState()
	if err != nil {
		m.previewRaw = err.Error()
	} else {
		m.previewRaw = body
	}
	m.preview.GotoTop()
	return m.reflowPreview(true), nil
}

// resetPreviewState clears the editor and language fields when a commit or a
// patch replaces a file in the preview.
func (m Model) resetPreviewState() Model {
	m.previewNumbered = false
	m.previewLang = syntax.Language{}
	m.previewSyntax = false
	m.previewEditing = false
	m.previewEditLines = nil
	m.previewDirty = false
	m.previewSaved = false
	m.previewSaveErr = false
	m.previewCursor = 0
	m.previewCol = 0
	m.previewFolds = map[int]bool{}
	m.previewBlocks = nil
	m.previewSpans = nil
	m.previewRows = nil
	m.previewVisible = nil
	return m
}

// openHistorySelection opens a commit or a patch depending on the visible tab.
func (m Model) openHistorySelection() (tea.Model, tea.Cmd) {
	if m.historyTab == histTabChanges {
		return m.openChange()
	}
	return m.openCommit()
}

// switchHistoryTab flips between commits and changes; the direction is kept so
// the two tab keys stay symmetric.
func (m Model) switchHistoryTab(delta int) Model {
	if delta == 0 {
		return m
	}
	if m.historyTab == histTabCommits {
		m.historyTab = histTabChanges
	} else {
		m.historyTab = histTabCommits
	}
	return m
}

// switchBuildTab flips between the command list and the terminal.
func (m Model) switchBuildTab(delta int) Model {
	if delta == 0 {
		return m
	}
	if m.buildTab == buildTabCommands {
		m.buildTab = buildTabTerminal
	} else {
		m.buildTab = buildTabCommands
	}
	return m
}

// moveHistory moves the cursor of the visible history tab.
func (m Model) moveHistory(delta int) Model {
	if m.historyTab == histTabChanges {
		return m.moveChange(delta)
	}
	return m.moveCommit(delta)
}

func (m Model) moveChange(delta int) Model {
	n := len(m.gitSnap.Status.Files)
	if n == 0 {
		m.changeCursor = 0
		return m
	}
	m.changeCursor = clamp(m.changeCursor+delta, 0, n-1)
	m.changeScroll = ensure(m.changeCursor, m.changeScroll, m.historyPage())
	return m
}

func (m Model) prepareBlocks() Model {
	m.previewBlocks = nil
	if !m.previewNumbered {
		return m
	}
	m.previewBlocks = syntax.Outline(m.previewLang, m.sourceLines())
	return m
}

func (m Model) movePreview(delta int) Model {
	if m.previewKind == previewCommit || m.previewKind == previewDiff {
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
	lines := m.sourceLines()
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
	lines := m.sourceLines()
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
	m.runTarget = targetBuild
	m.buildState = stateRunning
	m.follow = true
	copied := c
	m.last = &copied
	m.outLines = nil
	m = m.appendRun("$ " + c.Name)
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
		m = m.setRunState(stateFailed)
		m.cancel = nil
		return m.appendRun("✗ " + msg.err.Error()), nil
	}
	m.cancel = msg.cancel
	m.running = true
	if msg.events == nil {
		m.running = false
		m = m.setRunState(stateFailed)
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
		m = m.setRunState(stateCanceled)
		m = m.appendRun("✗ cancelled")
	case msg.err == nil:
		m = m.setRunState(stateOK)
		m = m.appendRun("✓ done")
	default:
		m = m.setRunState(stateFailed)
		m = m.appendRun("✗ " + msg.err.Error())
	}
	return m
}

func (m Model) appendOut(line string) Model {
	clean := sanitizeLine(line)
	if line != "" && clean == "" {
		// The command wrote nothing but control sequences.
		return m
	}
	m.outLines = append(m.outLines, clean)
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
	var keepChange string
	if m.changeCursor >= 0 && m.changeCursor < len(m.gitSnap.Status.Files) {
		keepChange = m.gitSnap.Status.Files[m.changeCursor].Path
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
	m.changeCursor = 0
	if keepChange != "" {
		for i, f := range snap.Status.Files {
			if f.Path == keepChange {
				m.changeCursor = i
				break
			}
		}
	}
	if n := len(snap.Status.Files); n > 0 {
		m.changeCursor = clamp(m.changeCursor, 0, n-1)
	}
	m.changeScroll = ensure(m.changeCursor, m.changeScroll, m.historyPage())
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
	termW := lay.Build.W - 2
	termH := lay.Build.H - 5
	if termW < 1 {
		termW = 1
	}
	if termH < 1 {
		termH = 1
	}
	prevH := m.termOutput.Height
	m.termOutput.Width = termW
	m.termOutput.Height = termH
	if m.termFollow && prevH != termH {
		m.termOutput.GotoBottom()
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
		if m.previewKind == previewDiff {
			m.previewLines = diffLines(m.previewRaw, m.preview.Width)
		} else {
			m.previewLines = wrappedLines(m.previewRaw, m.preview.Width)
		}
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
