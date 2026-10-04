package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoadRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-b", "main")
	gitCmd(t, dir, "config", "user.email", "bento@example.com")
	gitCmd(t, dir, "config", "user.name", "Bento terminal")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "README.md")
	gitCmd(t, dir, "commit", "-m", "Initial commit")

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\nchanged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "new.txt")

	snap := Load(dir)
	if !snap.Repo || snap.Err != "" {
		t.Fatalf("%+v", snap)
	}
	if snap.Status.Branch != "main" {
		t.Fatalf("branch %q", snap.Status.Branch)
	}
	if snap.Status.Staged != 1 || snap.Status.Unstaged != 1 {
		t.Fatalf("staged=%d unstaged=%d", snap.Status.Staged, snap.Status.Unstaged)
	}
	if len(snap.Commits) != 1 || snap.Commits[0].Subject != "Initial commit" {
		t.Fatalf("commits %+v", snap.Commits)
	}
	if _, ok := snap.Status.Dirty["README.md"]; !ok {
		t.Fatal("README.md should be dirty")
	}
	if _, ok := snap.Status.Dirty["new.txt"]; !ok {
		t.Fatal("new.txt should be dirty")
	}

	gitCmd(t, dir, "stash", "push", "-m", "wip", "--", "README.md")
	stashed := Load(dir)
	if stashed.Status.Stash != 1 {
		t.Fatalf("stash = %d", stashed.Status.Stash)
	}
}

func TestLoadNotRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	snap := Load(t.TempDir())
	if snap.Repo {
		t.Fatal("expected a non-repository")
	}
}

func TestShowRejectsBadHash(t *testing.T) {
	if _, err := Show(t.TempDir(), "not a hash"); err == nil {
		t.Fatal("expected error")
	}
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
