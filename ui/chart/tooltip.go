package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// TooltipRow is one line of a [Tooltip]: a series' name and the value under
// the pointer.
type TooltipRow struct {
	// Name is the series' name.
	Name string
	// Color is the colour its line is drawn in, so the row and the line agree.
	Color ui.Color
	// Value is what the series has at this point, already formatted: the
	// tooltip is the one place a value is written out in full.
	Value string
}

// TooltipOptions configure a [Tooltip].
type TooltipOptions struct {
	// Frame is the chart's frame: the tooltip stays inside its plot area.
	Frame FrameResult
	// Hover is the chart's pointer. A nil hover, or one that is not over the
	// plot, draws nothing at all.
	Hover *Hover
	// Title names what the values belong to — a category, a date, a branch.
	Title string
	// Rows are the values, one per series.
	Rows []TooltipRow
	// X and Y are the axes' scales, used to place a pinned At; a tooltip that
	// follows the pointer does not need them.
	X, Y Scale
	// At hangs the tooltip off a point rather than off the pointer, for a
	// chart that keeps its tooltip on a value the reader has chosen. nil
	// follows the hover.
	At *Point
	// MaxWidth caps how wide the card may get; zero lets it be as wide as its
	// rows need, which is as wide as the plot. A cap narrower than the text
	// lets the rows overlap, so only set one that the values fit inside.
	MaxWidth float32
	// Offset moves the card off its anchor along the axis, so it clears the
	// point it is about; zero puts it just clear of it.
	Offset float32
}

// Tooltip is what a chart says about the point the pointer is over: a small
// card with the values under it, set beside the point and kept inside the
// plot.
//
// It is drawn rather than laid out, because it belongs to a plot area whose
// size is only known while the chart paints — and because a tooltip that
// measured its own text before the frame knew where it was would be a frame
// behind.
func Tooltip(c *ui.Context, opts TooltipOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Title
	for _, row := range opts.Rows {
		if label != "" {
			label += " "
		}
		label += row.Name + " " + row.Value
	}

	e := ui.Box(c).Absolute().Fill()
	if label != "" {
		e = e.Label(label)
	}
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := opts.Frame.Plot()
		if plot.W <= 0 || plot.H <= 0 || (opts.Hover == nil && opts.At == nil) {
			return
		}
		anchor := plot.X + plot.W/2
		switch {
		case opts.At != nil && opts.X.OK():
			anchor = opts.X.Span(plot.X, plot.X+plot.W).At(opts.At.X)
		case opts.Hover != nil:
			if !opts.Hover.Over {
				// No pointer over the chart is no tooltip: one that stayed
				// behind would be a card claiming a point nobody is on.
				return
			}
			anchor = opts.Hover.X
		}
		card := tooltipCard(c, opts, plot)
		// Beside the point, and back the other way when there is no room on
		// that side: a tooltip that runs off the chart is worse than one that
		// covers the line it is about.
		x := anchor + card.W + opts.Offset
		if x+card.W > plot.X+plot.W {
			x = anchor - card.W - opts.Offset
		}
		y := plot.Y + (plot.H-card.H)/2
		x = clampLabel(x, card.W, plot.X, plot.X+plot.W)
		y = maxf(plot.Y, minf(y, plot.Y+plot.H-card.H))

		p.Clip(plot, 0, func() {
			p.Fill(ui.Rect{X: x, Y: y, W: card.W, H: card.H}, k.Fill, theme.ControlRadius)
			name, value := LabelHeight(c, theme.MetaSize), LabelHeight(c, theme.StatSize)
			text := y + u*1.5
			if opts.Title != "" {
				p.Text(x+u*2.5, text, opts.Title, theme.RowSize, k.OnFill.Alpha(0.7))
				text += LabelHeight(c, theme.RowSize) + u*0.5
			}
			for _, row := range opts.Rows {
				internal.Dot(p, x+u*3.25, text+name/2, u, row.Color)
				p.Text(x+u*5.25, text, row.Name, theme.MetaSize, k.OnFill.Alpha(0.75))
				w := LabelWidth(c, row.Value, theme.StatSize)
				p.Text(x+card.W-u*2.5-w, text+(name-value)/2, row.Value, theme.StatSize, k.OnFill)
				text += name + u*0.5
			}
		})
	})
}

// tooltipCard measures the card: as wide as the widest row plus its margins,
// as tall as its lines. Measuring is the same job as drawing it, done once
// here so the drawing only has to place what this returns.
func tooltipCard(c *ui.Context, opts TooltipOptions, plot ui.Rect) ui.Rect {
	u := core.Density(c).Unit()
	maxW := opts.MaxWidth
	if maxW <= 0 {
		maxW = plot.W
	}
	lines := len(opts.Rows)
	if opts.Title != "" {
		lines++
	}
	height := float32(lines)*LabelHeight(c, theme.MetaSize) + u*3 +
		(u * 0.5 * float32(max(lines-1, 0)))
	if opts.Title != "" {
		height += LabelHeight(c, theme.RowSize)
	}
	w := LabelWidth(c, opts.Title, theme.RowSize)
	for _, row := range opts.Rows {
		w = maxf(w, u*5.25+LabelWidth(c, row.Name, theme.MetaSize)+
			u*2+LabelWidth(c, row.Value, theme.StatSize))
	}
	return ui.Rect{W: minf(maxf(w+u*5, 1), maxf(maxW, 1)), H: maxf(height, 1)}
}
