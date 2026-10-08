package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// DialogOptions configure a Dialog.
type DialogOptions struct {
	// Title heads the dialog; empty draws a panel of body alone.
	Title string
	// Subtitle is the second line under the title: what the dialog is for.
	Subtitle string
	// Body is the content. Nil draws a panel of only the actions, which is
	// what a dialog that only asks something looks like.
	Body func()
	// Actions is the button row, drawn right-aligned under the body and
	// given room to sit apart from it.
	Actions func()
	// Width is the panel's width; zero lets it fit its content.
	Width float32
	// MaxWidth caps it; zero leaves it to the window.
	MaxWidth float32
	// Rule draws a hairline under the title, for a form under a heading.
	Rule bool
	// NonModal leaves the window behind live: a press outside the panel goes
	// to the page rather than to the scrim. A dialog blocks it by default.
	NonModal bool
}

// Dialog is a panel over a window asking for something: a form, a choice, a
// confirmation the caller wants room to word.
//
// It is modal unless NonModal is set — the scrim covers the window, the page
// behind goes inert, and a press on the scrim or Escape writes false into the
// caller's *bool. There is no result, because those are the only two things a
// person does to a dialog that is not one of its own buttons, and the buttons
// report themselves.
//
//	core.Use(c, core.Settings{})
//	Dialog(c, &app.asking, DialogOptions{
//	    Title: "Log callback",
//	    Body:  func() { fields() },
//	    Actions: func() {
//	        if input.Button(c, "Cancel", input.ButtonOptions{}).Clicked() {
//	            app.asking = false
//	        }
//	    },
//	})
//
// Dialog returns nil while closed, and asks for nothing when it is: a modal
// layer takes part in the layout whether or not it is showing, so one built to
// be closed spends a frame's layout on a panel nobody can see. The caller who
// needs the call in the same place as the rest of the view guards it the way
// internal/board guards ui.Modal.
func Dialog(c *ui.Context, open *bool, opts DialogOptions) *ui.Element {
	if opts.Body == nil && opts.Actions == nil {
		// A panel with no content is what a caller gets from forgetting the
		// body, and it is indistinguishable from a layer that has nothing to
		// say — which is exactly the case worth stopping for.
		panic("overlay: Dialog needs a Body or Actions; an empty panel says nothing")
	}
	u := core.Density(c).Unit()

	panel := layer(c, open, !opts.NonModal, ui.Center, func(_, p *ui.Element) {
		Panel(c, p, PanelOptions{
			Title:     opts.Title,
			Subtitle:  opts.Subtitle,
			Rule:      opts.Rule,
			MaxWidth:  opts.MaxWidth,
			TitleSize: theme.SheetSize,
		}, func() {
			if opts.Body != nil {
				opts.Body()
			}
			if opts.Actions != nil {
				ui.Row(c).FillWidth().Gap(u*2).Justify(ui.End).Margin(u*2.5, 0, 0, 0).
					Children(opts.Actions)
			}
		})
	})
	if panel == nil {
		return nil
	}
	if opts.Width > 0 {
		panel.Width(opts.Width)
	}
	return panel
}

// AlertDialogOptions configure an AlertDialog.
type AlertDialogOptions struct {
	// Title is the question, as a heading: "Delete “Notes”?"
	Title string
	// Body is one sentence saying what follows from the answer.
	Body string
	// Actions names the buttons from left to right. It is required, and the
	// last of them is the default — the one Escape picks and Enter takes the
	// focus to.
	Actions []string
	// Destructive marks the default as the one that destroys something, so it
	// is drawn in the danger colour rather than as the plain affirmative.
	Destructive bool
	// NonModal leaves the window behind live.
	NonModal bool
}

// AlertDialogResult carries an AlertDialog and what the user answered.
type AlertDialogResult struct {
	// Element is the whole alert, or nil while it is closed.
	Element *ui.Element
	chosen  int
	// answered records that a button was pressed this frame, which is what
	// separates "Cancel, the second one" from "nothing was pressed".
	answered bool
}

// Chosen returns the index in AlertDialogOptions.Actions of the button that
// was pressed, or -1 in a frame in which none was.
func (r AlertDialogResult) Chosen() int {
	if !r.answered {
		return -1
	}
	return r.chosen
}

// AlertDialog is a dialog that asks about one thing and is answered with a
// button, as AppKit's: a title, a sentence, and the buttons.
//
// It is modal unless NonModal is set, and Escape and a press on the scrim
// pick the button labelled "Cancel" — the answer a dismissing key is expected
// to mean — writing false into the caller's *bool. An alert with no Cancel
// cannot be dismissed that way; that is a decision about what the answer
// means, so it belongs to the caller rather than to this component.
//
// Escape is read from the backdrop rather than left to layer, because an
// alert has to answer a dismissal with the button a person meant rather than
// with nothing happening.
func AlertDialog(c *ui.Context, open *bool, opts AlertDialogOptions) AlertDialogResult {
	if open == nil {
		panic("overlay: AlertDialog needs the *bool it opens and closes")
	}
	if len(opts.Actions) == 0 {
		panic("overlay: AlertDialog needs at least one action; a dialog nobody can answer is not a dialog")
	}
	if !*open {
		return AlertDialogResult{}
	}

	cancel := -1
	for i, a := range opts.Actions {
		if a == cancelLabel {
			cancel = i
		}
	}

	res := AlertDialogResult{}
	panel := layer(c, open, !opts.NonModal, ui.Center, func(back, p *ui.Element) {
		if cancel >= 0 && back.OverlayShortcut(0, ui.KeyEscape) {
			res.chosen, res.answered = cancel, true
			*open = false
		}
		Panel(c, p, PanelOptions{
			Title:     opts.Title,
			Subtitle:  opts.Body,
			Role:      ui.RoleAlertDialog,
			TitleSize: theme.SheetSize,
		}, func() {
			u := core.Density(c).Unit()
			row := ui.Row(c).FillWidth().Gap(u*2).Justify(ui.End).
				Margin(u*2.5, 0, 0, 0)
			row.Children(func() {
				for i, label := range opts.Actions {
					btn := alertButton(c, label, i == len(opts.Actions)-1, opts.Destructive)
					if btn.Clicked() {
						res.chosen, res.answered = i, true
					}
				}
			})
			// The message is what the alert is for, so it is what a screen
			// reader reads after the title rather than only something shown.
			row.Description(opts.Body)
		})
	})
	res.Element = panel
	return res
}

// alertButton draws one button of an alert.
//
// The default — the last — takes the focus, so Enter answers without the
// pointer, and carries the danger colour when the answer destroys something:
// an alert exists because this answer is not like the others, and drawing it
// the same as Cancel would say it was.
func alertButton(c *ui.Context, label string, default_, destructive bool) *ui.Element {
	switch {
	case !default_:
		return input.Button(c, label, input.ButtonOptions{Label: label})
	case destructive:
		return input.Button(c, label, input.ButtonOptions{Danger: true, Label: label}).AutoFocus()
	default:
		return input.Button(c, label, input.ButtonOptions{Primary: true, Label: label}).AutoFocus()
	}
}

// cancelLabel is the button Escape picks in an alert. It is matched by name
// rather than by place, because which button dismisses is a property of what
// it is called and not of where it sits in the row.
const cancelLabel = "Cancel"
