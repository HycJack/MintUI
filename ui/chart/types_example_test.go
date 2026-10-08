package chart_test

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
)

// Example_types shows what a chart type is: data in, an element out, and
// nothing else for the caller to hold. Every type below brings its own data
// and leaves the axes to be built from it, so the page that puts these on
// screen writes six numbers rather than a frame, two scales and a legend.
//
// The options are still there for a caller who wants them — a fixed axis to
// line two charts up, a grid, a legend, the pointer for a crosshair — but
// nothing below asks for one, and every chart still gets measured gutters and
// axis labels that do not collide.
func Example_types() {
	days := []float64{0, 1, 2, 3, 4, 5, 6}
	open := []float64{4, 7, 5, 9, 8, 12, 11}
	resolved := []float64{2, 4, 3, 6, 5, 7, 6}
	latency := []float64{120, 180, 90, 240, 410, 150, 95}

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})

		ui.Column(c).Gap(8).Children(func() {
			// Two series over a week, with the axes worked out from the
			// data and a legend named from the series themselves.
			chart.LineChart(c, chart.LineOptions{
				ChartOptions: chart.ChartOptions{Height: 200, Grid: true, Label: "A week of callbacks"},
				Series: []chart.Series{
					chart.SeriesFrom("Opened", chart.FormLine, days, open),
					chart.SeriesFrom("Resolved", chart.FormArea, days, resolved),
				},
				Dots: true,
			})

			// A bar chart of the same week: the baseline is zero, which is
			// why its y axis starts there whatever the values are.
			chart.BarChart(c, chart.BarOptions{
				ChartOptions: chart.ChartOptions{Height: 180, Label: "Opened by day"},
				Series:       []chart.Series{chart.SeriesFrom("Opened", chart.FormBar, days, open)},
			})

			// A share of a whole, sorted into a Pareto, and a distribution
			// of latencies: three different questions, three different
			// shapes, one set of rules underneath them.
			ui.Row(c).Gap(8).Children(func() {
				chart.PieChart(c, chart.PieOptions{
					ChartOptions: chart.ChartOptions{Height: 200},
					Labels:       []string{"Engine", "Routing", "Auth", "Docs", "Other"},
					Values:       []float64{38, 24, 17, 12, 9},
				})
				chart.Histogram(c, chart.HistogramOptions{
					ChartOptions: chart.ChartOptions{Height: 200,
						Y: chart.AxisOptions{Side: chart.Left, Label: "Callbacks"}},
					Values: latency,
					Bins:   6,
				})
			})
		})
	}, 560, 720)

	// Output:
}

// Example_types_marks shows the one thing every chart type is made of: the
// numbers, put through a pure function, before anything is drawn. A caller
// can use these to size something else around a chart — a summary line
// beside it, a badge on it — and be sure it agrees with what the chart draws.
func Example_types_marks() {
	// Every column of a percent stack comes to a hundred, exactly.
	percent := chart.Stack([][]float64{
		{4, 2, 6}, // opened
		{1, 3, 4}, // re-opened
	})
	// A Pareto's line ends at a hundred, and a pie's slices come to a
	// whole turn.
	share := chart.Cumulative([]float64{50, 30, 20})
	sweep := chart.Sweep([]float64{1, 1, 2})

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		for _, row := range percent {
			fmt.Println(row)
		}
		fmt.Println(share, sweep)
	}, 320, 40)

	// Output:
	// [80 40 60]
	// [20 60 40]
	// [50 80 100] [90 90 180]
}
