package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Two components that put a number or a string out one piece at a time. Both
// take what they have shown as an argument rather than keeping it, because
// the component that owns a counter has to be told when to start and when to
// stop — and a caller refreshing a value every frame already has both.

// countUpKey and caretKey name the animations on the elements that run them.
// The state they name lives on the element, so two figures on one screen do
// not share a count and the caller never has to invent a key to keep them
// apart.
const (
	countUpKey = "feedback.countUp"
	caretKey   = "feedback.caret"
	// caretPeriod is how long the caret takes to fade out and back in. It is
	// slower than a loader's loop on purpose: a caret is not work in
	// progress, and a fast blink next to a real spinner is noise.
	caretPeriod float32 = 1100
	// caretGap is the space between the text and its caret, in density
	// units. Without it the caret sits on the last glyph.
	caretGap float32 = 0.75
)

// caretSize returns a caret's width and height at a type size: a thin bar as
// tall as the line. It is a function rather than two numbers because the
// component and the tests that measure it have to agree on where it is.
func caretSize(size float32) (w, h float32) {
	h = size * 1.15
	return h * 0.14, h
}

// CountUpOptions configure a CountUp.
type CountUpOptions struct {
	// Value is the number to show. When it changes the number moves to it
	// rather than cutting to it, so a figure that arrives in steps — a queue
	// draining, a page of results being counted — reads as growing rather
	// than as flickering.
	Value int
	// Prefix and Suffix sit around the digits: "$", " callbacks", "%". They
	// are part of the number as far as anyone reading it aloud is concerned.
	Prefix, Suffix string
	// Duration is how long a change takes, in milliseconds. Zero takes 400.
	Duration float32
	// Size is the type size in points; zero takes theme.StatSize, which is
	// the size a figure that matters is drawn at.
	Size float32
	// Bold sets the weight, for a figure that is the point of a header.
	Bold bool
	// Label names the figure for assistive technology. Empty takes the
	// number with its prefix and suffix at the value it is heading for — not
	// the digit in flight, because a number still counting is not yet a
	// number.
	Label string
}

// CountUp is a figure that moves to its new value instead of cutting to it.
//
// It counts when the value changes, not when it first appears. The first
// frame shows the value as given, for two reasons: a figure that climbs out
// of nothing every time a panel opens is a number pretending to be a
// process, and nothing in this package knows when the climb should have
// started — only the caller does.
func CountUp(c *ui.Context, opts CountUpOptions) *ui.Element {
	k := core.Tokens(c)

	size := opts.Size
	if size <= 0 {
		size = theme.StatSize
	}
	dur := opts.Duration
	if dur <= 0 {
		dur = 400
	}
	label := opts.Label
	if label == "" {
		label = formatCount(opts.Value, opts.Prefix, opts.Suffix)
	}

	// The figure is a box with a label and a text inside it rather than one
	// text element, because those are two different numbers: what it draws
	// is where the count has got to, and what it says is the value it is
	// heading for. Reduced motion is not a shorter count-up either, it is
	// the number now — a figure that ticks part way there leaves the reader
	// holding a number that is not the number.
	fig := ui.Box(c).Label(label).FontSize(core.FontSize(c, size))
	shown := opts.Value
	if !core.Reduced(c) {
		// The duration goes through core.Motion as well, so a window that
		// only shortened the motion still gets an instant count.
		shown = int(fig.AnimateWith(countUpKey, float32(opts.Value),
			ms(core.Motion(c, dur)), ui.EaseOut) + 0.5)
	}
	fig.Children(func() {
		ui.Text(c, formatCount(shown, opts.Prefix, opts.Suffix)).
			TextColor(k.Text).FontSize(core.FontSize(c, size))
	})
	if opts.Bold {
		fig.Bold()
	}
	return fig
}

// formatCount renders a figure the way the rest of the interface does:
// thousands separated, so four digits read as four digits.
func formatCount(v int, prefix, suffix string) string {
	return prefix + internal.Commas(v) + suffix
}

// TypewriterOptions configure a Typewriter.
type TypewriterOptions struct {
	// Text is the whole line. It is required: a typewriter with nothing to
	// type is a caret.
	Text string
	// Shown is how many of its characters are on screen. More than the text
	// has is the whole text; fewer is a prefix.
	//
	// It is the caller's number and not a clock inside the component,
	// because the caller is the only thing that knows when the line began —
	// a reply arriving, an echo of a search, a status that restarts when a
	// second job starts. A component that counted on its own would need
	// somewhere to keep when it started, and this package keeps nothing.
	Shown int
	// NoCaret turns off the caret after the text. It is on by default: a
	// line arriving letter by letter with nothing at its end reads as a
	// fade, not as typing.
	NoCaret bool
	// Size and Color override the type; zero takes theme.BodySize and the
	// window's text colour.
	Size  float32
	Color ui.Color
	// Label names the line for assistive technology. Empty takes the whole
	// text, because a reader should hear the line rather than the prefix of
	// it that happens to be on screen.
	Label string
}

// Typewriter is a line of text shown up to a given character, with a caret at
// its end. It is for text that is genuinely arriving — a reply being written
// out, a search echoing what was typed — and not for a line someone is typing
// into, which is an editable text and takes the keyboard.
func Typewriter(c *ui.Context, opts TypewriterOptions) *ui.Element {
	if opts.Text == "" {
		panic("feedback: Typewriter needs Text; a caret on its own says the " +
			"interface is listening to nobody")
	}
	if opts.Shown < 0 {
		panic("feedback: Typewriter Shown is how many characters show, and it " +
			"cannot be before the start of the text")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	// Cut on characters rather than bytes, so a line cut mid-character is
	// fine and one cut mid-grapheme is not: a replacement character is the
	// one thing a placeholder must never show.
	shown := opts.Text
	if runes := []rune(shown); opts.Shown < len(runes) {
		shown = string(runes[:opts.Shown])
	}

	size := opts.Size
	if size <= 0 {
		size = theme.BodySize
	}
	ink := opts.Color
	if ink.A == 0 {
		ink = k.Text
	}
	still := core.Reduced(c)
	label := opts.Label
	if label == "" {
		label = opts.Text
	}

	// Shrink(0) so the row is as wide as the line and the caret rather than
	// the whole panel: a line of arriving text has an end, and it is the end
	// of the text.
	return ui.Row(c).AlignItems(ui.Center).Shrink(0).Label(label).Children(func() {
		ui.Text(c, shown).TextColor(ink).FontSize(core.FontSize(c, size))

		if opts.NoCaret {
			return
		}
		// The caret is an element rather than a character in the string, so
		// it can be as tall as the type and blink without the text being
		// re-measured every frame.
		w, h := caretSize(core.FontSize(c, size))
		cb := ui.Box(c).Width(w).Height(h).Radius(1).Shrink(0)
		// The margin is on the left: an element's margins are given top,
		// right, bottom, left, and a caret wants its space before it rather
		// than under it — a top margin here would push it out of line with
		// the text it belongs to.
		cb.Margin(0, 0, 0, u*caretGap)
		cb.Background(ink.Alpha(caretInk(still,
			cb.AnimateWith(caretKey, 1, ms(core.Motion(c, caretPeriod)), ui.Bounce(ui.Linear)))))
	})
}

// caretInk is how inked the caret is at a point in its blink.
//
// Still, the caret is fully drawn. A caret that blinks is this component's
// only motion, and under reduced motion the thing it is saying — the line is
// still arriving — is better said by a caret that is simply there than by a
// line with a gap at the end of it.
func caretInk(still bool, at float32) float32 {
	if still {
		return 1
	}
	return 0.2 + 0.8*at
}
