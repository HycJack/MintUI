package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// This file holds the frame the components in this package that show rows
// sit on: a head that stays where it is while the rows scroll under it, and
// a list that builds only the rows in view.
//
// A table of thousands of callbacks cannot draw thousands of rows, and the
// two ways of not drawing them are both wrong in their own way. Drawing
// every row and letting a viewport crop them costs the frame everything
// below the crop; drawing the viewport's worth by hand means the rows shown
// are the rows of the frame before, so scrolling shows blank rows for a
// frame. MyGo's List is built for this: it keeps its place by a row rather
// than by an offset into its content, so it builds the rows its place shows
// and measures them as it lays them out.

// windowOptions configure rowsWindow.
type windowOptions struct {
	// rows is how many there are; row builds the ones in view.
	rows int
	row  func(row int, widths []float32)
	// height is the height of the whole window, head and rows together. It
	// is required: a window with no height grows to fit its rows and never
	// scrolls, which is every row drawn rather than a window.
	height float32
	// width is the width the rows are laid out in when the caller knows it;
	// zero measures the window instead.
	width float32
	// columns are the columns the rows are made of, and each row builder is
	// told how wide each of them is. Nil for a window of rows with no
	// columns, which then fill it.
	columns []Column
	// head builds the row of column names above the rows, and headH is how
	// high it is. A nil head is a window of rows with nothing above them.
	head  func(widths []float32)
	headH float32
	// role is what assistive technology is told the window is: a table, a
	// tree or a list. The list the rows are built in says nothing itself —
	// it is the machinery, not a level of the thing the caller is showing.
	role ui.Role
	gap  float32
	// rowH is the height of one row, when the rows are all the same height.
	// The window then shows whole rows only: a viewport that ends mid-row
	// offers half a record a scroll would finish, which reads as a drawing
	// mistake rather than as somewhere to scroll.
	rowH float32
	// state is the list's place and its choice between frames; nil keeps
	// them in the table's own.
	state *ui.ListState
	// scroll is where the window is scrolled sideways, when the caller keeps
	// one to restore where a user was.
	scroll *ui.ScrollState
	// empty draws instead of the rows when there are none.
	empty func()
}

// rowsWindow is rows in a window: a head that does not scroll away, and a
// list that builds only the rows in view, scrolling down and — when the
// columns are wider than the window — sideways.
//
// The head is a sibling of the scroller rather than a layer over it. A layer
// would be pinned over the rows and would have to be moved by hand as they
// scrolled, a frame behind them; as a sibling it is above them, outside
// their scroll, and cannot move at all.
func rowsWindow(c *ui.Context, o windowOptions) *ui.Element {
	k := core.Tokens(c)
	if o.height <= 0 {
		panic("data: a rows window needs a Height; without one it cannot scroll")
	}
	if o.rows <= 0 {
		// Not a scroll view: there is nothing to scroll, and a control
		// inside a scroll view is hard to press.
		e := ui.Box(c).FillWidth().Height(o.height).Background(k.Background)
		if o.empty != nil {
			e.Children(o.empty)
		}
		return e
	}
	body := o.height - o.headH
	if body <= 0 {
		panic("data: a rows window needs room for a row under its head")
	}
	if o.rowH > 0 {
		// Whole rows only: a viewport that stops mid-row leaves half a
		// record hanging at its edge.
		body -= float32(int(body) % int(o.rowH))
	}

	view := layout.ScrollArea(c, layout.ScrollAreaOptions{
		Vertical: true, Horizontal: true, Height: o.height, State: o.scroll,
	}, nil).Element.FillWidth()

	// The width the rows are laid out in, which is the one rule that makes a
	// table a table: a scroll viewport hands its content its own width and
	// never more, so a window wider than its window has to be told how wide
	// it is — otherwise every column is squeezed into the window and the
	// table stops scrolling sideways, which is the one thing that would have
	// made it readable.
	//
	// The caller's width is the first answer. Failing that, the viewport's
	// own width is the one it had in the frame before, which is nothing on
	// the frame a window appears on: that frame lays the rows out at the
	// width their columns need and asks for the frame that measures the
	// viewport, or a window in a panel wider than its columns would stay at
	// their width until something else happened to repaint it. A window in
	// a panel with no width asks every frame, which costs a frame of
	// nothing drawn: it is the one case where measuring never ends.
	table := o.width
	if table <= 0 {
		table = view.Bounds().W
		if table <= 0 {
			c.Invalidate()
		}
	}
	widths, row := measure(o.columns, table)

	view.Children(func() {
		// The width of the content is said outright: a scroll window hands
		// its content what fits rather than what the content asked for, so
		// a column that asked for the window's width with a percentage would
		// be given the width of its own rows and a table that asked to fill
		// its panel would be as narrow as its narrowest column.
		col := ui.Column(c).Height(o.height)
		if row > 0 {
			col.Width(row).Shrink(0)
		} else {
			col.FillWidth()
		}
		col.Children(func() {
			if o.head != nil {
				o.head(widths)
			}
			rows := ui.List(c, o.state, o.rows, func(i int) { o.row(i, widths) }).
				Height(body).Shrink(0).Role(ui.RoleNone)
			if row > 0 {
				rows.Width(row)
			} else {
				rows.FillWidth()
			}
			if o.gap > 0 {
				rows.Gap(o.gap)
			}
		})
	})
	// What the caller is showing is a table or a tree or a list; the list the
	// rows are built in is how they are built, not a level of it.
	if o.role != ui.RoleNone {
		view.Role(o.role)
	}
	return view
}

// headerHeight is the height of a table's head: a row of column names and
// the hairline under it.
func headerHeight(u float32) float32 { return u*7 + theme.BorderWidth }

// rowHeight is the height of a table's row, and of a list's: one line of
// body text with the space above and below it that makes it read as a row
// rather than as a line of text.
func rowHeight(u float32) float32 { return u * 12 }

// headRow lays a table's head out in its columns' widths and draws the
// hairline under it. cell builds what a column's head shows, which is the
// caller's because only it knows whether the column sorts.
//
// Each head cell is a row so that what it holds is centred in it: a name
// beside a mark, in a box a fixed height, with nothing telling the box where
// to put them otherwise.
func headRow(c *ui.Context, cols []Column, widths []float32, cell func(col int)) *ui.Element {
	u := core.Density(c).Unit()
	return ui.Column(c).FillWidth().Shrink(0).Children(func() {
		// Stretch, not the Row's own centring: a head cell is as tall as
		// the head, so that the line under it runs under the whole of it.
		ui.Row(c).FillWidth().Height(headerHeight(u)).Shrink(0).
			AlignItems(ui.Stretch).Role(ui.RoleRow).Children(func() {
			for col := range widths {
				// The name is on the cell rather than on the text in it,
				// so what is announced for the column is the column, and
				// what a test measures is the column's width rather than
				// the width of the word in it.
				ui.Row(c).Width(widths[col]).Shrink(0).AlignItems(ui.Center).Clip().
					Padding(0, u*2).Role(ui.RoleColumnHeader).
					Label(cols[col].Title).
					Children(func() { cell(col) })
			}
		})
		layout.Divider(c, layout.DividerOptions{})
	})
}
