package input

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// The two floating things a window wears over its content: the bar that rises
// over a list once anything in it is chosen, and the panel hung from an edge.
//
// Both of them are layout.Panel and nothing else. That is the reason they are
// in this file together: a floating layer in this library is one radius, one
// hairline, one shadow and one set of padding, and a component that drew its
// own would be the second answer to that question.

// BulkAction is one action in a BulkActionBar.
type BulkAction struct {
	// Label is the words on the button, and is required: a bulk bar is a row
	// of the most destructive things a window can do, and an unnamed one of
	// them is the worst possible mystery.
	Label string
	// Icon draws before the label.
	Icon *ui.SVG
	// Danger draws it in the danger colour, for "Delete" in a bar whose
	// other actions are "Archive" and "Assign".
	Danger bool
	// Disabled takes this action out for this bar. A bar's actions are
	// usually all-or-nothing on the selection, and a disabled one says which.
	Disabled bool
}

// BulkActionBarOptions configure a BulkActionBar.
type BulkActionBarOptions struct {
	// Label names the bar for assistive technology. It is required, and it
	// should be the action rather than the number — "Archive selected" — so
	// that a reader arriving at one of the buttons is told what they do
	// before they do it.
	Label string
	// Count is what the bar says: how many things are chosen. It defaults to
	// the length of the caller's selection, which is almost always what the
	// caller wanted, and is an option because a bar sometimes reports
	// something else: "12 of 340 selected".
	Count int
	// Noun is what is being chosen, and it makes the count read as English:
	// "3 selected" with no noun is a number and a status word.
	Noun string
	// Actions is what the bar offers, left to right after the count and
	// before the clear. Empty draws a bar with a count and nothing to do,
	// which is not useless: it is how a window says what the selection is
	// before the keyboard shortcuts arrive.
	Actions []BulkAction
	// Clearable puts a button at the trailing edge that empties the
	// selection, which is the way out of a selection nobody can undo and a
	// list of a thousand rows is otherwise walked back one press at a time.
	Clearable bool
	// Width bounds the bar; zero takes what its content needs.
	Width float32
}

// BulkActionBarResult carries a BulkActionBar and what was pressed in it.
type BulkActionBarResult struct {
	// Element is the bar, or nil while nothing is chosen: the bar is drawn
	// floating over the list it belongs to, so a list with nothing chosen in
	// it has no bar, rather than a bar with an empty count on it.
	Element *ui.Element
	// action is the value of the action pressed this frame, "" for none.
	action string
	// cleared reports that the clear button was pressed this frame.
	cleared bool
}

// Action reports which action was pressed this frame, by its Label. It is "" on
// every other frame, so a caller reads it inside its own view rather than
// after the frame is over.
func (r BulkActionBarResult) Action() string { return r.action }

// Cleared reports that the clear button was pressed this frame, which is the
// caller's cue to drop its selection.
func (r BulkActionBarResult) Cleared() bool { return r.cleared }

// BulkActionBar is the bar that rises over a list once something in it is
// chosen: the count, the actions, and a way back to choosing nothing.
//
// It writes nothing itself. The count is the length of selected — the caller's
// slice — and the clear button writes false into the caller's *bool, so what a
// caller does with "these are chosen" stays where it already is. A bar that
// kept its own copy of the selection would be a second answer to which rows
// are chosen, and two answers to that is the bug every bulk-select ever had.
//
// It returns nil while the selection is empty, which is what lets a caller
// write it unconditionally inside the row that holds the list.
func BulkActionBar(c *ui.Context, chosen *bool, selected []string, opts BulkActionBarOptions) BulkActionBarResult {
	if chosen == nil {
		panic("input: BulkActionBar needs the *bool it rises and settles on")
	}
	if opts.Label == "" {
		panic("input: BulkActionBar needs options.Label; a row of destructive actions " +
			"that says nothing about what it acts on is the worst place in the interface " +
			"to say nothing")
	}
	for i, a := range opts.Actions {
		if a.Label == "" {
			panic("input: BulkActionBar action " + strconv.Itoa(i) + " has no Label; the button would say nothing about what it does")
		}
	}

	var r BulkActionBarResult
	if !*chosen || len(selected) == 0 {
		// Nothing is chosen, so there is no bar. Building one anyway would
		// cost a frame's layout on a panel nobody can see — the same guard
		// ui/layout's layer keeps before ui.Modal.
		return r
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	n := opts.Count
	if n <= 0 {
		n = len(selected)
	}
	say := core.Msg(c, "input.selected", core.Def("selected"))
	noun := opts.Noun
	count := itoa(n) + " " + say
	if noun != "" {
		// The noun is the caller's own subject, not the library's, so it is
		// pluralised by the caller and only the pattern is ours.
		count = itoa(n) + " " + noun + " " + say
	}

	// The bar is built inside an Overlay so it floats over the list rather
	// than pushing it: a bar that took a row out of the layout would move
	// every row under it down by its own height the moment anything was
	// chosen, which is the single most jarring thing a list can do.
	ui.Overlay(c, func() {
		host := ui.Row(c).Absolute().Left(0).Right(0).Bottom(u * 2).
			Justify(ui.Center)
		if opts.Width > 0 {
			host.Width(opts.Width).Shrink(0)
		}
		panel := layout.Panel(c, host, layout.PanelOptions{
			Compact: true, Pad: u, MaxWidth: opts.Width, Label: opts.Label,
		}, nil)
		panel.Row().AlignItems(ui.Center).Gap(u).Children(func() {
			ui.Text(c, count).TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize)).
				FontWeight(600).SingleLine()

			for _, a := range opts.Actions {
				act := a
				fg, bg := k.Text, k.Background
				if act.Danger {
					fg = k.Danger
				}
				if act.Disabled {
					fg = k.TextFaint
				}
				btn := ui.ButtonBase(c).Height(core.ControlHeight(c)-u*2).Shrink(0).
					Radius(theme.PillRadius).Padding(0, u*2.5, 0, u*2.5).
					Background(bg).TextColor(fg).Role(ui.RoleButton).
					Label(act.Label).Tooltip(act.Label).Disabled(act.Disabled)
				if btn.Hovered() && !act.Disabled {
					btn.Background(k.SurfaceHover)
				}
				if btn.Clicked() {
					r.action = act.Label
				}
				btn.Children(func() {
					if act.Icon != nil {
						ui.Icon(c, act.Icon).Size(u*4, u*4).TextColor(fg).Shrink(0)
					}
					ui.Text(c, act.Label).SingleLine().
						FontSize(core.FontSize(c, theme.RowSize)).TextColor(fg)
				})
			}

			if opts.Clearable {
				clear := core.Msg(c, "input.clearSelection", core.Def("Clear selection"))
				btn := ui.ButtonBase(c).Height(core.ControlHeight(c)-u*2).Shrink(0).
					Radius(theme.PillRadius).Padding(0, u*2.5, 0, u*2.5).
					Background(ui.Transparent).TextColor(k.TextMuted).
					Role(ui.RoleButton).Label(clear).Tooltip(clear)
				if btn.Hovered() {
					btn.Background(k.SurfaceHover)
				}
				if btn.Clicked() {
					r.cleared = true
					*chosen = false
					c.Announce(clear)
				}
				btn.Children(func() {
					ui.Icon(c, glyphCross).Size(u*3.5, u*3.5).TextColor(k.TextMuted).Shrink(0)
				})
			}
		})
		r.Element = panel
	})
	return r
}

// DockOptions configure a Dock.
type DockOptions struct {
	// Title heads the panel. It is required for the same reason a dialog's
	// is: a panel with no name is read out as a group, and a drawer a reader
	// cannot name is a drawer they cannot close.
	Title string
	// Subtitle is the second line under the title.
	Subtitle string
	// Side is which edge the dock hangs from. Left and Right take the whole
	// height and are the usual answers; Top is a bar and Bottom is a status
	// area, and both are square along the edge they meet the window.
	//
	// It is named for the edge rather than for the direction the panel's
	// content runs, because "vertical" reads both ways — the rule
	// ui/layout's SplitPaneOptions follows for the same reason.
	Side Side
	// Width is the panel's width along the edge; zero is a quarter of the
	// window, which is what a drawer wants and what a drawer too narrow for
	// its own labels never is.
	Width float32
	// Closeable puts a button at the trailing edge of the header that writes
	// false into the caller's *bool. A dock with no way out is a panel that
	// has taken the window.
	Closeable bool
	// Compact wears the smaller panel — tighter padding and the control
	// radius — which is right for a peek and wrong for a drawer holding a
	// form.
	Compact bool
}

// Side is which edge of the window a Dock hangs from.
type Side int

const (
	// DockLeft is the edge down the left-hand side, which is where a
	// navigation drawer and a details pane live.
	DockLeft Side = iota
	// DockRight is the opposite edge.
	DockRight
	// DockTop is a bar across the top of the window.
	DockTop
	// DockBottom is a bar across the bottom, for a console.
	DockBottom
)

func (s Side) String() string {
	switch s {
	case DockRight:
		return "Right"
	case DockTop:
		return "Top"
	case DockBottom:
		return "Bottom"
	}
	return "Left"
}

// DockResult carries a Dock and whether it was closed.
type DockResult struct {
	// Element is the panel, or nil while it is shut.
	Element *ui.Element
	closed  bool
}

// Closed reports that the close button was pressed this frame, so a caller
// that keeps the panel's width somewhere can put it back.
func (r DockResult) Closed() bool { return r.closed }

// Dock is a panel hung from one edge of the window.
//
// It is a panel and not a container because of where it sits: a dock is over
// the content, it is square where it meets the window's edge, and it floats —
// which is exactly what ui/layout.Panel already decides, so Dock asks it for
// the radius rather than inventing a second one. Round is what does the
// squaring, which is why PanelOptions has it at all.
//
// It returns nil while shut. A dock that stayed in the layout with its
// contents built would cost every frame's layout on a panel nobody can see.
func Dock(c *ui.Context, open *bool, opts DockOptions, body func()) DockResult {
	if open == nil {
		panic("input: Dock needs the *bool it opens and closes")
	}
	if !*open {
		return DockResult{}
	}
	if opts.Title == "" {
		panic("input: Dock needs a Title; a panel a reader cannot name is a panel " +
			"they cannot close")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	w := opts.Width
	if w <= 0 {
		w = theme.SidebarWidth * 0.75
	}

	var r DockResult
	host := ui.Column(c).FillHeight().Shrink(0)
	if opts.Side == DockLeft || opts.Side == DockRight {
		host.Width(w)
	} else {
		host.Height(w).FillWidth()
	}
	// The panel's own radius, knocked back along the edge it is hung from:
	// a panel square against the window's edge and round against its own
	// would show a curve where there is nothing to curve away from.
	panel := layout.Panel(c, host, layout.PanelOptions{
		Title: opts.Title, Subtitle: opts.Subtitle, Compact: opts.Compact,
		Label: opts.Title,
		Round: func(radius float32) (tl, tr, br, bl float32) {
			tl, tr, br, bl = radius, radius, radius, radius
			switch opts.Side {
			case DockLeft:
				tl, bl = 0, 0
			case DockRight:
				tr, br = 0, 0
			case DockTop:
				tl, tr = 0, 0
			case DockBottom:
				br, bl = 0, 0
			}
			return tl, tr, br, bl
		},
	}, nil)
	panel.Children(func() {
		if opts.Closeable {
			closeName := core.Msg(c, "input.close", core.Def("Close"))
			// Above the body rather than beside the title, so a long title
			// cannot push the close button out of the panel it belongs to.
			btn := ui.ButtonBase(c).Size(u*6.5, u*6.5).Shrink(0).Radius(theme.PillRadius).
				Background(k.Surface).Role(ui.RoleButton).
				Label(closeName).Tooltip(closeName).Children(func() {
				ui.Icon(c, glyphCross).Size(u*3.5, u*3.5).TextColor(k.TextMuted)
			})
			if btn.Clicked() {
				r.closed = true
				*open = false
			}
		}
		if body != nil {
			body()
		}
	})
	r.Element = panel
	return r
}
