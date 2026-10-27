package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
	"github.com/StarkDroid/Bento-Terminal/internal/files"
)

func TestViewEmptyStates(t *testing.T) {
	dir := t.TempDir()
	m := readyEmpty(dir)
	m.width = 100
	m.height = 32
	view := m.View()
	if lipgloss.Width(view) != 100 || lipgloss.Height(view) != 32 {
		t.Fatalf("size %dx%d\n%s", lipgloss.Width(view), lipgloss.Height(view), view)
	}
	plain := stripANSI(view)
	for _, want := range []string{"≡ Files", "History", "Changes", "Build", "Terminal", "Empty directory", "Not a git repository", "No commands detected", "Idle. Select a command and press enter.", "bento"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("missing %q\n%s", want, plain)
		}
	}
}

func TestViewTooSmall(t *testing.T) {
	m := New(t.TempDir(), "test")
	m.width = 40
	m.height = 12
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "larger") {
		t.Fatalf("%s", plain)
	}
}

func TestFileSelectionAndPreview(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tree, err := files.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := readyEmpty(dir)
	m.tree = tree
	m.treeReady = true
	m.width = 100
	m.height = 32
	m.focus = PaneFiles
	if tree.Rows()[0].Name != "sub" {
		t.Fatalf("rows %+v", tree.Rows())
	}
	next, _ := m.Update(keyMsg("j"))
	m = next.(Model)
	if m.fileCursor != 1 {
		t.Fatalf("cursor %d", m.fileCursor)
	}
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.previewKind != previewFile || !strings.Contains(m.previewRaw, "alpha") {
		t.Fatalf("preview %q kind %d", m.previewRaw, m.previewKind)
	}
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "a.txt") || !strings.Contains(plain, "alpha") {
		t.Fatalf("view missing preview\n%s", plain)
	}
	if !strings.Contains(plain, "1") {
		t.Fatal("expected a line number")
	}
}

func TestToggleBuildMode(t *testing.T) {
	m := New(t.TempDir(), "test")
	m.autoCmds = []build.Command{{Name: "go test ./...", Argv: []string{"go", "test", "./..."}}}
	m.manualCmds = []build.Command{{Name: "go build ./...", Argv: []string{"go", "build", "./..."}}}
	m.mode = build.ModeAuto
	m.buildReady = true
	next, _ := m.Update(keyMsg("a"))
	m = next.(Model)
	if m.mode != build.ModeManual {
		t.Fatal("expected manual mode")
	}
	next, _ = m.Update(keyMsg("a"))
	m = next.(Model)
	if m.mode != build.ModeAuto {
		t.Fatal("expected auto mode")
	}
}

func TestRunEchoCommand(t *testing.T) {
	dir := t.TempDir()
	m := readyEmpty(dir)
	m.width = 100
	m.height = 32
	m.focus = PaneBuild
	m.mode = build.ModeManual
	m.manualCmds = []build.Command{{Name: "echo hi", Argv: build.ShellCommand("echo hi").Argv, Source: "test"}}
	m.buildReady = true
	next, cmd := m.Update(keyMsg("enter"))
	m = drive(t, next.(Model), cmd)
	if m.buildState != stateOK {
		t.Fatalf("state %d output %#v", m.buildState, m.outLines)
	}
	joined := strings.Join(m.outLines, "\n")
	if !strings.Contains(joined, "hi") || !strings.Contains(joined, "✓ done") {
		t.Fatalf("%s", joined)
	}
}

func readyEmpty(dir string) Model {
	m := New(dir, "0.2.0")
	m.treeReady = true
	m.gitReady = true
	m.buildReady = true
	tree, err := files.Load(dir)
	if err == nil {
		m.tree = tree
	}
	return m
}

func drive(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	queue := []tea.Cmd{cmd}
	for steps := 0; len(queue) > 0 && steps < 40; steps++ {
		c := queue[0]
		queue = queue[1:]
		if c == nil {
			continue
		}
		msg := execCmd(t, c)
		if msg == nil {
			continue
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			queue = append(queue, []tea.Cmd(batch)...)
			continue
		}
		next, follow := m.Update(msg)
		m = next.(Model)
		if _, ok := msg.(spinner.TickMsg); ok && !m.running {
			continue
		}
		if follow != nil {
			queue = append(queue, follow)
		}
	}
	return m
}

func execCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(5 * time.Second):
		t.Fatal("command timed out")
		return nil
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != 0x1b {
			b.WriteByte(s[i])
			continue
		}
		j := i + 1
		if j < len(s) && s[j] == '[' {
			j++
			for j < len(s) && s[j] != 'm' {
				j++
			}
		}
		i = j
	}
	return b.String()
}
