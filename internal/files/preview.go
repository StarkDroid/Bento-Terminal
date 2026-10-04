package files

import (
	"bytes"
	"errors"
	"io"
	"os"
)

const maxPreview = 256 * 1024

// ErrBinary is returned when a file contains a NUL byte.
var ErrBinary = errors.New("binary file")

// Preview reads a bounded text preview of path.
func Preview(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", errors.New("not a file")
	}
	buf := make([]byte, maxPreview+1)
	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}
	data := buf[:n]
	if bytes.IndexByte(data, 0) >= 0 {
		return "", ErrBinary
	}
	if n > maxPreview {
		return string(data[:maxPreview]) + "\n… truncated …\n", nil
	}
	return string(data), nil
}
