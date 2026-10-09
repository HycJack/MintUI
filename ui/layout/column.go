package layout

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ColumnOptions configure a Column.
type ColumnOptions[T any] struct {
	// Title is the column's name, shown at its head.
	Title string
	// Count is how many items the column holds in all, which may be more
	// than its body shows. Nil draws no count.
	Count *int
	// Menu names a trailing button; nil draws none.
	Menu string
	// More is the label of the button that reveals the rest. Nil draws none.
	More string
	// Empty draws instead of the body when there is nothing to scroll — the
	// place a column keeps its own explanation, so the layout around it does
	// not have to know what absence looks like.
	Empty func()
}

// ColumnResult carries a Column and what the user did with it. T is the type
// of value the column takes in a drop.
type ColumnResult[T any] struct {
	// Element is the whole column.
	Element *ui.Element
	// Revealed reports a press of the More button.
	revealed bool
	// Dropped is the value dropped onto the column's body this frame.
	dropped T
	// got records a drop: T need not be comparable, so a zero test would not
	// compile for a slice or map payload.
	got bool
	// over is whether a drag is hovering it.
	over bool
}

// Revealed reports a press of the More button.
func (r ColumnResult[T]) Revealed() bool { return r.revealed }

// Dropped returns the value dropped onto the column, and whether one was.
func (r ColumnResult[T]) Dropped() (T, bool) { return r.dropped, r.got }

// Hovered reports a drag hovering the column, which is the moment to show
// that it would take the drop.
func (r ColumnResult[T]) Hovered() bool { return r.over }

// Column is one lane of a board: a head naming it, a body that scrolls, and
// a button that reveals what the body is not showing.
//
// The body is not wrapped in a scroll view when Empty is drawing, because
// there is nothing to scroll and a button inside a scroll view is hard to
// press.
//
// A body wrapped in a scroll view of its own must be given a definite height,
// or it collapses to nothing — see the package note on nesting scrolls.
func Column[T any](c *ui.Context, opts ColumnOptions[T], body func()) ColumnResult[T] {
	k, u := core.Tokens(c), core.Density(c).Unit()
	var r ColumnResult[T]

	col := ui.Column(c).Width(theme.ColumnWidth).FillHeight().Background(k.Surface).
		Radius(theme.PanelRadius).Padding(u*4, u*3.5, u*4, u*3.5).Gap(u * 3)

	if _, over := ui.DragOver[T](col); over {
		r.over = true
		col.Border(2, k.TextMuted)
	}
	if dropped, ok := ui.Drop[T](col); ok {
		r.dropped, r.got = dropped, true
	}

	col.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Padding(u/2, u, 0, u).Children(func() {
			ui.Text(c, opts.Title).TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize)).Bold().
				Grow(1).SingleLine()
			if opts.Count != nil {
				ui.Box(c).Padding(u*0.75, u*2.25, u*0.75, u*2.25).Radius(theme.PillRadius).
					Background(k.Border).Children(func() {
					ui.Text(c, itoa(*opts.Count)).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
				})
			}
			if opts.Menu != "" {
				ui.Box(c).Size(u*9.5, u*9.5).Radius(u * 4.75).Background(k.Background).
					Center().Label(opts.Menu).Children(func() {
					ui.Text(c, "⋯").TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize))
				})
			}
		})

		if opts.Empty != nil {
			opts.Empty()
		} else {
			ui.Scroll(c).FillWidth().Grow(1).Children(body)
		}

		if opts.More != "" {
			more := ui.Row(c).FillWidth().Height(u * 12).Radius(theme.CardRadius).
				Background(k.Border).Center().Label(opts.More).Children(func() {
				ui.Text(c, opts.More).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			})
			r.revealed = more.Clicked()
		}
	})

	r.Element = col
	return r
}

// Board lays columns out in a row that scrolls sideways rather than
// shrinking them. Columns below theme.ColumnWidth wrap their titles, and a
// column of wrapped titles reads as a wall of text.
//
// The row is given a height because the columns fill their parent: a
// horizontal scroll view passes its content whatever height it likes, and a
// child asking to fill it gets nothing.
func Board(c *ui.Context, columns func()) *ui.Element {
	u := core.Density(c).Unit()
	return ui.ScrollHorizontal(c).Fill().Grow(1).Children(func() {
		ui.Row(c).FillHeight().Gap(u*4).Padding(u*4, u*7, u*5.5, u*7).
			Children(columns)
	})
}
