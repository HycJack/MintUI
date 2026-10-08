// Package chart is the data-visualisation layer of the interface: the axes,
// the grid, the legend, the series and the pieces that follow the pointer.
//
// It is deliberately the bottom of a chart and not the charts themselves.
// There are no line charts here and no bar charts — there is a [Frame] that
// measures its gutters, a [Scale] that says where a value falls, a set of
// [Ticks] that say where the lines go, and a [Plot] that draws a [Series]
// wherever those say. A chart type is then a handful of lines over these, and
// every chart type gets the axis label that does not collide for free.
//
// # How a chart is put together
//
// The one thing worth knowing before writing a chart is that a frame measures
// its plot area as it paints, because a frame does not know how big it is
// until the layout has run. So the scales a caller hands in are written in
// their own space — 0 to 1 — and the frame places them in the plot as it draws:
//
//	var hover chart.Hover
//	series := []chart.Series{
//		chart.SeriesFrom("Open", chart.FormLine, days, open),
//		chart.SeriesFrom("Closed", chart.FormArea, days, closed),
//	}
//	d, _ := chart.SeriesDomain(series)
//	xs := chart.NewLinear(chart.NewBand(...), 0, 1)          // or a band scale
//	ys := chart.NewLinear(chart.Nice(d.Min, d.Max, 4), 0, 1)
//
//	chart.Frame(c, chart.FrameOptions{
//		Height: 220,
//		X:      chart.AxisOptions{Scale: xs, Count: 7, Label: "Day"},
//		Y:      chart.AxisOptions{Side: chart.Left, Scale: ys, Count: 5, Label: "Callbacks"},
//		Legend: chart.LegendOptions{Entries: chart.EntriesOf(c, series)},
//		Hover:  &hover,
//	}, func(frame chart.FrameResult) {
//		chart.Grid(c, chart.GridOptions{Frame: frame, X: …, Y: …})
//		for _, s := range series {
//			chart.Plot(c, chart.PlotOptions{Frame: frame, Series: s, X: xs, Y: ys})
//		}
//		chart.Crosshair(c, chart.CrosshairOptions{Frame: frame, Hover: &hover})
//	})
//
// Everything inside the frame reads its plot area from the frame result as it
// paints, which is why the pieces take a FrameResult rather than a rectangle:
// a parent's drawing runs before its children's, so by the time a series draws,
// the frame knows where it is going.
//
// # The rules of the package
//
// Four of them, all of them inherited rather than invented here.
//
// Components hold no state. A pointer's place is the caller's [Hover], a
// brush's selection is the caller's [Range], a pressed thing is a bool the
// caller owns. Nothing is kept between frames, so what a chart draws is always
// what it was told to draw.
//
// The data is the caller's. Nothing here generates a number, reads the clock
// or fills in a value: a chart with no data says so with [Empty] rather than
// drawing an axis over a range nothing reaches.
//
// Colours are tokens. The one exception is a chart's series, which take their
// colours from [Palette] — a ladder stepped out from the window's accent that
// reads on a white window and a near-black one alike, so a chart needs no
// second palette for the dark appearance.
//
// What cannot be drawn is a mistake, not a shrug: a scale without a domain, a
// band axis without categories, an annotation with nothing to say, a brush with
// nowhere to put what it catches — each panics with "chart: …" rather than
// drawing a frame that quietly looks wrong.
package chart
