package datetime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// The layouts every date in the library is written with.
//
// One set, used by every component and available to every caller: a day shown
// in a picker and the same day in a table row have to be the same string, or
// the two stop reading as the same date. They are ISO — 2026-10-07 — and
// 24-hour — 14:30 — because a calendar column that mixes "2:30 PM" and "14:30"
// is a column nobody can sort by eye, and because neither depends on a locale
// the package does not know about.
const (
	// DateLayout is a plain day: 2026-10-07.
	DateLayout = "2006-01-02"
	// TimeLayout is a wall clock: 14:30.
	TimeLayout = "15:04"
	// TimeSecondsLayout is a wall clock with seconds: 14:30:07. A stopwatch
	// needs them; nothing else does.
	TimeSecondsLayout = "15:04:05"
	// DateTimeLayout is both: 2026-10-07 14:30.
	DateTimeLayout = DateLayout + " " + TimeLayout
	// MonthLayout names a month and a year: 2026-10.
	MonthLayout = "2006-01"
	// TimezoneLayout is a zone abbreviation: CEST.
	TimezoneLayout = "MST"
)

// FormatDate writes a day as 2026-10-07.
func FormatDate(t time.Time) string { return t.Format(DateLayout) }

// FormatTime writes a wall clock as 14:30.
func FormatTime(t time.Time) string { return t.Format(TimeLayout) }

// FormatSeconds writes a wall clock as 14:30:07.
func FormatSeconds(t time.Time) string { return t.Format(TimeSecondsLayout) }

// FormatDateTime writes a day and a clock as 2026-10-07 14:30.
func FormatDateTime(t time.Time) string { return t.Format(DateTimeLayout) }

// ParseDate reads a day written by FormatDate.
func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("datetime: %q is not a date (want 2006-01-02)", s)
	}
	return t, nil
}

// ParseTime reads a clock written by FormatTime. The day is zero, because a
// time of day says nothing about a date and a caller pairing it with one
// should not have to notice that the zero day is there.
func ParseTime(s string) (time.Time, error) {
	t, err := time.Parse(TimeLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("datetime: %q is not a time (want 15:04)", s)
	}
	return t, nil
}

// ParseDateTime reads a day and a clock written by FormatDateTime.
func ParseDateTime(s string) (time.Time, error) {
	t, err := time.Parse(DateTimeLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("datetime: %q is not a date and time (want 2006-01-02 15:04)", s)
	}
	return t, nil
}

// ── calendar arithmetic ──────────────────────────────────────────────────────
//
// Everything below answers a question a person asks rather than a machine:
// which week is this in, how many days does February have, what is the Monday
// of this week. The answers are computed, never read off a clock.

// IsLeapYear reports whether y has a 29th of February.
//
// The rule is the one everybody half-remembers, so it is written out: every
// fourth year, except every hundredth, except every four hundredth. 2024 and
// 2000 are leap years; 2100 and 2026 are not.
func IsLeapYear(y int) bool {
	return y%4 == 0 && (y%100 != 0 || y%400 == 0)
}

// DaysInMonth returns how many days the month of t has, 28, 29, 30 or 31.
//
// February is the only month this matters for, and it is the only month the
// rule is about: 29 in a leap year, 28 otherwise.
func DaysInMonth(t time.Time) int {
	switch t.Month() {
	case time.January, time.March, time.May, time.July,
		time.August, time.October, time.December:
		return 31
	case time.April, time.June, time.September, time.November:
		return 30
	}
	if IsLeapYear(t.Year()) {
		return 29
	}
	return 28
}

// Weekday returns t's column in a seven-column week, Monday as 0 and Sunday
// as 6.
//
// It is the whole reason this package can talk about weeks at all:
// [time.Time.Weekday] counts from Sunday, so using it directly puts Sunday in
// the first column of every grid in the library — which is not what "the week
// starts on Monday" means.
func Weekday(t time.Time) int { return (int(t.Weekday()) + 6) % 7 }

// StartOfDay returns midnight at the start of t's day, in t's own location.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the last representable moment of t's day, for a range test
// that is written as after(start) and before(end).
func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

// StartOfWeek returns the Monday of t's week, at midnight.
func StartOfWeek(t time.Time) time.Time {
	d := StartOfDay(t)
	return d.AddDate(0, 0, -Weekday(d))
}

// StartOfMonth returns the first of t's month at midnight.
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// StartOfYear returns the first of t's year at midnight.
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// AddDays moves t by n days as calendar days, keeping the wall clock.
//
// AddDate, not Add: on a day-saving boundary Add(48*time.Hour) can land on the
// same wall clock twice or skip it, and a calendar that draws the same day in
// two columns has a bug that only appears twice a year.
func AddDays(t time.Time, n int) time.Time { return t.AddDate(0, 0, n) }

// AddMonths moves t by n months, clamping the day to the month's length.
//
// 31 January plus one month is 28 February, not 3 March: a picker stepping
// through months should stop on the last day of a short one rather than skip
// over it. Go's AddDate would overflow into the next month instead, which is
// how a month grid ends up drawing March twice.
func AddMonths(t time.Time, n int) time.Time {
	first := time.Date(t.Year(), t.Month(), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	shifted := first.AddDate(0, n, 0)
	day := min(t.Day(), DaysInMonth(shifted))
	return time.Date(shifted.Year(), shifted.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// SameDay reports whether a and b are the same calendar day in their own
// locations. Two instants on the same day in two zones are not the same day.
func SameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// SameMonth reports whether a and b are in the same month of the same year.
func SameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}

// SameWeek reports whether a and b fall in the same Monday-to-Sunday week.
func SameWeek(a, b time.Time) bool { return StartOfWeek(a).Equal(StartOfWeek(b)) }

// DaysBetween returns the whole calendar days from a to b: negative when b is
// before a, and counting nights crossed rather than dividing two instants.
//
// "Three days ago" is three nights, so a day either side of a day-saving
// boundary is still one day. Dividing the two instants by 24 hours would say
// two, and the countdown under a deadline would disagree with the person
// reading it.
func DaysBetween(a, b time.Time) int {
	return dayNumber(b) - dayNumber(a)
}

// dayNumber counts whole days from a fixed epoch for a time's calendar date.
// It is the civil-date to day-number conversion, which is arithmetic rather
// than a subtraction so that a location, a zone offset and the length of the
// months cannot get into it.
func dayNumber(t time.Time) int {
	y, m, d := t.Year(), int(t.Month()), t.Day()
	// Howard Hinnant's days_from_civil: shift the year so that March is the
	// first month, which puts the leap day at the end and makes the month
	// lengths after February irrelevant.
	if m <= 2 {
		y--
		m += 12
	}
	era := floorDiv(y, 400)
	yoe := y - era*400
	doy := (153*(m-3)+2)/5 + d - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// WeekOf returns the ISO 8601 year and week number of t, where week 1 is the
// week holding 4 January — which is the same as the week holding the first
// Thursday, and is the rule that makes a week belong to the year of most of
// its days.
//
// It is the standard's rule and not an invented one because week numbers are
// shown next to real bookings: a year with 53 weeks and a 31 December that
// belongs to week 1 of the next are both ordinary, and a grid that disagrees
// with the booking system about them is worse than one that shows no number.
func WeekOf(t time.Time) (year, week int) { return t.ISOWeek() }

// WeekStart returns the Monday of the ISO week number (year, week).
func WeekStart(year, week int) time.Time {
	// 4 January is always in week 1, so the Monday of week 1 is within three
	// days before it.
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	monday := StartOfWeek(jan4)
	return monday.AddDate(0, 0, (week-1)*7)
}

// WeekOfYear returns the ISO week number of t alone.
func WeekOfYear(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// InRange reports whether t falls in [start, end], with a zero bound meaning
// "open". It is the one range test the pickers use, so a caller who passes one
// bound and leaves the other zero gets the half-open range they meant.
func InRange(t, start, end time.Time) bool {
	if !start.IsZero() && t.Before(start) {
		return false
	}
	if !end.IsZero() && t.After(end) {
		return false
	}
	return true
}

// ── durations ────────────────────────────────────────────────────────────────

// DurationText writes a length the way a person says it: the two largest
// units that are not zero, such as "1h 30m", "2d 4h" or "45s".
//
// Below a second it counts in milliseconds, microseconds and nanoseconds
// rather than rounding, because a rounded answer is a different length — and
// rounding far enough left a length under a millisecond with no unit to be
// counted in, which came out as an empty string rather than as nothing at all.
// The table ends at the nanosecond because that is the smallest duration there
// is, so every length has an answer.
//
// Two units is the limit because a booking form has room for "1d 2h 30m" and
// nothing for the rest. Zero is "0m" rather than an empty string, because an
// empty duration field reads as a missing value rather than as no time at all.
// A negative length keeps its sign in front, as "-30m", and it keeps a length
// behind the sign: "-1ns", never a sign on its own.
func DurationText(d time.Duration) string {
	neg := d < 0
	// The magnitude, in unsigned arithmetic, and never a recursion on the
	// negated value: the most negative duration negates back to itself, so
	// this recursed on its own answer until the stack overflowed, and a stack
	// overflow is not a panic recover() can catch. Unsigned is also the only
	// arithmetic that holds that one length, which int64 cannot as a
	// positive number.
	left := uint64(d)
	if neg {
		left = uint64(-(d + 1)) + 1
	}
	if left == 0 {
		return "0m"
	}
	units := []struct {
		size   time.Duration
		suffix string
	}{
		{24 * time.Hour, "d"}, {time.Hour, "h"}, {time.Minute, "m"},
		{time.Second, "s"}, {time.Millisecond, "ms"},
		// The same U+00B5 micro sign Go's own Duration writes, so a length
		// reads the same here as it does in a log line or a test failure.
		{time.Microsecond, "µs"}, {time.Nanosecond, "ns"},
	}
	var parts []string
	for _, u := range units {
		if n := left / uint64(u.size); n > 0 {
			parts = append(parts, strconv.FormatUint(n, 10)+u.suffix)
			left -= n * uint64(u.size)
			if len(parts) == 2 {
				break
			}
		}
	}
	out := strings.Join(parts, " ")
	if neg {
		return "-" + out
	}
	return out
}

// ParseDurationText reads what DurationText wrote, in the same vocabulary.
func ParseDurationText(s string) (time.Duration, error) {
	neg := false
	if strings.HasPrefix(s, "-") {
		neg, s = true, s[1:]
	}
	if s == "" {
		return 0, errors.New("datetime: no duration")
	}
	var total time.Duration
	for _, part := range strings.Fields(s) {
		// The unit is taken off the end by what it is rather than by its last
		// byte. "ms", "µs" and "ns" are two letters where every other unit is
		// one, and "ns" ends in "s": read by its last byte, "500ns" is five
		// hundred seconds.
		num, unit := "", ""
		for _, u := range [...]string{"ms", "µs", "ns"} {
			if strings.HasSuffix(part, u) {
				num, unit = part[:len(part)-len(u)], u
				break
			}
		}
		if num == "" && len(part) > 1 {
			num, unit = part[:len(part)-1], part[len(part)-1:]
		}
		if num == "" {
			return 0, fmt.Errorf("datetime: %q is not a duration", s)
		}
		n, err := strconv.Atoi(num)
		if err != nil {
			return 0, fmt.Errorf("datetime: %q is not a duration", s)
		}
		switch unit {
		case "ms":
			total += time.Duration(n) * time.Millisecond
		case "µs":
			total += time.Duration(n) * time.Microsecond
		case "ns":
			total += time.Duration(n) * time.Nanosecond
		case "s":
			total += time.Duration(n) * time.Second
		case "m":
			total += time.Duration(n) * time.Minute
		case "h":
			total += time.Duration(n) * time.Hour
		case "d":
			total += time.Duration(n) * 24 * time.Hour
		default:
			return 0, fmt.Errorf("datetime: %q is not a duration", s)
		}
	}
	if neg {
		total = -total
	}
	return total, nil
}

// ── relative time ────────────────────────────────────────────────────────────

// RelativeSpan is how far one instant is from another, in the units a reader
// names them in.
//
// It is a value rather than a string on purpose. "3 days ago", "in 3 days" and
// "vor 3 Tagen" are the same measurement written three ways, and which one
// appears is the window's business — the component that knows the difference
// has no business choosing the wording. The fields are the largest whole
// units first, so Days is what is left after Years and Months, not the total.
type RelativeSpan struct {
	// Past reports that the earlier instant came first, which is the half of
	// the answer a sentence has to get right ("ago" versus "in").
	Past bool
	// Years, Months, Days, Hours, Minutes and Seconds are whole calendar
	// units, largest first.
	Years, Months, Days, Hours, Minutes, Seconds int
	// TotalDays is the whole days between the two, ignoring the clock. It is
	// what a picker disables against — "no bookings in the next 30 days".
	TotalDays int
}

// NonZero returns the largest non-zero unit as a count and a name, and whether
// there was one. A caller formatting a span needs the same answer for every
// language it might be in, so the choice of unit is made here and the wording
// is not.
func (s RelativeSpan) NonZero() (n int, unit string, ok bool) {
	for _, f := range []struct {
		n    int
		unit string
	}{
		{s.Years, "year"}, {s.Months, "month"}, {s.Days, "day"},
		{s.Hours, "hour"}, {s.Minutes, "minute"}, {s.Seconds, "second"},
	} {
		if f.n != 0 {
			return f.n, f.unit, true
		}
	}
	return 0, "", false
}

// Relative measures then against now.
//
// The units are calendar units, and the calendar is now's: "3 days ago" counts
// three nights in the location the reader is in, and a timestamp in another
// zone has to be read before it can be counted. Years and months are found by
// adding them until they would overshoot rather than by dividing, because a
// month is not a number of days and 2026-01-31 is one month after 2025-12-31
// by anyone's reckoning but zero months by a 30.44-day average.
func Relative(then, now time.Time) RelativeSpan {
	later, earlier := then, now
	past := then.Before(now)
	if past {
		later, earlier = now, then
	}

	// Both read in one location, the earlier one's: a calendar day is a day
	// in somebody's zone, and a timestamp arriving from a booking server in
	// another one has to be read before it can be counted.
	later, earlier = later.In(earlier.Location()), earlier.In(earlier.Location())

	// Whole years, then whole months, each found by adding one and asking
	// whether it has overshot. Dividing by 365.25 and by 30.44 would be
	// shorter and wrong at both ends: it makes 31 January 2026 one month
	// after 31 December 2025 into zero months and thirty-one days, and it
	// makes 29 February into a year that is still a day short.
	years := 0
	for guess := later.Year() - earlier.Year(); guess > 0; guess-- {
		if !AddMonths(earlier, guess*12).After(later) {
			years = guess
			break
		}
	}
	months := 0
	for m := 1; m < 12; m++ {
		if AddMonths(earlier, years*12+m).After(later) {
			break
		}
		months = m
	}

	// What is left after the whole months is a number of days and a wall
	// clock, which are the two things a calendar is made of.
	cursor := AddMonths(earlier, years*12+months)
	days := dayNumber(later) - dayNumber(cursor)

	// The remainder is measured from the cursor rather than from midnight,
	// because when nothing but minutes separates the two instants the whole
	// of the difference is in the clock: measured from midnight, two moments
	// that are the same would report the hour they happen to be in.
	rest := later.Sub(cursor) - time.Duration(days)*24*time.Hour
	// A day-saving boundary in between moves the two by an hour, so the
	// remainder is clamped back inside one day rather than allowed to report
	// 25 hours or minus one.
	rest = min(max(rest, 0), 24*time.Hour-time.Nanosecond)

	return RelativeSpan{
		Past:      past,
		Years:     years,
		Months:    months,
		Days:      days,
		Hours:     int(rest / time.Hour),
		Minutes:   int(rest%time.Hour) / int(time.Minute),
		Seconds:   int(rest%time.Minute) / int(time.Second),
		TotalDays: dayNumber(later) - dayNumber(earlier),
	}
}

// RelativeText is the library's own wording for a span: "3 days ago", "in 2
// hours", "just now".
//
// It exists so a window has something to show before it has decided on a
// language, and so the tests have a fixed sentence to assert. A window that
// has its own wording passes a function to RelativeTime instead, and this one
// is never called.
func RelativeText(s RelativeSpan) string {
	n, unit, ok := s.NonZero()
	if !ok {
		return "just now"
	}
	// Under three quarters of a minute reads as now rather than as "43
	// seconds ago", which is noise in a list of recent activity.
	if unit == "second" && n < 45 {
		return "just now"
	}
	if n != 1 {
		unit += "s"
	}
	if s.Past {
		return fmt.Sprintf("%d %s ago", n, unit)
	}
	return fmt.Sprintf("in %d %s", n, unit)
}
