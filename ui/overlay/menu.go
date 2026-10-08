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
