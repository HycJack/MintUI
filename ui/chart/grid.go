package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// GridOptions configure a [Grid].
type GridOptions struct {
	// Frame is the chart's frame: the lines are drawn across its plot area,
	// and take their places from its axes.
	Frame FrameResult
	// X and Y are the two axes whose ticks become lines. An axis with no scale
	// draws no lines, which is how a chart says "there are no grid lines on
	// this side" without needing a flag.
	X, Y AxisOptions
	// Color is what the lines are drawn in; the zero colour uses the tokens'
	// border, which is the one hairline colour the whole interface shares.
	Color ui.Color
}

// Grid draws the lines behind the data: a vertical one at each tick of the x
// axis and a horizontal one at each tick of the y axis.
//
// It draws the ticks that were drawn — the ones whose labels fitted. A line
// whose label was dropped would be a line with no number on it, and a chart
// full of those reads as noise rather than as a scale.
func Grid(c *ui.Context, opts GridOptions) *ui.Element {
	k := core.Tokens(c)
	color := opts.Color
	if color == (ui.Color{}) {
		color = k.Border
	}
	e := ui.Box(c).Absolute().Fill().Label("Grid")
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := opts.Frame.Plot()
		if plot.W <= 0 || plot.H <= 0 {
			return
		}
		p.Clip(plot, 0, func() {
			lines(c, p, plot, opts.X, color)
			lines(c, p, plot, opts.Y, color)
		})
	})
}

// lines draws one line across the plot at each of the axis' drawn ticks: along
// the axis for a left or right one, across it for a bottom or top one.
func lines(c *ui.Context, p *ui.Painter, plot ui.Rect, axis AxisOptions, color ui.Color) {
	if !axis.Scale.OK() {
		return
	}
	along := axis.Side == Left || axis.Side == Right
	for _, t := range Ticks(c, place(axis.Scale, plot, axis.Side),
		TickOptions{Count: axis.Count, Format: axis.Format, Size: tickSize(axis), Vertical: along}) {
		if !t.Drawn {
			continue
		}
		if along {
			p.Line(plot.X, t.Pos, plot.X+plot.W, t.Pos, theme.BorderWidth, color)
		} else {
			p.Line(t.Pos, plot.Y, t.Pos, plot.Y+plot.H, theme.BorderWidth, color)
		}
	}
}
