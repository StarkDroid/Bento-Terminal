package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
)

func TestBuildTabsSwitchWithArrows(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 140
	m.height = 34
	m.focus = PaneBuild

	if m.buildTab != buildTabCommands {
		t.Fatal("build should start on the commands tab")
	}
	m = press(t, m, "right")
	if m.buildTab != buildTabOutput {
		t.Fatal("right should select the output tab")
	}
	if plain := stripANSI(m.View()); !strings.Contains(plain, "No output yet") {
		t.Fatalf("output tab missing its empty state\n%s", plain)
	}
	m = press(t, m, "left")
	if m.buildTab != buildTabCommands {
		t.Fatal("left should select the commands tab")
	}
}

func TestBuildTabClickSwitches(t *testing.T) {
	m, _ := previewSample(t)
	m.width = 140
	m.height = 34
	m.focus = PaneBuild
	lay := computeLayout(m.width, m.height)

	x := -1
	for rel := 0; rel < lay.Build.W; rel++ {
		if tab, ok := m.buildTabAt(rel); ok && tab == buildTabOutput {
			x = rel + 1
			break
		}
	}
	if x < 0 {
		t.Fatal("output tab not laid out")
	}
	next, _ := m.Update(tea.MouseMsg{
		X: x, Y: lay.BuildBodyTop() - 1,
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
	m = next.(Model)
	if m.buildTab != buildTabOutput {
		t.Fatal("clicking Output should select it")
	}
}

func TestBuildLaunchSwitchesToOutput(t *testing.T) {
	m, _ := previewSample(t)
	m.focus = PaneBuild
	m.buildReady = true
	m.mode = build.ModeAuto
	m.autoCmds = []build.Command{{Name: "go test", Source: "go.mod", Argv: []string{"go", "test"}}}
	m.buildTab = buildTabCommands

	next, cmd := m.launch(m.commands()[0])
	m = next.(Model)
	if cmd == nil {
		t.Fatal("launch should return a command")
	}
	if m.buildTab != buildTabOutput {
		t.Fatal("running a build should switch to the output tab")
	}
	if len(m.outLines) == 0 || !strings.Contains(m.outLines[0], "go test") {
		t.Fatalf("output did not start with the command: %#v", m.outLines)
	}
}
