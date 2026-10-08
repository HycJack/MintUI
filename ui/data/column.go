package data

import (
	"github.com/egoist/mygo/ui"
)

// Column is one column of a DataTable or a TreeTable.
//
// Its width is the same rule the board's columns follow, and for the same
// reason: a number, or a share of a number the caller gives, that a narrow
// window cannot take away. A column narrower than its content shows a name
// as three letters and a figure as a dash, and a table of those cannot be
// read at all — so a table that does not fit scrolls sideways instead, and
// every column is as wide as it was in a window twice the size.
type Column struct {
	// Title heads the column. Empty is a column with no head, which is a
	// column of a row nobody is meant to name.
	Title string
	// ID names the column to the rest of the library: it is what Sort.Column
	// gives it, so the arrow over a header and the order of the rows cannot
	// name the same column two ways. Empty means Title, and an untitled
	// sortable column is one nothing can sort by.
	ID string
	// Width is the column's width in DIPs: a fixed number, whatever the
	// window does. Zero shares what the fixed columns leave, by Share.
	Width float32
	// Share is how many shares of the leftover room a shared column takes —
	// one is the default, two is twice what one takes. It says something
	// only when the table is laid out in a width: the shares are taken
	// against the width of the table, which the caller usually knows
	// because it is the panel it put the table in.
	Share float32
	// MinWidth is the least a shared column is squeezed to; zero is the
	// library's floor, about two words of the row's own text.
	MinWidth float32
	// Align lines the cells up: ui.End for a figure, which people compare
	// by their last digit, and the default for anything else.
	Align ui.Align
	// Sortable lets a click on the column's header ask for the rows ordered
	// by it. Asking for a column with no comparator in Rows is a panic there,
	// where the sort actually happens.
	Sortable bool
}

// key is how the rest of the library names the column.
func (col Column) key() string {
	if col.ID != "" {
		return col.ID
	}
	return col.Title
}

// floor is the least width a shared column is laid out in: the column's own
// MinWidth, or the width two of the row's words need. A column below it
// stops being able to show what is in it.
func (col Column) floor(least float32) float32 {
	return max(col.MinWidth, least)
}

// least is the width of two words of the row's own text, the floor of a
// column the caller gave no MinWidth: below it a name truncates to
// something that is no longer a name.
const least float32 = 60

// measure lays a table's columns out in a viewport table wide, and returns
// the width of each column and of the row they make. A viewport of 0 — the
// frame before the table has been measured — gives the width the columns
// need, which is the least a table is ever.
//
// The two rules are these, and they are the whole of the column-width story:
//
//   - a column with a Width keeps it in a window too narrow for it; the row
//     is then wider than the viewport and the table scrolls sideways
//   - a shared column takes its share of what the fixed columns leave, over
//     its floor, and never less than that floor
func measure(cols []Column, table float32) (widths []float32, row float32) {
	if len(cols) == 0 {
		// No columns: the rows are as wide as the window they are in.
		return nil, table
	}
	widths = make([]float32, len(cols))
	var fixed, shares, floors float32
	for i, col := range cols {
		if col.Width > 0 {
			widths[i], fixed = col.Width, fixed+col.Width
			continue
		}
		shares += max(col.Share, 1)
		floors += col.floor(least)
	}

	free := table - fixed - floors
	if table <= 0 || free < 0 {
		// No width yet, or the fixed columns alone already overrun the
		// viewport: the row is as wide as its columns need, which is what
		// makes the table scroll rather than squeeze.
		for i, col := range cols {
			if col.Width <= 0 {
				widths[i] = col.floor(least)
			}
		}
		return widths, fixed + floors
	}
	for i, col := range cols {
		if col.Width <= 0 {
			widths[i] = col.floor(least) + free*max(col.Share, 1)/shares
		}
	}
	return widths, table
}
