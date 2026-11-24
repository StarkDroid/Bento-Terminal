package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hinshun/vt10x"

	"github.com/StarkDroid/Bento-Terminal/internal/agent"
)

type fakeSession struct {
	written strings.Builder
	closed  bool
	term    vt10x.Terminal
}

func (f *fakeSession) Write(p []byte) error {
	f.written.Write(p)
	return nil
}

func (f *fakeSession) Resize(int, int) error { return nil }

func (f *fakeSession) Screen() vt10x.Terminal { return f.term }

func (f *fakeSession) Close() { f.closed = true }

func TestAIPanelOpensWithCtrlA(t *testing.T) {
	m, _ := previewSample(t)
	if m.aiOpen {
		t.Fatal("the AI panel should start closed")
	}
	m.aiReady = true
	m.aiFound = map[string]string{}
	next, cmd := m.Update(keyMsg("ctrl+a"))
	m = next.(Model)
	if !m.aiOpen || m.focus != PaneAI {
		t.Fatalf("open %v focus %d", m.aiOpen, m.focus)
	}
	if cmd == nil {
		t.Fatal("opening should detect installed agents")
	}
	plain := stripANSI(m.View())
	for _, want := range []string{"AI", "Claude Code", "OpenCode", "install"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("AI panel missing %q\n%s", want, plain)
		}
	}
	next, _ = m.Update(keyMsg("ctrl+a"))
	m = next.(Model)
	if m.aiOpen || m.focus != PaneMain {
		t.Fatalf("close %v focus %d", m.aiOpen, m.focus)
	}
}

func TestAIPanelFooterShowsToggleShortcut(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.aiOpen = true
	m.aiReady = true
	m.focus = PaneAI
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "ctrl+a close") {
		t.Fatalf("AI footer should show the toggle shortcut\n%s", plain)
	}
}

func TestAIRowsPutInstalledFirst(t *testing.T) {
	m := New(t.TempDir(), "test")
	m.aiFound = map[string]string{"gemini": "/bin/gemini"}
	rows := m.aiRows()
	if len(rows) != len(agent.All()) {
		t.Fatalf("rows %d want %d", len(rows), len(agent.All()))
	}
	if !rows[0].installed || rows[0].provider.ID != "gemini" {
		t.Fatalf("first row %+v", rows[0])
	}
	if rows[1].installed {
		t.Fatalf("second row should need an install: %+v", rows[1])
	}
	ready := stripANSI(m.aiLine(rows[0], 30, false))
	if !strings.Contains(ready, "ready") {
		t.Fatalf("installed row %q", ready)
	}
	missing := stripANSI(m.aiLine(rows[1], 30, false))
	if !strings.Contains(missing, "install") {
		t.Fatalf("missing row %q", missing)
	}
}

func TestAIInstallRunsInThePanel(t *testing.T) {
	m, _ := previewSample(t)
	m.aiOpen = true
	p, ok := agent.Find("claude")
	if !ok {
		t.Fatal("claude should be in the catalogue")
	}
	next, cmd := m.startAIInstall(p)
	m = next.(Model)
	if cmd == nil {
		t.Fatal("install should return a command")
	}
	if !m.running || m.runTarget != targetAI {
		t.Fatalf("running %v target %d", m.running, m.runTarget)
	}
	if m.aiInstall != "claude" || m.aiState != stateRunning {
		t.Fatalf("install %q state %d", m.aiInstall, m.aiState)
	}
	if len(m.aiLines) == 0 || !strings.Contains(m.aiLines[0], "npm install -g @anthropic-ai/claude-code") {
		t.Fatalf("install output %#v", m.aiLines)
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "installing") {
		t.Fatalf("title should say installing\n%s", plain)
	}
}

func TestAISessionHidesFooterAndForwardsKeys(t *testing.T) {
	m, _ := previewSample(t)
	m.aiOpen = true
	m.focus = PaneAI
	fake := &fakeSession{term: vt10x.New(vt10x.WithSize(40, 10))}
	m.aiSession = fake
	m.aiSessionID = "claude"
	m.aiCols, m.aiScreenH = 40, 10

	plain := stripANSI(m.View())
	if strings.Contains(plain, "rescan") {
		t.Fatalf("a running agent should hide the panel footer\n%s", plain)
	}
	if !strings.Contains(plain, "ctrl+\\ detach") {
		t.Fatalf("status bar should carry the detach hint\n%s", plain)
	}

	next, cmd := m.Update(keyMsg("q"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("q should go to the agent, not quit Bento")
	}
	if fake.written.String() != "q" {
		t.Fatalf("agent received %q", fake.written.String())
	}
	next, _ = m.Update(keyMsg("ctrl+a"))
	m = next.(Model)
	if m.aiOpen {
		t.Fatal("ctrl+a should close the AI panel while an agent runs")
	}
	if m.aiSession != fake {
		t.Fatal("closing the panel should keep the agent session alive")
	}
	if fake.written.String() != "q" {
		t.Fatalf("ctrl+a should not reach the agent, got %q", fake.written.String())
	}
	next, _ = m.Update(keyMsg("ctrl+a"))
	m = next.(Model)
	if !m.aiOpen || m.focus != PaneAI {
		t.Fatalf("ctrl+a should reopen the panel: open=%v focus=%d", m.aiOpen, m.focus)
	}

	next, _ = m.Update(keyMsg("ctrl+\\"))
	m = next.(Model)
	if m.aiSession != nil || !fake.closed {
		t.Fatalf("detach should close the session: %v %v", m.aiSession, fake.closed)
	}
	if !strings.Contains(m.aiNote, "detached") {
		t.Fatalf("detach note %q", m.aiNote)
	}
}

func TestAgentScreenRendersCells(t *testing.T) {
	term := vt10x.New(vt10x.WithSize(20, 4))
	if _, err := term.Write([]byte("\x1b[2J\x1b[H\x1b[31mred\x1b[0m plain")); err != nil {
		t.Fatal(err)
	}
	out := stripANSI(renderAgentScreen(term, 20, 4, true))
	if !strings.Contains(out, "red plain") {
		t.Fatalf("screen %q", out)
	}
}

func TestEncodeKey(t *testing.T) {
	cases := []struct {
		msg  string
		want string
	}{
		{"a", "a"},
		{"enter", "\r"},
		{"tab", "\t"},
		{"shift+tab", "\x1b[Z"},
		{"up", "\x1b[A"},
		{"down", "\x1b[B"},
		{"left", "\x1b[D"},
		{"right", "\x1b[C"},
		{"backspace", "\x7f"},
		{"delete", "\x1b[3~"},
		{"esc", "\x1b"},
		{"ctrl+a", "\x01"},
		{"ctrl+c", "\x03"},
	}
	for _, c := range cases {
		if got := string(encodeKey(keyMsg(c.msg))); got != c.want {
			t.Fatalf("encodeKey(%s) = %q, want %q", c.msg, got, c.want)
		}
	}
}

func TestAIViewKeepsWindowSize(t *testing.T) {
	m, _ := previewSample(t)
	m.aiOpen = true
	m.aiReady = true
	m.focus = PaneAI
	for _, size := range []struct{ w, h int }{{72, 22}, {80, 24}, {120, 32}} {
		m.width, m.height = size.w, size.h
		view := m.View()
		if w, h := lipgloss.Width(view), lipgloss.Height(view); w != size.w || h != size.h {
			t.Fatalf("%dx%d rendered %dx%d", size.w, size.h, w, h)
		}
	}
}

func TestAIWheelScrollsTheAgent(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.aiOpen = true
	m.focus = PaneAI
	fake := &fakeSession{term: vt10x.New(vt10x.WithSize(40, 10))}
	m.aiSession = fake
	m.aiCols, m.aiScreenH = 40, 10
	lay := computeLayout(m.width, m.height)
	previewW, _ := lay.MainSplit()

	next, _ := m.Update(tea.MouseMsg{
		X:      previewW + 3,
		Y:      3,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	m = next.(Model)
	if got := fake.written.String(); !strings.Contains(got, "\x1b[5~") {
		t.Fatalf("wheel up wrote %q, want PageUp", got)
	}
	fake.written.Reset()
	next, _ = m.Update(tea.MouseMsg{
		X:      previewW + 3,
		Y:      3,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = next.(Model)
	if got := fake.written.String(); !strings.Contains(got, "\x1b[6~") {
		t.Fatalf("wheel down wrote %q, want PageDown", got)
	}
}
