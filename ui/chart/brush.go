package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// Range is a selection on a chart's x axis, and the drag that made it.
//
// It is the caller's value rather than something the brush keeps, for the
// same reason the hover is: a chart's brush, its crosshair and its tooltip
// all read the same pointer, and a caller that wants the selection — to
// filter a table by it, to put it in the URL — has it without asking a widget
// what it thought the user chose.
//
// [Range.Min] and [Range.Max] are written as the brush paints, which is the
// only moment the selection's place in the plot is known; a caller reading
// them while it builds a view sees the selection as it was a frame ago, which
// is the frame the reader has been looking at.
type Range struct {
	// Min and Max are the ends of the selection in the x axis' values, in
	// order however the drag went.
	Min, Max float64
	// Active reports whether there is a selection. A brush that has never
	// been dragged leaves it false, which is how a caller tells "everything"
	// from "this part of it".
	Active bool
	// Dragging reports a drag in progress, so a caller can tell a brush being
	// pulled from one that has settled.
	Dragging bool

	// The brush's own bookkeeping, in window DIPs: where the drag began and
	// how far it has reached. It lives here rather than in the brush because
	// the brush holds no state between frames.
	anchor, from, to float32
	moved            bool
}

// Span is the selected range, and is zero when there is no selection.
func (r Range) Span() float64 { return r.Max - r.Min }

// Reset clears the selection, which is what a double press on a brush does
// and what a caller does when the filters it came from change.
func (r *Range) Reset() {
	r.Min, r.Max, r.Active, r.Dragging = 0, 0, false, false
}

// BrushOptions configure a [Brush].
type BrushOptions struct {
	// Frame is the chart's frame, whose plot area the selection covers.
	Frame FrameResult
	// X is the axis being brushed: the selection is a range of its values,
	// which is why a chart brushes one axis and not two. A drag may start
	// anywhere over the chart — including in its gutters — and what it makes
	// is clamped to the axis' range.
	X Scale
	// Range is the caller's selection. The brush writes it and the caller
	// reads it; a nil one is a brush that has nowhere to put what it caught,
	// which is a mistake rather than an inert control.
	Range *Range
	// Color is what the selection is washed in; the zero colour uses the
	// accent, faded back so the data under it still reads.
	Color ui.Color
	// Handles draws the two grips at the ends of the selection.
	Handles bool
	// Min is the smallest drag that counts, in DIPs; zero uses four. A press
	// and a release in the same place is a click, not a range, and clears the
	// selection instead.
	Min float32
}

// BrushResult carries a [Brush] and what the drag in it did.
type BrushResult struct {
	// Element is the brush.
	Element *ui.Element
	changed bool
}

// Changed reports the selection having moved this frame. The selection itself
// is in the caller's [Range] either way; this is for a caller that wants to
// do something once per drag rather than every frame.
func (r BrushResult) Changed() bool { return r.changed }

// Brush is a drag across a chart that selects a range of its x axis.
//
// The values it catches go to the caller's [Range] and nowhere else: the
// brush draws the selection over the chart, and what it means — filter a
// table by it, zoom to it, fetch only that week — is the caller's business.
// The drag's own bookkeeping is in that same struct, so the brush survives a
// redraw without keeping anything.
func Brush(c *ui.Context, opts BrushOptions) BrushResult {
	k, step := core.Tokens(c), core.Density(c).Unit()
	if opts.Range == nil {
		panic("chart: Brush needs a Range to put what it catches")
	}
	color := opts.Color
	if color == (ui.Color{}) {
		color = k.Accent
	}
	min := opts.Min
	if min <= 0 {
		min = step
	}
	r := opts.Range

	e := ui.Box(c).Absolute().Fill().Label("Brush").
		Cursor(ui.CursorResizeColumn)
	// Input is read while the chart is being built, which is when MyGo asks
	// elements where the pointer is and what it is doing to them. The bounds
	// are this frame's — or the last frame's, on the first — which is the
	// same frame of reference the pointer is given in.
	b := e.Bounds()
	x, _, over := e.PointerPosition()
	at := b.X + x
	pressed := e.Pressed()

	var res BrushResult
	if over && pressed && !r.Dragging {
		r.Dragging, r.anchor, r.from, r.to = true, at, at, at
		r.Active = false
	}
	if r.Dragging {
		if at != r.to {
			r.to = at
			r.moved = true
			res.changed = true
		}
		r.from = r.anchor
		if absf32(r.to-r.from) >= min {
			r.Active = true
		}
		if !pressed {
			// The drag is over: a press that never moved was a click, and a
			// click on a brush clears it.
			r.Dragging = false
			if !r.moved || absf32(r.to-r.from) < min {
				r.Reset()
				res.changed = res.changed || r.moved
			}
		}
	}
	r.moved = r.Dragging

	res.Element = e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := opts.Frame.Plot()
		if plot.W <= 0 || plot.H <= 0 || !opts.X.OK() || !r.Active {
			return
		}
		xs := opts.X.Span(plot.X, plot.X+plot.W)
		// The values are written here, where the plot's size is known.
		lo, hi := order(r.from, r.to)
		r.Min, r.Max = xs.Value(lo), xs.Value(hi)
		if r.Min > r.Max {
			r.Min, r.Max = r.Max, r.Min
		}
		left, right := xs.At(r.Min), xs.At(r.Max)
		if right < left {
			left, right = right, left
		}
		width := right - left
		if width < 1 {
			width = 1
		}
		p.Clip(plot, 0, func() {
			band := ui.Rect{X: left, Y: plot.Y, W: width, H: plot.H}
			p.Fill(band, color.Alpha(0.12), 0)
			if opts.Handles {
				for _, x := range []float32{left, right} {
					p.Fill(ui.Rect{X: x - step/4, Y: plot.Y, W: step / 2, H: plot.H}, color.Alpha(0.5), 0)
				}
			}
		})
	})
	return res
}

func order(a, b float32) (lo, hi float32) {
	if a <= b {
		return a, b
	}
	return b, a
}

func absf32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func min64(a, b float64) float64 { return math.Min(a, b) }
func max64(a, b float64) float64 { return math.Max(a, b) }
