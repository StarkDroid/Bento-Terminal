package ui

import (
	"strings"
	"testing"
)

func TestSanitizeLineDropsTerminalControl(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"\x1b[31mred\x1b[0m", "red"},
		{"before\x1b[2J\x1b[3J\x1b[Hafter", "beforeafter"},
		{"\x1b]0;a title\x07kept", "kept"},
		{"\x1b]8;;https://x\x1b\\link\x1b]8;;\x1b\\", "link"},
		{"col\tone", "col    one"},
		{"bell\x07", "bell"},
		{"progress 10%\rprogress 90%\rdone", "done"},
		{"clear\x1b", "clear"},
	}
	for _, c := range cases {
		if got := sanitizeLine(c.in); got != c.want {
			t.Fatalf("sanitizeLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTerminalOutputCannotClearTheScreen(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneBuild
	m = press(t, m, "right")
	m = press(t, m, `printf 'before\033[2J\033[3J\033[Hafter\n'`)
	next, cmd := m.Update(keyMsg("enter"))
	m = drive(t, next.(Model), cmd)
	joined := strings.Join(m.termLines, "\n")
	if strings.ContainsRune(joined, 0x1b) {
		t.Fatalf("terminal kept an escape sequence: %q", joined)
	}
	if !strings.Contains(joined, "beforeafter") {
		t.Fatalf("terminal lost the text around the escapes: %q", joined)
	}
	view := m.View()
	for _, seq := range []string{"\x1b[2J", "\x1b[3J"} {
		if strings.Contains(view, seq) {
			t.Fatalf("view kept an erase sequence %q", seq)
		}
	}
}

func TestBuildOutputCannotClearTheScreen(t *testing.T) {
	m, _ := previewSample(t)
	m.outLines = nil
	m = m.appendOut("running\x1b[2Jdone")
	for _, line := range m.outLines {
		if strings.ContainsRune(line, 0x1b) {
			t.Fatalf("build output kept an escape sequence: %q", line)
		}
	}
	if !strings.Contains(m.output.View(), "runningdone") {
		t.Fatalf("build output lost the text: %q", m.output.View())
	}
}

func TestClearCommandKeepsTheInterface(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneBuild
	m = press(t, m, "right")
	m = press(t, m, "clear")
	next, cmd := m.Update(keyMsg("enter"))
	m = drive(t, next.(Model), cmd)
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "Terminal") || !strings.Contains(plain, "≡ Files") {
		t.Fatalf("clear removed the interface\n%s", plain)
	}
	if strings.ContainsRune(strings.Join(m.termLines, ""), 0x1b) {
		t.Fatalf("clear left escapes in the scrollback: %q", m.termLines)
	}
	for _, line := range m.termLines {
		if line == "" {
			t.Fatalf("clear left a blank line in the scrollback: %q", m.termLines)
		}
	}
	for _, seq := range []string{"\x1b[2J", "\x1b[3J", "\x1b[H"} {
		if strings.Contains(m.View(), seq) {
			t.Fatalf("clear left %q in the view", seq)
		}
	}
}
