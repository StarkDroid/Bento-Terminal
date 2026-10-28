package ui

import (
	"runtime"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func aboutModel(t *testing.T) Model {
	t.Helper()
	m, _ := previewSample(t)
	m.width = 120
	m.height = 32
	m.focus = PaneFiles
	m.version = "9.9.9"
	next, _ := m.Update(keyMsg(","))
	m = next.(Model)
	if !m.settingsOpen {
		t.Fatal("comma should open the about overlay")
	}
	return m
}

// aboutCell finds the screen cell of a label in the rendered overlay. The row
// holds box-drawing glyphs, so the byte index is measured as cells.
func aboutCell(t *testing.T, m Model, label string) (int, int) {
	t.Helper()
	view := stripANSI(m.View())
	for y, line := range strings.Split(view, "\n") {
		if at := strings.Index(line, label); at >= 0 {
			return lipgloss.Width(line[:at]), y
		}
	}
	t.Fatalf("the overlay does not show %q\n%s", label, view)
	return 0, 0
}

func aboutClick(m Model, x, y int) (Model, tea.Cmd) {
	next, cmd := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	return next.(Model), cmd
}

func TestAboutButtonsReplaceTheURLLine(t *testing.T) {
	m := aboutModel(t)
	plain := stripANSI(m.View())
	for _, want := range []string{"github", "? commands", "q quit", "esc close"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("about missing %q\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "github.com") {
		t.Fatalf("the about should link without printing the URL\n%s", plain)
	}
}

func TestAboutGitHubButtonOpensTheProject(t *testing.T) {
	m := aboutModel(t)
	original := openURL
	t.Cleanup(func() { openURL = original })

	var opened string
	openURL = func(url string) tea.Cmd {
		opened = url
		return nil
	}
	x, y := aboutCell(t, m, "github")
	_, cmd := aboutClick(m, x, y)
	if cmd != nil {
		t.Fatal("the stubbed browser should return no command")
	}
	if opened != projectURL {
		t.Fatalf("opened %q, want %q", opened, projectURL)
	}
}

func TestAboutCommandsButtonOpensTheKeys(t *testing.T) {
	m := aboutModel(t)
	x, y := aboutCell(t, m, "commands")
	next, _ := aboutClick(m, x, y)
	if next.settingsOpen {
		t.Fatal("commands should close the about")
	}
	if !next.paletteOpen {
		t.Fatal("commands should open the command list")
	}
}

func TestAboutCloseButtonCloses(t *testing.T) {
	m := aboutModel(t)
	x, y := aboutCell(t, m, "close")
	next, _ := aboutClick(m, x, y)
	if next.settingsOpen {
		t.Fatal("close should close the about")
	}
}

func TestAboutQuitButtonQuits(t *testing.T) {
	m := aboutModel(t)
	x, y := aboutCell(t, m, "quit")
	_, cmd := aboutClick(m, x, y)
	if cmd == nil {
		t.Fatal("quit should return a command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("quit should return tea.Quit")
	}
}

func TestAboutListsFeaturesInTwoColumns(t *testing.T) {
	m := aboutModel(t)
	view := stripANSI(m.View())
	for _, want := range []string{
		"Everything in reach, nothing in the way.",
		"Browse the tree",
		"Review changes",
		"Run detected builds",
		"Use your shell",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("about missing %q\n%s", want, view)
		}
	}
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Browse the tree") && strings.Contains(line, "Review changes") {
			return
		}
	}
	t.Fatalf("features should sit in two columns\n%s", view)
}

func TestAboutIgnoresClicksOutsideButtons(t *testing.T) {
	m := aboutModel(t)
	next, cmd := aboutClick(m, 0, 0)
	if cmd != nil {
		t.Fatal("a click on the background should do nothing")
	}
	if !next.settingsOpen {
		t.Fatal("a click on the background should keep the about open")
	}
}

func TestBrowserCommandCarriesTheURL(t *testing.T) {
	cmd := browserCommand(projectURL)
	if len(cmd.Args) < 2 {
		t.Fatalf("browser command has no URL: %#v", cmd.Args)
	}
	if got := cmd.Args[len(cmd.Args)-1]; got != projectURL {
		t.Fatalf("browser command ends with %q, want %q", got, projectURL)
	}
	if runtime.GOOS == "windows" {
		if cmd.Args[0] != "rundll32" || cmd.Args[1] != "url.dll,FileProtocolHandler" {
			t.Fatalf("windows browser command: %#v", cmd.Args)
		}
	}
}
