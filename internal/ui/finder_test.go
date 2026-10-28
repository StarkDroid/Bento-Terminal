package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRankFilesFuzzy(t *testing.T) {
	files := []string{"README.md", "internal/box/box.go", "main.go", "main_test.go"}
	got := rankFiles(files, "main", 10)
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	if got[0].path != "main.go" || got[1].path != "main_test.go" {
		t.Fatalf("got %#v", got)
	}
	if strings.Join(runeSlice(got[0].pos), ",") != "0,1,2,3" {
		t.Fatalf("positions %#v", got[0].pos)
	}

	all := rankFiles(files, "", 10)
	if len(all) != len(files) {
		t.Fatalf("empty query got %#v", all)
	}
	if rankFiles(files, "zzz", 10) != nil {
		t.Fatal("expected no matches")
	}
}

func runeSlice(v []int) []string {
	out := make([]string, len(v))
	for i, n := range v {
		out[i] = string(rune('0' + n))
	}
	return out
}

func TestFinderOpensFiltersAndOpensFile(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneFiles
	m.finderReady = true
	m.finderFiles = []string{"README.md", "internal/box/box.go", "main.go", "main_test.go"}
	next, cmd := m.Update(keyMsg("ctrl+p"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("no load command expected")
	}
	if !m.finderOpen {
		t.Fatal("ctrl+p should open the finder")
	}
	for _, r := range "main" {
		m = press(t, m, string(r))
	}
	if m.finderInput != "main" || len(m.finderMatches) != 2 {
		t.Fatalf("input %q matches %#v", m.finderInput, m.finderMatches)
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "Find file") || !strings.Contains(plain, "main_test.go") {
		t.Fatalf("finder view\n%s", plain)
	}
	m = press(t, m, "down")
	if m.finderCursor != 1 {
		t.Fatalf("cursor %d", m.finderCursor)
	}
	m = press(t, m, "up")
	next, _ = m.Update(keyMsg("enter"))
	m = next.(Model)
	if m.finderOpen {
		t.Fatal("enter should close the finder")
	}
	if m.previewPath != "main.go" || m.previewKind != previewFile {
		t.Fatalf("preview %q kind %d", m.previewPath, m.previewKind)
	}
	if m.focus != PaneMain {
		t.Fatalf("focus %d", m.focus)
	}
}

func TestFinderLoadsFilesOnFirstOpen(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := readyEmpty(dir)
	next, cmd := m.Update(keyMsg("ctrl+p"))
	m = next.(Model)
	if !m.finderOpen || !m.finderLoading || cmd == nil {
		t.Fatalf("open=%v loading=%v cmd=%v", m.finderOpen, m.finderLoading, cmd)
	}
	m = drive(t, m, cmd)
	if !m.finderReady || m.finderLoading || len(m.finderFiles) != 1 || m.finderFiles[0] != "a.txt" {
		t.Fatalf("ready=%v loading=%v files=%#v", m.finderReady, m.finderLoading, m.finderFiles)
	}
}

func TestFinderEscAndEditingGuard(t *testing.T) {
	m, _ := previewSample(t)
	m.finderReady = true
	m.finderFiles = []string{"main.go"}
	next, _ := m.Update(keyMsg("ctrl+p"))
	m = next.(Model)
	m = press(t, m, "esc")
	if m.finderOpen {
		t.Fatal("esc should close the finder")
	}

	m, _ = previewSample(t)
	m = press(t, m, "e")
	next, _ = m.Update(keyMsg("ctrl+p"))
	m = next.(Model)
	if m.finderOpen {
		t.Fatal("file search should not open while editing")
	}
}

func TestFinderShortcutsInPaletteAndFooter(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneFiles
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "ctrl+p find") {
		t.Fatalf("files footer\n%s", plain)
	}
	next, _ := m.Update(keyMsg("?"))
	m = next.(Model)
	plain = stripANSI(m.View())
	for _, want := range []string{"find a file in the project", "find in this file"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("palette missing %q\n%s", want, plain)
		}
	}
}

func TestFinderAndFindFitMinimumWindow(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 72
	m.height = 22
	m.finderReady = true
	m.finderFiles = []string{"main.go", "README.md"}
	next, _ := m.Update(keyMsg("ctrl+p"))
	m = next.(Model)
	if w, h := lipgloss.Width(m.View()), lipgloss.Height(m.View()); w != 72 || h != 22 {
		t.Fatalf("finder view %dx%d", w, h)
	}
	m = press(t, m, "esc")

	m.focus = PaneFiles
	next, _ = m.Update(keyMsg("ctrl+f"))
	m = next.(Model)
	if !m.findOpen || m.focus != PaneMain {
		t.Fatalf("find open=%v focus=%d", m.findOpen, m.focus)
	}
	if w, h := lipgloss.Width(m.View()), lipgloss.Height(m.View()); w != 72 || h != 22 {
		t.Fatalf("find view %dx%d", w, h)
	}
}
