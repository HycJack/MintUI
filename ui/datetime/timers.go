package datetime

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── counting ─────────────────────────────────────────────────────────────────

// RelativeTimeOptions configure a RelativeTime.
type RelativeTimeOptions struct {
	// Then is the instant being described and Now is the one it is described
	// from. Both are the caller's: nothing here reads the clock.
	Then, Now time.Time
	// Format writes the difference. It is required, and deliberately so: the
	// measurement and the wording are different things, and a component that
	// chose the wording would put "3 days ago" into a German window.
	Format func(RelativeSpan) string
	// Muted draws the text as secondary, which is what a list of recent
	// activity wants.
	Muted bool
	// Name is what a screen reader announces; empty uses the text.
	Name string
}

// RelativeTime is a moment described by how far it is from another: "3 days
// ago", "in 2 hours".
//
// The wording is the caller's. This component works out the difference — the
// part that is arithmetic and the same in every window — and hands it to a
// function the caller supplied. RelativeText is the library's own wording, and
// a window that has a better one passes it instead.
func RelativeTime(c *ui.Context, opts RelativeTimeOptions) *ui.Element {
	k := core.Tokens(c)
	if opts.Format == nil {
		panic("datetime: RelativeTime needs a Format; the wording of a difference is the caller's")
	}
	if opts.Then.IsZero() {
		return ui.Box(c)
	}
	span := Relative(opts.Then, opts.Now)
	text := opts.Format(span)
	colour := k.Text
	if opts.Muted {
		colour = k.TextMuted
	}
	name := opts.Name
	if name == "" {
		name = text
	}
	return ui.Text(c, text).TextColor(colour).FontSize(core.FontSize(c, theme.CaptionSize)).
		SingleLine().Label(name)
}

// CountdownOptions configure a Countdown.
type CountdownOptions struct {
	// Target is what is being counted to. Required.
	Target time.Time
	// Now is the caller's now, on the same clock as Target.
	Now time.Time
	// Format writes the difference; the same function RelativeTime takes.
	Format func(RelativeSpan) string
	// Start, when set, makes the count run from it rather than to Target: an
	// elapsed timer is the same control the other way up, and measuring it
	// with the same arithmetic is what keeps the two from disagreeing.
	Start time.Time
	// Bar draws a meter of how much of the wait is gone, which needs
	// From and Until to know the whole of it.
	From, Until time.Time
	// Name is what a screen reader announces.
	Name string
	// Compact draws the number without the prose around it.
	Compact bool
}

// Countdown is a target and the time left to it — or the time since it, when
// the control is the other way up.
//
// The meter underneath is the part worth having: a bare number has to be
// compared against a mental picture of how long the wait is, and a bar that
// empties is that comparison drawn. It needs both ends of the wait, so it is
// drawn only when From and Until are both given — a meter that guesses its
// own far end is a bar that is quietly wrong.
func Countdown(c *ui.Context, opts CountdownOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Target.IsZero() {
		panic("datetime: Countdown needs a Target")
	}
	if opts.Format == nil {
		panic("datetime: Countdown needs a Format; the wording of a difference is the caller's")
	}
	// An elapsed count is the same measurement with the ends swapped, so it
	// is measured the same way rather than by a second piece of arithmetic
	// that could disagree with the first.
	then, now := opts.Target, opts.Now
	if !opts.Start.IsZero() {
		then, now = opts.Start, opts.Now
	}
	span := Relative(then, now)
	text := opts.Format(span)
	if opts.Compact {
		text = CountdownText(span)
	}
	name := opts.Name
	if name == "" {
		name = text
	}

	return ui.Column(c).FillWidth().Gap(u).Label(name).Children(func() {
		ui.Text(c, text).TextColor(k.Text).FontSize(core.FontSize(c, theme.BodySize)).Bold().
			SingleLine().FontFeatures("tnum")
		if !opts.From.IsZero() && !opts.Until.IsZero() && opts.Until.After(opts.From) {
			whole := opts.Until.Sub(opts.From)
			done := opts.Now.Sub(opts.From)
			filled := 0
			if whole > 0 {
				filled = int(float64(done) / float64(whole) * 100)
			}
			// Clamped rather than wrapped: a clock that has jumped, or a
			// target that has passed, would otherwise fill a bar past its end
			// and draw outside it.
			filled = min(max(filled, 0), 100)
			ui.Box(c).FillWidth().Height(u).Radius(u / 2).Background(k.Surface).
				Shrink(0).Label(name).Draw(func(p *ui.Painter, rect ui.Rect) {
				p.Fill(ui.Rect{X: rect.X, Y: rect.Y, W: rect.W * float32(filled) / 100, H: rect.H},
					k.Accent, rect.H/2)
			})
		}
	})
}

// CountdownText writes just the number and its unit: "3 days", "in 2 hours"
// without the prose.
//
// It is for a column of counts where every line is the same kind of thing —
// four jobs due, all of them ahead — so the wording is the same on each and
// repeating it four times is noise. On its own, where a reader has to work out
// which way round it is, RelativeText is the one to use.
func CountdownText(s RelativeSpan) string {
	n, unit, ok := s.NonZero()
	if !ok {
		return "0s"
	}
	if n != 1 {
		unit += "s"
	}
	return itoa(n) + " " + unit
}

// StopwatchOptions configure a Stopwatch.
type StopwatchOptions struct {
	// Start is when the clock started, and the reset button zeroes it. It is a
	// pointer because a stopwatch that owns its own start is a stopwatch
	// whose reading cannot be restored from a save file.
	Start *time.Time
	// Now is the caller's now.
	Now time.Time
	// Laps, when given, is where the lap button records the moments it was
	// pressed. Nil draws no lap button, because a stopwatch that records laps
	// nowhere is a stopwatch that forgets them.
	Laps *[]time.Time
	// Precision is how many decimals the seconds are written with; zero is
	// two, which is what a person reading a stopwatch can use.
	Precision int
	// Name is what a screen reader announces.
	Name string
}

// StopwatchResult carries a Stopwatch and what the user did with it.
type StopwatchResult struct {
	// Element is the control.
	Element *ui.Element
	// lapped reports a press of the lap button this frame.
	lapped bool
	// reset reports a press of the reset button this frame.
	reset bool
	// started reports a press of start/stop this frame, and running says
	// whether the clock is now running.
	started bool
	running bool
}

// Lapped reports a press of the lap button.
func (r StopwatchResult) Lapped() bool { return r.lapped }

// Reset reports a press of the reset button.
func (r StopwatchResult) Reset() bool { return r.reset }

// Started reports a press of the start button, and Running whether the clock
// is running now.
func (r StopwatchResult) Started() (pressed, running bool) { return r.started, r.running }

// Stopwatch is a clock that counts up from a moment the caller holds.
//
// The start is the caller's, so the reading survives a rebuild, a save and a
// restart: a stopwatch that keeps its own start inside the widget loses the
// reading every time the view is rebuilt, which on a desktop is every time
// anything else on the window changes.
func Stopwatch(c *ui.Context, opts StopwatchOptions) StopwatchResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Start == nil {
		panic("datetime: Stopwatch needs a Start to point at")
	}
	precision := opts.Precision
	if precision <= 0 || precision > 3 {
		precision = 2
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.stopwatch", "Stopwatch")
	}
	// A zero start is a stopped clock, not one that began when the calendar
	// did: reading the gap between the year 1 and today would put seven
	// figures of minutes on a watch that is not running.
	elapsed := time.Duration(0)
	if !opts.Start.IsZero() {
		elapsed = max(opts.Now.Sub(*opts.Start), 0)
	}

	var r StopwatchResult
	// Running means a start that is not now: a zero start is stopped, and
	// pressing start puts it at the caller's now, which is the only moment
	// this component knows to be true.
	running := !opts.Start.IsZero() && !opts.Start.Equal(opts.Now)
	start := core.Msg(c, "datetime.start", "Start")
	stop := core.Msg(c, "datetime.stop", "Stop")
	label := start
	if running {
		label = stop
	}
	lap := core.Msg(c, "datetime.lap", "Lap")
	reset := core.Msg(c, "datetime.reset", "Reset")

	r.Element = ui.Column(c).FillWidth().Gap(u * 1.5).Label(name).Children(func() {
		ui.Text(c, StopwatchText(elapsed, precision)).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.DisplaySize)).Bold().FontFeatures("tnum").
			SingleLine().Label(name + " " + StopwatchText(elapsed, precision))
		ui.Row(c).FillWidth().Gap(u * 1.5).Children(func() {
			toggle := input.Button(c, label, input.ButtonOptions{Primary: !running, Label: label})
			if toggle.Clicked() {
				if running {
					*opts.Start = time.Time{}
				} else {
					*opts.Start = opts.Now
				}
				r.started, r.running = true, !running
			}
			if opts.Laps != nil {
				if b := input.Button(c, lap, input.ButtonOptions{Label: lap}); b.Clicked() {
					*opts.Laps = append(*opts.Laps, opts.Now)
					r.lapped = true
				}
			}
			if b := input.Button(c, reset, input.ButtonOptions{Label: reset}); b.Clicked() {
				*opts.Start = time.Time{}
				if opts.Laps != nil {
					// Truncated rather than nil: Reset is drawn whether or not
					// the caller wants laps, so the write has to be guarded —
					// and nilling the caller's slice would hand them back a nil
					// and throw away the array their own laps live in.
					*opts.Laps = (*opts.Laps)[:0]
				}
				r.reset = true
			}
		})
		if opts.Laps != nil && len(*opts.Laps) > 0 {
			// The newest lap first, which is the one a person wants when
			// they have just pressed the button.
			ui.Column(c).FillWidth().Gap(u / 2).Children(func() {
				for i := len(*opts.Laps) - 1; i >= 0; i-- {
					prev := time.Duration(0)
					if i > 0 {
						prev = (*opts.Laps)[i-1].Sub(*opts.Start)
					}
					gap := (*opts.Laps)[i].Sub(*opts.Start) - prev
					ui.Text(c, itoa(len(*opts.Laps)-i)+".  "+
						StopwatchText(gap, precision)).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).FontFeatures("tnum")
				}
			})
		}
	})
	return r
}

// StopwatchText writes an elapsed time as 01:23.4 — minutes, seconds and
// hundredths, with a leading zero on the minutes.
//
// Hundredths because that is the resolution a person can read off a stopwatch
// at a glance, and tenths because hundredths of a second is finer than the
// eye can follow a hand moving. A reading with more places than that looks
// precise and is not.
func StopwatchText(d time.Duration, precision int) string {
	if d < 0 {
		d = 0
	}
	minutes := int(d / time.Minute)
	seconds := (d % time.Minute).Seconds()
	switch precision {
	case 0:
		return pad2(minutes) + ":" + pad2(int(seconds))
	case 1:
		return pad2(minutes) + ":" + pad2(int(seconds)) + "." +
			itoa(int(seconds*10)%10)
	default:
		return pad2(minutes) + ":" + pad2(int(seconds)) + "." +
			pad2(int(seconds*100)%100)
	}
}

// ── reminders ────────────────────────────────────────────────────────────────

// ReminderChoices are the offsets a reminder is usually set at, in the order
// they are offered. Zero is "no reminder" and is a real choice: half the
// reminders people set are the ones they then turn off.
var ReminderChoices = []time.Duration{
	0, 5 * time.Minute, 15 * time.Minute, time.Hour, 24 * time.Hour, 7 * 24 * time.Hour,
}

// ReminderText writes how long before something a reminder fires: "15 minutes
// before", or "No reminder" for none.
//
// "Before" and not "before the event" because the row it sits in is about an
// event, and repeating the noun above it is what makes a form read as a form
// rather than as a sentence.
func ReminderText(d time.Duration) string {
	if d == 0 {
		return core.Def("No reminder")
	}
	return core.Def(DurationText(d)) + " before"
}

// ReminderPickerOptions configure a ReminderPicker.
type ReminderPickerOptions struct {
	// Value is the offset before the event, and is written as the choice
	// changes. Zero means no reminder. Required.
	Value *time.Duration
	// Choices are the offsets on offer; nil uses ReminderChoices.
	Choices []time.Duration
	// Custom, when set, is where a length the caller invents goes, and it is
	// shown below the choices when the chosen value is not one of them.
	Custom *time.Duration
	// Name is what a screen reader announces.
	Name string
	// Format writes a length; the default is ReminderText.
	Format func(time.Duration) string
}

// ReminderPickerResult carries a ReminderPicker and whether it changed.
type ReminderPickerResult struct {
	// Element is the control.
	Element *ui.Element
	// changed reports that the reminder moved this frame.
	changed bool
}

// Changed reports that the reminder moved this frame.
func (r ReminderPickerResult) Changed() bool { return r.changed }

// ReminderPicker is a row of offsets: when a reminder about this event should
// fire, in words rather than in a clock.
//
// It is a row of buttons and not a drop-down because the choices are known and
// few. A person setting a reminder for tomorrow's job wants "1 hour before"
// in the same glance as "No reminder", and a menu that hides six answers
// behind one of them is a menu that has to be opened to be read.
func ReminderPicker(c *ui.Context, opts ReminderPickerOptions) ReminderPickerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Value == nil {
		panic("datetime: ReminderPicker needs a Value to point at")
	}
	choices := opts.Choices
	if len(choices) == 0 {
		choices = ReminderChoices
	}
	format := opts.Format
	if format == nil {
		format = ReminderText
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.reminder", "Reminder")
	}

	// A value that is not on offer is written back into Custom, so that a
	// reminder set elsewhere in the application is shown rather than silently
	// rounded to the nearest choice.
	known := false
	for _, d := range choices {
		if d == *opts.Value {
			known = true
		}
	}
	if !known && opts.Custom != nil {
		*opts.Custom = *opts.Value
	}

	var r ReminderPickerResult
	r.Element = ui.Column(c).FillWidth().Gap(u).Label(name).Children(func() {
		ui.Row(c).FillWidth().Gap(u).Children(func() {
			for _, d := range choices {
				d := d
				on := known && d == *opts.Value
				bg, fg := k.Surface, k.Text
				if on {
					bg, fg = k.Fill, k.OnFill
				}
				label := format(d)
				btn := ui.Button(c, "").Grow(1).Height(core.ControlHeight(c)).
					Radius(theme.PillRadius).Background(bg).TextColor(fg).
					Label(name + " " + label).Tooltip(label).Children(func() {
					ui.Text(c, label).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				})
				if btn.Clicked() {
					*opts.Value = d
					r.changed = true
				}
			}
		})
		if !known && opts.Custom != nil {
			// The length the caller had, shown rather than hidden: a form
			// that quietly drops a value on open is a form that loses data.
			ui.Box(c).Radius(theme.PillRadius).Background(k.AccentBg).
				Padding(u, u*2).Children(func() {
				ui.Text(c, format(*opts.Value)).TextColor(k.AccentText).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
	})
	return r
}

// ── an event, in detail ──────────────────────────────────────────────────────

// EventPopoverOptions configure an EventPopover.
type EventPopoverOptions struct {
	// Anchor is the element the panel hangs on — a chip in a day column, a
	// row in an agenda. Required.
	Anchor *ui.Element
	// Open is the caller's state, which the panel also closes: a press
	// outside it and Escape both write false here.
	Open *bool
	// Event is what the panel is about.
	Event Event
	// Format writes a date in the panel; the default is FormatDateTime.
	Format func(time.Time) string
	// Edit names the button that opens the editor; empty draws none.
	Edit string
	// Delete names the destructive button; empty draws none.
	Delete string
	// Edited and Deleted report which of the two was pressed this frame.
	Edited, Deleted bool
}

// EventPopoverResult carries an EventPopover and what the user did with it.
type EventPopoverResult struct {
	// Element is the panel, or nil while it is closed.
	Element *ui.Element
	// edited and deleted are the two presses.
	edited, deleted bool
}

// Edited reports a press of the edit button.
func (r EventPopoverResult) Edited() bool { return r.edited }

// Deleted reports a press of the delete button.
func (r EventPopoverResult) Deleted() bool { return r.deleted }

// EventPopover is an event's details, hung off the chip that stands for it.
//
// It is a popover rather than a side panel because the chip is the thing the
// person pointed at. A details panel that took the window would replace the
// day they were reading to explain one hour of it, and the question — what is
// this, and who is coming — is answered in three lines.
func EventPopover(c *ui.Context, opts EventPopoverOptions) EventPopoverResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Anchor == nil {
		panic("datetime: EventPopover needs the element its panel hangs on")
	}
	if opts.Open == nil {
		panic("datetime: EventPopover needs the *bool it opens and closes in")
	}
	format := opts.Format
	if format == nil {
		format = FormatDateTime
	}
	e := opts.Event
	start, end := e.Span()
	_, fg := e.Severity.Pair(k)

	var r EventPopoverResult
	r.Element = overlay.Popover(c, opts.Anchor, opts.Open, overlay.PopoverOptions{
		Modal: false,
		Title: e.Title,
		Label: e.Title,
		Body: func() {
			ui.Column(c).FillWidth().Gap(u * 1.5).Children(func() {
				text := format(start)
				if !e.AllDay {
					text += " – " + format(end)
				}
				ui.Text(c, text).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).FontFeatures("tnum")
				if e.Location != "" {
					ui.Text(c, e.Location).TextColor(k.Text).
						FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
				}
				// The duration is here as well as the two ends, because the
				// ends of a booking are the two things a person gets wrong.
				ui.Text(c, DurationText(e.Duration())).
					TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize))
				if e.Cancelled {
					ui.Text(c, core.Msg(c, "datetime.cancelled", "Cancelled")).
						TextColor(k.Danger).FontSize(core.FontSize(c, theme.CaptionSize))
				}
				if opts.Edit != "" || opts.Delete != "" {
					ui.Row(c).FillWidth().Gap(u).Children(func() {
						if opts.Edit != "" {
							btn := input.Button(c, opts.Edit, input.ButtonOptions{
								Label:   opts.Edit,
								Primary: true,
							})
							r.edited = btn.Clicked()
						}
						if opts.Delete != "" {
							btn := input.Button(c, opts.Delete, input.ButtonOptions{
								Label:  opts.Delete,
								Danger: true,
							})
							r.deleted = btn.Clicked()
						}
					})
				}
			})
		},
	})
	return r
}

// EventEditorOptions configure an EventEditor.
type EventEditorOptions struct {
	// Event is the event being edited, and is written as the fields are
	// changed. Required.
	Event *Event
	// Save and Cancel name the two buttons; Save is required, because an
	// editor with nothing to confirm leaves the caller with no way to know
	// the change was meant.
	Save, Cancel string
	// Attendees are the people, shown under the fields.
	Attendees []Attendee
	// Reminder writes into the caller's offset as the reminder is chosen.
	Reminder *time.Duration
	// Timezone is the zone the times are shown in; empty keeps the event's
	// own.
	Timezone string
	// Open is the caller's state for the date calendar the editor holds open.
	// Required: an editor that opened a calendar on every frame would cover
	// the form it is part of.
	Open *bool
	// Duration is the length the start and end are kept to; zero leaves the
	// end alone.
	Duration time.Duration
	// Formatted names a date for a screen reader; empty uses the text.
	Name string
}

// EventEditorResult carries an EventEditor and what the user did with it.
type EventEditorResult struct {
	// Element is the form.
	Element *ui.Element
	// saved and cancelled are the two presses.
	saved, cancelled bool
}

// Saved reports a press of the save button.
func (r EventEditorResult) Saved() bool { return r.saved }

// Cancelled reports a press of the cancel button.
func (r EventEditorResult) Cancelled() bool { return r.cancelled }

// EventEditor is the form behind an event: its title, when it is, how long it
// lasts, who is coming, and when they should be reminded.
//
// It is assembled rather than drawn — a date picker, a clock, a reminder row,
// the attendee list — so that every one of those is the same control the rest
// of the application uses. An editor with its own private calendar would be a
// second calendar to keep right, and the two would disagree about February.
func EventEditor(c *ui.Context, opts EventEditorOptions) EventEditorResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Event == nil {
		panic("datetime: EventEditor needs an Event to point at")
	}
	if opts.Save == "" {
		panic("datetime: EventEditor needs a Save label")
	}
	if opts.Open == nil {
		panic("datetime: EventEditor needs the *bool its calendar opens in")
	}
	e := opts.Event
	cancel := opts.Cancel
	if cancel == "" {
		cancel = core.Msg(c, "datetime.cancel", "Cancel")
	}

	var r EventEditorResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).Label(opts.Name).Children(func() {
		fieldRow(c, core.Msg(c, "datetime.title", "Title"), func() {
			ui.Text(c, e.Title).TextColor(k.Text).FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
		})
		DateTimePicker(c, DateTimePickerOptions{
			Value:      &e.Start,
			Open:       opts.Open,
			Today:      e.Start,
			MinuteStep: 5,
		})
		if e.Start.Location() != nil && opts.Timezone != "" {
			ui.Text(c, ZoneLabel(opts.Timezone, e.Start)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		// The end follows the start, and only when it is not already right:
		// a field that rewrote its own value on every frame would be a field
		// a caller could never keep its own opinion about.
		if opts.Duration > 0 {
			end := e.Start.Add(opts.Duration)
			if !e.End.Equal(end) {
				e.End = end
			}
			ui.Text(c, DurationText(e.Duration())).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).FontFeatures("tnum")
		}
		if opts.Reminder != nil {
			ReminderPicker(c, ReminderPickerOptions{Value: opts.Reminder})
		}
		if len(opts.Attendees) > 0 {
			AttendeeList(c, AttendeeOptions{Attendees: opts.Attendees, Avatars: true})
		}
		ui.Row(c).FillWidth().Gap(u).Justify(ui.End).Children(func() {
			if b := input.Button(c, cancel, input.ButtonOptions{Label: cancel}); b.Clicked() {
				r.cancelled = true
			}
			if b := input.Button(c, opts.Save, input.ButtonOptions{
				Label: opts.Save, Primary: true,
			}); b.Clicked() {
				r.saved = true
			}
		})
	})
	return r
}

// fieldRow is a label and the value under it, which is how every field in this
// library is shaped: the label above, not beside, so that a form's labels are
// left-aligned down a column whatever the widths of the values.
func fieldRow(c *ui.Context, name string, value func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Column(c).FillWidth().Gap(u / 2).Label(name).Children(func() {
		ui.Text(c, name).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		ui.Box(c).FillWidth().Radius(theme.ControlRadius).Background(k.Surface).
			Padding(u, u*2).Children(value)
	})
}
