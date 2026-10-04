// Package ui is the full-screen interface for Bento terminal.
package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
	"github.com/StarkDroid/Bento-Terminal/internal/files"
	"github.com/StarkDroid/Bento-Terminal/internal/git"
)

type previewKind int

const (
	previewNone previewKind = iota
	previewFile
	previewCommit
)

type runState int

const (
	stateIdle runState = iota
	stateRunning
	stateOK
	stateFailed
	stateCanceled
)

// Model is the Bubble Tea model for Bento terminal.
type Model struct {
	root          string
	width, height int
	focus         Pane
	helpOpen      bool

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

	mode       build.Mode
	autoCmds   []build.Command
	manualCmds []build.Command
	systems    []string
	buildReady bool
	cmdCursor  int
	cmdScroll  int

	running         bool
	runID           int
	cancel          func()
	cancelRequested bool
	last            *build.Command
	buildState      runState
	outLines        []string
	output          viewport.Model
	follow          bool

	previewKind     previewKind
	previewTitle    string
	previewRaw      string
	previewNumbered bool
	preview         viewport.Model
	previewLines    []string
}

// New returns a TUI model rooted at the project directory.
func New(root string) Model {
	sp := spinner.New(
		spinner.WithSpinner(spinner.Line),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(colorPalm).Background(colorPanel)),
	)
	return Model{
		root:    root,
		focus:   PaneFiles,
		keys:    newKeys(),
		spinner: sp,
		mode:    build.ModeAuto,
		follow:  true,
		preview: viewport.New(0, 0),
		output:  viewport.New(0, 0),
	}
}

// Init loads the tree, git status, and detected commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadTreeCmd(m.root),
		loadGitCmd(m.root),
		loadBuildCmd(m.root),
		m.spinner.Tick,
		refreshCmd(),
	)
}

// Focus returns the pane that receives movement keys.
func (m Model) Focus() Pane { return m.focus }
