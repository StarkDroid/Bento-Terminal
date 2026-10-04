package ui

// Pane is a focusable region of the layout.
type Pane int

const (
	// PaneMain is the file preview or island home.
	PaneMain Pane = iota
	// PaneFiles is the directory tree.
	PaneFiles
	// PaneHistory is the commit list.
	PaneHistory
	// PaneBuild is the command list and output.
	PaneBuild
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
