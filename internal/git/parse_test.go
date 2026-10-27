package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestParseStatusFixture(t *testing.T) {
	st, err := ParseStatus(readFixture(t, "status_repo.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if st.Branch != "main" || st.Upstream != "origin/main" {
		t.Fatalf("branch: %+v", st)
	}
	if st.Ahead != 3 || st.Behind != 1 {
		t.Fatalf("ahead/behind = %d/%d", st.Ahead, st.Behind)
	}
	if st.Staged != 5 {
		t.Fatalf("staged = %d", st.Staged)
	}
	if st.Unstaged != 6 {
		t.Fatalf("unstaged = %d", st.Unstaged)
	}
	for _, path := range []string{
		"README.md",
		"main.go",
		"internal/app.go",
		"new.txt",
		"renamed.go",
		"scratch.txt",
		"my file.txt",
		"quote\"name.txt",
		"conflict.txt",
	} {
		if _, ok := st.Dirty[path]; !ok {
			t.Errorf("missing dirty path %q", path)
		}
	}
	if _, ok := st.Dirty["old.go"]; ok {
		t.Fatal("rename source should not be the dirty path")
	}
	if len(st.Files) != 9 {
		t.Fatalf("files = %d: %+v", len(st.Files), st.Files)
	}
	byPath := map[string]FileStatus{}
	for _, f := range st.Files {
		byPath[f.Path] = f
	}
	if got := byPath["README.md"]; got.Code() != "M " || !got.Staged() || got.Unstaged() {
		t.Fatalf("README.md = %+v code %q", got, got.Code())
	}
	if got := byPath["main.go"]; got.Code() != " M" || got.Staged() || !got.Unstaged() {
		t.Fatalf("main.go = %+v code %q", got, got.Code())
	}
	if got := byPath["internal/app.go"]; got.Code() != "MM" || !got.Staged() || !got.Unstaged() {
		t.Fatalf("app.go = %+v", got)
	}
	if got := byPath["new.txt"]; got.Code() != "A " {
		t.Fatalf("new.txt = %+v", got)
	}
	if got := byPath["renamed.go"]; got.Code() != "R " || got.OrigPath != "old.go" {
		t.Fatalf("renamed.go = %+v", got)
	}
	if got := byPath["scratch.txt"]; !got.Untracked || got.Code() != "??" {
		t.Fatalf("scratch.txt = %+v", got)
	}
	if got := byPath["conflict.txt"]; !got.Unmerged {
		t.Fatalf("conflict.txt = %+v", got)
	}
}

func TestParseStatusDetached(t *testing.T) {
	st, err := ParseStatus(readFixture(t, "status_detached.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !st.Detached || st.Branch != "HEAD" {
		t.Fatalf("%+v", st)
	}
	if st.Upstream != "" || st.Ahead != 0 || st.Behind != 0 {
		t.Fatalf("upstream data = %+v", st)
	}
}

func TestParseStatusRejectsGarbage(t *testing.T) {
	if _, err := ParseStatus(""); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ParseStatus("?? README.md\n"); err == nil {
		t.Fatal("expected missing header error")
	}
}

func TestParseLogFixture(t *testing.T) {
	commits, err := ParseLog(readFixture(t, "log.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 {
		t.Fatalf("%d commits", len(commits))
	}
	if commits[0].Hash != "abc1234" || commits[0].Subject != "Initial commit" {
		t.Fatalf("%+v", commits[0])
	}
	if !commits[0].When.Equal(time.Unix(1710000000, 0)) {
		t.Fatalf("when = %s", commits[0].When)
	}
	if commits[1].Subject != "Add the tree view" {
		t.Fatal(commits[1].Subject)
	}
}

func TestParseLogEmptyAndInvalid(t *testing.T) {
	commits, err := ParseLog("\n")
	if err != nil || commits != nil {
		t.Fatalf("commits=%v err=%v", commits, err)
	}
	if _, err := ParseLog("not-a-log\n"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatWhen(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		at   time.Time
		want string
	}{
		{now.Add(-20 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m ago"},
		{now.Add(-3 * time.Hour), "3h ago"},
		{now.Add(-48 * time.Hour), "2d ago"},
		{now.Add(-40 * 24 * time.Hour), "Aug 25"},
	}
	for _, tc := range cases {
		if got := FormatWhen(tc.at, now); got != tc.want {
			t.Errorf("FormatWhen(%s) = %q, want %q", tc.at, got, tc.want)
		}
	}
}

func TestCountLines(t *testing.T) {
	if CountLines("stash@{0}: On main: wip\n\nstash@{1}: On main: other\n") != 2 {
		t.Fatal("stash count")
	}
}
