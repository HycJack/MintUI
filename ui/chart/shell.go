package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// This file is the scaffold every chart type stands on. A chart type is not
// free to measure its own plot area: the whole reason this package has a
// [Frame], a [Measure] and a [Ticks] is that a chart's gutters are measured
// rather than guessed, and a chart type that guessed them would be the one
// place where the guessing came back. So a chart type asks for a canvas —
// [canvas] with axes, [panel] without them, [mark] with no chrome at all —
// and draws into what it is handed.

// ChartOptions are the parts of a chart every type shares: its box, its two
// axes, its legend and its pointer. Each type embeds it rather than spelling
// the fields out again, so a chart type is about its own shape and the frame
// underneath stays the same one.
type ChartOptions struct {
	// Height gives the chart its own height; zero lets the layout give it one,
	// which is what a chart in a column wants.
	Height float32
	// Pad is the inner padding between the frame's edge and everything in it.
	Pad float32
	// Label names the chart for assistive technology. A chart's numbers are
	// drawn rather than laid out, so this is the only thing a screen reader
	// has to go on: give it one.
	Label string
	// X and Y are the two axes. An axis left without a scale is not drawn at
	// all — which is how a chart says it has no numbers on that side — and a
	// chart type fills one in from its own data when the caller leaves it out.
	X, Y AxisOptions
	// Legend sits in the top gutter. [EntriesOf] builds it from the series so
	// that the names and the colours are the ones the chart draws with.
	Legend LegendOptions
	// Hover is the caller's pointer, filled by the frame as it paints and read
	// by the pieces that follow the pointer. Nothing here keeps it.
	Hover *Hover
	// Surface draws the chart a step below the window, for a chart that is a
	// panel rather than part of the page.
	Surface bool
	// Border draws a hairline around it.
	Border bool
	// Radius rounds the frame's own surface; zero uses [theme.SmallRadius].
	Radius float32
	// Grid draws the grid lines behind the data. It is off unless asked for,
	// because a filled chart — a pie, a treemap, a heatmap — has no room for
	// them and would only be crossed by them.
	Grid bool
	// EmptyTitle and EmptyBody are what the chart says when there is nothing
	// to draw. They are the caller's words, because "no callbacks resolved
	// today" is a sentence about their app and not about a chart; a chart
	// that says nothing in particular says it in English.
	EmptyTitle, EmptyBody string
}

// canvas draws a chart with two axes: the frame, the grid, the axes themselves
// and whatever the type draws inside the plot area.
//
// The type draws through [marks], which hands it the plot area and both scales
// as the frame paints — the only moment either is known. A chart type that
// placed a scale itself would be doing the frame's job a second time, and a
// second time means a second answer.
func canvas(c *ui.Context, opts ChartOptions, body func(f FrameResult)) *ui.Element {
	// A value axis left on Side's zero reads along the bottom, over the x
	// axis' own labels: a y scale that did not say where it goes reads on
	// the left. One placed deliberately — a right-hand scale, a top one —
	// stays where its caller put it.
	if opts.Y.Scale.OK() && opts.Y.Side == Bottom {
		opts.Y.Side = Left
	}
	return Frame(c, frameOptions(opts), func(f FrameResult) {
		if opts.Grid {
			Grid(c, GridOptions{Frame: f, X: opts.X, Y: opts.Y})
		}
		if body != nil {
			body(f)
		}
		Axis(c, f, opts.X)
		Axis(c, f, opts.Y)
	}).Element
}

// panel draws a chart with no axes at all — a pie, a treemap, a chord — where
// the plot area is the whole of the frame and the labels are drawn from the
// marks themselves. It still goes through the frame, so the padding, the
// surface and the border are the same measured ones every other chart has,
// and a chart type with no axes is still named for assistive technology.
func panel(c *ui.Context, opts ChartOptions, body func(f FrameResult)) *ui.Element {
	return canvas(c, opts, body)
}

// marksBox is the element a chart with no axes draws through: an invisible
// layer over the frame's own box that paints as the frame does, so that the
// marks of a pie, a treemap or a ring land inside the plot area the frame
// measured rather than wherever the window happens to be.
func marksBox(c *ui.Context, f FrameResult, name string, body func(p *ui.Painter, plot ui.Rect)) *ui.Element {
	e := ui.Box(c).Absolute().Fill()
	if name != "" {
		e = e.Label(name)
	}
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := f.Plot()
		if plot.W <= 0 || plot.H <= 0 {
			return
		}
		p.Clip(plot, 0, func() { body(p, plot) })
	})
}

// marksWide is marksBox with the clip on the frame rather than on the plot
// area, for a chart that writes its own axis names and outer ring labels
// outside the plot they measure. Clipped to the plot they are cut in half —
// a label that is written across an edge is not a label that is in the
// gutter, it is a label that has to have room to cross.
func marksWide(c *ui.Context, f FrameResult, name string, body func(p *ui.Painter, plot ui.Rect)) *ui.Element {
	e := ui.Box(c).Absolute().Fill()
	if name != "" {
		e = e.Label(name)
	}
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := f.Plot()
		if plot.W <= 0 || plot.H <= 0 {
			return
		}
		p.Clip(f.Bounds(), 0, func() { body(p, plot) })
	})
}

// mark draws into a box rather than into a frame: the charts with no axes and
// no gutters to measure — a sparkline in a table row, a gauge on a card, a
// countdown between two pieces of text. Their box is their plot area, so
// there is nothing here to measure, and going through a frame would only add
// a gutter nobody asked for.
func mark(c *ui.Context, name string, w, h float32, body func(p *ui.Painter, r ui.Rect)) *ui.Element {
	e := ui.Box(c)
	switch {
	case w > 0 && h > 0:
		e.Size(w, h)
	case w > 0:
		e.Width(w).FillHeight()
	case h > 0:
		e.Height(h).FillWidth()
	default:
		e.Fill()
	}
	if name != "" {
		e.Label(name)
	}
	return e.Draw(body)
}

// frameOptions is the chart's own options as a frame's: the shell's fields
// with the two the frame does not have left out.
func frameOptions(opts ChartOptions) FrameOptions {
	return FrameOptions{
		Height:  opts.Height,
		Pad:     opts.Pad,
		Label:   opts.Label,
		X:       opts.X,
		Y:       opts.Y,
		Legend:  opts.Legend,
		Hover:   opts.Hover,
		Surface: opts.Surface,
		Border:  opts.Border,
		Radius:  opts.Radius,
	}
}

// banded is the x axis of a chart over labels: one equal share per category,
// with each category's tick in the middle of its share rather than on its
// edge, so that a bar has room on both sides of itself.
func banded(labels []string) Scale { return NewBand(labels, 0, 1) }

// counted is a y axis over a data's span, rounded out to numbers somebody
// would have chosen. It is what a chart type falls back on when the caller
// brings no scale of their own: the numbers are the chart type's to know, and
// asking the caller to restate them is how the two drift apart.
func counted(d Domain, count int) Scale {
	return NewLinear(Nice(d.Min, d.Max, max(count, 1)), 0, 1)
}

// axisX fills in the scale a categorical axis was left without.
func axisX(opts AxisOptions, labels []string) AxisOptions {
	if opts.Scale.OK() {
		return opts
	}
	opts.Scale = banded(labels)
	return opts
}

// axisY fills in the scale a value axis was left without, from a span that
// already includes zero where the data calls for it — a bar or a histogram
// whose baseline is its own smallest value says nothing about it.
func axisY(opts AxisOptions, d Domain, count int) AxisOptions {
	if opts.Scale.OK() {
		return opts
	}
	opts.Scale = NewLinear(Nice(FromZero(d).Min, d.Max, max(count, 1)), 0, 1)
	return opts
}

// axisValues fills in the scale a value axis was left without, from a span
// that may sit anywhere: a line's axis is not obliged to reach zero.
func axisValues(opts AxisOptions, d Domain, count int) AxisOptions {
	if opts.Scale.OK() {
		return opts
	}
	opts.Scale = counted(d, count)
	return opts
}

// coloured gives a set of series their colours, and the legend to go with
// them. It is the one place a chart type resolves which colour is which, so
// that the marks on the plot and the names above them cannot disagree: a
// series that chose no colour takes the palette's at its place in the set,
// rather than the first colour of the palette whichever set it is in.
func coloured(c *ui.Context, series []Series) ([]Series, []LegendEntry) {
	entries := EntriesOf(c, series)
	out := make([]Series, len(series))
	for i, s := range series {
		out[i] = s
		out[i].Color = entries[i].Color
	}
	return out, entries
}

// named fills in the legend of a chart over labels and values: what a pie,
// a treemap or a funnel needs, where there are no series to name.
func named(c *ui.Context, labels []string, count int) LegendOptions {
	if len(labels) == 0 {
		return LegendOptions{}
	}
	return LegendOptions{Entries: paletteEntries(c, min(count, len(labels)), labels)}
}

// paletteEntries is a legend for labels rather than for series: as many
// colours as the chart needs, in the order they are drawn.
func paletteEntries(c *ui.Context, n int, labels []string) []LegendEntry {
	colors := Palette(c, n)
	out := make([]LegendEntry, min(n, len(labels)))
	for i := range out {
		out[i] = LegendEntry{Name: labels[i], Color: colors[i], Mark: MarkSquare}
	}
	return out
}

// colorAt is the nth of a set of palette colours, and the first of them for
// any set with nothing in it.
func colorAt(colors []ui.Color, i int) ui.Color {
	if len(colors) == 0 {
		return ui.Color{}
	}
	return colors[i%len(colors)]
}

// dots draws a mark at each of a set of points, in one colour: what a
// scatter's, a box plot's and a violin's shared.
func dots(p *ui.Painter, xs, ys Scale, points []Point, radius float32, color ui.Color) {
	for _, pt := range points {
		if x, y, ok := screen(xs, ys, pt); ok {
			internal.Dot(p, x, y, radius, color)
		}
	}
}

// baselineAt is where bars stand: zero where the axis crosses it, and the
// edge of the plot where a range that never crosses zero would otherwise draw
// its bars down past the frame.
func baselineAt(ys Scale, plot ui.Rect) float32 {
	switch d := ys.Domain(); {
	case d.Min > 0:
		return plot.Y + plot.H
	case d.Max < 0:
		return plot.Y
	}
	return ys.At(0)
}

// ringAround is a mark's own outline, in a colour of the caller's: what
// separates a bubble from its neighbour and a node from its edges.
func ringAround(p *ui.Painter, x, y, r float32, color ui.Color) {
	internal.Ring(p, x, y, r, theme.BorderWidth, color)
}

// wedge is the path of a slice of a ring: the two radii and the curve between
// them, from angle a0 to a1 in degrees. A ring from zero is the same path with
// no inner radius, which is why the two share one function.
func wedge(cx, cy, rIn, rOut, a0, a1 float32, steps int) ui.Path {
	path := ring(cx, cy, rOut, a0, a1, steps)
	if rIn > 0 {
		path.LineTo(point(cx, cy, rIn, a1))
		for i := steps; i >= 0; i-- {
			path.LineTo(point(cx, cy, rIn, a0+(a1-a0)*float32(i)/float32(steps)))
		}
	}
	path.Close()
	return path
}

// ring is the curve of a circle from a0 to a1 degrees, with no radii: what a
// gauge's needle is drawn along and a chord's ribbon is bent around.
//
// The steps are fixed rather than worked out from the span, because the ring
// is flattened into line segments by the renderer and a curve half a degree
// wide is a curve nobody can see the difference on.
func ring(cx, cy, r, a0, a1 float32, steps int) ui.Path {
	var path ui.Path
	steps = clampSteps(steps)
	for i := range steps + 1 {
		x, y := point(cx, cy, r, a0+(a1-a0)*float32(i)/float32(steps))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	return path
}

// clampSteps is how finely a curve is drawn: enough for a small arc to look
// like a curve, and never so many that a full circle costs a thousand points.
func clampSteps(steps int) int {
	return min(max(steps, 8), 180)
}

// point is one place on a curve.
func point(cx, cy, r, deg float32) (x, y float32) {
	return cx + r*float32(math.Cos(radians(deg))), cy + r*float32(math.Sin(radians(deg)))
}

// radians turns degrees into radians. Angles are written in degrees everywhere
// a caller can see them — a sweep, a gauge's span — because degrees are what a
// pie is measured in, and radians are what only a trigonometry table is.
func radians(deg float32) float64 { return float64(deg) * math.Pi / 180 }

// ink is the colour a mark is drawn in, or the accent when the caller left it
// out: the one colour in the interface that means "look here".
func ink(color ui.Color, k theme.Tokens) ui.Color {
	if color == (ui.Color{}) {
		return k.Accent
	}
	return color
}

// unit is one step of the window's spacing scale, which is what every gap,
// mark and thickness in this package is built from.
func unit(c *ui.Context) float32 { return core.Density(c).Unit() }
