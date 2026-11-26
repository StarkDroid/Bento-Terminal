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
	actUpdate
	actNewTerm
	actNextTerm
	actPrevTerm
	actQuit

	actFileUp
	actFileDown
	actFilePageUp
	actFilePageDown
	actFileCollapse
	actFileExpand
	actFileOpen
	actFileRename
	actFileMove
	actFileCopy
	actFileDelete

	actHistoryUp
	actHistoryDown
	actHistoryPageUp
	actHistoryPageDown
	actHistoryTabPrev
	actHistoryTabNext
	actHistoryOpen

	actBuildTabPrev
	actBuildTabNext
	actBuildUp
	actBuildDown
	actBuildPageUp
	actBuildPageDown
	actBuildRun
	actBuildMode
	actBuildRerun
	actBuildCancel
	actTermRun
	actTermCancel
	actTermClear
	actTermClose

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
	actRecent

	actAIToggle
	actAIUp
	actAIDown
	actAIPageUp
	actAIPageDown
	actAIEnter
	actAIRescan
	actAIBack
	actAIStop

	actPaneMain
	actPaneAI
	actPaneFiles
	actPaneHistory
	actPaneBuild
	actPaneTerminal
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
		if m.previewKind != previewNone {
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
			{actFileRename, m.keys.Rename, "rename the selection"},
			{actFileMove, m.keys.Move, "move the selection"},
			{actFileCopy, m.keys.Copy, "copy the selection"},
			{actFileDelete, m.keys.Delete, "delete the selection"},
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
		return m.buildPaneActions()
	case PaneTerminal:
		return m.terminalActions()
	case PaneAI:
		return m.aiActions()
	}
	return nil
}

func (m Model) buildPaneActions() []action {
	tabs := []action{
		{actBuildTabPrev, m.keys.Left, "previous tab"},
		{actBuildTabNext, m.keys.Right, "next tab"},
	}
	return append(tabs, m.buildActions()...)
}

func (m Model) buildActions() []action {
	if m.buildTab == buildTabOutput {
		return []action{
			{actBuildUp, m.keys.Up, "scroll output up"},
			{actBuildDown, m.keys.Down, "scroll output down"},
			{actBuildPageUp, m.keys.PageUp, "scroll output up a page"},
			{actBuildPageDown, m.keys.PageDown, "scroll output down a page"},
			{actBuildRerun, m.keys.Rerun, "rerun last command"},
			{actBuildCancel, m.keys.Cancel, "cancel the run"},
		}
	}
	return []action{
		{actBuildUp, m.keys.Up, "move up"},
		{actBuildDown, m.keys.Down, "move down"},
		{actBuildPageUp, m.keys.PageUp, "move up a page"},
		{actBuildPageDown, m.keys.PageDown, "move down a page"},
		{actBuildRun, m.keys.Enter, "run selected command"},
		{actBuildMode, m.keys.Mode, "switch auto or manual"},
		{actBuildRerun, m.keys.Rerun, "rerun last command"},
		{actBuildCancel, m.keys.Cancel, "cancel the run"},
	}
}

func (m Model) terminalActions() []action {
	return []action{
		{actTermRun, m.keys.Enter, "run the typed command"},
		{actTermCancel, m.keys.Interrupt, "stop the running command"},
		{actTermClear, m.keys.Close, "clear the input or go back to Build"},
		{actTermClose, m.keys.CloseTab, "close this shell tab"},
		{actNextTerm, m.keys.NextTerm, "next shell tab"},
		{actPrevTerm, m.keys.PrevTerm, "previous shell tab"},
	}
}

// aiActions is the idle AI panel. A running agent owns its own keys, so the
// panel lists nothing then.
func (m Model) aiActions() []action {
	if m.aiSession != nil {
		return nil
	}
	if m.aiInstall != "" {
		if m.running && m.runTarget == targetAI {
			return []action{
				{actAIStop, m.keys.Close, "cancel the install"},
			}
		}
		return []action{
			{actAIStop, m.keys.Close, "close the install output"},
		}
	}
	enter := "start or install the selected agent"
	if row, ok := m.aiCurrent(); ok {
		if row.installed {
			enter = "start " + row.provider.Name
		} else {
			enter = "install " + row.provider.Name
		}
	}
	return []action{
		{actAIUp, m.keys.Up, "move up"},
		{actAIDown, m.keys.Down, "move down"},
		{actAIPageUp, m.keys.PageUp, "page up"},
		{actAIPageDown, m.keys.PageDown, "page down"},
		{actAIEnter, m.keys.Enter, enter},
		{actAIRescan, m.keys.Rerun, "rescan installed agents"},
		{actAIBack, m.keys.Close, "back to the preview"},
	}
}

// paneJumpActions are the number keys that focus a pane directly. They work
// everywhere the pane shortcuts work, including while an agent owns the AI
// panel.
func paneJumpActions() []action {
	return []action{
		{actPaneMain, key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "main pane")), "focus the Main pane"},
		{actPaneAI, key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "AI pane")), "focus the AI pane"},
		{actPaneFiles, key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "files pane")), "focus the Files pane"},
		{actPaneHistory, key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "history pane")), "focus the History pane"},
		{actPaneBuild, key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "build pane")), "focus the Build pane"},
		{actPaneTerminal, key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "terminal pane")), "focus the Terminal pane"},
	}
}

func (m Model) globalActions() []action {
	list := []action{
		{actNext, m.keys.Next, "focus next pane"},
		{actPrev, m.keys.Prev, "focus previous pane"},
		{actKeys, m.keys.Help, "command list for this pane"},
		{actSettings, m.keys.Settings, "settings and about"},
		{actFindFile, m.keys.FindFile, "find a file in the project"},
		{actFindText, m.keys.FindText, "find in this file"},
		{actRecent, m.keys.Recent, "open a recent project"},
		{actAIToggle, m.keys.ToggleAI, "open or close the AI panel"},
		{actNewTerm, m.keys.NewTerm, "new shell tab"},
		{actQuit, m.keys.Quit, "quit Bento terminal"},
	}
	list = append(list, paneJumpActions()...)
	if m.updateAvailable() {
		list = append(list, action{actUpdate, m.keys.Update, "update Bento"})
	}
	return list
}

// actions returns every command that may run now. Build commands work from any
// pane, so they are kept in the list while another pane has focus.
func (m Model) actions() []action {
	list := m.paneActions(m.focus)
	if m.focus != PaneBuild && m.focus != PaneTerminal {
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
		m.focus = m.nextPane(1)
		return m.reflowPreview(false), nil
	case actPrev:
		m.focus = m.nextPane(-1)
		return m.reflowPreview(false), nil
	case actKeys:
		return m.openPalette(), nil
	case actSettings:
		return m.openSettings(), nil
	case actUpdate:
		return m.startUpdate()
	case actNewTerm:
		return m.newTerminalTab()
	case actNextTerm:
		return m.cycleTerminalTab(1), nil
	case actPrevTerm:
		return m.cycleTerminalTab(-1), nil
	case actQuit:
		m = m.stopRun()
		m = m.closeAISession()
		m.SaveState()
		return m, tea.Quit

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
	case actFileRename:
		return m.openFileOp(fileOpRename), nil
	case actFileMove:
		return m.openFileOp(fileOpMove), nil
	case actFileCopy:
		return m.openFileOp(fileOpCopy), nil
	case actFileDelete:
		return m.openFileOp(fileOpDelete), nil

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

	case actBuildTabPrev:
		return m.switchBuildTab(-1), nil
	case actBuildTabNext:
		return m.switchBuildTab(1), nil
	case actBuildUp:
		if m.buildTab == buildTabOutput {
			return m.scrollOutput(-1), nil
		}
		return m.moveCmd(-1), nil
	case actBuildDown:
		if m.buildTab == buildTabOutput {
			return m.scrollOutput(1), nil
		}
		return m.moveCmd(1), nil
	case actBuildPageUp:
		if m.buildTab == buildTabOutput {
			return m.scrollOutput(-m.outputPage()), nil
		}
		return m.moveCmd(-m.commandPage()), nil
	case actBuildPageDown:
		if m.buildTab == buildTabOutput {
			return m.scrollOutput(m.outputPage()), nil
		}
		return m.moveCmd(m.commandPage()), nil
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
	case actTermRun:
		return m.runTerminal()
	case actTermCancel:
		return m.stopRun(), nil
	case actTermClear:
		m.termInput = ""
		m.termCursor = 0
		return m, nil
	case actTermClose:
		return m.closeTerminalTab()

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
		return m.closePreview(), nil

	case actFindFile:
		return m.openFinder()
	case actFindText:
		return m.openFind(), nil
	case actRecent:
		return m.openRecent()

	case actAIToggle:
		return m.toggleAI()
	case actPaneMain:
		return m.jumpPane(PaneMain)
	case actPaneAI:
		return m.jumpPane(PaneAI)
	case actPaneFiles:
		return m.jumpPane(PaneFiles)
	case actPaneHistory:
		return m.jumpPane(PaneHistory)
	case actPaneBuild:
		return m.jumpPane(PaneBuild)
	case actPaneTerminal:
		return m.jumpPane(PaneTerminal)
	case actAIUp:
		return m.moveAI(-1), nil
	case actAIDown:
		return m.moveAI(1), nil
	case actAIPageUp:
		return m.moveAI(-m.aiPage()), nil
	case actAIPageDown:
		return m.moveAI(m.aiPage()), nil
	case actAIEnter:
		return m.activateAI()
	case actAIRescan:
		return m, detectAgentsCmd()
	case actAIBack:
		m.focus = PaneMain
		return m.reflowPreview(false), nil
	case actAIStop:
		if m.aiInstall != "" {
			if m.running && m.runTarget == targetAI {
				m = m.stopRun()
				m.aiNote = "install cancelled"
			}
			m.aiInstall = ""
			m.aiLines = nil
			return m, nil
		}
		return m.stopAISession("detached"), nil
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
	case PaneTerminal:
		return "Terminal"
	case PaneAI:
		return "AI"
	}
	return "Bento"
}
