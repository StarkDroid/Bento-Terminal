package files

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestRenameAndMoveFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")

	rel, err := Rename(dir, "main.go", "cmd/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if rel != "cmd/main.go" {
		t.Fatalf("rel = %q", rel)
	}
	if got := readFile(t, dir, "cmd/main.go"); got != "package main\n" {
		t.Fatalf("body = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source still exists: %v", err)
	}

	rel, err = Move(dir, "cmd/main.go", "cmd/app.go")
	if err != nil {
		t.Fatal(err)
	}
	if rel != "cmd/app.go" {
		t.Fatalf("move rel = %q", rel)
	}
}

func TestCopyFileAndDirectory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/a.go", "package pkg\n")
	writeFile(t, dir, "pkg/sub/b.txt", "b\n")

	if _, err := Copy(dir, "pkg/a.go", "pkg/a_test.go"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, dir, "pkg/a_test.go"); got != "package pkg\n" {
		t.Fatalf("copy body = %q", got)
	}
	if got := readFile(t, dir, "pkg/a.go"); got != "package pkg\n" {
		t.Fatalf("source changed: %q", got)
	}

	if _, err := Copy(dir, "pkg", "pkg-copy"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, dir, "pkg-copy/sub/b.txt"); got != "b\n" {
		t.Fatalf("directory copy body = %q", got)
	}
}

func TestOpsRefuseOverwriteAndOutsidePaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "b.txt", "b")

	if _, err := Rename(dir, "a.txt", "b.txt"); !errors.Is(err, ErrExists) {
		t.Fatalf("rename onto existing = %v", err)
	}
	if _, err := Copy(dir, "a.txt", "b.txt"); !errors.Is(err, ErrExists) {
		t.Fatalf("copy onto existing = %v", err)
	}
	if _, err := Move(dir, "a.txt", "../escape.txt"); err == nil {
		t.Fatal("move outside the project should fail")
	}
	if _, err := Remove(dir, "../escape.txt"); err == nil {
		t.Fatal("remove outside the project should fail")
	}
	if _, err := Rename(dir, "missing.txt", "new.txt"); err == nil {
		t.Fatal("rename of a missing file should fail")
	}
}

func TestOpsRefuseDirectoryIntoItself(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/a.go", "package pkg\n")

	if _, err := Copy(dir, "pkg", "pkg/inner"); err == nil {
		t.Fatal("copying a directory into itself should fail")
	}
	if _, err := Move(dir, "pkg", "pkg/inner"); err == nil {
		t.Fatal("moving a directory into itself should fail")
	}
}

func TestRemoveFileAndDirectory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "pkg/b.txt", "b")

	if _, err := Remove(dir, "a.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file still exists: %v", err)
	}
	if _, err := Remove(dir, "pkg"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pkg")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("directory still exists: %v", err)
	}
}
