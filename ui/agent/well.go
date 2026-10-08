package agent

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// RowKind is what one row of a run's transcript is. A transcript is a column
// of rows of several different shapes, and the shapes are what this enum is
// for: a caller assembling a transcript holds one slice of one type rather
// than one slice per kind, and a component reading it switches on the kind
// instead of type-asserting eight different things.
type RowKind int

const (
	// RowStep is a step of the run.
	RowStep RowKind = iota
	// RowThinking is a run's own account of what it is doing, before it has
	// done any of it.
	RowThinking
	// RowToolCall is a call to a tool.
	RowToolCall
	// RowCommand is a command run at a shell.
	RowCommand
	// RowFileChange is a file the run wrote to.
	RowFileChange
	// RowNote is something said rather than done: a decision, a warning, a
	// question nobody has answered yet.
	RowNote
	// RowRequest is the run stopping to ask a person something.
	RowRequest
)

func (k RowKind) String() string {
	switch k {
	case RowThinking:
		return "Thinking"
	case RowToolCall:
		return "Tool call"
	case RowCommand:
		return "Command"
	case RowFileChange:
		return "File change"
	case RowNote:
		return "Note"
	case RowRequest:
		return "Question"
	case RowStep:
		return "Step"
	}
	panic("agent: unknown RowKind " + itoa(int(k)))
}

// WellRow is one row of a run's transcript.
//
// It is one type with a field per kind rather than seven types behind an
// interface. An interface here would be an interface with seven
// implementations and no method on any of them, so it would buy nothing: the
// only thing a row can be asked is to draw itself, and every drawer here is
// in this package. What one type buys is that a caller can append a row of
// any kind to the same slice, which is what a transcript arriving over time
// actually is.
type WellRow struct {
	// Kind is which of the fields below is the row.
	Kind RowKind
	// At is when it happened, already formatted, drawn in the gutter.
	At string
	// Agent is who produced it, drawn as a caption above the content when
	// the row's kind has one.
	Agent string
	// Status is where the thing is, and is drawn as the row's mark.
	Status Status
	// Step is the row when Kind is RowStep.
	Step Step
	// Call is the row when Kind is RowToolCall.
	Call ToolCall
	// Exec is the row when Kind is RowCommand.
	Exec CommandExecution
	// File is the row when Kind is RowFileChange, with Added and Removed its
	// line counts.
	File    ReviewFile
	Added   int
	Removed int
	// Text is the row when Kind is RowThinking, RowNote or RowRequest.
	Text string
	// Draft is the row's free-text field, when a RowRequest has one to
	// answer into. It is the caller's, and nil leaves the question read-only
	// — which is what a transcript row of somebody else's question is, and
	// what a window that answers elsewhere wants.
	Draft *string
}

// AgentWellRowsOptions configure an AgentWellRows.
type AgentWellRowsOptions struct {
	// Rows are the rows, in the order they happened. The order is the whole
	// of a transcript: there is no other way to read one.
	Rows []WellRow
	// Selected is the row marked as the chosen one, -1 for none.
	Selected *int
	// Gap is the space between rows; zero takes the density's.
	Gap float32
	// Empty draws instead of the rows when there are none, and is required in
	// that case: a transcript that has not started yet is a state somebody
	// has to describe, and this package cannot guess whether they want "no
	// activity yet" or "the run has not reported".
	Empty string
}

// AgentWellRows is the column a run's transcript flows down.
//
// Every row is one shape — a step, a call, a command, a file — and this is
// the component that decides what they all look like together. The decision
// that matters is the gutter: the time and the status mark are drawn at the
// left of every row whatever the row is, so a reader scanning the column is
// scanning one column of marks rather than re-learning a new one at every
// kind of thing.
func AgentWellRows(c *ui.Context, opts AgentWellRowsOptions) *ui.Element {
	u := core.Density(c).Unit()
	if len(opts.Rows) == 0 && opts.Empty == "" {
		panic("agent: AgentWellRows with no rows needs opts.Empty")
	}
	if len(opts.Rows) == 0 {
		return emptyCard(c, opts.Empty)
	}
	gap := opts.Gap
	if gap <= 0 {
		gap = u * 0.5
	}

	return ui.Column(c).FillWidth().Gap(gap).Children(func() {
		for i, row := range opts.Rows {
			i, row := i, row
			wellRow(c, row, opts.Selected != nil && *opts.Selected == i)
		}
	})
}

// wellRow is one row: the gutter, and whatever the kind's content is.
func wellRow(c *ui.Context, row WellRow, selected bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	_, ink := StatusTone(row.Status).Pair(k)
	if StatusTone(row.Status) == core.Neutral {
		ink = k.TextFaint
	}

	e := ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u*1.5).
		Padding(u*0.75, u).Radius(theme.SmallRadius).Children(func() {
		// The gutter is a fixed width so that every row's content starts in
		// the same place, which is what makes the column scannable at all.
		ui.Column(c).Width(u * 13).Shrink(0).AlignItems(ui.End).Gap(u * 0.25).Children(func() {
			if row.At != "" {
				ui.Text(c, row.At).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			}
			ui.Box(c).Size(u*2.25, u*2.25).Radius(u * 1.25).Shrink(0).
				Background(ink).Label(row.Kind.String() + ": " + row.Status.String())
		})
		// The content is built here rather than beside the row: an element
		// belongs to whatever was being built when it was made, so a column
		// made out here would land in the caller's column instead of beside
		// the gutter, and the row would measure as an empty strip.
		ui.Column(c).Grow(1).Shrink(0).Gap(u * 0.5).Children(func() {
			if row.Agent != "" {
				AgentCaption(c, AgentCaptionOptions{Name: row.Agent, At: row.At})
			}
			switch row.Kind {
			case RowStep:
				stepRow(c, stepRowOptions{step: row.Step, showNo: true}, nil, nil)
			case RowThinking:
				display.Text(c, row.Text, display.TextOptions{Muted: true, MaxLines: 4})
			case RowToolCall:
				ToolCallCard(c, ToolCallCardOptions{Call: row.Call})
			case RowCommand:
				CommandExecutionCard(c, CommandExecutionCardOptions{Exec: row.Exec})
			case RowFileChange:
				FileChangeCard(c, FileChangeCardOptions{
					Path: row.File.Path, Change: row.File.Change,
					Added: row.Added, Removed: row.Removed,
				})
			case RowNote:
				panel(c, layout.ContainerOptions{
					Surface: true, Radius: theme.SmallRadius, Pad: u * 1.5,
				}, func() {
					display.Text(c, row.Text, display.TextOptions{MaxLines: 4})
				})
			case RowRequest:
				HumanInputRequest(c, HumanInputRequestOptions{
					From: row.Agent, Question: row.Text, Urgency: row.Status,
					Draft: row.Draft,
				})
			}
		})
	})
	e.Label(row.Kind.String() + " row")
	if selected {
		e.Background(k.SurfacePressed)
	}
}
