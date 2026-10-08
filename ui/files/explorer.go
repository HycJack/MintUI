package files

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// Entry is one thing in a folder: a file, or another folder.
type Entry struct {
	// Name is the thing's own name, without any path in front of it.
	Name string
	// Path is where it is, from wherever the caller's tree is rooted. It is
	// what a click reports, because a name is only unique inside one folder.
	Path string
	// Kind is what it is. Leave it Binary for a file and let it be worked
	// out from the name; set it for a folder, which has no extension to
	// read.
	Kind Kind
	// Dir says it is a folder. A folder is drawn with the folder glyph and
	// says nothing about its size, because a folder's size is the size of
	// everything inside it and is not what the row is about.
	Dir bool
	// Size is the file's size in bytes.
	Size int64
	// Modified is when it changed, already formatted for display.
	Modified string
	// Children are the entries of a folder, when the caller has opened it.
	// An empty slice with Dir set is a folder nobody has opened yet, which
	// is drawn the same as an empty one — the difference is not knowable
	// from here and is not this component's business.
	Children []Entry
	// Count is how many things a folder holds, for a row that shows a
	// folder: "12 items" beside a name is what tells somebody whether
	// opening it is worth the click.
	Count int
	// Expanded says the folder is open. It is the caller's, because which
	// folders are open is the tree's state and not the tree's own — a
	// folder that remembered being open would reopen in a different
	// window and would be wrong there.
	Expanded bool
}

// kindOf is the entry's kind, working it out from the name when the caller
// did not say and it is not a folder.
func (e Entry) kindOf() Kind {
	if e.Dir {
		return Archive // a folder has no extension; the archive glyph is the stored thing
	}
	if e.Kind != Binary {
		return e.Kind
	}
	return KindOf(e.Name)
}

// FileExplorerOptions configure a FileExplorer.
type FileExplorerOptions struct {
	// Height is the height of the tree; required, as everywhere in this
	// package that scrolls.
	Height float32
	// Width is the width of the tree; zero measures the window.
	Width float32
	// ShowSize puts the size on a row as well as the time. Off by default
	// because the size is the wrong column to sort a folder by and the right
	// one to read in a search result.
	ShowSize bool
	// Indent is how far one level in is, in DIPs; zero gives the library's.
	Indent float32
}

// FileExplorerResult carries a FileExplorer and what was pressed in it.
type FileExplorerResult struct {
	// Element is the tree.
	Element *ui.Element
	// picked is the path pressed this frame, empty for none.
	picked string
	// toggled is the folder whose open state changed this frame, empty for
	// none.
	toggled string
}

// Picked returns the path pressed this frame, empty for none. It is a path
// rather than a row number because the caller has the entries and the
// numbers stop meaning the same thing as soon as the tree is re-sorted.
func (r FileExplorerResult) Picked() string { return r.picked }

// Toggled returns the path of the folder opened or closed this frame, empty
// for none.
func (r FileExplorerResult) Toggled() string { return r.toggled }

// FileExplorer is a folder as a tree: the folders open, their entries under
// them, one level further in for each open folder.
//
// Nothing is loaded. An entry that has children and is not expanded is drawn
// as a closed folder and costs one row, so a tree of a hundred thousand
// entries behind nine closed folders draws nine rows. Whoever wants them
// fetched — on a click, on a search, on a timer — decides, and this component
// draws whatever it is handed.
//
// The expand/collapse state is an Entry field rather than something kept
// here, because a tree that remembered its own state would reopen in the
// window the caller opened last, in the place they left it, which is right
// for a file dialog and wrong for everything else.
func FileExplorer(c *ui.Context, selected *string, entries []Entry, opts FileExplorerOptions) FileExplorerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Height <= 0 {
		panic("files: FileExplorer needs a Height; a tree with no height is every folder there is")
	}
	step := opts.Indent
	if step <= 0 {
		step = u * 3.5
	}

	var res FileExplorerResult
	tree := ui.Scroll(c).Height(opts.Height).FillWidth().Role(ui.RoleTree)
	tree.Children(func() {
		if len(entries) == 0 {
			ui.Text(c, core.Msg(c, "files.empty", core.Def("This folder is empty"))).
				TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
			return
		}
		for _, e := range entries {
			explorerEntry(c, selected, e, 0, step, opts, &res)
		}
	})
	res.Element = tree
	return res
}

// explorerEntry is one entry and, if it is an open folder, the entries under
// it. level is how many folders deep this one is, which is the only state
// this function carries — everything else is in the entry.
func explorerEntry(c *ui.Context, selected *string, e Entry, level int, step float32, opts FileExplorerOptions, res *FileExplorerResult) {
	u := core.Density(c).Unit()
	chosen := selected != nil && *selected == e.Path

	row := entryRow(c, e.Name, entryDetail(c, e, opts), e.kindOf(), chosen,
		u*1.5+float32(level)*step, func() {
			if selected != nil {
				*selected = e.Path
			}
			res.picked = e.Path
			if e.Dir {
				res.toggled = e.Path
			}
		})
	row.Label(e.Path)

	// The twisty is drawn beside the glyph rather than inside it, and only
	// for a folder: a file with an arrow beside it is a promise the tree
	// will not keep.
	if e.Dir {
		mark := display.IconChevron
		if e.Expanded {
			mark = display.IconChevronUp
		}
		display.Icon(c, mark, display.IconOptions{
			Size: u * 3, Muted: true,
			Name: e.Name + " folder, " + openWord(c, e.Expanded),
		})
	}
	if !e.Dir || !e.Expanded {
		return
	}
	for _, child := range e.Children {
		explorerEntry(c, selected, child, level+1, step, opts, res)
	}
}

// entryDetail is the second line of a row: a folder's count, a file's size
// and time, or nothing at all when the caller gave neither.
func entryDetail(c *ui.Context, e Entry, opts FileExplorerOptions) string {
	switch {
	case e.Dir:
		if e.Count <= 0 {
			return ""
		}
		return core.Msg(c, "files.items", core.Def(itoa(e.Count)+" items"))
	case opts.ShowSize && e.Size > 0:
		return meta(e.Size, e.Modified)
	}
	return e.Modified
}

func openWord(c *ui.Context, open bool) string {
	if open {
		return core.Msg(c, "files.open", core.Def("open"))
	}
	return core.Msg(c, "files.closed", core.Def("closed"))
}

// FileGridOptions configure a FileGrid.
type FileGridOptions struct {
	// Height is the height of the grid; zero grows to fit the rows, which
	// is right for a small folder in a window and wrong for a big one. Give
	// it one and the grid scrolls.
	Height float32
	// Side is the side of one tile in DIPs; zero gives the library's.
	Side float32
	// ShowSize puts the size under the name.
	ShowSize bool
}

// FileGridResult carries a FileGrid and what was pressed in it.
type FileGridResult struct {
	// Element is the grid.
	Element *ui.Element
	// picked is the path pressed this frame, empty for none.
	picked string
}

// Picked returns the path pressed this frame, empty for none.
func (r FileGridResult) Picked() string { return r.picked }

// FileGrid is the same folder as tiles rather than rows: glyph, name, and
// the size under it.
//
// It is a grid of a fixed tile side rather than a column layout, because a
// folder icon is a picture and a picture needs a shape. The number of tiles
// across comes from the width the caller gives, and the gap is the density's
// own step so that Comfortable really does give a bigger tile.
//
// Folders come before files and both come before nothing else. That is what
// every file manager does and nobody agrees with in the abstract: opening a
// folder is the common action, and it should not be below a folder of
// screenshots called 2024.
func FileGrid(c *ui.Context, selected *string, entries []Entry, opts FileGridOptions) FileGridResult {
	u := core.Density(c).Unit()
	side := opts.Side
	if side <= 0 {
		side = u * 22
	}
	gap := u * 1.5

	var res FileGridResult
	// The tiles are built inside the Children call rather than beside it:
	// MyGo parents a child by where it was CREATED, so a tile made before
	// the row that should hold it is that row's sibling and no amount of
	// putting it in afterwards will move it.
	if opts.Height > 0 {
		res.Element = layout.ScrollArea(c, layout.ScrollAreaOptions{
			Vertical: true, Horizontal: true, Height: opts.Height, Pad: u * 0.5,
		}, func() {
			gridTiles(c, selected, entries, opts, side, gap, &res)
		}).Element
		return res
	}
	grid := ui.Column(c).FillWidth().Role(ui.RoleList)
	grid.Children(func() { gridTiles(c, selected, entries, opts, side, gap, &res) })
	res.Element = grid
	return res
}

// gridTiles is the wrapped row of tiles. It is its own function so that the
// scrolling and the non-scrolling grids build them through exactly one path.
func gridTiles(c *ui.Context, selected *string, entries []Entry, opts FileGridOptions, side, gap float32, res *FileGridResult) {
	ui.Row(c).FillWidth().Wrap().Gap(gap).Children(func() {
		for _, e := range sortEntries(entries) {
			e := e
			tileRow(c, e.Name, gridDetail(c, e, opts), e.kindOf(), side, func() {
				if selected != nil {
					*selected = e.Path
				}
				res.picked = e.Path
			})
		}
	})
}

func gridDetail(c *ui.Context, e Entry, opts FileGridOptions) string {
	if e.Dir || !opts.ShowSize || e.Size <= 0 {
		return ""
	}
	return HumanSize(e.Size)
}

// sortEntries puts the folders first, and leaves the order of each group
// alone — a grid that re-sorted by date would make a folder jump about under
// the pointer as new files arrive.
func sortEntries(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.Dir {
			out = append(out, e)
		}
	}
	for _, e := range entries {
		if !e.Dir {
			out = append(out, e)
		}
	}
	return out
}

// PathBarOptions configure a PathBar.
type PathBarOptions struct {
	// Label names the bar; it is required, because a row of folder names is
	// a set of words with no indication of what they are.
	Label string
	// Crumb is the segment pressed this frame, empty for none.
	Crumb string
	// Root is the name of the top segment — "callbacks", or the name of a
	// place — rather than a slash.
	Root string
	// MaxSegments collapses the middle into an ellipsis past this many
	// segments. Zero shows all of them.
	MaxSegments int
}

// PathBar is where you are, as the path's own segments: each one a press
// that would take you there.
//
// The whole path is never a piece of text. A path of seven segments is too
// long for the bar and the part that matters is the end of it, so each
// segment is its own target and the bar can collapse its middle when there
// are too many — which is what a Finder path bar does and is the only way a
// deep path stays usable in a narrow window.
//
// The segments are joined by Join rather than by string concatenation, so a
// path bar showing a Windows path shows the same path the rest of the
// interface is holding.
func PathBar(c *ui.Context, path string, opts PathBarOptions) PathBarOptions {
	if opts.Label == "" {
		panic("files: PathBar needs a Label; a row of folder names says nothing about what it is")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	parts := splitPath(path)
	shown := parts
	if opts.MaxSegments > 0 && len(parts) > opts.MaxSegments {
		// The tail is what the reader came for; the head is what they came
		// from. Keeping the head and one of the tail and joining them in
		// the middle says the path is longer than the bar, which is true
		// and is better than silently dropping the part they are in.
		shown = append([]string{parts[0]}, parts[len(parts)-1:]...)
	}

	bar := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 0.25).
		Label(opts.Label).Role(ui.RoleNone)
	bar.Children(func() {
		for i, seg := range shown {
			if i > 0 {
				display.Icon(c, display.IconChevron, display.IconOptions{
					Size: u * 2.5, Muted: true, Name: "in",
				})
			}
			if i == 1 && len(shown) != len(parts) {
				ui.Text(c, "…").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.RowSize))
				display.Icon(c, display.IconChevron, display.IconOptions{
					Size: u * 2.5, Muted: true, Name: "in",
				})
			}
			name := seg
			if i == 0 && opts.Root != "" {
				name = opts.Root
			}
			// The last segment is where you are, and it is drawn in body
			// ink rather than as a button: a press on the folder you are
			// already in goes nowhere, and a control that goes nowhere
			// should not look like one.
			last := i == len(shown)-1
			crumb := ui.ButtonBase(c).Padding(u*0.25, u).Radius(theme.SmallRadius).
				TextColor(k.Text).Label(name).Disabled(last)
			if last {
				crumb.Background(k.Surface)
			} else if crumb.Hovered() {
				crumb.Background(k.SurfaceHover)
			}
			// The path this segment leads to is built by walking the path
			// from the top, so it is the path the caller is standing in
			// rather than one re-derived from the name.
			crumb.Children(func() {
				ui.Text(c, name).FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
			})
			if crumb.Clicked() {
				opts.Crumb = crumbOf(path, i)
			}
		}
	})
	return opts
}

// crumbOf is the path a segment of path leads to, by index. It is a function
// so that the bar and anything else that has to point at a folder agree.
func crumbOf(path string, i int) string {
	parts := splitPath(path)
	if i < 0 || i >= len(parts) {
		return path
	}
	return Join("/", parts[:i+1]...)
}
