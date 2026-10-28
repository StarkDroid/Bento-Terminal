//go:build windows

package build

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// shellArgv wraps a terminal line in the command interpreter: COMSPEC, or
// cmd.exe when the environment does not name one.
func shellArgv(script string) []string {
	shell := os.Getenv("COMSPEC")
	if shell == "" {
		shell = "cmd.exe"
	}
	return []string{shell, "/C", script}
}

func prepareCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	cmd.WaitDelay = 2 * time.Second
	cmd.Cancel = func() error {
		signalGroup(cmd)
		return nil
	}
}

// signalGroup ends the process and its children. Windows has no signal for a
// process group, so taskkill walks the tree; killing the root is the fallback.
func signalGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := kill.Run(); err != nil {
		_ = cmd.Process.Kill()
	}
}
