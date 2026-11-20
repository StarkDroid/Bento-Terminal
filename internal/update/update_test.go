package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.8.1", "0.8.2", -1},
		{"v0.9.0", "0.8.9", 1},
		{"0.8.1", "0.8.1", 0},
		{"0.8", "0.8.0", 0},
		{"0.8.1-rc.1", "0.8.1", -1},
		{"0.8.1+build", "0.8.1", 0},
		{"dev", "0.8.1", 0},
	}
	for _, tc := range cases {
		if got := Compare(tc.a, tc.b); got != tc.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNewer(t *testing.T) {
	if !Newer("0.8.1", "0.8.2") {
		t.Fatal("0.8.2 should be newer than 0.8.1")
	}
	if Newer("0.8.2", "0.8.2") {
		t.Fatal("the same version is not newer")
	}
	if Newer("dev", "0.8.2") {
		t.Fatal("a development build should not prompt")
	}
}

func TestDisplay(t *testing.T) {
	if got := Display("0.8.1"); got != "v0.8.1" {
		t.Fatalf("Display = %q", got)
	}
	if got := Display("dev"); got != "development build" {
		t.Fatalf("Display = %q", got)
	}
}

func TestCheckAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing user agent")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.8.2","html_url":"https://example.test/release","body":"notes"}`))
	}))
	defer server.Close()

	release, newer, err := checkAt(context.Background(), server.Client(), server.URL, "0.8.1")
	if err != nil {
		t.Fatal(err)
	}
	if !newer || release.Version != "0.8.2" || release.URL == "" || release.Notes != "notes" {
		t.Fatalf("release = %+v, newer = %v", release, newer)
	}

	release, newer, err = checkAt(context.Background(), server.Client(), server.URL, "0.8.2")
	if err != nil || newer || release.Version != "0.8.2" {
		t.Fatalf("same version: release = %+v, newer = %v, err = %v", release, newer, err)
	}
}

func TestCheckAtErrors(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"not-a-version"}`))
	}))
	defer bad.Close()
	if _, _, err := checkAt(context.Background(), bad.Client(), bad.URL, "0.8.1"); err == nil {
		t.Fatal("invalid tag should error")
	}

	fail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer fail.Close()
	if _, _, err := checkAt(context.Background(), fail.Client(), fail.URL, "0.8.1"); err == nil {
		t.Fatal("non-200 should error")
	}
}

func TestCheckSkipsDevelopmentBuilds(t *testing.T) {
	release, newer, err := Check(context.Background(), "dev")
	if err != nil || newer || release.Version != "" {
		t.Fatalf("release = %+v, newer = %v, err = %v", release, newer, err)
	}
}

func TestForExecutable(t *testing.T) {
	cases := []struct {
		path string
		name string
		want string
	}{
		{"/opt/homebrew/Cellar/bento/0.8.1/bin/bento", "Homebrew", "brew upgrade"},
		{`C:\Users\a\scoop\apps\bento\current\bento.exe`, "Scoop", "scoop update"},
		{`C:\ProgramData\chocolatey\lib\bento-terminal\tools\bento.exe`, "Chocolatey", "choco upgrade"},
		{"/home/a/.npm/_npx/x/node_modules/bento-terminal/vendor/bento", "npm", "npm install -g"},
	}
	for _, tc := range cases {
		method := ForExecutable(tc.path)
		if method.Name != tc.name || !strings.Contains(method.Command, tc.want) {
			t.Errorf("ForExecutable(%q) = %+v, want %s containing %q", tc.path, method, tc.name, tc.want)
		}
	}

	t.Setenv("GOBIN", t.TempDir())
	goBin := ForExecutable(filepath.Join(os.Getenv("GOBIN"), "bento"))
	if goBin.Name != "Go" || !strings.Contains(goBin.Command, "go install") {
		t.Fatalf("GOBIN method = %+v", goBin)
	}

	fallback := ForExecutable("/tmp/bento")
	if runtime.GOOS == "windows" {
		if fallback.Name != "installer" || !strings.Contains(fallback.Command, "install.ps1") {
			t.Fatalf("fallback = %+v", fallback)
		}
	} else {
		if fallback.Name != "installer" || !strings.Contains(fallback.Command, "install.sh") {
			t.Fatalf("fallback = %+v", fallback)
		}
	}
}
