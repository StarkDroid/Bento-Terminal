//go:build !windows

package agent

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

// startProcess runs the command on a Unix pseudo-terminal.
func startProcess(spec processSpec) (termProcess, error) {
	cmd := exec.Command(spec.command)
	cmd.Dir = spec.dir
	cmd.Env = spec.env
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(spec.cols),
		Rows: uint16(spec.rows),
	})
	if err != nil {
		return nil, err
	}
	return &unixProcess{ptmx: ptmx, cmd: cmd}, nil
}

type unixProcess struct {
	ptmx *os.File
	cmd  *exec.Cmd
}

func (p *unixProcess) Read(b []byte) (int, error)  { return p.ptmx.Read(b) }
func (p *unixProcess) Write(b []byte) (int, error) { return p.ptmx.Write(b) }

func (p *unixProcess) Resize(cols, rows int) error {
	return pty.Setsize(p.ptmx, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})
}

func (p *unixProcess) Wait() error {
	return p.cmd.Wait()
}

// Close ends the agent's whole process group. The pseudo-terminal starts the
// child as a session leader, so the group id is the child's pid.
func (p *unixProcess) Close() error {
	if p.cmd.Process != nil {
		if pgid, err := syscall.Getpgid(p.cmd.Process.Pid); err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			_ = p.cmd.Process.Kill()
		}
	}
	return p.ptmx.Close()
}
