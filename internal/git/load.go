package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

// Snapshot is a point-in-time read of a working tree.
type Snapshot struct {
	Repo    bool
	Status  Status
	Commits []Commit
	Err     string
}

var hashRe = regexp.MustCompile(`^[0-9a-fA-F]{4,64}$`)

// Load runs git in root. Repo is false when root is not a repository.
func Load(root string) Snapshot {
	if _, err := exec.LookPath("git"); err != nil {
		return Snapshot{Err: "git is not installed"}
	}
	if _, err := runGit(root, "rev-parse", "--is-inside-work-tree"); err != nil {
		return Snapshot{}
	}
	snap := Snapshot{Repo: true}
	statusOut, err := runGit(root, "status", "--porcelain=v2", "--branch")
	if err != nil {
		snap.Err = err.Error()
		return snap
	}
	st, err := ParseStatus(statusOut)
	if err != nil {
		snap.Err = err.Error()
		return snap
	}
	if stashOut, err := runGit(root, "stash", "list"); err == nil {
		st.Stash = CountLines(stashOut)
	}
	snap.Status = st

	logOut, err := runGit(root, "log", "-n", "100", "--pretty=format:%h%x09%ct%x09%s")
	if err != nil {
		return snap
	}
	commits, err := ParseLog(logOut)
	if err != nil {
		snap.Err = err.Error()
		return snap
	}
	snap.Commits = commits
	return snap
}

// Show returns a commit summary suitable for the preview pane.
func Show(root, hash string) (string, error) {
	if !hashRe.MatchString(hash) {
		return "", fmt.Errorf("invalid hash")
	}
	return runGit(root, "show", "--stat", "--format=fuller", hash)
}

func runGit(root string, args ...string) (string, error) {
	full := append([]string{"--no-pager", "-C", root, "-c", "core.quotepath=false"}, args...)
	cmd := exec.Command("git", full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := stringsTrim(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	return stdout.String(), nil
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}
