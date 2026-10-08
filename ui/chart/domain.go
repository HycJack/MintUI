package chart

import "math"

// Domain is a closed interval of values: what a chart shows.
//
// It is the one thing a scale, a tick set and a hover all agree on, so it is
// built from the data rather than guessed at: [DomainOf] finds the span of a
// slice, and [Nice] rounds it out to numbers a person would choose.
type Domain struct {
	Min, Max float64
}

// Valid reports whether d is an interval a scale can map: finite, and not
// upside down.
func (d Domain) Valid() bool {
	return !math.IsNaN(d.Min) && !math.IsNaN(d.Max) &&
		!math.IsInf(d.Min, 0) && !math.IsInf(d.Max, 0) && d.Min <= d.Max
}

// Span is how much room d covers. A flat domain spans nothing, which every
// caller has to survive rather than divide by.
func (d Domain) Span() float64 { return d.Max - d.Min }

// Union is the span of d and o, and whether either of them is usable. A
// domain built from values is only as good as the values in it, so the
// result says so instead of pretending an empty slice is 0 to 0.
func (d Domain) Union(o Domain) Domain {
	if !d.Valid() {
		return o
	}
	if !o.Valid() {
		return d
	}
	return Domain{Min: math.Min(d.Min, o.Min), Max: math.Max(d.Max, o.Max)}
}

// DomainOf is the span of vs, ignoring the values that are not numbers, and
// whether it holds any. Callers that plot data panic on the false rather than
// charting an empty rectangle — a chart of nothing is a mistake, and [Empty]
// is the way to say so on screen.
func DomainOf(vs []float64) (Domain, bool) {
	var d Domain
	seen := false
	for _, v := range vs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if !seen {
			d.Min, d.Max, seen = v, v, true
			continue
		}
		d.Min = math.Min(d.Min, v)
		d.Max = math.Max(d.Max, v)
	}
	return d, seen
}
