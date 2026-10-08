package feedback

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// TraceEvent is one span on a TraceViewer's timeline: what happened, when,
// and how long it took.
type TraceEvent struct {
	// Name is what happened on the timeline: a fetch, a model call, a tool
	// run. It is required: a dot with a time and no span beside it cannot
	// be traced back to the code that made it.
	Name string
	// At is when it happened, already formatted by the caller, for the same
	// reason a Notification's is: the library has no clock and will not
	// keep one, and a trace is as much a log as a drawing.
	At string
	// Duration is how long it took, formatted the same way. Empty draws
	// the row without a duration, for a span that is still open when the
	// timeline is drawn — a trace of a run in flight.
	Duration string
	// Severity marks the row with a dot at its leading edge. Zero is
	// core.Neutral, which is right for the ordinary spans; a trace where
	// every dot is danger is a trace of nothing interesting, and the
	// neutral ink is what keeps the loud ones loud.
	Severity core.Severity
}

// TraceViewerOptions configure a TraceViewer.
type TraceViewerOptions struct {
	// Events are the spans, in the order they are shown, which is usually
	// the order they happened. A timeline that reorders itself is a
	// timeline the reader cannot follow with the log next to it.
	Events []TraceEvent
	// Title heads the timeline; empty takes the library's "Trace".
	Title string
	// Rule draws a hairline between rows, for a timeline long enough that
	// the reader has to lose and find the place. Off by default, because a
	// few rows are read as one list, and a rule between every pair of them
	// makes the list into a table it is not.
	Rule bool
	// Width bounds the timeline; zero lets it fill its parent.
	Width float32
}

// TraceViewer is a run's timeline: one row per span, each with its time,
// its duration, and a mark for how the span went.
//
// It is a list and not a Gantt, deliberately: a trace is read the way a
// log is read, top to bottom, and what the reader wants from a row is
// which span was the slow or the failed one, which a mark and a duration
// say without an axis. A span that took long enough to matter has a
// duration the reader can read as a number, and a chart whose bars are
// too small to compare against each other is decoration with gridlines.
//
// The times come in formatted, for the reason a Notification's do:
// rendering "14:02:01" from a moment is a decision about time zones and
// clocks that belongs to the caller's log, and a component that made it
// would be a component whose rows disagreed with the log beside it.
func TraceViewer(c *ui.Context, opts TraceViewerOptions) *ui.Element {
	if len(opts.Events) == 0 {
		panic("feedback: TraceViewer needs at least one TraceEvent; a " +
			"timeline of nothing is a heading with nothing under it")
	}
	for _, ev := range opts.Events {
		if ev.Name == "" {
			panic("feedback: a trace event needs a Name; a dot with a time " +
				"and no span beside it cannot be traced")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	title := opts.Title
	if title == "" {
		title = core.Msg(c, "feedback.traceViewer.title", "Trace")
	}

	col := ui.Column(c).FillWidth().Gap(u * 1.5).Role(ui.RoleList)
	if opts.Width > 0 {
		col.Width(opts.Width)
	}
	col.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.TitleSize)).Bold()
		})
		for i, ev := range opts.Events {
			if i > 0 && opts.Rule {
				layout.Divider(c, layout.DividerOptions{})
			}
			traceRow(c, ev)
		}
	})
	return col
}

// traceRow draws one span: its mark, its name, and its figures on the far
// side of the row — the time in the faint ink, the duration beside it.
func traceRow(c *ui.Context, ev TraceEvent) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	dot := u * 2.25

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(ev.Name).Children(func() {
		ui.Box(c).Size(dot, dot).Radius(dot/2).Shrink(0).
			Background(severityInk(k, ev.Severity)).
			Margin(u*0.5, 0, 0, 0)
		ui.Text(c, ev.Name).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize))
		ui.Box(c).Grow(1)
		if ev.At != "" {
			ui.Text(c, ev.At).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		if ev.Duration != "" {
			ui.Text(c, ev.Duration).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
}
