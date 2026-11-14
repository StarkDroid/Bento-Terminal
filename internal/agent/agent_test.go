package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestProviderCatalogue(t *testing.T) {
	all := All()
	if len(all) < 6 {
		t.Fatalf("catalogue has %d providers", len(all))
	}
	for _, p := range all {
		if p.ID == "" || p.Name == "" || p.Command == "" || p.Install == "" {
			t.Fatalf("incomplete provider %+v", p)
		}
		if found, ok := Find(p.ID); !ok || found.Command != p.Command {
			t.Fatalf("Find(%q) = %+v %v", p.ID, found, ok)
		}
	}
	if _, ok := Find("nope"); ok {
		t.Fatal("unknown provider should not be found")
	}
}

func TestInstallLinePicksWindowsOverride(t *testing.T) {
	p := Provider{Install: "unix line", InstallWindows: "windows line"}
	if got := p.InstallLine("windows"); got != "windows line" {
		t.Fatalf("windows install %q", got)
	}
	if got := p.InstallLine("darwin"); got != "unix line" {
		t.Fatalf("darwin install %q", got)
	}
	plain := Provider{Install: "unix line"}
	if got := plain.InstallLine("windows"); got != "unix line" {
		t.Fatalf("fallback install %q", got)
	}
}

func TestDetectFindsCommandsOnPath(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"claude", "claude.exe"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
	found := Detect()
	if _, ok := found["claude"]; !ok {
		t.Fatalf("Detect missed claude: %#v", found)
	}
	if _, ok := found["opencode"]; ok {
		t.Fatalf("Detect invented opencode: %#v", found)
	}
}

func TestSessionRunsOnAPseudoTerminal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a Unix shell")
	}
	done := make(chan error, 1)
	session, err := StartSession(t.TempDir(), Provider{ID: "sh", Name: "sh", Command: "sh"}, 40, 10,
		nil,
		func(err error) { done <- err },
	)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := session.Resize(50, 12); err != nil {
		t.Fatalf("resize: %v", err)
	}
	if err := session.Write([]byte("printf 'agent says hi'\r")); err != nil {
		t.Fatal(err)
	}
	if err := session.Write([]byte("exit\r")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("session exit: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("session did not exit")
	}
	if screen := session.Screen().String(); !strings.Contains(screen, "agent says hi") {
		t.Fatalf("screen %q", screen)
	}
}
