package files

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// maxIndexedFiles bounds the file search index on very large trees.
const maxIndexedFiles = 20000

// List returns every regular file under root as a project-relative slash path.
// It follows the tree's rules: dot entries and heavy tool directories are
// skipped, and symlinks are not followed. The result is sorted by path,
// case-insensitively.
func List(root string) ([]string, error) {
	out := make([]string, 0)
	err := walkFiles(root, "", &out)
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := strings.ToLower(out[i]), strings.ToLower(out[j])
		if a == b {
			return out[i] < out[j]
		}
		return a < b
	})
	return out, nil
}

func walkFiles(root, rel string, out *[]string) error {
	dir := root
	if rel != "" {
		dir = filepath.Join(root, filepath.FromSlash(rel))
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if len(*out) >= maxIndexedFiles {
			return nil
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || skip(name) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		child := name
		if rel != "" {
			child = rel + "/" + name
		}
		if entry.IsDir() {
			if err := walkFiles(root, child, out); err != nil {
				continue
			}
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		*out = append(*out, child)
		if len(*out) >= maxIndexedFiles {
			return nil
		}
	}
	return nil
}
