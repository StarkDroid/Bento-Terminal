// Package syntax detects source languages, highlights code, and finds foldable
// function and class blocks.
package syntax

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Language is a detected source language.
type Language struct {
	// Name is the chroma lexer name, for example "Go" or "Python".
	Name string
	// Family groups languages that share a folding strategy: "go", "python",
	// "brace", or "" when the language has no folding support.
	Family string
}

// Detected reports whether a language was found.
func (l Language) Detected() bool { return l.Name != "" }

// ignored lists lexers that only mean "plain text" and are not worth a badge.
var ignored = map[string]struct{}{
	"plaintext":   {},
	"text only":   {},
	"txt":         {},
	"fallback":    {},
	"other":       {},
	"generated":   {},
	"ignore list": {},
}

// Detect picks a lexer from the file name first, then from the content.
func Detect(name, content string) Language {
	var lexer chroma.Lexer
	if name != "" {
		lexer = lexers.Match(filepath.Base(name))
		if lexer != nil && isIgnored(lexer) {
			lexer = nil
		}
	}
	if lexer == nil && strings.TrimSpace(content) != "" {
		if found := lexers.Analyse(content); found != nil && !isIgnored(found) {
			lexer = found
		}
	}
	if lexer == nil {
		lexer = shebangLexer(content)
	}
	if lexer == nil {
		return Language{}
	}
	name = displayName(lexer.Config().Name)
	return Language{Name: name, Family: family(name)}
}

var displayNames = map[string]string{
	"markdown":      "Markdown",
	"javascript":    "JavaScript",
	"typescript":    "TypeScript",
	"json":          "JSON",
	"yaml":          "YAML",
	"toml":          "TOML",
	"html":          "HTML",
	"css":           "CSS",
	"xml":           "XML",
	"sql":           "SQL",
	"php":           "PHP",
	"c++":           "C++",
	"c#":            "C#",
	"objective-c":   "Objective-C",
	"objective-c++": "Objective-C++",
	"makefile":      "Makefile",
	"docker":        "Docker",
	"dockerfile":    "Dockerfile",
	"bash":          "Bash",
	"zsh":           "Zsh",
	"fish":          "Fish",
}

func displayName(raw string) string {
	if pretty, ok := displayNames[strings.ToLower(raw)]; ok {
		return pretty
	}
	if raw == "" {
		return ""
	}
	runes := []rune(raw)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// shebangLexer recognises the common interpreter lines that chroma's content
// analysis misses.
func shebangLexer(content string) chroma.Lexer {
	if !strings.HasPrefix(content, "#!") {
		return nil
	}
	line := content
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	for _, c := range []struct {
		hint string
		lang string
	}{
		{"python", "python"},
		{"node", "javascript"},
		{"deno", "typescript"},
		{"bash", "bash"},
		{"zsh", "zsh"},
		{"ruby", "ruby"},
		{"perl", "perl"},
		{"php", "php"},
		{"lua", "lua"},
	} {
		if strings.Contains(line, c.hint) {
			return lexers.Get(c.lang)
		}
	}
	return nil
}

func isIgnored(lexer chroma.Lexer) bool {
	_, ok := ignored[strings.ToLower(lexer.Config().Name)]
	return ok
}

func family(name string) string {
	switch strings.ToLower(name) {
	case "go":
		return "go"
	case "python", "python 3":
		return "python"
	case "javascript", "typescript", "react", "java", "c", "c++", "c#",
		"rust", "swift", "kotlin", "scala", "dart", "php", "groovy",
		"solidity", "zig", "bash", "shell", "zsh":
		return "brace"
	}
	return ""
}
