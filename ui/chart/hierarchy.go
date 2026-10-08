package chart

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The charts in this file draw parts of something whole: a treemap of areas,
// a sunburst of rings, a tree cut into rows, and a graph of nodes and the
// edges between them. None of them has a scale — the mark is the thing itself
// — so all of them work out their own geometry from the caller's tree or list,
// and all of them take that structure as it is given rather than rebuilding it.

// Node is one node of a tree: a name, a value, and what is under it.
//
// The value is the caller's rather than the sum of the children, because a
// node in a treemap and a node in a sunburst are asked for different things —
// "how much is this, including what it holds" and "how much of this does its
// own children hold" — and a caller that means the sum can say it.
type Node struct {
	// Label is the node's name, written inside it where there is room.
	Label string
	// Value is what the node is worth on its own scale.
	Value float64
	// Children are what it holds, in the order they are laid out.
	Children []Node
	// Color is what the node is drawn in; the zero colour takes the palette's
	// at its place in the set.
	Color ui.Color
}

// total is what a node and everything under it add up to, which is what a
// treemap lays out by: the parent, not the sum of the parts, so that a node's
// own slice is visible as a gap inside its parent's.
func total(n Node) float64 {
	value := n.Value
	if len(n.Children) > 0 {
		sum := 0.0
		for _, child := range n.Children {
			sum += total(child)
		}
		value = sum
	}
	return value
}

// ── treemap ────────────────────────────────────────────────────────────────

// Tile is one rectangle of a treemap: a name and what it is worth.
type Tile struct {
	// Label is the tile's name, written inside it where it fits.
	Label string
	// Value is what the tile's area is in proportion to.
	Value float64
	// Color is what it is drawn in; the zero colour takes the palette's at
	// its place in the set.
	Color ui.Color
}

// TreemapOptions configure a [Treemap].
type TreemapOptions struct {
	ChartOptions
	// Tiles are the rectangles. They are laid out biggest-first unless the
	// caller has put them in an order of their own — a treemap grouped by
	// parent keeps the caller's order, which is the only thing that makes
	// neighbouring tiles read as related.
	Tiles []Tile
	// Labels writes each tile's name inside it; off by default, because a
	// treemap of a hundred small tiles is a chart of a hundred numbers and
	// not a chart.
	Labels bool
}

// Treemap is one rectangle per part, each as nearly square as it can be, with
// its area in proportion to what it is worth: the shape of a part-to-whole
// that has too many parts for a pie and too little room for a bar chart.
func Treemap(c *ui.Context, opts TreemapOptions) *ui.Element {
	if len(opts.Tiles) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	values := make([]float64, len(opts.Tiles))
	colors := Palette(c, len(opts.Tiles))
	for i, t := range opts.Tiles {
		values[i] = t.Value
		if t.Color != (ui.Color{}) {
			colors[i] = t.Color
		}
	}
	k := core.Tokens(c)
	size := theme.CaptionSize
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Treemap", func(p *ui.Painter, plot ui.Rect) {
			// The gap is inside each tile rather than between them, so the
			// areas still say what they say: a treemap that shrinks its tiles
			// to leave gaps is a treemap of slightly wrong numbers.
			gap := unit(c) / 2
			boxes := Squarify(values, ui.Rect{X: plot.X + gap, Y: plot.Y + gap,
				W: maxf(plot.W-gap*2, 0), H: maxf(plot.H-gap*2, 0)})
			for i, box := range boxes {
				if box.W <= 0 || box.H <= 0 {
					continue
				}
				p.Fill(box, colors[i].Alpha(0.85), minf(theme.SmallRadius, box.W/6))
				if !opts.Labels || opts.Tiles[i].Label == "" {
					continue
				}
				label := opts.Tiles[i].Label
				w, h := LabelWidth(c, label, size), LabelHeight(c, size)
				// A name that does not fit inside its own tile is left to
				// the legend: written anyway it would sit on the tile
				// next door and name the wrong thing.
				if w <= box.W-unit(c)*2 && h <= box.H {
					p.Text(box.X+(box.W-w)/2, box.Y+(box.H-h)/2, label, size, k.OnFill)
				}
			}
		})
	})
}

// ── sunburst ───────────────────────────────────────────────────────────────

// SunburstOptions configure a [SunburstChart].
type SunburstOptions struct {
	ChartOptions
	// Root is the top of the tree. A chart with no root has no rings to
	// draw, so a root with neither a value nor any children is empty.
	Root Node
	// Depth is how many rings to draw; zero draws them all.
	Depth int
	// Hole is the inner radius as a share of the outer one; zero uses 0.2,
	// which is room for the total in the middle.
	Hole float32
}

// SunburstChart is a tree as rings: the root at the centre, its children round
// it, theirs round them, so that a part's share of the whole is an angle and
// its depth is a distance from the middle.
func SunburstChart(c *ui.Context, opts SunburstOptions) *ui.Element {
	if total(opts.Root) <= 0 {
		return nothing(c, opts.ChartOptions)
	}
	k := core.Tokens(c)
	hole := opts.Hole
	if hole <= 0 {
		hole = 0.2
	}
	depth := opts.Depth
	if depth < 1 {
		depth = depthOf(opts.Root)
	}
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Sunburst", func(p *ui.Painter, plot ui.Rect) {
			radius := minf(plot.W, plot.H) / 2
			if radius <= 0 {
				return
			}
			cx, cy := plot.X+plot.W/2, plot.Y+plot.H/2
			colors := Palette(c, len(opts.Root.Children))
			band := (radius - radius*hole) / float32(max(depth, 1))
			// The root in the middle, then a ring for each level under it:
			// the rings are the same width whatever is in them, so a depth
			// is a distance from the middle and not a share of anything.
			inner := radius * hole
			if depth > 1 {
				// The root's own slice, all the way round: every child
				// of it is inside it, so the ring is unbroken.
				full := wedge(cx, cy, inner, inner+band, 0, 360, 64)
				p.FillPath(&full, opts.Root.Color.Alpha(0.9))
				ringChildren(c, p, cx, cy, inner+band, band, depth-1, opts.Root, colors, k)
			}
			// The whole in the middle, which is the one number a sunburst
			// is usually read for.
			value := Group(opts.Root.Value, 0)
			w := LabelWidth(c, value, theme.StatSize)
			p.Text(cx-w/2, cy-LabelHeight(c, theme.StatSize)/2, value, theme.StatSize, k.Text)
		})
	})
}

// ringChildren lays one node's children out in the ring between inner and
// outer, each in proportion to what it is worth, and then does the same
// inside each of them.
func ringChildren(c *ui.Context, p *ui.Painter, cx, cy, inner, band float32, depth int, parent Node, colors []ui.Color, k theme.Tokens) {
	if depth <= 0 || len(parent.Children) == 0 {
		return
	}
	values := childValues(parent.Children)
	ends := Ends(values)
	at := float32(0)
	for i, child := range parent.Children {
		from, to := at, ends[i]
		at = to
		if to <= from {
			continue
		}
		color := child.Color
		if color == (ui.Color{}) && len(colors) > 0 {
			color = colors[i%len(colors)]
		}
		// A gap of most of a degree at each end, so that two neighbours
		// in a ring read as two things rather than as one long arc.
		ring2(p, cx, cy, inner, inner+band, from, to, 0.5, color.Alpha(0.9))
		if child.Label != "" && to-from > 14 {
			// Only where the slice is big enough for the name: a label
			// written across a two-degree arc lies on its side across
			// three other labels.
			x, y := point(cx, cy, inner+band/2, (from+to)/2)
			w := LabelWidth(c, child.Label, theme.CaptionSize)
			p.Text(x-w/2, y-LabelHeight(c, theme.CaptionSize)/2, child.Label,
				theme.CaptionSize, k.OnFill)
		}
		ringChildren(c, p, cx, cy, inner+band, band, depth-1, child, colors, k)
	}
}

// childValues is a node's children as numbers, which is what a ring is laid
// out by.
func childValues(children []Node) []float64 {
	out := make([]float64, len(children))
	for i, c := range children {
		out[i] = total(c)
	}
	return out
}

// depthOf is how many rings a tree needs: one for the root and one for each
// level under it.
func depthOf(n Node) int {
	if len(n.Children) == 0 {
		return 1
	}
	most := 0
	for _, child := range n.Children {
		most = max(most, depthOf(child))
	}
	return most + 1
}

// ring2 fills one segment of a ring: the band between two radii, from one
// angle to another, a gap at each end unless it runs all the way round.
func ring2(p *ui.Painter, cx, cy, inner, outer, from, to, gap float32, color ui.Color) {
	if to <= from || outer <= inner {
		return
	}
	path := wedge(cx, cy, inner, outer, from+gap, to-gap, clampSteps(int(to-from)))
	p.FillPath(&path, color)
}

// ── decomposition tree ─────────────────────────────────────────────────────

// DecompositionOptions configure a [DecompositionTree].
type DecompositionOptions struct {
	ChartOptions
	// Root is the top of the tree, cut into rows: the root along the top, its
	// children under it, and so on.
	Root Node
	// Depth is how many rows to draw; zero draws them all.
	Depth int
}

// DecompositionTree is a tree as bands across the plot, one row per level: the
// same data as a sunburst, laid out so that the rows can be read left to right
// and each level compared against the one above it.
//
// It is the treemap's answer for data that nests — each row's bands line up
// with the parent above them, which a ring cannot show and a row of aggregates
// cannot either.
func DecompositionTree(c *ui.Context, opts DecompositionOptions) *ui.Element {
	if total(opts.Root) <= 0 {
		return nothing(c, opts.ChartOptions)
	}
	k := core.Tokens(c)
	size := theme.CaptionSize
	depth := opts.Depth
	if depth < 1 {
		depth = depthOf(opts.Root)
	}
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Decomposition tree", func(p *ui.Painter, plot ui.Rect) {
			colors := Palette(c, len(opts.Root.Children))
			rowHeight := plot.H / float32(max(depth, 1))
			gap := unit(c) / 2
			// The root is written rather than drawn: it is the whole of what
			// is below it, and a band for it would say nothing.
			head := float32(0)
			if opts.Root.Label != "" {
				head = LabelHeight(c, theme.RowSize) + unit(c)
				p.Text(plot.X, plot.Y, opts.Root.Label, theme.RowSize, k.Text)
			}
			rows(c, p, plot.X, plot.Y+head, plot.W, plot.H-head, rowHeight, gap, depth,
				opts.Root, colors, k, size)
		})
	})
}

// rows draws one level of the tree and then draws the levels beneath it,
// each inside the band of the parent it belongs to — which is what lines the
// rows up with each other.
func rows(c *ui.Context, p *ui.Painter, left, top, width, height, rowHeight, gap float32, depth int,
	parent Node, colors []ui.Color, k theme.Tokens, size float32) {
	if depth <= 0 || len(parent.Children) == 0 || width <= 0 {
		return
	}
	values := childValues(parent.Children)
	grand := 0.0
	for _, v := range values {
		grand += v
	}
	if grand <= 0 {
		return
	}
	usable := width - gap*float32(len(values)-1)
	at := left
	for i, child := range parent.Children {
		w := usable * float32(values[i]/grand)
		color := child.Color
		if color == (ui.Color{}) && len(colors) > 0 {
			color = colors[i%len(colors)]
		}
		band := ui.Rect{X: at + gap/2, Y: top, W: maxf(w-gap, 1), H: rowHeight * 0.62}
		p.Fill(band, color.Alpha(0.85), minf(theme.SmallRadius, band.H/4))
		if child.Label != "" {
			lw, lh := LabelWidth(c, child.Label, size), LabelHeight(c, size)
			if lw <= band.W-unit(c) && lh <= band.H {
				p.Text(band.X+(band.W-lw)/2, band.Y+(band.H-lh)/2, child.Label, size, k.OnFill)
			}
		}
		rows(c, p, band.X, band.Y+rowHeight*0.75, band.W, height-rowHeight*0.75, rowHeight, gap,
			depth-1, child, colors, k, size)
		at += w
	}
}
