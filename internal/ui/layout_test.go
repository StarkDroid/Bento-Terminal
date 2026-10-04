package ui

import "testing"

func TestComputeLayout(t *testing.T) {
	sizes := []struct{ w, h int }{{100, 32}, {80, 24}, {72, 20}}
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
		if l.History.W != l.Files.W || l.Build.W != sz.w || l.Status.W != sz.w {
			t.Fatalf("%dx%d boxes %+v", sz.w, sz.h, l)
		}
		if l.Main.H+1+l.Build.H+l.Status.H != sz.h {
			t.Fatalf("%dx%d height sum %d", sz.w, sz.h, l.Main.H+1+l.Build.H+l.Status.H)
		}
		listH, outH := buildParts(l.Build.H)
		if listH < 2 || outH < 1 {
			t.Fatalf("%dx%d build parts %d %d", sz.w, sz.h, listH, outH)
		}
	}
	if computeLayout(40, 10).OK {
		t.Fatal("tiny terminal should not lay out")
	}
}
