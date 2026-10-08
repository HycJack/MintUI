package chart

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The charts in this file are the ones whose marks are not one number each:
// a grid of cells, a wall of days, a set of candles with two numbers apiece,
// and a set of axes that are not an x and a y at all. They share one thing —
// every one of them turns a value into a colour or a box, so the rule they all
// obey is that the value belongs to the caller's data and the drawing belongs
// to the frame.

// ── heatmap ────────────────────────────────────────────────────────────────

// HeatmapOptions configure a [HeatmapChart].
type HeatmapOptions struct {
	ChartOptions
	// Rows are the names down the side and Cols the names across the top;
	// Values[i][j] is the cell at their intersection. A row that is short
	// leaves the cells past its end empty rather than reading off the end of
	// another row.
	Rows, Cols []string
	Values     [][]float64
	// Write puts each cell's number inside it, where the cell is big enough
	// for it — which is the only place a number in a heatmap should go.
	Write bool
	// Min and Max are the ends of the colour scale; zero means the data's own
	// span. Setting them is how two heatmaps of different data are put on one
	// scale.
	Min, Max float64
	// Reverse runs the colour scale from the accent at the low end to the
	// surface at the high one, for a scale where the small numbers are the
	// ones to notice.
	Reverse bool
}

// HeatmapChart is a grid of cells, each coloured by its value: a table that a
// reader can take a pattern off without reading a single number.
func HeatmapChart(c *ui.Context, opts HeatmapOptions) *ui.Element {
	d, ok := gridDomain(opts.Values)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	if opts.Min == 0 && opts.Max == 0 {
		opts.Min, opts.Max = d.Min, d.Max
	}
	if opts.Max <= opts.Min {
		opts.Max = opts.Min + 1
	}
	// Both axes are band scales — the columns across the top and the rows up
	// the side — so the frame measures both gutters by the widest name in
	// them, which is the one thing a grid of words cannot guess at.
	opts.X = axisX(opts.X, opts.Cols)
	opts.Y = axisX(opts.Y, opts.Rows)
	k := core.Tokens(c)
	size := theme.CaptionSize
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Heatmap", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			gap := unit(c) / 2
			for i, row := range opts.Values {
				top, bottom, ok := cell(ys, i)
				if !ok {
					continue
				}
				height := maxf(bottom-top-2*gap, 1)
				for j, v := range row {
					left, right, ok := cell(xs, j)
					if !ok {
						continue
					}
					width := maxf(right-left-2*gap, 1)
					fill := HeatColor(v, opts.Min, opts.Max, k)
					if opts.Reverse {
						fill = HeatColor(v, opts.Min, opts.Max, reversed(k))
					}
					p.Fill(ui.Rect{X: left + gap, Y: top + gap, W: width, H: height},
						fill, theme.SmallRadius/2)
					if !opts.Write {
						continue
					}
					label := Group(v, decimalsFor(opts.Max-opts.Min))
					w, h := LabelWidth(c, label, size), LabelHeight(c, size)
					// Only written where it fits: a number written over
					// a cell too small for it is a number over the cell
					// next door as well.
					if w <= width && h <= height {
						p.Text(left+gap+(width-w)/2, top+gap+(height-h)/2, label, size, k.Text)
					}
				}
			}
		})
	})
}

// reversed is the tokens with the heat scale the other way round, which is a
// pair of tokens rather than a flag every caller has to remember to pass on
// down to [HeatColor].
func reversed(k theme.Tokens) theme.Tokens {
	k.Surface, k.Accent = k.Accent, k.Surface
	return k
}

// cell is where a cell of the grid is: its two edges along the axis and
// whether there is one at all.
//
// The two edges come back the smaller first, because a row's axis runs
// upwards while a column's runs rightwards — and a grid of cells is made of
// rectangles, which need a top and a bottom rather than a direction.
func cell(s Scale, i int) (lo, hi float32, ok bool) {
	if !s.OK() || s.Kind() != KindBand {
		return 0, 0, false
	}
	start, _, end := s.Share(i)
	if end < start {
		start, end = end, start
	}
	return start, end, end > start
}

// gridDomain is the span of a grid's values, and whether it holds any. The
// rows need not be the same length: a short row is a grid with some cells
// missing, which is a thing that happens and not a mistake.
func gridDomain(values [][]float64) (Domain, bool) {
	flat := make([]float64, 0, len(values)*4)
	for _, row := range values {
		flat = append(flat, row...)
	}
	return DomainOf(flat)
}

// ── calendar heatmap ───────────────────────────────────────────────────────

// CalendarHeatmapOptions configure a [CalendarHeatmap].
type CalendarHeatmapOptions struct {
	ChartOptions
	// Days are the cells, each with the date it belongs to. The dates decide
	// where every cell falls, so a day that is missing from the list is a day
	// the chart has nothing to say about.
	Days []Day
	// Weekday is the day the weeks start on; Sunday unless the caller says
	// Monday, which is what a reader of a working week expects.
	Weekday time.Weekday
	// MonthLabels writes the month over the first week of it.
	MonthLabels bool
}

// CalendarHeatmap is a wall of days: a year of them, week by week, each cell
// coloured by what that day was worth — the shape of a contribution graph
// without the axes or the grid to get in the way.
func CalendarHeatmap(c *ui.Context, opts CalendarHeatmapOptions) *ui.Element {
	grid, ok := NewCalendarGrid(opts.Days, opts.Weekday)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	d, ok := DomainOf(dayValues(opts.Days))
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	weekdays := make([]string, 7)
	for i := range weekdays {
		weekdays[i] = (opts.Weekday + time.Weekday(i)).String()[:3]
	}
	opts.X = axisX(opts.X, weekLabels(grid))
	opts.Y = axisX(opts.Y, weekdays)
	if opts.MonthLabels {
		// The month names draw just above the first row of cells, which is
		// outside the plot: the frame keeps room for them there, or the
		// labels land on whatever the frame put above the chart.
		opts.Pad += LabelHeight(c, theme.CaptionSize) + unit(c)
	}
	k := core.Tokens(c)
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Calendar heatmap", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			for _, day := range opts.Days {
				week, weekday, ok := grid.Cell(day.Date)
				if !ok {
					continue
				}
				x, right, _ := cell(xs, week)
				top, bottom, _ := cell(ys, weekday)
				gap := unit(c) / 2
				p.Fill(ui.Rect{X: x + gap, Y: top + gap,
					W: maxf(right-x-2*gap, 1), H: maxf(bottom-top-2*gap, 1)},
					HeatColor(day.Value, d.Min, d.Max, k), gap)
			}
			if !opts.MonthLabels {
				return
			}
			// The month over the first week it appears in, and not over
			// every week of it: twelve labels a year is a calendar, and
			// fifty-two is noise.
			last := time.Month(0)
			for week := range grid.Weeks {
				at := grid.From.AddDate(0, 0, week*7)
				if at.Month() == last {
					continue
				}
				last = at.Month()
				x, _, _ := cell(xs, week)
				p.Text(x, cellTop(ys, 0)-LabelHeight(c, theme.CaptionSize)-unit(c),
					at.Month().String()[:3], theme.CaptionSize, k.TextMuted)
			}
		})
	})
}

func cellTop(s Scale, i int) float32 {
	start, _, _ := s.Share(i)
	return start
}

func dayValues(days []Day) []float64 {
	out := make([]float64, len(days))
	for i, d := range days {
		out[i] = d.Value
	}
	return out
}

// weekLabels names each column by the date its week starts on, so that a
// reader can find April in a wall of cells. The label is short because the
// columns are one week wide and there are fifty of them.
func weekLabels(grid CalendarGrid) []string {
	out := make([]string, grid.Weeks)
	for i := range out {
		out[i] = grid.From.AddDate(0, 0, i*7).Format("Jan 2")
	}
	return out
}

// ── candles ────────────────────────────────────────────────────────────────

// Candle is one period of a market: what it opened at, where it got to, and
// where it closed.
type Candle struct {
	// Label names the period — a date, a minute, a bar number.
	Label string
	// Open is where it started.
	Open float64
	// High is the highest it reached, which is the top of the wick.
	High float64
	// Low is the lowest it fell to, which is the bottom of the wick.
	Low float64
	// Close is where it ended, and with Open says which way the candle
	// points.
	Close float64
}

// CandlestickOptions configure a [CandlestickChart].
type CandlestickOptions struct {
	ChartOptions
	// Candles are the periods, in order.
	Candles []Candle
	// Up and Down are what a rising and a falling candle are drawn in; zero
	// takes the success and danger tokens, which are the two colours in the
	// interface that mean a thing went well and a thing went wrong.
	Up, Down ui.Color
	// Ratio is how much of its band a candle takes; zero uses 0.6, which is
	// narrower than a bar's because a candle has a wick either side of it.
	Ratio float32
}

// CandlestickChart is a chart of what a price did in each period: a body
// between where it opened and where it closed, a wick out to the highest and
// the lowest it reached.
//
// A candle carries four numbers where a line carried one, and the body is the
// only part of it that is not the same colour twice — so the body is drawn
// hollow when it rises and filled when it falls, the way a chart on paper
// does it, which is the convention every reader of one already knows.
func CandlestickChart(c *ui.Context, opts CandlestickOptions) *ui.Element {
	if len(opts.Candles) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	k := core.Tokens(c)
	up, down := opts.Up, opts.Down
	if up == (ui.Color{}) {
		up = k.Success
	}
	if down == (ui.Color{}) {
		down = k.Danger
	}
	labels := make([]string, len(opts.Candles))
	lo, hi := 0.0, 0.0
	seen := false
	for i, candle := range opts.Candles {
		labels[i] = candle.Label
		for _, v := range []float64{candle.Low, candle.High, candle.Open, candle.Close} {
			// The first value seeds the span; after that each value widens
			// it. The seed has to happen only once — an unconditional
			// assignment here would leave the span as the last candle's
			// alone, and every earlier candle would draw outside the axis
			// and be clipped away.
			if !seen {
				lo, hi, seen = v, v, true
				continue
			}
			lo, hi = minf64(lo, v), maxf64(hi, v)
		}
	}
	if !seen {
		return nothing(c, opts.ChartOptions)
	}
	opts.X = axisX(opts.X, labels)
	opts.Y = axisValues(opts.Y, Domain{Min: lo, Max: hi}, opts.Y.Count)
	ratio := opts.Ratio
	if ratio <= 0 {
		ratio = 0.6
	}
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Candlestick", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			for i, candle := range opts.Candles {
				rising := candle.Close >= candle.Open
				color := down
				if rising {
					color = up
				}
				mid := midOf(xs, i)
				width := bandWidth(xs, i, ratio)
				if width <= 0 {
					continue
				}
				// The wick first, so the body's own edge is what is left
				// on top of it.
				p.Line(mid, ys.At(candle.High), mid, ys.At(candle.Low),
					theme.BorderWidth, color.Alpha(0.8))
				top, bottom := Body(candle.Open, candle.Close)
				rect := inside(ui.Rect{X: mid - width/2, Y: ys.At(top),
					W: width, H: maxf(ys.At(bottom)-ys.At(top), 1)}, plot)
				if rising {
					// Hollow: a rising candle drawn solid would be
					// indistinguishable from a falling one at this size.
					p.Fill(rect, k.Background, minf(width/4, theme.SmallRadius/2))
					p.Stroke(rect, color, minf(width/4, theme.SmallRadius/2), theme.BorderWidth)
				} else {
					p.Fill(rect, color, minf(width/4, theme.SmallRadius/2))
				}
			}
		})
	})
}

// ── candle countdown ───────────────────────────────────────────────────────

// CandleCountdownOptions configure a [CandleCountdown].
type CandleCountdownOptions struct {
	// Left is how much of the countdown is still to run.
	Left time.Duration
	// Total is how long it was to begin with; zero means the countdown runs
	// from left to empty. Both come from the caller: a chart that read the
	// clock would be a chart that drew a different number every time it was
	// screenshotted, and a countdown nobody can test is a countdown that
	// ships broken.
	Total time.Duration
	// Width and Height are the box it takes; zero lets the layout give it
	// one.
	Width, Height float32
	// Label is what the countdown says, written under it.
	Label string
	// Lit and Out are the colours of a candle that has time left and one
	// that has burnt down; zero takes the accent and the border.
	Lit, Out ui.Color
	// Candles is how many candles to draw; zero works it out from the box.
	Candles int
}

// CandleCountdown is a countdown drawn as a row of candles: one for each
// stretch of the time left, burnt down from the end. It answers "how long
// until this" with something a reader can see emptying rather than with a
// number they have to subtract from another number.
func CandleCountdown(c *ui.Context, opts CandleCountdownOptions) *ui.Element {
	k := core.Tokens(c)
	total := opts.Total
	if total <= 0 {
		total = opts.Left
	}
	left := opts.Left
	if left < 0 {
		left = 0
	}
	if total > 0 && left > total {
		left = total
	}
	lit, out := opts.Lit, opts.Out
	if lit == (ui.Color{}) {
		lit = k.Accent
	}
	if out == (ui.Color{}) {
		out = k.Border
	}
	spent := total - left
	return mark(c, dial("Countdown", opts.Label), opts.Width, opts.Height, func(p *ui.Painter, r ui.Rect) {
		count := opts.Candles
		if count < 1 {
			count = max(int(r.W/(unit(c)*4)), 1)
		}
		if count < 1 || r.W <= 0 || r.H <= 0 {
			return
		}
		burnt := int(float64(count) * float64(spent) / float64(max(total, 1)))
		gap := r.W / float32(count*3)
		width := (r.W - gap*float32(count-1)) / float32(count)
		height := r.H * 0.7
		y := r.Y + r.H - height
		// Drawn straight into the box it was given: a dial, a
		// bullet and a countdown have no frame, and an element made
		// while painting is one that never paints.
		for i := range count {
			x := r.X + float32(i)*(width+gap)
			color := out
			if i >= burnt {
				color = lit
			}
			// A wick, so that each mark reads as a candle rather than
			// as a bar of a bar chart.
			mid := x + width/2
			p.Line(mid, y-unit(c), mid, y+height+unit(c), theme.BorderWidth, color.Alpha(0.7))
			p.Fill(ui.Rect{X: x, Y: y, W: width, H: height}, color.Alpha(0.85),
				minf(width/3, theme.SmallRadius/2))
		}
		if opts.Label == "" {
			return
		}
		w := LabelWidth(c, opts.Label, theme.CaptionSize)
		p.Text(r.X+(r.W-w)/2, r.Y, opts.Label, theme.CaptionSize, k.TextMuted)
	})
}

// ── parallel coordinates ───────────────────────────────────────────────────

// ParallelOptions configure a [ParallelCoordinates].
type ParallelOptions struct {
	ChartOptions
	// Axes are the measures, one per value of each series, in the order they
	// are drawn across the plot.
	Axes []string
	// Series are the lines: one per row of a table, with one value per axis.
	Series []Series
	// Rings is how many labels each axis carries; zero uses three.
	Rings int
}

// ParallelCoordinates is a table drawn as lines: one axis per column, each
// scaled by itself, and one line per row across all of them. It is the way to
// see which rows of a table are alike, which is the one question a table on
// its own cannot answer.
func ParallelCoordinates(c *ui.Context, opts ParallelOptions) *ui.Element {
	series, entries := coloured(c, opts.Series)
	if len(opts.Axes) == 0 || len(series) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	opts.ChartOptions = charted(opts.ChartOptions, entries)
	rings := opts.Rings
	if rings < 2 {
		rings = 3
	}
	k := core.Tokens(c)
	// A row of the frame's own padding for the axis names, which are written
	// above the plot and so need room the frame's legend gutter does not
	// hold: the panel charts' names are the chart's, not a legend's.
	if pad := LabelHeight(c, theme.RowSize) + unit(c); opts.Pad < pad {
		opts.Pad = pad
	}
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		// The axis names go along the top and the outer ring's values hang
		// off the top and bottom edges of the plot, so this one draws into
		// the frame rather than into the plot: clipped to the plot, the
		// names are gone and the outer values are cut in half.
		marksWide(c, f, "Parallel coordinates", func(p *ui.Painter, plot ui.Rect) {
			// Each axis is its own scale over its own column, because a
			// count and a duration have nothing in common to share a scale
			// over. The axis names go along the top, the values up the side of
			// their own axis.
			columns := make([]Scale, len(opts.Axes))
			for i := range opts.Axes {
				d, _ := Column(series, i)
				columns[i] = counted(d, rings).Span(plot.Y+plot.H, plot.Y)
			}
			for i, axis := range opts.Axes {
				x := spread(plot, len(opts.Axes), i)
				p.Line(x, plot.Y, x, plot.Y+plot.H, theme.BorderWidth, k.Border)
				if axis != "" {
					w := LabelWidth(c, axis, theme.RowSize)
					p.Text(x-w/2, plot.Y-LabelHeight(c, theme.RowSize)-unit(c), axis, theme.RowSize, k.Text)
				}
				// The values come off the same tick machinery as every
				// other axis in this package, so they are spaced, rounded
				// and thinned the same way — and two of them never collide.
				for _, t := range Ticks(c, columns[i], TickOptions{Count: rings, Size: theme.CaptionSize}) {
					if !t.Drawn {
						continue
					}
					p.Text(x+unit(c), t.Pos-LabelHeight(c, theme.CaptionSize)/2, t.Label,
						theme.CaptionSize, k.TextMuted)
				}
			}
			for _, s := range series {
				var path ui.Path
				for i, pt := range s.Points {
					if i >= len(columns) {
						break
					}
					x := spread(plot, len(opts.Axes), i)
					y := columns[i].At(pt.Y)
					if i == 0 {
						path.MoveTo(x, y)
					} else {
						path.LineTo(x, y)
					}
				}
				p.StrokePath(&path, theme.BorderWidth, s.Color.Alpha(0.7))
			}
		})
	})
}

// spread is where the i-th of n axes stands across a plot area: the same
// division a band scale makes, written out because these axes are drawn rather
// than measured and there is no scale to ask.
func spread(plot ui.Rect, n, i int) float32 {
	if n <= 0 {
		return plot.X
	}
	return plot.X + plot.W*float32(i+1)/float32(n+1)
}

// ── range highlight ────────────────────────────────────────────────────────

// RangeHighlightOptions configure a [RangeHighlight].
type RangeHighlightOptions struct {
	// Frame is the chart's frame: the band is drawn across its plot area, in
	// the same place as everything else on the chart.
	Frame FrameResult
	// X is the axis the range is a range of.
	X Scale
	// From and To are the two ends of the range, in the axis' own values and
	// in whatever order the reader dragged it.
	From, To float64
	// Color is what the band is washed in; the zero colour takes the accent,
	// faded back so the data under it still reads.
	Color ui.Color
	// Label is written along the top of the band, where the data is least
	// likely to be.
	Label string
	// Edges draws the two lines the band starts and ends on, for a highlight
	// whose ends are the point.
	Edges bool
}

// RangeHighlight washes a range of a chart's x axis: "this week", "the hours
// between 9 and 5", "since the deploy". It draws nothing of its own and
// means nothing on its own — what the range means is the caller's, which is
// the same rule a brush's [Range] follows.
func RangeHighlight(c *ui.Context, opts RangeHighlightOptions) *ui.Element {
	k := core.Tokens(c)
	color := opts.Color
	if color == (ui.Color{}) {
		color = k.Accent
	}
	return marks(c, opts.Frame, "Range", func(p *ui.Painter, plot ui.Rect, _, _ Scale) {
		if !opts.X.OK() || plot.W <= 0 || plot.H <= 0 {
			return
		}
		if opts.From == opts.To {
			return
		}
		xs := opts.X.Span(plot.X, plot.X+plot.W)
		left, right := xs.At(minf64(opts.From, opts.To)), xs.At(maxf64(opts.From, opts.To))
		if right <= left {
			return
		}
		band := inside(ui.Rect{X: left, Y: plot.Y, W: right - left, H: plot.H}, plot)
		if band.W <= 0 {
			return
		}
		p.Fill(band, color.Alpha(0.12), 0)
		if opts.Edges {
			for _, x := range []float32{band.X, band.X + band.W} {
				p.Line(x, band.Y, x, band.Y+band.H, theme.BorderWidth, color.Alpha(0.6))
			}
		}
		if opts.Label == "" {
			return
		}
		size := theme.CaptionSize
		w, h := LabelWidth(c, opts.Label, size), LabelHeight(c, size)
		x := clampLabel(band.X+(band.W-w)/2, w, plot.X, plot.X+plot.W-w)
		p.Fill(ui.Rect{X: x - unit(c)/2, Y: plot.Y, W: w + unit(c), H: h}, k.Fill, theme.SmallRadius/2)
		p.Text(x, plot.Y, opts.Label, size, k.OnFill)
	})
}
