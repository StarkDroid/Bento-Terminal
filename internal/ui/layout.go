package ui

type box struct {
	W int
	H int
}

type layout struct {
	Main     box
	Files    box
	History  box
	Build    box
	Terminal box
	Status   box
	OK       bool
}

func computeLayout(width, height int) layout {
	const (
		statusH = 1
		gap     = 1
	)
	var l layout
	if width < 72 || height < 22 {
		return l
	}
	sideW := width * 34 / 100
	if sideW < 28 {
		sideW = 28
	}
	if sideW > 40 {
		sideW = 40
	}
	mainW := width - gap - sideW
	if mainW < 32 {
		return l
	}
	inner := height - statusH - gap
	buildH := inner * 34 / 100
	if buildH < 8 {
		buildH = 8
	}
	if buildH > 16 {
		buildH = 16
	}
	topH := inner - buildH
	if topH < 9 {
		buildH -= 9 - topH
		topH = 9
	}
	if buildH < 8 {
		return l
	}
	filesH := topH * 55 / 100
	if filesH < 6 {
		filesH = 6
	}
	histH := topH - gap - filesH
	if histH < 5 {
		filesH -= 5 - histH
		histH = 5
	}
	if filesH < 5 || histH < 5 {
		return l
	}
	// The bottom row is two panels, Build on the left and Terminal on the
	// right, split as evenly as the width allows.
	bottomW := width - gap
	buildW := bottomW / 2
	termW := bottomW - buildW
	l.OK = true
	l.Main = box{W: mainW, H: topH}
	l.Files = box{W: sideW, H: filesH}
	l.History = box{W: sideW, H: histH}
	l.Build = box{W: buildW, H: buildH}
	l.Terminal = box{W: termW, H: buildH}
	l.Status = box{W: width, H: statusH}
	return l
}

// MainSplit divides the main box into the file preview and the AI panel. The
// two halves differ by at most one cell because of the gap between them, and
// the AI panel takes the extra cell when the main box is even.
func (l layout) MainSplit() (previewW, aiW int) {
	aiW = l.Main.W / 2
	previewW = l.Main.W - 1 - aiW
	return previewW, aiW
}

// Body tops are the first content row of a pane, after its border and title.
func (l layout) MainBodyTop() int { return 2 }

// FilesBodyTop is the first content row of the files pane.
func (l layout) FilesBodyTop() int { return 2 }

// HistoryBodyTop is the first content row of the history pane.
func (l layout) HistoryBodyTop() int { return l.Files.H + 1 + 2 }

// BuildBodyTop is the first content row of the build pane.
func (l layout) BuildBodyTop() int { return l.Main.H + 1 + 2 }

// TerminalBodyTop is the first content row of the terminal pane.
func (l layout) TerminalBodyTop() int { return l.Main.H + 1 + 2 }

// paneAt maps a screen cell to the pane that owns it. Gaps and the status bar
// return false.
func paneAt(l layout, x, y int, aiOpen bool) (Pane, bool) {
	previewW := l.Main.W
	if aiOpen {
		previewW, _ = l.MainSplit()
	}
	switch {
	case x >= 0 && x < previewW && y >= 0 && y < l.Main.H:
		return PaneMain, true
	case aiOpen && x > previewW && x < l.Main.W && y >= 0 && y < l.Main.H:
		return PaneAI, true
	case x > l.Main.W && x < l.Main.W+1+l.Files.W && y >= 0 && y < l.Files.H:
		return PaneFiles, true
	case x > l.Main.W && x < l.Main.W+1+l.Files.W && y > l.Files.H && y < l.Files.H+1+l.History.H:
		return PaneHistory, true
	case x >= 0 && x < l.Build.W && y > l.Main.H && y < l.Main.H+1+l.Build.H:
		return PaneBuild, true
	case x > l.Build.W && x < l.Build.W+1+l.Terminal.W && y > l.Main.H && y < l.Main.H+1+l.Build.H:
		return PaneTerminal, true
	}
	return 0, false
}
