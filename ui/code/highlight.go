package code

import (
	"sort"
	"strings"
)

// Highlight is a token-level highlighter: it splits each line into runs of
// like meaning and nothing more.
//
// It is not a parser, and that is a decision rather than a shortfall. A
// parser is a dependency, a tree, and a cost that grows with the file rather
// than with the line, and everything a reader uses highlighting for — telling
// a string from an identifier, a comment from code, a number from a version —
// is decided by the line in front of them. A half-written line is the common
// case in an editor and the one a parser refuses most often, so a component
// that has to draw a file as it is being typed into gets a scanner.
//
// The one thing carried across lines is whether a block comment is open,
// because `/*` on one line and the words on the next are one comment to a
// reader and cannot be highlighted one line at a time without it.

// TokenKind is what a run of text is.
type TokenKind int

const (
	// TokenPlain is code that is none of the others: identifiers,
	// punctuation, operators, whitespace.
	TokenPlain TokenKind = iota
	// Keyword is a word the language reserves.
	Keyword
	// Type is a name that looks like a type — capitalised in the languages
	// whose types are capitalised, and nothing else. It is a guess, and it is
	// labelled as one: the alternative is either no distinction at all or a
	// parser.
	Type
	// String is a string, a character, or a template's text.
	String
	// Number is a numeric literal, in any of the bases.
	Number
	// Comment is a line comment or a block comment.
	Comment
	// Function is a name followed by an opening parenthesis.
	Function
	// Punct is punctuation and operators, kept apart from Plain so that a
	// reader can tell structure from words without the two being one colour.
	Punct
)

func (k TokenKind) String() string {
	switch k {
	case Keyword:
		return "Keyword"
	case Type:
		return "Type"
	case String:
		return "String"
	case Number:
		return "Number"
	case Comment:
		return "Comment"
	case Function:
		return "Function"
	case Punct:
		return "Punct"
	}
	return "TokenPlain"
}

// Lang is which language's words the scanner knows.
type Lang int

const (
	// LangPlain is no language: identifiers and punctuation only, which is
	// what a viewer shows for a file it has never heard of, and what a log
	// shows.
	LangPlain Lang = iota
	// Go is the language this repository is written in.
	Go
	// JavaScript covers TypeScript's keywords too, since the two differ in
	// types and not in words.
	JavaScript
	// Python has no braces and no line comments that start with anything but
	// a hash, which the scanner handles by having its own comment sigils.
	Python
	// Rust adds a lifetime sigil and a raw string, both of which are string
	// starts a naive scanner gets wrong.
	Rust
	// SQL has a line comment of its own and case-insensitive words.
	SQL
	// Shell has a comment of its own and a quote of its own.
	Shell
)

func (l Lang) String() string {
	switch l {
	case Go:
		return "Go"
	case JavaScript:
		return "JavaScript"
	case Python:
		return "Python"
	case Rust:
		return "Rust"
	case SQL:
		return "SQL"
	case Shell:
		return "Shell"
	}
	return "LangPlain"
}

// Token is a run of a line of one kind.
type Token struct {
	// Text is the run itself, never empty: an empty token is a thing a
	// splitter produces by accident and every consumer then has to skip.
	Text string
	// Kind is what it is.
	Kind TokenKind
}

// The languages' reserved words, as sets.
//
// They are sets rather than maps because the test that matters is "is this
// word reserved", and a map lookup answers that; the slices are sorted so the
// binary search in isKeyword is right, which is why there is a test that
// checks they are.
var keywordSets = map[Lang][]string{
	Go: {
		"break",
		"case",
		"chan",
		"const",
		"continue",
		"default",
		"defer",
		"else",
		"fallthrough",
		"for",
		"func",
		"go",
		"goto",
		"if",
		"import",
		"interface",
		"map",
		"package",
		"range",
		"return",
		"select",
		"struct",
		"switch",
		"type",
		"var",
	},
	JavaScript: {
		"async",
		"await",
		"break",
		"case",
		"catch",
		"class",
		"const",
		"continue",
		"debugger",
		"default",
		"delete",
		"do",
		"else",
		"export",
		"extends",
		"finally",
		"for",
		"from",
		"function",
		"get",
		"if",
		"import",
		"in",
		"instanceof",
		"let",
		"new",
		"of",
		"return",
		"set",
		"static",
		"super",
		"switch",
		"this",
		"throw",
		"try",
		"typeof",
		"var",
		"void",
		"while",
		"with",
		"yield",
	},
	Python: {
		"and",
		"as",
		"assert",
		"async",
		"await",
		"break",
		"class",
		"continue",
		"def",
		"del",
		"elif",
		"else",
		"except",
		"finally",
		"for",
		"from",
		"global",
		"if",
		"import",
		"in",
		"is",
		"lambda",
		"nonlocal",
		"not",
		"or",
		"pass",
		"raise",
		"return",
		"try",
		"while",
		"with",
		"yield",
	},
	Rust: {
		"Self",
		"as",
		"async",
		"await",
		"break",
		"const",
		"continue",
		"crate",
		"dyn",
		"else",
		"enum",
		"extern",
		"fn",
		"for",
		"if",
		"impl",
		"in",
		"let",
		"loop",
		"match",
		"mod",
		"move",
		"mut",
		"pub",
		"ref",
		"return",
		"self",
		"static",
		"struct",
		"super",
		"trait",
		"type",
		"unsafe",
		"use",
		"where",
		"while",
	},
	SQL: {
		"add",
		"all",
		"alter",
		"and",
		"as",
		"asc",
		"begin",
		"between",
		"by",
		"case",
		"cast",
		"column",
		"commit",
		"create",
		"cross",
		"default",
		"delete",
		"desc",
		"distinct",
		"drop",
		"else",
		"end",
		"exists",
		"from",
		"full",
		"group",
		"having",
		"if",
		"in",
		"index",
		"inner",
		"insert",
		"into",
		"is",
		"join",
		"key",
		"left",
		"like",
		"limit",
		"not",
		"null",
		"on",
		"or",
		"order",
		"outer",
		"primary",
		"references",
		"right",
		"rollback",
		"select",
		"set",
		"table",
		"then",
		"transaction",
		"union",
		"unique",
		"update",
		"values",
		"view",
		"when",
		"where",
		"with",
	},
	Shell: {
		"case",
		"do",
		"done",
		"elif",
		"else",
		"esac",
		"export",
		"fi",
		"for",
		"function",
		"if",
		"in",
		"local",
		"return",
		"select",
		"then",
		"until",
		"while",
	},
}

// Highlight splits one line into tokens, given whether the line before it
// left a block comment open.
//
// The answer about the line before is the caller's to hold, because only the
// caller knows the order the lines are shown in — a viewer scrolls, a diff
// skips, a snippet is given three lines of the middle of a file — and a
// highlighter that kept that itself would be state in a place this package
// keeps nothing.
func Highlight(line string, lang Lang, inComment bool) []Token {
	if lang == LangPlain {
		return highlightPlain(line, lang, inComment)
	}
	rule := rulesFor(lang)
	return highlightLine(line, lang, rule, inComment)
}

// HighlightAll splits a whole file, returning the tokens of each line. The
// comment state is threaded through, so a block comment opened on line four
// is still a comment on line five.
func HighlightAll(lines []string, lang Lang) [][]Token {
	out := make([][]Token, len(lines))
	inComment := false
	for i, line := range lines {
		out[i] = Highlight(line, lang, inComment)
		inComment = blockCommentOpenAfter(lines[i], lang, inComment)
	}
	return out
}

// HighlightSpans is Highlight for a caller that wants the kinds but not the
// text — a minimap, which paints one block per kind and needs to know what
// each block is rather than what it says.
func HighlightSpans(line string, lang Lang, inComment bool) []TokenKind {
	tokens := Highlight(line, lang, inComment)
	kinds := make([]TokenKind, 0, len(tokens))
	for _, t := range tokens {
		kinds = append(kinds, t.Kind)
	}
	return kinds
}

// langRules is everything about a language that the scanner needs and that is
// not its word list: where its comments are, what quotes a string, and which
// of the punctuation counts as structure.
type langRules struct {
	lineComment  []string
	blockOpen    string
	blockClose   string
	blockComment bool
	quotes       string
	// escape is the character that makes the next one literal, which is a
	// backslash in every language here except Python's raw strings.
	escape byte
	// hashComment is Python's and Shell's comment sigil, kept apart from the
	// line-comment list because it is not followed by anything: `#comment`
	// and `# comment` are both comments and only the first is a word.
	hashComment bool
}

func rulesFor(lang Lang) langRules {
	switch lang {
	case Go:
		return langRules{
			lineComment:  []string{"//"},
			blockOpen:    "/*",
			blockClose:   "*/",
			blockComment: true,
			quotes:       "\"`",
			escape:       '\\',
		}
	case JavaScript:
		return langRules{
			lineComment:  []string{"//"},
			blockOpen:    "/*",
			blockClose:   "*/",
			blockComment: true,
			quotes:       "\"'`",
			escape:       '\\',
		}
	case Rust:
		return langRules{
			lineComment:  []string{"//"},
			blockOpen:    "/*",
			blockClose:   "*/",
			blockComment: true,
			quotes:       "\"'",
			escape:       '\\',
		}
	case Python:
		return langRules{
			lineComment:  []string{"#"},
			blockComment: false,
			quotes:       "\"'",
			// A backslash still escapes in Python, and the scanner cannot tell
			// a raw string from a cooked one without parsing; treating it as
			// escaping is wrong for a raw string and right for the rest.
			escape: '\\',
		}
	case SQL:
		return langRules{
			lineComment:  []string{"--"},
			blockComment: false,
			quotes:       "'\"",
		}
	case Shell:
		return langRules{
			lineComment:  []string{"#"},
			blockComment: false,
			quotes:       "\"'",
			escape:       '\\',
		}
	}
	return langRules{quotes: "\"`", escape: '\\'}
}

// blockCommentOpenAfter is whether the line leaves a block comment open,
// which is the one piece of state the scanner carries and the only reason
// Highlight takes a flag.
func blockCommentOpenAfter(line string, lang Lang, wasOpen bool) bool {
	rule := rulesFor(lang)
	if !rule.blockComment {
		return false
	}
	if wasOpen {
		return !strings.Contains(line, rule.blockClose)
	}
	// Scan for an opener the same way Highlight does, so that an opener inside
	// a string is not counted: a line holding `s := "/*"` opens nothing.
	return commentDepth(line, lang) > 0
}

// commentDepth is how many block comments the line opens without closing,
// counting strings so that one inside a string does not count.
func commentDepth(line string, lang Lang) int {
	rule := rulesFor(lang)
	if !rule.blockComment {
		return 0
	}
	depth := 0
	for i := 0; i < len(line); {
		if c := line[i]; strings.ContainsRune(rule.quotes, rune(c)) {
			i = skipString(line, i, rule)
			continue
		}
		if strings.HasPrefix(line[i:], rule.blockOpen) {
			depth++
			i += len(rule.blockOpen)
			continue
		}
		if depth > 0 && strings.HasPrefix(line[i:], rule.blockClose) {
			depth--
			i += len(rule.blockClose)
			continue
		}
		i++
	}
	return depth
}

// highlightLine is the scanner itself: one pass over the line, emitting a run
// each time the kind changes.
func highlightLine(line string, lang Lang, rule langRules, inComment bool) []Token {
	var out []Token
	emit := func(text string, kind TokenKind) {
		if text == "" {
			return
		}
		if n := len(out); n > 0 && out[n-1].Kind == kind {
			// Runs of the same kind are merged, so `func main()` is two tokens
			// rather than nine. A caller that draws runs does not care, and one
			// that counts them should not be counting punctuation characters.
			out[n-1].Text += text
			return
		}
		out = append(out, Token{Text: text, Kind: kind})
	}

	if inComment {
		// The whole line is comment until the closer, and whatever is after it
		// is code again — which is what makes `/* */ x := 1` highlight the
		// trailing statement.
		end := strings.Index(line, rule.blockClose)
		if end < 0 {
			emit(line, Comment)
			return out
		}
		emit(line[:end+len(rule.blockClose)], Comment)
		return append(out, highlightLine(line[end+len(rule.blockClose):], lang, rule, false)...)
	}

	for i := 0; i < len(line); {
		rest := line[i:]
		c := line[i]
		if startsAny(rest, rule.lineComment) {
			// A line comment runs to the end and there is nothing after it, so
			// it is the last token of the line by construction.
			emit(rest, Comment)
			return out
		}
		if rule.hashComment && c == '#' {
			// Python's and Shell's: the sigil is enough, with or without a
			// space, which is why it is not in lineComment — there `//` has to
			// be two characters for the same reason.
			emit(line[i:], Comment)
			return out
		}
		if rule.blockComment && strings.HasPrefix(rest, rule.blockOpen) {
			emit(rest, Comment)
			return out
		}
		if strings.ContainsRune(rule.quotes, rune(c)) {
			end := skipString(line, i, rule)
			emit(line[i:end], String)
			i = end
			continue
		}
		if isDigit(c) {
			end := scanNumber(line, i)
			emit(line[i:end], Number)
			i = end
			continue
		}
		if isIdentStart(c) {
			end := scanWord(line, i)
			word := line[i:end]
			switch {
			case isKeyword(word, lang):
				emit(word, Keyword)
			case looksLikeType(word):
				emit(word, Type)
			case isCall(line, end):
				emit(word, Function)
			default:
				emit(word, TokenPlain)
			}
			i = end
			continue
		}
		if isPunct(c) {
			emit(string(c), Punct)
			i++
			continue
		}
		// Everything else — whitespace above all — is Plain, and it is run
		// forward so that the indent of a line is one token rather than one
		// per space.
		end := i
		for end < len(line) && !isDigit(line[end]) && !isIdentStart(line[end]) &&
			!isPunct(line[end]) && !startsAny(line[end:], rule.lineComment) &&
			!strings.ContainsRune(rule.quotes, rune(line[end])) &&
			!(rule.hashComment && line[end] == '#') &&
			!(rule.blockComment && strings.HasPrefix(line[end:], rule.blockOpen)) {
			end++
		}
		if end == i {
			end++
		}
		emit(line[i:end], TokenPlain)
		i = end
	}
	return out
}

// highlightPlain is what a line of a language with no words gets: strings and
// numbers still read as themselves, because a log line with a URL in it
// should not be one flat colour.
func highlightPlain(line string, _ Lang, _ bool) []Token {
	rule := rulesFor(LangPlain)
	var out []Token
	emit := func(text string, kind TokenKind) {
		if text == "" {
			return
		}
		if n := len(out); n > 0 && out[n-1].Kind == kind {
			out[n-1].Text += text
			return
		}
		out = append(out, Token{Text: text, Kind: kind})
	}
	for i := 0; i < len(line); {
		c := line[i]
		if strings.ContainsRune(rule.quotes, rune(c)) {
			end := skipString(line, i, rule)
			emit(line[i:end], String)
			i = end
			continue
		}
		if isDigit(c) {
			end := scanNumber(line, i)
			emit(line[i:end], Number)
			i = end
			continue
		}
		emit(string(c), TokenPlain)
		i++
	}
	return out
}

// skipString is where a string that starts at i ends, the opening quote
// included.
//
// The escape is honoured where the language has one, and the closing quote is
// the first one that is not escaped. A string with no closer runs to the end
// of the line, which is what a half-typed line looks like and is why a scanner
// is the right tool for a file as it is being written.
func skipString(line string, i int, rule langRules) int {
	quote := line[i]
	j := i + 1
	for j < len(line) {
		if rule.escape != 0 && line[j] == rule.escape {
			j += 2
			continue
		}
		if line[j] == quote {
			return j + 1
		}
		j++
	}
	return len(line)
}

// scanNumber is where the literal that starts at i ends: digits, underscores
// and a decimal point, plus an exponent and a suffix where there is one.
//
// The base prefix matters more than it looks. `0xff` and `0b1011` and `0755`
// are all numbers, and a scanner that stopped at the `x` would colour half of
// each of them as an identifier — which is the kind of small wrong that makes
// a highlighter look broken rather than simple.
func scanNumber(line string, i int) int {
	if line[i] == '0' && i+1 < len(line) {
		switch line[i+1] {
		case 'x', 'X', 'b', 'B', 'o', 'O':
			j := i + 2
			for j < len(line) && (isHexDigit(line[j]) || line[j] == '_') {
				j++
			}
			return j
		}
	}
	j := i
	for j < len(line) && (isDigit(line[j]) || line[j] == '_' || line[j] == '.') {
		j++
	}
	if j < len(line) && (line[j] == 'e' || line[j] == 'E') {
		k := j + 1
		if k < len(line) && (line[k] == '+' || line[k] == '-') {
			k++
		}
		for k < len(line) && isDigit(line[k]) {
			k++
		}
		if k > j+1 {
			j = k
		}
	}
	// The suffixes a number can carry: a length in Go, a type in Rust, a
	// width in C. Without them `1i64` is a number and an identifier, and the
	// identifier wins because it is scanned next.
	for _, suffix := range []string{"i64", "i32", "i16", "i8", "u64", "u32", "u16", "u8",
		"f64", "f32", "usize", "isize", "L", "l", "f", "u"} {
		if strings.HasPrefix(line[j:], suffix) {
			return j + len(suffix)
		}
	}
	return j
}

// scanWord is where the identifier that starts at i ends. Dots are inside a
// word because `core.Tokens` is one name to a reader, and a selector chain
// broken into six colours is worse than one.
func scanWord(line string, i int) int {
	j := i
	for j < len(line) && isIdentPart(line[j]) {
		j++
	}
	if j < len(line) && line[j] == '.' && j+1 < len(line) && isIdentStart(line[j+1]) {
		k := j + 1
		for k < len(line) && isIdentPart(line[k]) {
			k++
		}
		return k
	}
	return j
}

// isCall is whether the word that ended at end is followed by an opening
// parenthesis, which is what makes it a call rather than a variable.
//
// It is a guess with a good hit rate and no false positives worth the name:
// `if (x)` is a keyword and never reaches here, and `func` is a keyword too.
// What it does get wrong is a declaration — `func main(` names a function and
// is coloured as a keyword, not as a call — which is a name worth having a
// comment about.
func isCall(line string, end int) bool {
	j := end
	for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
		j++
	}
	return j < len(line) && line[j] == '('
}

// looksLikeType is whether a word is written the way its language writes its
// types: capitalised. It is a convention rather than a rule, and the only
// thing that makes it worth having is that it separates `Callback` from
// `callback` without a parser.
func looksLikeType(word string) bool {
	if word == "" {
		return false
	}
	c := word[0]
	return c >= 'A' && c <= 'Z' && word != "True" && word != "False" && word != "None"
}

// caseFolded is the languages whose reserved words are written in lower case
// but match in any case. SQL is the only one here, and it is the reason a
// scanner cannot just look the word up: `SELECT` is as much a keyword as
// `select`, and a keyword list that only knew the lower case would highlight
// every query in a file as one long run of capitals.
var caseFolded = map[Lang]bool{SQL: true}

// isKeyword reports whether word is reserved in that language.
func isKeyword(word string, lang Lang) bool {
	words, ok := keywordSets[lang]
	if !ok {
		return false
	}
	if caseFolded[lang] {
		word = strings.ToLower(word)
	}
	i := sort.SearchStrings(words, word)
	return i < len(words) && words[i] == word
}

// startsAny reports whether s begins with any of the given sigils.
func startsAny(s string, sigils []string) bool {
	for _, sigil := range sigils {
		if strings.HasPrefix(s, sigil) {
			return true
		}
	}
	return false
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == '_'
}

func isIdentStart(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}

func isIdentPart(c byte) bool { return isIdentStart(c) || isDigit(c) }

func isPunct(c byte) bool {
	return strings.IndexByte("+-*/%=<>!&|^~?:;,.()[]{}@#$\\", c) >= 0
}

// MatchSpans is where the matched runs of a search are in a line, for a
// viewer or a diff that has to mark them.
//
// It is on its own and not a field of Token because most callers have no
// search at all: a viewer with no query in it would carry a slice of empty
// ranges for every line of a file of three thousand.
type MatchSpans struct {
	// Line is which line the runs are on, counted from zero.
	Line int
	// From and To are byte offsets into that line, To exclusive.
	From, To int
}

// FindMatches is where the occurrences of needle are in line, left to right
// and not overlapping.
//
// The comparison is case-insensitive, because a search box that finds nothing
// when somebody types the word in the wrong case is a search box that looks
// broken — and case matters for the code itself, which is why the caller's
// own text is left alone and only the offsets come back.
func FindMatches(line, needle string) []MatchSpans {
	if needle == "" {
		return nil
	}
	var out []MatchSpans
	lowerLine := strings.ToLower(line)
	lowerNeedle := strings.ToLower(needle)
	for at := 0; ; {
		i := strings.Index(lowerLine[at:], lowerNeedle)
		if i < 0 {
			return out
		}
		from := at + i
		out = append(out, MatchSpans{From: from, To: from + len(needle)})
		// Past the whole match rather than past its first character, so that
		// searching for "aa" in "aaaa" finds two and not three.
		at = from + len(needle)
		if at > len(line) {
			return out
		}
	}
}

// AnsiSpans is the colouring of a run of terminal output, as the escape
// sequences say it.
//
// It is a separate type from Token because the two answer different
// questions: a Token is what the scanner decided a piece of source is, and an
// AnsiSpan is what a program asked to be drawn a certain way. Merging them
// would mean a terminal output and a source file could not both be shown by
// the same code, which is the whole of what a shared drawing routine is for.
type AnsiSpan struct {
	// Text is the run, with no escape sequences in it.
	Text string
	// Bold, Dim, Italic and Underline are what the sequences asked for.
	Bold, Dim, Italic, Underline bool
	// Colour is the foreground, and Background the ground behind it. Zero
	// alpha means the terminal's own colour rather than one that was asked
	// for.
	Colour, Background AnsiColour
	// Bold is what brightens a colour to its strong form, which is why it is
	// separate from the colour itself: a terminal has sixteen and the pair is
	// how it has sixteen.
	Bright bool
}

// AnsiColour is one of the sixteen colours a terminal has, or none.
type AnsiColour int

const (
	// AnsiDefault is the terminal's own colour: whatever the window makes
	// text that was not coloured look, which is the only right answer for
	// output a program never coloured.
	AnsiDefault AnsiColour = iota
	AnsiBlack
	AnsiRed
	AnsiGreen
	AnsiYellow
	AnsiBlue
	AnsiMagenta
	AnsiCyan
	AnsiWhite
	// AnsiBrightBlack and the seven after it are the bright forms.
	AnsiBrightBlack
	AnsiBrightRed
	AnsiBrightGreen
	AnsiBrightYellow
	AnsiBrightBlue
	AnsiBrightMagenta
	AnsiBrightCyan
	AnsiBrightWhite
)

// ParseAnsi is a run of terminal output split into what to draw, with the
// escape sequences taken out.
//
// The sequences handled are the ones a build tool and a test runner actually
// emit: SGR colour and attribute changes, and nothing else. A CSI sequence of
// any other kind — a cursor move, a screen clear — is dropped, because a log
// view is a list of lines and not a screen: replaying a cursor move would
// move text that is not on a screen, and keeping the bytes would show
// `[2J` to the reader.
func ParseAnsi(s string) []AnsiSpan {
	var out []AnsiSpan
	cur := AnsiSpan{}
	emit := func(text string) {
		if text == "" {
			return
		}
		if n := len(out); n > 0 && out[n-1] == cur {
			out[n-1].Text += text
			return
		}
		cur.Text = text
		out = append(out, cur)
	}

	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			at := ansiEnd(s, i)
			if at < 0 {
				// An escape with no final byte is not an escape; the rest of
				// the line is text.
				emit(s[i:])
				return out
			}
			// A sequence that is not a colour or attribute change is dropped
			// rather than shown and rather than acted on. Showing it puts
			// "[2J" in front of a reader; acting on it would move text that is
			// not on a screen, because a log view is a list of lines.
			if s[at] == 'm' {
				if params := s[i+2 : at]; params == "" || params == "0" {
					cur = AnsiSpan{}
				} else {
					applySGR(&cur, params)
				}
			}
			i = at + 1
			continue
		}
		// Ordinary text up to the next escape.
		end := i
		for end < len(s) && s[end] != 0x1b {
			end++
		}
		emit(s[i:end])
		i = end
	}
	return out
}

// ansiEnd is where the escape sequence starting at i ends, or -1 when there is
// none: a CSI sequence runs from ESC [ to a byte in the range 0x40 to 0x7e,
// which is the letter saying what kind of sequence it is.
func ansiEnd(s string, i int) int {
	for j := i + 2; j < len(s); j++ {
		if s[j] >= 0x40 && s[j] <= 0x7e {
			return j
		}
	}
	return -1
}

// applySGR is what one SGR run of parameters asks for. Several may be in one
// escape, separated by semicolons, and a reset in the middle of them means
// the ones after it are from zero.
func applySGR(cur *AnsiSpan, params string) {
	if params == "" {
		*cur = AnsiSpan{}
		return
	}
	for _, p := range strings.Split(params, ";") {
		n := 0
		for _, c := range p {
			if c < '0' || c > '9' {
				n = -1
				break
			}
			n = n*10 + int(c-'0')
		}
		if n < 0 {
			continue
		}
		switch {
		case n == 0:
			*cur = AnsiSpan{}
		case n == 1:
			cur.Bold = true
		case n == 2:
			cur.Dim = true
		case n == 3:
			cur.Italic = true
		case n == 4:
			cur.Underline = true
		case n == 22:
			cur.Bold, cur.Dim = false, false
		case n == 23:
			cur.Italic = false
		case n == 24:
			cur.Underline = false
		case n >= 30 && n <= 37:
			cur.Colour = AnsiColour(n - 30 + 1)
			cur.Bright = false
		case n == 39:
			cur.Colour = AnsiDefault
		case n >= 40 && n <= 47:
			cur.Background = AnsiColour(n - 40 + 1)
		case n == 49:
			cur.Background = AnsiDefault
		case n >= 90 && n <= 97:
			cur.Colour = AnsiColour(n - 90 + 9)
			cur.Bright = true
		case n >= 100 && n <= 107:
			cur.Background = AnsiColour(n - 100 + 9)
			cur.Bright = true
		}
	}
}
