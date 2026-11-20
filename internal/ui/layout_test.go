package ui

import "testing"

func TestComputeLayout(t *testing.T) {
	sizes := []struct{ w, h int }{{100, 32}, {80, 24}, {72, 22}}
	for _, sz := range sizes {
		l := computeLayout(sz.w, sz.h)
		if !l.OK {
			t.Fatalf("%dx%d not ok", sz.w, sz.h)
		}
		if l.Main.H != l.Files.H+1+l.History.H {
			t.Fatalf("%dx%d vertical right column %d vs %d", sz.w, sz.h, l.Main.H, l.Files.H+1+l.History.H)
		}
		if l.Main.W+1+l.Files.W != sz.w {
			t.Fatalf("%dx%d width %d+%d", sz.w, sz.h, l.Main.W, l.Files.W)
		}
		if l.History.W != l.Files.W || l.Status.W != sz.w {
			t.Fatalf("%dx%d boxes %+v", sz.w, sz.h, l)
		}
		if l.Build.W+1+l.Terminal.W != sz.w {
			t.Fatalf("%dx%d bottom row %d+%d", sz.w, sz.h, l.Build.W, l.Terminal.W)
		}
		if l.Build.H != l.Terminal.H {
			t.Fatalf("%dx%d bottom heights %d %d", sz.w, sz.h, l.Build.H, l.Terminal.H)
		}
		if l.Main.H+1+l.Build.H+l.Status.H != sz.h {
			t.Fatalf("%dx%d height sum %d", sz.w, sz.h, l.Main.H+1+l.Build.H+l.Status.H)
		}
		previewW, aiW := l.MainSplit()
		if previewW+1+aiW != l.Main.W {
			t.Fatalf("%dx%d main split %d+%d", sz.w, sz.h, previewW, aiW)
		}
		if previewW < 1 || aiW < 1 {
			t.Fatalf("%dx%d empty half %d %d", sz.w, sz.h, previewW, aiW)
		}
		if aiW != l.Main.W/2 {
			t.Fatalf("%dx%d ai half %d of %d", sz.w, sz.h, aiW, l.Main.W)
		}
	}
	if computeLayout(40, 10).OK {
		t.Fatal("tiny terminal should not lay out")
	}
	if computeLayout(72, 20).OK {
		t.Fatal("too-short terminal should not lay out")
	}
}

func TestPaneAtSeparatesEveryBox(t *testing.T) {
	l := computeLayout(100, 32)
	if !l.OK {
		t.Fatal("layout not ok")
	}
	previewW, _ := l.MainSplit()
	cases := []struct {
		name   string
		x, y   int
		aiOpen bool
		want   Pane
		ok     bool
	}{
		{"preview", 2, 2, true, PaneMain, true},
		{"ai", previewW + 2, 2, true, PaneAI, true},
		{"ai closed", previewW + 2, 2, false, PaneMain, true},
		{"files", l.Main.W + 2, 2, false, PaneFiles, true},
		{"history", l.Main.W + 2, l.Files.H + 3, false, PaneHistory, true},
		{"build", 2, l.Main.H + 2, false, PaneBuild, true},
		{"terminal", l.Build.W + 2, l.Main.H + 2, false, PaneTerminal, true},
		{"gap", l.Main.W, 2, false, 0, false},
		{"status", 2, l.Main.H + 1 + l.Build.H, false, 0, false},
	}
	for _, c := range cases {
		got, ok := paneAt(l, c.x, c.y, c.aiOpen)
		if ok != c.ok || (ok && got != c.want) {
			t.Fatalf("%s: got %d %v want %d %v", c.name, got, ok, c.want, c.ok)
		}
	}
}
