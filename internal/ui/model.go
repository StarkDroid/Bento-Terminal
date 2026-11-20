// Package ui is the full-screen interface for Bento terminal.
package ui

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
	"github.com/StarkDroid/Bento-Terminal/internal/files"
	"github.com/StarkDroid/Bento-Terminal/internal/git"
	appstate "github.com/StarkDroid/Bento-Terminal/internal/state"
	"github.com/StarkDroid/Bento-Terminal/internal/syntax"
	upd "github.com/StarkDroid/Bento-Terminal/internal/update"
)

type previewKind int

const (
	previewNone previewKind = iota
	previewFile
	previewCommit
	previewDiff
)

type runState int

const (
	stateIdle runState = iota
	stateRunning
	stateOK
	stateFailed
	stateCanceled
)

// historyTab selects the commit list or the working-tree changes.
type historyTab int

const (
	histTabCommits historyTab = iota
	histTabChanges
)

// buildTab selects the command list or the last run's output.
type buildTab int

const (
	buildTabCommands buildTab = iota
	buildTabOutput
)

// terminalTab is one terminal buffer. The active tab is mirrored in the
// Model's term* fields so rendering and input keep their old shape.
type terminalTab struct {
	title   string
	lines   []string
	follow  bool
	scroll  int
	input   string
	cursor  int
	history []string
	histPos int
	state   runState
}

// runTarget says where a running process writes its output.
type runTarget int

const (
	targetBuild runTarget = iota
	targetTerminal
	targetAI
)

// Model is the Bubble Tea model for Bento terminal.
type Model struct {
	root          string
	width, height int
	focus         Pane
	paletteOpen   bool
	palettePane   Pane
	paletteCursor int
	settingsOpen  bool
	version       string

	updateRelease   upd.Release
	updateOpen      bool
	updateDismissed bool
	updateRunning   bool
	updateDone      bool
	updateMethod    upd.Method

	keys    keyMap
	spinner spinner.Model

	tree       *files.Tree
	treeReady  bool
	treeErr    string
	fileCursor int
	fileScroll int

	gitSnap      git.Snapshot
	gitReady     bool
	commitCursor int
	commitScroll int
	changeCursor int
	changeScroll int
	historyTab   historyTab

	mode       build.Mode
	autoCmds   []build.Command
	manualCmds []build.Command
	systems    []string
	buildReady bool
	buildTab   buildTab
	cmdCursor  int
	cmdScroll  int

	running         bool
	runID           int
	runTarget       runTarget
	cancel          func()
	cancelRequested bool
	last            *build.Command
	buildState      runState
	termState       runState
	outLines        []string
	output          viewport.Model
	follow          bool
	termTabs        []terminalTab
	termIndex       int
	termRunIndex    int
	termLines       []string
	termOutput      viewport.Model
	termFollow      bool
	termInput       string
	termCursor      int
	termHistory     []string
	termHistPos     int

	aiOpen      bool
	aiCursor    int
	aiScroll    int
	aiFound     map[string]string
	aiReady     bool
	aiErr       string
	aiNote      string
	aiInstall   string
	aiLines     []string
	aiState     runState
	aiSession   aiSession
	aiSessionID string
	aiEvents    chan aiEvent
	aiRun       int
	aiCols      int
	aiScreenH   int

	store            *appstate.Store
	saved            appstate.ProjectState
	restoreScroll    int
	restoreScrollSet bool

	recentOpen   bool
	recent       []appstate.Project
	recentCursor int
	recentScroll int

	previewKind      previewKind
	previewTitle     string
	previewPath      string
	previewRaw       string
	previewNumbered  bool
	previewLang      syntax.Language
	previewSyntax    bool
	previewCursor    int
	previewCol       int
	previewFolds     map[int]bool
	previewBlocks    []syntax.Block
	previewSpans     [][]codeSpan
	previewRows      []previewRow
	previewVisible   []int
	preview          viewport.Model
	previewLines     []string
	previewEditing   bool
	previewEditLines []string
	previewDirty     bool
	previewSaved     bool
	previewSaveErr   bool

	newFileOpen bool
	newFilePath string
	newFileErr  string

	finderOpen        bool
	finderInput       string
	finderCursor      int
	finderInputCursor int
	finderScroll      int
	finderFiles       []string
	finderReady       bool
	finderMatches     []fileMatch
	finderLoading     bool
	finderErr         string

	findOpen    bool
	findInput   string
	findCursor  int
	findMatches []findMatch
	findIndex   int
}

// New returns a TUI model rooted at the project directory.
// New builds a model with no persistent project state. Tests and callers that
// do not need recent projects use this; cmd/bento uses NewWithState.
func New(root, version string) Model {
	return NewWithState(root, version, appstate.NewMemory())
}

// NewWithState builds a model and restores the last file and focus saved for
// root. The store is also given the root as the newest recent project.
func NewWithState(root, version string, store *appstate.Store) Model {
	if store == nil {
		store = appstate.NewMemory()
	}
	sp := spinner.New(
		spinner.WithSpinner(spinner.Line),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(colorPalm).Background(colorPanel)),
	)
	m := Model{
		root:         root,
		version:      version,
		focus:        PaneFiles,
		keys:         newKeys(),
		spinner:      sp,
		mode:         build.ModeAuto,
		follow:       true,
		termTabs:     []terminalTab{{title: "shell 1", follow: true, histPos: -1}},
		termRunIndex: -1,
		termFollow:   true,
		termHistPos:  -1,
		previewFolds: map[int]bool{},
		preview:      viewport.New(0, 0),
		output:       viewport.New(0, 0),
		termOutput:   viewport.New(0, 0),
		store:        store,
		saved:        store.Project(root),
	}
	_ = store.Touch(root)
	return m.restore()
}

// restore reopens the file and focus saved for this project. A missing file or
// an AI focus falls back to a safe pane.
func (m Model) restore() Model {
	ps := m.saved
	if ps.File != "" {
		abs := filepath.Join(m.root, filepath.FromSlash(ps.File))
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			m = m.openFile(ps.File, info.Name())
			if len(m.previewLines) > 0 {
				m.previewCursor = clamp(ps.Line, 0, len(m.previewLines)-1)
				line := m.previewLines[m.previewCursor]
				m.previewCol = clamp(ps.Col, 0, len([]rune(line)))
			}
			if ps.Scroll > 0 {
				m.restoreScroll = ps.Scroll
				m.restoreScrollSet = true
			}
		}
	}
	if p, ok := paneFromName(ps.Focus); ok && p != PaneAI {
		m.focus = p
	} else if m.previewKind == previewFile {
		m.focus = PaneMain
	}
	m.SaveState()
	return m
}

// paneName returns the stable name saved in the state file.
func paneName(p Pane) string {
	switch p {
	case PaneMain:
		return "main"
	case PaneAI:
		return "ai"
	case PaneFiles:
		return "files"
	case PaneHistory:
		return "history"
	case PaneBuild:
		return "build"
	case PaneTerminal:
		return "terminal"
	}
	return ""
}

func paneFromName(name string) (Pane, bool) {
	switch name {
	case "main":
		return PaneMain, true
	case "ai":
		return PaneAI, true
	case "files":
		return PaneFiles, true
	case "history":
		return PaneHistory, true
	case "build":
		return PaneBuild, true
	case "terminal":
		return PaneTerminal, true
	}
	return 0, false
}

// SaveState writes the current file, caret, scroll, and focus for this project.
// It is safe to call on a model with a memory store.
func (m *Model) SaveState() {
	if m == nil || m.store == nil {
		return
	}
	ps := m.saved
	if m.previewKind == previewFile {
		ps.File = m.previewPath
		ps.Line = m.previewCursor
		ps.Col = m.previewCol
		ps.Scroll = m.preview.YOffset
	} else if m.previewKind == previewNone {
		ps.File = ""
		ps.Line = 0
		ps.Col = 0
		ps.Scroll = 0
	}
	ps.Focus = paneName(m.focus)
	m.saved = ps
	_ = m.store.SetProject(m.root, ps)
}

// Init loads the tree, git status, and detected commands.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		loadTreeCmd(m.root),
		loadGitCmd(m.root),
		loadBuildCmd(m.root),
		detectAgentsCmd(),
		m.spinner.Tick,
		refreshCmd(),
	}
	if updateCmd := checkUpdateCmd(m.version); updateCmd != nil {
		cmds = append(cmds, updateCmd)
	}
	return tea.Batch(cmds...)
}

// Focus returns the pane that receives movement keys.
func (m Model) Focus() Pane { return m.focus }
