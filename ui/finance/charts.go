package finance

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// Every chart in this file is a [chart] component with this package's data in
// it. None of them draws anything itself.
//
// That is not a shortcut and it is not a shortage of effort; it is the reason
// they are correct. A chart's gutters are *measured* — the widest y-axis label
// of one chart is a date and of the next is "1,200,000", and a gutter guessed
// rather than measured puts the labels over the data at one window size and not
// at another. ui/chart measures them, once, for fifty-four types. A hand-drawn
// chart here would have to answer that question a fifty-fifth time and would
// answer it slightly differently.
//
// What this file owns is the other half of the job: the shapes of the data, the
// scales' domains, the palette and the labels.

// ── performance ────────────────────────────────────────────────────────────

// PerformanceSeries is one line on a performance chart: the value of a thing
// over a set of points.
type PerformanceSeries struct {
	// Name is what the legend calls it.
	Name string
	// Points are the values, in order.
	Points []Point
}

// PerformanceChartOptions configure a PerformanceChart.
type PerformanceChartOptions struct {
	// Series are the lines, the caller's. A series with no colour takes the
	// palette's, from [chart.EntriesOf] — so the legend and the lines cannot
	// disagree about which colour is which.
	Series []PerformanceSeries
	// Baseline is the value the portfolio is measured against, usually a
	// fixed deposit or an index. A performance chart without one is a price
	// chart, and a price chart does not answer "did I do better than doing
	// nothing".
	Baseline *float64
	// Height is the chart's height; zero lets the layout give it one.
	Height float32
	// Labels are the x axis' categories, one per point. Empty puts a plain
	// index there, which is right for a chart whose points are evenly spaced.
	Labels []string
	// Hover is the caller's pointer, filled by the frame and shared with the
	// tooltip.
	Hover *chart.Hover
	// Grid draws the grid lines behind the data.
	Grid bool
	// Label names the chart for assistive technology, and is required: a
	// chart's numbers are drawn rather than laid out, so the name is the only
	// thing a screen reader has to go on.
	Label string
	// YLabel names the y axis.
	YLabel string
}

// PerformanceChart is what the account was worth over time, against whatever
// it would have been if nothing had been done.
//
// It is [chart.LineChart] with a band axis, because a performance chart's x
// axis is a set of dates and a set of dates takes an equal share each whether
// or not the gaps between them are equal — and a linear x axis over unequally
// spaced days draws a straight line across a weekend, which is a claim about
// the market being open on Saturday.
func PerformanceChart(c *ui.Context, opts PerformanceChartOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: PerformanceChart needs a Label; a chart's numbers are drawn rather " +
			"than laid out, so the name is all a screen reader has")
	}
	if len(opts.Series) == 0 {
		panic("finance: PerformanceChart needs at least one Series; an empty chart is a " +
			"frame around nothing")
	}

	series := make([]chart.Series, 0, len(opts.Series)+1)
	for _, s := range opts.Series {
		series = append(series, chart.Series{Name: s.Name, Form: chart.FormLine, Points: s.Points})
	}
	if opts.Baseline != nil {
		flat := make([]Point, len(opts.Series[0].Points))
		for i := range flat {
			flat[i] = Point{X: float64(i), Y: *opts.Baseline}
		}
		series = append(series, chart.Series{
			Name: "Baseline", Form: chart.FormLine, Points: flat, Color: core.Tokens(c).Border,
		})
	}
	// Colours are resolved once for the whole set, by the same index rule the
	// legend uses: each series calling Plot for itself would fall back to the
	// palette's first colour and every line would draw in it while the legend
	// handed out a different one.
	palette := chart.Palette(c, len(series))
	for i := range series {
		if series[i].Color == (ui.Color{}) {
			series[i].Color = palette[i]
		}
	}

	// One domain for every series, or two lines drawn against two
	// independently-rounded axes would cross somewhere neither of them passes
	// through and the crossing — the one thing a performance chart is for —
	// would be an artefact of the rounding.
	domain, _ := chart.SeriesDomain(series)
	x := chart.NewBand(axisLabels(opts.Labels, len(opts.Series[0].Points)), 0, 1)
	ys := chart.NewLinear(chart.FromZero(domain), 0, 1)

	frame := chart.Frame(c, chart.FrameOptions{
		Height: opts.Height,
		Label:  opts.Label,
		X:      chart.AxisOptions{Side: chart.Bottom, Scale: x, Count: 7},
		Y: chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 5,
			Label: opts.YLabel, Format: chart.Compact()},
		Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
		Hover:  opts.Hover,
		Border: opts.Grid,
	}, func(frame chart.FrameResult) {
		chart.Grid(c, chart.GridOptions{Frame: frame, X: chart.AxisOptions{Scale: x}})
		for _, s := range series {
			// The scales go in as they were built, not as frame.X() and
			// frame.Y(): the frame only places them while it paints, which is
			// after this closure has run, so reading them back here reads a
			// pair of zero scales and the chart comes out with its frame, its
			// axes and no data on it.
			chart.Plot(c, chart.PlotOptions{Frame: frame, Series: s, X: x, Y: ys})
		}
		chart.Axis(c, frame, chart.AxisOptions{Side: chart.Bottom, Scale: x, Count: 7})
		chart.Axis(c, frame, chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 5,
			Label: opts.YLabel, Format: chart.Compact()})
	})
	return frame.Element
}

// axisLabels is the caller's categories, filled out to the number of points a
// series has. A band scale with fewer labels than points leaves the last
// points in a category of their own, which draws them as a band with no
// label — the reader sees an empty column.
func axisLabels(labels []string, points int) []string {
	if len(labels) >= points {
		return labels[:points]
	}
	out := append([]string(nil), labels...)
	for len(out) < points {
		out = append(out, "")
	}
	return out
}

// ── allocation ─────────────────────────────────────────────────────────────

// AssetAllocationOptions configure an AssetAllocationChart.
type AssetAllocationOptions struct {
	// Positions are the holdings, and Cash what is not invested. Both are the
	// caller's and both are in the ring: a portfolio's cash is an allocation,
	// and a ring that left it out would say the whole is invested when part of
	// it is not.
	Positions []Position
	Cash      float64
	// Width and Height are the ring's box; zero lets it fill what it is given.
	Width, Height float32
	// Group caps how many slices there are; zero takes eight, past which a
	// reader is comparing angles by eye and a bar chart would be better. The
	// smallest holdings are folded into one "Other" slice rather than dropped,
	// so the ring still adds up to the whole.
	Group int
	// Label names the chart.
	Label string
	// Hole is the ring's inner radius as a share of the outer; zero takes the
	// library's own, which is drawn with the total in the middle.
	Hole float32
}

// AssetAllocationChart is the portfolio as a ring of what it is made of.
//
// It is [chart.DonutChart], which draws the slices, the labels and the hole,
// and the only thing decided here is the aggregation: a portfolio of forty
// holdings is not a chart, it is a list of colours, so the holdings are
// grouped by sector where they have one and by name where they do not, and the
// smallest are folded into a single slice.
func AssetAllocationChart(c *ui.Context, opts AssetAllocationOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: AssetAllocationChart needs a Label; a ring of unnamed slices is a " +
			"gradient with a hole in it")
	}
	labels, values := allocation(opts.Positions, opts.Cash, opts.Group)
	if len(values) == 0 {
		// The chart's own empty state rather than a zero-sized donut: an
		// empty ring and a ring of nothing look identical, and one of them is
		// a portfolio that has not been set up and the other is a bug.
		return chart.Frame(c, chart.FrameOptions{
			Height: opts.Height, Label: opts.Label, Border: true,
		}, func(frame chart.FrameResult) {
			chart.Empty(c, chart.EmptyOptions{
				Frame: frame, Title: "Nothing to allocate",
				Body: "This portfolio holds nothing yet.",
			})
		}).Element
	}
	e := chart.DonutChart(c, chart.DonutOptions{
		PieOptions: chart.PieOptions{
			ChartOptions: chart.ChartOptions{
				Label: opts.Label, Height: opts.Height, Border: true,
				Legend:     chart.LegendOptions{Entries: legendFor(c, labels)},
				EmptyTitle: "Nothing to allocate",
				EmptyBody:  "This portfolio holds nothing yet.",
			},
			Labels: labels, Values: values,
		},
		Hole: opts.Hole,
	})
	if opts.Width > 0 {
		e.Width(opts.Width).Shrink(0)
	}
	return e
}

// legendFor builds a legend for slices that carry no series of their own. It
// goes through EntriesOf so that the legend's colours and the slices' colours
// come from the same place — a legend built here with its own colour list
// would be a legend that can disagree with the chart it is a key to.
func legendFor(c *ui.Context, labels []string) []chart.LegendEntry {
	series := make([]chart.Series, len(labels))
	for i, l := range labels {
		series[i] = chart.Series{Name: l, Form: chart.FormBar}
	}
	return chart.EntriesOf(c, series)
}

// allocation is the grouping: sector where a position has one, the instrument
// where it does not, and cash as itself.
//
// It returns the labels and the values in the same order, which is the only
// thing that keeps a slice's name attached to it — [chart.PieChart] takes two
// parallel slices and a reordering of one without the other is a chart of the
// right numbers with the wrong names.
func allocation(positions []Position, cash float64, group int) ([]string, []float64) {
	byKey := map[string]float64{}
	for _, p := range positions {
		key := p.Sector
		if key == "" {
			key = p.Instrument
		}
		byKey[key] += p.MarketValue()
	}
	if cash != 0 {
		byKey["Cash"] += cash
	}

	all := make([]slice, 0, len(byKey))
	for name, value := range byKey {
		if value > 0 {
			all = append(all, slice{name: name, value: value})
		}
	}
	// Biggest first, then by name, so the order does not depend on Go's map
	// iteration: two frames of the same portfolio must not show two
	// different orderings of the same slices.
	for i := 1; i < len(all); i++ {
		for j := i; j > 0 && bigger(all[j], all[j-1]); j-- {
			all[j], all[j-1] = all[j-1], all[j]
		}
	}
	if group <= 0 {
		group = 8
	}
	if len(all) > group {
		rest := 0.0
		for _, s := range all[group-1:] {
			rest += s.value
		}
		all = append(all[:group-1], slice{name: "Other", value: rest})
	}

	labels := make([]string, len(all))
	values := make([]float64, len(all))
	for i, s := range all {
		labels[i] = s.name
		values[i] = s.value
	}
	return labels, values
}

// slice is one wedge of the allocation ring: a name and what it is worth.
type slice struct {
	name  string
	value float64
}

// bigger is the sort: larger first, and alphabetical between equals so that a
// tie has a stable answer.
func bigger(a, b slice) bool {
	if a.value != b.value {
		return a.value > b.value
	}
	return a.name < b.name
}

// ── the calendar ───────────────────────────────────────────────────────────

// Event is one dated thing on a calendar: an earnings report, a rate
// decision, a payroll run.
type Event struct {
	// Title is what it is called, and is required: a calendar is a list of
	// dates and an event with no name is a date with a mark on it.
	Title string
	// When is the date, as the caller writes it. It is a string rather than a
	// time.Time because the caller has an exchange calendar and its holidays
	// and this package has no opinion about either.
	When string
	// Impact is how much the event is expected to move the price, 0 to 1, and
	// is what the bar beside it is drawn from.
	Impact float64
	// Forecast, Actual and Previous are the three figures an earnings row has,
	// written out because a calendar has to show the miss and it cannot
	// compute it from a typed expectation.
	Forecast, Actual, Previous float64
	// Released says the actual is in, which is the difference between a
	// forecast and a result and is why the Actual is dimmed rather than
	// hidden while a figure has not been announced.
	Released bool
}

// calendarColumns are a calendar's columns, all of them fixed widths for the
// same reason every table here has them: a figure that wraps is a figure
// nobody can compare against the one above it.
var calendarColumns = []tableColumn{
	// Narrow on purpose: a calendar lives in half a page, and price columns
	// sized for the whole width pushed the last column past the block's
	// edge, where the header clipped to "Act" and the numbers to one digit.
	{Title: "Date", ID: "when", Width: 96},
	{Title: "Event", ID: "title", Share: 1},
	{Title: "Impact", ID: "impact", Width: 88, Align: ui.End},
	{Title: "Forecast", ID: "forecast", Width: 76, Align: ui.End},
	{Title: "Actual", ID: "actual", Width: 76, Align: ui.End},
	{Title: "Previous", ID: "previous", Width: 76, Align: ui.End},
}

// EarningsCalendarOptions configure an EarningsCalendar.
type EarningsCalendarOptions struct {
	// Events are the caller's, in the order they should be shown.
	Events []Event
	// Height is the table's height, and is required.
	Height float32
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Label names the table; empty takes the library's "Earnings".
	Label string
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// EarningsCalendar is the season's reports, with what was expected, what
// happened and what happened last time.
//
// It is a table and not a chart because the four figures have to line up in
// columns to be compared — a calendar of events as a bar chart of impacts
// would say which events matter and not what any of them reported.
func EarningsCalendar(c *ui.Context, opts EarningsCalendarOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: EarningsCalendar needs a Height; a table with no height grows to fit " +
			"every row rather than scrolling")
	}
	k := core.Tokens(c)
	name := opts.Label
	if name == "" {
		name = "Earnings"
	}
	empty := opts.Empty
	if empty == nil {
		empty = func() {
			ui.Text(c, "Nothing scheduled").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		}
	}

	return tableOf(c, tableOptions{
		columns: calendarColumns,
		rows:    len(opts.Events),
		name:    name,
		cell: func(row, col int) {
			e := opts.Events[row]
			switch calendarColumns[col].ID {
			case "when":
				ui.Text(c, e.When).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			case "title":
				ui.Text(c, e.Title).TextColor(k.Text).Grow(1).MaxLines(1).
					FontSize(core.FontSize(c, theme.RowSize))
			case "impact":
				// The bar is drawn from the impact and the impact is *not*
				// also written as a number: a reader scanning this column is
				// scanning for "how much", and a bar answers it in a shape a
				// glance can compare, where five pairs of digits cannot.
				ui.Box(c).FillWidth().Height(theme.RowSize).Shrink(0).
					Label("Impact " + impactWord(e.Impact)).
					Draw(func(p *ui.Painter, r ui.Rect) {
						p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}, k.SurfacePressed, r.H/2)
						w := r.W * clamp01(e.Impact)
						p.Fill(ui.Rect{X: r.X, Y: r.Y, W: w, H: r.H},
							impactInk(e.Impact, k), r.H/2)
					})
			case "forecast":
				mono(c, FormatPrice(e.Forecast), theme.RowSize, k.TextMuted)
			case "actual":
				if !e.Released {
					ui.Text(c, "—").TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.RowSize))
					return
				}
				// The miss is coloured by which side of the forecast it fell.
				// A calendar whose actual column is one flat colour is a
				// calendar where a beat and a miss look the same, and the beat
				// is the half of that comparison a reader acts on.
				mono(c, FormatPrice(e.Actual), theme.RowSize, beatInk(e, k))
			case "previous":
				mono(c, FormatPrice(e.Previous), theme.RowSize, k.TextFaint)
			}
		},
		key:   func(row int) any { return row },
		row:   func(row int) string { return opts.Events[row].Title },
		state: opts.State, scroll: opts.Scroll, height: opts.Height, empty: empty,
	})
}

// impactWord is how much an event is expected to move a price, as the word a
// calendar would use for it rather than as a percentage: "High" is a sentence
// a reader can act on and "0.85" is a number they have to interpret.
func impactWord(impact float64) string {
	switch {
	case impact >= 0.8:
		return "High"
	case impact >= 0.4:
		return "Medium"
	}
	return "Low"
}

// beatInk is the tone a released actual is in against its forecast.
func beatInk(e Event, k theme.Tokens) ui.Color {
	switch {
	case e.Actual > e.Forecast:
		return k.Success
	case e.Actual < e.Forecast:
		return k.Danger
	}
	return k.Text
}

// impactInk is the tone an impact bar is in. Above 0.8 is the danger tone
// rather than the success one, because a scheduled event that will move a
// price by a fifth is not good news and should not be painted as though it
// were.
func impactInk(impact float64, k theme.Tokens) ui.Color {
	switch {
	case impact >= 0.8:
		return k.Danger
	case impact >= 0.4:
		return k.Warning
	}
	return k.Success
}

func clamp01(v float64) float32 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 1
	}
	return float32(v)
}

// EconomicCalendarOptions configure an EconomicCalendar.
type EconomicCalendarOptions struct {
	// Events are the caller's, in the order they should be shown. They are
	// [Event]s because the shape is the same — a date, a name, an expected
	// and an actual — and a second type for the same four fields would be a
	// second set of rules about how a forecast column is drawn.
	Events []Event
	// Height is the table's height, and is required.
	Height float32
	// State and Scroll keep the table's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Label names the table; empty takes the library's "Economic calendar".
	Label string
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// EconomicCalendar is the macro schedule: rate decisions, inflation prints,
// payrolls, in the order they land.
//
// It is the same table as the earnings calendar with a different column set,
// because that is what it is: a list of dated expectations with results
// against them. The impact column is gone — a CPI print is not "impactful" in
// the way a report is, it either moved the market or it did not, and the
// actual column says which.
func EconomicCalendar(c *ui.Context, opts EconomicCalendarOptions) *ui.Element {
	if opts.Height <= 0 {
		panic("finance: EconomicCalendar needs a Height; a table with no height grows to fit " +
			"every row rather than scrolling")
	}
	k := core.Tokens(c)
	name := opts.Label
	if name == "" {
		name = "Economic calendar"
	}
	empty := opts.Empty
	if empty == nil {
		empty = func() {
			ui.Text(c, "Nothing scheduled").TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.BodySize))
		}
	}

	econColumns := []tableColumn{
		{Title: "Date", ID: "when", Width: 128},
		{Title: "Event", ID: "title", Share: 1},
		{Title: "Forecast", ID: "forecast", Width: 116, Align: ui.End},
		{Title: "Actual", ID: "actual", Width: 116, Align: ui.End},
		{Title: "Previous", ID: "previous", Width: 116, Align: ui.End},
	}

	return tableOf(c, tableOptions{
		columns: econColumns,
		rows:    len(opts.Events),
		name:    name,
		cell: func(row, col int) {
			e := opts.Events[row]
			switch econColumns[col].ID {
			case "when":
				ui.Text(c, e.When).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
			case "title":
				ui.Text(c, e.Title).TextColor(k.Text).Grow(1).MaxLines(1).
					FontSize(core.FontSize(c, theme.RowSize))
			case "forecast":
				mono(c, FormatPrice(e.Forecast), theme.RowSize, k.TextMuted)
			case "actual":
				if !e.Released {
					ui.Text(c, "—").TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.RowSize))
					return
				}
				mono(c, FormatPrice(e.Actual), theme.RowSize, beatInk(e, k))
			case "previous":
				mono(c, FormatPrice(e.Previous), theme.RowSize, k.TextFaint)
			}
		},
		key:   func(row int) any { return row },
		row:   func(row int) string { return opts.Events[row].Title },
		state: opts.State, scroll: opts.Scroll, height: opts.Height, empty: empty,
	})
}

// ── the payoff diagram ─────────────────────────────────────────────────────

// PayoffOptions configure a PayoffDiagram.
type PayoffOptions struct {
	// Spot is where the underlying trades now, and Strike the strike. The
	// break-even point is computed from them rather than being given, because
	// a payoff diagram whose break-even is a separate number is a diagram
	// whose break-even can disagree with its own curve.
	Spot, Strike float64
	// Expiry is the option's own date, drawn in the header.
	Expiry string
	// Premium is what was paid or received, and it sets where the diagram's
	// zero line is. A long option's break-even is the strike plus the premium
	// and a short one's is the strike minus it, and that difference is the
	// whole of why the diagram exists.
	Premium float64
	// Put says it is a put. A call and a put have mirror-image curves and a
	// diagram that drew both as the same shape would be worse than no
	// diagram: the reader would take away exactly the wrong conclusion about
	// which way the position benefits from a move.
	Put bool
	// Long says the position is long. A short position's diagram is the long
	// one's upside down, and drawing it that way keeps the two comparable
	// rather than being a separate picture.
	Long bool
	// Steps is how many prices the curve is drawn at; zero takes 41, which is
	// odd so that the strike can land on one of them.
	Steps int
	// Width and Height are the chart's box.
	Width, Height float32
	// Hover is the caller's pointer, shared with the tooltip.
	Hover *chart.Hover
	// Grid draws the grid lines behind the data; it is on by default here
	// because a payoff curve is read against both axes and there is nothing
	// else to read it against.
	Grid bool
	// Label names the chart, and is required. The break-even is put into that
	// name rather than only into the line's painted text, because painted text
	// is invisible to a screen reader and the break-even is the one number
	// somebody reading this diagram out loud needs.
	Label string
}

// PayoffDiagram is what an option is worth at every price, and where it stops
// being worth having.
//
// It is [chart.LineChart] over a linear price axis with a reference line at
// the break-even, drawn by [chart.Annotation]. The curve is computed here
// because the *shape* is the option's business — the two straight lines that
// meet at the strike are the payoff, and the chart does not know what an
// option is — but the axes, the gutters and the label measuring are the
// chart's, and are not re-decided here.
func PayoffDiagram(c *ui.Context, opts PayoffOptions) *ui.Element {
	if opts.Label == "" {
		panic("finance: PayoffDiagram needs a Label; a curve with no axis on it says nothing " +
			"about what it is a curve of")
	}
	if !opts.Grid {
		opts.Grid = true
	}
	k := core.Tokens(c)
	steps := opts.Steps
	if steps <= 0 {
		steps = 41
	}
	spread := opts.Spot * 0.3
	if spread <= 0 {
		spread = 1
	}
	lo, hi := opts.Spot-spread, opts.Spot+spread

	xs := make([]float64, steps)
	ys := make([]float64, steps)
	for i := range steps {
		t := float64(i) / float64(steps-1)
		price := lo + (hi-lo)*t
		// The x values are prices, in the axis' own domain: a normalised
		// 0..1 here would land every point outside the price axis and draw
		// nothing at all.
		xs[i] = price
		ys[i] = payoff(price, opts.Strike, opts.Premium, opts.Put, opts.Long)
	}

	series := chart.SeriesFrom("Payoff", chart.FormLine, xs, ys)
	px := chart.NewLinear(chart.Domain{Min: lo, Max: hi}, 0, 1)
	py := chart.NewLinear(chart.FromZero(chart.Domain{Min: lowest(ys), Max: highest(ys)}), 0, 1)
	breakEven := opts.Strike
	if opts.Long {
		breakEven += opts.Premium
	} else {
		breakEven -= opts.Premium
	}

	// The frame is built here rather than through chart.LineChart so that the
	// break-even reference line can be drawn inside the same plot area. The
	// scales, the gutters and the label measuring are still ui/chart's — the
	// frame is the thing that measures them — so nothing about the axis is
	// re-decided; what is added is one line across the plot that says where
	// the position stops being worth having.
	frame := chart.Frame(c, chart.FrameOptions{
		Height: opts.Height, Border: true,
		// The break-even is put into the chart's name rather than only into
		// the line's painted text: painted text is invisible to a screen reader,
		// and the break-even is the one number somebody reading this diagram
		// out loud needs.
		Label: opts.Label + ", break-even " + FormatPrice(breakEven),
		Hover: opts.Hover,
		X: chart.AxisOptions{Side: chart.Bottom, Scale: px, Count: 6,
			Label: "Underlying at expiry"},
		Y: chart.AxisOptions{Side: chart.Left, Scale: py, Count: 5,
			Label: "Profit or loss"},
		Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, []chart.Series{series})},
	}, func(f chart.FrameResult) {
		chart.Grid(c, chart.GridOptions{Frame: f, X: chart.AxisOptions{Scale: px}})
		// As in PerformanceChart: the frame's own scales are only placed
		// while it paints, so the ones built here are the ones to hand over.
		chart.Plot(c, chart.PlotOptions{
			Frame: f, Series: series, X: px, Y: py, Width: 2.5,
		})
		chart.Annotation(c, chart.AnnotationOptions{
			Frame: f, X: px, Y: py,
			Line: &chart.Line{
				Value: breakEven, Dashed: true, Color: k.TextMuted,
				Text: "Break-even " + FormatPrice(breakEven),
			},
		})
		chart.Axis(c, f, chart.AxisOptions{Side: chart.Bottom, Scale: px, Count: 6,
			Label: "Underlying at expiry"})
		chart.Axis(c, f, chart.AxisOptions{Side: chart.Left, Scale: py, Count: 5,
			Label: "Profit or loss"})
	})
	e := frame.Element
	if opts.Width > 0 {
		e.Width(opts.Width)
	}
	return e
}

// payoff is what an option is worth at one price, in money.
//
// It is the definition and not an approximation: at expiry a call is worth
// max(price − strike, 0) and a put max(strike − price, 0), and the premium is
// what was paid for it. Long and short differ by a sign on the whole thing,
// which is why one function serves both — a short call is a long call with
// everything negated, and that is exactly true.
//
// Both the call branch and the put branch can return zero, and they have to.
// A long call below the strike is worth nothing at expiry, and a function that
// computed max(strike − price, 0) for it — as a naive version of this does —
// would draw a long call that makes money as the price falls, which is a put.
func payoff(price, strike, premium float64, put, long bool) float64 {
	intrinsic := 0.0
	if put {
		if strike > price {
			intrinsic = strike - price
		}
	} else if price > strike {
		intrinsic = price - strike
	}
	v := intrinsic - premium
	if long {
		return v
	}
	return -v
}

// lowest and highest are a series' two ends, written out because the two are
// one piece of arithmetic and a chart's y domain needs them together.
func lowest(vs []float64) float64 {
	out := vs[0]
	for _, v := range vs[1:] {
		if v < out {
			out = v
		}
	}
	return out
}

func highest(vs []float64) float64 {
	out := vs[0]
	for _, v := range vs[1:] {
		if v > out {
			out = v
		}
	}
	return out
}

// ── one table builder ──────────────────────────────────────────────────────

// tableColumn is [data.Column] under a name this package can use. data.Column
// identifies itself with a key() that is deliberately unexported, so a column
// built here carries its own ID and the cell builder switches on that. It is
// the same struct — an alias, not a copy — so a field added to data.Column
// cannot be silently absent from every table in this package.
type tableColumn = data.Column

// tableOptions is what every table in this file needs, gathered so that each
// component's own function is about its own columns and nothing else.
type tableOptions struct {
	columns []tableColumn
	rows    int
	name    string
	// cell builds one cell, given a row and a column of the columns above.
	cell func(row, col int)
	// key identifies a row's record so its place and its choice follow the
	// record rather than its number when the rows are reordered.
	key func(row int) any
	// row names a row for assistive technology.
	row   func(row int) string
	state *ui.ListState
	// scroll is where the table is scrolled sideways, for a table wider than
	// the panel it is in.
	scroll *ui.ScrollState
	height float32
	empty  func()
}

// tableOf is [data.DataTable] with the fields this package's tables all fill
// in, so that the six components above are six sets of columns and cells
// rather than six copies of the same option struct.
func tableOf(c *ui.Context, opts tableOptions) *ui.Element {
	return ui.Box(c).FillWidth().Label(opts.name).Role(ui.RoleTable).Children(func() {
		data.DataTable(c, data.DataTableOptions{
			Columns: opts.columns,
			Rows:    opts.rows,
			Cell:    opts.cell,
			Key:     opts.key,
			Label:   opts.row,
			State:   opts.state,
			Scroll:  opts.scroll,
			Height:  opts.height,
			Empty:   opts.empty,
		}).Element.FillWidth()
	})
}
