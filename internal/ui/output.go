package ui

import "strings"

// sanitizeLine makes captured process output safe to draw inside Bento. A
// command writes to a pipe, but it still believes it owns a terminal, so it can
// emit colour, erase, and cursor sequences; letting those reach the screen
// would repaint or clear the app around the pane. The escapes are dropped
// instead, which costs the colours but keeps Bento's own display intact.
//
// A carriage return is a progress-bar redraw, so only the last segment is kept,
// tabs become four spaces, and the remaining control characters are removed.
func sanitizeLine(line string) string {
	if i := strings.LastIndexByte(line, '\r'); i >= 0 {
		line = line[i+1:]
	}
	var b strings.Builder
	b.Grow(len(line))
	for i := 0; i < len(line); {
		if line[i] == 0x1b {
			i += escapeLen(line[i:])
			continue
		}
		c := line[i]
		switch {
		case c == '\t':
			b.WriteString("    ")
		case c < 0x20 || c == 0x7f:
			// Drop control characters.
		default:
			b.WriteByte(c)
		}
		i++
	}
	return b.String()
}

// escapeLen measures the escape sequence at the start of s: a CSI run up to
// its final byte, a string sequence up to BEL or ST, or a two-byte escape.
func escapeLen(s string) int {
	if len(s) < 2 {
		return len(s)
	}
	switch s[1] {
	case '[':
		for i := 2; i < len(s); i++ {
			if s[i] >= 0x40 && s[i] <= 0x7e {
				return i + 1
			}
		}
		return len(s)
	case ']', 'P', 'X', '^', '_':
		for i := 2; i < len(s); i++ {
			if s[i] == 0x07 {
				return i + 1
			}
			if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '\\' {
				return i + 2
			}
		}
		return len(s)
	default:
		return 2
	}
}
