package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appstate "github.com/StarkDroid/Bento-Terminal/internal/state"
)

const goSample = `package main

func alpha() {
	println("a")
	println("b")
}

func beta() {}
`

func TestFootersShowPaneShortcuts(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	plain := stripANSI(m.View())
	for _, want := range []string{"↑↓←→ move", "[ ] fold", "? keys", "↑↓ pick", "enter run"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("footer missing %q\n%s", want, plain)
		}
	}
	if !strings.Contains(plain, "ctrl+a ai") {
		t.Fatalf("footer should show the AI shortcut\n%s", plain)
	}
}

func TestFooterKeysAndLabelsUseDifferentColours(t *testing.T) {
	for _, focused := range []bool{false, true} {
		key, label := footerColours(focused)
		if key == label {
			t.Fatalf("focused %v: key and label share a colour", focused)
		}
	}
}

func TestCommitPreviewClosesWithEsc(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneHistory
	m.previewKind = previewCommit
	m.previewTitle = "abc1234"
	m.previewRaw = "commit abc1234\n\nfix the thing"
	m = m.reflowPreview(true)

	if !strings.Contains(stripANSI(m.View()), "abc1234") {
		t.Fatal("the commit should be in the preview")
	}
	next, _ := m.Update(keyMsg("esc"))
	m = next.(Model)
	if m.previewKind != previewNone {
		t.Fatalf("esc should close the commit, got kind %d", m.previewKind)
	}
	plain := stripANSI(m.View())
	if strings.Contains(plain, "abc1234") {
		t.Fatalf("the commit should be gone\n%s", plain)
	}
	if !strings.Contains(plain, "tab panes") {
		t.Fatalf("Main should show the home footer\n%s", plain)
	}
}

func TestEscClosesAFilePreview(t *testing.T) {
	m, _ := previewSample(t)
	if m.previewKind != previewFile {
		t.Fatal("the sample should open a file")
	}
	next, _ := m.Update(keyMsg("esc"))
	m = next.(Model)
	if m.previewKind != previewNone {
		t.Fatal("esc should close a file preview")
	}
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "Select a file to preview it.") {
		t.Fatalf("Main should return to the home view\n%s", plain)
	}
}

func TestEscInAIPanelLeavesTheFileOpen(t *testing.T) {
	m, _ := previewSample(t)
	m.aiOpen = true
	m.aiReady = true
	m.focus = PaneAI
	next, _ := m.Update(keyMsg("esc"))
	m = next.(Model)
	if m.previewKind != previewFile {
		t.Fatal("esc in the AI panel should not close the file")
	}
	if m.focus != PaneMain {
		t.Fatalf("esc should return focus to Main, got %d", m.focus)
	}
}

func TestFilePreviewFooterOffersClose(t *testing.T) {
	m, _ := previewSample(t)
	if !strings.Contains(stripANSI(m.View()), "esc close") {
		t.Fatal("the file footer should offer esc close")
	}
}

func TestCommitPreviewFooterOffersClose(t *testing.T) {
	m, _ := previewSample(t)
	m.previewKind = previewCommit
	m.previewTitle = "abc1234"
	m = m.reflowPreview(true)
	if !strings.Contains(stripANSI(m.View()), "esc close") {
		t.Fatal("the commit footer should offer esc close")
	}
}

func TestPaletteListsFocusedPaneCommands(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneFiles
	next, _ := m.Update(keyMsg("?"))
	m = next.(Model)
	plain := stripANSI(m.View())
	for _, want := range []string{"? Keys · Files", "FILES", "GLOBAL", "preview file", "focus next pane", "settings and about"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("palette missing %q\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "switch auto or manual") {
		t.Fatal("files palette should not list build commands")
	}
	next, _ = m.Update(keyMsg("esc"))
	m = next.(Model)
	m.focus = PaneBuild
	next, _ = m.Update(keyMsg("?"))
	m = next.(Model)
	if plain := stripANSI(m.View()); !strings.Contains(plain, "switch auto or manual") {
		t.Fatalf("build palette should list mode\n%s", plain)
	}
}

func TestPaletteRunsListedShortcut(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneMain
	next, _ := m.Update(keyMsg("?"))
	m = next.(Model)
	if !m.paletteOpen {
		t.Fatal("palette should be open")
	}
	next, _ = m.Update(keyMsg("s"))
	m = next.(Model)
	if m.paletteOpen {
		t.Fatal("running a shortcut should close the palette")
	}
	if m.previewSyntax {
		t.Fatal("s should turn highlighting off")
	}
}

func TestFoldGoBlockWithKeys(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneMain
	m.previewCursor = 2
	next, _ := m.Update(keyMsg("["))
	m = next.(Model)
	if !m.previewFolds[2] {
		t.Fatalf("block 2 should be folded: %+v", m.previewFolds)
	}
	for _, row := range m.previewRows {
		if row.line >= 3 && row.line <= 5 {
			t.Fatalf("folded lines should be hidden, found line %d", row.line)
		}
	}
	next, _ = m.Update(keyMsg("]"))
	m = next.(Model)
	if m.previewFolds[2] {
		t.Fatal("] should unfold the block")
	}
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if !m.previewFolds[2] {
		t.Fatal("enter should fold the block again")
	}
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.previewFolds[2] {
		t.Fatal("enter should unfold the block")
	}
}

func TestArrowKeysMovePreviewCaret(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneMain
	m.previewCursor = 2
	m.previewCol = 0
	m = m.reflowPreview(false)

	before := m.previewRows[2].text
	next, _ := m.Update(keyMsg("right"))
	m = next.(Model)
	if m.previewCursor != 2 || m.previewCol != 1 {
		t.Fatalf("right should move one column: line %d col %d", m.previewCursor, m.previewCol)
	}
	if m.previewRows[2].text == before {
		t.Fatal("the caret should redraw the cursor row")
	}

	next, _ = m.Update(keyMsg("down"))
	m = next.(Model)
	if m.previewCursor != 3 {
		t.Fatalf("down should move to line 3, got %d", m.previewCursor)
	}

	m.previewCursor, m.previewCol = 0, 0
	m = m.reflowPreview(false)
	next, _ = m.Update(keyMsg("up"))
	m = next.(Model)
	if m.previewCursor != 0 || m.previewCol != 0 {
		t.Fatalf("up at the first line should stay put: line %d col %d", m.previewCursor, m.previewCol)
	}
	next, _ = m.Update(keyMsg("left"))
	m = next.(Model)
	if m.previewCursor != 0 || m.previewCol != 0 {
		t.Fatalf("left at the start of the file should stay put: line %d col %d", m.previewCursor, m.previewCol)
	}

	m.previewCursor, m.previewCol = 1, 0
	m = m.reflowPreview(false)
	next, _ = m.Update(keyMsg("left"))
	m = next.(Model)
	if m.previewCursor != 0 || m.previewCol != len([]rune("package main")) {
		t.Fatalf("left should wrap to the previous line end: line %d col %d", m.previewCursor, m.previewCol)
	}
	next, _ = m.Update(keyMsg("right"))
	m = next.(Model)
	if m.previewCursor != 1 || m.previewCol != 0 {
		t.Fatalf("right should wrap to the next line start: line %d col %d", m.previewCursor, m.previewCol)
	}
}

func TestFoldGoBlockWithMouse(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneFiles
	next, _ := m.Update(tea.MouseMsg{X: 4, Y: 4, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = next.(Model)
	if m.focus != PaneMain {
		t.Fatalf("click should focus main, got %d", m.focus)
	}
	if !m.previewFolds[2] {
		t.Fatalf("click on the func line should fold it: %+v", m.previewFolds)
	}
	next, _ = m.Update(tea.MouseMsg{X: 4, Y: 4, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = next.(Model)
	if m.previewFolds[2] {
		t.Fatal("second click should unfold")
	}
}

func TestMouseSelectsFile(t *testing.T) {
	m, dir := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneHistory
	row := 0
	for i, r := range m.tree.Rows() {
		if r.Name == "main.go" {
			row = i
		}
	}
	lay := computeLayout(m.width, m.height)
	x := lay.Main.W + 3
	y := lay.FilesBodyTop() + row - m.fileScroll
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = next.(Model)
	if m.focus != PaneFiles {
		t.Fatalf("focus %d", m.focus)
	}
	if m.previewKind != previewFile || m.previewTitle != "main.go" {
		t.Fatalf("preview %q kind %d", m.previewTitle, m.previewKind)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err != nil {
		t.Fatal(err)
	}
}

func TestSyntaxTitleAndToggle(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "·  Go") {
		t.Fatalf("title should name the language\n%s", plain)
	}
	m.focus = PaneMain
	next, _ := m.Update(keyMsg("s"))
	m = next.(Model)
	if m.previewSyntax {
		t.Fatal("s should disable highlighting")
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "·  Go off") {
		t.Fatalf("title should show highlighting off\n%s", plain)
	}
}

func TestPaletteEnterRunsSelectedCommand(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneFiles
	m.fileCursor = 1
	next, _ := m.Update(keyMsg("?"))
	m = next.(Model)
	for i := 0; i < len(m.paneActions(PaneFiles))-1; i++ {
		next, _ = m.Update(keyMsg("down"))
		m = next.(Model)
	}
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.paletteOpen {
		t.Fatal("enter should close the palette")
	}
	if m.previewKind != previewFile || m.previewTitle != "main.go" {
		t.Fatalf("preview %q kind %d", m.previewTitle, m.previewKind)
	}
}

func TestPalettePrefersNavigationOverListedKeys(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneHistory
	next, _ := m.Update(keyMsg("?"))
	m = next.(Model)
	next, _ = m.Update(keyMsg("j"))
	m = next.(Model)
	if m.paletteCursor != 1 {
		t.Fatalf("j should move the palette cursor, got %d", m.paletteCursor)
	}
	if m.commitCursor != 0 {
		t.Fatal("j should not move the history selection while the palette is open")
	}
}

func TestMouseClickTogglesDirectory(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneMain
	lay := computeLayout(m.width, m.height)
	rows := m.tree.Rows()
	dirRow := 0
	for i, r := range rows {
		if r.IsDir {
			dirRow = i
		}
	}
	x := lay.Main.W + 3
	y := lay.FilesBodyTop() + dirRow
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = next.(Model)
	if m.focus != PaneFiles || !m.tree.Rows()[dirRow].Expanded {
		t.Fatalf("click should open the directory: focus %d rows %+v", m.focus, m.tree.Rows())
	}
	next, _ = m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = next.(Model)
	if m.tree.Rows()[dirRow].Expanded {
		t.Fatal("second click should close the directory")
	}
}

func TestSyntaxToggleKeepsFolds(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneMain
	m.previewCursor = 2
	next, _ := m.Update(keyMsg("enter"))
	m = next.(Model)
	next, _ = m.Update(keyMsg("s"))
	m = next.(Model)
	if !m.previewFolds[2] {
		t.Fatal("highlight toggle should keep the fold state")
	}
	if len(m.previewSpans) != 0 {
		t.Fatal("highlight toggle off should drop cached spans")
	}
}

func TestPlainTextHasNoLanguageBadge(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := readyEmpty(dir)
	m.width = 120
	m.height = 32
	m.focus = PaneFiles
	m.fileCursor = 0
	next, _ := m.Update(keyMsg("enter"))
	m = next.(Model)
	m.focus = PaneMain
	plain := stripANSI(m.View())
	if strings.Contains(plain, "notes.txt  ·") {
		t.Fatalf("plain text should not show a language badge\n%s", plain)
	}
	if m.previewLang.Detected() {
		t.Fatalf("plain text should not detect a language: %+v", m.previewLang)
	}
	if len(m.previewBlocks) != 0 {
		t.Fatalf("plain text should not have folds: %+v", m.previewBlocks)
	}
}

func TestSettingsShowsAboutAndVersion(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneFiles
	m.version = "9.9.9"
	next, _ := m.Update(keyMsg(","))
	m = next.(Model)
	if !m.settingsOpen {
		t.Fatal("comma should open settings")
	}
	plain := stripANSI(m.View())
	for _, want := range []string{"Settings · About", "v9.9.9", "github", "? commands", "q quit", "esc close"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("settings missing %q\n%s", want, plain)
		}
	}
	if !strings.Contains(plain, "████") {
		t.Fatal("settings should draw the wordmark")
	}
	next, _ = m.Update(keyMsg("esc"))
	m = next.(Model)
	if m.settingsOpen {
		t.Fatal("esc should close settings")
	}
}

func TestViewSizeWithFooter(t *testing.T) {
	m, _ := previewSample(t)
	for _, size := range []struct{ w, h int }{{120, 32}, {80, 24}, {72, 20}} {
		m.width, m.height = size.w, size.h
		view := m.View()
		if w, h := lipgloss.Width(view), lipgloss.Height(view); w != size.w || h != size.h {
			t.Fatalf("%dx%d rendered %dx%d", size.w, size.h, w, h)
		}
	}
}

// previewSample builds a ready model with main.go open in the preview.
func previewSample(t *testing.T) (Model, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(goSample), 0o644); err != nil {
		t.Fatal(err)
	}
	m := readyEmpty(dir)
	m.width = 120
	m.height = 32
	m.focus = PaneFiles
	i := 0
	for _, r := range m.tree.Rows() {
		if r.Name == "main.go" {
			m.fileCursor = i
		}
		i++
	}
	next, _ := m.Update(keyMsg("enter"))
	m = next.(Model)
	m.focus = PaneMain
	m.previewCursor = 2
	return m, dir
}

func TestClosingAFileClearsSavedState(t *testing.T) {
	m, _ := previewSample(t)
	store := appstate.NewMemory()
	m.store = store
	m.SaveState()
	if got := store.Project(m.root).File; got != "main.go" {
		t.Fatalf("saved file %q", got)
	}
	next, _ := m.Update(keyMsg("esc"))
	m = next.(Model)
	if got := store.Project(m.root).File; got != "" {
		t.Fatalf("closed file should be cleared, got %q", got)
	}
}

func TestRecentProjectSwitchesAndRestores(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "app.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := appstate.NewMemory()
	if err := store.Touch(other); err != nil {
		t.Fatal(err)
	}
	if err := store.SetProject(other, appstate.ProjectState{File: "app.go", Line: 2, Col: 0, Focus: "main"}); err != nil {
		t.Fatal(err)
	}

	m, _ := previewSample(t)
	m.store = store
	m.width, m.height = 100, 32
	m.focus = PaneFiles
	next, _ := m.Update(keyMsg("ctrl+r"))
	m = next.(Model)
	if !m.recentOpen {
		t.Fatal("ctrl+r should open recent projects")
	}
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "Recent projects") || !strings.Contains(plain, "other") {
		t.Fatalf("recent modal missing the project\n%s", plain)
	}
	next, cmd := m.Update(keyMsg("enter"))
	m = next.(Model)
	if cmd == nil {
		t.Fatal("switching projects should return the load commands")
	}
	if m.root != other {
		t.Fatalf("root %q want %q", m.root, other)
	}
	if m.previewKind != previewFile || m.previewPath != "app.go" {
		t.Fatalf("preview %d %q", m.previewKind, m.previewPath)
	}
	if m.focus != PaneMain {
		t.Fatalf("restored focus %d", m.focus)
	}
}
