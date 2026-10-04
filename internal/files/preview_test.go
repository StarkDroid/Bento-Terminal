package files

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewTextAndBinary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	mustWrite(t, path, "alpha\r\nbeta\n")
	got, err := Preview(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "alpha\r\nbeta\n" {
		t.Fatalf("%q", got)
	}

	bin := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(bin, []byte{'h', 0, 'i'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Preview(bin); !errors.Is(err, ErrBinary) {
		t.Fatalf("err = %v", err)
	}
}

func TestPreviewTruncates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	body := strings.Repeat("a", maxPreview+32)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Preview(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "truncated") {
		t.Fatal("expected truncation marker")
	}
	if len(got) > maxPreview+64 {
		t.Fatalf("preview len = %d", len(got))
	}
}
