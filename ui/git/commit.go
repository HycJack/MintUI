package git

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// CommitListOptions configure a CommitList.
type CommitListOptions struct {
	// Height is the height of the table. It is required, and for the reason
	// data.DataTable requires one: a commit list with no height is every
	// commit in the history.
	Height float32
	// Width is the width the columns share; zero measures the window.
	Width float32
	// Selected is the row the keys move from, -1 for none. A click chooses
	// it and the caller reads it.
	Selected *int
	// Choice is the set of rows a table that chooses several needs; nil for
	// the one-row-at-a-time list this usually is.
	Choice *data.Selectable
	// Sort is the column the rows are ordered by. A click on a header writes
	// it and the caller reorders with it — the table keeps no order.
	Sort *data.Sort
	// ShowRefs draws the branch and tag names pointing at each commit.
	ShowRefs bool
	// State is where the list keeps its place between frames.
	State *ui.ListState
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// CommitListResult carries a CommitList and what was asked of it.
type CommitListResult struct {
	// Element is the whole table.
	Element *ui.Element
	// sorted is the column whose head was clicked this frame.
	sorted string
}

// Sorted returns the column the rows were asked to be ordered by this frame,
// empty when none was. The caller reorders with data.Rows and asks again.
func (r CommitListResult) Sorted() string { return r.sorted }

// CommitList is a table of commits: the graph's dot, the short hash, the
// subject, and who and when.
//
// It is data.DataTable rather than a column of rows, which is the point of
// the whole package: a log of ten thousand commits has to cost what a log of
// thirty does, has to scroll under a head that stays, and has to sort by the
// author with a click on the author's column. Reusing the table gives all
// three, and reusing the table means this component has nothing to get wrong.
//
// The columns are the log's own four, and their widths follow the same rule
// every column in the library follows: the subject shares what the others
// leave, and the others are the width their content needs.
func CommitList(c *ui.Context, commits []Commit, opts CommitListOptions) CommitListResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	cols := []data.Column{
		{ID: "subject", Title: "Commit", Share: 3},
	}
	if opts.ShowRefs {
		cols = append(cols, data.Column{ID: "refs", Title: "Refs", Width: 140})
	}
	cols = append(cols,
		data.Column{ID: "author", Title: "Author", Width: 140},
		data.Column{ID: "when", Title: "When", Width: 110},
	)

	cellLabel := func(row, col int) string {
		cm := commits[row]
		switch cols[col].ID {
		case "author":
			return cm.Author
		case "when":
			return cm.When
		case "refs":
			return strings.Join(cm.Refs, ", ")
		}
		return cm.short() + " " + cm.Subject
	}

	tr := data.DataTable(c, data.DataTableOptions{
		Columns: cols, Rows: len(commits), Height: opts.Height, Width: opts.Width,
		Selected: opts.Selected, Choice: opts.Choice, Sort: opts.Sort,
		State: opts.State,
		Label: func(row int) string { return cellLabel(row, 0) },
		CellLabel: func(row, col int) string {
			return commits[row].Subject + ": " + cellLabel(row, col)
		},
		Key:  func(row int) any { return commits[row].Hash },
		Cell: func(row, col int) { commitCell(c, commits[row], cols[col].ID, u) },
		Empty: func() {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, noRows(c, "git.noCommits", core.Def("No commits"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
		},
	})
	return CommitListResult{Element: tr.Element, sorted: tr.Sorted()}
}

// commitCell is one cell of the log, by column. The subject cell is the only
// one with anything in it: the author and the date are strings, and a string
// in a table cell needs nothing from the component that drew it.
//
// "commit" and "subject" name the same cell. FileHistory calls its first
// column "commit" and CommitList calls it "subject", and a switch carrying
// only one of the two names leaves the other table's first column empty —
// which no test caught, because the row's accessible label carries the
// subject either way and HasText reads labels too.
func commitCell(c *ui.Context, cm Commit, id string, u float32) {
	k := core.Tokens(c)
	switch id {
	case "subject", "commit":
		row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5)
		row.Children(func() {
			commitDot(c, len(cm.Parents) > 1)
			hashText(c, cm.short())
			ui.Text(c, cm.Subject).Grow(1).Ellipsis(cm.Subject).
				FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
		})
	case "refs":
		ui.Row(c).Gap(u * 0.5).Children(func() {
			for _, ref := range cm.Refs {
				// A ref the shape of whose name says what it is: refs/heads
				// and refs/tags are two words here rather than four.
				tone := core.Accent
				if strings.HasPrefix(ref, "tag:") {
					tone = core.Warning
				}
				refChip(c, strings.TrimPrefix(ref, "tag:"), tone)
			}
		})
	case "author":
		ui.Text(c, cm.Author).TextColor(k.TextMuted).SingleLine().
			FontSize(core.FontSize(c, theme.MetaSize))
	case "when":
		ui.Text(c, cm.When).TextColor(k.TextMuted).SingleLine().
			FontSize(core.FontSize(c, theme.MetaSize))
	}
}

// CommitGraphOptions configure a CommitGraph.
type CommitGraphOptions struct {
	// Width is how many lane columns the graph draws across. Zero lays the
	// history out in as many lanes as it needs.
	Width int
	// Lane is the distance between two lanes in DIPs; zero gives the
	// library's own.
	Lane float32
	// Selected is the commit the caller has chosen, by hash. Empty selects
	// none, which draws no row differently from any other.
	Selected string
	// ShowRefs draws the branch and tag names beside each commit.
	ShowRefs bool
	// RowHeight is the height of one row; zero is the library's own.
	RowHeight float32
	// Empty draws instead of the graph when there are no commits.
	Empty func()
}

// CommitGraphResult carries a CommitGraph and the commit chosen in it.
type CommitGraphResult struct {
	// Element is the graph.
	Element *ui.Element
	// picked is the commit chosen this frame, empty for none.
	picked string
	// rows is the layout the graph was drawn from, so a caller can put a
	// cursor on a lane or measure a row without computing the graph again.
	rows []GraphRow
}

// Picked returns the hash of the commit chosen this frame, empty for none.
func (r CommitGraphResult) Picked() string { return r.picked }

// Rows returns the layout the graph was drawn from. It is the same slice
// CommitsFor computed, handed back so that whatever else wants to draw in
// the lanes — a cursor, a highlight, a selection — draws them where the graph
// drew them rather than working it out a second time.
func (r CommitGraphResult) Rows() []GraphRow { return r.rows }

// CommitGraph is a column of commits with the lane lines beside them.
//
// The lines are not drawn here: where each node sits and which lanes are
// still open is computed by CommitsFor, and this component only paints what
// it said. That is what makes the awkward histories testable — a merge that
// opens a lane, a branch that closes one — without a picture of a graph at
// all.
//
// A commit is chosen by clicking its row, and the choice is reported rather
// than held: the graph has no idea what a caller does with "the commit under
// the pointer", which is sometimes a diff, sometimes a reset, and sometimes
// only a row to scroll into view.
func CommitGraph(c *ui.Context, commits []Commit, opts CommitGraphOptions) CommitGraphResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	rows := CommitsFor(commits, opts.Width)

	lane := opts.Lane
	if lane <= 0 {
		lane = u * 2.5
	}
	h := opts.RowHeight
	if h <= 0 {
		h = u * 7
	}

	var res CommitGraphResult
	res.rows = rows

	col := ui.Column(c).FillWidth().Gap(0)
	col.Children(func() {
		if len(rows) == 0 {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, noRows(c, "git.noCommits", core.Def("No commits"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			return
		}
		for _, row := range rows {
			cm := row.Commit
			chosen := opts.Selected != "" && opts.Selected == cm.Hash
			line := pickRow(c, cm.Subject, chosen, func() {
				graphGutter(c, row, lane)
				ui.Text(c, cm.Subject).Grow(1).Ellipsis(cm.Subject).
					FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
				if opts.ShowRefs && len(cm.Refs) > 0 {
					for _, ref := range cm.Refs {
						refChip(c, ref, core.Accent)
					}
				}
			})
			line.MinHeight(h)
			if line.Clicked() {
				res.picked = cm.Hash
			}
		}
	})
	res.Element = col
	return res
}

// graphGutter is the lane drawing of one row: a line down every open lane and
// a node in this row's own.
//
// The node is drawn over the line rather than beside it, which is what makes
// a merge read as a merge: two lines arriving at one dot is a shape, and the
// same two lines drawn with a gap between them is just two lines.
func graphGutter(c *ui.Context, row GraphRow, lane float32) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	widest := row.Width
	if widest < 1 {
		widest = 1
	}
	w := float32(widest) * lane

	ui.Box(c).Width(w).FillHeight().Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			mid := r.Y + r.H/2
			for _, l := range row.Lanes {
				x := r.X + float32(l)*lane + lane/2
				p.Line(x, r.Y, x, r.Y+r.H, u*0.8, k.Border)
			}
			x := r.X + float32(row.Column)*lane + lane/2
			if row.Merge {
				// A merge gets a ring rather than a dot: the same shape as a
				// tip means "this commit joins two lines", which a filled
				// circle cannot say.
				var path ui.Path
				path.Circle(x, mid, lane*0.28)
				p.FillPath(&path, k.Background)
				p.StrokePath(&path, u*0.9, k.Text)
				return
			}
			p.Fill(ui.Rect{X: x - lane*0.22, Y: mid - lane*0.22, W: lane * 0.44, H: lane * 0.44}, k.Text, lane*0.22)
		})
}

// CommitInputOptions configure a CommitInput.
type CommitInputOptions struct {
	// Label names the message field. It is required, for the same reason
	// every text field in the library requires one.
	Label string
	// Placeholder is what the empty field says.
	Placeholder string
	// Amend is the "commit to the last commit instead of a new one" box, in
	// the caller's state — because which commit is last is the caller's
	// repository, and whether the box starts ticked is the caller's too.
	Amend *bool
	// Staged counts what would go into the commit, shown under the button so
	// that a commit of nothing is visible before it is pressed rather than
	// after it fails.
	Staged int
	// Height is the height of the message field; zero gives the library's
	// own.
	Height float32
}

// CommitInputResult carries a CommitInput and whether it was committed.
type CommitInputResult struct {
	// Element is the whole panel.
	Element *ui.Element
	// committed reports the commit button being pressed this frame.
	committed bool
}

// Committed reports the commit button being pressed this frame.
func (r CommitInputResult) Committed() bool { return r.committed }

// CommitInput is the box a commit is written in: the message, the box for
// amending, the count of what is staged, and the button.
//
// The message is the caller's string the moment it is typed, so a caller that
// wants to validate as somebody types needs no commit step and no second copy
// to keep in step. The button reports the press and nothing else: what
// committing means — running git, refusing an empty message, asking about a
// hook — belongs to whoever owns the repository, and a component that ran it
// would be the one place in the interface with a working directory.
//
// The button is disabled while the message is blank, which is the one thing
// this component decides on its own. It is decided rather than left to the
// caller because an empty commit is refused by git anyway, and a button that
// offers it teaches the wrong thing about what a commit is.
func CommitInput(c *ui.Context, message *string, opts CommitInputOptions) CommitInputResult {
	if message == nil {
		panic("git: CommitInput needs a message to point at; it keeps no message of its own")
	}
	if opts.Label == "" {
		panic("git: CommitInput needs a Label for the message field")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	h := opts.Height
	if h <= 0 {
		h = u * 10
	}

	var res CommitInputResult
	blank := strings.TrimSpace(*message) == ""

	panel := ui.Column(c).FillWidth().Gap(u * 1.5)
	panel.Children(func() {
		messageField(c, message, opts.Label, opts.Placeholder,
			fieldError(c, blank), h)

		if opts.Amend != nil {
			tickBox(c, opts.Amend, core.Msg(c, "git.amendLast", core.Def("Amend last commit")))
		}

		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			ui.Text(c, stagedWord(c, opts.Staged)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize))
			ui.Box(c).Grow(1)
			btn := press(c, commitWord(c, opts.Amend), true, core.Neutral)
			if btn.Clicked() && !blank {
				res.committed = true
			}
		})
	})
	res.Element = panel
	return res
}

// fieldError is what the field says when it is blank. An error rather than a
// hint, because the button beside it is the only other place the rule is
// written and a message that is not an error reads as a suggestion.
func fieldError(c *ui.Context, blank bool) string {
	if !blank {
		return ""
	}
	return core.Msg(c, "git.commitMessageEmpty", core.Def("A commit needs a message"))
}

// stagedWord is what would go into the commit, counting the word, so that
// "staged: 1" never reads as a claim about a file.
func stagedWord(c *ui.Context, n int) string {
	if n <= 0 {
		return core.Msg(c, "git.nothingStaged", core.Def("Nothing staged"))
	}
	return core.Msg(c, "git.staged", core.Def("Staged: "+itoa(n)))
}

func commitWord(c *ui.Context, amend *bool) string {
	if amend != nil && *amend {
		return core.Msg(c, "git.amend", core.Def("Amend"))
	}
	return core.Msg(c, "git.commit", core.Def("Commit"))
}
