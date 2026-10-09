package datetime

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── recurrence ───────────────────────────────────────────────────────────────

// Frequency is how often a recurring event happens.
type Frequency int

const (
	// Never is the zero value: an event that happens once. It is the zero so
	// that a zero Recurrence is a sensible event rather than a daily one.
	Never Frequency = iota
	// Daily.
	Daily
	// Weekly.
	Weekly
	// Monthly.
	Monthly
	// Yearly.
	Yearly
	// Custom is a rule this component does not draw: a cron, or a set of
	// dates the caller holds. It is a frequency of its own so that "every
	// other Tuesday except in December" has somewhere to be.
	Custom
)

func (f Frequency) String() string {
	switch f {
	case Daily:
		return "Day"
	case Weekly:
		return "Week"
	case Monthly:
		return "Month"
	case Yearly:
		return "Year"
	case Custom:
		return "Custom"
	}
	return "Never"
}

// Recurrence is a rule for an event that happens more than once.
//
// It is the caller's value, written by RecurrenceEditor, and read by
// RecurrenceSummary to say it in a sentence. A recurring booking that cannot
// be printed as a sentence cannot be checked by the person reading it, and an
// unchecked recurrence is how a technician ends up turning up on the wrong
// Tuesday for a month.
type Recurrence struct {
	// Frequency is how often it happens.
	Frequency Frequency
	// Interval is how many of those units between times; 1 is every one of
	// them. Zero is read as 1, because a rule that means "every zero weeks"
	// means nothing.
	Interval int
	// Weekdays are the days a weekly rule lands on, Monday first. They are a
	// set of seven flags rather than a list, because a set has one obvious
	// answer to "is Tuesday in it" and a list does not.
	Weekdays [7]bool
	// Count is how many times it happens, and Until is when it stops. Zero in
	// both means it goes on for ever, which is the common case and the one
	// that has to be visible in the summary rather than implied.
	Count int
	Until time.Time
	// Custom is the rule a Custom frequency holds — a cron expression, say —
	// and is shown as it stands rather than interpreted.
	Custom string
}

// step returns the interval, never zero.
func (r Recurrence) step() int {
	if r.Interval < 1 {
		return 1
	}
	return r.Interval
}

// RecurrenceSummary writes a rule as a person would say it: "Every 2 weeks on
// Mon, Wed until 31 Dec 2026", or "Does not repeat".
//
// The wording is English, and deliberately so: it is a summary for a caller to
// read before saving, and a window that wants it in another language passes
// its own summary function to the editor rather than having this one guessed
// at. What is not a matter of taste is that the sentence must contain the
// interval, the days and the end — a rule summarised as "weekly" is one that
// cannot be checked.
func RecurrenceSummary(r Recurrence) string {
	if r.Frequency == Never {
		return core.Def("Does not repeat")
	}
	every := "Every "
	if n := r.step(); n > 1 {
		every += itoa(n) + " "
	}
	// Lower case, because it is a common noun inside a sentence and not a
	// column heading: "Every 2 weeks", not "Every 2 Weeks".
	unit := strings.ToLower(r.Frequency.String()) + "s"
	if r.step() == 1 {
		unit = strings.ToLower(r.Frequency.String())
	}
	out := every + unit

	if r.Frequency == Weekly {
		if days := weekdayList(r.Weekdays); days != "" {
			out += " on " + days
		}
	}
	out += r.endText()
	return out
}

func (r Recurrence) endText() string {
	switch {
	case r.Count > 0:
		if r.Count == 1 {
			return core.Def(", once")
		}
		return core.Def(" for ") + itoa(r.Count) + core.Def(" times")
	case !r.Until.IsZero():
		return core.Def(", until ") + FormatDate(r.Until)
	}
	return core.Def("")
}

// weekdayList writes the days of a weekly rule: "Mon, Wed, Fri".
func weekdayList(days [7]bool) string {
	names := DefaultWeekdays()
	var parts []string
	for i, on := range days {
		if on {
			parts = append(parts, names[i])
		}
	}
	return strings.Join(parts, ", ")
}

// RecurrenceEditorOptions configure a RecurrenceEditor.
type RecurrenceEditorOptions struct {
	// Rule is the rule being edited, and is written as the controls change.
	// Required.
	Rule *Recurrence
	// Anchor is the event the rule repeats from: the days a weekly rule offers
	// are the days between here and six days later. The zero time means
	// today, as far as the component can tell — which is nothing, so it means
	// no days are offered.
	Anchor time.Time
	// Name is what a screen reader announces.
	Name string
	// Frequencies are the frequencies on offer; nil uses all of them.
	Frequencies []Frequency
	// CustomLabel names the Custom frequency's field; empty uses a default.
	CustomLabel string
	// Intervals are the intervals on offer; nil means 1, 2, 3, 6 and 12.
	Intervals []int
}

// RecurrenceEditorResult carries a RecurrenceEditor and whether it changed.
type RecurrenceEditorResult struct {
	// Element is the control.
	Element *ui.Element
	// changed reports that the rule moved this frame.
	changed bool
}

// Changed reports that the rule moved this frame.
func (r RecurrenceEditorResult) Changed() bool { return r.changed }

// RecurrenceEditor is the rule a recurring event follows: how often, every how
// many, on which days, and until when.
//
// The days of the week are offered only for a weekly rule, because they are
// the only frequency where they mean anything, and a control that shows seven
// toggles for a daily rule is seven controls that do nothing.
func RecurrenceEditor(c *ui.Context, opts RecurrenceEditorOptions) RecurrenceEditorResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Rule == nil {
		panic("datetime: RecurrenceEditor needs a Rule to point at")
	}
	rule := *opts.Rule
	frequencies := opts.Frequencies
	if len(frequencies) == 0 {
		frequencies = []Frequency{Never, Daily, Weekly, Monthly, Yearly, Custom}
	}
	intervals := opts.Intervals
	if len(intervals) == 0 {
		intervals = []int{1, 2, 3, 6, 12}
	}
	name := opts.Name
	if name == "" {
		name = core.Msg(c, "datetime.repeat", "Repeat")
	}
	custom := opts.CustomLabel
	if custom == "" {
		custom = core.Msg(c, "datetime.cronRule", "Rule")
	}

	var r RecurrenceEditorResult
	r.Element = ui.Column(c).FillWidth().Gap(u * 1.5).Label(name).Children(func() {
		ui.Row(c).FillWidth().Gap(u * 1.5).Children(func() {
			ui.Text(c, core.Msg(c, "datetime.repeats", "Repeats")).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
			// Frequency is a row of buttons rather than a menu: there are six
			// of them and they are the whole point of the control, and a
			// person choosing "weekly" should not have to open something to
			// find out that weekly is in there.
			for _, f := range frequencies {
				f := f
				on := rule.Frequency == f
				bg, fg := k.Surface, k.Text
				if on {
					bg, fg = k.Fill, k.OnFill
				}
				label := f.String()
				btn := ui.Button(c, "").Grow(1).Height(core.ControlHeight(c)).
					Radius(theme.PillRadius).Background(bg).TextColor(fg).
					Label(name + " " + label).Children(func() {
					ui.Text(c, label).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				})
				if btn.Clicked() {
					rule.Frequency = f
					r.changed = true
				}
			}
		})
		if rule.Frequency == Never {
			return
		}
		if len(intervals) > 1 {
			ui.Row(c).FillWidth().Gap(u * 1.5).Children(func() {
				ui.Text(c, core.Msg(c, "datetime.every", "Every")).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				for _, n := range intervals {
					n := n
					on := rule.step() == n
					bg, fg := k.Surface, k.Text
					if on {
						bg, fg = k.AccentBg, k.AccentText
					}
					btn := ui.Button(c, "").Grow(1).Height(u * 8).
						Radius(theme.SmallRadius).Background(bg).TextColor(fg).
						Label(itoa(n)).Children(func() {
						ui.Text(c, itoa(n)).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize))
					})
					if btn.Clicked() {
						rule.Interval = n
						r.changed = true
					}
				}
			})
		}
		if rule.Frequency == Weekly {
			// Monday first, because a week here is Monday to Sunday and the
			// toggles have to read in that order.
			ui.Row(c).FillWidth().Gap(u).Children(func() {
				names := DefaultWeekdays()
				for i := range 7 {
					i := i
					on := rule.Weekdays[i]
					bg, fg := k.Surface, k.Text
					if on {
						bg, fg = k.Fill, k.OnFill
					}
					btn := ui.Button(c, "").Grow(1).MinWidth(MinCellSize).Height(u * 8).
						Radius(theme.SmallRadius).Background(bg).TextColor(fg).
						Label(WeekdayFullNames[i]).
						Children(func() {
							ui.Text(c, names[i]).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
						})
					if btn.Clicked() {
						rule.Weekdays[i] = !rule.Weekdays[i]
						r.changed = true
					}
				}
			})
		}
		if rule.Frequency == Custom {
			ui.Box(c).FillWidth().Radius(theme.ControlRadius).Background(k.Surface).
				Padding(u, u*2).Label(custom).Children(func() {
				ui.Text(c, rule.Custom).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).FontFeatures("tnum").SingleLine()
			})
		}
		// The sentence is drawn under the controls, always. It is the only
		// place the whole rule is visible at once, and a rule that cannot be
		// read back is a rule nobody checked.
		ui.Text(c, RecurrenceSummary(rule)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
	*opts.Rule = rule
	return r
}

// ── cron ─────────────────────────────────────────────────────────────────────

// CronFields are the five fields of a cron expression, each the numbers it
// matches, in order: minute, hour, day of month, month, day of week.
type CronFields struct {
	Minute, Hour, Day, Month, Weekday []int
	// Any marks a field that was "*", which is worth keeping because "every
	// minute" and "minute 0" are different rules that both match minute 0.
	Any [5]bool
}

// String writes the fields back as a cron expression.
func (f CronFields) String() string {
	parts := make([]string, 5)
	for i, set := range [][]int{f.Minute, f.Hour, f.Day, f.Month, f.Weekday} {
		if f.Any[i] {
			parts[i] = "*"
			continue
		}
		nums := make([]string, len(set))
		for j, n := range set {
			nums[j] = strconv.Itoa(n)
		}
		parts[i] = strings.Join(nums, ",")
	}
	return strings.Join(parts, " ")
}

// ParseCron reads a five-field cron expression: minute, hour, day of month,
// month, day of week.
//
// It is the standard form and not a variant, because the expression comes out
// of a scheduling service and has to mean there what it means here. Ranges,
// lists and steps are accepted; names such as "MON" are not, and neither is
// the "@daily" shorthand — a rule a person cannot read back in the field it
// came from is a rule they cannot check.
func ParseCron(expr string) (CronFields, error) {
	parts := strings.Fields(strings.TrimSpace(expr))
	if len(parts) != 5 {
		return CronFields{}, fmt.Errorf("datetime: %q is not a cron expression: want 5 fields, got %d",
			expr, len(parts))
	}
	limits := [5][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 6}}
	var out CronFields
	targets := [5]*[]int{&out.Minute, &out.Hour, &out.Day, &out.Month, &out.Weekday}
	for i, part := range parts {
		if part == "*" {
			out.Any[i] = true
			*targets[i] = rangeAll(limits[i][0], limits[i][1])
			continue
		}
		nums, err := parseCronField(part, limits[i][0], limits[i][1])
		if err != nil {
			return CronFields{}, fmt.Errorf("datetime: %q field %d: %w", expr, i+1, err)
		}
		*targets[i] = nums
	}
	return out, nil
}

// ValidateCron reports whether an expression is one this package can read.
func ValidateCron(expr string) error {
	_, err := ParseCron(expr)
	return err
}

func parseCronField(part string, lo, hi int) ([]int, error) {
	seen := map[int]bool{}
	var out []int
	for _, term := range strings.Split(part, ",") {
		step := 1
		if i := strings.Index(term, "/"); i >= 0 {
			n, err := strconv.Atoi(term[i+1:])
			if err != nil || n < 1 {
				return nil, fmt.Errorf("%q is not a step", term)
			}
			step = n
			term = term[:i]
		}
		from, to := lo, hi
		switch {
		case term == "*":
		case strings.Contains(term, "-"):
			bounds := strings.SplitN(term, "-", 2)
			var err error
			if from, err = strconv.Atoi(bounds[0]); err != nil {
				return nil, fmt.Errorf("%q is not a number", bounds[0])
			}
			if to, err = strconv.Atoi(bounds[1]); err != nil {
				return nil, fmt.Errorf("%q is not a number", bounds[1])
			}
		default:
			n, err := strconv.Atoi(term)
			if err != nil {
				return nil, fmt.Errorf("%q is not a number", term)
			}
			from, to = n, n
			if step > 1 {
				// "5/15" means from 5 to the end of the field, every 15.
				to = hi
			}
		}
		if from < lo || to > hi || from > to {
			return nil, fmt.Errorf("%d-%d is outside %d-%d", from, to, lo, hi)
		}
		for n := from; n <= to; n += step {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%q matches nothing", part)
	}
	return out, nil
}

func rangeAll(lo, hi int) []int {
	out := make([]int, 0, hi-lo+1)
	for n := lo; n <= hi; n++ {
		out = append(out, n)
	}
	return out
}

// CronSummary writes a cron expression as a person would say it: "Every day
// at 09:30", "On Monday, Tuesday, Wednesday, Thursday, Friday at 07:00".
//
// It is a summary, not a translation: a field that is a list of thirty numbers
// is written as a phrase, because the alternative is a sentence nobody reads.
// A rule this package cannot parse is written back exactly as it was given,
// so an unreadable rule is visible rather than summarised into something else.
func CronSummary(expr string) string {
	f, err := ParseCron(expr)
	if err != nil {
		return expr
	}
	return daySummary(f) + " " + clockSummary(f)
}

// daySummary names the days a rule lands on. The two day fields are both
// starred or neither in every rule anybody writes by hand, and a field that
// says "on Mondays and the 13th" is ambiguous in the original too — so the
// two are reported together rather than one of them silently winning.
func daySummary(f CronFields) string {
	switch {
	case f.Any[2] && f.Any[4]:
		return core.Def("Every day")
	case f.Any[2]:
		return "On " + weekdayNames(f.Weekday)
	case f.Any[4]:
		return "On day " + numberList(f.Day) + core.Def(" of ") + numberList(f.Month)
	default:
		return "On " + numberList(f.Month) + " " + numberList(f.Day) +
			core.Def(", and on ") + weekdayNames(f.Weekday)
	}
}

// weekdayNames spells cron's day-of-week numbers, which count from Sunday,
// into the Monday-first names the rest of this package uses. The shift is the
// whole conversion, and getting it wrong moves a nightly job onto the wrong
// night — which is the kind of bug nobody notices until a backup has not run.
func weekdayNames(nums []int) string {
	out := make([]string, 0, len(nums))
	for _, d := range nums {
		out = append(out, WeekdayFullNames[(d+6)%7])
	}
	return strings.Join(out, ", ")
}

// clockSummary names the times. Two single fields are a clock, which is what
// almost every rule is, and it reads as one rather than as two.
func clockSummary(f CronFields) string {
	if len(f.Minute) == 1 && len(f.Hour) == 1 {
		return core.Def("at ") + pad2(f.Hour[0]) + ":" + pad2(f.Minute[0])
	}
	switch {
	case len(f.Minute) == 1 && f.Any[1]:
		return core.Def("every hour at ") + pad2(f.Minute[0]) + core.Def(" past")
	case len(f.Minute) == 1:
		return hoursPhrase(f.Hour) + core.Def(" at ") + pad2(f.Minute[0]) + core.Def(" past")
	case f.Any[1] && f.Minute[1]-f.Minute[0] == 1:
		return core.Def("every minute")
	}
	return hoursPhrase(f.Hour) + ", " + minutesPhrase(f.Minute)
}

func hoursPhrase(nums []int) string {
	if len(nums) == 1 {
		return core.Def("at ") + pad2(nums[0]) + ":00"
	}
	return core.Def("from ") + pad2(nums[0]) + ":00" +
		core.Def(" to ") + pad2(nums[len(nums)-1]) + ":00"
}

func minutesPhrase(nums []int) string {
	return core.Def("from minute ") + itoa(nums[0]) + core.Def(" to minute ") +
		itoa(nums[len(nums)-1]) + core.Def(" every ") + itoa(nums[1]-nums[0])
}

func numberList(nums []int) string {
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = itoa(n)
	}
	return strings.Join(parts, ", ")
}

// CronEditorOptions configure a CronEditor.
type CronEditorOptions struct {
	// Expression is the caller's cron string, and is written as the fields
	// are chosen. Required.
	Expression *string
	// Fields are the caller's five choices, each a list of the values that
	// field may take. Nil uses the defaults — every minute, every hour, every
	// day, every month, every weekday — which makes the editor a place to
	// narrow a rule rather than to build one from nothing.
	Fields [5][]int
	// Names labels the five rows; nil uses Minute, Hour, Day, Month, Weekday.
	Names [5]string
	// OnChange is called with the expression whenever it is valid, so a caller
	// can see what the rule is before saving it.
	OnChange func(string)
}

// CronEditorResult carries a CronEditor, its expression and whether it is one
// this package can read.
type CronEditorResult struct {
	// Element is the control.
	Element *ui.Element
	// changed reports that the expression moved this frame.
	changed bool
	// err is why the expression is not readable, or nil.
	err error
}

// Changed reports that the expression moved this frame.
func (r CronEditorResult) Changed() bool { return r.changed }

// Err returns why the expression cannot be read, or nil when it can. A
// scheduler that is handed a rule it cannot parse fails silently at three in
// the morning, so the editor says so on the screen rather than on save.
func (r CronEditorResult) Err() error { return r.err }

// CronEditor is the five fields of a cron rule, one row each, with the
// sentence it means underneath.
//
// Each row is a row of buttons rather than a text field, because a cron
// expression is the one place in this package where a typo is silent: "0 9 * *"
// never fires and never complains, and the complaint only arrives as a job
// that did not run. Choosing from the values that are allowed cannot be typed
// wrong.
func CronEditor(c *ui.Context, opts CronEditorOptions) CronEditorResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Expression == nil {
		panic("datetime: CronEditor needs an Expression to point at")
	}
	defaults := [5][]int{
		rangeAll(0, 59), rangeAll(0, 23), rangeAll(1, 31), rangeAll(1, 12), rangeAll(0, 6),
	}
	fields := opts.Fields
	for i := range defaults {
		if len(fields[i]) == 0 {
			fields[i] = defaults[i]
		}
	}
	names := opts.Names
	defaultNames := [5]string{
		core.Msg(c, "datetime.cronMinute", "Minute"),
		core.Msg(c, "datetime.cronHour", "Hour"),
		core.Msg(c, "datetime.cronDay", "Day of month"),
		core.Msg(c, "datetime.cronMonth", "Month"),
		core.Msg(c, "datetime.cronWeekday", "Day of week"),
	}
	for i, n := range defaultNames {
		if names[i] == "" {
			names[i] = n
		}
	}

	chosen := [5]int{}
	// The error is kept as it is, with the reason in it: "not readable" tells
	// a person nothing, and the reason — five fields, a number out of range —
	// is the half of it they can act on.
	err := ValidateCron(*opts.Expression)
	if err == nil {
		f, _ := ParseCron(*opts.Expression)
		picked := [5][]int{f.Minute, f.Hour, f.Day, f.Month, f.Weekday}
		for i := range picked {
			chosen[i] = picked[i][0]
		}
	}

	var r CronEditorResult
	r.err = err
	r.Element = ui.Column(c).FillWidth().Gap(u * 1.5).
		Label(core.Msg(c, "datetime.cron", "Cron rule")).Children(func() {
		for row := range 5 {
			row := row
			ui.Column(c).FillWidth().Gap(u / 2).Label(names[row]).Children(func() {
				ui.Text(c, names[row]).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				// The strip wraps: Minute is sixty chips, and one row of
				// them ran off the page's right edge, cutting the last chip
				// in half and dropping the rest of the hour.
				ui.Row(c).FillWidth().Wrap().AlignContent(ui.Start).Gap(u / 2).Children(func() {
					for _, v := range fields[row] {
						v := v
						on := v == chosen[row]
						bg, fg := k.Surface, k.Text
						if on {
							bg, fg = k.Fill, k.OnFill
						}
						label := itoa(v)
						btn := ui.Button(c, "").MinWidth(MinCellSize).
							Height(u * 8).Radius(theme.SmallRadius).
							Background(bg).TextColor(fg).Label(label).Children(func() {
							ui.Text(c, label).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize))
						})
						if btn.Clicked() {
							chosen[row] = v
							out := CronFields{
								Minute:  []int{chosen[0]},
								Hour:    []int{chosen[1]},
								Day:     []int{chosen[2]},
								Month:   []int{chosen[3]},
								Weekday: []int{chosen[4]},
							}
							*opts.Expression = out.String()
							r.err = nil
							r.changed = true
							if opts.OnChange != nil {
								opts.OnChange(*opts.Expression)
							}
						}
					}
				})
			})
		}
		// The sentence is the whole point: five rows of numbers are not a
		// rule anybody can check, and this is the line that lets them.
		summary := CronSummary(*opts.Expression)
		if r.err != nil {
			summary = r.err.Error()
		}
		colour := k.TextMuted
		if r.err != nil {
			colour = k.Danger
		}
		ui.Text(c, summary).TextColor(colour).FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
	return r
}
