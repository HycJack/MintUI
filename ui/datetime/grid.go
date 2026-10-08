package datetime

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// DefaultWeeks is how many week rows a month grid draws when the caller does
// not say.
//
// Six, not "as many as this month needs": a month that needs five rows would
// otherwise be two rows shorter than the month before it, and a month view
// that changes height as you page through it moves every click target under
// the pointer. Six rows is the largest any month can need, so the box never
// changes size. Callers who would rather have the shorter box — a picker in a
// popover, where the height matters more than the stability — pass Weeks: 5 or
// a number of their own.
const DefaultWeeks = 6

// WeekdayNames are the seven column headings, Monday first, matching Weekday.
//
// They are the caller's rather than the library's because they are words: a
// window showing a German calendar passes "Mo, Di, …" and one showing a
// Chinese one passes "一, 二, …". The zero value — seven empty strings — means
// the defaults, so a caller who has no opinion says nothing.
type WeekdayNames [7]string

// DefaultWeekdays are the English abbreviations, as a calendar prints them.
func DefaultWeekdays() WeekdayNames {
	return WeekdayNames{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
}

// Valid reports whether every name is filled in.
func (w WeekdayNames) Valid() bool {
	for _, s := range w {
		if s == "" {
			return false
		}
	}
	return true
}

// or returns w, or the default when w was left zero.
func (w WeekdayNames) or() WeekdayNames {
	if w.Valid() {
		return w
	}
	return DefaultWeekdays()
}

// Shorten clips a name to n characters, for a column too narrow for the whole
// of it. It clips rather than wrapping: a wrapped "Wed" breaks the row and the
// grid stops being square. It counts runes, not bytes, so a name written in a
// script whose letters are more than one byte is never cut in half.
func (w WeekdayNames) Shorten(n int) WeekdayNames {
	var out WeekdayNames
	for i, s := range w {
		if r := []rune(s); len(r) > n {
			s = string(r[:n])
		}
		out[i] = s
	}
	return out
}

// MonthNames are the twelve month names, January first.
type MonthNames [12]string

// DefaultMonths are the English month names.
func DefaultMonths() MonthNames {
	return MonthNames{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
}

func (m MonthNames) or() MonthNames {
	for _, s := range m {
		if s == "" {
			return DefaultMonths()
		}
	}
	return m
}

// monthIndex is a month as an index into MonthNames, January as 0.
func monthIndex(t time.Time) int { return int(t.Month()) - 1 }

// WeekdayFullNames are the seven weekday names in full, Monday first. They
// are here for the places that need a name without a date to take it from —
// the toggle of a weekly recurrence, the label of a column heading.
var WeekdayFullNames = [7]string{
	"Monday", "Tuesday", "Wednesday", "Thursday",
	"Friday", "Saturday", "Sunday",
}

// WeekdayFullName is the uncut name of t's weekday, used to name a day cell
// for a screen reader. A grid cell is a digit in a circle: without a name it
// is announced as "7", which is not a date to anyone.
func WeekdayFullName(t time.Time) string { return WeekdayFullNames[Weekday(t)] }

// Cell is one square of a month grid: a day, where it sits, and whether it is
// the month's own day or a neighbour borrowed from the month before or after.
type Cell struct {
	// Day is the date in the cell, at midnight in Month's location.
	Day time.Time
	// InMonth is false for the days borrowed from the neighbouring months to
	// make the first and last rows whole.
	InMonth bool
	// Row and Column place the cell: seven columns, Monday first.
	Row, Column int
	// Week is the ISO week number of Day, for a grid that shows week numbers.
	Week int
}

// MonthCells lays a month out as whole weeks starting on Monday, which is what
// every month grid in the library draws from.
//
// A grid is weeks, not days: the first row starts on the Monday on or before
// the first of the month and the last ends on the Sunday on or after the last,
// so every cell has a column and the rows are all seven wide. Leading and
// trailing days are Cell.InMonth false — the days a caller draws quieter
// because they belong to the month you are not looking at.
//
// weeks below zero, or zero, means DefaultWeeks; a caller that wants a fixed
// height across months gets one by not counting rows off the month itself.
func MonthCells(month time.Time, weeks int) []Cell {
	if weeks <= 0 {
		weeks = DefaultWeeks
	}
	first := StartOfMonth(month)
	// How many weeks the month itself spans, so a six-week request is not
	// filled with empty rows for a February that fits in four.
	need := (Weekday(first) + DaysInMonth(first) + 6) / 7
	weeks = max(weeks, need)

	start := StartOfWeek(first)
	cells := make([]Cell, 0, weeks*7)
	for i := range weeks * 7 {
		day := start.AddDate(0, 0, i)
		_, week := day.ISOWeek()
		cells = append(cells, Cell{
			Day:     day,
			InMonth: SameMonth(day, first),
			Row:     i / 7,
			Column:  i % 7,
			Week:    week,
		})
	}
	return cells
}

// WeekCells lays out the seven days of the week t falls in.
func WeekCells(t time.Time) []Cell {
	start := StartOfWeek(t)
	cells := make([]Cell, 0, 7)
	for i := range 7 {
		day := start.AddDate(0, 0, i)
		_, week := day.ISOWeek()
		cells = append(cells, Cell{Day: day, InMonth: true, Row: 0, Column: i, Week: week})
	}
	return cells
}

// GridOptions configure a Grid.
type GridOptions struct {
	// Month is any day in the month to draw; only its year and month are read.
	Month time.Time
	// Selected is the chosen day, drawn filled. Nil draws no selection, which
	// is what a month view showing every month of a year wants.
	Selected *time.Time
	// Today is the caller's today, drawn as a ring. The zero time draws no
	// today at all — a component here never reads the clock to find out.
	Today time.Time
	// Start and End, when both are set, shade the days between them
	// inclusive. One without the other draws no range.
	Start, End time.Time
	// Min and Max bound the days that can be chosen. A zero bound is open.
	Min, Max time.Time
	// Weeks is how many rows to draw; zero means DefaultWeeks.
	Weeks int
	// Weekdays names the columns. Zero means the default.
	Weekdays WeekdayNames
	// Months names the months, for naming cells and for the day names read
	// aloud. Zero means the default.
	Months MonthNames
	// ShowWeekNumbers adds the ISO week number down the left.
	ShowWeekNumbers bool
	// WeekHeader is a label for the column of week numbers.
	WeekHeader string
	// Busy marks a day as having something on it with a dot under its number —
	// the cheap way to show which days a month view should open.
	Busy func(time.Time) bool
	// Disabled greys a day out and takes its clicks away.
	Disabled func(time.Time) bool
}

// GridResult carries a Grid and the day the user pressed in it.
type GridResult struct {
	// Element is the whole grid, headings included.
	Element *ui.Element
	// picked is the day pressed this frame, and got whether there was one.
	picked time.Time
	got    bool
}

// Picked returns the day the user pressed, and whether they pressed one.
func (r GridResult) Picked() (time.Time, bool) { return r.picked, r.got }

// Grid draws one month as seven columns of day cells, the base every calendar
// in the library stands on: the month view, the date picker, the range picker
// and the year view's twelve small months are all this with a different
// heading.
//
// It writes nothing. A press comes back through Picked, and it is up to the
// caller to put the day in the variable the grid is drawing as selected —
// which is what lets one grid serve a read-only month view and a form field
// without either of them knowing about the other.
//
// The four states a reader has to tell apart are drawn apart on purpose:
//
//	selected    filled with ink, the number in the ink's own colour
//	today       a ring in the accent colour and an accent-coloured number
//	in range    the accent's pale background, the same colour as today
//	out of month  faint, on nothing
//
// Today and a range share a colour because they are related — both are "the
// accent is talking about this day" — and they are told apart by the ring and
// by the fill. Disabled days are the fourth state and the one that cannot be
// told apart by colour at all: they are filled with the surface colour and
// drawn faint, and they take no clicks, so what a person notices is that the
// day does not answer.
func Grid(c *ui.Context, opts GridOptions) GridResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	months := opts.Months.or()
	side := cellSide(u)
	cells := MonthCells(opts.Month, opts.Weeks)
	rows := cells[len(cells)-1].Row + 1

	var r GridResult
	gap := u / 2

	grid := ui.Column(c).FillWidth().Gap(gap).Children(func() {
		gridHead(c, opts, side, gap)
		for row := range rows {
			line := ui.Row(c).FillWidth().Gap(gap).AlignItems(ui.Center)
			if opts.ShowWeekNumbers {
				_, week := cells[row*7].Day.ISOWeek()
				// The number goes in the margin: it belongs to the row, not to
				// any one of its seven days, so a column of its own is the
				// only place it can go without moving with the days.
				line.Children(func() {
					ui.Text(c, itoa(week)).TextColor(k.TextFaint).
						FontSize(theme.CaptionSize).SingleLine().
						Width(side).TextAlign(ui.Center)
				})
			}
			for column := range 7 {
				cell := cells[row*7+column]
				line.Children(func() {
					if day, ok := dayCell(c, opts, cell, side, months); ok {
						r.picked, r.got = day, true
					}
				})
			}
		}
	})
	r.Element = grid
	return r
}

// dayNameOf spells a date out for a screen reader.
func dayNameOf(t time.Time, months MonthNames) string {
	return WeekdayFullName(t) + " " + itoa(t.Day()) + " " + months[monthIndex(t)] + " " + itoa(t.Year())
}

// gridHead is the row of column headings above a grid, and the first column
// when the grid carries week numbers. It sits in its own row rather than being
// the grid's first item so that a caller who wraps a grid in a scroll area gets
// the headings scrolled with it — which is the point, since a heading that
// stays while the days move under it names the wrong days.
func gridHead(c *ui.Context, opts GridOptions, side, gap float32) {
	k := core.Tokens(c)
	weekdays := opts.Weekdays.or()
	head := ui.Row(c).FillWidth().Gap(gap).AlignItems(ui.Center).
		Label(core.Msg(c, "datetime.weekdays", "Days of the week"))
	head.Children(func() {
		if opts.ShowWeekNumbers {
			name := opts.WeekHeader
			if name == "" {
				name = core.Msg(c, "datetime.week", "Wk")
			}
			ui.Text(c, name).TextColor(k.TextFaint).FontSize(theme.CaptionSize).
				Width(side).TextAlign(ui.Center).SingleLine()
		}
		// The heading is clipped to three characters to fit the narrowest
		// column, but named in full: a screen reader should hear "Wednesday"
		// rather than the "Wed" the eye has to make do with.
		short := weekdays.Shorten(3)
		for i, name := range short {
			full := weekdays[i]
			ui.Text(c, name).TextColor(k.TextMuted).FontSize(theme.CaptionSize).
				Bold().Grow(1).TextAlign(ui.Center).SingleLine().Label(full)
		}
	})
}

// dayCell draws one day and reports a press of it.
//
// It returns the day and whether it was pressed, rather than writing the
// selection: the grid is a view of a selection, and letting it move the
// selection itself would mean two copies of the same fact.
func dayCell(c *ui.Context, opts GridOptions, cell Cell, side float32, months MonthNames) (time.Time, bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	day := cell.Day
	busy := opts.Busy != nil && opts.Busy(day)
	off := !cell.InMonth
	disabled := (opts.Disabled != nil && opts.Disabled(day)) || (!InRange(day, opts.Min, opts.Max))
	selected := opts.Selected != nil && SameDay(*opts.Selected, day)
	today := !opts.Today.IsZero() && SameDay(opts.Today, day)
	inRange := !opts.Start.IsZero() && !opts.End.IsZero() &&
		!day.Before(StartOfDay(opts.Start)) && !day.After(StartOfDay(opts.End))

	// The four states, and the one that is none of them.
	bg, fg, dot := ui.Color{}, k.Text, k.Text
	switch {
	case selected:
		bg, fg, dot = k.Fill, k.OnFill, k.OnFill
	case disabled:
		bg, fg, dot = k.Surface, k.TextFaint, k.Border
	case inRange:
		bg, fg, dot = k.AccentBg, k.AccentText, k.AccentText
	case off:
		fg, dot = k.TextFaint, k.Border
	case today:
		fg, dot = k.AccentText, k.Accent
	}

	btn := ui.Button(c, "").Grow(1).MinWidth(MinCellSize).Height(side).Shrink(0).
		Radius(side / 2).Background(bg).TextColor(fg).Disabled(disabled).
		Label(dayNameOf(day, months)).Center().Children(func() {
		ui.Text(c, itoa(day.Day())).FontSize(theme.RowSize).Bold().SingleLine().
			FontFeatures("tnum")
		// The dot for a busy day sits under the number rather than replacing
		// it: the number is what the day is, the dot is what is on it. It
		// takes the number's own colour, so a busy day is one colour and the
		// state underneath it still reads through it.
		if busy && !selected {
			ui.Box(c).Width(u*1.25).Height(u*1.25).Radius(u).Margin(0, u*0.5, 0, 0).
				Background(dot)
		}
	})
	// The ring is drawn rather than bordered so it can sit inside the cell and
	// leave the number where it is: a border would push the number a pixel.
	if today && !selected {
		btn.Draw(func(p *ui.Painter, rect ui.Rect) {
			internal.Ring(p, rect.X+rect.W/2, rect.Y+rect.H/2, rect.W/2-u*0.5,
				u*0.5, k.Accent)
		})
	}
	return day, btn.Clicked()
}
