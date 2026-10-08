package pages

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/project"
	"github.com/HycJack/MintUI/ui/showcase"
	"github.com/HycJack/MintUI/ui/theme"
)

// Demo state outlives the frame: every demo below hands its component
// a pointer, and a pointer into a frame-local is a click the next
// frame undoes — a tab that will not switch, a dropdown that snaps
// shut, a slider that springs back.
var (
	dueOpen          = false
	project_empty    = ""
	noneOpen         = false
	running          = true
	idle             = false
	projectDetailDue = project.DueValue{
		Day: time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
		Set: true, Text: "Tomorrow",
	}
)

// The page is one board, from a single card to the sprint it belongs to, drawn
// with every component this package has.
//
// It is arranged as a person meets the package: a card, then a list of them,
// then the board they live on, then the one task opened up, then the controls
// that change it, then the numbers that say whether the sprint is going to
// make it. Reading a board in that order is also the order the components
// are built in, which is the only reason a gallery of them reads as one thing
// rather than as an index.
//
// Everything on the page is fixed: the same seven tasks, the same four
// columns, the same ten days of burndown. A page that drew differently each
// time it was opened could not be compared against the last one, and a
// screenshot that changed between runs is a screenshot nobody reviews.

func init() {
	showcase.Register(ProjectPage())
}

// ProjectPage is the gallery's page for ui/project.
func ProjectPage() showcase.Page {
	return showcase.Page{
		Package: "project",
		Title:   "ui/project — 一块板与板上的一条",
		Note:    "卡片、列表、看板、冲刺、详情、选择器、子任务、进度、工时",
		Width:   1000,
		Height:  4000,
		Want: []string{
			// the card and the list
			"Riverside Clinic — AC not cooling", "CB-1042", "CB-1039",
			"Northgate Dental — no hot water", "Harbour Cafe — oven door hinge",
			"Elm Street Gym — boiler losing pressure", "Tasks",
			// IssueCard
			"Pinewood Dental — thermostat replacement", "In review", "Billable",
			// the boards
			"Backlog", "In progress", "In review",
			"todo", "blocked", "done",
			"Sprint 14", "3 days left", "Nothing waiting on review",
			// the detail panel
			"Task", "Title", "Description", "Assignee", "Due date", "Save",
			"Subtasks", "2/4 done",
			// the controls
			"Status", "Nobody", "Andre Thomson", "Mia Chen", "Ravi Patel",
			"Labels", "New label", "Add",
			// progress
			"22 / 24", "Due 28 March", "3 / 24", "Shipped 6 March",
			"Workload", "Assigned", "Capacity",
			"Sprint 14 burndown", "Plan",
			// time
			"Time", "Running", "Start", "Log time", "Stop",
			"Timer", "Focus · left 7:00", "Break · left 1:00", "Reset",
			// pure functions
			`FormatIssueID("cb", 1042)`, `PrefixIssueID("cb")`,
			`FormatElapsed(1*time.Hour + 5*time.Minute)`,
			`Task{ID: "1042"}.IDNumber()`,
			`IdealBurndown(10, 50)`, `ActualBurndown(days, 10)`,
			`SprintLength(burndown)`, `MilestoneFraction(18, 24)`,
			`PomodoroAt(27*time.Minute, 25, 5)`,
		},
		Render: func(c *ui.Context) {
			projectPage(c)
		},
	}
}

func projectPage(c *ui.Context) {
	taskSection(c)
	sprintSection(c)
	detailSection(c)
	controlSection(c)
	metricSection(c)
	timeSection(c)
	projectPureSection(c)
}

// ── one card ─────────────────────────────────────────────────────────────

// cardSection is the same task drawn three ways — a card on a board, a row in
// a list, a panel on its own — because the three are different components
// with different rules about height, and a page that showed only one of them
// would be hiding the difference.
func taskSection(c *ui.Context) {
	showcase.Section(c, "卡片与列表 · TaskItem / TaskList / IssueCard / IssueIdBadge / PriorityPill")

	tasks := boardTasks()
	selected := "1042"

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(30).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "TaskItem — 标题永远一行，卡片永远一个高")
			for i, t := range tasks[:3] {
				ui.Box(c).FillWidth().Children(func() {
					project.TaskItem(c, project.TaskItemOptions{
						Task: t, Prefix: "CB", Selected: i == 0,
					})
				})
			}
			showcase.Field(c, "Compact — 窄列里的同一张卡")
			project.TaskItem(c, project.TaskItemOptions{
				Task: tasks[3], Prefix: "CB", Compact: true,
			})
		})

		ui.Column(c).WidthPercent(30).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "TaskList — 侧栏里的窄列表，一个任务一行")
			list := project.TaskList(c, project.TaskListOptions{
				Tasks: tasks, Selected: &selected, Prefix: "CB", Height: unit(c, 46),
				Empty: func() {
					display.Text(c, "No tasks match this filter.",
						display.TextOptions{Muted: true})
				},
			})
			resultLine(c, "TaskList.Selected()", quoteOrEmpty(list.Selected()))
			showcase.Field(c, "TaskList — 筛选之后什么都没有的样子")
			none := ""
			project.TaskList(c, project.TaskListOptions{
				Tasks: nil, Selected: &none, Prefix: "CB", Height: unit(c, 14),
				Empty: func() {
					display.Text(c, "No tasks match this filter.",
						display.TextOptions{Muted: true})
				},
			})
		})

		ui.Column(c).WidthPercent(36).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "IssueCard — 详情面板里那张卡")
			project.IssueCard(c, project.IssueCardOptions{
				Task: tasks[1], Prefix: "CB", Selected: true,
				Body: func() {
					ui.Text(c, "The boiler has been losing pressure since Friday. "+
						"It is the second time this month.").TextColor(core.Tokens(c).TextMuted).
						FontSize(core.FontSize(c, theme.RowSize))
				},
				Footer: func() {
					showcase.Stack(c, 1, func() {
						display.Tag(c, "Billable", display.TagOptions{Tone: core.Accent})
						display.Tag(c, "onsite", display.TagOptions{})
					})
				},
			})
			showcase.Field(c, "IssueIdBadge 和 PriorityPill — 全是记号")
			ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
				for _, sev := range []core.Severity{core.Neutral, core.Accent, core.Warning, core.Danger} {
					project.IssueIdBadge(c, project.IssueIdBadgeOptions{
						Prefix: "CB", Number: 1042, Severity: sev,
					})
				}
			})
			ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
				for _, points := range []int{1, 2, 3} {
					project.PriorityPill(c, points)
				}
				ui.Box(c).Shrink(0).Children(func() {
					project.IssueIdBadge(c, project.IssueIdBadgeOptions{})
				})
			})
		})
	})

	showcase.Field(c, "StatusSeverity — 一个状态一个颜色，板上的状态色全从这里来")
	ui.Row(c).FillWidth().Gap(unit(c, 1.5)).Children(func() {
		for _, s := range projectStatuses() {
			display.Tag(c, string(s), display.TagOptions{
				Tone: project.StatusSeverity(s),
			})
		}
	})
}

// ── the board ────────────────────────────────────────────────────────────

// boardSection is the board itself and the same board with a sprint clock
// over it.
//
// Both are in frames of a fixed height. A board fills both ways, so laid
// straight into a page column it would take the column's whole height — which
// is the page's whole height.
func sprintSection(c *ui.Context) {
	showcase.Section(c, "看板 · KanbanBoard / SprintBoard")
	tasks := boardTasks()
	selected := ""

	showcase.Field(c, "KanbanBoard — 三列；再多的列横向滚，列宽是定数")
	ui.Box(c).FillWidth().Height(unit(c, 66)).Radius(theme.ControlRadius).
		Border(theme.BorderWidth, core.Tokens(c).Border).Children(func() {
		board := project.KanbanBoard(c, project.KanbanBoardOptions{
			Columns: boardColumns(c), Tasks: tasks, Selected: &selected, Prefix: "CB",
		})
		resultLine(c, "KanbanBoard.Moved() · Opened()",
			quoteOrEmpty(board.Moved())+" · "+quoteOrEmpty(board.Opened()))
	})

	showcase.Field(c, "SprintBoard — 同一块板，多一个倒计时")
	ui.Box(c).FillWidth().Height(unit(c, 66)).Radius(theme.ControlRadius).
		Border(theme.BorderWidth, core.Tokens(c).Border).Children(func() {
		project.SprintBoard(c, project.SprintBoardOptions{
			Sprint: "Sprint 14", DaysLeft: 3,
			Columns: boardColumns(c), Tasks: tasks, Selected: &selected, Prefix: "CB",
		})
	})
	showcase.Field(c, "同一个倒计时，三种颜色")
	ui.Row(c).Gap(unit(c, 1.5)).Children(func() {
		for _, d := range []int{9, 3, -2} {
			k := core.Tokens(c)
			_, fg := project.SprintTone(d).Pair(k)
			ui.Box(c).Padding(unit(c, 0.75), unit(c, 2)).Radius(theme.PillRadius).
				Background(k.Surface).Label(itoaOf(d) + " days left").Children(func() {
				ui.Text(c, itoaOf(d)+" days left").TextColor(fg).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold().Shrink(0)
			})
		}
	})
}

// ── the one task ─────────────────────────────────────────────────────────

// detailSection is a task opened up: everything that decides where it sits
// and who is on it, in the order the questions are asked.
func detailSection(c *ui.Context) {
	showcase.Section(c, "详情 · TaskDetailPanel")

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(58).Shrink(0).Children(func() {
			title := "Northgate Dental — no hot water"
			body := "Boiler lost pressure again. The tenant has no hot water " +
				"and the pressure drops again within a day of a repressurising."
			status := project.StatusReview
			assignee := "Mia Chen"
			today := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
			due := project.DueValue{
				Day: time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
				Set: true, Text: "Tomorrow",
			}
			toggled := showcase.State(c, "project.261.toggled", "")
			saved := showcase.State(c, "project.261.saved", false)
			detail := project.TaskDetailPanel(c, project.TaskDetailPanelOptions{
				Task: boardTasks()[2], Prefix: "CB", Title: &title, Body: &body,
				Status: &status, Assignee: &assignee, People: people(),
				Due: &due, Today: &today, Save: saved, SubtaskToggled: toggled,
				Subtasks: subtasks(),
			})
			resultLine(c, "TaskDetailPanel.Saved()", fmtBool(detail.Saved()))
		})

		ui.Column(c).WidthPercent(40).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "同一个面板，只读：没有状态、没有人、没有日期")
			title := "Harbour Cafe — oven door"
			body := "The hinge is bent. Parts are on order until the 18th."
			detail := project.TaskDetailPanel(c, project.TaskDetailPanelOptions{
				Task: boardTasks()[3], Prefix: "CB", Title: &title, Body: &body,
				Save: new(bool), Busy: true,
			})
			resultLine(c, "TaskDetailPanel.Saved()", fmtBool(detail.Saved()))
		})
	})
}

// ── the controls ─────────────────────────────────────────────────────────

// controlSection is everything a person touches to change a task, each drawn
// on its own so that a reader can tell which control belongs to which
// question.
func controlSection(c *ui.Context) {
	showcase.Section(c, "选择与编辑 · StatusSelect / AssigneePicker / DueDatePicker / LabelManager / SubtaskList")

	today := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	status := project.StatusDoing
	assignee := "Ravi Patel"
	due := project.DueValue{
		Day: time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
		Set: true, Text: "Fri 14 Mar",
	}

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(32).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "StatusSelect — 每个选项带自己的颜色")
			project.StatusSelect(c, project.StatusSelectOptions{
				Status: &status, Columns: boardColumns(c),
			})
			showcase.Field(c, "AssigneePicker — 名单长到记不住，所以能打字")
			project.AssigneePicker(c, project.AssigneeOptions{
				Value: &assignee, People: people(),
			})
			showcase.Field(c, "一个都没选的样子")
			project.AssigneePicker(c, project.AssigneeOptions{
				Value: &project_empty, People: people(),
			})
		})

		ui.Column(c).WidthPercent(32).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "DueDatePicker — 日期可以不存在")
			project.DueDatePicker(c, project.DueDatePickerOptions{
				Due: &due, Open: &dueOpen, Today: today,
				Min: today, Max: today.AddDate(0, 3, 0), Clearable: true,
			})
			showcase.Field(c, "没有日期的那一个")
			none := project.DueValue{}
			project.DueDatePicker(c, project.DueDatePickerOptions{
				Due: &none, Open: &noneOpen, Today: today, Clearable: true,
			})
			showcase.Field(c, "DueLabel 和 DueTone — 同一个值的三种读法")
			for _, d := range []project.DueValue{due, none, {
				Day: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
				Set: true, Text: "Yesterday",
			}} {
				dueToneRow(c, d, today)
			}
		})

		ui.Column(c).WidthPercent(32).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "LabelManager — 板认识的全部标签")
			label := showcase.State(c, "project.338.label", "")
			newLabel := showcase.State(c, "project.338.newLabel", "")
			manager := project.LabelManager(c, project.LabelManagerOptions{
				Labels:  []string{"hvac", "plumbing", "urgent", "recurring", "onsite", "warranty"},
				Toggled: label, Name: newLabel,
				Added: func(name string) {},
			})
			resultLine(c, "LabelManager.Toggled() · Added()",
				quoteOrEmpty(manager.Toggled())+" · "+quoteOrEmpty(manager.Added()))

			showcase.Field(c, "SubtaskList — 勾选框是库里那个，不是手画的方块")
			toggled := showcase.State(c, "project.348.toggled", "")
			done := showcase.State(c, "project.348.done", 2)
			total := showcase.State(c, "project.348.total", 4)
			list := project.SubtaskList(c, project.SubtaskListOptions{
				Subtasks: subtasks(), Toggled: toggled,
				Done: done, Total: total,
			})
			resultLine(c, "SubtaskList.Toggled()", quoteOrEmpty(list.Toggled()))
		})
	})
}

// ── the numbers ──────────────────────────────────────────────────────────

// progressSection is the three ways a sprint is measured against: one
// milestone, one person, and the whole thing.
func metricSection(c *ui.Context) {
	showcase.Section(c, "进度 · MilestoneProgress / WorkloadView / BurndownChart")

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(52).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "BurndownChart — 计划那条线是斜的，实际那条是数据")
			sprint := burndown()
			project.BurndownChart(c, project.BurndownChartOptions{
				Burndown: sprint, Height: unit(c, 56),
			})
			showcase.Field(c, "HideIdeal — 只有实际那条线的时候是什么样")
			sprint.Today = -1
			sprint.Days = nil
			project.BurndownChart(c, project.BurndownChartOptions{
				Burndown: sprint, Height: unit(c, 26), HideIdeal: true,
			})
		})

		ui.Column(c).WidthPercent(46).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "MilestoneProgress — 90% 是「快好了」，不是「出事了」")
			for _, m := range []struct {
				title     string
				done, all int
				due       string
			}{
				{"March release", 24, 24, "Shipped 6 March"},
				{"AC service contract", 22, 24, "Due 28 March"},
				{"Q2 onboarding", 3, 24, "Due 30 June"},
			} {
				project.MilestoneProgress(c, project.MilestoneProgressOptions{
					Title: m.title, Done: m.done, Total: m.all, Due: m.due,
				})
			}

			showcase.Field(c, "WorkloadView — 有 capacity 时是图，没有时是行")
			project.WorkloadView(c, project.WorkloadViewOptions{
				Entries: workload(), Height: unit(c, 44),
			})
			showcase.Field(c, "同一个视图，谁都没说自己能接多少")
			project.WorkloadView(c, project.WorkloadViewOptions{
				Entries: workload(),
				Empty: func() {
					display.Text(c, "Nobody has anything assigned.",
						display.TextOptions{Muted: true})
				},
			})
		})
	})
}

// ── the clock ────────────────────────────────────────────────────────────

// timeSection is the two clocks a person leaves running while they work.
func timeSection(c *ui.Context) {
	showcase.Section(c, "计时 · TimeTracker / PomodoroTimer")

	ui.Row(c).FillWidth().Gap(unit(c, 4)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(46).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "TimeTracker — 屏幕上的数字就是会记下来的那个")
			tracker := project.TimeTracker(c, project.TimeTrackerOptions{
				Running: &running, Elapsed: 1*time.Hour + 5*time.Minute,
				Started: func() {}, Stopped: func() {}, Logged: func() {},
				Entries: []project.TimeEntry{
					{Task: "CB-1042", Who: "Andre Thomson",
						Spent: 45 * time.Minute, Billable: true},
					{Task: "CB-1039", Who: "Mia Chen", Spent: 20 * time.Minute},
					{Task: "CB-1031", Who: "Ravi Patel",
						Spent: 3 * time.Hour, Billable: true},
				},
			})
			resultLine(c, "TimeTracker.Running()", fmtBool(tracker.Running()))
			showcase.Field(c, "停着的时候，Log time 是禁用的")
			project.TimeTracker(c, project.TimeTrackerOptions{
				Running: &idle, Elapsed: 0,
				Started: func() {}, Stopped: func() {}, Logged: func() {},
			})
		})

		ui.Column(c).WidthPercent(52).Shrink(0).Gap(unit(c, 2)).Children(func() {
			showcase.Field(c, "PomodoroTimer — 环是剩下的，不是走完的")
			// A frame of the dial's own proportions: a dial is drawn twice as
			// tall as the box it is given, so a timer given the whole column
			// draws a ring that runs over the buttons under it.
			ui.Box(c).Width(unit(c, 24)).Shrink(0).Children(func() {
				focusing := showcase.State(c, "project.446.focusing", true)
				broken := showcase.State(c, "project.446.broken", false)
				timer := project.PomodoroTimer(c, project.PomodoroOptions{
					Elapsed: 18 * time.Minute, Running: focusing,
					Started: func() {}, Stopped: func() {}, Reset: func() {},
				})
				resultLine(c, "PomodoroTimer.State()",
					phaseName(timer.State().Phase)+" · left "+
						project.FormatElapsed(timer.State().Left))
				showcase.Field(c, "休息那五分钟：环是它自己的长度，不是专注的")
				rest := project.PomodoroTimer(c, project.PomodoroOptions{
					Elapsed: 27*time.Minute + 2*time.Minute, Running: broken,
					Started: func() {}, Stopped: func() {}, Reset: func() {},
				})
				resultLine(c, "PomodoroTimer.State()",
					phaseName(rest.State().Phase)+" · left "+
						project.FormatElapsed(rest.State().Left))
			})
		})
	})
}

// ── the functions ────────────────────────────────────────────────────────

// pureSection is the part of the package with nothing a person presses: the
// rules the components above are made of.
func projectPureSection(c *ui.Context) {
	showcase.Section(c, "纯函数 · 只调不测的那一半")

	ui.Row(c).FillWidth().Gap(unit(c, 6)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(50).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			specimen(c, `FormatIssueID("cb", 1042)`, project.FormatIssueID("cb", 1042))
			specimen(c, `FormatIssueID("", 0)`, quoteOrEmpty(project.FormatIssueID("", 0)))
			specimen(c, `PrefixIssueID("cb")`, project.PrefixIssueID("cb"))
			specimen(c, `FormatIssueID("cb", 0)`, project.FormatIssueID("cb", 0))
			specimen(c, `Task{ID: "1042"}.IDNumber()`,
				itoaOf(project.Task{ID: "1042"}.IDNumber()))
			specimen(c, `Task{ID: "cb-2871-x9"}.IDNumber()`,
				itoaOf(project.Task{ID: "cb-2871-x9"}.IDNumber()))
			specimen(c, `FormatElapsed(1*time.Hour + 5*time.Minute)`,
				project.FormatElapsed(1*time.Hour+5*time.Minute))
			specimen(c, `FormatElapsed(9*time.Second)`, project.FormatElapsed(9*time.Second))
			specimen(c, `FormatElapsed(26*time.Hour)`, project.FormatElapsed(26*time.Hour))
			specimen(c, `MilestoneFraction(18, 24)`, fmtFloat(project.MilestoneFraction(18, 24)))
			specimen(c, `MilestoneFraction(0, 0)`, fmtFloat(project.MilestoneFraction(0, 0)))
		})

		ui.Column(c).WidthPercent(50).Shrink(0).Gap(unit(c, 1.5)).Children(func() {
			specimen(c, `IdealBurndown(10, 50)`, joinFloats(project.IdealBurndown(10, 50)))
			specimen(c, `ActualBurndown(days, 10)`, joinFloats(project.ActualBurndown(sprintDays(), 10)))
			specimen(c, `SprintLength(burndown)`, itoaOf(project.SprintLength(burndown())))
			// The points a chart is fed: one x per day, from zero. The function
			// that builds them is private to the package, so what is on the
			// page is the run of values above it and the chart that drew them.
			specimen(c, "chart points for those values",
				"[{0 50} {1 45} {2 40} … {9 5}]")
			at := project.PomodoroAt(27*time.Minute, 25, 5)
			specimen(c, `PomodoroAt(27*time.Minute, 25, 5)`,
				phaseName(at.Phase)+" · left "+project.FormatElapsed(at.Left))
			at2 := project.PomodoroAt(2*time.Minute, 25, 5)
			specimen(c, `PomodoroAt(2*time.Minute, 25, 5)`,
				phaseName(at2.Phase)+" · left "+project.FormatElapsed(at2.Left)+
					" · fraction "+fmtFloat(at2.Fraction))
			specimen(c, `MilestoneTone(22, 24)`, project.MilestoneTone(22, 24).String())
			specimen(c, `SprintTone(3)`, project.SprintTone(3).String())
			specimen(c, `StatusSeverity("blocked")`, project.StatusSeverity("blocked").String())
		})
	})
}

// ── the data ─────────────────────────────────────────────────────────────

// boardTasks is the seven tasks every drawing of a task on this page uses.
// One set, written once: a gallery whose board and whose detail panel showed
// different tasks would be two galleries in a trench coat.
func boardTasks() []project.Task {
	return []project.Task{
		{ID: "1042", Title: "Riverside Clinic — AC not cooling",
			Status: project.StatusDoing, Assignee: "Andre Thomson",
			Labels: []string{"hvac", "urgent"}, Due: "Today", Priority: 2,
			Subtasks: 4, SubtasksDone: 2, Comments: 3},
		{ID: "1039", Title: "Elm Street Gym — boiler losing pressure",
			Status: project.StatusBacklog, Assignee: "Mia Chen",
			Labels: []string{"plumbing"}, Due: "Fri 20 Mar",
			Subtasks: 2, SubtasksDone: 0},
		{ID: "1037", Title: "Northgate Dental — no hot water",
			Status: project.StatusReview, Assignee: "Mia Chen",
			Labels: []string{"plumbing", "recurring"}, Due: "Tomorrow",
			Priority: 1, Comments: 8},
		{ID: "1031", Title: "Harbour Cafe — oven door hinge",
			Status: project.StatusBacklog, Labels: []string{"appliance"},
			Due: "18 Mar"},
		{ID: "1028", Title: "Pinewood Dental — thermostat replacement",
			Status: project.StatusDoing, Assignee: "Ravi Patel",
			Labels: []string{"hvac", "warranty"}, Due: "Mon 16 Mar", Priority: 3},
		{ID: "1019", Title: "Harbour Cafe — fridge not cooling",
			Status: project.StatusBlocked, Assignee: "Andre Thomson",
			Labels: []string{"refrigeration"}, Due: "overdue", Overdue: true},
		{ID: "1012", Title: "Riverside Clinic — quarterly service",
			Status: project.StatusDone, Assignee: "Ravi Patel",
			Labels: []string{"contract"}, Due: "6 Mar", Subtasks: 5, SubtasksDone: 5},
	}
}

// boardColumns is three lanes rather than six, so that every one of them is
// on the page. A fourth would be off the right edge, which is a fact about
// the horizontal scroll rather than about the component.
func boardColumns(c *ui.Context) []project.Column {
	return []project.Column{
		{Status: project.StatusBacklog, Title: "Backlog", Limit: 4},
		{Status: project.StatusDoing, Title: "In progress", Menu: "⋯"},
		{Status: project.StatusReview, Title: "In review", Empty: func() {
			ui.Column(c).Grow(1).Center().Gap(unit(c, 1)).Children(func() {
				ui.Text(c, "Nothing waiting on review").TextColor(core.Tokens(c).TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize))
				ui.Text(c, "拖到这里的卡片会等一个人看").TextColor(core.Tokens(c).TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}},
	}
}

func projectStatuses() []project.Status {
	return []project.Status{
		project.StatusBacklog, project.StatusTodo, project.StatusDoing,
		project.StatusReview, project.StatusBlocked, project.StatusDone,
	}
}

func people() []string {
	return []string{"Andre Thomson", "Mia Chen", "Ravi Patel", "Lena Ford", "Sam Ortiz"}
}

func subtasks() []project.Subtask {
	return []project.Subtask{
		{ID: "s-1", Title: "Photograph the gauge reading", Done: true,
			Assignee: "Mia Chen"},
		{ID: "s-2", Title: "Check the expansion vessel", Done: true},
		{ID: "s-3", Title: "Book a second visit", Assignee: "Mia Chen"},
		{ID: "s-4", Title: "Tell the letting agent in writing"},
	}
}

func workload() []project.WorkloadEntry {
	return []project.WorkloadEntry{
		{Who: "Andre Thomson", Assigned: 6, Done: 3, Capacity: 6},
		{Who: "Mia Chen", Assigned: 4, Done: 2, Capacity: 5},
		{Who: "Ravi Patel", Assigned: 7, Done: 6, Capacity: 6},
		{Who: "Lena Ford", Assigned: 2, Done: 0},
	}
}

// burndown is ten days of a sprint that ran a little behind the plan: the
// shape is the point, and a sprint that hit it exactly would make the gap
// between the two lines look like a rendering artefact.
func burndown() project.Burndown {
	return project.Burndown{
		Name: "Sprint 14", Scope: 50, Today: 9, Unit: "points",
		Days: sprintDays(),
	}
}

func sprintDays() []project.BurndownDay {
	// Day 5 and day 6 are missing on purpose: a weekend is a gap in a
	// burndown, not a line at zero.
	return []project.BurndownDay{
		{Day: 0, Remaining: 50}, {Day: 1, Remaining: 46}, {Day: 2, Remaining: 41},
		{Day: 3, Remaining: 38}, {Day: 4, Remaining: 33}, {Day: 7, Remaining: 26},
		{Day: 8, Remaining: 21}, {Day: 9, Remaining: 17},
	}
}

// ── small pieces ─────────────────────────────────────────────────────────

// dueToneRow is one due date as the two components read it: the words beside
// the title, and the colour the whole row is drawn at.
func dueToneRow(c *ui.Context, due project.DueValue, now time.Time) {
	k := core.Tokens(c)
	sev := project.DueTone(due, now)
	_, fg := sev.Pair(k)
	label := project.DueLabel(projectDetailDue)

	ui.Row(c).FillWidth().Gap(unit(c, 2)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "DueLabel").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize)).Width(unit(c, 14)).Shrink(0)
		ui.Text(c, label).TextColor(k.Text).Width(unit(c, 16)).Shrink(0).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		ui.Text(c, "DueTone").TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize)).Width(unit(c, 14)).Shrink(0)
		ui.Box(c).Padding(unit(c, 0.5), unit(c, 1.5)).Radius(theme.PillRadius).
			Background(k.Surface).Label(sev.String()).Children(func() {
			ui.Text(c, sev.String()).TextColor(fg).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		})
	})
}

func phaseName(p project.PomodoroPhase) string {
	if p == project.Break {
		return "Break"
	}
	return "Focus"
}

// joinFloats is a run of numbers as a person writes them down, so a table of
// pure functions has something to read on the right of the arrow.
func joinFloats(vals []float64) string {
	out := "["
	for i, v := range vals {
		if i > 0 {
			out += " "
		}
		out += trimFloat(v)
	}
	return out + "]"
}

func trimFloat(f float64) string {
	tenths := int64(f*10 + 0.5)
	whole := tenths / 10
	frac := tenths % 10
	if frac == 0 {
		return itoaOf(int(whole))
	}
	return itoaOf(int(whole)) + "." + itoaOf(int(frac))
}

func fmtFloat(f float32) string { return trimFloat(float64(f)) }
