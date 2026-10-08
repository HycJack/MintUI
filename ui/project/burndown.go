package project

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
)

// BurndownDay is one day's reading of a sprint: how much was left and how many
// points it was to begin with.
type BurndownDay struct {
	// Day is where in the sprint this was, counting from 0. It is an index
	// rather than a date because the chart's x axis is "days of the sprint"
	// and a calendar axis would put a gap in it at every weekend.
	Day int
	// Remaining is how many points were still open at the end of the day.
	// It may exceed Scope on a day the scope grew, which is the normal case
	// for a sprint that took on more work.
	Remaining int
}

// Burndown is a sprint's shape: what was in it, and what was left on each day.
type Burndown struct {
	// Name is what the chart is called at its head.
	Name string
	// Scope is how many points the sprint began with.
	Scope int
	// Days are the readings, in order. Fewer days than the sprint has is
	// fine — the chart ends where the readings stop.
	Days []BurndownDay
	// Today is the day the sprint is on now, negative before it started. It
	// is the caller's because this package has no clock, and a burndown
	// whose "today" moved with the wall clock would be a different chart in
	// every screenshot.
	Today int
	// Unit is what the numbers count — "points", "tickets", "days". It is
	// the caller's word because every tracker counts something different.
	Unit string
}

// IdealBurndown is the line a sprint would trace if everything went exactly
// to plan: straight from Scope down to zero, one day's worth off per day.
//
// It is a function and not a series drawn by hand because it is the same
// shape whatever the data says, and a caller who draws it themselves will
// draw it a different way each time. The last point is exactly zero rather
// than "nearly zero": a plan that lands at one point left is a plan that did
// not land.
func IdealBurndown(days, scope int) []float64 {
	if days < 0 {
		return nil
	}
	out := make([]float64, days+1)
	if days == 0 {
		out[0] = float64(scope)
		return out
	}
	step := float64(scope) / float64(days)
	for i := range out {
		v := float64(scope) - step*float64(i)
		if i == days {
			v = 0
		}
		if v < 0 {
			v = 0
		}
		out[i] = v
	}
	return out
}

// ActualBurndown is the line the sprint actually traced, over a band axis
// from 0 to the last day there is a reading for.
//
// A missing day is left out rather than drawn as zero. A gap in a burndown is
// a weekend or a day nobody recorded, and drawing it as zero would say the
// sprint finished on the Friday — which is the one thing a burndown chart
// exists not to claim.
func ActualBurndown(days []BurndownDay, lastDay int) []float64 {
	if lastDay < 0 {
		return nil
	}
	out := make([]float64, 0, len(days))
	for _, d := range days {
		if d.Day < 0 || d.Day > lastDay {
			continue
		}
		out = append(out, float64(d.Remaining))
	}
	return out
}

// SprintLength is how many days the chart covers: the last reading there is,
// or Today, whichever is further along.
//
// Taking the later of the two is what makes the chart finish where the sprint
// does rather than where its last recorded reading does. A sprint on day six
// with readings only to day four has two days of flat line ahead of it, and
// drawing that gap as a straight continuation would be a claim about work
// nobody has measured.
func SprintLength(b Burndown) int {
	last := b.Today
	for _, d := range b.Days {
		if d.Day > last {
			last = d.Day
		}
	}
	if last < 0 {
		return 0
	}
	return last
}

// BurndownChartOptions configure a BurndownChart.
type BurndownChartOptions struct {
	// Burndown is the sprint being drawn.
	Burndown Burndown
	// Hover is the caller's pointer into the chart, passed through to the
	// frame so that a tooltip follows it. Nil draws the chart with no
	// pointer, which is what a chart inside a screenshot wants.
	Hover *chart.Hover
	// Height is the chart's height; zero lets the layout give it one.
	Height float32
	// HideIdeal drops the straight line of the plan.
	//
	// It is phrased as hiding rather than showing because the plan line is
	// the default and a burndown without it is a line going down, which is
	// the one thing a reader can already work out. A field called ShowIdeal
	// would be false until somebody remembered to set it, and every chart
	// would quietly lose the half of the comparison the chart exists for.
	HideIdeal bool
}

// BurndownChartResult carries a BurndownChart.
type BurndownChartResult struct {
	// Element is the whole chart.
	Element *ui.Element
}

// BurndownChart is a sprint's remaining work against the plan.
//
// It is ui/chart's LineChart and nothing of its own. That is the whole point:
// the frame in ui/chart is the thing that measures the gutters for the axis
// labels, thins those labels so they do not collide, clips a mark that runs
// past the plot instead of letting it draw over the axis, and shows an empty
// state when there is nothing to plot. A burndown chart that drew its own
// axes would be a second answer to every one of those, and the second one
// would be the worse one.
//
// Two series: the sprint's own line and the straight line of the plan. They
// are given the palette's colours in that order, so the plan is the pale one
// and the real one is the accent — which is the arrangement that makes the
// gap between them the first thing the eye finds.
func BurndownChart(c *ui.Context, opts BurndownChartOptions) BurndownChartResult {
	b := opts.Burndown
	if b.Name == "" {
		panic("project: BurndownChart needs a Name; a chart's numbers are drawn " +
			"rather than laid out, so the name is all a screen reader has")
	}
	unit := b.Unit
	if unit == "" {
		unit = "points"
	}

	length := SprintLength(b)
	actual := ActualBurndown(b.Days, length)

	// Nothing recorded and the sprint has not started is the one case with
	// nothing at all to draw. The rule is deliberately narrow: a sprint that
	// HAS started with no readings still shows the plan line, because "you
	// have done nothing and here is the shape you were supposed to have" is
	// the most useful thing a burndown can say at the start of a day.
	if len(b.Days) == 0 && b.Today < 0 {
		return BurndownChartResult{Element: emptyBurndown(c, opts)}
	}

	series := make([]chart.Series, 0, 2)
	if !opts.HideIdeal {
		series = append(series,
			chart.NewSeries("Plan", chart.FormLine,
				pointsFrom(IdealBurndown(length, b.Scope))...))
	}
	if len(actual) > 0 {
		series = append(series,
			chart.NewSeries(b.Name, chart.FormLine, pointsFrom(actual)...))
	}

	if len(series) == 0 {
		return BurndownChartResult{Element: emptyBurndown(c, opts)}
	}

	return BurndownChartResult{Element: chart.LineChart(c, chart.LineOptions{
		ChartOptions: chart.ChartOptions{
			Label:   b.Name + " burndown",
			Height:  opts.Height,
			Surface: true,
			Border:  true,
			Grid:    true,
			Hover:   opts.Hover,
			X:       chart.AxisOptions{Label: "Day", Ticks: true},
			// Side, because a Side of zero is Bottom and a y axis with no
			// side said for it lands on the bottom edge, on top of the day
			// numbers it is supposed to be read against.
			Y:      chart.AxisOptions{Side: chart.Left, Label: unit, Count: 6, Ticks: true},
			Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
		},
		Series: series,
		// Dots: a sprint is a dozen readings, not a thousand, and the dots are
		// what let somebody read the number off a flat part of the line.
		Dots: true,
	})}
}

// emptyBurndown is the chart with nothing to draw. It is LineChart with no
// series at all, so what a caller sees is ui/chart's own empty state — which
// says what is missing rather than drawing an axis over a range nothing
// reaches — rather than an empty panel this package drew for itself.
func emptyBurndown(c *ui.Context, opts BurndownChartOptions) *ui.Element {
	b := opts.Burndown
	unit := b.Unit
	if unit == "" {
		unit = "points"
	}
	return chart.LineChart(c, chart.LineOptions{
		ChartOptions: chart.ChartOptions{
			Label:      b.Name,
			Height:     opts.Height,
			Surface:    true,
			Border:     true,
			Grid:       true,
			EmptyTitle: core.Msg(c, "project.noReadings", "No readings yet"),
			EmptyBody:  core.Msg(c, "project.noReadingsBody", "Nothing has been recorded for this sprint."),
			X:          chart.AxisOptions{Label: "Day"},
			Y:          chart.AxisOptions{Side: chart.Left, Label: unit},
		},
	})
}

// pointsFrom is a run of values as the chart's own points, indexed from zero.
// The index is the day: a band axis would say the same thing and read worse
// for a chart whose x values are always 0, 1, 2.
func pointsFrom(values []float64) []chart.Point {
	out := make([]chart.Point, len(values))
	for i, v := range values {
		out[i] = chart.Point{X: float64(i), Y: v}
	}
	return out
}
