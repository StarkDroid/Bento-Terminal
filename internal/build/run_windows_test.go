//go:build windows

package build

import (
	"strings"
	"testing"
	"time"
)

func TestStartEcho(t *testing.T) {
	cancel, events, err := Start(t.TempDir(), ShellCommand("echo hello-bento").Argv)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	lines, done := collect(t, events, 5*time.Second)
	if done != nil {
		t.Fatal(done)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "hello-bento") {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestStartMergesStderrAndExit(t *testing.T) {
	cancel, events, err := Start(t.TempDir(), ShellCommand("echo out & echo err 1>&2 & exit /b 3").Argv)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	lines, done := collect(t, events, 5*time.Second)
	if done == nil {
		t.Fatal("expected exit error")
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "out") || !strings.Contains(joined, "err") {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestStartCancel(t *testing.T) {
	cancel, events, err := Start(t.TempDir(), ShellCommand("ping -n 30 127.0.0.1").Argv)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	cancel()
	_, done := collect(t, events, 5*time.Second)
	if done == nil {
		t.Fatal("expected cancel error")
	}
}
