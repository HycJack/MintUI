package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Side is which edge of a frame's plot area an axis runs along.
//
// It is named for the plot it belongs to rather than for the line's
// direction, because "vertical" reads both ways and getting it backwards puts
// the numbers on the wrong side of the chart.
type Side uint8

const (
	// Bottom is where an x axis goes: the values increase to the right.
	Bottom Side = iota
	// Left is where a y axis goes: the values increase upwards.
	Left
	// Top and Right are the other two, for a chart that puts its numbers
	// somewhere else — a second scale along the top of a plot, or a
	// temperature axis down the right.
	Top
	Right
)

func (s Side) String() string {
	switch s {
	case Left:
		return "Left"
	case Top:
		return "Top"
	case Right:
		return "Right"
	}
	return "Bottom"
}

// AxisOptions configure an [Axis], and are what a [Frame] measures its gutters
// from: the same options that say what the labels look like say how much room
// they need.
type AxisOptions struct {
	// Side is which edge of the plot this axis runs along. Zero is the
	// bottom, which is where an x axis goes.
	Side Side
	// Scale says where the values fall: linear, logarithmic, or the categories
	// of a band scale. It is written in the axis' own space — 0 to 1 — and the
	// frame places it in the plot as it paints. The zero scale, which means no
	// axis at all, is what a caller leaves in an axis they do not want.
	Scale Scale
	// Count is how many steps it asks for; zero asks for [DefaultTickCount].
	Count int
	// Format renders the values; nil uses [Auto].
	Format Format
	// Label names the axis. A y axis' name is written at the head of the
	// numbers it names, above them, rather than turned on its side — a word
	// on its side is a word nobody scans.
	Label string
	// Ticks draws the short marks where each tick meets the plot's edge. A
	// chart with a grid does not need them, which is why it is off unless
	// asked for, the same way a border is.
	Ticks bool
	// Size is the labels' font size; zero uses [theme.CaptionSize].
	Size float32
}

// Axis draws one of a chart's axes: its labels, its tick marks and its name,
// in the gutter the frame left it.
//
// It takes the frame rather than a rectangle because an axis belongs to the
// plot it is measuring: its labels hang off the plot's edge, and an edge is
// the frame's to give.
func Axis(c *ui.Context, frame FrameResult, opts AxisOptions) *ui.Element {
	k := core.Tokens(c)
	size := tickSize(opts)
	e := ui.Box(c).Absolute().Fill()
	if opts.Label != "" {
		e = e.Label(opts.Label + " axis")
	}
	return e.Draw(func(p *ui.Painter, _ ui.Rect) {
		plot, bounds := frame.Plot(), frame.Bounds()
		if !opts.Scale.OK() || plot.W <= 0 || plot.H <= 0 {
			return
		}
		ticks := Ticks(c, place(opts.Scale, plot, opts.Side),
			TickOptions{Count: opts.Count, Format: opts.Format, Size: size,
				Vertical: opts.Side == Left || opts.Side == Right})
		if len(ticks) == 0 {
			return
		}
		gap := axisGap(c, opts)
		height := LabelHeight(c, size)

		// Everything here is in the gutter — the frame's own edge to the
		// plot's — so no label can spill into the data or off the frame.
		p.Clip(gutter(bounds, plot, opts.Side), 0, func() {
			if opts.Ticks {
				axisMarks(p, plot, ticks, gap/3, opts.Side, k.Border)
			}
			for _, t := range ticks {
				if !t.Drawn {
					continue
				}
				width := LabelWidth(c, t.Label, size)
				switch opts.Side {
				case Bottom:
					p.Text(clampLabel(t.Pos, width, bounds.X, bounds.X+bounds.W), plot.Y+plot.H+gap, t.Label, size, k.TextMuted)
				case Top:
					p.Text(clampLabel(t.Pos, width, bounds.X, bounds.X+bounds.W), plot.Y-gap-height, t.Label, size, k.TextMuted)
				case Left:
					// A y axis' tick positions run up the plot, so the
					// number is placed against the plot's own left edge
					// and not against the position — which is a line
					// height, and would put every label somewhere over
					// the data or off the frame entirely.
					p.Text(plot.X-gap-width, centre(t.Pos, height, bounds), t.Label, size, k.TextMuted)
				case Right:
					p.Text(plot.X+plot.W+gap, centre(t.Pos, height, bounds), t.Label, size, k.TextMuted)
				}
			}
		})
		// The name is clipped to the frame and not to the gutter: a y axis'
		// name is written at the head of the numbers it names, which is
		// above them and clear across the plot, and clipping it to a strip
		// as wide as the widest number would cut every word after the first
		// two letters.
		if opts.Label != "" {
			p.Clip(bounds, 0, func() {
				switch opts.Side {
				case Bottom:
					p.Text(bounds.X, plot.Y+plot.H+gap+height, opts.Label, theme.RowSize, k.Text)
				default:
					p.Text(bounds.X, bounds.Y, opts.Label, theme.RowSize, k.Text)
				}
			})
		}
	})
}

// place puts a scale in a plot area along the side its axis runs on. A y
// scale's ends are the other way round: its first value is at the bottom.
func place(s Scale, plot ui.Rect, side Side) Scale {
	switch side {
	case Bottom, Top:
		return s.Span(plot.X, plot.X+plot.W)
	default: // Left and Right both read upwards.
		return s.Span(plot.Y+plot.H, plot.Y)
	}
}

// gutter is the strip of a frame an axis writes in: everything between the
// plot's edge and the frame's own.
func gutter(bounds, plot ui.Rect, side Side) ui.Rect {
	switch side {
	case Bottom:
		y := plot.Y + plot.H
		return ui.Rect{X: bounds.X, Y: y, W: bounds.W, H: max(bounds.Y+bounds.H-y, 0)}
	case Top:
		return ui.Rect{X: bounds.X, Y: bounds.Y, W: bounds.W, H: max(plot.Y-bounds.Y, 0)}
	case Left:
		return ui.Rect{X: bounds.X, Y: bounds.Y, W: max(plot.X-bounds.X, 0), H: bounds.H}
	default:
		x := plot.X + plot.W
		return ui.Rect{X: x, Y: bounds.Y, W: max(bounds.X+bounds.W-x, 0), H: bounds.H}
	}
}

// axisMarks draws the short tick marks where the tick lines meet the plot's
// edge, each one on the labels' side of the line.
func axisMarks(p *ui.Painter, plot ui.Rect, ticks []Tick, mark float32, side Side, col ui.Color) {
	if mark <= 0 {
		return
	}
	for _, t := range ticks {
		if !t.Drawn {
			continue
		}
		switch side {
		case Bottom:
			p.Line(t.Pos, plot.Y+plot.H, t.Pos, plot.Y+plot.H-mark, theme.BorderWidth, col)
		case Top:
			p.Line(t.Pos, plot.Y, t.Pos, plot.Y+mark, theme.BorderWidth, col)
		case Left:
			p.Line(plot.X, t.Pos, plot.X+mark, t.Pos, theme.BorderWidth, col)
		default:
			p.Line(plot.X+plot.W, t.Pos, plot.X+plot.W-mark, t.Pos, theme.BorderWidth, col)
		}
	}
}

// clampLabel keeps a label written under its tick inside the frame: the first
// and the last are centred on the plot's edges, which are the frame's own, so
// half of each would otherwise fall off the chart.
func clampLabel(at, width, lo, hi float32) float32 {
	if hi < lo {
		return lo
	}
	return max(lo, min(at-width/2, hi-width))
}

// centre is where a label of that height belongs beside a tick, kept inside
// the frame — so the label at the very top of a y axis does not climb into
// the legend above it.
func centre(at, height float32, bounds ui.Rect) float32 {
	lo, hi := bounds.Y, bounds.Y+bounds.H-height
	if hi < lo {
		return lo
	}
	return max(lo, min(at-height/2, hi))
}
