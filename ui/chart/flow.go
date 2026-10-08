package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The charts in this file move something from one place to another — a
// funnel of steps that narrow, ribbons of flow between columns, and the
// chords of a circle — so they all share one problem: the marks are areas
// rather than lengths, so they have to be laid out from a count of what flows
// rather than from two scales. That count is [Depths] and [Link], which are
// plain numbers and are tested as such.

// ── funnel ─────────────────────────────────────────────────────────────────

// FunnelStep is one stage of a funnel: what it is called and how much got that
// far.
type FunnelStep struct {
	// Label is the stage's name, written in it.
	Label string
	// Value is how much reached it. A funnel's widths are read against each
	// other, so the numbers come from the caller rather than from a scale.
	Value float64
	// Color is what the stage is drawn in; the zero colour takes the
	// palette's at its place in the set.
	Color ui.Color
}

// FunnelOptions configure a [FunnelChart].
type FunnelOptions struct {
	ChartOptions
	// Steps are the stages, widest first — in the order the thing went
	// through them, which is not necessarily the order of their widths.
	Steps []FunnelStep
}

// FunnelChart is the shape of a thing getting smaller: each stage as wide as
// what got that far, so the drop between two stages is a gap a reader can see
// rather than two numbers they have to subtract.
func FunnelChart(c *ui.Context, opts FunnelOptions) *ui.Element {
	if len(opts.Steps) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	colors := Palette(c, len(opts.Steps))
	k := core.Tokens(c)
	size := theme.RowSize
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Funnel", func(p *ui.Painter, plot ui.Rect) {
			widest := 0.0
			for _, s := range opts.Steps {
				widest = maxf64(widest, s.Value)
			}
			if widest <= 0 {
				return
			}
			rows := len(opts.Steps)
			height := plot.H / float32(rows)
			// Half the width a stage's value earns, and the funnel's own
			// centre: every edge is written as centre plus or minus half a
			// width, so the two ends of an edge cannot end up in different
			// coordinate spaces — which is what folded every stage into a
			// self-crossing path and filled it as a small triangle.
			half := func(v float64) float32 { return plot.W * float32(v/widest) / 2 }
			cx := plot.X + plot.W/2
			for i, s := range opts.Steps {
				above, below := s.Value, s.Value
				if i > 0 {
					above = opts.Steps[i-1].Value
				}
				if i+1 < rows {
					below = opts.Steps[i+1].Value
				}
				hwTop, hwBottom := half(above), half(below)
				y := plot.Y + float32(i)*height
				color := s.Color
				if color == (ui.Color{}) {
					color = colors[i]
				}
				// Each stage is a trapezoid from the width of the one above
				// it to the width of the one below, so the funnel's sides
				// are one line all the way down.
				path := ui.Path{}
				path.MoveTo(cx-hwTop, y)
				path.LineTo(cx+hwTop, y)
				path.LineTo(cx+hwBottom, y+height)
				path.LineTo(cx-hwBottom, y+height)
				path.Close()
				p.FillPath(&path, color.Alpha(0.85))
				label := s.Label
				if s.Value > 0 {
					label = Group(s.Value, 0) + "  " + label
				}
				if label == "" {
					continue
				}
				w, h := LabelWidth(c, label, size), LabelHeight(c, size)
				// Inside the stage where there is room for the name — the
				// narrower of its two edges is the room it has — and beside
				// it where there is not, rather than over its neighbours
				// either way.
				if narrow := 2 * minf(hwTop, hwBottom); w < narrow {
					p.Text(cx-w/2, y+(height-h)/2, label, size, k.OnFill)
					continue
				}
				p.Text(minf(plot.X+plot.W-w-unit(c), plot.X), y+(height-h)/2, label, size, k.TextMuted)
			}
		})
	})
}

// ── sankey ─────────────────────────────────────────────────────────────────

// SankeyOptions configure a [SankeyChart].
type SankeyOptions struct {
	ChartOptions
	// Names are the nodes, in the order the caller wants them laid out down a
	// column. Their columns are worked out from the links, so a node with no
	// links of its own is a column of nothing.
	Names []string
	// Links are what flows from which node to which, and is the whole of the
	// data: a flow diagram has no other numbers.
	Links []Link
	// Node is how wide a node's column is, in DIPs; zero uses two thirds of
	// what is left after the columns and the gaps.
	Node float32
}

// SankeyChart is a set of nodes in columns with ribbons between them: how much
// of what one thing holds arrives at another, with the width of every ribbon
// in the same units as the node it leaves.
func SankeyChart(c *ui.Context, opts SankeyOptions) *ui.Element {
	if len(opts.Names) == 0 || len(opts.Links) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	k := core.Tokens(c)
	depth := Depths(len(opts.Names), opts.Links)
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Sankey", func(p *ui.Painter, plot ui.Rect) {
			gap := unit(c) * 2
			columns := 1
			for _, d := range depth {
				columns = max(columns, d+1)
			}
			width := opts.Node
			if width <= 0 {
				width = (plot.W - gap*float32(columns-1)) / float32(columns) * 0.6
			}
			if width <= 0 {
				return
			}
			span := plot.W - gap*float32(columns-1) - width*float32(columns)
			// A strip along the top for the nodes' own names, so the
			// topmost node is named rather than having its name clipped
			// off by the edge of the plot.
			head := LabelHeight(c, theme.CaptionSize) + unit(c)*2
			space := plot.H - head - gap*float32(max(deepest(depth), 1)-1)
			scale := ribbonScale(depth, opts.Links, space)
			if scale <= 0 {
				return
			}
			xOf := func(d int) float32 {
				return plot.X + float32(d)*(width+gap) + span*float32(d)/float32(max(columns, 1))
			}
			// Ribbons under the nodes, so that a ribbon which is wider
			// than its node still reads as leaving it.
			outs, ins := laidOut(depth, opts.Links, scale, gap, plot.Y+head, space)
			colors := Palette(c, len(opts.Names))
			firstOfColumn := map[int]bool{}
			seenColumn := map[int]bool{}
			for i, d := range depth {
				if !seenColumn[d] {
					seenColumn[d] = true
					firstOfColumn[i] = true
				}
			}
			for _, l := range opts.Links {
				from, to := outs[l.Source], ins[l.Target]
				height := float32(l.Value) * scale
				drawRibbon(p, xOf(depth[l.Source])+width, xOf(depth[l.Target]),
					from.top+from.used, to.top+to.used, height, colors[l.Source%len(colors)])
				from.used += height
				to.used += height
			}
			for i, name := range opts.Names {
				node := outs[i]
				x := xOf(depth[i])
				p.Fill(ui.Rect{X: x, Y: node.top, W: width, H: maxf(node.height, 1)},
					colors[i%len(colors)], minf(width/4, theme.SmallRadius/2))
				if name == "" {
					continue
				}
				w, h := LabelWidth(c, name, theme.CaptionSize), LabelHeight(c, theme.CaptionSize)
				// Only the first node of a column has room above it — the
				// strip the layout reserved — and every node after it has
				// just the inter-node gap, where a label prints straight
				// through the node above. The rest carry their name inside
				// themselves when they are tall enough, and beside
				// themselves when they are not.
				switch {
				case firstOfColumn[i]:
					p.Text(x+width/2-w/2, node.top-h-unit(c)*0.5, name, theme.CaptionSize, k.TextMuted)
				case h <= node.height && w < width:
					p.Text(x+width/2-w/2, node.top+(node.height-h)/2, name, theme.CaptionSize, k.OnFill)
				default:
					p.Text(x+width+unit(c), node.top, name, theme.CaptionSize, k.TextMuted)
				}
			}
		})
	})
}

// slot is where a node sits and how much of it its links have used up so far.
type slot struct {
	top, height, used float32
}

// laidOut is every node's place: its column's stack, its height in the same
// units as the links, and the offsets its own ribbons start and end at.
func laidOut(depth []int, links []Link, scale, gap, top, space float32) (out, in []slot) {
	totals := make([]float64, len(depth))
	for _, l := range links {
		totals[l.Source] += l.Value
		totals[l.Target] += l.Value
	}
	out = make([]slot, len(depth))
	in = make([]slot, len(depth))
	byColumn := map[int][]int{}
	for i, d := range depth {
		byColumn[d] = append(byColumn[d], i)
	}
	for _, column := range byColumn {
		y := top
		for _, i := range column {
			height := float32(totals[i]) * scale
			out[i] = slot{top: y, height: height}
			in[i] = slot{top: y, height: height}
			y += height + gap
		}
	}
	return out, in
}

// ribbonScale is the number of DIPs a unit of flow is worth: the biggest
// column's total fills the space the columns have, and every other one is
// smaller than that by however much less it holds.
func ribbonScale(depth []int, links []Link, space float32) float32 {
	totals := map[int]float64{}
	for _, l := range links {
		totals[depth[l.Source]] += l.Value
		totals[depth[l.Target]] += l.Value
	}
	tallest := 0.0
	for _, v := range totals {
		tallest = maxf64(tallest, v)
	}
	if tallest <= 0 || space <= 0 {
		return 0
	}
	return space / float32(tallest)
}

// deepest is how many columns there are, and so how many gaps the stack has to
// leave between the nodes in one.
func deepest(depth []int) int {
	most := 0
	for _, d := range depth {
		most = max(most, d+1)
	}
	return most
}

// drawRibbon is one flow between two columns: a band whose height at each end
// is the flow's own, curved between them so that it reads as moving rather
// than as two bars that happen to line up.
func drawRibbon(p *ui.Painter, x0, x1, y0, y1, height float32, color ui.Color) {
	if height <= 0 {
		return
	}
	mid := (x0 + x1) / 2
	path := ui.Path{}
	path.MoveTo(x0, y0)
	path.CubeTo(mid, y0, mid, y1, x1, y1)
	path.LineTo(x1, y1+height)
	path.CubeTo(mid, y1+height, mid, y0+height, x0, y0+height)
	path.Close()
	p.FillPath(&path, color.Alpha(0.35))
}

// ── alluvial ───────────────────────────────────────────────────────────────

// AlluvialOptions configure an [AlluvialChart].
type AlluvialOptions struct {
	ChartOptions
	// Stages are the columns, in order.
	Stages []string
	// Flows is what passes from one stage to the next: flows[i][j] is the
	// amount leaving stage i for stage j+1. The rows therefore overlap by
	// one, which is what makes it an alluvial rather than a Sankey — the
	// stages are the whole story and each one hands on what it kept.
	Flows [][]float64
	// Node is how wide a stage's column is; zero uses a fifth of the space a
	// column's share of the plot leaves.
	Node float32
}

// AlluvialChart is a flow between stages, each stage a full-height column:
// what reached the first one, what was kept, what passed on, and how much of
// it went which way.
func AlluvialChart(c *ui.Context, opts AlluvialOptions) *ui.Element {
	if len(opts.Stages) == 0 || len(opts.Flows) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	k := core.Tokens(c)
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Alluvial", func(p *ui.Painter, plot ui.Rect) {
			stages := len(opts.Stages)
			if len(opts.Flows) != stages-1 {
				panic("chart: an alluvial needs one row of flows between each pair of stages")
			}
			width := opts.Node
			if width <= 0 {
				width = plot.W / float32(stages*4) * 2
			}
			gap := unit(c) * 3
			space := plot.W - width*float32(stages) - gap*float32(stages-1)
			if space <= 0 || width <= 0 {
				return
			}
			// A strip along the top for the stages' names, and one along
			// the bottom for the flows that leave the last stage, so that
			// neither is clipped off by the edge of the plot.
			head := LabelHeight(c, theme.CaptionSize) + unit(c)
			// Every column is scaled by the same unit, so that a band leaving
			// one stage arrives at the next the same thickness.
			tallest := 0.0
			for _, row := range opts.Flows {
				var total float64
				for _, v := range row {
					total += v
				}
				tallest = maxf64(tallest, total)
			}
			if tallest <= 0 {
				return
			}
			scale := (plot.H - head*2) / float32(tallest)
			colors := Palette(c, stages)
			xOf := func(i int) float32 {
				return plot.X + float32(i)*(width+gap) + space*float32(i)/float32(max(stages-1, 1))
			}
			for i, row := range opts.Flows {
				y := plot.Y + head
				for _, v := range row {
					height := float32(v) * scale
					if height <= 0 {
						continue
					}
					drawRibbon(p, xOf(i)+width, xOf(i+1), y, y, height,
						colors[i%len(colors)])
					y += height
				}
			}
			for i, name := range opts.Stages {
				x := xOf(i)
				p.Fill(ui.Rect{X: x, Y: plot.Y + head, W: width, H: plot.H - head*2},
					colors[i%len(colors)], width/4)
				if name == "" {
					continue
				}
				w := LabelWidth(c, name, theme.CaptionSize)
				p.Text(x+width/2-w/2, plot.Y, name, theme.CaptionSize, k.TextMuted)
			}
		})
	})
}

// ── chords ─────────────────────────────────────────────────────────────────

// ChordOptions configure a [ChordDiagram].
type ChordOptions struct {
	ChartOptions
	// Names are the members, one per row and column of the matrix.
	Names []string
	// Matrix is what flows between them: matrix[i][j] arrives at i from j.
	// It has to be square and as wide as the names, or the chords would be
	// drawn between members that are not there.
	Matrix [][]float64
}

// ChordDiagram is the members of a set round a circle with a ribbon between
// every pair that has traffic: the same numbers a Sankey draws, with the
// members themselves laid out in a ring because they have no order of their
// own to be laid out in.
func ChordDiagram(c *ui.Context, opts ChordOptions) *ui.Element {
	if len(opts.Names) == 0 || len(opts.Matrix) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	if len(opts.Matrix) != len(opts.Names) {
		panic("chart: a chord diagram needs a square matrix, one row and column per name")
	}
	for _, row := range opts.Matrix {
		if len(row) != len(opts.Names) {
			panic("chart: a chord diagram's matrix has to be square, one row and column per name")
		}
	}
	totals := make([]float64, len(opts.Names))
	for i, row := range opts.Matrix {
		for j, v := range row {
			// A member's own traffic is not a chord to itself: a
			// diagram of a network does not draw a ribbon from every
			// node back to itself.
			if i != j {
				totals[i] += v
			}
		}
	}
	k := core.Tokens(c)
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Chords", func(p *ui.Painter, plot ui.Rect) {
			radius := minf(plot.W, plot.H)/2 - ringRoom(c, opts.Names)
			if radius <= 0 {
				return
			}
			cx, cy := plot.X+plot.W/2, plot.Y+plot.H/2
			// The members are laid out in proportion to what flows through
			// them, so a big member is a big arc and the chords underneath
			// have somewhere to land.
			angles := Sweep(totals)
			colors := Palette(c, len(opts.Names))
			at := float32(0)
			starts := make([]float32, len(angles))
			for i, a := range angles {
				starts[i] = at
				if a > 0 {
					segment := wedge(cx, cy, radius*0.82, radius, at+1, at+a-1, clampSteps(int(a)))
					p.FillPath(&segment, colors[i])
					p.StrokePath(&segment, theme.BorderWidth, k.Surface)
					if opts.Names[i] != "" {
						mid := at + a/2
						x, y := point(cx, cy, radius+unit(c)*2, mid)
						w, h := LabelWidth(c, opts.Names[i], theme.CaptionSize), LabelHeight(c, theme.CaptionSize)
						// Written to whichever side of the circle its own
						// arc is on, so that a label is never inside the
						// ring it names.
						if x < cx {
							p.Text(x-w-unit(c), y-h/2, opts.Names[i], theme.CaptionSize, k.TextMuted)
						} else {
							p.Text(x+unit(c), y-h/2, opts.Names[i], theme.CaptionSize, k.TextMuted)
						}
					}
				}
				at += a
			}
			// The chords, each one leaving as much of its own member's
			// arc as its share of that member's traffic: two members
			// with five flows between them get five chords side by side
			// rather than five stacked on one line.
			used := make([]float32, len(angles))
			for i, row := range opts.Matrix {
				for j, v := range row {
					if i == j || v <= 0 || totals[i] <= 0 {
						continue
					}
					span := angles[i] * float32(chordShare(i, j, opts.Matrix)/totals[i])
					if used[i]+span > angles[i] {
						span = angles[i] - used[i]
					}
					if span <= 0 {
						continue
					}
					a0, a1 := starts[i]+used[i], starts[j]+used[j]
					drawChord(p, cx, cy, radius*0.82, a0, a1, span, colors[i%len(colors)])
					used[i] += span
				}
			}
		})
	})
}

// chordShare is how much of a member's arc a chord with one other member takes
// — the traffic between the two as a share of all the traffic through it.
func chordShare(from, to int, matrix [][]float64) float64 {
	var pair float64
	for j, v := range matrix[from] {
		if j != from {
			pair += v
		}
	}
	if pair <= 0 {
		return 0
	}
	return matrix[from][to] / pair
}

// drawChord is one flow between two members of a ring: a band from each
// member's arc to the other's, curving towards the middle.
func drawChord(p *ui.Painter, cx, cy, r, a0, a1, width float32, color ui.Color) {
	steps := 24
	var path ui.Path
	for i := range steps + 1 {
		at := a0 + (a1-a0)*float32(i)/float32(steps)
		x, y := point(cx, cy, r, at)
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	for i := steps; i >= 0; i-- {
		// The way back is a curve through the middle rather than round the
		// ring, which is what makes a chord a chord — and it comes back
		// the width of the flow away from where it went, so two chords
		// out of the same member do not lie on one another.
		at := a0 + (a1-a0)*float32(i)/float32(steps)
		x, y := point(cx, cy, r*0.35, at)
		path.LineTo(x, y)
	}
	for i := range steps + 1 {
		at := a0 + width + (a1-a0-width)*float32(i)/float32(steps)
		x, y := point(cx, cy, r, at)
		path.LineTo(x, y)
	}
	path.Close()
	p.FillPath(&path, color.Alpha(0.3))
}
