//go:build windows

package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/UserExistsError/conpty"
)

// startProcess runs the command on a Windows pseudo-console. npm installs
// agents as .cmd shims, which ConPTY cannot start directly, so those go
// through the command interpreter.
func startProcess(spec processSpec) (termProcess, error) {
	path, err := exec.LookPath(spec.command)
	if err != nil {
		return nil, err
	}
	cpty, err := conpty.Start(commandLine(path),
		conpty.ConPtyDimensions(spec.cols, spec.rows),
		conpty.ConPtyWorkDir(spec.dir),
		conpty.ConPtyEnv(spec.env),
	)
	if err != nil {
		return nil, err
	}
	return &windowsProcess{cpty: cpty}, nil
}

// commandLine quotes the resolved command for CreateProcess, wrapping a
// batch file in the command interpreter first.
func commandLine(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".cmd", ".bat":
		shell := os.Getenv("COMSPEC")
		if shell == "" {
			shell = "cmd.exe"
		}
		return syscall.EscapeArg(shell) + " /C " + syscall.EscapeArg(path)
	}
	return syscall.EscapeArg(path)
}

type windowsProcess struct {
	cpty *conpty.ConPty
}

func (p *windowsProcess) Read(b []byte) (int, error)  { return p.cpty.Read(b) }
func (p *windowsProcess) Write(b []byte) (int, error) { return p.cpty.Write(b) }

func (p *windowsProcess) Resize(cols, rows int) error {
	return p.cpty.Resize(cols, rows)
}

func (p *windowsProcess) Wait() error {
	code, err := p.cpty.Wait(context.Background())
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("exit status %d", code)
	}
	return nil
}

func (p *windowsProcess) Close() error {
	if pid := p.cpty.Pid(); pid > 0 {
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Kill()
			_ = proc.Release()
		}
	}
	return p.cpty.Close()
}
