package datetime

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── a day ────────────────────────────────────────────────────────────────────

// CalendarDayViewOptions configure a CalendarDayView.
type CalendarDayViewOptions struct {
	// Day is the day the column shows. Required: a day view with no day is a
	// column of hours belonging to nothing.
	Day time.Time
	// Events are the day's events, in any order.
	Events []Event
	// Now draws the current-time line. The zero time draws none — nothing
	// here reads the clock to find out what time it is.
	Now time.Time
	// StartHour, EndHour and HourHeight are the visible scale; zero means a
	// working day at HourHeight for the density.
	StartHour, EndHour int
	HourHeight         float32
	// AllDay draws the all-day events in a row above the clock.
	AllDay bool
	// Pickable makes the hours report a press, for a day view that books.
	Pickable bool
	// Busy greys an hour out and takes its clicks away: the clearest thing
	// there is to show someone asking when they can come in.
	Busy func(time.Time) bool
	// Selected is the event the caller is looking at, or nil.
	Selected *Event
	// Format writes the clock on a chip.
	Format func(time.Time) string
	// Months names the months, for the heading; zero means the default.
	Months MonthNames
}

// CalendarDayViewResult carries a CalendarDayView and what the user did with it.
type CalendarDayViewResult struct {
	// Element is the whole view, heading included.
	Element *ui.Element
	// events is the day's own result, so that Chosen and Empty can be asked
	// of the day view without it re-exposing the column.
	events CalendarEventsResult
	// grid is the hours underneath it.
	grid TimeGridResult
	// stepped is -1 or 1 when the day's arrows were pressed, 0 for neither.
	stepped int
}

// Picked returns the hour the user pressed, and whether they pressed one.
func (r CalendarDayViewResult) Picked() (time.Time, bool) { return r.grid.Picked() }

// Chosen returns the event the user pressed, and whether they pressed one.
func (r CalendarDayViewResult) Chosen() (*Event, bool) { return r.events.Chosen() }

// Stepped reports which day arrow was pressed: -1 for the day before, 1 for
// the day after, 0 for neither.
func (r CalendarDayViewResult) Stepped() int { return r.stepped }

// Empty reports whether the day has nothing on it, which is what a caller
// uses to decide between the day and an explanation of its own.
func (r CalendarDayViewResult) Empty() bool { return r.events.Empty() }

// CalendarDayView is one day: its name, the hours down the side, and the
// day's events at the heights their clocks give them.
//
// It is the view that answers a question the other two cannot — what does
// Tuesday look like — so it is what a callback's detail page is built from.
func CalendarDayView(c *ui.Context, opts CalendarDayViewOptions) CalendarDayViewResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Day.IsZero() {
		panic("datetime: CalendarDayView needs a Day")
	}
	day := StartOfDay(opts.Day)
	start, end := hourWindow(opts.StartHour, opts.EndHour)
	hourHeight := opts.HourHeight
	if hourHeight <= 0 {
		hourHeight = HourHeight(u)
	}
	prev := core.Msg(c, "datetime.prevDay", "Previous day")
	next := core.Msg(c, "datetime.nextDay", "Next day")
	months := opts.Months.or()

	var r CalendarDayViewResult
	head := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u)
	r.Element = ui.Column(c).FillWidth().FillHeight().Gap(u).Children(func() {
		head.Children(func() {
			if stepButton(c, iconPrev, prev, core.ControlHeight(c)).Clicked() {
				r.stepped = -1
			}
			ui.Column(c).Grow(1).Children(func() {
				ui.Text(c, WeekdayFullName(day)).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
				ui.Text(c, itoa(day.Day())+" "+months[monthIndex(day)]+" "+
					itoa(day.Year())).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			})
			if stepButton(c, iconNext, next, core.ControlHeight(c)).Clicked() {
				r.stepped = 1
			}
		}).Label(WeekdayFullName(day) + " " + FormatDate(day) + " " +
			months[monthIndex(day)])

		ui.Row(c).FillWidth().FillHeight().Children(func() {
			hourRuler(c, start, end, hourHeight, true)
			dayColumn(c, CalendarEventsOptions{
				Events:     opts.Events,
				Day:        day,
				StartHour:  start,
				EndHour:    end,
				HourHeight: hourHeight,
				AllDay:     opts.AllDay,
				Format:     opts.Format,
				Selected:   opts.Selected,
			}, start, end, hourHeight, &r.events, func() {
				g := CalendarTimeGrid(c, TimeGridOptions{
					Day:        day,
					StartHour:  start,
					EndHour:    end,
					HourHeight: hourHeight,
					Pickable:   opts.Pickable,
					Now:        opts.Now,
					Disabled:   opts.Busy,
				})
				r.grid = g
			})
		})
	})
	return r
}

// hourRuler is the column of hour numbers down the side of a time view.
//
// It is a function rather than a component because the week view needs one
// ruler for seven columns: a ruler drawn seven times is seven sets of numbers
// to read past, and the only one that matters is the leftmost.
func hourRuler(c *ui.Context, start, end int, hourHeight float32, head bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	col := ui.Column(c).Width(u * 9).Shrink(0).FillHeight()
	col.Children(func() {
		if head {
			// A spacer the height of the day headings, so the numbers start
			// on the first hour's line rather than under the names.
			ui.Box(c).Width(1).Height(u * 9)
		}
		for hour := start; hour < end; hour++ {
			ui.Text(c, HourLabel(hour)).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Height(hourHeight).
				TextAlign(ui.End).SingleLine()
		}
	})
}

// ── a week ───────────────────────────────────────────────────────────────────

// WeekViewTitle names the week t falls in: its first and last day, and its
// number.
//
// The number is in it because a week is the unit a booking is made in, and a
// person reading "W41" next to a date has to be able to check it against the
// calendar they are looking at. It is the ISO number, the one a booking system
// exports, rather than a rule invented here that would disagree with it.
func WeekViewTitle(t time.Time, months MonthNames) string {
	names := months.or()
	start := StartOfWeek(t)
	end := start.AddDate(0, 0, 6)
	_, week := t.ISOWeek()
	number := " · W" + itoa(week)
	switch {
	case SameMonth(start, end) && start.Year() == end.Year():
		return itoa(start.Day()) + "–" + itoa(end.Day()) + " " +
			names[monthIndex(end)] + " " + itoa(end.Year()) + number
	case start.Year() == end.Year():
		return itoa(start.Day()) + " " + names[monthIndex(start)] + " – " +
			itoa(end.Day()) + " " + names[monthIndex(end)] + " " + itoa(end.Year()) + number
	default:
		// A week that runs over New Year is named by its two dates rather
		// than by a month and a day, because "28–3 January" is a lie.
		return FormatDate(start) + " – " + FormatDate(end) + number
	}
}

// CalendarWeekViewOptions configure a CalendarWeekView.
type CalendarWeekViewOptions struct {
	// Day is any day in the week to show: the Monday-to-Sunday week it falls
	// in is the one drawn.
	Day time.Time
	// Events are the week's events, in any order; each day is sorted into its
	// own column.
	Events []Event
	// Now draws the current-time line on today's column.
	Now time.Time
	// StartHour, EndHour and HourHeight are the visible scale.
	StartHour, EndHour int
	HourHeight         float32
	// AllDay draws the week's all-day events in a row above the columns.
	AllDay bool
	// Pickable makes the hours report a press, for a week that books.
	Pickable bool
	// Title draws the week's name above the columns.
	Title bool
	// Selected is the event the caller is looking at, or nil.
	Selected *Event
	// Format writes the clock on a chip.
	Format func(time.Time) string
	// Weekdays and Months name the days; zero means the default.
	Weekdays WeekdayNames
	Months   MonthNames
}

// CalendarWeekViewResult carries a CalendarWeekView and what the user did with it.
type CalendarWeekViewResult struct {
	// Element is the whole view, title included.
	Element *ui.Element
	// chosen is the event pressed this frame, and got whether there was one.
	chosen *Event
	got    bool
	// picked is the time pressed this frame, and pick whether there was one.
	picked time.Time
	pick   bool
}

// Chosen returns the event the user pressed, and whether they pressed one.
func (r CalendarWeekViewResult) Chosen() (*Event, bool) { return r.chosen, r.got }

// Picked returns the time the user pressed, and whether they pressed one.
func (r CalendarWeekViewResult) Picked() (time.Time, bool) { return r.picked, r.pick }

// CalendarWeekView is seven days side by side: the hours down the left, one
// column each, and the week's events at the heights their clocks give them.
//
// The hour numbers are drawn once for the whole week rather than seven times.
// That is the whole difference between a week that reads as seven days and a
// week that reads as seven narrow calendars: one ruler, seven columns.
func CalendarWeekView(c *ui.Context, opts CalendarWeekViewOptions) CalendarWeekViewResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Day.IsZero() {
		panic("datetime: CalendarWeekView needs a Day")
	}
	start, end := hourWindow(opts.StartHour, opts.EndHour)
	hourHeight := opts.HourHeight
	if hourHeight <= 0 {
		hourHeight = HourHeight(u)
	}
	weekdays := opts.Weekdays.or()
	months := opts.Months.or()
	cells := WeekCells(opts.Day)
	today := StartOfDay(opts.Now)
	headHeight := u * 9
	// One result per column, in the order the columns are built, so a press
	// in any of the seven can be reported without the seven sharing a
	// struct and overwriting each other.
	cols := make([]CalendarEventsResult, 0, len(cells))

	var r CalendarWeekViewResult
	r.Element = ui.Column(c).FillWidth().FillHeight().Gap(u * 1.5).Children(func() {
		if opts.Title {
			ui.Text(c, WeekViewTitle(opts.Day, months)).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
		}
		if opts.AllDay {
			// The all-day row spans the whole week: an all-day event belongs
			// to the week, not to a column of hours.
			ui.Column(c).FillWidth().Gap(u / 2).Children(func() {
				for _, e := range EventsOn(opts.Events, opts.Day) {
					if !e.AllDay {
						continue
					}
					e := e
					row := chip(c, e, theme.CaptionSize, false, false)
					if row.Clicked() {
						r.chosen, r.got = &e, true
					}
				}
			})
		}
		ui.Row(c).FillWidth().FillHeight().Children(func() {
			hourRuler(c, start, end, hourHeight, true)
			ui.Row(c).FillWidth().FillHeight().Gap(u / 2).Children(func() {
				for _, cell := range cells {
					day := cell.Day
					isToday := !today.IsZero() && SameDay(day, today)
					fg, chipBG := k.Text, k.Surface
					if isToday {
						fg, chipBG = k.AccentText, k.AccentBg
					}
					slot := len(cols)
					cols = append(cols, CalendarEventsResult{})
					ui.Column(c).FillWidth().FillHeight().Gap(u / 2).Children(func() {
						// One heading at one fixed height, so that all seven
						// columns' hours begin on the same line.
						ui.Box(c).FillWidth().Height(headHeight).
							Radius(theme.SmallRadius).Background(chipBG).Center().
							Label(dayNameOf(day, months)).
							Children(func() {
								ui.Text(c, weekdays[Weekday(day)]).TextColor(k.TextMuted).
									FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
								ui.Text(c, itoa(day.Day())).TextColor(fg).
									FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
							})
						// Only today's column carries the current-time
						// line: a line across all seven says the time seven
						// times and means it once.
						line := time.Time{}
						if isToday {
							line = opts.Now
						}
						dayColumn(c, CalendarEventsOptions{
							Events:     opts.Events,
							Day:        day,
							StartHour:  start,
							EndHour:    end,
							HourHeight: hourHeight,
							Format:     opts.Format,
							Selected:   opts.Selected,
						}, start, end, hourHeight, &cols[slot], func() {
							g := CalendarTimeGrid(c, TimeGridOptions{
								Day:        day,
								StartHour:  start,
								EndHour:    end,
								HourHeight: hourHeight,
								Pickable:   opts.Pickable,
								Now:        line,
							})
							if t, ok := g.Picked(); ok {
								r.picked, r.pick = t, true
							}
						})
					})
				}
			})
		})
	})
	for i := range cols {
		if e, ok := cols[i].Chosen(); ok {
			r.chosen, r.got = e, true
		}
	}
	return r
}

// ── a month ──────────────────────────────────────────────────────────────────

// CalendarMonthViewOptions configure a CalendarMonthView.
type CalendarMonthViewOptions struct {
	// Month is any day in the month to draw.
	Month time.Time
	// Today is the caller's today; the zero time draws no today.
	Today time.Time
	// Selected is the chosen day, or nil.
	Selected *time.Time
	// Start and End shade a range between them, when both are set.
	Start, End time.Time
	// Weeks is how many rows to draw; zero means DefaultWeeks.
	Weeks int
	// Busy marks a day as having something on it.
	Busy func(time.Time) bool
	// Disabled greys a day out.
	Disabled func(time.Time) bool
	// Caption draws the month's name and year above the days.
	Caption bool
	// Weekdays and Months name the days and months; zero means the default.
	Weekdays WeekdayNames
	Months   MonthNames
}

// CalendarMonthViewResult carries a CalendarMonthView and the day the user pressed.
type CalendarMonthViewResult struct {
	// Element is the month and its caption.
	Element *ui.Element
	// picked is the day pressed this frame, and got whether there was one.
	picked time.Time
	got    bool
}

// Picked returns the day the user pressed, and whether they pressed one.
func (r CalendarMonthViewResult) Picked() (time.Time, bool) { return r.picked, r.got }

// CalendarMonthView is one month of days with its name above them and nothing
// else: no arrows, no heading, no jump to today.
//
// It is the month without the furniture, for a page that has a heading and a
// way of moving of its own — a dashboard, a booking page with its own month
// switcher. It is also the smallest thing a month can be, which is what a year
// view's twelve months are made of.
func CalendarMonthView(c *ui.Context, opts CalendarMonthViewOptions) CalendarMonthViewResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	caption := monthCaption(opts.Month, opts.Months.or())
	var r CalendarMonthViewResult
	r.Element = ui.Column(c).FillWidth().Gap(u).Label(caption).Children(func() {
		if opts.Caption {
			ui.Text(c, caption).TextColor(k.Text).FontSize(core.FontSize(c, theme.RowSize)).
				Bold().SingleLine()
		}
		g := Grid(c, GridOptions{
			Month:    opts.Month,
			Selected: opts.Selected,
			Today:    opts.Today,
			Start:    opts.Start,
			End:      opts.End,
			Weeks:    opts.Weeks,
			Weekdays: opts.Weekdays,
			Months:   opts.Months,
			Busy:     opts.Busy,
			Disabled: opts.Disabled,
		})
		if d, ok := g.Picked(); ok {
			r.picked, r.got = d, true
		}
	})
	return r
}

func monthCaption(month time.Time, months MonthNames) string {
	return months[monthIndex(month)] + " " + itoa(month.Year())
}

// ── a year ───────────────────────────────────────────────────────────────────

// CalendarYearViewOptions configure a CalendarYearView.
type CalendarYearViewOptions struct {
	// Year is any day in the year to draw; the arrows move it. It is a
	// pointer because paging is a change to it.
	Year *time.Time
	// Today is the caller's today, so the day of the year can be found where
	// it belongs. The zero time draws no today.
	Today time.Time
	// Selected is the chosen day, or nil.
	Selected *time.Time
	// Start and End shade a range between them, when both are set.
	Start, End time.Time
	// Columns is how many months across; zero means three, which is twelve
	// months in four rows of three.
	Columns int
	// Highlighted is the month the caller has open; it is ringed.
	Highlighted *time.Time
	// Busy marks a day as having something on it, in the small months.
	Busy func(time.Time) bool
	// Weekdays and Months name the days and months; zero means the default.
	Weekdays WeekdayNames
	Months   MonthNames
}

// CalendarYearViewResult carries a CalendarYearView and what the user did with it.
type CalendarYearViewResult struct {
	// Element is the whole year.
	Element *ui.Element
	// picked is the day pressed this frame, and got whether there was one.
	picked time.Time
	got    bool
	// stepped is -1 or 1 for the arrows, 0 for neither.
	stepped int
}

// Picked returns the day the user pressed, and whether they pressed one.
func (r CalendarYearViewResult) Picked() (time.Time, bool) { return r.picked, r.got }

// Stepped reports which year arrow was pressed: -1 for the year before, 1 for
// the year after, 0 for neither.
func (r CalendarYearViewResult) Stepped() int { return r.stepped }

// CalendarYearView is twelve months of days, small, with the year between two
// arrows.
//
// A year is the largest span a grid can usefully show: a decade is a list of
// years rather than a picture of one, and a year of days is a picture a
// person can read. The months are drawn with the same grid as a month view, so
// a day that reads wrongly here reads wrongly in both.
func CalendarYearView(c *ui.Context, opts CalendarYearViewOptions) CalendarYearViewResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Year == nil {
		panic("datetime: CalendarYearView needs a Year to point at")
	}
	year := opts.Year.Year()
	columns := opts.Columns
	if columns <= 0 || columns > 6 {
		columns = 3
	}
	prev := core.Msg(c, "datetime.prevYear", "Previous year")
	next := core.Msg(c, "datetime.nextYear", "Next year")
	loc := opts.Year.Location()

	var r CalendarYearViewResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 2).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Label(itoa(year)).
			Children(func() {
				if stepButton(c, iconPrev, prev, core.ControlHeight(c)).Clicked() {
					r.stepped = -1
				}
				ui.Text(c, itoa(year)).TextColor(k.Text).FontSize(core.FontSize(c, theme.TitleSize)).
					Bold().Grow(1).TextAlign(ui.Center).SingleLine()
				if stepButton(c, iconNext, next, core.ControlHeight(c)).Clicked() {
					r.stepped = 1
				}
			})
		ui.Column(c).FillWidth().Gap(u * 2).Children(func() {
			for row := 0; row < 12; row += columns {
				ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
					for col := range columns {
						month := time.Date(year, time.Month(row+col+1), 1, 0, 0, 0, 0, loc)
						open := opts.Highlighted != nil && SameMonth(*opts.Highlighted, month)
						m := CalendarMonthView(c, CalendarMonthViewOptions{
							Month:    month,
							Today:    opts.Today,
							Selected: opts.Selected,
							Start:    opts.Start,
							End:      opts.End,
							Caption:  true,
							Weeks:    6,
							Weekdays: opts.Weekdays,
							Months:   opts.Months,
							Busy:     opts.Busy,
						})
						if d, ok := m.Picked(); ok {
							r.picked, r.got = d, true
						}
						// The open month is ringed rather than filled: a
						// filled month reads as selected, and the selected
						// day is already filled inside it.
						if open {
							m.Element.Padding(u*1.5).Radius(theme.ControlRadius).
								Border(theme.BorderWidth*2, k.Accent)
						}
					}
				})
			}
		})
	})
	// After the frame, so that the heading and the months always agree.
	if r.stepped != 0 {
		*opts.Year = time.Date(year+r.stepped, opts.Year.Month(), opts.Year.Day(),
			opts.Year.Hour(), opts.Year.Minute(), 0, 0, loc)
	}
	return r
}
