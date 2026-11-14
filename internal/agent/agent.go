// Package agent knows the coding-agent CLIs Bento can install and run.
package agent

import (
	"os/exec"
	"runtime"
)

// Provider is one coding agent with a command-line interface.
type Provider struct {
	ID      string
	Name    string
	Command string
	// Install is the shell line that installs the agent. Windows uses
	// InstallWindows when it is set.
	Install        string
	InstallWindows string
}

// InstallLine returns the install command for a platform.
func (p Provider) InstallLine(goos string) string {
	if goos == "windows" && p.InstallWindows != "" {
		return p.InstallWindows
	}
	return p.Install
}

// providers is the catalogue shown in the AI panel. Detection is a PATH lookup
// of Command, so a provider installed by any package manager is found.
var providers = []Provider{
	{
		ID:      "claude",
		Name:    "Claude Code",
		Command: "claude",
		Install: "npm install -g @anthropic-ai/claude-code",
	},
	{
		ID:      "opencode",
		Name:    "OpenCode",
		Command: "opencode",
		Install: "npm i -g opencode-ai",
	},
	{
		ID:      "gemini",
		Name:    "Gemini CLI",
		Command: "gemini",
		Install: "npm install -g @google/gemini-cli",
	},
	{
		ID:      "codex",
		Name:    "Codex CLI",
		Command: "codex",
		Install: "npm install -g @openai/codex",
	},
	{
		ID:      "copilot",
		Name:    "GitHub Copilot CLI",
		Command: "copilot",
		Install: "npm install -g @github/copilot",
	},
	{
		ID:      "crush",
		Name:    "Crush",
		Command: "crush",
		Install: "go install github.com/charmbracelet/crush@latest",
	},
	{
		ID:             "aider",
		Name:           "Aider",
		Command:        "aider",
		Install:        "python3 -m pip install -U aider-chat",
		InstallWindows: "py -m pip install -U aider-chat",
	},
	{
		ID:             "cursor",
		Name:           "Cursor CLI",
		Command:        "cursor-agent",
		Install:        "curl -fsSL https://cursor.com/install | bash",
		InstallWindows: "powershell -NoProfile -Command \"irm https://cursor.com/install | iex\"",
	},
}

// All returns the provider catalogue in display order.
func All() []Provider {
	out := make([]Provider, len(providers))
	copy(out, providers)
	return out
}

// Find returns the provider with id.
func Find(id string) (Provider, bool) {
	for _, p := range providers {
		if p.ID == id {
			return p, true
		}
	}
	return Provider{}, false
}

// Installed reports whether the provider's command is on PATH.
func Installed(p Provider) (string, bool) {
	path, err := exec.LookPath(p.Command)
	if err != nil {
		return "", false
	}
	return path, true
}

// Detect returns the installed providers keyed by ID, with the command path.
func Detect() map[string]string {
	found := make(map[string]string, len(providers))
	for _, p := range providers {
		if path, ok := Installed(p); ok {
			found[p.ID] = path
		}
	}
	return found
}

// InstallCommand returns the shell line for the current platform.
func (p Provider) InstallCommand() string {
	return p.InstallLine(runtime.GOOS)
}
