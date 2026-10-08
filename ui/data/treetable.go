package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// TreeTableOptions configure a TreeTable. K is what an item of the tree is:
// something comparable, so that the open set can hold it.
type TreeTableOptions[K comparable] struct {
	// Roots are the items at the top, and Children says what is below an
	// item — nil for a leaf. Asking for the children of an item only as it
	// is opened is what keeps a tree of a hundred thousand rows cheap.
	Roots    []K
	Children func(item K) []K
	// Open holds the items showing their children. It is the caller's.
	Open *Open[K]
	// Columns are the columns, and Cell builds one cell of an item. The
	// first column is where the tree goes: the arrow and the depth of the
	// item are drawn in it, before what Cell makes of it.
	Columns []Column
	Cell    func(item K, col int)
	// CellLabel names a cell for assistive technology.
	CellLabel func(item K, col int) string
	// Label names an item for assistive technology.
	Label func(item K) string
	// Sort is the column the rows are ordered by, in the caller's state. A
	// click on the head of a sortable column writes it, and the caller
	// reorders its own items with Rows and asks again.
	Sort *Sort
	// Selected is the row the table marks as the one the keys move from.
	Selected *int
	// Choice is the set of chosen rows, when the table chooses several.
	Choice *Selectable
	// State is the list's place between frames; nil keeps it in the table's
	// own.
	State *ui.ListState
	// Scroll is where the table is scrolled sideways, when the caller keeps
	// one — a column dragged out of sight stays out of sight.
	Scroll *ui.ScrollState
	// Height is the height of the whole table, head and rows; it is
	// required. Width is the width the rows are laid out in, and zero
	// measures the window the table is in.
	Height    float32
	Width     float32
	RowHeight float32
	// Gap is the space between rows; zero for rows that touch.
	Gap float32
	// NoRules draws no hairline between the rows.
	NoRules bool
	// Indent is how far a level is in from the one above; zero is the
	// library's own.
	Indent float32
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// TreeTableResult carries a TreeTable and what the user did with it.
type TreeTableResult[K comparable] struct {
	// Element is the whole table.
	Element *ui.Element
	// sorted is the column whose head was clicked this frame.
	sorted string
	// toggled is the item whose arrow was pressed this frame, and was says
	// whether there was one.
	toggled K
	was     bool
}

// Sorted returns the column whose head was clicked this frame, empty when
// none was.
func (r TreeTableResult[K]) Sorted() string { return r.sorted }

// Toggled returns the item whose arrow was pressed this frame, and whether
// one was.
func (r TreeTableResult[K]) Toggled() (K, bool) { return r.toggled, r.was }

// TreeTable is a table whose rows are items in a hierarchy: the first
// column carries the arrow and the depth, and the rest are columns like any
// table's.
//
// It is a Tree and a DataTable at once because a hierarchy in a table is
// one thing, not two: the columns keep their widths, the head stays, only
// the rows in view are built, and choosing a row works the way it does in
// either.
func TreeTable[K comparable](c *ui.Context, opts TreeTableOptions[K]) TreeTableResult[K] {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if len(opts.Columns) == 0 {
		panic("data: TreeTable needs at least one column")
	}
	if opts.Cell == nil {
		panic("data: TreeTable needs a Cell to build each cell with")
	}
	if opts.Children == nil {
		panic("data: TreeTable needs a Children function to ask what is below an item")
	}
	if opts.Open == nil {
		panic("data: TreeTable needs the Open set to point at")
	}
	if opts.Sort != nil && opts.Sort.Column != "" {
		if _, ok := column(opts.Columns, opts.Sort.Column); !ok {
			panic("data: TreeTable is sorted by " + opts.Sort.Column + ", which is not one of its columns")
		}
	}
	h := opts.RowHeight
	if h <= 0 {
		h = rowHeight(u)
	}
	indent := opts.Indent
	if indent <= 0 {
		indent = u * 4
	}
	rows := branches(opts.Roots, opts.Children, opts.Open)

	if opts.State != nil {
		opts.State.Key = func(row int) any { return rows[row].item }
		if opts.Label != nil {
			opts.State.Label = func(row int) string { return opts.Label(rows[row].item) }
		}
		opts.State.Selected = opts.Selected
	}

	var res TreeTableResult[K]
	res.Element = rowsWindow(c, windowOptions{
		rows: len(rows), height: opts.Height, width: opts.Width,
		columns: opts.Columns, role: ui.RoleTree,
		head: func(widths []float32) {
			headRow(c, opts.Columns, widths, func(col int) {
				headCell(c, opts.Columns, opts.Sort, col, &res.sorted)
			})
		},
		row: func(i int, widths []float32) {
			b := rows[i]
			open := b.kids && opts.Open.Has(b.item)
			r := indentRow(c, opts.Selected, opts.Choice, i, u, h, ui.RoleTreeItem)
			choice(c, r, opts.Choice, opts.Selected, i)
			r.Children(func() {
				for col := range opts.Columns {
					cell := ui.Box(c).Width(widths[col]).Shrink(0).Justify(ui.Center).
						AlignItems(aligned(opts.Columns[col].Align)).Role(ui.RoleCell)
					if opts.CellLabel != nil {
						cell.Label(opts.CellLabel(b.item, col))
					}
					cell.Children(func() {
						if col > 0 {
							opts.Cell(b.item, col)
							return
						}
						// The tree goes in the first column, in a row
						// before what the cell says: the name of an item is
						// what its children hang off, and a name that is not
						// behind its own arrow is a list with some lines
						// missing.
						ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
							if b.depth > 0 {
								ui.Box(c).Width(float32(b.depth) * indent).Shrink(0)
							}
							if b.kids {
								mark, pressed := twisty(c, u*4, open)
								if pressed {
									setBranch(opts.Open, b.item, !open,
										mark.ClickModifiers()&ui.Alt != 0, opts.Children)
									res.toggled, res.was = b.item, true
								}
							}
							opts.Cell(b.item, col)
						})
					})
				}
				if !opts.NoRules {
					ui.Box(c).FillWidth().Height(theme.BorderWidth).Shrink(0).
						Background(k.Border)
				}
			})
		},
		state: opts.State, scroll: opts.Scroll, empty: opts.Empty,
	})
	return res
}
