package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTreeOrderExpandAndSkip(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "b.txt"), "b")
	mustWrite(t, filepath.Join(dir, "a.txt"), "a")
	mustMkdir(t, filepath.Join(dir, "sub"))
	mustWrite(t, filepath.Join(dir, "sub", "child.go"), "package child\n")
	mustMkdir(t, filepath.Join(dir, "node_modules"))
	mustWrite(t, filepath.Join(dir, "node_modules", "skip.js"), "nope")
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustWrite(t, filepath.Join(dir, ".git", "config"), "hidden")
	if err := os.Symlink(filepath.Join(dir, "a.txt"), filepath.Join(dir, "link.txt")); err != nil {
		t.Fatal(err)
	}

	tree, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	rows := tree.Rows()
	if len(rows) != 3 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].Name != "sub" || !rows[0].IsDir {
		t.Fatalf("first row = %+v", rows[0])
	}
	if rows[1].Name != "a.txt" || rows[2].Name != "b.txt" {
		t.Fatalf("files = %+v", rows)
	}

	tree.Expand(0)
	rows = tree.Rows()
	if len(rows) != 4 || rows[1].Name != "child.go" || rows[1].Depth != 1 {
		t.Fatalf("expanded = %+v", rows)
	}
	tree.Collapse(0)
	if len(tree.Rows()) != 3 {
		t.Fatalf("collapsed = %+v", tree.Rows())
	}
	tree.Toggle(1)
	if len(tree.Rows()) != 3 {
		t.Fatal("toggle on a file should not change the tree")
	}
}

func TestLoadEmpty(t *testing.T) {
	tree, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Rows()) != 0 {
		t.Fatalf("%+v", tree.Rows())
	}
}

func TestLoadNotDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	mustWrite(t, path, "x")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
