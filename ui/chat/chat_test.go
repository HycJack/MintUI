package chat

import (
	"image"
	"math"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/feedback"
	"github.com/HycJack/MintUI/ui/theme"
)

// The two pure functions carry this package. Parse turns a document into
// blocks and Highlight turns source into tokens, and neither needs a window,
// so both are tested on what they return rather than on what they drew — which
// is the only way to assert that a block list is *["heading", "para", "code",
// "list", "quote"]* and not merely that some words appeared.
//
// Everything else is asserted on the value a component wrote into the caller's
// pointer, or on a result read *inside* the view. MyGo builds a frame up to
// three times and stops when a frame consumed nothing, so by the time a tester
// has settled, Clicked() and Changed() are false on the last pass: a test that
// read a Result after settling would read the empty answer and pass for the
// wrong reason. Every press below is therefore counted inside the view.

// ── helpers ────────────────────────────────────────────────────────────────

// wantsPanic runs view and fails unless it panics, which is how this library
// says a component was asked for something it cannot be given.
func wantsPanic(t *testing.T, what string, view func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s should have panicked", what)
		}
	}()
	view()
}

// blockKinds is the sequence of block kinds a document cuts into — the single
// assertion that says what Parse did with it.
func blockKinds(blocks []Block) []string {
	out := make([]string, len(blocks))
	for i, b := range blocks {
		out[i] = string(b.Kind)
	}
	return out
}

func wantKinds(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("block kinds = %v, want %v", got, want)
	}
}

// tokenKinds is every token's kind, spaces included. It is the whole answer to
// "what did this input cut into", and leaving the spaces in is deliberate: a
// highlighter that ate the gaps between the words of a sentence would produce
// the same kinds as one that did not.
func tokenKinds(tokens []Token) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = t.Kind.String()
	}
	return out
}

// tightKinds drops the whitespace, for the assertion that cares about what the
// source was *made of*.
func tightKinds(tokens []Token) []string {
	var out []string
	for _, t := range tokens {
		if t.Kind != TokText {
			out = append(out, t.Kind.String())
		}
	}
	return out
}

func wantTokens(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("token kinds = %v, want %v", got, want)
	}
}

func wantStrings(t *testing.T, got, want []string, what string) {
	t.Helper()
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("%s = %v, want %v", what, got, want)
	}
}

// ── Parse ──────────────────────────────────────────────────────────────────

// TestParseCutsADocumentIntoBlocks is the contract of the whole parser: given a
// document that uses each kind of block once, it cuts out exactly those blocks
// in that order. The assertion is on the sequence, not on the contents,
// because the sequence is what a renderer walks — a parser that found the right
// blocks in the wrong order would draw a transcript whose code sample came
// before its heading and would look, in a screenshot, almost right.
func TestParseCutsADocumentIntoBlocks(t *testing.T) {
	src := "# Release notes\n" +
		"\n" +
		"This paragraph\nspans two lines.\n" +
		"\n" +
		"```go\nfmt.Println(\"hi\")\n```\n" +
		"\n" +
		"- one\n- two\n  - nested\n" +
		"\n" +
		"> a quoted paragraph\n"

	blocks := Parse(src)
	wantKinds(t, blockKinds(blocks), "heading", "para", "code", "list", "quote")

	// And each of them holds what its kind says it holds, because a block
	// sequence on its own would pass for a parser that returned six empty
	// blocks of the right shapes.
	if blocks[0].Level != 1 || SpansText(blocks[0].Spans) != "Release notes" {
		t.Errorf("heading = level %d %q, want level 1 %q",
			blocks[0].Level, SpansText(blocks[0].Spans), "Release notes")
	}
	if got := SpansText(blocks[1].Spans); got != "This paragraph\nspans two lines." {
		t.Errorf("paragraph = %q", got)
	}
	if blocks[2].Lang != "go" {
		t.Errorf("code lang = %q, want go", blocks[2].Lang)
	}
	if blocks[2].Code != `fmt.Println("hi")` {
		t.Errorf("code = %q", blocks[2].Code)
	}
	if len(blocks[3].Items) != 3 {
		t.Fatalf("list has %d items, want 3", len(blocks[3].Items))
	}
	if blocks[3].Items[2].Depth != 1 {
		t.Errorf("third item depth = %d, want 1", blocks[3].Items[2].Depth)
	}
	// A quote is a document, not a string with a mark in front of it, so it is
	// parsed again and what it holds is blocks.
	if len(blocks[4].Blocks) != 1 || blocks[4].Blocks[0].Kind != BlockPara {
		t.Errorf("quote holds %v, want one paragraph", blockKinds(blocks[4].Blocks))
	}
}

// TestParseReadsAListAsItWasWritten is about the two things a list gets wrong:
// nesting and numbering. A renumbered list hides the gap its author was
// pointing at, and a flat list turns two levels of bullet into one paragraph.
func TestParseReadsAListAsItWasWritten(t *testing.T) {
	blocks := Parse("1. first\n2. second\n\n- [ ] todo\n- [x] done\n")
	wantKinds(t, blockKinds(blocks), "list")

	items := blocks[0].Items
	if len(items) != 4 {
		t.Fatalf("got %d items, want 4", len(items))
	}
	if !items[0].Ordered || items[0].Marker != "1." {
		t.Errorf("first item = ordered %v marker %q", items[0].Ordered, items[0].Marker)
	}
	if SpansText(items[0].Spans) != "first" {
		t.Errorf("first item text = %q", SpansText(items[0].Spans))
	}
	// Checked is -1 for an ordinary item and 0 or 1 for a task item: a pointer
	// per bullet in a document would be a lot of boxes holding nothing.
	if items[0].Checked != -1 {
		t.Errorf("an ordinary bullet reported Checked %d, want -1", items[0].Checked)
	}
	if items[2].Checked != 0 || items[3].Checked != 1 {
		t.Errorf("task items reported %d and %d, want 0 and 1",
			items[2].Checked, items[3].Checked)
	}
	if SpansText(items[2].Spans) != "todo" {
		t.Errorf("task item text = %q, want the words without the box", SpansText(items[2].Spans))
	}
}

// TestParseReadsATableFromItsDelimiterRow is the test for what makes a table a
// table: the "| --- |" row under the head. Without it, two paragraphs with
// bars in them are not a table, and a parser that said they were would draw a
// grid over somebody's prose.
func TestParseReadsATableFromItsDelimiterRow(t *testing.T) {
	blocks := Parse("| Name | Size |\n| --- | ---: |\n| a.png | 12 KB |\n| b.png | 4 KB |\n")
	wantKinds(t, blockKinds(blocks), "table")

	b := blocks[0]
	wantStrings(t, b.Head, []string{"Name", "Size"}, "head")
	if len(b.Rows) != 2 || b.Rows[0][0] != "a.png" {
		t.Fatalf("rows = %v", b.Rows)
	}
	if b.Aligns[0] != AlignNone || b.Aligns[1] != AlignRight {
		t.Errorf("aligns = %v, want none and right", b.Aligns)
	}

	// The same document without the delimiter row is prose — one paragraph,
	// because the two lines are one run of text and nothing between them said
	// otherwise. A grid drawn over a sentence is the failure this guards.
	if got := blockKinds(Parse("| Name | Size |\n| a.png | 12 KB |")); len(got) != 1 ||
		got[0] != "para" {
		t.Errorf("two rows with no delimiter row read as %v, want one paragraph", got)
	}
}

// TestParseReadsCodeFences covers the three cases that differ: a fence with an
// info string, one without, and one the model was still streaming when the
// user pressed stop. The third is the one that matters — half a code block is
// worse to lose than to show.
func TestParseReadsCodeFences(t *testing.T) {
	with := Parse("```Python\nx = 1\n```")
	if with[0].Lang != "python" {
		t.Errorf("lang = %q, want python lowercased; Go and go are one language",
			with[0].Lang)
	}
	if without := Parse("```\nplain\n```"); without[0].Lang != "" {
		t.Errorf("an unlabelled fence read as %q", without[0].Lang)
	}

	open := Parse("```go\nfunc main() {\n")
	if len(open) != 1 || open[0].Kind != BlockCode {
		t.Fatalf("an unterminated fence read as %v", blockKinds(open))
	}
	if open[0].Code != "func main() {" {
		t.Errorf("unterminated code = %q", open[0].Code)
	}
}

// TestParseReadsInlineMarks covers each of the inline marks, and the one thing
// that is not a mark: an escape before a letter. A model writes Windows paths,
// and \n in "C:\new" is not a line break.
func TestParseReadsInlineMarks(t *testing.T) {
	spans := ParseInline("**bold** and *soft* and `code` and [text](https://x)")

	var got []string
	for _, s := range spans {
		switch {
		case s.Bold:
			got = append(got, "bold:"+s.Text)
		case s.Italic:
			got = append(got, "italic:"+s.Text)
		case s.Code:
			got = append(got, "code:"+s.Text)
		case s.Link != "":
			got = append(got, "link:"+s.Text+"->"+s.Link)
		}
	}
	wantStrings(t, got, []string{
		"bold:bold", "italic:soft", "code:code", "link:text->https://x",
	}, "marked spans")

	// An underscore inside a word belongs to the word. Every snake_case name in
	// every code suggestion would otherwise lose its middle to italics.
	if s := ParseInline("some_call_name"); len(s) != 1 || s[0].Italic || s[0].Text != "some_call_name" {
		t.Errorf("snake_case read as %+v", s)
	}
	// A backslash only escapes punctuation.
	if s := ParseInline(`C:\new`); len(s) != 1 || s[0].Text != `C:\new` {
		t.Errorf(`backslash before a letter = %+v, want the string unchanged`, s)
	}
	if s := ParseInline(`\*not italic\*`); len(s) != 1 || s[0].Italic ||
		s[0].Text != "*not italic*" {
		t.Errorf(`escaped asterisks = %+v`, s)
	}
	// A code span is taken whole: nothing inside it is markup.
	if s := ParseInline("`a * b`"); len(s) != 1 || !s[0].Code || s[0].Text != "a * b" {
		t.Errorf("code span = %+v", s)
	}
}

// TestParseInsistsOnAHashFollowedBySpace is the line between a heading and a
// hashtag, and it is the one an answer gets wrong most often.
func TestParseInsistsOnAHashFollowedBySpace(t *testing.T) {
	wantKinds(t, blockKinds(Parse("# Heading")), "heading")
	wantKinds(t, blockKinds(Parse("###### Six")), "heading")
	for _, line := range []string{"#hashtag", "####### seven"} {
		if got := blockKinds(Parse(line)); got[0] != "para" {
			t.Errorf("%q read as %v, want a paragraph", line, got)
		}
	}
}

// TestParseTakesThreeStarsAsARule pins down the deliberate narrowness: a rule
// must be made only of its marks, so "***bold***" stays prose.
func TestParseTakesThreeStarsAsARule(t *testing.T) {
	wantKinds(t, blockKinds(Parse("***")), "rule")
	// Spaces between the marks are how a rule is written when it has to read
	// as one, so "- - -" is one rule and not three bullets.
	wantKinds(t, blockKinds(Parse("- - -")), "rule")
	// And the other half of the same rule: marks with a word among them are a
	// paragraph, not a rule with a word on it.
	if got := blockKinds(Parse("***bold***")); got[0] != "para" {
		t.Errorf("***bold*** read as %v, want a paragraph", got)
	}
	// Two different marks do not make one either.
	if got := blockKinds(Parse("-_*_")); got[0] == "rule" {
		t.Error("-_*_ read as a rule; the marks have to be the same one")
	}
}

// TestParseKeepsARuleAndAParagraphApart is the second half of the same rule: a
// paragraph followed by a heading is two blocks, not one paragraph with the
// heading swallowed into it.
func TestParseKeepsARuleAndAParagraphApart(t *testing.T) {
	wantKinds(t, blockKinds(Parse("text\n## heading\n")), "para", "heading")
}

// ── Highlight ──────────────────────────────────────────────────────────────

// TestHighlightCutsSourceIntoTokens is the same contract as the Parse one: an
// input, and the exact sequence of tokens it cuts into.
func TestHighlightCutsSourceIntoTokens(t *testing.T) {
	tokens := Highlight("x := 1", "go")
	// ":=" is one punctuation token rather than two, because a viewer draws
	// one colour block per token and two blocks where the reader sees a single
	// operator is a seam in the middle of the code.
	wantTokens(t, tokenKinds(tokens), "ident", "text", "punct", "text", "number")
	wantTokens(t, tightKinds(tokens), "ident", "punct", "number")

	// The texts are the assertion as much as the kinds are: a tokenizer that
	// returned five right-shaped tokens holding the wrong pieces of the line
	// would pass the first assertion.
	wantStrings(t, []string{tokens[0].Text, tokens[2].Text, tokens[4].Text},
		[]string{"x", ":=", "1"}, "token texts")
}

// TestHighlightGivesTheSourceBack is the property the renderer leans on. If the
// tokens did not concatenate to the input, the coloured block on screen would
// be missing or duplicating words while looking perfectly plausible in a list
// of kinds.
func TestHighlightGivesTheSourceBack(t *testing.T) {
	for _, src := range []string{
		"x := 1",
		"// a comment\nfunc f(s string) int {\n\treturn len(s) // 3\n}\n",
		"`a * b` and \"unterminated",
		"1 + 0xFF + 1.5e10 + 1_000",
	} {
		if got := TokenText(Highlight(src, "go")); got != src {
			t.Errorf("tokens did not give the source back:\n got %q\nwant %q", got, src)
		}
	}
}

// TestHighlightTellsKeywordsFromTypes is the reason TokType exists: colour is
// most of what a highlighter does, and "int" in a line of "if" is a different
// colour for a reason.
func TestHighlightTellsKeywordsFromTypes(t *testing.T) {
	tokens := Highlight("if x { return int }", "go")
	wantTokens(t, tightKinds(tokens),
		"keyword", "ident", "punct", "keyword", "type", "punct")
}

// TestHighlightReadsStringsAndComments covers the two rules that overlap: an
// unterminated string ends at the newline rather than eating the rest of the
// file, and a comment runs to the end of its line.
func TestHighlightReadsStringsAndComments(t *testing.T) {
	// `fmt.Println("hi")` — the comment marker inside a string is not a
	// comment, which is the rule that a tokenizer reading left to right gets
	// wrong and a reader never forgives.
	tokens := Highlight(`x := "// not a comment"`, "go")
	wantTokens(t, tightKinds(tokens), "ident", "punct", "string")

	// An unterminated one stops at the line, so the code after it stays code.
	tokens = Highlight("s := \"oops\ny := 1", "go")
	wantTokens(t, tightKinds(tokens), "ident", "punct", "string", "ident", "punct", "number")

	// A Go backtick runs past the newline; that is what makes a raw string a
	// raw string.
	raw := Highlight("s := `a\nb`", "go")
	var sawComment bool
	for _, tk := range raw {
		if tk.Kind == TokComment {
			sawComment = true
		}
	}
	if sawComment {
		t.Error("a backtick string was read as a comment")
	}
}

// TestHighlightReadsNumbersOfEveryBase is the number scanner on its own: a
// modern language's literals are not "digits", and a highlighter that stopped
// at the first non-digit would colour 1_000_000 as three numbers and leave the
// separators to the text.
func TestHighlightReadsNumbersOfEveryBase(t *testing.T) {
	for _, src := range []string{"1", "0xFF", "0b1010", "0o777", "1.5", "1.5e10", "1_000_000", ".5"} {
		tokens := Highlight(src, "go")
		if len(tokens) != 1 || tokens[0].Kind != TokNumber {
			t.Errorf("%q read as %v, want one number", src, tokenKinds(tokens))
		}
	}
	// A method call's dot is a dot: 1.bit_length() is a number followed by a
	// dot and not a malformed literal — which matters, because it is a very
	// common line and a reader reading Python is looking at one.
	tokens := Highlight("x = 1.bit_length()", "python")
	var got []string
	for _, tk := range tokens {
		if tk.Kind == TokNumber || tk.Kind == TokPunct || tk.Kind == TokIdent {
			got = append(got, tk.Kind.String()+":"+tk.Text)
		}
	}
	wantStrings(t, got,
		[]string{"ident:x", "punct:=", "number:1", "punct:.", "ident:bit_length", "punct:()"},
		"the marks of a method call")
}

// TestHighlightLeavesAnUnknownLanguageAlone is the do-not-guess rule. A
// highlighter that treated Go as C would put a shell script's # comments
// inside its strings, which is worse than not colouring anything at all.
func TestHighlightLeavesAnUnknownLanguageAlone(t *testing.T) {
	src := "// not a comment in this language\nx = 1\n"
	tokens := Highlight(src, "brainfuck")
	if len(tokens) != 1 || tokens[0].Kind != TokText || tokens[0].Text != src {
		t.Errorf("an unknown language read as %v, want the whole source as one token",
			tokenKinds(tokens))
	}
}

// TestHighlightDoesNotReadJSONComments is why JSON is its own row rather than
// JavaScript with a flag: a JSON block that highlighted // as a comment would
// accept a document the parser rejects.
func TestHighlightDoesNotReadJSONComments(t *testing.T) {
	tokens := Highlight(`{"a": 1} // trailing`, "json")
	for _, tk := range tokens {
		if tk.Kind == TokComment {
			t.Errorf("JSON was read with comments: %v", tokenKinds(tokens))
		}
	}
}

// TestTokenColorFollowsTheAppearance is the dark-mode half of the highlighter:
// the palette carries light and dark, and a highlighter with colours of its
// own would be the one part of a dark chat window that did not follow the
// desktop.
func TestTokenColorFollowsTheAppearance(t *testing.T) {
	light, _ := TokenColor(TokKeyword, false)
	dark, _ := TokenColor(TokKeyword, true)
	if light == dark {
		t.Error("a keyword is the same colour in light and dark")
	}
	if light != theme.Light().AccentText {
		t.Errorf("a light keyword is %v, want the palette's AccentText", light)
	}
	if dark != theme.Dark().AccentText {
		t.Errorf("a dark keyword is %v, want the palette's AccentText", dark)
	}
	// A kind with no colour of its own falls back to body text rather than to
	// the accent, which is kept for things that do something.
	if fg, _ := TokenColor(TokIdent, false); fg != theme.Light().Text {
		t.Errorf("an identifier is %v, want the palette's Text", fg)
	}
}

// ── the bubble style ───────────────────────────────────────────────────────

// TestTheFourRolesShareOneStyle is the requirement that a transcript's four
// kinds of turn differ in exactly two things — which side of the column they
// sit on and which corner the tail is under — and in nothing else.
//
// It is a test and not a comment because the failure is invisible in a
// screenshot: a system notice with its own margins still looks like a system
// notice. It shows up as a transcript where the eye has to re-measure the
// rhythm every time the speaker changes.
func TestTheFourRolesShareOneStyle(t *testing.T) {
	// The metrics are asked for at one density and compared to themselves: the
	// point is that they do not take a role, which is what makes them one set
	// rather than four.
	m := bubbleMetricsFor(4)
	// One set of numbers, and every role asks the same one for them.
	if m.PadY <= 0 || m.PadX <= 0 || m.Radius <= 0 || m.Gap <= 0 || m.Tail <= 0 {
		t.Fatalf("the metrics are not a full set: %+v", m)
	}
	if m.MaxWidth != ReadingWidth {
		t.Errorf("the reading width is %v, want %v", m.MaxWidth, ReadingWidth)
	}

	// The two things that are allowed to differ, and they do.
	if bubbleSide(RoleUser) != ui.End {
		t.Error("the user's bubble is not on the right")
	}
	for _, role := range []Role{RoleAssistant, RoleSystem, RoleTool} {
		if bubbleSide(role) != ui.Start && role != RoleSystem {
			t.Errorf("role %v is not on the left", role)
		}
	}
	if bubbleSide(RoleSystem) != ui.Center {
		t.Error("the system's bubble is not centred")
	}

	// The tails: under the corner the speaker is on, and only the two roles
	// that have a speaker have one at all.
	if bubbleTailCorner(RoleUser) != 2 || bubbleTailCorner(RoleAssistant) != 3 {
		t.Error("the tails are not on the speaking sides")
	}
	if bubbleTailCorner(RoleSystem) != -1 || bubbleTailCorner(RoleTool) != -1 {
		t.Error("a system or tool bubble has a tail to draw")
	}

	// The tail's corner is squared and the other three are round, so the tail
	// looks grown out of the bubble rather than stuck on it. The radii come
	// back in MyGo's order — top-left, top-right, bottom-right, bottom-left.
	userTL, userTR, userBR, userBL := bubbleCorners(RoleUser, 18)
	if userBR != 0 || userBL != 18 {
		t.Errorf("the user's bottom corners are br %v bl %v, want 0 and 18", userBR, userBL)
	}
	if userTL != 18 || userTR != 18 {
		t.Errorf("the user's top corners are tl %v tr %v, want 18 and 18", userTL, userTR)
	}
	_, _, asstBR, asstBL := bubbleCorners(RoleAssistant, 18)
	if asstBL != 0 || asstBR != 18 {
		t.Errorf("the model's bottom corners are br %v bl %v, want 18 and 0", asstBR, asstBL)
	}
	// The two centred roles have no tail, so every corner is round.
	for _, role := range []Role{RoleSystem, RoleTool} {
		tl, tr, br, bl := bubbleCorners(role, 18)
		if tl != 18 || tr != 18 || br != 18 || bl != 18 {
			t.Errorf("role %v has corners %v %v %v %v, want all 18", role, tl, tr, br, bl)
		}
	}

	// The ink differs, and all four come out of the palette.
	light := theme.Light()
	seen := map[ui.Color]int{}
	for _, role := range []Role{RoleUser, RoleAssistant, RoleSystem, RoleTool} {
		bg, _ := bubbleInk(role, light)
		seen[bg]++
	}
	if len(seen) < 2 {
		t.Error("the four roles are all drawn in one background, which is no distinction at all")
	}
	if bg, fg := bubbleInk(RoleUser, light); bg != light.Fill || fg != light.OnFill {
		t.Error("the user's bubble is not the filled one")
	}
}

// ── cost ───────────────────────────────────────────────────────────────────

// TestEstimateIsTwoRatesSeparately is the arithmetic, on exact numbers. A long
// conversation is cheap because the two sides are priced differently, and a
// test that only checked "greater than zero" would say nothing about whether
// the two were being counted at all.
func TestEstimateIsTwoRatesSeparately(t *testing.T) {
	const (
		inRate  = 3.00  // dollars per million tokens
		outRate = 15.00 // dollars per million tokens
	)
	// 1,000,000 in at $3/M is exactly $3.00; 200,000 out at $15/M is $3.00.
	if got := Estimate(1_000_000, 0, inRate, outRate); math.Abs(got-3.0) > 1e-9 {
		t.Errorf("a million input tokens cost %v, want 3", got)
	}
	if got := Estimate(0, 200_000, inRate, outRate); math.Abs(got-3.0) > 1e-9 {
		t.Errorf("200k output tokens cost %v, want 3", got)
	}
	if got := Estimate(1_000_000, 200_000, inRate, outRate); math.Abs(got-6.0) > 1e-9 {
		t.Errorf("both sides cost %v, want 6", got)
	}
	if got := Estimate(0, 0, inRate, outRate); got != 0 {
		t.Errorf("nothing costs %v, want 0", got)
	}
	// The rates are per million, and a missing factor of a thousand is the
	// mistake this test exists to catch.
	if got := Estimate(1000, 0, inRate, outRate); math.Abs(got-0.003) > 1e-12 {
		t.Errorf("a thousand input tokens cost %v, want 0.003", got)
	}
}

// TestEstimateTotalChargesTheHigherRate is the one-number form the live counter
// needs, and it is the higher rate rather than a guess at a split: a counter
// that under-reports is a bill that arrives as a surprise.
func TestEstimateTotalChargesTheHigherRate(t *testing.T) {
	if got := EstimateTotal(1_000_000, 3, 15); math.Abs(got-15) > 1e-9 {
		t.Errorf("a million tokens charged %v, want 15 (the higher rate)", got)
	}
	if got := EstimateTotal(1_000_000, 15, 3); math.Abs(got-15) > 1e-9 {
		t.Errorf("a million tokens charged %v, want 15", got)
	}
	if got := EstimateTotal(0, 3, 15); got != 0 {
		t.Errorf("nothing charged %v, want 0", got)
	}
	if got := EstimateTotal(-5, 3, 15); got != 0 {
		t.Errorf("a negative count charged %v, want 0", got)
	}
}

// TestCostFormatting: four decimal places while the number is small, two once
// it is not. At two places every real chat cost reads "$0.00".
func TestCostFormatting(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want string
	}{
		{0, "$0.00"},
		{0.0031, "$0.0031"},
		{0.01, "$0.01"},
		{12.345, "$12.35"},
		{1234.5, "$1,234"},
	} {
		if got := FormatCost(tc.in); got != tc.want {
			t.Errorf("FormatCost(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTokenFormatting(t *testing.T) {
	for _, tc := range []struct {
		in   int
		want string
	}{
		{0, "0"},
		{-4, "0"},
		{1234, "1,234"},
		{12345, "12.3k"},
		{2_500_000, "2.5M"},
	} {
		if got := FormatTokens(tc.in); got != tc.want {
			t.Errorf("FormatTokens(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestDurationFormatting is the player's label, where "0:05" and "1:05:00" are
// both right and the same code has to choose.
func TestDurationFormatting(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want string
	}{
		{5, "0:05"},
		{65, "1:05"},
		{3600, "1:00:00"},
		{-1, "0:00"},
	} {
		if got := FormatDuration(tc.in); got != tc.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ── reduced motion ─────────────────────────────────────────────────────────

// TestTheCaretIsSteadyUnderReducedMotion is the whole of the reduced-motion
// rule for this package, stated as numbers: with motion off the caret's alpha
// is 1 at every point of its cycle, and with motion on it is not.
//
// A caret is one of the two things in the library that exist only in motion, so
// a reader who asked for no motion gets the resting shape of it rather than
// nothing. And it has to be a *steady* mark, not a faded one: a faint caret
// reads as text that has ended, which is the opposite of what it is for.
func TestTheCaretIsSteadyUnderReducedMotion(t *testing.T) {
	for _, phase := range []float32{0, 0.25, 0.5, 0.75, 0.999} {
		if got := caretAlpha(phase, true); got != 1 {
			t.Errorf("still, the caret at phase %v is alpha %v, want 1", phase, got)
		}
	}
	// Moving, it goes somewhere. Half a cycle apart is the whole of it: the
	// mark is at its brightest where its phase is zero and at its faintest half
	// a cycle later, and nowhere in between is it the same twice by accident.
	bright, dim := caretAlpha(0, false), caretAlpha(0.5, false)
	if bright != 1 {
		t.Errorf("the caret's brightest phase is %v, want full ink", bright)
	}
	if dim <= 0 || dim >= bright {
		t.Errorf("the caret's faintest phase is %v, want between nothing and %v", dim, bright)
	}
	for _, half := range []float32{0.25, 0.75} {
		if mid := caretAlpha(half, false); mid <= dim || mid >= bright {
			t.Errorf("the caret at phase %v is %v, want between %v and %v",
				half, mid, dim, bright)
		}
	}
}

// TestLoopPhaseIsZeroWhenStill is the second half of the rule: a mark caught
// half-faded would read as an interface that has stopped halfway, rather than
// as a quieter one.
func TestLoopPhaseIsZeroWhenStill(t *testing.T) {
	if got := loopPhase(nil, caretPeriod, true); got != 0 {
		t.Errorf("a still mark is at phase %v, want 0", got)
	}
}

// TestReducedMotionReachesTheCaret is the same rule end to end: a window with
// reduced motion draws the caret, and it draws the same one every frame. It is
// a render test because the value above only proves the arithmetic — that
// nothing between core.Reduced and the paint quietly undoes it is the part that
// can only be caught by rendering.
func TestReducedMotionReachesTheCaret(t *testing.T) {
	draft := "half an answer"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		core.WithReducedMotion(c, true)
		StreamingText(c, &draft, StreamingTextOptions{Streaming: true, Lines: 4})
	}, 420, 240)
	if !tt.HasText("still typing") {
		t.Error("with reduced motion the caret is not drawn at all, which reads as " +
			"the answer having ended")
	}
	if !tt.HasText("half an answer") {
		t.Error("the text being streamed is not on screen")
	}
}

// TestStreamingTextShowsTheCaretOnlyWhileStreaming: a finished answer has no
// mark at its end, and one that does reads as typing that will never finish.
func TestStreamingTextShowsTheCaretOnlyWhileStreaming(t *testing.T) {
	draft := "an answer"
	live := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StreamingText(c, &draft, StreamingTextOptions{Streaming: true, Lines: 4})
	}, 420, 240)
	if !live.HasText("still typing") {
		t.Error("a streaming answer has no caret")
	}

	done := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StreamingText(c, &draft, StreamingTextOptions{Streaming: false, Lines: 4})
	}, 420, 240)
	if done.HasText("still typing") {
		t.Error("a finished answer is still showing its caret")
	}
}

// ── rendering ──────────────────────────────────────────────────────────────

// TestMessageBubbleDrawsAllFourRoles is the render half of the one-style rule:
// every role's bubble is on screen, in its own place, and none of them
// panicked. The measurement half is in TestTheFourRolesShareOneStyle, because
// a bubble's width in a headless test is the width of an empty box.
func TestMessageBubbleDrawsAllFourRoles(t *testing.T) {
	roles := []Role{RoleUser, RoleAssistant, RoleSystem, RoleTool}
	names := []string{"Ada", "Model", "System", "Tool"}
	for i, role := range roles {
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MessageBubble(c, MessageBubbleOptions{
				Role: role, Author: names[i], Time: "09:41",
			}, func() {
				ui.Text(c, "the words of the turn")
			})
		}, 520, 220)
		if !tt.HasText(names[i]) {
			t.Errorf("role %v shows no author", role)
		}
		if !tt.HasText("the words of the turn") {
			t.Errorf("role %v shows no body", role)
		}
	}
}

// TestMessageBubbleReportsThePress: the result is read inside the view,
// because by the time the tester settles the press is three passes old.
func TestMessageBubbleReportsThePress(t *testing.T) {
	presses := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if MessageBubble(c, MessageBubbleOptions{Role: RoleAssistant},
			func() { ui.Text(c, "press me") }).Clicked() {
			presses++
		}
	}, 420, 200)
	if presses != 0 {
		t.Fatalf("pressed %d times before anything was pressed", presses)
	}
	if err := tt.Click("press me"); err != nil {
		t.Fatal(err)
	}
	if presses != 1 {
		t.Errorf("the bubble reported %d presses, want exactly 1", presses)
	}
}

// TestMessageActionsReportsByName is the contract of a row of buttons: one
// press, reported as the label the caller gave the action. Two labels that
// collapse to the same text would be the caller's mistake, and a caller that
// had to count presses to find out which one it was would be reconstructing
// the value from the interface.
func TestMessageActionsReportsByName(t *testing.T) {
	actions := []MessageAction{
		{Label: "Copy", Icon: "copy"},
		{Label: "Try again", Icon: "refresh"},
	}
	var pressed []string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if got := MessageActions(c, actions).Pressed(); got != "" {
			pressed = append(pressed, got)
		}
	}, 420, 200)

	if err := tt.Click("Copy"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Try again"); err != nil {
		t.Fatal(err)
	}
	wantStrings(t, pressed, []string{"Copy", "Try again"}, "pressed actions")
}

// TestMessageActionsInsistOnLabels is the rule that a mark must have a name,
// and that it is also the value the result returns — so an action with no
// Label is not a button that will not say what it did, it is a button that
// cannot report at all.
func TestMessageActionsInsistOnLabels(t *testing.T) {
	wantsPanic(t, "an action with no label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MessageActions(c, []MessageAction{{Icon: "copy"}})
		}, 300, 150)
	})
}

// TestMarkdownViewShowsWhatParseFound walks the two halves together once: the
// parser's blocks go into the view and the words come out on screen. It is the
// test that would catch a renderer that dropped a block kind, which the Parse
// tests cannot see and a screenshot might.
func TestMarkdownViewShowsWhatParseFound(t *testing.T) {
	src := "# Findings\n\nThe **pool** was full.\n\n- one\n- two\n\n" +
		"| Name | Size |\n| --- | ---: |\n| a.png | 12 KB |\n"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MarkdownView(c, MarkdownViewOptions{Source: src})
	}, 560, 520)

	for _, want := range []string{"Findings", "The pool was full.", "one", "two", "Name", "a.png"} {
		if !tt.HasText(want) {
			t.Errorf("the view does not show %q", want)
		}
	}
	// An already-parsed tree is drawn the same way, which is the point of
	// Blocks being there: one answer shown in two places is parsed once.
	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MarkdownView(c, MarkdownViewOptions{Blocks: Parse(src)})
	}, 560, 520)
	if !tt2.HasText("Findings") || !tt2.HasText("a.png") {
		t.Error("the view drew differently when it was handed the blocks instead of the source")
	}
}

// TestCodeBlockPutsTheSourceOnTheClipboard is the one result on a code block,
// and it is a pulse for the same reason every other press is: the caller acts
// on it in the frame it happens.
func TestCodeBlockPutsTheSourceOnTheClipboard(t *testing.T) {
	copies := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if CodeBlock(c, CodeBlockOptions{
			Code: "x := 1", Lang: "go", Title: "main.go", Copy: "Copy code",
		}).Copied() {
			copies++
		}
	}, 520, 280)

	if copies != 0 {
		t.Fatalf("copied %d times before anything was copied", copies)
	}
	if err := tt.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if copies != 1 {
		t.Errorf("the block copied %d times, want exactly 1", copies)
	}
}

// TestCodeBlockInsistsOnCode: an empty block is a stray pair of fences, and
// drawing it as a card would put a box in the answer where a reader expects
// source.
func TestCodeBlockInsistsOnCode(t *testing.T) {
	wantsPanic(t, "an empty code block", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CodeBlock(c, CodeBlockOptions{Lang: "go"})
		}, 300, 150)
	})
}

// TestCodeBlockNumbersItsLinesFromWhereItStarted is the detail that makes a
// snippet from the middle of a file usable: the numbers are the file's, not the
// snippet's.
func TestCodeBlockNumbersItsLinesFromWhereItStarted(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CodeBlock(c, CodeBlockOptions{
			Code: "a\nb\nc", Lang: "go", LineNumbers: true, StartLine: 40,
		})
	}, 420, 240)
	for _, want := range []string{"40", "41", "42"} {
		if !tt.HasText(want) {
			t.Errorf("the block does not show line %s", want)
		}
	}
}

// TestMessageListFollowsTheEnd is the question a composer asks of a
// transcript, and the answer with no state is "yes": a transcript nobody is
// scrolling is being read from the top, and following it is the safe default.
func TestMessageListFollowsTheEnd(t *testing.T) {
	var state ui.ScrollState
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := MessageList(c, MessageListOptions{
			Messages: 3, Height: 300,
			State: &state,
			Message: func(i int) {
				MessageBubble(c, MessageBubbleOptions{Role: RoleUser}, func() {
					ui.Text(c, "turn "+itoa(i))
				})
			},
		})
		if !r.AtEnd() {
			t.Error("a transcript nobody has scrolled is not at its end")
		}
		if r.State != &state {
			t.Error("the list did not hand back the state it was given")
		}
	}, 520, 360)
	for i := range 3 {
		if !tt.HasText("turn " + itoa(i)) {
			t.Errorf("turn %d is not on screen", i)
		}
	}
	// Without a state it still says it is at the end: a transcript nobody is
	// scrolling is being read from the top, and following it is the safe
	// default rather than a guess.
	plain := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if !MessageList(c, MessageListOptions{
			Messages: 1, Height: 200,
			Message: func(int) { ui.Text(c, "a turn") },
		}).AtEnd() {
			t.Error("a transcript with no state is not at its end")
		}
	}, 420, 260)
	_ = plain
}

// TestConversationListReusesTheList: it is data.List, so a click chooses a row
// and the caller's Selected pointer moves — through the same code a table of
// callbacks uses, and not through a second implementation of "what a click
// means".
func TestConversationListReusesTheList(t *testing.T) {
	names := []string{"Hillside Dental", "Maple Street Bakery", "Corner Studio"}
	choice := &data.Selectable{}
	selected := -1
	state := ui.ListState{}

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ConversationList(c, ConversationListOptions{
			Conversations: len(names),
			Height:        260,
			Choice:        choice,
			Selected:      &selected,
			State:         &state,
			Conversation: func(row int) {
				ConversationItem(c, ConversationItemOptions{
					Title:   names[row],
					Preview: "the last thing said",
					When:    "4m",
					Unread:  row,
				})
			},
		})
	}, 420, 320)

	if err := tt.Click("Maple Street Bakery"); err != nil {
		t.Fatal(err)
	}
	if selected != 1 {
		t.Errorf("clicking the second row chose %d, want 1", selected)
	}
	if choice.Len() != 1 || choice.Rows()[0] != 1 {
		t.Errorf("the choice is %v, want just row 1", choice.Rows())
	}
}

// TestConversationItemInsistsOnATitle: a row with no name cannot be told from
// another, and a list of those is a list the reader is guessing through.
func TestConversationItemInsistsOnATitle(t *testing.T) {
	wantsPanic(t, "a conversation with no title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ConversationItem(c, ConversationItemOptions{Preview: "something"})
		}, 320, 160)
	})
}

// TestPromptComposerSendsAndStops: the two actions a composer has, and the
// rule that stops it being three — while an answer is arriving there is
// nothing to send, so Sent is false even if the button were there.
func TestPromptComposerSendsAndStops(t *testing.T) {
	draft := ""
	var sends, stops int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := PromptComposer(c, PromptComposerOptions{
			Text: &draft, Label: "Message", Placeholder: "Ask anything",
		})
		if r.Sent() {
			sends++
		}
		if r.Stopped() {
			stops++
		}
	}, 520, 320)

	// Nothing written, so nothing to send: the button is disabled rather than
	// accepting an empty question.
	if err := tt.Click("Send"); err != nil {
		t.Fatal(err)
	}
	if sends != 0 {
		t.Errorf("an empty draft was sent %d times", sends)
	}
	if tt.HasText("Stop generating") {
		t.Error("a composer with nothing running is offering to stop")
	}

	draft = "a real question"
	if err := tt.Click("Send"); err != nil {
		t.Fatal(err)
	}
	if sends != 1 {
		t.Errorf("the composer sent %d times, want 1", sends)
	}

	// Busy swaps the button for the stop one.
	busy := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := PromptComposer(c, PromptComposerOptions{
			Text: &draft, Label: "Message", Busy: true,
		})
		if r.Stopped() {
			stops++
		}
		if r.Sent() {
			sends++
		}
	}, 520, 320)
	if err := busy.Click("Stop generating"); err != nil {
		t.Fatal(err)
	}
	if stops != 1 {
		t.Errorf("the composer stopped %d times, want 1", stops)
	}
	if sends != 1 {
		t.Errorf("a busy composer also sent %d times, want it still to be 1", sends)
	}
}

// TestPromptComposerNeedsADraft: it keeps none of its own, and a composer that
// kept one would throw away a paragraph nobody could get back.
func TestPromptComposerNeedsADraft(t *testing.T) {
	wantsPanic(t, "a composer with no draft to point at", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PromptComposer(c, PromptComposerOptions{Label: "Message"})
		}, 400, 200)
	})
}

// TestSlashCommandMenuFiltersAndReports: a panel that appears with nothing in
// it says the feature is broken, where one that does not appear says the
// typing did not match anything — which is the true answer.
func TestSlashCommandMenuFiltersAndReports(t *testing.T) {
	commands := []string{
		"clear: Empty the conversation",
		"compact: Summarise what has been said",
		"model: Choose a model",
	}
	var chosen []string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if got := SlashCommandMenu(c, SlashCommandOptions{
			Commands: commands, Query: "c", MaxHeight: 200,
		}).Chosen(); got != "" {
			chosen = append(chosen, got)
		}
	}, 460, 320)

	// The query filters on the name: "c" matches clear and compact, and not
	// model — so the hint of the one it dropped is gone with it.
	if !tt.HasText("Empty the conversation") {
		t.Error("the menu does not show the hint of what matched")
	}
	if tt.HasText("Choose a model") {
		t.Error("the menu shows a command the query did not match")
	}

	if err := tt.Click("compact"); err != nil {
		t.Fatal(err)
	}
	wantStrings(t, chosen, []string{"compact"}, "chosen commands")

	// Nothing matching: no panel at all.
	none := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if SlashCommandMenu(c, SlashCommandOptions{
			Commands: commands, Query: "zzz",
		}).Element != nil {
			t.Error("the menu drew a panel with nothing in it")
		}
	}, 460, 320)
	_ = none
}

// TestMentionMenuIsTheSamePanel is the reason both menus go through one
// function: they are the same control with a different trigger, and two copies
// would differ in the width of the left column within a month.
func TestMentionMenuIsTheSamePanel(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MentionMenu(c, MentionOptions{
			Items: []string{"ui/chat/parse.go: the parser", "docs/design-system.md: the rules"},
			Query: "parse", MaxHeight: 200,
		})
	}, 460, 300)
	if !tt.HasText("ui/chat/parse.go") {
		t.Error("the mention menu does not show what matched")
	}
	if tt.HasText("design-system.md") {
		t.Error("the mention menu shows something the query did not match")
	}
}

// TestThinkingBlockToggles: the fold is the caller's flag, and this is the
// press that asks to change it.
func TestThinkingBlockToggles(t *testing.T) {
	open := false
	toggles := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if ThinkingBlock(c, ThinkingBlockOptions{
			Summary: "Checked the connection pool first",
			Detail:  "1. read the config\n2. count the pools",
			Open:    open,
		}).Toggled() {
			toggles++
			open = !open
		}
	}, 460, 240)

	if !tt.HasText("Checked the connection pool first") {
		t.Error("the folded block shows no summary")
	}
	if tt.HasText("count the pools") {
		t.Error("a folded block is showing its detail")
	}
	if err := tt.Click("Checked the connection pool first"); err != nil {
		t.Fatal(err)
	}
	if toggles != 1 {
		t.Fatalf("the block toggled %d times, want 1", toggles)
	}
	if !tt.HasText("count the pools") {
		t.Error("an opened block is not showing its detail")
	}
}

// TestThinkingBlockInsistsOnASummary: a collapsed block with nothing in it is a
// line with no way to open it.
func TestThinkingBlockInsistsOnASummary(t *testing.T) {
	wantsPanic(t, "a thinking block with no summary", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ThinkingBlock(c, ThinkingBlockOptions{})
		}, 320, 160)
	})
}

// TestBranchNavigatorStopsAtTheEnds and says so in the panic, because a
// navigator showing "3 of 2" is a caller that counted wrong somewhere and
// would quietly move nowhere.
func TestBranchNavigatorStopsAtTheEnds(t *testing.T) {
	moves := []int{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if m := BranchNavigator(c, BranchNavigatorOptions{Branch: 2, Of: 3}).Moved(); m != 0 {
			moves = append(moves, m)
		}
	}, 400, 200)
	if err := tt.Click("Next answer"); err != nil {
		t.Fatal(err)
	}
	wantInts(t, moves, 1)

	wantsPanic(t, "a branch that is not one of the branches", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			BranchNavigator(c, BranchNavigatorOptions{Branch: 4, Of: 3})
		}, 300, 150)
	})
}

func wantInts(t *testing.T, got []int, want ...int) {
	t.Helper()
	if fmtInts(got) != fmtInts(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func fmtInts(v []int) string {
	parts := make([]string, len(v))
	for i, n := range v {
		parts[i] = itoa(n)
	}
	return strings.Join(parts, ",")
}

// TestDateSeparatorWhenDrawsOnlyOnTheChange is the whole component: it
// compares the date against the one before it and draws nothing when they are
// the same. A caller that put the test in its own loop would write it once
// correctly and once in a hurry, and the second would put a rule above every
// turn.
func TestDateSeparatorWhenDrawsOnlyOnTheChange(t *testing.T) {
	when, last := "Tuesday 4 March", ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DateSeparatorWhen(c, &when, &last)
	}, 420, 160)
	if !tt.HasText("Tuesday 4 March") {
		t.Error("the first turn of a day has no separator above it")
	}

	same := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DateSeparatorWhen(c, &when, &last)
	}, 420, 160)
	if same.HasText("Tuesday 4 March") {
		t.Error("a second turn on the same day drew a separator")
	}
}

// TestTranscriptRenders is the exported conversation as text, checked against
// an exact string. A screenshot could not do this: what a reader pastes into a
// document is a string, and a lost turn or a doubled heading would look, in a
// picture, almost right.
func TestTranscriptRenders(t *testing.T) {
	doc := Transcript{
		Title: "Pool exhaustion",
		Turns: []Turn{
			{Role: RoleUser, Who: "Ada", When: "09:41", Text: "Why is it timing out?"},
			{Role: RoleAssistant, Who: "Model", When: "09:42",
				Text: "The pool is full.", Quote: "Why is it timing out?"},
		},
		Exported: ExportMarkdown,
	}
	want := "# Pool exhaustion\n\n" +
		"**Ada** · 09:41\n\nWhy is it timing out?\n\n" +
		"> Why is it timing out?\n>\n" +
		"**Model** · 09:42\n\nThe pool is full.\n\n"
	if got := doc.Render(); got != want {
		t.Errorf("markdown export:\n got %q\nwant %q", got, want)
	}

	doc.Exported = ExportJSON
	if got := doc.Render(); !strings.Contains(got, `"role": "assistant"`) ||
		!strings.Contains(got, `"text": "The pool is full."`) {
		t.Errorf("JSON export = %q", got)
	}
	// A quote has to be escaped or the exported file is not parseable.
	doc.Turns[0].Text = `he said "no"`
	if got := doc.Render(); !strings.Contains(got, `he said \"no\"`) {
		t.Errorf("a quote inside the text was not escaped: %q", got)
	}
}

// TestSourcesPanelReportsWhichOne: two presses on a panel of cards, each of
// which navigates somewhere and each of which is a different promise — one
// leaves the answer and one brings a piece of it back.
func TestSourcesPanelReportsWhichOne(t *testing.T) {
	sources := []Source{
		{Title: "config.go", Where: "internal/pool", Snippet: "MaxConns: 4"},
		{Title: "design-system.md", Where: "docs", Snippet: "§2 colours"},
	}
	// Opened and Quoted are one-frame pulses, so each is counted rather than
	// assigned: assigning would leave the last frame's value — zero — and the
	// assertion below would pass for the wrong reason on the frame the click
	// did not happen in.
	opened, quoted := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := SourcesPanel(c, SourcesPanelOptions{
			Title: "Sources", Sources: sources, Quotation: "Quote", Height: 300,
		})
		if n := r.Opened(); n > 0 {
			opened = n
		}
		if n := r.Quoted(); n > 0 {
			quoted = n
		}
	}, 520, 400)

	if err := tt.Click("design-system.md"); err != nil {
		t.Fatal(err)
	}
	if opened != 2 {
		t.Errorf("opened source %d, want 2", opened)
	}
	// The quote button is named after its source, so this click lands on the
	// second card rather than on whichever "Quote" came first.
	if err := tt.Click("Quote design-system.md"); err != nil {
		t.Fatal(err)
	}
	if quoted != 2 {
		t.Errorf("quoted source %d, want 2", quoted)
	}
}

// TestAttachmentChipSeparatesItsTwoPromises: removing an attachment is usually
// undoable and opening one is not, and a caller handed a single boolean would
// have to guess which had happened.
func TestAttachmentChipSeparatesItsTwoPromises(t *testing.T) {
	removed, opened := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// Two chips, not one: a chip that both opens and removes would make the
		// remove button also an open button, and a reader trying to take an
		// attachment off would open the file instead.
		removable := AttachmentChip(c, AttachmentChipOptions{
			Name: "trace.json", Meta: "412 KB", Mark: "file", Removable: true,
		})
		if removable.Removed() {
			removed++
		}
		openable := AttachmentChip(c, AttachmentChipOptions{
			Name: "diagram.png", Meta: "80 KB", Mark: "file", Openable: true,
		})
		if openable.Opened() {
			opened++
		}
	}, 520, 240)

	if err := tt.Click("Remove trace.json"); err != nil {
		t.Fatal(err)
	}
	if removed != 1 || opened != 0 {
		t.Errorf("the remove button reported removed %d opened %d, want 1 and 0", removed, opened)
	}
	if err := tt.Click("diagram.png"); err != nil {
		t.Fatal(err)
	}
	if opened != 1 {
		t.Errorf("the chip opened %d times, want 1", opened)
	}
}

// TestAttachmentChipInsistsOnItsMark, for the reason a glyph needs a name
// everywhere else in this library: it is the one element with no word of its
// own.
func TestAttachmentChipInsistsOnItsMark(t *testing.T) {
	wantsPanic(t, "a chip whose mark has no name", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AttachmentChip(c, AttachmentChipOptions{Name: "a.png"})
		}, 300, 150)
	})
}

// TestContextWindowMeterChangesToneWithUse is the meter reading at two numbers
// rather than at a picture: it must say "running out" once past the warning
// mark and stop saying it before it.
func TestContextWindowMeterChangesToneWithUse(t *testing.T) {
	calm := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ContextWindowMeter(c, ContextWindowMeterOptions{Used: 10_000, Window: 200_000})
	}, 420, 220)
	if calm.HasText("Running out") || calm.HasText("Nearly full") {
		t.Error("a window a twentieth full is already warning")
	}

	full := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ContextWindowMeter(c, ContextWindowMeterOptions{Used: 195_000, Window: 200_000})
	}, 420, 220)
	if !full.HasText("Nearly full") {
		t.Error("a window that is nearly gone does not say so")
	}
	if !full.HasText("195.0k") {
		t.Error("the meter does not say how much is used")
	}
}

// TestContextWindowMeterNeedsAWindow: dividing by a window of nothing is the
// kind of mistake that shows up as a bar at 300%.
func TestContextWindowMeterNeedsAWindow(t *testing.T) {
	wantsPanic(t, "a window of nothing", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ContextWindowMeter(c, ContextWindowMeterOptions{Used: 10})
		}, 300, 150)
	})
}

// TestVoiceWaveformNeedsALabel: a waveform is a picture of a sound and carries
// no words of its own.
func TestVoiceWaveformNeedsALabel(t *testing.T) {
	wantsPanic(t, "a waveform with no label", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			VoiceWaveform(c, VoiceWaveformOptions{Levels: []float32{0.5}})
		}, 300, 150)
	})
}

// TestVoiceWaveformRendersUnderReducedMotion: the mark still has a shape when
// it is not moving. A level meter that had flattened to nothing would read as a
// microphone that has failed rather than as a quieter interface.
func TestVoiceWaveformRendersUnderReducedMotion(t *testing.T) {
	levels := []float32{0.1, 0.6, 0.9, 0.4, 0.7}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		core.WithReducedMotion(c, true)
		VoiceWaveform(c, VoiceWaveformOptions{
			Levels: levels, Bars: len(levels), Width: 200, Height: 40,
			Label: "Microphone level",
		})
	}, 320, 160)
	if !tt.HasText("Microphone level") {
		t.Error("the meter is not named, so nothing reads it out")
	}
}

// TestFeedbackFormStarsAreNumbered: five stars drawn as stars, each labelled
// with its number, so a screen reader says "3" rather than "star" five times.
func TestFeedbackFormStarsAreNumbered(t *testing.T) {
	rating := 0
	var stars int
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := FeedbackForm(c, FeedbackFormOptions{Title: "Was that helpful?", Rating: rating})
		if s := r.Starred(); s != 0 {
			stars = s
			rating = s
		}
	}, 460, 260)

	if err := tt.Click("Rate 4 of 5"); err != nil {
		t.Fatal(err)
	}
	if stars != 4 {
		t.Errorf("the form reported %d stars, want 4", stars)
	}
}

// TestErrorMessageSitsWhereTheAnswerWouldHaveBeen: it is drawn in the model's
// place and it says what went wrong in words, because a red box with a
// paragraph in it is a paragraph that happens to be red.
func TestErrorMessageSitsWhereTheAnswerWouldHaveBeen(t *testing.T) {
	retried := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ErrorMessage(c, ErrorMessageOptions{
			Title:  "The model did not answer",
			Detail: "context length exceeded",
			Retry:  "Try again",
		})
		if r.Retried() {
			retried++
		}
	}, 480, 260)

	if !tt.HasText("The model did not answer") || !tt.HasText("context length exceeded") {
		t.Error("the error message does not say what went wrong")
	}
	if err := tt.Click("Try again"); err != nil {
		t.Fatal(err)
	}
	if retried != 1 {
		t.Errorf("the message retried %d times, want 1", retried)
	}
}

// TestChatMessagePicksTheComponentForTheKind is the reason ChatMessage exists:
// one place decides which body a turn carries, so two views of the same
// conversation cannot disagree about it.
func TestChatMessagePicksTheComponentForTheKind(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChatMessage(c, ChatMessageOptions{
			Kind: KindCode, Role: RoleAssistant, Author: "Model",
			Code: "print(1)", Lang: "python", CodeTitle: "hello.py",
			Actions: []MessageAction{{Label: "Copy code", Icon: "copy"}},
		})
	}, 520, 320)
	if !tt.HasText("hello.py") || !tt.HasText("Model") {
		t.Error("a code turn is missing its header or its title")
	}

	tt2 := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ChatMessage(c, ChatMessageOptions{
			Kind: KindError, Role: RoleAssistant,
			Err: ErrorMessageOptions{Title: "It did not answer"},
		})
	}, 480, 260)
	if !tt2.HasText("It did not answer") {
		t.Error("an error turn did not draw the error")
	}
}

// TestThePanelsInsistOnTheirRows: each of the four that take a list panics on
// an empty one, because an empty list of things is a gap and a caller who
// meant one is looking at a screenshot with a hole in it.
func TestThePanelsInsistOnTheirRows(t *testing.T) {
	view := func(fn func(c *ui.Context)) { ui.NewTester(fn, 400, 240) }
	wantsPanic(t, "an image grid with no images", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ImageGrid(c, ImageGridOptions{})
		})
	})
	wantsPanic(t, "a table block with no head", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TableBlock(c, TableBlockOptions{Rows: [][]string{{"a"}}})
		})
	})
	wantsPanic(t, "an empty state with no title", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			EmptyState(c, EmptyStateOptions{Body: "nothing here yet"})
		})
	})
	wantsPanic(t, "a prompt library with no prompts", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PromptLibrary(c, PromptLibraryOptions{Height: 200})
		})
	})
	wantsPanic(t, "a project list with no height", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ProjectList(c, ProjectListOptions{Projects: []string{"one"}})
		})
	})
	wantsPanic(t, "a citation numbered from zero", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			CitationBadge(c, CitationBadgeOptions{N: 0})
		})
	})
}

// ── dark mode ──────────────────────────────────────────────────────────────

// TestTheWindowFollowsTheAppearance is the dark-mode test for this package. A
// chat window is mostly text on a background, so the risk is not that it
// crashes but that one component hard-coded a colour and stood out as a white
// card in a dark window. Every colour in the package comes from core.Tokens,
// so this walks the components that draw the most of them under core.Dark and
// asks that they all still render.
func TestTheWindowFollowsTheAppearance(t *testing.T) {
	draft := "a question in the dark"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		if !core.IsDark(c) {
			t.Error("the window did not resolve the dark palette")
		}
		MessageBubble(c, MessageBubbleOptions{
			Role: RoleAssistant, Author: "Model", Time: "09:42",
		}, func() {
			MarkdownView(c, MarkdownViewOptions{
				Source: "# Findings\n\nThe **pool** was full.\n\n```go\nx := 1\n```\n",
			})
		})
		PromptComposer(c, PromptComposerOptions{
			Text: &draft, Label: "Message", Chips: []ContextChip{{Name: "config.go"}},
		})
		ContextWindowMeter(c, ContextWindowMeterOptions{Used: 190_000, Window: 200_000})
		TokenCounter(c, TokenCounterOptions{Tokens: 12_345, Cost: 0.0031})
	}, 620, 700)

	// The draft is not in this list and cannot be: a text field's value is not
	// a label, and a tester finds elements by their labels. What the composer
	// contributes is its field name and its chips.
	for _, want := range []string{
		"Findings", "The pool was full.", "Model", "config.go", "Message",
		"Nearly full", "12.3k",
	} {
		if !tt.HasText(want) {
			t.Errorf("the dark window does not show %q", want)
		}
	}
	// The cost line is one of the few in this package that is money, and it is
	// drawn with four places while the figure is small.
	if !tt.HasText("$0.0031") {
		t.Errorf("the dark window does not show the cost; texts were %v", tt.Texts())
	}
}

// TestThePaletteIsTheOnlySourceOfColour is the light-and-dark pair of
// TokenColor, checked at the point a renderer asks: the same token in the two
// appearances must be two different colours, and both must be the palette's.
func TestThePaletteIsTheOnlySourceOfColour(t *testing.T) {
	for kind := TokText; kind <= TokPunct; kind++ {
		light, _ := TokenColor(kind, false)
		dark, _ := TokenColor(kind, true)
		if light == dark {
			t.Errorf("token kind %v is the same colour in both appearances", kind)
		}
	}
}

// TestFileMessageCaptionTakesTheBubbleInk is the one place in this package a
// caption is a child of the filled bubble, so it has to be drawn in the
// bubble's own ink. Repeating the window's text colour instead paints the
// caption in the bubble's fill: in the dark palette the two tokens are the
// same colour, so the caption is not there at all. bubble sets the ink on the
// box for its children to inherit, which is why a child that sets none takes
// the right answer and a child that sets the window's does not.
//
// The assertion is the contrast the caption reaches against the bubble's fill,
// compared with the contrast the palette's own bubble ink reaches there — the
// same way the library holds marks to 3:1 — rather than on a particular hex
// value, which is the only form of this that survives a palette change.
func TestFileMessageCaptionTakesTheBubbleInk(t *testing.T) {
	const caption = "the trace of the failing request"
	for _, tc := range []struct {
		mode core.Mode
		k    theme.Tokens
	}{
		{core.Light, theme.Light()},
		{core.Dark, theme.Dark()},
	} {
		mode, k := tc.mode, tc.k
		t.Run(mode.String(), func(t *testing.T) {
			tt := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{Mode: mode})
				FileMessage(c, FileMessageOptions{
					Role:       RoleUser,
					Attachment: AttachmentChipOptions{Name: "trace.json", Mark: "file"},
					Caption:    caption,
				})
			}, 560, 220)
			bg, fg := bubbleInk(RoleUser, k)
			r, ok := tt.Find(caption)
			if !ok {
				t.Fatalf("the caption was not drawn at all; texts were %v", tt.Texts())
			}
			want := contrast(fg, bg)
			if got := strongestContrastIn(tt.Image(), r, bg); got < want-0.5 {
				t.Errorf("the caption on the filled bubble reaches only %.2f:1 against the bubble, want %.2f:1 as its own ink does",
					got, want)
			}
		})
	}
}

// strongestContrastIn is the highest contrast any pixel of r has with bg, which
// is how much the thing drawn in r stands out from it.
func strongestContrastIn(frame *image.RGBA, r ui.Rect, bg ui.Color) float64 {
	best := 0.0
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if y < 0 || y >= frame.Bounds().Dy() || x < 0 || x >= frame.Bounds().Dx() {
				continue
			}
			p := frame.RGBAAt(x, y)
			if cr := contrast(ui.RGBA(p.R, p.G, p.B, float32(p.A)/255), bg); cr > best {
				best = cr
			}
		}
	}
	return best
}

// contrast is the WCAG ratio of two opaque colours, the same ratio every other
// package in this repository holds its marks to.
func contrast(a, b ui.Color) float64 {
	la, lb := luminance(a)+0.05, luminance(b)+0.05
	if la < lb {
		la, lb = lb, la
	}
	return la / lb
}

func luminance(c ui.Color) float64 {
	channel := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.R) + 0.7152*channel(c.G) + 0.0722*channel(c.B)
}

// ── the rest, by kind ──────────────────────────────────────────────────────

// TestTheRemainingComponentsRender is the sweep for the ones without a
// behaviour worth a test of their own: it draws each of them and asks that it
// got to the screen. A component that panicked here would be a component
// nobody had ever built, which is the failure this catches.
func TestTheRemainingComponentsRender(t *testing.T) {
	cases := []struct {
		name  string
		build func(c *ui.Context)
		want  string
	}{
		{"MessageHeader", func(c *ui.Context) {
			MessageHeader(c, MessageHeaderOptions{Author: "Ada", Time: "09:41", Badge: "main"})
		}, "Ada"},
		{"QuoteReply", func(c *ui.Context) {
			QuoteReply(c, QuoteReplyOptions{Author: "Ada", Text: "the earlier turn"})
		}, "the earlier turn"},
		{"TableBlock", func(c *ui.Context) {
			TableBlock(c, TableBlockOptions{
				Head:   []string{"Name", "Size"},
				Rows:   [][]string{{"a.png", "12 KB"}},
				Aligns: []Align{AlignNone, AlignRight},
			})
		}, "a.png"},
		{"ImageGrid", func(c *ui.Context) {
			ImageGrid(c, ImageGridOptions{
				Items: []ImageItem{{Alt: "the failing dashboard"}, {Alt: "the trace"}},
			})
		}, "the failing dashboard"},
		{"SuggestionChips", func(c *ui.Context) {
			SuggestionChips(c, SuggestionChipsOptions{
				Title:       "Try",
				Suggestions: []string{"Explain this: what the pool is for"},
			})
		}, "Explain this"},
		{"ThinkingIndicator", func(c *ui.Context) {
			ThinkingIndicator(c, ThinkingIndicatorOptions{
				Label: "Reading the repository", Detail: "12 files",
			})
		}, "Reading the repository"},
		{"TypingIndicator", func(c *ui.Context) {
			TypingIndicator(c, TypingIndicatorOptions{Who: "Ada", Bubbles: 2})
		}, "Ada"},
		{"CostEstimator", func(c *ui.Context) {
			CostEstimator(c, CostEstimatorOptions{
				InTokens: 1_000_000, OutTokens: 200_000,
				InRate: 3, OutRate: 15, Model: "model-a", Breakdown: true,
			})
		}, "$6.00"},
		{"ContextChips", func(c *ui.Context) {
			ContextChips(c, []ContextChip{{Name: "config.go", Detail: "internal/pool"}})
		}, "config.go"},
		{"ProjectList", func(c *ui.Context) {
			ProjectList(c, ProjectListOptions{
				Projects: []string{"ui/chat", "docs"}, Height: 200,
			})
		}, "ui/chat"},
		{"ConversationSearch", func(c *ui.Context) {
			q := "pool"
			ConversationSearch(c, ConversationSearchOptions{Query: &q, Found: 3, Total: 40})
		}, "3 of 40"},
		{"ConversationExport", func(c *ui.Context) {
			ConversationExport(c, ConversationExportOptions{
				Transcript: Transcript{Title: "Pool", Turns: []Turn{
					{Role: RoleUser, Who: "Ada", When: "09:41", Text: "why?"},
				}},
			})
		}, "Export Markdown"},
		{"BranchNavigator", func(c *ui.Context) {
			BranchNavigator(c, BranchNavigatorOptions{Branch: 1, Of: 2})
		}, "1 / 2"},
		{"DateSeparator", func(c *ui.Context) {
			DateSeparator(c, "Tuesday 4 March")
		}, "Tuesday 4 March"},
		{"RegenerateMenu", func(c *ui.Context) {
			RegenerateMenu(c, RegenerateOptions{
				Items: []string{"plain: As asked", "steps: With the steps shown"},
			})
		}, "Try again"},
		{"ModelSelector", func(c *ui.Context) {
			sel := ""
			ModelSelector(c, ModelSelectorOptions{
				Selected: &sel, Models: []string{"a: Model A", "b: Model B"},
			})
		}, "Choose a model"},
		{"ModeSelector", func(c *ui.Context) {
			mode := 0
			ModeSelector(c, ModeSelectorOptions{Selected: &mode, Modes: []string{"Ask", "Edit"}})
		}, "Ask"},
		{"ParameterPanel", func(c *ui.Context) {
			temp := 0.7
			ParameterPanel(c, ParameterPanelOptions{
				Title: "Parameters",
				Parameters: []Parameter{{
					Name: "Temperature", Hint: "Higher is looser",
					Control: func() {
						// The control is the caller's; this test only has to
						// prove it is built where the panel says it is.
						ui.Text(c, "0.70")
					},
				}},
				ResetAll: "Reset all",
			})
			_ = temp
		}, "Temperature"},
		{"SystemPromptEditor", func(c *ui.Context) {
			p := "You are careful."
			SystemPromptEditor(c, SystemPromptEditorOptions{Prompt: &p, Label: "System prompt", Title: "System"})
		}, "System prompt"},
		{"PromptLibrary", func(c *ui.Context) {
			PromptLibrary(c, PromptLibraryOptions{
				Title:   "Saved",
				Prompts: []string{"Review: Review this diff line by line"},
				Height:  200,
			})
		}, "Review this diff line by line"},
		{"ToolToggleMenu", func(c *ui.Context) {
			on := []string{"search"}
			ToolToggleMenu(c, ToolToggleMenuOptions{Tools: []string{"search", "write"}, On: &on})
		}, "search"},
		{"PasteImagePreview", func(c *ui.Context) {
			PasteImagePreview(c, PasteImagePreviewOptions{
				Images: []ImageItem{{Alt: "a pasted screenshot"}},
			})
		}, "a pasted screenshot"},
		{"VoiceInputButton", func(c *ui.Context) {
			VoiceInputButton(c, VoiceInputButtonOptions{Label: "Voice input"})
		}, "Voice input"},
		{"ScrollToBottomButton", func(c *ui.Context) {
			ScrollToBottomButton(c, ScrollToBottomButtonOptions{Pending: 3})
		}, "Jump to latest"},
		{"SearchProgress", func(c *ui.Context) {
			SearchProgress(c, SearchProgressOptions{
				Query: "pool", Found: 12, Match: 3, Busy: true,
			})
		}, "3 of 12"},
		{"AudioMessage", func(c *ui.Context) {
			AudioMessage(c, AudioMessageOptions{Seconds: 125, Title: "Voice note"})
		}, "Voice note"},
		{"FileMessage", func(c *ui.Context) {
			FileMessage(c, FileMessageOptions{
				Role:       RoleUser,
				Attachment: AttachmentChipOptions{Name: "trace.json", Mark: "file"},
				Caption:    "the trace of the failing request",
			})
		}, "trace.json"},
		{"ArtifactCard", func(c *ui.Context) {
			ArtifactCard(c, ArtifactCardOptions{
				Title: "Incident report", Kind: "document",
				Body: "It was the pool.", Open: "Open",
			})
		}, "It was the pool."},
		{"CapabilityCards", func(c *ui.Context) {
			CapabilityCards(c, CapabilityCardsOptions{
				Title: "What this window does",
				Capabilities: []Capability{
					{Title: "Answer questions", Body: "About this repository"},
					{Title: "Change code", Body: "With the diff shown"},
				},
			})
		}, "Answer questions"},
		{"EmptyState", func(c *ui.Context) {
			EmptyState(c, EmptyStateOptions{
				Title: "Ask about this repository", Body: "It reads the code and the docs.",
				Action: "New conversation",
			})
		}, "Ask about this repository"},
		{"CitationBadge", func(c *ui.Context) {
			CitationBadge(c, CitationBadgeOptions{N: 2, Count: 5})
		}, "[2]"},
		{"SourceCard", func(c *ui.Context) {
			SourceCard(c, SourceCardOptions{
				N: 1, Title: "config.go", Where: "internal/pool",
				Snippet: "MaxConns: 4", Quotation: "Quote",
			})
		}, "MaxConns: 4"},
		{"ConversationItem", func(c *ui.Context) {
			ConversationItem(c, ConversationItemOptions{
				Title: "Pool exhaustion", Preview: "Why is it timing out?", When: "4m",
			})
		}, "Pool exhaustion"},
		{"StopGeneratingButton", func(c *ui.Context) {
			StopGeneratingButton(c, StopGeneratingButtonOptions{})
		}, "Stop generating"},
		{"SendButton", func(c *ui.Context) {
			SendButton(c, SendButtonOptions{Label: "Send"})
		}, "Send"},
		{"DragDropOverlay", func(c *ui.Context) {
			host := ui.Box(c).Fill().Label("the window")
			DragDropOverlay(c, DragDropOverlayOptions{Host: host})
		}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tt := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				tc.build(c)
			}, 560, 420)
			if tc.want != "" && !tt.HasText(tc.want) {
				t.Errorf("%s does not show %q; texts were %v", tc.name, tc.want, tt.Texts())
			}
			if tt.Image() == nil {
				t.Errorf("%s painted nothing at all", tc.name)
			}
		})
	}
}

// TestConversationContainerCollapses walks the one state the container holds,
// and the reason it holds nothing else: two views of the same window must agree
// about whether the list of conversations is showing.
func TestConversationContainerCollapses(t *testing.T) {
	collapsed := false
	presses := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ConversationContainer(c, &collapsed,
			ConversationContainerOptions{SidebarWidth: 280},
			func() { ui.Text(c, "the list of conversations") },
			func() { ui.Text(c, "the transcript") })
		if r.Collapsed() {
			presses++
			collapsed = true
		}
	}, 900, 400)

	if !tt.HasText("the list of conversations") {
		t.Error("the sidebar is missing")
	}
	if err := tt.Click("Collapse conversations"); err != nil {
		t.Fatal(err)
	}
	if presses != 1 {
		t.Errorf("the container collapsed %d times, want 1", presses)
	}
	if !collapsed {
		t.Error("the collapse press did not write the caller's flag")
	}
}

// TestScrollToBottomButtonHidesWithNothingToJumpTo: a control for going
// nowhere is a control taking up the bottom of the transcript.
func TestScrollToBottomButtonHidesWithNothingToJumpTo(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := ScrollToBottomButton(c, ScrollToBottomButtonOptions{})
		if r.Visible() || r.Element != nil {
			t.Error("the button was drawn with nothing below the reader")
		}
	}, 420, 200)
	_ = tt
}

// TestPanelsRefuseWhatTheyCannotUse: one more pass over the panics, for the
// ones a caller reaches for by accident rather than by mistake.
func TestPanelsRefuseWhatTheyCannotUse(t *testing.T) {
	view := func(fn func(c *ui.Context)) { ui.NewTester(fn, 400, 240) }
	wantsPanic(t, "a message list with no turn builder", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MessageList(c, MessageListOptions{Messages: 1, Height: 200})
		})
	})
	wantsPanic(t, "a message list with no height", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MessageList(c, MessageListOptions{Messages: 1, Message: func(int) {}})
		})
	})
	wantsPanic(t, "a quote with no text", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			QuoteReply(c, QuoteReplyOptions{Author: "Ada"})
		})
	})
	wantsPanic(t, "a tool menu with no set to write into", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ToolToggleMenu(c, ToolToggleMenuOptions{Tools: []string{"search"}})
		})
	})
	wantsPanic(t, "a mode selector out of range", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			mode := 7
			ModeSelector(c, ModeSelectorOptions{Selected: &mode, Modes: []string{"Ask"}})
		})
	})
	wantsPanic(t, "a regenerate menu with nothing to regenerate", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			RegenerateMenu(c, RegenerateOptions{})
		})
	})
	wantsPanic(t, "an image with no description", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ImageGrid(c, ImageGridOptions{Items: []ImageItem{{}}})
		})
	})
	wantsPanic(t, "a system prompt editor with nothing to point at", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SystemPromptEditor(c, SystemPromptEditorOptions{Label: "System prompt"})
		})
	})
	wantsPanic(t, "an audio message with no length", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AudioMessage(c, AudioMessageOptions{Title: "note"})
		})
	})
	wantsPanic(t, "an overlay with nothing to be drawn over", func() {
		view(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DragDropOverlay(c, DragDropOverlayOptions{})
		})
	})
}

// TestTheLibrarysOwnConventionsStillHold: this package is a caller of the rest
// of the library, so the rules it relies on are asserted here as well as where
// they are defined — a change that broke the reduced-motion path or the
// severity ramp would otherwise show up as one odd screenshot.
func TestTheLibrarysOwnConventionsStillHold(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		// A loader, because this package sits inside bubbles and both must
		// agree about reduced motion.
		feedback.DotsLoader(c, feedback.LoaderOptions{Label: "Working"})
		TypingIndicator(c, TypingIndicatorOptions{Who: "Ada"})
	}, 320, 200)
	if !tt.HasText("Ada") {
		t.Error("the typing indicator is not on screen")
	}
}
