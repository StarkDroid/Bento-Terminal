package ui

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/StarkDroid/Bento-Terminal/internal/files"
)

// openFileOp starts a rename, move, copy, or delete prompt for the selected
// tree row. The source is captured now so the prompt still works if the tree
// refreshes underneath it.
func (m Model) openFileOp(kind fileOpKind) Model {
	row, ok := m.selectedFileRow()
	if !ok {
		return m
	}
	m.fileOpOpen = true
	m.fileOpKind = kind
	m.fileOpSource = row.Rel
	m.fileOpErr = ""
	if kind == fileOpDelete {
		m.fileOpInput = ""
	} else {
		m.fileOpInput = row.Rel
	}
	m.paletteOpen = false
	m.settingsOpen = false
	return m
}

// selectedFileRow is the tree row under the file cursor.
func (m Model) selectedFileRow() (files.Row, bool) {
	if m.tree == nil {
		return files.Row{}, false
	}
	rows := m.tree.Rows()
	if m.fileCursor < 0 || m.fileCursor >= len(rows) {
		return files.Row{}, false
	}
	return rows[m.fileCursor], true
}

// onFileOpKey edits the destination path and runs the operation on enter. A
// delete is confirmed with enter or y and cancelled with esc or n.
func (m Model) onFileOpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.fileOpOpen = false
		return m, nil
	case tea.KeyEnter:
		return m.applyFileOp()
	case tea.KeyBackspace:
		if m.fileOpKind == fileOpDelete {
			return m, nil
		}
		runes := []rune(m.fileOpInput)
		if len(runes) > 0 {
			m.fileOpInput = string(runes[:len(runes)-1])
		}
		m.fileOpErr = ""
		return m, nil
	case tea.KeySpace:
		if m.fileOpKind == fileOpDelete {
			return m, nil
		}
		m.fileOpInput += " "
		m.fileOpErr = ""
		return m, nil
	case tea.KeyRunes:
		if m.fileOpKind == fileOpDelete {
			switch strings.ToLower(string(msg.Runes)) {
			case "y":
				return m.applyFileOp()
			case "n":
				m.fileOpOpen = false
			}
			return m, nil
		}
		m.fileOpInput += string(msg.Runes)
		m.fileOpErr = ""
		return m, nil
	}
	return m, nil
}

// applyFileOp runs the selected operation and refreshes the tree, the preview,
// and the git snapshot.
func (m Model) applyFileOp() (tea.Model, tea.Cmd) {
	src := m.fileOpSource
	dst := strings.TrimSpace(m.fileOpInput)
	var (
		rel string
		err error
	)
	switch m.fileOpKind {
	case fileOpRename:
		rel, err = files.Rename(m.root, src, dst)
	case fileOpMove:
		rel, err = files.Move(m.root, src, dst)
	case fileOpCopy:
		rel, err = files.Copy(m.root, src, dst)
	case fileOpDelete:
		rel, err = files.Remove(m.root, src)
	}
	if err != nil {
		m.fileOpErr = err.Error()
		return m, nil
	}
	m.fileOpOpen = false
	m.fileOpErr = ""
	if m.tree != nil {
		_ = m.tree.Refresh()
	}
	m = m.afterFileOp(src, rel)
	return m, loadGitCmd(m.root)
}

// afterFileOp keeps the preview and the file cursor on the moved or copied
// file, and closes the preview when the file it showed was deleted.
func (m Model) afterFileOp(src, dst string) Model {
	if m.previewKind == previewFile && withinRel(m.previewPath, src) {
		switch m.fileOpKind {
		case fileOpDelete:
			m = m.closePreview()
		case fileOpRename, fileOpMove:
			suffix := strings.TrimPrefix(strings.TrimPrefix(m.previewPath, src), "/")
			next := dst
			if suffix != "" {
				next = dst + "/" + suffix
			}
			if info, err := os.Stat(filepath.Join(m.root, filepath.FromSlash(next))); err == nil && !info.IsDir() {
				m = m.openFile(next, info.Name())
			} else {
				m = m.closePreview()
			}
		}
	}
	if m.tree != nil {
		for i, row := range m.tree.Rows() {
			if row.Rel == dst {
				m.fileCursor = i
				break
			}
		}
		rows := m.tree.Rows()
		if m.fileCursor >= len(rows) {
			m.fileCursor = len(rows) - 1
		}
		if m.fileCursor < 0 {
			m.fileCursor = 0
		}
		m.fileScroll = ensure(m.fileCursor, m.fileScroll, m.filePage())
	}
	m.focus = PaneFiles
	return m
}

// withinRel reports whether rel is path itself or sits under it.
func withinRel(rel, path string) bool {
	if path == "" {
		return false
	}
	return rel == path || strings.HasPrefix(rel, path+"/")
}

// fileOpTitle is the prompt heading.
func (m Model) fileOpTitle() string {
	switch m.fileOpKind {
	case fileOpRename:
		return "R Rename"
	case fileOpMove:
		return "M Move"
	case fileOpCopy:
		return "Y Copy"
	case fileOpDelete:
		return "D Delete"
	}
	return "File"
}

// fileOpVerb is the action word used in the prompt hint.
func (m Model) fileOpVerb() string {
	switch m.fileOpKind {
	case fileOpRename:
		return "rename"
	case fileOpMove:
		return "move"
	case fileOpCopy:
		return "copy"
	case fileOpDelete:
		return "delete"
	}
	return "apply"
}

// fileOpView is the modal prompt for rename, move, copy, and delete.
func (m Model) fileOpView() string {
	innerW := m.width - 6
	if innerW > 64 {
		innerW = 64
	}
	if innerW < 24 {
		innerW = m.width - 2
	}
	if innerW < 8 {
		innerW = 8
	}
	var rows []string
	if m.fileOpKind == fileOpDelete {
		rows = append(rows, noteLine(" Delete "+m.fileOpSource+"?", innerW))
		if m.fileOpErr != "" {
			rows = append(rows, noteLine(" "+m.fileOpErr, innerW))
		} else {
			rows = append(rows, noteLine(" enter or y delete · esc cancel", innerW))
		}
	} else {
		rows = append(rows, noteLine(" "+m.fileOpSource, innerW))
		rows = append(rows, m.fileOpInputLine(innerW))
		if m.fileOpErr != "" {
			rows = append(rows, noteLine(" "+m.fileOpErr, innerW))
		} else {
			rows = append(rows, noteLine(" enter "+m.fileOpVerb()+" · esc cancel", innerW))
		}
	}
	boxH := len(rows) + 3
	if boxH > m.height {
		boxH = m.height
	}
	if boxH < 5 {
		boxH = 5
	}
	box := renderPanel(paintTitle(" "+m.fileOpTitle(), innerW, true), strings.Join(rows, "\n"), innerW+2, boxH, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(colorOcean),
	)
}

// fileOpInputLine draws the editable destination path.
func (m Model) fileOpInputLine(width int) string {
	text := truncateCells(" "+m.fileOpInput+"▌", width)
	return lipgloss.NewStyle().
		Foreground(colorLagoonBright).
		Background(colorSelectDim).
		Width(width).
		Render(text)
}
