package ui

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/trishiraj/bentro/internal/build"
	"github.com/trishiraj/bentro/internal/files"
	"github.com/trishiraj/bentro/internal/git"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.reflowPreview(false), nil
	case tea.KeyMsg:
		return m.onKey(msg)
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
	if m.helpOpen {
		if key.Matches(msg, m.keys.Help, m.keys.Close) {
			m.helpOpen = false
		}
		return m, nil
	}
	switch {
	case key.Matches(msg, m.keys.Help):
		m.helpOpen = true
		return m, nil
	case key.Matches(msg, m.keys.Next):
		m.focus = Cycle(m.focus, 1)
		return m, nil
	case key.Matches(msg, m.keys.Prev):
		m.focus = Cycle(m.focus, -1)
		return m, nil
	case key.Matches(msg, m.keys.Mode):
		return m.toggleMode(), nil
	case key.Matches(msg, m.keys.Rerun):
		return m.rerun()
	case key.Matches(msg, m.keys.Cancel):
		return m.stopRun(), nil
	}
	switch m.focus {
	case PaneFiles:
		return m.onFiles(msg)
	case PaneHistory:
		return m.onHistory(msg)
	case PaneBuild:
		return m.onBuild(msg)
	default:
		return m.onMain(msg)
	}
}

func (m Model) onFiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		return m.moveFile(-1), nil
	case key.Matches(msg, m.keys.Down):
		return m.moveFile(1), nil
	case key.Matches(msg, m.keys.PageUp):
		return m.moveFile(-m.filePage()), nil
	case key.Matches(msg, m.keys.PageDown):
		return m.moveFile(m.filePage()), nil
	case key.Matches(msg, m.keys.Left):
		if m.tree != nil {
			m.tree.Collapse(m.fileCursor)
		}
		return m, nil
	case key.Matches(msg, m.keys.Right):
		if m.tree != nil {
			m.tree.Expand(m.fileCursor)
		}
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		return m.openSelection()
	default:
		return m, nil
	}
}

func (m Model) onHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		return m.moveCommit(-1), nil
	case key.Matches(msg, m.keys.Down):
		return m.moveCommit(1), nil
	case key.Matches(msg, m.keys.PageUp):
		return m.moveCommit(-m.historyPage()), nil
	case key.Matches(msg, m.keys.PageDown):
		return m.moveCommit(m.historyPage()), nil
	case key.Matches(msg, m.keys.Enter):
		return m.openCommit()
	default:
		return m, nil
	}
}

func (m Model) onBuild(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		return m.moveCmd(-1), nil
	case key.Matches(msg, m.keys.Down):
		return m.moveCmd(1), nil
	case key.Matches(msg, m.keys.PageUp):
		m = m.syncSizes()
		m.output.PageUp()
		m.follow = m.output.AtBottom()
		return m, nil
	case key.Matches(msg, m.keys.PageDown):
		m = m.syncSizes()
		m.output.PageDown()
		m.follow = m.output.AtBottom()
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		cmds := m.commands()
		if m.cmdCursor < 0 || m.cmdCursor >= len(cmds) {
			return m, nil
		}
		return m.launch(cmds[m.cmdCursor])
	default:
		return m, nil
	}
}

func (m Model) onMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.previewKind == previewNone {
		return m, nil
	}
	m = m.syncSizes()
	switch {
	case key.Matches(msg, m.keys.Up):
		m.preview.ScrollUp(1)
	case key.Matches(msg, m.keys.Down):
		m.preview.ScrollDown(1)
	case key.Matches(msg, m.keys.PageUp):
		m.preview.PageUp()
	case key.Matches(msg, m.keys.PageDown):
		m.preview.PageDown()
	}
	return m, nil
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
	}
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
	m.preview.Height = lay.Main.H - 3
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
		m.preview.SetContent("")
	case m.previewNumbered:
		m.previewLines = numberedLines(m.previewRaw, m.preview.Width)
		m.preview.SetContent(strings.Join(m.previewLines, "\n"))
	default:
		m.previewLines = wrappedLines(m.previewRaw, m.preview.Width)
		m.preview.SetContent(strings.Join(m.previewLines, "\n"))
	}
	if reset {
		m.preview.GotoTop()
	} else if y > 0 {
		m.preview.SetYOffset(y)
	}
	return m
}

func (m Model) filePage() int {
	lay := computeLayout(m.width, m.height)
	h := lay.Files.H - 3
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) historyPage() int {
	lay := computeLayout(m.width, m.height)
	h := lay.History.H - 3
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
