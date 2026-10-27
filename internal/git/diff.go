package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxDiffBytes bounds the size of a synthesised new-file patch.
const maxDiffBytes = 256 * 1024

// Diff returns the patch for one changed path. When both the index and the
// working tree changed it joins the staged and unstaged diffs, and for an
// untracked file it synthesises a new-file patch.
func Diff(root string, f FileStatus) (string, error) {
	if f.Untracked {
		return untrackedDiff(root, f.Path)
	}
	var b strings.Builder
	if f.Staged() {
		out, err := runGit(root, "diff", "--cached", "--", f.Path)
		if err != nil {
			return "", err
		}
		b.WriteString(out)
	}
	if f.Unstaged() {
		out, err := runGit(root, "diff", "--", f.Path)
		if err != nil {
			return "", err
		}
		b.WriteString(out)
	}
	if b.Len() == 0 {
		return "(no changes to show)", nil
	}
	return b.String(), nil
}

func untrackedDiff(root, rel string) (string, error) {
	data, err := readBounded(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return "Binary file " + rel, nil
	}
	text := strings.TrimSuffix(string(data), "\n")
	lines := strings.Split(text, "\n")
	var b strings.Builder
	fmt.Fprintf(&b, "diff --git a/%s b/%s\nnew file mode 100644\n--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n", rel, rel, rel, len(lines))
	for _, line := range lines {
		b.WriteString("+" + line + "\n")
	}
	return b.String(), nil
}

func readBounded(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxDiffBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxDiffBytes {
		data = data[:maxDiffBytes]
	}
	return data, nil
}
