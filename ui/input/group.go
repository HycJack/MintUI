package input

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The two controls that put something else inside a field, and the one that
// takes a field's value out of it.
//
// An InputGroup is the same well as every other field with a mark at one end
// and a button at the other, which is why it is written on top of the same
// box rather than as a control of its own: a group that looked different
// would read as a different kind of field.

// InputGroupOptions configure an InputGroup.
type InputGroupOptions struct {
	// Label names the field; it is required.
	Label string
	// Placeholder shows while the value is empty.
	Placeholder string
	// Error marks the value as not valid.
	Error string
	// Icon marks the leading edge, for a field about a named thing — a
	// branch, a customer, a URL. It is decoration: the field is named by
	// Label, and never by what it happens to show.
	Icon *ui.SVG
	// Prefix and Suffix sit inside the well, before and after the text.
	// Neither is part of the value: "https://" and ".com" are what the
	// field is about rather than what it holds.
	Prefix, Suffix string
	// Trailing is drawn at the trailing edge, after the suffix — a
	// CopyButton, a clear button, whatever the field gives you as a last
	// act. It is built inside the well, so it counts as part of this field
	// and not as the next control in the form.
	Trailing func()
	// Password hides what is typed. A group that should also be revealable
	// wants PasswordInput instead, which owns that switch and its key.
	Password bool
	// ReadOnly keeps the value selectable and copyable but not editable.
	ReadOnly bool
	// Disabled greys the field and takes it out of the tab order.
	Disabled bool
	// Width is the field's own width; zero fills the row it is in.
	Width float32
}

// InputGroup is a text field with furniture inside it: a leading icon, a
// scheme or a unit at either end, and a button at the trailing edge.
//
// Everything it draws around the text belongs to the field — the well, the
// hairline, the height, the error — so a group set beside a plain TextInput
// in the same form lines up with it exactly.
func InputGroup(c *ui.Context, value *string, opts InputGroupOptions) *ui.Element {
	if value == nil {
		panic("input: InputGroup needs a value to point at; it keeps no value of its own")
	}
	if opts.Label == "" {
		panic("input: InputGroup needs a Label, or nothing can name the field")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	s := fieldSkin{
		label: opts.Label, err: opts.Error,
		disabled: opts.Disabled, readOnly: opts.ReadOnly, width: opts.Width,
	}

	return textWell(c, s, func(_ *ui.Element) *ui.Element {
		if opts.Icon != nil {
			ui.Icon(c, opts.Icon).TextColor(k.TextMuted).Size(u*4.25, u*4.25).Shrink(0)
		}
		if opts.Prefix != "" {
			fieldText(c, opts.Prefix, k.TextMuted)
		}
		in := bareInput(c, value, opts.Label, opts.Placeholder).
			Grow(1).MinHeight(core.ControlHeight(c) - u*3)
		if opts.Password {
			in.Password()
		}
		dressInput(in, opts.Error, opts.ReadOnly)
		if opts.Suffix != "" {
			fieldText(c, opts.Suffix, k.TextMuted)
		}
		if opts.Trailing != nil {
			opts.Trailing()
		}
		return in
	})
}

// copyKey is where a CopyButton remembers until when it should stop showing
// that it did its work.
type copyKey struct{}

// copiedFor is how long a CopyButton shows its tick: long enough to be seen
// by somebody who looked away for the click, short enough that it is not
// still ticking when the next thing happens.
const copiedFor = 1400 * time.Millisecond

// CopyButtonOptions configure a CopyButton.
type CopyButtonOptions struct {
	// Label names the button; it is required, like every control here that
	// shows nothing but a mark. It should say what is copied rather than
	// what the mark is: "Copy callback URL", not "Copy".
	Label string
	// Icon draws instead of the two sheets.
	Icon *ui.SVG
	// Plain draws the mark on nothing, for a button that belongs to the
	// field it sits in rather than to a toolbar — a plate inside a well is
	// a plate inside a plate.
	Plain bool
	// Disabled takes the button out of play. Copying an empty string is a
	// real thing to want — it is how a clipboard is emptied — so nothing is
	// disabled on its own; a caller who would rather the button were dead
	// on an empty value says so here.
	Disabled bool
}

// CopyButtonResult carries a CopyButton and whether it copied.
type CopyButtonResult struct {
	// Element is the button.
	Element *ui.Element
	copied  bool
}

// Copied reports that the press put the value on the clipboard this frame.
func (r CopyButtonResult) Copied() bool { return r.copied }

// CopyButton puts the string the caller points at on the clipboard, and says
// that it did: the mark becomes a tick for a moment and a screen reader is
// told, because a button that gives no sign of having worked is a button
// people press twice.
//
// While it is showing that, it answers to a different name — the word for
// having done it — which is how a reader hears it and how a test knows to
// stop clicking.
//
// What lands on the clipboard is the caller's string as it stands, not
// whatever it said when the button was built.
func CopyButton(c *ui.Context, text *string, opts CopyButtonOptions) CopyButtonResult {
	if text == nil {
		panic("input: CopyButton needs a text to point at; it copies nothing of its own")
	}
	if opts.Label == "" {
		panic("input: CopyButton needs a Label, or nothing can name the button")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	copied := core.Msg(c, "input.copied", "Copied")
	mark := glyphCopy
	if opts.Icon != nil {
		mark = opts.Icon
	}
	side := u * 11
	var r CopyButtonResult

	btn := ui.ButtonBase(c).Size(side, side).Shrink(0).Radius(theme.PillRadius).
		Label(opts.Label).Tooltip(opts.Label).Disabled(opts.Disabled)
	if !opts.Plain {
		btn.Background(k.SurfaceHover)
	}
	btn.Children(func() {
		// One store for the tick, so the mark and the frame that takes it
		// away cannot disagree about whether it is showing.
		until := ui.Local(btn, copyKey{}, func() time.Time { return time.Time{} })
		shown := time.Now().Before(*until)

		if btn.Clicked() {
			c.WriteClipboard(*text)
			r.copied = true
			*until = time.Now().Add(copiedFor)
			shown = true
			c.Announce(copied)
		}
		glyph := mark
		if shown {
			glyph = glyphTick
			btn.Label(copied).Tooltip(copied)
			// The frame that takes the tick away is asked for by the clock,
			// not by anything the user did, and it is asked for again every
			// frame it is still shown, so the mark cannot outlive its moment.
			c.After(time.Until(*until))
		}
		ui.Icon(c, glyph).TextColor(k.TextMuted).Size(u*4.25, u*4.25)
	})
	r.Element = btn
	return r
}
