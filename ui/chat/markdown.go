package chat

import "strings"

// Parse turns markdown into blocks.
//
// It is the one part of a chat view with no colour in it: it takes a string
// and returns a tree, and a Context would give it nothing to say. That is why
// it can be tested exactly — the tests in chat_test.go assert the sequence of
// block kinds a document cuts into, which is the whole contract, and which a
// rendering test could never tell you.
//
// The shape of the parser is the usual one, and the reason is that the other
// shape does not work: the document is first cut into blocks by blank lines
// and fences, and only then is the inside of each block read. Trying to answer
// "is this a heading or a paragraph?" at every character means asking a
// question whose answer depends on lines the reader has not seen yet — which is
// how regular expressions for markdown become a hundred lines of lookaround
// that nobody can change without breaking a case they never saw.
//
// Two deliberate narrownesses, both chosen so that a test can be written
// against them:
//
//   - Setext headings ("Title\n=====") are not headings. They are a rule, or a
//     paragraph. A model emits ATX headings and a drawer would rather show a
//     paragraph than a rule for a line it mistook.
//   - Emphasis is one or two of the same character. "***bold italic***" is
//     bold, and the third marker is text.
func Parse(src string) []Block {
	lines := splitLines(src)
	var out []Block
	for i := 0; i < len(lines); {
		line := lines[i]
		switch {
		case blank(line):
			i++

		case fenceOf(line) != "":
			b, next := readFence(lines, i)
			out = append(out, b)
			i = next

		case isRule(line):
			out = append(out, Block{Kind: BlockRule})
			i++

		case isHeading(line):
			level, text, _ := headingLevel(line)
			out = append(out, Block{Kind: BlockHeading, Level: level, Spans: ParseInline(text)})
			i++

		case quotePrefix(line) >= 0:
			b, next := readQuote(lines, i)
			out = append(out, b)
			i = next

		case listItem(line).ok:
			b, next := readList(lines, i)
			out = append(out, b)
			i = next

		case isTableHead(lines, i):
			b, next := readTable(lines, i)
			out = append(out, b)
			i = next

		default:
			b, next := readPara(lines, i)
			out = append(out, b)
			i = next
		}
	}
	return out
}

// BlockKind is what a block is, and the thing Parse's tests assert on. The
// strings are the ones a caller would write in a switch, which is why they are
// lowercase words rather than a numbered enum.
type BlockKind string

const (
	// BlockHeading is "# text", with Level counting the hashes.
	BlockHeading BlockKind = "heading"
	// BlockPara is running prose, possibly over several lines.
	BlockPara BlockKind = "para"
	// BlockCode is a fenced block, with Code and Lang.
	BlockCode BlockKind = "code"
	// BlockList is one run of list items, ordered or not.
	BlockList BlockKind = "list"
	// BlockQuote is a run of ">" lines, parsed with Parse again.
	BlockQuote BlockKind = "quote"
	// BlockTable is a pipe table with a head row and a delimiter row.
	BlockTable BlockKind = "table"
	// BlockRule is a horizontal rule: three or more of -, _ or * alone.
	BlockRule BlockKind = "rule"
)

// Block is one thing in a document. Which fields are meaningful is decided by
// Kind, and only by Kind — a list's Rows is always nil, a paragraph's Items is
// always nil — so a reader can hold one struct in their head instead of seven.
type Block struct {
	Kind BlockKind

	// Level is a heading's depth, 1 to 6.
	Level int
	// Spans are a heading's or paragraph's styled runs.
	Spans []Span

	// Code is a fenced block's text, without the fences, and Lang its info
	// string lowercased — "```Go" and "```go" are the same language, and a
	// highlighter that treated them as two would colour half the code wrong.
	Code string
	Lang string

	// Items are a list's items in document order. Ordered is per item rather
	// than per block because a model that switches from bullets to numbers
	// halfway through a list means it, and joining two blocks would lose the
	// indentation that says the second half belongs to the first.
	Items []Item

	// Blocks are a quote's contents, run through Parse again: a quote is a
	// document, not a string with a mark in front of it.
	Blocks []Block

	// Head, Rows and Aligns are a table's head row, its body rows, and the
	// per-column alignment read off its delimiter row.
	Head   []string
	Rows   [][]string
	Aligns []Align
}

// Item is one list item.
type Item struct {
	// Depth is how far in it sits, counted in two spaces of indent.
	Depth int
	// Ordered says it came off a "1." rather than a "-".
	Ordered bool
	// Marker is the text the source used, kept so a caller can re-render a
	// list the way its author numbered it rather than renumbering it.
	Marker string
	// Spans are the item's styled runs.
	Spans []Span
	// Checked is -1 for an ordinary item, 0 and 1 for a task item's boxes.
	// It is an int rather than a *bool because "not a task item" is the common
	// case and a pointer for every bullet in a document is a lot of boxes
	// holding nothing.
	Checked int
}

// Span is a run of text sharing one set of marks.
type Span struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
	// Link is the destination when the span is a link's label, empty
	// otherwise.
	Link string
}

// Align is a table column's alignment, from its delimiter row.
type Align int

const (
	// AlignNone is a column with no alignment said, which is left alone
	// rather than forced to one side.
	AlignNone Align = iota
	AlignLeft
	AlignCenter
	AlignRight
)

func (a Align) String() string {
	switch a {
	case AlignLeft:
		return "left"
	case AlignCenter:
		return "center"
	case AlignRight:
		return "right"
	}
	return "none"
}

// ── line helpers ───────────────────────────────────────────────────────────

// splitLines normalises the three line endings onto one, so a document that
// came off a Windows clipboard is read the same as one that did not.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	// A source ending in a newline has one more element than it has lines, and
	// the last of those is the absence of a line rather than a line. Left in,
	// it becomes an extra empty line inside an unterminated code block and a
	// trailing newline on every code block that has one.
	if n := len(lines); n > 1 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}

func blank(s string) bool { return strings.TrimSpace(s) == "" }

// indentOf returns how many columns of leading space a line has, counting a
// tab as four so a list indented with tabs is not read as unindented.
func indentOf(line string) int {
	n := 0
	for _, r := range line {
		switch r {
		case ' ':
			n++
		case '\t':
			n += 4
		default:
			return n
		}
	}
	return n
}

// fenceOf returns the fence a line opens or closes a code block with, or "".
func fenceOf(line string) string {
	t := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(t, "```"):
		return "```"
	case strings.HasPrefix(t, "~~~"):
		return "~~~"
	}
	return ""
}

// headingLevel returns how deep a heading is and its text, or 0 for a line
// that is not one.
//
// The space after the hashes is required, which is what keeps "####### seven
// hashes" and "#hashtag" out of the heading branch.
func headingLevel(line string) (level int, text string, ok bool) {
	t := strings.TrimLeft(line, " ")
	n := 0
	for n < len(t) && t[n] == '#' {
		n++
	}
	if n == 0 || n > 6 || n == len(t) {
		return 0, "", false
	}
	if t[n] != ' ' && t[n] != '\t' {
		return 0, "", false
	}
	// Trailing hashes are the author closing the heading, not part of it.
	return n, strings.TrimSpace(strings.TrimRight(strings.TrimSpace(t[n:]), "#")), true
}

// isHeading reports whether a line opens a heading. It is a function because a
// selector cannot be taken from a call that returns three values, and the
// switch in Parse needs the answer as a boolean.
func isHeading(line string) bool {
	_, _, ok := headingLevel(line)
	return ok
}

// isRule reports a horizontal rule: three or more of one mark, with spaces
// allowed between them so that "- - -" is one.
//
// All the marks must be the *same* one, and nothing else may be on the line.
// Both halves are load-bearing: the first is what keeps "- - -" from reading as
// three bullets, and the second is what keeps "***bold***" a paragraph.
func isRule(line string) bool {
	t := strings.TrimSpace(line)
	if t == "" {
		return false
	}
	var mark byte
	n := 0
	for i := range len(t) {
		switch t[i] {
		case ' ', '\t':
		case '-', '_', '*':
			switch {
			case mark == 0:
				mark = t[i]
			case mark != t[i]:
				return false
			}
			n++
		default:
			return false
		}
	}
	return n >= 3
}

// quotePrefix returns where the ">" is on a quote line, or -1 for a line that
// is not one.
func quotePrefix(line string) int {
	t := strings.TrimLeft(line, " ")
	if !strings.HasPrefix(t, ">") {
		return -1
	}
	return len(line) - len(t)
}

// itemStart is the answer to "does this line begin a list item, and what
// number is it".
type itemStart struct {
	ok      bool
	indent  int
	ordered bool
	marker  string
	text    string
}

// listItem reads a line as the start of a list item.
func listItem(line string) itemStart {
	indent := indentOf(line)
	rest := strings.TrimLeft(line, " \t")
	if rest == "" {
		// Whitespace only. TrimLeft leaves nothing and the check below would
		// read the first character of the empty string.
		return itemStart{}
	}
	switch rest[0] {
	case '-', '*', '+':
		if !spaceAfter(rest, 1) {
			return itemStart{}
		}
		return itemStart{ok: true, indent: indent, marker: string(rest[0]),
			text: strings.TrimSpace(rest[1:])}
	}
	n := 0
	for n < len(rest) && rest[n] >= '0' && rest[n] <= '9' {
		n++
	}
	if n == 0 || n+1 >= len(rest) {
		return itemStart{}
	}
	if rest[n] != '.' && rest[n] != ')' {
		return itemStart{}
	}
	if !spaceAfter(rest, n+1) {
		return itemStart{}
	}
	return itemStart{ok: true, indent: indent, ordered: true,
		marker: rest[:n+1], text: strings.TrimSpace(rest[n+1:])}
}

func spaceAfter(s string, i int) bool {
	return i < len(s) && (s[i] == ' ' || s[i] == '\t')
}

// isTableHead reports that line i is the head row of a table: a row with pipes
// in it whose next line is the delimiter row that says which columns line up
// how. The delimiter row is what makes it a table rather than two paragraphs
// that happen to contain bars.
func isTableHead(lines []string, i int) bool {
	if i+1 >= len(lines) || !strings.Contains(lines[i], "|") {
		return false
	}
	return delimiterAligns(lines[i+1]) != nil
}

// delimiterAligns reads a "| --- | :-: |" row, returning each column's
// alignment. It returns nil for anything that is not one, which is what keeps
// a paragraph of "a --- b" out of the table branch.
func delimiterAligns(line string) []Align {
	t := strings.TrimSpace(line)
	if !strings.Contains(t, "-") {
		return nil
	}
	cells := splitRow(t)
	if len(cells) == 0 {
		return nil
	}
	out := make([]Align, len(cells))
	for i, cell := range cells {
		c := strings.TrimSpace(cell)
		if c == "" {
			return nil
		}
		left, right := strings.HasPrefix(c, ":"), strings.HasSuffix(c, ":")
		body := strings.Trim(c, ":")
		if body == "" || strings.Trim(body, "-") != "" {
			return nil
		}
		switch {
		case left && right:
			out[i] = AlignCenter
		case right:
			out[i] = AlignRight
		case left:
			out[i] = AlignLeft
		default:
			out[i] = AlignNone
		}
	}
	return out
}

// splitRow cuts one table row into its cells, dropping the empty ends a row
// with an outer pipe has and keeping an empty cell in the middle, because
// "| a |  | b |" is three columns and not two.
func splitRow(line string) []string {
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")
	if t == "" {
		return nil
	}
	parts := strings.Split(t, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// ── block readers ──────────────────────────────────────────────────────────

// readFence reads a fenced code block, returning the block and the line after
// it. An unterminated fence runs to the end of the document rather than
// disappearing: half a code block is worse to lose than to show, because the
// text a model was streaming when the user pressed stop is still worth reading.
func readFence(lines []string, i int) (Block, int) {
	fence := fenceOf(lines[i])
	lang := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), fence)))

	var body []string
	j := i + 1
	for ; j < len(lines); j++ {
		if t := strings.TrimSpace(lines[j]); fenceOf(t) == fence && t == fence {
			j++
			break
		}
		body = append(body, lines[j])
	}
	return Block{Kind: BlockCode, Code: strings.Join(body, "\n"), Lang: lang}, j
}

// readQuote reads the run of ">" lines and parses what is inside it.
func readQuote(lines []string, i int) (Block, int) {
	var inner []string
	j := i
	for ; j < len(lines); j++ {
		at := quotePrefix(lines[j])
		if at < 0 {
			// A blank line inside a quote continues it; anything else ends it.
			if blank(lines[j]) {
				inner = append(inner, "")
				continue
			}
			break
		}
		text := lines[j][at+1:]
		inner = append(inner, strings.TrimPrefix(text, " "))
	}
	return Block{Kind: BlockQuote, Blocks: Parse(strings.Join(inner, "\n"))}, j
}

// readPara reads the run of lines that is prose: until a blank line or the
// first line that opens another kind of block. The second test is the one that
// matters — a paragraph followed by "## A heading" is a paragraph and a
// heading, not a paragraph that swallowed the heading.
func readPara(lines []string, i int) (Block, int) {
	var buf []string
	j := i
	for ; j < len(lines); j++ {
		if blank(lines[j]) || startsBlock(lines, j) {
			break
		}
		buf = append(buf, strings.TrimSpace(lines[j]))
	}
	return Block{Kind: BlockPara, Spans: ParseInline(strings.Join(buf, "\n"))}, j
}

// startsBlock reports that line i opens a block, which is how a paragraph
// knows where it ends.
func startsBlock(lines []string, i int) bool {
	line := lines[i]
	switch {
	case blank(line), fenceOf(line) != "", isRule(line):
		return true
	case isHeading(line), quotePrefix(line) >= 0, listItem(line).ok:
		return true
	case i+1 < len(lines) && delimiterAligns(lines[i+1]) != nil && strings.Contains(line, "|"):
		return true
	}
	return false
}

// readList reads one run of list items, however they are indented, into a
// single block. Nesting is depth on an item rather than a block inside a
// block, because a nested list drawn as its own block would break the
// paragraph rhythm of the list holding it.
func readList(lines []string, i int) (Block, int) {
	var b Block
	b.Kind = BlockList

	j := i
	for j < len(lines) {
		it := listItem(lines[j])
		if !it.ok {
			// A blank line only ends the list if what follows is not more of
			// it; two paragraphs of bullets are one list to a reader, and two
			// blocks with a gap between them to a layout.
			if blank(lines[j]) {
				k := j
				for k < len(lines) && blank(lines[k]) {
					k++
				}
				if k >= len(lines) || !listItem(lines[k]).ok {
					return b, j
				}
				j = k
				continue
			}
			// An indented, non-item line belongs to the item above it, which is
			// how a wrapped bullet continues without a two-space continuation
			// marker. The newline between them is a span of its own so the
			// renderer can break there rather than run the two together.
			if len(b.Items) > 0 && indentOf(lines[j]) > 0 {
				last := &b.Items[len(b.Items)-1]
				last.Spans = append(last.Spans, Span{Text: "\n"})
				last.Spans = append(last.Spans, ParseInline(strings.TrimSpace(lines[j]))...)
				j++
				continue
			}
			return b, j
		}

		var item Item
		item.Depth = it.indent / 2
		item.Ordered = it.ordered
		item.Marker = it.marker
		item.Checked = -1
		text := it.text
		// A task item: "- [ ] do it", "- [x] did it".
		if len(text) >= 3 && text[0] == '[' && text[2] == ']' &&
			(text[1] == ' ' || text[1] == 'x' || text[1] == 'X') {
			item.Checked = 0
			if text[1] != ' ' {
				item.Checked = 1
			}
			text = text[3:]
		}
		item.Spans = ParseInline(strings.TrimSpace(text))
		b.Items = append(b.Items, item)
		j++
	}
	return b, j
}

// readTable reads a head row, its delimiter row and the body under it.
func readTable(lines []string, i int) (Block, int) {
	b := Block{Kind: BlockTable, Head: splitRow(lines[i]), Aligns: delimiterAligns(lines[i+1])}
	j := i + 2
	for ; j < len(lines); j++ {
		if blank(lines[j]) || !strings.Contains(lines[j], "|") {
			break
		}
		b.Rows = append(b.Rows, splitRow(lines[j]))
	}
	return b, j
}

// ── inline ─────────────────────────────────────────────────────────────────

// ParseInline splits one line's markdown text into styled spans.
//
// It is exported because the callers that build a message by hand — a search
// hit, a system notice, a snippet in a list — need the same reading of the
// same markup that Parse gives a model's answer. One reader for both is the
// point: two would disagree about what "**bold**" means and one of them would
// be wrong.
func ParseInline(s string) []Span {
	var out []Span
	var lit []byte

	flush := func() {
		if len(lit) > 0 {
			out = append(out, Span{Text: string(lit)})
			lit = lit[:0]
		}
	}
	for i := 0; i < len(s); {
		switch ch := s[i]; ch {
		case '\\':
			// A backslash only escapes punctuation; before a letter it is
			// itself, because models write Windows paths and \n in them is not
			// a line break.
			if i+1 < len(s) && isASCIIPunct(s[i+1]) {
				lit = append(lit, s[i+1])
				i += 2
				continue
			}
			lit = append(lit, ch)
			i++

		case '`':
			n := 1
			for i+n < len(s) && s[i+n] == '`' {
				n++
			}
			end := findRun(s, '`', n, i+n)
			if end < 0 || end == i+n {
				// No closing fence: the backtick is text, not markup.
				lit = append(lit, s[i:i+n]...)
				i += n
				continue
			}
			// A code span is taken whole. Nothing inside it is markup, which
			// is what lets `a * b` and `*emphasis*` in the same sentence read
			// the way their author meant them.
			flush()
			out = append(out, Span{Text: s[i+n : end], Code: true})
			i = end + n

		case '[':
			if sp, next, ok := parseLink(s, i); ok {
				flush()
				out = append(out, sp)
				i = next
				continue
			}
			lit = append(lit, ch)
			i++

		case '~':
			if strings.HasPrefix(s[i:], "~~") {
				if end := strings.Index(s[i+2:], "~~"); end > 0 {
					flush()
					for _, sp := range ParseInline(s[i+2 : i+2+end]) {
						out = append(out, Span{Text: sp.Text, Bold: sp.Bold, Italic: sp.Italic, Code: sp.Code})
					}
					i += 2 + end + 2
					continue
				}
			}
			lit = append(lit, ch)
			i++

		case '*', '_':
			n := 1
			for i+n < len(s) && s[i+n] == ch {
				n++
			}
			// One or two. Three is the second one's partner plus a stray, and
			// reading "***" as bold is what every reader does with it anyway.
			if n > 2 {
				n = 2
			}
			if ch == '_' && !opensAt(s, i) {
				// Inside a word, an underscore belongs to the word:
				// some_call_name must not lose its middle to italics.
				lit = append(lit, ch)
				i++
				continue
			}
			end := findRun(s, ch, n, i+n)
			if end <= i+n {
				lit = append(lit, s[i:i+n]...)
				i += n
				continue
			}
			flush()
			for _, sp := range ParseInline(s[i+n : end]) {
				if n == 2 {
					sp.Bold = true
				} else {
					sp.Italic = true
				}
				out = append(out, sp)
			}
			i = end + n

		default:
			lit = append(lit, ch)
			i++
		}
	}
	flush()
	return out
}

// parseLink reads "[label](url)" at i. A link whose label holds markup keeps
// the text and loses the styling: a nested bold inside a link label is a
// thing a model emits by accident, and flattening it is the only reading that
// cannot come out as an unclosed pair of asterisks in a transcript.
func parseLink(s string, i int) (Span, int, bool) {
	close := strings.IndexByte(s[i:], ']')
	if close < 0 {
		return Span{}, 0, false
	}
	close += i
	if close+1 >= len(s) || s[close+1] != '(' {
		return Span{}, 0, false
	}
	end := strings.IndexByte(s[close+2:], ')')
	if end < 0 {
		return Span{}, 0, false
	}
	end += close + 2
	url := strings.TrimSpace(s[close+2 : end])
	if sp := strings.IndexAny(url, " \t"); sp >= 0 {
		url = url[:sp] // drop the optional "title"
	}
	return Span{Text: plainText(s[i+1 : close]), Link: url}, end + 1, true
}

// SpansText is a run of spans as one string — the words with the markup taken
// out. It is exported because a caller holding a parsed document has exactly
// that problem: a heading's text to put in a label, a list item's words to put
// in a row, a search index to build — and re-joining the marks by hand is how
// a bold run turns up with its asterisks in.
func SpansText(spans []Span) string {
	var b strings.Builder
	for _, sp := range spans {
		b.WriteString(sp.Text)
	}
	return b.String()
}

// plainText is a string's words with its markup taken out. It is the same
// thing SpansText does after parsing, for the places that start with a string.
func plainText(s string) string { return SpansText(ParseInline(s)) }

// opensAt reports whether an underscore at i may open emphasis: only where a
// word does not already run. Without it every snake_case identifier in a code
// suggestion would lose its middle.
func opensAt(s string, i int) bool {
	if i == 0 {
		return true
	}
	return !isWordByte(s[i-1])
}

// findRun returns where a run of exactly n of ch closes at or after from, or
// -1. "Exactly" is what makes “ `a` “ inside a double-backtick span work,
// and what stops a single backtick from closing a span opened with two.
func findRun(s string, ch byte, n, from int) int {
	for i := from; i < len(s); {
		if s[i] != ch {
			i++
			continue
		}
		run := 1
		for i+run < len(s) && s[i+run] == ch {
			run++
		}
		if run == n {
			return i
		}
		i += run
	}
	return -1
}

func isWordByte(b byte) bool {
	return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

func isASCIIPunct(b byte) bool {
	switch {
	case b >= '!' && b <= '/', b >= ':' && b <= '@', b >= '[' && b <= '`', b >= '{' && b <= '~':
		return true
	}
	return false
}
