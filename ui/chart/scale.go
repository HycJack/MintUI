package chart

import "math"

// AxisKind is how a scale turns values into positions.
type AxisKind uint8

const (
	// KindLinear spreads values evenly: 0 to 100 covers the axis twice over.
	KindLinear AxisKind = iota
	// KindLog is a logarithmic scale, for data that spans decades: 1, 10,
	// 100 take the same room each. Its domain has to be positive.
	KindLog
	// KindBand is a categorical scale. Its values are the indices of the
	// categories — 0, 1, 2 — and each takes an equal share of the axis, with
	// its tick in the middle of that share rather than on its edge, so a bar
	// chart's bars have room on both sides.
	KindBand
)

func (k AxisKind) String() string {
	switch k {
	case KindLog:
		return "Log"
	case KindBand:
		return "Band"
	}
	return "Linear"
}

// Scale maps values to positions along an axis, and back.
//
// A scale is built before the chart knows how big it is: the frame measures
// its plot area as it paints. So a scale carries the axis' own coordinate
// space, from [Scale.From] to [Scale.To], and [Scale.Span] returns the same
// scale placed across a plot area:
//
//	xs := chart.NewLinear(chart.Nice(0, 97, 5), 0, 1)   // a proportion
//	// …as the chart paints, with plot being the frame's plot area:
//	x := xs.Span(plot.X, plot.X + plot.W)                 // in window DIPs
//
// An x scale runs left to right and a y scale runs bottom to top, which is
// why the second Span of a y scale gives its ends the other way round.
type Scale struct {
	kind   AxisKind
	d      Domain
	base   float64
	labels []string
	from   float32
	to     float32
	// ok says a constructor built this. The zero Scale is not a scale over
	// anything, and a caller who leaves an axis out writes it.
	ok bool
}

// NewLinear is a scale over d, spread evenly from from to to.
func NewLinear(d Domain, from, to float32) Scale {
	if !d.Valid() {
		panic("chart: a linear scale needs a domain that is not upside down or not a number")
	}
	return Scale{kind: KindLinear, d: d, from: from, to: to, ok: true}
}

// NewLog is a scale over a positive domain, spread by logarithm. The base is
// 10 unless base is one of 2 or 10 — a log axis that reads as anything else
// is a chart nobody can read off.
//
// Zero and negative values have no place on a log axis, so a domain
// containing one is a mistake rather than something to clamp.
func NewLog(d Domain, base float64, from, to float32) Scale {
	if !d.Valid() {
		panic("chart: a log scale needs a domain that is not upside down or not a number")
	}
	if d.Min <= 0 {
		panic("chart: a log scale needs a positive domain; zero has no logarithm")
	}
	switch base {
	case 2, 10:
	default:
		base = 10
	}
	return Scale{kind: KindLog, d: d, base: base, from: from, to: to, ok: true}
}

// NewBand is a scale over labels, one equal share each. It panics without
// labels: a categorical axis with no categories has nothing to draw.
func NewBand(labels []string, from, to float32) Scale {
	if len(labels) == 0 {
		panic("chart: a band scale needs at least one category")
	}
	return Scale{
		kind:   KindBand,
		d:      Domain{Min: 0, Max: float64(len(labels) - 1)},
		labels: append([]string(nil), labels...),
		from:   from,
		to:     to,
		ok:     true,
	}
}

// OK reports whether a constructor built this scale. A caller who leaves an
// axis out writes the zero [Scale], which is not a scale over anything, and a
// chart with no x axis is an ordinary thing to draw.
func (s Scale) OK() bool { return s.ok }

// Kind is what kind of scale this is.
func (s Scale) Kind() AxisKind { return s.kind }

// Domain is the range of values the scale covers: the indices, 0 to the last
// category, for a band scale.
func (s Scale) Domain() Domain { return s.d }

// Base is the base of a log scale, and 1 for the others.
func (s Scale) Base() float64 {
	if s.kind == KindLog {
		return s.base
	}
	return 1
}

// Labels are a band scale's categories, and nil for the others.
func (s Scale) Labels() []string { return s.labels }

// From and To are the axis positions the scale's ends sit at, in the axis'
// own coordinate space.
func (s Scale) From() float32 { return s.from }
func (s Scale) To() float32   { return s.to }

// Span returns the same scale placed from lo to hi: the plot area's left and
// right edges for an x scale, its bottom and top for a y scale. It is the one
// call a component makes to turn the scale a caller handed it into the pixels
// of this frame.
func (s Scale) Span(lo, hi float32) Scale {
	s.from, s.to = lo, hi
	return s
}

// At is where v falls on the axis. Values outside the domain are not
// clamped: a series may well draw past its axis' range, and where it does is
// the caller's business, not this one's.
func (s Scale) At(v float64) float32 {
	return s.from + s.t(v)*(s.to-s.from)
}

// t is where v falls as a fraction of the axis: 0 at the domain's first end,
// 1 at its last. For a y scale those are the bottom and the top.
func (s Scale) t(v float64) float32 {
	switch s.kind {
	case KindLog:
		if v <= 0 || s.d.Min <= 0 {
			return 0
		}
		return float32(math.Log(v/s.d.Min) / math.Log(s.d.Max/s.d.Min))
	case KindBand:
		n := float32(len(s.labels))
		if n <= 1 {
			return 0
		}
		i := float32(v) + 0.5
		return i / n
	}
	if s.d.Span() == 0 {
		return 0
	}
	return float32((v - s.d.Min) / s.d.Span())
}

// Value is the value under the axis position at: the inverse of [Scale.At],
// and what a hover or a brush turns a pointer's pixel back into.
//
// A band scale answers with the index of the category under at, or -1 when
// at is off the ends of the axis.
func (s Scale) Value(at float32) float64 {
	if s.kind == KindBand {
		return float64(s.Index(at))
	}
	span := s.to - s.from
	if span == 0 {
		return s.d.Min
	}
	f := float64((at - s.from) / span)
	switch s.kind {
	case KindLog:
		return s.d.Min * math.Pow(s.d.Max/s.d.Min, f)
	}
	return s.d.Min + f*s.d.Span()
}

// Index is the category a band scale's position falls in, and -1 for a scale
// that has no categories or a position off its ends. It is how a caller picks
// the datum a bar or a column belongs to.
func (s Scale) Index(at float32) int {
	if s.kind != KindBand || len(s.labels) == 0 {
		return -1
	}
	n := len(s.labels)
	w := (s.to - s.from) / float32(n)
	if w == 0 {
		return -1
	}
	i := int(math.Floor(float64((at - s.from) / w)))
	if i < 0 || i >= n {
		return -1
	}
	return i
}

// Share is the slice of the axis category i takes: its two edges and its
// middle. A bar chart draws its bars inside it; an index of -1 gives nothing,
// which is what a caller should expect off the ends of the axis.
func (s Scale) Share(i int) (start, mid, end float32) {
	if s.kind != KindBand || i < 0 || i >= len(s.labels) {
		return 0, 0, 0
	}
	w := (s.to - s.from) / float32(len(s.labels))
	start = s.from + float32(i)*w
	return start, start + w/2, start + w
}
