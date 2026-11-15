package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bento", "state.json")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(dir, "alpha")
	b := filepath.Join(dir, "beta")
	if err := store.Touch(a); err != nil {
		t.Fatal(err)
	}
	if err := store.Touch(b); err != nil {
		t.Fatal(err)
	}
	if err := store.SetProject(a, ProjectState{File: "src/main.go", Line: 12, Col: 3, Scroll: 4, Focus: "main"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	recent := reopened.Recent()
	if len(recent) != 2 || recent[0].Path != b || recent[1].Path != a {
		t.Fatalf("recent order %+v", recent)
	}
	ps := reopened.Project(a)
	if ps.File != "src/main.go" || ps.Line != 12 || ps.Col != 3 || ps.Scroll != 4 || ps.Focus != "main" {
		t.Fatalf("project state %+v", ps)
	}
}

func TestTouchMovesExistingProjectToFront(t *testing.T) {
	store := NewMemory()
	a := filepath.Join(t.TempDir(), "a")
	b := filepath.Join(t.TempDir(), "b")
	_ = store.Touch(a)
	_ = store.Touch(b)
	_ = store.Touch(a)
	recent := store.Recent()
	if len(recent) != 2 || recent[0].Path != a {
		t.Fatalf("recent %+v", recent)
	}
	if recent[0].LastOpened.Before(recent[1].LastOpened) {
		t.Fatal("the newest project should be first")
	}
}

func TestOpenMissingFileIsEmpty(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Recent()) != 0 {
		t.Fatal("missing file should have no recent projects")
	}
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenCorruptFileFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil {
		t.Fatal("corrupt state should be reported")
	}
}

func TestShortPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	if got := ShortPath(home); got != "~" {
		t.Fatalf("home = %q", got)
	}
	if got := ShortPath(filepath.Join(home, "Projects", "app")); got != filepath.Join("~", "Projects", "app") {
		t.Fatalf("nested = %q", got)
	}
	if got := ShortPath("/tmp/app"); got != "/tmp/app" {
		t.Fatalf("outside = %q", got)
	}
}

func TestRecentLimit(t *testing.T) {
	store := NewMemory()
	for i := 0; i < maxRecent+5; i++ {
		_ = store.Touch(filepath.Join(t.TempDir(), time.Now().Format("150405.000000000")))
	}
	if len(store.Recent()) > maxRecent {
		t.Fatalf("recent list grew to %d", len(store.Recent()))
	}
}
