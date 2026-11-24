package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/StarkDroid/Bento-Terminal/internal/build"
)

// maxTermLines bounds the terminal scrollback.
const maxTermLines = 2000

// onTerminalKey owns every key while the terminal panel is focused, so typing
// a command never triggers a pane shortcut.
func (m Model) onTerminalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.termNote = ""
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
		m.focus = PaneBuild
		return m.reflowPreview(false), nil
	case tea.KeyCtrlA:
		if m.termInput == "" {
			return m.toggleAI()
		}
		return m, nil
	case tea.KeyCtrlR:
		if m.termInput == "" {
			return m.openRecent()
		}
		return m, nil
	case tea.KeyCtrlT:
		return m.newTerminalTab()
	case tea.KeyCtrlW:
		return m.closeTerminalTab()
	case tea.KeyCtrlRight, tea.KeyCtrlN:
		return m.cycleTerminalTab(1), nil
	case tea.KeyCtrlLeft, tea.KeyCtrlP:
		return m.cycleTerminalTab(-1), nil
	case tea.KeyTab:
		m.focus = m.nextPane(1)
		return m.reflowPreview(false), nil
	case tea.KeyShiftTab:
		m.focus = m.nextPane(-1)
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

// saveTermTab copies the active terminal fields into its tab slot.
func (m Model) saveTermTab() Model {
	if m.termIndex < 0 || m.termIndex >= len(m.termTabs) {
		return m
	}
	m.termTabs[m.termIndex] = terminalTab{
		title:   m.termTabs[m.termIndex].title,
		lines:   append([]string(nil), m.termLines...),
		follow:  m.termFollow,
		scroll:  m.termOutput.YOffset,
		input:   m.termInput,
		cursor:  m.termCursor,
		history: append([]string(nil), m.termHistory...),
		histPos: m.termHistPos,
		state:   m.termState,
	}
	return m
}

// loadTermTab copies a tab slot into the active terminal fields.
func (m Model) loadTermTab(index int) Model {
	if index < 0 || index >= len(m.termTabs) {
		return m
	}
	tab := m.termTabs[index]
	m.termIndex = index
	m.termLines = append([]string(nil), tab.lines...)
	m.termFollow = tab.follow
	m.termInput = tab.input
	m.termCursor = tab.cursor
	m.termHistory = append([]string(nil), tab.history...)
	m.termHistPos = tab.histPos
	m.termState = tab.state
	m.termOutput.SetContent(strings.Join(m.termLines, "\n"))
	m = m.syncSizes()
	if m.termFollow {
		m.termOutput.GotoBottom()
	} else {
		m.termOutput.SetYOffset(tab.scroll)
	}
	return m
}

func (m Model) switchTerminalTab(index int) Model {
	if index < 0 || index >= len(m.termTabs) || index == m.termIndex {
		return m
	}
	m.termNote = ""
	m = m.saveTermTab()
	m = m.loadTermTab(index)
	m.focus = PaneTerminal
	return m
}

func (m Model) cycleTerminalTab(delta int) Model {
	if len(m.termTabs) < 2 || delta == 0 {
		return m
	}
	index := (m.termIndex + delta) % len(m.termTabs)
	if index < 0 {
		index += len(m.termTabs)
	}
	return m.switchTerminalTab(index)
}

// terminalTabChipWidth is the cell width of the numbered tab for a 1-based
// index: one cell of fill on each side of the number.
func terminalTabChipWidth(index int) int {
	return len(fmt.Sprintf("%d", index)) + 2
}

// terminalRightWidth is the width of the label at the right of the Terminal
// title. The tab limit reserves it so a new tab never pushes the label out.
func (m Model) terminalRightWidth() int {
	switch {
	case m.running && m.runTarget == targetTerminal && m.termRunIndex == m.termIndex:
		return len(" running ")
	case m.running && m.runTarget == targetTerminal:
		return len(fmt.Sprintf(" shell %d running ", m.termRunIndex+1))
	}
	return len(" $SHELL ")
}

// maxTerminalTabs is how many numbered shell tabs fit in the title beside the
// glyph, the right label, and the + tab. A pane always allows at least one.
func (m Model) maxTerminalTabs(width int) int {
	avail := width - m.runGlyph(m.termState, false).width - len(" Terminal ") - m.terminalRightWidth() - 3
	used := 0
	count := 0
	for i := 1; ; i++ {
		next := terminalTabChipWidth(i)
		if used+next > avail {
			break
		}
		used += next
		count++
	}
	if count < 1 {
		return 1
	}
	return count
}

// terminalTabsFull reports whether another shell tab would not fit the panel
// at this width.
func (m Model) terminalTabsFull(width int) bool {
	return len(m.termTabs) >= m.maxTerminalTabs(width)
}

func (m Model) newTerminalTab() (tea.Model, tea.Cmd) {
	if lay := computeLayout(m.width, m.height); lay.OK {
		capacity := m.maxTerminalTabs(lay.Terminal.W - 2)
		if len(m.termTabs) >= capacity {
			m.termNote = fmt.Sprintf("no room · %d shell tabs max", capacity)
			return m, nil
		}
	}
	m.termNote = ""
	m = m.saveTermTab()
	m.termTabs = append(m.termTabs, terminalTab{
		title:   fmt.Sprintf("shell %d", len(m.termTabs)+1),
		follow:  true,
		histPos: -1,
	})
	m.termIndex = len(m.termTabs) - 1
	m.termLines = nil
	m.termInput = ""
	m.termCursor = 0
	m.termHistory = nil
	m.termHistPos = -1
	m.termState = stateIdle
	m.termFollow = true
	m.termOutput.SetContent("")
	m.termOutput.GotoTop()
	m.focus = PaneTerminal
	return m, nil
}

// closeTerminalTab removes the active shell tab. The last tab is cleared
// instead of removed, so the panel always keeps one shell.
func (m Model) closeTerminalTab() (tea.Model, tea.Cmd) {
	if len(m.termTabs) == 0 {
		return m, nil
	}
	m.termNote = ""
	index := m.termIndex
	if index < 0 || index >= len(m.termTabs) {
		index = 0
	}
	if m.running && m.runTarget == targetTerminal && m.termRunIndex == index {
		m = m.stopRun()
		m.termRunIndex = -1
	}
	m = m.saveTermTab()
	remaining := append(append([]terminalTab(nil), m.termTabs[:index]...), m.termTabs[index+1:]...)
	if m.running && m.runTarget == targetTerminal && m.termRunIndex > index {
		m.termRunIndex--
	}
	if len(remaining) == 0 {
		m.termTabs = []terminalTab{{title: "shell 1", follow: true, histPos: -1}}
		m.termIndex = 0
		m.termLines = nil
		m.termFollow = true
		m.termInput = ""
		m.termCursor = 0
		m.termHistory = nil
		m.termHistPos = -1
		m.termState = stateIdle
		m.termOutput.SetContent("")
		m.termOutput.GotoTop()
		m.focus = PaneTerminal
		return m, nil
	}
	for i := range remaining {
		remaining[i].title = fmt.Sprintf("shell %d", i+1)
	}
	m.termTabs = remaining
	if index >= len(remaining) {
		index = len(remaining) - 1
	}
	m = m.loadTermTab(index)
	m.focus = PaneTerminal
	return m, nil
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
	wasTerminalRun := m.running && m.runTarget == targetTerminal
	m = m.stopRun()
	m.cancelRequested = false
	if wasTerminalRun && m.termRunIndex >= 0 && m.termRunIndex < len(m.termTabs) && m.termRunIndex != m.termIndex {
		tab := m.termTabs[m.termRunIndex]
		tab.state = stateCanceled
		m.termTabs[m.termRunIndex] = tab
	}
	m.runID++
	id := m.runID
	m.running = true
	m.runTarget = targetTerminal
	m.termRunIndex = m.termIndex
	m.termState = stateRunning
	m.termFollow = true
	m.termLines = append(m.termLines, sanitizeLine("$ "+script))
	m = m.syncTermOutput()
	return m, tea.Batch(startCmd(id, m.root, build.ShellCommand(script)), m.spinner.Tick)
}

func (m Model) appendTerm(line string) Model {
	clean := sanitizeLine(line)
	if line != "" && clean == "" {
		// The command wrote nothing but control sequences, as "clear" does.
		return m
	}
	if m.termRunIndex >= 0 && m.termRunIndex != m.termIndex && m.termRunIndex < len(m.termTabs) {
		tab := m.termTabs[m.termRunIndex]
		tab.lines = append(tab.lines, clean)
		if len(tab.lines) > maxTermLines {
			tab.lines = tab.lines[len(tab.lines)-maxTermLines:]
		}
		m.termTabs[m.termRunIndex] = tab
		return m
	}
	m.termLines = append(m.termLines, clean)
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
	h := lay.Terminal.H - 5
	if h < 1 {
		return 1
	}
	return h
}

// appendRun writes process output to the buffer owned by the running target.
func (m Model) appendRun(line string) Model {
	switch m.runTarget {
	case targetTerminal:
		return m.appendTerm(line)
	case targetAI:
		return m.appendAILine(line)
	}
	return m.appendOut(line)
}

// setRunState records a result on the running target.
func (m Model) setRunState(state runState) Model {
	switch m.runTarget {
	case targetTerminal:
		if m.termRunIndex >= 0 && m.termRunIndex != m.termIndex && m.termRunIndex < len(m.termTabs) {
			tab := m.termTabs[m.termRunIndex]
			tab.state = state
			m.termTabs[m.termRunIndex] = tab
		} else {
			m.termState = state
		}
	case targetAI:
		m.aiState = state
	default:
		m.buildState = state
	}
	return m
}
