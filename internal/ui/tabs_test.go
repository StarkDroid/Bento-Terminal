package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/StarkDroid/Bento-Terminal/internal/git"
)

func changesSample(t *testing.T) (Model, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := readyEmpty(dir)
	m.width = 120
	m.height = 32
	m.gitSnap = git.Snapshot{
		Repo: true,
		Status: git.Status{
			Branch: "main",
			Files: []git.FileStatus{
				{Path: "main.go", Index: ' ', Worktree: 'M'},
				{Path: "note.md", Index: '?', Worktree: '?', Untracked: true},
			},
		},
	}
	return m, dir
}

func TestHistoryChangesTabListsFiles(t *testing.T) {
	m, _ := changesSample(t)
	m.focus = PaneHistory
	m = press(t, m, "right")
	if m.historyTab != histTabChanges {
		t.Fatal("right should select the Changes tab")
	}
	plain := stripANSI(m.View())
	for _, want := range []string{"History", "Changes", "main.go", "note.md", " M", "??"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("changes tab missing %q\n%s", want, plain)
		}
	}
	m = press(t, m, "down")
	if m.changeCursor != 1 {
		t.Fatalf("down should move the change cursor, got %d", m.changeCursor)
	}
	m = press(t, m, "left")
	if m.historyTab != histTabCommits {
		t.Fatal("left should return to the commits tab")
	}
}

func TestChangesEnterShowsDiff(t *testing.T) {
	m, _ := changesSample(t)
	m.focus = PaneHistory
	m = press(t, m, "right")
	m = press(t, m, "down")
	next, _ := m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.previewKind != previewDiff || m.previewTitle != "note.md" {
		t.Fatalf("preview %q kind %d", m.previewTitle, m.previewKind)
	}
	plain := stripANSI(m.View())
	for _, want := range []string{"± note.md", "+++ b/note.md", "+hello", "+world"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("diff missing %q\n%s", want, plain)
		}
	}
}

func TestHistoryTabClickSwitches(t *testing.T) {
	m, _ := changesSample(t)
	lay := computeLayout(m.width, m.height)
	next, _ := m.Update(tea.MouseMsg{
		X:      lay.Main.W + 2 + 13,
		Y:      lay.HistoryBodyTop() - 1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if m.historyTab != histTabChanges || m.focus != PaneHistory {
		t.Fatalf("tab %d focus %d", m.historyTab, m.focus)
	}
	next, _ = m.Update(tea.MouseMsg{
		X:      lay.Main.W + 2 + 3,
		Y:      lay.HistoryBodyTop() - 1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if m.historyTab != histTabCommits {
		t.Fatal("clicking History should show commits again")
	}
}

func TestTerminalPanelClickFocuses(t *testing.T) {
	m, _ := previewSample(t)
	lay := computeLayout(m.width, m.height)
	next, _ := m.Update(tea.MouseMsg{
		X:      lay.Build.W + 3,
		Y:      lay.TerminalBodyTop() - 1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if m.focus != PaneTerminal {
		t.Fatalf("focus %d", m.focus)
	}
	plain := stripANSI(m.View())
	for _, want := range []string{"Terminal", "Type a command and press enter.", "enter run", "ctrl+c stop"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("terminal missing %q\n%s", want, plain)
		}
	}
}

func TestTerminalPanelRunsCommands(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneTerminal
	next, cmd := m.Update(keyMsg("q"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("q should be typed into the terminal, not quit")
	}
	if m.termInput != "q" {
		t.Fatalf("typed input %q", m.termInput)
	}
	m = press(t, m, "esc")
	if m.termInput != "" || m.focus != PaneTerminal {
		t.Fatal("esc should clear the input before leaving the panel")
	}
	m = press(t, m, "echo bento-term")
	next, cmd = m.Update(keyMsg("enter"))
	m = drive(t, next.(Model), cmd)
	if m.termState != stateOK {
		t.Fatalf("terminal state %d output %#v", m.termState, m.termLines)
	}
	joined := strings.Join(m.termLines, "\n")
	if !strings.Contains(joined, "bento-term") || !strings.Contains(joined, "✓ done") {
		t.Fatalf("terminal output %q", joined)
	}
	if len(m.outLines) != 0 {
		t.Fatalf("build output should stay separate: %#v", m.outLines)
	}
	if m.termInput != "" {
		t.Fatalf("running should clear the prompt, got %q", m.termInput)
	}
	m = press(t, m, "esc")
	if m.focus != PaneBuild {
		t.Fatal("esc on an empty prompt should return to the build panel")
	}
}

func TestTerminalHistoryRecallsCommands(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneTerminal
	m.termHistory = []string{"echo one", "echo two"}
	m.termHistPos = -1
	m = press(t, m, "up")
	if m.termInput != "echo two" || m.termCursor != len([]rune("echo two")) {
		t.Fatalf("up recalled %q cursor %d", m.termInput, m.termCursor)
	}
	m = press(t, m, "up")
	if m.termInput != "echo one" {
		t.Fatalf("second up recalled %q", m.termInput)
	}
	m = press(t, m, "down")
	if m.termInput != "echo two" {
		t.Fatalf("down recalled %q", m.termInput)
	}
	m = press(t, m, "down")
	if m.termInput != "" {
		t.Fatalf("past the newest entry should clear the input, got %q", m.termInput)
	}
}

func TestPanelViewsKeepWindowSize(t *testing.T) {
	m, _ := previewSample(t)
	for _, size := range []struct{ w, h int }{{72, 22}, {80, 24}, {120, 32}} {
		m.width, m.height = size.w, size.h
		views := []string{m.View()}
		m.newFileOpen = true
		views = append(views, m.View())
		m.newFileOpen = false
		m.previewEditing = true
		views = append(views, m.View())
		m.previewEditing = false
		for i, view := range views {
			if w, h := lipgloss.Width(view), lipgloss.Height(view); w != size.w || h != size.h {
				t.Fatalf("%dx%d view %d rendered %dx%d", size.w, size.h, i, w, h)
			}
		}
	}
}

func TestFooterHintsCoverTabsAndEditing(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	if plain := stripANSI(m.View()); !strings.Contains(plain, "e edit") {
		t.Fatalf("file footer missing the edit hint\n%s", plain)
	}
	m = press(t, m, "e")
	if plain := stripANSI(m.View()); !strings.Contains(plain, "ctrl+s save") {
		t.Fatalf("editor footer missing the save hint\n%s", plain)
	}
	m = press(t, m, "esc")
	m.focus = PaneHistory
	if plain := stripANSI(m.View()); !strings.Contains(plain, "←→ tab") {
		t.Fatalf("history footer missing the tab hint\n%s", plain)
	}
}

func TestChangeRowClickShowsDiff(t *testing.T) {
	m, _ := changesSample(t)
	m.focus = PaneHistory
	m.historyTab = histTabChanges
	m.gitReady = true
	lay := computeLayout(m.width, m.height)

	next, _ := m.Update(tea.MouseMsg{
		X:      lay.Main.W + 3,
		Y:      lay.HistoryBodyTop() + 1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if m.changeCursor != 1 {
		t.Fatalf("click selected %d", m.changeCursor)
	}
	if m.previewKind != previewDiff || m.previewTitle != "note.md" {
		t.Fatalf("preview %q kind %d", m.previewTitle, m.previewKind)
	}
	if m.focus != PaneHistory {
		t.Fatalf("click should keep browsing focus in History, got %d", m.focus)
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "+hello") {
		t.Fatalf("diff body missing\n%s", plain)
	}
}

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestOpenDiffFollowsTheWorkingTree(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("one\n")
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "config", "user.email", "test@example.com")
	gitIn(t, dir, "config", "user.name", "Test")
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-qm", "init")

	write("one\ntwo\n")
	m := readyEmpty(dir)
	m.width, m.height = 120, 32
	m.gitReady = true
	m.gitSnap = git.Load(dir)
	m.historyTab = histTabChanges
	for i, f := range m.gitSnap.Status.Files {
		if f.Path == "a.txt" {
			m.changeCursor = i
		}
	}
	next, _ := m.openChange()
	m = next.(Model)
	if !strings.Contains(m.previewRaw, "+two") {
		t.Fatalf("initial diff %q", m.previewRaw)
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "± a.txt") {
		t.Fatalf("diff title missing\n%s", plain)
	}

	write("one\ntwo\nthree\n")
	next, cmd := m.Update(gitLoadedMsg{snap: git.Load(dir)})
	m = drive(t, next.(Model), cmd)
	if !strings.Contains(m.previewRaw, "+three") {
		t.Fatalf("diff did not refresh: %q", m.previewRaw)
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "+2") {
		t.Fatalf("diff title should count additions\n%s", plain)
	}

	gitIn(t, dir, "checkout", "--", "a.txt")
	next, cmd = m.Update(gitLoadedMsg{snap: git.Load(dir)})
	m = drive(t, next.(Model), cmd)
	if m.previewRaw != "(no changes to show)" {
		t.Fatalf("reverted file should clear the diff, got %q", m.previewRaw)
	}
}
