package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/StarkDroid/Bento-Terminal/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// version is stamped by the release build. A plain go build reports dev, and
// make build stamps the latest tag.
var version = "dev"

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
	p := tea.NewProgram(ui.New(abs, version), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}
