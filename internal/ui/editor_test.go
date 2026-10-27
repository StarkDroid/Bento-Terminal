package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// press sends one key and returns the updated model without running commands.
func press(t *testing.T, m Model, key string) Model {
	t.Helper()
	next, _ := m.Update(keyMsg(key))
	return next.(Model)
}

func TestFileTitleHasNewFileShortcut(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "≡ main.go") {
		t.Fatalf("title missing the file name\n%s", plain)
	}
	if !strings.Contains(plain, "+") {
		t.Fatalf("title missing the new-file shortcut\n%s", plain)
	}
}

func TestNewFileFromPlusPromptSavesAndHighlights(t *testing.T) {
	m, dir := previewSample(t)
	m.focus = PaneMain
	lay := computeLayout(m.width, m.height)
	next, _ := m.Update(tea.MouseMsg{
		X:      lay.Main.W - 3,
		Y:      1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if !m.newFileOpen {
		t.Fatal("the + title shortcut should open the new-file prompt")
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "+ New file") {
		t.Fatalf("prompt missing its title\n%s", plain)
	}

	for _, r := range "pkg/hello.py" {
		m = press(t, m, string(r))
	}
	if m.newFilePath != "pkg/hello.py" {
		t.Fatalf("typed path %q", m.newFilePath)
	}
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.newFileOpen {
		t.Fatal("enter should close the prompt")
	}
	if !m.previewEditing || m.previewPath != "pkg/hello.py" {
		t.Fatalf("new file should open in the editor: editing %v path %q", m.previewEditing, m.previewPath)
	}
	if _, err := os.Stat(filepath.Join(dir, "pkg", "hello.py")); err != nil {
		t.Fatalf("new file was not created: %v", err)
	}
	if m.previewLang.Name != "Python" {
		t.Fatalf("new .py file should detect Python, got %+v", m.previewLang)
	}

	m = press(t, m, "def hi():")
	m = press(t, m, "enter")
	m = press(t, m, "    print(1)")
	next, _ = m.Update(keyMsg("ctrl+s"))
	m = next.(Model)
	if m.previewDirty {
		t.Fatal("save should clear the dirty flag")
	}
	if !m.previewSaved {
		t.Fatal("save should mark the file saved")
	}
	if len(m.previewSpans) == 0 {
		t.Fatal("saving a Python file should refresh its highlighting")
	}
	body, err := os.ReadFile(filepath.Join(dir, "pkg", "hello.py"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "def hi():\n    print(1)\n" {
		t.Fatalf("saved %q", got)
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "print(1)") {
		t.Fatalf("preview missing the saved line\n%s", plain)
	}
}

func TestEditorKeysChangeTheBuffer(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	m = press(t, m, "e")
	if !m.previewEditing {
		t.Fatal("e should start editing")
	}
	m = press(t, m, "X")
	if got := m.previewEditLines[2]; got != "Xfunc alpha() {" {
		t.Fatalf("typed line %q", got)
	}
	m = press(t, m, "enter")
	if m.previewCursor != 3 || m.previewEditLines[2] != "X" {
		t.Fatalf("enter should split at the caret: line %d %+v", m.previewCursor, m.previewEditLines[:4])
	}
	m = press(t, m, "backspace")
	if m.previewCursor != 2 || m.previewEditLines[2] != "Xfunc alpha() {" {
		t.Fatalf("backspace should join the line: line %d %q", m.previewCursor, m.previewEditLines[2])
	}
	m.previewCol = 0
	m = press(t, m, "[")
	if m.previewFolds[2] {
		t.Fatal("[ should type instead of folding while editing")
	}
	if !strings.HasPrefix(m.previewEditLines[2], "[X") {
		t.Fatalf("bracket was not typed: %q", m.previewEditLines[2])
	}
	m = press(t, m, "esc")
	if m.previewEditing {
		t.Fatal("esc should leave the editor")
	}
	if !m.previewDirty {
		t.Fatal("unsaved edits should mark the preview dirty")
	}
	if !strings.Contains(m.previewRaw, "[Xfunc") {
		t.Fatalf("preview text should keep unsaved edits: %q", m.previewRaw)
	}
}

func TestEditorHomeEndAndDelete(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	m = press(t, m, "e")
	m.previewCursor, m.previewCol = 2, 0
	m = press(t, m, "end")
	if got := m.previewCol; got != len([]rune("func alpha() {")) {
		t.Fatalf("end moved to %d", got)
	}
	m = press(t, m, "home")
	if m.previewCol != 0 {
		t.Fatalf("home moved to %d", m.previewCol)
	}
	m = press(t, m, "delete")
	if got := m.previewEditLines[2]; got != "unc alpha() {" {
		t.Fatalf("delete should remove the first character: %q", got)
	}
}

func TestNewFilePromptRejectsEscapedPaths(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	m = press(t, m, "n")
	if !m.newFileOpen {
		t.Fatal("n should open the prompt")
	}
	for _, r := range "../escape.go" {
		m = press(t, m, string(r))
	}
	next, _ := m.Update(keyMsg("enter"))
	m = next.(Model)
	if !m.newFileOpen || m.newFileErr == "" {
		t.Fatalf("prompt should stay open with an error: %q", m.newFileErr)
	}
}
