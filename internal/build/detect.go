// Package build detects project commands and runs them.
package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Mode selects the build pane list.
type Mode int

const (
	// ModeAuto lists commands detected from project files.
	ModeAuto Mode = iota
	// ModeManual lists the primary run commands for detected ecosystems.
	ModeManual
)

// Command is a runnable command shown in the build pane.
type Command struct {
	Name   string
	Argv   []string
	Source string
}

// Result is the automatic and manual command lists for a project root.
type Result struct {
	Auto    []Command
	Manual  []Command
	Systems []string
}

// Inspect reads project files under root and returns detected commands.
// It does not execute build tools.
func Inspect(root string) Result {
	var result Result
	if text, ok := readFile(filepath.Join(root, "Cargo.toml")); ok {
		result.Systems = append(result.Systems, "Cargo")
		result.Auto = append(result.Auto,
			cmd("cargo test", "Cargo.toml", "cargo", "test"),
			cmd("cargo build", "Cargo.toml", "cargo", "build"),
			cmd("cargo check", "Cargo.toml", "cargo", "check"),
		)
		if cargoBinary(root, text) {
			result.Auto = append(result.Auto, cmd("cargo run", "Cargo.toml", "cargo", "run"))
		}
		result.Manual = append(result.Manual,
			cmd("cargo test", "Cargo.toml", "cargo", "test"),
			cmd("cargo build", "Cargo.toml", "cargo", "build"),
		)
	}
	if _, ok := readFile(filepath.Join(root, "go.mod")); ok {
		result.Systems = append(result.Systems, "Go")
		result.Auto = append(result.Auto,
			cmd("go test ./...", "go.mod", "go", "test", "./..."),
			cmd("go build ./...", "go.mod", "go", "build", "./..."),
			cmd("go vet ./...", "go.mod", "go", "vet", "./..."),
		)
		result.Manual = append(result.Manual,
			cmd("go test ./...", "go.mod", "go", "test", "./..."),
			cmd("go build ./...", "go.mod", "go", "build", "./..."),
		)
	}
	if gradle, ok := gradleWrapper(root); ok {
		result.Systems = append(result.Systems, "Gradle")
		tasks := gradleTasks(root)
		if len(tasks) == 0 {
			tasks = []string{"build", "test", "check"}
		}
		for _, task := range tasks {
			result.Auto = append(result.Auto, cmd(gradle+" "+task, "gradlew", gradle, task))
		}
		for _, task := range []string{"test", "build", "check"} {
			result.Manual = append(result.Manual, cmd(gradle+" "+task, "gradlew", gradle, task))
		}
	} else if fileExists(filepath.Join(root, "build.gradle")) || fileExists(filepath.Join(root, "build.gradle.kts")) {
		result.Systems = append(result.Systems, "Gradle")
	}
	if targets, name, ok := makeTargets(root); ok {
		result.Systems = append(result.Systems, "Make")
		for _, target := range targets {
			result.Auto = append(result.Auto, cmd("make "+target, name, "make", target))
		}
		result.Manual = append(result.Manual, cmd("make", name, "make"))
		if contains(targets, "test") {
			result.Manual = append(result.Manual, cmd("make test", name, "make", "test"))
		}
		if contains(targets, "build") {
			result.Manual = append(result.Manual, cmd("make build", name, "make", "build"))
		}
	}
	if scripts, ok := npmScripts(root); ok {
		result.Systems = append(result.Systems, "npm")
		names := make([]string, 0, len(scripts))
		for name := range scripts {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			result.Auto = append(result.Auto, cmd("npm run "+name, "package.json", "npm", "run", name))
		}
		if _, ok := scripts["start"]; ok {
			result.Manual = append(result.Manual, cmd("npm start", "package.json", "npm", "start"))
		}
		if _, ok := scripts["test"]; ok {
			result.Manual = append(result.Manual, cmd("npm test", "package.json", "npm", "test"))
		}
		if _, ok := scripts["build"]; ok {
			result.Manual = append(result.Manual, cmd("npm run build", "package.json", "npm", "run", "build"))
		}
	}
	if text, ok := readFile(filepath.Join(root, "pyproject.toml")); ok {
		result.Systems = append(result.Systems, "Python")
		result.Auto = append(result.Auto, cmd("python -m pytest", "pyproject.toml", "python", "-m", "pytest"))
		if strings.Contains(text, "[build-system]") || strings.Contains(text, "build-system") {
			result.Auto = append(result.Auto, cmd("python -m build", "pyproject.toml", "python", "-m", "build"))
		}
		result.Manual = append(result.Manual, cmd("python -m pytest", "pyproject.toml", "python", "-m", "pytest"))
	}
	return result
}

func cmd(name, source string, argv ...string) Command {
	return Command{Name: name, Argv: argv, Source: source}
}

func cargoBinary(root, text string) bool {
	if strings.Contains(text, "[[bin]]") {
		return true
	}
	_, err := os.Stat(filepath.Join(root, "src", "main.rs"))
	return err == nil
}

func npmScripts(root string) (map[string]string, bool) {
	text, ok := readFile(filepath.Join(root, "package.json"))
	if !ok {
		return nil, false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal([]byte(text), &pkg); err != nil {
		return map[string]string{}, true
	}
	if pkg.Scripts == nil {
		pkg.Scripts = map[string]string{}
	}
	return pkg.Scripts, true
}

func makeTargets(root string) ([]string, string, bool) {
	for _, name := range []string{"Makefile", "makefile", "GNUmakefile"} {
		text, ok := readFile(filepath.Join(root, name))
		if !ok {
			continue
		}
		return parseMakeTargets(text), name, true
	}
	return nil, "", false
}

var makeTarget = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

func parseMakeTargets(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "#") {
			continue
		}
		colon := strings.Index(line, ":")
		if colon <= 0 {
			continue
		}
		if colon+1 < len(line) && line[colon+1] == '=' {
			continue
		}
		left := line[:colon]
		if strings.ContainsAny(left, "=?+") {
			continue
		}
		for _, name := range strings.Fields(left) {
			if strings.HasPrefix(name, ".") || strings.ContainsAny(name, "$%*") || !makeTarget.MatchString(name) {
				continue
			}
			if seen[name] {
				continue
			}
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

func gradleWrapper(root string) (string, bool) {
	if fileExists(filepath.Join(root, "gradlew")) {
		return "./gradlew", true
	}
	if fileExists(filepath.Join(root, "gradlew.bat")) {
		return `.\gradlew.bat`, true
	}
	return "", false
}

var (
	reGradleRegister = regexp.MustCompile(`tasks\.register\(\s*"([A-Za-z_][A-Za-z0-9_]*)"`)
	reGradleTask     = regexp.MustCompile(`(?m)^\s*task\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
)

func gradleTasks(root string) []string {
	var chunks []string
	for _, name := range []string{"build.gradle", "build.gradle.kts"} {
		if text, ok := readFile(filepath.Join(root, name)); ok {
			chunks = append(chunks, text)
		}
	}
	text := strings.Join(chunks, "\n")
	var out []string
	seen := map[string]bool{}
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}
	for _, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "//") || strings.HasPrefix(trim, "#") || strings.HasPrefix(trim, "*") {
			continue
		}
		if m := reGradleTask.FindStringSubmatch(line); len(m) == 2 {
			add(m[1])
		}
	}
	for _, m := range reGradleRegister.FindAllStringSubmatch(text, -1) {
		add(m[1])
	}
	return out
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

func readFile(path string) (string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(b), true
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
