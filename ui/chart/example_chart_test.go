package chart_test

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chart"
	"github.com/HycJack/MintUI/ui/core"
)

// Example shows a week of callbacks: two series, a bar axis of days, a
// nicest-ranged count axis, a grid, a legend, a target line, and a crosshair
// and tooltip following the pointer.
//
// The scales are written 0 to 1 — the frame measures its plot area as it
// paints, so nothing is placed in pixels before it is time — and every piece
// inside the frame reads that plot area from the frame result as it draws.
func Example() {
	days := []float64{0, 1, 2, 3, 4, 5, 6}
	open := []float64{4, 7, 5, 9, 8, 12, 11}
	closed := []float64{2, 4, 3, 6, 5, 7, 6}

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})

		series := []chart.Series{
			chart.SeriesFrom("Open", chart.FormLine, days, open),
			chart.SeriesFrom("Closed", chart.FormArea, days, closed),
		}
		// The y axis is nice-ranged before anything is drawn, so its ends are
		// numbers somebody would have chosen.
		d, _ := chart.SeriesDomain(series)
		ys := chart.NewLinear(chart.Nice(d.Min, d.Max, 4), 0, 1)
		xs := chart.NewLinear(chart.Domain{Min: 0, Max: 6}, 0, 1)

		xAxis := chart.AxisOptions{Scale: xs, Count: 7, Label: "Day"}
		yAxis := chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 5, Label: "Callbacks"}
		var hover chart.Hover
		var selection chart.Range

		chart.Frame(c, chart.FrameOptions{
			Height: 240,
			Label:  "Open callbacks by day",
			X:      xAxis,
			Y:      yAxis,
			Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
			Hover:  &hover,
			Border: true,
		}, func(frame chart.FrameResult) {
			chart.Grid(c, chart.GridOptions{Frame: frame, X: xAxis, Y: yAxis})

			// "Eight a day" is the target the week is read against.
			chart.Annotation(c, chart.AnnotationOptions{Frame: frame, X: xs, Y: ys,
				Line: &chart.Line{Value: 8, Text: "target", Dashed: true}})

			for _, s := range series {
				chart.Plot(c, chart.PlotOptions{Frame: frame, Series: s, X: xs, Y: ys, Dots: true})
			}

			// The pointer's place is the caller's, so the brush, the
			// crosshair and the tooltip cannot disagree about it.
			chart.Brush(c, chart.BrushOptions{Frame: frame, X: xs, Range: &selection, Handles: true})
			chart.Crosshair(c, chart.CrosshairOptions{Frame: frame, Hover: &hover})
			chart.Tooltip(c, chart.TooltipOptions{Frame: frame, Hover: &hover,
				Title: "Tuesday",
				Rows: []chart.TooltipRow{
					{Name: "Open", Color: chart.EntriesOf(c, series)[0].Color, Value: "9"},
					{Name: "Closed", Color: chart.EntriesOf(c, series)[1].Color, Value: "6"},
				}})

			chart.Axis(c, frame, xAxis)
			chart.Axis(c, frame, yAxis)
		})
	}, 560, 280)

	// Output:
}

// Example_empty shows what a chart says when it has nothing to plot: the
// empty state takes the plot area whole rather than drawing an axis over a
// range no data reaches.
func Example_empty() {
	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		chart.Frame(c, chart.FrameOptions{Height: 220, Surface: true, Border: true},
			func(frame chart.FrameResult) {
				chart.Empty(c, chart.EmptyOptions{
					Frame: frame,
					Title: "Nothing resolved yet",
					Body:  "No callbacks were resolved today. A bar chart of nothing is still a chart.",
				})
			})
	}, 460, 240)

	// Output:
}

// Example_ticks shows the maths on its own, away from a window: a range
// nicened to round numbers, and where each of its ticks lands in a plot 400
// wide.
//
// These are pure functions, so a caller can use them to size something else
// around a chart — a range slider under it, a summary line beside it — and be
// sure it agrees with what the chart draws.
func Example_ticks() {
	d := chart.Nice(0, 97, 5) // 0, 97 → 0, 100
	scale := chart.NewLinear(d, 0, 400)

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		for _, tick := range chart.Ticks(c, scale.Span(0, 400), chart.TickOptions{Count: 5}) {
			fmt.Printf("%6.0f %s\n", tick.Pos, tick.Label)
		}
	}, 400, 40)

	// Output:
	//      0 0
	//     80 20
	//    160 40
	//    240 60
	//    320 80
	//    400 100
}
