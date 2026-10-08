package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The field box: the one piece every text control in this package is built
// on. A search box, a money field and a one-time code differ in what goes
// inside the well, never in how the well looks, so the ground, the hairline,
// the height at a density and the type inside are written once here.

// fieldSkin is what the box around a text control has to know. It is one
// struct for the same reason the box is one: a control that styled its own
// ground could not stay in step with the others, and a form of six fields
// would come out of it looking like six designs.
type fieldSkin struct {
	// label names the control for assistive technology and for tests to
	// find it by. Every field in this package requires one.
	label string
	// err is the validation message. Empty means the value is fine; a
	// message turns the hairline to the danger colour and is read out
	// before the field's own description.
	err string
	// disabled greys the whole field and takes it out of the tab order.
	disabled bool
	// readOnly keeps the value selectable and copyable but not editable,
	// for a field the app fills in and the user cannot change.
	readOnly bool
	// width is the well's own width; zero fills what the row gives it.
	width float32
	// height is the well's height; zero is a standard control at this
	// window's density.
	height float32
	// lines is how many lines a text area shows; one makes it a field.
	lines int
	// padY overrides the well's vertical padding.
	padY float32
	// mono draws the control in the system's monospace face at the
	// monospace size, for code and fragments: the text is not prose, and
	// prose type says nothing about it. It is what a TextInputEditor is.
	mono bool
}

// textWell is the box a single-line control sits in.
func textWell(c *ui.Context, s fieldSkin, build func(w *ui.Element) *ui.Element) *ui.Element {
	return fieldWell(c, s, false, build)
}

// areaWell is textWell's counterpart for a text area: the same box, with the
// control as tall as its lines rather than as tall as one line.
func areaWell(c *ui.Context, s fieldSkin, build func(w *ui.Element) *ui.Element) *ui.Element {
	return fieldWell(c, s, true, build)
}

// fieldWell is the box a text control sits in: the ground, the height a
// control has at this density, the type inside it, and the outline. build
// fills it and returns the element that takes the typing; the well focuses
// that element when the user presses anywhere in it, so the padding around
// the text reads as part of the field rather than as a margin beside it.
//
// Role is RoleNone because the well is furniture: assistive technology
// should meet the control inside it, not a group that presses like a button.
func fieldWell(c *ui.Context, s fieldSkin, column bool, build func(w *ui.Element) *ui.Element) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	padY := s.padY
	if padY <= 0 {
		padY = u * 1.5
	}
	h := s.height
	if h <= 0 {
		h = core.ControlHeight(c)
		if column {
			// A line of text and the padding above and below it, which is
			// the same height a single-line well gives its control: the
			// text does not grow when the field does, the count of lines
			// does.
			lines := s.lines
			if lines < 1 {
				lines = 4
			}
			lineH := lineHeight(c)
			if s.mono {
				lineH = monoLineHeight(c)
			}
			h = float32(lines)*lineH + padY*2
		}
	}

	var well *ui.Element
	if column {
		well = ui.Column(c).Gap(0)
	} else {
		well = ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5)
	}
	well = well.Padding(padY, u*2.5, padY, u*2.5).Radius(theme.ControlRadius).
		Background(k.Surface).FontSize(core.FontSize(c, theme.BodySize)).
		MinHeight(h).FillWidth().Cursor(ui.CursorText).Role(ui.RoleNone)
	if s.mono {
		// The editor inside takes its face from the well, as every other
		// text control here does: the type is decided once, at the box.
		well.Font("monospace").FontSize(core.FontSize(c, theme.MonoSize))
	}
	if s.width > 0 {
		well.Width(s.width).Shrink(0)
	}

	var in *ui.Element
	well.Children(func() {
		if s.disabled {
			// Disabled walks up the tree for painting and for presses, but
			// the focus a click picks is the element's own flag, so the
			// control inside has to be disabled in its own right too.
			well.Disabled(true)
		}
		if build != nil {
			in = build(well)
		}
		if in == nil {
			return
		}
		if s.disabled {
			in.Disabled(true).ReadOnly(true)
		}
		// A press in the padding — beside the caret, under the placeholder,
		// on the unit — belongs to the field, not to whatever the press
		// happened to land on.
		if well.Clicked() {
			in.Focus()
		}
		ringField(c, well, in, s.err != "")
	})
	return well
}

// ringField gives a field's box its hairline: the danger colour while the
// field carries an error, the accent while it has the focus, a shade darker
// under the pointer, and the ordinary border otherwise.
//
// It is the only place a field outline is drawn. A field marked by FormField
// and a field marked through its own Error option therefore come out the
// same, which is what stops a form from growing two borders per field.
func ringField(c *ui.Context, box, in *ui.Element, invalid bool) {
	k := core.Tokens(c)
	col := k.Border
	switch {
	case invalid:
		col = k.Danger
	case in == nil:
	case in.IsDisabled():
		col = k.Border.Alpha(0.7)
	case in.Focused():
		col = k.Accent
	case box.Hovered():
		col = k.Border.Mix(k.Text, 0.25)
	}
	box.BorderWidth(theme.BorderWidth).BorderColor(col)
}

// bareInput is a text control with MyGo's editor and none of its chrome, so
// the well is the only thing around it that paints.
func bareInput(c *ui.Context, value *string, label, placeholder string) *ui.Element {
	in := ui.TextInputBase(c, value).Label(label)
	if placeholder != "" {
		in.Placeholder(placeholder)
	}
	return in
}

// lineHeight is how tall one line of a field's own text is, which is what a
// text area is measured in lines by.
func lineHeight(c *ui.Context) float32 {
	return core.FontSize(c, theme.BodySize) * 1.45
}

// monoLineHeight is the same measure for the monospace face, which is what a
// code editor is measured in.
func monoLineHeight(c *ui.Context) float32 {
	return core.FontSize(c, theme.MonoSize) * 1.45
}

// fieldText is the small label text a control carries: a unit after a number,
// the hint inside a group.
func fieldText(c *ui.Context, s string, col ui.Color) *ui.Element {
	return ui.Text(c, s).TextColor(col).
		FontSize(core.FontSize(c, theme.RowSize)).Shrink(0)
}

// ── glyphs ─────────────────────────────────────────────────────────────────
//
// Drawn as strokes on a 24-unit grid at the current colour, so they take the
// palette rather than carrying colours of their own, and keep the same
// weight as each other at every size a control uses.

func mustGlyph(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	glyphSearch = mustGlyph(`<circle cx="10.5" cy="10.5" r="6"/><path d="M15 15l4.5 4.5"/>`)
	glyphEye    = mustGlyph(`<path d="M2.5 12S6 5.75 12 5.75 21.5 12 21.5 12 18 18.25 12 18.25 2.5 12 2.5 12Z"/><circle cx="12" cy="12" r="2.75"/>`)
	glyphBlind  = mustGlyph(`<path d="M4 12s3.2-5.25 8-5.25c1.3 0 2.5.4 3.5 1M20 12s-1.1 1.8-2.9 3.2"/><path d="M6.2 6.9C4 8.3 2.5 12 2.5 12s3.5 6.25 9.5 6.25c1.2 0 2.3-.25 3.3-.7M9.9 9.95a2.75 2.75 0 0 0 4.2 3.2M4 4l16 16"/>`)
	glyphCopy   = mustGlyph(`<rect x="9" y="9" width="11" height="11" rx="2.5"/><path d="M15 6.4A2.5 2.5 0 0 0 12.5 4H6.5A2.5 2.5 0 0 0 4 6.5v6A2.5 2.5 0 0 0 6.5 15"/>`)
	glyphTick   = mustGlyph(`<path d="M5 12.5 9.75 17.5 19 7"/>`)
	glyphCross  = mustGlyph(`<path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/>`)
	glyphLess   = mustGlyph(`<path d="M6 12h12"/>`)
	glyphMore   = mustGlyph(`<path d="M12 6v12M6 12h12"/>`)
	glyphInfo   = mustGlyph(`<circle cx="12" cy="12" r="8.5"/><path d="M12 11.25V17"/><path d="M12 7.6v.6"/>`)
	glyphWarn   = mustGlyph(`<path d="M12 4.75 21 19.5H3Z"/><path d="M12 10.5V14.5"/><path d="M12 17v.6"/>`)
	glyphDanger = mustGlyph(`<circle cx="12" cy="12" r="8.5"/><path d="M12 7.5v5.5"/><path d="M12 16.4v.6"/>`)
)
