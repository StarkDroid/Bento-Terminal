package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListSkipsAndSorts(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "internal"))
	mustWrite(t, filepath.Join(dir, "internal", "a.go"), "package internal\n")
	mustWrite(t, filepath.Join(dir, "root.md"), "root\n")
	mustWrite(t, filepath.Join(dir, "z.txt"), "z\n")
	mustMkdir(t, filepath.Join(dir, "node_modules"))
	mustWrite(t, filepath.Join(dir, "node_modules", "skip.js"), "nope\n")
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustWrite(t, filepath.Join(dir, ".git", "config"), "hidden\n")
	if err := os.Symlink(filepath.Join(dir, "root.md"), filepath.Join(dir, "link.md")); err != nil {
		t.Logf("symlinks unavailable: %v", err)
	}

	got, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"internal/a.go", "root.md", "z.txt"}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	}
}

func TestListEmpty(t *testing.T) {
	got, err := List(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %#v", got)
	}
}
