package agent

import (
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/hinshun/vt10x"
)

// Session is one running coding agent attached to a pseudo-terminal. Its
// output is parsed into a terminal emulator the UI can draw cell by cell.
type Session struct {
	provider Provider
	proc     termProcess
	emu      vt10x.Terminal
	onExit   func(error)

	closeOnce sync.Once
	detached  atomic.Bool
}

// processSpec is one agent launch, handed to the platform-specific starter.
type processSpec struct {
	dir     string
	command string
	env     []string
	cols    int
	rows    int
}

// termProcess is a child process attached to a pseudo-terminal.
type termProcess interface {
	io.ReadWriteCloser
	// Resize tells the process how many cells the terminal has.
	Resize(cols, rows int) error
	// Wait blocks until the process exits and reports a non-zero status as
	// an error.
	Wait() error
}

// StartSession launches the provider's command in dir on a pseudo-terminal of
// the given cell size. onOutput is called (from a background goroutine) after
// each read, and onExit once when the process ends. Both may fire
// concurrently, so callers must make them safe.
func StartSession(dir string, p Provider, width, height int, onOutput func(), onExit func(error)) (*Session, error) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	proc, err := startProcess(processSpec{
		dir:     dir,
		command: p.Command,
		env:     sessionEnv(),
		cols:    width,
		rows:    height,
	})
	if err != nil {
		return nil, err
	}
	s := &Session{
		provider: p,
		proc:     proc,
		emu:      vt10x.New(vt10x.WithSize(width, height), vt10x.WithWriter(io.Discard)),
		onExit:   onExit,
	}
	go s.read(onOutput)
	return s, nil
}

// sessionEnv tells the agent a colour terminal is attached even though the
// panel, not the user's terminal, draws the screen.
func sessionEnv() []string {
	return append(os.Environ(), "TERM=xterm-256color", "COLORTERM=truecolor")
}

// read copies the agent's output into the emulator until the terminal closes.
// vt10x returns fewer bytes than it was given when the buffer ends in half a
// rune, so the remainder is carried into the next read.
func (s *Session) read(onOutput func()) {
	buf := make([]byte, 32*1024)
	var pending []byte
	for {
		n, err := s.proc.Read(buf)
		if n > 0 {
			pending = append(pending, buf[:n]...)
			written, _ := s.emu.Write(pending)
			if written < 0 {
				written = 0
			}
			pending = append(pending[:0], pending[written:]...)
			if onOutput != nil {
				onOutput()
			}
		}
		if err != nil {
			break
		}
	}
	werr := s.proc.Wait()
	s.closeProc()
	if s.detached.Load() {
		werr = nil
	}
	if s.onExit != nil {
		s.onExit(werr)
	}
}

// closeProc releases the process and its terminal once.
func (s *Session) closeProc() {
	s.closeOnce.Do(func() {
		_ = s.proc.Close()
	})
}

// Write sends input bytes to the agent.
func (s *Session) Write(p []byte) error {
	if s == nil || s.proc == nil {
		return nil
	}
	_, err := s.proc.Write(p)
	return err
}

// Resize changes the pseudo-terminal and the emulator together.
func (s *Session) Resize(width, height int) error {
	if s == nil || s.proc == nil {
		return nil
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	s.emu.Resize(width, height)
	return s.proc.Resize(width, height)
}

// Screen returns the emulator the UI draws. Callers lock it with Lock and
// Unlock while reading cells.
func (s *Session) Screen() vt10x.Terminal {
	if s == nil {
		return nil
	}
	return s.emu
}

// Provider returns the agent this session runs.
func (s *Session) Provider() Provider {
	if s == nil {
		return Provider{}
	}
	return s.provider
}

// Close ends the process and its terminal. It is safe to call more than once
// and from the read goroutine.
func (s *Session) Close() {
	if s == nil {
		return
	}
	s.detached.Store(true)
	s.closeProc()
}
