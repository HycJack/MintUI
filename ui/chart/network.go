package chart

import (
	"math"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The network graph is the one chart here with no data of its own: the caller
// hands over nodes and the edges between them and the chart works out where
// to put them. It lays them out rather than running a simulation, because a
// layout that moves when the window repaints is a chart that cannot be read
// twice the same way — and because a caller's node order is a grouping the
// caller already knows about.

// GraphNode is one node of a network: what it is called, how big it is, and
// which group of the network it belongs to.
type GraphNode struct {
	// Label is the node's name, written inside it where it fits and outside
	// it where it does not.
	Label string
	// Value is what sizes the mark; the largest node's value sets the rest.
	Value float64
	// Group is which part of the network the node is in. Nodes of one group
	// are laid out together in one arc of the ring, so that a community is
	// something a reader can see rather than something the colours imply.
	Group int
}

// GraphLink is one edge of a network, from node Source to node Target.
type GraphLink struct {
	// Source and Target are the two nodes' indices.
	Source, Target int
	// Value is how much runs along the edge, which is what its width is.
	Value float64
}

// NetworkOptions configure a [NetworkGraph].
type NetworkOptions struct {
	ChartOptions
	// Nodes are the things. Their order is kept within each group, biggest
	// first, so that the caller can put the ones that matter at the front.
	Nodes []GraphNode
	// Links are the edges between them.
	Links []GraphLink
	// Radius is the radius of the ring the nodes are laid out on; zero takes
	// as much of the plot as the marks and their names leave.
	Radius float32
	// Curvature bends the edges towards the middle, which is what stops a
	// dense network being a disc of straight lines nobody can follow.
	Curvature float32
}

// NetworkGraph is a set of nodes laid out in a ring with the edges between
// them: a picture of how things are related rather than of how much of
// anything there is.
func NetworkGraph(c *ui.Context, opts NetworkOptions) *ui.Element {
	if len(opts.Nodes) == 0 {
		return nothing(c, opts.ChartOptions)
	}
	layout := layOut(opts.Nodes)
	edges := edgesOf(opts.Nodes, opts.Links)
	k := core.Tokens(c)
	return panel(c, opts.ChartOptions, func(f FrameResult) {
		marksBox(c, f, "Network", func(p *ui.Painter, plot ui.Rect) {
			widest := float32(0)
			for _, n := range layout {
				widest = maxf(widest, n.r)
			}
			radius := opts.Radius
			if radius <= 0 {
				radius = minf(plot.W, plot.H)/2 - widest*2 - ringRoom(c, nodeNames(layout))
			}
			if radius <= 0 {
				return
			}
			cx, cy := plot.X+plot.W/2, plot.Y+plot.H/2
			curvature := opts.Curvature
			if curvature == 0 {
				curvature = 0.25
			}
			groups := groupCount(opts.Nodes)
			colors := Palette(c, groups)
			placed := onRing(layout, cx, cy, radius)
			// The edges first, so that a node is never hidden behind the
			// traffic going into it.
			for _, e := range edges {
				from, to := placed[e.source], placed[e.target]
				width := maxf(unit(c)*0.75*e.weight, theme.BorderWidth)
				curve := ui.Path{}
				curve.MoveTo(from.x, from.y)
				curve.CubeTo(
					from.x+(cx-from.x)*curvature, from.y+(cy-from.y)*curvature,
					to.x+(cx-to.x)*curvature, to.y+(cy-to.y)*curvature,
					to.x, to.y)
				p.StrokePath(&curve, width, colors[e.group%len(colors)].Alpha(0.45))
			}
			for _, n := range placed {
				internal.Dot(p, n.x, n.y, n.r, colors[n.group%len(colors)].Alpha(0.9))
				ringAround(p, n.x, n.y, n.r, k.Background)
				if n.label == "" {
					continue
				}
				size := theme.CaptionSize
				w, h := LabelWidth(c, n.label, size), LabelHeight(c, size)
				switch {
				case w <= n.r*2 && h <= n.r*2:
					p.Text(n.x-w/2, n.y-h/2, n.label, size, k.OnFill)
				default:
					// Outside the ring, on whichever side its own node is
					// on, so that two names never sit on top of one
					// another in the middle of the circle.
					x, y := point(cx, cy, radius+n.r+unit(c), n.angle)
					if x < cx {
						p.Text(x-w-unit(c), y-h/2, n.label, size, k.TextMuted)
					} else {
						p.Text(x+unit(c), y-h/2, n.label, size, k.TextMuted)
					}
				}
			}
		})
	})
}

// node is a node with its place on the ring worked out, but not yet its
// place in the plot — which cannot be known until the frame has measured it.
type node struct {
	group int
	value float64
	label string
	angle float32
	x, y  float32
	r     float32
}

// layOut is a network's nodes ordered for the ring: grouped, and with every
// mark sized against the largest value there is.
func layOut(nodes []GraphNode) []node {
	widest := 0.0
	for _, n := range nodes {
		widest = maxf64(widest, n.Value)
	}
	if widest <= 0 {
		widest = 1
	}
	out := make([]node, len(nodes))
	for i, n := range nodes {
		// The root, because the area of a circle is what a reader
		// compares: a node drawn twice the radius looks four times as
		// important, and it is twice.
		out[i] = node{group: n.Group, value: n.Value, label: n.Label,
			r: minNodeRadius + nodeRadiusRange*float32(math.Sqrt(max(n.Value, 0)/widest))}
	}
	// Each group's share of the ring is its share of the nodes, so a big
	// community takes more of the circle instead of being squeezed into
	// the same arc as a small one.
	groups := map[int][]int{}
	order := make([]int, 0, len(nodes))
	for i, n := range out {
		if _, seen := groups[n.group]; !seen {
			order = append(order, n.group)
		}
		groups[n.group] = append(groups[n.group], i)
	}
	at := float32(0)
	for _, g := range order {
		step := 360 / float32(len(out))
		for _, i := range groups[g] {
			out[i].angle = at + step/2
			at += step
		}
	}
	return out
}

// The marks' size: the smallest any node is drawn, and how much room the
// biggest one may take, both in DIPs off the window's own spacing scale.
const (
	minNodeRadius   float32 = 4
	nodeRadiusRange float32 = 20
)

// onRing is the ring itself: every node's own place in the plot, which cannot
// be worked out until the frame has measured the plot area.
func onRing(in []node, cx, cy, radius float32) []node {
	out := make([]node, len(in))
	copy(out, in)
	for i, n := range out {
		out[i].x, out[i].y = point(cx, cy, radius, n.angle)
	}
	return out
}

// edge is a link with the layout's own numbering.
type edge struct {
	source, target, group int
	weight                float32
}

// edgesOf is a network's links as edges, with an index of the group the edge
// belongs to: the source's, because an edge is seen leaving somewhere.
func edgesOf(nodes []GraphNode, links []GraphLink) []edge {
	out := make([]edge, 0, len(links))
	for _, l := range links {
		if l.Source < 0 || l.Source >= len(nodes) || l.Target < 0 || l.Target >= len(nodes) {
			panic("chart: a network link points at a node that is not there")
		}
		out = append(out, edge{source: l.Source, target: l.Target,
			group: nodes[l.Source].Group, weight: float32(l.Value)})
	}
	return out
}

// nodeNames is a laid-out network's names, which is what the ring has to be
// small enough to leave room for.
func nodeNames(in []node) []string {
	out := make([]string, 0, len(in))
	for _, n := range in {
		out = append(out, n.label)
	}
	return out
}

// groupCount is how many groups a network has, which is how many colours its
// nodes are spread across.
func groupCount(nodes []GraphNode) int {
	seen := map[int]bool{}
	for _, n := range nodes {
		seen[n.Group] = true
	}
	return max(len(seen), 1)
}
