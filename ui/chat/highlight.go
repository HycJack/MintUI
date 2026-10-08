package chat

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/theme"
)

// Highlight splits source into the tokens a code block is coloured by.
//
// Like Parse it is pure — a Context would give it nothing to say — and like
// Parse its value is in what it refuses to guess. A language it has no table
// for comes back as one undifferentiated token, because a highlighter that
// guessed would colour Go as C and put a shell script's # comments inside its
// strings, which is worse than not colouring anything at all. Colouring what
// it knows and leaving the rest as text is the only reading a reader cannot be
// misled by.
func Highlight(src, lang string) []Token {
	if src == "" {
		return nil
	}
	spec, ok := specs[strings.ToLower(strings.TrimSpace(lang))]
	if !ok {
		return []Token{{Kind: TokText, Text: src}}
	}

	var out []Token
	emit := func(kind TokenKind, text string) {
		if text == "" {
			return
		}
		// Runs of the same kind and touching in the source are one token:
		// "return value" is two words but not three tokens, and a viewer that
		// draws one colour block per token should not have two blocks where
		// the reader sees a space.
		if n := len(out); n > 0 && out[n-1].Kind == kind && touching(out[n-1].Text, text) {
			out[n-1].Text += text
			return
		}
		out = append(out, Token{Kind: kind, Text: text})
	}

	for i := 0; i < len(src); {
		rest := src[i:]
		switch {
		case spec.lineComment != "" && strings.HasPrefix(rest, spec.lineComment):
			end := lineEnd(src, i)
			emit(TokComment, src[i:end])
			i = end

		case spec.block[0] != "" && strings.HasPrefix(rest, spec.block[0]):
			end := strings.Index(src[i+len(spec.block[0]):], spec.block[1])
			if end < 0 {
				// Unterminated: run to the end, which is what the reader is
				// looking at anyway. A token that stops at the end of the line
				// would colour the rest of the file as code.
				emit(TokComment, src[i:])
				i = len(src)
			} else {
				end = i + len(spec.block[0]) + end + len(spec.block[1])
				emit(TokComment, src[i:end])
				i = end
			}

		case spec.quotes != "" && strings.IndexByte(spec.quotes, rest[0]) >= 0:
			text, next := readString(src, i, rest[0], spec.raw)
			emit(TokString, text)
			i = next

		case rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\n' || rest[0] == '\r':
			end := i
			for end < len(src) && isSpaceByte(src[end]) {
				end++
			}
			emit(TokText, src[i:end])
			i = end

		case isDigit(rest[0]) || (rest[0] == '.' && i+1 < len(src) && isDigit(src[i+1]) &&
			// A method call's dot is a dot: "1 .bit_length()" must not read the
			// dot and the following digits as one number.
			(i == 0 || !isIdentByte(src[i-1]))):
			end := readNumber(src, i)
			emit(TokNumber, src[i:end])
			i = end

		case isIdentStart(rest[0]):
			end := i
			for end < len(src) && isIdentByte(src[end]) {
				end++
			}
			word := src[i:end]
			switch {
			case spec.keywords[word]:
				emit(TokKeyword, word)
			case spec.types[word]:
				emit(TokType, word)
			default:
				emit(TokIdent, word)
			}
			i = end

		default:
			emit(TokPunct, src[i:i+1])
			i++
		}
	}
	return out
}

// TokenKind is what a token is, for colouring and for the tests that say which
// tokens an input cut into.
type TokenKind int

const (
	// TokText is everything that carries no meaning of its own: the spaces
	// between the words, and the whole of a language with no table here.
	TokText TokenKind = iota
	// TokKeyword is a reserved word: if, return, select.
	TokKeyword
	// TokType is a builtin type name. It is separate from TokKeyword because
	// colour is most of what a highlighter does, and int in a line of `if` is
	// a different colour for a reason.
	TokType
	// TokString is a quoted literal, escapes and all.
	TokString
	// TokNumber is any numeric literal, in any base.
	TokNumber
	// TokComment is a line or block comment.
	TokComment
	// TokIdent is any other name.
	TokIdent
	// TokPunct is one character of punctuation.
	TokPunct
)

func (k TokenKind) String() string {
	switch k {
	case TokKeyword:
		return "keyword"
	case TokType:
		return "type"
	case TokString:
		return "string"
	case TokNumber:
		return "number"
	case TokComment:
		return "comment"
	case TokIdent:
		return "ident"
	case TokPunct:
		return "punct"
	}
	return "text"
}

// Token is one run of source sharing one kind.
//
// Concatenating every token's Text gives the source back exactly, spaces and
// all. That is the property the renderer leans on — it draws each token in its
// own colour and the gaps are simply the tokens in between — and the property
// the tests check, because a highlighter that dropped the spaces between the
// words of a sentence would look correct in a list of kinds and be a word
// salad on screen.
type Token struct {
	Kind TokenKind
	Text string
}

// String is the token's text, so a token can be printed as its own value.
func (t Token) String() string { return t.Text }

// ── the scanner ────────────────────────────────────────────────────────────

// readString reads a quoted literal, returning its text and where it ends. raw
// says the quote does not end at a newline, which is what a Go backtick and a
// JavaScript template literal do and what every other quote does not.
func readString(src string, i int, quote byte, raw bool) (string, int) {
	j := i + 1
	for j < len(src) {
		if src[j] == '\\' && j+1 < len(src) {
			j += 2
			continue
		}
		if src[j] == quote {
			return src[i : j+1], j + 1
		}
		if src[j] == '\n' && !raw {
			// An unterminated one ends at the line rather than eating the
			// rest of the file: a model that streamed half a string should
			// colour half a string, and the code after the newline is code.
			return src[i:j], j
		}
		j++
	}
	return src[i:], len(src)
}

// readNumber reads a numeric literal of any base: hex, binary, octal, decimal
// with a point and an exponent, and the digit separators a modern language
// allows so 1_000_000 is one token and not a run of digits around underscores.
func readNumber(src string, i int) int {
	j := i
	if src[j] == '0' && j+1 < len(src) {
		switch src[j+1] {
		case 'x', 'X', 'b', 'B', 'o', 'O':
			j += 2
			for j < len(src) && (isHexDigit(src[j]) || src[j] == '_') {
				j++
			}
			return j
		}
	}
	for j < len(src) && (isDigit(src[j]) || src[j] == '_') {
		j++
	}
	if j < len(src) && src[j] == '.' && j+1 < len(src) && isDigit(src[j+1]) {
		j++
		for j < len(src) && (isDigit(src[j]) || src[j] == '_') {
			j++
		}
	}
	if j < len(src) && (src[j] == 'e' || src[j] == 'E') {
		k := j + 1
		if k < len(src) && (src[k] == '+' || src[k] == '-') {
			k++
		}
		if k < len(src) && isDigit(src[k]) {
			for k < len(src) && (isDigit(src[k]) || src[k] == '_') {
				k++
			}
			return k
		}
	}
	return j
}

// touching reports whether two texts end up next to one another in the
// source, which is what lets two runs of the same kind merge without eating
// the text between them.
func touching(a, b string) bool {
	if a == "" || b == "" {
		return true
	}
	last, first := a[len(a)-1], b[0]
	return !isSpaceByte(last) && !isSpaceByte(first)
}

func lineEnd(src string, i int) int {
	if end := strings.IndexByte(src[i:], '\n'); end >= 0 {
		return i + end
	}
	return len(src)
}

func isSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

func isHexDigit(b byte) bool {
	return isDigit(b) || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}

func isIdentStart(b byte) bool {
	return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= 0x80
}

func isIdentByte(b byte) bool { return isIdentStart(b) || isDigit(b) }

// ── the language tables ────────────────────────────────────────────────────

// langSpec is what one language's highlighting needs to know. Everything about
// a language that the scanner cannot read off the source itself — which words
// are reserved, which quote characters mean a string — is here, so adding a
// language is writing a row and not touching the scanner.
type langSpec struct {
	keywords    map[string]bool
	types       map[string]bool
	lineComment string
	block       [2]string
	// quotes are the characters that open a string.
	quotes string
	// raw says a quote runs past the end of its line, as a Go backtick does.
	raw bool
}

// words builds a lookup set from a space-separated list.
func words(s string) map[string]bool {
	m := make(map[string]bool)
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

// cLike is what every curly-brace language shares. It is a function rather
// than a var because Go initialises vars in file order and the maps below are
// built by calling it: a var built by a func declared in another file has no
// defined order to run in.
func cLike(keywords, types, line string, block [2]string, quotes string) langSpec {
	return langSpec{
		keywords: words(keywords), types: words(types),
		lineComment: line, block: block, quotes: quotes,
	}
}

var specs = map[string]langSpec{
	"go": cLike(
		"break case chan const continue default defer else fallthrough for func go goto if "+
			"import interface map package range return select struct switch type var",
		"any bool byte complex64 complex128 comparable error float32 float64 int int8 int16 "+
			"int32 int64 rune string uint uint8 uint16 uint32 uint64 uintptr",
		"//", [2]string{"/*", "*/"}, "\"'`"),

	"js": cLike(
		"async await break case catch class const continue debugger default delete do else "+
			"export extends finally for function if import in instanceof let new return static "+
			"super switch this throw try typeof var void while with yield",
		"Array Boolean Date Error JSON Map Number Object Promise RegExp Set String Symbol bigint",
		"//", [2]string{"/*", "*/"}, "\"'`"),

	"ts": cLike(
		"abstract any as async await boolean break case catch class const continue declare "+
			"default delete do else enum export extends finally for from function if implements "+
			"import in instanceof interface keyof let namespace never new number of private "+
			"protected public readonly return static string super switch symbol this throw try "+
			"type typeof undefined unknown var void while yield",
		"Array Object Promise Record Partial Readonly Set Map string number boolean bigint "+
			"unknown never",
		"//", [2]string{"/*", "*/"}, "\"'`"),

	"python": {
		keywords: words("and as assert async await break class continue def del elif else " +
			"except finally for from global if import in is lambda match nonlocal not or pass " +
			"raise return try while with yield"),
		types: words("None True False Self bool bytes complex dict float frozenset int list " +
			"object set str tuple"),
		lineComment: "#",
		quotes:      `"'`,
	},

	"rust": cLike(
		"as async await break const continue crate dyn else enum extern fn for if impl in let "+
			"loop match mod move mut pub ref return self static struct super trait type unsafe "+
			"use where while",
		"bool char f32 f64 i8 i16 i32 i64 i128 isize str u8 u16 u32 u64 u128 usize",
		"//", [2]string{"/*", "*/"}, "\""),

	"c": cLike(
		"auto break case const continue default do else enum extern for goto if inline register "+
			"restrict return sizeof static struct switch typedef union volatile while",
		"bool char double float int long short signed size_t ssize_t unsigned void uint8_t "+
			"uint16_t uint32_t uint64_t int8_t int16_t int32_t int64_t",
		"//", [2]string{"/*", "*/"}, "\"'"),

	"java": cLike(
		"abstract assert break case catch class const continue default do else enum extends "+
			"final finally for goto if implements import instanceof interface native new package "+
			"private protected public return static strictfp super switch synchronized this throw "+
			"throws transient try var void volatile while record sealed permits yield",
		"boolean byte char double float int long short String Integer Double Float Long Boolean "+
			"List Map Set Object",
		"//", [2]string{"/*", "*/"}, "\"'"),

	// JSON has three reserved words and no comments, which is exactly why it
	// is its own row rather than JavaScript with a flag: a JSON block that
	// highlighted // as a comment would accept a document the parser rejects.
	"json": {
		keywords: words("true false null"),
		quotes:   `"`,
	},

	"sql": cLike(
		"select from where insert into values update set delete create table drop alter add "+
			"column join inner left right outer full on as and or not null is in between like "+
			"order by group having limit offset distinct union all case when then else end "+
			"primary key foreign references default unique index view with",
		"int integer bigint smallint decimal numeric float real double char varchar text date "+
			"time timestamp boolean serial",
		"--", [2]string{"/*", "*/"}, `'"`),

	"sh": {
		keywords: words("if then else elif fi for while until do done case esac in function " +
			"return export local readonly declare source alias unset trap shift exit set"),
		lineComment: "#",
		quotes:      `"'`,
	},
}

// TokenText joins every token's text, which gives the source back. It is
// exported because a caller that wants to hand a highlighted block on — to
// copy it, to search it, to hash it — needs the whole string and not a list of
// coloured pieces.
func TokenText(tokens []Token) string {
	var b strings.Builder
	for _, t := range tokens {
		b.WriteString(t.Text)
	}
	return b.String()
}

// TokenColor is the colour a token is drawn in, out of the window's palette.
//
// Nothing here is a raw hex: the palette is what carries light and dark, and a
// highlighter with colours of its own would be the one part of a dark chat
// window that did not follow the desktop. A kind with no colour of its own
// falls back to body text rather than to the accent, which is kept for things
// that do something rather than things that are something.
//
// It is a function of two plain values rather than a method, so that a
// caller can colour a token without a window — which is what a test of the
// highlighter's output needs, and what makes the palette the only place a
// chat block's colour comes from.
func TokenColor(kind TokenKind, dark bool) (fg, bg ui.Color) {
	k := theme.Light()
	if dark {
		k = theme.Dark()
	}
	fg, bg = k.Text, ui.Transparent
	switch kind {
	case TokKeyword:
		fg = k.AccentText
	case TokType:
		fg = k.Success
	case TokString:
		fg = k.Warning
	case TokNumber:
		fg = k.Accent
	case TokComment:
		fg = k.TextFaint
	case TokPunct, TokIdent, TokText:
		fg = k.Text
	}
	return fg, bg
}
