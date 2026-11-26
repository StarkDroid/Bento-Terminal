package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/StarkDroid/Bento-Terminal/internal/agent"
	"github.com/StarkDroid/Bento-Terminal/internal/build"
	"github.com/StarkDroid/Bento-Terminal/internal/files"
	"github.com/StarkDroid/Bento-Terminal/internal/git"
)

type treeLoadedMsg struct {
	tree *files.Tree
	err  error
}

type gitLoadedMsg struct {
	snap git.Snapshot
}

type diffLoadedMsg struct {
	path string
	body string
	err  error
}

type buildLoadedMsg struct {
	result build.Result
}

type filesListedMsg struct {
	files []string
	err   error
}

type agentsDetectedMsg struct {
	found map[string]string
}

func detectAgentsCmd() tea.Cmd {
	return func() tea.Msg {
		return agentsDetectedMsg{found: agent.Detect()}
	}
}

func waitAICmd(events <-chan aiEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return nil
		}
		return aiEventMsg{event: event}
	}
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

func loadDiffCmd(root string, f git.FileStatus) tea.Cmd {
	return func() tea.Msg {
		body, err := git.Diff(root, f)
		return diffLoadedMsg{path: f.Path, body: body, err: err}
	}
}

func loadFilesListCmd(root string) tea.Cmd {
	return func() tea.Msg {
		files, err := files.List(root)
		return filesListedMsg{files: files, err: err}
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
