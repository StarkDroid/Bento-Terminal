package syntax

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Span is a run of source text that shares one colour.
type Span struct {
	Text   string
	Colour string
	Bold   bool
	Italic bool
}

// MaxHighlightLines caps the work done for one file. Longer files stay plain.
const MaxHighlightLines = 8000

// Highlight tokenises lines with the lexer named lang. The result is one slice
// of spans per source line. ok is false when the language is unknown, the file
// is too long, or tokenising failed.
func Highlight(lang string, lines []string) (out [][]Span, ok bool) {
	if lang == "" || len(lines) == 0 || len(lines) > MaxHighlightLines {
		return nil, false
	}
	lexer := lexers.Get(lang)
	if lexer == nil {
		return nil, false
	}
	content := strings.Join(lines, "\n")
	tokens, err := chroma.Tokenise(lexer, nil, content)
	if err != nil {
		return nil, false
	}
	out = [][]Span{nil}
	for _, token := range tokens {
		entry := bentoStyle.Get(token.Type)
		span := Span{
			Colour: entry.Colour.String(),
			Bold:   entry.Bold == chroma.Yes,
			Italic: entry.Italic == chroma.Yes,
		}
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				out = append(out, nil)
			}
			if part == "" {
				continue
			}
			span.Text = part
			out[len(out)-1] = append(out[len(out)-1], span)
		}
	}
	if len(out) > len(lines) {
		out = out[:len(lines)]
	}
	for len(out) < len(lines) {
		out = append(out, nil)
	}
	return out, true
}

// bentoStyle is the lagoon palette expressed as chroma token colours.
var bentoStyle = chroma.MustNewStyle("bento-lagoon", chroma.StyleEntries{
	chroma.Background:          "#F4E4C4",
	chroma.Text:                "#F4E4C4",
	chroma.Error:               "underline #FF6F61",
	chroma.Comment:             "italic #5F7D7B",
	chroma.CommentPreproc:      "#1F8F86",
	chroma.CommentSpecial:      "italic #7DCE82",
	chroma.Keyword:             "bold #3DDECE",
	chroma.KeywordConstant:     "#FF6F61",
	chroma.KeywordDeclaration:  "#3DDECE",
	chroma.KeywordNamespace:    "#3DDECE",
	chroma.KeywordType:         "#5FB3C4",
	chroma.NameAttribute:       "#3DDECE",
	chroma.NameBuiltin:         "#5FB3C4",
	chroma.NameClass:           "bold #7DCE82",
	chroma.NameConstant:        "#FF6F61",
	chroma.NameDecorator:       "#E9C46A",
	chroma.NameFunction:        "#7DCE82",
	chroma.NameLabel:           "#7DCE82",
	chroma.NameNamespace:       "#7DCE82",
	chroma.NameTag:             "#FF6F61",
	chroma.NameVariable:        "#F4E4C4",
	chroma.LiteralString:       "#E9C46A",
	chroma.LiteralStringEscape: "#3DDECE",
	chroma.LiteralNumber:       "#F4A261",
	chroma.Operator:            "#B7A88A",
	chroma.Punctuation:         "#B7A88A",
	chroma.GenericDeleted:      "#FF6F61",
	chroma.GenericEmph:         "italic #F4E4C4",
	chroma.GenericHeading:      "bold #3DDECE",
	chroma.GenericInserted:     "#7DCE82",
	chroma.GenericStrong:       "bold #F4E4C4",
	chroma.GenericSubheading:   "bold #7DCE82",
})
