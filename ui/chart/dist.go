package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The charts in this file show a distribution rather than a value: where a
// sample's middle is, how far it spreads, what shape it has. Both take their
// samples as [Sample]s — a name and the values behind it — because every
// distribution chart compares groups, and a chart that could only take one
// would be a chart that could not compare anything.

// Sample is one named set of readings: what a box plot or a violin plot is
// comparing one group of numbers with another.
type Sample struct {
	// Label is the category's name, written under it.
	Label string
	// Values are the readings in it. They are the caller's: nothing here
	// generates a number, and a group with none is a group the chart leaves
	// out.
	Values []float64
	// Color is what the group is drawn in; the zero colour takes the
	// palette's, at its place in the set.
	Color ui.Color
}

// samples is a set of groups with their colours resolved, the range they need
// between them, and whether there was anything in them at all.
//
// The range is the values' own, outliers and all, because a chart that cut
// its outliers off the axis to make room would be hiding the very marks it
// draws.
func samples(c *ui.Context, in []Sample) ([]Sample, Domain, bool) {
	out := make([]Sample, len(in))
	colors := Palette(c, len(in))
	var d Domain
	seen := false
	for i, g := range in {
		out[i] = g
		if out[i].Color == (ui.Color{}) {
			out[i].Color = colors[i]
		}
		if one, ok := DomainOf(g.Values); ok {
			if !seen {
				d, seen = one, true
				continue
			}
			d = d.Union(one)
		}
	}
	return out, d, seen
}

// groupLabels is a set of samples' names: the categories of the x axis.
func groupLabels(in []Sample) []string {
	out := make([]string, len(in))
	for i, g := range in {
		out[i] = g.Label
	}
	return out
}

// legendOf is the legend of a set of samples: each one named and coloured, so
// that a chart of three samples says which is which without its labels having
// to be read against the x axis.
func legendOf(in []Sample) []LegendEntry {
	out := make([]LegendEntry, 0, len(in))
	for _, g := range in {
		if g.Label == "" {
			continue
		}
		out = append(out, LegendEntry{Name: g.Label, Color: g.Color, Mark: MarkSquare})
	}
	return out
}

// ── box plot ───────────────────────────────────────────────────────────────

// BoxPlotOptions configure a [BoxPlot].
type BoxPlotOptions struct {
	ChartOptions
	// Samples are the groups of readings, one per category.
	Samples []Sample
}

// BoxPlot is the five-number summary of each sample drawn as a box and two
// whiskers: the median in the middle, the middle half of the values in the
// box, and the ends of the range at the whiskers.
//
// It is a chart about spread rather than about size, and it says so: the box
// is as tall as the middle half is wide whatever the sample's total is, so two
// groups of very different sizes but the same spread look alike here. That is
// the reason to reach for this rather than for a bar of averages.
func BoxPlot(c *ui.Context, opts BoxPlotOptions) *ui.Element {
	in, d, ok := samples(c, opts.Samples)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	opts.X = axisX(opts.X, groupLabels(in))
	opts.Y = axisValues(opts.Y, d, opts.Y.Count)
	opts.ChartOptions = charted(opts.ChartOptions, legendOf(in))
	k, u := core.Tokens(c), unit(c)
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Box plot", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			for i, g := range in {
				s, ok := Summary(g.Values)
				if !ok {
					continue
				}
				whiskers(p, xs, ys, i, s, g.Color)
				box(p, xs, ys, i, s, g.Color)
				for _, out := range s.Outliers {
					internal.Ring(p, midOf(xs, i), ys.At(out), u*0.35, theme.BorderWidth, k.TextMuted)
				}
			}
		})
	})
}

// midOf is the middle of a category's share of a band axis — where a mark
// belonging to that category belongs.
func midOf(xs Scale, i int) float32 {
	_, mid, _ := xs.Share(i)
	return mid
}

// box is the middle half of a sample, with its median across it.
func box(p *ui.Painter, xs, ys Scale, i int, s Stats, color ui.Color) {
	_, mid, _ := xs.Share(i)
	width := bandWidth(xs, i, 1) * 0.55
	x := mid - width/2
	p.Fill(ui.Rect{X: x, Y: ys.At(s.Q3), W: width, H: maxf(ys.At(s.Q1)-ys.At(s.Q3), 1)},
		color.Alpha(0.4), minf(width/4, theme.SmallRadius/2))
	p.Line(x, ys.At(s.Median), x+width, ys.At(s.Median), theme.BorderWidth*2, color)
}

// whiskers are the line from the lowest value inside the fences to the
// highest, with a cap at each end. They go down before the box does, so what
// is left on top is the box's own edge rather than a line through it.
func whiskers(p *ui.Painter, xs, ys Scale, i int, s Stats, color ui.Color) {
	mid := midOf(xs, i)
	width := bandWidth(xs, i, 1) * 0.55
	lo, hi := ys.At(s.Lower), ys.At(s.Upper)
	ink := color.Alpha(0.7)
	p.Line(mid, lo, mid, hi, theme.BorderWidth, ink)
	p.Line(mid-width*0.25, lo, mid+width*0.25, lo, theme.BorderWidth, ink)
	p.Line(mid-width*0.25, hi, mid+width*0.25, hi, theme.BorderWidth, ink)
}

// ── violin plot ────────────────────────────────────────────────────────────

// ViolinOptions configure a [ViolinPlot].
type ViolinOptions struct {
	ChartOptions
	// Samples are the groups of readings, one per category.
	Samples []Sample
	// Bandwidth is how wide each sample's kernel is; zero works it out from
	// the sample's own spread, which is what keeps two violins of different
	// spreads comparable.
	Bandwidth float64
	// Steps is how finely each shape is drawn; zero uses thirty-two, which is
	// as much detail as a violin a hundred pixels across can show.
	Steps int
}

// ViolinPlot is a sample's shape on its side: wide where its values pile up,
// narrow where there are few, mirrored about the category's own place so that
// two of them can be laid side by side and compared as silhouettes.
func ViolinPlot(c *ui.Context, opts ViolinOptions) *ui.Element {
	in, d, ok := samples(c, opts.Samples)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	opts.X = axisX(opts.X, groupLabels(in))
	opts.Y = axisValues(opts.Y, d, opts.Y.Count)
	opts.ChartOptions = charted(opts.ChartOptions, legendOf(in))
	steps := opts.Steps
	if steps < 2 {
		steps = 32
	}
	k := core.Tokens(c)
	return canvas(c, opts.ChartOptions, func(f FrameResult) {
		marks(c, f, "Violin plot", func(p *ui.Painter, plot ui.Rect, xs, ys Scale) {
			for i, g := range in {
				violin(p, xs, ys, i, g.Values, opts.Bandwidth, steps, g.Color.Alpha(0.7), k.Text)
			}
		})
	})
}

// violin draws one sample's mirrored density. Every violin is scaled to the
// same width whatever its own peak density is, because what two of them are
// compared on is the shape of each and not how many values happened to be in
// it.
func violin(p *ui.Painter, xs, ys Scale, i int, values []float64, bandwidth float64, steps int, color ui.Color, median ui.Color) {
	one, ok := DomainOf(values)
	if !ok || one.Span() == 0 {
		return
	}
	width := maxf64(bandwidth, one.Span()/6)
	half := bandWidth(xs, i, 1) / 2
	if half <= 0 {
		return
	}
	densest := 0.0
	for step := range steps + 1 {
		at := one.Min + one.Span()*float64(step)/float64(steps)
		densest = maxf64(densest, Density(at, values, width))
	}
	if densest <= 0 {
		return
	}
	// One closed outline rather than a stack of segments: a violin is filled,
	// and a path of horizontal segments has no area to fill — it draws as a
	// hairline down the median and nothing else. The two edges come from the
	// same estimate, so they cannot cross.
	mid := midOf(xs, i)
	left := make([]float32, steps+1)
	right := make([]float32, steps+1)
	edge := make([]float32, steps+1)
	for step := range steps + 1 {
		at := one.Min + one.Span()*float64(step)/float64(steps)
		wide := half * float32(Density(at, values, width)/densest)
		edge[step] = ys.At(at)
		left[step] = mid - wide
		right[step] = mid + wide
	}
	path := ui.Path{}
	path.MoveTo(left[0], edge[0])
	for step := 1; step <= steps; step++ {
		path.LineTo(left[step], edge[step])
	}
	for step := steps; step >= 0; step-- {
		path.LineTo(right[step], edge[step])
	}
	path.Close()
	p.FillPath(&path, color)
	// The median across it, which is the one number a reader of a violin
	// always wants and the one a shape on its own does not say.
	if s, ok := Summary(values); ok {
		y := ys.At(s.Median)
		p.Line(midOf(xs, i)-half*0.35, y, midOf(xs, i)+half*0.35, y, theme.BorderWidth, median)
	}
}
