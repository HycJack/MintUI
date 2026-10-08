package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Line is a reference line across a plot: the line a target was set at, a
// budget, the average, the moment something happened.
type Line struct {
	// Value is where the line falls on the axis it is drawn against.
	Value float64
	// Vertical draws it up the plot at that x instead of across it at that y.
	Vertical bool
	// Text is written at the end of the line, inside the plot.
	Text string
	// Color is what it is drawn in; the zero colour uses the accent, which is
	// the one colour in the interface that means "look here".
	Color ui.Color
	// Dashed breaks the line into dashes, for a reference rather than a
	// reading of the data.
	Dashed bool
}

// Callout is a label on one datum: the peak, the outage, the callback worth
// naming.
type Callout struct {
	// At is the datum it belongs to, in the axis' own values.
	At Point
	// Text is what it says.
	Text string
	// Color is what the label is drawn in; the zero colour uses the accent.
	Color ui.Color
	// Dx and Dy move the label off its point, in DIPs — up and to the right,
	// away from the data, which is the usual place for a callout. The default
	// is half a point's height above the point.
	Dx, Dy float32
}

// AnnotationOptions configure an [Annotation].
type AnnotationOptions struct {
	// Frame is the chart's frame, whose plot area everything is drawn in.
	Frame FrameResult
	// X and Y are the axes' scales, placed in the plot as this draws.
	X, Y Scale
	// Line draws a reference line across the plot; nil draws none.
	Line *Line
	// Label draws a callout on one point; nil draws none.
	Label *Callout
}

// Annotation draws the things a chart says about its own data: a reference
// line across it, a label on one point, or both.
//
// It is a component rather than something the chart types do themselves,
// because "the target was 40 and this one day went past it" is the same
// piece of drawing whichever chart it is written over.
func Annotation(c *ui.Context, opts AnnotationOptions) *ui.Element {
	k := core.Tokens(c)
	if opts.Line == nil && opts.Label == nil {
		panic("chart: Annotation needs a line, a label, or both")
	}
	e := ui.Box(c).Absolute().Fill().Label("Annotation")
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot := opts.Frame.Plot()
		if plot.W <= 0 || plot.H <= 0 {
			return
		}
		xs := opts.X.Span(plot.X, plot.X+plot.W)
		ys := opts.Y.Span(plot.Y+plot.H, plot.Y)
		p.Clip(plot, 0, func() {
			if line := opts.Line; line != nil {
				drawLine(p, c, plot, *line, xs, ys, k)
			}
			if call := opts.Label; call != nil {
				drawCallout(p, c, plot, *call, xs, ys, k)
			}
		})
	})
}

// drawLine draws the reference line and its label at the end of it: a pill in
// the fill colour, so the label reads over any colour the data happens to be.
func drawLine(p *ui.Painter, c *ui.Context, plot ui.Rect, line Line, xs, ys Scale, k theme.Tokens) {
	step := core.Density(c).Unit()
	color := line.Color
	if color == (ui.Color{}) {
		color = k.Accent
	}
	size := theme.CaptionSize
	if line.Vertical {
		x := xs.At(line.Value)
		dashedLine(p, x, plot.Y, x, plot.Y+plot.H, theme.BorderWidth, color, line.Dashed)
		if line.Text == "" {
			return
		}
		w, h := LabelWidth(c, line.Text, size), LabelHeight(c, size)
		y := plot.Y + plot.H - h - step*2
		p.Fill(ui.Rect{X: x + step/2, Y: y, W: w + step, H: h}, k.Fill, theme.SmallRadius)
		p.Text(x+step, y, line.Text, size, k.OnFill)
		return
	}
	y := ys.At(line.Value)
	dashedLine(p, plot.X, y, plot.X+plot.W, y, theme.BorderWidth, color, line.Dashed)
	if line.Text == "" {
		return
	}
	w, h := LabelWidth(c, line.Text, size), LabelHeight(c, size)
	p.Fill(ui.Rect{X: plot.X + plot.W - w - step, Y: y - h - step/2, W: w + step, H: h},
		k.Fill, theme.SmallRadius)
	p.Text(plot.X+plot.W-w-step/2, y-h-step/2, line.Text, size, k.OnFill)
}

// drawCallout draws a label on a point, with a short leader from the point to
// the label when it has been moved off it.
func drawCallout(p *ui.Painter, c *ui.Context, plot ui.Rect, call Callout, xs, ys Scale, k theme.Tokens) {
	x, y, ok := screen(xs, ys, call.At)
	if !ok {
		return
	}
	color := call.Color
	if color == (ui.Color{}) {
		color = k.Accent
	}
	step := core.Density(c).Unit()
	size := theme.CaptionSize
	w := LabelWidth(c, call.Text, size)
	h := LabelHeight(c, size)
	dx, dy := call.Dx, call.Dy
	if dx == 0 && dy == 0 {
		dy = -(h + step*2)
	}
	box := ui.Rect{X: x + dx, Y: y + dy, W: w + step, H: h + step/2}
	// Keep the callout inside the plot: one that ran off would be the one
	// annotation a reader could not read.
	box.X = clampLabel(box.X, box.W, plot.X, plot.X+plot.W)
	box.Y = maxf(plot.Y, minf(box.Y, plot.Y+plot.H-box.H))
	if dx != 0 || dy != 0 {
		dashedLine(p, x, y, box.X+box.W/2, box.Y+box.H/2, theme.BorderWidth, color.Alpha(0.5), true)
	}
	p.Fill(box, k.Fill, theme.SmallRadius)
	p.Text(box.X+step/2, box.Y+step/4, call.Text, size, k.OnFill)
}

// dashedLine draws a straight line, or the same line in dashes: on, off, on
// off. A path in this library is stroked solid, so a dashed reference line is
// a run of short ones.
func dashedLine(p *ui.Painter, x0, y0, x1, y1, width float32, color ui.Color, dashed bool) {
	if !dashed {
		p.Line(x0, y0, x1, y1, width, color)
		return
	}
	const on, off = 4.0, 3.0
	dx, dy := x1-x0, y1-y0
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length == 0 {
		return
	}
	for at := float32(0); at < length; at += on + off {
		end := minf(at+on, length)
		t0, t1 := at/length, end/length
		p.Line(x0+dx*t0, y0+dy*t0, x0+dx*t1, y0+dy*t1, width, color)
	}
}
