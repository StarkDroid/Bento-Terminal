package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type actionID int

const (
	actNext actionID = iota
	actPrev
	actKeys
	actSettings
	actQuit

	actFileUp
	actFileDown
	actFilePageUp
	actFilePageDown
	actFileCollapse
	actFileExpand
	actFileOpen

	actHistoryUp
	actHistoryDown
	actHistoryPageUp
	actHistoryPageDown
	actHistoryTabPrev
	actHistoryTabNext
	actHistoryOpen

	actBuildUp
	actBuildDown
	actBuildPageUp
	actBuildPageDown
	actBuildRun
	actBuildMode
	actBuildRerun
	actBuildCancel
	actBuildTabPrev
	actBuildTabNext
	actTermRun
	actTermCancel
	actTermClear

	actMainUp
	actMainDown
	actMainLeft
	actMainRight
	actMainPageUp
	actMainPageDown
	actMainFold
	actMainUnfold
	actMainToggleFold
	actMainSyntax
	actMainEdit
	actMainNew
	actMainSave
	actMainClose

	actFindFile
	actFindText
)

// action is one keyboard command. Binding supplies the keys and their label;
// desc is the pane-specific wording used by the command list.
type action struct {
	id      actionID
	binding key.Binding
	desc    string
}

// paneActions lists the commands owned by one pane.
func (m Model) paneActions(p Pane) []action {
	switch p {
	case PaneMain:
		actions := []action{
			{actMainUp, m.keys.Up, "move up a line"},
			{actMainDown, m.keys.Down, "move down a line"},
			{actMainLeft, m.keys.Left, "move left one character"},
			{actMainRight, m.keys.Right, "move right one character"},
			{actMainPageUp, m.keys.PageUp, "page up"},
			{actMainPageDown, m.keys.PageDown, "page down"},
			{actMainFold, m.keys.Fold, "fold function or class"},
			{actMainUnfold, m.keys.Unfold, "unfold function or class"},
			{actMainToggleFold, m.keys.Enter, "fold or unfold"},
			{actMainEdit, m.keys.Edit, "edit this file"},
			{actMainNew, m.keys.New, "create a new file"},
			{actMainSave, m.keys.Save, "save the file"},
			{actMainSyntax, m.keys.Syntax, "toggle code highlighting"},
		}
		if m.previewKind == previewCommit || m.previewKind == previewDiff {
			actions = append(actions, action{actMainClose, m.keys.Close, "close the preview"})
		}
		return actions
	case PaneFiles:
		return []action{
			{actFileUp, m.keys.Up, "move up"},
			{actFileDown, m.keys.Down, "move down"},
			{actFilePageUp, m.keys.PageUp, "page up"},
			{actFilePageDown, m.keys.PageDown, "page down"},
			{actFileCollapse, m.keys.Left, "close directory"},
			{actFileExpand, m.keys.Right, "open directory"},
			{actFileOpen, m.keys.Enter, "preview file"},
		}
	case PaneHistory:
		return []action{
			{actHistoryUp, m.keys.Up, "move up"},
			{actHistoryDown, m.keys.Down, "move down"},
			{actHistoryPageUp, m.keys.PageUp, "page up"},
			{actHistoryPageDown, m.keys.PageDown, "page down"},
			{actHistoryTabPrev, m.keys.Left, "previous tab"},
			{actHistoryTabNext, m.keys.Right, "next tab"},
			{actHistoryOpen, m.keys.Enter, "open the selection"},
		}
	case PaneBuild:
		if m.buildTab == buildTabTerminal {
			return m.terminalActions()
		}
		return m.buildActions()
	}
	return nil
}

func (m Model) buildActions() []action {
	return []action{
		{actBuildUp, m.keys.Up, "move up"},
		{actBuildDown, m.keys.Down, "move down"},
		{actBuildPageUp, m.keys.PageUp, "scroll output up"},
		{actBuildPageDown, m.keys.PageDown, "scroll output down"},
		{actBuildRun, m.keys.Enter, "run selected command"},
		{actBuildMode, m.keys.Mode, "switch auto or manual"},
		{actBuildRerun, m.keys.Rerun, "rerun last command"},
		{actBuildCancel, m.keys.Cancel, "cancel the run"},
		{actBuildTabPrev, m.keys.Left, "previous tab"},
		{actBuildTabNext, m.keys.Right, "next tab"},
	}
}

func (m Model) terminalActions() []action {
	return []action{
		{actTermRun, m.keys.Enter, "run the typed command"},
		{actTermCancel, m.keys.Interrupt, "stop the running command"},
		{actTermClear, m.keys.Close, "clear the input"},
	}
}

func (m Model) globalActions() []action {
	return []action{
		{actNext, m.keys.Next, "focus next pane"},
		{actPrev, m.keys.Prev, "focus previous pane"},
		{actKeys, m.keys.Help, "command list for this pane"},
		{actSettings, m.keys.Settings, "settings and about"},
		{actFindFile, m.keys.FindFile, "find a file in the project"},
		{actFindText, m.keys.FindText, "find in this file"},
		{actQuit, m.keys.Quit, "quit Bento terminal"},
	}
}

// actions returns every command that may run now. Build commands work from any
// pane, so they are kept in the list while another pane has focus.
func (m Model) actions() []action {
	list := m.paneActions(m.focus)
	if m.focus != PaneBuild {
		list = append(list, m.buildActions()...)
	}
	return append(list, m.globalActions()...)
}

func (m Model) paletteItems() []paletteItem {
	var items []paletteItem
	for _, a := range m.paneActions(m.palettePane) {
		items = append(items, paletteItem{section: m.paneTitle(m.palettePane), action: a})
	}
	for _, a := range m.globalActions() {
		items = append(items, paletteItem{section: "Global", action: a})
	}
	return items
}

func (m Model) runAction(id actionID) (tea.Model, tea.Cmd) {
	m.paletteOpen = false
	switch id {
	case actNext:
		m.focus = Cycle(m.focus, 1)
		return m.reflowPreview(false), nil
	case actPrev:
		m.focus = Cycle(m.focus, -1)
		return m.reflowPreview(false), nil
	case actKeys:
		return m.openPalette(), nil
	case actSettings:
		return m.openSettings(), nil
	case actQuit:
		return m.stopRun(), tea.Quit

	case actFileUp:
		return m.moveFile(-1), nil
	case actFileDown:
		return m.moveFile(1), nil
	case actFilePageUp:
		return m.moveFile(-m.filePage()), nil
	case actFilePageDown:
		return m.moveFile(m.filePage()), nil
	case actFileCollapse:
		if m.tree != nil {
			m.tree.Collapse(m.fileCursor)
		}
	case actFileExpand:
		if m.tree != nil {
			m.tree.Expand(m.fileCursor)
		}
	case actFileOpen:
		return m.openSelection()

	case actHistoryUp:
		return m.moveHistory(-1), nil
	case actHistoryDown:
		return m.moveHistory(1), nil
	case actHistoryPageUp:
		return m.moveHistory(-m.historyPage()), nil
	case actHistoryPageDown:
		return m.moveHistory(m.historyPage()), nil
	case actHistoryTabPrev:
		return m.switchHistoryTab(-1), nil
	case actHistoryTabNext:
		return m.switchHistoryTab(1), nil
	case actHistoryOpen:
		return m.openHistorySelection()

	case actBuildUp:
		return m.moveCmd(-1), nil
	case actBuildDown:
		return m.moveCmd(1), nil
	case actBuildPageUp:
		return m.scrollOutput(-m.outputPage()), nil
	case actBuildPageDown:
		return m.scrollOutput(m.outputPage()), nil
	case actBuildRun:
		cmds := m.commands()
		if m.cmdCursor < 0 || m.cmdCursor >= len(cmds) {
			return m, nil
		}
		return m.launch(cmds[m.cmdCursor])
	case actBuildMode:
		return m.toggleMode(), nil
	case actBuildRerun:
		return m.rerun()
	case actBuildCancel:
		return m.stopRun(), nil
	case actBuildTabPrev:
		return m.switchBuildTab(-1), nil
	case actBuildTabNext:
		return m.switchBuildTab(1), nil
	case actTermRun:
		return m.runTerminal()
	case actTermCancel:
		return m.stopRun(), nil
	case actTermClear:
		m.termInput = ""
		m.termCursor = 0
		return m, nil

	case actMainUp:
		return m.movePreview(-1), nil
	case actMainDown:
		return m.movePreview(1), nil
	case actMainLeft:
		return m.movePreviewCol(-1), nil
	case actMainRight:
		return m.movePreviewCol(1), nil
	case actMainPageUp:
		return m.movePreview(-m.previewPage()), nil
	case actMainPageDown:
		return m.movePreview(m.previewPage()), nil
	case actMainFold:
		return m.setFold(true), nil
	case actMainUnfold:
		return m.setFold(false), nil
	case actMainToggleFold:
		return m.toggleFold(), nil
	case actMainSyntax:
		return m.toggleSyntax(), nil
	case actMainEdit:
		return m.startEditing(), nil
	case actMainNew:
		return m.openNewFile(), nil
	case actMainSave:
		return m.savePreview()
	case actMainClose:
		return m.closeCommitPreview(), nil

	case actFindFile:
		return m.openFinder()
	case actFindText:
		return m.openFind(), nil
	}
	return m, nil
}

func keyLabel(b key.Binding) string {
	return b.Help().Key
}

func (m Model) paneTitle(p Pane) string {
	switch p {
	case PaneMain:
		return "Main"
	case PaneFiles:
		return "Files"
	case PaneHistory:
		return "History"
	case PaneBuild:
		return "Build"
	}
	return "Bento"
}
