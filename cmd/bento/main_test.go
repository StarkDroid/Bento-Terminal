package main

import "testing"

func TestVersionFromBuildInfo(t *testing.T) {
	cases := map[string]string{
		"v0.8.1":        "0.8.1",
		"0.8.1":         "0.8.1",
		"0.8.1-rc.1":    "",
		"(devel)":       "",
		"devel":         "",
		"":              "",
		"v":             "",
		"not-a-version": "",
	}
	for in, want := range cases {
		if got := versionFromBuildInfo(in); got != want {
			t.Errorf("versionFromBuildInfo(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveVersionKeepsStampedVersion(t *testing.T) {
	if got := resolveVersion("0.8.1"); got != "0.8.1" {
		t.Fatalf("resolveVersion = %q", got)
	}
}
