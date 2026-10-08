package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The charts in this file are the cartesian ones: two axes, a plot area
// between them, and a scale for each. None of them measures anything. A type
// asks [canvas] for a frame, takes the plot area and the two scales out of it
// as it draws, and puts its marks down inside — which is why all of them get
// gutters sized to their own numbers, axis labels that do not collide, and an
// empty state that says what is missing rather than an axis over a range
// nothing reaches.

// marks is the element a cartesian chart type draws through: an invisible
// layer over the frame's own box that paints as the frame does.
//
// It paints after the frame has worked out where the plot is, and it clips to
// that plot, so a mark whose data runs past the axis stops at the edge instead
// of drawing over the axis' labels — which is what a chart that does not clip
// looks like, and why it looks wrong without anybody being able to say why.
func marks(c *ui.Context, f FrameResult, name string, body func(p *ui.Painter, plot ui.Rect, xs, ys Scale)) *ui.Element {
	e := ui.Box(c).Absolute().Fill()
	if name != "" {
		e = e.Label(name)
	}
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := f.Plot()
		if plot.W <= 0 || plot.H <= 0 {
			return
		}
		// The scales are the frame's own, already placed in this plot:
		// a chart type cannot place them itself, because it does not know
		// how big the plot is until the frame says so.
		p.Clip(plot, 0, func() { body(p, plot, f.X(), f.Y()) })
	})
}

// nothing is what a chart with no data draws: the frame as it was asked for,
// and then [Empty] over the plot area.
//
// The caller writes the words, because the words are theirs — "no callbacks
// resolved today" is a sentence about their app, not about this package.
func nothing(c *ui.Context, opts ChartOptions) *ui.Element {
	title := opts.EmptyTitle
	if title == "" {
		title = "Nothing to plot"
	}
	body := opts.EmptyBody
	if body == "" {
		body = "There is nothing in range to draw."
	}
	// A grid behind an empty state is a grid over nothing, which is the one
	// thing an empty state exists to avoid.
	opts.Grid = false
	return canvas(c, opts, func(f FrameResult) {
		Empty(c, EmptyOptions{Frame: f, Title: title, Body: body})
	})
}

// xDomain is the span of a set of series' x values: the range a chart with no
// x axis of its own builds one from.
func xDomain(series []Series) (Domain, bool) {
	values := make([]float64, 0, len(series))
	for _, s := range series {
		for _, pt := range s.Points {
			values = append(values, pt.X)
		}
	}
	return DomainOf(values)
}

// axisXValues fills in the scale a continuous x axis was left without.
func axisXValues(opts AxisOptions, d Domain) AxisOptions {
	if opts.Scale.OK() {
		return opts
	}
	opts.Scale = counted(d, DefaultTickCount)
	return opts
}

// valueRows is a set of series' values as rows: the shape the stacking maths
// works on.
func valueRows(series []Series) [][]float64 {
	rows := make([][]float64, len(series))
	for i, s := range series {
		rows[i] = make([]float64, len(s.Points))
		for j, pt := range s.Points {
			rows[i][j] = pt.Y
		}
	}
	return rows
}

// categories is a set of series' categories: one per column, named after the x
// value in it. On a band axis a point's X is which category it belongs to, so
// a chart of days written as 0…6 comes out with its days in order.
func categories(series []Series) []string {
	longest := 0
	for _, s := range series {
		longest = max(longest, len(s.Points))
	}
	out := make([]string, longest)
	for _, s := range series {
		for i, pt := range s.Points {
			if i < longest {
				out[i] = Group(pt.X, 0)
			}
		}
	}
	return out
}

// ── line ───────────────────────────────────────────────────────────────────

// LineOptions configure a [LineChart].
type LineOptions struct {
	ChartOptions
	// Series are the lines. A series that chose no colour takes the palette's
	// at its place in the set, so two lines next to each other are two
	// colours rather than the same one twice.
	Series []Series
	// Width is the line's weight; zero uses two DIPs.
	Width float32
	// Dots marks every point of a line as well as joining them, which is what
	// a series of a dozen readings wants and a series of a thousand does not.
	Dots bool
	// DotRadius is how big those marks are; zero uses two and a half DIPs.
	DotRadius float32
}

// LineChart is values over an axis: the one chart every other cartesian chart
// is a variation of.
//
//	var hover chart.Hover
//	series := []chart.Series{
//		chart.SeriesFrom("Open", chart.FormLine, days, open),
//		chart.SeriesFrom("Closed", chart.FormLine, days, closed),
//	}
//	chart.LineChart(c, chart.LineOptions{
//		ChartOptions: chart.ChartOptions{
//			Height: 220, Grid: true, Hover: &hover,
//			X: chart.AxisOptions{Count: 7, Label: "Day"},
//			Y: chart.AxisOptions{Side: chart.Left, Count: 5, Label: "Callbacks"},
//			Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
//		},
//		Series: series, Dots: true,
//	})
//
// The axes may be left out: the x one is built from the points' x values and
// the y one from their span, both rounded out to round numbers. Bring your own
// scales when several charts have to sit on one axis.
func LineChart(c *ui.Context, opts LineOptions) *ui.Element {
	chart, series, ok := cartesian(c, opts.Series, opts.ChartOptions)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	return canvas(c, chart, func(f FrameResult) {
		marks(c, f, "Line", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			for _, s := range series {
				line(p, s, xs, ys, s.Color, lineWidth(opts.Width))
				if opts.Dots {
					dots(p, xs, ys, s.Points, dotRadius(c, opts.DotRadius), s.Color)
				}
			}
		})
	})
}

// cartesian is the front half every cartesian type shares: the series with
// their colours on, both axes filled in from the data unless the caller
// brought scales of their own, the legend filled in from the same series,
// and false when there is nothing to draw at all.
//
// It is a function rather than three lines repeated in each type because a
// chart type that builds its axis differently from its neighbour is how two
// charts of the same data end up with different gutters. It hands the options
// back rather than filling them in behind the caller's back, because the axes
// it worked out are the axes the frame is given — a type that threw them away
// would get a frame with no axis on it and marks drawn against nothing.
func cartesian(c *ui.Context, in []Series, opts ChartOptions) (ChartOptions, []Series, bool) {
	series, entries := coloured(c, in)
	d, ok := SeriesDomain(series)
	if !ok {
		return opts, nil, false
	}
	xd, hasX := xDomain(series)
	if !hasX {
		return opts, nil, false
	}
	opts.X = axisXValues(opts.X, xd)
	opts.Y = axisValues(opts.Y, d, opts.Y.Count)
	return charted(opts, entries), series, true
}

// charted fills a chart's legend in from its series, when the caller did not
// write one: a chart whose legend and marks disagree about which colour is
// which is worse than a chart with no legend at all.
func charted(opts ChartOptions, entries []LegendEntry) ChartOptions {
	if opts.Legend.Entries == nil {
		opts.Legend.Entries = entries
	}
	return opts
}

// ── area ───────────────────────────────────────────────────────────────────

// AreaOptions configure an [AreaChart].
type AreaOptions struct {
	ChartOptions
	// Series are the areas, each filling from the ground up to its own line.
	Series []Series
	// Stacked puts each series on top of the ones below it, so the plot reads
	// as one total split up rather than as several fills fighting over the
	// same ground.
	Stacked bool
	// Fill is how solid the areas are; zero uses 0.18, a wash that still shows
	// the grid through it.
	Fill float32
	// Width is the line along the top of an area; zero uses two DIPs.
	Width float32
}

// AreaChart is a line with the ground under it filled in: the same chart as
// [LineChart], read for the size of what is under the curve rather than for
// where the curve went.
func AreaChart(c *ui.Context, opts AreaOptions) *ui.Element {
	chart, series, ok := cartesian(c, opts.Series, opts.ChartOptions)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	return canvas(c, chart, func(f FrameResult) {
		marks(c, f, "Area", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			if opts.Stacked {
				stackAreas(p, series, xs, ys, areaFill(opts.Fill))
				return
			}
			base := baselineAt(ys, plot)
			for _, s := range series {
				area(p, s, xs, ys, base, s.Color.Alpha(areaFill(opts.Fill)))
				line(p, s, xs, ys, s.Color, lineWidth(opts.Width))
			}
		})
	})
}

// stackAreas draws areas on top of one another: each filled from the running
// total of the series below it, so the top of the stack is the total of them
// all and the line along each edge is the edge of that band's own sum.
func stackAreas(p *ui.Painter, series []Series, xs, ys Scale, fill float32) {
	_, upper := Bounds(valueRows(series))
	for i, s := range series {
		if i >= len(upper) {
			continue
		}
		// The edge of this band is the running total, which is what a
		// stacked area's line is: not its own values, but the total of
		// everything below it as well.
		edge := s
		edge.Points = make([]Point, len(s.Points))
		for j, pt := range s.Points {
			y := 0.0
			if j < len(upper[i]) {
				y = upper[i][j]
			}
			edge.Points[j] = Point{X: pt.X, Y: y}
		}
		area(p, edge, xs, ys, ys.At(0), s.Color.Alpha(fill))
		line(p, edge, xs, ys, s.Color, theme.BorderWidth*2)
	}
}

// AreaMountainOptions configure an [AreaMountain].
type AreaMountainOptions struct {
	ChartOptions
	// Series are the ridges. Each is filled to the ground, so where two of
	// them overlap the one behind shows through the one in front.
	Series []Series
	// Fill is how solid each ridge is; zero uses 0.3, faint enough to layer
	// four deep without the front one hiding the rest.
	Fill float32
}

// AreaMountain is a set of areas over one another, biggest at the back: the
// shape a set of overlapping distributions has when nobody has cut one of them
// out of the others.
func AreaMountain(c *ui.Context, opts AreaMountainOptions) *ui.Element {
	chart, series, ok := cartesian(c, opts.Series, opts.ChartOptions)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	// Biggest first, because the mountain behind is the one that has to
	// show, and the drawing order is the only thing that says which is which.
	series = tallest(series)
	return canvas(c, chart, func(f FrameResult) {
		marks(c, f, "Areas", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			base := baselineAt(ys, plot)
			for _, s := range series {
				area(p, s, xs, ys, base, s.Color.Alpha(mountainFill(opts.Fill)))
				line(p, s, xs, ys, s.Color, theme.BorderWidth)
			}
		})
	})
}

// tallest is a set of series ordered by how far they rise: the order layered
// areas are drawn in, the biggest at the back.
func tallest(series []Series) []Series {
	out := append([]Series(nil), series...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && riseOf(out[j]) > riseOf(out[j-1]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func riseOf(s Series) float64 {
	d, ok := s.Domain()
	if !ok {
		return 0
	}
	return d.Span()
}

// ── bars ───────────────────────────────────────────────────────────────────

// BarOptions configure a [BarChart].
type BarOptions struct {
	ChartOptions
	// Series are the bars. One series with a point per category is a chart of
	// one measure; several are drawn side by side within each category.
	Series []Series
	// Ratio is how much of its band a bar takes; zero uses 0.7.
	Ratio float32
	// Horizontal turns the chart on its side: the categories down the left and
	// the values across the bottom, for a set of names too long to write
	// underneath a column.
	Horizontal bool
}

// BarChart is one bar per category, standing on a baseline.
//
// The baseline is zero, not the smallest value: a bar's length is read as a
// length, and a length that starts at 40 says nothing about whether 40 is a
// lot. A chart of all negative values stands on zero too, at the top.
func BarChart(c *ui.Context, opts BarOptions) *ui.Element {
	series, entries := coloured(c, opts.Series)
	d, ok := SeriesDomain(series)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	if opts.Horizontal {
		// On its side the two axes swap jobs rather than the chart being
		// drawn sideways: the categories are the band axis and go up the
		// left, the values are the linear one and go along the bottom.
		opts.Y = axisX(opts.Y, categories(series))
		opts.X = axisY(opts.X, FromZero(d), opts.X.Count)
	} else {
		opts.X = axisX(opts.X, categories(series))
		opts.Y = axisY(opts.Y, FromZero(d), opts.Y.Count)
	}
	opts.ChartOptions = charted(opts.ChartOptions, entries)
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Bars", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			if opts.Horizontal {
				sideBySide(p, series, xs, ys, plot, barRatio(opts.Ratio), true)
				return
			}
			sideBySide(p, series, xs, ys, plot, barRatio(opts.Ratio), false)
		})
	})
}

// sideBySide draws a set of series as bars sharing each category's band
// between them, so that several series never draw over one another, and
// standing each one on the baseline.
//
// It is one function for both orientations because a horizontal bar chart is
// the same drawing with the two axes swapped — which is exactly what the
// caller did by putting the band scale on the other axis.
func sideBySide(p *ui.Painter, series []Series, xs, ys Scale, plot ui.Rect, ratio float32, horizontal bool) {
	groups := float32(max(len(series), 1))
	// A chart of one measure takes the whole band less its gap; several take
	// a slice each, so that neither covers the other.
	share := ratio / groups
	for i, s := range series {
		for _, pt := range s.Points {
			along := xs
			if horizontal {
				along = ys
			}
			start, _, end := along.Share(int(pt.X))
			// The band runs upwards on a chart turned on its side, so the
			// thickness is the distance between the two edges rather than
			// the second less the first.
			thickness := absf32(end-start) * share
			if thickness <= 0 {
				continue
			}
			// The slice this series takes sits at its own place in the
			// band, measured from where the whole stack would have started.
			at := start + (end-start)*ratio/float32(len(series))*float32(i)
			if !horizontal {
				rect := inside(ui.Rect{X: at, W: thickness, H: 1}, plot)
				y := ys.At(pt.Y)
				rect.Y, rect.H = minf(y, ys.At(0)), maxf(absf32(y-ys.At(0)), 1)
				rect.Y, rect.H = maxf(rect.Y, plot.Y), minf(rect.H, plot.Y+plot.H-rect.Y)
				p.Fill(rect, s.Color.Alpha(0.85), minf(thickness/3, theme.SmallRadius))
				continue
			}
			rect := inside(ui.Rect{Y: at, H: thickness, W: 1}, plot)
			x := xs.At(pt.Y)
			rect.X, rect.W = minf(x, xs.At(0)), maxf(absf32(x-xs.At(0)), 1)
			rect.X, rect.W = maxf(rect.X, plot.X), minf(rect.W, plot.X+plot.W-rect.X)
			p.Fill(rect, s.Color.Alpha(0.85), minf(thickness/3, theme.SmallRadius))
		}
	}
}

// StackedBarOptions configure a [StackedBar] and a [PercentBar].
type StackedBarOptions struct {
	ChartOptions
	// Labels are the categories, one per column.
	Labels []string
	// Series are the stack, in the order it stacks: the first is on the
	// baseline and the last is at the top.
	Series []Series
	// Ratio is how much of its band a column takes; zero uses 0.7.
	Ratio float32
}

// StackedBar is a column per category split into a stack, so that the top of
// each column is the total of it and each band is one part of that total.
func StackedBar(c *ui.Context, opts StackedBarOptions) *ui.Element {
	return stacked(c, opts, opts.Series)
}

// PercentBar is a [StackedBar] in shares rather than in totals: every column
// comes to the same height and its bands are percentages of it, which is what
// makes two columns of very different sizes comparable at a glance.
func PercentBar(c *ui.Context, opts StackedBarOptions) *ui.Element {
	return stacked(c, opts, sharesOf(opts.Series))
}

// sharesOf is a set of series with every column turned into its share of that
// column, which is all a percent bar is: the same stack, in percentages.
func sharesOf(series []Series) []Series {
	shares := Stack(valueRows(series))
	out := make([]Series, len(series))
	for i, s := range series {
		out[i] = s
		out[i].Points = make([]Point, len(s.Points))
		for j, pt := range s.Points {
			y := 0.0
			if j < len(shares[i]) {
				y = shares[i][j]
			}
			out[i].Points[j] = Point{X: pt.X, Y: y}
		}
	}
	return out
}

func stacked(c *ui.Context, opts StackedBarOptions, in []Series) *ui.Element {
	series, entries := coloured(c, in)
	d, ok := SeriesDomain(series)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	opts.X = axisX(opts.X, opts.Labels)
	opts.Y = axisY(opts.Y, FromZero(d), opts.Y.Count)
	opts.ChartOptions = charted(opts.ChartOptions, entries)
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Stacked bars", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			stack(p, series, xs, ys, plot, barRatio(opts.Ratio))
		})
	})
}

// stack draws each category as a run of bands between the running totals of
// the series below it. The running totals come from [Bounds] rather than from
// the caller, so a band can never float above the one on top of it, and an
// ordinary column chart is the same code with every bottom at the baseline.
func stack(p *ui.Painter, series []Series, xs, ys Scale, plot ui.Rect, ratio float32) {
	lower, upper := Bounds(valueRows(series))
	for i, s := range series {
		if i >= len(upper) {
			continue
		}
		for j, pt := range s.Points {
			if j >= len(upper[i]) {
				continue
			}
			_, mid, _ := xs.Share(int(pt.X))
			w := bandWidth(xs, int(pt.X), ratio)
			if w <= 0 {
				continue
			}
			top, bottom := ys.At(upper[i][j]), ys.At(lower[i][j])
			rect := inside(ui.Rect{X: mid - w/2, Y: minf(top, bottom), W: w, H: maxf(absf32(top-bottom), 1)}, plot)
			p.Fill(rect, s.Color.Alpha(0.9), minf(w/3, theme.SmallRadius))
		}
	}
}

// bandWidth is how wide a bar of category i is: its share of the axis less
// what the bars leave as a gap.
func bandWidth(xs Scale, i int, ratio float32) float32 {
	start, _, end := xs.Share(i)
	// The distance between the two edges rather than the second less the
	// first: a band axis that runs upwards — the rows of a heatmap, the
	// categories of a bar chart on its side — comes back the other way.
	return maxf(absf32(end-start)*ratio, 1)
}

// inside keeps a mark inside the plot area even when its value fell outside
// the axis, so that a bar never draws over the axis' labels.
func inside(r, plot ui.Rect) ui.Rect {
	r.X = maxf(r.X, plot.X)
	r.Y = maxf(r.Y, plot.Y)
	r.W = minf(r.W, plot.X+plot.W-r.X)
	r.H = minf(r.H, plot.Y+plot.H-r.Y)
	return r
}

// ── histogram ──────────────────────────────────────────────────────────────

// HistogramOptions configure a [Histogram].
type HistogramOptions struct {
	ChartOptions
	// Values are the readings binned into buckets.
	Values []float64
	// Bins is how many buckets to split them into; zero means
	// [DefaultTickCount].
	Bins int
}

// Histogram is how many readings fell in each stretch of the scale: the shape
// of a distribution rather than its order in time.
//
// The buckets are the caller's count rather than the caller's widths, because
// equal buckets are the only kind two histograms of two different things can
// be compared by.
func Histogram(c *ui.Context, opts HistogramOptions) *ui.Element {
	bins := Bins(opts.Values, max(opts.Bins, DefaultTickCount))
	if len(bins) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	labels := make([]string, len(bins))
	series := make([]Series, 1)
	series[0].Name = "Count"
	series[0].Points = make([]Point, len(bins))
	for i, b := range bins {
		labels[i] = binLabel(b)
		series[0].Points[i] = Point{X: float64(i), Y: float64(b.Count)}
	}
	series, entries := coloured(c, series)
	opts.X = axisX(opts.X, labels)
	// A count starts at zero: the height of a bucket is how many readings are
	// in it, and a bucket with none in it is empty rather than short.
	opts.Y = axisY(opts.Y, Domain{Min: 0, Max: maxCount(bins)}, opts.Y.Count)
	opts.ChartOptions = charted(opts.ChartOptions, entries)
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Histogram", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			// The buckets touch. A gap between them would read as a missing
			// stretch of the scale, which is the one thing a histogram of a
			// continuous range cannot have.
			stack(p, series, xs, ys, plot, 1)
		})
	})
}

// binLabel is a bucket's two edges as one label: what a histogram's axis says
// under a bucket rather than where its edges happen to fall.
func binLabel(b Bin) string {
	decimals := decimalsFor(b.Width())
	return Group(b.Lo, decimals) + "–" + Group(b.Hi, decimals)
}

func maxCount(bins []Bin) float64 {
	most := 0
	for _, b := range bins {
		most = max(most, b.Count)
	}
	return float64(most)
}

// ── points ─────────────────────────────────────────────────────────────────

// ScatterOptions configure a [ScatterChart].
type ScatterOptions struct {
	ChartOptions
	// Series are the marks; each point is one observation.
	Series []Series
	// Radius is how big the marks are; zero uses three DIPs.
	Radius float32
}

// ScatterChart is a mark per observation and nothing between them: the shape
// of a relationship, which a line drawn through the points would hide.
func ScatterChart(c *ui.Context, opts ScatterOptions) *ui.Element {
	chart, series, ok := cartesian(c, opts.Series, opts.ChartOptions)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	return canvas(c, chart, func(f FrameResult) {
		marks(c, f, "Scatter", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			for _, s := range series {
				dots(p, xs, ys, s.Points, dotRadius(c, opts.Radius)*1.2, s.Color)
			}
		})
	})
}

// Bubble is one mark in a bubble chart: where it falls, how big it is and
// what it is called.
type Bubble struct {
	Point
	// Size scales the mark against the largest of the set. Its area, not its
	// width, is what Size is read as, which is why the radius takes the root
	// of it.
	Size float64
	// Name is what the legend and a tooltip call it.
	Name string
	// Color is what it is drawn in; the zero colour takes the palette's.
	Color ui.Color
}

// BubblesFrom is a set of bubbles from the three parallel slices their data
// arrives in. They have to be the same length: a bubble with no size is a
// mark that cannot be drawn, and a size with no bubble is one that is lost.
func BubblesFrom(name string, xs, ys, sizes []float64) []Bubble {
	if len(xs) != len(ys) || len(xs) != len(sizes) {
		panic("chart: BubblesFrom needs an x, a y and a size for every bubble")
	}
	out := make([]Bubble, len(xs))
	for i := range out {
		out[i] = Bubble{Point: Point{X: xs[i], Y: ys[i]}, Size: sizes[i], Name: name}
	}
	return out
}

// BubbleOptions configure a [BubbleChart].
type BubbleOptions struct {
	ChartOptions
	// Bubbles are the marks.
	Bubbles []Bubble
	// Radius is the radius of the largest bubble, in DIPs; zero uses a sixth
	// of the plot's shorter side.
	Radius float32
}

// BubbleChart is a scatter whose marks are sized: where each thing falls on
// both axes, and how big it is, all in the same plot.
func BubbleChart(c *ui.Context, opts BubbleOptions) *ui.Element {
	bubbles := opts.Bubbles
	if len(bubbles) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	ys := make([]float64, len(bubbles))
	xs := make([]float64, len(bubbles))
	biggest := 0.0
	labels := make([]string, 0, len(bubbles))
	for i, b := range bubbles {
		ys[i], xs[i] = b.Y, b.X
		biggest = max(biggest, b.Size)
		if b.Name != "" {
			labels = append(labels, b.Name)
		}
	}
	yd, _ := DomainOf(ys)
	xd, _ := DomainOf(xs)
	opts.X = axisXValues(opts.X, xd)
	opts.Y = axisValues(opts.Y, yd, opts.Y.Count)
	if opts.Legend.Entries == nil && len(labels) > 0 {
		opts.Legend = named(c, labels, len(labels))
	}
	k := core.Tokens(c)
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Bubbles", func(p *ui.Painter, plot ui.Rect, bx, by Scale) {
			radius := opts.Radius
			if radius <= 0 {
				radius = minf(plot.W, plot.H) / 6
			}
			// The root, because area is what a bubble's size is read as: a
			// bubble twice the radius has four times the area, and a chart
			// that drew it twice as wide would be lying about that.
			colors := Palette(c, len(bubbles))
			for i, b := range bubbles {
				x, y, ok := screen(bx, by, b.Point)
				if !ok {
					continue
				}
				r := radius
				if biggest > 0 {
					r = radius * float32(math.Sqrt(max(b.Size, 0)/biggest))
				}
				color := b.Color
				if color == (ui.Color{}) {
					color = colorAt(colors, i)
				}
				internal.Dot(p, x, y, r, color.Alpha(0.75))
				// A hairline in the background's own colour, because two
				// bubbles that overlap are two things and a chart that
				// merged them has lost one.
				internal.Ring(p, x, y, r, theme.BorderWidth, k.Background)
			}
		})
	})
}

// ── waterfall ──────────────────────────────────────────────────────────────

// WaterfallOptions configure a [WaterfallChart].
type WaterfallOptions struct {
	ChartOptions
	// Labels are the steps, one per value.
	Labels []string
	// Values are what each step moved the running total by. A negative one is
	// a fall, and is drawn in the danger colour rather than the success one.
	Values []float64
	// Total adds a bar at the end standing from zero, for where the walk
	// ended up.
	Total bool
}

// WaterfallChart is a walk: each bar how far the running total moved from one
// step to the next, so that a rise and a fall read as one story rather than as
// a hundred bars a reader has to add up in their head.
func WaterfallChart(c *ui.Context, opts WaterfallOptions) *ui.Element {
	running := Running(opts.Values)
	if len(running) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	k := core.Tokens(c)
	labels := opts.Labels
	if opts.Total {
		labels = append(append([]string(nil), labels...), "Total")
	}
	opts.X = axisX(opts.X, labels)
	opts.Y = axisY(opts.Y, FromZero(walkDomain(running)), opts.Y.Count)
	if opts.Legend.Entries == nil {
		// The legend reads the colours the bars are drawn in — the rise and
		// fall of the walk are semantic, the same green and red a
		// candlestick wears, and a legend that handed out the palette's
		// own would key three colours that are on no bar at all.
		opts.Legend = LegendOptions{Entries: []LegendEntry{
			{Name: "Step", Color: k.Success, Mark: MarkSquare},
			{Name: "Fall", Color: k.Danger, Mark: MarkSquare},
			{Name: "Total", Color: k.Accent, Mark: MarkSquare},
		}}
	}
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Waterfall", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			bottom := 0.0
			for i, top := range running {
				color := k.Success
				if i < len(opts.Values) && opts.Values[i] < 0 {
					color = k.Danger
				}
				column(p, xs, i, minf64(bottom, top), maxf64(bottom, top), ys,
					color.Alpha(0.85), barRatio(0.7), plot)
				bottom = top
			}
			if opts.Total {
				// The closing bar stands from zero rather than from where
				// the walk stopped, because that is what a total is.
				column(p, xs, len(running), 0, running[len(running)-1], ys,
					k.Accent.Alpha(0.9), barRatio(0.7), plot)
			}
		})
	})
}

// walkDomain is the range a walk's bars need. The values themselves say
// nothing about it: a step of +5 can land anywhere.
func walkDomain(running []float64) Domain {
	d, ok := DomainOf(running)
	if !ok {
		return Domain{Min: 0, Max: 1}
	}
	return d
}

// column draws one bar of a band axis between two values.
func column(p *ui.Painter, xs Scale, i int, lo, hi float64, ys Scale, color ui.Color, ratio float32, plot ui.Rect) {
	_, mid, _ := xs.Share(i)
	w := bandWidth(xs, i, ratio)
	if w <= 0 {
		return
	}
	top, bottom := ys.At(hi), ys.At(lo)
	p.Fill(inside(ui.Rect{X: mid - w/2, Y: minf(top, bottom), W: w, H: maxf(absf32(top-bottom), 1)}, plot),
		color, minf(w/3, theme.SmallRadius))
}

// ── pareto ─────────────────────────────────────────────────────────────────

// ParetoOptions configure a [ParetoChart].
type ParetoOptions struct {
	ChartOptions
	// Labels are the categories. They are sorted along with the values, so
	// the caller's order is the tie-break rather than the chart's.
	Labels []string
	// Values are what each category is worth.
	Values []float64
}

// ParetoChart is the bars in order of size with the running share over them:
// which few of the categories are most of the total.
//
// It is two scales on one plot, so the frame measures a gutter on its right
// for the percentage axis and the bar scale is placed in the plot below it.
// The frame is given the percentage scale because that is the one whose labels
// have to be measured; the bar scale is a range and a plot area, and turning
// one into the other is what [Scale.Span] is for.
func ParetoChart(c *ui.Context, opts ParetoOptions) *ui.Element {
	values, labels := Descending(opts.Values, opts.Labels)
	if len(values) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	k := core.Tokens(c)
	share := NewLinear(Domain{Min: 0, Max: 100}, 0, 1)
	// Cumulative is already in percent, so this axis writes its values as
	// they are rather than through Percent — which is for values that are
	// fractions, and would turn 100 into 10,000%.
	right := AxisOptions{Side: Right, Scale: share, Count: DefaultTickCount, Format: percentOf100}
	across := axisX(opts.X, labels)
	chart := opts.ChartOptions
	chart.X, chart.Y = across, right
	if chart.Legend.Entries == nil {
		chart.Legend = named(c, labels, len(labels))
	}
	bars := make([]Series, 1)
	bars[0].Name = "Count"
	bars[0].Points = make([]Point, len(values))
	for i, v := range values {
		bars[0].Points[i] = Point{X: float64(i), Y: v}
	}
	bars, _ = coloured(c, bars)
	valuesDomain, _ := DomainOf(values)
	barAxis := counted(FromZero(valuesDomain), DefaultTickCount)

	return Frame(c, frameOptions(chart), func(f FrameResult) {
		marks(c, f, "Pareto", func(p *ui.Painter, plot ui.Rect, xs, _ Scale) {
			// The bars stand on the frame's own x axis and on a bar scale
			// placed here, because the frame was given the percentage
			// scale to measure its right-hand gutter with.
			ys := barAxis.Span(plot.Y+plot.H, plot.Y)
			ps := share.Span(plot.Y+plot.H, plot.Y)
			stack(p, bars, xs, ys, plot, 1)
			// The share line is drawn over the bars and in the accent,
			// because it is the one thing on this plot that is not a count.
			line(p, SeriesFrom("Share", FormLine, indices(len(values)), Cumulative(values)),
				xs, ps, k.Accent, theme.BorderWidth*2)
			if at := crossing(Cumulative(values), 80); at >= 0 {
				// The eight tenths line, which is the whole reason a Pareto
				// is drawn: where a few categories make up most of it.
				line(p, SeriesFrom("", FormLine, []float64{at}, []float64{at}),
					xs, ps, k.TextMuted.Alpha(0.7), theme.BorderWidth)
			}
		})
		Axis(c, f, across)
		Axis(c, f, right)
	}).Element
}

// percentOf100 writes a value that is already in percent as a percent.
func percentOf100(v, step float64) string {
	return tidy(Group(v, decimalsFor(step))) + "%"
}

func indices(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = float64(i)
	}
	return out
}

// crossing is where a running share first reaches percent, taken straight
// across the step that crossed it so the line sits where a reader would have
// put it by eye, and -1 when it never does.
func crossing(share []float64, percent float64) float64 {
	for i, v := range share {
		if v < percent {
			continue
		}
		if i == 0 {
			return 0
		}
		before := share[i-1]
		if v == before {
			return float64(i)
		}
		return float64(i-1) + (percent-before)/(v-before)
	}
	return -1
}

// ── sparks ─────────────────────────────────────────────────────────────────

// SparklineOptions configure a [Sparkline].
type SparklineOptions struct {
	// Values are the readings, drawn in order.
	Values []float64
	// Width and Height are the box it takes; zero lets the layout give it
	// one. A sparkline has no axes and no gutters, so its box is its plot
	// area and there is nothing to measure.
	Width, Height float32
	// Color is what the line is drawn in; the zero colour takes the accent.
	Color ui.Color
	// Fill washes the ground under the line, which is what makes a shape
	// readable at thirty pixels across.
	Fill bool
	// Baseline is what the fill runs down to; nil means the bottom of the
	// box for a line of positive readings and the top for a line of negative
	// ones, the same rule a bar chart's baseline follows.
	Baseline *float64
}

// Sparkline is a line with no axes, no labels and no numbers: the shape of a
// trend, for a table row or a card corner where a whole chart would not fit.
func Sparkline(c *ui.Context, opts SparklineOptions) *ui.Element {
	k := core.Tokens(c)
	color := ink(opts.Color, k)
	d, ok := DomainOf(opts.Values)
	if !ok {
		// A sparkline with no readings draws no line rather than an axis
		// over nothing: at this size there is no room to say so in, but it
		// still says what it is for assistive technology.
		return mark(c, "Sparkline", opts.Width, opts.Height, func(*ui.Painter, ui.Rect) {})
	}
	lo, hi := d.Min, d.Max
	if b := opts.Baseline; b != nil {
		lo, hi = minf64(lo, *b), maxf64(hi, *b)
	}
	return mark(c, "Sparkline", opts.Width, opts.Height, func(p *ui.Painter, r ui.Rect) {
		if r.W <= 0 || r.H <= 0 {
			return
		}
		ys := NewLinear(Domain{Min: lo, Max: hi}, 0, 1).Span(r.Y, r.Y+r.H)
		xs := NewLinear(Domain{Min: 0, Max: float64(max(len(opts.Values)-1, 1))}, r.X, r.X+r.W)
		series := SeriesFrom("", FormLine, indices(len(opts.Values)), opts.Values)
		p.Clip(r, 0, func() {
			if opts.Fill {
				area(p, series, xs, ys, r.Y+r.H, color.Alpha(0.16))
			}
			line(p, series, xs, ys, color, theme.BorderWidth*2)
		})
	})
}

// SparkBarOptions configure a [SparkBar].
type SparkBarOptions struct {
	// Values are the readings; each takes a bar's height in proportion to the
	// largest of them.
	Values []float64
	// Width and Height are the box it takes; zero lets the layout give it
	// one.
	Width, Height float32
	// Color is what the bars are drawn in; the zero colour takes the accent.
	Color ui.Color
}

// SparkBar is a run of bars with no axes: which of a set of readings is high,
// at a glance, in the same corner of a row a sparkline would take.
func SparkBar(c *ui.Context, opts SparkBarOptions) *ui.Element {
	color := ink(opts.Color, core.Tokens(c))
	values := opts.Values
	return mark(c, "Spark bar", opts.Width, opts.Height, func(p *ui.Painter, r ui.Rect) {
		if len(values) == 0 || r.W <= 0 || r.H <= 0 {
			return
		}
		tallestBar := 0.0
		for _, v := range values {
			tallestBar = maxf64(tallestBar, v)
		}
		// The bars share the box out by hand rather than through a scale: a
		// spark bar's axis is the width of the box, and a scale over nothing
		// would be a lie about the one thing it has to say.
		gap := r.W / float32(len(values)*3)
		width := (r.W - gap*float32(len(values)-1)) / float32(len(values))
		for i, v := range values {
			h := float32(1)
			if tallestBar > 0 {
				h = maxf(float32(float64(r.H)*max(v, 0)/tallestBar), 1)
			}
			x := r.X + float32(i)*(width+gap)
			p.Fill(ui.Rect{X: x, Y: r.Y + r.H - h, W: width, H: h},
				color.Alpha(0.8), minf(width/2, theme.SmallRadius/2))
		}
	})
}

// ── shared marks ───────────────────────────────────────────────────────────

// lineWidth is a line's weight: two DIPs unless the caller says otherwise,
// which is thick enough to read against a grid and thin enough not to hide
// the marks underneath it.
func lineWidth(width float32) float32 {
	if width > 0 {
		return width
	}
	return theme.BorderWidth * 2
}

// barRatio is how much of its band a bar takes: seven tenths, so that a row of
// bars has a visible gap in it rather than touching into one shape.
func barRatio(ratio float32) float32 {
	if ratio > 0 {
		return ratio
	}
	return 0.7
}

// dotRadius is a mark's size, off the window's own spacing scale so that a
// chart's marks get bigger with the density a reader has set.
func dotRadius(c *ui.Context, radius float32) float32 {
	if radius > 0 {
		return radius
	}
	return unit(c) * 0.625
}

// areaFill is how solid a filled series is: a wash, because a fill at full
// strength hides the grid lines a reader takes their values off.
func areaFill(fill float32) float32 {
	if fill > 0 {
		return fill
	}
	return 0.18
}

// mountainFill is a layered area's opacity: faint enough that four of them
// stacked do not turn into one solid shape.
func mountainFill(fill float32) float32 {
	if fill > 0 {
		return fill
	}
	return 0.3
}

func minf64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxf64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
