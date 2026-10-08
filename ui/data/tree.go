package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
)

// Open is the set of items a Tree or a TreeTable is showing the children of.
//
// It is the caller's, and the tree writes to it: an arrow opens an item and
// the caller's set has the item in it, which is why a tree holding a set of
// its own would have no way of being told which items were open when the
// view was built again from scratch. Its zero value is closed all over, so
// a tree opens on nothing and the caller opens what it means to.
type Open[K comparable] struct {
	items map[K]bool
}

// Has reports whether an item is showing its children.
func (o *Open[K]) Has(item K) bool { return o.items[item] }

// Open puts an item's children on show.
func (o *Open[K]) Open(item K) {
	if o.items == nil {
		o.items = map[K]bool{}
	}
	o.items[item] = true
}

// Close puts an item's children away.
func (o *Open[K]) Close(item K) { delete(o.items, item) }

// Set opens or closes an item.
func (o *Open[K]) Set(item K, open bool) {
	if open {
		o.Open(item)
		return
	}
	o.Close(item)
}

// Clear closes everything.
func (o *Open[K]) Clear() { clear(o.items) }

// Len returns how many items are showing their children.
func (o *Open[K]) Len() int { return len(o.items) }

// branch is one row of a tree: the item, how deep it is, and whether it has
// children to open.
type branch[K comparable] struct {
	item  K
	depth int
	kids  bool
}

// branches lists the rows the tree shows: its items at the top, and below
// each open one its children, and so on.
//
// It is done every frame, from the caller's set, because the rows a tree
// shows are decided by what is open and nothing else — and because a set of
// open items is small next to the tree, which is what makes it cheap enough
// to do every frame at all.
func branches[K comparable](roots []K, children func(K) []K, open *Open[K]) []branch[K] {
	var rows []branch[K]
	var walk func(items []K, depth int)
	walk = func(items []K, depth int) {
		for _, item := range items {
			kids := children(item)
			rows = append(rows, branch[K]{item: item, depth: depth, kids: len(kids) > 0})
			if len(kids) > 0 && open.Has(item) {
				walk(kids, depth+1)
			}
		}
	}
	walk(roots, 0)
	return rows
}

// setBranch opens or closes an item, and with all it holds when all is set:
// every item below it that has children of its own.
func setBranch[K comparable](open *Open[K], item K, on, all bool, children func(K) []K) {
	open.Set(item, on)
	if !all {
		return
	}
	for _, kid := range children(item) {
		if len(children(kid)) > 0 {
			setBranch(open, kid, on, all, children)
		}
	}
}

// TreeOptions configure a Tree. K is what an item of the tree is: something
// comparable, so that the open set can hold it.
type TreeOptions[K comparable] struct {
	// Roots are the items at the top, and Children says what is below an
	// item — nil for a leaf. Asking for the children of an item only as it
	// is opened is what lets a tree of a hundred thousand items cost what a
	// small one does, so a tree that already holds its items should hand
	// them over as it is.
	Roots    []K
	Children func(item K) []K
	// Open holds the items showing their children. It is the caller's.
	Open *Open[K]
	// Row builds what an item's row says, given the item and how deep it is.
	Row func(item K, depth int)
	// Label names an item for assistive technology.
	Label func(item K) string
	// Selected is the row the tree marks as the one the keys move from, -1
	// for none. A click chooses it.
	Selected *int
	// Choice is the set of chosen rows, when the tree chooses several.
	Choice *Selectable
	// State is the list's place between frames; nil keeps it in the tree's
	// own. Give one to scroll a row into view, or to keep its place as the
	// items above it open and close.
	State *ui.ListState
	// Scroll is where the tree is scrolled, when the caller keeps one.
	Scroll    *ui.ScrollState
	Height    float32
	RowHeight float32
	Gap       float32
	// Indent is how far a level is in from the one above; zero is the
	// library's own.
	Indent float32
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// TreeResult carries a Tree and what the user did with it.
type TreeResult[K comparable] struct {
	// Element is the whole tree.
	Element *ui.Element
	// toggled is the item whose arrow was pressed this frame, and was says
	// whether there was one.
	toggled K
	was     bool
}

// Toggled returns the item whose arrow was pressed this frame, and whether
// one was: the item whose children to load, or to forget.
func (r TreeResult[K]) Toggled() (K, bool) { return r.toggled, r.was }

// Tree is items in a hierarchy: the items at the top, and below each open
// one its children, in from the side behind an arrow that opens and closes
// them. It builds only the rows in view, as the other row views do.
//
// What is open is the caller's — an Open set in the caller's state — so a
// tree rebuilt from scratch comes back as it was, and the tree holds nothing
// at all.
func Tree[K comparable](c *ui.Context, opts TreeOptions[K]) TreeResult[K] {
	u := core.Density(c).Unit()
	if opts.Children == nil {
		panic("data: Tree needs a Children function to ask what is below an item")
	}
	if opts.Open == nil {
		panic("data: Tree needs the Open set to point at")
	}
	if opts.Row == nil {
		panic("data: Tree needs a Row function to build each row with")
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

	var res TreeResult[K]
	res.Element = rowsWindow(c, windowOptions{
		rows: len(rows), height: opts.Height, gap: opts.Gap,
		role:  ui.RoleTree,
		state: opts.State, scroll: opts.Scroll, empty: opts.Empty,
		row: func(i int, _ []float32) {
			b := rows[i]
			open := b.kids && opts.Open.Has(b.item)
			r := indentRow(c, opts.Selected, opts.Choice, i, u, h, ui.RoleTreeItem)
			choice(c, r, opts.Choice, opts.Selected, i)
			// The space a level is in from the one above, and the arrow that
			// opens the item, are in the row before what the item says: the
			// name of an item is what its children hang off, and a tree whose
			// names start a line further right than their arrows reads as
			// one column too far.
			r.Children(func() {
				if b.depth > 0 {
					ui.Box(c).Width(float32(b.depth) * indent).Shrink(0)
				}
				if b.kids {
					mark, pressed := twisty(c, u*4, open)
					if pressed {
						// Alt opens or closes the whole of what is below,
						// which is the one thing a tree of thousands cannot
						// do by clicking its way down.
						setBranch(opts.Open, b.item, !open,
							mark.ClickModifiers()&ui.Alt != 0, opts.Children)
						res.toggled, res.was = b.item, true
					}
				}
			})
			r.Children(func() { opts.Row(b.item, b.depth) })
		},
	})
	return res
}

// indentRow is the box a row's own content is in: its surface, its height
// and the marks down its side, before anything is put in it.
func indentRow(c *ui.Context, selected *int, choice *Selectable, i int, u, h float32, role ui.Role) *ui.Element {
	k := core.Tokens(c)
	chosen := choice != nil && choice.Has(i)
	lead := selected != nil && *selected == i
	r := ui.Row(c).FillWidth().MinHeight(h).Shrink(0).AlignItems(ui.Center).
		Gap(u*1.5).Padding(0, u*1.5).Role(role)
	if choice != nil {
		// See DataTable: a row painted here is ordinary text on a light
		// accent; a row the library paints takes the library's own.
		r.TextColor(k.Text)
	}
	paintRow(c, r, chosen, lead)
	if choice != nil && r.Hovered() && !chosen {
		r.Background(k.SurfaceHover)
	}
	return r
}

// choice applies a click on a row to the set of chosen rows, in the way the
// modifiers of the click ask for.
func choice(c *ui.Context, r *ui.Element, chosen *Selectable, selected *int, row int) {
	if chosen == nil {
		return
	}
	if r.Clicked() {
		// Shift reaches from the row last chosen, not from the top of the
		// table: a range from the top picks rows the user never touched.
		mods := r.ClickModifiers()
		chosen.Click(row, mods&ui.Cmd != 0, mods&ui.Shift != 0)
		if selected != nil {
			*selected = row
		}
	}
}
