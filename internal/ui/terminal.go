package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
)

// maxTermLines bounds the terminal scrollback.
const maxTermLines = 2000

// onTerminalKey owns every key while the terminal tab is visible, so typing a
// command never triggers a pane shortcut.
func (m Model) onTerminalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m.runTerminal()
	case tea.KeyCtrlC:
		if m.running {
			return m.stopRun(), nil
		}
		m.termInput = ""
		m.termCursor = 0
		m.termHistPos = -1
		return m, nil
	case tea.KeyEsc:
		if m.termInput != "" {
			m.termInput = ""
			m.termCursor = 0
			m.termHistPos = -1
			return m, nil
		}
		m.buildTab = buildTabCommands
		return m, nil
	case tea.KeyTab:
		m.focus = Cycle(m.focus, 1)
		return m.reflowPreview(false), nil
	case tea.KeyShiftTab:
		m.focus = Cycle(m.focus, -1)
		return m.reflowPreview(false), nil
	case tea.KeyUp:
		return m.termHistoryStep(-1), nil
	case tea.KeyDown:
		return m.termHistoryStep(1), nil
	case tea.KeyLeft:
		m.termCursor = clamp(m.termCursor-1, 0, len([]rune(m.termInput)))
		return m, nil
	case tea.KeyRight:
		m.termCursor = clamp(m.termCursor+1, 0, len([]rune(m.termInput)))
		return m, nil
	case tea.KeyHome:
		m.termCursor = 0
		return m, nil
	case tea.KeyEnd:
		m.termCursor = len([]rune(m.termInput))
		return m, nil
	case tea.KeyBackspace:
		return m.termBackspace(), nil
	case tea.KeyDelete:
		return m.termDelete(), nil
	case tea.KeySpace:
		return m.termInsert(" "), nil
	case tea.KeyRunes:
		if m.termInput == "" {
			switch string(msg.Runes) {
			case "?":
				return m.openPalette(), nil
			case ",":
				return m.openSettings(), nil
			}
		}
		return m.termInsert(string(msg.Runes)), nil
	}
	return m, nil
}

func (m Model) termInsert(text string) Model {
	runes := []rune(m.termInput)
	col := clamp(m.termCursor, 0, len(runes))
	typed := []rune(text)
	next := make([]rune, 0, len(runes)+len(typed))
	next = append(next, runes[:col]...)
	next = append(next, typed...)
	next = append(next, runes[col:]...)
	m.termInput = string(next)
	m.termCursor = col + len(typed)
	return m
}

func (m Model) termBackspace() Model {
	runes := []rune(m.termInput)
	col := clamp(m.termCursor, 0, len(runes))
	if col == 0 {
		return m
	}
	next := make([]rune, 0, len(runes)-1)
	next = append(next, runes[:col-1]...)
	next = append(next, runes[col:]...)
	m.termInput = string(next)
	m.termCursor = col - 1
	return m
}

func (m Model) termDelete() Model {
	runes := []rune(m.termInput)
	col := clamp(m.termCursor, 0, len(runes))
	if col >= len(runes) {
		return m
	}
	next := make([]rune, 0, len(runes)-1)
	next = append(next, runes[:col]...)
	next = append(next, runes[col+1:]...)
	m.termInput = string(next)
	return m
}

// termHistoryStep walks previously run commands. Going past the newest entry
// clears the input.
func (m Model) termHistoryStep(delta int) Model {
	if len(m.termHistory) == 0 {
		return m
	}
	pos := m.termHistPos
	if pos < 0 {
		pos = len(m.termHistory)
	}
	pos = clamp(pos+delta, 0, len(m.termHistory))
	if pos == len(m.termHistory) {
		m.termHistPos = -1
		m.termInput = ""
	} else {
		m.termHistPos = pos
		m.termInput = m.termHistory[pos]
	}
	m.termCursor = len([]rune(m.termInput))
	return m
}

// runTerminal sends the typed line to the user's shell inside the project.
func (m Model) runTerminal() (tea.Model, tea.Cmd) {
	script := strings.TrimSpace(m.termInput)
	if script == "" {
		return m, nil
	}
	if len(m.termHistory) == 0 || m.termHistory[len(m.termHistory)-1] != script {
		m.termHistory = append(m.termHistory, script)
	}
	m.termHistPos = -1
	m.termInput = ""
	m.termCursor = 0
	m = m.stopRun()
	m.cancelRequested = false
	m.runID++
	id := m.runID
	m.running = true
	m.runTarget = targetTerminal
	m.termState = stateRunning
	m.termFollow = true
	m.termLines = append(m.termLines, "$ "+script)
	m = m.syncTermOutput()
	return m, tea.Batch(startCmd(id, m.root, build.ShellCommand(script)), m.spinner.Tick)
}

func (m Model) appendTerm(line string) Model {
	m.termLines = append(m.termLines, line)
	if len(m.termLines) > maxTermLines {
		m.termLines = m.termLines[len(m.termLines)-maxTermLines:]
	}
	return m.syncTermOutput()
}

// syncTermOutput sizes the terminal viewport, then keeps it pinned to the end
// while following.
func (m Model) syncTermOutput() Model {
	m = m.syncSizes()
	m.termOutput.SetContent(strings.Join(m.termLines, "\n"))
	if m.termFollow {
		m.termOutput.GotoBottom()
	}
	return m
}

func (m Model) scrollTerminal(delta int) Model {
	m = m.syncSizes()
	if delta > 0 {
		m.termOutput.ScrollDown(delta)
	} else {
		m.termOutput.ScrollUp(-delta)
	}
	m.termFollow = m.termOutput.AtBottom()
	return m
}

func (m Model) terminalPage() int {
	lay := computeLayout(m.width, m.height)
	h := lay.Build.H - 5
	if h < 1 {
		return 1
	}
	return h
}

// appendRun writes process output to the buffer owned by the running target.
func (m Model) appendRun(line string) Model {
	if m.runTarget == targetTerminal {
		return m.appendTerm(line)
	}
	return m.appendOut(line)
}

// setRunState records a result on the running target.
func (m Model) setRunState(state runState) Model {
	if m.runTarget == targetTerminal {
		m.termState = state
		return m
	}
	m.buildState = state
	return m
}
