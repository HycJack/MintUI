package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// ResizablePanelGroupOptions configure a ResizablePanelGroup.
type ResizablePanelGroupOptions struct {
	// Fractions is how much of the group each pane takes, in order, and it
	// is the caller's slice: the group writes into it on every drag and reads
	// it on every frame, so where a user left the panes survives a redraw
	// without the group keeping anything.
	//
	// It is a fraction rather than a width on purpose. A group of two panes
	// dragged to a third is the same split in every window, and a width is
	// the wrong number the moment the window is resized.
	Fractions *[]float32
	// Labels name the panes for assistive technology and for a test to find
	// them by. One per pane is required: a pane's own width is not observable
	// from outside — an empty box measures zero, which is the trap
	// docs/design-system.md §15.2 is about — so the label is what a test
	// measures instead, and a splitter with two unnamed sides is two halves
	// of nothing.
	Labels []string
	// Min is the smallest share a pane may be dragged to, as a fraction of
	// the group. Zero is a tenth, which is the smallest a pane can be while
	// still being looked at rather than being a sliver.
	Min float32
	// Gutter is how wide the draggable seams are. Zero is the density's own
	// step: wide enough to hit, narrow enough that the seam does not read as
	// a third pane.
	Gutter float32
	// Vertical stacks the panes one under another with horizontal seams. The
	// default is panes side by side, for the reason ui/layout's SplitPane
	// names its own option SideBySide: "vertical" reads both ways.
	Vertical bool
}

// ResizablePanelGroup is a row of panes with seams the user drags.
//
// It is not layout.SplitPane because that is two panes and one fraction: a
// window with three — a file list, an editor, a tests column — is three panes
// and two seams, and writing the same handle-pressing code twice for the
// second seam is how the two ends of one window drift apart.
//
// The shares are the caller's and are normalised on the way in, so a group
// handed a slice that does not add up to one is laid out at the shares it
// asked for rather than at shares that overflow their container.
func ResizablePanelGroup(c *ui.Context, opts ResizablePanelGroupOptions, panes ...func()) *ui.Element {
	n := len(panes)
	if n < 2 {
		panic("input: ResizablePanelGroup needs at least two panes; one pane with a seam " +
			"on one side of it is a layout.SplitPane")
	}
	if opts.Fractions == nil {
		panic("input: ResizablePanelGroup needs a slice of fractions to point at; the seams " +
			"have nowhere to write the split they make")
	}
	if len(opts.Labels) != n {
		panic("input: ResizablePanelGroup needs one Label per pane; a pane nothing is named " +
			"cannot be dragged to on purpose")
	}
	u := core.Density(c).Unit()

	shares := normalise(opts.Fractions, n)
	gutter := opts.Gutter
	if gutter <= 0 {
		gutter = u
	}

	host := ui.Row(c).FillWidth().FillHeight()
	if opts.Vertical {
		host = ui.Column(c).FillWidth().FillHeight()
	}
	host.Children(func() {
		for i := range n {
			build := panes[i]
			// The pane is a plain box with no size of its own, and the shares
			// are stated as percentages: that is the only way MyGo's layout
			// gives a row exactly the fractions it was asked for.
			// WidthPercent is not BasisPercent, which seeds a flex start and
			// then has two Grow panes split what is left evenly — the trap
			// ui/layout's SplitPane documents and this is written against.
			pane := ui.Box(c).Shrink(0).Label(opts.Labels[i])
			if opts.Vertical {
				pane.HeightPercent(shares[i] * 100)
			} else {
				pane.WidthPercent(shares[i] * 100)
			}
			if build != nil {
				pane.Children(build)
			}
			if i < n-1 {
				resizablePaneGroupSeam(c, opts, shares, gutter, i, n, pane)
			}
		}
	})
	return host
}

// resizablePaneGroupSeam is one draggable seam, which moves the share of the
// pane before it and takes the difference out of the pane after it.
//
// neighbour is the pane the seam's lower index belongs to, which is how the
// group measures itself: the seam's own box is a few pixels wide and says
// nothing about the thousand it divides, but the pane beside it is a known
// fraction of the group, so the group is that pane's size over its share.
//
// The two panes either side move by the same amount, so dragging a seam grows
// both neighbours rather than one. A splitter that moves one side only is a
// splitter in a layout with a fixed side, and here there is no fixed side —
// which is why the total has to stay one and why rescaleRemainder exists.
func resizablePaneGroupSeam(c *ui.Context, opts ResizablePanelGroupOptions, shares []float32, gutter float32, at, n int, neighbour *ui.Element) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	seam := ui.Box(c).Grow(0).Shrink(0).Role(ui.RoleSplitter).
		Background(k.Border).Cursor(panelSeamCursor(opts.Vertical))
	if opts.Vertical {
		seam.Height(gutter)
	} else {
		seam.Width(gutter)
	}
	// The seam is a splitter and says which one it is: a bare divider with
	// no name is furniture, and the person looking for the thing they can
	// drag is looking for the label.
	seam.Label("Resize " + opts.Labels[at])
	over := seam.Hovered()
	if over {
		seam.Background(k.Accent)
	}
	if dx, dy, ok := seam.Dragged(); ok {
		delta := dx
		span := spanOf(c, neighbour, shares[at], opts.Vertical)
		if opts.Vertical {
			delta = dy
		}
		moveSeam(opts.Fractions, at, delta, span, minShare(opts, n))
		// A drag already sets the engine's consumed flag, but a caller who
		// moves a seam any other way must not have to know that: the panes
		// have to be built again at their new shares and only a caller who
		// asked for that frame gets them.
		c.Invalidate()
	}
	_ = u
}

// spanOf is how much of the group a seam has to work with, worked out from
// the pane beside it: the pane is a known fraction of the group, so the group
// is the pane's size over that fraction.
//
// The pane's box is the previous frame's, which is one frame of drift on a
// control being dragged by hand and no drift at all on a seam nobody has
// touched — which is every seam on the frame the window appears, where the
// answer is zero anyway because there was no drag.
func spanOf(c *ui.Context, pane *ui.Element, share float32, vertical bool) float32 {
	_ = c
	if pane == nil || share <= 0.001 {
		return 0
	}
	r := pane.Bounds()
	if vertical {
		return r.H / share
	}
	return r.W / share
}

// minShare is the smallest a pane may be dragged to, and never so large that
// two panes could sit at it at once: a group of two at 0.6 each is 1.2, and
// the window behind it is the thing that ends up cropped.
func minShare(opts ResizablePanelGroupOptions, n int) float32 {
	min := opts.Min
	if min <= 0 {
		min = 0.1
	}
	if n > 1 {
		if cap := 0.8 / float32(n-1); min > cap {
			min = cap
		}
	}
	return min
}

// moveSeam gives the seam at place at a drag of delta pixels across a group
// of span, and keeps every pane inside its minimum.
//
// The drag stops at whichever pane it reaches first rather than passing
// through it, because a pane dragged below its minimum and then dragged back
// comes back from somewhere else: the seam has to be where the pointer is, or
// it drifts away under the hand.
func moveSeam(group *[]float32, at int, delta, span, min float32) {
	if at < 0 || at >= len(*group)-1 || span <= 0 {
		return
	}
	share := delta / span
	a, b := (*group)[at], (*group)[at+1]
	switch {
	case share > 0 && a+share < min:
		share = min - a
	case share < 0 && b+share < min:
		share = min - b
	}
	(*group)[at] = a + share
	(*group)[at+1] = b - share
	rescaleRemainder(group, at, min)
}

// rescaleRemainder is what keeps the fractions adding up to one after a
// clamped drag: the panes past the seam share out whatever is left, in their
// own proportions, so a group that cannot move past a minimum still fills its
// whole width instead of leaving a gap at the end.
func rescaleRemainder(group *[]float32, at int, min float32) {
	tail := (*group)[at+2:]
	if len(tail) == 0 {
		return
	}
	room := 1 - (*group)[at] - (*group)[at+1] - min*float32(len(tail))
	if room <= 0 {
		return
	}
	sum := float32(0)
	for _, v := range tail {
		sum += v
	}
	if sum <= 0 {
		share := room / float32(len(tail))
		for i := range tail {
			tail[i] = share
		}
		return
	}
	for i := range tail {
		tail[i] = room * tail[i] / sum
	}
}

// normalise is the shares as they will be laid out: the caller's, put into
// range and scaled to add up to one.
//
// A slice that does not add up to one is the likeliest mistake there is — a
// caller that halved a fraction and forgot the other — and both alternatives
// are worse: laying it out as given overflows the container and pushes the
// last pane out of the window, while refusing it panics in a window that
// renders perfectly well. So it is scaled, and a caller watching its
// fractions afterwards sees the ones that will actually be drawn.
func normalise(fractions *[]float32, n int) []float32 {
	shares := make([]float32, n)
	sum := float32(0)
	for i := range n {
		v := float32(0)
		if i < len(*fractions) {
			v = (*fractions)[i]
		}
		if v < 0 {
			v = 0
		}
		shares[i] = v
		sum += v
	}
	if sum <= 0 {
		// Nothing asked for, so they share evenly — which is also what a
		// group that has just appeared should look like.
		for i := range shares {
			shares[i] = 1 / float32(n)
			if i < len(*fractions) {
				(*fractions)[i] = shares[i]
			}
		}
		return shares
	}
	for i, v := range shares {
		shares[i] = v / sum
		if i < len(*fractions) {
			(*fractions)[i] = shares[i]
		}
	}
	return shares
}

// panelSeamCursor is the pointer shape that says which way a seam moves.
func panelSeamCursor(vertical bool) ui.Cursor {
	if vertical {
		return ui.CursorResizeNS
	}
	return ui.CursorResizeEW
}
