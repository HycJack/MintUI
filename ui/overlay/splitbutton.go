package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// SplitButtonOptions configure a SplitButton.
type SplitButtonOptions struct {
	// Build fills the menu the arrow opens.
	Build func(m *ui.Menu)
	// Primary fills the whole control with ink, for the one action a surface
	// is about.
	Primary bool
	// Danger draws both labels in the danger colour, for the action that
	// destroys something.
	Danger bool
	// Disabled greys the whole control out: it takes neither clicks nor
	// focus.
	Disabled bool
	// Label names the arrow for assistive technology. It is required: the
	// arrow shows nothing but a chevron, and a menu with no name is a list
	// nobody can be told about.
	Label string
}

// SplitButtonResult carries a SplitButton and the press of its action half.
type SplitButtonResult struct {
	// Element is the whole control.
	Element *ui.Element
	pressed bool
}

// Pressed reports the action — the half carrying the label — being pressed.
// The arrow half opens a menu instead and reports nothing: what a menu item
// does is asked inside Build, where the item is.
func (r SplitButtonResult) Pressed() bool { return r.pressed }

// SplitButton is one control doing two things: the action, in the half that
// says what it is, and the menu of the less usual ones, behind the arrow.
//
// It is one control rather than two buttons because the two are one decision
// — there is one thing to do here and a shortlist of the rest — and two
// separate controls would let a person pick the wrong one.
//
//	core.Use(c, core.Settings{})
//	if SplitButton(c, "Log callback", SplitButtonOptions{
//	    Label:  "More ways to log a callback",
//	    Primary: true,
//	    Build: func(m *ui.Menu) {
//	        if m.Item("Log from a template").Chosen() {
//	            app.fromTemplate()
//	        }
//	    },
//	}).Pressed() {
//	    app.log()
//	}
//
// The menu is NonModal in the sense this package documents: it is a list of
// things to do to the page, and it belongs to the desktop.
func SplitButton(c *ui.Context, label string, opts SplitButtonOptions) SplitButtonResult {
	if label == "" {
		panic("overlay: SplitButton needs a label; the half carrying the words is the action")
	}
	if opts.Build == nil {
		panic("overlay: SplitButton needs a Build; an arrow opening nothing is not an arrow")
	}
	if opts.Label == "" {
		panic("overlay: SplitButton needs a Label for the arrow; it shows a chevron and says nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := triggerInk(k, opts.Primary, opts.Danger)
	// The action half's own height. A split button whose halves are not the
	// same height is not one control, so the arrow takes what input.Button
	// gives the label rather than a number of its own.
	height := u * 11

	res := SplitButtonResult{}
	// The row carries no name of its own: the two halves are named, and a
	// name on the row as well would leave a test or a reader two elements
	// answering to one.
	res.Element = ui.Row(c).AlignItems(ui.Center).Grow(0).Children(func() {
		action := input.Button(c, label, input.ButtonOptions{
			Primary:  opts.Primary,
			Danger:   opts.Danger,
			Disabled: opts.Disabled,
			Label:    label,
		}).Radius(splitRound(theme.PillRadius, true))
		if action.Clicked() {
			res.pressed = true
		}
		ui.ButtonBase(c).
			Height(height).Radius(splitRound(theme.PillRadius, false)).
			Padding(0, u*1.75, 0, u*1.75).Background(bg).TextColor(fg).
			Label(opts.Label).Disabled(opts.Disabled).
			Children(func() { chevron(c, fg, u) }).
			Menu(opts.Build)
	})
	return res
}

// splitRound rounds the two corners of one half of a control that was split
// in two and squares the two where the halves meet.
//
// It is the one asymmetry the radius family has to spell out, and it earns
// it: a control that is round at its ends and square where it was cut reads
// as one object, where two halves each keeping the full radius read as two
// buttons sitting next to each other.
func splitRound(r float32, left bool) (topLeft, topRight, bottomRight, bottomLeft float32) {
	if left {
		// Round on the left, square where the halves meet: (r, 0, r, 0)
		// rounds a diagonal, and the left cap reads as a leaf cut on the
		// bias rather than as half of one pill.
		return r, 0, 0, r
	}
	return 0, r, r, 0
}
