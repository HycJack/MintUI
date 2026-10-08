package feedback

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// EvalRow is one line of an EvalResultTable: what was evaluated, how it
// scored, and whether it passed.
type EvalRow struct {
	// Name is what was evaluated: a suite, a checkpoint, a criterion. It is
	// required: a score with no name beside it is a figure with no claim.
	Name string
	// Score is the figure to show for it, already formatted by the caller —
	// a percentage, a time, a verdict like "over budget" — for the same
	// reason a TraceEvent's duration is: the library does not guess which
	// of the three a score is.
	Score string
	// Pass is whether it passed, which decides the row's mark: the success
	// ink when it did, the danger ink when it did not.
	Pass bool
}

// EvalResultTableOptions configure an EvalResultTable.
type EvalResultTableOptions struct {
	// Rows are the results, in the order to show them, which is the
	// caller's: the library does not sort pass before fail, because a suite
	// is often read in the order it ran, and reordering it would make the
	// table disagree with the log next to it.
	Rows []EvalRow
	// Title heads the table; empty takes the library's "Evaluation".
	Title string
	// Width bounds the table; zero lets it fill its parent.
	Width float32
}

// EvalResultTable is the outcome of a run of checks: one row per result, a
// mark for each, and a count over the list of how many passed.
//
// The count is in the header rather than the footer, the way a window's
// heading states the window, because the number a caller is usually after
// is the verdict — "three of five passed" — and a verdict at the bottom of
// a list is a verdict the reader has to read the whole list to get. Each
// row's mark is a plain dot in the severity's ink rather than a pill, for
// the reason a notification's is: a column of pills would be a column of
// opinions, and this is a column of outcomes.
func EvalResultTable(c *ui.Context, opts EvalResultTableOptions) *ui.Element {
	if len(opts.Rows) == 0 {
		panic("feedback: EvalResultTable needs at least one EvalRow; a " +
			"header counting zero results claims a run that never happened")
	}
	for _, row := range opts.Rows {
		if row.Name == "" {
			panic("feedback: an eval row needs a Name; a score with no name " +
				"beside it is a figure with no claim")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	passed := 0
	for _, row := range opts.Rows {
		if row.Pass {
			passed++
		}
	}
	title := opts.Title
	if title == "" {
		title = core.Msg(c, "feedback.evalResultTable.title", "Evaluation")
	}
	// The count is a template through core.Msg, the way a window localises
	// a sentence with numbers in it: the translation has to carry the same
	// two verbs, and a translation that drops one is a translation that
	// reads "3 passed" for "3 of 5 passed", which is a different claim.
	count := fmt.Sprintf(
		core.Msg(c, "feedback.evalResultTable.passed", "%d of %d passed"),
		passed, len(opts.Rows))

	e := ui.Column(c).FillWidth().Gap(u * 2).Role(ui.RoleTable)
	if opts.Width > 0 {
		e.Width(opts.Width)
	}
	e.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).Children(func() {
			ui.Text(c, title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.TitleSize)).Bold()
			ui.Box(c).Grow(1)
			ui.Text(c, count).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		})
		for _, row := range opts.Rows {
			evalRow(c, row)
		}
	})
	return e
}

// evalRow draws one result: its mark, its name, and its figure on the far
// side of the row.
func evalRow(c *ui.Context, row EvalRow) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	dot := u * 2.25

	sev, word := core.Success,
		core.Msg(c, "feedback.evalResultTable.pass", "passed")
	if !row.Pass {
		sev, word = core.Danger,
			core.Msg(c, "feedback.evalResultTable.fail", "failed")
	}

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 2).
		Label(row.Name + " " + word).Children(func() {
		ui.Box(c).Size(dot, dot).Radius(dot / 2).Shrink(0).
			Background(severityInk(k, sev))
		ui.Text(c, row.Name).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.RowSize))
		ui.Box(c).Grow(1)
		ui.Text(c, row.Score).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
	})
}
