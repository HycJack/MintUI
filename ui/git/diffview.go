package git

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// DiffViewerOptions configure a DiffViewer.
type DiffViewerOptions struct {
	// Height is the height of the viewport. It is required: a diff has no
	// natural height, and one with none grows to fit every line of it, which
	// for a generated file is a window that cannot be closed.
	Height float32
	// Width is the width of the viewport; zero measures the window.
	Width float32
	// File names the diff, shown above the lines when the caller knows it.
	// The diff's own "diff --git" line is not a header: a viewer that shows
	// the raw patch shows the user the plumbing.
	File string
	// OldName and NewName are the two paths a rename moved between, shown
	// in the file header when both are given.
	OldName, NewName string
	// WithHunks draws each @@ line as a separator between the hunks, which
	// is what makes a diff of two distant parts of a file readable.
	WithHunks bool
	// Collapsed is how many unchanged lines to show either side of a change
	// before skipping to the next hunk. Zero shows everything, which is what
	// a small diff wants.
	Collapsed int
	// State is where the viewport keeps its place between frames.
	State *ui.ScrollState
}

// DiffViewer is a diff, drawn one line at a time from the lines SplitDiff
// cut out of it.
//
// The component decides nothing about what a diff says: the operations, the
// line numbers and the counts were all decided by the pure functions in this
// package before a single element was made. What is left here is two gutters,
// a tint behind the lines that changed, and a scrollbar.
//
// The tint is the tint rather than a coloured letter because a line of a diff
// is not read one character at a time. Tinting the whole row is what lets a
// person see the shape of a change — where it starts, how far it runs —
// before reading any of it.
func DiffViewer(c *ui.Context, unified string, opts DiffViewerOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Height <= 0 {
		panic("git: DiffViewer needs a Height; a diff with no height is every line of it")
	}
	lines := SplitDiff(unified)
	width := gutter(c, lines)
	hunks := Hunks(unified)
	// The hunk header belongs to the line that starts its hunk, which is
	// how a viewer with the headers inline draws them without a second walk
	// of the diff.
	headers := map[int]string{}
	if opts.WithHunks {
		n := 0
		for _, h := range hunks {
			headers[n] = h.Header
			n += len(h.Lines)
		}
	}

	view := layout.ScrollArea(c, layout.ScrollAreaOptions{
		Vertical: true, Horizontal: true, Height: opts.Height,
		Pad: u, State: opts.State,
	}, func() {
		if opts.File != "" || opts.OldName != "" {
			diffHeader(c, opts)
		}
		if len(lines) == 0 {
			ui.Text(c, noRows(c, "git.noDiff", core.Def("No changes to show"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			return
		}
		for i, l := range lines {
			if head, ok := headers[i]; ok {
				hunkSeparator(c, head)
			}
			diffRow(c, l, width)
		}
		if last := len(lines) - 1; last >= 0 {
			if _, ok := headers[last+1]; ok {
				hunkSeparator(c, headers[last+1])
			}
		}
	})
	return view.Element
}

// diffHeader is the one line above the lines that says which files these are.
// A rename shows both, because showing only the new one loses the fact that
// anything was renamed at all.
func diffHeader(c *ui.Context, opts DiffViewerOptions) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	name := opts.File
	if name == "" {
		name = opts.NewName
	}
	ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Center).Padding(u*0.5, 0).Children(func() {
		ui.Text(c, name).FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
		if opts.OldName != "" && opts.NewName != "" && opts.OldName != opts.NewName {
			ui.Text(c, "← "+opts.OldName).TextColor(k.TextMuted).Grow(1).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
	})
	layout.Divider(c, layout.DividerOptions{})
}

// hunkSeparator is the @@ line, kept as it was written because the function
// name git puts after the second @@ is the fastest way to know which function
// a change is in.
func hunkSeparator(c *ui.Context, head string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Box(c).FillWidth().Padding(u*0.5, 0).Children(func() {
		ui.Text(c, strings.TrimSpace(head)).TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
}

// diffRow is one line of a diff: the old number, the new number, and the
// text, tinted by what happened to it.
func diffRow(c *ui.Context, l DiffLine, gutterWidth float32) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	w := digits(max(l.OldNo, l.NewNo))

	bg := k.Background
	ink := k.Text
	switch l.Op {
	case OpAdd:
		bg, ink = k.SuccessBg, k.Success
	case OpDel:
		bg, ink = k.DangerBg, k.Danger
	}

	row := ui.Row(c).FillWidth().Shrink(0).Background(bg).Children(func() {
		no := ui.Text(c, padNo(l.OldNo, w)).TextColor(k.TextFaint)
		no.Width(gutterWidth).TextAlign(ui.End).FontSize(core.FontSize(c, theme.CaptionSize))
		// The two gutters need a step between them: two right-aligned
		// numbers back to back read as one long number — "4143", not
		// "41" and "43".
		ui.Box(c).Width(u)
		no2 := ui.Text(c, padNo(l.NewNo, w)).TextColor(k.TextFaint)
		no2.Width(gutterWidth).TextAlign(ui.End).FontSize(core.FontSize(c, theme.CaptionSize))
		ui.Box(c).Width(u * 0.5)
		ui.Text(c, l.Text).TextColor(ink).Grow(1).
			FontSize(core.FontSize(c, theme.RowSize))
	})
	row.Label(diffLineLabel(l))
}

// diffLineLabel is a diff line read out loud: what happened to it, where, and
// what it says. A screen reader that reads "plus, forty two, return" cannot
// review a diff; one that reads "added at line 42, return value" can.
func diffLineLabel(l DiffLine) string {
	switch l.Op {
	case OpAdd:
		return "added at line " + itoa(l.NewNo) + ", " + l.Text
	case OpDel:
		return "removed from line " + itoa(l.OldNo) + ", " + l.Text
	}
	return "line " + itoa(l.NewNo) + ", " + l.Text
}

// DiffStatOptions configure a DiffStat.
type DiffStatOptions struct {
	// Width is the width the bar is drawn in. Zero fills the row it is in.
	Width float32
	// Height is the height of the bar; zero gives the library's own.
	Height float32
	// Slots is how many units the bar is divided into. Zero gives eight,
	// which is as many as stay countable at a glance.
	Slots int
	// WithCounts puts "+4 -1" beside the bar rather than leaving the bar to
	// be read on its own.
	WithCounts bool
}

// DiffStat is the bar a repository puts beside every file it has changed:
// one run of added lines and one of deleted ones, in the same weights the
// status colours use.
//
// The bar is the summary and the two figures are the detail, and the split
// between them is made by a function so that the bar beside a file in a
// list and the numbers in a summary cannot round the same change two ways
// into two different pictures.
func DiffStat(c *ui.Context, added, deleted int, opts DiffStatOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	slots := opts.Slots
	if slots <= 0 {
		slots = 8
	}
	h := opts.Height
	if h <= 0 {
		h = u * 1.75
	}
	adds, dels := StatSlices(added, deleted, slots)

	bar := ui.Box(c).FillWidth().Height(h).Shrink(0).Role(ui.RoleNone).
		Label(StatWord(added, deleted) + ", " + itoa(adds) + " of " + itoa(slots) + " added").
		Draw(func(p *ui.Painter, r ui.Rect) {
			step := r.W / float32(slots)
			for i := range adds {
				internal.Dot(p, r.X+float32(i)*step+step/2, r.Y+r.H/2, h/2, k.Success)
			}
			for i := range dels {
				internal.Dot(p, r.X+float32(adds+i)*step+step/2, r.Y+r.H/2, h/2, k.Danger)
			}
		})
	if opts.Width > 0 {
		bar.Width(opts.Width).FillWidth()
	}

	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5)
	row.Children(func() {
		if opts.WithCounts {
			ui.Text(c, StatWord(added, deleted)).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
		bar.Grow(1)
	})
	return row
}

// StatSlices is how a change is split across the slots of a stat bar: adds
// first, then deletes, the whole divided in proportion and never adding to
// more slots than there are.
//
// The rule is proportional with the leftovers going to the additions, because
// a bar that has rounded its additions down is the one that understates the
// thing people care about. A diff of one line in a thousand is a single dot
// on either side, which is honest: it is one line.
//
// A diff with nothing in it takes no slots at all rather than all of them,
// which is the one case where "proportional" would say otherwise.
func StatSlices(added, deleted, slots int) (adds, dels int) {
	if slots <= 0 || added+deleted <= 0 {
		return 0, 0
	}
	// A change on one side only fills the bar. That is not a rounding: a
	// file with fifty deletions and none added has half its history gone,
	// and a bar showing a quarter of itself would say the opposite.
	switch {
	case added == 0:
		return 0, slots
	case deleted == 0:
		return slots, 0
	}
	total := added + deleted
	dels = deleted * slots / total
	adds = slots - dels
	// Both sides get at least one slot. Without it a hundred additions and
	// one deletion draw a full green bar and no red at all, which is the
	// one shape a stat bar must never be: the deletion is what the reader
	// is looking for.
	if dels == 0 {
		dels, adds = 1, slots-1
	}
	if adds == 0 {
		adds, dels = 1, slots-1
	}
	return adds, dels
}

// InlineDiffOptions configure an InlineDiff.
type InlineDiffOptions struct {
	// Width is the width of the line; zero lets it fill the row.
	Width float32
	// WithLineNumbers draws the two numbers beside the text, for an inline
	// diff that has to be pointed at in a review.
	WithLineNumbers bool
	// MaxLines caps how much of each side is shown; zero is unlimited.
	MaxLines int
}

// InlineDiff is two versions of one line, side by side: what was taken out in
// the danger tint, what went in the success tint, word by word.
//
// The word-level split is computed by InlineSplit rather than drawn by
// comparison, because "changed" and "inserted" are different things and only
// one of them can be underlined. A whole line painted as removed when one
// word of it changed is the reason inline diffs got a reputation for being
// noise.
//
// It is for one line or one short paragraph. Anything longer wants
// DiffViewer, which has gutters, hunks and a scrollbar.
func InlineDiff(c *ui.Context, before, after string, opts InlineDiffOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	removed, added := InlineSplit(before, after)

	row := ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Stretch).Shrink(0)
	row.Children(func() {
		side := func(words []InlineWord, tint, bg ui.Color, name string) {
			// A wrapping row, not a column: this component is two versions
			// of ONE line, and a column of its words draws a ladder that
			// reads as a paragraph per side rather than a diff. The gap is
			// a space's width because each word carries its trailing
			// whitespace in the string but not in the box it is measured
			// in, so without it the line reads as one word.
			col := ui.Row(c).Grow(1).Wrap().AlignItems(ui.Start).Gap(u*0.9).
				Padding(u*0.5, u).
				Radius(theme.SmallRadius).Background(bg).Role(ui.RoleNone).Label(name)
			col.Children(func() {
				if opts.MaxLines > 0 && len(words) > opts.MaxLines {
					words = words[:opts.MaxLines]
				}
				for _, word := range words {
					t := ui.Text(c, word.Text).TextColor(tint).FontSize(core.FontSize(c, theme.RowSize))
					if word.Changed {
						// The word that moved is underlined rather than
						// filled: filling the changed word makes the two
						// sides the same block of colour, which is the thing
						// this component exists to avoid.
						t.Underline().DecorationColor(tint)
					}
					if opts.WithLineNumbers {
						no := ui.Text(c, padNo(word.No, 3)).TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.CaptionSize))
						no.Width(u * 3).TextAlign(ui.End)
					}
				}
			})
			if opts.Width > 0 {
				col.Width(opts.Width / 2).Shrink(0)
			}
		}
		side(removed, k.Danger, k.DangerBg, "before: "+before)
		side(added, k.Success, k.SuccessBg, "after: "+after)
	})
	return row
}

// InlineWord is one word of one side of an inline diff, and whether it is the
// part that changed.
type InlineWord struct {
	// Text is the word with the whitespace that follows it, so that joining
	// every word of a side gives the side back exactly.
	Text string
	// No is the word's 1-based number in that side.
	No int
	// Changed reports that this word is in the part that differs. It is the
	// distinction the whole component is about.
	Changed bool
}

// InlineSplit is the word-level difference between two strings: the words of
// the old one that are not in the new, and the words of the new one that are
// not in the old, each keeping its own numbering.
//
// It is a longest common subsequence over words, which is what makes a
// changed word one changed word rather than the whole rest of the line. That
// matters: "the quick brown fox" to "the slow brown fox" has one changed
// word, and any cheaper comparison would report the other two as well.
//
// Words are split on whitespace with the whitespace kept, so the two sides
// can be put back together exactly as they were — a diff that silently drops
// a space is a diff of a different string.
func InlineSplit(before, after string) (removed, added []InlineWord) {
	a, b := tokenize(before), tokenize(after)

	// The LCS table is (len(a)+1) by (len(b)+1), filled from the bottom up:
	// the length of the common run ending at a[i] and b[j].
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i].trimmed == b[j].trimmed {
				lcs[i][j] = lcs[i+1][j+1] + 1
				continue
			}
			lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
		}
	}

	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i].trimmed == b[j].trimmed:
			removed = append(removed, InlineWord{Text: a[i].text, No: i + 1})
			added = append(added, InlineWord{Text: b[j].text, No: j + 1})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			removed = append(removed, InlineWord{Text: a[i].text, No: i + 1, Changed: true})
			i++
		default:
			added = append(added, InlineWord{Text: b[j].text, No: j + 1, Changed: true})
			j++
		}
	}
	for ; i < len(a); i++ {
		removed = append(removed, InlineWord{Text: a[i].text, No: i + 1, Changed: true})
	}
	for ; j < len(b); j++ {
		added = append(added, InlineWord{Text: b[j].text, No: j + 1, Changed: true})
	}
	return removed, added
}

// token is one word with its trailing whitespace, and the word without it.
type token struct {
	text    string
	trimmed string
}

// tokenize splits s into words, keeping the whitespace that follows each one.
func tokenize(s string) []token {
	var out []token
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			continue
		}
		out = append(out, token{text: s[start : i+1], trimmed: strings.TrimSpace(s[start : i+1])})
		start = i + 1
	}
	if start < len(s) {
		out = append(out, token{text: s[start:], trimmed: strings.TrimSpace(s[start:])})
	}
	return out
}
