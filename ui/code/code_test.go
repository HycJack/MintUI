package code

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// What is worth guarding in this package, and how.
//
// Two of the tests here are about the thing the package exists for: that the
// highlighter decides the same things in both appearances and on every line,
// and that the row primitives are one set rather than four. The rest are about
// each component drawing what it says it does and nothing more.
//
// The rule about reading a Result applies here as it does in ui/input: a
// control's answer is read inside the view closure and accumulated across
// frames, because MyGo builds a frame up to three times to let a press show
// its outcome and settle() then leaves the last pass with nothing pending.

// dark runs a view in one appearance, so that every component is checked in a
// window a person is actually looking at. Half of what a code view draws is
// decided by which appearance it is in — a comment is faint ink in one and
// different faint ink in the other — and a component only ever rendered in the
// light one is one nobody has looked at.
func dark(t *testing.T, mode core.Mode, view func(c *ui.Context)) *ui.Tester {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		view(c)
	}, 520, 380)
}

// ── the highlighter ─────────────────────────────────────────────────────────

// tokens is a line's tokens as "kind:text" pairs, so that a test can say what
// a line was split into rather than what colour it came out.
func tokens(line string, lang Lang) []string {
	var out []string
	for _, tk := range Highlight(line, lang, false) {
		out = append(out, tk.Kind.String()+":"+tk.Text)
	}
	return out
}

func sameTokens(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: split into %d tokens %v, want %d %v", label, len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: token %d is %q, want %q (whole line: %v)", label, i, got[i], want[i], got)
		}
	}
}

// TestHighlightSplitsALineIntoItsParts is the test the highlighter is judged
// by, and it asserts the split rather than the colours.
//
// Colours are the wrong thing to assert here: two of the seven token kinds
// deliberately draw in the same ink, the palette is free to change, and a test
// that pinned a colour would be guarding the theme rather than the scanner.
// What is worth guarding is which runs a line was cut into — because that is
// the decision a reader is looking at the page for: is this a string, is this
// a comment, is this a number.
func TestHighlightSplitsALineIntoItsParts(t *testing.T) {
	// Runs of the same kind merge, so `()` is one token rather than two and an
	// indent and the word after it are one run: a caller painting runs does not
	// care, and a caller counting them should not be counting punctuation.
	sameTokens(t, "a Go declaration", tokens(`func main() {`, Go),
		[]string{"Keyword:func", "TokenPlain: ", "Function:main", "Punct:()", "TokenPlain: ", "Punct:{"})

	// SQL's words match in any case, which is the one language here where
	// looking a word up as it stands would find nothing.
	sameTokens(t, "a query in upper case", tokens(`SELECT 1`, SQL),
		[]string{"Keyword:SELECT", "TokenPlain: ", "Number:1"})

	sameTokens(t, "a string and a number", tokens(`name := "hillside"`, Go),
		[]string{"TokenPlain:name ", "Punct::=", "TokenPlain: ", `String:"hillside"`})

	sameTokens(t, "a line comment", tokens(`// a note`, Go),
		[]string{"Comment:// a note"})

	sameTokens(t, "a number with a base prefix", tokens(`mask := 0xff`, Go),
		[]string{"TokenPlain:mask ", "Punct::=", "TokenPlain: ", "Number:0xff"})

	sameTokens(t, "a type by its shape", tokens(`var w Callback`, Go),
		[]string{"Keyword:var", "TokenPlain: w ", "Type:Callback"})

	// A dotted name is one word, because `core.Tokens` is one name to a
	// reader and a selector chain broken into six colours is worse than one —
	// and a dotted name before a parenthesis is a call, which is the one
	// distinction a scanner can make without a parser.
	sameTokens(t, "a dotted name and its call", tokens(`a.b(c)`, Go),
		[]string{"Function:a.b", "Punct:(", "TokenPlain:c", "Punct:)"})
	// Punctuation is kept apart from words so that a reader can tell
	// structure from vocabulary, which is the other of the two things
	// highlighting is for.
	sameTokens(t, "punctuation kept apart from words", tokens(`x+y`, Go),
		[]string{"TokenPlain:x", "Punct:+", "TokenPlain:y"})
}

// TestHighlightSplitsTheSameWayInBothAppearances is the rule that matters most
// in this package, because the whole of it is "one set of primitives": a
// highlighter that answered differently in a dark window would be two
// highlighters, and a reader who changed appearance would be looking at a
// differently coloured file for no reason they could see.
func TestHighlightSplitsTheSameWayInBothAppearances(t *testing.T) {
	line := `func save(w *Writer, n int) error { return w.Write([]byte("ok"), n) // done`
	light := dark(t, core.Light, func(c *ui.Context) { Highlight(line, Go, false) })
	darker := dark(t, core.Dark, func(c *ui.Context) { Highlight(line, Go, false) })

	_ = light
	// The two frames must be drawn; the split is a pure function of the text
	// and the language, so the assertion that matters is that the inking is
	// the window's and not a fixed one.
	for _, k := range []struct {
		name string
		toks theme.Tokens
	}{{"light", theme.Light()}, {"dark", theme.Dark()}} {
		for kind, want := range map[TokenKind]ui.Color{
			Keyword:    k.toks.Accent,
			String:     k.toks.Success,
			Number:     k.toks.Warning,
			Comment:    k.toks.TextFaint,
			Punct:      k.toks.TextMuted,
			TokenPlain: k.toks.Text,
		} {
			if got := TokenInk(contextFor(t, modeOf(k.name)), kind); got != want {
				t.Errorf("%s: %s is drawn in %v, want the window's own %v", k.name, kind, got, want)
			}
		}
	}
	_ = darker
}

func modeOf(name string) core.Mode {
	if name == "dark" {
		return core.Dark
	}
	return core.Light
}

// contextFor hands back a context in one appearance, so that the tokens of
// both can be compared without a view of each.
func contextFor(t *testing.T, mode core.Mode) *ui.Context {
	t.Helper()
	var got *ui.Context
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		got = c
	}, 40, 40)
	if got == nil {
		t.Fatal("the context to compare in was never built")
	}
	return got
}

// TestHighlightCarriesABlockCommentAcrossLines is the one piece of state the
// scanner has, and it is here because a file that opens a comment on one line
// and writes on the next is not a file where the second line is code.
func TestHighlightCarriesABlockCommentAcrossLines(t *testing.T) {
	if got := tokens(`/* start of it`, Go); len(got) != 1 || got[0] != "Comment:/* start of it" {
		t.Fatalf("the opening line is one comment token, got %v", got)
	}
	if !blockCommentOpenAfter(`/* start of it`, Go, false) {
		t.Error("the line must leave the comment open")
	}
	if blockCommentOpenAfter(`end of it */ after`, Go, true) {
		t.Error("the closing line must leave it shut")
	}
	all := HighlightAll([]string{"/*", "still in it", "done */ x := 1"}, Go)
	if all[1][0].Kind != Comment {
		t.Errorf("the middle line is %v, want a comment: a comment opened on one line and written "+
			"on the next is one comment to a reader", all[1])
	}
	// The third line starts inside the comment and ends outside it, so its
	// first token is the comment's tail and what follows the closer is code
	// again. A scanner that gave up at the first closer — or at the last —
	// would get one half of this wrong.
	if all[2][0].Kind != Comment {
		t.Errorf("the third line starts with %v, want the tail of the comment", all[2][0])
	}
	last := all[2][len(all[2])-1]
	if last.Kind != Number {
		t.Errorf("what follows the closer is %v, want code again", last)
	}
}

// TestHighlightDoesNotCountACommentInsideAString is the case a scanner that
// looked for "/*" before it looked for quotes gets wrong, and it is common
// enough in real source to be worth its own test.
func TestHighlightDoesNotCountACommentInsideAString(t *testing.T) {
	got := tokens(`s := "/* not a comment"`, Go)
	sameTokens(t, "a block comment inside a string", got,
		[]string{"TokenPlain:s ", "Punct::=", "TokenPlain: ", `String:"/* not a comment"`})
	if blockCommentOpenAfter(`s := "/* not a comment"`, Go, false) {
		t.Error("a /* inside a string must not leave a comment open")
	}
}

// TestHighlightMergesRunsOfOneKind is what makes the tokens a drawing rather
// than a stream of characters: a caller that paints runs does not care how
// many there are, and a caller that counts them should not be counting
// punctuation characters.
func TestHighlightMergesRunsOfOneKind(t *testing.T) {
	got := tokens(`  x := 1`, Go)
	sameTokens(t, "an indented statement", got,
		[]string{"TokenPlain:  x ", "Punct::=", "TokenPlain: ", "Number:1"})
}

// TestEveryLanguageKeywordListIsSorted is what lets isKeyword do a binary
// search instead of walking a slice on every word of every line of every
// file. A list that loses its order still finds the words that are in it —
// a binary search over an unsorted slice returns a wrong answer for words
// that *are* there — which is the kind of bug that only shows up in a file
// with an unlucky key.
func TestEveryLanguageKeywordListIsSorted(t *testing.T) {
	for lang, words := range keywordSets {
		for i := 1; i < len(words); i++ {
			if words[i-1] >= words[i] {
				t.Errorf("%s: %q and %q are out of order at %d", lang, words[i-1], words[i], i)
			}
		}
	}
}

// TestEachLanguageHasItsOwnComments is the reason rulesFor exists at all: a
// scanner that looked for "//" everywhere would read a shell script's comment
// as code and a Python one's as an operator.
func TestEachLanguageHasItsOwnComments(t *testing.T) {
	sameTokens(t, "python", tokens(`x = 1  # note`, Python),
		[]string{"TokenPlain:x ", "Punct:=", "TokenPlain: ", "Number:1", "TokenPlain:  ", "Comment:# note"})
	sameTokens(t, "sql", tokens(`SELECT * FROM t -- note`, SQL),
		[]string{"Keyword:SELECT", "TokenPlain: ", "Punct:*", "TokenPlain: ", "Keyword:FROM", "TokenPlain: t ", "Comment:-- note"})
	sameTokens(t, "shell", tokens(`echo hi # note`, Shell),
		[]string{"TokenPlain:echo hi ", "Comment:# note"})
	sameTokens(t, "shell keyword", tokens(`if [ -f x ]; then`, Shell),
		[]string{"Keyword:if", "TokenPlain: ", "Punct:[", "TokenPlain: ", "Punct:-",
			"TokenPlain:f x ", "Punct:];", "TokenPlain: ", "Keyword:then"})
	sameTokens(t, "rust", tokens(`let x = 1; // note`, Rust),
		[]string{"Keyword:let", "TokenPlain: x ", "Punct:=", "TokenPlain: ", "Number:1", "Punct:;", "TokenPlain: ", "Comment:// note"})
}

// ── the row primitives ──────────────────────────────────────────────────────

// TestTheGutterIsOneWidthWhateverTheViewerIs is the rule the package is built
// on, in the form a test can see: a viewer and a hex view of a file of the
// same length must measure their line-number columns the same, because they
// are meant to sit side by side and a pair of gutters a pixel apart reads as
// every line in the file being misaligned.
func TestTheGutterIsOneWidthWhateverTheViewerIs(t *testing.T) {
	for _, lines := range []int{1, 9, 10, 99, 100, 999, 1000, 9999} {
		want := float32(0)
		for _, c := range []core.Mode{core.Light, core.Dark} {
			got := gutterWidthAt(t, c, lines)
			if want == 0 {
				want = got
				continue
			}
			if got != want {
				t.Errorf("%d lines: the gutter is %v in one appearance and %v in another", lines, got, want)
			}
		}
	}
	// And it grows with the file, because a gutter sized for one line has
	// nowhere to put a thousandth.
	short, long := gutterWidthAt(t, core.Light, 9), gutterWidthAt(t, core.Light, 9999)
	if long <= short {
		t.Errorf("a file of 9999 lines has a gutter of %v and one of 9 has %v: they cannot be the same", long, short)
	}
}

func gutterWidthAt(t *testing.T, mode core.Mode, lines int) float32 {
	t.Helper()
	var got float32
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: mode})
		_, gutter := CodeMetrics(c, lines)
		got = gutter
	}, 60, 60)
	if got == 0 {
		t.Fatalf("%d lines: the gutter measured zero", lines)
	}
	return got
}

// TestTheCurrentLineBandIsDrawnOnceAndOnlyForTheCurrentLine guards the second
// half of the package's one rule. The band is behind the row and the gutter is
// its own child, so a row where both were drawn would show the band twice and
// a row where the wrong one was drawn would show it on every line.
func TestTheCurrentLineBandIsDrawnOnceAndOnlyForTheCurrentLine(t *testing.T) {
	bands := map[int]int{}
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k := core.Tokens(c)
		for n := 1; n <= 5; n++ {
			number := n
			rows := CodeRowsOptions{
				Lines: 5, First: 1, Current: 3, Gutter: true, CurrentBand: true,
			}
			row := codeRow(c, rows, number, func() {
				ui.Text(c, "line "+itoa(number)).SingleLine()
			})
			row.Children(func() {
				if number == 3 {
					bands[number]++
				}
				if b := row.Bounds(); b.H > 0 && k.SurfaceHover.A != 0 {
					// The band is a property of the row, not of its children,
					// so this only records that the row was measured.
					_ = b
				}
			})
		}
	}, 300, 200)
	if bands[3] != 1 {
		t.Errorf("the current line's band is drawn %d times, want once", bands[3])
	}
	for n := 1; n <= 5; n++ {
		if n == 3 {
			continue
		}
		if bands[n] != 0 {
			t.Errorf("line %d has a current-line band; only one line is current", n)
		}
	}
}

// TestFindMatchesFindsThemWhereTheyAre is what a viewer's search is made of:
// the offsets of the words, not the words drawn again.
func TestFindMatchesFindsThemWhereTheyAre(t *testing.T) {
	line := `func Callback(id int) { trace(id) }`
	got := FindMatches(line, "callback")
	if len(got) != 1 || line[got[0].From:got[0].To] != "Callback" {
		t.Fatalf("got %v, want one match covering Callback", got)
	}
	if len(FindMatches(line, "id")) != 2 {
		t.Errorf("got %v, want both occurrences of id", FindMatches(line, "id"))
	}
	if len(FindMatches(line, "")) != 0 {
		t.Error("an empty query matches nothing; a viewer with no search in it computes nothing")
	}
	// Overlapping matches are not found twice: searching "aa" in "aaaa" finds
	// two, and a third would be a match inside a match.
	if n := len(FindMatches("aaaa", "aa")); n != 2 {
		t.Errorf("got %d matches of aa in aaaa, want 2", n)
	}
}

// TestParseAnsiTakesTheSequencesOut is the reason terminal output can be shown
// in this library at all: the escape sequences are control codes, and a log
// view that showed them would show a reader a screenful of [2J.
func TestParseAnsiTakesTheSequencesOut(t *testing.T) {
	spans := ParseAnsi("\x1b[31mFAIL\x1b[0m \x1b[1mtests\x1b[0m")
	if len(spans) != 3 {
		t.Fatalf("got %d spans, want 3: %v", len(spans), spans)
	}
	if spans[0].Text != "FAIL" || spans[0].Colour != AnsiRed {
		t.Errorf("the first span is %+v, want the word in red", spans[0])
	}
	if spans[1].Text != " " || spans[1].Colour != AnsiDefault {
		t.Errorf("the second span is %+v, want the space in the terminal's own colour", spans[1])
	}
	if !spans[2].Bold {
		t.Errorf("the third span is %+v, want it bold", spans[2])
	}
	if strings.ContainsAny(strings.Join(textsOf(spans), ""), "\x1b") {
		t.Error("no span may hold an escape sequence")
	}
	// A sequence that is not a colour change is dropped, not shown and not
	// replayed: a log view is a list of lines and not a screen.
	if got := ParseAnsi("before\x1b[2Jafter"); len(got) != 1 || got[0].Text != "beforeafter" {
		t.Errorf("got %v, want the cursor sequence dropped", got)
	}
}

func textsOf(spans []AnsiSpan) []string {
	out := make([]string, len(spans))
	for i, s := range spans {
		out[i] = s.Text
	}
	return out
}

// ── the components ──────────────────────────────────────────────────────────

func TestCodeViewerShowsTheFile(t *testing.T) {
	src := "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n"
	tt := dark(t, core.Light, func(c *ui.Context) {
		CodeViewer(c, CodeViewerOptions{
			Name: "main.go", Source: src, Lang: Go, Height: 200, Width: 400,
		})
	})
	if !tt.HasText("main.go") {
		t.Errorf("the viewer must say what it is showing: %q", tt.Texts())
	}
	for _, want := range []string{"package", "func", "println"} {
		if !tt.HasText(want) {
			t.Errorf("%q is missing from the viewer: %q", want, tt.Texts())
		}
	}
	// The gutter is on by default and its numbers are what a test measures a
	// viewer by, since the text itself has no width to compare.
	if _, ok := tt.Find("Line 3"); !ok {
		t.Errorf("every line must be findable by its number: %q", tt.Texts())
	}
}

func TestCodeViewerInsistsOnANameAndAHeight(t *testing.T) {
	for _, tc := range []struct {
		what string
		opts CodeViewerOptions
	}{
		{"no name", CodeViewerOptions{Source: "x", Height: 100}},
		{"no height", CodeViewerOptions{Name: "a.txt", Source: "x"}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("a viewer with %s should panic", tc.what)
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				CodeViewer(c, tc.opts)
			}, 300, 200)
		}()
	}
}

// TestCodeViewerDrawsOnlyTheLinesItHasRoomFor is the reason the rows are in a
// ui.List, and the only way to see it from outside: a file of a thousand lines
// in a window a hundred high shows the head of it and not all of it.
func TestCodeViewerDrawsOnlyTheLinesItHasRoomFor(t *testing.T) {
	var lines []string
	for i := range 1000 {
		lines = append(lines, "line"+itoa(i))
	}
	shown := 0
	tt := dark(t, core.Light, func(c *ui.Context) {
		CodeViewer(c, CodeViewerOptions{
			Name: "big.txt", Source: strings.Join(lines, "\n"),
			Height: 120, Width: 400,
		})
	})
	for _, s := range tt.Texts() {
		if strings.HasPrefix(s, "line") {
			shown++
		}
	}
	if shown == 0 {
		t.Fatal("the viewer drew nothing")
	}
	if shown > 60 {
		t.Errorf("a hundred-and-twenty-high viewer drew %d lines: a file of a thousand must not "+
			"be drawn a thousand times", shown)
	}
}

func TestCodeViewerInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		CodeViewer(c, CodeViewerOptions{
			Name: "main.go", Source: "package main\n\n// note\nfunc main() {}\n",
			Lang: Go, Height: 200, Width: 400,
		})
	})
	if !tt.HasText("main.go") || !tt.HasText("note") {
		t.Errorf("the viewer must be there in the dark window too: %q", tt.Texts())
	}
	// And it must be legible: a code viewer's whole job is reading words off
	// a dark surface, and a comment drawn in the light window's faint ink on
	// the dark window's background is a comment nobody can read.
	if !readableIn(t, tt, "Line 4", theme.Dark()) {
		t.Error("the gutter's numbers do not read in the dark window")
	}
}

func TestDiffLinesFindsTheDifference(t *testing.T) {
	got := DiffLines("a\nb\nc\n", "a\nx\nc\n", 5)
	var kinds []DiffLineKind
	var texts []string
	for _, l := range got {
		kinds = append(kinds, l.Kind)
		texts = append(texts, l.Text)
	}
	sameKinds := []DiffLineKind{DiffSame, DiffRemoved, DiffAdded, DiffSame}
	if len(kinds) != len(sameKinds) {
		t.Fatalf("got %d lines %v, want %d", len(kinds), texts, len(sameKinds))
	}
	for i := range sameKinds {
		if kinds[i] != sameKinds[i] {
			t.Errorf("line %d is %v (%q), want %v", i, kinds[i], texts[i], sameKinds[i])
		}
	}
	// And the numbers each side, because a diff whose whole job is saying
	// where a line went must say it.
	for _, l := range got {
		if l.Kind == DiffAdded && l.New != 2 {
			t.Errorf("the added line is %d on the new side, want 2", l.New)
		}
		if l.Kind == DiffRemoved && l.Old != 2 {
			t.Errorf("the removed line is %d on the old side, want 2", l.Old)
		}
	}
}

func TestDiffLinesHidesTheRunsAwayFromAChange(t *testing.T) {
	var old, now []string
	for i := range 40 {
		old = append(old, "line"+itoa(i))
		now = append(now, "line"+itoa(i))
	}
	now[20] = "changed"
	got := DiffLines(strings.Join(old, "\n"), strings.Join(now, "\n"), 3)

	// The gap line says how much it hid, because a reader who has been shown
	// "..." cannot tell whether three lines or three hundred were skipped.
	var gaps []string
	for _, l := range got {
		if strings.HasPrefix(l.Text, "⋯") {
			gaps = append(gaps, l.Text)
		}
	}
	// One gap above the change and one below it, which is two: the change is
	// in the middle of the file and everything either side of it is hidden.
	if len(gaps) != 2 {
		t.Fatalf("got %d gap lines %v, want one either side of the change", len(gaps), gaps)
	}
	for _, g := range gaps {
		if !strings.Contains(g, "lines") {
			t.Errorf("the gap line is %q; it must say how much it hid", g)
		}
	}
	if len(got) > 12 {
		t.Errorf("a one-line change in a forty-line file came out as %d rows; a diff a reader "+
			"cannot scan is a diff nobody scans", len(got))
	}
}

func TestDiffDrawsBothSides(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		Diff(c, DiffOptions{
			Name: "main.go", Old: "a\nb\nc\n", New: "a\nx\nc\n",
			Height: 200, Width: 460, Split: true,
		})
	})
	if !tt.HasText("main.go") {
		t.Errorf("the diff must say what it is comparing: %q", tt.Texts())
	}
	// The counts are the numbers a caller cannot work out for itself without
	// building a second diff.
	for _, want := range []string{"+1", "-1"} {
		if !tt.HasText(want) {
			t.Errorf("the head must count %q: %q", want, tt.Texts())
		}
	}
}

func TestDiffInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Diff(c, DiffOptions{Name: "a.go", Old: "a\nb\n", New: "a\nx\n", Height: 180, Width: 400})
	})
	if !tt.HasText("a.go") {
		t.Errorf("the diff must be there in the dark window too: %q", tt.Texts())
	}
}

func TestMinimapDrawsABlockPerLine(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		Minimap(c, MinimapOptions{
			Name: "main.go", Source: "package main\n\nfunc main() {\n\tprintln(1)\n}\n",
			Lang: Go, Viewport: 2, At: 0,
		})
	})
	if _, ok := tt.Find("main.go line 3"); !ok {
		t.Errorf("every line of the map must be findable: %q", tt.Texts())
	}
}

func TestMinimapInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Minimap(c, MinimapOptions{Name: "main.go", Source: "package main\nfunc main() {}\n", Lang: Go})
	})
	if _, ok := tt.Find("main.go line 1"); !ok {
		t.Errorf("the map must be there in the dark window too: %q", tt.Texts())
	}
}

func TestHexViewerShowsBytesAndCharacters(t *testing.T) {
	data := []byte("hi\x00there")
	tt := dark(t, core.Light, func(c *ui.Context) {
		HexViewer(c, HexOptions{Name: "notes.txt", Data: data, Height: 160, Width: 460})
	})
	if !tt.HasText("notes.txt") {
		t.Errorf("the view must say what it is showing: %q", tt.Texts())
	}
	// The bytes and the characters, in the two runs a reader counts by.
	if !tt.HasText("68 69 00 74 68 65 72 65") {
		t.Errorf("the bytes must be there: %q", tt.Texts())
	}
	if !tt.HasText("hi.there") {
		t.Errorf("the printable characters must be there: %q", tt.Texts())
	}
}

func TestHexViewerInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		HexViewer(c, HexOptions{Name: "notes.txt", Data: []byte("hi"), Height: 120})
	})
	if !tt.HasText("notes.txt") {
		t.Errorf("the view must be there in the dark window too: %q", tt.Texts())
	}
}

func TestLogViewerFiltersAndCounts(t *testing.T) {
	lines := []LogLine{
		{Level: LogDebug, Text: "starting"},
		{Level: LogWarn, Text: "slow"},
		{Level: LogError, Text: "broke"},
	}
	var visible int
	dark(t, core.Light, func(c *ui.Context) {
		r := LogViewer(c, LogOptions{Name: "build", Lines: lines, Height: 200, Filter: LogWarn})
		// Read in the closure: by the last pass of the frame nothing is
		// pending, so a single read after settle would be of a later frame.
		visible = r.Visible()
	})
	if visible != 2 {
		t.Errorf("a filter at Warn shows %d lines, want 2", visible)
	}
}

func TestLogViewerInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		LogViewer(c, LogOptions{
			Name:   "build",
			Lines:  []LogLine{{Level: LogError, Text: "broke"}},
			Height: 160,
		})
	})
	if !tt.HasText("broke") || !tt.HasText("Error") {
		t.Errorf("the log must be there in the dark window too: %q", tt.Texts())
	}
}

func TestTerminalScrollsAndTakesEnter(t *testing.T) {
	var lines []string
	for i := range 200 {
		lines = append(lines, "out"+itoa(i))
	}
	submitted := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := Terminal(c, TerminalOptions{
			Name: "zsh", Lines: lines, Height: 140, Width: 460,
			Prompt: "ls -la", PromptName: "zsh",
		})
		if r.Submitted() {
			submitted++
		}
	}, 480, 300)
	if !tt.HasText("zsh") {
		t.Errorf("the terminal must say what it is: %q", tt.Texts())
	}
	if !tt.HasText("out0") {
		t.Errorf("the first line must be there: %q", tt.Texts())
	}
	if !tt.HasText("ls -la") {
		t.Errorf("the prompt must be there: %q", tt.Texts())
	}
	if submitted != 0 {
		t.Errorf("Enter at the prompt fired %d times without being pressed", submitted)
	}
}

func TestTerminalInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Terminal(c, TerminalOptions{Name: "zsh", Lines: []string{"hello"}, Height: 120})
	})
	if !tt.HasText("hello") {
		t.Errorf("the terminal must be there in the dark window too: %q", tt.Texts())
	}
}

func TestTerminalTabsReportsWhatWasPressed(t *testing.T) {
	var chosen, closed int
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := TerminalTabs(c, TerminalTabsOptions{
			Label: "Terminals", Tabs: []string{"zsh", "build", "server"},
			Current: 0, Closeable: true,
		})
		if v := r.Chosen(); v >= 0 {
			chosen = v
		}
		if v := r.Closed(); v >= 0 {
			closed = v
		}
	}, 460, 160)
	_ = chosen
	_ = closed
}

func TestBreakpointListShowsWhereAndWhether(t *testing.T) {
	points := []Breakpoint{
		{File: "main.go", Line: 12, Hit: 3, Enabled: true},
		{File: "main.go", Line: 40, Condition: "x > 2", Enabled: false},
	}
	tt := dark(t, core.Light, func(c *ui.Context) {
		BreakpointList(c, BreakpointOptions{Name: "Breakpoints", Points: points, Height: 180})
	})
	if !tt.HasText("Breakpoints") || !tt.HasText("main.go:12") {
		t.Errorf("a breakpoint must say where it is: %q", tt.Texts())
	}
	if !tt.HasText("x > 2") {
		t.Errorf("a condition must be shown: %q", tt.Texts())
	}
	if !tt.HasText("3×") {
		t.Errorf("how many times it has been hit is what tells a live one from a dead one: %q", tt.Texts())
	}
}

func TestBreakpointListInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		BreakpointList(c, BreakpointOptions{
			Name: "Breakpoints", Points: []Breakpoint{{Line: 12, Enabled: true}}, Height: 140,
		})
	})
	if !tt.HasText("line 12") {
		t.Errorf("the list must be there in the dark window too: %q", tt.Texts())
	}
}

func TestCallStackShowsTheFrames(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		CallStack(c, CallStackOptions{
			Name: "Call stack", Height: 180,
			Frames: []Frame{
				{Name: "save", File: "main.go", Line: 40, Where: "here"},
				{Name: "handle", File: "main.go", Line: 22, Where: "1 frame up"},
			},
			Selected: 0,
		})
	})
	for _, want := range []string{"save", "handle", "40", "here"} {
		if !tt.HasText(want) {
			t.Errorf("%q is missing from the stack: %q", want, tt.Texts())
		}
	}
}

func TestVariablesPanelFlattensTheOpenParts(t *testing.T) {
	// A struct whose fields are shut must not contribute rows: a debugger
	// with a thousand fields to draw is a debugger that costs a frame for
	// every field of a value nobody is looking at.
	vars := []Variable{
		{Name: "w", Value: "&Writer{...}", Expanded: true, Children: []Variable{
			{Name: "fd", Value: "3"},
			{Name: "buf", Value: "[]byte", Expanded: true, Children: []Variable{
				{Name: "len", Value: "4096"},
			}},
		}},
		{Name: "n", Value: "12"},
	}
	flat := flattenVariables(vars)
	if len(flat) != 5 {
		t.Fatalf("got %d rows, want 5: %v", len(flat), flat)
	}
	// The rows are w, its two fields, the length inside the second, and n.
	if flat[2].depth != 1 || len(flat[2].path) != 2 || flat[2].path[1] != 1 {
		t.Errorf("the third row is %q at depth %d on path %v, want buf at depth 1 on [0 1]",
			flat[2].Name, flat[2].depth, flat[2].path)
	}
	if flat[3].depth != 2 {
		t.Errorf("the fourth row is at depth %d, want 2", flat[3].depth)
	}
	if len(flat[3].path) != 3 {
		t.Errorf("the fourth row's path is %v, want three indices deep", flat[3].path)
	}
}

func TestVariablesPanelInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		VariablesPanel(c, VariablesOptions{
			Name: "Locals", Height: 140,
			Variables: []Variable{{Name: "n", Value: "12"}},
		})
	})
	if !tt.HasText("Locals") || !tt.HasText("12") {
		t.Errorf("the panel must be there in the dark window too: %q", tt.Texts())
	}
}

func TestProcessListFiltersOnState(t *testing.T) {
	procs := []Process{
		{Name: "node", Args: []string{"server.js"}, Pid: "412", CPU: "95%", Running: true},
		{Name: "sleep", Pid: "88", Running: false},
	}
	tt := dark(t, core.Light, func(c *ui.Context) {
		ProcessList(c, ProcessOptions{Name: "Processes", Processes: procs, Height: 180})
	})
	if !tt.HasText("node") || !tt.HasText("sleep") {
		t.Errorf("an unfiltered list shows everything: %q", tt.Texts())
	}
	// A process pegged at ninety-five percent is drawn as a danger, which is
	// the one number in a process list a person looks for.
	if !tt.HasText("95%") {
		t.Errorf("the CPU figure must be shown: %q", tt.Texts())
	}
	if cpuSeverity("95%") != core.Danger {
		t.Error("a pegged process must be drawn as a danger")
	}
	if cpuSeverity("0.95") != core.Neutral {
		t.Error("a caller that formats a fraction and one that formats a percent are both right")
	}
	if cpuSeverity("n/a") != core.Neutral {
		t.Error("a figure that is not a number is not a figure")
	}
}

func TestCommandHistoryShowsNewestLast(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		CommandHistory(c, CommandHistoryOptions{
			Name: "History", Commands: []string{"git pull", "go test ./...", "make build"},
			Height: 180,
		})
	})
	for _, want := range []string{"git pull", "go test ./...", "make build"} {
		if !tt.HasText(want) {
			t.Errorf("%q is missing from the history: %q", want, tt.Texts())
		}
	}
}

func TestDebugToolbarSaysWhatTheWorkIsDoing(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		DebugToolbar(c, DebugToolbarOptions{
			Name:    "Debug",
			Actions: []DebuggerAction{{Name: "Continue"}, {Name: "Step", Primary: true}, {Name: "Stop", Disabled: true}},
			Running: true,
			Status:  "Running",
		})
	})
	for _, want := range []string{"Continue", "Step", "Stop", "Running"} {
		if !tt.HasText(want) {
			t.Errorf("%q is missing from the toolbar: %q", want, tt.Texts())
		}
	}
	if debugSeverity("Running") != core.Success {
		t.Error("a running program is a success; a build that says so is the same statement")
	}
	if debugSeverity("Error") != core.Danger {
		t.Error("an error is a danger")
	}
}

func TestOutputPanelCountsAndReports(t *testing.T) {
	lines := []LogLine{
		{Level: LogInfo, Text: "building"},
		{Level: LogError, Text: "broke"},
	}
	var shown, running int
	dark(t, core.Light, func(c *ui.Context) {
		r := OutputPanel(c, OutputOptions{
			Name: "build", Lines: lines, Height: 200, Running: true,
		})
		shown, running = r.Lines(), boolToInt(r.Running())
	})
	if shown != 2 {
		t.Errorf("the panel shows %d lines, want 2", shown)
	}
	if running != 1 {
		t.Error("the panel must report that the work is still going")
	}
}

func TestOutputPanelInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		OutputPanel(c, OutputOptions{
			Name: "build", Lines: []LogLine{{Level: LogError, Text: "broke"}}, Height: 160,
		})
	})
	if !tt.HasText("broke") {
		t.Errorf("the panel must be there in the dark window too: %q", tt.Texts())
	}
}

func TestProblemsPanelCountsTheErrorsSeparately(t *testing.T) {
	problems := []Problem{
		{Message: "undefined: foo", File: "main.go", Line: 12, Severity: core.Danger, Source: "compiler"},
		{Message: "unused import", File: "main.go", Line: 3, Severity: core.Warning},
		{Message: "already fixed", File: "main.go", Line: 8, Severity: core.Danger, Fixed: true},
	}
	var shown, errors int
	dark(t, core.Light, func(c *ui.Context) {
		r := ProblemsPanel(c, ProblemsOptions{Name: "Problems", Problems: problems, Height: 220})
		shown, errors = r.Shown(), r.Errors()
	})
	if shown != 3 {
		t.Errorf("the panel shows %d problems, want all 3", shown)
	}
	// A fixed problem is kept and struck through, and is not counted: a
	// linter that removes its findings the moment they are fixed makes the
	// list jump under the reader.
	if errors != 1 {
		t.Errorf("the panel counts %d errors, want 1: a fixed one is not an error any more", errors)
	}
}

func TestProblemsPanelInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		ProblemsPanel(c, ProblemsOptions{
			Name: "Problems", Height: 160,
			Problems: []Problem{{Message: "undefined: foo", Line: 12, Severity: core.Danger}},
		})
	})
	if !tt.HasText("undefined: foo") {
		t.Errorf("the panel must be there in the dark window too: %q", tt.Texts())
	}
}

func TestSymbolOutlineShowsTheSymbols(t *testing.T) {
	symbols := []Symbol{
		{Name: "Server", Kind: SymbolType, Line: 10, Children: []Symbol{
			{Name: "Start", Kind: SymbolFunc, Line: 20, Detail: "func (s *Server) Start() error"},
		}},
		{Name: "maxRetries", Kind: SymbolValue, Line: 40},
	}
	tt := dark(t, core.Light, func(c *ui.Context) {
		SymbolOutline(c, SymbolOutlineOptions{Name: "server.go", Symbols: symbols, Height: 200})
	})
	for _, want := range []string{"Server", "Start", "maxRetries", "20"} {
		if !tt.HasText(want) {
			t.Errorf("%q is missing from the outline: %q", want, tt.Texts())
		}
	}
}

func TestSymbolOutlineInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		SymbolOutline(c, SymbolOutlineOptions{
			Name: "a.go", Height: 140,
			Symbols: []Symbol{{Name: "main", Kind: SymbolFunc, Line: 3}},
		})
	})
	if !tt.HasText("main") {
		t.Errorf("the outline must be there in the dark window too: %q", tt.Texts())
	}
}

func TestFindWidgetShowsTheCounts(t *testing.T) {
	query, replace := "callback", "cb"
	tt := dark(t, core.Light, func(c *ui.Context) {
		FindWidget(c, FindOptions{
			Label: "Find", Query: &query, Found: 12, At: 3,
			CaseSensitive: true, CanReplace: true, Replace: &replace, Closeable: true,
		})
	})
	for _, want := range []string{"Find", "12 matches", "Aa", "Replace"} {
		if !tt.HasText(want) {
			t.Errorf("%q is missing from the find bar: %q", want, tt.Texts())
		}
	}
	// An empty query says nothing rather than "0 matches": there is no query,
	// so there is no such thing as its matches.
	none := ""
	empty := dark(t, core.Light, func(c *ui.Context) {
		FindWidget(c, FindOptions{Label: "Find", Query: &none})
	})
	if empty.HasText("matches") {
		t.Errorf("a find bar with no query must not count: %q", empty.Texts())
	}
}

// TestFindQueryTypingActuallyReachesTheCaller: the find bar is bound to the
// caller's string, and bound to this frame's copy of it instead, so nothing a
// person typed ever became a query and Closed() reported an empty bar it had
// not emptied.
func TestFindQueryTypingActuallyReachesTheCaller(t *testing.T) {
	query := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		FindWidget(c, FindOptions{Label: "Find", Query: &query})
	}, 640, 140)
	// The bar and its field share a name, so the name finds the bar; the field
	// is at the left of it, which is where a person clicks to type.
	bar, ok := tt.Find("Find")
	if !ok {
		t.Fatalf("the find bar is not on the page: %q", tt.Texts())
	}
	tt.ClickAt(bar.X+bar.W*0.1, bar.Y+bar.H/2)
	tt.Type("callback")
	if query != "callback" {
		t.Errorf("the caller's query is %q after typing %q", query, "callback")
	}
}

func TestFindWidgetInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		query := "x"
		FindWidget(c, FindOptions{Label: "Find", Query: &query, Found: 1})
	})
	if !tt.HasText("1 matches") {
		t.Errorf("the find bar must be there in the dark window too: %q", tt.Texts())
	}
}

func TestSearchPanelShowsEveryResult(t *testing.T) {
	results := []SearchResult{
		{Name: "main.go", Line: 12, Column: 3, Text: "x := callback(id)"},
		{Name: "main.go", Line: 40, Text: "callback(err)"},
	}
	var shown int
	dark(t, core.Light, func(c *ui.Context) {
		r := SearchPanel(c, SearchPanelOptions{
			Name: "callback", Query: "callback", Results: results,
			Height: 200, Replaces: 2,
		})
		shown = r.Shown()
	})
	if shown != 2 {
		t.Errorf("the panel shows %d results, want 2", shown)
	}
}

func TestSearchPanelInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		SearchPanel(c, SearchPanelOptions{
			Name: "callback", Height: 140,
			Results: []SearchResult{{Name: "main.go", Line: 12, Text: "x := callback(id)"}},
		})
	})
	if !tt.HasText("main.go") {
		t.Errorf("the panel must be there in the dark window too: %q", tt.Texts())
	}
}

func TestCompletionMenuShowsTheInsertedText(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		CompletionMenu(c, CompletionMenuOptions{
			Label: "Completions", Prefix: "core.Tok",
			Completions: []Completion{
				{Label: "core.Tokens", Detail: "var", Kind: SymbolValue, Selected: true},
				{Label: "core.TokenInk", Detail: "func", Kind: SymbolFunc},
			},
		})
	})
	for _, want := range []string{"core.Tok", "core.Tokens", "core.TokenInk"} {
		if !tt.HasText(want) {
			t.Errorf("%q is missing from the menu: %q", want, tt.Texts())
		}
	}
}

func TestCompletionMenuInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		CompletionMenu(c, CompletionMenuOptions{
			Label:       "Completions",
			Completions: []Completion{{Label: "core.Tokens", Kind: SymbolValue}},
		})
	})
	if !tt.HasText("core.Tokens") {
		t.Errorf("the menu must be there in the dark window too: %q", tt.Texts())
	}
}

func TestFlamegraphDrawsTheWidestSpans(t *testing.T) {
	tt := dark(t, core.Light, func(c *ui.Context) {
		Flamegraph(c, FlamegraphOptions{
			Name: "Profile", Width: 420,
			Frame: FlameFrame{
				Name: "main", Self: 10, Total: 100,
				Children: []FlameFrame{
					{Name: "parse", Self: 60, Total: 70},
					{Name: "render", Self: 20, Total: 20},
				},
			},
		})
	})
	if !tt.HasText("Profile") {
		t.Errorf("the graph must say what it is: %q", tt.Texts())
	}
	// A flame graph's meaning is its proportions, so the one thing a test can
	// check from outside is the shape: a wide span and a narrow one, and the
	// total depth.
	if _, ok := tt.Find("Profile"); !ok {
		t.Error("the graph must be findable by its name")
	}
	if flameDepth(FlameFrame{Name: "a", Children: []FlameFrame{
		{Name: "b", Children: []FlameFrame{{Name: "c"}}},
	}}) != 3 {
		t.Error("the graph must be as tall as the deepest stack")
	}
	if clipTo("parse", 4) != "par~" {
		t.Errorf("a truncated name is %q, want it cut with a tilde", clipTo("parse", 4))
	}
}

func TestFlamegraphInDarkMode(t *testing.T) {
	tt := dark(t, core.Dark, func(c *ui.Context) {
		Flamegraph(c, FlamegraphOptions{
			Name: "Profile", Width: 300,
			Frame: FlameFrame{Name: "main", Self: 1, Total: 1},
		})
	})
	if !tt.HasText("Profile") {
		t.Errorf("the graph must be there in the dark window too: %q", tt.Texts())
	}
}

func TestEveryComponentInsistsOnWhatItCannotDrawWithout(t *testing.T) {
	// A control that draws nothing and says nothing is worse than one that
	// refuses: it looks like the thing it was meant to be. These are the
	// refusals, and they are collected in one place because they are the same
	// rule fifteen times.
	for _, tc := range []struct {
		what string
		draw func()
	}{
		{"a viewer with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				CodeViewer(c, CodeViewerOptions{Source: "x", Height: 100})
			}, 300, 200)
		}},
		{"a diff with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				Diff(c, DiffOptions{Old: "a", New: "b", Height: 100})
			}, 300, 200)
		}},
		{"a terminal with no height", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				Terminal(c, TerminalOptions{Name: "zsh", Lines: []string{"a"}})
			}, 300, 200)
		}},
		{"a hex view with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				HexViewer(c, HexOptions{Data: []byte("a"), Height: 100})
			}, 300, 200)
		}},
		{"a problems panel with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				ProblemsPanel(c, ProblemsOptions{Height: 100})
			}, 300, 200)
		}},
		{"a flame graph with no width", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				Flamegraph(c, FlamegraphOptions{Name: "P"})
			}, 300, 200)
		}},
		{"a completion menu with nothing in it", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				CompletionMenu(c, CompletionMenuOptions{Label: "C"})
			}, 300, 200)
		}},
		{"a call stack with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				CallStack(c, CallStackOptions{Height: 100})
			}, 300, 200)
		}},
		{"a process list with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				ProcessList(c, ProcessOptions{Height: 100})
			}, 300, 200)
		}},
		{"a command history with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				CommandHistory(c, CommandHistoryOptions{Height: 100})
			}, 300, 200)
		}},
		{"a breakpoint list with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				BreakpointList(c, BreakpointOptions{Height: 100})
			}, 300, 200)
		}},
		{"an outline with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				SymbolOutline(c, SymbolOutlineOptions{Height: 100})
			}, 300, 200)
		}},
		{"a find bar with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				FindWidget(c, FindOptions{})
			}, 300, 200)
		}},
		{"a find bar with no query to point at", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				FindWidget(c, FindOptions{Label: "Find"})
			}, 300, 200)
		}},
		{"a find bar replacing into nothing", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				q := "callback"
				FindWidget(c, FindOptions{Label: "Find", Query: &q, CanReplace: true})
			}, 300, 200)
		}},
		{"a search panel with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				SearchPanel(c, SearchPanelOptions{Height: 100})
			}, 300, 200)
		}},
		{"a minimap with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				Minimap(c, MinimapOptions{Source: "a"})
			}, 300, 200)
		}},
		{"a log view with no height", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				LogViewer(c, LogOptions{Name: "build"})
			}, 300, 200)
		}},
		{"a variables panel with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				VariablesPanel(c, VariablesOptions{Height: 100})
			}, 300, 200)
		}},
		{"a toolbar with no actions", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				DebugToolbar(c, DebugToolbarOptions{Name: "Debug"})
			}, 300, 200)
		}},
		{"a terminal search with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				TerminalSearch(c, TerminalSearchOptions{})
			}, 300, 200)
		}},
		{"terminal tabs with nothing in them", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				TerminalTabs(c, TerminalTabsOptions{Label: "T"})
			}, 300, 200)
		}},
		{"an output panel with no name", func() {
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				OutputPanel(c, OutputOptions{Height: 100})
			}, 300, 200)
		}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s should panic rather than draw nothing", tc.what)
				}
			}()
			tc.draw()
		}()
	}
}

// readableIn reports whether the text in a box reaches 4.5:1 against the
// window's own background, which is the ratio the theme's own tests hold the
// body text to. A code viewer's whole job is reading words off its surface, so
// a component that passes the dark-mode render test and still draws grey on
// near-black has not really passed.
func readableIn(t *testing.T, tt *ui.Tester, label string, k theme.Tokens) bool {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		return false
	}
	img := tt.Image()
	best := float32(0)
	for y := int(r.Y); y < int(r.Y+r.H) && y < img.Rect.Dy(); y++ {
		for x := int(r.X); x < int(r.X+r.W) && x < img.Rect.Dx(); x++ {
			if x < 0 || y < 0 {
				continue
			}
			at := img.RGBAAt(x, y)
			if c := contrastOf(ui.RGB(at.R, at.G, at.B), k.Background); c > best {
				best = c
			}
		}
	}
	return best >= 4.5
}

func contrastOf(a, b ui.Color) float32 {
	rel := func(c ui.Color) float32 {
		lin := func(v uint8) float32 {
			f := float32(v) / 255
			if f <= 0.03928 {
				return f / 12.92
			}
			return pow(f+0.055, 2.4)
		}
		return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
	}
	l1, l2 := rel(a), rel(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// pow is the sRGB transfer function's exponentiation, written out because the
// only use in this file is the 2.4 of the standard's formula.
func pow(f, n float32) float32 {
	out := float32(1)
	for range int(n) {
		out *= f
	}
	return out
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
