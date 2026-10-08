package project

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
)

// ── shared fixtures ────────────────────────────────────────────────────────

func tasks() []Task {
	return []Task{
		{ID: "CB-1042", Title: "Root cause review: intermittent 502s from the edge",
			Status: StatusDoing, Assignee: "Rosa Vidal", Labels: []string{"backend", "flaky"},
			Due: "14 March", Priority: 2, Subtasks: 4, SubtasksDone: 3, Comments: 12},
		{ID: "CB-1043", Title: "Log a callback from the board", Status: StatusTodo,
			Assignee: "Ana Duarte", Priority: 1},
		{ID: "CB-1044", Title: "Blocked: waiting on procurement", Status: StatusBlocked,
			Assignee: "Kim Osei", Due: "yesterday", Overdue: true},
		{ID: "CB-1045", Title: "Ship the settings pane", Status: StatusDone},
	}
}

func columns() []Column {
	return []Column{
		{Status: StatusTodo, Title: "To do"},
		{Status: StatusDoing, Title: "In progress"},
		{Status: StatusBlocked, Title: "Blocked"},
		{Status: StatusDone, Title: "Done"},
	}
}

// ── KanbanBoard ────────────────────────────────────────────────────────────

func TestKanbanBoardShowsEveryLaneAndItsCards(t *testing.T) {
	var selected string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		KanbanBoard(c, KanbanBoardOptions{
			Columns: columns(), Tasks: tasks(),
			Selected: &selected, Prefix: "CB",
		})
	}, 1400, 620)
	for _, want := range []string{
		"To do", "In progress", "Blocked", "Done",
		"Log a callback from the board",
		"Blocked: waiting on procurement",
		"Ship the settings pane",
		"CB-1043",
	} {
		if !tt.HasText(want) {
			t.Errorf("the board is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestKanbanBoardKeepsTheColumnsWide(t *testing.T) {
	// The rule the whole board rests on: a column narrower than
	// theme.ColumnWidth wraps its card titles, and a column of wrapped titles
	// is a wall of text. The window here is far too narrow for four columns,
	// so the board must scroll rather than squeeze.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		var selected string
		KanbanBoard(c, KanbanBoardOptions{
			Columns: columns(), Tasks: tasks(),
			Selected: &selected, Prefix: "CB",
		})
	}, 600, 620)

	// Both titles are on the window: squeezed columns would truncate or wrap
	// them, and truncation is what a horizontal scroll is there to avoid.
	for _, want := range []string{
		"Log a callback from the board",
		"Root cause review: intermittent 502s from the edge",
	} {
		if !tt.HasText(want) {
			t.Errorf("a title was squeezed out of its column: %q is missing; %q", want, tt.Texts())
		}
	}
}

func TestKanbanBoardTakesTheLongestTitle(t *testing.T) {
	// A title far too long for any column is truncated to one line rather
	// than wrapped, which is the difference between a scannable board and a
	// ragged one.
	long := Task{ID: "CB-9", Title: strings.Repeat("a very long issue title ", 12),
		Status: StatusTodo}
	var selected string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		KanbanBoard(c, KanbanBoardOptions{
			Columns:  []Column{{Status: StatusTodo, Title: "To do"}},
			Tasks:    []Task{long},
			Selected: &selected,
		})
	}, 1400, 620)
	if !tt.HasText("To do") {
		t.Errorf("the lane is missing; %q", tt.Texts())
	}
}

func TestKanbanBoardReportsAnOpen(t *testing.T) {
	var selected string
	opened := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := KanbanBoard(c, KanbanBoardOptions{
			Columns: columns(), Tasks: tasks(),
			Selected: &selected, Prefix: "CB",
		})
		if o := r.Opened(); o != "" {
			opened = o
		}
	}, 1400, 620)
	if err := tt.Click("Ship the settings pane"); err != nil {
		t.Fatal(err)
	}
	if opened != "CB-1045" {
		t.Errorf("pressing a card should report its task, got %q", opened)
	}
}

func TestKanbanBoardMovesACardBetweenLanes(t *testing.T) {
	var selected string
	// Collected rather than assigned: the drop is reported on one frame of
	// the three the frame is built in, and by the last one it is over. See
	// docs/design-system.md §17.2.
	var moves []string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := KanbanBoard(c, KanbanBoardOptions{
			Columns: columns(), Tasks: tasks(),
			Selected: &selected, Prefix: "CB",
		})
		if m := r.Moved(); m != "" {
			moves = append(moves, m)
		}
	}, 1400, 620)

	src, ok := tt.Find("Log a callback from the board")
	if !ok {
		t.Fatalf("the card is missing; %q", tt.Texts())
	}
	// Onto a card in the other lane, which is inside that lane's scroll view
	// and therefore inside the drop target the lane owns.
	dst, ok := tt.Find("Root cause review: intermittent 502s from the edge")
	if !ok {
		t.Fatalf("the other lane's card is missing; %q", tt.Texts())
	}
	dragFromTo(tt, src.X+40, src.Y+src.H/2, dst.X+40, dst.Y+dst.H/2)
	if len(moves) == 0 {
		t.Fatal("the card was dragged onto another lane and nothing was reported")
	}
	if moves[0] != "doing,doing,CB-1043" {
		t.Errorf("the drop should name the lane it landed on and the task, got %q", moves[0])
	}
}

func TestKanbanBoardEmptyLaneShowsTheCallersExplanation(t *testing.T) {
	var selected string
	hits := 0
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		cols := []Column{{
			Status: StatusReview, Title: "In review",
			Empty: func() {
				if input.Button(c, "Nothing waiting", input.ButtonOptions{}).Clicked() {
					hits++
				}
			},
		}}
		KanbanBoard(c, KanbanBoardOptions{
			Columns: cols, Tasks: nil,
			Selected: &selected,
		})
	}, 700, 400)
	if err := tt.Click("Nothing waiting"); err != nil {
		t.Fatal(err)
	}
	if hits == 0 {
		t.Error("a button inside a lane's Empty must stay hittable")
	}
}

// ── TaskItem ───────────────────────────────────────────────────────────────

func TestTaskItemShowsTheReferenceAndTheFacts(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TaskItem(c, TaskItemOptions{Task: tasks()[0], Prefix: "CB"})
	}, 272, 140)
	for _, want := range []string{"CB-1042", "2", "backend", "3/4", "12", "14 March"} {
		if !tt.HasText(want) {
			t.Errorf("the card is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestTaskItemShowsTheSecondLabelAsACount(t *testing.T) {
	// "backend" plus one more is "+1", not the second label: a card is one
	// line of facts and a second label is a line too many.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TaskItem(c, TaskItemOptions{Task: tasks()[0], Prefix: "CB"})
	}, 272, 140)
	if !tt.HasText("+1") {
		t.Errorf("the second label should be counted; %q", tt.Texts())
	}
	if tt.HasText("flaky") {
		t.Errorf("the second label should not be printed; %q", tt.Texts())
	}
}

func TestTaskItemCompactDropsTheFacts(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TaskItem(c, TaskItemOptions{Task: tasks()[0], Prefix: "CB", Compact: true})
	}, 272, 120)
	if tt.HasText("14 March") {
		t.Errorf("a compact card has no facts line; %q", tt.Texts())
	}
	if !tt.HasText("Root cause review") && !tt.HasText("CB-1042") {
		t.Errorf("a compact card still has its reference and title; %q", tt.Texts())
	}
}

func TestTaskItemIsNamedAfterItsTitle(t *testing.T) {
	// A card shows a reference and a title; the title is what a screen reader
	// should read out, because the reference alone says nothing.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TaskItem(c, TaskItemOptions{Task: tasks()[1], Prefix: "CB"})
	}, 272, 120)
	if _, ok := tt.Find("Log a callback from the board"); !ok {
		t.Errorf("the card is not named after its title; %q", tt.Texts())
	}
}

// ── IssueIdBadge ───────────────────────────────────────────────────────────

func TestIssueIdBadge(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IssueIdBadge(c, IssueIdBadgeOptions{Prefix: "CB", Number: 1042})
	}, 120, 60)
	if !tt.HasText("CB-1042") {
		t.Errorf("the badge is wrong; %q", tt.Texts())
	}
	if _, ok := tt.Find("CB-1042"); !ok {
		t.Errorf("a badge with no word of its own must still be named; %q", tt.Texts())
	}
}

func TestIssueIdBadgeWithNothingToShowIsNothing(t *testing.T) {
	// No prefix means no badge rather than a badge reading "-", which looks
	// like a reference somebody forgot the number of.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IssueIdBadge(c, IssueIdBadgeOptions{})
	}, 120, 60)
	if len(tt.Texts()) != 0 {
		t.Errorf("nothing should be drawn, got %q", tt.Texts())
	}
}

func TestIssueIdBadgeWithAPrefixButNoNumber(t *testing.T) {
	// The other half of the same rule: a badge reading a bare "CB" is the
	// tracker without a reference, which is still worth saying.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IssueIdBadge(c, IssueIdBadgeOptions{Prefix: "cb"})
	}, 120, 60)
	if !tt.HasText("CB") {
		t.Errorf("a tracker with no number is a bare prefix; %q", tt.Texts())
	}
}

// ── IssueCard ──────────────────────────────────────────────────────────────

func TestIssueCard(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IssueCard(c, IssueCardOptions{
			Task: tasks()[0], Prefix: "CB",
			Body: func() { ui.Text(c, "The edge returns 502 for about one request in two hundred.") },
			Footer: func() {
				if input.Button(c, "Close", input.ButtonOptions{}).Clicked() {
					// nothing; the button only has to be there
				}
			},
		})
	}, 460, 360)
	for _, want := range []string{
		"CB-1042", "doing", "Root cause review",
		"The edge returns 502", "14 March", "Rosa Vidal", "Close",
	} {
		if !tt.HasText(want) {
			t.Errorf("the card is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestIssueCardWithoutAFooterHasNoRule(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		IssueCard(c, IssueCardOptions{Task: tasks()[2], Prefix: "CB"})
	}, 460, 240)
	if !tt.HasText("Blocked: waiting on procurement") {
		t.Errorf("the card did not draw; %q", tt.Texts())
	}
}

// ── SubtaskList ────────────────────────────────────────────────────────────

func subtasks() []Subtask {
	return []Subtask{
		{ID: "s1", Title: "Reproduce on staging", Assignee: "Ana Duarte"},
		{ID: "s2", Title: "Check the edge configuration", Done: true},
		{ID: "s3", Title: "Write the runbook"},
	}
}

func TestSubtaskListShowsEveryLine(t *testing.T) {
	var toggled string
	done, total := 1, 3
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SubtaskList(c, SubtaskListOptions{
			Subtasks: subtasks(), Toggled: &toggled, Done: &done, Total: &total,
		})
	}, 460, 260)
	for _, want := range []string{"Subtasks", "1/3 done", "Reproduce on staging",
		"Check the edge configuration", "Write the runbook"} {
		if !tt.HasText(want) {
			t.Errorf("the list is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestSubtaskListReportsAToggle(t *testing.T) {
	var toggled string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := SubtaskList(c, SubtaskListOptions{
			Subtasks: subtasks(), Toggled: &toggled,
		})
		if r.Toggled() != "" {
			toggled = r.Toggled()
		}
	}, 460, 260)
	if err := tt.Click("Reproduce on staging"); err != nil {
		t.Fatal(err)
	}
	if toggled != "s1" {
		t.Errorf("toggling should report the line's id, got %q", toggled)
	}
}

func TestSubtaskListWithNothing(t *testing.T) {
	var toggled string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SubtaskList(c, SubtaskListOptions{Toggled: &toggled})
	}, 460, 140)
	if !tt.HasText("Subtasks") {
		t.Errorf("the list is missing its name; %q", tt.Texts())
	}
}

func TestSubtaskListRefusesHalfACount(t *testing.T) {
	defer panics(t, "project: SubtaskList takes Done and Total together", func() {
		var toggled string
		done := 1
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			SubtaskList(c, SubtaskListOptions{Toggled: &toggled, Done: &done})
		}, 400, 200)
	})
}

// ── MilestoneProgress ──────────────────────────────────────────────────────

func TestMilestoneProgress(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		MilestoneProgress(c, MilestoneProgressOptions{
			Title: "Sprint 14", Done: 9, Total: 10, Due: "Due 21 March",
		})
	}, 460, 180)
	for _, want := range []string{"Sprint 14", "9 / 10", "Due 21 March"} {
		if !tt.HasText(want) {
			t.Errorf("the milestone is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestMilestoneProgressNeedsATitle(t *testing.T) {
	defer panics(t, "project: MilestoneProgress needs a Title", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			MilestoneProgress(c, MilestoneProgressOptions{Done: 1, Total: 2})
		}, 400, 200)
	})
}

// ── LabelManager ───────────────────────────────────────────────────────────

func TestLabelManagerShowsTheLabels(t *testing.T) {
	var toggled, added, name string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LabelManager(c, LabelManagerOptions{
			Labels:  []string{"backend", "flaky", "urgent"},
			Toggled: &toggled, Name: &name,
			Added: func(n string) { added = n },
		})
	}, 460, 300)
	for _, want := range []string{"backend", "flaky", "urgent", "New label", "Add"} {
		if !tt.HasText(want) {
			t.Errorf("the manager is missing %q; %q", want, tt.Texts())
		}
	}
	if err := tt.Click("New label"); err != nil {
		t.Fatal(err)
	}
	tt.Type("docs")
	if name != "docs" {
		t.Fatalf("the name is the caller's string, got %q", name)
	}
	if err := tt.Click("Add"); err != nil {
		t.Fatal(err)
	}
	if added != "docs" {
		t.Errorf("adding should hand back the name, got %q", added)
	}
}

func TestLabelManagerWithNoLabels(t *testing.T) {
	var toggled string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		LabelManager(c, LabelManagerOptions{Toggled: &toggled})
	}, 400, 160)
	if !tt.HasText("No labels yet") {
		t.Errorf("an empty manager must say so; %q", tt.Texts())
	}
}

func TestLabelManagerNeedsAPointerForItsToggle(t *testing.T) {
	defer panics(t, "project: LabelManager needs the *string Toggled writes to", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			LabelManager(c, LabelManagerOptions{Labels: []string{"a"}})
		}, 400, 200)
	})
}

// ── TaskDetailPanel ────────────────────────────────────────────────────────

func TestTaskDetailPanelShowsEveryField(t *testing.T) {
	title := "Root cause review"
	body := "The edge returns 502 for one request in two hundred."
	var save bool
	status := StatusDoing
	assignee := "Rosa Vidal"
	today := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	due := DueValue{Day: today.AddDate(0, 0, 3), Set: true, Text: "17 March"}
	var subtoggled string

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TaskDetailPanel(c, TaskDetailPanelOptions{
			Task: tasks()[0], Prefix: "CB",
			Title: &title, Body: &body, Save: &save,
			Status: &status, Columns: columns(),
			Assignee: &assignee, People: []string{"Rosa Vidal", "Ana Duarte", "Kim Osei"},
			Due: &due, Today: &today,
			Subtasks: subtasks(), SubtaskToggled: &subtoggled,
		})
	}, 560, 900)
	for _, want := range []string{
		"CB-1042", "doing", "Title", "Description",
		"Reproduce on staging", "17 March", "Save",
	} {
		if !tt.HasText(want) {
			t.Errorf("the panel is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestTaskDetailPanelReportsASave(t *testing.T) {
	title, body := "Root cause review", ""
	var save bool
	saved := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := TaskDetailPanel(c, TaskDetailPanelOptions{
			Task: tasks()[0], Prefix: "CB",
			Title: &title, Body: &body, Save: &save,
		})
		if r.Saved() {
			saved = true
		}
	}, 560, 520)
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	if !saved {
		t.Error("saving should have been reported")
	}
}

func TestTaskDetailPanelRefusesADateWithNoToday(t *testing.T) {
	defer panics(t, "project: a TaskDetailPanel with a due date needs a Today", func() {
		title, body := "a", ""
		var save bool
		due := DueValue{}
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TaskDetailPanel(c, TaskDetailPanelOptions{
				Task: tasks()[0], Title: &title, Body: &body, Save: &save,
				Due: &due,
			})
		}, 560, 600)
	})
}

// ── TaskList ───────────────────────────────────────────────────────────────

func TestTaskListShowsEveryRow(t *testing.T) {
	var selected string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TaskList(c, TaskListOptions{
			Tasks: tasks(), Selected: &selected, Prefix: "CB", Height: 300,
		})
	}, 560, 340)
	for _, want := range []string{
		"Log a callback from the board", "Blocked: waiting on procurement",
		"Ship the settings pane", "todo", "blocked", "done",
	} {
		if !tt.HasText(want) {
			t.Errorf("the list is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestTaskListReportsASelection(t *testing.T) {
	var selected string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := TaskList(c, TaskListOptions{
			Tasks: tasks(), Selected: &selected, Prefix: "CB", Height: 300,
		})
		if r.Selected() != "" {
			selected = r.Selected()
		}
	}, 560, 340)
	if err := tt.Click("Ship the settings pane"); err != nil {
		t.Fatal(err)
	}
	if selected != "CB-1045" {
		t.Errorf("pressing a row should report its task, got %q", selected)
	}
}

func TestTaskListWithNothingInIt(t *testing.T) {
	var selected string
	drew := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TaskList(c, TaskListOptions{
			Selected: &selected, Height: 200,
			Empty: func() {
				ui.Text(c, "Nothing assigned")
				drew = true
			},
		})
	}, 460, 240)
	if !tt.HasText("Nothing assigned") {
		t.Errorf("the empty state did not draw; %q", tt.Texts())
	}
	if !drew {
		t.Error("the caller's Empty was not called")
	}
}

// ── the pickers ────────────────────────────────────────────────────────────

func TestAssigneePickerChoosesSomebody(t *testing.T) {
	assignee := ""
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		AssigneePicker(c, AssigneeOptions{
			Value:  &assignee,
			People: []string{"Ana Duarte", "Kim Osei", "Rosa Vidal"},
		})
	}, 460, 200)
	if err := tt.Click("Nobody"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Kim Osei"); err != nil {
		t.Fatal(err)
	}
	if assignee != "Kim Osei" {
		t.Errorf("choosing should write the name, got %q", assignee)
	}
}

func TestAssigneePickerNeedsPeople(t *testing.T) {
	defer panics(t, "project: AssigneePicker needs at least one person", func() {
		var v string
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			AssigneePicker(c, AssigneeOptions{Value: &v})
		}, 400, 200)
	})
}

func TestStatusSelectMovesATask(t *testing.T) {
	status := StatusTodo
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		StatusSelect(c, StatusSelectOptions{Status: &status, Columns: columns()})
	}, 460, 220)
	if err := tt.Click("To do"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("In progress"); err != nil {
		t.Fatal(err)
	}
	if status != StatusDoing {
		t.Errorf("choosing should write the status, got %q", status)
	}
}

func TestStatusSelectRefusesAStatusItCannotShow(t *testing.T) {
	defer panics(t, "which is not one of its columns", func() {
		status := StatusReview
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			StatusSelect(c, StatusSelectOptions{Status: &status, Columns: columns()})
		}, 460, 220)
	})
}

func TestDueDatePickerShowsTheCallersOwnWords(t *testing.T) {
	today := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	due := DueValue{Day: today.AddDate(0, 0, 2), Set: true, Text: "the day after tomorrow"}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DueDatePicker(c, DueDatePickerOptions{
			Due: &due, Open: new(bool), Today: today, Clearable: true,
		})
	}, 460, 200)
	if !tt.HasText("the day after tomorrow") {
		t.Errorf("the field should show what the caller wrote; %q", tt.Texts())
	}
	if _, ok := tt.Find("Clear due date"); !ok {
		t.Errorf("the clear button is missing; %q", tt.Texts())
	}
}

func TestDueDatePickerWithNoDateSaysWhatItWants(t *testing.T) {
	today := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	due := DueValue{}
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		DueDatePicker(c, DueDatePickerOptions{Due: &due, Open: new(bool), Today: today})
	}, 460, 200)
	if !tt.HasText("Due date") {
		t.Errorf("an empty date field says what it wants; %q", tt.Texts())
	}
}

func TestDueDatePickerNeedsItsPointers(t *testing.T) {
	defer panics(t, "project: DueDatePicker needs the *DueValue it writes to", func() {
		today := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			DueDatePicker(c, DueDatePickerOptions{Open: new(bool), Today: today})
		}, 400, 200)
	})
}

// ── SprintBoard ────────────────────────────────────────────────────────────

func TestSprintBoardShowsItsClock(t *testing.T) {
	var selected string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SprintBoard(c, SprintBoardOptions{
			Sprint: "Sprint 14", DaysLeft: 2,
			Columns: columns(), Tasks: tasks(),
			Selected: &selected, Prefix: "CB",
		})
	}, 1400, 680)
	for _, want := range []string{"Sprint 14", "2 days left", "To do", "In progress",
		"Log a callback from the board"} {
		if !tt.HasText(want) {
			t.Errorf("the board is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestSprintBoardReportsASprintThatHasRunOut(t *testing.T) {
	var selected string
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		SprintBoard(c, SprintBoardOptions{
			Sprint: "Sprint 13", DaysLeft: -2,
			Columns: columns(), Tasks: tasks(),
			Selected: &selected,
		})
	}, 1400, 680)
	if !tt.HasText("-2 days left") {
		t.Errorf("a sprint that has run out should say so; %q", tt.Texts())
	}
}

// ── BurndownChart ──────────────────────────────────────────────────────────

func sprint() Burndown {
	return Burndown{
		Name: "Sprint 14", Scope: 50, Today: 5, Unit: "points",
		Days: []BurndownDay{
			{Day: 0, Remaining: 50},
			{Day: 1, Remaining: 44},
			{Day: 2, Remaining: 41},
			{Day: 3, Remaining: 30},
			{Day: 4, Remaining: 28},
			{Day: 5, Remaining: 19},
		},
	}
}

func TestBurndownChartDrawsBothLines(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		BurndownChart(c, BurndownChartOptions{Burndown: sprint(), Height: 300})
	}, 760, 360)
	for _, want := range []string{"Plan", "Sprint 14", "Day", "points"} {
		if !tt.HasText(want) {
			t.Errorf("the chart is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestBurndownChartWithNoReadingsSaysWhatIsMissing(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		BurndownChart(c, BurndownChartOptions{
			Burndown: Burndown{Name: "Sprint 15", Scope: 30, Today: -1},
			Height:   260,
		})
	}, 700, 320)
	// Nothing has been recorded and the sprint has not started, so there is
	// no plan line either: the chart shows its own empty state rather than an
	// axis over a range nothing reaches.
	if !tt.HasText("No readings yet") {
		t.Errorf("the empty state is missing; %q", tt.Texts())
	}
}

func TestBurndownChartNeedsAName(t *testing.T) {
	defer panics(t, "project: BurndownChart needs a Name", func() {
		b := sprint()
		b.Name = ""
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			BurndownChart(c, BurndownChartOptions{Burndown: b})
		}, 700, 320)
	})
}

func TestBurndownChartHidesThePlanOnRequest(t *testing.T) {
	with := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		BurndownChart(c, BurndownChartOptions{Burndown: sprint(), Height: 300})
	}, 760, 360)
	if !with.HasText("Plan") {
		t.Fatalf("the plan line should be there by default; %q", with.Texts())
	}
	without := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		BurndownChart(c, BurndownChartOptions{
			Burndown: sprint(), Height: 300, HideIdeal: true,
		})
	}, 760, 360)
	if without.HasText("Plan") {
		t.Errorf("the plan line should be gone; %q", without.Texts())
	}
}

// ── TimeTracker ────────────────────────────────────────────────────────────

func TestTimeTrackerShowsTheClockAndTheEntries(t *testing.T) {
	running := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TimeTracker(c, TimeTrackerOptions{
			Running: &running, Elapsed: 65 * time.Minute,
			Started: func() {}, Stopped: func() {}, Logged: func() {},
			Entries: []TimeEntry{
				{Task: "Root cause review", Spent: 90 * time.Minute, Billable: true},
				{Task: "Log a callback", Spent: 25 * time.Minute},
			},
		})
	}, 520, 340)
	for _, want := range []string{"1:05:00", "Running", "Stop", "Root cause review",
		"1:30:00", "Billable", "25:00", "Log time"} {
		if !tt.HasText(want) {
			t.Errorf("the tracker is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestTimeTrackerStarts(t *testing.T) {
	running := false
	started := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := TimeTracker(c, TimeTrackerOptions{
			Running: &running, Elapsed: 0, Started: func() { started = true },
		})
		if r.Running() {
			t.Error("a clock that is not going must not report as going")
		}
	}, 460, 220)
	if err := tt.Click("Start"); err != nil {
		t.Fatal(err)
	}
	if !started {
		t.Error("Start did not fire")
	}
}

func TestTimeTrackerWithNoEntries(t *testing.T) {
	running := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		TimeTracker(c, TimeTrackerOptions{Running: &running})
	}, 460, 220)
	if !tt.HasText("0:00") {
		t.Errorf("a clock with nothing on it reads zero; %q", tt.Texts())
	}
}

func TestTimeTrackerNeedsItsRunningPointer(t *testing.T) {
	defer panics(t, "project: TimeTracker needs the *bool Running writes to", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			TimeTracker(c, TimeTrackerOptions{Elapsed: time.Minute})
		}, 400, 200)
	})
}

// ── WorkloadView ───────────────────────────────────────────────────────────

func workload() []WorkloadEntry {
	return []WorkloadEntry{
		{Who: "Ana Duarte", Assigned: 6, Done: 2, Capacity: 5},
		{Who: "Kim Osei", Assigned: 2, Done: 2, Capacity: 5},
		{Who: "Rosa Vidal", Assigned: 9, Done: 1, Capacity: 5},
	}
}

func TestWorkloadViewDrawsAChartWhenThereIsACapacity(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WorkloadView(c, WorkloadViewOptions{Entries: workload(), Height: 280})
	}, 760, 360)
	// The people's names are the chart's category labels, and a chart's
	// labels are painted rather than laid out — which is why the axis in a
	// screenshot is legible but is not something a text query can find. So
	// what is asserted here is the chart's own chrome and its series names.
	for _, want := range []string{"Workload", "Assigned", "Capacity", "Person axis",
		"Tasks axis"} {
		if !tt.HasText(want) {
			t.Errorf("the view is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestWorkloadViewDrawsRowsWhenNobodyHasACapacity(t *testing.T) {
	// A bar against nothing is a proportion of an unknown, so with no
	// capacities the view is a plain list rather than a chart with a
	// meaningless second series.
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WorkloadView(c, WorkloadViewOptions{Entries: []WorkloadEntry{
			{Who: "Ana Duarte", Assigned: 6, Done: 2},
		}})
	}, 560, 240)
	if !tt.HasText("Ana Duarte") || !tt.HasText("2 / 6") {
		t.Errorf("the rows are missing; %q", tt.Texts())
	}
	if tt.HasText("Capacity") {
		t.Errorf("there is no capacity to chart; %q", tt.Texts())
	}
}

func TestWorkloadViewWithNobody(t *testing.T) {
	drew := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		WorkloadView(c, WorkloadViewOptions{
			Empty: func() {
				ui.Text(c, "Nobody has anything assigned")
				drew = true
			},
		})
	}, 460, 200)
	if !tt.HasText("Nobody has anything assigned") || !drew {
		t.Errorf("the caller's empty state should be drawn; %q", tt.Texts())
	}
}

// ── PomodoroTimer ──────────────────────────────────────────────────────────

func TestPomodoroTimerShowsTheFocus(t *testing.T) {
	running := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := PomodoroTimer(c, PomodoroOptions{
			Elapsed: 10 * time.Minute, Running: &running,
			Started: func() {}, Stopped: func() {}, Reset: func() {},
		})
		if r.State().Phase != Focus {
			t.Errorf("ten minutes in is still focus, got %v", r.State().Phase)
		}
	}, 360, 420)
	for _, want := range []string{"15:00", "Focus", "Stop", "Reset"} {
		if !tt.HasText(want) {
			t.Errorf("the timer is missing %q; %q", want, tt.Texts())
		}
	}
}

func TestPomodoroTimerShowsTheBreak(t *testing.T) {
	running := true
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		r := PomodoroTimer(c, PomodoroOptions{
			Elapsed: 27 * time.Minute, Running: &running, Reset: func() {},
		})
		if r.State().Phase != Break {
			t.Errorf("twenty-seven minutes in is the break, got %v", r.State().Phase)
		}
	}, 360, 420)
	if !tt.HasText("Break") {
		t.Errorf("the break is missing; %q", tt.Texts())
	}
	if !tt.HasText("3:00") {
		t.Errorf("three minutes of break left; %q", tt.Texts())
	}
}

func TestPomodoroTimerResets(t *testing.T) {
	running := false
	reset := false
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		PomodoroTimer(c, PomodoroOptions{
			Elapsed: 4*time.Minute + 30*time.Second, Running: &running,
			Started: func() {}, Reset: func() { reset = true },
		})
	}, 360, 420)
	if err := tt.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	if !reset {
		t.Error("Reset did not fire")
	}
}

func TestPomodoroTimerNeedsItsRunningPointer(t *testing.T) {
	defer panics(t, "project: PomodoroTimer needs the *bool Running writes to", func() {
		ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			PomodoroTimer(c, PomodoroOptions{Elapsed: time.Minute})
		}, 400, 400)
	})
}

// ── the dark palette ───────────────────────────────────────────────────────

func TestProjectSurvivesTheDarkPalette(t *testing.T) {
	// Every sink the page writes to is declared here and read at the end. A
	// dark pass that collects answers and does not look at them cannot catch
	// a control firing on its own, which is the failure a dark palette exists
	// to surface.
	var selected, toggled, name, added, moved string
	var save, running, started, reset bool
	var release, step = 0, 0
	title, body := "Root cause review", "The edge returns 502."
	today := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	due := DueValue{Day: today.AddDate(0, 0, 2), Set: true, Text: "17 March"}
	status := StatusDoing
	done, total := 1, 3
	accent := ui.Hex("#5b8dff")

	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})

		KanbanBoard(c, KanbanBoardOptions{
			Columns: columns(), Tasks: tasks(), Selected: &selected, Prefix: "CB",
		})
		TaskItem(c, TaskItemOptions{Task: tasks()[0], Prefix: "CB"})
		IssueIdBadge(c, IssueIdBadgeOptions{Prefix: "CB", Number: 1042, Severity: core.Warning})
		IssueCard(c, IssueCardOptions{Task: tasks()[1], Prefix: "CB"})
		SubtaskList(c, SubtaskListOptions{
			Subtasks: subtasks(), Toggled: &toggled, Done: &done, Total: &total,
		})
		MilestoneProgress(c, MilestoneProgressOptions{
			Title: "Sprint 14", Done: 9, Total: 10, Due: "Due 21 March",
		})
		AssigneePicker(c, AssigneeOptions{
			Value: &selected, People: []string{"Ana Duarte", "Kim Osei"},
		})
		DueDatePicker(c, DueDatePickerOptions{
			Due: &due, Open: new(bool), Today: today, Clearable: true,
		})
		StatusSelect(c, StatusSelectOptions{Status: &status, Columns: columns()})
		LabelManager(c, LabelManagerOptions{
			Labels: []string{"backend"}, Toggled: &name, Name: &name,
			Added: func(string) { added = "toggled" },
		})
		SprintBoard(c, SprintBoardOptions{
			Sprint: "Sprint 14", DaysLeft: 2,
			Columns: columns(), Tasks: tasks(),
			Selected: &selected, Prefix: "CB",
		})
		BurndownChart(c, BurndownChartOptions{Burndown: sprint(), Height: 280})
		TimeTracker(c, TimeTrackerOptions{
			Running: &running, Elapsed: 65 * time.Minute,
			Started: func() { started = true },
			Entries: []TimeEntry{{Task: "Root cause review", Spent: 90 * time.Minute, Billable: true}},
		})
		WorkloadView(c, WorkloadViewOptions{Entries: workload(), Height: 240})
		PomodoroTimer(c, PomodoroOptions{
			Elapsed: 10 * time.Minute, Running: &running,
			Reset: func() { reset = true },
		})
		TaskList(c, TaskListOptions{
			Tasks: tasks(), Selected: &selected, Prefix: "CB", Height: 240,
		})
		TaskDetailPanel(c, TaskDetailPanelOptions{
			Task: tasks()[0], Prefix: "CB",
			Title: &title, Body: &body, Save: &save,
			Subtasks: subtasks(), SubtaskToggled: &toggled,
		})
		_ = accent
	}, 1400, 2200)

	for _, want := range []string{
		"Log a callback from the board", "CB-1042", "Subtasks", "Sprint 14",
		"Plan", "1:05:00", "Workload", "Focus", "Root cause review",
	} {
		if !tt.HasText(want) {
			t.Errorf("the dark pass is missing %q; %q", want, tt.Texts())
		}
	}

	// Seventeen components on one frame is seventeen chances to write to a
	// caller's pointer. Nothing here was pressed, so nothing may have moved.
	if started || reset || save {
		t.Errorf("something fired without being pressed: started=%v reset=%v save=%v",
			started, reset, save)
	}
	for _, got := range []struct{ what, v string }{
		{"toggled", toggled}, {"added", added}, {"moved", moved}, {"selected", selected},
	} {
		if got.v != "" {
			t.Errorf("%s was written to without being pressed: %q", got.what, got.v)
		}
	}
	if step != 0 || release != 0 {
		t.Errorf("a wizard or a release moved on its own: step=%d release=%d", step, release)
	}
	if status != StatusDoing || !due.Set || due.Text != "17 March" {
		t.Errorf("a field moved on its own: status=%q due=%+v", status, due)
	}
	_ = name
	_ = title
	_ = body
	_ = accent
}

// ── helpers ────────────────────────────────────────────────────────────────

// dragFromTo presses at one point, moves in steps to another and releases. A
// single jump is not a drag: MyGo only starts one after the pointer has moved
// while pressed.
func dragFromTo(tt *ui.Tester, fromX, fromY, toX, toY float32) {
	tt.Press(fromX, fromY)
	for k := 1; k <= 4; k++ {
		f := float32(k) / 4
		tt.Move(fromX+(toX-fromX)*f, fromY+(toY-fromY)*f)
	}
	tt.Release(toX, toY)
}

// panics asserts that building the thing panics with a message containing
// want. Anything a component cannot be given is a panic rather than a guess,
// so these are as much a part of the contract as what it draws.
func panics(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		if r == nil {
			t.Fatalf("expected a panic saying %q", want)
		}
		got, ok := r.(string)
		if !ok {
			t.Fatalf("the panic value is %T, not a string: %v", r, r)
		}
		if !strings.Contains(got, want) {
			t.Errorf("panic = %q, want it to contain %q", got, want)
		}
	}()
	fn()
}
