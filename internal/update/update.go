// Package update checks the latest Bento release and picks the right update
// command for the way this binary was installed.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// Repo is the GitHub repository that publishes Bento releases.
	Repo = "StarkDroid/Bento-Terminal"
	// APIURL is the latest-release endpoint.
	APIURL = "https://api.github.com/repos/" + Repo + "/releases/latest"

	installSh  = "https://raw.githubusercontent.com/" + Repo + "/main/scripts/install.sh"
	installPs1 = "https://raw.githubusercontent.com/" + Repo + "/main/scripts/install.ps1"
)

// Release is the information the UI needs about the latest version.
type Release struct {
	Version string
	URL     string
	Notes   string
}

// Method describes how the running binary should be updated.
type Method struct {
	Name    string
	Command string
}

// Versioned reports whether v looks like a release version, rather than a
// development build.
func Versioned(v string) bool {
	v = Normalize(v)
	switch v {
	case "", "dev", "development", "(devel)":
		return false
	}
	_, ok := parse(v)
	return ok
}

// Normalize trims spaces and one leading v.
func Normalize(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// Display formats a version for people. Development builds stay readable.
func Display(v string) string {
	n := Normalize(v)
	if !Versioned(n) {
		return "development build"
	}
	return "v" + n
}

// Newer reports whether latest is a later semver than current.
func Newer(current, latest string) bool {
	return Compare(current, latest) < 0
}

// Compare returns -1, 0, or 1 for a versus b. Unparseable versions compare
// equal, so a malformed release tag never produces an update prompt.
func Compare(a, b string) int {
	pa, oka := parse(Normalize(a))
	pb, okb := parse(Normalize(b))
	if !oka || !okb {
		return 0
	}
	for i := range pa.nums {
		switch {
		case pa.nums[i] < pb.nums[i]:
			return -1
		case pa.nums[i] > pb.nums[i]:
			return 1
		}
	}
	switch {
	case pa.pre == "" && pb.pre != "":
		return 1
	case pa.pre != "" && pb.pre == "":
		return -1
	}
	return strings.Compare(pa.pre, pb.pre)
}

type parsed struct {
	nums [3]int
	pre  string
}

func parse(v string) (parsed, bool) {
	var p parsed
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	if i := strings.IndexByte(v, '-'); i >= 0 {
		p.pre = v[i+1:]
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) == 0 || len(parts) > len(p.nums) {
		return p, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return p, false
		}
		p.nums[i] = n
	}
	return p, true
}

// Check fetches the latest release. It returns a nil check for development
// builds; callers should not prompt those.
func Check(ctx context.Context, current string) (Release, bool, error) {
	if !Versioned(current) {
		return Release{}, false, nil
	}
	return checkAt(ctx, &http.Client{Timeout: 8 * time.Second}, APIURL, current)
}

func checkAt(ctx context.Context, client *http.Client, url, current string) (Release, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "bento-terminal")
	resp, err := client.Do(req)
	if err != nil {
		return Release{}, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Release{}, false, fmt.Errorf("github: %s", resp.Status)
	}
	var body struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return Release{}, false, err
	}
	version := Normalize(body.TagName)
	if !Versioned(version) {
		return Release{}, false, fmt.Errorf("github: invalid release tag %q", body.TagName)
	}
	return Release{Version: version, URL: body.HTMLURL, Notes: body.Body}, Newer(current, version), nil
}

// ForExecutable picks the update command from the install path. Package
// managers win over the generic install script because they own the binary.
func ForExecutable(exe string) Method {
	paths := []string{strings.ToLower(filepath.ToSlash(exe))}
	resolved, resolveErr := filepath.EvalSymlinks(exe)
	if resolveErr == nil {
		paths = append(paths, strings.ToLower(filepath.ToSlash(resolved)))
	}
	joined := strings.Join(paths, "|")

	switch {
	case strings.Contains(joined, "/cellar/bento/") || strings.Contains(joined, "homebrew"):
		return Method{Name: "Homebrew", Command: "brew upgrade StarkDroid/tap/bento"}
	case strings.Contains(joined, "/scoop/apps/bento/") || (strings.Contains(joined, "scoop") && strings.Contains(joined, "bento")):
		return Method{Name: "Scoop", Command: "scoop update bento"}
	case strings.Contains(joined, "chocolatey"):
		return Method{Name: "Chocolatey", Command: "choco upgrade bento-terminal -y"}
	case strings.Contains(joined, "node_modules") && strings.Contains(joined, "bento-terminal"):
		return Method{Name: "npm", Command: "npm install -g bento-terminal@latest"}
	}

	if isGoBin(exe) || (resolveErr == nil && isGoBin(resolved)) {
		return Method{
			Name:    "Go",
			Command: "go install github.com/" + Repo + "/cmd/bento@latest",
		}
	}

	if runtime.GOOS == "windows" {
		return Method{
			Name:    "installer",
			Command: `powershell -NoProfile -ExecutionPolicy Bypass -Command "irm ` + installPs1 + ` | iex"`,
		}
	}
	return Method{
		Name:    "installer",
		Command: "curl -fsSL " + installSh + " | sh",
	}
}

// isGoBin recognizes the directories go install writes to. Build info is not
// enough: a go build from a checkout also carries a pseudo-version.
func isGoBin(exe string) bool {
	dir := filepath.Dir(exe)
	var candidates []string
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		candidates = append(candidates, gobin)
	}
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		for _, p := range filepath.SplitList(gopath) {
			candidates = append(candidates, filepath.Join(p, "bin"))
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "go", "bin"))
	}
	for _, candidate := range candidates {
		if samePath(dir, candidate) {
			return true
		}
	}
	return false
}

func samePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
