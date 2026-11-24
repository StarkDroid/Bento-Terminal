package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hinshun/vt10x"
)

func TestPaneNumberKeysJump(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain

	next, _ := m.Update(keyMsg("5"))
	m = next.(Model)
	if m.focus != PaneBuild {
		t.Fatalf("5 should focus Build, got %d", m.focus)
	}

	next, _ = m.Update(keyMsg("3"))
	m = next.(Model)
	if m.focus != PaneFiles {
		t.Fatalf("3 should focus Files, got %d", m.focus)
	}

	next, _ = m.Update(keyMsg("2"))
	m = next.(Model)
	if m.focus != PaneAI || !m.aiOpen {
		t.Fatalf("2 should open and focus AI: open=%v focus=%d", m.aiOpen, m.focus)
	}

	next, _ = m.Update(keyMsg("1"))
	m = next.(Model)
	if m.focus != PaneMain {
		t.Fatalf("1 should focus Main, got %d", m.focus)
	}
}

func TestPaneNumbersTypeInsteadOfJumpingWhileEditing(t *testing.T) {
	m, _ := previewSample(t)
	next, _ := m.Update(keyMsg("e"))
	m = next.(Model)
	if !m.previewEditing {
		t.Fatal("e should start editing")
	}
	next, _ = m.Update(keyMsg("5"))
	m = next.(Model)
	if m.focus != PaneMain {
		t.Fatalf("a digit should type, not jump, while editing: focus=%d", m.focus)
	}
}

func TestAIAgentKeepsEscapeKeys(t *testing.T) {
	m, _ := previewSample(t)
	m.aiOpen = true
	m.focus = PaneAI
	fake := &fakeSession{term: vt10x.New(vt10x.WithSize(40, 10))}
	m.aiSession = fake
	m.aiSessionID = "claude"
	m.aiCols, m.aiScreenH = 40, 10

	next, _ := m.Update(keyMsg("tab"))
	m = next.(Model)
	if m.focus != PaneFiles {
		t.Fatalf("tab should leave the AI panel, focus=%d", m.focus)
	}
	if fake.written.Len() != 0 {
		t.Fatalf("tab should not reach the agent: %q", fake.written.String())
	}

	m.focus = PaneAI
	next, _ = m.Update(keyMsg("4"))
	m = next.(Model)
	if m.focus != PaneHistory {
		t.Fatalf("4 should leave the AI panel, focus=%d", m.focus)
	}

	m.focus = PaneAI
	next, _ = m.Update(keyMsg("q"))
	m = next.(Model)
	if fake.written.String() != "q" {
		t.Fatalf("q should still reach the agent, got %q", fake.written.String())
	}
}

func TestAIClickForwardsSGRMouse(t *testing.T) {
	m, _ := previewSample(t)
	m.aiOpen = true
	m.focus = PaneMain
	term := vt10x.New(vt10x.WithSize(40, 10))
	if _, err := term.Write([]byte("\x1b[?1006h")); err != nil {
		t.Fatal(err)
	}
	fake := &fakeSession{term: term}
	m.aiSession = fake
	m.aiSessionID = "claude"
	m.aiCols, m.aiScreenH = 40, 10
	lay := computeLayout(m.width, m.height)
	previewW, _ := lay.MainSplit()

	next, _ := m.Update(tea.MouseMsg{
		X:      previewW + 4,
		Y:      4,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = next.(Model)
	if m.focus != PaneAI {
		t.Fatalf("click should focus the AI panel, got %d", m.focus)
	}
	press := fake.written.String()
	if !strings.HasPrefix(press, "\x1b[<0;") || !strings.HasSuffix(press, "M") {
		t.Fatalf("press wrote %q, want an SGR press", press)
	}
	fake.written.Reset()

	next, _ = m.Update(tea.MouseMsg{
		X:      previewW + 4,
		Y:      4,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
	})
	m = next.(Model)
	if got := fake.written.String(); !strings.HasPrefix(got, "\x1b[<0;") || !strings.HasSuffix(got, "m") {
		t.Fatalf("release wrote %q, want an SGR release", got)
	}
}

func TestTerminalCloseTab(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneTerminal
	m.termInput = "first"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m = next.(Model)
	m.termInput = "second"

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	m = next.(Model)
	if len(m.termTabs) != 1 || m.termIndex != 0 {
		t.Fatalf("tabs=%d index=%d", len(m.termTabs), m.termIndex)
	}
	if m.termInput != "first" {
		t.Fatalf("closing should restore the previous tab, input=%q", m.termInput)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	m = next.(Model)
	if len(m.termTabs) != 1 {
		t.Fatalf("the last tab should be kept, got %d", len(m.termTabs))
	}
	if m.termInput != "" || len(m.termLines) != 0 {
		t.Fatalf("closing the last tab should clear it: input=%q lines=%d", m.termInput, len(m.termLines))
	}
}

func TestTerminalCloseTabStopsItsRun(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneTerminal
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m = next.(Model)
	m.running = true
	m.runTarget = targetTerminal
	m.termRunIndex = m.termIndex
	cancelled := false
	m.cancel = func() { cancelled = true }

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	m = next.(Model)
	if !cancelled || !m.cancelRequested {
		t.Fatal("closing the running shell tab should cancel its command")
	}
	if m.termRunIndex != -1 {
		t.Fatalf("a closed run should not point at another tab: %d", m.termRunIndex)
	}
}

func TestFileRenameFromFilesPane(t *testing.T) {
	m, dir := previewSample(t)
	m.focus = PaneFiles
	selectRow(t, &m, "main.go")

	next, _ := m.Update(keyMsg("R"))
	m = next.(Model)
	if !m.fileOpOpen || m.fileOpKind != fileOpRename {
		t.Fatalf("R should open rename: open=%v kind=%d", m.fileOpOpen, m.fileOpKind)
	}
	if m.fileOpInput != "main.go" {
		t.Fatalf("rename should prefill the path, got %q", m.fileOpInput)
	}
	m.fileOpInput = "cmd/app.go"

	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.fileOpOpen {
		t.Fatalf("rename should close on enter: %q", m.fileOpErr)
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "app.go")); err != nil {
		t.Fatalf("renamed file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Fatalf("old file still there: %v", err)
	}
	if m.previewPath != "cmd/app.go" {
		t.Fatalf("preview should follow the rename, got %q", m.previewPath)
	}
}

func TestFileCopyFromFilesPane(t *testing.T) {
	m, dir := previewSample(t)
	m.focus = PaneFiles
	selectRow(t, &m, "main.go")

	next, _ := m.Update(keyMsg("y"))
	m = next.(Model)
	if !m.fileOpOpen || m.fileOpKind != fileOpCopy {
		t.Fatalf("y should open copy: open=%v kind=%d", m.fileOpOpen, m.fileOpKind)
	}
	m.fileOpInput = "main_copy.go"

	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.fileOpErr != "" {
		t.Fatalf("copy failed: %q", m.fileOpErr)
	}
	if _, err := os.Stat(filepath.Join(dir, "main_copy.go")); err != nil {
		t.Fatalf("copied file missing: %v", err)
	}
	if m.previewPath != "main.go" {
		t.Fatalf("copy should keep the preview, got %q", m.previewPath)
	}
}

func TestFileDeleteNeedsConfirmation(t *testing.T) {
	m, dir := previewSample(t)
	m.focus = PaneFiles
	selectRow(t, &m, "sub")

	next, _ := m.Update(keyMsg("d"))
	m = next.(Model)
	if !m.fileOpOpen || m.fileOpKind != fileOpDelete {
		t.Fatalf("d should open delete: open=%v kind=%d", m.fileOpOpen, m.fileOpKind)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub")); err != nil {
		t.Fatalf("delete should wait for confirmation: %v", err)
	}

	next, _ = m.Update(keyMsg("n"))
	m = next.(Model)
	if m.fileOpOpen {
		t.Fatal("n should cancel the delete")
	}
	if _, err := os.Stat(filepath.Join(dir, "sub")); err != nil {
		t.Fatalf("cancelled delete removed the directory: %v", err)
	}

	next, _ = m.Update(keyMsg("d"))
	m = next.(Model)
	next, _ = m.Update(keyMsg("y"))
	m = next.(Model)
	if _, err := os.Stat(filepath.Join(dir, "sub")); !os.IsNotExist(err) {
		t.Fatalf("confirmed delete left the directory: %v", err)
	}
}

func TestFileOpsOnlyRunInFilesPane(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	next, _ := m.Update(keyMsg("d"))
	m = next.(Model)
	if m.fileOpOpen {
		t.Fatal("d should not open delete outside the Files pane")
	}
}

func TestFileOpPromptsRender(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneFiles
	selectRow(t, &m, "main.go")

	m = m.openFileOp(fileOpRename)
	plain := stripANSI(m.View())
	for _, want := range []string{"R Rename", "main.go", "enter rename", "esc cancel"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("rename prompt missing %q\n%s", want, plain)
		}
	}

	m.fileOpOpen = false
	m = m.openFileOp(fileOpDelete)
	plain = stripANSI(m.View())
	for _, want := range []string{"D Delete", "Delete main.go", "enter or y delete"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("delete prompt missing %q\n%s", want, plain)
		}
	}
}

// selectRow puts the file cursor on a tree row by name.
func selectRow(t *testing.T, m *Model, name string) {
	t.Helper()
	if m.tree == nil {
		t.Fatal("tree is not loaded")
	}
	for i, row := range m.tree.Rows() {
		if row.Name == name {
			m.fileCursor = i
			return
		}
	}
	t.Fatalf("row %q not found", name)
}
