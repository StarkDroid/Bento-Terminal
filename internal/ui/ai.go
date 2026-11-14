package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hinshun/vt10x"

	"github.com/StarkDroid/Bento-Terminal/internal/agent"
	"github.com/StarkDroid/Bento-Terminal/internal/build"
)

// aiEventKind is a message from a running agent session.
type aiEventKind int

const (
	aiOutputEvent aiEventKind = iota
	aiExitEvent
)

type aiEvent struct {
	id   int
	kind aiEventKind
	err  error
}

type aiEventMsg struct {
	event aiEvent
}

// aiRow is one selectable agent. Installed agents come first.
type aiRow struct {
	provider  agent.Provider
	path      string
	installed bool
}

// aiSession is one attached coding agent. The interface lets tests drive the
// panel without a real pseudo-terminal.
type aiSession interface {
	Write(p []byte) error
	Resize(width, height int) error
	Screen() vt10x.Terminal
	Close()
}

// aiRows orders installed agents first, then the rest, both in catalogue
// order.
func (m Model) aiRows() []aiRow {
	all := agent.All()
	rows := make([]aiRow, 0, len(all))
	for _, p := range all {
		if path, ok := m.aiFound[p.ID]; ok {
			rows = append(rows, aiRow{provider: p, path: path, installed: true})
		}
	}
	for _, p := range all {
		if _, ok := m.aiFound[p.ID]; !ok {
			rows = append(rows, aiRow{provider: p})
		}
	}
	return rows
}

func (m Model) aiCurrent() (aiRow, bool) {
	rows := m.aiRows()
	if len(rows) == 0 {
		return aiRow{}, false
	}
	return rows[clamp(m.aiCursor, 0, len(rows)-1)], true
}

// toggleAI opens or closes the panel and focuses it when it opens.
func (m Model) toggleAI() (tea.Model, tea.Cmd) {
	if m.aiOpen {
		m.aiOpen = false
		if m.focus == PaneAI {
			m.focus = PaneMain
		}
		return m.reflowPreview(false), nil
	}
	m.aiOpen = true
	m.focus = PaneAI
	m = m.clampAI()
	return m.reflowPreview(false), detectAgentsCmd()
}

func (m Model) clampAI() Model {
	n := len(m.aiRows())
	if n == 0 {
		m.aiCursor = 0
		m.aiScroll = 0
		return m
	}
	m.aiCursor = clamp(m.aiCursor, 0, n-1)
	m.aiScroll = ensure(m.aiCursor, m.aiScroll, m.aiPage())
	return m
}

func (m Model) moveAI(delta int) Model {
	n := len(m.aiRows())
	if n == 0 {
		return m
	}
	m.aiCursor = clamp(m.aiCursor+delta, 0, n-1)
	m.aiScroll = ensure(m.aiCursor, m.aiScroll, m.aiPage())
	return m
}

func (m Model) aiPage() int {
	lay := computeLayout(m.width, m.height)
	h := lay.Main.H - 4
	if h < 1 {
		return 1
	}
	return h
}

// aiSessionSize is the emulator size for the agent panel, in cells.
func (m Model) aiSessionSize() (int, int) {
	lay := computeLayout(m.width, m.height)
	if !lay.OK {
		return 80, 24
	}
	_, aiW := lay.MainSplit()
	w := aiW - 2
	h := lay.Main.H - 3
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

func (m Model) onAgentsDetected(found map[string]string) Model {
	m.aiFound = found
	m.aiReady = true
	if m.aiInstall != "" && !m.running {
		name := providerName(m.aiInstall)
		if _, ok := found[m.aiInstall]; ok {
			m.aiNote = name + " is ready"
			m.aiErr = ""
		} else {
			m.aiErr = name + " is not on PATH yet"
		}
		m.aiInstall = ""
		m.aiLines = nil
	}
	return m.clampAI()
}

// activateAI starts the selected agent, or installs it when it is missing.
func (m Model) activateAI() (tea.Model, tea.Cmd) {
	row, ok := m.aiCurrent()
	if !ok {
		return m, nil
	}
	m.aiErr = ""
	m.aiNote = ""
	if row.installed {
		return m.startAISession(row.provider)
	}
	return m.startAIInstall(row.provider)
}

// startAIInstall runs the provider's installer through the user's shell and
// streams its output into the panel.
func (m Model) startAIInstall(p agent.Provider) (tea.Model, tea.Cmd) {
	script := p.InstallCommand()
	if script == "" {
		m.aiErr = "No installer for " + p.Name
		return m, nil
	}
	m = m.stopRun()
	m.cancelRequested = false
	m.runID++
	id := m.runID
	m.running = true
	m.runTarget = targetAI
	m.aiState = stateRunning
	m.aiInstall = p.ID
	m.aiErr = ""
	m.aiNote = ""
	m.aiLines = nil
	m.aiCursor = 0
	m.aiScroll = 0
	m = m.appendAILine("$ " + script)
	return m, tea.Batch(startCmd(id, m.root, build.ShellCommand(script)), m.spinner.Tick)
}

func (m Model) appendAILine(line string) Model {
	clean := sanitizeLine(line)
	if line != "" && clean == "" {
		return m
	}
	m.aiLines = append(m.aiLines, clean)
	if len(m.aiLines) > 500 {
		m.aiLines = m.aiLines[len(m.aiLines)-500:]
	}
	return m
}

// startAISession runs the selected agent on a pseudo-terminal inside the panel.
func (m Model) startAISession(p agent.Provider) (tea.Model, tea.Cmd) {
	m = m.closeAISession()
	m.aiRun++
	id := m.aiRun
	events := make(chan aiEvent, 128)
	m.aiEvents = events
	width, height := m.aiSessionSize()
	session, err := agent.StartSession(m.root, p, width, height,
		func() {
			select {
			case events <- aiEvent{id: id, kind: aiOutputEvent}:
			default:
			}
		},
		func(err error) {
			events <- aiEvent{id: id, kind: aiExitEvent, err: err}
		},
	)
	if err != nil {
		m.aiErr = "Could not start " + p.Name + ": " + err.Error()
		m.aiSession = nil
		m.aiSessionID = ""
		m.aiState = stateIdle
		return m, nil
	}
	m.aiSession = session
	m.aiSessionID = p.ID
	m.aiState = stateRunning
	m.aiCols = width
	m.aiScreenH = height
	m.aiErr = ""
	m.aiNote = ""
	m.aiInstall = ""
	m.focus = PaneAI
	return m, waitAICmd(events)
}

// onAISessionKey forwards every key to the agent except the detach chord.
func (m Model) onAISessionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Detach) {
		return m.stopAISession("detached from " + providerName(m.aiSessionID)), nil
	}
	if m.aiSession != nil {
		_ = m.aiSession.Write(encodeKey(msg))
	}
	return m, nil
}

func (m Model) onAIEvent(msg aiEventMsg) (tea.Model, tea.Cmd) {
	ev := msg.event
	if m.aiSession == nil || ev.id != m.aiRun {
		return m, nil
	}
	if ev.kind == aiOutputEvent {
		return m, waitAICmd(m.aiEvents)
	}
	name := providerName(m.aiSessionID)
	m.aiSession = nil
	m.aiSessionID = ""
	m.aiState = stateIdle
	m.aiRun++
	if ev.err != nil {
		m.aiNote = name + " exited: " + ev.err.Error()
	} else {
		m.aiNote = name + " exited"
	}
	return m, nil
}

func (m Model) stopAISession(note string) Model {
	if m.aiSession == nil {
		return m
	}
	m.aiSession.Close()
	m.aiSession = nil
	m.aiSessionID = ""
	m.aiState = stateIdle
	m.aiRun++
	if note != "" {
		m.aiNote = note
	}
	return m
}

func (m Model) closeAISession() Model {
	return m.stopAISession("")
}

func providerName(id string) string {
	if p, ok := agent.Find(id); ok {
		return p.Name
	}
	if id == "" {
		return "agent"
	}
	return id
}

// renderAI draws either the agent list or the running agent's screen.
func (m Model) renderAI(width, height int) string {
	if width < 4 || height < 4 {
		return ""
	}
	focused := m.focus == PaneAI
	innerW := width - 2
	if m.aiSession != nil {
		body := renderAgentScreen(m.aiSession.Screen(), innerW, height-3, focused)
		return renderPanel(m.aiTitle(innerW), body, width, height, focused)
	}
	bodyH := height - 4
	if bodyH < 0 {
		bodyH = 0
	}
	var lines []string
	switch {
	case m.aiInstall != "":
		lines = m.aiInstallLines(innerW, bodyH)
	case !m.aiReady:
		lines = []string{noteLine(" Looking for coding agents…", innerW)}
	default:
		lines = m.aiListLines(innerW, bodyH)
	}
	for len(lines) < bodyH {
		lines = append(lines, noteLine("", innerW))
	}
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	if bodyH > 0 {
		lines = append(lines, m.footer(PaneAI, innerW))
	}
	return renderPanel(m.aiTitle(innerW), strings.Join(lines, "\n"), width, height, focused)
}

func (m Model) aiTitle(width int) string {
	focused := m.focus == PaneAI
	left := []piece{
		pieceText(" ▾ ", colorLagoonBright, colorPanel),
		tabPiece("AI", true, focused),
	}
	var right []piece
	switch {
	case m.aiSession != nil:
		right = append(right,
			pieceText("● ", colorPalm, colorPanel),
			pieceText(providerName(m.aiSessionID)+" ", colorPalm, colorPanel),
		)
	case m.aiInstall != "" && m.aiState == stateRunning:
		right = append(right,
			pieceText(m.spinner.View(), colorLagoonBright, colorPanel),
			pieceText(" installing ", colorLagoon, colorPanel),
		)
	case m.aiInstall != "" && m.aiState == stateFailed:
		right = append(right, pieceText(" install failed ", colorCoral, colorPanel))
	case !m.aiReady:
		right = append(right, pieceText(" detecting ", colorSandDim, colorPanel))
	default:
		right = append(right, pieceText(fmt.Sprintf(" %d ready ", len(m.aiFound)), colorPalm, colorPanel))
	}
	return joinLeftRight(width, colorPanel, left, right)
}

// aiListLines draws the notice row (if any) and the scrollable agent list.
func (m Model) aiListLines(width, height int) []string {
	var lines []string
	switch {
	case m.aiErr != "":
		lines = append(lines, lipgloss.NewStyle().
			Foreground(colorCoral).
			Background(colorPanel).
			Width(width).
			Render(truncateCells(" "+m.aiErr, width)))
	case m.aiNote != "":
		lines = append(lines, noteLine(" "+m.aiNote, width))
	}
	rows := m.aiRows()
	if len(rows) == 0 {
		lines = append(lines, noteLine(" No coding agents found", width))
		return lines
	}
	listH := height - len(lines)
	if listH < 1 {
		listH = 1
	}
	start := ensure(m.aiCursor, m.aiScroll, listH)
	end := min(start+listH, len(rows))
	for i := start; i < end; i++ {
		lines = append(lines, m.aiLine(rows[i], width, i == m.aiCursor))
	}
	return lines
}

// aiLine shows install state, name, and whether enter starts or installs.
func (m Model) aiLine(row aiRow, width int, selected bool) string {
	if width < 1 {
		return ""
	}
	bg := lipgloss.TerminalColor(colorPanel)
	markerFG := lipgloss.TerminalColor(colorMoss)
	marker := "  "
	if selected {
		if m.focus == PaneAI {
			bg = colorSelect
		} else {
			bg = colorSelectDim
		}
		marker = "▸ "
		markerFG = colorLagoonBright
	}
	dot := "○"
	dotFG := lipgloss.TerminalColor(colorMoss)
	label := "install"
	labelFG := lipgloss.TerminalColor(colorLagoon)
	if row.installed {
		dot = "●"
		dotFG = colorPalm
		label = "ready"
		labelFG = colorPalm
	}
	labelW := lipgloss.Width(label) + 1
	if width < 12 {
		label = ""
		labelW = 0
	}
	nameW := width - 4 - labelW
	if nameW < 1 {
		nameW = 1
	}
	name := lipgloss.NewStyle().Foreground(colorSand).Background(bg)
	if selected {
		name = name.Bold(true)
	}
	line := lipgloss.NewStyle().Foreground(markerFG).Background(bg).Render(marker) +
		lipgloss.NewStyle().Foreground(dotFG).Background(bg).Render(dot+" ") +
		name.Width(nameW).Render(truncateCells(row.provider.Name, nameW))
	if label != "" {
		line += lipgloss.NewStyle().Foreground(labelFG).Background(bg).Render(label + " ")
	}
	if used := lipgloss.Width(line); used < width {
		line += lipgloss.NewStyle().Background(bg).Width(width - used).Render("")
	}
	return line
}

func (m Model) aiInstallLines(width, height int) []string {
	if height <= 0 {
		return nil
	}
	if len(m.aiLines) == 0 {
		return []string{noteLine(" Installing…", width)}
	}
	start := len(m.aiLines) - height
	if start < 0 {
		start = 0
	}
	return strings.Split(windowLines(stylePlain(m.aiLines, width), start, width, height), "\n")
}

// renderAgentScreen turns the emulator's cells into styled rows. Only the
// emulator's own styles are used, so the agent draws its own interface.
func renderAgentScreen(emu vt10x.Terminal, width, height int, focused bool) string {
	if emu == nil || width < 1 || height < 1 {
		return ""
	}
	emu.Lock()
	defer emu.Unlock()
	cols, rows := emu.Size()
	if width > cols {
		width = cols
	}
	if height > rows {
		height = rows
	}
	cursorX, cursorY := -1, -1
	if focused && emu.CursorVisible() {
		pos := emu.Cursor()
		if pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height {
			cursorX, cursorY = pos.X, pos.Y
		}
	}
	out := make([]string, height)
	for y := 0; y < height; y++ {
		var b strings.Builder
		var run strings.Builder
		var runStyle cellStyle
		flush := func() {
			if run.Len() == 0 {
				return
			}
			b.WriteString(runStyle.render(run.String()))
			run.Reset()
		}
		used := 0
		for x := 0; x < width; x++ {
			cell := emu.Cell(x, y)
			text := " "
			if cell.Char != 0 {
				text = string(cell.Char)
			}
			style := cellStyleFor(cell)
			if x == cursorX && y == cursorY {
				style = cellStyle{fg: colorOcean, bg: colorLagoonBright, bold: true}
			}
			if style != runStyle && run.Len() > 0 {
				flush()
			}
			runStyle = style
			run.WriteString(text)
			used++
		}
		flush()
		if used < width {
			b.WriteString(lipgloss.NewStyle().Background(colorPanel).Width(width - used).Render(""))
		}
		out[y] = b.String()
	}
	return strings.Join(out, "\n")
}

type cellStyle struct {
	fg        lipgloss.TerminalColor
	bg        lipgloss.TerminalColor
	bold      bool
	italic    bool
	underline bool
	strike    bool
	faint     bool
}

func (s cellStyle) render(text string) string {
	fg := s.fg
	if fg == nil {
		fg = colorSand
	}
	bg := s.bg
	if bg == nil {
		bg = colorPanel
	}
	st := lipgloss.NewStyle().Foreground(fg).Background(bg)
	if s.bold {
		st = st.Bold(true)
	}
	if s.italic {
		st = st.Italic(true)
	}
	if s.underline {
		st = st.Underline(true)
	}
	if s.strike {
		st = st.Strikethrough(true)
	}
	if s.faint {
		st = st.Faint(true)
	}
	return st.Render(text)
}

// vt10x keeps its glyph attribute bits private and applies reverse video to
// the stored colours itself, so only these bits are left to render.
const (
	glyphUnderline = 1 << 1
	glyphBold      = 1 << 2
	glyphItalic    = 1 << 4
	glyphBlink     = 1 << 5
)

func cellStyleFor(cell vt10x.Glyph) cellStyle {
	mode := int(cell.Mode)
	return cellStyle{
		fg:        cellColor(cell.FG, colorSand),
		bg:        cellColor(cell.BG, colorPanel),
		bold:      mode&glyphBold != 0,
		italic:    mode&glyphItalic != 0,
		underline: mode&glyphUnderline != 0,
	}
}

// cellColor converts a terminal cell colour to a lipgloss colour. The
// emulator stores ANSI indices below 256, packed RGB above that, and the
// default sentinels past every real colour.
func cellColor(c vt10x.Color, def lipgloss.TerminalColor) lipgloss.TerminalColor {
	switch {
	case c == vt10x.DefaultFG || c == vt10x.DefaultBG || c == vt10x.DefaultCursor:
		return def
	case c < 256:
		return lipgloss.Color(strconv.Itoa(int(c)))
	default:
		return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", byte(c>>16), byte(c>>8), byte(c)))
	}
}

// encodeKey turns a Bubble Tea key into the bytes a terminal application
// expects.
func encodeKey(msg tea.KeyMsg) []byte {
	if msg.Alt && msg.Type == tea.KeyRunes {
		return append([]byte{0x1b}, []byte(string(msg.Runes))...)
	}
	switch msg.Type {
	case tea.KeyRunes:
		return []byte(string(msg.Runes))
	case tea.KeySpace:
		return []byte(" ")
	case tea.KeyEnter:
		return []byte("\r")
	case tea.KeyTab:
		return []byte("\t")
	case tea.KeyShiftTab:
		return []byte("\x1b[Z")
	case tea.KeyBackspace:
		return []byte("\x7f")
	case tea.KeyEsc:
		return []byte("\x1b")
	case tea.KeyUp:
		return []byte("\x1b[A")
	case tea.KeyDown:
		return []byte("\x1b[B")
	case tea.KeyRight:
		return []byte("\x1b[C")
	case tea.KeyLeft:
		return []byte("\x1b[D")
	case tea.KeyHome:
		return []byte("\x1b[H")
	case tea.KeyEnd:
		return []byte("\x1b[F")
	case tea.KeyPgUp:
		return []byte("\x1b[5~")
	case tea.KeyPgDown:
		return []byte("\x1b[6~")
	case tea.KeyDelete:
		return []byte("\x1b[3~")
	case tea.KeyInsert:
		return []byte("\x1b[2~")
	}
	if msg.Type >= tea.KeyCtrlA && msg.Type <= tea.KeyCtrlZ {
		return []byte{byte(msg.Type)}
	}
	if msg.Type == tea.KeyCtrlAt {
		return []byte{0}
	}
	return nil
}
