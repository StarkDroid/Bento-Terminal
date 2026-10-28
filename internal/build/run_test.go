package build

import (
	"strings"
	"testing"
	"time"
)

func TestShellCommandWrapsTheScript(t *testing.T) {
	cmd := ShellCommand("echo shell-ok")
	if len(cmd.Argv) != 3 || cmd.Argv[2] != "echo shell-ok" {
		t.Fatalf("argv = %#v", cmd.Argv)
	}
	if cmd.Argv[0] == "" {
		t.Fatal("the shell must be named")
	}
}

func TestShellCommandRunsThroughTheShell(t *testing.T) {
	cancel, events, err := Start(t.TempDir(), ShellCommand("echo shell-ok").Argv)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	lines, done := collect(t, events, 5*time.Second)
	if done != nil {
		t.Fatal(done)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "shell-ok") {
		t.Fatalf("lines = %#v", lines)
	}
}

func collect(t *testing.T, events <-chan Event, timeout time.Duration) ([]string, error) {
	t.Helper()
	deadline := time.After(timeout)
	var lines []string
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				t.Fatal("events closed before done")
			}
			if ev.Done {
				return lines, ev.Err
			}
			lines = append(lines, ev.Line)
		case <-deadline:
			t.Fatal("timed out waiting for process")
		}
	}
}
