package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// ListOptions configure a List.
type ListOptions struct {
	// Rows is how many there are; the row function builds each of them.
	Rows int
	// Height is the height of the list. It is required: a list with no height
	// grows to fit its rows, which is every row drawn rather than a list.
	Height float32
	// RowHeight is the height of a row. Zero is the library's own; a row
	// taller than that is as tall as what is in it.
	RowHeight float32
	// Gap is the space between rows; zero for rows that touch, as the rows
	// of a table do.
	Gap float32
	// Selected is the row the list marks as the one the keys move from, -1
	// for none. A click chooses it, and Up and Down move it.
	Selected *int
	// Choice is the set of chosen rows, when the list chooses several: a
	// click chooses one, Cmd-click adds one, and Shift-click takes the range
	// from the row last chosen to the one clicked.
	Choice *Selectable
	// Key identifies a row's record, so a row's place and its choice follow
	// the record rather than its number.
	Key func(row int) any
	// Label names a row for assistive technology.
	Label func(row int) string
	// State is the list's place between frames; nil keeps it in the list's
	// own. Give one to scroll a row into view, or to come back to where a
	// user was.
	State *ui.ListState
	// Scroll is where the list is scrolled, when the caller keeps one — a
	// row scrolled out of sight stays out of it.
	Scroll *ui.ScrollState
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// List is a column of rows of one height: chosen one at a time or several
// at a time, scrolling, and building only the rows in view.
//
// It is the plainest of the row views, and it is here for the same reason a
// table is: because a list of a thousand things should cost what a list of
// thirty does, and because a list and a table that choose rows differently
// are two things a user has to learn twice.
func List(c *ui.Context, opts ListOptions, row func(row int)) *ui.Element {
	u := core.Density(c).Unit()
	if row == nil {
		panic("data: List needs a row function")
	}
	if opts.Rows < 0 {
		panic("data: List cannot show a negative number of rows")
	}
	h := opts.RowHeight
	if h <= 0 {
		h = rowHeight(u)
	}
	if opts.State != nil {
		if opts.Key != nil {
			opts.State.Key = opts.Key
		}
		if opts.Label != nil {
			opts.State.Label = opts.Label
		}
		opts.State.Selected = opts.Selected
	}
	return rowsWindow(c, windowOptions{
		rows: opts.Rows, height: opts.Height, gap: opts.Gap,
		role:  ui.RoleList,
		state: opts.State, scroll: opts.Scroll, empty: opts.Empty,
		row: func(i int, _ []float32) { listRow(c, opts, u, h, i, row) },
	})
}

// listRow builds one row: its surface, then whatever the caller puts in it.
func listRow(c *ui.Context, opts ListOptions, u, h float32, i int, row func(row int)) {
	k := core.Tokens(c)
	chosen := opts.Choice != nil && opts.Choice.Has(i)
	lead := opts.Selected != nil && *opts.Selected == i
	r := ui.Row(c).FillWidth().MinHeight(h).Shrink(0).AlignItems(ui.Center).
		Gap(u*2).Padding(0, u*1.5).Role(ui.RoleListItem)
	if opts.Choice != nil {
		// See DataTable: a row this list paints itself is ordinary text on
		// the light accent of a chosen row; a row the library paints takes
		// the library's own text for it.
		r.TextColor(k.Text)
	}
	paintRow(c, r, chosen, lead)
	choice(c, r, opts.Choice, opts.Selected, i)
	r.Children(func() { row(i) })
}
