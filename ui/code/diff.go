package code

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Diff shows the difference between two texts, either as one column with the
// changes marked or as two side by side.
//
// It computes the difference itself rather than taking a list of changes,
// because a caller with two files does not have a diff and a caller with a
// diff has one that is in a different shape from every other diff tool's. The
// lines in and the lines out are the two things a caller has.

// DiffLineKind is what a line of a diff is.
type DiffLineKind int

const (
	// DiffSame is a line in both.
	DiffSame DiffLineKind = iota
	// DiffAdded is only in the new one.
	DiffAdded
	// DiffRemoved is only in the old one.
	DiffRemoved
)

func (k DiffLineKind) String() string {
	switch k {
	case DiffAdded:
		return "Added"
	case DiffRemoved:
		return "Removed"
	}
	return "Same"
}

// DiffLine is one line of a diff.
type DiffLine struct {
	// Kind is what happened to it.
	Kind DiffLineKind
	// Old and New are the line's numbers on either side, counted from one.
	// Zero on a side means the line is not on that side, which is what draws
	// the gutter's gap rather than a wrong number.
	Old, New int
	// Text is the line itself, without its newline.
	Text string
}

// DiffOptions configure a Diff.
type DiffOptions struct {
	// Old and New are the two texts.
	Old, New string
	// Name is what the diff is called — a path, a pair of revisions. It is
	// required, for the reason a viewer's is.
	Name string
	// Lang is which language's words the highlighter knows, so that a diff of
	// code is coloured like the code rather than like the differences.
	Lang Lang
	// Split puts the two versions side by side rather than in one column.
	// Off by default: one column is the one that can be read on a narrow
	// window and that a whole-file diff fits in at all.
	Split bool
	// Context is how many unchanged lines are shown either side of a change.
	// Zero is three, which is enough to see what the change is in.
	Context int
	// First and Second are the numbers the two sides' first lines are given,
	// for a diff of fragments of two larger files.
	First, Second int
	// Height is the viewport's own height, and is required for the reason a
	// viewer's is.
	Height float32
	// Width is the diff's own width.
	Width float32
	// Gutter draws the line numbers of both sides.
	Gutter *bool
	// State and Scroll are where the diff is, and the caller's to keep.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Caption says what is being compared.
	Caption string
}

// DiffResult carries a Diff and what was done in it.
type DiffResult struct {
	// Element is the diff.
	Element *ui.Element
	// Lines is what the diff decided, which is what a caller needs for a
	// status bar and for a stat line: "12 added, 3 removed" is a question
	// about the difference and not about the drawing.
	Lines []DiffLine
	// added and removed are the counts.
	added, removed int
	// selected is the row pressed this frame, or -1.
	selected int
}

// Added is how many lines are only in the new one.
func (r DiffResult) Added() int { return r.added }

// Removed is how many lines are only in the old one.
func (r DiffResult) Removed() int { return r.removed }

// Selected is the row pressed this frame, counted from zero, and -1 for none.
// It is a row rather than a line because a side-by-side diff's rows are two
// lines and a caller opening something wants the one that is there.
func (r DiffResult) Selected() int { return r.selected }

// Diff shows two texts and what changed between them.
//
// The rows are built in a ui.List, so a diff of two files of three thousand
// lines draws the forty it has room for. The lines are computed once and
// kept, because computing a longest-common-subsequence over six thousand
// lines every frame would be the whole frame's cost for something that has
// not changed.
func Diff(c *ui.Context, opts DiffOptions) DiffResult {
	if opts.Name == "" {
		panic("code: Diff needs a Name; two columns of text with nothing saying what they are " +
			"is not a difference")
	}
	if opts.Height <= 0 {
		panic("code: Diff needs a Height; without one it draws every line of both files")
	}
	if opts.First <= 0 {
		opts.First = 1
	}
	if opts.Second <= 0 {
		opts.Second = 1
	}
	if opts.Context <= 0 {
		opts.Context = 3
	}
	gutter := true
	if opts.Gutter != nil {
		gutter = *opts.Gutter
	}

	lines := DiffLines(opts.Old, opts.New, opts.Context)
	var added, removed int
	for _, l := range lines {
		switch l.Kind {
		case DiffAdded:
			added++
		case DiffRemoved:
			removed++
		}
	}

	var r DiffResult
	r.Lines, r.added, r.removed, r.selected = lines, added, removed, -1

	// One column or two is one list and two lists, and which is which changes
	// only how a row is built — so the row builder takes the line or the pair
	// and the two callers differ in nothing else.
	// The body is a function, not an element, so that CodePanel can draw the
	// header above it rather than below it.
	body := func() {
		host := layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Horizontal: true, Height: opts.Height, State: opts.Scroll,
		}, nil).Element.FillWidth()
		if opts.Width > 0 {
			host.Width(opts.Width).Shrink(0)
		}
		host.Children(func() {
			if opts.Split {
				ui.List(c, opts.State, len(lines), func(i int) {
					if lines[i].Kind == DiffSame {
						splitPair(c, opts, lines, i, gutter, &r)
						return
					}
					splitSingle(c, opts, lines, i, gutter, &r)
				}).FillWidth().Height(opts.Height).Role(ui.RoleNone)
				return
			}
			ui.List(c, opts.State, len(lines), func(i int) {
				unifiedRow(c, opts, lines[i], gutter, &r)
			}).FillWidth().Height(opts.Height).Role(ui.RoleNone)
		})

	}
	host := CodePanel(c, body, layout.ContainerOptions{Width: opts.Width}, func() {
		CodeHeader(c, opts.Name, nil, func() {
			if opts.Caption != "" {
				CodeCaption(c, opts.Caption)
			}
			if added > 0 {
				CodeBadge(c, "+"+itoa(added), core.Success)
			}
			if removed > 0 {
				CodeBadge(c, "-"+itoa(removed), core.Danger)
			}
		})
	})

	r.Element = host
	return r
}

// unifiedRow is one row of a one-column diff: a number on the left, a number
// on the right, and the line between them.
//
// Both numbers are drawn even for a line that is only on one side, because a
// reader counting down a file wants the line they are on on both sides and a
// gap is not a number. The side a line is missing from is drawn blank rather
// than as a zero, and that is the same rule the numbers themselves follow.
func unifiedRow(c *ui.Context, opts DiffOptions, line DiffLine, gutter bool, r *DiffResult) {
	k := core.Tokens(c)
	rows := CodeRowsOptions{
		Lines: len(opts.Old) + 1, First: opts.First,
		CurrentBand: false, Gutter: false,
	}
	band, base := diffFace(k, line.Kind)
	number := line.Old
	if line.Kind == DiffAdded {
		number = line.New
	}
	// Everything goes in the content slot codeRow makes, and not in a second
	// Children call on the row itself: the row already holds one box that
	// grows to fill it, and a second set of children beside that one splits
	// the row in half — which is how a line of forty characters ends up drawn
	// twenty characters wide.
	row := codeRow(c, rows, number, func() {
		if line.Kind != DiffSame {
			ui.Box(c).Width(metricsOf(c, len(linesOf(opts.New))).gutter).
				Shrink(0).Background(band)
		}
		if gutter {
			diffNumbers(c, line.Old, line.New, gutterOf(c, opts), false)
		}
		mark := diffMark(c, line.Kind)
		if mark != nil {
			mark()
		}
		ui.Box(c).Grow(1).Shrink(0).Children(func() {
			codeTokens(c, Highlight(line.Text, opts.Lang, false), theme.RowSize, base, nil)
		})
	})
	if band.A != 0 {
		row.Background(band)
	}
	// The press target is absolute: an element made beside the row rather
	// than inside it is its sibling, and a sibling that grows takes half the
	// row's width away from the code it is meant to be a target for.
	hit := ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Role(ui.RoleNone)
	hit.Label(diffRowLabel(line))
	if hit.Clicked() {
		r.selected = line.New - 1
	}
}

// splitPair is one unchanged row of a side-by-side diff: the same line on both
// sides.
func splitPair(c *ui.Context, opts DiffOptions, lines []DiffLine, i int, gutter bool, r *DiffResult) {
	k := core.Tokens(c)
	rows := CodeRowsOptions{Lines: len(opts.Old) + 1, First: opts.First, Gutter: gutter}
	codeRow(c, rows, lines[i].Old, func() {
		diffSide(c, opts, lines[i], gutter)
		layout.Divider(c, layout.DividerOptions{Vertical: true})
		diffSide(c, opts, lines[i], gutter)
	})
	_ = k
	hit := ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Role(ui.RoleNone)
	hit.Label(diffRowLabel(lines[i]))
	if hit.Clicked() {
		r.selected = lines[i].New - 1
	}
}

// splitSingle is one changed row of a side-by-side diff: the removed line on
// the left, the added one on the right, and nothing blank between them.
//
// They are on one row rather than on two because a side-by-side diff is read
// by looking across: a change whose two halves are on different rows is a
// change the reader has to remember the top half of while reading the bottom.
func splitSingle(c *ui.Context, opts DiffOptions, lines []DiffLine, i int, gutter bool, r *DiffResult) {
	rows := CodeRowsOptions{Lines: len(opts.Old) + 1, First: opts.First, Gutter: gutter}
	codeRow(c, rows, lines[i].New, func() {
		diffSide(c, opts, lines[i], gutter)
		layout.Divider(c, layout.DividerOptions{Vertical: true})
		diffSide(c, opts, lines[i], gutter)
	})
	hit := ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Role(ui.RoleNone)
	hit.Label(diffRowLabel(lines[i]))
	if hit.Clicked() {
		r.selected = lines[i].New - 1
	}
}

// diffSide is one half of a side-by-side row: the line if it is on this side,
// and nothing if it is not.
//
// The side is a Row, not a Column: the number, the mark and the code share one
// line, and a column here stacks them instead — a row of code three rows tall,
// spilling out of its own line into the two below it.
func diffSide(c *ui.Context, opts DiffOptions, line DiffLine, gutter bool) {
	k := core.Tokens(c)
	band, base := diffFace(k, line.Kind)
	side := ui.Row(c).Grow(1).Shrink(0).AlignItems(ui.Center).Background(band)
	side.Children(func() {
		if gutter {
			// The number that belongs to this half, and nothing for the half
			// the line is not on: a gap reads as "not here", where a zero
			// reads as a line zero.
			switch line.Kind {
			case DiffRemoved:
				diffNumbers(c, line.Old, 0, gutterOf(c, opts), true)
			case DiffAdded:
				diffNumbers(c, 0, line.New, gutterOf(c, opts), true)
			default:
				diffNumbers(c, line.Old, line.New, gutterOf(c, opts), true)
			}
		}
		mark := diffMark(c, line.Kind)
		if mark != nil {
			mark()
		}
		ui.Box(c).Grow(1).Shrink(0).Children(func() {
			codeTokens(c, Highlight(line.Text, opts.Lang, false), theme.RowSize, base, nil)
		})
	})
}

// diffNumbers draws the two line numbers a diff row carries, each in its own
// half of the gutter, aligned to the ends.
//
// They are two numbers and not one because a diff's whole job is saying where
// a line came from and where it went, and a single number can only say one of
// those.
func diffNumbers(c *ui.Context, old, at int, width float32, right bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	e := ui.Row(c).Width(width).Shrink(0).AlignItems(ui.Center).Role(ui.RoleNone)
	if right {
		e.Justify(ui.End)
	}
	e.Children(func() {
		if old == 0 && at == 0 {
			// Neither side has this line: the gap, not a zero. A zero beside a
			// line number is a line zero, and no file has one.
			ui.Box(c).Grow(1)
			return
		}
		text := itoa(maxInt(old, at))
		ui.Text(c, text).Font(MonoStack).SingleLine().Shrink(0).
			FontSize(core.FontSize(c, theme.CaptionSize)).
			TextColor(k.TextFaint).Padding(0, u, 0, u)
	})
}

// diffMark is the + or − in the margin of a changed line.
//
// It is in the text and not in the colour because a colour is the one thing
// somebody cannot see, and a diff is the one thing in a window where that
// matters most: a diff nobody can see is not a diff.
func diffMark(c *ui.Context, kind DiffLineKind) func() {
	k, u := core.Tokens(c), core.Density(c).Unit()
	switch kind {
	case DiffAdded:
		return func() {
			ui.Text(c, "+").Font(MonoStack).SingleLine().Shrink(0).
				FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Success)
		}
	case DiffRemoved:
		return func() {
			ui.Text(c, "−").Font(MonoStack).SingleLine().Shrink(0).
				FontSize(core.FontSize(c, theme.RowSize)).TextColor(k.Danger)
		}
	}
	_ = u
	return nil
}

// diffRowLabel names a diff row for assistive technology and for a test.
func diffRowLabel(l DiffLine) string {
	switch l.Kind {
	case DiffAdded:
		return "Added line " + itoa(l.New)
	case DiffRemoved:
		return "Removed line " + itoa(l.Old)
	}
	return "Line " + itoa(l.New)
}

// diffFace is the band a line of a diff is drawn on, and the ink its text
// takes on it.
//
// A changed line is tinted rather than filled with its severity's pair,
// because a whole row of a strong colour behind a row of text is a wall, and
// a wall of red in the middle of a file is a thing a reader learns to skip
// past. The mark in the margin is what says which way the line went.
func diffFace(k theme.Tokens, kind DiffLineKind) (band, base ui.Color) {
	switch kind {
	case DiffAdded:
		return k.SuccessBg, k.Success
	case DiffRemoved:
		return k.DangerBg, k.Danger
	}
	return ui.Color{}, k.Text
}

// linesOf is the lines of a text, for the one place that needs the count.
func linesOf(s string) []string { return splitLines(s) }

func gutterOf(c *ui.Context, opts DiffOptions) float32 {
	n := len(linesOf(opts.Old))
	if m := len(linesOf(opts.New)); m > n {
		n = m
	}
	return metricsOf(c, n).gutter
}

// DiffLines is the difference between two texts, as the rows a Diff shows.
//
// It is exported because a caller needs it without the drawing: a status bar
// saying "12 added, 3 removed", a commit message, a file picker showing which
// files changed and by how much. All three want the same answer and none of
// them wants a viewer.
//
// The algorithm is a longest common subsequence over lines, which is the
// definition of a line diff and costs O(n·m) — a few hundred thousand cells
// for two files of a thousand lines, and nothing at all for two files of
// forty. It is not a heuristic and it is not Myers' algorithm; it is the one
// that is obviously right, which for a diff shown to a person and not measured
// against a clock is the better half of the trade.
func DiffLines(old, new string, context int) []DiffLine {
	a, b := splitLines(old), splitLines(new)
	if context <= 0 {
		context = 3
	}

	// The common subsequence, filled in from the bottom right. Every cell is
	// the length of the longest common run ending there.
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
				continue
			}
			if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	var out []DiffLine
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			out = append(out, DiffLine{Kind: DiffSame, Old: i + 1, New: j + 1, Text: a[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, DiffLine{Kind: DiffRemoved, Old: i + 1, Text: a[i]})
			i++
		default:
			out = append(out, DiffLine{Kind: DiffAdded, New: j + 1, Text: b[j]})
			j++
		}
	}
	for ; i < len(a); i++ {
		out = append(out, DiffLine{Kind: DiffRemoved, Old: i + 1, Text: a[i]})
	}
	for ; j < len(b); j++ {
		out = append(out, DiffLine{Kind: DiffAdded, New: j + 1, Text: b[j]})
	}
	return withContext(out, context)
}

// withContext is the diff with the unchanged runs away from every change
// replaced by one marker line, which is what makes a diff of two files of a
// thousand lines readable in a window a hundred high.
//
// The marker says how many lines are hidden rather than nothing at all: "⋯
// 12 lines" tells a reader that the gap is a gap and how big it is, which a
// bare ellipsis does not.
func withContext(all []DiffLine, context int) []DiffLine {
	keep := make([]bool, len(all))
	for i, l := range all {
		if l.Kind == DiffSame {
			continue
		}
		from, to := maxInt(i-context, 0), minInt(i+context+1, len(all))
		for j := from; j < to; j++ {
			keep[j] = true
		}
	}
	out := make([]DiffLine, 0, len(all))
	hidden := 0
	flush := func() {
		if hidden == 0 {
			return
		}
		out = append(out, DiffLine{Kind: DiffSame, Text: gapText(hidden)})
		hidden = 0
	}
	for i, l := range all {
		if keep[i] {
			flush()
			out = append(out, l)
			continue
		}
		hidden++
	}
	flush()
	return out
}

func gapText(n int) string {
	if n == 1 {
		return "⋯ 1 line"
	}
	return "⋯ " + itoa(n) + " lines"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
