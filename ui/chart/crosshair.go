package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// CrosshairOptions configure a [Crosshair].
type CrosshairOptions struct {
	// Frame is the chart's frame, whose plot area the rules are drawn in.
	Frame FrameResult
	// Hover is the chart's pointer. A nil hover, or one that is not over the
	// plot, draws nothing: a crosshair with no pointer to follow is two
	// lines through the middle of somebody's data.
	Hover *Hover
	// X and Y are the axes' scales, used to label the rules with values.
	X, Y Scale
	// Color is what the rules are drawn in; the zero colour uses the muted
	// text, which is faint enough to read data through.
	Color ui.Color
	// ValueX and ValueY, when set, are written on cards at the ends of the
	// rules: the x value under the vertical one, the y value beside the
	// horizontal one. Empty strings write nothing.
	ValueX, ValueY string
}

// Crosshair is the pair of rules that follow the pointer: one down the plot at
// its x, one across it at its y, so a reader can take a value off either axis
// without moving anything.
//
// It follows the caller's [Hover], which the [Frame] fills, rather than
// reading the pointer itself — so a chart's crosshair, tooltip and brush can
// never disagree about where the pointer is.
func Crosshair(c *ui.Context, opts CrosshairOptions) *ui.Element {
	k, step := core.Tokens(c), core.Density(c).Unit()
	color := opts.Color
	if color == (ui.Color{}) {
		color = k.TextMuted
	}
	e := ui.Box(c).Absolute().Fill().Label("Crosshair")
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		h := opts.Hover
		plot, bounds := opts.Frame.Plot(), opts.Frame.Bounds()
		if h == nil || !h.Over || plot.W <= 0 || plot.H <= 0 {
			return
		}
		// The rules are the chart's own, and stop at the plot's edge. The
		// ring at their crossing is where the pointer is, which a rule on
		// its own cannot say — a pair of lines has no middle to it.
		p.Clip(plot, 0, func() {
			p.Line(h.X, plot.Y, h.X, plot.Y+plot.H, theme.BorderWidth, color.Alpha(0.55))
			internal.Rule(p, ui.Rect{X: plot.X, Y: h.Y, W: plot.W, H: 1}, color.Alpha(0.55))
			internal.Ring(p, h.X, h.Y, step*0.75, theme.BorderWidth*1.5, color)
		})
		if opts.ValueX == "" && opts.ValueY == "" {
			return
		}
		// The values are read off the rules, so they belong where the axis'
		// own labels are — in the gutters, taking the numbers' place rather
		// than covering the data. A chart too narrow for them puts them over
		// the plot's edge instead, which is better than not showing them.
		size := theme.CaptionSize
		card := func(x, y float32, label string) {
			w, hgt := LabelWidth(c, label, size), LabelHeight(c, size)+step/2
			// A chart too small for the cards puts them over the plot's edge
			// rather than off the chart; they still stay inside the frame.
			x = maxf(bounds.X, minf(x, bounds.X+bounds.W-w-step))
			y = maxf(bounds.Y, minf(y, bounds.Y+bounds.H-hgt))
			p.Fill(ui.Rect{X: x, Y: y, W: w + step, H: hgt}, k.Fill, theme.SmallRadius)
			p.Text(x+step/2, y+step/4, label, size, k.OnFill)
		}
		p.Clip(bounds, 0, func() {
			if opts.ValueX != "" {
				w := LabelWidth(c, opts.ValueX, size)
				card(h.X-w/2, plot.Y+plot.H+step/2, opts.ValueX)
			}
			if opts.ValueY != "" {
				w := LabelWidth(c, opts.ValueY, size)
				card(plot.X-w-step, h.Y-LabelHeight(c, size)/2-step/4, opts.ValueY)
			}
		})
	})
}
