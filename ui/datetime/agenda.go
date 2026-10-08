package datetime

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── the agenda ───────────────────────────────────────────────────────────────

// AgendaDay is one day of an agenda: the date and the events on it.
type AgendaDay struct {
	// Day is the date.
	Day time.Time
	// Events are the day's events, in any order.
	Events []Event
}

// AgendaOptions configure an AgendaView.
type AgendaOptions struct {
	// Days are the days to show, in the order given. The caller has already
	// decided which days those are — the next seven with bookings, say — and
	// this component does not go looking for them, because which days are
	// worth a person's attention is the application's decision.
	Days []AgendaDay
	// Today is the caller's today; the zero time highlights nothing.
	Today time.Time
	// Format writes a date in a day heading; the default is FormatDate.
	Format func(time.Time) string
	// Weekdays and Months name the days; zero means the default.
	Weekdays WeekdayNames
	Months   MonthNames
	// Limit is how many events one day shows before the rest become a count.
	// Zero shows them all.
	Limit int
	// Empty draws instead of a day that has nothing on it, so a caller can
	// say "free" in its own words. Nil skips the day entirely, because a list
	// of empty days is a list of nothing.
	Empty func(day time.Time)
	// Pressed is called with the event a chip was pressed on, so a caller can
	// open it. Nil draws chips that take no clicks.
	Pressed func(Event)
}

// AgendaView is a list of days and what is on them, as a booking page reads
// downwards rather than across.
//
// It is the view for a question the calendar views cannot answer without a lot
// of scrolling: what is coming, in what order, and how much of it is there. A
// day with nothing on it is skipped when the caller left Empty nil, because a
// list of empty days is a list of nothing.
func AgendaView(c *ui.Context, opts AgendaOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	format := opts.Format
	if format == nil {
		format = FormatDate
	}
	months := opts.Months.or()

	return ui.Column(c).FillWidth().Gap(u * 2).Children(func() {
		for _, d := range opts.Days {
			events := append([]Event(nil), d.Events...)
			SortEvents(events)
			if len(events) == 0 {
				if opts.Empty != nil {
					opts.Empty(d.Day)
				}
				continue
			}
			day := StartOfDay(d.Day)
			isToday := !opts.Today.IsZero() && SameDay(opts.Today, day)
			ui.Column(c).FillWidth().Gap(u * 1.5).Label(
				WeekdayFullName(day) + " " + FormatDate(day)).Children(func() {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
					// The date is a chip of its own rather than text: a reader
					// scanning the list is looking for the shape of a day, and
					// today's has to be findable without reading it.
					badge, fg := k.Surface, k.Text
					if isToday {
						badge, fg = k.Fill, k.OnFill
					}
					ui.Box(c).Padding(u*0.75, u*1.5).Radius(theme.PillRadius).
						Background(badge).Label(format(day)).Children(func() {
						ui.Text(c, format(day)).TextColor(fg).FontSize(theme.CaptionSize)
					})
					ui.Text(c, WeekdayFullName(day)+" "+months[monthIndex(day)]).
						TextColor(k.TextMuted).FontSize(theme.CaptionSize).SingleLine()
				})
				ui.Column(c).FillWidth().Gap(u).Children(func() {
					shown := len(events)
					if opts.Limit > 0 && shown > opts.Limit {
						shown = opts.Limit
					}
					for _, e := range events[:shown] {
						e := e
						ch := EventChip(c, EventChipOptions{Event: e, Format: format})
						if opts.Pressed != nil && ch.Pressed() {
							opts.Pressed(e)
						}
					}
					if rest := len(events) - shown; rest > 0 {
						ui.Text(c, internal.Commas(rest)+" "+
							core.Msg(c, "datetime.more", "more")).TextColor(k.TextMuted).
							FontSize(theme.CaptionSize)
					}
				})
			})
		}
	})
}

// ── the people ───────────────────────────────────────────────────────────────

// Response is what an attendee has said about an invitation.
type Response int

const (
	// NoReply is the zero value: nobody has answered, which is most of them.
	NoReply Response = iota
	// Accepted.
	Accepted
	// Maybe.
	Maybe
	// Declined.
	Declined
)

func (r Response) String() string {
	switch r {
	case Accepted:
		return "Yes"
	case Maybe:
		return "Maybe"
	case Declined:
		return "No"
	}
	return "No reply"
}

// severity is the response's colour: a declined is quieter than an accepted,
// and a nobody-said is quieter than both. A list of guests that is mostly grey
// says what the meeting is missing.
func (r Response) severity() core.Severity {
	switch r {
	case Accepted:
		return core.Success
	case Maybe:
		return core.Warning
	case Declined:
		return core.Danger
	}
	return core.Neutral
}

// Attendee is one person on an invitation.
type Attendee struct {
	// Name is the person, and the initials in their circle.
	Name string
	// Response is what they said.
	Response Response
	// Note is anything the organiser wrote about them, which belongs under
	// the name rather than in a tooltip nobody on a desktop will ever see.
	Note string
	// Timezone is their zone, when the caller's data has it. It is shown so
	// that somebody in Auckland can see they are the one waking up early.
	Timezone string
}

// AttendeeOptions configure an AttendeeList.
type AttendeeOptions struct {
	// Attendees are the people, in the order the caller wants them.
	Attendees []Attendee
	// Max is how many are shown before the rest become a count. Zero shows
	// them all.
	Max int
	// Avatars draws each person's initials in a circle, which is faster to
	// scan than a column of names.
	Avatars bool
	// ShowResponse writes each answer beside the name.
	ShowResponse bool
	// Empty draws when there is nobody; nil draws nothing at all, which is what
	// a form wants for an invitation nobody has been sent yet.
	Empty func()
}

// AttendeeList is the people on an invitation: their names, their answers, and
// a count for the ones too many to draw.
//
// The answers are words and not colours alone — "Yes", "No", "Maybe" — because
// a list of green and grey circles is a puzzle, and a person deciding whether
// to chase someone needs to know that the grey one has not answered rather
// than that they have said no.
func AttendeeList(c *ui.Context, opts AttendeeOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if len(opts.Attendees) == 0 {
		if opts.Empty != nil {
			return ui.Box(c).FillWidth().Children(opts.Empty)
		}
		return ui.Box(c)
	}
	shown := len(opts.Attendees)
	if opts.Max > 0 && shown > opts.Max {
		shown = opts.Max
	}

	return ui.Column(c).FillWidth().Gap(u).Children(func() {
		for _, a := range opts.Attendees[:shown] {
			row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).
				Label(a.Name)
			row.Children(func() {
				if opts.Avatars {
					circle := ui.Box(c).Size(u*7, u*7).Radius(u * 3.5).
						Background(k.Surface).Label(a.Name)
					circle.Children(func() {
						if in := internal.Initials(a.Name); in != "" {
							ui.Text(c, in).TextColor(k.Text).
								FontSize(theme.MonoSize).Bold()
						}
					})
				}
				ui.Column(c).Grow(1).Children(func() {
					ui.Text(c, a.Name).TextColor(k.Text).
						FontSize(theme.RowSize).SingleLine()
					if a.Note != "" {
						ui.Text(c, a.Note).TextColor(k.TextMuted).
							FontSize(theme.CaptionSize).SingleLine()
					}
				})
				if a.Timezone != "" {
					ui.Text(c, a.Timezone).TextColor(k.TextFaint).
						FontSize(theme.CaptionSize).SingleLine()
				}
				if opts.ShowResponse {
					_, fg := a.Response.severity().Pair(k)
					ui.Text(c, a.Response.String()).TextColor(fg).
						FontSize(theme.CaptionSize).SingleLine()
				}
			})
		}
		if rest := len(opts.Attendees) - shown; rest > 0 {
			ui.Text(c, "+"+internal.Commas(rest)).TextColor(k.TextMuted).
				FontSize(theme.CaptionSize)
		}
	})
}

// ── availability ─────────────────────────────────────────────────────────────

// AvailabilityOptions configure an AvailabilityPicker.
type AvailabilityOptions struct {
	// Days are the days to show, in the order given.
	Days []time.Time
	// StartHour and EndHour are the hours offered; zero means a working day.
	StartHour, EndHour int
	// HourHeight is the scale of the rows; zero uses HourHeight for the
	// density.
	HourHeight float32
	// Free reports whether an hour is available. It is the caller's, because
	// what is free is the caller's data: this component draws the answer and
	// reports the press, and it has no opinion about bookings.
	Free func(day time.Time, hour int) bool
	// Marked is the caller's own set of hours, the ones a person has already
	// picked. A press is reported rather than written, so a caller that wants
	// a different rule — toggle, or only ever add — can have it.
	Marked func(day time.Time, hour int) bool
	// Toggle writes a press into the caller's set; nil draws cells that take
	// no clicks.
	Toggle func(day time.Time, hour int)
	// Weekdays and Months name the days; zero means the default.
	Weekdays WeekdayNames
	Months   MonthNames
}

// AvailabilityResult carries an AvailabilityPicker and the cell pressed.
type AvailabilityResult struct {
	// Element is the grid.
	Element *ui.Element
	// picked is the hour pressed this frame, and got whether there was one.
	picked time.Time
	got    bool
}

// Picked returns the hour the user pressed, and whether they pressed one.
func (r AvailabilityResult) Picked() (time.Time, bool) { return r.picked, r.got }

// AvailabilityPicker is a grid of days against hours, for a person saying when
// they can come in.
//
// It is the only component here that draws its cells as hours rather than
// days, and it is what a booking page shows somebody who has been offered a
// choice of times. Three states, told apart without colour alone: a free hour
// is empty, a taken one is filled, and a picked one is filled with ink and
// carries the hour's own number in the fill's colour.
func AvailabilityPicker(c *ui.Context, opts AvailabilityOptions) AvailabilityResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	start, end := hourWindow(opts.StartHour, opts.EndHour)
	weekdays := opts.Weekdays.or()
	months := opts.Months.or()
	free := opts.Free
	if free == nil {
		// With nothing to go on, everything is free. The alternative is a
		// grid of nothing with no way to say why, which reads as a bug.
		free = func(time.Time, int) bool { return true }
	}

	var r AvailabilityResult
	r.Element = ui.Row(c).Row().FillWidth().Children(func() {
		// The gutter is built to the day columns' exact rhythm — same gap,
		// same header height, and per hour the height of one availability
		// cell. HourHeight here reserved a full agenda hour per label while
		// the cells advanced at their own (smaller) pitch, so the labels
		// drifted out of alignment and kept printing after the grid ended.
		ui.Column(c).Width(u * 9).Shrink(0).FillHeight().Gap(u / 2).Children(func() {
			ui.Box(c).Width(1).Height(u * 9)
			for hour := start; hour < end; hour++ {
				ui.Text(c, HourLabel(hour)).TextColor(k.TextFaint).
					FontSize(theme.CaptionSize).Height(u * 8).
					TextAlign(ui.End).SingleLine()
			}
		})
		ui.Row(c).FillWidth().FillHeight().Gap(u / 2).Children(func() {
			for _, d := range opts.Days {
				day := StartOfDay(d)
				ui.Column(c).FillWidth().FillHeight().Gap(u / 2).Children(func() {
					ui.Box(c).FillWidth().Height(u * 9).Radius(theme.SmallRadius).
						Background(k.Surface).Center().
						Label(dayNameOf(day, months)).Children(func() {
						ui.Text(c, weekdays[Weekday(day)]).TextColor(k.TextMuted).
							FontSize(theme.CaptionSize).SingleLine()
						ui.Text(c, itoa(day.Day())).TextColor(k.Text).
							FontSize(theme.RowSize).Bold().SingleLine()
					})
					for hour := start; hour < end; hour++ {
						at := time.Date(day.Year(), day.Month(), day.Day(), hour, 0, 0, 0, day.Location())
						open := free(day, hour)
						marked := opts.Marked != nil && opts.Marked(day, hour)
						cell := availabilityCell(c, HourLabel(hour), at, open, marked)
						if cell && opts.Toggle != nil {
							opts.Toggle(day, hour)
							r.picked, r.got = at, true
						}
					}
				})
			}
		})
	})
	return r
}

func availabilityCell(c *ui.Context, label string, at time.Time, open, marked bool) bool {
	k, u := core.Tokens(c), core.Density(c).Unit()
	bg, fg := k.Surface, k.TextFaint
	switch {
	case marked:
		bg, fg = k.Fill, k.OnFill
	case open:
		bg, fg = ui.Color{}, k.Text
	}
	btn := ui.Button(c, "").FillWidth().Height(u * 8).Shrink(0).
		Radius(theme.SmallRadius).Background(bg).TextColor(fg).
		Label(label + " " + FormatDate(at)).Children(func() {
		ui.Text(c, label).FontSize(theme.CaptionSize).FontFeatures("tnum")
	})
	return btn.Clicked()
}

// ── lengths ──────────────────────────────────────────────────────────────────

// DurationPickerOptions configure a DurationPicker.
type DurationPickerOptions struct {
	// Value is the length being edited, and is written as the steppers move.
	// Required.
	Value *time.Duration
	// Step is how far the steppers go; zero means fifteen minutes, which is
	// the length of a callback appointment.
	Step time.Duration
	// Presets are the lengths offered as buttons under the steppers, in
	// order. Nil draws none.
	Presets []time.Duration
	// Name is what a screen reader announces.
	Name string
	// Format writes a length; the default is DurationText.
	Format func(time.Duration) string
	// Min and Max bound the length; a zero bound is open.
	Min, Max time.Duration
}

// DurationPickerResult carries a DurationPicker and whether it changed.
type DurationPickerResult struct {
	// Element is the control.
	Element *ui.Element
	// changed reports that the length moved this frame.
	changed bool
}

// Changed reports that the length moved this frame.
func (r DurationPickerResult) Changed() bool { return r.changed }

// DurationPicker is a length: a value, two steppers, and the few lengths that
// get used every day as buttons under it.
//
// The presets are the reason to build this rather than reach for a pair of
// steppers. "30 minutes", "1 hour", "2 hours", "half a day" cover almost every
// job a callback takes, and a person picking one of them should not have to
// press a button twelve times to get there.
func DurationPicker(c *ui.Context, opts DurationPickerOptions) DurationPickerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Value == nil {
		panic("datetime: DurationPicker needs a Value to point at")
	}
	step := opts.Step
	if step <= 0 {
		step = 15 * time.Minute
	}
	format := opts.Format
	if format == nil {
		format = DurationText
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.duration", "Duration")
	}
	clamp := func(d time.Duration) time.Duration {
		if opts.Min > 0 && d < opts.Min {
			return opts.Min
		}
		if opts.Max > 0 && d > opts.Max {
			return opts.Max
		}
		return d
	}
	move := func(by time.Duration) {
		*opts.Value = clamp(*opts.Value + by)
	}

	var r DurationPickerResult
	r.Element = ui.Column(c).FillWidth().Gap(u).Label(name).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			if stepButton(c, iconDown, core.Msg(c, "datetime.shorter", "Shorter"),
				core.ControlHeight(c)).Clicked() {
				move(-step)
				r.changed = true
			}
			ui.Text(c, format(*opts.Value)).TextColor(k.Text).
				FontSize(theme.BodySize).Bold().Grow(1).TextAlign(ui.Center).
				FontFeatures("tnum")
			if stepButton(c, iconUp, core.Msg(c, "datetime.longer", "Longer"),
				core.ControlHeight(c)).Clicked() {
				move(step)
				r.changed = true
			}
		})
		if len(opts.Presets) > 0 {
			ui.Row(c).FillWidth().Gap(u).Children(func() {
				for _, p := range opts.Presets {
					p := p
					chosen := *opts.Value == p
					bg, fg := k.Surface, k.Text
					if chosen {
						bg, fg = k.Fill, k.OnFill
					}
					label := format(p)
					// The button is built here rather than through
					// input.Button, because the chosen one carries a dot as
					// well as its length: an input.Button brings its own text
					// with it and would show the length twice.
					btn := ui.Button(c, "").Grow(1).Height(core.ControlHeight(c)).
						Radius(theme.PillRadius).Background(bg).TextColor(fg).
						Label(name + " " + label).Children(func() {
						ui.Text(c, label).TextColor(fg).FontSize(theme.CaptionSize).SingleLine()
						if chosen {
							ui.Box(c).Width(u * 1.25).Height(u * 1.25).
								Radius(u).Background(k.Lively)
						}
					})
					if btn.Clicked() {
						*opts.Value = clamp(p)
						r.changed = true
					}
				}
			})
		}
	})
	return r
}

// ── zones ────────────────────────────────────────────────────────────────────

// LoadZone returns the named zone, or UTC when the platform has no such zone.
//
// It returns rather than panics because a zone database is a thing the
// operating system may not ship: an application that cannot find "Europe/Berlin"
// should still run, in UTC, rather than stop on a machine that has never had
// a reason to install the files.
func LoadZone(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// ZoneOffset returns how far ahead of UTC a zone is at the given moment, in
// minutes, and its abbreviation.
//
// The moment matters: a zone that is one hour ahead in March is two hours
// ahead in July, and a booking form that shows the winter offset all summer
// is a booking that is an hour wrong twice a year.
func ZoneOffset(loc *time.Location, at time.Time) (int, string) {
	if loc == nil {
		loc = time.UTC
	}
	name, offset := at.In(loc).Zone()
	return int(offset) / 60, name
}

// ZoneLabel writes a zone as it is shown beside a time: the city, and how far
// ahead of UTC it is at the moment being shown.
func ZoneLabel(name string, at time.Time) string {
	minutes, abbr := ZoneOffset(LoadZone(name), at)
	sign := "+"
	if minutes < 0 {
		sign = "−"
		minutes = -minutes
	}
	offset := "UTC" + sign + pad2(minutes/60)
	if minutes%60 != 0 {
		offset += ":" + pad2(minutes%60)
	}
	if abbr != "" && abbr != "UTC" {
		return name + " · " + offset + " " + abbr
	}
	return name + " · " + offset
}

// TimezoneSelectOptions configure a TimezoneSelect.
type TimezoneSelectOptions struct {
	// Value is the chosen zone's name, as the IANA database spells it, and is
	// written as the selection changes. Required.
	Value *string
	// Zones are the names on offer. The caller's list, because which zones a
	// booking page offers is a business decision and a list of every zone in
	// the database is a list of six hundred.
	Zones []string
	// At is the moment the offsets beside the names are worked out for.
	At time.Time
	// Name is what a screen reader announces.
	Name string
	// ShowOffsets writes the offset beside each zone; without it a person
	// choosing a zone for a call has to know the cities by heart.
	ShowOffsets bool
	// Label writes a zone's own name; the default is the name itself.
	Label func(zone string) string
}

// TimezoneSelect is a drop-down of the caller's zones, each with how far
// ahead of UTC it is at the moment being shown.
//
// It is a Select rather than a picker of its own because a zone is a long list
// of names with no order a person can navigate by eye except alphabetically,
// and a list that has to be scrolled in seven columns to find a city is a
// list nobody uses.
func TimezoneSelect(c *ui.Context, opts TimezoneSelectOptions) *ui.Element {
	if opts.Value == nil {
		panic("datetime: TimezoneSelect needs a Value to point at")
	}
	if len(opts.Zones) == 0 {
		panic("datetime: TimezoneSelect needs at least one zone to choose from")
	}
	if *opts.Value == "" {
		*opts.Value = opts.Zones[0]
	}
	label := opts.Label
	if label == nil {
		label = func(zone string) string {
			if !opts.ShowOffsets {
				return zone
			}
			return ZoneLabel(zone, opts.At)
		}
	}
	// MyGo's Select chooses among the strings it is given, so it is given the
	// labels and the choice is translated back: a window with a German
	// calendar wants to read "Berlin · UTC+01:00" and the form to save
	// "Europe/Berlin". The labels have to be distinct for that to be
	// unambiguous, which is what Label is for.
	labels := make([]string, len(opts.Zones))
	for i, z := range opts.Zones {
		labels[i] = label(z)
	}
	picked := 0
	for i, z := range opts.Zones {
		if z == *opts.Value {
			picked = i
		}
	}
	shown := append([]string(nil), labels...)
	sel := ui.Select(c, &shown[picked], shown)
	if shown[picked] != labels[picked] {
		for i, l := range labels {
			if l == shown[picked] {
				*opts.Value = opts.Zones[i]
				break
			}
		}
	}
	return sel.Label(opts.Name)
}
