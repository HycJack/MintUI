package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Masonry lays items out in columns of their own heights, each item sitting on
// top of the shortest column so far.
//
// It is not a Grid and not a Wrap. A Grid gives every cell the same track, so
// a row is as tall as its tallest item and the page ends up with holes in it
// under the short ones; Wrap puts items side by side in the order they were
// given, so a tall item in the middle leaves a ragged edge for everything
// after it. Both are answers to "how many across", and this is an answer to
// "how do the holes disappear", which is the whole reason a masonry exists.
//
// The heights are the caller's, because only the caller knows them: a card's
// height is its title's line count plus its image's ratio, and a library that
// measured it would have to measure text.

// MasonryItem is one thing in a Masonry.
type MasonryItem struct {
	// Label names the item for assistive technology and for a test to find it
	// by. It is required: a masonry is a stack of boxes with no grid to say
	// what a row is, and an unnamed box in a stack is a box in a stack.
	Label string
	// Height is how tall the item is, in DIPs. It is what makes this a
	// masonry and not a grid, so it is not optional: an item of no stated
	// height would have to be measured, and only the caller can measure it.
	Height float32
}

// MasonryOptions configure a Masonry.
type MasonryOptions struct {
	// Columns is how many columns the items are packed into. Zero is as many
	// as fit at Column's width in the space there is, which is what a masonry
	// that reflows with the window wants; a fixed number is what one that
	// keeps its shape wants, and is what a caller whose items have to be
	// compared across a resize should give.
	Columns int
	// Column is how wide each column is; zero shares what the row gives it.
	Column float32
	// Gap is the space between items, on both axes.
	Gap float32
	// Width is the whole masonry's own width. It is required whenever Columns
	// is zero, because how many columns fit cannot be worked out from nothing
	// — and answering it from the width the parent happened to have last
	// frame is the reason a masonry reflows one frame late.
	Width float32
	// MaxHeight caps how far down the masonry goes; zero is no cap.
	MaxHeight float32
}

// Masonry is a set of items packed into columns, each on the shortest one.
//
// The packing is the greedy rule — shortest column wins — and it is written
// out rather than delegated to a flex layout because MyGo has no such layout
// and because the rule is the component: the same list of heights packed
// greedily always comes out the same, which is what lets a caller scroll and a
// test measure.
func Masonry(c *ui.Context, items []MasonryItem, opts MasonryOptions) *ui.Element {
	if len(items) == 0 {
		panic("input: Masonry needs at least one item; an empty masonry is a gap")
	}
	for i, it := range items {
		if it.Label == "" {
			panic("input: Masonry item " + itoa(i) + " has no Label; a masonry is a stack of " +
				"boxes with no grid to say what a row is, and an unnamed box in a stack is a " +
				"box in a stack")
		}
		if it.Height <= 0 {
			panic("input: Masonry item " + it.Label + " has no Height; an item of no stated " +
				"height would have to be measured, and only the caller can measure it")
		}
	}
	u := core.Density(c).Unit()

	gap := opts.Gap
	if gap == 0 {
		gap = u * 2
	}
	track := opts.Column
	columns := opts.Columns

	if track <= 0 {
		if opts.Width <= 0 {
			panic("input: Masonry needs a Width when it has no Column; how many columns fit " +
				"cannot be worked out from nothing")
		}
		// As many as fit, with at least one: a masonry narrower than its own
		// gap has no column left to put anything in, and a zero-column
		// masonry divides by nothing and drops every item on the floor.
		fit := int((opts.Width + gap) / (masonryDefaultWidth(c) + gap))
		if columns == 0 {
			columns = max(fit, 1)
		}
		track = (opts.Width - gap*float32(columns-1)) / float32(columns)
	}
	if columns < 1 {
		columns = 1
	}

	packed := packMasonry(items, columns)
	build := func() *ui.Element {
		// The row is made here rather than out here, because an element made
		// before its container is that container's sibling: a masonry built
		// as a sibling of its own columns is a column of nothing.
		row := ui.Row(c).AlignItems(ui.Start).Gap(gap)
		if opts.Width > 0 {
			row.Width(opts.Width).Shrink(0)
		}
		row.Children(func() { masonryColumns(c, packed, track) })
		return row
	}
	if opts.MaxHeight > 0 {
		// Reusing ui/layout's scroll area rather than a scroll of our own, so
		// a masonry inside a panel scrolls with the same thumb and the same
		// overscroll as every other long thing in this library.
		return layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Height: opts.MaxHeight,
		}, func() { build() }).Element
	}
	// The row, not nil: the caller sizes a masonry the way it sizes anything
	// else, and Masonry(...).Grow(1) on the nil this used to return was a nil
	// dereference.
	return build()
}

// masonryColumns builds the columns inside the host's own Children call,
// because an element made out here would be the caller's child and not this
// control's.
func masonryColumns(c *ui.Context, packed [][]MasonryItem, track float32) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	for _, column := range packed {
		col := ui.Column(c).Shrink(0).Gap(u * 2)
		if track > 0 {
			col.Width(track)
		}
		col.Children(func() {
			for _, it := range column {
				// The label is on the box and not only on the words inside it,
				// so what a reader is given and what a test measures is the
				// item rather than the width of the label in it.
				ui.Box(c).FillWidth().Height(it.Height).Shrink(0).
					Radius(theme.SmallRadius).Background(k.Surface).
					BorderWidth(theme.BorderWidth).BorderColor(k.Border).
					Label(it.Label).Role(ui.RoleNone).
					Children(func() {
						ui.Text(c, it.Label).SingleLine().TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize))
					})
			}
		})
	}
}

// packMasonry is the packing rule: every item goes on whichever column is
// currently the shortest, which is what makes the bottom of a masonry even.
//
// Ties go to the leftmost column, so the same list always packs the same way.
// A masonry that changed its layout when nothing changed would make every
// scroll position a guess.
func packMasonry(items []MasonryItem, columns int) [][]MasonryItem {
	if columns < 1 {
		columns = 1
	}
	heights := make([]float32, columns)
	packed := make([][]MasonryItem, columns)
	for _, it := range items {
		at := 0
		for i := 1; i < columns; i++ {
			if heights[i] < heights[at] {
				at = i
			}
		}
		packed[at] = append(packed[at], it)
		heights[at] += it.Height
	}
	return packed
}

// masonryDefaultWidth is how wide a column is when the caller gave neither a
// column nor a width: wide enough for a card of a title and two lines, which
// is the smallest a card of this library stops being a line of text.
func masonryDefaultWidth(c *ui.Context) float32 {
	return theme.ColumnWidth * 0.6
}
