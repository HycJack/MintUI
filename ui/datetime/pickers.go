package datetime

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// field is the control-height button every picker collapses to: a label, a
// chevron, nothing else. It is the shape MyGo's own date and time inputs have,
// so a form mixing a picker with a text field does not look like two
// different programs.
func field(c *ui.Context, value, placeholder, name string) *ui.Element {
	k := core.Tokens(c)
	show, fg := value, k.Text
	if show == "" {
		show, fg = placeholder, k.TextFaint
	}
	return ui.Button(c, "").Height(core.ControlHeight(c)).
		Radius(theme.ControlRadius).Padding(0, 12).
		Background(k.Surface).TextColor(fg).Label(name).Tooltip(name).
		Children(func() {
			ui.Text(c, show).TextColor(fg).SingleLine().Grow(1).TextAlign(ui.Start)
			ui.Icon(c, iconDown).TextColor(k.TextMuted).Size(10, 10)
		})
}

// DatePickerOptions configure a DatePicker.
type DatePickerOptions struct {
	// Value is the chosen day. Only its date is read and written: the clock
	// is left alone, because a picker that set it to midnight would silently
	// move a timed booking to the small hours.
	Value *time.Time
	// Open is the caller's popover state. Required, and the caller closes it:
	// whether a pick ends the popover is a decision about the form, not about
	// the calendar inside it.
	Open *bool
	// Today is the caller's today; the zero time draws no today.
	Today time.Time
	// Min and Max bound the days that can be chosen.
	Min, Max time.Time
	// Placeholder is shown when Value is zero.
	Placeholder string
	// Name is what a screen reader announces; empty uses the field's text.
	Name string
	// Format writes the value in the field. The default is FormatDate.
	Format func(time.Time) string
	// Weekdays and Months name the days and months; zero means the default.
	Weekdays WeekdayNames
	Months   MonthNames
	// Weeks is how many rows the calendar draws; zero means DefaultWeeks.
	Weeks int
	// Busy marks a day as having something on it.
	Busy func(time.Time) bool
	// Disabled greys a day out and takes its clicks away.
	Disabled func(time.Time) bool
	// Footer draws under the calendar, where a picker puts the controls that
	// are about the pick rather than about the day — clear, jump to today.
	Footer func()
}

// DatePickerResult carries a DatePicker and the day the user chose.
type DatePickerResult struct {
	// Element is the field, and the popover's panel when it is open.
	Element *ui.Element
	// picked is the day chosen this frame, and got whether there was one.
	picked time.Time
	got    bool
}

// Picked returns the day the user chose, and whether they chose one.
func (r DatePickerResult) Picked() (time.Time, bool) { return r.picked, r.got }

// DatePicker is a field showing a date, and a calendar that chooses it.
//
// The chosen day is written straight into the caller's Value, the way
// Segmented writes its index: a form field is a view of the value it edits,
// and a second copy of that value inside the widget is a second thing to keep
// in step. Picked reports the press so the caller can close the popover, check
// whether the form is now valid, or save.
//
// The calendar inside is the ordinary one at the ordinary size. A picker with a
// smaller, prettier calendar of its own would be a second date grid to keep
// right.
func DatePicker(c *ui.Context, opts DatePickerOptions) DatePickerResult {
	if opts.Value == nil {
		panic("datetime: DatePicker needs a Value to point at")
	}
	if opts.Open == nil {
		panic("datetime: DatePicker needs the *bool its calendar opens in")
	}
	u := core.Density(c).Unit()
	format := opts.Format
	if format == nil {
		format = FormatDate
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.date", "Date")
	}
	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = name
	}
	// The calendar pages on a copy of the month, so browsing is not editing:
	// a caller who has paged to March without choosing has not changed
	// anything, and a form that saved the browsed month would be wrong.
	month, ok := StartOfMonth(*opts.Value), !opts.Value.IsZero()
	if !ok && !opts.Today.IsZero() {
		month = StartOfMonth(opts.Today)
	}
	if !ok && opts.Today.IsZero() {
		panic("datetime: DatePicker needs a Value or a Today to open its calendar on")
	}

	var r DatePickerResult
	value := ""
	if !opts.Value.IsZero() {
		value = format(*opts.Value)
	}
	anchor := field(c, value, placeholder, name)
	// The field opens its own calendar: the caller's bool is the popover's
	// state and the field is the only thing that can open it, so writing it
	// here is the control telling the form what the user just did rather than
	// holding a second copy of the state.
	if anchor.Clicked() {
		*opts.Open = !*opts.Open
	}
	r.Element = anchor
	overlay.Popover(c, anchor, opts.Open, overlay.PopoverOptions{
		Modal: false,
		Width: MinCellSize*7 + u*6,
		Body: func() {
			cal := Calendar(c, CalendarOptions{
				Month:    &month,
				Selected: opts.Value,
				Today:    opts.Today,
				Min:      opts.Min,
				Max:      opts.Max,
				Weeks:    opts.Weeks,
				Weekdays: opts.Weekdays,
				Months:   opts.Months,
				Busy:     opts.Busy,
				Disabled: opts.Disabled,
				Header:   true,
			})
			// The day chosen is written straight into the caller's value,
			// keeping the clock: a picker that set the time to midnight would
			// silently move a timed booking to the small hours. The days
			// outside Min and Max are disabled in the grid, so there is
			// nothing to check here that the grid has not already refused.
			if day, picked := cal.Picked(); picked {
				*opts.Value = time.Date(day.Year(), day.Month(), day.Day(),
					opts.Value.Hour(), opts.Value.Minute(), opts.Value.Second(),
					0, opts.Value.Location())
				r.picked, r.got = day, true
			}
			if opts.Footer != nil {
				opts.Footer()
			}
		},
	})
	return r
}

// DateRangePickerOptions configure a DateRangePicker.
type DateRangePickerOptions struct {
	// Start and End are the chosen days, and the caller's own: a range is
	// chosen in two presses and both halves are written straight into them.
	Start, End *time.Time
	// Open is the caller's popover state.
	Open *bool
	// Today is the caller's today; the zero time draws no today.
	Today time.Time
	// Min and Max bound the days that can be chosen.
	Min, Max time.Time
	// Panels is how many months side by side: one or two. Three months of a
	// date range is a calendar, not a picker.
	Panels int
	// Name is what a screen reader announces.
	Name string
	// Format writes a date in the field. The default is FormatDate.
	Format func(time.Time) string
	// Span writes the length of the range under the calendars. The default is
	// DurationText.
	Span func(time.Duration) string
	// Weekdays and Months name the days and the months; zero means default.
	Weekdays WeekdayNames
	Months   MonthNames
	// Busy marks a day as having something on it.
	Busy func(time.Time) bool
	// Disabled greys a day out and takes its clicks away.
	Disabled func(time.Time) bool
}

// DateRangePickerResult carries a DateRangePicker and whether it changed.
type DateRangePickerResult struct {
	// Element is the field, and the popover's panel when it is open.
	Element *ui.Element
	// changed reports that the range is not what it was.
	changed bool
}

// Changed reports that the range moved this frame.
func (r DateRangePickerResult) Changed() bool { return r.changed }

// DateRangePicker is a field showing two dates, and one or two calendars that
// choose them.
//
// A range takes two presses — the first night, then the last. A press on a
// range that is already complete starts a new one, which is the only way to
// correct a mis-click without an undo, and a press before the start swaps the
// two ends, because a range drawn backwards is one a person meant forwards.
func DateRangePicker(c *ui.Context, opts DateRangePickerOptions) DateRangePickerResult {
	if opts.Start == nil || opts.End == nil {
		panic("datetime: DateRangePicker needs a Start and an End to point at")
	}
	if opts.Open == nil {
		panic("datetime: DateRangePicker needs the *bool its calendars open in")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	format := opts.Format
	if format == nil {
		format = FormatDate
	}
	span := opts.Span
	if span == nil {
		span = DurationText
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.range", "Dates")
	}
	panels := opts.Panels
	if panels < 1 || panels > 2 {
		panels = 1
	}
	// The panels start on the month the range starts in, the first one on the
	// month after: a range is read forwards.
	from := *opts.Start
	if from.IsZero() {
		from = opts.Today
	}
	first := StartOfMonth(from)
	months := make([]time.Time, panels)
	for i := range months {
		months[i] = AddMonths(first, i)
	}

	text := name
	switch {
	case !opts.Start.IsZero() && !opts.End.IsZero():
		text = format(*opts.Start) + "  –  " + format(*opts.End)
	case !opts.Start.IsZero():
		text = format(*opts.Start) + "  –  " + core.Msg(c, "datetime.end", "end")
	}

	var r DateRangePickerResult
	anchor := field(c, text, name, name)
	if anchor.Clicked() {
		*opts.Open = !*opts.Open
	}
	r.Element = anchor
	overlay.Popover(c, anchor, opts.Open, overlay.PopoverOptions{
		Modal: false,
		Width: float32(panels) * (MinCellSize*7 + u*6),
		Body: func() {
			ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
				for i := range months {
					cal := Calendar(c, CalendarOptions{
						Month:    &months[i],
						Selected: opts.Start,
						Today:    opts.Today,
						Start:    *opts.Start,
						End:      *opts.End,
						Min:      opts.Min,
						Max:      opts.Max,
						Weekdays: opts.Weekdays,
						Months:   opts.Months,
						Busy:     opts.Busy,
						Disabled: opts.Disabled,
						Header:   true,
					})
					if day, picked := cal.Picked(); picked {
						switch {
						case opts.Start.IsZero() || !opts.End.IsZero():
							*opts.Start, *opts.End = day, time.Time{}
						case day.Before(*opts.Start):
							*opts.Start, *opts.End = day, *opts.Start
						default:
							*opts.End = day
						}
						r.changed = true
					}
				}
			})
			if !opts.Start.IsZero() && !opts.End.IsZero() {
				ui.Text(c, span(time.Duration(DaysBetween(*opts.Start, *opts.End))*24*time.Hour)).
					TextColor(k.TextMuted).FontSize(theme.CaptionSize).SingleLine()
			}
		},
	})
	return r
}

// ── clocks ───────────────────────────────────────────────────────────────────

// TimePickerOptions configure a TimePicker.
type TimePickerOptions struct {
	// Value is the time being edited, and is written as the steppers move.
	// Required.
	Value *time.Time
	// MinuteStep is how far the minute steppers go; zero means five, which is
	// the difference between a time a clinic runs on and one it does not.
	MinuteStep int
	// Name is what a screen reader announces; empty uses the time's text.
	Name string
	// ShowSeconds adds a third pair of steppers.
	ShowSeconds bool
	// SecondsStep is how far the seconds steppers go; zero means one.
	SecondsStep int
	// Disabled takes the whole control out of play.
	Disabled bool
}

// TimePickerResult carries a TimePicker and whether it changed.
type TimePickerResult struct {
	// Element is the control.
	Element *ui.Element
	// changed reports that the time moved this frame.
	changed bool
}

// Changed reports that the time moved this frame.
func (r TimePickerResult) Changed() bool { return r.changed }

// TimePicker is a clock, edited with steppers: hour, minute, and seconds if
// asked for.
//
// Steppers and not a list of every half hour, because a list of every half
// hour is fourteen rows for five characters, and because a picker that only
// offers times a clinic does not open is a picker that cannot set a callback
// that runs late. The stepper wraps at the ends — 23 goes to 0 — so that a
// caller booking at 23:50 does not have to think about what comes next.
func TimePicker(c *ui.Context, opts TimePickerOptions) TimePickerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Value == nil {
		panic("datetime: TimePicker needs a Value to point at")
	}
	minuteStep := opts.MinuteStep
	if minuteStep <= 0 {
		minuteStep = 5
	}
	secondStep := opts.SecondsStep
	if secondStep <= 0 {
		secondStep = 1
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.time", "Time")
	}

	// Each part gets its own pair, and the clock is put back together from
	// them, so a step that crosses midnight moves the day rather than
	// inventing an hour that is not on the clock.
	move := func(dh, dm, ds int) {
		next := *opts.Value
		next = next.Add(time.Duration(dh) * time.Hour)
		next = next.Add(time.Duration(dm) * time.Minute)
		next = next.Add(time.Duration(ds) * time.Second)
		*opts.Value = next
	}

	var r TimePickerResult
	r.Element = ui.Row(c).Row().AlignItems(ui.Center).Gap(u).
		Label(name).Disabled(opts.Disabled).Children(func() {
		part := func(label string, value int, up, down string, dh, dm, ds int) {
			ui.Column(c).AlignItems(ui.Center).Gap(u / 2).Label(label).
				Children(func() {
					if stepButton(c, iconUp, up, core.ControlHeight(c)).Clicked() {
						move(dh, dm, ds)
						r.changed = true
					}
					ui.Text(c, pad2(value)).TextColor(k.Text).
						FontSize(theme.BodySize).Bold().FontFeatures("tnum")
					if stepButton(c, iconDown, down, core.ControlHeight(c)).Clicked() {
						move(-dh, -dm, -ds)
						r.changed = true
					}
				})
		}
		part(core.Msg(c, "datetime.hour", "Hour"), opts.Value.Hour(),
			core.Msg(c, "datetime.hourLater", "One hour later"),
			core.Msg(c, "datetime.hourEarlier", "One hour earlier"), 1, 0, 0)
		ui.Text(c, ":").TextColor(k.TextMuted).FontSize(theme.BodySize).Bold()
		part(core.Msg(c, "datetime.minute", "Minute"), opts.Value.Minute(),
			core.Msg(c, "datetime.minuteLater", "One minute later"),
			core.Msg(c, "datetime.minuteEarlier", "One minute earlier"), 0, minuteStep, 0)
		if opts.ShowSeconds {
			ui.Text(c, ":").TextColor(k.TextMuted).FontSize(theme.BodySize).Bold()
			part(core.Msg(c, "datetime.second", "Second"), opts.Value.Second(),
				core.Msg(c, "datetime.secondLater", "One second later"),
				core.Msg(c, "datetime.secondEarlier", "One second earlier"), 0, 0, secondStep)
		}
	})
	return r
}

// pad2 writes a clock part as two digits, so that 09:05 does not read 9:5 in a
// column of times.
func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

// TimeRangePickerOptions configure a TimeRangePicker.
type TimeRangePickerOptions struct {
	// Start and End are the two times, both the caller's.
	Start, End *time.Time
	// MinuteStep is how far the steppers go; zero means five.
	MinuteStep int
	// Name is what a screen reader announces.
	Name string
	// Span writes the length of the range. The default is DurationText.
	Span func(time.Duration) string
	// ShowSeconds adds seconds to both ends.
	ShowSeconds bool
	// Disabled takes the whole control out of play.
	Disabled bool
}

// TimeRangePickerResult carries a TimeRangePicker and whether it changed.
type TimeRangePickerResult struct {
	// Element is the control.
	Element *ui.Element
	// changed reports that the range moved this frame.
	changed bool
}

// Changed reports that the range moved this frame.
func (r TimeRangePickerResult) Changed() bool { return r.changed }

// TimeRangePicker is two clocks and the length between them.
//
// The length is the reason this is one component and not two: a range whose
// two halves are on their own tells a person nothing about whether the meeting
// fits, and the number that tells them is the difference.
func TimeRangePicker(c *ui.Context, opts TimeRangePickerOptions) TimeRangePickerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Start == nil || opts.End == nil {
		panic("datetime: TimeRangePicker needs a Start and an End to point at")
	}
	span := opts.Span
	if span == nil {
		span = DurationText
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.timeRange", "From and until")
	}

	var r TimeRangePickerResult
	// The two clocks are built inside the row, so each lands in it: an
	// element is parented by the Children callback it is created in, and
	// building one outside and placing it later is not a thing MyGo does.
	r.Element = ui.Row(c).Row().FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(name).Children(func() {
		from := TimePicker(c, TimePickerOptions{
			Value:       opts.Start,
			MinuteStep:  opts.MinuteStep,
			Name:        core.Msg(c, "datetime.from", "From"),
			ShowSeconds: opts.ShowSeconds,
			Disabled:    opts.Disabled,
		})
		until := TimePicker(c, TimePickerOptions{
			Value:       opts.End,
			MinuteStep:  opts.MinuteStep,
			Name:        core.Msg(c, "datetime.until", "Until"),
			ShowSeconds: opts.ShowSeconds,
			Disabled:    opts.Disabled,
		})
		r.changed = from.Changed() || until.Changed()
		ui.Text(c, "\u2013").TextColor(k.TextMuted).FontSize(theme.BodySize)
		// The length is the point of the component: two clocks on their own
		// do not say whether the meeting fits.
		ui.Text(c, span(Until(opts.Start, opts.End))).TextColor(k.TextMuted).
			FontSize(theme.CaptionSize)
	})
	return r
}

// Until returns the length from start to end, and never a negative one: a
// range drawn backwards is still a range, and "-1h 30m" in a booking form
// reads as a mistake in the form rather than in the dates.
func Until(start, end *time.Time) time.Duration {
	if end == nil || start == nil {
		return 0
	}
	return max(end.Sub(*start), 0)
}

// DateTimePickerOptions configure a DateTimePicker.
type DateTimePickerOptions struct {
	// Value is the moment being edited: the date picker writes the day into it
	// and the time picker the clock, and neither touches the other's half.
	Value *time.Time
	// Open is the caller's popover state, shared by the date half.
	Open *bool
	// Today is the caller's today.
	Today time.Time
	// Min and Max bound the days that can be chosen.
	Min, Max time.Time
	// MinuteStep is how far the time's minute steppers go; zero means five.
	MinuteStep int
	// Name is what a screen reader announces.
	Name string
	// Format writes the date in its field; the default is FormatDate.
	Format func(time.Time) string
	// Weekdays and Months name the days and months; zero means the default.
	Weekdays WeekdayNames
	Months   MonthNames
	// Busy marks a day as having something on it.
	Busy func(time.Time) bool
	// Disabled greys a day out and takes its clicks away.
	Disabled func(time.Time) bool
}

// DateTimePickerResult carries a DateTimePicker and whether it changed.
type DateTimePickerResult struct {
	// Element is the two controls side by side.
	Element *ui.Element
	// changed reports that the moment moved this frame.
	changed bool
}

// Changed reports that the moment moved this frame.
func (r DateTimePickerResult) Changed() bool { return r.changed }

// DateTimePicker is a date field and a clock side by side, editing one moment
// between them.
//
// The two halves each own their own part of the value — the calendar writes
// the date and leaves the clock alone, the steppers write the clock and leave
// the date alone — because a combined widget that wrote both would reset the
// time to midnight every time somebody picked a day, and a booking that
// silently moved to 00:00 is worse than no widget.
func DateTimePicker(c *ui.Context, opts DateTimePickerOptions) DateTimePickerResult {
	if opts.Value == nil {
		panic("datetime: DateTimePicker needs a Value to point at")
	}
	if opts.Open == nil {
		panic("datetime: DateTimePicker needs the *bool its calendar opens in")
	}
	u := core.Density(c).Unit()
	r := DateTimePickerResult{}
	// Both halves are built inside the row. Each writes its own part of the
	// value and leaves the other alone: a calendar that reset the clock to
	// midnight, or steppers that moved the date, would make a booking change
	// in a way nobody asked for.
	r.Element = ui.Row(c).Row().FillWidth().AlignItems(ui.Center).Gap(u).
		Label(opts.Name).Children(func() {
		day := DatePicker(c, DatePickerOptions{
			Value:    opts.Value,
			Open:     opts.Open,
			Today:    opts.Today,
			Min:      opts.Min,
			Max:      opts.Max,
			Format:   opts.Format,
			Name:     core.Msg(c, "datetime.date", "Date"),
			Weekdays: opts.Weekdays,
			Months:   opts.Months,
			Busy:     opts.Busy,
			Disabled: opts.Disabled,
		})
		clock := TimePicker(c, TimePickerOptions{
			Value:      opts.Value,
			MinuteStep: opts.MinuteStep,
			Name:       core.Msg(c, "datetime.time", "Time"),
		})
		if _, picked := day.Picked(); picked {
			r.changed = true
		}
		if clock.Changed() {
			r.changed = true
		}
	})
	return r
}

// ── months, years, weeks ─────────────────────────────────────────────────────

// MonthPickerOptions configure a MonthPicker.
type MonthPickerOptions struct {
	// Month is the month on screen and the month chosen. The arrows move it
	// by a year, so a caller can page without holding a second variable.
	Month *time.Time
	// Today is the caller's today, drawn as the ring on its month.
	Today time.Time
	// Name is what a screen reader announces.
	Name string
	// Months names the months; zero means the default.
	Months MonthNames
	// Highlighted is the month the caller has open, drawn filled.
	Highlighted *time.Time
	// Columns is how many months across; zero means three, which is twelve in
	// four rows.
	Columns int
	// Disabled greys a month out and takes its clicks away.
	Disabled func(time.Time) bool
}

// MonthPickerResult carries a MonthPicker and the month the user chose.
type MonthPickerResult struct {
	// Element is the whole control.
	Element *ui.Element
	// picked is the month chosen this frame, and got whether there was one.
	picked time.Time
	got    bool
	// stepped is -1 or 1 for the year arrows, 0 for neither.
	stepped int
}

// Picked returns the month the user chose, and whether they chose one.
func (r MonthPickerResult) Picked() (time.Time, bool) { return r.picked, r.got }

// Stepped reports which year arrow was pressed: -1 for the year before, 1 for
// the year after, 0 for neither.
func (r MonthPickerResult) Stepped() int { return r.stepped }

// MonthPicker is twelve months in a grid, with the year between two arrows.
//
// The arrows move the year rather than keeping a cursor of their own, because
// a month picker whose paging is remembered in a place the caller cannot see
// is a component holding state: press next, pick nothing, and the next frame
// would have to agree with the last one about which year it is showing.
func MonthPicker(c *ui.Context, opts MonthPickerOptions) MonthPickerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Month == nil {
		panic("datetime: MonthPicker needs a Month to point at")
	}
	month := *opts.Month
	names := opts.Months.or()
	columns := opts.Columns
	if columns <= 0 || columns > 4 {
		columns = 3
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.month", "Month")
	}
	prev := core.Msg(c, "datetime.prevYear", "Previous year")
	next := core.Msg(c, "datetime.nextYear", "Next year")

	var r MonthPickerResult
	r.Element = ui.Column(c).FillWidth().Gap(u).Label(name).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			if stepButton(c, iconPrev, prev, core.ControlHeight(c)).Clicked() {
				r.stepped = -1
			}
			ui.Text(c, itoa(month.Year())).TextColor(k.Text).
				FontSize(theme.RowSize).Bold().Grow(1).TextAlign(ui.Center).SingleLine()
			if stepButton(c, iconNext, next, core.ControlHeight(c)).Clicked() {
				r.stepped = 1
			}
		})
		for row := 0; row < 12; row += columns {
			ui.Row(c).FillWidth().Gap(u).Children(func() {
				for col := range columns {
					at := time.Date(month.Year(), time.Month(row+col+1), 1, 0, 0, 0, 0, month.Location())
					cell := monthCell(c, names[monthIndex(at)], at,
						opts.Highlighted != nil && SameMonth(*opts.Highlighted, at), opts.Today, opts.Disabled)
					if cell {
						*opts.Month = at
						r.picked, r.got = at, true
					}
				}
			})
		}
	})
	if r.stepped != 0 {
		*opts.Month = AddMonths(month, r.stepped*12)
	}
	return r
}

// monthCell is one month in a month picker. It reports a press rather than
// writing, so the picker can decide what a press means.
func monthCell(c *ui.Context, name string, at time.Time, open bool, today time.Time,
	disabled func(time.Time) bool,
) bool {
	k, u := core.Tokens(c), core.Density(c).Unit()
	off := disabled != nil && disabled(at)
	isToday := !today.IsZero() && SameMonth(at, today)
	bg, fg := ui.Color{}, k.Text
	switch {
	case open:
		bg, fg = k.Fill, k.OnFill
	case off:
		fg = k.TextFaint
	case isToday:
		fg = k.AccentText
	}
	btn := ui.Button(c, "").Grow(1).Height(u * 9).Radius(theme.SmallRadius).
		Background(bg).TextColor(fg).Disabled(off).Label(name + " " + itoa(at.Year())).
		Children(func() {
			ui.Text(c, name).FontSize(theme.RowSize).SingleLine()
		})
	if isToday && !open {
		btn.Border(theme.BorderWidth, k.Accent)
	}
	return btn.Clicked()
}

// YearPickerOptions configure a YearPicker.
type YearPickerOptions struct {
	// Year is the year on screen and the year chosen; the arrows move it by
	// Span.
	Year *time.Time
	// Span is how many years the grid shows; zero means twelve.
	Span int
	// Columns is how many years across; zero means four, which is twelve in
	// three rows.
	Columns int
	// Today is the caller's today, drawn as this year's ring.
	Today time.Time
	// Name is what a screen reader announces.
	Name string
	// Highlighted is the year the caller has open, drawn filled.
	Highlighted *time.Time
	// Disabled greys a year out and takes its clicks away.
	Disabled func(time.Time) bool
}

// YearPickerResult carries a YearPicker and the year the user chose.
type YearPickerResult struct {
	// Element is the whole control.
	Element *ui.Element
	// picked is the year chosen this frame, and got whether there was one.
	picked time.Time
	got    bool
	// stepped is -1 or 1 for the page arrows, 0 for neither.
	stepped int
}

// Picked returns the year the user chose, and whether they chose one.
func (r YearPickerResult) Picked() (time.Time, bool) { return r.picked, r.got }

// Stepped reports which page arrow was pressed: -1 for the page before, 1 for
// the page after, 0 for neither.
func (r YearPickerResult) Stepped() int { return r.stepped }

// YearPicker is a page of years, a page either side of the one on screen.
//
// The page is a whole number of spans from the year zero, so it does not
// creep: a picker whose first year is "the year on screen minus five" gives a
// different set of years every time the year changes, and a person who has
// learned where their year sits in the grid loses it.
func YearPicker(c *ui.Context, opts YearPickerOptions) YearPickerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Year == nil {
		panic("datetime: YearPicker needs a Year to point at")
	}
	span := opts.Span
	if span <= 0 {
		span = 12
	}
	columns := opts.Columns
	if columns <= 0 || columns > 6 {
		columns = 4
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.year", "Year")
	}
	prev := core.Msg(c, "datetime.prevYears", "Earlier years")
	next := core.Msg(c, "datetime.nextYears", "Later years")
	year := opts.Year.Year()
	first := floorDiv(year, span) * span

	var r YearPickerResult
	r.Element = ui.Column(c).FillWidth().Gap(u).Label(name).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			if stepButton(c, iconPrev, prev, core.ControlHeight(c)).Clicked() {
				r.stepped = -1
			}
			ui.Text(c, itoa(first)+" – "+itoa(first+span-1)).TextColor(k.Text).
				FontSize(theme.RowSize).Bold().Grow(1).TextAlign(ui.Center).SingleLine()
			if stepButton(c, iconNext, next, core.ControlHeight(c)).Clicked() {
				r.stepped = 1
			}
		})
		for row := 0; row < span; row += columns {
			ui.Row(c).FillWidth().Gap(u).Children(func() {
				for col := range columns {
					at := time.Date(first+row+col, 1, 1, 0, 0, 0, 0, opts.Year.Location())
					open := opts.Highlighted != nil && opts.Highlighted.Year() == at.Year()
					cell := yearCell(c, at, open, opts.Today, opts.Disabled)
					if cell {
						*opts.Year = at
						r.picked, r.got = at, true
					}
				}
			})
		}
	})
	if r.stepped != 0 {
		*opts.Year = time.Date(year+r.stepped*span, opts.Year.Month(), opts.Year.Day(),
			opts.Year.Hour(), opts.Year.Minute(), 0, 0, opts.Year.Location())
	}
	return r
}

func yearCell(c *ui.Context, at time.Time, open bool, today time.Time,
	disabled func(time.Time) bool,
) bool {
	k, u := core.Tokens(c), core.Density(c).Unit()
	off := disabled != nil && disabled(at)
	isToday := !today.IsZero() && today.Year() == at.Year()
	bg, fg := ui.Color{}, k.Text
	switch {
	case open:
		bg, fg = k.Fill, k.OnFill
	case off:
		fg = k.TextFaint
	case isToday:
		fg = k.AccentText
	}
	btn := ui.Button(c, "").Grow(1).Height(u * 9).Radius(theme.SmallRadius).
		Background(bg).TextColor(fg).Disabled(off).Label(itoa(at.Year())).
		Children(func() {
			ui.Text(c, itoa(at.Year())).FontSize(theme.RowSize).FontFeatures("tnum")
		})
	if isToday && !open {
		btn.Border(theme.BorderWidth, k.Accent)
	}
	return btn.Clicked()
}

// WeekPickerOptions configure a WeekPicker.
type WeekPickerOptions struct {
	// Value is the chosen day, and any day in the chosen week will do: what a
	// week is for is the week, not the day inside it.
	Value *time.Time
	// Weeks is how many weeks to show at once; zero means four.
	Weeks int
	// Today is the caller's today; the zero time draws no today.
	Today time.Time
	// Name is what a screen reader announces.
	Name string
	// Weekdays names the columns; zero means the default.
	Weekdays WeekdayNames
	// Months names the months, for the day names read aloud; zero means the
	// default.
	Months MonthNames
	// Disabled greys a day out and takes its clicks away.
	Disabled func(time.Time) bool
}

// WeekPickerResult carries a WeekPicker and the day the user pressed.
type WeekPickerResult struct {
	// Element is the row of weeks.
	Element *ui.Element
	// picked is the day pressed this frame, and got whether there was one.
	picked time.Time
	got    bool
}

// Picked returns the day the user pressed, and whether they pressed one.
func (r WeekPickerResult) Picked() (time.Time, bool) { return r.picked, r.got }

// WeekPicker is a row of weeks, each a column of seven days — the compact
// form a booking page uses when somebody has said "some day next month" and
// needs to be shown the days without being shown a month.
//
// The week on screen is the middle of the row, so that there are always weeks
// both before and after the chosen one: a week picker that can only look
// forward is a forward-only booking system.
func WeekPicker(c *ui.Context, opts WeekPickerOptions) WeekPickerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Value == nil {
		panic("datetime: WeekPicker needs a Value to point at")
	}
	weeks := opts.Weeks
	if weeks <= 0 {
		weeks = 4
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.week", "Week")
	}
	weekdays := opts.Weekdays.or()
	months := opts.Months.or()
	// The chosen week sits in the middle; with an even number of weeks the
	// extra one goes after, so the row reads forwards from where the choice
	// was.
	anchor := *opts.Value
	if anchor.IsZero() {
		anchor = opts.Today
	}
	if anchor.IsZero() {
		panic("datetime: WeekPicker needs a Value or a Today to show weeks around")
	}
	first := StartOfWeek(anchor).AddDate(0, 0, -weeks/2)

	var r WeekPickerResult
	r.Element = ui.Row(c).Row().FillWidth().Gap(u * 1.5).
		Label(name + " " + WeekViewTitle(anchor, months)).Children(func() {
		for w := range weeks {
			ui.Column(c).FillWidth().Gap(u / 2).Children(func() {
				ui.Text(c, itoa(WeekOfYear(first.AddDate(0, 0, w*7)))).
					TextColor(k.TextFaint).FontSize(theme.CaptionSize).
					TextAlign(ui.Center).SingleLine()
				for d := range 7 {
					day := first.AddDate(0, 0, w*7+d)
					cell := weekCell(c, day, weekdays[Weekday(day)],
						opts.Value != nil && SameDay(*opts.Value, day),
						!opts.Today.IsZero() && SameDay(opts.Today, day),
						opts.Disabled)
					if cell {
						*opts.Value = day
						r.picked, r.got = day, true
					}
				}
			})
		}
	})
	return r
}

func weekCell(c *ui.Context, day time.Time, weekday string, selected, today bool,
	disabled func(time.Time) bool,
) bool {
	k, u := core.Tokens(c), core.Density(c).Unit()
	off := disabled != nil && disabled(day)
	bg, fg := ui.Color{}, k.Text
	switch {
	case selected:
		bg, fg = k.Fill, k.OnFill
	case off:
		fg = k.TextFaint
	case today:
		fg = k.AccentText
	}
	// No Grow: the cells are stacked in a column, where a flex grow shares the
	// leftover space downwards and a cell with a fixed height has none to
	// share. The width is what has to fill here, and FillWidth does that.
	//
	// Each cell carries its weekday above its number, which is what turns a
	// row of days into a week: seven numbers in a column are a column.
	btn := ui.Button(c, "").FillWidth().MinWidth(MinCellSize).Height(u * 9).
		Radius(theme.SmallRadius).Background(bg).TextColor(fg).Disabled(off).
		Label(dayNameOf(day, DefaultMonths())).
		Children(func() {
			ui.Text(c, weekday).TextColor(k.TextFaint).FontSize(theme.CaptionSize).
				SingleLine()
			ui.Text(c, itoa(day.Day())).FontSize(theme.RowSize).Bold().SingleLine()
		})
	if today && !selected {
		btn.Border(theme.BorderWidth, k.Accent)
	}
	return btn.Clicked()
}
