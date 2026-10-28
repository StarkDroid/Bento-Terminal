package ui

import (
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

// projectURL is the repository the about overlay links to.
const projectURL = "https://github.com/StarkDroid/Bento-Terminal"

// browserCommand builds the program that hands url to the default browser.
func browserCommand(url string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		return exec.Command("open", url)
	default:
		return exec.Command("xdg-open", url)
	}
}

// openURL hands a URL to the default browser. It is a variable so tests can
// record the request instead of launching one.
var openURL = func(url string) tea.Cmd {
	return func() tea.Msg {
		if cmd := browserCommand(url); cmd != nil {
			_ = cmd.Run()
		}
		return nil
	}
}
