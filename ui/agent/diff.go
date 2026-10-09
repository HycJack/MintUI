package agent

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/code"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// FileStatus is what a run did to a file. It is one type for the paths a run
// wrote as well as the ones it touched, because a review reads them all
// together and "Modified" next to "Added" next to "Deleted" is one column.
type FileStatus int

const (
	// FileModified is a file both versions still have, with different
	// contents. It is the zero value so that a path with no status attached
	// is treated as changed rather than as created.
	FileModified FileStatus = iota
	// FileAdded is a file only one version has.
	FileAdded
	// FileRemoved is a file the other version has and this one has deleted.
	FileRemoved
	// FileRenamed is a file that moved and did not change.
	FileRenamed
)

func (f FileStatus) String() string {
	switch f {
	case FileAdded:
		return "Added"
	case FileRemoved:
		return "Removed"
	case FileRenamed:
		return "Renamed"
	case FileModified:
		return "Modified"
	}
	panic("agent: unknown FileStatus " + itoa(int(f)))
}

// FileTone is the severity a file's status is marked with.
//
// Renamed is Accent rather than a status of its own: a move is the one change
// that is neither the file appearing nor the file going, and wearing the
// system's highlight says "look here" without claiming anything went wrong.
func FileTone(f FileStatus) core.Severity {
	switch f {
	case FileAdded:
		return core.Success
	case FileRemoved:
		return core.Danger
	case FileRenamed:
		return core.Accent
	case FileModified:
		return core.Neutral
	}
	panic("agent: unknown FileStatus " + itoa(int(f)))
}

// splitLines is a text as its lines, without the newline that ended the last
// one. An empty text is no lines rather than one empty line, because a file
// that has never been written is not a file whose single line is blank.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// ── the three-way merge ─────────────────────────────────────────────────────

// sideEdits is what one branch did to the lines of the base: which it left,
// which it rewrote, which it deleted, and what it put in between them.
type sideEdits struct {
	// repl[i] is the whole of what the branch has in place of base[i]: the
	// lines a rewrite turned one line into, or nil when it left the line
	// alone. A slice rather than a single string because a rewrite is
	// routinely more than one line — gofmt folding a call across three is the
	// commonest edit in a Go repository — and a merge that can only hold one
	// line per base line turns every such edit into a conflict.
	repl [][]string
	// gone[i] says the branch deleted base[i]. It is separate from repl being
	// nil because "left it alone" and "took it out" are different answers and
	// the merge treats them differently.
	gone []bool
	// after[i] is what the branch put in before base[i], and after[n] is what
	// it left at the end. Indexing by the line a thing comes *before* rather
	// than after is what makes the merge walk a single forward pass.
	after [][]string
}

// mergeCells is how large a longest-common-subsequence table may get before
// the merge stops trying to be minimal.
//
// Two thousand by two thousand lines is four million cells of int32 — sixteen
// megabytes, once per side, on every frame that recomputes a merge. Past that
// the alignment falls back to pairing lines by position, which is wrong on a
// file with a big move in it and instant on a generated one. The fallback is
// the honest choice: a merge that is slightly wrong is fixable by a person
// looking at it, and one that hangs the window is not.
const mergeCells = 4 << 20

// op is one move in the edit script that editSide builds: a line that stayed,
// a line that went, or a line that arrived. The script is the raw form; the
// pairing of a deletion with an insertion that follows it is what turns "this
// went and that arrived" into "this was rewritten as that".
type op struct {
	kind byte // ' ' stayed, '-' went, '+' arrived
	base int  // the base line, for '-' and ' '
	text string
	pos  int // the base line an arrival sits before, for '+'
}

// editSide is what one branch did to the base.
func editSide(base, side []string) sideEdits {
	e := sideEdits{
		repl:  make([][]string, len(base)),
		gone:  make([]bool, len(base)),
		after: make([][]string, len(base)+1),
	}
	n, m := len(base), len(side)
	switch {
	case n == 0:
		e.after[0] = append(e.after[0], side...)
		return e
	case m == 0:
		for i := range base {
			e.gone[i] = true
		}
		return e
	case (n+1)*(m+1) > mergeCells:
		editSideByPosition(e, base, side)
		return e
	}

	// lcs[i][j] is the length of the longest common subsequence of base[i:]
	// and side[j:]. The walk below is forward and asks one question at every
	// mismatch: is the longer common tail reached by leaving out base[i], or
	// by leaving out side[j]? Either answer is a correct alignment — an LCS
	// is ambiguous wherever two of them are the same length — but answering
	// "leave out the base line" on a tie is what keeps a one-line edit as
	// one removal beside one addition instead of the whole file rewritten.
	lcs := make([][]int32, n+1)
	for i := range lcs {
		lcs[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if base[i] == side[j] {
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

	script := make([]op, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case base[i] == side[j]:
			script = append(script, op{kind: ' ', base: i})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			script = append(script, op{kind: '-', base: i})
			i++
		default:
			script = append(script, op{kind: '+', text: side[j], pos: i})
			j++
		}
	}
	for i < n {
		script = append(script, op{kind: '-', base: i})
		i++
	}
	for j < m {
		script = append(script, op{kind: '+', text: side[j], pos: n})
		j++
	}

	// The pairing pass. A deletion followed by a run of arrivals in the same
	// place is one rewrite — one base line became those lines — and it is the
	// only thing that turns a script of separate removals and arrivals into
	// the changes a person made. Arrivals that got paired are marked so they
	// are not also counted as insertions.
	//
	// "A deletion followed by arrivals" rather than "a deletion followed by
	// exactly one arrival" is the difference between a merge that understands
	// gofmt and one that reports every reformat of a line as a conflict with
	// the other branch's reformat of it.
	paired := make([]bool, len(script))
	for n := 0; n < len(script); n++ {
		if script[n].kind != '-' {
			continue
		}
		b := script[n].base
		j := n + 1
		var repl []string
		for j < len(script) && script[j].kind == '+' && script[j].pos == b+1 {
			repl = append(repl, script[j].text)
			paired[j] = true
			j++
		}
		if len(repl) == 0 {
			e.gone[b] = true
			continue
		}
		e.repl[b] = repl
		n = j - 1
	}
	for n, o := range script {
		if o.kind == '+' && !paired[n] {
			e.after[o.pos] = append(e.after[o.pos], o.text)
		}
	}
	return e
}

// sameLines reports whether two replacements are the same lines, which is the
// question that decides convergence: two branches that arrived at the same
// text — down to every line of a multi-line rewrite — have not disagreed.
func sameLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// editSideByPosition pairs the two texts line for line. It is the fallback
// for a pair too large to align properly, and it is deterministic, which is
// the only property a merge is allowed to keep at that size.
func editSideByPosition(e sideEdits, base, side []string) {
	for i := range base {
		switch {
		case i < len(side) && base[i] == side[i]:
		case i < len(side):
			e.repl[i] = []string{side[i]}
		default:
			e.gone[i] = true
		}
	}
	if len(side) > len(base) {
		e.after[len(base)] = append(e.after[len(base)], side[len(base):]...)
	}
}

// Merge3 folds theirs into ours given the base the two branches share, and
// returns the result as a diff against that base.
//
// The return is code.DiffLine rather than a merged string, because the thing
// a reviewer needs is not the merged file — they have that, it is on disk —
// it is what the merge decided about each line of the base: which lines
// survived untouched, which one branch rewrote, which both rewrote the same
// way, and which one neither branch would concede.
//
// That last case is the whole reason this function exists. When both branches
// rewrite one base line to different text there is no correct answer, so
// rather than silently preferring one side the merge emits the base line as
// removed and both candidates as added, ours first. Nothing is lost, the
// order says which way a person should lean if they have to, and the shape
// is recognisable: a removal followed by two additions is a conflict, and
// Conflicts finds them again without a second walk.
//
// A deletion is not a conflict. One branch deleting a line the other only
// touched is resolved in favour of the deletion, because the alternative —
// keeping the line with the other branch's text on it — puts back a line
// somebody deliberately took out.
//
// The line numbers are the diff's own: Old is where the line was in the base
// and New is where it ended up in the merge, zero on the side it is not on.
func Merge3(base, ours, theirs string) []code.DiffLine {
	merged, _ := merge3(base, ours, theirs)
	return merged
}

// merge3 is Merge3 and the conflicts it found, in one walk — the two are the
// same walk, and running it twice to get the second answer would double the
// cost of every review pane on screen for nothing.
func merge3(base, ours, theirs string) ([]code.DiffLine, []Conflict) {
	b := splitLines(base)
	o := editSide(b, splitLines(ours))
	t := editSide(b, splitLines(theirs))

	out := make([]code.DiffLine, 0, len(b)+8)
	var found []Conflict
	newNo := 0

	add := func(text string) {
		newNo++
		out = append(out, code.DiffLine{Kind: code.DiffAdded, New: newNo, Text: text})
	}

	for i := range b {
		// Insertions first: a line a branch put in before this one comes
		// before it, whichever branch it came from. Ours go first so that a
		// reader scanning an insertion sees their own side before the other.
		for _, text := range o.after[i] {
			add(text)
		}
		for _, text := range t.after[i] {
			add(text)
		}

		oldNo := i + 1
		oursGone, theirsGone := o.gone[i], t.gone[i]

		// Both branches deleted it: the merged file does not have it, and a
		// diff from the base to the merge has to say so. Silently dropping it
		// would make the merge of a deleted file look like a merge that
		// changed nothing.
		if oursGone && theirsGone {
			out = append(out, code.DiffLine{Kind: code.DiffRemoved, Old: oldNo, Text: b[i]})
			continue
		}
		if oursGone || theirsGone {
			// One deleted and the other rewrote. The delete stands; the
			// rewrite is dropped along with it, which is the only resolution
			// that does not put back a line a branch took out.
			out = append(out, code.DiffLine{Kind: code.DiffRemoved, Old: oldNo, Text: b[i]})
			continue
		}

		ours, theirs := o.repl[i], t.repl[i]
		switch {
		case sameLines(ours, theirs):
			// Both branches left it, or both made the same edit. There is
			// nothing to choose between them either way.
			if ours == nil {
				newNo++
				out = append(out, code.DiffLine{
					Kind: code.DiffSame, Old: oldNo, New: newNo, Text: b[i],
				})
				continue
			}
			out = append(out, code.DiffLine{Kind: code.DiffRemoved, Old: oldNo, Text: b[i]})
			for _, text := range ours {
				add(text)
			}
		case ours != nil && theirs != nil:
			// The conflict: the base line goes, both sides stay, ours first.
			found = append(found, Conflict{At: len(out), Base: b[i], Ours: ours, Theirs: theirs})
			out = append(out, code.DiffLine{Kind: code.DiffRemoved, Old: oldNo, Text: b[i]})
			for _, text := range ours {
				add(text)
			}
			for _, text := range theirs {
				add(text)
			}
		case ours != nil:
			out = append(out, code.DiffLine{Kind: code.DiffRemoved, Old: oldNo, Text: b[i]})
			for _, text := range ours {
				add(text)
			}
		default:
			out = append(out, code.DiffLine{Kind: code.DiffRemoved, Old: oldNo, Text: b[i]})
			for _, text := range theirs {
				add(text)
			}
		}
	}
	for _, text := range o.after[len(b)] {
		add(text)
	}
	for _, text := range t.after[len(b)] {
		add(text)
	}
	return out, found
}

// Conflict is one line of the base that both branches rewrote to different
// text, with both of the rewrites.
// Conflict is one hunk a merge could not merge.
type Conflict struct {
	// At is where the conflict starts in the line list Merge3 returned: the
	// removed base line, then ours, then theirs.
	At int
	// Base is what the two branches started from, Ours is the whole of what
	// this branch made of it and Theirs is the whole of what the other did.
	//
	// Rewrites are slices rather than strings because a rewrite is routinely
	// more than one line, and a merge that reported only the first of them
	// would be hiding the edit it was asked to show.
	Base   string
	Ours   []string
	Theirs []string
}

// Lines is how many lines of the merged diff this conflict occupies: the base
// line, then both sides. It is what a caller needs to draw a band across the
// whole conflict rather than a rule beside its first line.
func (cf Conflict) Lines() int { return 1 + len(cf.Ours) + len(cf.Theirs) }

// Conflicts are the hunks a merge could not merge, found by merging again.
//
// It takes the three texts rather than the merged line list on purpose. A
// conflict in the line list looks like a removal followed by additions, and
// where one side's additions stop and the other's begin cannot be read back
// out of it: both sides may have rewritten the base line into several lines,
// and the diff says only that they are there. So the question is answered
// from the three texts it can actually be answered from, and it costs the
// same walk Merge3 already makes for the drawing.
func Conflicts(base, ours, theirs string) []Conflict {
	_, found := merge3(base, ours, theirs)
	return found
}

// ── one file's change ────────────────────────────────────────────────────────

// AgentDiffResult carries an AgentDiff and what was pressed in it.
type AgentDiffResult struct {
	// Element is the diff.
	Element *ui.Element
	// Lines is what the merge decided, for a caller that wants the counts
	// without reading them off the drawing.
	Lines                    []code.DiffLine
	added, removed, selected int
}

// Added is how many lines the change introduced.
func (r AgentDiffResult) Added() int { return r.added }

// Removed is how many lines it took away.
func (r AgentDiffResult) Removed() int { return r.removed }

// Selected is the row pressed this frame, counted from zero, and -1 for none.
func (r AgentDiffResult) Selected() int { return r.selected }

// AgentDiffOptions configure an AgentDiff.
type AgentDiffOptions struct {
	// Path is the file the change is in. It is required: a diff with nothing
	// saying which file it is a diff of nothing.
	Path string
	// Change is what the run did to the file. It drives the chip beside the
	// path; the diff itself is a diff either way.
	Change FileStatus
	// Old and New are the file's two versions.
	Old, New string
	// Lang is which language's words the highlighter knows.
	Lang code.Lang
	// Caption says what the change was for, under the path.
	Caption string
	// Context is how many unchanged lines show either side of a change.
	Context int
	// Height is the viewport's height, and is required for the reason a
	// scroll area's is: without one the whole file is drawn.
	Height float32
	// Scroll is where the diff is scrolled to, and the caller's to keep.
	Scroll *ui.ScrollState
	// State is the list's place, and the caller's to keep.
	State *ui.ListState
}

// AgentDiff is one file's change: its path, what happened to it, and the
// lines.
//
// It is code.Diff with an agent's header on top, which is the whole
// difference: a reader of a run's changes is scanning a list of files first
// and a file second, so the path and the count have to be readable before any
// of the code is. Handing them a bare diff would mean every caller re-drawing
// the same two lines.
func AgentDiff(c *ui.Context, opts AgentDiffOptions) AgentDiffResult {
	if opts.Path == "" {
		panic("agent: AgentDiff needs a Path; a diff of nothing says nothing")
	}
	if opts.Height <= 0 {
		panic("agent: AgentDiff needs a Height; without one it draws every line of the file")
	}
	lines := code.DiffLines(opts.Old, opts.New, opts.Context)
	var added, removed int
	for _, l := range lines {
		switch l.Kind {
		case code.DiffAdded:
			added++
		case code.DiffRemoved:
			removed++
		}
	}

	d := code.Diff(c, code.DiffOptions{
		Old: opts.Old, New: opts.New, Name: opts.Path, Lang: opts.Lang,
		Height: opts.Height, Context: opts.Context,
		State: opts.State, Scroll: opts.Scroll, Caption: opts.Caption,
	})
	return AgentDiffResult{
		Element: d.Element, Lines: lines,
		added: added, removed: removed, selected: d.Selected(),
	}
}

// ── the review of several files ──────────────────────────────────────────────

// ReviewFile is one file in a review: its path, what happened to it, and the
// three versions a merge is made from.
type ReviewFile struct {
	// Path is the file. It is required.
	Path string
	// Change is what happened to it.
	Change FileStatus
	// Base, Ours and Theirs are the merge base, this branch's version and
	// the other branch's. A file that was not in the base has an empty Base.
	Base, Ours, Theirs string
}

// MultiFileDiffReviewResult carries a MultiFileDiffReview and what was done
// in it.
type MultiFileDiffReviewResult struct {
	// Element is the review.
	Element  *ui.Element
	selected int
	approved string
	answered bool
}

// Selected is the file whose diff is on show, counted from zero, and -1 in a
// frame in which nothing was pressed.
func (r MultiFileDiffReviewResult) Selected() int {
	if !r.answered {
		return -1
	}
	return r.selected
}

// Approved is the path that was marked approved this frame, empty for none.
// The set itself is the caller's; this only names the key that changed.
func (r MultiFileDiffReviewResult) Approved() string { return r.approved }

// MultiFileDiffReviewOptions configure a MultiFileDiffReview.
type MultiFileDiffReviewOptions struct {
	// Files are the files under review, in the order they should be listed.
	Files []ReviewFile
	// Ours and Theirs name the two branches, and appear as the headings over
	// the two sides of every file that has a conflict. Empty takes "ours" and
	// "theirs".
	Ours, Theirs string
	// Selected is the file whose merged diff is on show, -1 for none. When it
	// is out of range the first file is shown anyway, because a review pane
	// with nothing in it reads as a broken one.
	Selected *int
	// Approved is the set of paths a person has signed off on, and the
	// caller's. It is a set rather than a count because a review is worked
	// through in an order nobody chose, and the tenth file approved has to
	// still be approved after the eleventh is looked at. It is required, and
	// has to be the caller's own map: Options arrives by value, so a set
	// made here would be this frame's copy and gone by the next one.
	Approved map[string]bool
	// Lang is which language's words the highlighter knows.
	Lang code.Lang
	// ListWidth is the share the file list takes of the pane. Zero takes a
	// third.
	ListWidth float32
	// Height is the height of the merged diff on the right, and is required.
	Height float32
	// Scroll is where the merged diff is scrolled to, and the caller's.
	Scroll *ui.ScrollState
}

// MultiFileDiffReview is the thing a person is asked about before a run's
// work is taken: every file it touched, each one merged against what the
// other branch did, and a place to say yes.
//
// The file list and the diff are side by side rather than stacked because
// approving means moving down a list while looking at what each entry is —
// and in a stacked layout, every file's diff is a scroll away from its name.
//
// Approving writes into the caller's set rather than holding a counter,
// because the set is the state a window is rebuilt from: a reviewer who has
// signed off on six of nine files must not lose that by closing and reopening
// the pane.
func MultiFileDiffReview(c *ui.Context, opts MultiFileDiffReviewOptions) MultiFileDiffReviewResult {
	u := core.Density(c).Unit()
	if len(opts.Files) == 0 {
		panic("agent: MultiFileDiffReview needs at least one file; a review of nothing is " +
			"indistinguishable from one that failed to draw")
	}
	if opts.Height <= 0 {
		panic("agent: MultiFileDiffReview needs a Height for the diff pane")
	}
	if opts.Approved == nil {
		// Checked before anything is drawn rather than at the click: a review
		// that refuses to approve is a broken review, and finding out at the
		// press is a frame too late to see which call site did it. See the
		// approve handler for why the set cannot be allocated here.
		panic("agent: Review needs a non-nil Approved map to write into")
	}
	for i, f := range opts.Files {
		if f.Path == "" {
			panic("agent: MultiFileDiffReview file " + itoa(i) + " has no Path")
		}
	}
	ours, theirs := opts.Ours, opts.Theirs
	if ours == "" {
		ours = "ours"
	}
	if theirs == "" {
		theirs = "theirs"
	}

	first := 0
	if opts.Selected != nil && *opts.Selected >= 0 && *opts.Selected < len(opts.Files) {
		first = *opts.Selected
	}
	share := opts.ListWidth
	if share <= 0 || share >= 1 {
		share = 0.34
	}

	var r MultiFileDiffReviewResult
	// The two panes are laid out here rather than through a SplitPane. A
	// split pane fills the window it is in, and a review is not the window:
	// putting one on the page would make it take the whole height and push
	// everything after it off the bottom.
	row := ui.Row(c).FillWidth().Gap(u).AlignItems(ui.Start).Children(func() {
		ui.Column(c).WidthPercent(share * 100).Shrink(0).Gap(u).Children(func() {
			for i, f := range opts.Files {
				i, f := i, f
				reviewFileRow(c, reviewFileRowOptions{
					file:     f,
					selected: i == first,
					approved: opts.Approved[f.Path],
				}, func() {
					r.answered = true
					r.selected = i
				}, func() {
					if opts.Approved == nil {
						// Options arrives by value, so a map we allocated here
						// would be written into this frame's copy and lost at
						// frame's end. Ask for the map instead of inventing one.
						panic("agent: Review needs a non-nil Approved map to write into")
					}
					opts.Approved[f.Path] = !opts.Approved[f.Path]
					r.approved = f.Path
				})
			}
		})
		layout.Divider(c, layout.DividerOptions{Vertical: true})
		ui.Box(c).Grow(1).Shrink(0).Children(func() {
			f := opts.Files[first]
			merged, conflicts := merge3(f.Base, f.Ours, f.Theirs)
			mergedDiff(c, mergedDiffOptions{
				path:      f.Path,
				lines:     merged,
				conflicts: conflicts,
				lang:      opts.Lang,
				ours:      ours,
				theirs:    theirs,
				height:    opts.Height,
				scroll:    opts.Scroll,
			})
		})
	})
	r.Element = row
	return r
}

// reviewFileRowOptions is one file row's own settings.
type reviewFileRowOptions struct {
	file     ReviewFile
	selected bool
	approved bool
}

// reviewFileRow is one file in the list: its path, its counts, and whether it
// has been signed off on.
func reviewFileRow(c *ui.Context, opts reviewFileRowOptions, onPress, onApprove func()) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	f := opts.file

	merged, conflicts := merge3(f.Base, f.Ours, f.Theirs)
	added, removed := 0, 0
	for _, l := range merged {
		switch l.Kind {
		case code.DiffAdded:
			added++
		case code.DiffRemoved:
			removed++
		}
	}

	name := f.Path + ", " + f.Change.String() + ", " + signed(added) + " " + signed(-removed)
	row := ui.Box(c).FillWidth().Radius(theme.SmallRadius).Children(func() {
		ui.Column(c).FillWidth().Gap(u * 0.25).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1).Children(func() {
				ui.Text(c, f.Path).TextColor(k.Text).Font(code.MonoStack).
					FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1).Grow(1)
				display.Tag(c, f.Change.String(), display.TagOptions{Tone: FileTone(f.Change)})
			})
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				if added > 0 {
					ui.Text(c, signed(added)).TextColor(k.Success).
						Font(code.MonoStack).FontSize(core.FontSize(c, theme.CaptionSize))
				}
				if removed > 0 {
					ui.Text(c, signed(-removed)).TextColor(k.Danger).
						Font(code.MonoStack).FontSize(core.FontSize(c, theme.CaptionSize))
				}
				if n := len(conflicts); n > 0 {
					display.Tag(c, itoa(n)+" conflict", display.TagOptions{Tone: core.Warning})
				}
				ui.Box(c).Grow(1)
				if f.Path != "" {
					approve := display.Icon(c, approveMark(opts.approved), display.IconOptions{
						Name: approveName(f.Path, opts.approved), Size: u * 4,
						Tone: approveTone(opts.approved), Muted: !opts.approved,
					})
					if approve.Clicked() && onApprove != nil {
						onApprove()
					}
				}
			})
		})
	})
	row.Label(name)
	if opts.selected {
		row.Background(k.SurfacePressed)
	}
	if onPress != nil && row.Clicked() {
		onPress()
	}
}

// approveMark is the tick an approved file wears and the empty circle it
// wears before that.
func approveMark(approved bool) display.IconName {
	if approved {
		return display.IconCheck
	}
	return display.IconStar
}

// approveTone is what an approve mark is inked in.
func approveTone(approved bool) core.Severity {
	if approved {
		return core.Success
	}
	return core.Neutral
}

// approveName is what the approve mark says out loud, and what a test clicks.
func approveName(path string, approved bool) string {
	if approved {
		return "Withdraw approval for " + path
	}
	return "Approve " + path
}

// mergedDiffOptions configures the merged diff on the right of a review.
type mergedDiffOptions struct {
	path   string
	lines  []code.DiffLine
	lang   code.Lang
	ours   string
	theirs string
	height float32
	scroll *ui.ScrollState
	// conflicts are the hunks in lines the two branches disagreed about, and
	// they are passed in rather than recomputed here: the caller already
	// merged to draw the lines, and a merge is not something to run twice in
	// one frame because a row wanted the count.
	conflicts []Conflict
}

// mergedDiff draws a merge's lines: the gutter, the marks, the bands, and a
// rule across every conflict.
//
// It draws the rows itself rather than calling code.Diff, and that is the one
// place this package does its own diffing: the conflicts are a property of
// the merge rather than of a pair of texts, and a hunk the reviewer has to
// notice by reading is a hunk that will be missed. The lines are code's, the
// chrome is code's, and only the band across a conflict is new.
func mergedDiff(c *ui.Context, opts mergedDiffOptions) *ui.Element {
	// The whole conflict is banded, not just its first line: a hunk whose
	// middle is unmarked is a hunk that reads as agreed-on where the rest of
	// it is not.
	conflictAt := make(map[int]bool, len(opts.conflicts)*4)
	for _, cf := range opts.conflicts {
		for i := cf.At; i < cf.At+cf.Lines() && i < len(opts.lines); i++ {
			conflictAt[i] = true
		}
	}

	var added, removed int
	for _, l := range opts.lines {
		switch l.Kind {
		case code.DiffAdded:
			added++
		case code.DiffRemoved:
			removed++
		}
	}

	// The body is a function so that CodePanel can draw the header above
	// it: a container appends children, and a header added afterwards
	// lands at the bottom of the panel.
	// The gutter is one width for both numbers on every row, measured from
	// the largest line number the merge can show: numbers sized by their
	// own digits start the code at a different x on every row.
	maxNo := 0
	for _, l := range opts.lines {
		maxNo = max(maxNo, max(l.Old, l.New))
	}
	u := core.Density(c).Unit()
	gutter := float32(len(itoa(maxNo)))*theme.CaptionSize*0.62 + u*2
	body := func() {
		host := layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Height: opts.height, State: opts.scroll,
		}, nil).Element.FillWidth()

		host.Children(func() {
			for i, l := range opts.lines {
				mergedRow(c, mergedRowOptions{
					line: l, ours: opts.ours, theirs: opts.theirs,
					lang: opts.lang, conflict: conflictAt[i], gutter: gutter,
				})
			}
		})

	}
	host := code.CodePanel(c, body, layout.ContainerOptions{}, func() {
		badges := make([]string, 0, 2)
		if added > 0 {
			badges = append(badges, "+"+itoa(added))
		}
		if removed > 0 {
			badges = append(badges, "-"+itoa(removed))
		}
		if len(opts.conflicts) > 0 {
			badges = append(badges, itoa(len(opts.conflicts))+" to resolve")
		}
		code.CodeHeader(c, opts.path, badges, func() {
			if len(opts.conflicts) > 0 {
				code.CodeBadge(c, opts.ours+" / "+opts.theirs, core.Warning)
			}
		})
	})
	return host
}

// mergedRowOptions is one merged line's own settings.
type mergedRowOptions struct {
	line     code.DiffLine
	ours     string
	theirs   string
	lang     code.Lang
	conflict bool
	// gutter is the width both line numbers take, right-aligned: numbers
	// sized by their own digits start the code at a different x on every
	// row, and a single-digit row reads its number as glued to the code.
	gutter float32
}

// mergedRow is one line of a merge, tinted by what happened to it and ruled
// across when it is one half of a conflict.
func mergedRow(c *ui.Context, opts mergedRowOptions) {
	k := core.Tokens(c)
	l := opts.line

	band, ink := ui.Color{}, k.Text
	switch l.Kind {
	case code.DiffAdded:
		band, ink = k.SuccessBg, k.Success
	case code.DiffRemoved:
		band, ink = k.DangerBg, k.Danger
	}
	mark := " "
	switch l.Kind {
	case code.DiffAdded:
		mark = "+"
	case code.DiffRemoved:
		mark = "−"
	}

	row := ui.Row(c).FillWidth().Shrink(0).Gap(1).AlignItems(ui.Center).Children(func() {
		// The number the line had in the base and the number it has in the
		// merge, each blank on the side the line is not on. A zero there
		// would read as line zero, which no file has. Both take the same
		// slot, right-aligned, so the code starts at one x all the way
		// down.
		for _, n := range [2]int{l.Old, l.New} {
			label := " "
			ink := k.TextFaint
			if n != 0 {
				label = itoa(n)
			}
			ui.Text(c, label).TextColor(ink).Font(code.MonoStack).Shrink(0).
				FontSize(core.FontSize(c, theme.CaptionSize)).
				Width(opts.gutter).TextAlign(ui.End)
		}
		ui.Text(c, mark).TextColor(ink).Font(code.MonoStack).Shrink(0).
			FontSize(core.FontSize(c, theme.RowSize))
		// Highlighted rather than plain: a merged diff is mostly Go, and a
		// reviewer reading a conflict wants the code to look like the code
		// they wrote it in, not like a wall of one colour.
		ui.Box(c).Grow(1).Shrink(0).Children(func() {
			// Laid out exactly as ui/code lays out a highlighted line — a
			// row of SingleLine runs, centred and shrunk — because that is the
			// one arrangement of several coloured pieces of a line that does
			// not break a tab-indented line in the middle of its own
			// indentation.
			ui.Row(c).AlignItems(ui.Center).Gap(0).Shrink(0).Children(func() {
				for _, tk := range code.Highlight(expandTabs(l.Text), opts.lang, false) {
					ui.Text(c, tk.Text).TextColor(code.TokenInk(c, tk.Kind)).
						Font(code.MonoStack).FontSize(core.FontSize(c, theme.RowSize)).
						SingleLine().Shrink(0)
				}
			})
		})
	})
	row.Label(mergedRowLabel(opts))
	if band.A != 0 {
		row.Background(band)
	}
	if opts.conflict {
		// A rule down the side rather than a colour behind the row: the row
		// already wears its band's colour, and the thing a reader needs to
		// see is where one conflict stops and the next begins.
		row.BorderWidth(theme.BorderWidth, theme.BorderWidth, 0, theme.BorderWidth*2).
			BorderColor(k.Warning)
	}
}

// expandTabs is a line's indentation drawn as spaces.
//
// A tab is measured by the text engine as a stop to the next multiple of its
// own width, and in a row of several coloured runs one tab-indented line comes
// out measurably wider than its neighbours and gets its tail ellipsised. Two
// spaces are what a tab means in Go source, so the line reads the same and
// every line in the hunk is the same length on screen.
func expandTabs(line string) string {
	if !strings.Contains(line, "\t") {
		return line
	}
	return strings.ReplaceAll(line, "\t", "  ")
}

// mergedRowLabel names a merged line for assistive technology and for a test.
func mergedRowLabel(opts mergedRowOptions) string {
	l := opts.line
	var what string
	switch l.Kind {
	case code.DiffAdded:
		what = "Added line " + itoa(l.New)
		if opts.conflict {
			what = "Conflict, " + opts.ours + ": line " + itoa(l.New)
		}
	case code.DiffRemoved:
		what = "Removed line " + itoa(l.Old)
		if opts.conflict {
			what = "Conflict, base line " + itoa(l.Old)
		}
	default:
		what = "Line " + itoa(l.New)
	}
	return l.Text + ", " + what
}
