package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// EmptyOptions configure an [Empty].
type EmptyOptions struct {
	// Frame is the chart's frame: the empty state fills its plot area.
	Frame FrameResult
	// Title says what is missing.
	Title string
	// Body explains what would fill it, in one sentence. It is the caller's
	// copy, not the library's: "No callbacks resolved today" is theirs.
	Body string
	// Art draws the mark above the title; nil draws the default, which is a
	// few blocks of nothing and one bright dot — the shape of data that is
	// not there yet.
	Art func(p *ui.Painter, area ui.Rect)
}

// Empty is what a chart shows when it has nothing to plot.
//
// A chart with no data is the one case where the axes are a lie — they would
// draw a scale over a range nothing reaches — so the empty state takes the
// plot area whole and says what is missing instead.
func Empty(c *ui.Context, opts EmptyOptions) *ui.Element {
	k, step := core.Tokens(c), core.Density(c).Unit()
	label := opts.Title
	if opts.Body != "" {
		if label != "" {
			label += ". "
		}
		label += opts.Body
	}
	e := ui.Box(c).Absolute().Fill()
	if label != "" {
		e = e.Label(label)
	}
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := opts.Frame.Plot()
		if plot.W <= 0 || plot.H <= 0 {
			return
		}
		pad := step * 4
		width := minf(plot.W-pad*2, 320)
		if width <= 0 {
			return
		}
		x := plot.X + (plot.W-width)/2

		// The text is measured before anything is placed, because everything
		// else is placed relative to it: the mark above the title, and the
		// whole thing centred in the plot rather than hanging from its top.
		titleH, bodyH := float32(0), float32(0)
		if opts.Title != "" {
			titleH = LabelHeight(c, theme.LeadSize)
		}
		if opts.Body != "" {
			_, bodyH = p.MeasureText(width, ui.Span{Text: opts.Body, Size: theme.MetaSize})
		}
		art := step * 11
		total := art + titleH + bodyH + step*3
		y := plot.Y + (plot.H-total)/2
		if y < plot.Y {
			y = plot.Y
		}

		if opts.Art != nil {
			opts.Art(p, ui.Rect{X: x + (width-art)/2, Y: y, W: art, H: art * 0.6})
			y += art + step*1.5
		} else {
			defaultChartArt(p, ui.Rect{X: x + (width-art)/2, Y: y + art*0.2, W: art, H: art * 0.4}, k, step)
			y += art + step*1.5
		}
		if opts.Title != "" {
			w := LabelWidth(c, opts.Title, theme.LeadSize)
			p.Text(x+(width-w)/2, y, opts.Title, theme.LeadSize, k.Text)
			y += titleH + step
		}
		if opts.Body != "" {
			p.RichText(x, y, width, ui.Span{Text: opts.Body, Size: theme.MetaSize, Color: k.TextMuted})
		}
	})
}

// defaultChartArt draws a few blocks that are not there, and the one bright
// dot that says they would be. They are the same marks the interface's other
// empty states are drawn from, so a chart with nothing to say says it the way
// the rest of the interface says it.
func defaultChartArt(p *ui.Painter, area ui.Rect, k theme.Tokens, step float32) {
	dot := step * 3
	blocks := ui.Rect{X: area.X, Y: area.Y, W: area.W - dot - step, H: area.H}
	internal.Tiles(p, blocks, []float32{1, 0.85, 0.7}, step*1.5, step*0.5, k.TextMuted.Alpha(0.22))
	internal.Dot(p, area.X+area.W-dot/2, area.Y+area.H/2, dot/2, k.Lively)
}
