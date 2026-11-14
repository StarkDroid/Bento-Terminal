package ui

// Pane is a focusable region of the layout.
type Pane int

const (
	// PaneMain is the file preview or island home.
	PaneMain Pane = iota
	// PaneAI is the coding-agent panel beside the preview.
	PaneAI
	// PaneFiles is the directory tree.
	PaneFiles
	// PaneHistory is the commit list.
	PaneHistory
	// PaneBuild is the command list and output.
	PaneBuild
	// PaneTerminal is the free-form shell.
	PaneTerminal
	paneCount
)

// Cycle moves focus by delta panes and wraps at both ends.
func Cycle(p Pane, delta int) Pane {
	n := int(paneCount)
	d := delta % n
	if d < 0 {
		d += n
	}
	return Pane((int(p) + d) % n)
}

// nextPane is Cycle that also skips the AI panel while it is closed.
func (m Model) nextPane(delta int) Pane {
	for i := 0; i < int(paneCount); i++ {
		p := Cycle(m.focus, delta*(i+1))
		if p == PaneAI && !m.aiOpen {
			continue
		}
		return p
	}
	return m.focus
}
