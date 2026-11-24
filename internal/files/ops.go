package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Rename moves a file or directory to a new project-relative path.
func Rename(root, from, to string) (string, error) {
	return transfer(root, from, to, false)
}

// Move moves a file or directory to a new project-relative path. It is the
// same filesystem operation as Rename; the two exist so the interface can use
// the word that fits the prompt.
func Move(root, from, to string) (string, error) {
	return transfer(root, from, to, false)
}

// Copy duplicates a file or directory at a new project-relative path. Directories
// are copied recursively.
func Copy(root, from, to string) (string, error) {
	return transfer(root, from, to, true)
}

// Remove deletes a file or directory at a project-relative path.
func Remove(root, rel string) (string, error) {
	abs, clean, err := resolve(root, rel)
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(abs); err != nil {
		return "", err
	}
	return clean, nil
}

// transfer is the shared body of Rename, Move, and Copy. It refuses to
// overwrite an existing destination and to copy or move a directory into
// itself.
func transfer(root, from, to string, duplicate bool) (string, error) {
	srcAbs, srcRel, err := resolve(root, from)
	if err != nil {
		return "", err
	}
	dstAbs, dstRel, err := resolve(root, to)
	if err != nil {
		return "", err
	}
	if srcRel == dstRel {
		return "", fmt.Errorf("%s is already at that path", srcRel)
	}
	info, err := os.Lstat(srcAbs)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(dstAbs); err == nil {
		return "", fmt.Errorf("%s: %w", dstRel, ErrExists)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if info.IsDir() && within(dstAbs, srcAbs) {
		return "", fmt.Errorf("%s is inside %s", dstRel, srcRel)
	}
	if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
		return "", err
	}
	if duplicate {
		if err := copyPath(srcAbs, dstAbs, info); err != nil {
			return "", err
		}
	} else if err := os.Rename(srcAbs, dstAbs); err != nil {
		return "", err
	}
	return dstRel, nil
}

// copyPath writes src to dst. Directories are recreated and walked; symlinks
// are recreated with the same target rather than followed.
func copyPath(src, dst string, info os.FileInfo) error {
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	case info.IsDir():
		if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			child, err := entry.Info()
			if err != nil {
				return err
			}
			if err := copyPath(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()), child); err != nil {
				return err
			}
		}
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// within reports whether path is dir itself or sits under it.
func within(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
