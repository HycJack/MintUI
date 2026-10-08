package datetime_test

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/datetime"
	"github.com/HycJack/MintUI/ui/theme"
)

// The clock every test in this file is written against. Nothing here calls
// time.Now and nothing here is random: a calendar component that reads a
// different second on every run is a component whose tests only ever pass by
// luck, and a test file that flakes teaches the next person to re-run it.
func fixed() time.Time {
	return time.Date(2026, 10, 7, 14, 30, 0, 0, time.UTC)
}

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// render builds a frame with the library's theme installed, which is what
// every view does first and what a component that reads the theme without it
// panics on.
func render(w, h int, body func(c *ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		body(c)
	}, w, h)
}

// renderDark is render with the dark palette, so that a colour chosen by a
// token rather than by a literal is checked in both appearances.
func renderDark(w, h int, body func(c *ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		body(c)
	}, w, h)
}

// countText counts the elements showing exactly s, which is how a test tells
// "the ruler is drawn once" from "a chip happens to say 08 too".
func countText(tt *ui.Tester, s string) int {
	n := 0
	for _, got := range tt.Texts() {
		if got == s {
			n++
		}
	}
	return n
}

func hasText(t *testing.T, tt *ui.Tester, want ...string) {
	t.Helper()
	for _, w := range want {
		if !tt.HasText(w) {
			t.Errorf("missing %q; the frame shows %q", w, tt.Texts())
		}
	}
}

// colorsIn samples a box on a 8×8 grid and returns the distinct colours it
// finds, so that two pieces of the same frame can be compared by what they
// are painted rather than by what they say.
func colorsIn(img *image.RGBA, r ui.Rect) map[color.RGBA]bool {
	out := map[color.RGBA]bool{}
	if r.W <= 0 || r.H <= 0 {
		return out
	}
	for i := 1; i < 8; i++ {
		for j := 1; j < 8; j++ {
			x := int(r.X + r.W*float32(i)/8)
			y := int(r.Y + r.H*float32(j)/8)
			out[img.RGBAAt(x, y)] = true
		}
	}
	return out
}

func sameColors(a, b map[color.RGBA]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// ── the format layer ─────────────────────────────────────────────────────────

func TestFormatWritesTheSameStringsEverywhere(t *testing.T) {
	now := fixed()
	for _, tc := range []struct{ what, got, want string }{
		{"date", datetime.FormatDate(now), "2026-10-07"},
		{"time", datetime.FormatTime(now), "14:30"},
		{"seconds", datetime.FormatSeconds(now), "14:30:00"},
		{"date and time", datetime.FormatDateTime(now), "2026-10-07 14:30"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.what, tc.got, tc.want)
		}
	}
}

func TestParseReadsWhatFormatWrites(t *testing.T) {
	now := fixed()
	got, err := datetime.ParseDateTime(datetime.FormatDateTime(now))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(now) {
		t.Errorf("round trip: got %v, want %v", got, now)
	}
	if _, err := datetime.ParseDate("7 October"); err == nil {
		t.Error("a date in words should not parse as an ISO date")
	}
	if _, err := datetime.ParseTime("2:30 pm"); err == nil {
		t.Error("a 12-hour clock should not parse as this package's 24-hour one")
	}
}

func TestLeapYearsAndMonthEnds(t *testing.T) {
	// The rule everybody half-remembers, written out: every fourth year,
	// except every hundredth, except every four hundredth.
	for _, tc := range []struct {
		year int
		want bool
	}{{2024, true}, {2026, false}, {2000, true}, {1900, false}, {2100, false}} {
		if got := datetime.IsLeapYear(tc.year); got != tc.want {
			t.Errorf("IsLeapYear(%d) = %v, want %v", tc.year, got, tc.want)
		}
	}
	// February is the only month the rule is about.
	if got := datetime.DaysInMonth(day(2028, time.February, 1)); got != 29 {
		t.Errorf("February 2028 has %d days, want 29", got)
	}
	if got := datetime.DaysInMonth(day(2026, time.February, 1)); got != 28 {
		t.Errorf("February 2026 has %d days, want 28", got)
	}
	if got := datetime.DaysInMonth(day(2026, time.April, 1)); got != 30 {
		t.Errorf("April has %d days, want 30", got)
	}
	// Adding a month clamps to the end of a short one rather than skipping
	// over it: 31 January plus one month is 28 February, not 3 March.
	from := time.Date(2026, time.January, 31, 9, 0, 0, 0, time.UTC)
	if got := datetime.AddMonths(from, 1); got.Day() != 28 || got.Month() != time.February {
		t.Errorf("31 January plus a month is %v, want 28 February", got)
	}
	leap := time.Date(2028, time.January, 31, 9, 0, 0, 0, time.UTC)
	if got := datetime.AddMonths(leap, 1); got.Day() != 29 || got.Month() != time.February {
		t.Errorf("31 January 2028 plus a month is %v, want 29 February", got)
	}
	// And the year rolls rather than sticking in December.
	dec := time.Date(2026, time.December, 15, 0, 0, 0, 0, time.UTC)
	if got := datetime.AddMonths(dec, 1); got.Year() != 2027 || got.Month() != time.January {
		t.Errorf("15 December plus a month is %v, want January 2027", got)
	}
}

func TestDaysBetweenCountsNightsNotHours(t *testing.T) {
	// Across the 29th of February: a day either side of it is still one day,
	// which is why this counts calendar days rather than dividing instants.
	if got := datetime.DaysBetween(day(2028, time.February, 28), day(2028, time.March, 1)); got != 2 {
		t.Errorf("28 Feb to 1 Mar 2028 is %d days, want 2", got)
	}
	if got := datetime.DaysBetween(day(2026, time.March, 1), day(2026, time.February, 28)); got != -1 {
		t.Errorf("backwards is %d days, want -1", got)
	}
	// And across a day-saving boundary, where 24 hours is not a day.
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skipf("no zone database: %v", err)
	}
	before := time.Date(2026, time.March, 28, 12, 0, 0, 0, berlin)
	after := time.Date(2026, time.March, 29, 12, 0, 0, 0, berlin)
	if got := datetime.DaysBetween(before, after); got != 1 {
		t.Errorf("28 to 29 March across a clock change is %d days, want 1", got)
	}
}

func TestWeeksStartOnMonday(t *testing.T) {
	// 7 October 2026 is a Wednesday: Monday is 0, so a Wednesday is 2.
	if got := datetime.Weekday(fixed()); got != 2 {
		t.Errorf("Wednesday is %d in this package, want 2 (Monday 0)", got)
	}
	if got := datetime.Weekday(day(2026, time.October, 5)); got != 0 {
		t.Errorf("Monday is %d, want 0", got)
	}
	if got := datetime.Weekday(day(2026, time.October, 11)); got != 6 {
		t.Errorf("Sunday is %d, want 6", got)
	}
	// The week is the Monday-to-Sunday one, and StartOfWeek lands on it.
	start := datetime.StartOfWeek(fixed())
	if want := day(2026, time.October, 5); !start.Equal(want) {
		t.Errorf("the week starts %v, want %v", start, want)
	}
	if !datetime.SameWeek(day(2026, time.October, 5), day(2026, time.October, 11)) {
		t.Error("Monday and Sunday of the same week should be one week")
	}
	if datetime.SameWeek(day(2026, time.October, 11), day(2026, time.October, 12)) {
		t.Error("Sunday and the next Monday are not one week")
	}
	// A month that begins on a Sunday needs six rows; one that begins on a
	// Monday needs five. This is why the grid is seven columns wide.
	// February 2026 begins on a Sunday, so it needs five rows of seven; the
	// default is six, because six is the most any month needs and a grid that
	// changes height as you page moves its own click targets.
	cells := datetime.MonthCells(day(2026, time.February, 1), 5)
	if len(cells) != 35 {
		t.Errorf("February 2026 laid out in %d cells, want 35", len(cells))
	}
	if cells[0].Day.Day() != 26 || cells[0].InMonth {
		t.Error("the first cell of February 2026 should be 26 January, out of month")
	}
	// Five rows of seven from Monday 26 January ends on Sunday 1 March.
	if !datetime.SameDay(cells[len(cells)-1].Day, day(2026, time.March, 1)) {
		t.Errorf("the last cell is %v, want 1 March", cells[len(cells)-1].Day)
	}
	// A month that needs more rows than were asked for is given them rather
	// than losing its last days: August 2026 begins on a Saturday and so
	// spills into a sixth row, whatever the caller asked for.
	if got := len(datetime.MonthCells(day(2026, time.August, 1), 4)); got != 42 {
		t.Errorf("August 2026 laid out in %d cells, want 42", got)
	}
}

func TestISOWeekNumbersCrossTheYearBoundary(t *testing.T) {
	// 1 January 2027 is a Friday, and it belongs to week 53 of 2026: a week
	// belongs to the year of most of its days, and a booking system that says
	// otherwise is a booking system that disagrees with the calendar.
	year, week := datetime.WeekOf(day(2027, time.January, 1))
	if year != 2026 || week != 53 {
		t.Errorf("1 January 2027 is week %d of %d, want 53 of 2026", week, year)
	}
	year, week = datetime.WeekOf(day(2026, time.January, 1))
	if year != 2026 || week != 1 {
		t.Errorf("1 January 2026 is week %d of %d, want 1 of 2026", week, year)
	}
	if got := datetime.WeekOfYear(day(2026, time.October, 7)); got != 41 {
		t.Errorf("7 October 2026 is week %d, want 41", got)
	}
	// And the Monday of a week comes back from its number.
	if got := datetime.WeekStart(2026, 41); got.Day() != 5 || got.Month() != time.October {
		t.Errorf("week 41 of 2026 starts %v, want 5 October", got)
	}
}

func TestRelativeMeasuresInCalendarUnits(t *testing.T) {
	now := fixed()
	for _, tc := range []struct {
		what string
		then time.Time
		span datetime.RelativeSpan
		text string
	}{
		{
			what: "three days earlier",
			then: now.AddDate(0, 0, -3),
			span: datetime.RelativeSpan{Past: true, Days: 3, TotalDays: 3},
			text: "3 days ago",
		},
		{
			what: "two hours later",
			then: now.Add(2 * time.Hour),
			span: datetime.RelativeSpan{Hours: 2},
			text: "in 2 hours",
		},
		{
			// The span is a decomposition and not a rounded figure: a day
			// before, earlier in the day, is a day and the five hours that
			// are still there. RelativeText says "1 day ago" because a day is
			// the largest unit that is not zero, which is what a reader means.
			what: "one day before, earlier in the day",
			then: day(2026, time.October, 6).Add(9 * time.Hour),
			span: datetime.RelativeSpan{Past: true, Days: 1, Hours: 5, Minutes: 30, TotalDays: 1},
			text: "1 day ago",
		},
		{
			what: "the same moment",
			then: now,
			span: datetime.RelativeSpan{},
			text: "just now",
		},
	} {
		got := datetime.Relative(tc.then, now)
		if got != tc.span {
			t.Errorf("%s: span %+v, want %+v", tc.what, got, tc.span)
		}
		if text := datetime.RelativeText(got); text != tc.text {
			t.Errorf("%s: text %q, want %q", tc.what, text, tc.text)
		}
	}
	// Nine months and a week is not "280 days", and a month is not 30.44 of
	// them: the span keeps the units a reader would use.
	span := datetime.Relative(day(2026, time.October, 7), day(2025, time.December, 31))
	if span.Months != 9 || span.Days != 7 {
		t.Errorf("31 Dec 2025 to 7 Oct 2026 is %d months %d days, want 9 and 7",
			span.Months, span.Days)
	}
	// A month from a 31st is one month, however short that month is.
	span = datetime.Relative(day(2026, time.February, 28), day(2026, time.January, 31))
	if span.Months != 1 || span.Days != 0 {
		t.Errorf("31 Jan to 28 Feb is %+v, want one month and no days", span)
	}
}

func TestDurationTextWritesTwoUnits(t *testing.T) {
	for _, tc := range []struct {
		d    time.Duration
		want string
	}{
		{0, "0m"},
		{45 * time.Second, "45s"},
		{90 * time.Minute, "1h 30m"},
		{2 * time.Hour, "2h"},
		{50 * time.Hour, "2d 2h"},
		{-30 * time.Minute, "-30m"},
	} {
		if got := datetime.DurationText(tc.d); got != tc.want {
			t.Errorf("DurationText(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
	// What it writes, it reads back.
	back, err := datetime.ParseDurationText("2d 2h")
	if err != nil {
		t.Fatal(err)
	}
	if back != 50*time.Hour {
		t.Errorf("read back %v, want 50h", back)
	}
	if _, err := datetime.ParseDurationText("soon"); err == nil {
		t.Error("a word is not a duration")
	}
}

// The most negative duration there is has no positive: negating it is itself,
// so a DurationText that recursed on the negation of its argument never came
// back, and the frame that asked it took the process with it. It has to answer
// with a length as well — a bare "-" would be a sign with nothing after it.
func TestDurationTextSurvivesTheMostNegativeDuration(t *testing.T) {
	got := datetime.DurationText(math.MinInt64)
	if got == "" || got == "-" {
		t.Errorf("DurationText(math.MinInt64) = %q, which says no length at all", got)
	}
}

// Under a minute there is no unit left to count, so the answer has to come
// from the seconds rather than from nothing: the doc says zero is "0m" rather
// than an empty string, and half a second is a length, not a blank.
func TestDurationTextSaysSomethingAboutSubSecondDurations(t *testing.T) {
	if got := datetime.DurationText(500 * time.Millisecond); got == "" {
		t.Error("DurationText(500ms) returned an empty string, but its doc says " +
			"zero is \"0m\" rather than an empty string")
	}
	// What it writes, it reads back — the two are one vocabulary, and the
	// milliseconds are now part of it.
	back, err := datetime.ParseDurationText(datetime.DurationText(1500 * time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	if back != 1500*time.Millisecond {
		t.Errorf("read back %v, want 1.5s", back)
	}
}

// ── the grid, and the four states a day can be in ────────────────────────────

func TestGridDrawsTheMonthAndReportsTheDay(t *testing.T) {
	now := fixed()
	sel := day(2026, time.October, 9)
	tt := render(320, 260, func(c *ui.Context) {
		datetime.Grid(c, datetime.GridOptions{
			Month: now, Today: now, Selected: &sel, ShowWeekNumbers: true,
		})
	})
	hasText(t, tt, "Mon", "Tue", "Sun", "October", "30")
	if _, ok := tt.Find("Wk"); !ok {
		t.Errorf("the week-number column is missing its heading: %q", tt.Texts())
	}
	// Every day is named for a screen reader, which is also how a test can
	// tell two 9s apart.
	if _, ok := tt.Find("Friday 9 October 2026"); !ok {
		t.Error("a chosen day should be named in full")
	}

	var got time.Time
	ok := false
	press := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if d, hit := datetime.Grid(c, datetime.GridOptions{Month: now, Today: now}).Picked(); hit {
			got, ok = d, true
		}
	}, 320, 260)
	if err := press.Click("Thursday 8 October 2026"); err != nil {
		t.Fatal(err)
	}
	if !ok || !datetime.SameDay(got, day(2026, time.October, 8)) {
		t.Errorf("pressing the 8th reported %v", got)
	}
}

func TestGridDaysAreBigEnoughToHit(t *testing.T) {
	tt := render(320, 260, func(c *ui.Context) {
		datetime.Grid(c, datetime.GridOptions{Month: fixed()})
	})
	rect, ok := tt.Find("Wednesday 7 October 2026")
	if !ok {
		t.Fatal("no cell for the 7th")
	}
	if rect.W < datetime.MinCellSize || rect.H < datetime.MinCellSize {
		t.Errorf("a day cell is %v, want at least %v on both axes",
			rect, datetime.MinCellSize)
	}
}

func TestGridTellsItsFourStatesApart(t *testing.T) {
	// A calendar a person cannot read is a calendar they cannot use, so the
	// four states are compared by what is painted, not by what is written:
	// the day numbers are all two digits in the same face.
	now := fixed()
	sel := day(2026, time.October, 9)
	tt := render(320, 260, func(c *ui.Context) {
		datetime.Grid(c, datetime.GridOptions{
			Month:    now,
			Today:    now,
			Selected: &sel,
			Start:    day(2026, time.October, 12),
			End:      day(2026, time.October, 16),
			Disabled: func(d time.Time) bool { return d.Day() == 21 },
		})
	})
	img := tt.Image()
	cell := func(name string) map[color.RGBA]bool {
		t.Helper()
		r, ok := tt.Find(name)
		if !ok {
			t.Fatalf("no cell %q", name)
		}
		return colorsIn(img, r)
	}
	plain := cell("Thursday 8 October 2026")  // an ordinary day
	today := cell("Wednesday 7 October 2026") // a ring in the accent
	chosen := cell("Friday 9 October 2026")   // filled with ink
	inRange := cell("Monday 12 October 2026") // the accent's pale fill
	outside := cell("Monday 5 October 2026")  // not this month
	off := cell("Wednesday 21 October 2026")  // greyed out and dead

	states := []struct {
		what string
		cols map[color.RGBA]bool
	}{
		{"today", today}, {"chosen", chosen}, {"in range", inRange},
		{"out of month", outside}, {"disabled", off},
	}
	for i := range states {
		for j := i + 1; j < len(states); j++ {
			if sameColors(states[i].cols, states[j].cols) {
				t.Errorf("a %s day and a %s day are painted the same way",
					states[i].what, states[j].what)
			}
		}
	}
	// A chosen day is filled, so it cannot be mistaken for a day that is only
	// marked: the plain day is the window's own background and nothing else.
	if len(chosen) == 1 && sameColors(chosen, plain) {
		t.Error("a chosen day is painted like an ordinary one")
	}
}

func TestGridDisabledDayTakesNoClicks(t *testing.T) {
	now := fixed()
	var got time.Time
	picked := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		g := datetime.Grid(c, datetime.GridOptions{
			Month:    now,
			Disabled: func(d time.Time) bool { return d.Day() == 21 },
		})
		if d, ok := g.Picked(); ok {
			got, picked = d, true
		}
	}, 320, 260)
	if err := tt.Click("Wednesday 21 October 2026"); err != nil {
		t.Fatal(err)
	}
	if picked {
		t.Errorf("a disabled day reported a press of %v", got)
	}
	// A day that is not disabled still answers.
	if err := tt.Click("Wednesday 28 October 2026"); err != nil {
		t.Fatal(err)
	}
	if !picked || got.Day() != 28 {
		t.Errorf("pressing the 28th reported %v", got)
	}
}

func TestGridNamesItsWeekdays(t *testing.T) {
	// A window that speaks another language passes its own names, and the
	// heading is clipped to fit the column while the full name is what a
	// screen reader hears.
	custom := datetime.WeekdayNames{"一", "二", "三", "四", "五", "六", "日"}
	tt := render(320, 260, func(c *ui.Context) {
		datetime.Grid(c, datetime.GridOptions{Month: fixed(), Weekdays: custom})
	})
	hasText(t, tt, "一", "二", "日")
	// The first column is a Monday: 5 October 2026.
	if _, ok := tt.Find("Monday 5 October 2026"); !ok {
		t.Error("the first column of the grid should be the Monday")
	}
}

func TestCalendarPagesOnAMonthTheCallerHolds(t *testing.T) {
	month := day(2026, time.October, 15)
	tt := render(360, 300, func(c *ui.Context) {
		r := datetime.Calendar(c, datetime.CalendarOptions{
			Month: &month, Today: fixed(), TodayButton: true, Header: true,
		})
		if day, ok := r.Picked(); ok {
			t.Logf("picked %v", day)
		}
	})
	hasText(t, tt, "October 2026", "Previous month", "Next month", "Today")

	// The arrows write through the caller's month, and the frame they move
	// still shows the old one: a heading that has moved and a grid that has
	// not is the bug this ordering exists to prevent.
	before := month
	arrow := render(360, 300, func(c *ui.Context) {
		m := before
		datetime.Calendar(c, datetime.CalendarOptions{Month: &m, Header: true})
	})
	if err := arrow.Click("Next month"); err != nil {
		t.Fatal(err)
	}
	if !month.Equal(before) {
		t.Error("the month should move after the frame, not inside it")
	}
	_ = tt
}

func TestCalendarNextArrowMovesTheMonthAndTheDayIsKept(t *testing.T) {
	month := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	tt := render(360, 300, func(c *ui.Context) {
		datetime.Calendar(c, datetime.CalendarOptions{Month: &month, Header: true})
	})
	if err := tt.Click("Next month"); err != nil {
		t.Fatal(err)
	}
	// 31 January plus a month clamps to 28 February rather than skipping to
	// March, so paging never loses the day a caller was looking at.
	if month.Month() != time.February || month.Day() != 28 {
		t.Errorf("paging from 31 January landed on %v, want 28 February", month)
	}
}

// ── one render per component ─────────────────────────────────────────────────

func TestTimeGridDrawsHoursAndTheNowLine(t *testing.T) {
	tt := render(220, 640, func(c *ui.Context) {
		datetime.CalendarTimeGrid(c, datetime.TimeGridOptions{
			Day: fixed(), StartHour: 8, EndHour: 18, Hours: true, Now: fixed(),
		})
	})
	hasText(t, tt, "08", "17", "Now")
	// 14:30 is six and a half hours down an eleven-hour column.
	if _, ok := tt.Find("Now"); !ok {
		t.Error("the current-time line is missing")
	}
}

func TestTimeGridPicksTheHour(t *testing.T) {
	var got time.Time
	ok := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		g := datetime.CalendarTimeGrid(c, datetime.TimeGridOptions{
			Day: fixed(), StartHour: 8, EndHour: 18, Pickable: true,
		})
		if t2, hit := g.Picked(); hit {
			got, ok = t2, true
		}
	}, 220, 640)
	if err := tt.Click("14 2026-10-07"); err != nil {
		t.Fatal(err)
	}
	if !ok || got.Hour() != 14 || got.Minute() != 0 {
		t.Errorf("pressing the 14:00 row reported %v", got)
	}
}

func TestCurrentTimeIndicatorSitsAtTheRightMinute(t *testing.T) {
	tt := render(220, 640, func(c *ui.Context) {
		datetime.CalendarTimeGrid(c, datetime.TimeGridOptions{
			Day: fixed(), StartHour: 8, EndHour: 18, Now: fixed(), Pickable: true,
		})
	})
	now, ok := tt.Find("Now")
	if !ok {
		t.Fatal("no now line")
	}
	eight, ok := tt.Find("08 2026-10-07")
	if !ok {
		t.Fatal("no 08:00 row")
	}
	// 14:30 is 6.5 hours below 08:00, and an hour is one row.
	want := eight.Y + (8-8)*56 + 6.5*56
	if mathAbs(float64(now.Y-want)) > 2 {
		t.Errorf("the now line is at %.1f, want about %.1f", now.Y, want)
	}
}

func mathAbs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func TestCalendarDayViewShowsItsEvents(t *testing.T) {
	at9 := day(2026, time.October, 7).Add(9 * time.Hour)
	tt := render(320, 700, func(c *ui.Context) {
		datetime.CalendarDayView(c, datetime.CalendarDayViewOptions{
			Day:    at9,
			Events: []datetime.Event{{Title: "Maple Street Bakery", Start: at9, End: at9.Add(time.Hour)}},
			Now:    fixed(),
			AllDay: true,
		})
	})
	hasText(t, tt, "Wednesday", "7 October 2026", "09:00  Maple Street Bakery", "Now")
}

func TestCalendarWeekViewShowsSevenDays(t *testing.T) {
	at9 := day(2026, time.October, 7).Add(9 * time.Hour)
	at14 := day(2026, time.October, 7).Add(14 * time.Hour)
	tt := render(900, 700, func(c *ui.Context) {
		datetime.CalendarWeekView(c, datetime.CalendarWeekViewOptions{
			Day:   at9,
			Now:   fixed(),
			Title: true,
			Events: []datetime.Event{
				{Title: "Oak Lane Cafe", Start: at9, End: at9.Add(time.Hour)},
				{Title: "Riverside Clinic", Start: at14, End: at14.Add(time.Hour)},
			},
		})
	})
	// The week of Wednesday 7 October 2026 runs from Monday the 5th to
	// Sunday the 11th, and the title names the ISO week.
	hasText(t, tt, "5–11 October 2026 · W41", "Mon", "Sun", "09:00  Oak Lane Cafe")
	// One ruler for seven columns, not seven: the only element showing "08"
	// is the hour at the top of it.
	if n := countText(tt, "08"); n != 1 {
		t.Errorf("the hour numbers are drawn %d times, want once for the week", n)
	}
	// And the current-time line belongs to today's column alone.
	if n := countText(tt, "Now"); n != 1 {
		t.Errorf("the current-time line is drawn %d times, want once", n)
	}
}

func TestCalendarMonthAndYearViews(t *testing.T) {
	now := fixed()
	open := day(2026, time.October, 1)
	tt := render(900, 1400, func(c *ui.Context) {
		datetime.CalendarMonthView(c, datetime.CalendarMonthViewOptions{
			Month: now, Today: now, Selected: &open, Caption: true, Weeks: 5,
		})
	})
	hasText(t, tt, "October 2026")

	year := day(2026, time.January, 1)
	wide := render(1200, 1600, func(c *ui.Context) {
		datetime.CalendarYearView(c, datetime.CalendarYearViewOptions{
			Year: &year, Today: now, Highlighted: &open, Columns: 3,
		})
	})
	hasText(t, wide, "2026", "January", "December", "Previous year", "Next year")
}

func TestDatePickerChoosesADayAndKeepsTheClock(t *testing.T) {
	value := fixed()
	open := false
	var picked time.Time
	ok := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := datetime.DatePicker(c, datetime.DatePickerOptions{
			Value: &value, Open: &open, Today: fixed(),
		})
		if d, hit := r.Picked(); hit {
			picked, ok = d, true
		}
	}, 360, 300)
	// The field shows what it holds before anything is chosen.
	hasText(t, tt, "2026-10-07")
	if err := tt.Click("Date"); err != nil {
		t.Fatal(err)
	}
	if !open {
		t.Fatal("the calendar did not open")
	}
	hasText(t, tt, "October 2026", "Friday 9 October 2026")
	if err := tt.Click("Friday 9 October 2026"); err != nil {
		t.Fatal(err)
	}
	if !ok || picked.Day() != 9 {
		t.Errorf("the picker reported %v", picked)
	}
	// The clock is untouched: a date picker that reset the time would move a
	// timed booking to the small hours without saying so.
	if value.Hour() != 14 || value.Minute() != 30 {
		t.Errorf("picking a day moved the clock to %v", value.Format("15:04"))
	}
}

func TestDateRangePickerTakesTwoPresses(t *testing.T) {
	var from, to time.Time
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.DateRangePicker(c, datetime.DateRangePickerOptions{
			Start: &from, End: &to, Open: &open, Today: fixed(), Panels: 2,
		})
	}, 760, 320)
	if err := tt.Click("Dates"); err != nil {
		t.Fatal(err)
	}
	hasText(t, tt, "October 2026", "November 2026")
	if err := tt.Click("Monday 12 October 2026"); err != nil {
		t.Fatal(err)
	}
	if from.Day() != 12 || !to.IsZero() {
		t.Errorf("the first press gave %v – %v", from, to)
	}
	// A second press completes it.
	if err := tt.Click("Thursday 15 October 2026"); err != nil {
		t.Fatal(err)
	}
	if to.IsZero() || to.Day() != 15 {
		t.Fatalf("the second press left the range %v – %v", from, to)
	}
	// A third press on a range that is already complete starts a new one,
	// which is the only way to correct a mis-click without an undo.
	if err := tt.Click("Tuesday 20 October 2026"); err != nil {
		t.Fatal(err)
	}
	if from.Day() != 20 || !to.IsZero() {
		t.Errorf("a press on a finished range gave %v – %v", from, to)
	}
	// A press before the start swaps the two ends rather than drawing a range
	// backwards: 12 October, then the 11th, is the 11th to the 12th.
	if err := tt.Click("Friday 16 October 2026"); err != nil {
		t.Fatal(err)
	}
	if from.Day() != 16 || to.IsZero() {
		t.Fatalf("finishing the range gave %v – %v", from, to)
	}
	if err := tt.Click("Monday 12 October 2026"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Sunday 11 October 2026"); err != nil {
		t.Fatal(err)
	}
	if from.Day() != 11 || to.Day() != 12 {
		t.Errorf("a backwards press gave %v – %v", from, to)
	}
}

func TestTimePickerStepsTheClock(t *testing.T) {
	value := fixed()
	var changed bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		changed = datetime.TimePicker(c, datetime.TimePickerOptions{
			Value: &value, MinuteStep: 15,
		}).Changed() || changed
	}, 320, 220)
	hasText(t, tt, "14", "30")
	if err := tt.Click("One hour later"); err != nil {
		t.Fatal(err)
	}
	if !changed || value.Hour() != 15 {
		t.Errorf("an hour later gave %v, changed %v", value, changed)
	}
	if err := tt.Click("One minute later"); err != nil {
		t.Fatal(err)
	}
	if value.Minute() != 45 {
		t.Errorf("a quarter of an hour later gave %v", value)
	}
}

func TestTimeAndDateTimePickers(t *testing.T) {
	from := fixed()
	to := fixed().Add(90 * time.Minute)
	tt := render(420, 200, func(c *ui.Context) {
		datetime.TimeRangePicker(c, datetime.TimeRangePickerOptions{
			Start: &from, End: &to,
		})
	})
	// The length is the reason this is one component: two clocks on their own
	// do not say whether the job fits.
	hasText(t, tt, "1h 30m", "From", "Until")

	moment := fixed()
	open := false
	both := render(520, 220, func(c *ui.Context) {
		datetime.DateTimePicker(c, datetime.DateTimePickerOptions{
			Value: &moment, Open: &open, Today: fixed(),
		})
	})
	hasText(t, both, "2026-10-07", "14", "30")
}

func TestMonthYearAndWeekPickers(t *testing.T) {
	month := day(2026, time.October, 1)
	mt := render(420, 320, func(c *ui.Context) {
		datetime.MonthPicker(c, datetime.MonthPickerOptions{Month: &month, Today: fixed()})
	})
	hasText(t, mt, "2026", "January", "October", "December", "Previous year", "Next year")

	year := day(2026, time.January, 1)
	yt := render(420, 320, func(c *ui.Context) {
		datetime.YearPicker(c, datetime.YearPickerOptions{Year: &year, Today: fixed()})
	})
	// A page is a whole number of spans from zero, so it does not creep: the
	// same year is always on the same page, whatever the arrows did.
	hasText(t, yt, "2016 – 2027", "2026")

	chosen := fixed()
	wt := render(700, 220, func(c *ui.Context) {
		datetime.WeekPicker(c, datetime.WeekPickerOptions{
			Value: &chosen, Today: fixed(), Weeks: 3,
		})
	})
	hasText(t, wt, "5–11 October 2026 · W41")
	if _, ok := wt.Find("Wednesday 7 October 2026"); !ok {
		t.Errorf("the chosen week's days are missing: %q", wt.Texts())
	}
}

func TestWeekPickerWritesTheDayItIsGiven(t *testing.T) {
	chosen := fixed()
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.WeekPicker(c, datetime.WeekPickerOptions{
			Value: &chosen, Today: fixed(), Weeks: 3,
		})
	}, 700, 220)
	if err := tt.Click("Friday 9 October 2026"); err != nil {
		t.Fatal(err)
	}
	if chosen.Day() != 9 {
		t.Errorf("pressing the 9th left the value at %v", chosen)
	}
}

func TestAgendaAttendeesAndAvailability(t *testing.T) {
	now := fixed()
	today := day(2026, time.October, 7)
	tomorrow := today.AddDate(0, 0, 1)
	at9 := today.Add(9 * time.Hour)
	at11 := today.Add(11 * time.Hour)
	tt := render(420, 620, func(c *ui.Context) {
		datetime.AgendaView(c, datetime.AgendaOptions{
			Today: now,
			Days: []datetime.AgendaDay{
				{Day: today, Events: []datetime.Event{
					{Title: "Maple Street Bakery", Start: at9, End: at9.Add(time.Hour)},
					{Title: "Oak Lane Cafe", Start: at11, End: at11.Add(time.Hour)},
				}},
				{Day: tomorrow, Events: []datetime.Event{
					{Title: "Team stand-up", Start: tomorrow.Add(9 * time.Hour)},
				}},
			},
			Empty: func(day time.Time) {
				core.Use(c, core.Settings{})
			},
		})
	})
	hasText(t, tt, "2026-10-07", "09:00  Maple Street Bakery", "2026-10-08", "Team stand-up")

	people := render(420, 320, func(c *ui.Context) {
		datetime.AttendeeList(c, datetime.AttendeeOptions{
			Avatars:      true,
			ShowResponse: true,
			Attendees: []datetime.Attendee{
				{Name: "Dana Reyes", Response: datetime.Accepted, Timezone: "Europe/Berlin"},
				{Name: "Nate Coleman", Response: datetime.NoReply, Note: "no reply yet"},
			},
		})
	})
	// The answers are words, because a row of green and grey circles is a
	// puzzle rather than an answer.
	hasText(t, people, "Dana Reyes", "Yes", "Nate Coleman", "No reply", "Europe/Berlin")

	days := []time.Time{today, tomorrow}
	free := map[string]bool{"2026-10-07T09:00:00Z": true}
	av := render(420, 700, func(c *ui.Context) {
		datetime.AvailabilityPicker(c, datetime.AvailabilityOptions{
			Days: days, StartHour: 8, EndHour: 12,
			Free: func(d time.Time, hour int) bool {
				return free[d.Add(time.Duration(hour)*time.Hour).Format(time.RFC3339)]
			},
		})
	})
	hasText(t, av, "08", "09", "11", "7 October 2026", "8 October 2026")
}

func TestDurationAndTimezoneControls(t *testing.T) {
	length := 45 * time.Minute
	var changed bool
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		changed = datetime.DurationPicker(c, datetime.DurationPickerOptions{
			Value:   &length,
			Presets: []time.Duration{30 * time.Minute, time.Hour, 2 * time.Hour},
		}).Changed() || changed
	}, 420, 220)
	hasText(t, tt, "45m", "30m", "1h", "2h")
	if err := tt.Click("Longer"); err != nil {
		t.Fatal(err)
	}
	if !changed || length != time.Hour {
		t.Errorf("a step gave %v, changed %v", length, changed)
	}

	zone := "Europe/Berlin"
	zt := render(420, 120, func(c *ui.Context) {
		datetime.TimezoneSelect(c, datetime.TimezoneSelectOptions{
			Value: &zone, At: fixed(), ShowOffsets: true,
			Zones: []string{"Europe/Berlin", "UTC", "America/New_York"},
		})
	})
	// The zone in a label is UTC, which needs no database; a city needs one,
	// and the offset is worked out for the moment being shown.
	hasText(t, zt, "UTC")
	if !strings.Contains(strings.Join(zt.Texts(), " "), "UTC") {
		t.Errorf("the zone is not shown: %q", zt.Texts())
	}
}

func TestCountdownStopwatchAndRelativeTime(t *testing.T) {
	now := fixed()
	tt := render(420, 420, func(c *ui.Context) {
		datetime.Countdown(c, datetime.CountdownOptions{
			Target: now.AddDate(0, 0, 3), Now: now,
			From: now.Add(-24 * time.Hour), Until: now.AddDate(0, 0, 3),
			Format: datetime.RelativeText,
		})
		datetime.RelativeTime(c, datetime.RelativeTimeOptions{
			Then: now.AddDate(0, 0, -3), Now: now, Format: datetime.RelativeText,
		})
		datetime.RelativeTime(c, datetime.RelativeTimeOptions{
			Then: now.Add(2 * time.Hour), Now: now, Format: datetime.RelativeText,
		})
	})
	hasText(t, tt, "in 3 days", "3 days ago", "in 2 hours")

	start := now.Add(-83 * time.Second)
	laps := []time.Time{now.Add(-40 * time.Second)}
	st := render(420, 420, func(c *ui.Context) {
		datetime.Stopwatch(c, datetime.StopwatchOptions{
			Start: &start, Now: now, Laps: &laps,
		})
	})
	// 83 seconds is one minute and twenty-three seconds, to the hundredth a
	// person can read off a stopwatch.
	hasText(t, st, "01:23.00", "1.  00:43.00")

	// The wording is the caller's, which is the whole point of passing a
	// function in rather than a library string.
	own := render(420, 120, func(c *ui.Context) {
		datetime.RelativeTime(c, datetime.RelativeTimeOptions{
			Then: now.AddDate(0, 0, -3), Now: now,
			Format: func(s datetime.RelativeSpan) string {
				return fmt.Sprintf("vor %d Tagen", s.Days)
			},
		})
	})
	hasText(t, own, "vor 3 Tagen")
}

func TestReminderPickerAndRecurrence(t *testing.T) {
	now := fixed()
	reminder := 15 * time.Minute
	rt := render(700, 220, func(c *ui.Context) {
		datetime.ReminderPicker(c, datetime.ReminderPickerOptions{Value: &reminder})
	})
	hasText(t, rt, "No reminder", "15m before", "1h before")
	if err := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.ReminderPicker(c, datetime.ReminderPickerOptions{Value: &reminder})
	}, 700, 220).Click("1h before"); err != nil {
		t.Fatal(err)
	}
	if reminder != time.Hour {
		t.Errorf("choosing an hour left the reminder at %v", reminder)
	}

	rule := datetime.Recurrence{
		Frequency: datetime.Weekly, Interval: 2,
		Weekdays: [7]bool{true, false, false, false, true, false, false},
		Count:    10,
	}
	ct := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.RecurrenceEditor(c, datetime.RecurrenceEditorOptions{
			Rule: &rule, Anchor: now,
		})
	}, 700, 320)
	// The sentence under the controls is the only place the whole rule is
	// visible at once, and it has to be checkable.
	hasText(t, ct, "Mon, Fri", "Every 2 weeks on Mon, Fri for 10 times")
	if err := ct.Click("Month"); err != nil {
		t.Fatal(err)
	}
	if rule.Frequency != datetime.Monthly {
		t.Errorf("choosing monthly left the rule at %v", rule.Frequency)
	}
	if got := datetime.RecurrenceSummary(datetime.Recurrence{}); got != "Does not repeat" {
		t.Errorf("an empty rule reads %q", got)
	}
}

func TestCronEditorAndItsSummary(t *testing.T) {
	expr := "30 9 * * 1-5"
	if err := datetime.ValidateCron(expr); err != nil {
		t.Fatal(err)
	}
	if got := datetime.CronSummary(expr); !strings.Contains(got, "Monday") ||
		!strings.Contains(got, "09:30") {
		t.Errorf("%q reads %q", expr, got)
	}
	// A rule this package cannot read is said so rather than saved.
	for _, bad := range []string{"30 9 * *", "99 9 * * *", "a b c d e", "30 9 * * MON"} {
		if err := datetime.ValidateCron(bad); err == nil {
			t.Errorf("%q should not be a valid cron expression", bad)
		}
	}

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.CronEditor(c, datetime.CronEditorOptions{Expression: &expr})
	}, 700, 520)
	hasText(t, tt, "Minute", "Day of week", "On Monday")
	if err := tt.Click("0"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(expr, "0 9") {
		t.Errorf("choosing minute zero gave %q", expr)
	}
}

func TestEventChipPopoverAndEditor(t *testing.T) {
	now := fixed()
	e := datetime.Event{
		Title: "Maple Street Bakery", Start: now, End: now.Add(time.Hour),
		Location: "12 Maple Street", Severity: core.Accent,
	}
	ct := render(420, 160, func(c *ui.Context) {
		datetime.EventChip(c, datetime.EventChipOptions{Event: e})
	})
	hasText(t, ct, "14:30  Maple Street Bakery · 12 Maple Street", "Maple Street Bakery")

	open := true
	anchor := render(420, 320, func(c *ui.Context) {
		row := ui.Box(c).FillWidth().Label("chip").Children(func() {
			ui.Text(c, "Maple Street Bakery")
		})
		datetime.EventPopover(c, datetime.EventPopoverOptions{
			Anchor: row, Open: &open, Event: e,
			Edit: "Edit", Delete: "Delete",
		})
	})
	hasText(t, anchor, "2026-10-07 14:30 – 2026-10-07 15:30", "1h", "12 Maple Street", "Edit", "Delete")

	edited := false
	editorOpen := false
	et := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		e2 := e
		r := datetime.EventEditor(c, datetime.EventEditorOptions{
			Event: &e2, Save: "Save", Open: &editorOpen,
			Duration: time.Hour,
			Attendees: []datetime.Attendee{
				{Name: "Dana Reyes", Response: datetime.Accepted},
			},
		})
		edited = r.Saved() || edited
	}, 520, 700)
	hasText(t, et, "Title", "Maple Street Bakery", "Save", "Cancel", "Dana Reyes")
	if err := et.Click("Save"); err != nil {
		t.Fatal(err)
	}
	if !edited {
		t.Error("the editor did not report its save")
	}
}

func TestEventOverlapsGoIntoLanes(t *testing.T) {
	now := fixed()
	a := datetime.Event{Title: "A", Start: now, End: now.Add(time.Hour)}
	b := datetime.Event{Title: "B", Start: now.Add(30 * time.Minute), End: now.Add(90 * time.Minute)}
	c := datetime.Event{Title: "C", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)}
	lanes := datetime.OverlapLanes([]datetime.Event{a, b, c})
	if len(lanes) != 2 {
		t.Fatalf("three events, two of them overlapping, made %d lanes, want 2", len(lanes))
	}
	// The first lane holds the first and the third: they do not overlap, and
	// putting them side by side would waste the width of the column.
	if len(lanes[0]) != 2 {
		t.Errorf("the first lane holds %d events, want 2", len(lanes[0]))
	}
	if !datetime.Overlaps(a, b) {
		t.Error("A and B overlap and should say so")
	}
	// Two events that only touch do not: 09:00–10:00 and 10:00–11:00 are a
	// schedule, not a conflict.
	if datetime.Overlaps(a, datetime.Event{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)}) {
		t.Error("events that touch at an instant should not be an overlap")
	}
	// And the day's events come back in the order they start.
	got := datetime.EventsOn([]datetime.Event{c, a, b}, now)
	if len(got) != 3 || got[0].Title != "A" {
		t.Errorf("EventsOn returned %v", titles(got))
	}
}

func titles(events []datetime.Event) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = e.Title
	}
	return out
}

// ── appearance ───────────────────────────────────────────────────────────────

// Every component, once, in the dark palette. A colour read from a token is a
// different value in the two appearances, so a component that reaches for a
// literal instead looks right in one of them and wrong in the other — and
// nothing about the drawing fails to say so.
func TestEveryComponentRendersInDarkMode(t *testing.T) {
	now := fixed()
	sel := day(2026, time.October, 9)
	openMonth := day(2026, time.October, 1)
	year := yearAt(2026)
	open := false
	moment := now
	rule := datetime.Recurrence{Frequency: datetime.Weekly, Interval: 2}
	expr := "0 9 * * 1-5"
	length := time.Hour
	reminder := time.Hour
	zone := "UTC"
	rangeStart, rangeEnd := day(2026, time.October, 7), day(2026, time.October, 9)
	stopwatchStart := now.Add(-83 * time.Second)
	editorOpen := false
	laps := []time.Time{now.Add(-40 * time.Second)}

	tt := renderDark(1000, 2400, func(c *ui.Context) {
		u := core.Density(c).Unit()
		datetime.Grid(c, datetime.GridOptions{Month: now, Today: now, Selected: &sel})
		datetime.Calendar(c, datetime.CalendarOptions{Month: &openMonth, Today: now, Header: true})
		datetime.CalendarMonthView(c, datetime.CalendarMonthViewOptions{Month: now, Today: now})
		datetime.CalendarYearView(c, datetime.CalendarYearViewOptions{
			Year: &year, Today: now, Highlighted: &openMonth, Columns: 3,
		})
		datetime.CalendarDayView(c, datetime.CalendarDayViewOptions{
			Day: now, Now: now, Pickable: true, AllDay: true,
			Events: []datetime.Event{{Title: "Maple Street Bakery", Start: now}},
		})
		datetime.CalendarWeekView(c, datetime.CalendarWeekViewOptions{
			Day: now, Now: now, Title: true, Pickable: true,
		})
		datetime.CalendarEvents(c, datetime.CalendarEventsOptions{
			Day: now, AllDay: true,
			Events: []datetime.Event{{Title: "Oak Lane Cafe", Start: now, Severity: core.Warning}},
		})
		datetime.CalendarTimeGrid(c, datetime.TimeGridOptions{Day: now, Now: now, Hours: true})
		datetime.CurrentTimeIndicator(c, datetime.CurrentTimeIndicatorOptions{Now: now})
		datetime.EventChip(c, datetime.EventChipOptions{Event: datetime.Event{
			Title: "Oak Lane Cafe", Start: now, Severity: core.Danger,
		}})
		// The anchor has to be built inside the frame, because that is when
		// an element exists at all.
		anchor := ui.Box(c).Label("anchor").Children(func() {
			ui.Text(c, "Maple Street Bakery")
		})
		datetime.EventPopover(c, datetime.EventPopoverOptions{
			Anchor: anchor, Open: &open, Event: datetime.Event{
				Title: "Maple Street Bakery", Start: now, End: now.Add(time.Hour),
			},
			Edit: "Edit", Delete: "Delete",
		})
		datetime.EventEditor(c, datetime.EventEditorOptions{
			Event: &datetime.Event{Title: "Maple Street Bakery", Start: now},
			Save:  "Save",
			Open:  &editorOpen,
			Attendees: []datetime.Attendee{
				{Name: "Dana Reyes", Response: datetime.Accepted},
			},
			Duration: time.Hour,
		})
		datetime.DatePicker(c, datetime.DatePickerOptions{Value: &moment, Open: &open, Today: now})
		datetime.DateRangePicker(c, datetime.DateRangePickerOptions{
			Start: &rangeStart, End: &rangeEnd, Open: &open, Today: now, Panels: 2,
		})
		datetime.DateTimePicker(c, datetime.DateTimePickerOptions{Value: &moment, Open: &open, Today: now})
		datetime.TimePicker(c, datetime.TimePickerOptions{Value: &moment})
		datetime.MonthPicker(c, datetime.MonthPickerOptions{Month: &openMonth, Today: now})
		datetime.YearPicker(c, datetime.YearPickerOptions{Year: &year, Today: now})
		datetime.WeekPicker(c, datetime.WeekPickerOptions{Value: &moment, Today: now, Weeks: 2})
		datetime.AvailabilityPicker(c, datetime.AvailabilityOptions{
			Days: []time.Time{day(2026, time.October, 7)}, StartHour: 8, EndHour: 12,
			Free: func(time.Time, int) bool { return true },
		})
		datetime.DurationPicker(c, datetime.DurationPickerOptions{
			Value: &length, Presets: []time.Duration{time.Hour},
		})
		datetime.TimezoneSelect(c, datetime.TimezoneSelectOptions{
			Value: &zone, At: now, ShowOffsets: true, Zones: []string{"UTC"},
		})
		datetime.ReminderPicker(c, datetime.ReminderPickerOptions{Value: &reminder})
		datetime.RecurrenceEditor(c, datetime.RecurrenceEditorOptions{Rule: &rule, Anchor: now})
		datetime.CronEditor(c, datetime.CronEditorOptions{Expression: &expr})
		datetime.Countdown(c, datetime.CountdownOptions{
			Target: now.AddDate(0, 0, 2), Now: now, Format: datetime.RelativeText,
		})
		datetime.Stopwatch(c, datetime.StopwatchOptions{
			Start: &stopwatchStart, Now: now, Laps: &laps,
		})
		datetime.RelativeTime(c, datetime.RelativeTimeOptions{
			Then: now.AddDate(0, 0, -1), Now: now, Format: datetime.RelativeText,
		})
		datetime.AttendeeList(c, datetime.AttendeeOptions{
			Avatars: true, ShowResponse: true,
			Attendees: []datetime.Attendee{{Name: "Dana Reyes", Response: datetime.Maybe}},
		})
		datetime.AgendaView(c, datetime.AgendaOptions{
			Today: now,
			Days: []datetime.AgendaDay{{Day: now, Events: []datetime.Event{
				{Title: "Oak Lane Cafe", Start: now, End: now.Add(time.Hour)},
			}}},
		})
		_ = u
	})
	// Every one of them drew: the month, the day's job, the reminders, the
	// rule, the reading.
	hasText(t, tt, "October 2026", "Maple Street Bakery", "Oak Lane Cafe",
		"15m before", "Every 2 weeks", "in 2 days", "1 day ago", "No reminder",
		"Dana Reyes", "Now")
}

func yearAt(y int) time.Time { return day(y, time.January, 1) }

func TestDarkModePicksTheDarkTokens(t *testing.T) {
	// The same chosen day, in both appearances, filled with two different
	// inks: a component that chose its own colours would paint the same one
	// twice and be unreadable in one of them.
	now := fixed()
	sel := day(2026, time.October, 9)
	paint := func(dark bool) color.RGBA {
		tt := ui.NewTester(func(c *ui.Context) {
			mode := core.Light
			if dark {
				mode = core.Dark
			}
			core.Use(c, core.Settings{Mode: mode})
			datetime.Grid(c, datetime.GridOptions{Month: now, Selected: &sel})
		}, 320, 260)
		r, ok := tt.Find("Friday 9 October 2026")
		if !ok {
			t.Fatal("no chosen day")
		}
		return tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+4))
	}
	light, dark := paint(false), paint(true)
	if light == dark {
		t.Errorf("a chosen day is painted %v in both appearances", light)
	}
	for _, tc := range []struct {
		name string
		got  color.RGBA
		want ui.Color
	}{
		{"light", light, theme.Light().Fill},
		{"dark", dark, theme.Dark().Fill},
	} {
		w := tc.want
		if tc.got != (color.RGBA{R: w.R, G: w.G, B: w.B, A: 255}) {
			t.Errorf("in %s a chosen day is filled %v, want the ink %v",
				tc.name, tc.got, w)
		}
	}
}

// ── what a component refuses to guess ────────────────────────────────────────

func TestComponentsPanicRatherThanGuess(t *testing.T) {
	now := fixed()
	month := now
	noValue := map[string]func(c *ui.Context){
		"Calendar": func(c *ui.Context) {
			datetime.Calendar(c, datetime.CalendarOptions{})
		},
		"CalendarYearView": func(c *ui.Context) {
			datetime.CalendarYearView(c, datetime.CalendarYearViewOptions{})
		},
		"DatePicker": func(c *ui.Context) {
			open := false
			datetime.DatePicker(c, datetime.DatePickerOptions{Open: &open, Today: now})
		},
		"TimePicker": func(c *ui.Context) {
			datetime.TimePicker(c, datetime.TimePickerOptions{})
		},
		"MonthPicker": func(c *ui.Context) {
			datetime.MonthPicker(c, datetime.MonthPickerOptions{})
		},
		"YearPicker": func(c *ui.Context) {
			datetime.YearPicker(c, datetime.YearPickerOptions{})
		},
		"EventChip": func(c *ui.Context) {
			datetime.EventChip(c, datetime.EventChipOptions{
				Event: datetime.Event{Start: now},
			})
		},
		"RelativeTime": func(c *ui.Context) {
			datetime.RelativeTime(c, datetime.RelativeTimeOptions{Then: now, Now: now})
		},
		"Countdown": func(c *ui.Context) {
			datetime.Countdown(c, datetime.CountdownOptions{
				Now: now, Format: datetime.RelativeText,
			})
		},
		"TimezoneSelect": func(c *ui.Context) {
			zone := "UTC"
			datetime.TimezoneSelect(c, datetime.TimezoneSelectOptions{Value: &zone})
		},
	}
	for name, body := range noValue {
		t.Run(name, func(t *testing.T) {
			defer func() {
				got := recover()
				if got == nil {
					t.Fatalf("%s drew itself with nothing to draw", name)
				}
				if msg, _ := got.(string); !strings.HasPrefix(msg, "datetime: ") {
					t.Errorf("%s panicked with %v, want a datetime: message", name, got)
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				body(c)
			}, 400, 300)
		})
	}
	_ = month
}

func TestAComponentThatForgetsUsePanics(t *testing.T) {
	// core.Use is the first call of every frame, and a component that reads
	// the theme without it is a missing call rather than a wrong colour.
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("a grid drew itself with no theme installed")
		}
		if msg, _ := got.(string); !strings.Contains(msg, "core.Use") {
			t.Errorf("panicked with %v, want a message about core.Use", got)
		}
	}()
	ui.NewTester(func(c *ui.Context) {
		datetime.Grid(c, datetime.GridOptions{Month: fixed()})
	}, 320, 260)
}

// ── the rest of what the components do when they are pressed ────────────────

func TestMonthAndYearPickersWriteTheMonth(t *testing.T) {
	month := day(2026, time.October, 1)
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.MonthPicker(c, datetime.MonthPickerOptions{Month: &month, Today: fixed()})
	}, 460, 340)
	if err := tt.Click("February 2026"); err != nil {
		t.Fatal(err)
	}
	if month.Month() != time.February || month.Year() != 2026 {
		t.Errorf("choosing February gave %v", month)
	}
	// The arrows move the year, and they do it after the frame, so the title
	// and the months always agree: the frame that carries the press still
	// shows 2026, and the one after it shows 2027.
	paged := month
	pager := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.MonthPicker(c, datetime.MonthPickerOptions{Month: &paged, Today: fixed()})
	}, 460, 340)
	before := paged
	if err := pager.Click("Next year"); err != nil {
		t.Fatal(err)
	}
	if paged.Year() != before.Year()+1 {
		t.Errorf("paging gave %v, want %d", paged, before.Year()+1)
	}

	year := day(2026, time.May, 1)
	yt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.YearPicker(c, datetime.YearPickerOptions{Year: &year, Today: fixed()})
	}, 460, 340)
	if err := yt.Click("2020"); err != nil {
		t.Fatal(err)
	}
	if year.Year() != 2020 {
		t.Errorf("choosing 2020 gave %v", year)
	}
	// A page arrow moves twelve years, because the page is twelve years wide.
	if err := yt.Click("Later years"); err != nil {
		t.Fatal(err)
	}
	if year.Year() != 2032 {
		t.Errorf("a page later gave %v, want 2032", year)
	}
}

func TestAvailabilityPickerReportsTheHour(t *testing.T) {
	today := day(2026, time.October, 7)
	taken := map[int]bool{9: true}
	var got time.Time
	ok := false
	marked := map[int]bool{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := datetime.AvailabilityPicker(c, datetime.AvailabilityOptions{
			Days: []time.Time{today}, StartHour: 8, EndHour: 12,
			Free:   func(_ time.Time, hour int) bool { return !taken[hour] },
			Marked: func(_ time.Time, hour int) bool { return marked[hour] },
			Toggle: func(_ time.Time, hour int) {
				marked[hour] = !marked[hour]
			},
		})
		if at, hit := r.Picked(); hit {
			got, ok = at, true
		}
	}, 400, 700)
	if err := tt.Click("10 2026-10-07"); err != nil {
		t.Fatal(err)
	}
	if !ok || got.Hour() != 10 {
		t.Errorf("pressing 10:00 reported %v", got)
	}
	if !marked[10] {
		t.Error("the press did not reach the caller's set of marked hours")
	}
	// The write is the caller's, so the next frame draws the hour as picked.
	if !tt.HasText("10 2026-10-07") {
		t.Error("the marked hour is missing from the next frame")
	}
}

func TestEventChipReportsAPress(t *testing.T) {
	now := fixed()
	pressed := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		if datetime.EventChip(c, datetime.EventChipOptions{
			Event: datetime.Event{Title: "Maple Street Bakery", Start: now},
		}).Pressed() {
			pressed = true
		}
	}, 420, 160)
	// The chip names the whole moment for a screen reader, and that name is
	// what a test aims at: the bare title is only part of what it says.
	if err := tt.Click("Wednesday 2026-10-07 14:30  Maple Street Bakery"); err != nil {
		t.Fatal(err)
	}
	if !pressed {
		t.Error("a pressed chip said nothing")
	}
}

func TestEventPopoverReportsEditAndDelete(t *testing.T) {
	now := fixed()
	edited, deleted := false, false
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		anchor := ui.Box(c).Label("job").Children(func() {
			ui.Text(c, "Oak Lane Cafe")
		})
		r := datetime.EventPopover(c, datetime.EventPopoverOptions{
			Anchor: anchor, Open: &open, Delete: "Delete", Edit: "Edit",
			Event: datetime.Event{Title: "Oak Lane Cafe", Start: now, End: now.Add(time.Hour)},
		})
		edited = r.Edited() || edited
		deleted = r.Deleted() || deleted
	}, 460, 320)
	// Both ends of the booking, the length between them, and the two things a
	// person can do about it.
	hasText(t, tt, "Oak Lane Cafe", "2026-10-07 14:30 – 2026-10-07 15:30", "1h", "Delete", "Edit")
	if err := tt.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	if !edited || deleted {
		t.Errorf("Edit gave edited %v deleted %v", edited, deleted)
	}
}

func TestGridRespectsItsBounds(t *testing.T) {
	now := fixed()
	var got time.Time
	ok := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		g := datetime.Grid(c, datetime.GridOptions{
			Month: now, Today: now,
			Min: day(2026, time.October, 10), Max: day(2026, time.October, 20),
		})
		if d, hit := g.Picked(); hit {
			got, ok = d, true
		}
	}, 320, 260)
	// A day before the range is out, and so is one after it.
	if err := tt.Click("Thursday 8 October 2026"); err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Errorf("a day before the range reported %v", got)
	}
	if err := tt.Click("Tuesday 13 October 2026"); err != nil {
		t.Fatal(err)
	}
	if !ok || got.Day() != 13 {
		t.Errorf("a day inside the range reported %v", got)
	}
}

func TestStopwatchStartsAndStops(t *testing.T) {
	now := fixed()
	start := time.Time{}
	pressed, running := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		p, r := datetime.Stopwatch(c, datetime.StopwatchOptions{
			Start: &start, Now: now, Laps: new([]time.Time),
		}).Started()
		pressed, running = pressed || p, running || r
	}, 420, 260)
	// Stopped: the reading is zero — a zero start is a clock that is not
	// running, not one that began when the calendar did — and the button
	// offers to start it.
	hasText(t, tt, "00:00.00", "Start")
	if err := tt.Click("Start"); err != nil {
		t.Fatal(err)
	}
	if !pressed || !running {
		t.Errorf("starting gave pressed %v running %v", pressed, running)
	}
	if !start.Equal(now) {
		t.Errorf("the clock started at %v, want the caller's now %v", start, now)
	}
	// A reset zeroes the caller's start, and the reading with it.
	if err := tt.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	if !start.IsZero() {
		t.Errorf("a reset left the start at %v", start)
	}
}

// Laps is optional — nil is what draws no lap button — so Reset, which is
// drawn either way, has to cope with the stopwatch that was given none.
func TestStopwatchResetWithoutALapListDoesNotPanic(t *testing.T) {
	now := fixed()
	start := now
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.Stopwatch(c, datetime.StopwatchOptions{Start: &start, Now: now})
	}, 420, 260)
	if err := tt.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	if !start.IsZero() {
		t.Errorf("a reset left the start at %v", start)
	}
}

func TestRecurrenceEditorWritesTheRule(t *testing.T) {
	rule := datetime.Recurrence{Frequency: datetime.Weekly}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.RecurrenceEditor(c, datetime.RecurrenceEditorOptions{
			Rule: &rule, Anchor: fixed(),
		})
	}, 700, 340)
	// A weekly rule is the one that offers days, and the days are Monday
	// first whatever day the rule is anchored on.
	if err := tt.Click("Thursday"); err != nil {
		t.Fatal(err)
	}
	if !rule.Weekdays[3] {
		t.Error("Thursday was not set")
	}
	if rule.Weekdays[0] {
		t.Error("Monday was set by a press on Thursday")
	}
	// And the summary says the whole thing, which is the point of it.
	hasText(t, tt, "Every week on Thu")
	// Choosing an interval writes it.
	if err := tt.Click("3"); err != nil {
		t.Fatal(err)
	}
	if rule.Interval != 3 {
		t.Errorf("the interval is %d, want 3", rule.Interval)
	}
}

func TestCronEditorSaysWhenARuleCannotBeRead(t *testing.T) {
	// A rule that does not parse is shown as it stands, and Err says why,
	// rather than being summarised into something that looks fine.
	expr := "0 9 * *"
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := datetime.CronEditor(c, datetime.CronEditorOptions{Expression: &expr})
		if r.Err() == nil {
			t.Error("an unreadable rule reported no error")
		}
	}, 700, 520)
	hasText(t, tt, "want 5 fields")

	// A readable one reports no error, and says what it means.
	expr = "0 9 * * 1-5"
	readable := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := datetime.CronEditor(c, datetime.CronEditorOptions{Expression: &expr})
		if r.Err() != nil {
			t.Errorf("a valid rule reported %v", r.Err())
		}
	}, 700, 520)
	hasText(t, readable, "On Monday, Tuesday, Wednesday, Thursday, Friday at 09:00")
}

func TestAgendaViewCountsWhatItDoesNotShow(t *testing.T) {
	day := day(2026, time.October, 7)
	events := []datetime.Event{
		{Title: "Maple Street Bakery", Start: day.Add(9 * time.Hour)},
		{Title: "Oak Lane Cafe", Start: day.Add(10 * time.Hour)},
		{Title: "Riverside Clinic", Start: day.Add(11 * time.Hour)},
	}
	tt := render(420, 400, func(c *ui.Context) {
		datetime.AgendaView(c, datetime.AgendaOptions{
			Today: fixed(), Limit: 2,
			Days: []datetime.AgendaDay{{Day: day, Events: events}},
		})
	})
	// Two drawn, one counted: a list that silently drops the third job is a
	// list a person cannot trust.
	hasText(t, tt, "Maple Street Bakery", "Oak Lane Cafe", "1 more")
	if tt.HasText("Riverside Clinic") {
		t.Error("the third job should be counted rather than drawn")
	}
}

func TestZoneLabelsUseTheMoment(t *testing.T) {
	// UTC needs no zone database, so this is the part that can be asserted
	// anywhere; a city is the same arithmetic with a database behind it.
	if got := datetime.ZoneLabel("UTC", fixed()); !strings.Contains(got, "UTC+00") {
		t.Errorf("UTC reads %q", got)
	}
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skipf("no zone database: %v", err)
	}
	// The offset is worked out for the moment shown, because a zone that is
	// one hour ahead in March is two ahead in July.
	winter, _ := datetime.ZoneOffset(berlin, time.Date(2026, time.January, 15, 12, 0, 0, 0, berlin))
	summer, _ := datetime.ZoneOffset(berlin, time.Date(2026, time.July, 15, 12, 0, 0, 0, berlin))
	if winter != 60 {
		t.Errorf("Berlin is UTC%+d in January, want +1", winter)
	}
	if summer != 120 {
		t.Errorf("Berlin is UTC%+d in July, want +2", summer)
	}
	// A zone the platform does not have falls back to UTC rather than
	// stopping: a machine that has never needed the files still runs.
	if datetime.LoadZone("Middle/Earth") != time.UTC {
		t.Error("an unknown zone should fall back to UTC")
	}
}

func TestTimeGridPicksHoursNotQuarters(t *testing.T) {
	// The grid divides an hour no finer than an hour, because a quarter of
	// an hour inside 56 DIP is four bands of fourteen — not a target anybody
	// can hit. The minutes belong to TimePicker, which has a control for
	// them, and the two are tested apart on purpose.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		datetime.CalendarTimeGrid(c, datetime.TimeGridOptions{
			Day: fixed(), StartHour: 9, EndHour: 10, Pickable: true,
		})
	}, 220, 120)
	if n := countText(tt, "09 2026-10-07"); n != 1 {
		t.Errorf("an hour row is drawn %d times, want once", n)
	}
}

func TestUntilNeverGoesBackwards(t *testing.T) {
	start := fixed()
	end := fixed().Add(90 * time.Minute)
	if got := datetime.Until(&start, &end); got != 90*time.Minute {
		t.Errorf("a range of ninety minutes is %v", got)
	}
	earlier := fixed().Add(-time.Hour)
	if got := datetime.Until(&end, &earlier); got != 0 {
		t.Errorf("a backwards range is %v, want zero", got)
	}
	if got := datetime.Until(&start, nil); got != 0 {
		t.Errorf("a range with no end is %v, want zero", got)
	}
}

func TestEventPopoverAndEditorRefuseToGuess(t *testing.T) {
	now := fixed()
	more := map[string]func(c *ui.Context){
		"CalendarDayView": func(c *ui.Context) {
			datetime.CalendarDayView(c, datetime.CalendarDayViewOptions{})
		},
		"CalendarWeekView": func(c *ui.Context) {
			datetime.CalendarWeekView(c, datetime.CalendarWeekViewOptions{})
		},
		"EventPopover": func(c *ui.Context) {
			open := false
			datetime.EventPopover(c, datetime.EventPopoverOptions{Open: &open,
				Event: datetime.Event{Title: "x", Start: now}})
		},
		"EventEditor": func(c *ui.Context) {
			e := datetime.Event{Title: "x", Start: now}
			datetime.EventEditor(c, datetime.EventEditorOptions{Event: &e, Save: "Save"})
		},
		"DateRangePicker": func(c *ui.Context) {
			from, open := now, false
			datetime.DateRangePicker(c, datetime.DateRangePickerOptions{
				Start: &from, Open: &open, Today: now,
			})
		},
		"DateTimePicker": func(c *ui.Context) {
			m := now
			datetime.DateTimePicker(c, datetime.DateTimePickerOptions{Value: &m})
		},
		"WeekPicker": func(c *ui.Context) {
			datetime.WeekPicker(c, datetime.WeekPickerOptions{Today: now})
		},
		"DurationPicker": func(c *ui.Context) {
			datetime.DurationPicker(c, datetime.DurationPickerOptions{})
		},
		"ReminderPicker": func(c *ui.Context) {
			datetime.ReminderPicker(c, datetime.ReminderPickerOptions{})
		},
		"RecurrenceEditor": func(c *ui.Context) {
			datetime.RecurrenceEditor(c, datetime.RecurrenceEditorOptions{})
		},
		"CronEditor": func(c *ui.Context) {
			datetime.CronEditor(c, datetime.CronEditorOptions{})
		},
		"Stopwatch": func(c *ui.Context) {
			datetime.Stopwatch(c, datetime.StopwatchOptions{Now: now})
		},
		"TimeRangePicker": func(c *ui.Context) {
			from := now
			datetime.TimeRangePicker(c, datetime.TimeRangePickerOptions{Start: &from})
		},
	}
	for name, body := range more {
		t.Run(name, func(t *testing.T) {
			defer func() {
				got := recover()
				if got == nil {
					t.Fatalf("%s drew itself with nothing to draw", name)
				}
				if msg, _ := got.(string); !strings.HasPrefix(msg, "datetime: ") {
					t.Errorf("%s panicked with %v", name, got)
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				body(c)
			}, 500, 400)
		})
	}
}

func TestCalendarEventsPutsTheDayOnTheClock(t *testing.T) {
	day := day(2026, time.October, 7)
	at := func(h, m int) time.Time { return day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
	tt := render(360, 700, func(c *ui.Context) {
		r := datetime.CalendarEvents(c, datetime.CalendarEventsOptions{
			Day: day, AllDay: true, StartHour: 8, EndHour: 18,
			Events: []datetime.Event{
				{Title: "Maple Street Bakery", Start: at(9, 0), End: at(10, 0)},
				{Title: "Oak Lane Cafe", Start: at(9, 30), End: at(10, 30)},
				{Title: "Riverside Clinic", Start: at(14, 0)},
				{Title: "Team stand-up", Start: day, AllDay: true},
				{Title: "Cancelled job", Start: at(16, 0), Cancelled: true},
			},
		})
		if r.Empty() {
			t.Error("a day with five events on it is not empty")
		}
	})
	hasText(t, tt, "Team stand-up", "09:00  Maple Street Bakery", "14:00  Riverside Clinic")

	// The position is the information: 09:00 sits an hour below the top of
	// the column and 14:00 six, on the same scale as the hours.
	early, ok := tt.Find("Wednesday 2026-10-07 09:00  Maple Street Bakery")
	if !ok {
		t.Fatal("no chip for the 09:00 job")
	}
	later, ok := tt.Find("Wednesday 2026-10-07 14:00  Riverside Clinic")
	if !ok {
		t.Fatal("no chip for the 14:00 job")
	}
	if mathAbs(float64(later.Y-early.Y-5*56)) > 2 {
		t.Errorf("the 14:00 job is %.0f below the 09:00 one, want about %d",
			later.Y-early.Y, 5*56)
	}
	// Two jobs at once are side by side: the 09:30 job goes into the second
	// lane, so a double booking looks like one rather than like a chip drawn
	// on top of another.
	over, ok := tt.Find("Wednesday 2026-10-07 09:30  Oak Lane Cafe")
	if !ok {
		t.Fatal("no chip for the 09:30 job")
	}
	if over.X-early.X < 100 {
		t.Errorf("the lanes are %.0f apart, want half a column", over.X-early.X)
	}
	// And the 14:00 job, which clashes with nothing, is back in the first lane
	// and the full width of it.
	if later.X != early.X {
		t.Errorf("a job that clashes with nothing is in lane %d, want the first",
			int(later.X-early.X))
	}
}

func TestTimezoneSelectWritesTheZoneName(t *testing.T) {
	// With nothing chosen it starts on the first zone, because a select with
	// no value selected is a select that says nothing.
	zone := ""
	tt := render(460, 120, func(c *ui.Context) {
		datetime.TimezoneSelect(c, datetime.TimezoneSelectOptions{
			Value: &zone, At: fixed(), ShowOffsets: true,
			Zones: []string{"UTC", "Asia/Tokyo"},
		})
	})
	if zone != "UTC" {
		t.Errorf("an empty zone is %q, want the first on offer", zone)
	}
	// The name that is saved is the zone's own, whatever the label says: a
	// window that shows "Asia/Tokyo · UTC+09:00" still has to store what a
	// scheduler understands.
	hasText(t, tt, "UTC · UTC+00")
}
