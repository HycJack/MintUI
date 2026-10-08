package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// DataTableOptions configure a DataTable.
type DataTableOptions struct {
	// Columns are the columns, in the order they are shown. There must be at
	// least one: a table with no columns has no rows worth showing.
	Columns []Column
	// Rows is how many rows there are, in the order the caller has them.
	Rows int
	// Cell builds one cell, given a row in Rows and a column in Columns. It
	// is called for the rows in view and no others, so it may read the
	// record behind a row without a table's worth of records being read.
	Cell func(row, col int)
	// CellLabel names a cell for assistive technology, as "Hillside Dental:
	// $1,280". A cell of a bare figure says nothing when read out.
	CellLabel func(row, col int) string
	// Key is what identifies a row's record — an id, a path, anything
	// comparable. The rows' place and their choice follow their records as
	// a sort moves them, rather than following their numbers.
	Key func(row int) any
	// Label names a row for assistive technology, as "row 3: Maple Street
	// Bakery". A row with no name is a row nobody can hear.
	Label func(row int) string
	// Sort is the column the rows are ordered by, in the caller's state. A
	// click on the head of a sortable column writes it — the table asks, by
	// changing it — and the caller reorders Rows with it and asks again.
	Sort *Sort
	// Selected is the row the table marks as the one the keys move from, -1
	// for none. A click chooses it, and Up and Down move it.
	Selected *int
	// Choice is the set of chosen rows, when the table chooses several: a
	// click chooses one, Cmd-click adds one, and Shift-click takes the
	// range from the row last chosen to the one clicked.
	Choice *Selectable
	// State is the list's place and choice between frames; nil keeps them in
	// the table's own state. Give one to scroll a row into view, or to come
	// back to where a user was.
	State *ui.ListState
	// Scroll is where the table is scrolled sideways, when the caller keeps
	// one — a column dragged out of sight stays out of sight.
	Scroll *ui.ScrollState
	// Height is the height of the whole table, head and rows. It is
	// required: a table with no height grows to fit its rows, which is
	// every row drawn rather than a table.
	Height float32
	// Width is the width the rows are laid out in, which is what the
	// columns sharing the leftover room divide. Zero measures the window the
	// table is in, which is what a table filling a panel wants.
	Width float32
	// RowHeight is the least a row is high; zero is the library's own.
	RowHeight float32
	// NoRules draws no hairline between the rows, for a table whose rows
	// are separated by something of their own.
	NoRules bool
	// Empty draws instead of the rows when there are none. The place that
	// says "nothing here" is the caller's, because only it knows what the
	// table is of.
	Empty func()
}

// DataTableResult carries a DataTable and what the user did with it.
type DataTableResult struct {
	// Element is the whole table.
	Element *ui.Element
	// sorted is the column whose head was clicked this frame.
	sorted string
}

// Sorted returns the column whose head was clicked this frame, empty when
// none was. The caller reorders its rows with Rows and asks again: the
// table holds no order of its own, so what it shows is always what it was
// handed.
func (r DataTableResult) Sorted() string { return r.sorted }

// DataTable is a table of records under a head of column names: the rows
// scroll under a head that stays, only the rows in view are built, and the
// columns are as wide as they were given.
//
// Three rules make it a table rather than a grid of text, and each of them
// is a rule rather than a look:
//
//   - the columns keep their width. A column narrower than its content
//     shows names as letters and figures as dashes, so a window too narrow
//     for the columns scrolls sideways rather than squeezing them — the
//     same rule a board's columns follow, for the same reason.
//   - the head is above the scroll, not over it, so it cannot scroll away
//     and cannot be a frame behind the rows under it.
//   - the rows in view are the rows built, so a table of ten thousand
//     callbacks costs what a table of thirty does.
//
// Choosing is the caller's: Selected for the row the keys move from, Choice
// for a set of them, and Sort for the order. The table writes the pointers
// it is given and reports the head click; it keeps no order and no choice
// of its own, which is why what it shows is always what it was handed.
func DataTable(c *ui.Context, opts DataTableOptions) DataTableResult {
	u := core.Density(c).Unit()
	if len(opts.Columns) == 0 {
		panic("data: DataTable needs at least one column")
	}
	if opts.Cell == nil {
		panic("data: DataTable needs a Cell to build each cell with")
	}
	if opts.Rows < 0 {
		panic("data: DataTable cannot show a negative number of rows")
	}
	if opts.Sort != nil && opts.Sort.Column != "" {
		if _, ok := column(opts.Columns, opts.Sort.Column); !ok {
			panic("data: DataTable is sorted by " + opts.Sort.Column + ", which is not one of its columns")
		}
	}

	h := opts.RowHeight
	if h <= 0 {
		h = rowHeight(u)
	}
	state := opts.State
	if state != nil {
		if opts.Key != nil {
			state.Key = opts.Key
		}
		if opts.Label != nil {
			state.Label = opts.Label
		}
		// The keys move from the chosen row, so this is what gives the
		// table its Up, Down, Home and End.
		state.Selected = opts.Selected
	}

	var res DataTableResult
	res.Element = rowsWindow(c, windowOptions{
		rows: opts.Rows, height: opts.Height, width: opts.Width,
		rowH:    h,
		columns: opts.Columns,
		role:    ui.RoleTable,
		headH:   headerHeight(u),
		head: func(widths []float32) {
			headRow(c, opts.Columns, widths, func(col int) {
				headCell(c, opts.Columns, opts.Sort, col, &res.sorted)
			})
		},
		row:    func(i int, widths []float32) { tableRow(c, opts, widths, h, i) },
		state:  state,
		scroll: opts.Scroll,
		empty:  opts.Empty,
	})
	return res
}

// aligned is how a column's cells line up across it, with the middle and
// the far edge meaning the same thing in a cell (a box, aligned with
// AlignItems) as in a head cell (a row, placed with Justify).
func aligned(a ui.Align) ui.Align {
	if a == ui.Start {
		return ui.Start
	}
	return a
}

// column returns the column of that id.
func column(cols []Column, id string) (Column, bool) {
	for _, col := range cols {
		if col.key() == id {
			return col, true
		}
	}
	return Column{}, false
}

// headCell builds a column's head: its name, the arrow when the rows are
// ordered by it, and the press that asks to order them so.
//
// The head is built before the rows, so the click that changes the order is
// the same frame's answer to the order the rows were drawn in.
func headCell(c *ui.Context, cols []Column, sort *Sort, col int, sorted *string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	def := cols[col]
	id := def.key()

	// A head cell is a row, so what puts the name against the column's edge
	// is what places along it; a cell of the body is a box, where the same
	// thing is AlignItems. Two names for one idea in two shapes, which is
	// why aligned() is asked rather than decided at each site.
	// The head fills its column: a head cell that took the width of its own
	// name would be a control a third of the size of the column it sorts,
	// and the whole of a wide column's header would be dead.
	cell := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Role(ui.RoleNone).
		Justify(aligned(def.Align))
	if col < len(cols)-1 {
		cell.Padding(0, u, 0, 0)
	}
	if def.Sortable && sort != nil {
		// The head of a sortable column is a button, so that it takes the
		// keyboard and says so. A picture of a name is not a control.
		if cell.Clicked() {
			*sort = Toggle(*sort, id)
			*sorted = id
		}
		if sort.Column == id {
			cell.Background(k.SurfaceHover)
		}
	}
	cell.Children(func() {
		// The head is read before the rows, so it must not be read first as
		// body text: the smallest type in the table, in the muted tone, in
		// bold.
		name := ui.Text(c, def.Title).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Bold().SingleLine()

		if def.Align == ui.End {
			// A figure reads right against its edge, and the mark beside it
			// sits past the digits rather than among them.
			name.TextAlign(ui.End).Grow(1)
		}
		if sort != nil && sort.Column == id {
			sortMark(c, u*3.25, sort.Descending, k.TextMuted)
		}
	})
}

// tableRow builds one row: its surface, then its cells in the columns'
// widths, then whatever the caller's cell builder makes of them.
func tableRow(c *ui.Context, opts DataTableOptions, widths []float32, h float32, i int) {
	k := core.Tokens(c)
	chosen := opts.Choice != nil && opts.Choice.Has(i)
	lead := opts.Selected != nil && *opts.Selected == i

	// A row is a box and no taller than it has to be, so the cells line up
	// in columns however much text is in them.
	r := ui.Row(c).FillWidth().MinHeight(h).Shrink(0).AlignItems(ui.Stretch).
		Role(ui.RoleNone)
	if opts.Choice != nil {
		// A row the table paints itself is painted with ordinary text on the
		// accent of a chosen row, which is a light colour in both
		// appearances. A table that chooses one row lets the library's chosen
		// row stand, and the text that goes with it.
		r.TextColor(k.Text)
	}
	paintRow(c, r, chosen, lead)
	if opts.Choice != nil && r.Hovered() && !chosen {
		r.Background(k.SurfaceHover)
	}
	choice(c, r, opts.Choice, opts.Selected, i)
	r.Children(func() {
		for col := range opts.Columns {
			cell := ui.Box(c).Width(widths[col]).Shrink(0).Justify(ui.Center).
				AlignItems(aligned(opts.Columns[col].Align)).Role(ui.RoleCell)
			// A cell takes the whole of its column, so two columns that both
			// line up to the same edge meet with nothing between them: a
			// right-aligned Due runs straight into a right-aligned Cost and
			// the table reads as one column of run-together numbers. Every
			// cell but the last gives up a unit of room on its trailing
			// side, which is the gutter between columns.
			if col < len(opts.Columns)-1 {
				cell.Padding(0, core.Density(c).Unit()*2, 0, 0)
			}
			if opts.CellLabel != nil {
				cell.Label(opts.CellLabel(i, col))
			}
			cell.Children(func() { opts.Cell(i, col) })
		}
		if !opts.NoRules {
			// A hairline between the rows, and none around the table: the
			// head has its own, and a border around the whole of it would
			// say "this is a thing" over a table that is part of a page.
			ui.Box(c).FillWidth().Height(theme.BorderWidth).Shrink(0).
				Background(k.Border)
		}
	})
}
