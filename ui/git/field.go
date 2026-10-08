package git

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// Two form controls a git panel cannot do without, built on MyGo's own bases
// rather than on ui/input.
//
// The reason is a layering one rather than a dislike: ui/input sits above
// ui/layout and ui/overlay sits above ui/input, and this package is a peer of
// all three. A commit message box is the only form control a git interface
// owns, and it is a box with a text area in it — which is what
// ui/input's field is too. Duplicating a whole control to avoid an import
// would be wrong, so what is here is the smallest thing that wears the
// library's own field skin: the same ground, the same hairline, the same
// control height at this density, and the same tick box.
//
// Every glyph-free element these build is named, because a field with no name
// is a mystery box and a tick with no words beside it has nothing to read out.

// messageField is the box a commit message is written in.
//
// The value is the caller's the moment it is typed, which is what lets a
// caller validate as somebody writes rather than on a commit step.
//
// The error is both a danger-coloured hairline and a line of words under the
// well. The hairline alone is what every other field in the library does,
// and it is not enough here: this field's rule is about a button that is
// disabled rather than about a value that is wrong, and a disabled button
// with nothing written under it looks like a broken panel. So the message is
// said as well as shown.
func messageField(c *ui.Context, value *string, name, placeholder, err string, height float32) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	hair := k.Border
	if err != "" {
		hair = k.Danger
	}
	if height <= 0 {
		height = u * 10
	}

	field := ui.Column(c).FillWidth().Gap(u * 0.5)
	field.Children(func() {
		well := ui.Box(c).FillWidth().Height(height).Padding(u*1.5, u*2).
			Radius(theme.ControlRadius).Background(k.Surface).
			BorderWidth(theme.BorderWidth).BorderColor(hair).
			Role(ui.RoleNone)
		well.Children(func() {
			// TextAreaBase is MyGo's editor: the cursor, the selection, the
			// undo, the clipboard and the spell-check-free plain-text
			// contract are all its. What it does not have is a look, which is
			// what the well above gives it.
			area := ui.TextAreaBase(c, value)
			area.Fill().TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize)).
				Label(name).Placeholder(placeholder).Role(ui.RoleTextField)
			// Pressing anywhere in the well focuses the control inside it, so
			// the padding reads as part of the field rather than as a margin
			// beside it. Without it a message box has a band above the text
			// that does nothing, which is the classic "is that a field?".
			if well.Clicked() {
				area.Focus()
			}
		})
		if err != "" {
			ui.Text(c, err).TextColor(k.Danger).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
	return field
}

// tickBox is a labelled check box: the one beside "Amend last commit", and
// the one beside anything else in this package a person can switch on.
//
// It is the same shape as every check box in the library because it is the
// same base — MyGo's own — with the library's tick drawn on it. Nothing here
// reimplements a click, a Space or a focus: those are the desktop's.
func tickBox(c *ui.Context, on *bool, label string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	// The pointer into the caller's state is flipped by the base and read
	// back here, because the base works on a bool it owns. The copy is
	// written the same frame the click happened, which is what makes the
	// control a view of one value rather than a second copy of it.
	local := *on
	box := ui.CheckboxBase(c, &local).Gap(u*1.5).Radius(theme.SmallRadius).
		Padding(u*0.5, u).AlignItems(ui.Center).FillWidth().
		Label(label).Tooltip(label)
	if box.Changed() {
		*on = local
	}
	box.Children(func() {
		tickFace(c, local)
		ui.Text(c, label).TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize))
	})
	return box
}

// tickFace is the box in front of a tick: empty, or filled with the accent
// and a tick in the ink that reads on it.
func tickFace(c *ui.Context, on bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 4
	box := ui.Box(c).Size(side, side).Shrink(0).Radius(side * 0.28).
		Role(ui.RoleNone)
	if on {
		box.Background(k.Accent)
	} else {
		box.Background(k.Background).
			BorderWidth(theme.BorderWidth).BorderColor(k.Border.Mix(k.Text, 0.25))
	}
	box.DrawOver(func(p *ui.Painter, r ui.Rect) {
		if !on {
			return
		}
		// The ink is asked of the colour rather than taken from a token:
		// a window may set the accent to anything, and a tick in the
		// window's text colour disappears the moment that accent is dark.
		white, black := ui.RGB(255, 255, 255), ui.RGB(0, 0, 0)
		ink := black
		if lum(k.Accent) > 0.5 {
			ink = white
		}
		var path ui.Path
		path.MoveTo(r.X+r.W*0.24, r.Y+r.H*0.52).
			LineTo(r.X+r.W*0.44, r.Y+r.H*0.72).
			LineTo(r.X+r.W*0.78, r.Y+r.H*0.3)
		p.StrokePath(&path, 1.5, ink)
	})
	return box
}

// lum is the relative brightness of a colour, from 0 to 1, as a plain average
// of its bytes. It is enough to choose between black and white ink, and it is
// not the WCAG measure on purpose: the choice here is which of two inks is
// more visible on a swatch, and the answer is the same either way.
func lum(col ui.Color) float32 {
	return (float32(col.R) + float32(col.G) + float32(col.B)) / (3 * 255)
}

var _ = internal.Dot
