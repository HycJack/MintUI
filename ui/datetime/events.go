package datetime

import (
	"sort"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// DefaultSlot is how long an event lasts when it has a start and no end.
//
// Thirty minutes, because that is the length of a callbacks appointment and
// of a calendar's own default click. It is a default rather than an error:
// plenty of real events have no end time, and refusing to draw those would
// leave holes in a column.
const DefaultSlot = 30 * time.Minute

// Event is one thing happening at a time.
//
// It is the caller's type, not the library's: this package draws events and
// knows nothing about what one is for. The times carry a location, and it is
// the location they arrive in that the day columns are read in — a booking
// stored in UTC and a reader in Berlin need the view to say 15:00 rather than
// 14:00, and the only place that can be decided is here, where the event and
// the column meet.
type Event struct {
	// Title is what the event is called. It is required: an event with no
	// name is a block of colour.
	Title string
	// Start is when it begins.
	Start time.Time
	// End is when it finishes. Zero, or an end before the start, means
	// DefaultSlot.
	End time.Time
	// AllDay puts the event in the day's own row rather than on the clock.
	AllDay bool
	// Location is shown beside the title when there is room.
	Location string
	// Severity is the event's status colour, through core.Severity, so that
	// the urgent and the cancelled are told apart the same way here as they
	// are everywhere else in the library.
	Severity core.Severity
	// Cancelled strikes the title through and quiets the colour. It is a
	// field rather than a severity because a cancelled 3am maintenance is
	// still a thing worth noticing in a list, just not on the clock.
	Cancelled bool
}

// Span returns the event's start and end, with a missing end filled in.
func (e Event) Span() (start, end time.Time) {
	start = e.Start
	end = e.End
	if end.IsZero() || !end.After(start) {
		end = start.Add(DefaultSlot)
	}
	return start, end
}

// Duration returns how long the event lasts.
func (e Event) Duration() time.Duration {
	start, end := e.Span()
	return end.Sub(start)
}

// Overlaps reports whether a and b are on at the same time. Two events that
// only touch do not overlap: 09:00–10:00 and 10:00–11:00 are a schedule, not
// a conflict.
func Overlaps(a, b Event) bool {
	as, ae := a.Span()
	bs, be := b.Span()
	return as.Before(be) && bs.Before(ae)
}

// EventsOn returns the events touching the day of t, all-day ones included
// whenever they span it, ordered by when they start.
func EventsOn(events []Event, t time.Time) []Event {
	day := StartOfDay(t)
	next := day.AddDate(0, 0, 1)
	var out []Event
	for _, e := range events {
		start, end := e.Span()
		if start.Before(next) && end.After(day) {
			out = append(out, e)
		}
	}
	SortEvents(out)
	return out
}

// SortEvents orders events by start and then by title, so two events in the
// same minute do not swap places between frames.
func SortEvents(events []Event) {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Start.Equal(events[j].Start) {
			return events[i].Title < events[j].Title
		}
		return events[i].Start.Before(events[j].Start)
	})
}

// OverlapLanes splits events into the fewest groups within which no two events
// overlap — the lanes an hour column draws side by side.
//
// It is the standard greedy sweep: walk the events by start time and drop each
// into the first lane whose last event has finished, opening a lane when none
// has. The result is the smallest number of columns that fits, which is what a
// person means when they say a morning is "two deep". An event with nothing
// beside it gets a lane to itself and is drawn full width.
func OverlapLanes(events []Event) [][]Event {
	lanes := make([][]Event, 0, len(events))
	var ends []time.Time
	for _, e := range events {
		_, end := e.Span()
		placed := false
		for i, last := range ends {
			if !last.After(e.Start) {
				lanes[i] = append(lanes[i], e)
				ends[i] = end
				placed = true
				break
			}
		}
		if !placed {
			lanes = append(lanes, []Event{e})
			ends = append(ends, end)
		}
	}
	return lanes
}

// EventChipOptions configure an EventChip.
type EventChipOptions struct {
	// Event is what the chip shows. Required.
	Event Event
	// NoTime drops the clock, for a chip that is showing a day rather than a
	// moment.
	NoTime bool
	// Format writes the clock. It defaults to FormatTime.
	Format func(time.Time) string
	// Selected fills the chip, for the one the caller is looking at.
	Selected bool
}

// EventChipResult carries an EventChip and whether it was pressed.
type EventChipResult struct {
	// Element is the chip.
	Element *ui.Element
	// pressed is whether it took a click this frame.
	pressed bool
}

// Pressed reports a press of the chip.
func (r EventChipResult) Pressed() bool { return r.pressed }

// EventChip is one event as a bar: its colour from its severity, its title,
// and the clock it starts at.
//
// The colour comes from core.Severity rather than from a caller's own choice,
// because the alternative is a column of events whose colours mean whatever
// each author of them felt like. A neutral event stays on the surface, so the
// ordinary ones do not shout and the ones that matter are the ones in colour.
func EventChip(c *ui.Context, opts EventChipOptions) EventChipResult {
	if opts.Event.Title == "" {
		panic("datetime: EventChip needs a titled event; an event with no name is a block of colour")
	}
	e := chip(c, opts.Event, theme.CaptionSize, !opts.NoTime, opts.Selected)
	return EventChipResult{Element: e, pressed: e.Clicked()}
}

// chip draws the inside of an event chip, at the size the caller needs: a
// list shows the clock beside the title, a column inside a fixed height shows
// both in one line that the height then clips.
//
// The leading bar carries the severity. The chip's own background is pale, and
// a column of pale chips would read as a column of nothing at all without
// something in it saying which of them mattered.
func chip(c *ui.Context, e Event, size float32, showTime, selected bool) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	format := FormatTime

	bg, fg := e.Severity.Pair(k)
	if e.Severity == core.Neutral {
		bg, fg = k.Surface, k.Text
	}
	bar := fg
	if e.Severity == core.Neutral {
		bar = k.Border
	}
	switch {
	case e.Cancelled:
		bg, fg, bar = k.Surface, k.TextMuted, k.Border
	case selected:
		bg, fg, bar = k.Fill, k.OnFill, k.OnFill
	}

	text := e.Title
	label := e.Title
	switch {
	case !showTime:
		label = WeekdayFullName(e.Start) + " " + FormatDate(e.Start) + " " + text
	case e.Location != "":
		text = format(e.Start) + "  " + text + " · " + e.Location
		label = WeekdayFullName(e.Start) + " " + FormatDate(e.Start) + " " + text
	default:
		text = format(e.Start) + "  " + text
		label = WeekdayFullName(e.Start) + " " + FormatDate(e.Start) + " " + text
	}

	el := ui.Box(c).Row().FillWidth().Shrink(0).Padding(u*0.75, u*1.5, u*0.75, u*1.5).
		Radius(theme.SmallRadius).Background(bg).Label(label).Children(func() {
		// The bar is a child rather than a border: a border rounds with the
		// chip and draws a line along three of its sides as well.
		ui.Box(c).Width(u * 0.5).FillHeight().Shrink(0).Radius(u * 0.25).
			Background(bar)
		t := ui.Text(c, text).TextColor(fg).FontSize(size).SingleLine().Grow(1)
		if e.Cancelled {
			t.Strikethrough()
		}
	})
	return el
}

// CalendarEventsOptions configure a CalendarEvents.
type CalendarEventsOptions struct {
	// Events are the day's events as they are: unsorted input is sorted here
	// rather than demanded in order, because a caller's query is a query and
	// not a promise about what comes back.
	Events []Event
	// Day is the day the column belongs to, and is what the events' clocks
	// are read against.
	Day time.Time
	// StartHour and EndHour are the hours on screen; zero means 8 to 18.
	StartHour, EndHour int
	// HourHeight is the scale of the column; zero uses the default.
	HourHeight float32
	// AllDay draws the all-day events in a row of their own above the clock,
	// which is where a calendar puts them and the only place they fit.
	AllDay bool
	// Format writes the clock on a chip.
	Format func(time.Time) string
	// Selected is the event the caller is looking at, or nil.
	Selected *Event
}

// CalendarEventsResult carries the events of one day and what was pressed.
type CalendarEventsResult struct {
	// Element is the column: the all-day row, the hour rules, and a chip at
	// each event's own height.
	Element *ui.Element
	// chosen is the event pressed this frame, and got whether there was one.
	chosen *Event
	got    bool
	// empty is whether the day has nothing on it at all.
	empty bool
}

// Chosen returns the event the user pressed, and whether they pressed one.
func (r CalendarEventsResult) Chosen() (*Event, bool) { return r.chosen, r.got }

// Empty reports whether the day has nothing on it, which is what a caller
// uses to decide between the column and an explanation of its own.
func (r CalendarEventsResult) Empty() bool { return r.empty }

// CalendarEvents lays one day out as an hour column: a rule at every hour, a
// chip at each event's start and height, and the all-day events in a row
// above.
//
// The chips sit at the height their clock gives them rather than in a list,
// because the position is the information: a 09:00 drawn at the same height
// as a 14:00 tells a reader they are the same time of day, and a list cannot
// say that. Overlapping events go into lanes side by side, so a double-booked
// hour is visible as a double booking rather than as one event drawn on top of
// another.
func CalendarEvents(c *ui.Context, opts CalendarEventsOptions) CalendarEventsResult {
	u := core.Density(c).Unit()
	start, end := hourWindow(opts.StartHour, opts.EndHour)
	hourHeight := opts.HourHeight
	if hourHeight <= 0 {
		hourHeight = HourHeight(u)
	}
	var r CalendarEventsResult
	r.Element = dayColumn(c, opts, start, end, hourHeight, &r, nil)
	return r
}

// dayColumn is one day of a time view: the all-day events in a row, a
// divider, and the box the hour rules and the event chips share.
//
// rules draws the backdrop and is given by the caller, because the backdrop is
// the one part that differs between the three views: a day view's hours can be
// pressed and carry the now line, a week view's cannot, and a bare column's
// are only lines. Nil draws the plain ruled column.
func dayColumn(c *ui.Context, opts CalendarEventsOptions, start, end int,
	hourHeight float32, r *CalendarEventsResult, rules func(),
) *ui.Element {
	u := core.Density(c).Unit()
	day := StartOfDay(opts.Day)
	timed, allDay := splitEvents(opts.Events, day)
	r.empty = len(timed) == 0 && len(allDay) == 0
	lanes := OverlapLanes(timed)

	return ui.Column(c).FillWidth().FillHeight().Children(func() {
		if opts.AllDay && len(allDay) > 0 {
			ui.Column(c).FillWidth().Gap(u / 2).Children(func() {
				for _, e := range allDay {
					e := e
					row := chip(c, e, theme.CaptionSize, false, false)
					if row.Clicked() {
						r.chosen, r.got = &e, true
					}
				}
			})
			layout.Divider(c, layout.DividerOptions{})
		}
		// The body is a box rather than a column so the chips can be placed
		// absolutely inside it: the box's height is the hours and nothing
		// else, so a position inside it means a time of day.
		ui.Box(c).FillWidth().Height(float32(end-start) * hourHeight).Children(func() {
			if rules != nil {
				rules()
			} else {
				CalendarTimeGrid(c, TimeGridOptions{
					Day:        day,
					StartHour:  start,
					EndHour:    end,
					HourHeight: hourHeight,
				})
			}
			for lane, group := range lanes {
				for _, e := range group {
					e := e
					pos := chip(c, e, theme.CaptionSize, true,
						opts.Selected != nil && opts.Selected.Title == e.Title &&
							opts.Selected.Start.Equal(e.Start))
					share := 100.0 / float32(len(lanes))
					pos.Absolute().
						Top(HourY(e.Start, start, hourHeight)).
						// LeftPercent, not Left: a lane's offset is a share
						// of the column's width, and the column's width is
						// what an absolute inset is measured against.
						LeftPercent(float32(lane)*share).
						WidthPercent(share).
						Height(max(eventHeight(e, start, end, hourHeight), hourHeight/4)).
						Justify(ui.Center).Margin(0, u*0.5, 0, u*0.5)
					if pos.Clicked() {
						r.chosen, r.got = &e, true
					}
				}
			}
		})
	})
}

// splitEvents sorts a day's events and separates the all-day ones, which
// belong in a row above the clock rather than on it.
func splitEvents(events []Event, day time.Time) (timed, allDay []Event) {
	for _, e := range EventsOn(events, day) {
		if e.AllDay {
			allDay = append(allDay, e)
		} else {
			timed = append(timed, e)
		}
	}
	return timed, allDay
}

// eventHeight is how tall a chip is: its own length, clipped to the hours the
// column shows, because a chip running past the last hour would draw over
// whatever sits below the column.
func eventHeight(e Event, startHour, endHour int, hourHeight float32) float32 {
	s, en := e.Span()
	top := HourY(s, startHour, hourHeight)
	h := HourY(en, startHour, hourHeight) - top
	if bottom := float32(endHour-startHour) * hourHeight; top+h > bottom {
		h = bottom - top
	}
	return max(h, 0)
}

// hourWindow returns the hours to draw, defaulting to a working day.
//
// Eight to eighteen rather than zero to twenty-four, because a column of
// twenty-four hours is mostly empty and the hours a person looks at are the
// ones that should be on screen without scrolling.
//
// EndHour is what says whether a window was asked for at all, and zero is the
// way of saying it was not. Zero is not a usable end in its own right: a
// window that runs to midnight is written as twenty-four. StartHour may be
// zero, which is midnight, as long as the end says so — a caller who wants
// the small hours gets them.
func hourWindow(start, end int) (int, int) {
	if end == 0 {
		return 8, 18
	}
	if start < 0 || start > 23 {
		start = 0
	}
	if end > 24 || end <= start {
		end = min(24, start+10)
	}
	return start, end
}
