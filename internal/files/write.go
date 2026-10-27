package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrExists is returned when Create is asked to make a path that is already there.
var ErrExists = errors.New("file already exists")

// Create makes an empty file at a project-relative path and returns the cleaned
// slash-separated path. Missing parent directories are created.
func Create(root, rel string) (string, error) {
	abs, clean, err := resolve(root, rel)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err == nil {
		return "", fmt.Errorf("%s: %w", clean, ErrExists)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return clean, nil
}

// Save writes content to a project-relative path, creating parent directories.
func Save(root, rel, content string) (string, error) {
	abs, clean, err := resolve(root, rel)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return "", err
	}
	return clean, nil
}

// resolve turns a project-relative path into an absolute path and rejects
// absolute paths or paths that climb out of root.
func resolve(root, rel string) (abs, clean string, err error) {
	rel = strings.TrimSpace(rel)
	rel = strings.TrimPrefix(rel, "./")
	if rel == "" {
		return "", "", errors.New("empty path")
	}
	slashed := filepath.ToSlash(rel)
	if strings.HasPrefix(slashed, "/") || slashed == ".." || strings.HasPrefix(slashed, "../") {
		return "", "", fmt.Errorf("%s is outside the project", rel)
	}
	clean = filepath.ToSlash(filepath.Clean(filepath.FromSlash(slashed)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", "", fmt.Errorf("%s is outside the project", rel)
	}
	return filepath.Join(root, filepath.FromSlash(clean)), clean, nil
}
