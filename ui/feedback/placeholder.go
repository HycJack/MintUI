package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// A skeleton screen says "this is coming" in the shape of the thing that is
// coming. Three components, three jobs, and they compose:
//
//	Skeleton     the placeholder block — what a line of text will be
//	Shimmer      the highlight that crosses a block, to say it is a placeholder
//	ShimmerText  several skeletons, one per line, under one shimmer
//
// Nothing here breathes. A placeholder that pulses on its own is an
// animation the reader has to look at while waiting for content, and the
// shimmer already says the one thing worth saying.

// SkeletonOptions configure a Skeleton.
type SkeletonOptions struct {
	// Width is the block's width in DIPs.
	Width float32
	// Height is the block's height. Zero takes one line of text at the
	// body's size, which is what almost every placeholder is standing in
	// for.
	Height float32
	// Radius rounds the block. Zero takes a small radius, which is what a
	// block standing in for a line of text should be.
	Radius float32
	// FillWidth takes the width of the parent instead, which is how a row of
	// placeholders lines up with the row it is replacing without the caller
	// having to know that width.
	FillWidth bool
	// Color overrides the block's fill. Zero alpha takes the surface step
	// above the one it stands in for.
	Color ui.Color
	// Label names the block for assistive technology and for tests. A
	// placeholder is invisible to a screen reader by design, so a skeleton
	// that stands in for something meaningful should say what.
	Label string
}

// Skeleton is one rounded block standing in for content that has not arrived.
//
// It is deliberately inert: no shimmer, no pulse, no frame of its own. That
// is why it needs no reduced-motion branch — it is the resting state already,
// and Shimmer is the part that needs deciding whether to move.
func Skeleton(c *ui.Context, opts SkeletonOptions) *ui.Element {
	if opts.Width <= 0 && !opts.FillWidth {
		panic("feedback: Skeleton needs a Width, or FillWidth to take its " +
			"parent's; without one it has nothing to stand in for")
	}
	k := core.Tokens(c)

	h := opts.Height
	if h <= 0 {
		h = core.FontSize(c, theme.BodySize) * 1.4
	}
	rad := opts.Radius
	if rad <= 0 {
		rad = theme.SmallRadius
	}
	// SurfaceHover rather than Surface: a placeholder very often stands in
	// for a row on a Surface panel — a list row, a board column — and a
	// block the same colour as what it stands in for is invisible. The next
	// step up reads on both surfaces in both appearances, which is the
	// reason it is the hover colour and not a new token of its own.
	fill := opts.Color
	if fill.A == 0 {
		fill = k.SurfaceHover
	}

	block := ui.Box(c).Height(h).Radius(rad).Background(fill).Shrink(0).Label(opts.Label)
	if opts.FillWidth {
		block.FillWidth()
	} else {
		block.Width(opts.Width)
	}
	return block
}

// ShimmerOptions configure a Shimmer.
type ShimmerOptions struct {
	// Radius rounds the highlight to the block's own corners, because a
	// highlight clipped to a square drawn over a rounded block shows four
	// corners of shine where the block has none. Zero leaves it square,
	// which is right for a block that fills its panel.
	Radius float32
	// Period is how long the highlight takes to cross, in milliseconds.
	// Zero takes the loaders' period: the crossing should read at the same
	// rate as everything else that says "working".
	Period float32
	// Color is the colour the highlight is made of. Zero alpha takes the
	// window's own text colour, which is the one colour guaranteed to show
	// against a placeholder fill in both appearances; the highlight's own
	// softness is the ramp's business, not this option's.
	Color ui.Color
	// Width is how much of the block the highlight covers, as a fraction of
	// it. Zero takes 0.45, wide enough to be a band and narrow enough to
	// look like it is going somewhere.
	Width float32
	// Label names the block the highlight crosses.
	Label string
}

// Shimmer draws the highlight that crosses a placeholder, saying "this is
// still loading" in a way a static block cannot.
//
// child is what stands in the shimmer's place — almost always a Skeleton,
// which is why the two are separate components and not one with a flag:
// several blocks under one highlight is a row of text, not three loaders.
func Shimmer(c *ui.Context, opts ShimmerOptions, child func()) *ui.Element {
	k := core.Tokens(c)

	period := opts.Period
	if period <= 0 {
		period = loaderPeriod
	}
	band := opts.Width
	if band <= 0 {
		band = 0.45
	}
	ink := opts.Color
	if ink.A == 0 {
		ink = k.Text
	}
	still := core.Reduced(c)

	e := ui.Box(c).FillWidth().Label(opts.Label)
	if child != nil {
		e.Children(child)
	}
	e.DrawOver(func(p *ui.Painter, r ui.Rect) {
		if still {
			// Under reduced motion the highlight is not drawn at all rather
			// than drawn once in the middle: a highlight that does not move
			// is a bright band sitting on the content it stands in for,
			// which is the one thing a placeholder must never look like.
			return
		}
		sweep(p, r, opts.Radius, ink, band, loopPhase(p, ms(period), false))
	})
	return e
}

// ShimmerTextOptions configure a ShimmerText.
type ShimmerTextOptions struct {
	// Lines is how many lines of text are standing in. It is required: a
	// shimmer text with no lines is a shimmer with nothing under it.
	Lines int
	// Width is the longest line's width in DIPs. It is required, because a
	// placeholder's width is the only thing that tells a reader how much
	// text is coming.
	Width float32
	// LineHeight and Gap override the standard rhythm. Zero takes a line of
	// body text and two density units between lines.
	LineHeight, Gap float32
	// Label names the block for assistive technology.
	Label string
	// Period is the shimmer's crossing time in milliseconds.
	Period float32
}

// ShimmerText is the placeholder for a paragraph: one skeleton per line, the
// last one short, under a single highlight.
//
// It is Shimmer over Skeletons and nothing else — the reason the two exist
// separately is that this composition is the common case, and a caller who
// wants three custom blocks under one highlight should be able to write it
// with the two it already has.
func ShimmerText(c *ui.Context, opts ShimmerTextOptions) *ui.Element {
	if opts.Lines <= 0 {
		panic("feedback: ShimmerText needs at least one line")
	}
	if opts.Width <= 0 {
		panic("feedback: ShimmerText needs a Width; a placeholder's width is " +
			"how much text is coming")
	}
	u := core.Density(c).Unit()

	h := opts.LineHeight
	if h <= 0 {
		h = core.FontSize(c, theme.BodySize) * 1.4
	}
	gap := opts.Gap
	if gap <= 0 {
		gap = u * 2
	}

	return Shimmer(c, ShimmerOptions{Period: opts.Period, Label: opts.Label},
		func() {
			ui.Column(c).FillWidth().Gap(gap).Children(func() {
				for i := range opts.Lines {
					// The last line is short, as a paragraph's is: a run of
					// equal full-width lines reads as a table, not as prose.
					w := opts.Width
					if i == opts.Lines-1 {
						w *= 0.55
					}
					Skeleton(c, SkeletonOptions{Width: w, Height: h, Radius: h / 2})
				}
			})
		})
}

// sweep draws the crossing highlight over r.
//
// It is a run of thin slices whose alpha rises and falls across the band,
// rather than one LinearGradient: a sweep needs a ramp that comes back down,
// and a single linear gradient can only go one way. Twelve slices is enough
// that the edges are soft without being a per-pixel loop.
func sweep(p *ui.Painter, r ui.Rect, radius float32, ink ui.Color, band float32, phase float32) {
	const slices = 12
	w := r.W * band
	x := r.X - w + (r.W+w)*phase
	slice := w / slices

	p.Clip(r, radius, func() {
		for i := range slices {
			// A half-sine across the band: nothing at its edges, most of it
			// in the middle.
			t := float32(i) / float32(slices-1)
			a := 0.16 * (1 - (2*t-1)*(2*t-1))
			p.Fill(ui.Rect{X: x + slice*float32(i), Y: r.Y, W: slice, H: r.H},
				ink.Alpha(a), 0)
		}
	})
}
