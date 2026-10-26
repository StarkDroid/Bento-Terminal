package syntax

import (
	"regexp"
	"strings"
)

// Block is a foldable range of source lines.
type Block struct {
	// Start is the 0-based line that opens the block.
	Start int
	// End is the 0-based last line of the block, inclusive.
	End int
	// Kind is "func" or "type".
	Kind string
}

var (
	goFunc = regexp.MustCompile(`^\s*func\s`)
	goType = regexp.MustCompile(`^\s*type\s+\w+(?:\[[^\]]*\])?\s+(?:struct|interface)\b`)

	pyBlock = regexp.MustCompile(`^\s*(?:async\s+)?(?:def|class)\s+\w`)
	pyClass = regexp.MustCompile(`^\s*(?:async\s+)?class\s+\w`)

	braceFunc = []*regexp.Regexp{
		regexp.MustCompile(`^\s*(?:export\s+)?(?:default\s+)?(?:declare\s+)?(?:async\s+)?function\b`),
		regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?(?:const\s+)?(?:async\s+)?(?:unsafe\s+)?(?:extern\s+"[^"]*"\s+)?fn\s+\w+`),
		regexp.MustCompile(`^\s*(?:public|private|protected|internal|static|final|abstract|synchronized|native|override|open|suspend|inline|operator|virtual|explicit|constexpr|async|extern)\b[^;{}]*\([^;]*$`),
		regexp.MustCompile(`^\s*(?:function\s+)?[\w:.@-]+\s*\(\)\s*\{`),
	}
	braceType = []*regexp.Regexp{
		regexp.MustCompile(`^\s*(?:export\s+)?(?:default\s+)?(?:declare\s+)?(?:abstract\s+|final\s+|sealed\s+|public\s+|private\s+|protected\s+|internal\s+|static\s+|readonly\s+|data\s+|open\s+)*(?:class|interface|enum|struct|record|trait|impl|object|namespace|module|union)\b`),
	}
)

// Outline returns the foldable blocks in lines. It understands Go, Python, and
// the brace languages named in Language.Family.
func Outline(lang Language, lines []string) []Block {
	switch lang.Family {
	case "go":
		return braceBlocks(lines, []*regexp.Regexp{goFunc}, []*regexp.Regexp{goType})
	case "python":
		return indentBlocks(lines)
	case "brace":
		return braceBlocks(lines, braceFunc, braceType)
	}
	return nil
}

// BlockAt returns the innermost block that contains line.
func BlockAt(blocks []Block, line int) (Block, bool) {
	best, found := Block{}, false
	for _, b := range blocks {
		if line < b.Start || line > b.End {
			continue
		}
		if !found || b.Start >= best.Start {
			best, found = b, true
		}
	}
	return best, found
}

func braceBlocks(lines []string, funcs, types []*regexp.Regexp) []Block {
	var blocks []Block
	for i, line := range lines {
		kind := ""
		switch {
		case matchesAny(funcs, line):
			kind = "func"
		case matchesAny(types, line):
			kind = "type"
		default:
			continue
		}
		if end, ok := scanBraces(lines, i); ok && end > i {
			blocks = append(blocks, Block{Start: i, End: end, Kind: kind})
		}
	}
	return blocks
}

func matchesAny(exprs []*regexp.Regexp, line string) bool {
	for _, re := range exprs {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// scanBraces finds the line that closes the brace block opened at start. It
// skips braces inside strings, line comments, and block comments.
func scanBraces(lines []string, start int) (int, bool) {
	depth := 0
	opened := false
	inBlock := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		for j := 0; j < len(line); j++ {
			if inBlock {
				if strings.HasPrefix(line[j:], "*/") {
					inBlock = false
					j++
				}
				continue
			}
			switch {
			case strings.HasPrefix(line[j:], "//"):
				j = len(line)
			case strings.HasPrefix(line[j:], "/*"):
				inBlock = true
				j++
			case line[j] == '"' || line[j] == '\'' || line[j] == '`':
				quote := line[j]
				j++
				for j < len(line) {
					if line[j] == '\\' {
						j++
					} else if line[j] == quote {
						break
					}
					j++
				}
			case line[j] == '{':
				depth++
				opened = true
			case line[j] == '}':
				depth--
				if depth < 0 {
					return 0, false
				}
				if opened && depth == 0 {
					return i, true
				}
			}
		}
	}
	return 0, false
}

func indentBlocks(lines []string) []Block {
	var blocks []Block
	for i, line := range lines {
		if !pyBlock.MatchString(line) {
			continue
		}
		base := indentWidth(line)
		end := len(lines) - 1
		for j := i + 1; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == "" {
				continue
			}
			if indentWidth(lines[j]) <= base {
				end = j - 1
				break
			}
		}
		for end > i && strings.TrimSpace(lines[end]) == "" {
			end--
		}
		if end > i {
			kind := "func"
			if pyClass.MatchString(line) {
				kind = "type"
			}
			blocks = append(blocks, Block{Start: i, End: end, Kind: kind})
		}
	}
	return blocks
}

func indentWidth(line string) int {
	width := 0
	for _, r := range line {
		switch r {
		case ' ':
			width++
		case '\t':
			width += 4
		default:
			return width
		}
	}
	return width
}
