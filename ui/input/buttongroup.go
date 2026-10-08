package input

import (
	"strconv"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ButtonGroup is a run of buttons joined into one control: one plate, one
// hairline, no gap between the neighbours, and the buttons pressed together
// inside it.
//
// It exists next to ToggleGroup and next to Segmented because the three
// differ in one thing each, and choosing between them is a matter of saying
// which. Segmented is MyGo's own: equal-width segments on a track, with the
// keyboard moving between them. ToggleGroup is the same idea with a gap and a
// shadow on the chosen one, so a chosen segment reads as a button rather than
// as a notch cut in a track. ButtonGroup is the third shape: the buttons are
// cut out of one plate, which is what a toolbar of equal actions looks like
// and what a Segmented is not — a Segmented says "one of these", a ButtonGroup
// of equal buttons says "these, together".
//
// The chosen one is the caller's string, as RadioGroup's is: the value is what
// a form submits, and an icon is decoration that rides with a label.
type GroupButton struct {
	// Value is what the caller's state holds while this one is chosen. It is
	// required: a button the caller cannot recognise afterwards is not one it
	// can act on.
	Value string
	// Label is the words on the button.
	Label string
	// Icon draws before the label, and does not replace it. A run of
	// icon-only buttons cannot be scanned, and a screen reader meets a row of
	// them as a row of nameless controls — so an icon beside no label is
	// allowed, but nothing at all is not.
	Icon *ui.SVG
	// Disabled takes this one out of play without greying out the run. A
	// toolbar where the whole row is dead is a toolbar nobody looks at twice.
	Disabled bool
	// Tip is the button's tooltip, for a run whose labels are an icon's short
	// name and whose long one is worth saying.
	Tip string
}

// ButtonGroupOptions configure a ButtonGroup.
type ButtonGroupOptions struct {
	// Label names the group for assistive technology. It is required: a run
	// of pressed-together buttons is read out as a row of buttons, and a row
	// of buttons without a name says nothing about what they are for.
	Label string
	// Wrap lets the run break onto a second line rather than overflow, for a
	// toolbar with more actions than the window has room for.
	Wrap bool
	// Disabled greys out every button in the run.
	Disabled bool
	// Size scales the buttons. Zero is a standard control at this window's
	// density; the rest are for a run inside a sheet or a card smaller than a
	// window.
	Size float32
	// LockChoice refuses a value that is not one of the run's, and refuses to
	// let a press on the chosen button clear it — which is what a RadioGroup
	// does and what a filter row of view modes needs.
	//
	// It is a question here rather than the rule because a run of actions — a
	// toolbar of icons, the buttons on an empty state — quite often has no
	// business being one of them chosen at all.
	LockChoice bool
}

// ButtonGroup is a joined run of buttons with one of them chosen, writing the
// chosen value into the caller's string.
//
// A press writes straight into the caller's pointer, so there is no handler to
// keep in step with the state. Pressing the chosen button again clears the
// choice rather than re-asserting it, unless LockChoice says otherwise: a
// control that could only be turned on and not off would be a switch wearing
// a toolbar's clothes.
func ButtonGroup(c *ui.Context, selected *string, buttons []GroupButton, opts ButtonGroupOptions) *ui.Element {
	if selected == nil {
		panic("input: ButtonGroup needs a selection to point at")
	}
	if len(buttons) == 0 {
		panic("input: ButtonGroup needs at least one button")
	}
	if opts.Label == "" {
		panic("input: ButtonGroup needs options.Label; a run of buttons that says " +
			"nothing about what they are for is a row of marks")
	}
	for i, b := range buttons {
		if b.Value == "" {
			panic("input: ButtonGroup button " + strconv.Itoa(i) + " has no Value; a button nothing can be told apart from is not one")
		}
		if b.Label == "" && b.Icon == nil {
			panic("input: ButtonGroup button " + strconv.Itoa(i) + " has neither a Label nor an Icon; it would be an empty plate")
		}
	}
	if *selected != "" && !hasValue(checkGroupValues(buttons), *selected) {
		if opts.LockChoice {
			panic("input: ButtonGroup selected " + *selected + ", which is not one of its buttons")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	h := opts.Size
	if h <= 0 {
		h = core.ControlHeight(c)
	}
	// One plate for the whole run rather than a radius per button, so the
	// hairline runs the whole width of it instead of stopping at every seam.
	plate := ui.Row(c).Shrink(0).Radius(theme.ControlRadius).
		Background(k.Surface).BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Role(ui.RoleGroup).Label(opts.Label)
	if opts.Wrap {
		plate.Wrap().AlignContent(ui.Start)
	}
	plate.Children(func() {
		for i, b := range buttons {
			chosen := *selected == b.Value
			off := opts.Disabled || b.Disabled
			btn := ui.ButtonBase(c).Height(h).Padding(0, u*2.5).Shrink(0).
				Role(ui.RoleToggleButton).Checked(chosen).
				Label(buttonName(b, opts.Label)).Tooltip(buttonTip(b)).
				Disabled(off)
			// The inner corners are what join the run: a button with all four
			// of its own is a button, and one with none is a notch, and a
			// row of notches on a plate is one control rather than five.
			r := groupRadii(i, len(buttons))
			btn.Radius(r[0], r[1], r[2], r[3])
			if btn.Clicked() && !(chosen && opts.LockChoice) {
				if chosen {
					*selected = ""
				} else {
					*selected = b.Value
				}
			}
			switch {
			case chosen:
				btn.Background(k.Fill)
			case off:
				btn.Background(k.Surface)
			case btn.Hovered():
				btn.Background(k.SurfaceHover)
			}
			// The seam with the button on the left of it, drawn by this one,
			// so a run of five has four hairlines and not five to work out
			// which side of each is which.
			if i > 0 {
				btn.BorderWidth(theme.BorderWidth).BorderColor(k.Border)
			}
			ink := groupInk(c, chosen, off)
			btn.Children(func() {
				if b.Icon != nil {
					ui.Icon(c, b.Icon).Size(u*4, u*4).TextColor(ink).Shrink(0)
				}
				if b.Label != "" {
					ui.Text(c, b.Label).SingleLine().
						FontSize(core.FontSize(c, theme.BodySize)).TextColor(ink)
				}
			})
		}
	})
	return plate
}

// groupInk is the ink on a button of a run: the ink that reads on the chosen
// face, and the window's own text otherwise.
//
// The chosen face is Fill, and a window is free to set Fill to any colour at
// all, so the ink is asked for rather than assumed: OnFill is right for both
// the palettes the library ships and wrong for the pale Fill a caller may
// bring, where white-on-white is a button with nothing on it.
func groupInk(c *ui.Context, chosen, disabled bool) ui.Color {
	k := core.Tokens(c)
	switch {
	case disabled:
		return k.TextFaint
	case chosen:
		return inkOn(k.Fill)
	}
	return k.Text
}

// groupRadii is the radius the button at place at of a run of n takes: the
// first keeps its leading corners, the last its trailing ones, and anything in
// between keeps none.
//
// A run of one keeps all four, which is a button rather than a run, and is why
// this is asked with the count rather than being a constant.
func groupRadii(at, n int) [4]float32 {
	const flat float32 = 0
	switch {
	case n == 1:
		return [4]float32{theme.ControlRadius, theme.ControlRadius, theme.ControlRadius, theme.ControlRadius}
	case at == 0:
		return [4]float32{theme.ControlRadius, flat, flat, theme.ControlRadius}
	case at == n-1:
		return [4]float32{flat, theme.ControlRadius, theme.ControlRadius, flat}
	}
	return [4]float32{flat, flat, flat, flat}
}

// buttonName is what a button in a run is called out as: its own label, or
// its tip, or the group's name — so that an icon-only button is never the one
// thing in the row a reader cannot name.
func buttonName(b GroupButton, group string) string {
	switch {
	case b.Label != "":
		return b.Label
	case b.Tip != "":
		return b.Tip
	}
	return group
}

// buttonTip is a button's tooltip: the one it was given, or its label, so that
// every button in a run has one and none of them says the same thing twice.
func buttonTip(b GroupButton) string {
	if b.Tip != "" {
		return b.Tip
	}
	return b.Label
}

func checkGroupValues(buttons []GroupButton) []string {
	out := make([]string, len(buttons))
	for i, b := range buttons {
		out[i] = b.Value
	}
	return out
}
