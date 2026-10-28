package ui

import (
	"strings"
	"testing"
)

func TestFindInLines(t *testing.T) {
	lines := []string{"package main", "func main() {", "// main again", "nothing"}
	got := findInLines(lines, "main")
	if len(got) != 3 {
		t.Fatalf("got %#v", got)
	}
	if got[0] != (findMatch{line: 0, start: 8, end: 12}) {
		t.Fatalf("first %#v", got[0])
	}
	if got[1].line != 1 || got[1].start != 5 {
		t.Fatalf("second %#v", got[1])
	}
	if len(findInLines(lines, "MAIN")) != 3 {
		t.Fatal("find should be case-insensitive")
	}
	if len(findInLines(lines, "")) != 0 {
		t.Fatal("empty query should match nothing")
	}
}

func TestFindOpensJumpsAndSteps(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	next, _ := m.Update(keyMsg("ctrl+f"))
	m = next.(Model)
	if !m.findOpen {
		t.Fatal("ctrl+f should open find")
	}
	for _, r := range "println" {
		m = press(t, m, string(r))
	}
	if len(m.findMatches) != 2 {
		t.Fatalf("matches %#v", m.findMatches)
	}
	if m.previewCursor != 3 || m.previewCol != 1 {
		t.Fatalf("jumped to %d:%d", m.previewCursor, m.previewCol)
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "1/2") {
		t.Fatalf("view missing count\n%s", plain)
	}

	m = press(t, m, "enter")
	if m.findIndex != 1 || m.previewCursor != 4 {
		t.Fatalf("next index %d line %d", m.findIndex, m.previewCursor)
	}
	m = press(t, m, "enter")
	if m.findIndex != 0 || m.previewCursor != 3 {
		t.Fatalf("wrap index %d line %d", m.findIndex, m.previewCursor)
	}
	m = press(t, m, "up")
	if m.findIndex != 1 || m.previewCursor != 4 {
		t.Fatalf("prev index %d line %d", m.findIndex, m.previewCursor)
	}

	m = press(t, m, "esc")
	if m.findOpen || len(m.findMatches) != 0 {
		t.Fatal("esc should close find and clear matches")
	}
	if plain := stripANSI(m.View()); strings.Contains(plain, "1/2") {
		t.Fatalf("find row still visible\n%s", plain)
	}
}

func TestFindUsesEditingBuffer(t *testing.T) {
	m, _ := previewSample(t)
	m = press(t, m, "e")
	if !m.previewEditing {
		t.Fatal("expected editor")
	}
	next, _ := m.Update(keyMsg("ctrl+f"))
	m = next.(Model)
	for _, r := range "zzz" {
		m = press(t, m, string(r))
	}
	if !m.previewEditing {
		t.Fatal("find should keep the editor open underneath")
	}
	if len(m.findMatches) != 0 {
		t.Fatalf("unexpected matches %#v", m.findMatches)
	}
	m = press(t, m, "esc")
	if !m.previewEditing {
		t.Fatal("esc should return to the editor")
	}
}

func TestLineHighlightsExpandTabs(t *testing.T) {
	got := lineHighlights([]findMatch{{line: 0, start: 1, end: 5}}, 0, 0, "\tmain")
	if len(got) != 1 {
		t.Fatalf("got %#v", got)
	}
	if got[0].start != 4 || got[0].end != 8 || !got[0].current {
		t.Fatalf("got %#v", got[0])
	}
}

func TestFindIgnoresCommitPreview(t *testing.T) {
	m, _ := previewSample(t)
	m.previewKind = previewCommit
	m.previewNumbered = false
	next, _ := m.Update(keyMsg("ctrl+f"))
	m = next.(Model)
	if m.findOpen {
		t.Fatal("find should not open for a commit")
	}
}
