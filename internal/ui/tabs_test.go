package ui

import (
	"os"
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

func TestBuildTabClickSelectsTerminal(t *testing.T) {
	m, _ := previewSample(t)
	lay := computeLayout(m.width, m.height)
	next, _ := m.Update(tea.MouseMsg{
		X:      1 + 11,
		Y:      lay.BuildBodyTop() - 1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if m.buildTab != buildTabTerminal || m.focus != PaneBuild {
		t.Fatalf("tab %d focus %d", m.buildTab, m.focus)
	}
	plain := stripANSI(m.View())
	for _, want := range []string{"Terminal", "Type a command and press enter.", "enter run", "ctrl+c stop"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("terminal missing %q\n%s", want, plain)
		}
	}
}

func TestTerminalTabRunsCommands(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneBuild
	m = press(t, m, "right")
	if m.buildTab != buildTabTerminal {
		t.Fatal("right should select the terminal tab")
	}
	next, cmd := m.Update(keyMsg("q"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("q should be typed into the terminal, not quit")
	}
	if m.termInput != "q" {
		t.Fatalf("typed input %q", m.termInput)
	}
	m = press(t, m, "esc")
	if m.termInput != "" || m.buildTab != buildTabTerminal {
		t.Fatal("esc should clear the input before leaving the tab")
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
	if m.buildTab != buildTabCommands {
		t.Fatal("esc on an empty prompt should return to the command list")
	}
}

func TestTerminalHistoryRecallsCommands(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneBuild
	m = press(t, m, "right")
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

func TestTabViewsKeepWindowSize(t *testing.T) {
	m, _ := previewSample(t)
	for _, size := range []struct{ w, h int }{{72, 22}, {80, 24}, {120, 32}} {
		m.width, m.height = size.w, size.h
		m.buildTab = buildTabTerminal
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
