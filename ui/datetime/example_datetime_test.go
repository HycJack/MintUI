package datetime_test

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/datetime"
	"github.com/HycJack/MintUI/ui/theme"
)

// A booking page is the shape most of this package is drawn for: a month to
// pick a day in, a list of what is already on it, and the people who are
// coming.
func Example() {
	// A window that keeps its own clock. Nothing in this package reads one,
	// so a caller that wants "now" says what now is — and a test says what it
	// is too, which is what makes these components testable at all.
	now := time.Date(2026, 10, 7, 14, 30, 0, 0, time.UTC)
	day := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		k, u := core.Tokens(c), core.Density(c).Unit()

		ui.Column(c).FillWidth().Gap(u * 4).Children(func() {
			ui.Text(c, "Wednesday 7 October").TextColor(k.Text).
				FontSize(theme.TitleSize).Bold()

			datetime.Calendar(c, datetime.CalendarOptions{
				Month:    &day,
				Today:    now,
				Header:   true,
				Weekdays: datetime.DefaultWeekdays(),
			})

			datetime.AgendaView(c, datetime.AgendaOptions{
				Today: now,
				Days: []datetime.AgendaDay{{
					Day: day,
					Events: []datetime.Event{
						{
							Title:    "Maple Street Bakery",
							Start:    day.Add(9 * time.Hour),
							End:      day.Add(10 * time.Hour),
							Location: "12 Maple Street",
							Severity: core.Accent,
						},
						{
							Title: "Riverside Clinic",
							Start: day.Add(11 * time.Hour),
							End:   day.Add(12 * time.Hour),
						},
					},
				}},
			})

			datetime.AttendeeList(c, datetime.AttendeeOptions{
				Avatars:      true,
				ShowResponse: true,
				Attendees: []datetime.Attendee{
					{Name: "Dana Reyes", Response: datetime.Accepted, Timezone: "Europe/Berlin"},
					{Name: "Nate Coleman", Response: datetime.NoReply, Note: "chased on Tuesday"},
					{Name: "Ivy Ahmed", Response: datetime.Maybe},
				},
			})
		})
	}, 520, 900)

	// Output:
}

// A day on the clock rather than in a list: the shape a technician reads
// while deciding whether a job fits before four.
func ExampleCalendarDayView() {
	day := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	now := day.Add(14*time.Hour + 30*time.Minute)

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		view := datetime.CalendarDayView(c, datetime.CalendarDayViewOptions{
			Day:      day,
			Now:      now,
			Pickable: true,
			AllDay:   true,
			Events: []datetime.Event{
				{
					Title: "Maple Street Bakery",
					Start: day.Add(9 * time.Hour),
					End:   day.Add(10 * time.Hour),
				},
				{
					// Two jobs at once, which the column draws side by side
					// rather than one on top of the other.
					Title: "Oak Lane Cafe",
					Start: day.Add(9*time.Hour + 30*time.Minute),
					End:   day.Add(10*time.Hour + 30*time.Minute),
				},
				{Title: "Riverside Clinic", Start: day.Add(15 * time.Hour)},
			},
		})
		// Asking an hour and opening a job are both ordinary answers, and a
		// view asks for them after the component is built.
		if at, ok := view.Picked(); ok {
			_ = at
		}
		if e, ok := view.Chosen(); ok {
			_ = e.Title
		}
	}, 420, 760)

	// Output:
}

// The pickers are the caller's variables. A form holds the day and the
// minutes; the calendar and the steppers write them.
func ExampleDateTimePicker() {
	booking := time.Date(2026, 10, 7, 14, 30, 0, 0, time.UTC)
	open := false

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.DateTimePicker(c, datetime.DateTimePickerOptions{
			Value: &booking,
			Open:  &open,
			Today: booking,
			Name:  "Appointment",
		})
	}, 520, 240)

	// Output:
}

// A rule that can be read back. A recurring booking nobody can check is a
// recurring booking that turns up on the wrong Tuesday.
func ExampleRecurrenceSummary() {
	rule := datetime.Recurrence{
		Frequency: datetime.Weekly,
		Interval:  2,
		Weekdays:  [7]bool{true, false, false, false, true, false, false},
		Count:     10,
	}
	fmt.Println(datetime.RecurrenceSummary(rule))
	fmt.Println(datetime.RecurrenceSummary(datetime.Recurrence{Frequency: datetime.Monthly}))
	fmt.Println(datetime.CronSummary("30 9 * * 1-5"))

	// Output:
	// Every 2 weeks on Mon, Fri for 10 times
	// Every month
	// On Monday, Tuesday, Wednesday, Thursday, Friday at 09:30
}

// A difference is a measurement; the sentence is the window's. That is the
// rule this package follows everywhere: nothing here calls time.Now, and
// nothing here decides what "3 days ago" is called.
func ExampleRelative() {
	now := time.Date(2026, 10, 7, 14, 30, 0, 0, time.UTC)
	then := now.AddDate(0, 0, -3)

	fmt.Println(datetime.RelativeText(datetime.Relative(then, now)))
	fmt.Println(datetime.RelativeText(datetime.Relative(now.Add(2*time.Hour), now)))
	fmt.Println(datetime.DurationText(90 * time.Minute))

	// Output:
	// 3 days ago
	// in 2 hours
	// 1h 30m
}

// The four states of a day, and the week arithmetic under them. Everything
// here is fixed: 7 October 2026 is a Wednesday in week 41, and the next
// February is twenty-eight days long.
func ExampleGrid() {
	now := time.Date(2026, 10, 7, 14, 30, 0, 0, time.UTC)
	chosen := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.Grid(c, datetime.GridOptions{
			Month:           now,
			Today:           now,
			Selected:        &chosen,
			Start:           time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC),
			End:             time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
			ShowWeekNumbers: true,
			Busy:            func(d time.Time) bool { return d.Day() == 20 },
		})
	}, 340, 280)

	fmt.Println(datetime.Weekday(now), datetime.WeekOfYear(now),
		datetime.DaysInMonth(now), datetime.IsLeapYear(2028))

	// Output:
	// 2 41 31 true
}
