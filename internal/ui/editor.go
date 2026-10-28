package ui

import (
	"errors"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/StarkDroid/Bento-Terminal/internal/files"
	"github.com/StarkDroid/Bento-Terminal/internal/syntax"
)

// sourceLines returns the lines the preview draws. While a file is being edited
// they come from the edit buffer, which keeps a trailing empty line so a file
// can grow.
func (m Model) sourceLines() []string {
	if m.previewEditing && m.previewEditLines != nil {
		return m.previewEditLines
	}
	return fileLines(m.previewRaw)
}

// openFile previews a project-relative path and resets the editor state.
func (m Model) openFile(rel, name string) Model {
	body, err := files.Preview(filepath.Join(m.root, filepath.FromSlash(rel)))
	m.previewKind = previewFile
	m.previewPath = rel
	m.previewTitle = name
	m.previewNumbered = false
	m.previewLang = syntax.Language{}
	m.previewSyntax = false
	m.previewEditing = false
	m.previewEditLines = nil
	m.previewDirty = false
	m.previewSaved = false
	m.previewSaveErr = false
	m.previewCursor = 0
	m.previewCol = 0
	m.previewFolds = map[int]bool{}
	m.previewBlocks = nil
	switch {
	case errors.Is(err, files.ErrBinary):
		m.previewRaw = "Binary file"
	case err != nil:
		m.previewRaw = err.Error()
	default:
		m.previewRaw = body
		m.previewNumbered = true
		m.previewLang = syntax.Detect(name, body)
		m.previewSyntax = m.previewLang.Detected()
		m.previewEditLines = fileLines(body)
	}
	m = m.prepareBlocks()
	m = m.refreshSpans()
	m.preview.GotoTop()
	return m.reflowPreview(true)
}

// startEditing turns the open text file into an editable buffer with the caret
// at its current line.
func (m Model) startEditing() Model {
	if m.previewKind != previewFile || !m.previewNumbered {
		return m
	}
	m.previewEditing = true
	m.previewEditLines = append([]string(nil), fileLines(m.previewRaw)...)
	if len(m.previewEditLines) == 0 {
		m.previewEditLines = []string{""}
	}
	m.previewCursor = clamp(m.previewCursor, 0, len(m.previewEditLines)-1)
	m.previewDirty = false
	m.previewSaved = false
	m.previewSaveErr = false
	m.previewFolds = map[int]bool{}
	m.previewBlocks = nil
	m.previewCol = clamp(m.previewCol, 0, len([]rune(m.previewEditLines[m.previewCursor])))
	m = m.refreshSpans()
	m = m.reflowPreview(false)
	return m.scrollPreview()
}

// leaveEditing keeps the buffer in the preview, marks the title dirty, and
// restores folding and highlighting.
func (m Model) leaveEditing() Model {
	if !m.previewEditing {
		return m
	}
	m = m.syncEditRaw()
	m.previewEditing = false
	m = m.prepareBlocks()
	m = m.refreshSpans()
	m = m.reflowPreview(false)
	return m.scrollPreview()
}

// syncEditRaw copies the edit buffer into the preview text so the rendered
// preview and an unsaved file still agree.
func (m Model) syncEditRaw() Model {
	if m.previewEditLines == nil {
		return m
	}
	m.previewRaw = strings.Join(m.previewEditLines, "\n")
	m.previewEditLines = fileLines(m.previewRaw)
	return m
}

// onEditorKey is the editor mode of the preview pane. Printable keys type,
// arrows move the caret, enter splits a line, and ctrl+s saves.
func (m Model) onEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlF {
		return m.openFind(), nil
	}
	switch msg.Type {
	case tea.KeyCtrlS:
		return m.savePreview()
	case tea.KeyEsc:
		return m.leaveEditing(), nil
	case tea.KeyUp:
		return m.movePreview(-1), nil
	case tea.KeyDown:
		return m.movePreview(1), nil
	case tea.KeyLeft:
		return m.movePreviewCol(-1), nil
	case tea.KeyRight:
		return m.movePreviewCol(1), nil
	case tea.KeyPgUp:
		return m.movePreview(-m.previewPage()), nil
	case tea.KeyPgDown:
		return m.movePreview(m.previewPage()), nil
	case tea.KeyHome:
		m.previewCol = 0
		return m.scrollPreview(), nil
	case tea.KeyEnd:
		m.previewCol = len([]rune(m.previewEditLines[m.previewCursor]))
		return m.scrollPreview(), nil
	case tea.KeyEnter:
		return m.editNewline(), nil
	case tea.KeyBackspace:
		return m.editBackspace(), nil
	case tea.KeyDelete:
		return m.editDelete(), nil
	case tea.KeySpace:
		return m.editInsert(" "), nil
	case tea.KeyTab:
		return m.editInsert("    "), nil
	case tea.KeyRunes:
		return m.editInsert(string(msg.Runes)), nil
	}
	return m, nil
}

// editInsert types text at the caret.
func (m Model) editInsert(text string) Model {
	if !m.previewEditing || m.previewCursor < 0 || m.previewCursor >= len(m.previewEditLines) {
		return m
	}
	runes := []rune(m.previewEditLines[m.previewCursor])
	col := clamp(m.previewCol, 0, len(runes))
	typed := []rune(text)
	next := make([]rune, 0, len(runes)+len(typed))
	next = append(next, runes[:col]...)
	next = append(next, typed...)
	next = append(next, runes[col:]...)
	m.previewEditLines[m.previewCursor] = string(next)
	m.previewCol = col + len(typed)
	return m.afterEdit()
}

// editNewline splits the caret line in two.
func (m Model) editNewline() Model {
	if !m.previewEditing || m.previewCursor < 0 || m.previewCursor >= len(m.previewEditLines) {
		return m
	}
	runes := []rune(m.previewEditLines[m.previewCursor])
	col := clamp(m.previewCol, 0, len(runes))
	lines := make([]string, 0, len(m.previewEditLines)+1)
	lines = append(lines, m.previewEditLines[:m.previewCursor]...)
	lines = append(lines, string(runes[:col]), string(runes[col:]))
	lines = append(lines, m.previewEditLines[m.previewCursor+1:]...)
	m.previewEditLines = lines
	m.previewCursor++
	m.previewCol = 0
	return m.afterEdit()
}

// editBackspace removes the character before the caret, joining lines at a
// line start.
func (m Model) editBackspace() Model {
	if !m.previewEditing || m.previewCursor < 0 || m.previewCursor >= len(m.previewEditLines) {
		return m
	}
	runes := []rune(m.previewEditLines[m.previewCursor])
	col := clamp(m.previewCol, 0, len(runes))
	switch {
	case col > 0:
		next := make([]rune, 0, len(runes)-1)
		next = append(next, runes[:col-1]...)
		next = append(next, runes[col:]...)
		m.previewEditLines[m.previewCursor] = string(next)
		m.previewCol = col - 1
	case m.previewCursor > 0:
		prev := []rune(m.previewEditLines[m.previewCursor-1])
		m.previewEditLines[m.previewCursor-1] = string(prev) + string(runes)
		lines := make([]string, 0, len(m.previewEditLines)-1)
		lines = append(lines, m.previewEditLines[:m.previewCursor]...)
		lines = append(lines, m.previewEditLines[m.previewCursor+1:]...)
		m.previewEditLines = lines
		m.previewCursor--
		m.previewCol = len(prev)
	default:
		return m
	}
	return m.afterEdit()
}

// editDelete removes the character at the caret, joining the next line when the
// caret sits at a line end.
func (m Model) editDelete() Model {
	if !m.previewEditing || m.previewCursor < 0 || m.previewCursor >= len(m.previewEditLines) {
		return m
	}
	runes := []rune(m.previewEditLines[m.previewCursor])
	col := clamp(m.previewCol, 0, len(runes))
	switch {
	case col < len(runes):
		next := make([]rune, 0, len(runes)-1)
		next = append(next, runes[:col]...)
		next = append(next, runes[col+1:]...)
		m.previewEditLines[m.previewCursor] = string(next)
	case m.previewCursor < len(m.previewEditLines)-1:
		m.previewEditLines[m.previewCursor] = string(runes) + m.previewEditLines[m.previewCursor+1]
		lines := make([]string, 0, len(m.previewEditLines)-1)
		lines = append(lines, m.previewEditLines[:m.previewCursor+1]...)
		lines = append(lines, m.previewEditLines[m.previewCursor+2:]...)
		m.previewEditLines = lines
	default:
		return m
	}
	return m.afterEdit()
}

// afterEdit marks the buffer dirty, re-highlights it, and keeps the caret on
// screen.
func (m Model) afterEdit() Model {
	m.previewDirty = true
	m.previewSaved = false
	m.previewSaveErr = false
	m = m.refreshSpans()
	m = m.reflowPreview(false)
	return m.scrollPreview()
}

// savePreview writes the edit buffer back to its file and refreshes the tree
// and git summary.
func (m Model) savePreview() (tea.Model, tea.Cmd) {
	if m.previewKind != previewFile || m.previewPath == "" {
		return m, nil
	}
	if !m.previewEditing && !m.previewDirty {
		return m, nil
	}
	content := strings.Join(m.previewEditLines, "\n")
	if content != "" {
		content += "\n"
	}
	rel, err := files.Save(m.root, m.previewPath, content)
	if err != nil {
		m.previewSaveErr = true
		m.previewDirty = true
		return m, nil
	}
	m.previewPath = rel
	m.previewRaw = content
	m.previewEditLines = fileLines(content)
	m.previewDirty = false
	m.previewSaved = true
	m.previewSaveErr = false
	m = m.prepareBlocks()
	m = m.refreshSpans()
	m = m.reflowPreview(false)
	if m.tree != nil {
		_ = m.tree.Refresh()
	}
	return m, loadGitCmd(m.root)
}

// openNewFile shows the new-file prompt over the grid.
func (m Model) openNewFile() Model {
	m.newFileOpen = true
	m.newFilePath = ""
	m.newFileErr = ""
	m.paletteOpen = false
	m.settingsOpen = false
	return m
}

// onNewFileKey edits the new-file path, creates the file on enter, and closes
// on escape.
func (m Model) onNewFileKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.newFileOpen = false
		return m, nil
	case tea.KeyEnter:
		return m.createNewFile()
	case tea.KeyBackspace:
		runes := []rune(m.newFilePath)
		if len(runes) > 0 {
			m.newFilePath = string(runes[:len(runes)-1])
		}
		m.newFileErr = ""
		return m, nil
	case tea.KeySpace:
		m.newFilePath += " "
		m.newFileErr = ""
		return m, nil
	case tea.KeyRunes:
		m.newFilePath += string(msg.Runes)
		m.newFileErr = ""
		return m, nil
	}
	return m, nil
}

// createNewFile makes the typed path, opens it, and starts editing so the first
// line can be written straight away.
func (m Model) createNewFile() (tea.Model, tea.Cmd) {
	rel, err := files.Create(m.root, m.newFilePath)
	if err != nil {
		m.newFileErr = err.Error()
		return m, nil
	}
	m.newFileOpen = false
	m.focus = PaneMain
	if m.tree != nil {
		_ = m.tree.Refresh()
	}
	m = m.openFile(rel, rel)
	m = m.startEditing()
	return m, loadGitCmd(m.root)
}
