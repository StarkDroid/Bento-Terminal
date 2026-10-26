package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCycleFocus(t *testing.T) {
	order := []Pane{PaneMain, PaneFiles, PaneHistory, PaneBuild}
	p := PaneMain
	for i := 0; i < len(order)*2; i++ {
		want := order[(i+1)%len(order)]
		p = Cycle(p, 1)
		if p != want {
			t.Fatalf("step %d: got %d want %d", i, p, want)
		}
	}
	if Cycle(PaneMain, -1) != PaneBuild {
		t.Fatal("shift-tab from main should land on build")
	}
	if Cycle(PaneFiles, -1) != PaneMain {
		t.Fatal("shift-tab from files should land on main")
	}
}

func TestTabCyclesFocus(t *testing.T) {
	m := New(t.TempDir())
	if m.Focus() != PaneFiles {
		t.Fatalf("initial focus %d", m.Focus())
	}
	want := []Pane{PaneHistory, PaneBuild, PaneMain, PaneFiles}
	for i, pane := range want {
		next, _ := m.Update(keyMsg("tab"))
		m = next.(Model)
		if m.Focus() != pane {
			t.Fatalf("tab %d: got %d want %d", i, m.Focus(), pane)
		}
	}
	next, _ := m.Update(keyMsg("shift+tab"))
	m = next.(Model)
	if m.Focus() != PaneMain {
		t.Fatalf("shift+tab got %d", m.Focus())
	}
}

func TestPaletteBlocksFocusCycle(t *testing.T) {
	m := New(t.TempDir())
	next, _ := m.Update(keyMsg("?"))
	m = next.(Model)
	if !m.paletteOpen {
		t.Fatal("command list should open")
	}
	focus := m.Focus()
	next, _ = m.Update(keyMsg("tab"))
	m = next.(Model)
	if m.Focus() != focus || !m.paletteOpen {
		t.Fatal("tab should not move focus while the command list is open")
	}
	next, _ = m.Update(keyMsg("esc"))
	m = next.(Model)
	if m.paletteOpen {
		t.Fatal("esc should close the command list")
	}
}

func TestQuit(t *testing.T) {
	m := New(t.TempDir())
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("got %T", msg)
	}
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}
