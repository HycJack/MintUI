package display

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// HeadingOptions configure a Heading.
type HeadingOptions struct {
	// Level is 1 for the page's own heading, 2 for a panel's, 3 for a
	// section's. Anything above 3 is 3: past that it stops being a heading
	// and starts being text that happens to be large.
	Level int
	// Subtitle sits under the heading in the muted tone.
	Subtitle string
	// Divider draws a rule under it.
	Divider bool
}

// Heading is a titled block. It exists so a caller never reaches for a font
// size: "the title of a panel" is a role, and the sizes that go with roles
// live in one place.
func Heading(c *ui.Context, title string, opts HeadingOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Level < 1 {
		opts.Level = 2
	}
	if opts.Level > 3 {
		opts.Level = 3
	}
	size := theme.BodySize
	switch opts.Level {
	case 1:
		size = theme.DisplaySize
	case 2:
		size = theme.TitleSize
	}

	e := ui.Column(c).FillWidth().Gap(u).Children(func() {
		ui.Text(c, title).TextColor(k.Text).FontSize(core.FontSize(c, size)).Bold()
		if opts.Subtitle != "" {
			ui.Text(c, opts.Subtitle).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		}
	})
	if opts.Divider {
		e.Children(func() {
			ui.Box(c).FillWidth().Height(theme.BorderWidth).MarginY(u).Background(k.Border)
		})
	}
	return e
}

// TextOptions configure a Text.
type TextOptions struct {
	// MaxLines truncates with an ellipsis past this many lines; zero is
	// unlimited.
	MaxLines int
	// Muted, Faint and Bold pick from the palette rather than a raw colour.
	Muted, Faint, Bold bool
	// Mono uses the monospaced face, for a reference or a path.
	Mono bool
}

// Text is body copy. Every string in the interface goes through here rather
// than ui.Text directly, so the type scale and the muted tone are decided in
// one place.
func Text(c *ui.Context, s string, opts TextOptions) *ui.Element {
	k := core.Tokens(c)
	col := k.Text
	switch {
	case opts.Faint:
		col = k.TextFaint
	case opts.Muted:
		col = k.TextMuted
	}
	size := theme.BodySize
	if opts.Mono {
		size = theme.MonoSize
	}
	t := ui.Text(c, s).TextColor(col).FontSize(core.FontSize(c, size))
	if opts.Bold {
		t.Bold()
	}
	if opts.MaxLines > 0 {
		t.MaxLines(opts.MaxLines)
	}
	return t
}

// EditableTextOptions configure an EditableText.
type EditableTextOptions struct {
	// Placeholder shows when the value is empty.
	Placeholder string
	// Label names the field for assistive technology; an editable field with
	// no name is a mystery box.
	Label string
	// Multiline lets the value hold newlines.
	Multiline bool
	// Muted draws the value in the secondary tone, for a field whose value
	// is not the point.
	Muted bool
}

// EditableTextView carries an EditableText and what happened to it.
type EditableTextView struct {
	// Element is the field.
	Element *ui.Element
	// Committed reports that the user finished editing this frame.
	Committed bool
	// value is this frame's text.
	value string
}

// Value returns the field's current text.
func (v EditableTextView) Value() string { return v.value }

// EditableText is a text field whose value is the caller's. It is the same
// primitive as a form field, without the label above it.
func EditableText(c *ui.Context, value *string, opts EditableTextOptions) EditableTextView {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Label == "" {
		panic("display: EditableText needs a Label")
	}
	var v EditableTextView
	col := k.Text
	if opts.Muted {
		col = k.TextMuted
	}

	e := ui.Box(c).FillWidth().Padding(u*1.25, u*2).Radius(theme.ControlRadius).
		Background(k.Surface).BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Label(opts.Label).Children(func() {
		ui.Column(c).Grow(1).Gap(0).Children(func() {
			t := ui.Text(c, *value).TextColor(col).
				FontSize(core.FontSize(c, theme.RowSize))
			if opts.Multiline {
				t.MaxLines(0)
			}
		})
	})
	v.Element = e
	v.value = *value
	return v
}

// Kbd draws a keyboard shortcut: the keys, joined, in a small fixed box each.
//
// A shortcut written as plain text — "Cmd+K" — reads as one word, so it is
// impossible to see where one key ends and the next begins.
func Kbd(c *ui.Context, keys ...string) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	return ui.Row(c).AlignItems(ui.Center).Gap(u * 0.75).Children(func() {
		for _, key := range keys {
			ui.Box(c).Padding(u*0.75, u*1.25).Radius(theme.ControlRadius * 0.5).
				Background(k.Surface).BorderWidth(theme.BorderWidth).BorderColor(k.Border).
				Children(func() {
					ui.Text(c, key).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600)
				})
		}
	})
}

// LinkOptions configure a Link.
type LinkOptions struct {
	// URL is what a press opens. It is shown nowhere: a link that prints its
	// own address is a link nobody needs to click.
	URL string
	// Muted draws it in the secondary tone, for a link in running text.
	Muted bool
}

// Link is text that does something when pressed. It is always underlined:
// colour alone would leave it indistinguishable to anyone who cannot see it.
func Link(c *ui.Context, label string, opts LinkOptions) *ui.Element {
	k := core.Tokens(c)
	col := k.AccentText
	if opts.Muted {
		col = k.TextMuted
	}
	t := ui.Text(c, label).TextColor(col).
		FontSize(core.FontSize(c, theme.BodySize)).Cursor(ui.CursorPointer)
	if opts.URL != "" {
		t.Label(label)
	}
	return t
}

// TagView carries a Tag and what the user did to it.
type TagView struct {
	// Element is the tag.
	Element *ui.Element
	// closed reports a press of the close button.
	closed bool
}

// Closed reports a press of the tag's close button.
func (v TagView) Closed() bool { return v.closed }

// TagOptions configure a Tag.
type TagOptions struct {
	// Tone picks the tag's severity; Neutral is the quiet default.
	Tone core.Severity
	// Icon draws a mark before the label.
	Icon *ui.SVG
	// Closable draws a close button, which reports through TagView.Closed.
	Closable bool
	// Selected fills the tag, for a filter chip that is on.
	Selected bool
	// Disabled greys it out and blocks the press.
	Disabled bool
}

// Tag is a small label for a thing's state: a branch, a label, a filter. It is
// quieter than a Badge, which counts something.
func Tag(c *ui.Context, label string, opts TagOptions) TagView {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Tone == 0 {
		opts.Tone = core.Neutral
	}
	bg, fg := opts.Tone.Pair(k)
	if opts.Disabled {
		bg, fg = k.Surface, k.TextFaint
	}

	var v TagView
	// Row, not Box: ui.Box is ui.Column, so a Box here would stack the
	// icon, the label and the close button down the pill instead of across it.
	e := ui.Row(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).Background(bg).
		AlignItems(ui.Center).Gap(u).Children(func() {
		if opts.Icon != nil {
			sz := core.FontSize(c, theme.CaptionSize)
			ui.Icon(c, opts.Icon).Size(sz, sz).TextColor(fg)
		}
		ui.Text(c, label).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize))
		if opts.Closable {
			if IconClose(c, u).Clicked() {
				v.closed = true
			}
		}
	})
	if opts.Disabled {
		e = e.Disabled(true)
	}
	v.Element = e
	return v
}

// IconClose is the small cross a closable tag carries. It is its own component
// because it is the one mark that has to read at 10 points: too thin and it
// disappears, too heavy and it competes with the label.
func IconClose(c *ui.Context, u float32) *ui.Element {
	k := core.Tokens(c)
	s := u * 2.5
	return ui.Box(c).Size(s, s).Label("Remove").
		Cursor(ui.CursorPointer).Children(func() {
		ui.Box(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
			t := r.W * 0.09
			internal.Rule(p, ui.Rect{X: r.X + t, Y: r.Y + t, W: r.W - 2*t, H: r.H - 2*t}, k.TextMuted)
		})
	})
}
