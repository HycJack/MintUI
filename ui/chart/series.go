package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Form is what a series draws itself as. It is on the series rather than on
// the chart because the same series draws as a line on one axis and as bars
// on another, and a caller should not have to say so twice.
type Form uint8

const (
	// FormLine joins its points with a line.
	FormLine Form = iota
	// FormPoint draws a mark at each point and nothing between them, which is
	// what a scatter is.
	FormPoint
	// FormBar draws a bar per point, standing on the baseline.
	FormBar
	// FormArea fills the space under its line.
	FormArea
)

func (f Form) String() string {
	switch f {
	case FormPoint:
		return "Point"
	case FormBar:
		return "Bar"
	case FormArea:
		return "Area"
	}
	return "Line"
}

// Point is one datum: where it falls on each axis. On a band axis X is the
// index of a category rather than a value.
type Point struct {
	X, Y float64
}

// Series is one thing a chart plots: a name for the legend, a colour, and the
// points it draws.
//
// The colour is optional. A series that leaves it takes the next colour off
// the window's palette, which is what keeps a chart from being all the same
// blue — see [Palette] and [EntriesOf].
type Series struct {
	// Name is what the legend calls it.
	Name string
	// Color is what it is drawn in; the zero colour takes the palette's.
	Color ui.Color
	// Form is what it draws itself as.
	Form Form
	// Points are the data, in the order they are drawn.
	Points []Point
}

// NewSeries is a series of a form over its points:
//
//	chart.NewSeries("Open", chart.FormLine,
//		chart.Point{X: 0, Y: 12}, chart.Point{X: 1, Y: 15})
func NewSeries(name string, form Form, points ...Point) Series {
	return Series{Name: name, Form: form, Points: points}
}

// SeriesFrom is a series from two parallel slices, which is the shape data
// usually arrives in. The two have to be the same length: a series whose x's
// and y's are not lined up would draw points that do not exist.
func SeriesFrom(name string, form Form, xs, ys []float64) Series {
	if len(xs) != len(ys) {
		panic("chart: SeriesFrom needs one x for every y")
	}
	points := make([]Point, len(xs))
	for i := range points {
		points[i] = Point{X: xs[i], Y: ys[i]}
	}
	return Series{Name: name, Form: form, Points: points}
}

// Domain is the span of a series' y values, and whether it holds any. A
// series with nothing in it is an empty thing rather than a range, which is
// what stops a chart building an axis over 0 to 0.
func (s Series) Domain() (Domain, bool) {
	ys := make([]float64, len(s.Points))
	for i, p := range s.Points {
		ys[i] = p.Y
	}
	return DomainOf(ys)
}

// SeriesDomain is the span of a set of series' y values — the number a
// chart's y axis is built from — and whether there was anything to draw at
// all.
func SeriesDomain(series []Series) (Domain, bool) {
	var d Domain
	var seen bool
	for _, s := range series {
		if one, ok := s.Domain(); ok {
			if !seen {
				d, seen = one, true
			} else {
				d = d.Union(one)
			}
		}
	}
	return d, seen
}

// PlotOptions configure a [Plot].
type PlotOptions struct {
	// Frame is the chart's frame, whose plot area the series is drawn in.
	Frame FrameResult
	// Series is what to draw.
	Series Series
	// X and Y are the axes' scales, written in their own space: they are
	// placed in the plot as this draws, which is the only time its size is
	// known.
	X, Y Scale
	// Width is the line's weight; zero uses two DIPs.
	Width float32
	// Fill is how solid the bars and the area are; zero uses 0.85 for bars
	// and 0.18 for an area, which is what reads as solid and as wash at the
	// same time.
	Fill float32
	// Dots marks every point of a line or an area as well.
	Dots bool
	// DotRadius is how big those marks are; zero uses 2.5 DIPs.
	DotRadius float32
	// Baseline is what bars stand on and what an area fills down to; nil
	// means zero where the axis crosses it, and the edge of the plot where it
	// does not.
	Baseline *float64
	// BarRatio is how much of its band a bar takes; zero uses 0.7.
	BarRatio float32
}

// Plot draws one series into its frame's plot area.
//
// Everything is clipped to that area: a line whose data runs past the axis,
// an area over a negative baseline and a bar at the last category all stop at
// the plot's edge instead of drawing over the axis' labels — which is what a
// chart that does not clip looks like, and why it looks wrong without anybody
// being able to say why.
func Plot(c *ui.Context, opts PlotOptions) *ui.Element {
	u := core.Density(c).Unit()
	s := opts.Series
	color := s.Color
	if color == (ui.Color{}) {
		color = Palette(c, 1)[0]
	}
	width := opts.Width
	if width <= 0 {
		width = theme.BorderWidth * 2
	}
	fill := opts.Fill
	if fill <= 0 {
		if s.Form == FormArea {
			fill = 0.18
		} else {
			fill = 0.85
		}
	}
	dot := opts.DotRadius
	if dot <= 0 {
		dot = u * 0.625
	}
	ratio := opts.BarRatio
	if ratio <= 0 {
		ratio = 0.7
	}

	e := ui.Box(c).Absolute().Fill()
	if s.Name != "" {
		e = e.Label(s.Name + " series")
	}
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := opts.Frame.Plot()
		if plot.W <= 0 || plot.H <= 0 || len(s.Points) == 0 {
			return
		}
		xs := opts.X.Span(plot.X, plot.X+plot.W)
		ys := opts.Y.Span(plot.Y+plot.H, plot.Y)
		base := baselineOf(opts, ys, plot)

		p.Clip(plot, 0, func() {
			switch s.Form {
			case FormBar:
				bars(p, s, xs, ys, base, color.Alpha(fill), dot/2, ratio)
			case FormArea:
				area(p, s, xs, ys, base, color.Alpha(fill))
				line(p, s, xs, ys, color, width)
			case FormPoint:
				for _, pt := range s.Points {
					if x, y, ok := screen(xs, ys, pt); ok {
						internal.Dot(p, x, y, dot, color)
					}
				}
			default:
				line(p, s, xs, ys, color, width)
				if opts.Dots {
					for _, pt := range s.Points {
						if x, y, ok := screen(xs, ys, pt); ok {
							internal.Dot(p, x, y, dot, color)
						}
					}
				}
			}
		})
	})
}

// screen is where one datum falls in the plot. A point that is not a number —
// a gap in the data — has no place on a chart and is dropped rather than
// drawn at the edge.
func screen(xs, ys Scale, pt Point) (x, y float32, ok bool) {
	if math.IsNaN(pt.X) || math.IsNaN(pt.Y) || math.IsInf(pt.X, 0) || math.IsInf(pt.Y, 0) {
		return 0, 0, false
	}
	return xs.At(pt.X), ys.At(pt.Y), true
}

// line joins a series' points, starting a new path where the data has a gap
// in it rather than drawing a straight line across the gap.
func line(p *ui.Painter, s Series, xs, ys Scale, color ui.Color, width float32) {
	var path ui.Path
	drawing := false
	for _, pt := range s.Points {
		x, y, ok := screen(xs, ys, pt)
		if !ok {
			drawing = false
			continue
		}
		if drawing {
			path.LineTo(x, y)
		} else {
			path.MoveTo(x, y)
			drawing = true
		}
	}
	p.StrokePath(&path, width, color)
}

// area fills the space between a series' line and the baseline: down the
// baseline to the first point, along the line, and back down again.
func area(p *ui.Painter, s Series, xs, ys Scale, base float32, fill ui.Color) {
	var path ui.Path
	started := false
	var lastX float32
	for _, pt := range s.Points {
		x, y, ok := screen(xs, ys, pt)
		if !ok {
			continue
		}
		if !started {
			path.MoveTo(x, base)
			started = true
		}
		path.LineTo(x, y)
		lastX = x
	}
	if !started {
		return
	}
	path.LineTo(lastX, base)
	path.Close()
	p.FillPath(&path, fill)
}

// bars draws one bar per point, standing on the baseline.
func bars(p *ui.Painter, s Series, xs, ys Scale, base float32, fill ui.Color, radius, ratio float32) {
	width := measureBar(xs, s.Points, ratio)
	for _, pt := range s.Points {
		x, y, ok := screen(xs, ys, pt)
		if !ok {
			continue
		}
		top, bottom := minf(y, base), maxf(y, base)
		if bottom-top < 1 {
			bottom = top + 1
		}
		p.Fill(ui.Rect{X: x - width/2, Y: top, W: width, H: bottom - top}, fill, radius)
	}
}

// measureBar is how wide one bar is: on a band axis, its share of the axis
// less what the bars leave as gaps; on a continuous axis, the closest two
// points get, so that bars never overlap where the data is crowded.
func measureBar(xs Scale, points []Point, ratio float32) float32 {
	if xs.Kind() == KindBand {
		// A bar takes its share of the axis less what the bars leave between
		// them, which is the same share whatever else is on the chart.
		start, _, end := xs.Share(0)
		return maxf(maxf(end-start, 0)*ratio, 1)
	}
	closest := float32(0)
	for i := 1; i < len(points); i++ {
		gap := maxf(xs.At(points[i].X)-xs.At(points[i-1].X), xs.At(points[i-1].X)-xs.At(points[i].X))
		if gap <= 0 || math.IsNaN(float64(gap)) {
			continue
		}
		if closest == 0 || gap < closest {
			closest = gap
		}
	}
	if closest == 0 {
		closest = maxf(xs.To()-xs.From(), 0) / float32(max(len(points), 1))
	}
	return maxf(closest*ratio, 1)
}

// baselineOf is where bars stand and areas fill down to: zero where the axis
// crosses it, and the edge of the plot where a range that never crosses zero
// would otherwise draw its bars down past the frame.
func baselineOf(opts PlotOptions, ys Scale, plot ui.Rect) float32 {
	if opts.Baseline != nil {
		return ys.At(*opts.Baseline)
	}
	d := ys.Domain()
	switch {
	case d.Min > 0:
		return plot.Y + plot.H
	case d.Max < 0:
		return plot.Y
	}
	return ys.At(0)
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
