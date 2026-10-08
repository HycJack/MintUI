package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ToggleOptions configure a Toggle.
type ToggleOptions struct {
	// Icon draws before the label, for a toolbar button that says its thing
	// with a mark as much as with its words. Nil leaves it out.
	Icon *ui.SVG
	// Disabled greys the button out: it takes neither clicks nor focus.
	Disabled bool
}

// Toggle is a button that stays pressed: a press flips the *bool it points
// at, and while the flag is true the button is filled, so the state is the
// face rather than something to read off the app afterwards.
//
// It is the toolbar's control — "Only favourites", "Show hidden", "Wrap
// lines" — the state of a view the person switches on and off while looking
// at the thing the view is on. The state is the caller's *bool, flipped the
// moment the button is pressed, so there is no handler to keep in step with
// it and no way for the face to disagree with the value.
//
// The difference from a Switch is what the control says: a Switch is an
// on/off with a knob, the shape a setting takes — the thing is on, the
// thing is off — and it carries a label beside the knob that names the
// setting. A Toggle is a button in a run of actions whose pressed state
// marks the one that is active; it is the shape a filter or a display mode
// takes. The difference from a ToggleGroup is that a Toggle stands alone:
// it is one thing being on or off, not one of a set exactly one of which is
// chosen, which is what the group's segments are.
func Toggle(c *ui.Context, pressed *bool, label string, opts ToggleOptions) *ui.Element {
	if pressed == nil {
		panic("input: Toggle needs a pressed flag to point at; it keeps no state of its own")
	}
	if label == "" {
		panic("input: Toggle needs a label; a filled box with no words is a swatch")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	// ToggleBase is the behaviour underneath: it flips the bool it is given
	// on a click or on Space and reports it as a change, which is all a
	// toolbar button is. The face reads from the caller's flag, so a press
	// shows its answer and nothing the flag does can be a frame behind.
	on := *pressed
	btn := ui.ToggleBase(c, &on).Height(core.ControlHeight(c)).Radius(theme.ControlRadius).
		Padding(0, u*2.5, 0, u*2.5).Shrink(0).Role(ui.RoleToggleButton).
		Label(label).Tooltip(label).Disabled(opts.Disabled)
	if btn.Changed() {
		*pressed = on
	}

	bg, fg := k.Surface, k.Text
	switch {
	case *pressed:
		bg, fg = k.Fill, k.OnFill
	case btn.Hovered():
		bg = k.SurfaceHover
	}
	if opts.Disabled {
		bg, fg = k.Surface, k.TextFaint
	}
	btn.Background(bg).TextColor(fg)
	btn.Children(func() {
		if opts.Icon != nil {
			ui.Icon(c, opts.Icon).Size(u*4.25, u*4.25).TextColor(fg).Shrink(0)
		}
		ui.Text(c, label).SingleLine().FontSize(core.FontSize(c, theme.RowSize))
	})
	return btn
}
