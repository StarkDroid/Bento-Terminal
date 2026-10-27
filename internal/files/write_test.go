package files

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndSave(t *testing.T) {
	dir := t.TempDir()
	rel, err := Create(dir, "src/new.go")
	if err != nil {
		t.Fatal(err)
	}
	if rel != "src/new.go" {
		t.Fatalf("rel = %q", rel)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "new.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(dir, "src/new.go"); !errors.Is(err, ErrExists) {
		t.Fatalf("second create = %v", err)
	}
	if _, err := Save(dir, "src/new.go", "package main\n"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "src", "new.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "package main\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestCreateRejectsPathsOutsideProject(t *testing.T) {
	dir := t.TempDir()
	for _, rel := range []string{"", "   ", "..", "../escape.go", "/etc/passwd"} {
		if _, err := Create(dir, rel); err == nil {
			t.Errorf("Create(%q) should fail", rel)
		}
	}
}

func TestTreeRefreshKeepsOpenDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	tree, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	sub := 0
	for i, row := range tree.Rows() {
		if row.Name == "sub" {
			sub = i
		}
	}
	tree.Toggle(sub)
	if !tree.Rows()[sub].Expanded {
		t.Fatal("sub should be open")
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := tree.Refresh(); err != nil {
		t.Fatal(err)
	}
	rows := tree.Rows()
	if len(rows) != 3 || !rows[sub].Expanded || rows[sub+1].Name != "a.txt" || rows[sub+2].Name != "b.txt" {
		t.Fatalf("rows = %+v", rows)
	}
}
