package input

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// The form: the box its fields go in, the box one field goes in, and the
// notice that goes above them.
//
// The label sits above its control here and not beside it. That is a
// deliberate departure from a native settings pane, where labels line up in
// a column on the left: those panes are as wide as the window, while a form
// in this interface is most often in a drawer, and a long label — "Repeat
// failures in the last 30 days" — in a 200px column of labels is a label
// that wraps into three lines and misaligns every field below it. Above the
// control it wraps against the control's own width, which is the width it
// has to fit in anyway. The drawer is not a special case; it is the reason.

// FormOptions configure a Form.
type FormOptions struct {
	// Label names the form for assistive technology, which lands in it as a
	// group. It is required: a reader arriving at a field has to be able to
	// ask what the fields around it are for.
	Label string
	// Title heads the form, and Description says under it. Neither is
	// needed when the sheet or the panel the form is in already says it.
	Title, Description string
	// Fields is the order the fields are built in, and the order they are
	// read in: the tab order and the visual order are the same, because
	// anything else makes the keyboard walk a form nobody can see.
	//
	// Every control in this package takes the focus in the order it is
	// built, so putting the fields in reading order is all that is needed
	// for Tab to follow.
	Fields func()
	// Actions is the submit row under the fields. Empty draws none, and the
	// form is then only as tall as its fields.
	Actions func()
	// Divider draws a rule above the actions, which is what separates "what
	// this form is" from "what you do about it".
	Divider bool
	// Gap is the space between fields; zero is the density's own step,
	// which is what a form of six fields wants.
	Gap float32
	// Width bounds the form, for a drawer whose panel is wider than the
	// fields want to be.
	Width float32
}

// FormResult carries a Form and what was done in it.
type FormResult struct {
	// Element is the whole form, heading, fields and actions.
	Element *ui.Element
}

// Form is a column of fields with a submit row under it — the arrangement
// every record in this interface is edited in.
//
// It owns the spacing and the rule and nothing else. The values belong to
// the caller, the actions are the caller's buttons, and the rules that a
// particular field obeys are that field's business: Form is what holds the
// fields together, so it stays out of them.
func Form(c *ui.Context, opts FormOptions) FormResult {
	if opts.Label == "" {
		panic("input: Form needs a Label; a reader arriving at a field has to " +
			"be able to ask what the fields around it are for")
	}
	u := core.Density(c).Unit()
	gap := opts.Gap
	if gap <= 0 {
		gap = u * 4
	}

	var r FormResult
	form := ui.Column(c).FillWidth().Gap(gap).Role(ui.RoleGroup).Label(opts.Label)
	if opts.Width > 0 {
		form.Width(opts.Width).Shrink(0)
	}
	form.Children(func() {
		if opts.Title != "" {
			ui.Text(c, opts.Title).TextColor(core.Tokens(c).Text).
				FontSize(core.FontSize(c, theme.SheetSize)).Bold()
		}
		if opts.Description != "" {
			ui.Text(c, opts.Description).TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		}
		if opts.Fields != nil {
			ui.Column(c).FillWidth().Gap(gap).Children(opts.Fields)
		}
		if opts.Actions != nil {
			if opts.Divider {
				ui.Box(c).FillWidth().Height(theme.BorderWidth).
					Margin(gap/2, 0, 0, 0).Background(core.Tokens(c).Border)
			}
			ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(opts.Actions)
		}
	})
	r.Element = form
	return r
}

// FormFieldOptions configure a FormField.
type FormFieldOptions struct {
	// Label names the field, above it. It is required, and it is the text a
	// person reads: the control carries the same string for assistive
	// technology, and the two are meant to be the same words.
	Label string
	// Description says what the field wants, under it — "Whole minutes",
	// "As it appears on the invoice". It is the help text, so it reads
	// below the control and before any error.
	Description string
	// Error is what is wrong with the value. It is handed to the control,
	// which turns its hairline to the danger colour, and shown here as the
	// words. The message is written once, by the field; the outline is
	// drawn once, by the box the control sits in.
	Error string
	// Required marks a field that must be filled in. The mark is a word
	// rather than an asterisk, because an asterisk means nothing to
	// somebody who has not met one and means nothing at all to a screen
	// reader.
	Required bool
	// Gap is the space between the label, the control and the texts under
	// it; zero is the density's own smaller step, because these are one
	// thing rather than three.
	Gap float32
}

// FormField is one field: the label above it, the control under that, and
// whatever the field has to say below.
//
// The control is built by control, and it is handed the field's error: a
// field that drew its own hairline and was also told about the error would
// end up with two borders, or with none.
//
//	FormField(c, FormFieldOptions{Label: "Open cost", Error: app.costError},
//	    func(err string) *ui.Element {
//	        return input.CurrencyInput(c, &app.cost, input.CurrencyInputOptions{
//	            Label: "Open cost", Error: err,
//	        })
//	    })
//
// control takes an error argument rather than the caller repeating the
// message in two places, because a message said twice is a message that can
// disagree with itself: the field would show one error and explain another.
func FormField(c *ui.Context, opts FormFieldOptions, control func(err string) *ui.Element) *ui.Element {
	if opts.Label == "" {
		panic("input: FormField needs a Label; it is the text a person reads to " +
			"know what the field is for")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	gap := opts.Gap
	if gap <= 0 {
		gap = u * 1.5
	}
	required := core.Msg(c, "input.required", "Required")

	field := ui.Column(c).FillWidth().AlignItems(ui.Start).Gap(gap).Role(ui.RoleGroup)
	field.Children(func() {
		ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center).Children(func() {
			ui.Text(c, opts.Label).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).FontWeight(600)
			if opts.Required {
				// The word, not an asterisk: a mark is only a mark to
				// somebody who was told what it means. It sits next to the
				// label rather than at the far edge of the field, where a
				// word that qualifies the label would be read as belonging
				// to whatever happened to be last.
				ui.Text(c, required).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})

		var ctrl *ui.Element
		if control != nil {
			ctrl = control(opts.Error)
		}
		if ctrl != nil && opts.Error != "" {
			// The field itself marks the control; this is for a control that
			// is not one of ours, and for a group of them, where the error
			// belongs to the group rather than to any one part of it.
			ctrl.Error(opts.Error)
		}

		if opts.Description != "" {
			ui.Text(c, opts.Description).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
		}
		if opts.Error != "" {
			// Danger on DangerBg, the one pair in the palette meant to be
			// read as a warning, so the error is not a sentence in danger
			// red that has to be read carefully to be found at all.
			ui.Row(c).FillWidth().Padding(gap*0.75, gap, gap*0.75, gap).
				Radius(theme.SmallRadius).Background(k.DangerBg).Gap(u).
				AlignItems(ui.Start).Children(func() {
				// The mark repeats the words in shape as well as in colour,
				// which is the one thing that survives somebody who cannot
				// see the colour. It is nudged down to sit on the first
				// line rather than centred against all of them.
				ui.Icon(c, glyphDanger).TextColor(k.Danger).
					Size(u*4, u*4).Margin(gap*0.25, 0, 0, 0)
				ui.Text(c, opts.Error).TextColor(k.Danger).
					FontSize(core.FontSize(c, theme.MetaSize)).Grow(1)
			})
		}
	})
	return field
}

// BannerOptions configure a Banner.
type BannerOptions struct {
	// Text is what the banner says. It is required, and it should be one
	// sentence: a banner is read before anything around it and cannot be
	// skipped once it is noticed.
	Text string
	// Severity picks the mark's colour. Neutral is the quiet default, and
	// is the right answer for most of what a form has to say — a draft
	// that is not saved yet, a hint about what the next field wants.
	Severity core.Severity
	// Icon draws instead of the mark the severity asks for.
	Icon *ui.SVG
	// Action labels a button that does something about it — "Retry", "Read
	// more". Empty draws none: a banner is usually just saying something,
	// and a button on every one of them teaches people to ignore them.
	Action string
	// Dismissable adds a close button, reported through BannerResult.
	Dismissable bool
}

// BannerResult carries a Banner and what was pressed in it.
type BannerResult struct {
	// Element is the whole banner.
	Element             *ui.Element
	dismissed, actioned bool
}

// Dismissed reports a press of the close button.
func (r BannerResult) Dismissed() bool { return r.dismissed }

// Actioned reports a press of the action button.
func (r BannerResult) Actioned() bool { return r.actioned }

// Banner is the notice that goes above a form: why it is there, what is
// wrong with it, what it is about to do.
//
// It is quieter than an alert, and deliberately so — the difference is where
// it lives. An alert interrupts the content it is in; a banner is part of
// the form, read in the same breath as the fields, and a strip of colour at
// the top of a form would be shouting at the person who has not done
// anything wrong yet. The severity tints one mark rather than the whole
// surface, which is the same rule the pills follow.
func Banner(c *ui.Context, opts BannerOptions) BannerResult {
	if opts.Text == "" {
		panic("input: Banner needs Text; an empty banner is a coloured box " +
			"that looks like it is about to say something")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	_, ink := opts.Severity.Pair(k)

	mark := opts.Icon
	if mark == nil {
		switch opts.Severity {
		case core.Danger:
			mark = glyphDanger
		case core.Warning:
			mark = glyphWarn
		case core.Success:
			mark = glyphTick
		default:
			mark = glyphInfo
		}
	}

	var r BannerResult
	banner := ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u*2).
		Padding(u*2, u*3, u*2, u*3).Radius(theme.ControlRadius).
		Background(k.Surface).BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Role(ui.RoleGroup).Label(opts.Text)
	banner.Children(func() {
		// Before the words, because a mark that repeats them in shape as
		// well as in colour is the one thing that survives somebody who
		// cannot see the colour at all.
		ui.Box(c).Size(u*5, u*5).Shrink(0).Radius(u * 1.25).Background(k.Surface).
			Children(func() {
				ui.Icon(c, mark).TextColor(ink).Size(u*3.25, u*3.25)
			})
		ui.Text(c, opts.Text).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
		if opts.Dismissable {
			dismiss := ui.Box(c).Size(u*6, u*6).Shrink(0).Role(ui.RoleButton).
				Label(core.Msg(c, "input.dismiss", "Dismiss")).
				Tooltip(core.Msg(c, "input.dismiss", "Dismiss")).
				Cursor(ui.CursorPointer).Children(func() {
				ui.Icon(c, glyphCross).TextColor(k.TextMuted).Size(u*3.25, u*3.25)
			})
			r.dismissed = dismiss.Clicked()
		}
		if opts.Action != "" {
			btn := ui.ButtonBase(c).Height(core.ControlHeight(c)-u*2).Shrink(0).
				Radius(theme.PillRadius).Padding(0, u*3.5, 0, u*3.5).
				Background(k.Background).TextColor(k.Text).
				BorderWidth(theme.BorderWidth).BorderColor(k.Border).
				Label(opts.Action).Tooltip(opts.Action).
				Children(func() {
					ui.Text(c, opts.Action).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.RowSize))
				})
			r.actioned = btn.Clicked()
		}
	})
	r.Element = banner
	return r
}
