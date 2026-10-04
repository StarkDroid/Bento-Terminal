package build

import (
	"testing"
	"time"
)

func TestStartEcho(t *testing.T) {
	cancel, events, err := Start(t.TempDir(), []string{"echo", "hello-bento"})
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	lines, done := collect(t, events, 5*time.Second)
	if done != nil {
		t.Fatal(done)
	}
	if len(lines) != 1 || lines[0] != "hello-bento" {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestStartMergesStderrAndExit(t *testing.T) {
	cancel, events, err := Start(t.TempDir(), []string{"sh", "-c", "echo out; echo err >&2; exit 3"})
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	lines, done := collect(t, events, 5*time.Second)
	if done == nil {
		t.Fatal("expected exit error")
	}
	joined := map[string]bool{}
	for _, line := range lines {
		joined[line] = true
	}
	if !joined["out"] || !joined["err"] {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestStartCancel(t *testing.T) {
	cancel, events, err := Start(t.TempDir(), []string{"sleep", "30"})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	cancel()
	_, done := collect(t, events, 5*time.Second)
	if done == nil {
		t.Fatal("expected cancel error")
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
