package project_test

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/project"
)

// Example shows a sprint board and its burndown, which is the pair the two
// views exist to be read against: the board says where the cards are now and
// the chart says whether they will be gone by Friday.
//
// Three rules are visible on this page. A lane never narrows — the columns
// keep theme.ColumnWidth and the whole thing scrolls sideways when the window
// cannot hold them, because a wrapped card title is a wall of text. A drag
// carries the task's id as a string and the move is reported, not performed.
// And the chart is ui/chart's own: the frame there is what measures the
// gutters and thins the axis labels so they stop colliding.
func Example() {
	tasks := []project.Task{
		{ID: "CB-1042", Title: "Root cause review: intermittent 502s from the edge",
			Status: project.StatusDoing, Assignee: "Rosa Vidal",
			Labels: []string{"backend", "flaky"}, Priority: 2,
			Subtasks: 4, SubtasksDone: 3, Comments: 12, Due: "14 March"},
		{ID: "CB-1043", Title: "Log a callback from the board", Status: project.StatusTodo,
			Assignee: "Ana Duarte"},
		{ID: "CB-1044", Title: "Blocked: waiting on procurement", Status: project.StatusBlocked,
			Assignee: "Kim Osei", Due: "yesterday", Overdue: true},
	}
	lanes := []project.Column{
		{Status: project.StatusTodo, Title: "To do"},
		{Status: project.StatusDoing, Title: "In progress"},
		{Status: project.StatusBlocked, Title: "Blocked"},
		{Status: project.StatusDone, Title: "Done"},
	}

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		var open string

		board := project.SprintBoard(c, project.SprintBoardOptions{
			Sprint: "Sprint 14", DaysLeft: 2,
			Columns: lanes, Tasks: tasks,
			Selected: &open, Prefix: "CB",
		})

		// A drop is an event on one frame, not a state, so it is read from
		// the Result rather than written to by the board.
		if m := board.Moved(); m != "" {
			// "from,to,task" — re-order Tasks and ask again.
		}

		project.BurndownChart(c, project.BurndownChartOptions{
			Burndown: project.Burndown{
				Name: "Sprint 14", Scope: 50, Today: 5, Unit: "points",
				Days: []project.BurndownDay{
					{Day: 0, Remaining: 50},
					{Day: 1, Remaining: 44},
					{Day: 3, Remaining: 30}, // day 2 was a weekend
					{Day: 5, Remaining: 19},
				},
			},
			Height: 260,
		})
	}, 1500, 900)

	// Output:
}

// ExampleTaskDetailPanel shows one task, edited.
//
// Every field writes straight back into a string the caller owns, so the
// board behind the panel and the form in front of it can never disagree about
// a title.
func ExampleTaskDetailPanel() {
	var title, body string
	var saved bool
	today := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		title, body = "Root cause review", "The edge returns 502 for one request in two hundred."

		var toggled string
		project.TaskDetailPanel(c, project.TaskDetailPanelOptions{
			Task: project.Task{ID: "CB-1042", Status: project.StatusDoing,
				Assignee: "Rosa Vidal", Due: "14 March"},
			Prefix: "CB",
			Title:  &title, Body: &body, Save: &saved,
			Columns: []project.Column{
				{Status: project.StatusDoing, Title: "In progress"},
				{Status: project.StatusDone, Title: "Done"},
			},
			Subtasks: []project.Subtask{
				{ID: "s1", Title: "Reproduce on staging", Assignee: "Ana Duarte"},
				{ID: "s2", Title: "Check the edge configuration", Done: true},
			},
			SubtaskToggled: &toggled,
			Today:          &today,
		})
	}, 560, 700)

	// Output:
}

// ExampleIdealBurndown shows the straight line a sprint is measured against,
// at a scope that does not divide by the number of days — so the assertion
// about it being straight has to be one about a tolerance rather than about
// the last bit of a float.
func ExampleBurndown() {
	for i, v := range project.IdealBurndown(4, 37) {
		fmt.Println(i, v)
	}
	fmt.Println(project.FormatElapsed(65 * time.Minute))
	fmt.Println(project.FormatElapsed(25 * time.Minute))

	// Output:
	// 0 37
	// 1 27.75
	// 2 18.5
	// 3 9.25
	// 4 0
	// 1:05:00
	// 25:00
}
