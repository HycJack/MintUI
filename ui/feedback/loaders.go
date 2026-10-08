package feedback

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
)

// LoaderOptions configure every loader in this package.
//
// There is one options type rather than one per loader on purpose: the five
// loaders differ in nothing but the shape they draw, so a caller that swaps a
// spinner — a list for a button, a bar for a whole panel — changes one word
// and keeps its size and colour. Five near-identical option structs would be
// five places for the two things they share to drift apart.
type LoaderOptions struct {
	// Size is the square the shape is drawn in, in DIPs. Zero takes the
	// standard size for the window's density.
	Size float32
	// Color draws the shape. Zero alpha means the accent colour, which is
	// the one colour that reads as "the system is working" in both
	// appearances.
	Color ui.Color
	// Label names what is loading, for assistive technology. Empty takes
	// the library's "Loading", which is right for a spinner beside a
	// control and wrong beside a subject line, so anything with a subject
	// should say it.
	Label string
	// Period is how long one loop takes, in milliseconds. Zero takes 900:
	// slower than that a loader reads as stuck rather than busy.
	Period float32
}

// The proportions of the loaders. They are constants rather than options
// because they are the shapes themselves — a three-bar loader with four bars
// is a different mark — and a caller who wants that wants another component.
//
// The rest* values are the poses a loader is drawn in when the window asked
// for reduced motion. Each is a pose the mark is legible in on its own: bars
// level, dots equally inked, the orbit dot straight up, the pulse dot at
// middle size. None of them is a frozen frame of the loop, because a waveform
// caught mid-rise reads as a stuck interface rather than as a quieter one.
const (
	loaderDefaultSize float32 = 5   // × the density unit
	loaderPeriod      float32 = 900 // milliseconds per loop
	barsCount         int     = 3   // BarsLoader
	dotsCount         int     = 3   // DotsLoader
	waveCount         int     = 4   // WaveLoader

	// restBar is the fraction of its height a bar stands at when still.
	restBar float32 = 0.5
	// restFloor is how short the shortest bar of a run may get while it
	// moves: a bar at zero is a gap, and the run stops reading as a run.
	restFloor float32 = 0.35
	// restDot is the alpha a dot is drawn at when still — full ink, so three
	// still dots read as three dots rather than as a dim smudge.
	restDot float32 = 1
	// restDip is the alpha a dot falls to at the back of its turn.
	restDip float32 = 0.3
	// restAngle is straight up, −π/2: where an orbit's dot sits when it is
	// not going round.
	restAngle float32 = -1.5707963
	// restRadius is the fraction of its box a pulse's dot is drawn at when
	// still.
	restRadius float32 = 0.62
)

// tick is where a loader is in its loop, and whether it is moving at all.
// Every shape is handed one, so what a still loader draws is answered in the
// shape rather than by a branch around the whole loader, and a shape added
// later cannot forget the question: it has to take a tick to compile.
type tick struct {
	// phase is 0 to 1 of one loop.
	phase float32
	// still is core.Reduced, read where a Context is to hand, because a
	// Painter has none.
	still bool
}

// loaderShape draws one loader's shape. It is the type all five share, so
// loader can resolve the box, the ink, the label and the motion decision
// once.
type loaderShape func(p *ui.Painter, r ui.Rect, t tick)

// loader is what the five loaders share: the box they draw in, the colour
// they draw with, the name assistive technology reads, and the decision
// about whether they move at all.
func loader(c *ui.Context, opts LoaderOptions, shape loaderShape) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	side := opts.Size
	if side <= 0 {
		side = u * loaderDefaultSize
	}
	col := opts.Color
	if col.A == 0 {
		col = k.Accent
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "feedback.loader.loading", "Loading")
	}
	period := opts.Period
	if period <= 0 {
		period = loaderPeriod
	}
	still := core.Reduced(c)

	// RoleStatus: a loader is the one thing on a screen a screen reader
	// should not have to be told about on every poll, and the label is what
	// it says when it does.
	return ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleStatus).Label(label).
		Draw(func(p *ui.Painter, r ui.Rect) {
			shape(p, r, tick{phase: loopPhase(p, ms(period), still), still: still})
		})
}

// BarsLoader is three bars rising and falling one after another, the way a
// level meter that has nothing to read still moves. It is the default
// spinner: it reads at any size, which is why the others are alternatives
// rather than upgrades.
func BarsLoader(c *ui.Context, opts LoaderOptions) *ui.Element {
	return loader(c, opts, func(p *ui.Painter, r ui.Rect, t tick) {
		bars(p, r, loaderInk(opts, p.Theme()), t, barsCount, barHeight)
	})
}

// DotsLoader is three dots, one brightening at a time in turn. It is the
// quietest of the five: at small sizes, beside a caption, or inside a status
// pill where a jumping shape would fight the pill for attention.
func DotsLoader(c *ui.Context, opts LoaderOptions) *ui.Element {
	return loader(c, opts, func(p *ui.Painter, r ui.Rect, t tick) {
		dots(p, r, loaderInk(opts, p.Theme()), t, dotsCount)
	})
}

// OrbitLoader is a ring with one bright dot going round it. It is the only
// one of the five whose shape says which way time is going, so it is the one
// to reach for where a wait has a length the user is watching — a boot, an
// import, a build.
func OrbitLoader(c *ui.Context, opts LoaderOptions) *ui.Element {
	return loader(c, opts, func(p *ui.Painter, r ui.Rect, t tick) {
		col := loaderInk(opts, p.Theme())
		rad := min(r.W, r.H) / 2
		dot := rad * 0.28
		cx, cy := r.X+rad, r.Y+rad
		// The ring is a track, so the dot reads as going round something
		// rather than floating in a gap.
		internal.Ring(p, cx, cy, rad-dot, max(1, rad*0.16), col.Alpha(0.22))
		x, y := orbitPoint(cx, cy, rad-dot, orbitAngle(t))
		internal.Dot(p, x, y, dot, col)
	})
}

// PulseLoader is one dot breathing. It is for the smallest place a loader
// can go — a row, a cell, a button's trailing edge — where three marks of any
// kind would be noise and only one changing mark still reads as working.
func PulseLoader(c *ui.Context, opts LoaderOptions) *ui.Element {
	return loader(c, opts, func(p *ui.Painter, r ui.Rect, t tick) {
		col := loaderInk(opts, p.Theme())
		rad := min(r.W, r.H) / 2
		grow := pulseSize(t)
		internal.Dot(p, r.X+rad, r.Y+rad, rad*grow, col.Alpha(0.45+0.55*grow))
	})
}

// WaveLoader is four bars at different phases, so the run reads as a waveform
// rather than as one bar bouncing. It is the loader for work with no end in
// sight — a stream still arriving, an export — where a single back-and-forth
// would read as "almost done".
func WaveLoader(c *ui.Context, opts LoaderOptions) *ui.Element {
	return loader(c, opts, func(p *ui.Painter, r ui.Rect, t tick) {
		bars(p, r, loaderInk(opts, p.Theme()), t, waveCount, waveHeight)
	})
}

// loaderInk is the colour a loader draws with. It is resolved again inside
// the paint rather than captured, so a loader repainting across a palette
// change takes the colour the window has now rather than the one it was
// built with.
func loaderInk(opts LoaderOptions, t *ui.Theme) ui.Color {
	if opts.Color.A != 0 {
		return opts.Color
	}
	return t.Accent
}

// barHeight returns how tall bar i of a BarsLoader stands, as a fraction of
// the loader's height. It is a pure function of the tick so the shape can be
// checked without a clock: the three bars are a third of a cycle apart, so
// the run reads as a wave going across rather than as three things pulsing
// together. When the window asked for no motion they all stand at restBar —
// level, which is a shape rather than a frame of the wave.
func barHeight(t tick, i int) float32 {
	if t.still {
		return restBar
	}
	return restFloor + (1-restFloor)*wave(t.phase-float32(i)/float32(barsCount))
}

// waveHeight returns how tall bar i of a WaveLoader stands. Unlike the bar
// run the height is a sine through zero rather than a raised cosine, so the
// run has both crests and troughs and reads as a waveform rather than as
// three shapes taking turns to be tall. Still, it is level like the bar run's.
func waveHeight(t tick, i int) float32 {
	if t.still {
		return restBar
	}
	s := math.Sin(float64(twoPi * (t.phase + float32(i)/float32(waveCount))))
	return restFloor + (1-restFloor)*(0.5+0.5*float32(s))
}

// dotAlpha returns how inked dot i of a DotsLoader is. The same raised
// cosine the bar run uses, so the two loaders move together rather than each
// carrying a clock of its own. Still, the three dots are equally inked.
func dotAlpha(t tick, i int) float32 {
	if t.still {
		return restDot
	}
	return restDip + (1-restDip)*wave(t.phase-float32(i)/float32(dotsCount))
}

// orbitAngle returns where on the circle an OrbitLoader's dot is, in
// radians. Straight up when still, because "at the top" is the one place on
// a ring that does not read as a dot that fell off it.
func orbitAngle(t tick) float32 {
	if t.still {
		return restAngle
	}
	return t.phase*twoPi + restAngle
}

// pulseSize returns how far a PulseLoader's dot is drawn across its box, 0
// to 1. Still, it is drawn at restRadius rather than at the start or the end
// of its breath, where it would read as a dot that failed to grow.
func pulseSize(t tick) float32 {
	if t.still {
		return restRadius
	}
	return wave(t.phase)
}

// bars draws a run of n bars bottom-aligned in r, each as tall as height
// says. They are bottom-aligned because a bar run is a level reading: the
// marks move and the floor stays put, which is what lets a half-full meter
// and a spinner be the same shape.
func bars(p *ui.Painter, r ui.Rect, col ui.Color, t tick, n int, height func(tick, int) float32) {
	gap := r.W * 0.16
	bar := (r.W - gap*float32(n-1)) / float32(n)
	if bar <= 0 {
		return
	}
	for i := range n {
		h := r.H * clamp01(height(t, i))
		x := r.X + float32(i)*(bar+gap)
		p.Fill(ui.Rect{X: x, Y: r.Y + r.H - h, W: bar, H: h}, col, bar/2)
	}
}

// dots draws a run of n dots across the middle of r, each at the alpha
// dotAlpha gives it.
func dots(p *ui.Painter, r ui.Rect, col ui.Color, t tick, n int) {
	rad := min(r.W/float32(n), r.H) * 0.3
	if rad <= 0 {
		return
	}
	for i := range n {
		x := r.X + r.W*(float32(i)+0.5)/float32(n)
		internal.Dot(p, x, r.Y+r.H/2, rad, col.Alpha(dotAlpha(t, i)))
	}
}

// orbitPoint returns where on a circle of radius rad about (cx, cy) the
// angle points, in the radians Painter's circle paths take. It is separate
// from the drawing so a test can say where an orbit's dot rests without
// painting anything.
func orbitPoint(cx, cy, rad, ang float32) (x, y float32) {
	s, c := math.Sincos(float64(ang))
	return cx + rad*float32(c), cy + rad*float32(s)
}

// clamp01 keeps a value inside 0 to 1, the range every alpha and fraction in
// this package is written in.
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
