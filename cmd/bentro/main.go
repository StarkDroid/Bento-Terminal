package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/trishiraj/bentro/internal/ui"
)

var version = "0.1.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "bentro: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: bentro [directory]")
	}
	root := ""
	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			fmt.Printf("bentro %s\n\nUsage: bentro [directory]\n\nOpen Bentro for a project. With no argument, the current directory is the project root.\n", version)
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
	p := tea.NewProgram(ui.New(abs), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
