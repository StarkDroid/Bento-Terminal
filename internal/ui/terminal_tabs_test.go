package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTerminalNewTabAndSwitch(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 140
	m.height = 34
	m.focus = PaneTerminal
	m.termInput = "echo first"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m = next.(Model)
	if len(m.termTabs) != 2 || m.termIndex != 1 {
		t.Fatalf("tabs=%d index=%d", len(m.termTabs), m.termIndex)
	}
	if m.termInput != "" || m.focus != PaneTerminal {
		t.Fatalf("new tab input=%q focus=%d", m.termInput, m.focus)
	}
	m.termInput = "echo second"

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlLeft})
	m = next.(Model)
	if m.termIndex != 0 || m.termInput != "echo first" {
		t.Fatalf("previous tab index=%d input=%q", m.termIndex, m.termInput)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlRight})
	m = next.(Model)
	if m.termIndex != 1 || m.termInput != "echo second" {
		t.Fatalf("next tab index=%d input=%q", m.termIndex, m.termInput)
	}
}

func TestTerminalBackgroundOutputStaysInItsTab(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 140
	m.height = 34
	m.focus = PaneTerminal
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m = next.(Model)

	m.termRunIndex = 0
	m = m.appendTerm("from shell one")
	if strings.Join(m.termLines, "\n") != "" {
		t.Fatalf("inactive tab output leaked into active tab: %#v", m.termLines)
	}
	if got := strings.Join(m.termTabs[0].lines, "\n"); !strings.Contains(got, "from shell one") {
		t.Fatalf("tab one output %q", got)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlLeft})
	m = next.(Model)
	if got := strings.Join(m.termLines, "\n"); !strings.Contains(got, "from shell one") {
		t.Fatalf("switching back lost output: %q", got)
	}
}

func TestTerminalTabClickSwitchesAndPlusAdds(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 140
	m.height = 34
	m.focus = PaneTerminal
	lay := computeLayout(m.width, m.height)

	plusRel := -1
	secondRel := -1
	for rel := 0; rel < lay.Terminal.W; rel++ {
		index, plus, ok := m.terminalTabAt(rel)
		if !ok {
			continue
		}
		if plus {
			plusRel = rel
		} else if index == 0 && secondRel < 0 && rel > 0 {
			secondRel = rel
		}
	}
	if plusRel < 0 {
		t.Fatal("plus tab not laid out")
	}

	next, _ := m.Update(tea.MouseMsg{
		X: lay.Build.W + 2 + plusRel, Y: lay.TerminalBodyTop() - 1,
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if len(m.termTabs) != 2 {
		t.Fatalf("plus should add a tab, got %d", len(m.termTabs))
	}

	// Click the first tab.
	next, _ = m.Update(tea.MouseMsg{
		X: lay.Build.W + 2 + secondRel, Y: lay.TerminalBodyTop() - 1,
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if m.termIndex != 0 {
		t.Fatalf("clicking the first tab selected %d", m.termIndex)
	}
}

func TestCtrlTOpensTerminalTabFromAnyPane(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneMain
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m = next.(Model)
	if len(m.termTabs) != 2 || m.focus != PaneTerminal {
		t.Fatalf("tabs=%d focus=%d", len(m.termTabs), m.focus)
	}
}
