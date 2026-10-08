package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ContextMenuOptions configure a ContextMenu.
type ContextMenuOptions struct {
	// Build fills the menu, one item at a time.
	//
	// It runs when the menu opens, and again in the frame after an item was
	// chosen — which is the only frame in which Chosen reports. The same
	// function is asked to describe the menu and to read the answer, so what
	// it draws is what it answers about.
	//
	// The menu is the desktop's, not the library's: the system draws it in
	// its own look on the platform's own terms. Panel has nothing to say
	// about it, and asking the desktop to wear a hairline would be asking it
	// to be something it is not.
	Build func(m *ui.Menu)
}

// ContextMenu gives anchor a menu for the secondary click — and, while the
// keyboard focus is on it or inside it, for the menu key or Shift+F10.
//
//	core.Use(c, core.Settings{})
//	ContextMenu(c, row, ContextMenuOptions{Build: func(m *ui.Menu) {
//	    if m.Item("Rename").Chosen() {
//	        app.renaming = i
//	    }
//	    m.Separator()
//	    if m.Item("Delete").Disabled(app.locked).Chosen() {
//	        app.delete(i)
//	    }
//	}})
//
// It is NonModal in the sense this package documents: a menu describes the
// page it was opened from, and it belongs to the desktop, which shows it over
// the page and hands the page back when it closes.
func ContextMenu(c *ui.Context, anchor *ui.Element, opts ContextMenuOptions) *ui.Element {
	if anchor == nil {
		panic("overlay: ContextMenu needs the element it opens on")
	}
	if opts.Build == nil {
		panic("overlay: ContextMenu needs a Build; an empty menu opens on every secondary click")
	}
	return anchor.ContextMenu(opts.Build)
}

// DropdownMenuOptions configure a DropdownMenu.
type DropdownMenuOptions struct {
	// Build fills the menu, on the same terms as ContextMenuOptions.Build.
	Build func(m *ui.Menu)
	// Primary fills the trigger with ink, for the one action a surface is
	// about where the menu is the rest of it.
	Primary bool
	// Disabled greys the trigger out: it takes neither clicks nor focus.
	Disabled bool
	// Icon draws before the label; nil leaves it out.
	Icon *ui.SVG
	// Label names the trigger for assistive technology when it has no text
	// of its own. A trigger with a label says enough and needs none.
	Label string
}

// DropdownMenu is a button that opens a menu under itself.
//
// It is NonModal in the sense this package documents, and for the reason a
// menu always is: the items are things to do to the page in front of you, so
// covering that page would make them harder to act on, not easier.
func DropdownMenu(c *ui.Context, label string, opts DropdownMenuOptions) *ui.Element {
	if label == "" && opts.Label == "" {
		panic("overlay: DropdownMenu needs a label or options.Label")
	}
	if opts.Build == nil {
		panic("overlay: DropdownMenu needs a Build; a trigger opening nothing is just a button")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := triggerInk(k, opts.Primary, false)
	name := label
	if name == "" {
		name = opts.Label
	}

	// input.Button's height, so a drop-down set among buttons is the height
	// of its neighbours rather than a different kind of control.
	trigger := ui.ButtonBase(c).
		Height(u*11).Radius(theme.PillRadius).
		Padding(0, u*4.5, 0, u*4.5).Background(bg).TextColor(fg).
		Label(name).Disabled(opts.Disabled).
		Children(func() {
			if opts.Icon != nil {
				ui.Icon(c, opts.Icon).TextColor(fg).Size(u*4.25, u*4.25)
			}
			if label != "" {
				ui.Text(c, label)
			}
			chevron(c, fg, u)
		})
	return trigger.Menu(opts.Build)
}

// triggerInk is the background and the label colour a menu trigger wears. It
// is shared so a drop-down and a split button's arrow cannot drift apart:
// they are the same control in two arrangements.
func triggerInk(k theme.Tokens, primary, danger bool) (bg, fg ui.Color) {
	bg, fg = k.Surface, k.Text
	switch {
	case primary:
		bg, fg = k.Fill, k.OnFill
	case danger:
		fg = k.Danger
	}
	return bg, fg
}

// PopupMenuItem is one row of a PopupMenuOpen.
type PopupMenuItem struct {
	// Label is what the row says, and is required: a row a person cannot read
	// is a row they cannot choose.
	Label string
	// Destructive draws the label in the danger colour, for the row that
	// destroys something — the same mark the danger buttons in this library
	// wear, so a menu and a button row cannot disagree about which answer is
	// the one that does not come back.
	Destructive bool
	// Disabled greys the row out: it takes neither clicks nor focus, and a
	// row that cannot be chosen is the state the caller's rules put it in.
	Disabled bool
}

// PopupMenuOpenOptions configure a PopupMenuOpen.
type PopupMenuOpenOptions struct {
	// Anchor is the element the menu hangs off, and is required: a menu that
	// is not about something on the page is a list.
	Anchor *ui.Element
	// Open is the *bool the menu opens and closes with, and is required. The
	// caller owns it, which is the difference from a DropdownMenu and a
	// SplitButton, whose triggers open their menus on their own terms: a
	// caller that opens the menu from somewhere else — a key, a gesture, a
	// row the trigger is not — owns the *bool and draws the body here.
	Open *bool
	// Items are the rows, in the order they should be read. At least one is
	// required: an empty menu is a panel with nothing to choose.
	Items []PopupMenuItem
	// Label names the menu; empty leaves it to the caller's anchor, as a
	// menu with no title of its own is one that describes what it hangs off.
	Label string
	// Modal puts a scrim over the window and sends outside presses to it, on
	// the terms PopoverOptions.Modal spells out. It is off by default, and
	// that default is the point: a menu is a list of things to do to the page
	// in front of you, and covering that page would make them harder to act
	// on, not easier.
	Modal bool
	// Width is the menu's width; zero lets it fit its rows.
	Width float32
}

// PopupMenuOpenResult carries a PopupMenuOpen and the row that was pressed.
type PopupMenuOpenResult struct {
	// Element is the menu, or nil while it is closed.
	Element *ui.Element
	chosen  int
}

// Chosen returns the index in PopupMenuOpenOptions.Items of the row pressed
// this frame, or -1 in a frame in which none was. Read it inside the view, as
// every report in this library is: the view runs up to three times per frame
// and the settled pass has no press.
func (r PopupMenuOpenResult) Chosen() int {
	if r.chosen < 0 {
		return -1
	}
	return r.chosen
}

// PopupMenuOpen is the body of a menu in the open state, drawn by the library
// rather than by the desktop: the surface, and the list of rows in it.
//
// The other menus in this file — a ContextMenu, a DropdownMenu, a SplitButton's
// arrow — are the desktop's: the system draws them in its own look on the
// platform's own terms, and they are the right choice while the trigger opens
// the menu itself. PopupMenuOpen is for the menu the caller opens on its own
// terms, from a place a trigger is not, and wears the look the menu in this
// file draws: the panel's hairline and radius, the labels at the row size, the
// danger colour where a row destroys.
//
// A press on a row reports its index through Chosen; a press outside the
// panel and Escape write false into the *bool, as a Popover does, because the
// menu hangs off the page the same way one does.
//
//	core.Use(c, core.Settings{})
//	open := false
//	if row.SecondaryClicked() {
//	    open = true
//	}
//	if res := PopupMenuOpen(c, PopupMenuOpenOptions{
//	    Anchor: row, Open: &open,
//	    Items: []PopupMenuItem{{Label: "Rename"}, {Label: "Delete", Destructive: true}},
//	}); res.Chosen() == 1 {
//	    app.delete(i)
//	}
func PopupMenuOpen(c *ui.Context, opts PopupMenuOpenOptions) PopupMenuOpenResult {
	if opts.Anchor == nil {
		panic("overlay: PopupMenuOpen needs the Anchor it hangs off; a menu not about something " +
			"on the page is a list")
	}
	if opts.Open == nil {
		panic("overlay: PopupMenuOpen needs the *bool it opens and closes with")
	}
	if len(opts.Items) == 0 {
		panic("overlay: PopupMenuOpen needs at least one item; an empty menu is a panel with " +
			"nothing to choose")
	}
	for _, item := range opts.Items {
		if item.Label == "" {
			panic("overlay: PopupMenuOpen has an item with no label; a row a person cannot read " +
				"is not a row")
		}
	}

	res := PopupMenuOpenResult{chosen: -1}
	res.Element = anchored(c, opts.Anchor, opts.Open, opts.Modal, func(p *ui.Element) {
		Panel(c, p, PanelOptions{
			Compact: true,
			Label:   opts.Label,
		}, func() {
			u := core.Density(c).Unit()
			ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
				for i, item := range opts.Items {
					menuRow(c, item, &res, i, u)
				}
			})
		})
	})
	if res.Element != nil && opts.Width > 0 {
		res.Element.Width(opts.Width)
	}
	return res
}

// menuRow is one row of a menu: the label at the row's size, in the ordinary
// tone or the danger one, taking a press unless the caller's rules put it in
// the disabled state.
func menuRow(c *ui.Context, item PopupMenuItem, res *PopupMenuOpenResult, index int, u float32) {
	k := core.Tokens(c)
	fg := k.Text
	switch {
	case item.Destructive:
		fg = k.Danger
	case item.Disabled:
		fg = k.TextFaint
	}
	row := ui.Row(c).FillWidth().Gap(u).Padding(u*0.75, u*1.5).
		Radius(theme.SmallRadius).Label(item.Label).Role(ui.RoleNone)
	if !item.Disabled {
		// Clicked is what marks the row as taking a press, and it is the
		// report: the row that was pressed is the one that got it.
		row.Cursor(ui.CursorPointer)
		if row.Hovered() {
			row.Background(k.SurfaceHover)
		}
		if row.Clicked() {
			res.chosen = index
		}
	}
	row.Children(func() {
		ui.Text(c, item.Label).TextColor(fg).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
	})
}

// chevron draws the arrow that says a button opens something underneath.
//
// It is drawn here rather than taken from MyGo's menu button because this
// package builds its triggers out of the library's own controls — so that the
// arrow, its size and its colour live in one place along with the panel and
// the button beside it, instead of arriving with a widget that brings its own
// padding and its own palette with it.
func chevron(c *ui.Context, col ui.Color, u float32) {
	ui.Box(c).Size(u*3, u*2).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		var path ui.Path
		path.MoveTo(r.X+r.W*0.15, r.Y+r.H*0.32).
			LineTo(r.X+r.W*0.5, r.Y+r.H*0.66).
			LineTo(r.X+r.W*0.85, r.Y+r.H*0.32)
		p.StrokePath(&path, 1.5, col)
	})
}
