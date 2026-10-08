package project

import (
	"strings"

	"github.com/HycJack/MintUI/ui/core"
)

// Status is where a task sits. It is a string rather than an enum so that it
// goes to a URL, to a settings file and to a test without a conversion, and
// because the set of columns on a board is the caller's: a team that runs a
// "In review" lane between "Doing" and "Done" has to be able to say so.
type Status string

// The statuses this package knows how to colour. A caller may use any string
// as a status; these are the ones with a severity attached.
const (
	StatusBacklog Status = "backlog"
	StatusTodo    Status = "todo"
	StatusDoing   Status = "doing"
	StatusReview  Status = "review"
	StatusBlocked Status = "blocked"
	StatusDone    Status = "done"
)

// Severity is the tone a status is drawn at, which is a function rather than a
// field on the status: two boards can both have a "doing" and one of them can
// be a bottleneck the other treats as ordinary, and the caller says which.
//
// The two that matter are Blocked, which is a warning rather than a danger —
// blocked work is waiting on somebody, not broken — and Done, which is the
// only success in this package. A board where half the statuses are coloured
// is a board where none of them are.
func StatusSeverity(s Status) core.Severity {
	switch s {
	case StatusBlocked:
		return core.Warning
	case StatusDone:
		return core.Success
	case StatusDoing, StatusReview:
		return core.Accent
	default:
		return core.Neutral
	}
}

// Task is one piece of work as the board knows it.
type Task struct {
	// ID is the task's identity, and what a drag carries between columns.
	// It is a string so that it can be whatever the caller's system uses —
	// "CB-1042", a UUID, a database row id — and so that ui.Drop[string]
	// can carry it without a conversion on either end.
	ID string
	// Title is what the card shows, on one line.
	Title string
	// Status is the column it is in.
	Status Status
	// Assignee is who is on it, empty for nobody. It is a name rather than a
	// person record because the board's initials circle and the caller's
	// richer record are two views of the same thing, and the circle is
	// already the library's.
	Assignee string
	// Labels are the tags on it.
	Labels []string
	// Due is when it is wanted, already formatted and empty for no date. The
	// caller formats it because a due date read in a timezone that is not the
	// caller's is a due date that is wrong by a day.
	Due string
	// Overdue marks a task whose date has passed and is still open. It is
	// the caller's answer rather than this package's, because deciding that
	// a date has passed needs the clock and the timezone, and a board that
	// disagreed with the caller about it would be the more confusing of the
	// two.
	Overdue bool
	// Priority is a count of urgency points, zero for none. It is a count
	// rather than a word because every tracker counts them differently and
	// this package has no business choosing one.
	Priority int
	// Subtasks is how many there are, and Done how many of them are closed.
	Subtasks, SubtasksDone int
	// Comments is how many there are, shown on the card when it is not zero.
	Comments int
}

// Done reports whether the task is finished, which is what stops a card in a
// closed column from also being counted as open somewhere else in this
// package.
func (t Task) Done() bool { return t.Status == StatusDone }

// Column is one lane of the board.
type Column struct {
	// Status is the lane's key: what a task's Status has to be to sit here.
	Status Status
	// Title is what the lane is called. It defaults to Status, which is
	// enough for a board whose lanes are all the caller's own vocabulary.
	Title string
	// Limit is how many tasks may sit in the lane at once, zero for no
	// limit. It is drawn as the count's companion rather than enforced here:
	// a component that refused a card would be deciding a team's process.
	Limit int
	// Over marks a lane that is over its limit, so the caller can colour it.
	Over bool
	// Menu names a trailing button; empty draws none.
	Menu string
	// Empty draws instead of the body when the lane holds nothing. The place
	// a column keeps its own explanation — "nothing waiting on review" — so
	// the board around it does not have to know what absence looks like.
	Empty func()
}

// label is the column's own name, falling back to its key.
func (c Column) label() string {
	if c.Title != "" {
		return c.Title
	}
	return string(c.Status)
}

// FormatIssueID is a task's reference as it is printed: the prefix, a
// separator, and the number.
//
// The number is not padded. "CB-1042" and "CB-1043" are the same length and
// the next hundred will not need zeros in front of them, so padding buys
// nothing until the day it does and costs a wider column until then. A
// caller whose keys really are fixed width can pass a padded string and
// PrefixIssueID will not touch it.
func FormatIssueID(prefix string, n int) string {
	p := strings.ToUpper(strings.TrimSpace(prefix))
	switch {
	case p == "" && n <= 0:
		return ""
	case p == "":
		return itoa(n)
	case n <= 0:
		return p
	default:
		return p + "-" + itoa(n)
	}
}

// PrefixIssueID is the fixed part of a reference — "CB-" — which is what a
// search matches against and what the badge is built from. It is separate
// from FormatIssueID because the badge shows the prefix on its own and adding
// a number to it would make the widest badge depend on the smallest number.
func PrefixIssueID(prefix string) string {
	p := strings.ToUpper(strings.TrimSpace(prefix))
	if p == "" {
		return ""
	}
	return p + "-"
}

// MilestoneTone is the severity a milestone's progress is drawn at.
//
// The steps are the ones where somebody would move the date. A milestone
// halfway done is nothing remarkable; one with a day left and a third of its
// work open is a conversation.
func MilestoneTone(done, total int) core.Severity {
	if total <= 0 {
		return core.Neutral
	}
	switch {
	case done >= total:
		return core.Success
	case float64(done)/float64(total) >= 0.9:
		return core.Accent
	default:
		return core.Neutral
	}
}

// MilestoneFraction is how much of a milestone is done, 0 to 1. An empty
// milestone is 0 rather than a division by zero: a bar showing "nothing done"
// for a milestone nobody has started is true, and a bar showing nothing at
// all would not be.
func MilestoneFraction(done, total int) float32 {
	if total <= 0 || done <= 0 {
		return 0
	}
	if done >= total {
		return 1
	}
	return float32(done) / float32(total)
}

// itoa writes an int, sign and all.
//
// The sign matters because a sprint's countdown goes negative when it has run
// out, and "-2 days left" is the honest reading. An unsigned printer would
// render that as " days left", which is a countdown nobody can read.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
