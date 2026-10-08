package datetime

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// CalendarOptions configure a Calendar.
type CalendarOptions struct {
	// Month is the month on screen — any day in it — and the arrows move it.
	// It is a pointer because paging is a change to it, and the caller is the
	// one that has to know the view is now showing October when the form
	// behind it still says September.
	Month *time.Time
	// Header draws the title and the arrows. A Calendar inside another
	// calendar's header turns it off.
	Header bool
	// Today draws a button that jumps Month to Today. It needs Today, and
	// without one there is nothing to jump to.
	TodayButton bool
	// Selected, Today, Start, End, Min, Max, Weeks, Weekdays, Months,
	// ShowWeekNumbers, Busy and Disabled are the Grid's, passed straight
	// through: a calendar is a grid with a heading.
	Selected   *time.Time
	Today      time.Time
	Start, End time.Time
	Min, Max   time.Time
	Weeks      int
	Weekdays   WeekdayNames
	Months     MonthNames
	// ShowWeekNumbers adds the ISO week number down the left.
	ShowWeekNumbers bool
	// Busy marks a day as having something on it.
	Busy func(time.Time) bool
	// Disabled greys a day out and takes its clicks away.
	Disabled func(time.Time) bool
	// Title names the month. It defaults to "October 2026", and it takes a
	// function rather than a string because the title is drawn before the
	// caller knows whether a press has already moved the month this frame.
	Title func(time.Time) string
	// Prev and Next name the arrows for assistive technology.
	Prev, Next string
	// TodayName names the jump-to-today button.
	TodayName string
}

// CalendarResult carries a Calendar and what the user did with it.
type CalendarResult struct {
	// Element is the whole calendar, heading included.
	Element *ui.Element
	// picked is the day pressed this frame, and got whether there was one.
	picked time.Time
	got    bool
	// stepped is -1 for the previous arrow, 1 for the next, 0 for neither.
	stepped int
	// home reports a press of the jump-to-today button.
	home bool
}

// Picked returns the day the user pressed, and whether they pressed one.
func (r CalendarResult) Picked() (time.Time, bool) { return r.picked, r.got }

// Stepped reports which arrow was pressed: -1 for the previous month, 1 for
// the next, 0 for neither.
//
// It is a count rather than a value because Month has already been moved by the
// time the caller sees this — the caller wants to know that it moved, not to
// move it again.
func (r CalendarResult) Stepped() int { return r.stepped }

// Jumped reports a press of the jump-to-today button.
func (r CalendarResult) Jumped() bool { return r.home }

// Calendar is a month of days with a heading: the month name, the arrows that
// page it, and a grid underneath.
//
// The arrows write through Month rather than reporting a step and leaving the
// caller to apply it. A calendar is the only thing that knows what "the
// previous month" means — it has to clamp a day that the month has no room
// for, which is why the field is a pointer to a time and not an integer
// offset the caller keeps somewhere else.
//
// The move lands after the frame is built, so the heading and the grid always
// show the same month: a calendar whose title says October above a grid of
// September days is the bug this ordering exists to prevent.
func Calendar(c *ui.Context, opts CalendarOptions) CalendarResult {
	if opts.Month == nil {
		panic("datetime: Calendar needs a Month to point at")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	month := *opts.Month
	title := opts.Title
	if title == nil {
		title = func(t time.Time) string {
			return opts.Months.or()[monthIndex(t)] + " " + itoa(t.Year())
		}
	}
	prev, next := opts.Prev, opts.Next
	if prev == "" {
		prev = core.Msg(c, "datetime.prevMonth", "Previous month")
	}
	if next == "" {
		next = core.Msg(c, "datetime.nextMonth", "Next month")
	}

	var r CalendarResult
	cal := ui.Column(c).FillWidth().Gap(u * 1.5).Children(func() {
		if opts.Header {
			head := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).
				Label(title(month))
			head.Children(func() {
				if stepButton(c, iconPrev, prev, core.ControlHeight(c)).Clicked() {
					r.stepped = -1
				}
				ui.Text(c, title(month)).TextColor(k.Text).FontSize(theme.RowSize).
					Bold().Grow(1).TextAlign(ui.Center).SingleLine()
				if stepButton(c, iconNext, next, core.ControlHeight(c)).Clicked() {
					r.stepped = 1
				}
			})
			// The jump sits on the left, under the arrows, because a second
			// button in the heading would fight the title for the space the
			// month name needs.
			if opts.TodayButton && !opts.Today.IsZero() {
				name := opts.TodayName
				if name == "" {
					name = core.Msg(c, "datetime.today", "Today")
				}
				btn := ui.Button(c, name).Height(u*7).Radius(theme.PillRadius).
					Padding(0, u*3).Background(k.Surface).TextColor(k.Text).
					Label(name).Tooltip(name)
				btn.Children(func() {
					ui.Text(c, name).FontSize(theme.CaptionSize)
				})
				r.home = btn.Clicked()
			}
		}
		g := Grid(c, GridOptions{
			Month:           month,
			Selected:        opts.Selected,
			Today:           opts.Today,
			Start:           opts.Start,
			End:             opts.End,
			Min:             opts.Min,
			Max:             opts.Max,
			Weeks:           opts.Weeks,
			Weekdays:        opts.Weekdays,
			Months:          opts.Months,
			ShowWeekNumbers: opts.ShowWeekNumbers,
			Busy:            opts.Busy,
			Disabled:        opts.Disabled,
		})
		if d, ok := g.Picked(); ok {
			r.picked, r.got = d, true
		}
	})

	// After the frame, never during it: see the note on Calendar.
	switch {
	case r.stepped != 0:
		*opts.Month = AddMonths(month, r.stepped)
	case r.home:
		*opts.Month = StartOfMonth(opts.Today)
	}
	r.Element = cal
	return r
}
