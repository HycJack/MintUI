package account

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// WizardStep is one step of an onboarding wizard.
type WizardStep struct {
	// Title is what the step is called.
	Title string
	// Body is what it says, as the caller's sentences: an onboarding wizard
	// is the app explaining itself, and only the app knows what it is.
	Body string
	// Step builds the step's controls. Nil draws the body alone, which is
	// right for a step that is only ever a thing to read.
	Step func()
}

// OnboardingWizardOptions configure an OnboardingWizard.
type OnboardingWizardOptions struct {
	// Steps are the wizard's steps in order. At least one: a wizard with none
	// is a close button.
	Steps []WizardStep
	// Step is the caller's index into Steps, which is what makes the wizard
	// deep-linkable — "finish onboarding, step three" is a URL somebody can
	// be sent — and what lets a caller restore a half-finished wizard after
	// a restart.
	Step *int
	// Finished writes true when the last step is completed.
	Finished *bool
	// Skipped writes true when the wizard is skipped, so that somebody who
	// has seen all this before is not made to sit through it again. A
	// separate flag rather than a step of its own because skipping is not
	// completing: the app still has to remember that this person was shown
	// the wizard once.
	Skipped *bool
	// ShowSkip draws the skip link. It is off when there is no Skipped to
	// write to, and the wizard panics if one is asked for and not offered.
	ShowSkip bool
}

// OnboardingWizardResult carries an OnboardingWizard and where it is.
type OnboardingWizardResult struct {
	// Element is the whole wizard.
	Element *ui.Element
	// last is whether the step being drawn is the final one.
	last bool
	// finished and skipped are this frame's answers.
	finished, skipped bool
}

// Finished reports that the wizard was completed this frame.
func (r OnboardingWizardResult) Finished() bool { return r.finished }

// Skipped reports that the wizard was skipped this frame.
func (r OnboardingWizardResult) Skipped() bool { return r.skipped }

// AtEnd reports that the last step is the one showing, which is what a caller
// uses to decide whether its own "finish setup" button belongs on this screen
// as well.
func (r OnboardingWizardResult) AtEnd() bool { return r.last }

// OnboardingWizard is the first run: a few screens, a progress count, and a
// way out at every one of them.
//
// The step is the caller's index rather than something this component counts
// up through, because the wizard is not always the whole window — it is
// usually a sheet over a board somebody can already use — and a wizard that
// owns its own index cannot be closed and reopened on the same step.
//
// Skip is offered on every step including the last, because somebody who has
// read this before would rather skip than click through it a fourth time,
// and making them do that to reach a button they can already see is the kind
// of small hostility that makes people uninstall things.
func OnboardingWizard(c *ui.Context, opts OnboardingWizardOptions) OnboardingWizardResult {
	if len(opts.Steps) == 0 {
		panic("account: OnboardingWizard needs at least one step")
	}
	if opts.Step == nil {
		panic("account: OnboardingWizard needs the *int Step writes to")
	}
	if opts.Finished == nil {
		panic("account: OnboardingWizard needs the *bool Finished writes to")
	}
	if opts.ShowSkip && opts.Skipped == nil {
		panic("account: a wizard that offers Skip needs the *bool Skipped " +
			"writes to; a skip that is not remembered is shown again next run")
	}
	// An index outside the list would draw a step that is not there. It is
	// clamped to the last step rather than to the first, because the usual
	// way to get here is a saved index from a wizard that had more steps.
	//
	// The clamp is written back, not just applied to a local: a wizard that
	// drew its last step while the caller's own index stayed at 99 would be
	// showing one step to the person and claiming another to the rest of the
	// app, and the next frame would have to make the same choice again.
	idx := *opts.Step
	if idx < 0 {
		idx = 0
	}
	if idx >= len(opts.Steps) {
		idx = len(opts.Steps) - 1
	}
	*opts.Step = idx

	cur := opts.Steps[idx]
	last := idx == len(opts.Steps)-1

	var r OnboardingWizardResult
	r.Element = ui.Column(c).FillWidth().Gap(stepGap(c)).
		Label(core.Msg(c, "account.onboarding", "Welcome")).Children(func() {
		wizardCount(c, idx, len(opts.Steps))
		ui.Text(c, cur.Title).TextColor(core.Tokens(c).Text).Bold().
			FontSize(core.FontSize(c, theme.TitleSize))
		if cur.Body != "" {
			ui.Text(c, cur.Body).TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		}
		if cur.Step != nil {
			cur.Step()
		}
		wizardActions(c, opts, idx, last)
	})
	return r
}

// stepGap is the wizard's spacing between its parts, which is wider than the
// density's own step: the parts are whole blocks rather than a run of rows,
// and a column of them at the row gap reads as a form.
func stepGap(c *ui.Context) float32 { return core.Density(c).Unit() * 6 }

// wizardCount is "Step 2 of 4", with the count in the muted tone and the
// current step's number in the body tone — the one place in this package
// where two tones carry one sentence, and the split is worth it because
// "2" is what a person scans for.
func wizardCount(c *ui.Context, at, of int) {
	k := core.Tokens(c)
	label := core.Msg(c, "account.stepOf", "Step") + " " + itoa(at+1) + " / " + itoa(of)
	ui.Row(c).AlignItems(ui.Center).Gap(0).Label(label).Children(func() {
		ui.Text(c, itoa(at+1)).TextColor(k.Text).Bold().
			FontSize(core.FontSize(c, theme.CaptionSize))
		ui.Text(c, " / "+itoa(of)).TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// wizardActions is the row under a step: skip on the left, back and the
// primary action on the right.
func wizardActions(c *ui.Context, opts OnboardingWizardOptions, idx int, last bool) {
	u := core.Density(c).Unit()
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
		if opts.ShowSkip {
			if textButton(c, core.Msg(c, "account.skip", "Skip setup")).Clicked() {
				*opts.Skipped = true
			}
		}
		ui.Box(c).Grow(1)
		if idx > 0 {
			back := input.Button(c, core.Msg(c, "account.back", "Back"),
				input.ButtonOptions{})
			if back.Clicked() {
				*opts.Step = idx - 1
			}
		}
		label := core.Msg(c, "account.next", "Next")
		if last {
			label = core.Msg(c, "account.finish", "Finish")
		}
		forward := input.Button(c, label, input.ButtonOptions{Primary: true})
		if forward.Clicked() {
			if last {
				*opts.Finished = true
				return
			}
			*opts.Step = idx + 1
		}
	})
}

// CoachmarkOptions configure a Coachmark.
type CoachmarkOptions struct {
	// Anchor is the element the mark points at. It is required: a coachmark
	// with nothing to point at is a dialog, and the caller can use one.
	Anchor *ui.Element
	// Open is the caller's bool. Every way out writes false into it — the
	// button, the scrim, Escape — because a mark that cannot be dismissed is
	// a modal with a speech bubble on it.
	Open *bool
	// Title and Body are the caller's sentences.
	Title, Body string
	// Step is "2 of 4", shown under the body. Empty draws no count, which is
	// right for a one-off mark with no sequence behind it.
	Step string
	// ShowSkip draws a skip link for the last step of a sequence. On any step
	// but the last it is not drawn even when asked for: "skip" on the first
	// of four is not a thing anybody means.
	ShowSkip bool
	// Placement is which side of the anchor the panel hangs on: ui.Start for
	// one that hangs above or to the left, ui.End for one that hangs below
	// or to the right. Zero is Start, because the element a coachmark points
	// at is nearly always in the top-left of what is being explained and a
	// bubble over it is the one that fits.
	Placement ui.Align
}

// CoachmarkResult carries a Coachmark and what was done in it.
type CoachmarkResult struct {
	// Element is the trigger, with the panel built under it while it shows.
	Element *ui.Element
	// next and done report this frame's presses.
	next, done bool
}

// Next reports a press of the "next" button, for a caller walking a sequence.
func (r CoachmarkResult) Next() bool { return r.next }

// Done reports that the last mark was dismissed.
func (r CoachmarkResult) Done() bool { return r.done }

// Coachmark is the bubble that explains the thing under it, once.
//
// It is layout.Panel on a MyGo popover, which is the same arrangement a menu
// uses, and that is deliberate: the bubble has to look like the rest of the
// floating layers in the app rather than like a tooltip that grew, and the
// only way to guarantee that is to be built from the same two pieces.
//
// The anchor is an element the caller already has, never one built here. See
// docs/design-system.md §17.1: an element belongs to whatever container was
// current when it was made, so a coachmark that built its own anchor would
// land outside the thing it is meant to be attached to and point at nothing.
func Coachmark(c *ui.Context, opts CoachmarkOptions) CoachmarkResult {
	if opts.Anchor == nil {
		panic("account: Coachmark needs the Anchor it points at; a mark with " +
			"nothing to point at is a dialog")
	}
	if opts.Open == nil {
		panic("account: Coachmark needs the *bool it opens and closes")
	}
	if opts.Title == "" && opts.Body == "" {
		panic("account: Coachmark needs a Title or a Body; an empty bubble is " +
			"a rectangle somebody has to click past")
	}
	u := core.Density(c).Unit()
	name := opts.Title
	if name == "" {
		name = firstLine(opts.Body)
	}

	var r CoachmarkResult
	ui.PopoverBase(c, opts.Anchor, opts.Open, func(p *ui.Element) {
		panelFace(c, p, panelOptions{
			label: name,
			body: func() {
				if opts.Title != "" {
					ui.Text(c, opts.Title).TextColor(core.Tokens(c).Text).Bold().
						FontSize(core.FontSize(c, theme.RowSize))
				}
				if opts.Body != "" {
					ui.Text(c, opts.Body).TextColor(core.Tokens(c).TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize))
				}
				if opts.Step != "" {
					ui.Text(c, opts.Step).TextColor(core.Tokens(c).TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
				}
				ui.Row(c).FillWidth().Gap(u * 2).AlignItems(ui.Center).Children(func() {
					if opts.ShowSkip {
						if textButton(c, core.Msg(c, "account.skip", "Skip")).Clicked() {
							r.done = true
							*opts.Open = false
						}
					}
					ui.Box(c).Grow(1)
					next := input.Button(c, core.Msg(c, "account.gotIt", "Got it"),
						input.ButtonOptions{Primary: true})
					if next.Clicked() {
						r.next = true
						r.done = true
						*opts.Open = false
					}
				})
			},
		})
	})
	// The anchor is returned rather than built: a caller that needs the mark
	// in a particular place already knows where it is.
	r.Element = opts.Anchor
	return r
}

// firstLine is the first line of a body, used as a mark's name when it has no
// title of its own. It is the name a screen reader reads first, and a bubble
// whose first sentence is its subject is almost always the better name than
// "Coachmark".
func firstLine(body string) string {
	for i, r := range body {
		if r == '\n' {
			return body[:i]
		}
	}
	return body
}

// FeedbackOptions configure a FeedbackWidget.
type FeedbackWidgetOptions struct {
	// Note is the caller's string, written as it is typed.
	Note *string
	// Sent writes true when the message goes out. A widget with nowhere to
	// send to is a notes field, so this is required.
	Sent *bool
	// Categories are the chips above the field; the chosen one is the
	// caller's. Empty draws no chips, which is right for an app whose
	// feedback is only ever about one thing.
	Categories []string
	// Category is the chosen chip, empty for none.
	Category *string
	// MaxRunes caps the note. Zero is no cap. The count is shown under the
	// field rather than enforced silently: a field that just stops taking
	// keystrokes looks broken.
	MaxRunes int
	// Dismissed writes true when the widget is closed, so a caller showing
	// it in a corner can put it away.
	Dismissed *bool
	// Closeable draws the close button.
	Closeable bool
}

// FeedbackWidgetResult carries a FeedbackWidget and what was done in it.
type FeedbackWidgetResult struct {
	// Element is the widget.
	Element *ui.Element
	// sent and dismissed are this frame's answers.
	sent, dismissed bool
}

// Sent reports that the message was sent this frame.
func (r FeedbackWidgetResult) Sent() bool { return r.sent }

// Dismissed reports that the widget was closed this frame.
func (r FeedbackWidgetResult) Dismissed() bool { return r.dismissed }

// FeedbackWidget is a box in the corner asking what is wrong.
//
// It is a Form rather than a panel with a text box in it, so that the label,
// the field and the button are the library's and the tab order runs top to
// bottom. The chips are input.ChoiceChips: they are a choice from a fixed
// set, which is exactly what a chip row is, and drawing them by hand here
// would mean re-deciding what happens at the cap.
//
// The send button is disabled on an empty note, which is the one piece of
// state this component computes: "is there anything to send" is a fact about
// the field, and recomputing it in the caller would be one more copy of the
// same rule.
func FeedbackWidget(c *ui.Context, opts FeedbackWidgetOptions) FeedbackWidgetResult {
	if opts.Note == nil {
		panic("account: FeedbackWidget needs the *string Note writes to")
	}
	if opts.Sent == nil {
		panic("account: FeedbackWidget needs the *bool Sent writes to; a box " +
			"with nowhere to send to is a notes field")
	}
	if len(opts.Categories) > 0 && opts.Category == nil {
		panic("account: a feedback widget with categories needs the *string " +
			"Category writes to")
	}
	u := core.Density(c).Unit()

	var r FeedbackWidgetResult
	r.Element = ui.Column(c).Width(feedbackWidth).Shrink(0).Gap(u*2).
		Padding(u*2.5, u*3).Radius(theme.ControlRadius).Background(core.Tokens(c).Surface).
		Label(core.Msg(c, "account.feedback", "Feedback")).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, core.Msg(c, "account.feedback", "Feedback")).TextColor(core.Tokens(c).Text).
				Bold().FontSize(core.FontSize(c, theme.RowSize)).Grow(1)
			if opts.Closeable {
				close := input.IconButton(c, glyphDismiss, core.Msg(c, "account.close", "Close"),
					input.ButtonOptions{})
				if close.Clicked() && opts.Dismissed != nil {
					r.dismissed = true
					*opts.Dismissed = true
				}
			}
		})

		if len(opts.Categories) > 0 {
			choices := make([]input.Choice, 0, len(opts.Categories))
			for _, name := range opts.Categories {
				choices = append(choices, input.Choice{Value: name, Label: name})
			}
			input.ChoiceChips(c, categorySlice(opts.Category), choices,
				input.ChoiceChipsOptions{
					Label: core.Msg(c, "account.feedbackKind", "What is it about"),
					Max:   1,
					Wrap:  true,
				})
		}

		input.TextArea(c, opts.Note, input.TextAreaOptions{
			Label:       core.Msg(c, "account.feedbackNote", "What is going wrong?"),
			Placeholder: core.Msg(c, "account.feedbackHint", "Tell us what you were doing"),
			Lines:       4,
		})
		if opts.MaxRunes > 0 {
			ui.Text(c, itoa(runeCount(*opts.Note))+" / "+itoa(opts.MaxRunes)).
				TextColor(core.Tokens(c).TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}

		ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
			empty := trimmed(*opts.Note) == ""
			send := input.Button(c, core.Msg(c, "account.send", "Send"),
				input.ButtonOptions{Primary: true, Disabled: empty})
			if send.Clicked() && !empty {
				r.sent = true
				*opts.Sent = true
			}
		})
	})
	return r
}

// feedbackWidth is how wide the corner box is: wide enough for four lines of
// a note, narrow enough to sit over a board without hiding it.
const feedbackWidth float32 = 320

// categorySlice is the category choice as the slice ChoiceChips wants: a
// pointer to a slice with either no element or the one chip.
//
// It allocates each frame and that is the price of adapting a *string to a
// control whose value is a set. The alternative is a second category control
// of our own, which is a second thing to keep right about wrapping, about the
// cap and about what a chosen chip looks like.
func categorySlice(chosen *string) *[]string {
	if chosen == nil || *chosen == "" {
		empty := []string{}
		return &empty
	}
	one := []string{*chosen}
	return &one
}

// trimmed is the note with its surrounding space off, which is what decides
// whether there is anything to send: a field holding three spaces has
// something typed into it and nothing to say.
func trimmed(s string) string {
	start, end := 0, len(s)
	for start < end && isSpace(s[start]) {
		start++
	}
	for end > start && isSpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

// isSpace reports whether a byte is ASCII space or tab. The note's own
// whitespace is the only whitespace trimmed here; nothing in this package
// rewrites a value somebody typed beyond saying whether it was empty.
func isSpace(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }

// runeCount is a note's length in characters rather than bytes, so that a
// limit of 500 does not silently become 166 for a note of emoji.
func runeCount(s string) int { return len([]rune(s)) }
