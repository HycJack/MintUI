package project

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/datetime"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// DueValue is when a task is wanted. It is a struct rather than a bare
// time.Time so that "no date" is a value rather than a zero time somebody has
// to remember is special: a task with no due date is not due on 1 January
// 0001, and the difference is the whole reason this is a type.
type DueValue struct {
	// Day is the date wanted. The zero time means no date.
	Day time.Time
	// Set reports whether there is a date at all. It is separate from Day so
	// that a caller can hold "no date" without having to import time and
	// compare against the zero value.
	Set bool
	// Text is how the date is written on the card and in the list. It is the
	// caller's string because "Friday" and "14 March" and "in 3 days" are
	// all reasonable and the caller knows which one its users expect.
	Text string
}

// DueTone is the severity a due date is drawn at, against a day the caller
// calls now.
//
// It takes `now` rather than reading the clock so that the same task draws
// the same colour on a replayed screenshot and in a test. A component that
// read the time itself would be right once and untestable forever after.
func DueTone(due DueValue, now time.Time) core.Severity {
	if !due.Set || due.Day.IsZero() {
		return core.Neutral
	}
	switch {
	case !due.Day.After(now):
		return core.Danger
	case due.Day.Sub(now) <= 24*time.Hour:
		return core.Warning
	default:
		return core.Neutral
	}
}

// DueLabel is what a date reads as beside its task, and the empty string when
// there is no date. It is a function rather than a branch at each call site so
// that "a task with no date has no words after its title" is one rule rather
// than four sites each having to remember it.
func DueLabel(due DueValue) string {
	if !due.Set {
		return ""
	}
	return due.Text
}

// DueDatePickerOptions configure a DueDatePicker.
type DueDatePickerOptions struct {
	// Due is the caller's value, written when a day is chosen.
	Due *DueValue
	// Open is the popover's open state. It is the caller's, as it is for
	// every picker in the library: whether picking a date ends the popover is
	// a decision about the form the date is in, and a form that needs the
	// calendar to stay open for a second date cannot be served by a picker
	// that closes itself.
	Open *bool
	// Today is the caller's today; the zero time means the picker refuses,
	// because a calendar with nothing to open on has no month to show.
	Today time.Time
	// Min and Max bound the days that can be chosen.
	Min, Max time.Time
	// Clearable draws the button that takes the date away, for a task whose
	// due date has been set by mistake.
	Clearable bool
	// Label names the control.
	Label string
}

// DueDatePickerResult carries a DueDatePicker and what was picked.
type DueDatePickerResult struct {
	// Element is the field, with the calendar under it while it is open.
	Element *ui.Element
	// picked and cleared are this frame's answers.
	picked  bool
	cleared bool
}

// Picked reports that a day was chosen this frame, so the caller can close
// the popover or re-validate the form.
func (r DueDatePickerResult) Picked() bool { return r.picked }

// Cleared reports that the date was taken away this frame.
func (r DueDatePickerResult) Cleared() bool { return r.cleared }

// DueDatePicker is a task's due date: a field and a calendar that chooses it.
//
// It is datetime.DatePicker with one thing on top: a date that can be absent.
// The library's picker reads and writes the caller's time.Time and cannot
// express "no date", which is a state a task is in more often than most
// people expect — everything logged has a due date and everything logged has
// one. The Set flag carries that, and the field's own button clears it.
//
// The clear button is guarded as well as drawn, for the reason everywhere in
// this library: Disabled only greys a control, so the press is refused in
// this package's own code. See docs/design-system.md §17.3.
func DueDatePicker(c *ui.Context, opts DueDatePickerOptions) DueDatePickerResult {
	if opts.Due == nil {
		panic("project: DueDatePicker needs the *DueValue it writes to")
	}
	if opts.Open == nil {
		panic("project: DueDatePicker needs the *bool its calendar opens in")
	}
	u := core.Density(c).Unit()
	name := opts.Label
	if name == "" {
		name = core.Msg(c, "project.dueDate", "Due date")
	}

	var r DueDatePickerResult
	row := ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center)
	row.Children(func() {
		// The library's picker owns the calendar and writes a time.Time; the
		// DueValue is a view of the same value, so nothing is duplicated and
		// there is no second copy to keep in step.
		//
		// It is built inside the row's own Children call, because an element
		// belongs to whatever container was current when it was made — see
		// docs/design-system.md §17.1.
		day := opts.Due.Day
		res := datetime.DatePicker(c, datetime.DatePickerOptions{
			Value:       &day,
			Open:        opts.Open,
			Today:       opts.Today,
			Min:         opts.Min,
			Max:         opts.Max,
			Placeholder: name,
			Name:        name,
			Format: func(t time.Time) string {
				// What the caller wrote, if they wrote anything: "Fri 14 Mar"
				// and "2026-03-14" and "in 3 days" are all things a team
				// expects, and only the caller knows which.
				if opts.Due.Text != "" {
					return opts.Due.Text
				}
				return datetime.FormatDate(t)
			},
		})
		r.Element = res.Element
		if t, ok := res.Picked(); ok {
			r.picked = true
			opts.Due.Day, opts.Due.Set = t, true
			if opts.Due.Text == "" {
				opts.Due.Text = datetime.FormatDate(t)
			}
		}

		if opts.Clearable {
			clear := input.IconButton(c, glyphDismiss, core.Msg(c, "project.clearDue", "Clear due date"),
				input.ButtonOptions{})
			if clear.Clicked() && opts.Due.Set {
				// The guard as well as the grey. Disabled only draws; it does
				// not refuse a press. See docs/design-system.md §17.3.
				r.cleared = true
				*opts.Due = DueValue{}
			}
		}
	})
	return r
}

// AssigneeOptions configure an AssigneePicker.
type AssigneeOptions struct {
	// Value is the chosen person's name, empty for nobody. Empty is a real
	// answer — an unassigned task is the common case, not the absence of
	// one — so it is a value rather than a nil pointer.
	Value *string
	// People are the choices, as names. They are names rather than records
	// because the picker draws a name and an avatar and both come from a
	// name; a caller with a richer person type supplies the names.
	People []string
	// Label names the control.
	Label string
	// Placeholder is what the trigger says while nobody is chosen.
	Placeholder string
}

// AssigneePickerResult carries an AssigneePicker.
type AssigneePickerResult struct {
	// Element is the trigger, with the panel under it while it is open.
	Element *ui.Element
}

// AssigneePicker is who a task belongs to.
//
// It is input.SelectSearch rather than input.Select, because the list of
// people a task can be assigned to is the one choice list on a board that
// nobody can hold in their head: a team of forty has names that look alike,
// and a picker that can be typed into is the difference between assigning a
// task and giving up.
//
// The unassigned option is in the list rather than being the empty value,
// because "no assignee" is something somebody chooses deliberately and it
// should read as a row they can see and press.
func AssigneePicker(c *ui.Context, opts AssigneeOptions) AssigneePickerResult {
	if opts.Value == nil {
		panic("project: AssigneePicker needs the *string Value writes to")
	}
	if len(opts.People) == 0 {
		panic("project: AssigneePicker needs at least one person")
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "project.assignee", "Assignee")
	}
	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = core.Msg(c, "project.unassigned", "Nobody")
	}

	choices := make([]input.Choice, 0, len(opts.People)+1)
	// The empty value is the first choice so that "unassign" is the thing at
	// the top of a long list rather than something at the bottom of it.
	choices = append(choices, input.Choice{Value: "", Label: placeholder})
	for _, p := range opts.People {
		choices = append(choices, input.Choice{Value: p, Label: p})
	}

	query := ""
	elem := input.SelectSearch(c, opts.Value, &query, choices, input.SelectSearchOptions{
		Label:       label,
		Placeholder: placeholder,
		Search:      core.Msg(c, "project.searchPeople", "Search people"),
	})
	return AssigneePickerResult{Element: elem}
}

// StatusSelectOptions configure a StatusSelect.
type StatusSelectOptions struct {
	// Status is the chosen column, written by the picker.
	Status *Status
	// Columns are the lanes on offer. Empty offers every status this package
	// knows, which is the set a new board starts with.
	Columns []Column
	// Label names the control.
	Label string
	// Placeholder is what the trigger says while nothing is chosen.
	Placeholder string
}

// StatusSelectResult carries a StatusSelect.
type StatusSelectResult struct {
	// Element is the trigger, with the panel under it while it is open.
	Element *ui.Element
}

// StatusSelect is which column a task is in.
//
// Every choice carries its status' own colour, as a tag beside its name. That
// is the point of the control: the board is coloured by status and a picker
// that showed the same words in the same grey would make somebody read all
// of them to work out where they are. The colour is a second channel beside
// the name, not instead of it, so it is still legible without it.
func StatusSelect(c *ui.Context, opts StatusSelectOptions) StatusSelectResult {
	if opts.Status == nil {
		panic("project: StatusSelect needs the *Status it writes to")
	}
	columns := opts.Columns
	if len(columns) == 0 {
		columns = defaultColumns()
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "project.status", "Status")
	}
	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = label
	}

	// The trigger's own text is the column's name, so the label has to say
	// what the control is for rather than repeat it. That is why the
	// placeholder takes the label here: a trigger reading "Status" beside a
	// group called "Status" is announced twice as the same word.
	var chosen string
	for _, col := range columns {
		if col.Status == *opts.Status {
			chosen = col.label()
		}
	}
	if chosen == "" && *opts.Status != "" {
		panic("project: StatusSelect is set to " + string(*opts.Status) +
			", which is not one of its columns")
	}

	elem := input.Select(c, &chosen, choicesOfStatus(columns), input.SelectOptions{
		Label:       label,
		Placeholder: placeholder,
	})
	// A Select writes its own string, so the status is written back from it
	// here rather than through the control's value type. The mapping runs
	// both ways each frame, which keeps the caller's Status authoritative
	// and the control's text a view of it.
	for _, col := range columns {
		if col.label() == chosen {
			*opts.Status = col.Status
		}
	}
	return StatusSelectResult{Element: elem}
}

// defaultColumns is the lane set a board starts with. It is a copy handed out
// each call so that a caller mutating the result of one board's lane list
// cannot change the next board's.
func defaultColumns() []Column {
	return []Column{
		{Status: StatusBacklog, Title: "Backlog"},
		{Status: StatusTodo, Title: "To do"},
		{Status: StatusDoing, Title: "In progress"},
		{Status: StatusReview, Title: "In review"},
		{Status: StatusBlocked, Title: "Blocked"},
		{Status: StatusDone, Title: "Done"},
	}
}

// choicesOfStatus turns lanes into the choices a select takes.
func choicesOfStatus(cols []Column) []input.Choice {
	out := make([]input.Choice, 0, len(cols))
	for _, col := range cols {
		out = append(out, input.Choice{Value: col.label(), Label: col.label()})
	}
	return out
}

// TaskListOptions configure a TaskList.
type TaskListOptions struct {
	// Tasks are the rows, in the caller's order. The list does not sort
	// them: what a person arranged by hand is an arrangement.
	Tasks []Task
	// Selected is the task the list marks as the open one, empty for none. A
	// press asks for a selection; the list makes none of its own.
	Selected *string
	// Prefix is the tracker, for the reference on each row.
	Prefix string
	// Height is the viewport's height, which a scroll area needs before it
	// will scroll. Zero lets the layout give it one.
	Height float32
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// TaskListResult carries a TaskList and what was pressed in it.
type TaskListResult struct {
	// Element is the whole list.
	Element *ui.Element
	// selected is the task that was pressed this frame.
	selected string
}

// Selected returns the task that was pressed this frame, empty when none was.
func (r TaskListResult) Selected() string { return r.selected }

// TaskList is the tasks as a list of rows rather than as cards: one task a
// line, in a panel narrow enough for a sidebar or a search result.
//
// It is a different component from TaskItem on purpose. A card is scanned by
// comparing cards, and a list is read one row at a time; giving a list the
// card's padding and its avatar would make a sidebar of forty tasks four
// times taller than it needs to be for the same information. What both share
// is that the title is one line and never wraps.
//
// The rows are built inside a scroll area of the caller's height. Without a
// height the area grows to fit its content and never scrolls, which is almost
// never what is meant — see ui/layout/frame.go.
func TaskList(c *ui.Context, opts TaskListOptions) TaskListResult {
	if opts.Selected == nil {
		panic("project: TaskList needs the *string Selected writes to")
	}
	u := core.Density(c).Unit()

	var r TaskListResult
	r.Element = ui.Column(c).FillWidth().Gap(u).Role(ui.RoleList).
		Label(core.Msg(c, "project.tasks", "Tasks"))
	// Everything below is built inside the column's own Children call,
	// because an element belongs to whatever container was current when it
	// was made. See docs/design-system.md §17.1.
	r.Element.Children(func() {
		layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true,
			Height:   opts.Height,
		}, func() {
			ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
				if len(opts.Tasks) == 0 {
					if opts.Empty != nil {
						opts.Empty()
					}
					return
				}
				for _, task := range opts.Tasks {
					taskRow(c, opts, task, &r.selected)
				}
			})
		})
	})
	return r
}

// taskRow is one line of the list: the reference, the title, and the status as
// a coloured tag so the column is visible without opening anything.
func taskRow(c *ui.Context, opts TaskListOptions, task Task, selected *string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	on := task.ID == *opts.Selected

	row := ui.Row(c).FillWidth().Gap(u*2).AlignItems(ui.Center).
		Padding(u*1.25, u*2).Radius(theme.ControlRadius).
		Cursor(ui.CursorPointer).Role(ui.RoleListItem).Label(task.Title)
	if on {
		row.Background(k.SurfaceHover)
	}
	if row.Clicked() {
		*selected = task.ID
	}
	row.Children(func() {
		if PrefixIssueID(opts.Prefix) != "" {
			ui.Text(c, FormatIssueID(opts.Prefix, task.IDNumber())).
				TextColor(k.TextFaint).Shrink(0).Font(monoFamily).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		ui.Text(c, task.Title).TextColor(k.Text).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		if task.Assignee != "" {
			display.Avatar(c, task.Assignee)
		}
		display.Tag(c, string(task.Status), display.TagOptions{
			Tone: StatusSeverity(task.Status),
		})
	})
}

// glyphDismiss is the cross that takes a value away. It is the one mark here
// that has to read at ten points: too thin and it disappears, too heavy and
// it competes with the label beside it.
var glyphDismiss = ui.MustParseSVG([]byte(
	`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` +
		`stroke="currentColor" stroke-width="1.8" stroke-linecap="round">` +
		`<path d="m6.5 6.5 11 11M17.5 6.5l-11 11"/></svg>`))
