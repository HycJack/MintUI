package project

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// KanbanBoardOptions configure a KanbanBoard.
type KanbanBoardOptions struct {
	// Columns are the lanes, in order, and Tasks are the tasks among them.
	// The board does not sort either: a lane's order is the order somebody
	// dragged it into, and re-sorting it would undo the one gesture the
	// board exists to support.
	Columns []Column
	Tasks   []Task
	// Selected is the task the board marks as the open one, empty for none.
	// A press on a card asks to open it; the board does not open anything,
	// because what opening a task means — a drawer, a window, a route — is
	// the caller's decision and not this package's.
	//
	// The move is *not* a pointer here, unlike the selection. A drop is an
	// event that happened on one frame, not a state a window holds, so it is
	// reported through KanbanBoardResult.Moved the way layout.ColumnResult
	// reports its own drop — and giving it a sink as well would be two
	// answers to one question.
	Selected *string
	// Prefix is the task tracker, for the reference on each card. Empty
	// draws no reference at all, which is right for a board of tasks that
	// have no numbers.
	Prefix string
	// Height is the board's own height, which a scroll area needs before it
	// will scroll. Zero lets the layout give it one.
	Height float32
}

// KanbanBoardResult carries a KanbanBoard and what happened on it.
type KanbanBoardResult struct {
	// Element is the whole board.
	Element *ui.Element
	// moved is the drop that was made, "from,to,task".
	moved string
	// opened is the task that was pressed this frame.
	opened string
}

// Moved returns the drop made this frame as "from,to,task", empty when none
// was. The three parts are commas because a status and a task id are both
// strings with no commas in them, and because a caller that wants them
// separately should be holding the board's own state rather than be handed a
// slice of somebody else's.
func (r KanbanBoardResult) Moved() string { return r.moved }

// Opened returns the task that was pressed this frame, empty when none was.
func (r KanbanBoardResult) Opened() string { return r.opened }

// KanbanBoard is the board: lanes side by side, each one scrolling on its own,
// and the whole thing scrolling sideways when the window is too narrow.
//
// It is layout.Board and layout.Column[T] rather than a row of boxes written
// here, which is the decision that matters most in this package. Column keeps
// its width no matter what the window does, wraps the body in a scroll view,
// and — when a lane is empty — draws the caller's Empty *instead of* that
// scroll view, so that the button in "nothing waiting on review" is still
// hittable. All three of those are the rules of a board, and getting any of
// them wrong shows as a card title on two lines.
//
// Dropping is ui.Drop[string] through Column, so a lane says what it takes and
// the value that arrives is already a task id. Nothing here reorders a slice:
// the drop is reported through Moved and the caller reorders Tasks and asks
// again, which is what keeps the board a view of the caller's data rather than
// a second copy of it.
func KanbanBoard(c *ui.Context, opts KanbanBoardOptions) KanbanBoardResult {
	if opts.Selected == nil {
		panic("project: KanbanBoard needs the *string Selected writes to")
	}
	var r KanbanBoardResult
	board := layout.Board(c, func() {
		for _, col := range opts.Columns {
			in := tasksIn(opts.Tasks, col.Status)
			count := len(in)
			here := columnOf(opts, col)
			res := layout.Column(c, layout.ColumnOptions[string]{
				Title: col.label(),
				Count: &count,
				Menu:  col.Menu,
				Empty: here.Empty,
			}, func() {
				for i := range in {
					task := in[i]
					card := TaskItem(c, TaskItemOptions{
						Task: task, Prefix: opts.Prefix,
						Selected: task.ID == *opts.Selected,
					})
					// The card is the drag source and carries the task's id,
					// so what arrives at the other lane is a string that says
					// which task moved rather than the task itself.
					card.Drag(task.ID)
					if card.Clicked() {
						r.opened = task.ID
					}
				}
			})

			if dropped, ok := res.Dropped(); ok {
				// A drop onto the lane it is already in is not a move. It is
				// still reported — somebody did drag it and may want to
				// know it landed — but the board says so as "from,to,task"
				// with from and to equal, and leaves the decision to the
				// caller, because a board that ignored the drop would give
				// no answer at all to somebody whose drag did not register.
				r.moved = string(col.Status) + "," + string(col.Status) + "," + dropped
			}
		}
	})
	if opts.Height > 0 {
		board.Height(opts.Height)
	}
	r.Element = board
	return r
}

// columnOf is the lane with its derived fields filled in, so that a limit
// over-run is decided once per frame in one place rather than in every row.
func columnOf(opts KanbanBoardOptions, c Column) Column {
	if c.Limit <= 0 {
		return c
	}
	n := len(tasksIn(opts.Tasks, c.Status))
	c.Over = n > c.Limit
	return c
}

// tasksIn is the tasks in one lane, in the order the caller has them.
func tasksIn(all []Task, status Status) []Task {
	var out []Task
	for _, t := range all {
		if t.Status == status {
			out = append(out, t)
		}
	}
	return out
}

// TaskItemOptions configure a TaskItem.
type TaskItemOptions struct {
	// Task is what the card shows.
	Task Task
	// Prefix is the tracker, for the reference. Empty draws no reference.
	Prefix string
	// Selected draws the card as the open one: a hairline in the accent
	// colour rather than a filled background, because a filled card in a
	// column of white cards is the one card that is not the same kind of
	// thing as the rest.
	Selected bool
	// Compact drops the labels line, for a card in a narrow lane or a
	// sidebar list.
	Compact bool
}

// TaskItem is one card on the board: a reference, a title that stays on one
// line, and the few facts about it that fit.
//
// The title is SingleLine and the card is one row tall for it. That is the
// rule the whole board rests on — see docs/design-system.md §6.7 — because a
// column of cards of different heights cannot be scanned across, and the eye
// reads a board by comparing cards at the same height rather than by reading
// any of them.
//
// The card is not clickable on its own. The board attaches the drag and reads
// the press; a card used on its own — in a detail panel, in a search result —
// attaches its own. That is what ui.Drag returns for.
func TaskItem(c *ui.Context, opts TaskItemOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	task := opts.Task

	card := ui.Column(c).FillWidth().Gap(u).Padding(u*2, u*2.5).
		Radius(theme.CardRadius).Background(k.Background).
		BorderWidth(theme.BorderWidth).BorderColor(k.Border).
		Label(task.Title).Role(ui.RoleButton)
	if opts.Selected {
		// A ring rather than a fill: the card stays a card, and the thing
		// that marks it is on the edge where a card's edge is.
		card.BorderWidth(theme.BorderWidth * 2).BorderColor(k.Accent)
	}
	card.Children(func() {
		cardHead(c, opts)
		// The title claims no height: in a column Grow is vertical, and a
		// title grown into free height it does not have lays out at zero
		// tall while its glyphs still paint — the line under it draws
		// straight through it.
		ui.Text(c, task.Title).TextColor(k.Text).FillWidth().
			FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
		if !opts.Compact {
			cardFoot(c, task)
		}
	})
	return card
}

// cardHead is the reference and the assignee, on one line above the title.
func cardHead(c *ui.Context, opts TaskItemOptions) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	task := opts.Task

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
		if ref := FormatIssueID(opts.Prefix, task.IDNumber()); ref != "" {
			// Faint, not muted: the reference is a mark for the eye to
			// recognise a card by, and it must never be the first thing read
			// out of it.
			ui.Text(c, ref).TextColor(k.TextFaint).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		if task.Priority > 0 {
			PriorityPill(c, task.Priority)
		}
		ui.Box(c).Grow(1)
		if task.Assignee != "" {
			display.Avatar(c, task.Assignee)
		}
	})
}

// cardFoot is the line of small facts: the labels, the subtasks, the due date.
func cardFoot(c *ui.Context, task Task) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
		if len(task.Labels) > 0 {
			ui.Text(c, task.Labels[0]).TextColor(k.TextMuted).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			if len(task.Labels) > 1 {
				ui.Text(c, "+"+itoa(len(task.Labels)-1)).TextColor(k.TextFaint).Shrink(0).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
		}
		if task.Subtasks > 0 {
			ui.Text(c, itoa(task.SubtasksDone)+"/"+itoa(task.Subtasks)).
				TextColor(k.TextMuted).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		if task.Comments > 0 {
			ui.Text(c, itoa(task.Comments)).TextColor(k.TextFaint).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		if task.Due != "" {
			// The date is drawn in the danger colour only when the caller says
			// it has passed. This package has no clock, and a board that
			// decided for itself would disagree with the caller about which
			// side of midnight it was.
			ink := k.TextMuted
			if task.Overdue {
				ink = k.Danger
			}
			ui.Text(c, task.Due).TextColor(ink).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
}

// IDNumber is the number inside a task's id, which is what the reference is
// built from. An id that is not a number at all — a UUID, a slug — yields
// zero, and FormatIssueID then prints the bare prefix: a board whose ids are
// slugs shows "CB-" and no number rather than a number invented from the
// slug's letters.
func (t Task) IDNumber() int {
	start := len(t.ID)
	for start > 0 && t.ID[start-1] >= '0' && t.ID[start-1] <= '9' {
		start--
	}
	if start == len(t.ID) {
		return 0
	}
	// Parsed forwards from the run's own start. Walking backwards and
	// multiplying by ten assembles the digits in reverse, which turns
	// "1042" into 2401 — a reference that looks plausible and is wrong, and
	// is exactly the kind of bug that ships because every sample on the page
	// has a single digit.
	n := 0
	for i := start; i < len(t.ID); i++ {
		n = n*10 + int(t.ID[i]-'0')
	}
	return n
}

// PriorityPill is a task's urgency points as a small mark, in the severity the
// count implies.
//
// Two points is ordinary and three is the top of the scale, so the ladder
// stops at warning: a board where the highest priority is a failure colour
// has no way to say "this one actually matters".
func PriorityPill(c *ui.Context, points int) *ui.Element {
	if points < 0 {
		panic("project: PriorityPill needs a count of points; a negative " +
			"priority is not less urgent, it is a bug in the caller's data")
	}
	sev := core.Neutral
	switch {
	case points >= 3:
		sev = core.Danger
	case points == 2:
		sev = core.Warning
	case points == 1:
		sev = core.Accent
	}
	if points == 0 {
		sev = core.Neutral
	}
	bg, fg := sev.Pair(core.Tokens(c))
	name := core.Msg(c, "project.priority", "Priority") + ": " + itoa(points)
	return ui.Box(c).Padding(unit(c)*0.5, unit(c)*1.25).Radius(theme.PillRadius).
		Background(bg).Label(name).Children(func() {
		ui.Text(c, itoa(points)).TextColor(fg).
			FontSize(core.FontSize(c, theme.CaptionSize)).Bold().Shrink(0)
	})
}

// SprintBoardOptions configure a SprintBoard.
type SprintBoardOptions struct {
	// Sprint is what the board is of, drawn at the head of it. It is the
	// caller's own string because a sprint is named by whichever tracker
	// the team uses and "Sprint 14" is not universal.
	Sprint string
	// DaysLeft is how many days are left in it, negative for one that has
	// run out. It is drawn as a countdown beside the name because a sprint
	// board whose only date is in its title is a board nobody can read at a
	// glance on a Friday.
	DaysLeft int
	// Columns and Tasks are the lanes and their cards, exactly as for a
	// KanbanBoard. A sprint board is a board with a clock over it, not a
	// different board, and having two components for one shape would mean
	// two answers to how a card is drawn.
	Columns []Column
	Tasks   []Task
	// Selected is as KanbanBoard's; the move is reported through the Result.
	Selected *string
	// Prefix is the tracker, for the reference on each card.
	Prefix string
}

// SprintBoardResult carries a SprintBoard and what happened on it.
type SprintBoardResult struct {
	// Element is the whole board.
	Element *ui.Element
	// moved and opened are as KanbanBoardResult's.
	moved, opened string
}

// Moved returns the drop made this frame as "from,to,task", empty when none
// was.
func (r SprintBoardResult) Moved() string { return r.moved }

// Opened returns the task that was pressed this frame, empty when none was.
func (r SprintBoardResult) Opened() string { return r.opened }

// SprintBoard is a board with a header saying which sprint it is and how much
// of it is left.
//
// The board below it is a KanbanBoard, called rather than copied. The two
// share every rule that matters — column width, horizontal scroll, the drop
// being reported and not performed — and a sprint board that drifted from
// those would be the one place on a board where a card is drawn differently,
// which is the thing a board cannot afford.
//
// The countdown is coloured by SprintTone, so a sprint with a day left and a
// full board of unstarted cards says so in one mark rather than in a
// sentence somebody has to compose.
func SprintBoard(c *ui.Context, opts SprintBoardOptions) SprintBoardResult {
	if opts.Selected == nil {
		panic("project: SprintBoard needs the *string Selected writes to")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var r SprintBoardResult
	r.Element = ui.Column(c).Fill().Gap(u * 2).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, opts.Sprint).TextColor(k.Text).Bold().Grow(1).
				FontSize(core.FontSize(c, theme.TitleSize)).SingleLine()
			days := itoa(opts.DaysLeft) + " " + core.Msg(c, "project.daysLeft", "days left")
			_, fg := SprintTone(opts.DaysLeft).Pair(k)
			ui.Box(c).Padding(u*0.75, u*2).Radius(theme.PillRadius).
				Background(k.Surface).Label(days).Children(func() {
				ui.Text(c, days).TextColor(fg).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold().Shrink(0)
			})
		})

		board := KanbanBoard(c, KanbanBoardOptions{
			Columns:  opts.Columns,
			Tasks:    opts.Tasks,
			Selected: opts.Selected,
			Prefix:   opts.Prefix,
		})
		board.Element.Grow(1)
		if m := board.Moved(); m != "" {
			r.moved = m
		}
		if o := board.Opened(); o != "" {
			r.opened = o
		}
	})
	return r
}

// SprintTone is the severity a sprint's countdown is drawn at.
//
// Three days is the point at which somebody starts moving work out rather
// than in, and it is a point rather than a mood: the same rule on every
// board means two people reading two boards compare them the same way.
func SprintTone(daysLeft int) core.Severity {
	switch {
	case daysLeft < 0:
		return core.Danger
	case daysLeft <= 3:
		return core.Warning
	default:
		return core.Neutral
	}
}
