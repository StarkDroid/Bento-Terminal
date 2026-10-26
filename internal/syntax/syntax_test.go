package syntax

import (
	"strings"
	"testing"
)

func TestDetectByExtension(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"main.go", "Go"},
		{"app.py", "Python"},
		{"package.json", "JSON"},
		{"README.md", "Markdown"},
		{"Makefile", "Makefile"},
		{"notes.txt", ""},
	}
	for _, c := range cases {
		got := Detect(c.name, "")
		if got.Name != c.want {
			t.Fatalf("%s: got %q want %q", c.name, got.Name, c.want)
		}
	}
	if Detect("blob", "").Detected() {
		t.Fatal("unknown name should not detect a language")
	}
}

func TestDetectByContent(t *testing.T) {
	script := "#!/usr/bin/env python3\nprint('hi')\n"
	if got := Detect("script", script); got.Name != "Python" {
		t.Fatalf("got %q", got.Name)
	}
}

func TestFamilies(t *testing.T) {
	cases := map[string]string{"Go": "go", "Python": "python", "TypeScript": "brace", "Rust": "brace", "JSON": ""}
	for name, want := range cases {
		if got := family(name); got != want {
			t.Fatalf("%s: got %q want %q", name, got, want)
		}
	}
}

func TestHighlightKeepsText(t *testing.T) {
	lines := []string{"package main", "", "func main() {", "\tprintln(42)", "}"}
	spans, ok := Highlight("Go", lines)
	if !ok || len(spans) != len(lines) {
		t.Fatalf("ok %v lines %d", ok, len(spans))
	}
	for i, line := range lines {
		var b strings.Builder
		for _, span := range spans[i] {
			b.WriteString(span.Text)
		}
		if b.String() != line {
			t.Fatalf("line %d: %q != %q", i, b.String(), line)
		}
	}
	var colours []string
	for _, span := range spans[2] {
		colours = append(colours, span.Colour)
	}
	joined := strings.Join(colours, " ")
	if !strings.Contains(joined, "#f4e4c4") && !strings.Contains(joined, "#7dce82") && !strings.Contains(joined, "#3ddece") {
		t.Fatalf("no palette colours in %v", colours)
	}
}

func TestHighlightUnknownLanguage(t *testing.T) {
	if _, ok := Highlight("", []string{"x"}); ok {
		t.Fatal("empty language should not highlight")
	}
	if _, ok := Highlight("No Such Language", []string{"x"}); ok {
		t.Fatal("unknown language should not highlight")
	}
}

func TestOutlineGo(t *testing.T) {
	lines := []string{
		"package main",
		"",
		"type server struct {",
		"\tname string",
		"}",
		"",
		"func (s server) run() {",
		"\tif s.name != \"\" {",
		"\t\tprintln(s.name)",
		"\t}",
		"}",
		"",
		"func tail() int { return 1 }",
	}
	blocks := Outline(Language{Name: "Go", Family: "go"}, lines)
	if len(blocks) != 2 {
		t.Fatalf("blocks %+v", blocks)
	}
	if blocks[0] != (Block{Start: 2, End: 4, Kind: "type"}) {
		t.Fatalf("type block %+v", blocks[0])
	}
	if blocks[1] != (Block{Start: 6, End: 10, Kind: "func"}) {
		t.Fatalf("func block %+v", blocks[1])
	}
	if _, ok := BlockAt(blocks, 8); !ok {
		t.Fatal("line 8 should be inside the function")
	}
	if _, ok := BlockAt(blocks, 1); ok {
		t.Fatal("line 1 is outside every block")
	}
}

func TestOutlinePython(t *testing.T) {
	lines := []string{
		"import os",
		"",
		"class Store:",
		"    def get(self, key):",
		"        return key",
		"",
		"def top():",
		"    return 1",
		"# trailing",
	}
	blocks := Outline(Language{Name: "Python", Family: "python"}, lines)
	if len(blocks) != 3 {
		t.Fatalf("blocks %+v", blocks)
	}
	if blocks[0].Start != 2 || blocks[0].End != 4 || blocks[0].Kind != "type" {
		t.Fatalf("class block %+v", blocks[0])
	}
	if blocks[1].Start != 3 || blocks[1].End != 4 || blocks[1].Kind != "func" {
		t.Fatalf("method block %+v", blocks[1])
	}
	if blocks[2].Start != 6 || blocks[2].End != 7 {
		t.Fatalf("def block %+v", blocks[2])
	}
	if inner, ok := BlockAt(blocks, 4); !ok || inner.Start != 3 {
		t.Fatalf("innermost at line 4: %+v %v", inner, ok)
	}
}

func TestOutlineBraceStrings(t *testing.T) {
	lines := []string{
		"export function greet(name) {",
		"  const s = \"}\";",
		"  // } ignored",
		"  return s + name;",
		"}",
	}
	blocks := Outline(Language{Name: "JavaScript", Family: "brace"}, lines)
	if len(blocks) != 1 || blocks[0].End != 4 {
		t.Fatalf("blocks %+v", blocks)
	}
}

func TestOutlineUnknownFamily(t *testing.T) {
	if blocks := Outline(Language{Name: "JSON"}, []string{"{"}); len(blocks) != 0 {
		t.Fatalf("blocks %+v", blocks)
	}
}
