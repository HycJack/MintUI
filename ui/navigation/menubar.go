package navigation

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// MenuItem is one row of a menu.
type MenuItem struct {
	// Label is what the row says, and what NavigationMenu's caller gets
	// back — a desktop menu hands over "this item ran" and nothing else, so
	// the label is the only key a caller has.
	Label string
	// Disabled greys the row and stops it being chosen.
	Disabled bool
	// Separator draws a hairline instead of a row. It is a bool rather than
	// a row with an empty label because an empty label is an item that does
	// nothing when pressed, which is not what a separator is.
	Separator bool
}

// Menu is one title of a menu bar and the items under it.
type Menu struct {
	// Label is what the title shows, and what a screen reader calls it. A
	// menu bar of glyphs has no menu bar.
	Label string
	// Items are the rows, top to bottom, separators included.
	Items []MenuItem
}

// MenubarOptions configure a Menubar.
type MenubarOptions struct {
	// Label names the bar as a whole for assistive technology; empty uses
	// the library's.
	Label string
	// Fill draws a surface behind the bar, for a window whose top strip sits
	// on something other than the page. False leaves it on the window.
	Fill bool
	// Height overrides the standard bar height.
	Height float32
}

// MenubarResult carries a Menubar, what was chosen from it, and which of its
// menus is open.
type MenubarResult struct {
	// Element is the bar.
	Element *ui.Element
	// open is the index of the title whose menu was pressed this frame, or
	// -1. The bar does not decide that one title at a time is open: the
	// desktop owns the menu while it is up and dismisses it itself, so
	// telling the app "a menu is open" would be a guess about something
	// that happens elsewhere.
	open int
	// chosen is the label of the item that ran this frame.
	chosen string
	// answered separates a real answer from an empty label.
	answered bool
}

// Open returns the index of the title the desktop has a menu under, or -1.
//
// It is read from the focus, because that is what the bar can see: a menu
// button takes the press itself and opens the menu, so there is no click left
// for the bar to report. The title holding the focus is the title the menu is
// under — and, it has to be said, it is also where the focus stays after the
// menu closes, so a caller should treat this as "the last menu this title
// opened", not as a live open/closed flag.
func (r MenubarResult) Open() int { return r.open }

// Chosen returns the label of the item that ran this frame, and whether one
// ran at all.
func (r MenubarResult) Chosen() (string, bool) { return r.chosen, r.answered }

// Menubar is the row of menus along a window's top edge: the application menu
// first, then File, Edit, View.
//
// Its menus are the desktop's own, drawn through overlay.DropdownMenu. The
// system draws a menu in its own look on its own terms — on macOS a system
// menu, elsewhere the window's own — and asking it to wear this library's
// hairline would be asking it to be something it is not. So the bar keeps its
// own two jobs: draw the row of titles, and name them. Everything the menu
// itself shows, including which items are disabled right now, is described
// from the caller's state on the frame the desktop asks for it.
//
// The bar holds no open state. Which menu is up is the desktop's business,
// and the result says which title was pressed rather than keeping a field to
// contradict it with.
func Menubar(c *ui.Context, app string, menus []Menu, opts MenubarOptions) MenubarResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if app != "" {
		// The application menu comes first and is not one of the app's own
		// menus: it belongs to the application rather than to a document,
		// and a bar without it is a bar of document menus.
		menus = append([]Menu{appMenu(app)}, menus...)
	}
	if len(menus) == 0 {
		panic("navigation: Menubar needs at least one menu; a bar with nothing to open is a rule")
	}

	h := opts.Height
	if h <= 0 {
		h = core.ControlHeight(c) + u*0.5
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "menubar.label", "Menu bar")
	}

	var r MenubarResult
	r.open = -1
	bar := ui.Row(c).FillWidth().Height(h).AlignItems(ui.Center).
		Padding(u*0.5, u*2).Gap(u * 0.25).Label(label).Role(ui.RoleMenuBar)
	if opts.Fill {
		bar.Background(k.Surface)
	}

	bar.Children(func() {
		for i, m := range menus {
			idx, menu := i, m
			trigger := overlay.DropdownMenu(c, menu.Label, overlay.DropdownMenuOptions{
				// The desktop opens the menu on the press itself, so the
				// trigger's own Clicked is only reporting which title was
				// used — the app does not open anything.
				Build: func(m *ui.Menu) {
					for _, it := range menu.Items {
						if it.Separator {
							m.Separator()
							continue
						}
						if m.Item(it.Label).Disabled(it.Disabled).Chosen() {
							r.chosen, r.answered = it.Label, true
						}
					}
				},
			})
			// The desktop opens the menu on the press itself, so there is
			// no click for the bar to read: a menu button takes the press
			// and the menu with it. What the bar can see is where the focus
			// went, and the title the desktop opened its menu under is the
			// title holding it.
			if trigger.Focused() {
				r.open = idx
			}
		}
	})
	r.Element = bar
	return r
}

// appMenu is the application menu's own rows — the three every desktop puts
// first, so that no application has to write them out again. They are real
// items: a menu bar that shows them and does nothing with them would be a
// lie, and the caller's MenubarResult.Chosen is where the answer comes back.
func appMenu(app string) Menu {
	return Menu{
		Label: app,
		Items: []MenuItem{
			{Label: "About " + app},
			{Separator: true},
			{Label: "Settings"},
			{Separator: true},
			{Label: "Quit " + app},
		},
	}
}

// NavigationMenuOptions configure a NavigationMenu.
type NavigationMenuOptions struct {
	// Open is the caller's open state: the trigger's press and a press
	// outside the panel both write into it, and Escape closes it through
	// ui/overlay. Nil is not an option — a menu whose open state lived in
	// the component would be a menu that forgets it on the next rebuild.
	Open *bool
	// Label names the trigger for assistive technology; empty uses the
	// trigger's own label.
	Label string
	// MaxWidth caps the panel, which keeps a menu of one-word items from
	// stretching across a wide window.
	MaxWidth float32
	// Border draws a hairline down the panel's side, for a navigation menu
	// inside a panel that is already ruled.
	Border bool
	// Modal leaves the window behind live: a press outside the panel goes
	// through to the page, as a menu hanging off a control should.
	Modal bool
}

// NavigationMenuResult carries a NavigationMenu and what was chosen in it.
type NavigationMenuResult struct {
	// Element is the trigger.
	Element *ui.Element
	// chosen is the label of the item pressed this frame.
	chosen string
	// answered separates a real answer from an empty label.
	answered bool
}

// Chosen returns the label of the item chosen this frame, and whether one was.
func (r NavigationMenuResult) Chosen() (string, bool) { return r.chosen, r.answered }

// NavigationMenuItem is one destination of a NavigationMenu.
type NavigationMenuItem struct {
	// Label is what the row shows and what Chosen returns.
	Label string
	// Icon is the glyph beside the label; nil leaves it out.
	Icon *ui.SVG
	// Description is the line under the label: where it goes, what is there.
	Description string
	// Selected marks the row as where you are. It is a value, as everywhere
	// else — the menu reads the caller's app state and keeps none of its
	// own.
	Selected bool
	// Disabled greys the row and stops it being chosen.
	Disabled bool
}

// NavigationMenu is a menu that belongs to the page rather than to the
// desktop: a list of destinations, a line about each, opening under the
// control that triggered it.
//
// It differs from a menu bar's menus in exactly that. Those are things to do
// to the application — Quit, Settings — and the operating system owns them.
// This is a list of places to go, it changes with where you are, and it is
// drawn here: in the library's panel, in the library's palette, with the
// descriptions and the mark on the current row that make it a navigation
// rather than a list of buttons. Handing it to the desktop would lose both,
// and with them the reason to have it.
//
// The panel is overlay.Panel and the layer is overlay.Popover, so it wears
// the same floating layer a dialog does and goes away on Escape or a press
// outside it, by the same code.
func NavigationMenu(c *ui.Context, triggerLabel string, items []NavigationMenuItem,
	opts NavigationMenuOptions) NavigationMenuResult {

	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Open == nil {
		panic("navigation: NavigationMenu needs the *bool it opens and closes")
	}
	if triggerLabel == "" && opts.Label == "" {
		panic("navigation: NavigationMenu needs a label or options.Label; a trigger nobody can read is a mystery box")
	}
	if len(items) == 0 {
		panic("navigation: NavigationMenu needs at least one destination; a menu with nowhere to go is a label")
	}
	name := opts.Label
	if name == "" {
		name = triggerLabel
	}

	var r NavigationMenuResult
	// Row, not Box: the label and the arrow belong on one line.
	trigger := ui.Row(c).FillWidth().Height(core.ControlHeight(c)).
		Padding(0, u*3).Radius(theme.ControlRadius).Background(k.Surface).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		AlignItems(ui.Center).Gap(u * 2).Label(name)

	trigger.Children(func() {
		ui.Box(c).Grow(1).Children(func() {
			if triggerLabel != "" {
				ui.Text(c, triggerLabel).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			}
		})
		// The arrow is drawn here rather than borrowed from a desktop menu
		// button: this trigger is a box on a page, and a widget that brings
		// its own padding and its own palette would not match the rows below
		// it. The same reason overlay has its own chevron.
		menuChevron(c, k.TextMuted, u*1.5)
	})

	panel := overlay.Popover(c, trigger, opts.Open, overlay.PopoverOptions{
		Label:    name,
		MaxWidth: opts.MaxWidth,
		Modal:    opts.Modal,
		Body: func() {
			ui.Column(c).FillWidth().Children(func() {
				for _, it := range items {
					navRow(c, it, u, &r)
				}
			})
		},
	})
	if panel != nil && opts.Border {
		panel.BorderWidth(theme.BorderWidth).BorderColor(k.Border)
	}

	// Pressing the trigger opens the panel. The desktop's own menus are
	// opened by their widgets; this one is a box on a page, so the press is
	// ours to read and ours to answer — into the caller's *bool, like every
	// other way into a layer in this library.
	if trigger.Clicked() {
		*opts.Open = true
	}
	r.Element = trigger
	return r
}

// navRow draws one destination of a NavigationMenu.
func navRow(c *ui.Context, it NavigationMenuItem, u float32, r *NavigationMenuResult) {
	k := core.Tokens(c)

	bg, fg, sub := k.Background, k.Text, k.TextMuted
	if it.Selected {
		bg = k.SurfaceHover
	}
	if it.Disabled {
		fg, sub = k.TextFaint, k.TextFaint
	}

	// Row, not Box: the icon, the label and the current-row mark share one
	// line, and a Box here is a column — the row would stack its own parts,
	// grow several lines tall, and spill out of the panel.
	row := ui.Row(c).FillWidth().Padding(u*2, u*2.5).Radius(theme.ControlRadius).
		Background(bg).AlignItems(ui.Center).Gap(u * 2.5).
		Label(it.Label).Disabled(it.Disabled)

	row.Children(func() {
		if it.Icon != nil {
			ui.Box(c).Shrink(0).Children(func() {
				ui.Icon(c, it.Icon).TextColor(sub).Size(theme.IconSize, theme.IconSize)
			})
		}
		ui.Column(c).Grow(1).Gap(0).Children(func() {
			ui.Text(c, it.Label).TextColor(fg).
				FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			if it.Description != "" {
				ui.Text(c, it.Description).TextColor(sub).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			}
		})
		// The current row says so in a mark as well as in colour: a
		// highlighted row and the row you are on look the same otherwise,
		// and only one of them means "you are here".
		if it.Selected && !it.Disabled {
			display.Icon(c, display.IconCheck, display.IconOptions{
				Name: core.Msg(c, "navmenu.current", "Current"), Size: theme.IconSize, Muted: true,
			})
		}
	})
	if row.Clicked() && !it.Disabled {
		r.chosen, r.answered = it.Label, true
	}
}

// menuChevron draws the arrow that says a control opens a panel under it.
func menuChevron(c *ui.Context, col ui.Color, side float32) {
	ui.Box(c).Size(side, side).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		var path ui.Path
		path.MoveTo(r.X+r.W*0.18, r.Y+r.H*0.32).
			LineTo(r.X+r.W*0.5, r.Y+r.H*0.64).
			LineTo(r.X+r.W*0.82, r.Y+r.H*0.32)
		p.StrokePath(&path, 1.5, col)
	})
}
