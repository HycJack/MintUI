package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The charts in this file are drawn round a circle rather than across a
// square: the pie and its two cousins, the radar and the polar plot, and the
// two single-figure dials. None of them has a cartesian axis, so none of them
// asks for one — they take the plot area from the frame and work out their own
// centre in it, which is the one place a chart type has to do geometry
// rather than arithmetic.

// ── pie, donut, rose ───────────────────────────────────────────────────────

// PieOptions configure a [PieChart], a [DonutChart] and a [NightingaleChart].
type PieOptions struct {
	ChartOptions
	// Labels are the categories, one per value, and are what the slices are
	// named by in the legend.
	Labels []string
	// Values are what each category is worth. A value of zero takes no angle
	// at all, which is what a category with nothing in it should look like.
	Values []float64
	// Start is where the first slice begins, in degrees round from three
	// o'clock; zero is the default, so the slices start at the top-left the
	// way a clock face's quarters do.
	Start float32
	// Ratio is how much of its angle a slice is drawn, the rest being the gap
	// between it and the next; zero uses 0.92, which is enough to see the
	// slices apart and not enough to change what they say.
	Ratio float32
}

// DonutOptions configure a [DonutChart].
type DonutOptions struct {
	PieOptions
	// Hole is the inner radius as a share of the outer one; zero uses 0.6.
	Hole float32
}

// NightingaleOptions configure a [NightingaleChart].
type NightingaleOptions struct {
	PieOptions
}

// PieChart is a circle split into shares: one slice per category, each taking
// as much of the circle as its share of the total.
//
// It is a chart of a small number of parts of one whole, which is the only
// thing it is good at: past a handful of slices a reader is comparing angles
// by eye, and a bar chart of the same numbers is easier to read than either.
func PieChart(c *ui.Context, opts PieOptions) *ui.Element {
	return pie(c, opts, 0, false)
}

// DonutChart is a [PieChart] with its middle taken out: the same shares, with
// the total written in the hole, which is where the number a reader wants —
// "what was it in total" — has room to go.
func DonutChart(c *ui.Context, opts DonutOptions) *ui.Element {
	hole := opts.Hole
	if hole <= 0 {
		hole = 0.6
	}
	return pie(c, opts.PieOptions, minf(hole, 0.9), false)
}

// NightingaleChart is a rose: equal angles, radii out from the middle, so that
// each petal's area rather than its radius carries the value.
//
// It is the one chart where the square root is the point. A petal twice the
// radius has four times the area, and a reader comparing areas by eye is the
// whole reason for drawing it this way round — a rose drawn on radii would be
// a bar chart with the bars bent.
func NightingaleChart(c *ui.Context, opts NightingaleOptions) *ui.Element {
	return pie(c, opts.PieOptions, 0, true)
}

// pie is the one body all three of the round charts share: an angle per value,
// a radius per value, and a label beside each slice that has room for one.
func pie(c *ui.Context, opts PieOptions, hole float32, root bool) *ui.Element {
	angles := Sweep(opts.Values)
	if len(angles) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	if len(opts.Labels) != len(opts.Values) && len(opts.Labels) != 0 {
		panic("chart: a pie needs one label for every slice, or none at all")
	}
	colors := Palette(c, len(angles))
	if opts.Legend.Entries == nil {
		opts.Legend = named(c, opts.Labels, len(angles))
	}
	// A rose divides the circle evenly and spends a radius on each, so its
	// angles are an even split however the values fall.
	even := make([]float64, len(angles))
	for i := range even {
		even[i] = 1
	}
	if root {
		angles = Sweep(even)
	}
	ratio := opts.Ratio
	if ratio <= 0 {
		ratio = 0.92
	}
	largest := 0.0
	for _, v := range opts.Values {
		largest = maxf64(largest, v)
	}
	k := core.Tokens(c)
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Pie", func(p *ui.Painter, plot ui.Rect) {
			// The circle leaves room for the names written beside it,
			// which are the only thing that says what a slice is.
			radius := minf(plot.W, plot.H)/2 - ringRoom(c, opts.Labels)
			if radius <= 0 {
				return
			}
			cx, cy := plot.X+plot.W/2, plot.Y+plot.H/2
			at := opts.Start
			for i, a := range angles {
				to := at + a
				sliceRadius := radius
				if root {
					sliceRadius = radius * float32(math.Sqrt(max(opts.Values[i], 0)/largest))
				}
				from := at + (1-ratio)*a/2
				upto := to - (1-ratio)*a/2
				if upto > from {
					path := wedge(cx, cy, radius*hole, sliceRadius, from, upto,
						int(a/6)+2)
					p.FillPath(&path, colors[i])
					// A hairline in the surface colour: two slices that
					// meet are two things, and the seam between them is not
					// part of either.
					p.StrokePath(&path, theme.BorderWidth, k.Surface)
				}
				sliceLabel(c, p, cx, cy, sliceRadius, from, upto, opts.Labels, i, colors[i], radius, hole, root)
				at = to
			}
		})
	})
}

// sliceLabel writes a slice's name beside it, when there is room beside it for
// the name. A label that does not fit is left to the legend rather than
// written over its neighbours.
func sliceLabel(c *ui.Context, p *ui.Painter, cx, cy, radius, from, upto float32, labels []string, i int, color ui.Color, outer, hole float32, root bool) {
	if i >= len(labels) || labels[i] == "" {
		return
	}
	size := theme.CaptionSize
	width := LabelWidth(c, labels[i], size)
	height := LabelHeight(c, size)
	if upto-from < 12 {
		return
	}
	mid := (from + upto) / 2
	// A rose's petals are of different sizes, so its label hangs off the
	// outer edge of the ring rather than off the middle of the slice. The
	// label's own height is what pushes it clear of the rim: centred on the
	// rim it straddles the ring, half its letters on the colour and its
	// muted ink unreadable there.
	at := outer + height*0.75 + unit(c)*0.5
	if root {
		at = radius + (outer-radius)*0.15 + height*0.75 + unit(c)*0.5
	}
	x, y := point(cx, cy, at, mid)
	label := x - width/2
	if x-width/2 < 0 {
		label = x + width + unit(c)
	}
	p.Text(label, y-height/2, labels[i], size, core.Tokens(c).TextMuted)
}

// ── radar and polar ────────────────────────────────────────────────────────

// RadarOptions configure a [RadarChart].
type RadarOptions struct {
	ChartOptions
	// Axes are the spokes, one per value of each series, in the order they go
	// round.
	Axes []string
	// Series are the shapes. Every series has one value per axis, and the
	// values of different series are read against the same scale — which is
	// built from the largest of them all, since a radar of two things with
	// their own scales would compare nothing.
	Series []Series
	// Rings is how many circles the web has; zero uses three.
	Rings int
}

// RadarChart is one axis per spoke with a shape drawn across them: the way to
// put several measures of the same thing on one plot and let a profile be
// read off as a silhouette.
func RadarChart(c *ui.Context, opts RadarOptions) *ui.Element {
	series, entries := coloured(c, opts.Series)
	if len(opts.Axes) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	if opts.Legend.Entries == nil {
		opts.Legend.Entries = entries
	}
	// Each spoke is scaled by the largest value on it, so a chart of two
	// things whose measures differ by a hundredfold can still show both
	// shapes.
	peaks := make([]float64, len(opts.Axes))
	for i := range peaks {
		if d, ok := Column(series, i); ok {
			peaks[i] = maxf64(d.Max, 0)
		}
		if peaks[i] <= 0 {
			peaks[i] = 1
		}
	}
	rings := opts.Rings
	if rings < 1 {
		rings = 3
	}
	k := core.Tokens(c)
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Radar", func(p *ui.Painter, plot ui.Rect) {
			radius := minf(plot.W, plot.H)/2 - ringRoom(c, opts.Axes)
			if radius <= 0 {
				return
			}
			cx, cy := plot.X+plot.W/2, plot.Y+plot.H/2
			spokes := len(opts.Axes)
			// The web first: the rings and the spokes are what the shapes
			// are read against, so they have to be under them.
			for ring := 1; ring <= rings; ring++ {
				at := radius * float32(ring) / float32(rings)
				var web ui.Path
				web.Circle(cx, cy, at)
				p.StrokePath(&web, theme.BorderWidth, k.Border)
			}
			for i := range opts.Axes {
				x, y := point(cx, cy, radius, spokeAngle(i, spokes))
				p.Line(cx, cy, x, y, theme.BorderWidth, k.Border)
				if opts.Axes[i] == "" {
					continue
				}
				// The name goes outside its own spoke, and the box it is
				// written into is turned so that it is never upside down:
				// a label on the left of the circle reads rightwards and
				// the same label on the right reads leftwards.
				labelSpoke(c, p, cx, cy, radius, spokeAngle(i, spokes), opts.Axes[i], k)
			}
			for _, s := range series {
				var shape ui.Path
				for i := range opts.Axes {
					if i >= len(s.Points) {
						continue
					}
					at := radius * float32(max(s.Points[i].Y, 0)/peaks[i])
					x, y := point(cx, cy, at, spokeAngle(i, spokes))
					if i == 0 {
						shape.MoveTo(x, y)
					} else {
						shape.LineTo(x, y)
					}
				}
				shape.Close()
				p.FillPath(&shape, s.Color.Alpha(0.2))
				p.StrokePath(&shape, theme.BorderWidth*2, s.Color)
				for i, pt := range s.Points {
					if i >= len(opts.Axes) {
						break
					}
					at := radius * float32(max(pt.Y, 0)/peaks[i])
					x, y := point(cx, cy, at, spokeAngle(i, spokes))
					internal.Dot(p, x, y, unit(c)*0.5, s.Color)
				}
			}
		})
	})
}

// spokeAngle is where spoke i of n points, in degrees: the first spoke at the
// top and the rest evenly round from it, clockwise.
func spokeAngle(i, n int) float32 {
	if n <= 0 {
		return 0
	}
	return -90 + 360*float32(i)/float32(n)
}

// ringRoom is how far outside the circle a set of names is written, and so
// how much of the plot area is not the circle itself.
//
// It is the widest of the names rather than a constant, for the same reason a
// frame's gutter is the widest label on its axis and not a number: a ring
// drawn at a guessed radius writes half its names over the marks.
func ringRoom(c *ui.Context, names []string) float32 {
	widest := float32(0)
	for _, name := range names {
		widest = maxf(widest, LabelWidth(c, name, theme.CaptionSize))
	}
	if widest <= 0 {
		// Nothing is written outside: the numbers on a polar chart's rings
		// are two or three characters and sit inside the ring they name.
		return unit(c) * 2
	}
	return widest/2 + unit(c)*2
}

// labelSpoke writes a spoke's name at its end, inside the plot area and
// upright.
func labelSpoke(c *ui.Context, p *ui.Painter, cx, cy, radius, angle float32, name string, k theme.Tokens) {
	size := theme.CaptionSize
	w, h := LabelWidth(c, name, size), LabelHeight(c, size)
	x, y := point(cx, cy, radius+unit(c)*2, angle)
	// The label is written to the side its spoke points at, so that it never
	// crosses the web it names.
	switch {
	case angle > -135 && angle < -45: // the top
		p.Text(x-w/2, y-h, name, size, k.TextMuted)
	case angle >= -45 && angle <= 45: // the right
		p.Text(x, y-h/2, name, size, k.TextMuted)
	case angle > 45 && angle < 135: // the bottom
		p.Text(x-w/2, y, name, size, k.TextMuted)
	case angle >= 135 || angle <= -135: // the left
		p.Text(x-w, y-h/2, name, size, k.TextMuted)
	default:
		p.Text(x-w/2, y-h/2, name, size, k.TextMuted)
	}
}

// PolarOptions configure a [PolarChart].
type PolarOptions struct {
	ChartOptions
	// Series are the lines, wrapped round the circle: the x value is how far
	// round it is and the y value how far out from the middle.
	Series []Series
	// Rings is how many circles the grid has; zero uses [DefaultTickCount].
	Rings int
}

// PolarChart is a line chart bent round a circle: the x axis wrapped into an
// angle, for a value that repeats over a cycle — a day, a month, a wave.
func PolarChart(c *ui.Context, opts PolarOptions) *ui.Element {
	chart, series, ok := cartesian(c, opts.Series, opts.ChartOptions)
	if !ok {
		return nothing(c, opts.ChartOptions)
	}
	rings := opts.Rings
	if rings < 1 {
		rings = DefaultTickCount
	}
	k := core.Tokens(c)
	return panel(c, chart, func(f FrameResult) {
		marksBox(c, f, "Polar", func(p *ui.Painter, plot ui.Rect) {
			radius := minf(plot.W, plot.H)/2 - ringRoom(c, nil)
			if radius <= 0 {
				return
			}
			cx, cy := plot.X+plot.W/2, plot.Y+plot.H/2
			// The rings come off the same tick machinery as a cartesian axis,
			// so a polar chart's numbers are spaced and formatted the same way
			// and its labels still drop rather than collide.
			scale := NewLinear(Domain{Min: 0, Max: 1}, 0, radius)
			ticks := Ticks(c, scale, TickOptions{Count: rings})
			for _, t := range ticks {
				if !t.Drawn {
					continue
				}
				at := scale.At(t.Value)
				var ring ui.Path
				ring.Circle(cx, cy, at)
				p.StrokePath(&ring, theme.BorderWidth, k.Border)
				p.Text(cx+unit(c), cy-at-LabelHeight(c, theme.CaptionSize), t.Label,
					theme.CaptionSize, k.TextMuted)
			}
			for _, s := range series {
				ringLine(p, s, cx, cy, radius, s.Color, theme.BorderWidth*2)
			}
		})
	})
}

// ringLine is one series wrapped round a circle: each point's x is its place
// round it and its y how far out it sits, scaled to the whole.
func ringLine(p *ui.Painter, s Series, cx, cy, radius float32, color ui.Color, width float32) {
	var path ui.Path
	var peaks float64
	for _, pt := range s.Points {
		peaks = maxf64(peaks, pt.Y)
	}
	if peaks <= 0 {
		return
	}
	for i, pt := range s.Points {
		x, y := point(cx, cy, radius*float32(max(pt.Y, 0)/peaks), spokeAngle(i, len(s.Points)))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()
	p.StrokePath(&path, width, color)
}

// dial is the name a dial or a bullet carries for assistive technology: what
// it is, and what it says. The two are joined rather than written as one
// string with a gap, because a name with a trailing space is a caption.
func dial(kind, label string) string {
	if label == "" {
		return kind
	}
	return kind + ": " + label
}

// ── dials ──────────────────────────────────────────────────────────────────

// GaugeZone is one band of a gauge's range: "under", "close", "over", said in
// the dial's own colours.
type GaugeZone struct {
	// To is where the band ends, in the gauge's own values.
	To float64
	// Name is what the band is called, written beside the arc.
	Name string
	// Color is what the band is drawn in; the zero colour takes the palette's
	// at its place in the set.
	Color ui.Color
}

// GaugeOptions configure a [Gauge].
type GaugeOptions struct {
	// Value is where the needle points, between Min and Max.
	Value float64
	// Min and Max are the ends of the dial; zero to one hundred unless the
	// caller says otherwise.
	Min, Max float64
	// Width and Height are the box the dial takes; zero lets the layout give
	// it one.
	Width, Height float32
	// Label is written under the needle's value, and is the caller's sentence
	// — "72% within budget", "3 of 5 seats taken".
	Label string
	// Color is what the needle is drawn in; the zero colour takes the accent.
	Color ui.Color
	// Zones are the bands under the needle, smallest first.
	Zones []GaugeZone
}

// gaugeSpan is how much of a circle a dial sweeps, and where it starts: three
// quarters of the way round, from the lower left, which is where a dial has
// room for both its ends and its middle.
const (
	gaugeStart = 135.0
	gaugeSpan  = 270.0
)

// Gauge is one value on a dial: a needle on an arc, with the bands it falls
// among drawn under it.
//
// A dial has room for exactly one number, which is why it is a component and
// not a chart type: it answers "how is this one thing" and nothing else. Put
// a dial on a page to show a trend and the trend will be invisible.
func Gauge(c *ui.Context, opts GaugeOptions) *ui.Element {
	k := core.Tokens(c)
	lo, hi := opts.Min, opts.Max
	if hi <= lo {
		hi = lo + 1
	}
	needle := ArcAngle(float32(opts.Value), float32(lo), float32(hi), gaugeStart, gaugeSpan)
	colors := Palette(c, max(len(opts.Zones), 1))
	zoneColors := make([]ui.Color, len(opts.Zones))
	for i, z := range opts.Zones {
		zoneColors[i] = z.Color
		if zoneColors[i] == (ui.Color{}) {
			zoneColors[i] = colors[i]
		}
	}
	return mark(c, dial("Gauge", opts.Label), opts.Width, opts.Height, func(p *ui.Painter, r ui.Rect) {
		// The dial sweeps 270°, so its two ends dip 0.707 of the radius
		// below the centre while the crown of the arc reaches the radius
		// above it: the radius is what fits that whole sweep in the box,
		// and the centre sits one radius down from the top. Sized from the
		// width alone the ends painted straight through whatever the
		// layout put under the dial.
		radius := minf(r.W/2-unit(c)*2, (r.H-unit(c)*2)/(1+0.7071))
		if radius <= 0 {
			return
		}
		cx, cy := r.X+r.W/2, r.Y+radius
		// Drawn straight into the box it was given: a dial, a
		// bullet and a countdown have no frame, and an element made
		// while painting is one that never paints.
		// The track first, then the bands over it: a band is a part of
		// the range, not a thing floating over the dial.
		track := ring(cx, cy, radius, gaugeStart, gaugeStart+gaugeSpan, 64)
		p.StrokePath(&track, unit(c)*1.5, k.Fill)
		from := lo
		for i, z := range opts.Zones {
			at := minf64(max(z.To, from), hi)
			a0 := ArcAngle(float32(from), float32(lo), float32(hi), gaugeStart, gaugeSpan)
			a1 := ArcAngle(float32(at), float32(lo), float32(hi), gaugeStart, gaugeSpan)
			if a1 > a0 {
				arc := ring(cx, cy, radius, a0, a1, 32)
				p.StrokePath(&arc, unit(c)*1.5, zoneColors[i].Alpha(0.8))
			}
			from = at
		}
		x, y := point(cx, cy, radius, needle)
		p.Line(cx, cy, x, y, theme.BorderWidth*2, ink(opts.Color, k))
		internal.Dot(p, cx, cy, unit(c)*0.6, ink(opts.Color, k))
		value := Group(opts.Value, decimalsFor(opts.Max-opts.Min))
		w, h := LabelWidth(c, value, theme.StatSize), LabelHeight(c, theme.StatSize)
		// The value and its name travel as one block, anchored inside the
		// arc's crown: placed from the crown alone, a small dial's numbers
		// start above their own box and lose their first digit to the clip.
		vy := maxf(r.Y+unit(c), cy-radius*0.55-h)
		p.Text(cx-w/2, vy, value, theme.StatSize, k.Text)
		if opts.Label != "" {
			lw := LabelWidth(c, opts.Label, theme.CaptionSize)
			p.Text(cx-lw/2, vy+h+unit(c)*0.5, opts.Label, theme.CaptionSize, k.TextMuted)
		}
	})
}

// BulletBand is one range of a bullet chart: a qualitative band under the measure,
// from Min up to Max of its own.
type BulletBand struct {
	// Min and Max are the band's two ends, in the measure's own values.
	Min, Max float64
	// Name is what the band is called.
	Name string
	// Color is what it is drawn in; the zero colour takes the palette's.
	Color ui.Color
}

// BulletOptions configure a [BulletChart].
type BulletOptions struct {
	// Value is the measurement: what to say against the bands.
	Value float64
	// Min and Max are the ends of the scale; zero to one hundred unless the
	// caller says otherwise.
	Min, Max float64
	// Target is where the measure was meant to land; zero draws no target.
	Target float64
	// Bands are the qualitative ranges under the measure, smallest first.
	Bands []BulletBand
	// Width and Height are the box it takes; zero lets the layout give it
	// one.
	Width, Height float32
	// Label is what the value is called, written beside it.
	Label string
	// Color is what the measure is drawn in; the zero colour takes the
	// accent.
	Color ui.Color
}

// BulletChart is one measure against its scale, its target and the bands that
// say whether the number is good: what a gauge would say if it could show more
// than one thing at a time, in the space a gauge would have taken.
func BulletChart(c *ui.Context, opts BulletOptions) *ui.Element {
	k := core.Tokens(c)
	lo, hi := opts.Min, opts.Max
	if hi <= lo {
		hi = lo + 1
	}
	colors := Palette(c, max(len(opts.Bands), 1))
	return mark(c, dial("Bullet", opts.Label), opts.Width, opts.Height, func(p *ui.Painter, r ui.Rect) {
		if r.W <= 0 || r.H <= 0 {
			return
		}
		height := r.H * 0.42
		y := r.Y + r.H*0.3
		// Drawn straight into the box it was given: a dial, a
		// bullet and a countdown have no frame, and an element made
		// while painting is one that never paints.
		// The bands run the length of the scale, smallest first, and
		// the measure is drawn over all of them.
		from := lo
		for i, b := range opts.Bands {
			start := minf64(max(b.Min, lo), hi)
			end := minf64(max(b.Max, lo), hi)
			if end > start {
				x0 := r.X + r.W*float32((start-lo)/(hi-lo))
				p.Fill(ui.Rect{X: x0, Y: y, W: r.W * float32((end-start)/(hi-lo)), H: height},
					colors[i].Alpha(0.22), unit(c)*0.5)
			}
			from = minf64(max(b.Max, from), hi)
		}
		scale := NewLinear(Domain{Min: lo, Max: hi}, r.X, r.X+r.W)
		bar := ui.Rect{X: scale.At(lo), Y: y + height*0.2,
			W: maxf(scale.At(opts.Value)-scale.At(lo), unit(c)),
			H: height * 0.6}
		p.Fill(bar, ink(opts.Color, k), unit(c)*0.4)
		if opts.Target > lo && opts.Target < hi {
			at := scale.At(opts.Target)
			p.Line(at, y-height*0.25, at, y+height*1.25, theme.BorderWidth*2, k.Text)
		}
		value := Group(opts.Value, decimalsFor(hi-lo))
		if opts.Label != "" {
			value = opts.Label + " " + value
		}
		w := LabelWidth(c, value, theme.RowSize)
		p.Text(r.X+r.W-w, y-height-LabelHeight(c, theme.RowSize), value, theme.RowSize, k.Text)
	})
}
