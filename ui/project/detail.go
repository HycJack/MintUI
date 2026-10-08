package project

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
)

// TaskDetailPanelOptions configure a TaskDetailPanel.
type TaskDetailPanelOptions struct {
	// Task is the task as it is: every field below starts from this record
	// and writes back into the caller's own pointers. There is no copy of it
	// inside the panel.
	Task Task
	// Prefix is the tracker, for the reference at the head.
	Prefix string
	// Title and Body are the caller's strings, written as they are typed.
	// Title is the task's own title rather than Task.Title, because a panel
	// that edited a copy of the title would have two titles: the one on the
	// board and the one in the form.
	Title, Body *string
	// Status is the task's column. Nil draws no status control, which is what
	// a read-only panel wants.
	Status *Status
	// Columns are the lanes the status control offers.
	Columns []Column
	// Assignee is who is on it, empty for nobody.
	Assignee *string
	// People are the choices the assignee control offers.
	People []string
	// Due is when it is wanted. Nil draws no date control.
	Due *DueValue
	// Today is the caller's today, for the date picker and for the colour of
	// the date. Nil with a Due is refused, because a calendar with nothing
	// to open on has no month to show.
	Today *time.Time
	// Subtasks are the checklist, drawn under the fields when there are any.
	Subtasks []Subtask
	// SubtaskToggled writes the id of the line that was ticked, empty when
	// none was.
	SubtaskToggled *string
	// Save writes true when the save button is pressed.
	Save *bool
	// Busy greys the form out while a save is in flight, and is the caller's
	// because the request is the caller's.
	Busy bool
	// Width bounds the panel; zero lets it fill what it is in.
	Width float32
}

// TaskDetailPanelResult carries a TaskDetailPanel and what was done in it.
type TaskDetailPanelResult struct {
	// Element is the whole panel.
	Element *ui.Element
	// saved reports a press of the save button this frame.
	saved bool
}

// Saved reports a press of the save button. The caller reads it on the frame
// after, because the last pass of a frame reports nothing — see
// docs/design-system.md §17.2 — and this package has nowhere to save into.
func (r TaskDetailPanelResult) Saved() bool { return r.saved }

// TaskDetailPanel is one task, edited: its title, its description, and the
// four fields that decide where it sits and who is on it.
//
// It is a Form rather than a column of boxes, so the label, the control and
// the message under a field are the library's and the tab order runs top to
// bottom. The order of the fields is the order somebody reads them in: what
// it is, what is happening, where it is, who has it, when it is wanted —
// which is the order of the questions, not the order of the database columns.
//
// The pickers are the library's — StatusSelect, AssigneePicker and
// DueDatePicker above — so the choices behave the same as they do anywhere
// else in the app. A detail panel that re-drew them would be the one place
// where assigning a task works differently from assigning one in a form.
func TaskDetailPanel(c *ui.Context, opts TaskDetailPanelOptions) TaskDetailPanelResult {
	if opts.Title == nil || opts.Body == nil {
		panic("project: TaskDetailPanel needs a Title and a Body to point at")
	}
	if opts.Save == nil {
		panic("project: TaskDetailPanel needs the *bool Save writes to")
	}
	if opts.Due != nil && opts.Today == nil {
		panic("project: a TaskDetailPanel with a due date needs a Today; a " +
			"calendar with nothing to open on has no month to show")
	}
	if len(opts.Subtasks) > 0 && opts.SubtaskToggled == nil {
		panic("project: a TaskDetailPanel with subtasks needs the *string " +
			"SubtaskToggled writes to")
	}
	u := core.Density(c).Unit()
	task := opts.Task

	var r TaskDetailPanelResult
	r.Element = input.Form(c, input.FormOptions{
		Label: core.Msg(c, "project.detail", "Task"),
		Width: opts.Width,
		Fields: func() {
			// The reference and the status at the head, because they are what
			// the panel is about rather than fields in it.
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				IssueIdBadge(c, IssueIdBadgeOptions{
					Prefix:   opts.Prefix,
					Number:   task.IDNumber(),
					Severity: StatusSeverity(task.Status),
				})
				display.Tag(c, string(task.Status), display.TagOptions{
					Tone: StatusSeverity(task.Status),
				})
			})

			titleLabel := core.Msg(c, "project.taskTitle", "Title")
			input.FormField(c, input.FormFieldOptions{Label: titleLabel, Required: true},
				func(err string) *ui.Element {
					return input.TextInput(c, opts.Title, input.TextInputOptions{
						Label: titleLabel, Error: err, Disabled: opts.Busy,
					})
				})

			input.FormField(c, input.FormFieldOptions{
				Label: core.Msg(c, "project.description", "Description"),
			}, func(err string) *ui.Element {
				return input.TextArea(c, opts.Body, input.TextAreaOptions{
					Label:    core.Msg(c, "project.description", "Description"),
					Error:    err,
					Lines:    4,
					Disabled: opts.Busy,
				})
			})

			if opts.Status != nil {
				StatusSelect(c, StatusSelectOptions{
					Status:  opts.Status,
					Columns: opts.Columns,
				})
			}
			if opts.Assignee != nil && len(opts.People) > 0 {
				AssigneePicker(c, AssigneeOptions{
					Value: opts.Assignee, People: opts.People,
				})
			}
			if opts.Due != nil && opts.Today != nil {
				DueDatePicker(c, DueDatePickerOptions{
					Due: opts.Due, Open: new(bool), Today: *opts.Today,
					Clearable: true,
				})
			}
			if len(opts.Subtasks) > 0 {
				SubtaskList(c, SubtaskListOptions{
					Subtasks: opts.Subtasks, Toggled: opts.SubtaskToggled,
				})
			}
		},
		Actions: func() {
			ui.Box(c).Grow(1)
			if input.Button(c, core.Msg(c, "project.save", "Save"),
				input.ButtonOptions{Primary: true, Disabled: opts.Busy}).Clicked() {
				r.saved = true
				*opts.Save = true
			}
		},
	}).Element
	return r
}
