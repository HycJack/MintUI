package files

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/theme"
)

// The marks a file interface is built from. Half of them are drawn twice and
// the other half are the reason a folder list and a grid of the same folder
// look like two views of one thing.

// entryRow is one row of a file list: the glyph, the name, the size or the
// count, and the time. Every list in this package that has an entry draws it
// with this, so a folder in a tree and a folder in the recent list are the
// same row.
// padLeft is how far the row is inset from the left edge: 0 for a list at
// the edge of a panel, and one step per level for a tree. It is a parameter
// rather than something set afterwards because MyGo has no padding-left; it
// has Padding(top, right, bottom, left), and a row's other three edges are
// already set.
func entryRow(c *ui.Context, name, detail string, kind Kind, chosen bool, padLeft float32, picked func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	row := ui.ButtonBase(c).FillWidth().Justify(ui.Start).AlignItems(ui.Center).
		Gap(u*1.5).Padding(u*0.5, u*1.5, u*0.5, padLeft).
		Radius(theme.SmallRadius).TextColor(k.Text).Label(name)
	switch {
	case chosen:
		row.Background(k.SurfaceHover)
	case row.Hovered():
		row.Background(k.SurfaceHover)
	}
	row.Children(func() {
		FileIcon(c, name, FileIconOptions{Kind: kind, Size: u * 4.25})
		// The box is a column, so the name claims no height: grown into
		// height it does not have, it lays out at zero tall and the detail
		// line prints straight through it.
		ui.Box(c).Grow(1).Children(func() {
			ui.Text(c, name).Ellipsis(name).
				FontSize(core.FontSize(c, theme.BodySize)).SingleLine()
			if detail != "" {
				ui.Text(c, detail).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			}
		})
		ui.Box(c).Grow(1)
	})
	if row.Clicked() && picked != nil {
		picked()
	}
	return row
}

// tileRow is the grid's version of entryRow: a square with the glyph in the
// middle and the name under it.
//
// The name is under rather than beside because a grid has no room beside:
// beside is where the size goes in a list, and a grid has no size column
// until it stops being a grid.
func tileRow(c *ui.Context, name, detail string, kind Kind, side float32, picked func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	tile := ui.ButtonBase(c).Size(side, side).Shrink(0).Fill().
		Justify(ui.Start).AlignItems(ui.Center).Gap(u * 0.25).
		Padding(u * 1.25).Radius(theme.SmallRadius).TextColor(k.Text).Label(name)
	if tile.Hovered() {
		tile.Background(k.SurfaceHover)
	}
	tile.Children(func() {
		FileIcon(c, name, FileIconOptions{Kind: kind, Size: u * 8})
		ui.Text(c, name).FillWidth().Ellipsis(name).TextAlign(ui.Center).
			FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		if detail != "" {
			ui.Text(c, detail).FillWidth().Ellipsis(detail).TextAlign(ui.Center).
				TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
	if tile.Clicked() && picked != nil {
		picked()
	}
	return tile
}

// meta is the second line of a row: the size and the time, with the size
// first because it is the thing a person scrolls a list looking for.
func meta(size int64, when string) string {
	switch {
	case size > 0 && when != "":
		return HumanSize(size) + " · " + when
	case size > 0:
		return HumanSize(size)
	}
	return when
}

// FileIconOptions configure a FileIcon.
type FileIconOptions struct {
	// Kind overrides the one worked out from the name. A caller that knows
	// — because the entry came from a directory listing rather than from a
	// filename — passes it; empty works it out from the name.
	Kind Kind
	// Size is the glyph's side in DIPs; zero gives the navigation size.
	Size float32
	// Muted draws in the secondary tone, which is what a glyph in a grid of
	// tiles should wear so the names are what the eye lands on.
	Muted bool
}

// FileIcon is the glyph for a file: the one of the library's thirty-five that
// its kind wears.
//
// It never draws its own. Every file type in every operating system has its
// own little picture, and a desktop that mixed those with the library's set
// would read as two applications in one window. The fallback for an
// extension nobody has heard of is the stored-thing glyph, which says "a file
// whose contents nobody has looked at" and is true.
//
// The name is required and is the file's own name, because a glyph is the
// one element in this library with no word of its own: a screen reader given
// nothing here announces "image".
func FileIcon(c *ui.Context, name string, opts FileIconOptions) *ui.Element {
	if name == "" {
		panic("files: FileIcon needs the file's name; a glyph with no name says nothing out loud")
	}
	k := opts.Kind
	if opts.Kind == Binary && kindIsUnknown(k) {
		// A caller that did not say and a name that does not: work it out.
		// Binary is the zero value, so "did not say" and "is a binary" are
		// the same int and the name is the only way to tell them apart.
		k = KindOf(name)
	}
	glyph, ok := extIcons[k]
	if !ok {
		glyph = extIcons[fallbackKind]
	}
	return display.Icon(c, glyph, display.IconOptions{
		Size: opts.Size, Muted: opts.Muted, Name: name + " file",
	})
}

// kindIsUnknown reports a kind that was left at its zero value, which is the
// only way FileIcon can tell "the caller did not say" from "it is binary".
// It is here rather than a *Kind because a pointer parameter to every icon
// call would be a worse API than one name to check.
func kindIsUnknown(k Kind) bool {
	_, ok := extIcons[k]
	return !ok
}
