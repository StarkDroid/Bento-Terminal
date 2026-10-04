package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/trishiraj/bentro/internal/build"
	"github.com/trishiraj/bentro/internal/files"
	"github.com/trishiraj/bentro/internal/git"
)

type treeLoadedMsg struct {
	tree *files.Tree
	err  error
}

type gitLoadedMsg struct {
	snap git.Snapshot
}

type buildLoadedMsg struct {
	result build.Result
}

type runStartedMsg struct {
	id     int
	cancel func()
	events <-chan build.Event
	err    error
}

type runLineMsg struct {
	id     int
	line   string
	events <-chan build.Event
}

type runDoneMsg struct {
	id  int
	err error
}

type refreshMsg struct{}

func loadTreeCmd(root string) tea.Cmd {
	return func() tea.Msg {
		tree, err := files.Load(root)
		return treeLoadedMsg{tree: tree, err: err}
	}
}

func loadGitCmd(root string) tea.Cmd {
	return func() tea.Msg {
		return gitLoadedMsg{snap: git.Load(root)}
	}
}

func loadBuildCmd(root string) tea.Cmd {
	return func() tea.Msg {
		return buildLoadedMsg{result: build.Inspect(root)}
	}
}

func refreshCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return refreshMsg{}
	})
}

func startCmd(id int, root string, c build.Command) tea.Cmd {
	return func() tea.Msg {
		cancel, events, err := build.Start(root, c.Argv)
		return runStartedMsg{id: id, cancel: cancel, events: events, err: err}
	}
}

func listenCmd(id int, events <-chan build.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-events
		if !ok {
			return nil
		}
		if ev.Done {
			return runDoneMsg{id: id, err: ev.Err}
		}
		return runLineMsg{id: id, line: ev.Line, events: events}
	}
}
