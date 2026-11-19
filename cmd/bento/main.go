package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	appstate "github.com/StarkDroid/Bento-Terminal/internal/state"
	"github.com/StarkDroid/Bento-Terminal/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// version is stamped by the release build. A plain go build reports dev, and
// make build stamps the latest tag. go install of a tagged release carries the
// clean module version in build info, so use that when no ldflags version was
// supplied.
var version = "dev"

func init() {
	version = resolveVersion(version)
}

func resolveVersion(stamped string) string {
	stamped = strings.TrimSpace(stamped)
	if stamped != "" && stamped != "dev" {
		return stamped
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return stamped
	}
	if v := versionFromBuildInfo(info.Main.Version); v != "" {
		return v
	}
	return stamped
}

func versionFromBuildInfo(v string) string {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" || v == "(devel)" {
		return ""
	}
	// Only accept a clean release version. A local build's pseudo-version and
	// an untagged checkout must keep the development build marker.
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return ""
	}
	for _, part := range parts {
		if part == "" {
			return ""
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return ""
			}
		}
	}
	return v
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "bento: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: bento [directory]")
	}
	root := ""
	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			fmt.Printf("bento %s\n\nUsage: bento [directory]\n\nOpen Bento terminal for a project. With no argument, the current directory is the project root.\n", version)
			return nil
		case "-v", "--version":
			fmt.Println(version)
			return nil
		default:
			root = args[0]
		}
	}
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = wd
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", abs)
	}
	store, err := appstate.OpenDefault()
	if err != nil {
		store = appstate.NewMemory()
		fmt.Fprintf(os.Stderr, "bento: recent projects are unavailable: %s\n", err)
	}
	p := tea.NewProgram(ui.NewWithState(abs, version, store), tea.WithAltScreen(), tea.WithMouseCellMotion())
	final, err := p.Run()
	if model, ok := final.(ui.Model); ok {
		model.SaveState()
	}
	return err
}
