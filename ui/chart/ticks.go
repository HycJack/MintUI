package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// DefaultTickCount is how many steps an axis asks for when nobody says.
const DefaultTickCount = 5

// Tick is one line on an axis: the value it stands for, where it falls on the
// axis in window DIPs, and the label it would carry.
//
// Drawn says whether that label is one of the ones actually written. The
// ticks are computed for the whole range and then thinned to what fits, so a
// grid line can stand on a tick whose label was dropped — the value is still
// there to draw a line through, only the words ran out of room.
type Tick struct {
	// Value is the data value this tick stands for: an index on a band axis.
	Value float64
	// Pos is where it falls on the axis, in window DIPs.
	Pos float32
	// Label is what it says, formatted.
	Label string
	// Drawn reports whether the label is one of the ones written.
	Drawn bool
}

// BandTick is one category of a categorical axis: its index, its name, and
// the share of the axis it takes, which a bar chart draws inside.
type BandTick struct {
	Index int
	Label string
	// Start, Mid and End are the share's two edges and its middle, along the
	// axis.
	Start, Mid, End float32
}

// Nice rounds a range out to multiples of a 1, 2 or 5 times a power of ten, at
// about count steps across it: 0 to 97 becomes 0 to 100, and its ticks are
// 0, 20, 40, 60, 80, 100 rather than 0, 19.4, 38.8, 58.2 ….
//
// It is the first thing to reach for before plotting, because it is what
// makes a chart's ends look like somebody chose them. A range of nothing —
// every value the same — is widened by one either side, so a flat series
// still has an axis to sit on.
func Nice(lo, hi float64, count int) Domain {
	if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) || lo > hi {
		panic("chart: Nice needs a range that is not upside down or not a number")
	}
	if count < 1 {
		count = 1
	}
	if hi == lo {
		lo, hi = lo-1, hi+1
	}
	step := niceStep((hi - lo) / float64(count))
	return Domain{Min: math.Floor(lo/step) * step, Max: math.Ceil(hi/step) * step}
}

// niceStep is the round step at or above raw: a power of ten, twice one, or
// five times one, which is the series of steps that put 1 before 2 before 5
// before the next decade — the order people read numbers in.
func niceStep(raw float64) float64 {
	if raw <= 0 || math.IsInf(raw, 0) || math.IsNaN(raw) {
		return 1
	}
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	switch norm := raw / mag; {
	case norm <= 1:
		return mag
	case norm <= 2:
		return 2 * mag
	case norm <= 5:
		return 5 * mag
	}
	return 10 * mag
}

// Linear returns count+1 values evenly spaced across d, both ends included.
// It divides what it is given and rounds nothing, so a range that is already
// round stays exactly as it is and a range that is not shows it. Nice it
// first when that is what you want.
func Linear(d Domain, count int) []float64 {
	if count < 1 {
		count = 1
	}
	if d.Span() == 0 {
		return []float64{d.Min}
	}
	out := make([]float64, count+1)
	for i := range out {
		out[i] = d.Min + d.Span()*float64(i)/float64(count)
	}
	out[count] = d.Max
	return out
}

// Log returns the values a logarithmic axis over lo to hi draws: every power
// of base in range, and, for a base of ten, the 2 and 5 of each decade, which
// is what gives a log scale something between the decades to hang a grid line
// on.
//
// Zero and negative values have no logarithm, so a range containing one is a
// mistake rather than something to quietly drop.
func Log(lo, hi, base float64) []float64 {
	if lo <= 0 || hi < lo || math.IsNaN(lo) || math.IsNaN(hi) {
		panic("chart: a log axis needs a positive range that is not upside down")
	}
	switch base {
	case 2, 10:
	default:
		base = 10
	}
	steps := []float64{1}
	if base == 10 {
		steps = append(steps, 2, 5)
	}
	var out []float64
	for e := math.Floor(math.Log(lo) / math.Log(base)); ; e++ {
		power := math.Pow(base, e)
		for _, m := range steps {
			v := m * power
			if v < lo {
				continue
			}
			if v > hi {
				return out
			}
			out = append(out, v)
		}
		// A loop bounded by the range, not by any number of decades: a
		// misjudged log10 would otherwise run for ever.
		if power > hi {
			return out
		}
	}
}

// Band divides the axis from from to to into one share per label and returns
// them: a bar chart's bands, and a categorical axis's ticks.
func Band(labels []string, from, to float32) []BandTick {
	if len(labels) == 0 {
		panic("chart: a band axis needs at least one category")
	}
	w := (to - from) / float32(len(labels))
	out := make([]BandTick, len(labels))
	for i, name := range labels {
		start := from + float32(i)*w
		out[i] = BandTick{Index: i, Label: name, Start: start, Mid: start + w/2, End: start + w}
	}
	return out
}

// TicksFor is the pure half of an axis: the ticks a scale draws, with their
// positions and labels, all of them marked drawn. Nothing is dropped yet —
// measuring the labels and thinning the ticks is [Ticks]'s job, because it is
// the one that can measure.
//
// count is how many steps the axis asks for; a scale with a thousand values
// between them still asks for five or six lines.
func TicksFor(s Scale, count int, f Format) []Tick {
	if !s.OK() {
		// An axis with no scale has no ticks: the zero Scale is what a
		// caller leaves in an axis it does not want, not a scale over 0 to 0.
		return nil
	}
	if f == nil {
		f = Auto()
	}
	if count < 1 {
		count = DefaultTickCount
	}
	var values []float64
	switch s.Kind() {
	case KindBand:
		bands := Band(s.Labels(), s.From(), s.To())
		ticks := make([]Tick, len(bands))
		for i, b := range bands {
			ticks[i] = Tick{Value: float64(b.Index), Pos: b.Mid, Label: b.Label, Drawn: true}
		}
		return ticks
	case KindLog:
		d := s.Domain()
		values = Log(d.Min, d.Max, s.Base())
	default:
		d := s.Domain()
		values = Linear(d, count)
	}
	step := 0.0
	if n := len(values); n > 1 {
		step = (values[n-1] - values[0]) / float64(n-1)
	}
	ticks := make([]Tick, len(values))
	for i, v := range values {
		ticks[i] = Tick{Value: v, Pos: s.At(v), Label: f(v, step), Drawn: true}
	}
	return ticks
}

// LabelSkip marks which of a run of labels can be written without two of them
// touching: the ones drawn are true, and a label is dropped from the middle
// rather than from the end, so the first and the last — the two a reader
// looks for — stay.
//
// The run is read in the direction it goes. A y axis' ticks come back from
// the bottom of the plot, and a pass that only knew about left to right would
// drop every label of a vertical axis after the first.
//
// The last label is put back afterwards, dropping the one before it if that
// is what it takes: an axis that ends at an unlabelled tick looks broken at
// exactly the place a reader looks to see how far the data goes. The first is
// never the one dropped; when the two ends cannot both fit, the last goes.
func LabelSkip(positions, widths []float32, gap float32) []bool {
	if len(positions) != len(widths) {
		panic("chart: LabelSkip needs one width for every position")
	}
	n := len(positions)
	drawn := make([]bool, n)
	// The run is read from the scale's first value to its last, which on a y
	// axis is from the bottom up: the positions come back the other way, so
	// they are turned round before anything is measured against them.
	dir := float32(1)
	if n > 1 && positions[n-1] < positions[0] {
		dir = -1
	}
	// kept are the labels drawn so far, each with the position its right edge
	// reached — which is what the next one has to clear.
	type keptLabel struct {
		i   int
		end float32
	}
	var kept []keptLabel
	end := float32(math.Inf(-1))
	for i := range n {
		// Labels are written centred on their tick, so each runs half its
		// width either side of it.
		start := dir*positions[i] - widths[i]/2
		if start < end+gap {
			continue
		}
		drawn[i] = true
		end = start + widths[i]
		kept = append(kept, keptLabel{i: i, end: end})
	}
	if n > 0 && !drawn[n-1] && len(kept) > 0 {
		// Putting the last label back means dropping the one before it, so
		// what has to clear is the label before that one. The first is never
		// the one dropped: an axis that starts unlabelled looks as broken as
		// one that ends so, and when the two ends cannot both fit it is the
		// last that goes.
		drop := kept[len(kept)-1].i
		room := float32(math.Inf(-1))
		if len(kept) > 1 {
			room = kept[len(kept)-2].end
		}
		if drop != 0 && dir*positions[n-1]-widths[n-1]/2 >= room+gap {
			drawn[drop] = false
			drawn[n-1] = true
		}
	}
	return drawn
}

// TickOptions configure [Ticks].
type TickOptions struct {
	// Count is how many steps the axis asks for; zero means
	// [DefaultTickCount].
	Count int
	// Format renders the values; nil uses [Auto].
	Format Format
	// Gap is the room to leave between two labels; zero leaves one density
	// step, so labels never touch.
	Gap float32
	// Size is the label's font size; zero uses [theme.CaptionSize].
	Size float32
	// Vertical says the ticks run up or down the plot — a y axis — so two
	// labels collide along their height, not their width. An axis that left
	// it unset measures a wide number against a short label's worth of room
	// and drops ticks that would have fitted with room to spare.
	Vertical bool
}

// Ticks returns [TicksFor]'s ticks with their labels measured and the ones
// that would collide dropped, which is the set an axis writes and a grid
// draws.
func Ticks(c *ui.Context, s Scale, opts TickOptions) []Tick {
	ticks := TicksFor(s, opts.Count, opts.Format)
	size := opts.Size
	if size <= 0 {
		size = theme.CaptionSize
	}
	gap := opts.Gap
	if gap <= 0 {
		gap = core.Density(c).Unit()
	}
	widths := make([]float32, len(ticks))
	positions := make([]float32, len(ticks))
	for i, t := range ticks {
		// The extent a label reaches along the tick run: across the plot for
		// an x axis, up it for a y one — where every label is about as tall
		// as every other, and a wide number must not cost its neighbours
		// their tick.
		if opts.Vertical {
			widths[i] = LabelHeight(c, size)
		} else {
			widths[i] = LabelWidth(c, t.Label, size)
		}
		positions[i] = t.Pos
	}
	for i, keep := range LabelSkip(positions, widths, gap) {
		ticks[i].Drawn = keep
	}
	return ticks
}

// LabelWidth is how much room a tick label takes at that size, as an axis
// writes it — measured, not guessed, because a gutter sized by a guess is a
// gutter that clips the largest number in the chart.
func LabelWidth(c *ui.Context, label string, size float32) float32 {
	if label == "" {
		return 0
	}
	w, _ := c.MeasureText(0, ui.Span{Text: label, Size: size})
	return w
}

// LabelHeight is how tall a line of tick labels takes at that size. It is the
// measurement of a real glyph rather than the font's nominal size, so the
// bottom gutter holds the text and its leading and no more.
func LabelHeight(c *ui.Context, size float32) float32 {
	_, h := c.MeasureText(0, ui.Span{Text: "0", Size: size})
	return h
}

// Widest is the width of the largest of a set of ticks' labels, and what a
// frame's left gutter is measured from: the y axis has to be as wide as the
// biggest number on it, not as wide as a guess.
func Widest(c *ui.Context, ticks []Tick, size float32) float32 {
	var w float32
	for _, t := range ticks {
		w = max(w, LabelWidth(c, t.Label, size))
	}
	return w
}
