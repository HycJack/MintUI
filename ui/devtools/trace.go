package devtools

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// TraceWaterfallOptions configure a TraceWaterfall.
type TraceWaterfallOptions struct {
	// Spans are the trace, in the order the tracing system gave them: parent
	// before child, so that a row is drawn above the ones it contains. The
	// waterfall does not sort them, because that order is what makes the
	// nesting readable and a caller with a different one knows why.
	Spans []Span
	// Selected is the span the caller is looking at, empty for none, and
	// SelectedID is written when one is pressed. Both are the caller's: the
	// selection is a place in the trace somebody can come back to, and the
	// press is an event the caller acts on.
	Selected   *string
	SelectedID *string
	// Sort puts the slowest spans first instead of the caller's order. It is
	// off by default: a waterfall sorted by duration loses the nesting, and
	// the nesting is most of what a waterfall is for.
	Sort bool
	// ShowDepth indents each row by how deep its span sits.
	ShowDepth bool
	// Height is the viewport's height.
	Height float32
}

// TraceWaterfallResult carries a TraceWaterfall and what was pressed in it.
type TraceWaterfallResult struct {
	// Element is the whole waterfall.
	Element *ui.Element
}

// TraceWaterfall is a trace: one row per span, each bar placed where the span
// actually was inside the whole.
//
// Every bar's position is SpanBar's, which means the widths of a trace's bars
// are percentages of one number — the end of its last span — and a span
// drawn at 40% really did take 40% of the time. That is the whole claim of
// the view, and it is arithmetic rather than layout: nothing here measures
// anything, so a waterfall of a hundred spans costs the same as one of ten.
//
// The bars are drawn rather than laid out because a bar's position is a
// fraction of a total and not a length; a row of divs would have to be given
// a width that adds up to the whole row, and at six decimal places that
// rounding is visible.
func TraceWaterfall(c *ui.Context, opts TraceWaterfallOptions) TraceWaterfallResult {
	if opts.Selected == nil {
		panic("devtools: TraceWaterfall needs the *string Selected writes to")
	}
	if opts.SelectedID == nil {
		panic("devtools: TraceWaterfall needs the *string SelectedID writes to; " +
			"a span somebody pressed has to be nameable afterwards")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	spans := opts.Spans
	if opts.Sort {
		spans = SlowestSpans(spans, 0)
	}
	total := TotalOf(spans)

	var r TraceWaterfallResult
	r.Element = ui.Column(c).FillWidth().Gap(u).
		Label(core.Msg(c, "devtools.trace", "Trace")).Children(func() {
		ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
			ui.Text(c, core.Msg(c, "devtools.total", "Total")).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			ui.Text(c, RoundDuration(total)).TextColor(k.Text).Grow(1).Font(monoFamily).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			ui.Text(c, itoa(len(spans))+" "+core.Msg(c, "devtools.spans", "spans")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
		if len(spans) == 0 {
			ui.Text(c, core.Msg(c, "devtools.noTrace", "No spans")).
				TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			return
		}
		for i, s := range spans {
			spanRow(c, opts, spans, i, s, total)
		}
	})
	return r
}

// spanRow is one span: its name, how deep it is, and the bar.
func spanRow(c *ui.Context, opts TraceWaterfallOptions, spans []Span, at int, s Span, total time.Duration) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	left, width := SpanBar(s, total)
	depth := 0
	if opts.ShowDepth {
		depth = DepthOf(spans, at)
	}
	on := s.Name == *opts.Selected

	row := ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
		Padding(u*0.5, u).Radius(theme.SmallRadius).
		Cursor(ui.CursorPointer).Label(s.Name).Role(ui.RoleListItem)
	if on {
		row.Background(k.SurfaceHover)
	}
	if row.Clicked() {
		*opts.SelectedID = s.Name
	}
	row.Children(func() {
		ui.Box(c).Width(u * float32(depth) * 2).Shrink(0)
		name := s.Name
		if s.Error {
			name = "✕ " + name
		}
		ink := k.TextMuted
		switch {
		case s.Error:
			ink = k.Danger
		case on:
			ink = k.Text
		}
		ui.Text(c, name).TextColor(ink).Shrink(0).Font(monoFamily).
			Width(spanNameWidth).FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		// The track and the bar inside it. The track is the surface so that
		// a span taking a tenth of the trace is visibly a tenth of a line
		// rather than a tenth of nothing.
		ui.Box(c).Grow(1).Height(barHeight).Shrink(0).Radius(barHeight / 2).
			Background(k.Surface).Draw(func(p *ui.Painter, rect ui.Rect) {
			col := k.Accent
			if s.Error {
				col = k.Danger
			}
			if width <= 0 {
				return
			}
			p.Fill(ui.Rect{
				X: rect.X + rect.W*left/100,
				Y: rect.Y,
				W: rect.W * width / 100,
				H: rect.H,
			}, col, barHeight/2)
		})
		ui.Text(c, RoundDuration(s.Duration)).TextColor(k.TextFaint).Shrink(0).
			Font(monoFamily).Width(durationWidth).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// The three fixed measures of a waterfall row: the name's column, the bar's
// height and the duration's column. Numbers rather than expressions because
// they are what a reader compares between two traces, and three sites each
// spelling out their own idea of a width is three widths.
const (
	spanNameWidth float32 = 180
	barHeight     float32 = 10
	durationWidth float32 = 56
)

// Metric is one thing being watched.
type Metric struct {
	// Name is what it is called.
	Name string
	// Value is where it is now and Limit where it should stop being, both
	// already in the unit the caller wants drawn. A Limit of zero means
	// nothing: a queue depth or a goroutine count has no ceiling worth
	// drawing, and showing one would invent a target.
	Value, Limit float64
	// Unit is what the numbers count — "MB", "req/s", "" for a bare number.
	Unit string
	// HigherIsWorse is true for latency and queue depth and false for a
	// cache hit rate, which is the one fact a generic gauge cannot know.
	HigherIsWorse bool
}

// ResourceGaugeOptions configure a ResourceGauge.
type ResourceGaugeOptions struct {
	// Metric is what the gauge shows.
	Metric Metric
	// Width and Height bound the gauge; zero lets the layout give them.
	Width, Height float32
}

// ResourceGaugeResult carries a ResourceGauge.
type ResourceGaugeResult struct {
	// Element is the gauge.
	Element *ui.Element
}

// ResourceGauge is one metric against its limit: a bar, its number and how
// close it is to the edge.
//
// The fraction is MetricFraction's rather than arithmetic at each site,
// because "how close to the edge" is a question with a trap in it: a metric
// with no limit has no edge, and dividing by nothing is how a dashboard
// prints NaN next to a number somebody has to trust.
func ResourceGauge(c *ui.Context, opts ResourceGaugeOptions) ResourceGaugeResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	m := opts.Metric
	if m.Name == "" {
		panic("devtools: ResourceGauge needs a metric with a Name; a bar with " +
			"no name reaches a screen reader as \"progress bar\"")
	}

	frac := MetricFraction(m)
	sev := MetricTone(m)

	var r ResourceGaugeResult
	gauge := ui.Column(c).FillWidth().Gap(u * 0.5).Label(m.Name).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, m.Name).TextColor(k.Text).Grow(1).
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			ui.Text(c, MetricFigure(m)).TextColor(k.TextMuted).Shrink(0).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		})
		ui.Box(c).FillWidth().Height(gaugeHeight).Shrink(0).Radius(gaugeHeight / 2).
			Background(k.Surface).Draw(func(p *ui.Painter, rect ui.Rect) {
			_, ink := sev.Pair(k)
			p.Fill(ui.Rect{X: rect.X, Y: rect.Y, W: rect.W * frac, H: rect.H}, ink, gaugeHeight/2)
		})
	})
	if opts.Height > 0 {
		gauge.Height(opts.Height)
	}
	r.Element = gauge
	return r
}

// gaugeHeight is a gauge's bar. Thinner than a progress bar's, because a
// dashboard's gauges sit several to a row and the point of them is to be
// compared rather than read.
const gaugeHeight float32 = 8

// MetricFraction is how much of a metric's limit is used, 0 to 1.
//
// A metric with no limit is 0: the gauge draws as empty rather than as full,
// because a bar showing "everything" against no target is a claim the
// metric's owner never made.
func MetricFraction(m Metric) float32 {
	if m.Limit <= 0 || m.Value <= 0 {
		return 0
	}
	f := float32(m.Value / m.Limit)
	if f > 1 {
		return 1
	}
	return f
}

// MetricTone is the severity a metric is drawn at.
//
// It goes by the metric's own direction. A latency at eighty per cent of its
// limit is a warning and a cache hit rate at eighty per cent is not the same
// kind of fact, and a component that assumed both meant the same thing would
// be wrong about one of them in whichever direction it did not.
func MetricTone(m Metric) core.Severity {
	if m.Limit <= 0 {
		return core.Neutral
	}
	// Compared as hundredths of the limit rather than as a fraction.
	// float32(0.9) is a hair below 0.9, so a metric sitting exactly on the
	// warning step would draw as the step below it — and a metric at exactly
	// 90% is precisely the one somebody wants noticed.
	switch {
	case m.Value >= m.Limit:
		return core.Danger
	case m.Value*100 >= m.Limit*90:
		return core.Warning
	case m.Value*100 >= m.Limit*75:
		return core.Accent
	default:
		return core.Neutral
	}
}

// MetricFigure is a metric's number as a person reads it, with its unit.
func MetricFigure(m Metric) string {
	v := trimFloat(m.Value)
	switch {
	case m.Limit <= 0:
		return v + m.Unit
	case m.HigherIsWorse:
		return v + m.Unit + " / " + trimFloat(m.Limit) + m.Unit
	default:
		// Lower is worse, so the figure is what is left rather than what is
		// used: a cache hit rate at 20% out of 100% reads as "20 out of
		// 100" from the bar, and "20% used" from the words would not.
		return trimFloat(m.Limit-m.Value) + m.Unit + " left"
	}
}

// SystemMonitorOptions configure a SystemMonitor.
type SystemMonitorOptions struct {
	// Metrics are what is being watched, in the caller's order.
	Metrics []Metric
	// Refreshed is when the numbers were collected, already formatted.
	Refreshed string
	// Empty draws instead of the gauges when there are none.
	Empty func()
}

// SystemMonitorResult carries a SystemMonitor.
type SystemMonitorResult struct {
	// Element is the whole monitor.
	Element *ui.Element
}

// SystemMonitor is a set of gauges: what this machine or this service is
// doing right now.
//
// It draws the caller's numbers and nothing else. There is no sampling, no
// clock and no history here, because a widget that owns a clock cannot be
// asked "what were these numbers at 09:14" — and the moment somebody asks
// that, the answer has to come from the data, not from a label that moved.
// Refreshing is the caller's business and Refreshed is where it says so.
func SystemMonitor(c *ui.Context, opts SystemMonitorOptions) SystemMonitorResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r SystemMonitorResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).
		Label(core.Msg(c, "devtools.monitor", "System")).Children(func() {
		if opts.Refreshed != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
				ui.Text(c, core.Msg(c, "devtools.refreshed", "Refreshed")).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				ui.Text(c, opts.Refreshed).TextColor(k.TextFaint).Grow(1).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			})
		}
		if len(opts.Metrics) == 0 {
			if opts.Empty != nil {
				opts.Empty()
			} else {
				ui.Text(c, core.Msg(c, "devtools.noMetrics", "Nothing to watch")).
					TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			}
			return
		}
		// Two to a row: a list of full-width gauges is a column of hairlines
		// that has to be scrolled past, and a monitor's whole value is
		// comparing one metric with the next.
		ui.Column(c).FillWidth().Gap(u * 1.5).Children(func() {
			for _, m := range opts.Metrics {
				ResourceGauge(c, ResourceGaugeOptions{Metric: m})
			}
		})
	})
	return r
}

// ServiceStatusOptions configure a ServiceStatus.
type ServiceStatusOptions struct {
	// Name is the service's name.
	Name string
	// State is one of the states below, as a string the caller supplies.
	State string
	// Detail is the line under the name: the version, the region, the
	// reason. The caller's sentence, because only they know what is useful
	// beside a name.
	Detail string
	// Uptime is the availability from UptimeSlots, 0 to 1. A negative value
	// means "not measured", which draws no bar at all rather than an empty
	// one — see UptimeSlots on why an unmeasured service is not a perfect
	// one.
	Uptime float32
	// Slots are the days behind UptimeSlots, drawn as the bar of squares.
	Slots []SlotState
}

// ServiceStatusResult carries a ServiceStatus.
type ServiceStatusResult struct {
	// Element is the whole row.
	Element *ui.Element
}

// The states a service can be in. They are strings because a status page has
// its own vocabulary — "degraded", "partial outage", "maintenance" — and
// mapping it here would put this package in the middle of words that are not
// its.
const (
	// StateOperational is the good one.
	StateOperational = "operational"
	// StateDegraded is up and slower than it should be.
	StateDegraded = "degraded"
	// StateDown is not answering.
	StateDown = "down"
	// StateMaintenance is out on purpose.
	StateMaintenance = "maintenance"
)

// ServiceStateTone is the severity a service's state is drawn at.
func ServiceStateTone(state string) core.Severity {
	switch strings.ToLower(state) {
	case StateOperational:
		return core.Success
	case StateDegraded:
		return core.Warning
	case StateDown:
		return core.Danger
	case StateMaintenance:
		return core.Neutral
	}
	return core.Neutral
}

// ServiceStatus is one service: its name, what state it is in, and the ninety
// days behind it.
func ServiceStatus(c *ui.Context, opts ServiceStatusOptions) ServiceStatusResult {
	if opts.Name == "" {
		panic("devtools: ServiceStatus needs a Name; a row with no service in " +
			"it says nothing on a page of nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	sev := ServiceStateTone(opts.State)

	var r ServiceStatusResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 1.5).
		Label(opts.Name).Role(ui.RoleGroup).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			// The dot and the word, both. A dot alone is a colour, and the
			// state is the one thing on a status page a reader most needs
			// told plainly — so the row says it in words as well.
			_, ink := sev.Pair(k)
			ui.Box(c).Size(u*2.5, u*2.5).Shrink(0).Radius(u * 1.25).
				Background(ink).Label(opts.Name + " is " + opts.State)
			ui.Text(c, opts.Name).TextColor(k.Text).Grow(1).Bold().
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			ui.Text(c, opts.State).TextColor(stateInk(k, sev)).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
		if opts.Detail != "" {
			ui.Text(c, opts.Detail).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
		if len(opts.Slots) > 0 {
			UptimeBar(c, UptimeBarOptions{Slots: opts.Slots, Label: opts.Name})
		} else if opts.Uptime >= 0 {
			ui.Text(c, "uptime "+percentText(opts.Uptime)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
	return r
}

// stateInk is a service's state in words' colour.
func stateInk(k theme.Tokens, sev core.Severity) ui.Color {
	_, fg := sev.Pair(k)
	return fg
}

// percentText is a fraction written as a percentage, with at least two
// decimals: 99.9% and 100% are the two numbers an uptime bar is read at, and
// one decimal turns 99.95 into 100.
func percentText(f float32) string {
	p := f * 100
	whole := int(p)
	frac := int((p-float32(whole))*100 + 0.5)
	return itoa(whole) + "." + pad2(frac) + "%"
}

// pad2 writes a number in two digits.
func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

// UptimeBarOptions configure an UptimeBar.
type UptimeBarOptions struct {
	// Slots are the days, oldest first, as UptimeSlots produced them. A
	// shorter run is drawn from the left, so a service installed last week
	// shows a week rather than a fortnight of blank squares at the end — the
	// squares that are there are the days that were measured.
	Slots []SlotState
	// Label names the bar. It is required: a bar of ninety coloured squares
	// with no name announces as nothing.
	Label string
	// HidePercent drops the availability figure beside the bar.
	//
	// It is phrased as hiding because the figure is the default and a bar
	// alone answers "were there bad days" rather than "how bad". A field
	// called ShowPercent would be false until somebody set it, and every bar
	// would quietly lose the number the squares were drawn to be read with.
	HidePercent bool
	// Square is how wide one day is; zero is the library's own.
	Square float32
}

// UptimeBarResult carries a UptimeBar.
type UptimeBarResult struct {
	// Element is the bar.
	Element *ui.Element
}

// UptimeBar is ninety days of a service's availability as ninety squares.
//
// The squares are drawn, not laid out: a row of ninety boxes would be ninety
// elements and a great deal of layout for something that is four rectangles
// and a clip, and at this size the difference is a bar that stays smooth
// while a list of boxes does not.
//
// The colour of a square is its SlotState's severity pair, so a degraded day
// and a down day are different colours and an unmeasured day is the surface
// rather than a fifth colour — which is the answer that makes "we have not
// been watching this long" visible instead of merely absent.
func UptimeBar(c *ui.Context, opts UptimeBarOptions) UptimeBarResult {
	if opts.Label == "" {
		panic("devtools: UptimeBar needs a Label; ninety squares with no name " +
			"announce as nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	w := opts.Square
	if w <= 0 {
		w = uptimeSquare
	}
	pct := UptimePercent(opts.Slots)

	var r UptimeBarResult
	bar := ui.Column(c).FillWidth().Gap(u * 0.5).Label(opts.Label).Children(func() {
		if !opts.HidePercent {
			// A gap, because the figure and the sentence beside it are two
			// things said and not one word run together.
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
				ui.Text(c, percentText(pct)).TextColor(inkForUptime(k, UptimeTone(pct))).Bold().
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				ui.Text(c, core.Msg(c, "devtools.over90", "over 90 days")).
					TextColor(k.TextFaint).Grow(1).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			})
		}
		ui.Box(c).FillWidth().Height(w).Shrink(0).Radius(theme.SmallRadius * 0.5).
			Background(k.Surface).Draw(func(p *ui.Painter, rect ui.Rect) {
			n := len(opts.Slots)
			if n == 0 || rect.W <= 0 {
				return
			}
			gap := w * 0.15
			cell := (rect.W - gap*float32(n-1)) / float32(n)
			if cell <= 0 {
				// More days than there is room for: they are squeezed to
				// nothing rather than drawn overlapping, so the bar stays the
				// width it was given at any window size.
				return
			}
			for i, s := range opts.Slots {
				col := k.Surface
				switch s {
				case SlotUp:
					col = k.Success
				case SlotPartial:
					col = k.Warning
				case SlotDown:
					col = k.Danger
				}
				p.Fill(ui.Rect{
					X: rect.X + float32(i)*(cell+gap),
					Y: rect.Y,
					W: cell,
					H: rect.H,
				}, col, w*0.25)
			}
		})
	})
	r.Element = bar
	return r
}

// uptimeSquare is how tall one day's square is. A wide bar of ninety small
// squares is the shape everybody recognises from a status page, and the
// height is what makes it read as a row of days rather than as a ribbon.
const uptimeSquare float32 = 14

// inkForUptime is the text colour for an availability figure.
func inkForUptime(k theme.Tokens, sev core.Severity) ui.Color {
	_, fg := sev.Pair(k)
	return fg
}
