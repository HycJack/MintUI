package git

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Conflict is one hunk git could not merge, with the three versions of it.
//
// The three are separate fields rather than one string with markers already
// in it, because a resolver's whole job is to take the markers apart: a
// component handed a string has to re-parse it to highlight a side, and
// re-parsing a merge conflict is exactly the work that goes wrong on the one
// file that matters.
type Conflict struct {
	// Label names the hunk for assistive technology and for the header.
	Label string
	// Ours is what this branch had.
	Ours []string
	// Theirs is what came in from the other side.
	Theirs []string
	// Base is what both started from, when the caller has it. It is what
	// makes "keep both" reviewable rather than a guess.
	Base []string
}

// Resolution is what a caller decided to do with a conflict.
type Resolution int

const (
	// Unresolved is the zero value: nobody has decided, which is what a
	// freshly merged tree is full of.
	Unresolved Resolution = iota
	// KeepOurs takes this branch's side.
	KeepOurs
	// KeepTheirs takes the other side's.
	KeepTheirs
	// KeepBoth keeps every line of both, ours first.
	KeepBoth
	// KeepNeither drops the hunk, leaving the file without it.
	KeepNeither
)

func (r Resolution) String() string {
	switch r {
	case KeepOurs:
		return "ours"
	case KeepTheirs:
		return "theirs"
	case KeepBoth:
		return "both"
	case KeepNeither:
		return "neither"
	}
	return "unresolved"
}

// Apply is what the resolution does to the hunk: the lines the resolved file
// has where these three versions are.
//
// It is a function rather than a drawing because resolving a merge is
// arithmetic, and a file that has been resolved by a painter is a file
// nobody can check. Every branch of it is testable without a window, which
// is the only way to be sure "keep both" does not silently drop a line.
func (r Resolution) Apply(conf Conflict) []string {
	switch r {
	case KeepOurs:
		return append([]string(nil), conf.Ours...)
	case KeepTheirs:
		return append([]string(nil), conf.Theirs...)
	case KeepBoth:
		out := append([]string(nil), conf.Ours...)
		return append(out, conf.Theirs...)
	case KeepNeither:
		return nil
	}
	// Unresolved keeps ours rather than nothing: a file with the conflict
	// markers still in it can be opened and read, and one with the hunk
	// dropped cannot. Whoever resolves the rest writes the file.
	return append([]string(nil), conf.Ours...)
}

// ConflictResolverOptions configure a ConflictResolver.
type ConflictResolverOptions struct {
	// Choices is the caller's resolutions, one per conflict. A nil entry is
	// the zero value, Unresolved.
	Choices []Resolution
	// Width is the width of the panel; zero measures the window.
	Width float32
	// Ours and Theirs are the words for the two sides. They are the
	// caller's to word, because "ours" is "yours" in one window and
	// "upstream" in another, and a merge tool that insists on one word gets
	// it backwards at least once.
	Ours, Theirs string
}

// ConflictResolverResult carries a ConflictResolver and what was decided.
type ConflictResolverResult struct {
	// Element is the panel.
	Element *ui.Element
	// changed is the hunk resolved this frame, -1 for none.
	changed int
	// resolution is what it was resolved to.
	resolution Resolution
}

// Changed returns the index of the hunk decided this frame, -1 for none.
func (r ConflictResolverResult) Changed() int { return r.changed }

// Resolution returns what the hunk decided this frame was resolved to.
func (r ConflictResolverResult) Resolution() Resolution { return r.resolution }

// ConflictResolver is one conflict and the four ways out of it.
//
// The decision is the caller's: a slice of Resolutions, one per hunk, that
// this writes and a caller reads. Holding it here would mean the file the
// caller is about to write and the buttons somebody pressed were two
// separate pieces of state, which is the failure this package's convention
// exists to make impossible.
//
// All four buttons are on every hunk rather than one button and a menu: a
// merge is resolved in a hurry, often by somebody who has done it before, and
// a menu is a click that produces nothing visible before it has been chosen
// from.
func ConflictResolver(c *ui.Context, conflicts []Conflict, opts ConflictResolverOptions) ConflictResolverResult {
	if opts.Choices == nil {
		panic("git: ConflictResolver needs the caller's choices to point at; it resolves nothing on its own")
	}
	if len(opts.Choices) < len(conflicts) {
		panic("git: ConflictResolver has " + itoa(len(conflicts)) + " conflicts and " +
			itoa(len(opts.Choices)) + " choices; the caller's slice decides what each one is resolved to")
	}
	ours, theirs := opts.Ours, opts.Theirs
	if ours == "" {
		ours = core.Msg(c, "git.ours", core.Def("Ours"))
	}
	if theirs == "" {
		theirs = core.Msg(c, "git.theirs", core.Def("Theirs"))
	}

	var res ConflictResolverResult
	res.changed = -1

	panel := ui.Column(c).FillWidth().Gap(u2(c)).Width(opts.Width)
	panel.Children(func() {
		if len(conflicts) == 0 {
			ui.Text(c, noRows(c, "git.noConflicts", core.Def("No conflicts"))).
				TextColor(core.Tokens(c).TextMuted)
			return
		}
		for i, conf := range conflicts {
			i, conf := i, conf
			head := conf.Label
			if head == "" {
				head = core.Msg(c, "git.conflict", core.Def("Conflict")) + " " + itoa(i+1)
			}
			card := ui.Column(c).FillWidth().Gap(u2(c)).Padding(u2(c)).
				Radius(theme.CardRadius).Background(core.Tokens(c).Surface)
			card.Children(func() {
				ui.Text(c, head).Bold().FontSize(core.FontSize(c, theme.RowSize)).
					Label("conflict: " + head)
				mergeSides(c, ours, theirs, conf)
				layout.Divider(c, layout.DividerOptions{})
				ui.Row(c).FillWidth().Gap(u2(c) * 0.75).Children(func() {
					for _, opt := range []struct {
						r    Resolution
						name string
					}{
						{KeepOurs, ours},
						{KeepTheirs, theirs},
						{KeepBoth, core.Msg(c, "git.keepBoth", core.Def("Both"))},
						{KeepNeither, core.Msg(c, "git.keepNeither", core.Def("Neither"))},
					} {
						opt := opt
						chosen := opts.Choices[i] == opt.r
						btn := press(c, opt.name, false, severityOf(chosen))
						btn.Grow(1)
						if btn.Clicked() {
							opts.Choices[i] = opt.r
							res.changed = i
							res.resolution = opt.r
						}
					}
				})
			})
		}
	})
	res.Element = panel
	return res
}

// mergeSides is the two versions of a conflicting hunk, side by side, each
// under its own heading so that "ours" and "theirs" are attached to the
// text rather than to a column position somebody has to remember.
func mergeSides(c *ui.Context, ours, theirs string, conf Conflict) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Stretch).Children(func() {
		side := func(title string, lines []string, ink ui.Color) {
			col := ui.Column(c).Grow(1).Gap(u * 0.25).Children(func() {
				ui.Text(c, title).TextColor(ink).Bold().
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				for _, line := range lines {
					textOf(c, line, ink)
				}
			})
			_ = col
		}
		side(ours, conf.Ours, k.Accent)
		side(theirs, conf.Theirs, k.Warning)
	})
}

// severityOf is how loud a resolution button is: the chosen one takes the
// accent, so the state of every hunk can be read down the panel without
// reading a single word.
func severityOf(chosen bool) core.Severity {
	if chosen {
		return core.Accent
	}
	return core.Neutral
}

// ThreeWayMergeOptions configure a ThreeWayMerge.
type ThreeWayMergeOptions struct {
	// Height is the height of the scrolling area. It is required, as
	// everywhere else here that scrolls.
	Height float32
	// Base, Ours and Theirs are the three files, one line each, in the same
	// order. Their lengths may differ — that is what a merge is — but the
	// caller has to be able to say which line is which, which is why these
	// are lines and not a patch.
	Base, Ours, Theirs []string
	// OursLabel and TheirsLabel word the two sides.
	OursLabel, TheirsLabel string
	// Choices is the caller's resolution per row, where a row is a line of
	// ours; nil draws the three columns without the decision controls.
	Choices []Resolution
}

// ThreeWayMergeResult carries a ThreeWayMerge and what was decided.
type ThreeWayMergeResult struct {
	// Element is the whole thing.
	Element *ui.Element
	// changed is the row decided this frame, -1 for none.
	changed int
}

// Changed returns the index of the row decided this frame, -1 for none.
func (r ThreeWayMergeResult) Changed() int { return r.changed }

// ThreeWayMerge is the base, ours and theirs side by side, with the decision
// controls down the right.
//
// The three columns are the caller's three line slices rather than three
// patches, because a three-way view is a comparison and a comparison needs
// lines that line up. Deciding which lines correspond is the caller's job —
// it knows the merge base, the ancestry and what the tool actually did — and
// this component's job is to show the three and let each row be decided.
//
// A row decided is reported rather than applied here. Writing the resolved
// file is the one operation in a merge that must not be a side effect of
// drawing a picture: it happens when the caller says so, and not once per
// frame per rebuild.
func ThreeWayMerge(c *ui.Context, opts ThreeWayMergeOptions) ThreeWayMergeResult {
	if opts.Height <= 0 {
		panic("git: ThreeWayMerge needs a Height; a merge with no height is every line of it")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	ours, theirs := opts.OursLabel, opts.TheirsLabel
	if ours == "" {
		ours = core.Msg(c, "git.ours", core.Def("Ours"))
	}
	if theirs == "" {
		theirs = core.Msg(c, "git.theirs", core.Def("Theirs"))
	}

	var res ThreeWayMergeResult
	res.changed = -1

	area := layout.ScrollArea(c, layout.ScrollAreaOptions{
		Vertical: true, Horizontal: true, Height: opts.Height,
	}, func() {
		head := ui.Row(c).Gap(u * 2).FillWidth().Children(func() {
			headCell(c, "Base", k.TextMuted)
			headCell(c, ours, k.Accent)
			headCell(c, theirs, k.Warning)
		})
		head.MinHeight(u * 6)
		layout.Divider(c, layout.DividerOptions{})

		for i := range opts.Ours {
			i := i
			row := ui.Row(c).FillWidth().Gap(u * 2).Shrink(0).MinHeight(u * 5)
			row.Children(func() {
				textOf(c, lineAt(opts.Base, i), k.TextMuted)
				textOf(c, lineAt(opts.Ours, i), k.Text)
				textOf(c, lineAt(opts.Theirs, i), k.Text)
				if opts.Choices != nil {
					decision(c, opts.Choices, i, &res.changed)
				}
			})
		}
	})
	res.Element = area.Element
	return res
}

// decision is the four buttons of one row. They are only drawn when the
// caller gave a choices slice, because a three-way view used to look at
// rather than resolve is a legitimate thing to want and it should not carry
// forty buttons nobody pressed.
func decision(c *ui.Context, choices []Resolution, i int, changed *int) {
	u := core.Density(c).Unit()
	if i >= len(choices) {
		panic("git: a merge row has no resolution to write; give Choices one entry per row of ours")
	}
	ui.Row(c).Gap(u * 0.5).Children(func() {
		for _, r := range []Resolution{KeepOurs, KeepTheirs, KeepBoth, KeepNeither} {
			r := r
			btn := press(c, r.String(), false, severityOf(choices[i] == r))
			if btn.Clicked() {
				choices[i] = r
				*changed = i
			}
		}
	})
}

// headCell is one column's name over a column of lines.
func headCell(c *ui.Context, label string, ink ui.Color) {
	ui.Box(c).Grow(1).Children(func() {
		ui.Text(c, label).TextColor(ink).Bold().
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// textOf is one line of one side. A line the other side does not have is
// shown as a gap rather than as an empty string, so that "this side has
// nothing here" and "this side has an empty line here" do not look alike.
func textOf(c *ui.Context, s string, ink ui.Color) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	col := ui.Box(c).Grow(1).MinHeight(u*5).Radius(theme.SmallRadius).
		Background(k.Surface).Padding(u*0.5, u)
	col.Children(func() {
		if s == "" {
			ui.Text(c, "·").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.RowSize))
			return
		}
		ui.Text(c, s).TextColor(ink).FontSize(core.FontSize(c, theme.RowSize))
	})
}

// lineAt is the i'th line, or "" when that side is shorter. The two are
// different and the difference is the whole of a merge.
func lineAt(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}

// u2 is the unit at two steps, which is the gap between two things in the
// same panel rather than the gap inside one.
func u2(c *ui.Context) float32 { return core.Density(c).Unit() * 2 }

// BlameLine is one line of a file and who last wrote it.
type BlameLine struct {
	// Text is the line itself.
	Text string
	// Author is who last changed it, already formatted.
	Author string
	// Commit is the short hash of that change.
	Commit string
	// When is when, already formatted.
	When string
	// Date is when, in the second column's own format. Blame shows the
	// relative date by default and the absolute one beside it, because
	// "3 years ago" is what tells a reader whether to care and
	// "2019-04-02" is what tells them what to write in a bug report.
	Date string
}

// BlameViewOptions configure a BlameView.
type BlameViewOptions struct {
	// Height is the height of the viewport; required.
	Height float32
	// WithGutter draws the line numbers down the left. Off by default
	// because the left two columns of a blame are the author's name and the
	// date, and a third column of numbers is a third thing to read.
	WithGutter bool
	// WithDate draws the absolute date beside the author.
	WithDate bool
	// State is where the viewport keeps its place between frames.
	State *ui.ScrollState
}

// BlameView is a file with its author down the left, the way git blame draws
// it.
//
// The colour of an author's rows comes from their name rather than from the
// caller, so that two runs of the same author are the same colour: the whole
// point of the view is seeing one person's work as one block, and a caller
// who had to assign a colour per line would be the only thing standing
// between the view and that.
func BlameView(c *ui.Context, lines []BlameLine, opts BlameViewOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Height <= 0 {
		panic("git: BlameView needs a Height; blame with no height is the whole file")
	}
	w := u * 13

	area := layout.ScrollArea(c, layout.ScrollAreaOptions{
		Vertical: true, Horizontal: true, Height: opts.Height, State: opts.State,
	}, func() {
		for i, l := range lines {
			row := ui.Row(c).FillWidth().Shrink(0).Gap(u * 1.5).AlignItems(ui.Stretch)
			row.Children(func() {
				if opts.WithGutter {
					no := ui.Text(c, itoa(i+1)).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize))
					no.Width(u * 3).TextAlign(ui.End)
				}
				ui.Box(c).Width(w).Shrink(0).Children(func() {
					ui.Text(c, l.Author).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
				})
				hashText(c, l.Commit)
				ui.Text(c, l.When).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				if opts.WithDate && l.Date != "" {
					ui.Text(c, l.Date).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				}
				ui.Text(c, l.Text).Grow(1).FontSize(core.FontSize(c, theme.RowSize))
			})
			row.Label(l.Author + ", " + l.Commit + ": " + l.Text)
		}
	})
	return area.Element
}

// FileHistoryOptions configure a FileHistory.
type FileHistoryOptions struct {
	// Height is the height of the list; required.
	Height float32
	// Width is the width of the table; zero measures the window.
	Width float32
	// Selected is the row the keys move from, -1 for none.
	Selected *int
	// Path names the file the history belongs to, for assistive technology.
	Path string
	// Sort is the column the rows are ordered by.
	Sort *data.Sort
	// State is where the list keeps its place between frames.
	State *ui.ListState
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// FileHistoryResult carries a FileHistory and what was asked of it.
type FileHistoryResult struct {
	// Element is the whole table.
	Element *ui.Element
	// sorted is the column whose head was clicked this frame.
	sorted string
}

// Sorted returns the column the rows were asked to be ordered by this frame,
// empty when none was. The chosen commit is not here: it is the caller's
// Selected pointer, which is the one place a row's choice lives.
func (r FileHistoryResult) Sorted() string { return r.sorted }

// FileHistory is the commits that touched one file, newest first — the log
// with a path on it.
//
// It is CommitList over a shorter list, and saying so rather than writing a
// second one is the point: a file's history and the repository's history have
// the same columns in the same order for a reason, and two tables that could
// drift apart would drift.
func FileHistory(c *ui.Context, path string, commits []Commit, opts FileHistoryOptions) FileHistoryResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if path == "" {
		panic("git: FileHistory needs the path it is the history of")
	}
	if opts.Height <= 0 {
		panic("git: FileHistory needs a Height; a history with no height is every commit that touched the file")
	}

	cols := []data.Column{
		{ID: "commit", Title: "Commit", Share: 2},
		{ID: "author", Title: "Author", Width: 140},
		{ID: "when", Title: "When", Width: 110},
	}
	tr := data.DataTable(c, data.DataTableOptions{
		Columns: cols, Rows: len(commits), Height: opts.Height, Width: opts.Width,
		Selected: opts.Selected, Sort: opts.Sort, State: opts.State,
		Key:  func(row int) any { return commits[row].Hash },
		Cell: func(row, col int) { commitCell(c, commits[row], cols[col].ID, u) },
		Label: func(row int) string {
			return commits[row].Subject + ", by " + commits[row].Author + ", in " + path
		},
		CellLabel: func(row, col int) string {
			return commits[row].Subject + " in " + path
		},
		Empty: func() {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, noRows(c, "git.noHistory", core.Def("No history for this file"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
		},
	})
	return FileHistoryResult{Element: tr.Element, sorted: tr.Sorted()}
}

// AnsiTextOptions configure an AnsiText.
type AnsiTextOptions struct {
	// Size is the text size; zero gives the row size.
	Size float32
	// Muted draws the uncoloured parts in the secondary tone, which is what
	// terminal output wants: the colour is the message and everything else
	// is context.
	Muted bool
	// Mono draws in the monospaced face at the caption size, which is what
	// the face of a diff or a log is.
	Mono bool
	// MaxLines truncates with an ellipsis past this many lines; zero is
	// unlimited.
	MaxLines int
}

// AnsiText is text that came out of a program with colour in it.
//
// Git is the one program on a desktop that hands a window text already
// coloured by a palette that knows nothing about this window. There are two
// honest ways to deal with that and both are in this package: draw the
// colours, or throw them away. AnsiText draws them — ParseANSI has already
// cut the spans and AnsiInk has already said what each code means in this
// palette, so all that is left is to put the words down in order.
//
// The runs are measured by their stripped length, not by their raw one. A
// wrap computed over the escape sequences wraps in the wrong place, which is
// the whole reason StripANSI exists.
func AnsiText(c *ui.Context, text string, opts AnsiTextOptions) *ui.Element {
	k := core.Tokens(c)
	size := opts.Size
	if size <= 0 {
		size = theme.RowSize
	}
	if opts.Mono {
		size = theme.CaptionSize
	}

	spans := ParseANSI(text)
	col := ui.Column(c).FillWidth().Gap(0)
	col.Children(func() {
		for _, span := range spans {
			if span.Text == "" {
				continue
			}
			// A span is a run between two escapes, and text with no escapes
			// in it at all — which is most of what `git log` and `git
			// status` print — arrives as ONE span holding every line. Drawing
			// that as one Text and then calling SingleLine() on it threw all
			// but the first line away, so a CommandBlock over plain output
			// showed one line of a four-line command. Each line is its own
			// Text here, which is also what a terminal does: an escape
			// changes colour, not line count.
			for _, line := range strings.Split(span.Text, "\n") {
				row := ui.Text(c, line).TextColor(k.Text).
					FontSize(core.FontSize(c, size))
				// The colour is applied to the span as a whole rather than
				// to a run of text painted over it: an escape that changed
				// the foreground mid-word is honoured at the word, which is
				// as fine as text this small can be split without showing
				// the seam.
				ink, bold := spanInk(span.Codes, k)
				if ink != nil {
					row.TextColor(*ink)
				}
				if bold {
					row.Bold()
				}
				if opts.MaxLines > 0 {
					row.MaxLines(opts.MaxLines)
				} else {
					row.SingleLine()
				}
			}
		}
	})
	return col
}

// spanInk is the ink and the weight of one span's codes, the last one that
// says something winning — which is what a terminal does with SGR, where a
// later parameter overrides an earlier one.
func spanInk(codes []int, k theme.Tokens) (*ui.Color, bool) {
	var ink *ui.Color
	bold := false
	for _, code := range codes {
		col, b, ok := AnsiInk(code, k)
		if !ok {
			continue
		}
		// The local is a fresh variable each time round, so a span with
		// several colour codes points at the last one rather than at a
		// loop variable that has moved on.
		last := col
		ink = &last
		bold = b
	}
	if ink == nil {
		return nil, bold
	}
	return ink, bold
}

// CommandResult is what a command run said.
type CommandResult struct {
	// Code is what it exited with: 0 for success, anything else for not.
	Code int
	// Out is what it printed.
	Out string
	// Err is what it printed on its error stream, empty when there was
	// none.
	Err string
}

// OK reports the command having succeeded.
func (r CommandResult) OK() bool { return r.Code == 0 }

// CommandBlockOptions configure a CommandBlock.
type CommandBlockOptions struct {
	// Command is the line that was run, shown above the output. It is
	// required: a block of output with no command above it is a log, and
	// this is a record of one specific act.
	Command string
	// Result is what came back.
	Result CommandResult
	// Height is the height of the output area; zero grows to fit it, which
	// is right for a command that just finished.
	Height float32
	// Collapsed trims the output to the last few lines and says how many
	// were hidden. Zero shows all of it.
	Collapsed int
}

// CommandBlockResult carries a CommandBlock and whether it was re-run.
type CommandBlockResult struct {
	// Element is the whole block.
	Element *ui.Element
	// rerun reports the run button being pressed this frame.
	rerun bool
}

// Rerun reports the run button being pressed this frame. Running it is the
// caller's: this component knows what was run and what came back, and it has
// no working directory to run anything in.
func (r CommandBlockResult) Rerun() bool { return r.rerun }

// CommandBlock is one command that was run and what it said: the line above,
// the output below, and a button to run it again.
//
// The output is drawn with AnsiText rather than as plain text because git
// colours its own output, and the colours are the part that says "this is a
// path that does not exist" as against "this is a path".
//
// The exit code is a badge rather than a word, and it is drawn in the danger
// tone whenever it is non-zero: a command that failed and said nothing has
// still failed, and the failure has to be visible without reading the
// output.
func CommandBlock(c *ui.Context, opts CommandBlockOptions) CommandBlockResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Command == "" {
		panic("git: CommandBlock needs the command that was run; output with no command is a log")
	}

	out := opts.Result.Out
	if opts.Collapsed > 0 {
		out = tailLines(out, opts.Collapsed)
	}
	var res CommandBlockResult

	block := ui.Column(c).FillWidth().Gap(u).Padding(u * 1.5).
		Radius(theme.CardRadius).Background(k.Surface)
	block.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).Children(func() {
			ui.Text(c, "$ "+opts.Command).Grow(1).Ellipsis(opts.Command).
				FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
			if !opts.Result.OK() {
				exitBadge(c, opts.Result.Code)
			}
			btn := press(c, core.Msg(c, "git.rerun", core.Def("Run again")), false, core.Neutral)
			if btn.Clicked() {
				res.rerun = true
			}
		})
		layout.Divider(c, layout.DividerOptions{})

		body := func() {
			if out != "" {
				AnsiText(c, out, AnsiTextOptions{Mono: true})
			}
			if opts.Result.Err != "" {
				AnsiText(c, opts.Result.Err, AnsiTextOptions{Mono: true})
			}
		}
		if opts.Height > 0 {
			layout.ScrollArea(c, layout.ScrollAreaOptions{
				Vertical: true, Height: opts.Height,
			}, body)
		} else {
			ui.Column(c).FillWidth().Gap(u * 0.5).Children(body)
		}
	})
	res.Element = block
	return res
}

// exitBadge is what a command that failed wears. It says the number rather
// than a word, because that is what the caller has to put in a bug report
// and a word like "failed" would have to be turned back into one.
func exitBadge(c *ui.Context, code int) {
	bg, fg := core.Danger.Pair(core.Tokens(c))
	ui.Box(c).Padding(u2(c)*0.25, u2(c)).Radius(theme.PillRadius).
		Background(bg).Label("exit " + itoa(code)).Children(func() {
		ui.Text(c, "exit "+itoa(code)).TextColor(fg).Bold().
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// tailLines is the last n lines of s, with a note of how many were left out.
func tailLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if n <= 0 || len(lines) <= n {
		return s
	}
	hidden := len(lines) - n
	return core.Def("… ") + itoa(hidden) + core.Def(" earlier lines\n") + strings.Join(lines[hidden:], "\n")
}
