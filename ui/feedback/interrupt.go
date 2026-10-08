package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// InterruptButtonOptions configure an InterruptButton.
type InterruptButtonOptions struct {
	// Label names the button. Empty takes the library's "Stop", which is
	// right for the run being stopped and wrong for the thing being
	// cancelled — "Cancel upload" and "Stop" are different promises, and a
	// button that keeps its name when the action changes is how people stop
	// pressing it.
	Label string
	// Busy means the interrupt has been asked for and is on its way. The
	// button stops taking presses and says what is happening instead.
	//
	// That is the whole reason this is a component rather than a button in
	// the caller: the state where the press has been taken but the run has
	// not stopped is the state people click twice in, and a control that
	// looks the same in it invites the second click.
	Busy bool
	// BusyLabel is what the button says while Busy. Empty takes the
	// library's "Stopping…".
	BusyLabel string
	// Name is what assistive technology announces. Empty takes the label, or
	// the busy label while busy.
	Name string
	// Caution draws the button in the window's ordinary text rather than the
	// danger colour, for stopping something harmless — a sync, a download, a
	// preview. It is off by default, because the button exists to stop
	// something and the colour that says so is the safe default.
	//
	// It is a flag and not a Severity because a Severity's zero value is
	// core.Neutral, which would quietly make "the dangerous default" the
	// undangerous one for every caller who says nothing.
	Caution bool
}

// InterruptButtonResult carries an InterruptButton and its press.
type InterruptButtonResult struct {
	// Element is the whole button.
	Element *ui.Element
	pressed bool
}

// Pressed reports the button. It is false while Busy: the press has already
// been taken, and a caller that acted on it again would be sending the second
// interrupt this component exists to prevent.
func (r InterruptButtonResult) Pressed() bool { return r.pressed }

// InterruptButton is the button that stops a thing that is running: an agent
// on a long task, a deploy halfway out, a sync grinding through a backlog.
//
// It is the only button in this library with a state between the press and
// the outcome, and it says so in the label rather than only in the disable:
// "Stopping…" tells the reader their press landed, where a greyed-out "Stop"
// tells them they may have already pressed it.
func InterruptButton(c *ui.Context, opts InterruptButtonOptions) InterruptButtonResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	label := opts.Label
	if label == "" {
		label = core.Msg(c, "feedback.interrupt.stop", "Stop")
	}
	busyLabel := opts.BusyLabel
	if busyLabel == "" {
		busyLabel = core.Msg(c, "feedback.interrupt.stopping", "Stopping…")
	}
	name := opts.Name
	if name == "" {
		name = label
		if opts.Busy {
			name = busyLabel
		}
	}
	if opts.Busy {
		label = busyLabel
	}

	fg := k.Danger
	if opts.Caution {
		fg = k.Text
	}

	btn := ui.Button(c, "").Height(core.ControlHeight(c)).Shrink(0).
		Radius(theme.PillRadius).Padding(0, u*4, 0, u*4).
		TextColor(fg).Border(theme.BorderWidth, fg.Alpha(0.35)).
		Label(name).Tooltip(name).Disabled(opts.Busy).
		Children(func() {
			if opts.Busy {
				// A loader where the square was: the press has been taken and
				// something is still in flight, which is the only thing this
				// button's second state means.
				PulseLoader(c, LoaderOptions{Size: u * 3, Color: fg, Label: name}).
					Margin(0, u*1, 0, 0)
				return
			}
			// The square is drawn rather than set in the label: a stop glyph is
			// the one mark on this screen that is not a word, and at a size where
			// a character would be a smudge it has to be a block.
			ui.Box(c).Size(u*2, u*2).Radius(1).Shrink(0).Background(fg).
				Margin(0, u*1.75, 0, 0)
			ui.Text(c, label).TextColor(fg).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
		})

	return InterruptButtonResult{Element: btn, pressed: btn.Clicked()}
}
