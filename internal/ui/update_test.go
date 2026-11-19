package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	upd "github.com/StarkDroid/Bento-Terminal/internal/update"
)

func TestUpdateBadgeOnlyWhenAvailable(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	if strings.Contains(stripANSI(m.View()), "update v") {
		t.Fatal("an update badge appeared without a release")
	}

	m.updateRelease = upd.Release{Version: "9.9.10"}
	view := stripANSI(m.View())
	if !strings.Contains(view, "update v9.9.10 · u") {
		t.Fatalf("status badge missing\n%s", view)
	}

	m.updateDone = true
	view = stripANSI(m.View())
	if !strings.Contains(view, "restart Bento") || strings.Contains(view, "· u") {
		t.Fatalf("post-update badge is wrong\n%s", view)
	}
}

func TestAboutShowsCurrentVersionAndUpdate(t *testing.T) {
	m := aboutModel(t)
	m.updateRelease = upd.Release{Version: "9.9.10"}
	view := stripANSI(m.View())
	for _, want := range []string{
		"current v9.9.9",
		"update v9.9.10 available",
		"press u to update via",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("about missing %q\n%s", want, view)
		}
	}
}

func TestUpdatePopupRendersAndLaterCloses(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	m.updateRelease = upd.Release{Version: "9.9.10"}
	m.updateOpen = true

	view := stripANSI(m.View())
	for _, want := range []string{"Update available", "9.9.9", "9.9.10", "update now", "later"} {
		if !strings.Contains(view, want) {
			t.Fatalf("popup missing %q\n%s", want, view)
		}
	}

	next, _ := m.Update(keyMsg("esc"))
	m = next.(Model)
	if m.updateOpen || !m.updateDismissed {
		t.Fatalf("esc should close and dismiss the popup: open=%v dismissed=%v", m.updateOpen, m.updateDismissed)
	}
}

func TestUpdatePopupLaterClick(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	m.updateRelease = upd.Release{Version: "9.9.10"}
	m.updateOpen = true

	for _, hit := range m.updateFrame().buttons {
		if hit.id != laterButtonID {
			continue
		}
		next, _ := m.Update(tea.MouseMsg{
			X: hit.x + 1, Y: hit.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
		})
		m = next.(Model)
		if m.updateOpen || !m.updateDismissed {
			t.Fatalf("clicking later should close the popup: open=%v dismissed=%v", m.updateOpen, m.updateDismissed)
		}
		return
	}
	t.Fatal("later button was not rendered")
}

func TestUpdateCheckAutoOpensPopup(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	m.updateOpen = false
	next, _ := m.Update(updateCheckedMsg{
		release: upd.Release{Version: "9.9.10"},
		newer:   true,
	})
	m = next.(Model)
	if !m.updateOpen {
		t.Fatal("a newer release should open the popup when no overlay is open")
	}

	m = aboutModel(t) // settings is open
	next, _ = m.Update(updateCheckedMsg{
		release: upd.Release{Version: "9.9.11"},
		newer:   true,
	})
	m = next.(Model)
	if m.updateOpen {
		t.Fatal("the popup should not cover an already-open overlay")
	}
	if m.updateRelease.Version != "9.9.11" {
		t.Fatalf("release was not recorded: %+v", m.updateRelease)
	}
}

func TestUpdateActionOnlyWhenAvailable(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	if hasUpdateAction(m) {
		t.Fatal("u should not be listed without an update")
	}
	m.updateRelease = upd.Release{Version: "9.9.10"}
	if !hasUpdateAction(m) {
		t.Fatal("u should be listed when an update is available")
	}
}

func hasUpdateAction(m Model) bool {
	for _, a := range m.globalActions() {
		if a.id == actUpdate {
			return true
		}
	}
	return false
}

func TestStartUpdateRunsInTerminal(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	m.updateRelease = upd.Release{Version: "9.9.10"}
	m.updateMethod = upd.Method{Name: "test", Command: "true"}

	next, cmd := m.startUpdate()
	m = next.(Model)
	if !m.updateRunning {
		t.Fatal("update should be marked running")
	}
	if m.focus != PaneTerminal {
		t.Fatalf("focus = %v, want terminal", m.focus)
	}
	if m.updateOpen {
		t.Fatal("the popup should close when the update starts")
	}
	if cmd == nil {
		t.Fatal("the update should start a command")
	}
	if got := strings.Join(m.termLines, "\n"); !strings.Contains(got, "$ true") {
		t.Fatalf("terminal does not show the update command\n%s", got)
	}
}

func TestUpdateDoneShowsRestart(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	m.updateRelease = upd.Release{Version: "9.9.10"}
	m.runID = 7
	m.runTarget = targetTerminal
	m.updateRunning = true

	next, _ := m.Update(runDoneMsg{id: 7})
	m = next.(Model)
	if m.updateRunning || !m.updateDone {
		t.Fatalf("update state: running=%v done=%v", m.updateRunning, m.updateDone)
	}
	if got := strings.Join(m.termLines, "\n"); !strings.Contains(got, "Restart Bento") {
		t.Fatalf("terminal does not tell the user to restart\n%s", got)
	}
	if view := stripANSI(m.View()); !strings.Contains(view, "restart Bento") {
		t.Fatalf("status badge missing the restart prompt\n%s", view)
	}
}

func TestUpdateShortcutStartsUpdate(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	m.updateRelease = upd.Release{Version: "9.9.10"}
	m.updateMethod = upd.Method{Name: "test", Command: "true"}

	next, _ := m.Update(keyMsg("u"))
	m = next.(Model)
	if !m.updateRunning {
		t.Fatal("u should start the update when one is available")
	}
}

func TestUpdatePopupUpdateClick(t *testing.T) {
	m := aboutModel(t)
	m.settingsOpen = false
	m.updateRelease = upd.Release{Version: "9.9.10"}
	m.updateMethod = upd.Method{Name: "test", Command: "true"}
	m.updateOpen = true

	for _, hit := range m.updateFrame().buttons {
		if hit.id != updateButtonID {
			continue
		}
		next, _ := m.Update(tea.MouseMsg{
			X: hit.x + 1, Y: hit.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
		})
		m = next.(Model)
		if !m.updateRunning || m.updateOpen {
			t.Fatalf("update click: running=%v open=%v", m.updateRunning, m.updateOpen)
		}
		return
	}
	t.Fatal("update button was not rendered")
}

func TestAboutUpdateChipStartsUpdate(t *testing.T) {
	m := aboutModel(t)
	m.updateRelease = upd.Release{Version: "9.9.10"}
	m.updateMethod = upd.Method{Name: "test", Command: "true"}
	x, y := aboutCell(t, m, "u update")
	m, _ = aboutClick(m, x, y)
	if !m.updateRunning || m.settingsOpen {
		t.Fatalf("about update click: running=%v settings=%v", m.updateRunning, m.settingsOpen)
	}
}

func TestUpdateCheckCanBeDisabled(t *testing.T) {
	t.Setenv("BENTO_NO_UPDATE_CHECK", "1")
	if checkUpdateCmd("9.9.9") != nil {
		t.Fatal("BENTO_NO_UPDATE_CHECK should disable the check")
	}
}
