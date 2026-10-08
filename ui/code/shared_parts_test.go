package code

import (
	"image"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// sameImage reports whether two frames drew the same pixels, which is how a
// test says "the colour reached the screen" without pinning a palette value.
func sameImage(a, b *image.RGBA) bool {
	if a == nil || b == nil || a.Bounds() != b.Bounds() {
		return false
	}
	bnds := a.Bounds()
	for y := bnds.Min.Y; y < bnds.Max.Y; y++ {
		for x := bnds.Min.X; x < bnds.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}
	return true
}

// TestCodeAnsiSpansDrawsTheRunsTheSequencesAskedFor is the shared part a log
// line and a terminal's output are made of: the words, in the colours the
// program asked for, with no trace of the sequences that asked.
func TestCodeAnsiSpansDrawsTheRunsTheSequencesAskedFor(t *testing.T) {
	line := "\x1b[31mFAIL\x1b[0m \x1b[1mtests\x1b[0m"
	for _, mode := range []core.Mode{core.Light, core.Dark} {
		tt := dark(t, mode, func(c *ui.Context) {
			CodeAnsiSpans(c, line, theme.RowSize, ui.Color{}, ui.Color{})
		})
		if !tt.HasText("FAIL") || !tt.HasText("tests") {
			t.Errorf("%v: the line is missing its words: %q", mode, tt.Texts())
			continue
		}
		for _, s := range tt.Texts() {
			if strings.Contains(s, "\x1b") {
				t.Errorf("%v: a drawn run holds an escape sequence: %q", mode, s)
			}
		}
		// The colour a run asked for reaches the screen: the same words,
		// never coloured, are the control.
		plain := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Mode: mode})
			CodeAnsiSpans(c, "FAIL tests", theme.RowSize, ui.Color{}, ui.Color{})
		}, 520, 380)
		if sameImage(tt.Image(), plain.Image()) {
			t.Errorf("%v: the red and bold runs drew no difference at all", mode)
		}
	}
	// A line with nothing in it draws nothing: a part that drew an empty box
	// would leave a row in every log that is empty.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		CodeAnsiSpans(c, "", theme.RowSize, ui.Color{}, ui.Color{})
	}, 300, 60)
	if got := tt.Texts(); len(got) != 0 {
		t.Errorf("an empty line drew %q, want nothing", got)
	}
}

// TestCodeMatchSpansMarksItsMatches is the part a search hit is made of when
// the line was never tokenised: the mark behind the found words, and nothing
// behind the rest.
func TestCodeMatchSpansMarksItsMatches(t *testing.T) {
	line := "the quick brown fox"
	for _, mode := range []core.Mode{core.Light, core.Dark} {
		tt := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{Mode: mode})
			CodeMatchSpans(c, line, FindMatches(line, "quick"), theme.RowSize)
		}, 520, 380)
		if !tt.HasText("the") || !tt.HasText("quick") || !tt.HasText("brown fox") {
			t.Errorf("%v: the line is missing: %q", mode, tt.Texts())
			continue
		}
		mark := func(marks []MatchSpans) *image.RGBA {
			return ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{Mode: mode})
				CodeMatchSpans(c, line, marks, theme.RowSize)
			}, 520, 380).Image()
		}
		if sameImage(mark(nil), mark(FindMatches(line, "quick"))) {
			t.Errorf("%v: a match left no mark behind it", mode)
		}
		if sameImage(mark(FindMatches(line, "quick")), mark(FindMatches(line, "fox"))) {
			t.Errorf("%v: two different matches mark the same pixels", mode)
		}
	}
}
