// Package input holds the controls a person operates: buttons, segmented
// pickers, search fields and switches.
package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ButtonOptions configure a Button.
type ButtonOptions struct {
	// Primary fills the button with ink, for the one action a surface is
	// about. A plain Button leaves it to the surface.
	Primary bool
	// Danger draws the label in the danger colour, for a destructive action
	// that still wants to sit quietly until pressed.
	Danger bool
	// Disabled greys the button out: it takes neither clicks nor focus.
	Disabled bool
	// Label names the button for assistive technology. Buttons whose visible
	// text says enough need none; icon-only buttons always do.
	Label string
	// Icon draws before the label; nil leaves it out.
	Icon *ui.SVG
}

// Button is a press button showing a label.
//
// The label is required: an empty one leaves the control unnamed, which reads
// as nothing at all to a screen reader.
func Button(c *ui.Context, label string, opts ButtonOptions) *ui.Element {
	if label == "" && opts.Label == "" {
		panic("input: Button needs a label or options.Label")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	bg, fg := k.Surface, k.Text
	switch {
	case opts.Primary:
		bg, fg = k.Fill, k.OnFill
	case opts.Danger:
		fg = k.Danger
	}

	name := label
	if name == "" {
		name = opts.Label
	}
	btn := ui.Button(c, "").Height(u*11).Radius(theme.PillRadius).
		Padding(0, u*5, 0, u*5).Background(bg).TextColor(fg).
		Label(opts.Label).Disabled(opts.Disabled).Tooltip(opts.Label)
	if opts.Label != "" && label != "" {
		btn = btn.Tooltip("")
	}
	btn.Children(func() {
		if opts.Icon != nil {
			ui.Icon(c, opts.Icon).TextColor(fg).Size(u*4.25, u*4.25)
		}
		if label != "" {
			ui.Text(c, label)
		}
	})
	_ = name
	return btn
}

// IconButton is a square button showing only a glyph. Its label is required
// and is what a screen reader announces, so it should say what the button
// does rather than what it looks like.
func IconButton(c *ui.Context, glyph *ui.SVG, name string, opts ButtonOptions) *ui.Element {
	if glyph == nil {
		panic("input: IconButton needs an icon")
	}
	if name == "" {
		panic("input: IconButton needs a name for assistive technology")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	bg, fg := k.Surface, k.Text
	if opts.Primary {
		bg, fg = k.Fill, k.OnFill
	}
	side := u * 11
	btn := ui.Button(c, "").Size(side, side).Radius(theme.PillRadius).
		Background(bg).TextColor(fg).Label(name).Tooltip(name).
		Disabled(opts.Disabled).Children(func() {
		ui.Icon(c, glyph).TextColor(fg).Size(u*4.5, u*4.5)
	})
	return btn
}
