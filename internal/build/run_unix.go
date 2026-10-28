//go:build unix

package build

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// shellArgv wraps a terminal line in the user's shell, falling back to sh.
func shellArgv(script string) []string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return []string{shell, "-c", script}
}

func prepareCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
	cmd.Cancel = func() error {
		signalGroup(cmd)
		return nil
	}
}

func signalGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}
