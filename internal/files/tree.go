// Package files loads a project directory tree and text previews.
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxChildren = 2000

var skipNames = map[string]struct{}{
	"node_modules": {},
	"vendor":       {},
	"target":       {},
	"__pycache__":  {},
	"dist":         {},
	".next":        {},
	".venv":        {},
	"venv":         {},
}

// Row is one visible tree line.
type Row struct {
	Name     string
	Rel      string
	IsDir    bool
	Depth    int
	Expanded bool
}

// Tree is a lazy directory listing rooted at a project path.
type Tree struct {
	root string
	top  *node
}

type node struct {
	name     string
	rel      string
	isDir    bool
	expanded bool
	loaded   bool
	children []*node
}

// Load reads the first level of root. Dot entries and heavy tool directories are skipped.
func Load(root string) (*Tree, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	kids, err := readChildren(root, "")
	if err != nil {
		return nil, err
	}
	return &Tree{
		root: root,
		top: &node{
			name:     filepath.Base(root),
			isDir:    true,
			expanded: true,
			loaded:   true,
			children: kids,
		},
	}, nil
}

// Rows returns the expanded view of the tree.
func (t *Tree) Rows() []Row {
	if t == nil || t.top == nil {
		return nil
	}
	var rows []Row
	var walk func(*node, int)
	walk = func(n *node, depth int) {
		for _, c := range n.children {
			rows = append(rows, Row{
				Name:     c.name,
				Rel:      c.rel,
				IsDir:    c.isDir,
				Depth:    depth,
				Expanded: c.expanded,
			})
			if c.isDir && c.expanded {
				walk(c, depth+1)
			}
		}
	}
	walk(t.top, 0)
	return rows
}

// Toggle expands or collapses the directory at index.
func (t *Tree) Toggle(index int) {
	n := t.nodeAt(index)
	if n == nil || !n.isDir {
		return
	}
	if n.expanded {
		n.expanded = false
		return
	}
	t.ensureLoaded(n)
	n.expanded = true
}

// Expand opens the directory at index.
func (t *Tree) Expand(index int) {
	n := t.nodeAt(index)
	if n == nil || !n.isDir || n.expanded {
		return
	}
	t.ensureLoaded(n)
	n.expanded = true
}

// Collapse closes the directory at index.
func (t *Tree) Collapse(index int) {
	n := t.nodeAt(index)
	if n == nil || !n.isDir || !n.expanded {
		return
	}
	n.expanded = false
}

func (t *Tree) ensureLoaded(n *node) {
	if n.loaded {
		return
	}
	kids, err := readChildren(t.root, n.rel)
	if err == nil {
		n.children = kids
	}
	n.loaded = true
}

func (t *Tree) nodeAt(index int) *node {
	if t == nil || index < 0 {
		return nil
	}
	var found *node
	i := 0
	var walk func(*node)
	walk = func(n *node) {
		for _, c := range n.children {
			if found != nil {
				return
			}
			if i == index {
				found = c
				return
			}
			i++
			if c.isDir && c.expanded {
				walk(c)
			}
		}
	}
	walk(t.top)
	return found
}

func readChildren(root, rel string) ([]*node, error) {
	dir := root
	if rel != "" {
		dir = filepath.Join(root, filepath.FromSlash(rel))
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	nodes := make([]*node, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || skip(name) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		childRel := name
		if rel != "" {
			childRel = rel + "/" + name
		}
		nodes = append(nodes, &node{
			name:  name,
			rel:   childRel,
			isDir: e.IsDir(),
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].isDir != nodes[j].isDir {
			return nodes[i].isDir
		}
		return strings.ToLower(nodes[i].name) < strings.ToLower(nodes[j].name)
	})
	if len(nodes) > maxChildren {
		nodes = nodes[:maxChildren]
	}
	return nodes, nil
}

func skip(name string) bool {
	_, ok := skipNames[name]
	return ok
}
