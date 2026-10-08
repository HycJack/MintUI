package project

import (
	"testing"
	"time"
)

// ── FormatIssueID ──────────────────────────────────────────────────────────

func TestFormatIssueID(t *testing.T) {
	cases := []struct {
		prefix string
		n      int
		want   string
	}{
		{"CB", 1042, "CB-1042"},
		{"cb", 1042, "CB-1042"}, // upper-cased: references are shouted
		{" cb ", 7, "CB-7"},
		{"", 1042, "1042"}, // no tracker is a bare number, not a dash
		{"CB", 0, "CB"},    // no number is the bare prefix
		{"", 0, ""},        // neither is nothing
		{"", -1, ""},
		{"CB", -1, "CB"},
	}
	for _, c := range cases {
		if got := FormatIssueID(c.prefix, c.n); got != c.want {
			t.Errorf("FormatIssueID(%q, %d) = %q, want %q", c.prefix, c.n, got, c.want)
		}
	}
}

func TestPrefixIssueID(t *testing.T) {
	cases := []struct{ in, want string }{
		{"CB", "CB-"}, {"cb", "CB-"}, {"", ""}, {"  ", ""},
	}
	for _, c := range cases {
		if got := PrefixIssueID(c.in); got != c.want {
			t.Errorf("PrefixIssueID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── Task.IDNumber ──────────────────────────────────────────────────────────

func TestIDNumberReadsTheTrailingDigits(t *testing.T) {
	cases := []struct {
		id   string
		want int
	}{
		{"CB-1042", 1042},
		{"1042", 1042},
		{"cb-7", 7},
		{"TASK-000123", 123},
		// A UUID ends in digits, so those are the ones read: a slug is a
		// better answer than a number invented from its letters.
		{"9f2a-77", 77},
		{"no-digits", 0},
		{"", 0},
		// Digits that are not at the end are part of the slug, not the number.
		{"12ab34", 34},
	}
	for _, c := range cases {
		if got := (Task{ID: c.id}).IDNumber(); got != c.want {
			t.Errorf("IDNumber(%q) = %d, want %d", c.id, got, c.want)
		}
	}
}

// ── severity ───────────────────────────────────────────────────────────────

func TestStatusSeverity(t *testing.T) {
	cases := []struct {
		s    Status
		want string
	}{
		{StatusBacklog, "Neutral"},
		{StatusTodo, "Neutral"},
		{StatusDoing, "Accent"},
		{StatusReview, "Accent"},
		{StatusBlocked, "Warning"},
		{StatusDone, "Success"},
		// A status this package has never heard of is drawn quietly, not as
		// an error: a team that calls a lane "Waiting on procurement" gets
		// a grey lane rather than a panic.
		{Status("waiting"), "Neutral"},
	}
	for _, c := range cases {
		if got := StatusSeverity(c.s).String(); got != c.want {
			t.Errorf("StatusSeverity(%q) = %s, want %s", c.s, got, c.want)
		}
	}
}

func TestSprintTone(t *testing.T) {
	cases := []struct {
		days int
		want string
	}{
		{10, "Neutral"}, {4, "Neutral"}, {3, "Warning"},
		{1, "Warning"}, {0, "Warning"}, {-1, "Danger"},
	}
	for _, c := range cases {
		if got := SprintTone(c.days).String(); got != c.want {
			t.Errorf("SprintTone(%d) = %s, want %s", c.days, got, c.want)
		}
	}
}

func TestMilestoneToneAndFraction(t *testing.T) {
	cases := []struct {
		done, total int
		wantSev     string
		wantFrac    float32
	}{
		{0, 10, "Neutral", 0},
		{5, 10, "Neutral", 0.5},
		{9, 10, "Accent", 0.9},
		{10, 10, "Success", 1},
		{12, 10, "Success", 1},
		{0, 0, "Neutral", 0},
		{3, 0, "Neutral", 0}, // an empty milestone is not a division
	}
	for _, c := range cases {
		if got := MilestoneTone(c.done, c.total).String(); got != c.wantSev {
			t.Errorf("MilestoneTone(%d, %d) = %s, want %s", c.done, c.total, got, c.wantSev)
		}
		if got := MilestoneFraction(c.done, c.total); got != c.wantFrac {
			t.Errorf("MilestoneFraction(%d, %d) = %v, want %v", c.done, c.total, got, c.wantFrac)
		}
	}
}

// ── due dates ──────────────────────────────────────────────────────────────

func TestDueTone(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	day := func(d int) DueValue {
		return DueValue{Day: now.AddDate(0, 0, d), Set: true, Text: "then"}
	}
	cases := []struct {
		name string
		due  DueValue
		want string
	}{
		{"no date at all", DueValue{}, "Neutral"},
		{"set but empty", DueValue{Set: true}, "Neutral"},
		{"far away", day(30), "Neutral"},
		{"tomorrow", day(1), "Warning"},
		{"today", day(0), "Danger"},
		{"yesterday", day(-1), "Danger"},
	}
	for _, c := range cases {
		if got := DueTone(c.due, now).String(); got != c.want {
			t.Errorf("%s: DueTone = %s, want %s", c.name, got, c.want)
		}
	}
}

func TestDueLabel(t *testing.T) {
	if got := DueLabel(DueValue{}); got != "" {
		t.Errorf("a task with no date has no words after its title, got %q", got)
	}
	if got := DueLabel(DueValue{Set: true, Text: "Friday"}); got != "Friday" {
		t.Errorf("DueLabel = %q", got)
	}
}

// ── FormatElapsed ──────────────────────────────────────────────────────────

func TestFormatElapsed(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{0, "0:00"},
		{30 * time.Second, "0:30"},
		{65 * time.Second, "1:05"},
		{59*time.Minute + 59*time.Second, "59:59"},
		{time.Hour, "1:00:00"},
		{65*time.Minute + 5*time.Second, "1:05:05"},
		// Past a day: hours and no wrap, because the hours are what somebody
		// is checking.
		{25 * time.Hour, "25:00:00"},
		// A negative duration is a clock that has not started; it reads zero
		// rather than counting backwards.
		{-90 * time.Second, "0:00"},
	}
	for _, c := range cases {
		if got := FormatElapsed(c.in); got != c.want {
			t.Errorf("FormatElapsed(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPad2(t *testing.T) {
	for _, c := range []struct {
		in   int
		want string
	}{{0, "00"}, {7, "07"}, {10, "10"}, {59, "59"}} {
		if got := pad2(c.in); got != c.want {
			t.Errorf("pad2(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── the burndown's maths ───────────────────────────────────────────────────

func TestIdealBurndownLandsOnZero(t *testing.T) {
	// The whole claim of an ideal line: it starts at the scope and finishes
	// at nothing, whatever the scope and however many days.
	for _, c := range []struct{ days, scope int }{
		{10, 50}, {1, 8}, {5, 100}, {14, 37},
	} {
		got := IdealBurndown(c.days, c.scope)
		if len(got) != c.days+1 {
			t.Fatalf("IdealBurndown(%d, %d) has %d points, want %d",
				c.days, c.scope, len(got), c.days+1)
		}
		if got[0] != float64(c.scope) {
			t.Errorf("IdealBurndown(%d, %d) starts at %v, want %d", c.days, c.scope, got[0], c.scope)
		}
		if got[c.days] != 0 {
			t.Errorf("IdealBurndown(%d, %d) ends at %v, want 0", c.days, c.scope, got[c.days])
		}
		// And it is a straight line: the same fall every day. Compared with a
		// tolerance rather than exactly, because 37 points over 14 days is
		// 2.6428571428571... and float division does not land on the same
		// last bit twice. The line is straight; the arithmetic behind it is
		// not exact, and pretending otherwise would be the wrong assertion.
		fall := got[0] - got[1]
		for i := 1; i < len(got); i++ {
			if d := got[i-1] - got[i]; absDiff(d, fall) > 1e-9 {
				t.Errorf("IdealBurndown(%d, %d) is not straight at %d: %v vs %v",
					c.days, c.scope, i, d, fall)
			}
		}
	}
}

func TestIdealBurndownOfOneDay(t *testing.T) {
	// A one-day sprint has no slope to speak of; the line is the scope and
	// then nothing, and it must not divide by the day count twice.
	got := IdealBurndown(1, 12)
	if len(got) != 2 || got[0] != 12 || got[1] != 0 {
		t.Errorf("IdealBurndown(1, 12) = %v, want [12 0]", got)
	}
}

func TestIdealBurndownOfNothing(t *testing.T) {
	if got := IdealBurndown(-1, 10); got != nil {
		t.Errorf("a negative day count is nothing, got %v", got)
	}
	// No scope means a flat line at zero rather than a division by zero.
	got := IdealBurndown(3, 0)
	if len(got) != 4 {
		t.Fatalf("IdealBurndown(3, 0) = %v", got)
	}
	for _, v := range got {
		if v != 0 {
			t.Errorf("IdealBurndown(3, 0) = %v, want all zeros", got)
		}
	}
}

func TestIdealBurndownNeverGoesNegative(t *testing.T) {
	// A scope that does not divide by the day count would otherwise put the
	// last few points below the floor.
	got := IdealBurndown(7, 10)
	for i, v := range got {
		if v < 0 {
			t.Errorf("point %d is %v", i, v)
		}
	}
}

func TestActualBurndownSkipsTheDaysWithNoReading(t *testing.T) {
	// A weekend is not a sprint that finished. Drawing the gap as zero is
	// the one claim a burndown chart must not make, so the readings are
	// passed through as they are and the axis carries the days.
	days := []BurndownDay{
		{Day: 0, Remaining: 50},
		{Day: 1, Remaining: 44},
		// Day 2 and 3 are a weekend: no readings at all.
		{Day: 4, Remaining: 30},
	}
	got := ActualBurndown(days, 4)
	want := []float64{50, 44, 30}
	if len(got) != len(want) {
		t.Fatalf("ActualBurndown = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ActualBurndown = %v, want %v", got, want)
		}
	}
}

func TestActualBurndownDropsReadingsOutsideTheSprint(t *testing.T) {
	days := []BurndownDay{
		{Day: -1, Remaining: 99},
		{Day: 0, Remaining: 50},
		{Day: 5, Remaining: 10},
		{Day: 9, Remaining: 1},
	}
	got := ActualBurndown(days, 5)
	want := []float64{50, 10}
	if len(got) != len(want) {
		t.Fatalf("ActualBurndown = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ActualBurndown = %v, want %v", got, want)
		}
	}
}

func TestActualBurndownOfNothing(t *testing.T) {
	if got := ActualBurndown(nil, 4); len(got) != 0 {
		t.Errorf("no readings is an empty line, got %v", got)
	}
	if got := ActualBurndown([]BurndownDay{{Day: 0}}, -1); got != nil {
		t.Errorf("a sprint that has not started is nothing, got %v", got)
	}
}

func TestSprintLengthTakesTheFurtherOfTodayAndTheLastReading(t *testing.T) {
	cases := []struct {
		name string
		b    Burndown
		want int
	}{
		{"today is the last day", Burndown{Today: 4, Days: []BurndownDay{{Day: 4}}}, 4},
		{"a reading past today", Burndown{Today: 2, Days: []BurndownDay{{Day: 5}}}, 5},
		{"today past the last reading", Burndown{Today: 6, Days: []BurndownDay{{Day: 3}}}, 6},
		{"nothing at all", Burndown{Today: -1}, 0},
		{"before it started", Burndown{Today: -1, Days: []BurndownDay{{Day: -1}}}, 0},
	}
	for _, c := range cases {
		if got := SprintLength(c.b); got != c.want {
			t.Errorf("%s: SprintLength = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestPointsFromIndexesFromZero(t *testing.T) {
	got := pointsFrom([]float64{50, 44, 30})
	if len(got) != 3 {
		t.Fatalf("pointsFrom = %v", got)
	}
	for i, p := range got {
		if p.X != float64(i) {
			t.Errorf("point %d has x = %v, want %d", i, p.X, i)
		}
	}
}

func TestSequence(t *testing.T) {
	got := sequence(4)
	want := []float64{0, 1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("sequence(4) = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sequence(4) = %v, want %v", got, want)
		}
	}
}

// ── the pomodoro ───────────────────────────────────────────────────────────

func TestPomodoroAtTheStartOfFocus(t *testing.T) {
	got := PomodoroAt(0, 25, 5)
	if got.Phase != Focus {
		t.Error("a fresh timer is focusing, not resting")
	}
	if got.Left != 25*time.Minute {
		t.Errorf("left = %v, want the whole focus", got.Left)
	}
	if got.Fraction != 0 {
		t.Errorf("fraction = %v, want 0", got.Fraction)
	}
	if got.Done {
		t.Error("a full phase is not done")
	}
}

func TestPomodoroAtTheTurnBetweenFocusAndBreak(t *testing.T) {
	// The boundary is where two implementations disagree, so it is checked
	// from both sides.
	at25 := PomodoroAt(25*time.Minute, 25, 5)
	if at25.Phase != Break {
		t.Errorf("the twenty-fifth minute is the break, got %v", at25.Phase)
	}
	if at25.Left != 5*time.Minute {
		t.Errorf("the break starts full, got %v", at25.Left)
	}
	if at25.Fraction != 0 {
		t.Errorf("the break starts at nothing gone, got %v", at25.Fraction)
	}

	before := PomodoroAt(25*time.Minute-time.Second, 25, 5)
	if before.Phase != Focus {
		t.Error("a second before the turn is still focus")
	}
	if before.Left != time.Second {
		t.Errorf("left = %v, want a second", before.Left)
	}
	if before.Fraction < 0.999 {
		t.Errorf("fraction = %v, want nearly all of it gone", before.Fraction)
	}
}

func TestPomodoroWraps(t *testing.T) {
	// Left running overnight, a timer comes back round to the start rather
	// than sitting at zero, which would be indistinguishable from a broken
	// one.
	cycle := 30 * time.Minute
	got := PomodoroAt(cycle+5*time.Minute, 25, 5)
	if got.Phase != Focus {
		t.Errorf("after one whole cycle it is focusing again, got %v", got.Phase)
	}
	if got.Left != 20*time.Minute {
		t.Errorf("left = %v, want 20 minutes", got.Left)
	}
}

func TestPomodoroTakesTheDefaults(t *testing.T) {
	// Zero lengths fall back rather than dividing by zero, so a caller that
	// has not configured the timer gets a working one.
	got := PomodoroAt(0, 0, 0)
	if got.Left != DefaultFocusMinutes*time.Minute {
		t.Errorf("left = %v, want the %d minute default", got.Left, DefaultFocusMinutes)
	}
}

func TestPomodoroFractionNeverLeavesItsRange(t *testing.T) {
	for _, d := range []time.Duration{
		-time.Hour, 0, time.Second, 25 * time.Minute,
		30 * time.Minute, 29 * time.Hour, 37 * time.Minute,
	} {
		got := PomodoroAt(d, 25, 5)
		if got.Fraction < 0 || got.Fraction > 1 {
			t.Errorf("at %v the fraction is %v", d, got.Fraction)
		}
		if got.Left < 0 {
			t.Errorf("at %v the time left is %v", d, got.Left)
		}
	}
}

func TestPomodoroOfACustomShape(t *testing.T) {
	// Fifty and ten, so the phase boundaries are somewhere else and the code
	// cannot have been written against twenty-five and five alone.
	at50 := PomodoroAt(50*time.Minute, 50, 10)
	if at50.Phase != Break || at50.Left != 10*time.Minute {
		t.Errorf("at fifty minutes: phase %v, left %v", at50.Phase, at50.Left)
	}
	// Fifty and ten make a sixty minute cycle, so 55 minutes is five minutes
	// into the break rather than into a fresh focus.
	at55 := PomodoroAt(55*time.Minute, 50, 10)
	if at55.Phase != Break || at55.Left != 5*time.Minute {
		t.Errorf("fifty-five minutes in: phase %v, left %v", at55.Phase, at55.Left)
	}
	// And sixty is the turn back round to the beginning.
	at60 := PomodoroAt(60*time.Minute, 50, 10)
	if at60.Phase != Focus || at60.Left != 50*time.Minute {
		t.Errorf("after a full cycle: phase %v, left %v", at60.Phase, at60.Left)
	}
}

// ── the columns ────────────────────────────────────────────────────────────

func TestColumnLabelFallsBackToItsKey(t *testing.T) {
	if got := (Column{Status: StatusReview}).label(); got != "review" {
		t.Errorf("an untitled column is called by its key, got %q", got)
	}
	if got := (Column{Status: StatusReview, Title: "In review"}).label(); got != "In review" {
		t.Errorf("a titled column keeps its title, got %q", got)
	}
}

func TestTasksInKeepsTheCallersOrder(t *testing.T) {
	tasks := []Task{
		{ID: "a", Status: StatusTodo},
		{ID: "b", Status: StatusDoing},
		{ID: "c", Status: StatusTodo},
	}
	got := tasksIn(tasks, StatusTodo)
	if len(got) != 2 || got[0].ID != "a" || got[1].ID != "c" {
		t.Errorf("tasksIn = %v, want a and c in that order", got)
	}
	if len(tasksIn(tasks, Status("nope"))) != 0 {
		t.Error("a lane with nothing in it is empty")
	}
}

func TestTaskDone(t *testing.T) {
	if !(Task{Status: StatusDone}).Done() {
		t.Error("a task in Done is done")
	}
	for _, s := range []Status{StatusTodo, StatusDoing, StatusBlocked, StatusBacklog} {
		if (Task{Status: s}).Done() {
			t.Errorf("%s is not done", s)
		}
	}
}

// ── helpers ────────────────────────────────────────────────────────────────

// absDiff is how far two computed slopes are from each other.
func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

func TestItoa(t *testing.T) {
	for _, c := range []struct {
		in   int
		want string
	}{{0, "0"}, {7, "7"}, {42, "42"}, {1042, "1042"}, {1000000, "1000000"}} {
		if got := itoa(c.in); got != c.want {
			t.Errorf("itoa(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
