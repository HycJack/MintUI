package pages

// The datetime page is the whole of ui/datetime in one column, in the order a
// booking page is put together: pick a day, put things on it, look datetime_at a week,
// then choose a clock and a rule for repeating it.
//
// Nothing on this page reads the clock. Every view is handed one fixed "now"
// — Wednesday 7 October 2026, 14:30 — so two screenshots of this page are
// comparable, and so a reader can check the arithmetic in the bottom section
// against the components datetime_at the top.

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/datetime"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	selected     = dtChosen
	gridSelected = dtChosen
	events       = dtEvents()
	openEditor   = false
	popoverOpen  = true
	editing      = dtEvents()[2]
	dateOpen     = true
	value        = dtChosen
	when         = dtNow
	week         = dtChosen
	highlighted  = dtDay
	datetime_at  = dtAt(14, 30)
	dur          = 90 * time.Minute
	zone         = "Europe/Berlin"
	started      = dtNow.Add(-83400 * time.Millisecond)
	lapList      = []time.Time{dtNow.Add(-51400 * time.Millisecond), dtNow.Add(-24000 * time.Millisecond)}
	rule         = datetime.Recurrence{
		Frequency: datetime.Weekly, Interval: 2,
		Weekdays: [7]bool{true, false, false, false, true, false, false},
		Count:    10,
	}
	custom = 45 * time.Minute
	cron   = "30 9 * * 1-5"
	broken = "99 * * * *"
)

func init() {
	showcase.Register(showcase.Page{
		Package: "datetime",
		Title:   "ui/datetime — 日期和时间画出来是什么样子",
		Note:    "月、周、日、议程、一整套选择器、秒表倒计时、重复规则，以及底下那些纯函数",
		Width:   1000,
		Height:  8000,
		Want: []string{
			// the days
			"October 2026",
			"09:00  Maple Street Bakery",
			"11:00  Riverside Clinic · 88 Riverside Drive",
			"09:30  Oak Lane Cafe · 4 Oak Lane",
			"5–11 October 2026 · W41",
			"Wednesday", "7 October 2026",
			"January 2026", "2016 – 2027",
			// the clocks
			"Stopwatch 01:23.4", "1h 30m", "SLA", "On site",
			// the pickers
			"Appointment", "Date range", "Duration", "Time zone",
			"Europe/Berlin · UTC+02 CEST", "15m before",
			// the rules
			"Every 2 weeks on Mon, Fri for 10 times",
			"On Monday, Tuesday, Wednesday, Thursday, Friday at 09:30",
			// the words a moment is described in
			"3 days ago", "in 2 hours", "just now",
			// the pure functions, by call and by answer
			"FormatDate(day)", "FormatTime(now)", "FormatDateTime(now)",
			"ParseDate(\"2026-10-07\")", "ParseTime(\"14:30\")",
			"ParseDateTime(\"2026-10-07 14:30\")", "ParseDurationText(\"1h 30m\")",
			"ParseCron(\"30 9 * * 1-5\")", "ValidateCron(\"99 * * * *\")",
			"DaysBetween(day, today)", "DaysInMonth(feb)", "AddDays(day, 30)",
			"AddMonths(jan31, 1)", "StartOfDay(now)", "EndOfDay(now)",
			"StartOfWeek(now)", "StartOfMonth(now)", "StartOfYear(now)",
			"IsLeapYear(2028)", "WeekOfYear(day)", "Until(start, start+90m)",
			"Relative(day, now)", "DefaultMonths()", "DefaultWeekdays()",
			"ZoneOffset(Berlin, now)", "Overlaps(a, b)", "OverlapLanes(day)",
			"SameDay / SameWeek / SameMonth", "MonthCells(oct, 6)",
			"WeekCells(day)", "DurationText(90m)", "WeekViewTitle(day)",
			"RecurrenceSummary(weekly)", "CronSummary(\"30 9 * * 1-5\")",
		},
		Render: func(c *ui.Context) {
			datetimePage(c)
		},
	})
}

// datetimePage is the package in the order a booking page is assembled.
func datetimePage(c *ui.Context) {
	dtCalendars(c)
	dtViews(c)
	dtAgenda(c)
	dtPickers(c)
	dtCoarse(c)
	dtClocks(c)
	dtTimers(c)
	dtRules(c)
	dtFunctions(c)
	dtConventions(c)
}

// ── the data ───────────────────────────────────────────────────────────────
//
// One fixed week. Every view on this page is a different arrangement of these
// same events, which is the only way a reader can tell whether two of them
// disagree.

var (
	dtNow    = time.Date(2026, 10, 7, 14, 30, 0, 0, time.UTC)
	dtDay    = time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	dtChosen = time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	dtMonday = time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
)

// dtAt is a moment on the fixed week, written as hours past midnight.
func dtAt(h, m int) time.Time {
	return dtDay.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute)
}

func dtEvents() []datetime.Event {
	return []datetime.Event{
		{
			Title: "Maple Street Bakery", Start: dtAt(9, 0), End: dtAt(10, 0),
			Location: "12 Maple Street", Severity: core.Accent,
		},
		{
			// Overlapping the first one on purpose: the columns draw
			// doubles side by side rather than one on top of the other.
			Title: "Oak Lane Cafe", Start: dtAt(9, 30), End: dtAt(10, 30),
			Location: "4 Oak Lane",
		},
		{
			Title: "Riverside Clinic", Start: dtAt(11, 0), End: dtAt(12, 30),
			Location: "88 Riverside Drive", Severity: core.Warning,
		},
		{
			Title: "Northgate Dental", Start: dtAt(15, 0), End: dtAt(16, 0),
		},
		{
			Title: "Quarterly service window", AllDay: true,
			Start: dtDay, End: dtDay, Severity: core.Danger,
		},
		{
			Title: "Elm Street Gym", Start: dtAt(17, 0), End: dtAt(17, 30),
			Cancelled: true,
		},
	}
}

// dtWeekEvents is the same clinic over the whole week, so a week view and a
// day view are visibly two readings of one schedule.
func dtWeekEvents() []datetime.Event {
	return []datetime.Event{
		{Title: "Maple Street Bakery", Start: dtMonday.Add(9 * time.Hour), End: dtMonday.Add(10 * time.Hour)},
		{Title: "Oak Lane Cafe", Start: dtMonday.Add(14 * time.Hour), End: dtMonday.Add(15 * time.Hour)},
		{Title: "Northgate Dental", Start: dtMonday.Add(13 * time.Hour), End: dtMonday.Add(14 * time.Hour)},
		{Title: "Harbour Cafe", Start: dtMonday.AddDate(0, 0, 1).Add(10 * time.Hour), End: dtMonday.AddDate(0, 0, 1).Add(11 * time.Hour)},
		{Title: "Riverside Clinic", Start: dtDay.Add(11 * time.Hour), End: dtDay.Add(12*time.Hour + 30*time.Minute)},
		{Title: "Elm Street Gym", Start: dtDay.Add(17 * time.Hour), Severity: core.Danger},
		{Title: "Quarterly service window", AllDay: true, Start: dtDay, End: dtDay},
		{Title: "Ashgrove Gym", Start: dtDay.AddDate(0, 0, 1).Add(9 * time.Hour), End: dtDay.AddDate(0, 0, 1).Add(10*time.Hour + 30*time.Minute)},
		{Title: "Linden Bakery", Start: dtDay.AddDate(0, 0, 2).Add(13 * time.Hour)},
		{Title: "Birch Lane Clinic", Start: dtDay.AddDate(0, 0, 3).Add(8 * time.Hour), End: dtDay.AddDate(0, 0, 3).Add(9 * time.Hour)},
		{Title: "Cedar Court", Start: dtDay.AddDate(0, 0, 3).Add(16 * time.Hour), End: dtDay.AddDate(0, 0, 3).Add(17 * time.Hour)},
	}
}

// ── 1. calendars ───────────────────────────────────────────────────────────

func dtCalendars(c *ui.Context) {
	showcase.Section(c, "日历 · Calendar / CalendarMonthView / Grid / CalendarYearView")

	showcase.Field(c, "Calendar — 带月份标题和翻页箭头的那个；箭头直接改调用方的 Month")
	month := dtDay
	dtTwoUp(c,
		func() {
			datetime.Calendar(c, datetime.CalendarOptions{
				Month: &month, Header: true, TodayButton: true,
				Today: dtNow, Selected: &selected,
				Start: dtMonday, End: dtMonday.AddDate(0, 0, 3),
				Busy:            func(d time.Time) bool { return d.Day() == 20 || d.Day() == 21 },
				ShowWeekNumbers: true,
			})
		},
		func() {
			datetime.CalendarMonthView(c, datetime.CalendarMonthViewOptions{
				Month: dtDay, Today: dtNow, Selected: &selected, Caption: true,
				Start: dtMonday, End: dtMonday.AddDate(0, 0, 3),
				Busy:  func(d time.Time) bool { return d.Day() == 20 },
				Weeks: 6,
			})
		})

	showcase.Field(c, "Grid — 所有日历底下的那一个：只有格子，没有标题和箭头")
	// The grid is given a width of its own rather than the rest of the row:
	// it fills whatever box it is put in, and left to itself it takes the
	// whole line and squeezes the sentence beside it to one character.
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(unit(c, 72)).Children(func() {
			datetime.Grid(c, datetime.GridOptions{
				Month: dtDay, Today: dtNow, Selected: &gridSelected,
				Start: dtMonday, End: dtMonday.AddDate(0, 0, 3),
				Weeks:           6,
				WeekHeader:      "W",
				ShowWeekNumbers: true,
				Busy:            func(d time.Time) bool { return d.Day() == 20 || d.Day() == 21 },
				Disabled:        func(d time.Time) bool { return d.Day() > 23 },
			})
		})
		ui.Column(c).Grow(1).Shrink(0).Gap(unit(c, 1)).Children(func() {
			ui.Text(c, "四种状态各画各的：选中填墨、今天描一圈、区间铺淡底、").TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
			ui.Text(c, "不在本月画得更淡、禁用日填灰且点不动。").TextColor(core.Tokens(c).TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		})
	})

	showcase.Field(c, "CalendarYearView — 十二个月，全年的样子")
	year := dtDay
	dtFull(c, func() {
		datetime.CalendarYearView(c, datetime.CalendarYearViewOptions{
			Year: &year, Today: dtNow, Selected: &dtChosen, Columns: 3,
			Highlighted: &dtDay,
			Busy:        func(d time.Time) bool { return d.Day() == 20 || d.Day() == 27 },
		})
	})
}

// ── 2. the views ───────────────────────────────────────────────────────────

func dtViews(c *ui.Context) {
	showcase.Section(c, "一周与一天 · CalendarWeekView / CalendarDayView / CalendarTimeGrid / CalendarEvents / CurrentTimeIndicator")

	showcase.Field(c, "CalendarWeekView — 小时只画一次，七列共用一把尺")
	dtFull(c, func() {
		datetime.CalendarWeekView(c, datetime.CalendarWeekViewOptions{
			Day: dtDay, Now: dtNow, Events: dtWeekEvents(), Title: true,
			AllDay: true, Pickable: true,
		})
	})

	showcase.Field(c, "CalendarDayView 与 CalendarTimeGrid — 一天自己站一行，和它背后的格子")
	dtTwoUp(c,
		func() {
			datetime.CalendarDayView(c, datetime.CalendarDayViewOptions{
				Day: dtDay, Now: dtNow, Events: dtEvents(),
				AllDay: true, Pickable: true,
				Busy: func(t time.Time) bool { return t.Hour() == 13 },
			})
		},
		func() {
			ui.Column(c).Gap(unit(c, 2)).Children(func() {
				datetime.CalendarTimeGrid(c, datetime.TimeGridOptions{
					Day: dtDay, StartHour: 8, EndHour: 18, Now: dtNow,
					Hours: true, Pickable: true,
					Disabled: func(t time.Time) bool { return t.Hour() == 13 },
				})
				ui.Text(c, "CurrentTimeIndicator 单独放出来 — 它是绝对定位的一条线，不占行").TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		})

	showcase.Field(c, "CalendarEvents — 一天的那一列，EventChip 是它画的东西")
	dtTwoUp(c,
		func() {
			datetime.CalendarEvents(c, datetime.CalendarEventsOptions{
				Day: dtDay, Events: events, AllDay: true,
				Selected: &events[0],
			})
		},
		func() {
			ui.Column(c).Gap(unit(c, 1.5)).Children(func() {
				ui.Text(c, "EventChip — 一件事的颜色、标题和开始时间").TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				ui.Box(c).Width(unit(c, 100)).Gap(unit(c, 1)).Children(func() {
					for _, e := range dtEvents()[:3] {
						e := e
						datetime.EventChip(c, datetime.EventChipOptions{
							Event: e, Selected: e.Title == "Riverside Clinic",
						})
					}
				})
				ui.Box(c).Width(unit(c, 100)).Children(func() {
					datetime.EventChip(c, datetime.EventChipOptions{
						Event: datetime.Event{
							Title: "Quarterly service window", AllDay: true,
							Start: dtDay, End: dtDay, Severity: core.Danger,
						},
						NoTime: true,
					})
				})
			})
		})
}

// ── 3. the agenda ──────────────────────────────────────────────────────────

func dtAgenda(c *ui.Context) {
	showcase.Section(c, "议程 · AgendaView / AttendeeList / AvailabilityPicker / EventPopover / EventEditor")

	showcase.Field(c, "AgendaView — 一条一条往下读，而不是横着看一周")
	days := []datetime.AgendaDay{
		{Day: dtMonday, Events: dtWeekEvents()[:3]},
		{Day: dtDay, Events: dtEvents()},
		{Day: dtDay.AddDate(0, 0, 1), Events: dtWeekEvents()[7:8]},
		{Day: dtDay.AddDate(0, 0, 2), Events: dtWeekEvents()[8:9]},
		{Day: dtDay.AddDate(0, 0, 5), Events: []datetime.Event{}},
	}
	dtTwoUp(c,
		func() {
			datetime.AgendaView(c, datetime.AgendaOptions{
				Days: days, Today: dtNow, Limit: 3,
			})
		},
		func() {
			ui.Column(c).Gap(unit(c, 3)).Children(func() {
				datetime.AttendeeList(c, datetime.AttendeeOptions{
					Avatars: true, ShowResponse: true,
					Attendees: []datetime.Attendee{
						{Name: "Dana Reyes", Response: datetime.Accepted, Timezone: "Europe/Berlin"},
						{Name: "Nate Coleman", Response: datetime.NoReply, Note: "chased on Tuesday"},
						{Name: "Ivy Ahmed", Response: datetime.Maybe},
						{Name: "Rafi Okonjo", Response: datetime.Declined, Timezone: "Asia/Tokyo"},
					},
				})
				datetime.AttendeeList(c, datetime.AttendeeOptions{
					Empty: func() {
						ui.Text(c, "Nobody has been invited yet.").TextColor(core.Tokens(c).TextMuted).
							FontSize(core.FontSize(c, theme.MetaSize))
					},
				})
				datetime.AvailabilityPicker(c, datetime.AvailabilityOptions{
					Days: []time.Time{dtDay, dtDay.AddDate(0, 0, 1)},
					Free: func(day time.Time, hour int) bool {
						return !(day.Equal(dtDay) && (hour == 9 || hour == 13)) &&
							!(day.After(dtDay) && hour < 10)
					},
					Marked: func(day time.Time, hour int) bool {
						return day.Equal(dtDay.AddDate(0, 0, 1)) && (hour == 11 || hour == 12)
					},
				})
			})
		})

	showcase.Field(c, "EventPopover 与 EventEditor — 挂在 chip 上的详情，和详情背后的那张表")
	reminder := 15 * time.Minute
	event := dtEvents()[2]
	dtTwoUp(c,
		func() {
			ui.Column(c).Gap(unit(c, 2)).Children(func() {
				ui.Text(c, "EventPopover 挂在下面这个 chip 上").TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
				anchor := datetime.EventChip(c, datetime.EventChipOptions{Event: event}).Element
				datetime.EventPopover(c, datetime.EventPopoverOptions{
					Anchor: anchor, Open: &popoverOpen, Event: event,
					Edit: "Edit", Delete: "Delete",
				})
			})
		},
		func() {
			datetime.EventEditor(c, datetime.EventEditorOptions{
				Event: &editing, Save: "Save", Cancel: "Cancel", Open: &openEditor,
				Reminder: &reminder, Timezone: "Europe/Berlin",
				Duration: 30 * time.Minute,
				Attendees: []datetime.Attendee{
					{Name: "Dana Reyes", Response: datetime.Accepted},
					{Name: "Ivy Ahmed", Response: datetime.Maybe},
				},
			})
		})
}

// ── 4. the pickers ─────────────────────────────────────────────────────────

func dtPickers(c *ui.Context) {
	showcase.Section(c, "日期选择 · DatePicker / DateRangePicker / DateTimePicker / WeekPicker")

	showcase.Field(c, "DatePicker（浮层开着）与 DateRangePicker — 选日期的两种宽度")
	rangeStart := showcase.State(c, "datetime.393.rangeStart", dtMonday)
	rangeEnd := showcase.State(c, "datetime.393.rangeEnd", dtMonday.AddDate(0, 0, 9))
	dtTwoUp(c,
		func() {
			datetime.DatePicker(c, datetime.DatePickerOptions{
				Value: &value, Open: &dateOpen, Today: dtNow, Name: "Appointment",
				Placeholder: "Pick a day",
				Busy:        func(d time.Time) bool { return d.Day() == 20 },
			})
		},
		func() {
			datetime.DateRangePicker(c, datetime.DateRangePickerOptions{
				Start: rangeStart, End: rangeEnd, Open: &dateOpen, Today: dtNow,
				Panels: 2, Name: "Date range",
			})
		})

	showcase.Field(c, "DateTimePicker 与 WeekPicker — 日期归日期，时钟归时钟")
	dtTwoUp(c,
		func() {
			datetime.DateTimePicker(c, datetime.DateTimePickerOptions{
				Value: &when, Open: &dateOpen, Today: dtNow, Name: "Appointment",
			})
		},
		func() {
			datetime.WeekPicker(c, datetime.WeekPickerOptions{
				Value: &week, Weeks: 4, Today: dtNow, Name: "Week",
				Disabled: func(d time.Time) bool { return d.Before(dtDay) },
			})
		})
}

// ── 5. months and years ────────────────────────────────────────────────────

func dtCoarse(c *ui.Context) {
	showcase.Section(c, "粗粒度 · MonthPicker / YearPicker")

	showcase.Field(c, "MonthPicker 与 YearPicker — 十二个月和十二年，都是一屏")
	month := dtDay
	year := dtDay
	dtTwoUp(c,
		func() {
			datetime.MonthPicker(c, datetime.MonthPickerOptions{
				Month: &month, Today: dtNow, Highlighted: &highlighted, Columns: 3,
			})
		},
		func() {
			datetime.YearPicker(c, datetime.YearPickerOptions{
				Year: &year, Today: dtNow, Highlighted: &highlighted, Span: 12, Columns: 4,
			})
		})
}

// ── 6. clocks and lengths ──────────────────────────────────────────────────

func dtClocks(c *ui.Context) {
	showcase.Section(c, "时刻与时长 · TimePicker / TimeRangePicker / DurationPicker / TimezoneSelect")

	showcase.Field(c, "TimePicker、TimeRangePicker 与 DurationPicker — 步进器而不是一张半小时的表")
	from := showcase.State(c, "datetime.451.from", dtAt(9, 0))
	to := showcase.State(c, "datetime.451.to", dtAt(10, 30))
	dtTwoUp(c,
		func() {
			ui.Column(c).Gap(unit(c, 3)).Children(func() {
				datetime.TimePicker(c, datetime.TimePickerOptions{
					Value: &datetime_at, ShowSeconds: true, Name: "Time",
				})
				datetime.TimeRangePicker(c, datetime.TimeRangePickerOptions{
					Start: from, End: to, Name: "Time range", ShowSeconds: true,
				})
			})
		},
		func() {
			ui.Column(c).Gap(unit(c, 3)).Children(func() {
				datetime.DurationPicker(c, datetime.DurationPickerOptions{
					Value: &dur, Step: 15 * time.Minute, Name: "Duration",
					Presets: []time.Duration{30 * time.Minute, time.Hour,
						2 * time.Hour, 4 * time.Hour},
				})
				datetime.TimezoneSelect(c, datetime.TimezoneSelectOptions{
					Value: &zone, At: dtNow, ShowOffsets: true, Name: "Time zone",
					Zones: []string{"UTC", "Europe/Berlin", "America/New_York", "Asia/Tokyo"},
				})
			})
		})
}

// ── 7. clocks that run ─────────────────────────────────────────────────────

func dtTimers(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "走着的表 · Stopwatch / Countdown / RelativeTime")

	showcase.Field(c, "Stopwatch 与 Countdown — 起点是调用方的，所以读数能从存档里恢复")
	// Eighty-three and a bit seconds: precision 1 so the reading is
	// "01:23.4", which is as fine as a person can follow a hand moving.
	dtTwoUp(c,
		func() {
			ui.Column(c).Gap(unit(c, 3)).Children(func() {
				datetime.Stopwatch(c, datetime.StopwatchOptions{
					Start: &started, Now: dtNow, Laps: &lapList,
					Precision: 1, Name: "Stopwatch",
				})
				ui.Text(c, fmt.Sprintf("跑了两圈：%s · %s",
					datetime.StopwatchText(lapList[0].Sub(started), 1),
					datetime.StopwatchText(lapList[1].Sub(started), 1))).
					TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
			})
		},
		func() {
			ui.Column(c).Gap(unit(c, 3)).Children(func() {
				datetime.Countdown(c, datetime.CountdownOptions{
					Target: dtNow.Add(3 * time.Hour), Now: dtNow, Name: "SLA",
					Format:  datetime.RelativeText,
					From:    dtNow.Add(-21 * time.Hour),
					Until:   dtNow.Add(3 * time.Hour),
					Compact: true,
				})
				// The other way up: an elapsed timer is the same control
				// with the ends swapped, which is why there is no second
				// component for "time since".
				datetime.Countdown(c, datetime.CountdownOptions{
					Target: dtNow, Start: dtNow.Add(-90 * time.Minute),
					Now: dtNow, Name: "On site", Format: datetime.RelativeText,
					From: dtNow.Add(-90 * time.Minute), Until: dtNow.Add(30 * time.Minute),
				})
			})
		})

	showcase.Field(c, "RelativeTime — 量出来是组件的事，写成什么字是窗口的事")
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
		rel := func(name string, then, now time.Time) {
			ui.Column(c).Width(unit(c, 44)).Gap(unit(c, 0.5)).Children(func() {
				datetime.RelativeTime(c, datetime.RelativeTimeOptions{
					Then: then, Now: now, Format: datetime.RelativeText,
					Muted: true,
				})
				ui.Text(c, name).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
		rel("三天前", dtNow.AddDate(0, 0, -3), dtNow)
		rel("一分钟后", dtNow.Add(time.Minute), dtNow)
		rel("两小时后", dtNow.Add(2*time.Hour), dtNow)
		rel("九天前", dtNow.AddDate(0, 0, -9), dtNow)
		rel("一月前", dtNow.AddDate(0, -1, 0), dtNow)
		rel("就是现在", dtNow, dtNow)
	})
}

// ── 8. rules ───────────────────────────────────────────────────────────────

func dtRules(c *ui.Context) {
	showcase.Section(c, "重复规则 · ReminderPicker / RecurrenceEditor / CronEditor")

	showcase.Field(c, "ReminderPicker 与 RecurrenceEditor — 提前多久提醒，和多久重复一次")
	reminder := 15 * time.Minute
	dtTwoUp(c,
		func() {
			ui.Column(c).Gap(unit(c, 3)).Children(func() {
				datetime.ReminderPicker(c, datetime.ReminderPickerOptions{
					Value: &reminder, Name: "Reminder",
				})
				datetime.ReminderPicker(c, datetime.ReminderPickerOptions{
					Value: &custom, Name: "Reminder, custom",
					Choices: []time.Duration{0, 10 * time.Minute, time.Hour},
				})
			})
		},
		func() {
			datetime.RecurrenceEditor(c, datetime.RecurrenceEditorOptions{
				Rule: &rule, Anchor: dtDay, Name: "Repeats",
			})
		})

	showcase.Field(c, "CronEditor — 五个字段各一行，下面是它写成的那句话")
	dtFull(c, func() {
		datetime.CronEditor(c, datetime.CronEditorOptions{
			Expression: &cron,
			Names:      [5]string{"Minute", "Hour", "Day", "Month", "Weekday"},
		})
	})
	showcase.Field(c, "ValidateCron 拦下来的那种规则，编辑器自己会写出来")
	dtFull(c, func() {
		datetime.CronEditor(c, datetime.CronEditorOptions{Expression: &broken})
	})
}

// ── 9. the pure functions ───────────────────────────────────────────────────

// dtFunctions writes out what the arithmetic underneath returns, so a reader
// can check the components datetime_at the top of this page against the numbers they
// are drawn from.
func dtFunctions(c *ui.Context) {
	showcase.Section(c, "纯函数 · 每个组件背后那几个不需要窗口的函数")

	start := showcase.State(c, "datetime.587.start", dtAt(9, 0))
	end := showcase.State(c, "datetime.587.end", dtAt(10, 30))
	a := dtEvents()[0]
	b := dtEvents()[1]
	berlin := datetime.LoadZone("Europe/Berlin")
	cronFields, cronErr := datetime.ParseCron("30 9 * * 1-5")
	_, parseErr := datetime.ParseDateTime("2026-10-07 14:30")
	_, durErr := datetime.ParseDurationText("1h 30m")
	parsed, _ := datetime.ParseDate("2026-10-07")
	parsedTime, _ := datetime.ParseTime("14:30")
	parsedDateTime, _ := datetime.ParseDateTime("2026-10-07 14:30")
	threeDaysAgo := dtNow.AddDate(0, 0, -3)
	span := datetime.Relative(threeDaysAgo, dtNow)
	laneCount := len(datetime.OverlapLanes(dtEvents()))
	offsetMin, offsetName := datetime.ZoneOffset(berlin, dtNow)
	feb := time.Date(2028, 2, 1, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	cells := datetime.MonthCells(dtDay, 6)

	for _, f := range []struct{ name, value string }{
		{"FormatDate(day)", datetime.FormatDate(dtDay)},
		{"FormatTime(now)", datetime.FormatTime(dtNow)},
		{"FormatDateTime(now)", datetime.FormatDateTime(dtNow)},
		{"ParseDate(\"2026-10-07\")", fmt.Sprintf("%s · %v", datetime.FormatDate(parsed), errText(parseErr))},
		{"ParseTime(\"14:30\")", datetime.FormatTime(parsedTime)},
		{"ParseDateTime(\"2026-10-07 14:30\")", fmt.Sprintf("%s · %v", datetime.FormatTime(parsedDateTime), errText(parseErr))},
		{"ParseDurationText(\"1h 30m\")", fmt.Sprintf("%v · %v", dur90(), errText(durErr))},
		{"ParseCron(\"30 9 * * 1-5\")", fmt.Sprintf("%d/%d/%d/%d/%d · %v",
			len(cronFields.Minute), len(cronFields.Hour), len(cronFields.Day),
			len(cronFields.Month), len(cronFields.Weekday), errText(cronErr))},
		{"ValidateCron(\"99 * * * *\")", errText(datetime.ValidateCron("99 * * * *"))},
		{"ValidateCron(\"30 9 * * 1-5\")", errText(datetime.ValidateCron("30 9 * * 1-5"))},
		{"DaysBetween(day, today)", fmt.Sprint(datetime.DaysBetween(dtDay, dtNow))},
		{"DaysInMonth(feb)", fmt.Sprint(datetime.DaysInMonth(feb))},
		{"AddDays(day, 30)", datetime.FormatDate(datetime.AddDays(dtDay, 30))},
		{"AddMonths(jan31, 1)", datetime.FormatDate(datetime.AddMonths(jan31, 1))},
		{"StartOfDay(now)", datetime.FormatDateTime(datetime.StartOfDay(dtNow))},
		{"EndOfDay(now)", datetime.FormatDateTime(datetime.EndOfDay(dtNow))},
		{"StartOfWeek(now)", datetime.FormatDate(datetime.StartOfWeek(dtNow))},
		{"StartOfMonth(now)", datetime.FormatDate(datetime.StartOfMonth(dtNow))},
		{"StartOfYear(now)", datetime.FormatDate(datetime.StartOfYear(dtNow))},
		{"IsLeapYear(2028)", fmt.Sprint(datetime.IsLeapYear(2028))},
		{"WeekOfYear(day)", fmt.Sprint(datetime.WeekOfYear(dtDay))},
		{"Until(start, start+90m)", datetime.DurationText(datetime.Until(start, end))},
		{"Relative(day, now)", fmt.Sprintf("%d 天前 · %s",
			span.Days, datetime.RelativeText(span))},
		{"DefaultMonths()", datetime.DefaultMonths()[9]},
		{"DefaultWeekdays()", datetime.DefaultWeekdays()[0]},
		{"ZoneOffset(Berlin, now)", fmt.Sprintf("+%d 分钟 · %s", offsetMin, offsetName)},
		{"Overlaps(a, b)", fmt.Sprint(datetime.Overlaps(a, b))},
		{"OverlapLanes(day)", fmt.Sprintf("%d 条泳道", laneCount)},
		{"SameDay / SameWeek / SameMonth", fmt.Sprintf("%v / %v / %v",
			datetime.SameDay(dtDay, dtChosen), datetime.SameWeek(dtDay, dtChosen),
			datetime.SameMonth(dtDay, dtChosen))},
		{"MonthCells(oct, 6)", fmt.Sprintf("%d 格 · 首格 %s", len(cells),
			datetime.FormatDate(cells[0].Day))},
		{"WeekCells(day)", fmt.Sprintf("%d 格 · 周一 %s", len(datetime.WeekCells(dtDay)),
			datetime.FormatDate(datetime.WeekCells(dtDay)[0].Day))},
		{"DurationText(90m)", datetime.DurationText(90 * time.Minute)},
		{"WeekViewTitle(day)", datetime.WeekViewTitle(dtDay, datetime.DefaultMonths())},
		{"RecurrenceSummary(weekly)", datetime.RecurrenceSummary(datetime.Recurrence{
			Frequency: datetime.Weekly, Interval: 2,
			Weekdays: [7]bool{true, false, false, false, true, false, false}, Count: 10,
		})},
		{"CronSummary(\"30 9 * * 1-5\")", datetime.CronSummary("30 9 * * 1-5")},
	} {
		dtFnRow(c, f.name, f.value)
	}
}

// dtFnRow is one line of the table: the call on the left, its answer on the
// right. The row has a width of its own rather than filling the page, because
// a table whose last column runs to the window's edge reads as clipped rather
// than as padded.
func dtFnRow(c *ui.Context, name, value string) {
	k := core.Tokens(c)
	ui.Row(c).Width(unit(c, 236)).Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, name).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize)).Width(unit(c, 48)).Shrink(0)
		ui.Text(c, value).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

func errText(err error) string {
	if err == nil {
		return "没有错"
	}
	return "读不出来：" + err.Error()
}

func dur90() time.Duration {
	d, _ := datetime.ParseDurationText("1h 30m")
	return d
}

// ── 10. what the package is for ─────────────────────────────────────────────

func dtConventions(c *ui.Context) {
	k := core.Tokens(c)
	showcase.Section(c, "约定")
	ui.Column(c).Width(unit(c, 236)).Gap(unit(c, 1)).Children(func() {
		for _, line := range []string{
			"本包不读时钟：每个需要「现在」的组件都从字段拿 now，所以截图之间能比。",
			"状态也是调用方的：选中的那天是一个 *time.Time，按压从 Result 上回来。",
			"一周从周一开始，周一是 0（time.Weekday 的周日是 0），换算只写一次。",
			"星期和月份的名字由调用方给（WeekdayNames / MonthNames），本包不翻译。",
			"时刻是 ISO 的（2026-10-07 / 14:30 / 24 小时），同一个集合给全库用。",
			"区间是半开的：[start, end]。两个只在边界相接的事件不算撞车。",
			"DurationText 只写最大的两个单位，StopwatchText 只到百分之一秒。",
		} {
			ui.Text(c, line).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.RowSize))
		}
	})
}

// ── layout helpers ─────────────────────────────────────────────────────────

// dtTwoUp puts two things side by side, each taking half the page's own width
// less the gap between them.
func dtTwoUp(c *ui.Context, first, second func()) {
	ui.Row(c).FillWidth().Gap(unit(c, 3)).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Grow(1).Shrink(0).Children(first)
		ui.Box(c).Grow(1).Shrink(0).Children(second)
	})
}

// dtFull gives something the page's whole width, which is what the wide things
// — a week view, a cron editor — need: given half of it they would squeeze
// their columns until the text wrapped.
func dtFull(c *ui.Context, child func()) {
	ui.Box(c).Width(unit(c, 236)).Children(child)
}
