package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Inset is the space a [Frame] keeps around its plot area: how far in from
// each edge the drawing starts.
//
// A chart's margins are not decoration — they are where its numbers live — so
// they are measured rather than guessed, and each side has a job:
//
//	Left    the y axis: the widest label on it, the tick marks and the gap to
//	        the plot. It is the one number that cannot be a constant, because
//	        the widest label of one chart is a date and of the next is
//	        "1,200,000".
//	Bottom  the x axis: the height of one line of labels, the tick marks and
//	        the same gap.
//	Top     the legend, and the y axis' name above it.
//	Right   nothing. A chart's right edge is where its data stops, and a
//	        gutter there only pushes the last point away from the edge the
//	        reader is looking at.
type Inset struct {
	Top, Right, Bottom, Left float32
}

// PlotRect is the drawing area inside outer: outer less the inset, and never
// larger than outer — a frame too small for its axes draws no plot rather
// than an upside-down one.
func PlotRect(outer ui.Rect, in Inset) ui.Rect {
	plot := ui.Rect{
		X: outer.X + in.Left,
		Y: outer.Y + in.Top,
		W: outer.W - in.Left - in.Right,
		H: outer.H - in.Top - in.Bottom,
	}
	if plot.W < 0 {
		plot.W = 0
	}
	if plot.H < 0 {
		plot.H = 0
	}
	return plot
}

// Hover is a pointer's place on a chart: where it is and what it is over.
//
// It is the caller's, not the library's. A chart owns one, hands it to [Frame]
// to be filled and passes it on to the pieces that follow the pointer —
// [Crosshair], [Tooltip], [Brush] — and nothing keeps it between frames. That
// is what lets a chart's crosshair, tooltip and selection agree with each
// other: they are all reading one value rather than each guessing where the
// pointer is.
type Hover struct {
	// X, Y are the pointer's position in the window, which is the plot area's
	// own coordinates: the frame places them there as it paints.
	X, Y float32
	// Over reports that the pointer is inside the plot area. A crosshair that
	// followed the pointer into the gutter would be a chart marking the axis.
	Over bool
	// ValueX and ValueY are the data values under the pointer, read through
	// the frame's scales: a category's index on a band axis, and otherwise
	// the number that value stands for.
	ValueX, ValueY float64
	// Index is the category under the pointer, and -1 when the x axis is not
	// a band one or the pointer is off it.
	Index int
}

// FrameOptions configure a [Frame].
type FrameOptions struct {
	// Height gives the frame its own height; zero lets the layout give it one,
	// which is what a chart in a column wants.
	Height float32
	// Pad is the inner padding between the frame's edge and everything in it.
	Pad float32
	// Label names the chart for assistive technology. A chart's numbers are
	// drawn rather than laid out, so the name is the only thing a screen
	// reader has to go on: give it one.
	Label string
	// X and Y are the two axes, in the frame's own coordinate space — the
	// frame places them in the plot as it paints. An axis with no scale draws
	// nothing.
	X, Y AxisOptions
	// Legend sits in the top gutter. Entries in it and the frame's height
	// decide how much room it takes.
	Legend LegendOptions
	// Hover, when given, is the caller's pointer: the frame fills where it is
	// and what it is over as it paints, and the pieces inside the frame read
	// it as they paint, after it.
	Hover *Hover
	// Surface draws the frame a step below the window, for a chart that is a
	// panel rather than part of the page.
	Surface bool
	// Border draws a hairline around the frame.
	Border bool
	// Radius rounds the frame's own surface; zero uses [theme.SmallRadius]
	// when it has one.
	Radius float32
}

// FrameResult carries a [Frame] and the layout it worked out.
type FrameResult struct {
	// Element is the frame itself: put it in a layout like any other box.
	Element *ui.Element

	f *frame
}

// frame is what a frame works out as it paints, shared with the pieces inside
// it by pointer. They paint after the frame does — a parent's drawing runs
// before its children's — so by the time they ask, the plot area is known.
type frame struct {
	plot   ui.Rect
	bounds ui.Rect
	inset  Inset
	x, y   Scale
}

// Plot is the area everything inside the frame is drawn in: the frame's box
// less the gutters its axes and legend took.
//
// It is filled as the frame paints, so it is worth reading inside a Draw
// callback — which is what [Grid], [Plot] and the rest do, and what a caller
// does in the frame's own children. Read before the frame has painted, it is
// the zero rect and nothing is drawn.
func (r FrameResult) Plot() ui.Rect {
	if r.f == nil {
		return ui.Rect{}
	}
	return r.f.plot
}

// Bounds is the frame's own box, the plot area's frame of reference: where
// the axes and the legend live.
func (r FrameResult) Bounds() ui.Rect {
	if r.f == nil {
		return ui.Rect{}
	}
	return r.f.bounds
}

// Inset is the space the frame kept for its axes and legend, worked out
// before it paints.
func (r FrameResult) Inset() Inset {
	if r.f == nil {
		return Inset{}
	}
	return r.f.inset
}

// X and Y are the axes' scales placed in the plot area, filled as the frame
// paints like [FrameResult.Plot]. A caller that draws something of its own
// over the chart needs those two rather than the scales it handed in.
func (r FrameResult) X() Scale {
	if r.f == nil {
		return Scale{}
	}
	return r.f.x
}

func (r FrameResult) Y() Scale {
	if r.f == nil {
		return Scale{}
	}
	return r.f.y
}

// Frame is a chart's box: the surface and border around it, the legend in its
// top gutter, the axes down its left and under its bottom, and the plot area
// left over for the data.
//
//	chart.Frame(c, chart.FrameOptions{
//		Height: 220,
//		X:      chart.AxisOptions{Scale: xs, Count: 7, Label: "Day"},
//		Y:      chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 5, Label: "Callbacks"},
//		Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
//		Hover:  &hover,
//	}, func(frame chart.FrameResult) {
//		chart.Plot(c, chart.PlotOptions{Frame: frame, Series: series, X: xs, Y: ys})
//	})
//
// The scales are written in the axis' own space — 0 to 1 above, because the
// frame does not know how big it is until it paints — and the components
// inside it place them in the plot as they draw.
//
// children are the chart's own drawing. They are given the frame result, and
// each of them takes it: they read the plot area from it while painting,
// which is the only time it is known.
func Frame(c *ui.Context, opts FrameOptions, children func(FrameResult)) FrameResult {
	k := core.Tokens(c)
	inset := Measure(c, opts)
	f := &frame{inset: inset}
	res := FrameResult{f: f}

	bg := k.Background
	if opts.Surface {
		bg = k.Surface
	}
	e := ui.Box(c).FillWidth()
	if opts.Height > 0 {
		e.Height(opts.Height)
	}
	if opts.Surface || opts.Border {
		radius := opts.Radius
		if radius == 0 {
			radius = theme.SmallRadius
		}
		e.Background(bg).Radius(radius)
	}
	if opts.Border {
		e.BorderWidth(theme.BorderWidth).BorderColor(k.Border)
	}
	if opts.Label != "" {
		e = e.Label(opts.Label)
	}

	// The pointer is read here, while the frame is being built, because that
	// is when MyGo asks elements where it is; asking also gets the frame a
	// frame whenever the pointer moves over it, which is what a crosshair and
	// a tooltip need to follow along.
	px, py, over := e.PointerPosition()

	e = e.Draw(func(p *ui.Painter, r ui.Rect) {
		f.bounds = r
		f.plot = PlotRect(r, inset)
		f.x = opts.X.Scale.Span(f.plot.X, f.plot.X+f.plot.W)
		f.y = opts.Y.Scale.Span(f.plot.Y+f.plot.H, f.plot.Y)
		if h := opts.Hover; h != nil {
			h.X, h.Y = px+r.X, py+r.Y
			h.Over = over && f.plot.W > 0 && f.plot.H > 0 && f.plot.Contains(h.X, h.Y)
			h.ValueX, h.ValueY = f.x.Value(h.X), f.y.Value(h.Y)
			h.Index = f.x.Index(h.X)
		}
		// The plot's own two edges: down the left and along the bottom, the
		// way a chart is framed when it has no box drawn around it. The axes
		// draw their labels and marks; this is the line they hang from.
		if f.plot.W > 0 && f.plot.H > 0 {
			y := f.plot.Y + f.plot.H
			p.Line(f.plot.X, y, f.plot.X+f.plot.W, y, theme.BorderWidth, k.Border)
			p.Line(f.plot.X, f.plot.Y, f.plot.X, y, theme.BorderWidth, k.Border)
		}
	})

	e.Children(func() {
		if len(opts.Legend.Entries) > 0 {
			// A y axis' name is written at the head of the numbers it
			// names, which is the frame's own top-left corner — the same
			// corner the legend sits in. Measure already keeps room for
			// both, so the legend is pushed down by the name's height here
			// rather than the two being drawn on top of each other.
			name := float32(0)
			if opts.Y.Scale.OK() && opts.Y.Label != "" {
				name = LabelHeight(c, theme.RowSize) + core.Density(c).Unit()
			}
			row := ui.Box(c).FillWidth().Height(LegendHeight(c, opts.Legend) + name).
				Clip().Label("Series").Children(func() {
				Legend(c, opts.Legend)
			})
			if opts.Pad > 0 || name > 0 {
				row.Padding(opts.Pad+name, 0)
			}
		}
		if children != nil {
			children(res)
		}
	})

	res.Element = e
	return res
}

// Measure works out the inset a frame keeps for the axes and legend in opts,
// before anything is drawn: the left gutter from the y axis' widest label,
// the bottom from the x axis' label height, the top from the legend, the
// right from nothing.
//
// It is separate from [Frame] so a caller can size a chart's surroundings
// from the same numbers the frame used, and so a test can check the rule
// without a window.
func Measure(c *ui.Context, opts FrameOptions) Inset {
	u := core.Density(c).Unit()
	in := Inset{Right: opts.Pad, Top: opts.Pad, Bottom: opts.Pad, Left: opts.Pad}

	// The y axis: the widest of its labels is the one number that decides the
	// gutter, so it is measured rather than assumed.
	if opts.Y.Scale.OK() {
		size := tickSize(opts.Y)
		ticks := Ticks(c, opts.Y.Scale, TickOptions{Count: opts.Y.Count, Format: opts.Y.Format, Size: size})
		// The y axis takes its gutter on the side it is drawn on, which is
		// the left unless the caller put it on the right.
		if opts.Y.Side == Right {
			in.Right += Widest(c, ticks, size) + axisGap(c, opts.Y)
		} else {
			in.Left += Widest(c, ticks, size) + axisGap(c, opts.Y)
		}
	}
	if opts.X.Scale.OK() {
		height := LabelHeight(c, tickSize(opts.X)) + axisGap(c, opts.X)
		if opts.X.Side == Top {
			in.Top += height
		} else {
			in.Bottom += height
		}
	}
	// An axis' name goes on the gutter its axis is on: an x axis names itself
	// under the labels, a y axis at the head of the numbers it names.
	if opts.X.Scale.OK() && opts.X.Label != "" {
		name := LabelHeight(c, theme.RowSize) + u
		if opts.X.Side == Top {
			in.Top += name
		} else {
			in.Bottom += name
		}
	}
	// The legend sits above the plot, a step clear of it; a y axis' name sits
	// above that, at the head of the numbers it names.
	if len(opts.Legend.Entries) > 0 {
		in.Top += LegendHeight(c, opts.Legend) + u
	}
	if opts.Y.Scale.OK() && opts.Y.Label != "" {
		in.Top += LabelHeight(c, theme.RowSize) + u
	}
	// A y axis' topmost label is centred on its tick, and the top tick sits
	// on the plot's own top edge: half the label reaches above the plot,
	// where a legend or an axis name lives. Keep that half clear, or the
	// frame's first number reads straight through them.
	if opts.Y.Scale.OK() {
		in.Top += LabelHeight(c, tickSize(opts.Y)) / 2
	}
	return in
}

// axisGap is the room between an axis' labels and the plot: its tick marks
// and a step of space. A tick mark is on the plot's side of the labels, so a
// caller reading a value sees the tick it belongs to.
func axisGap(c *ui.Context, opts AxisOptions) float32 {
	u := core.Density(c).Unit()
	if opts.Ticks {
		return u*1.5 + u
	}
	return u
}

func tickSize(opts AxisOptions) float32 {
	if opts.Size > 0 {
		return opts.Size
	}
	return theme.CaptionSize
}
