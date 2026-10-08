package git

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/theme"
)

// Change is one path's worth of what is different from the commit.
type Change struct {
	// Path is the file's path, relative to the repository root and with
	// forward slashes on every platform — a path is compared against
	// strings, not opened, and "a\b" is not a path anywhere.
	Path string
	// Status is what happened to it. The letter, the word and the colour all
	// come from this one value, so a list cannot say "Modified" in a column
	// and paint it as a deletion.
	Status FileStatus
	// Staged says the change is in the index rather than only in the working
	// tree. It is what separates "git add" from "git commit" in a window
	// that shows both, and a list that does not draw it makes the two look
	// like one.
	Staged bool
	// Added and Deleted are how many lines the change adds and deletes.
	Added, Deleted int
}

// GitStatusBadgeOptions configure a GitStatusBadge.
type GitStatusBadgeOptions struct {
	// Severity overrides the status's own. It is for the one case the
	// status cannot know: a conflicted file that nobody has looked at yet
	// wants to be louder than the word "Conflicted" already makes it.
	Severity *core.Severity
	// Solid fills the pill with ink instead of tinting it, for the one
	// badge on a screen that should be the first thing the eye lands on.
	Solid bool
	// WithWord puts the word beside the letter. Off by default because the
	// letter is what a git reader looks for and the word is what stops them
	// reading it.
	WithWord bool
}

// GitStatusBadge is one file's status as a pill: the letter git prints, in
// the colour its severity says.
//
// The letter is not decoration and the word is not decoration either — they
// are the two halves of the same thing for two different people. Somebody who
// reads git status every day looks for M; somebody reviewing a pull request
// needs to know what M means. The badge can show either or both, and the
// colour is the severity's and never a colour of its own, so a Modified is
// the same amber in a status bar, in a row and in a badge.
func GitStatusBadge(c *ui.Context, status FileStatus, opts GitStatusBadgeOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	sev := StatusSeverity(status)
	if opts.Severity != nil {
		sev = *opts.Severity
	}
	bg, fg := sev.Pair(k)
	if status == Clean {
		bg, fg = k.Surface, k.TextMuted
	}
	if opts.Solid {
		bg, fg = k.Fill, k.OnFill
	}
	letter := StatusLetter(status)
	if letter == "" {
		letter = "·"
	}
	word := StatusWord(status)
	// The pill is named by the word whatever it shows: the letter is what a
	// reader looks for and the word is what a screen reader has to be given,
	// because two glyphs are not a description of anything.
	name := word
	if opts.WithWord {
		name = letter + " " + word
	}

	pill := ui.Box(c).Padding(u*0.25, u*1.5).Radius(theme.PillRadius).
		Background(bg).Label(name)
	pill.Children(func() {
		ui.Text(c, letter).TextColor(fg).Bold().
			FontSize(core.FontSize(c, theme.CaptionSize))
		if opts.WithWord {
			ui.Text(c, word).TextColor(fg).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
	})
	return pill
}

// ChangesListOptions configure a ChangesList.
type ChangesListOptions struct {
	// Height is the height of the table; required, as everywhere in this
	// package that scrolls.
	Height float32
	// Width is the width the columns share; zero measures the window.
	Width float32
	// Selected is the row the keys move from, -1 for none.
	Selected *int
	// Choice is the set of chosen rows when the list chooses several.
	Choice *data.Selectable
	// Sort is the column the rows are ordered by.
	Sort *data.Sort
	// WithWord puts the status word beside the letter in the status column.
	WithWord bool
	// State is where the list keeps its place between frames.
	State *ui.ListState
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// ChangesListResult carries a ChangesList and what was asked of it.
type ChangesListResult struct {
	// Element is the whole table.
	Element *ui.Element
	// sorted is the column whose head was clicked this frame.
	sorted string
	// picked is the path chosen this frame, empty for none.
	picked string
}

// Sorted returns the column the rows were asked to be ordered by this frame.
func (r ChangesListResult) Sorted() string { return r.sorted }

// Picked returns the path chosen this frame, empty for none. It is the path
// rather than the row number because a caller that stages a file wants the
// file, and a row number stops meaning that the moment the list is sorted.
func (r ChangesListResult) Picked() string { return r.picked }

// ChangesList is a table of what has changed: the status, the path, and the
// line counts.
//
// The row is chosen by clicking it, and Picked reports the path rather than
// the row: everything a caller does with a chosen change is done to the file,
// and a number that stops meaning the same file the moment the list is sorted
// is a number that will be used wrongly. The caller's *int follows the sort
// the way it follows every other table in the library.
func ChangesList(c *ui.Context, changes []Change, opts ChangesListOptions) ChangesListResult {
	k := core.Tokens(c)

	cols := []data.Column{
		{ID: "status", Title: "Status", Width: 108, Sortable: true},
		{ID: "path", Title: "Path", Share: 3, Sortable: true},
		{ID: "size", Title: "Changes", Width: 110, Align: ui.End, Sortable: true},
	}

	cell := func(row, col int) {
		ch := changes[row]
		switch cols[col].ID {
		case "status":
			GitStatusBadge(c, ch.Status, GitStatusBadgeOptions{
				WithWord: opts.WithWord,
				Severity: stagedSeverity(ch),
			})
		case "path":
			pathCell(c, ch.Path, ch.Staged)
		case "size":
			ui.Text(c, StatWord(ch.Added, ch.Deleted)).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
	}

	tr := data.DataTable(c, data.DataTableOptions{
		Columns: cols, Rows: len(changes), Height: opts.Height, Width: opts.Width,
		Selected: opts.Selected, Choice: opts.Choice, Sort: opts.Sort, State: opts.State,
		Key:  func(row int) any { return changes[row].Path },
		Cell: cell,
		Label: func(row int) string {
			ch := changes[row]
			return StatusWord(ch.Status) + ": " + ch.Path + ", " + StatWord(ch.Added, ch.Deleted)
		},
		CellLabel: func(row, col int) string {
			ch := changes[row]
			if cols[col].ID == "path" {
				return ch.Path
			}
			return StatusWord(ch.Status) + ": " + ch.Path
		},
		Empty: func() {
			if opts.Empty != nil {
				opts.Empty()
				return
			}
			ui.Text(c, noRows(c, "git.noChanges", core.Def("No changes"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
		},
	})
	return ChangesListResult{Element: tr.Element, sorted: tr.Sorted()}
}

// stagedSeverity is the badge's severity for a staged change, which is the
// accent rather than the status's own: the status says what happened and the
// accent says it is already in the index. A staged deletion is not louder for
// being staged, so nothing here touches the severity a status implies.
func stagedSeverity(ch Change) *core.Severity {
	if !ch.Staged {
		return nil
	}
	sev := core.Accent
	return &sev
}

// pathCell is a path as git prints it: the directories in the muted tone and
// the file's own name in body ink, so that the eye lands on the name in a
// list of twenty paths that share a prefix.
func pathCell(c *ui.Context, path string, staged bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	dir, name := splitPath(path)
	// A staged change wears an accent bar down its left edge rather than a
	// word in it: the point of the marker is to be scannable down the side of
	// a list, not to be read. The bar is a border and the gap is a padding,
	// because a row's other three edges already have their values and
	// MyGo has no padding-left to set on its own.
	bar := float32(0)
	if staged {
		bar = u
	}
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(0).
		Padding(0, 0, 0, bar).BorderWidth(0, 0, 0, bar).
		BorderColor(k.Accent).Children(func() {
		if dir != "" {
			ui.Text(c, dir).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
		}
		ui.Text(c, name).FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
	})
}

// splitPath divides a path into the part before the last separator and the
// name after it. A path with no separator is all name and no directory, which
// is a file at the repository root rather than an error.
func splitPath(path string) (dir, name string) {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i+1], path[i+1:]
		}
	}
	return "", path
}
