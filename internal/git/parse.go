package git

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Status is a branch and worktree summary from git status --porcelain=v2 --branch.
// Untracked paths count as unstaged.
type Status struct {
	Branch   string
	Upstream string
	Ahead    int
	Behind   int
	Staged   int
	Unstaged int
	Stash    int
	Detached bool
	Dirty    map[string]struct{}
	// Files lists each changed path in the order git reports it.
	Files []FileStatus
}

// FileStatus is one changed path. Index is the staged code and Worktree is the
// unstaged code, using the porcelain-v2 letters ('.' means unchanged).
type FileStatus struct {
	Path      string
	OrigPath  string
	Index     byte
	Worktree  byte
	Untracked bool
	Unmerged  bool
}

// Staged reports whether the index differs from HEAD.
func (f FileStatus) Staged() bool {
	return !f.Untracked && f.Index != 0 && f.Index != '.' && f.Index != ' '
}

// Unstaged reports whether the working tree differs from the index.
func (f FileStatus) Unstaged() bool {
	return !f.Untracked && f.Worktree != 0 && f.Worktree != '.' && f.Worktree != ' '
}

// Code renders the two status columns, for example "M ", " M", or "??".
func (f FileStatus) Code() string {
	if f.Untracked {
		return "??"
	}
	x, y := f.Index, f.Worktree
	if x == 0 || x == '.' {
		x = ' '
	}
	if y == 0 || y == '.' {
		y = ' '
	}
	return string([]byte{x, y})
}

// Commit is one git log row.
type Commit struct {
	Hash    string
	Subject string
	When    time.Time
}

// ParseStatus parses porcelain v2 output that includes the branch header.
func ParseStatus(out string) (Status, error) {
	st := Status{Dirty: map[string]struct{}{}}
	if strings.TrimSpace(out) == "" {
		return st, fmt.Errorf("empty status")
	}
	sawBranch := false
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			sawBranch = true
			applyHeader(&st, strings.TrimPrefix(line, "# "))
			continue
		}
		switch line[0] {
		case '?':
			st.Unstaged++
			if p := cleanPath(line[2:]); p != "" {
				st.Dirty[p] = struct{}{}
				st.Files = append(st.Files, FileStatus{
					Path:      p,
					Index:     '?',
					Worktree:  '?',
					Untracked: true,
				})
			}
		case '!':
			continue
		case '1':
			xy, path := splitOrdinary(line)
			applyXY(&st, xy, path)
			st.Files = append(st.Files, fileStatus(xy, path, ""))
		case '2':
			xy, path, orig := splitRename(line)
			applyXY(&st, xy, path)
			st.Files = append(st.Files, fileStatus(xy, path, orig))
		case 'u':
			xy, path := splitUnmerged(line)
			applyXY(&st, xy, path)
			fs := fileStatus(xy, path, "")
			fs.Unmerged = true
			st.Files = append(st.Files, fs)
		}
	}
	if !sawBranch {
		return st, fmt.Errorf("missing branch header")
	}
	return st, nil
}

func applyHeader(st *Status, rest string) {
	switch {
	case strings.HasPrefix(rest, "branch.head "):
		head := strings.TrimPrefix(rest, "branch.head ")
		if head == "(detached)" {
			st.Detached = true
			st.Branch = "HEAD"
			return
		}
		st.Branch = head
	case strings.HasPrefix(rest, "branch.upstream "):
		st.Upstream = strings.TrimPrefix(rest, "branch.upstream ")
	case strings.HasPrefix(rest, "branch.ab "):
		st.Ahead, st.Behind = parseAheadBehind(strings.TrimPrefix(rest, "branch.ab "))
	}
}

func parseAheadBehind(v string) (int, int) {
	var ahead, behind int
	_, _ = fmt.Sscanf(v, "+%d -%d", &ahead, &behind)
	return ahead, behind
}

func applyXY(st *Status, xy, path string) {
	if len(xy) < 2 {
		return
	}
	if xy[0] != '.' && xy[0] != ' ' {
		st.Staged++
	}
	if xy[1] != '.' && xy[1] != ' ' {
		st.Unstaged++
	}
	if path != "" {
		st.Dirty[path] = struct{}{}
	}
}

func splitOrdinary(line string) (string, string) {
	parts := strings.SplitN(line, " ", 9)
	if len(parts) < 9 {
		return "", ""
	}
	return parts[1], cleanPath(parts[8])
}

func splitRename(line string) (string, string, string) {
	parts := strings.SplitN(line, " ", 10)
	if len(parts) < 10 {
		return "", "", ""
	}
	path := parts[9]
	orig := ""
	if tab := strings.IndexByte(path, '\t'); tab >= 0 {
		orig = cleanPath(path[tab+1:])
		path = path[:tab]
	}
	return parts[1], cleanPath(path), orig
}

func splitUnmerged(line string) (string, string) {
	parts := strings.SplitN(line, " ", 11)
	if len(parts) < 11 {
		return "", ""
	}
	return parts[1], cleanPath(parts[10])
}

func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	if strings.HasPrefix(p, "\"") {
		if u, err := strconv.Unquote(p); err == nil {
			p = u
		}
	}
	return filepath.ToSlash(p)
}

func fileStatus(xy, path, orig string) FileStatus {
	fs := FileStatus{Path: path, OrigPath: orig}
	if len(xy) > 0 {
		fs.Index = xy[0]
	}
	if len(xy) > 1 {
		fs.Worktree = xy[1]
	}
	return fs
}

// ParseLog parses lines of hash, unix seconds, and subject separated by tabs.
func ParseLog(out string) ([]Commit, error) {
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var commits []Commit
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("bad log line %q", line)
		}
		sec, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad log time %q", parts[1])
		}
		commits = append(commits, Commit{
			Hash:    parts[0],
			Subject: parts[2],
			When:    time.Unix(sec, 0),
		})
	}
	return commits, nil
}

// FormatWhen renders a short relative time, or a month-day date after two weeks.
func FormatWhen(t, now time.Time) string {
	if t.After(now) {
		return t.Format("Jan 2")
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d / time.Minute)
		if mins < 1 {
			mins = 1
		}
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d/time.Hour))
	case d < 14*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d/(24*time.Hour)))
	default:
		return t.Format("Jan 2")
	}
}

// CountLines counts non-empty lines, used for git stash list.
func CountLines(out string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}
