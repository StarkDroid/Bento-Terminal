// Package state remembers the projects Bento has opened and where the user
// left off in each one.
package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// maxRecent bounds the recent-project list.
	maxRecent = 20
	// fileName is the file under the platform config directory.
	fileName = "state.json"
)

// Project is one entry in the recent-project list.
type Project struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	LastOpened time.Time `json:"last_opened"`
}

// ProjectState is what Bento restores when a project is opened again.
type ProjectState struct {
	File   string `json:"file,omitempty"`
	Line   int    `json:"line,omitempty"`
	Col    int    `json:"col,omitempty"`
	Scroll int    `json:"scroll,omitempty"`
	Focus  string `json:"focus,omitempty"`
}

type disk struct {
	Recent   []Project               `json:"recent"`
	Projects map[string]ProjectState `json:"projects"`
}

// Store reads and writes the recent-project state. A nil store and a store
// opened with NewMemory both behave as no-ops on Save.
type Store struct {
	mu   sync.Mutex
	path string
	data disk
}

// Open reads a state file. A missing file is an empty store; malformed JSON is
// an error so the caller can decide whether to overwrite it.
func Open(path string) (*Store, error) {
	s := &Store{path: path}
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		s.data = disk{Projects: map[string]ProjectState{}}
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &s.data); err != nil {
		return nil, err
	}
	if s.data.Projects == nil {
		s.data.Projects = map[string]ProjectState{}
	}
	return s, nil
}

// OpenDefault opens the state file under the platform config directory. A
// malformed file is moved aside so the next start begins with a clean list.
func OpenDefault() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "bento", fileName)
	store, err := Open(path)
	if err == nil {
		return store, nil
	}
	if renameErr := os.Rename(path, path+".bak"); renameErr != nil {
		return nil, err
	}
	return Open(path)
}

// NewMemory returns a store that never writes to disk. Tests use it so they
// never touch the user's real recent-project list.
func NewMemory() *Store {
	return &Store{data: disk{Projects: map[string]ProjectState{}}}
}

// Recent returns the recent projects in most-recently-opened order.
func (s *Store) Recent() []Project {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Project, len(s.data.Recent))
	copy(out, s.data.Recent)
	return out
}

// Touch moves root to the front of the recent list and saves.
func (s *Store) Touch(root string) error {
	if s == nil || root == "" {
		return nil
	}
	root = filepath.Clean(root)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Projects == nil {
		s.data.Projects = map[string]ProjectState{}
	}
	entry := Project{Path: root, Name: filepath.Base(root), LastOpened: time.Now().UTC()}
	out := make([]Project, 0, len(s.data.Recent)+1)
	out = append(out, entry)
	for _, p := range s.data.Recent {
		if filepath.Clean(p.Path) != root {
			out = append(out, p)
		}
	}
	if len(out) > maxRecent {
		out = out[:maxRecent]
	}
	s.data.Recent = out
	return s.saveLocked()
}

// Project returns the saved state for root.
func (s *Store) Project(root string) ProjectState {
	if s == nil {
		return ProjectState{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.Projects[filepath.Clean(root)]
}

// SetProject replaces the saved state for root and saves.
func (s *Store) SetProject(root string, ps ProjectState) error {
	if s == nil || root == "" {
		return nil
	}
	root = filepath.Clean(root)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Projects == nil {
		s.data.Projects = map[string]ProjectState{}
	}
	ps.File = filepath.ToSlash(ps.File)
	s.data.Projects[root] = ps
	return s.saveLocked()
}

// Save flushes the store. It is a no-op for a nil or in-memory store.
func (s *Store) Save() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	if s.path == "" {
		return nil
	}
	body, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".state-*.json")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, s.path)
}

// Path returns the state file path, or an empty string for memory stores.
func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// ShortPath replaces the home directory with "~" for display.
func ShortPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(path, prefix) {
		return "~" + string(filepath.Separator) + strings.TrimPrefix(path, prefix)
	}
	return path
}
