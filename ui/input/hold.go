package input

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// holdKey is where a HoldToConfirm remembers when the press began. It is a
// type of its own so that nothing else on the same element is found by its
// name, which is the rule every store in this package follows.
type holdKey struct{}

// holdDefaultFor is how long a confirm is held by default.
//
// It is long enough that nobody finishes it by accident and short enough that
// nobody reads it as a control that has stopped working: a hold is a
// deliberate act, and the wait is the deliberate part of it. A second is the
// number every desktop has settled on.
const holdDefaultFor = 900 * time.Millisecond

// HoldToConfirmOptions configure a HoldToConfirm.
type HoldToConfirmOptions struct {
	// Label is the words on the button, and is required: it is what the
	// control reads out and what a test clicks, and a button whose only mark
	// is a progress ring says nothing at all until it is finished.
	Label string
	// HoldFor is how long the press has to last. Zero is a second, which is
	// what a destructive action wants.
	HoldFor time.Duration
	// Done is the word for having finished, and replaces the label while the
	// control is in that state so that a screen reader hears it rather than
	// hearing the same word twice with nothing to say for the second one.
	// Empty takes the library's "Confirmed".
	Done string
	// Danger draws it in the danger colour, which is what every hold is: the
	// whole reason a hold exists is that the action is one a press should
	// not be enough for.
	Danger bool
	// Primary fills it with ink, for the one destructive action a window
	// wants a person to find rather than avoid.
	Primary bool
	// Disabled takes the control out of play, and is what a caller sets
	// while the thing being confirmed is already gone.
	Disabled bool
	// Width bounds the button; zero fits the label.
	Width float32
}

// HoldToConfirmResult carries a HoldToConfirm and how far along the hold is.
type HoldToConfirmResult struct {
	// Element is the button.
	Element *ui.Element
	// confirmed is that the hold finished this frame.
	confirmed bool
	// progress is how much of the hold has elapsed, 0 to 1, as of this
	// frame.
	progress float32
}

// Confirmed reports that the hold finished this frame. The control also
// writes true into the *bool it was given, so a caller that only wants to
// know "did it happen at all, whenever" can read that and ignore this; this
// is for the caller that has something to do on the frame it happened.
func (r HoldToConfirmResult) Confirmed() bool { return r.confirmed }

// Progress is how far the hold has got, 0 to 1. It is reported rather than
// left to the caller to measure off a ring, because the ring is a picture and
// this is the number — and a test that wants to know the hold is filling
// should not have to read pixels to say so.
func (r HoldToConfirmResult) Progress() float32 { return r.progress }

// HoldToConfirm is a button that only does its thing after it has been held.
//
// It is the control for the actions a single press should not be enough for:
// deleting a schedule, ending a session, sending to every device at once.
// Those are exactly the actions where a press is a mistake rather than a
// decision, and a confirmation dialog is too heavy a price for a typo.
//
// The hold is written into the caller's *bool and not kept here, so a caller
// that fires something on it does so from its own state. Everything else —
// how far the hold has got, and whether it finished this frame — is on the
// result, because neither of those is a fact about the thing being confirmed.
func HoldToConfirm(c *ui.Context, fired *bool, opts HoldToConfirmOptions) HoldToConfirmResult {
	if fired == nil {
		panic("input: HoldToConfirm needs the *bool the hold writes when it finishes")
	}
	if opts.Label == "" {
		panic("input: HoldToConfirm needs a Label; a button whose only mark is a progress ring " +
			"says nothing about what it will do")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	hold := opts.HoldFor
	if hold <= 0 {
		hold = holdDefaultFor
	}
	done := opts.Done
	if done == "" {
		done = core.Msg(c, "input.confirmed", core.Def("Confirmed"))
	}
	// A second hold cannot start while the last one is still going: the flag
	// is the caller's, and a control that cleared it would be the caller
	// holding its own state for this one.
	busy := *fired

	var r HoldToConfirmResult
	btn := ui.ButtonBase(c).Height(core.ControlHeight(c)).Radius(theme.PillRadius).
		Padding(0, u*4, 0, u*4).Role(ui.RoleButton).
		Disabled(opts.Disabled || busy)
	if opts.Width > 0 {
		btn.Width(opts.Width).Shrink(0)
	}
	btn.Children(func() {
		// The clock lives on the button, which MyGo keeps across frames, so
		// the start of a hold outlives the frame that started it. Without it
		// the hold would restart every frame and never finish.
		since := ui.Local(btn, holdKey{}, func() time.Time { return time.Time{} })
		now := c.Now()

		switch {
		case btn.Pressed():
			if since.IsZero() {
				*since = now
			}
		case !since.IsZero():
			// The pointer came off the button before the hold finished, so
			// the hold is off. Nothing was written: a hold abandoned half
			// way is the entire reason this control is not a button.
			*since = time.Time{}
		}

		progress := float32(0)
		if !since.IsZero() {
			elapsed := now.Sub(*since)
			if elapsed >= hold {
				progress = 1
				*since = time.Time{}
				*fired = true
				r.confirmed = true
				c.Announce(done)
			} else {
				progress = float32(elapsed) / float32(hold)
				// The frame that finishes the hold is asked for by the
				// clock, and it is asked for again every frame it is still
				// waiting, so the ring cannot stop at 99% and stay there
				// because nothing else in the window changed.
				c.After(hold - elapsed)
			}
		}
		r.progress = progress

		ink := k.Text
		switch {
		case opts.Primary:
			ink = inkOn(k.Fill)
			btn.Background(k.Fill)
		case opts.Danger:
			ink = k.Danger
			btn.Background(k.Surface)
		default:
			btn.Background(k.Surface)
		}
		if opts.Disabled {
			ink = k.TextFaint
		}
		// The hover face is a step above the resting one on the same surface
		// the rest of the library hovers on, so a hold that has been found is
		// visibly found rather than merely there.
		if btn.Hovered() && !opts.Disabled && !busy && progress == 0 {
			if opts.Primary {
				btn.Background(k.Fill.Mix(k.Text, 0.14))
			} else {
				btn.Background(k.SurfaceHover)
			}
		}
		if progress > 0 && !opts.Disabled {
			// The fill is drawn under the words rather than around them, so
			// the label never moves: a label that shifts as the hold fills
			// is a label the person has to keep re-reading.
			btn.DrawOver(func(p *ui.Painter, r ui.Rect) {
				p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * progress, H: r.H}, k.Accent.Alpha(0.25), 0)
			})
			ink = k.Accent
		}
		say := opts.Label
		if busy {
			say = done
			btn.Label(done).Tooltip(done)
		} else {
			btn.Label(opts.Label).Tooltip(opts.Label)
		}
		ui.Text(c, say).SingleLine().TextColor(ink).
			FontSize(core.FontSize(c, theme.BodySize))
	})
	r.Element = btn
	return r
}
