package files

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// FilePreviewOptions configure a FilePreview.
type FilePreviewOptions struct {
	// Name is the file's name. It is required: it is what says which file
	// this is, and a preview of nothing in particular is a picture of
	// nothing.
	Name string
	// Kind is what the file is. Leave it Binary to have it worked out from
	// the name — which is the right thing for anything a caller found on a
	// disk, and the wrong thing for a file it has already classified.
	Kind Kind
	// Text is the file's contents, for a file of text. It is the caller's
	// string and nothing here reads a disk: a component that opened a file
	// would block the frame it is drawn in, and would make every one of
	// these untestable.
	Text string
	// Image is the picture, already built. Same reason as Text: the bytes
	// were read by somebody who had to decide what to do when there were
	// too many, and this component draws what they built.
	Image *ui.Element
	// Size is the file's size in bytes, shown in the header.
	Size int64
	// Modified is when it changed, already formatted.
	Modified string
	// Width and Height are the preview's own size. They are required: a
	// preview that has no height is a preview as tall as its file, which
	// for a text file is a window that cannot be closed.
	Width, Height float32
	// MaxLines truncates a text preview past this many lines, with a note of
	// how many were left out. Zero shows the lot, which is right for a
	// small file and wrong for a log.
	MaxLines int
}

// FilePreview is a file as something to look at: a picture as a picture, a
// text file as its lines, and anything else as the reason it cannot be shown
// and what to do about it.
//
// Nothing is read. The text and the image are the caller's, because reading
// a file is the one thing a drawing function must not do: it would block the
// frame, it would happen again on every redraw, and it would make the whole
// component impossible to test without a disk. What this decides is what to
// draw, which is a decision about the kind and not about the bytes.
//
// The binary case is a real answer rather than an apology: it says what the
// file is, how big it is, and that it can still be downloaded or opened in
// its own program. "Cannot preview" with nothing after it is a dead end.
func FilePreview(c *ui.Context, opts FilePreviewOptions) *ui.Element {
	u := core.Density(c).Unit()
	if opts.Name == "" {
		panic("files: FilePreview needs the file's name; a preview of nothing in particular shows nothing")
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		panic("files: FilePreview needs a Width and a Height; a preview with no height is as tall as its file")
	}
	kind := opts.Kind
	if kind == Binary {
		kind = KindOf(opts.Name)
	}

	col := ui.Column(c).FillWidth().Height(opts.Height).Gap(u)
	col.Children(func() {
		previewHeader(c, opts, kind)

		body := ui.Column(c).FillWidth().Grow(1).Gap(u * 0.5)
		body.Children(func() {
			switch {
			case kind == Image && opts.Image != nil:
				layout.AspectRatio(c, layout.AspectRatioOptions{
					Ratio: 0, Cover: true,
				}, func() { opts.Image.Fill() })
			case kind == Image:
				// An image the caller has no picture for is the same problem
				// as a binary: say so, with the same words.
				noPreview(c, opts, "No preview of this image yet")
			case kind == Text:
				textPreview(c, opts)
			default:
				noPreview(c, opts, "No preview for "+strings.ToLower(kind.String())+" files")
			}
		})
		col.Width(opts.Width).Shrink(0)
	})
	return col
}

// previewHeader is the line above the preview: the glyph, the name, the size
// and the time.
func previewHeader(c *ui.Context, opts FilePreviewOptions, kind Kind) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
		FileIcon(c, opts.Name, FileIconOptions{Kind: kind, Size: u * 4.25})
		ui.Text(c, opts.Name).Grow(1).Ellipsis(opts.Name).
			FontSize(core.FontSize(c, theme.BodySize)).Bold().SingleLine()
		ui.Text(c, meta(opts.Size, opts.Modified)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
	})
	layout.Divider(c, layout.DividerOptions{})
}

// textPreview is a text file as its lines, in the monospaced face and the
// muted tone, with no wrapping: a line of source that wraps into two is
// source that cannot be read.
func textPreview(c *ui.Context, opts FilePreviewOptions) {
	k := core.Tokens(c)
	lines := splitTextLines(opts.Text)
	shown, hidden := lines, 0
	if opts.MaxLines > 0 && len(lines) > opts.MaxLines {
		shown, hidden = lines[:opts.MaxLines], len(lines)-opts.MaxLines
	}

	area := layout.ScrollArea(c, layout.ScrollAreaOptions{
		Vertical: true, Horizontal: true, Height: opts.Height * 0.8,
	}, func() {
		ui.Column(c).FillWidth().Gap(0).Children(func() {
			for _, line := range shown {
				ui.Text(c, line).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}
			if hidden > 0 {
				ui.Text(c, core.Msg(c, "files.moreLines",
					core.Def(itoa(hidden)+" more lines"))).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}
		})
	})
	area.Element.FillWidth()
}

// splitTextLines is a file's lines, without the empty piece a trailing
// newline leaves behind. A file that ends in a newline — which is every text
// file anybody writes — would otherwise report one line more than it has,
// and a preview that says "1 more line" about a file whose last line is on
// screen is lying.
func splitTextLines(s string) []string {
	lines := strings.Split(s, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}

// noPreview is what a file that cannot be shown says instead, and the two
// things that can be done about it.
func noPreview(c *ui.Context, opts FilePreviewOptions, why string) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Column(c).Fill().Center().Gap(u).Children(func() {
		previewGlyph(c, opts.Name)
		ui.Text(c, why).TextColor(k.TextMuted).TextAlign(ui.Center).
			FontSize(core.FontSize(c, theme.BodySize))
		ui.Text(c, HumanSize(opts.Size)).TextColor(k.TextFaint).
			FontSize(core.FontSize(c, theme.CaptionSize))
	})
}

// previewGlyph is the file's glyph at the size an empty state wants.
func previewGlyph(c *ui.Context, name string) *ui.Element {
	return FileIcon(c, name, FileIconOptions{Size: core.Density(c).Unit() * 12, Muted: true})
}

// ArchiveEntry is one file inside an archive.
type ArchiveEntry struct {
	// Name is the file's path inside the archive, with forward slashes.
	Name string
	// Size is how big it is uncompressed.
	Size int64
	// Compressed is how big it is in the archive, zero when unknown.
	Compressed int64
	// Modified is when it was written, already formatted.
	Modified string
	// Dir says it is a folder inside the archive.
	Dir bool
}

// ArchiveViewerOptions configure an ArchiveViewer.
type ArchiveViewerOptions struct {
	// Name is the archive's own name, shown in the header.
	Name string
	// Height is the height of the list; required.
	Height float32
	// Width is the width of the table; zero measures the window.
	Width float32
	// WithRatio draws how much each file is compressed by beside its size,
	// which is the one number in an archive anybody is looking for: the one
	// entry that did not compress is usually the reason the archive is as
	// big as it is.
	WithRatio bool
	// Empty draws instead of the rows when there are none.
	Empty func()
}

// ArchiveViewerResult carries an ArchiveViewer and what was pressed in it.
type ArchiveViewerResult struct {
	// Element is the whole thing.
	Element *ui.Element
	// picked is the entry pressed this frame, empty for none.
	picked string
}

// Picked returns the name of the entry pressed this frame, empty for none.
func (r ArchiveViewerResult) Picked() string { return r.picked }

// ArchiveViewer is what is inside an archive: the entries, their sizes, and
// what they compressed to.
//
// Nothing is extracted. An archive viewer that extracted to show a listing
// would write to a disk to answer a question, and the question — what is in
// here — does not need a disk. Extraction is a separate, deliberate act, and
// it belongs to whoever runs it.
//
// Folders come first for the same reason they do in a grid: opening one is
// the common action and it should not be below a screenshot.
func ArchiveViewer(c *ui.Context, entries []ArchiveEntry, opts ArchiveViewerOptions) ArchiveViewerResult {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Height <= 0 {
		panic("files: ArchiveViewer needs a Height; a listing with no height is every entry in the archive")
	}

	var res ArchiveViewerResult
	col := ui.Column(c).FillWidth().Gap(u)
	col.Children(func() {
		if opts.Name != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				FileIcon(c, opts.Name, FileIconOptions{Kind: Archive, Size: u * 4.25})
				ui.Text(c, opts.Name).Grow(1).Ellipsis(opts.Name).
					FontSize(core.FontSize(c, theme.BodySize)).Bold().SingleLine()
				ui.Text(c, core.Msg(c, "files.entries",
					core.Def(itoa(len(entries))+" entries"))).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			})
			layout.Divider(c, layout.DividerOptions{})
		}

		list := ui.Scroll(c).Height(opts.Height).FillWidth().Gap(u * 0.25).Role(ui.RoleList)
		list.Children(func() {
			if len(entries) == 0 {
				if opts.Empty != nil {
					opts.Empty()
					return
				}
				ui.Text(c, core.Msg(c, "files.emptyArchive", core.Def("This archive is empty"))).
					TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
				return
			}
			for _, e := range sortArchive(entries) {
				e := e
				row := entryRow(c, Base(e.Name), archiveDetail(c, e, opts.WithRatio),
					archiveKind(e), false, u*1.5, func() { res.picked = e.Name })
				row.Label(e.Name)
			}
		})
	})
	res.Element = col
	return res
}

// archiveDetail is a file's size and how much smaller it got in the archive.
func archiveDetail(c *ui.Context, e ArchiveEntry, withRatio bool) string {
	if !withRatio || e.Compressed <= 0 || e.Size <= 0 || e.Compressed >= e.Size {
		return HumanSize(e.Size)
	}
	saved := Percent(e.Size-e.Compressed, e.Size)
	return HumanSize(e.Size) + core.Def(" · ") + itoa(saved) + core.Def("% saved")
}

func archiveKind(e ArchiveEntry) Kind {
	if e.Dir {
		return Archive
	}
	return KindOf(e.Name)
}

// sortArchive puts the folders first and leaves the order of each group
// alone, for the same reason sortEntries does.
func sortArchive(entries []ArchiveEntry) []ArchiveEntry {
	out := make([]ArchiveEntry, 0, len(entries))
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
