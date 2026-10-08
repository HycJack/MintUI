package chat

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// What a model's answer is made of, and how it is drawn.
//
// Every component here reads the same tree [Parse] returns. That is deliberate:
// if MarkdownView, QuoteReply and SearchProgress each had their own idea of
// what a heading was, a transcript would show one answer three ways depending
// on which component happened to be holding it.

// ── markdown ───────────────────────────────────────────────────────────────

// MarkdownViewOptions configure a MarkdownView.
type MarkdownViewOptions struct {
	// Source is the markdown. Ignored when Blocks is given.
	Source string
	// Blocks is markdown already parsed by the caller — the same answer shown
	// in two places should be parsed once, and a caller already holding the
	// tree should not have to join it back into a string to draw it.
	Blocks []Block
	// Caption names the whole of the content for assistive technology. Empty
	// leaves it unnamed, which is right when the content reads as itself.
	Caption string
	// MaxWidth caps the measure. Zero takes the reading width — the same one
	// the bubbles use, so prose inside a bubble and prose outside one wrap at
	// the same column.
	MaxWidth float32
	// Tight drops the space above and below each block, for markdown drawn
	// inside something that already has room of its own.
	Tight bool
}

// MarkdownView draws markdown.
//
// It holds nothing: what it shows is exactly what Parse read out of the
// source, which is why the parser's tests and the drawing's tests can be
// written against different things and neither can drift into the other.
func MarkdownView(c *ui.Context, opts MarkdownViewOptions) *ui.Element {
	blocks := opts.Blocks
	if blocks == nil {
		blocks = Parse(opts.Source)
	}
	u := core.Density(c).Unit()

	width := opts.MaxWidth
	if width <= 0 {
		width = ReadingWidth
	}

	view := ui.Column(c).FillWidth().MaxWidth(width).Gap(u * 2)
	if opts.Caption != "" {
		view.Label(opts.Caption)
	}
	view.Children(func() {
		for _, b := range blocks {
			drawBlock(c, b, opts.Tight)
		}
	})
	return view
}

// drawBlock draws one block. It is the whole switch: every kind Parse can
// return has exactly one branch, so a kind added to the parser cannot be
// forgotten here — the compiler says so, which is the point of a switch on
// the kind rather than a chain of ifs.
func drawBlock(c *ui.Context, b Block, tight bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	gap := u * 2
	if tight {
		gap = u
	}

	switch b.Kind {
	case BlockHeading:
		size := headingSize(b.Level)
		// The role is set on the rich text rather than on a text element of
		// its own: a heading is not a line of bold, and assistive technology
		// has to meet it as a heading for the outline of an answer to be
		// navigable at all. The row under the heading is what gives the rich
		// text its width: inline grows to fill, and grown in a column that
		// is height — which it does not have — and the next block prints
		// straight through it.
		ui.Row(c).FillWidth().Children(func() {
			inline(c, b.Spans, size, k.Text, true).Role(ui.RoleHeading)
		})

	case BlockPara:
		// The row gives the paragraph its width, for the same reason the
		// heading's does: inline's grow means width, and only inside a row.
		ui.Row(c).FillWidth().Children(func() {
			inline(c, b.Spans, theme.BodySize, k.Text, false)
		})

	case BlockCode:
		CodeBlock(c, CodeBlockOptions{Code: b.Code, Lang: b.Lang})

	case BlockList:
		ui.Column(c).FillWidth().Gap(u * 0.75).Children(func() {
			for _, item := range b.Items {
				ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u).Children(func() {
					// The marker is a column of its own rather than a prefix
					// in the text: an item that wraps has to line its second
					// line up with its first, and a prefix cannot do that.
					marker(c, item)
					for range item.Depth {
						ui.Box(c).Width(u * 1.5).Shrink(0)
					}
					inline(c, item.Spans, theme.BodySize, k.Text, false).Grow(1)
				})
			}
		})

	case BlockQuote:
		// The rule is the quote's own child rather than a border on it, so it
		// runs the height of the text beside it. A border would draw a box
		// round each quote, and a quoted paragraph is not a box.
		ui.Row(c).FillWidth().AlignItems(ui.Stretch).Gap(u * 2).Children(func() {
			ui.Box(c).Width(theme.BorderWidth * 3).Shrink(0).Background(k.Border)
			ui.Column(c).Grow(1).Gap(gap).Children(func() {
				for _, inner := range b.Blocks {
					drawBlock(c, inner, true)
				}
			})
		})

	case BlockTable:
		TableBlock(c, TableBlockOptions{Head: b.Head, Rows: b.Rows, Aligns: b.Aligns})

	case BlockRule:
		ui.Box(c).FillWidth().Height(theme.BorderWidth).Shrink(0).
			Background(k.Border).MarginY(gap * 0.5)
	}
}

// headingSize is how large a heading of each level is. It is clamped to the
// reading column's own scale: a chat answer does not carry a page's 40pt
// headline, because the largest heading that still reads as prose inside a
// bubble is about a third bigger than the prose beside it, and past that the
// bubble stops being a bubble.
func headingSize(level int) float32 {
	switch level {
	case 1:
		return theme.BodySize * 1.35
	case 2:
		return theme.BodySize * 1.18
	}
	return theme.BodySize
}

// marker is what a list item is marked with, in its own column.
//
// The source's own marker is kept for a numbered list, because renumbering
// "1. 2. 4." would hide the thing its author was pointing at. A bullet is
// drawn rather than typed for the reason a checkbox is: the glyph every font
// happens to have is not the same glyph in every font, and it is not the same
// size either.
func marker(c *ui.Context, item Item) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	col := ui.Column(c).Width(u * 4).Shrink(0).AlignItems(ui.End)
	col.Children(func() {
		if item.Checked >= 0 {
			taskBox(c, item.Checked == 1)
			return
		}
		text := "•"
		if item.Ordered {
			text = item.Marker
			if text == "" {
				text = "1."
			}
		}
		ui.Text(c, text).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.MetaSize))
	})
}

// taskBox is the checkbox of a "- [ ]" item.
func taskBox(c *ui.Context, done bool) *ui.Element {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	side := u * 3.25
	name := "not done"
	if done {
		name = "done"
	}
	box := ui.Box(c).Size(side, side).Radius(side*0.28).Shrink(0).Label(name).
		Border(theme.BorderWidth, k.Border)
	if done {
		box = box.Background(k.Accent).Border(theme.BorderWidth, k.Accent)
		box.Children(func() {
			// The tick is the accent's own ink rather than the window's text
			// colour: a tick in body ink on an accent box disappears the
			// moment the accent is dark, which is exactly the case a dark
			// window hits.
			ui.Text(c, "✓").TextColor(k.OnFill).FontSize(side * 0.75)
		})
	}
	return box
}

// itoa is a small integer as text. The library has one for commas and a
// general one, but the places that want a bare number here want it in a label
// where a comma would be wrong — "[2 of 5]", not "[1,024 of 5]".
func itoa(n int) string { return strconv.Itoa(n) }

// inline draws a run of spans as one piece of rich text.
//
// It is the only place markdown's inline marks become type, which is what
// keeps bold the same weight in a heading, a list item and a quote — and keeps
// a code span the same size in all three. The marks carry no colours of their
// own: a link takes the palette's accent, so it is a link in a dark window
// too, and the same word cannot be one colour in a quote and another in the
// paragraph under it.
func inline(c *ui.Context, spans []Span, size float32, ink ui.Color, bold bool) *ui.Element {
	k := core.Tokens(c)
	out := make([]ui.Span, 0, len(spans))
	for _, s := range spans {
		sp := ui.Span{Text: s.Text, Size: core.FontSize(c, size), Color: ink}
		if s.Bold || bold {
			sp.Weight = 700
		}
		if s.Italic {
			sp.Italic = true
		}
		switch {
		case s.Code:
			// A code span in running text is a chip, not just a smaller face:
			// at paragraph size, monospaced text set as body type is hard to
			// tell from emphasis at a glance.
			sp.Size = core.FontSize(c, size) * 0.88
			sp.Background = k.SurfaceHover
			sp.Color = k.Text
		case s.Link != "":
			sp.Color = k.AccentText
			// Always underlined: colour alone would leave it
			// indistinguishable to anyone who cannot see it.
			sp.Underline = true
		}
		out = append(out, sp)
	}
	if len(out) == 0 {
		// An empty run still takes a line: a heading with nothing in it is
		// still a line of space between two paragraphs.
		return ui.Box(c).Height(core.FontSize(c, size) * 1.4)
	}
	e := ui.RichText(c, out...).TextColor(ink).FontSize(core.FontSize(c, size))
	// The grow is what makes it wrap: a rich text measured at its own widest
	// line would never break, and a paragraph that cannot break is a
	// transcript that scrolls sideways.
	e.Grow(1)
	return e
}

// ── code ───────────────────────────────────────────────────────────────────

// CodeBlockOptions configure a CodeBlock.
type CodeBlockOptions struct {
	// Code is the source. Required: a block with nothing in it is not a code
	// block, it is a stray pair of fences.
	Code string
	// Lang is the info string, which says how to colour it and is shown as a
	// caption when there is no title.
	Lang string
	// Title heads the block — a path, a function's name — and is what the
	// block is called for assistive technology.
	Title string
	// MaxHeight caps the block and scrolls what is over. Zero draws the whole
	// of it, which is right for a short answer and wrong for a pasted file.
	MaxHeight float32
	// LineNumbers numbers the lines, for code someone is about to talk about.
	LineNumbers bool
	// StartLine is the number of the first line, for a snippet lifted out of
	// the middle of a file.
	StartLine int
	// Copy names a copy button, which puts Code on the clipboard. Empty draws
	// no button: copying is the caller's business, and a button that does
	// something nobody asked for is a surprise.
	Copy string
	// Label names the block; empty takes the title.
	Label string
}

// CodeBlockResult carries a CodeBlock and the press of its copy button.
type CodeBlockResult struct {
	// Element is the whole block.
	Element *ui.Element
	copied  bool
}

// Copied reports the copy button being pressed this frame. It is a pulse, like
// every other press in this library: a caller reading it a frame later would
// read it as never having happened at all.
func (r CodeBlockResult) Copied() bool { return r.copied }

// CodeBlock is a fenced code block: coloured, bounded, and copyable if the
// caller wants it to be.
//
// The source is one rich text rather than an element per token, because an
// answer with three code blocks in it would otherwise be nine hundred elements
// inside a scroll view. One element carries the same colours for the same
// price as one element of plain text.
func CodeBlock(c *ui.Context, opts CodeBlockOptions) CodeBlockResult {
	if opts.Code == "" {
		panic("chat: CodeBlock needs Code; an empty block is a stray pair of fences")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	name := opts.Label
	if name == "" {
		name = opts.Title
	}

	var res CodeBlockResult
	block := ui.Column(c).FillWidth().Radius(theme.SmallRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border)
	if name != "" {
		block.Label(name)
	}

	block.Children(func() {
		if name != "" || opts.Copy != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u).
				Padding(u, u*2).Children(func() {
				if name != "" {
					ui.Text(c, name).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).Grow(1)
				} else {
					ui.Box(c).Grow(1)
				}
				if opts.Lang != "" {
					ui.Text(c, opts.Lang).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
				}
				if opts.Copy != "" {
					btn := ui.ButtonBase(c).Radius(theme.PillRadius).
						Padding(u*0.5, u*1.5).Label(opts.Copy).TextColor(k.TextMuted).
						Children(func() {
							ui.Text(c, opts.Copy).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize))
						})
					if btn.Clicked() {
						c.WriteClipboard(opts.Code)
						res.copied = true
					}
				}
			})
			ui.Box(c).FillWidth().Height(theme.BorderWidth).Shrink(0).Background(k.Border)
		}
		codeBody(c, opts, block)
	})
	res.Element = block
	return res
}

// codeBody puts the coloured source into host, scrolling it if the caller gave
// a height.
//
// The line numbers are a column beside the text rather than a prefix in it: a
// prefixed "12│" shifts every line by the width of the numbers, and code whose
// lines no longer line up with each other cannot be read as the shape it is.
func codeBody(c *ui.Context, opts CodeBlockOptions, host *ui.Element) {
	lines := func() {
		k, u := core.Tokens(c), core.Density(c).Unit()
		size := core.FontSize(c, theme.MetaSize)
		cut := splitTokenLines(Highlight(opts.Code, opts.Lang))

		if !opts.LineNumbers {
			ui.Column(c).FillWidth().Padding(u*2, u*2.5).Children(func() {
				for _, line := range cut {
					ui.RichText(c, uiSpans(line, size, k)...)
				}
			})
			return
		}
		first := max(opts.StartLine, 1)
		ui.Column(c).FillWidth().Padding(u*2, u*2.5).Gap(u * 0.5).Children(func() {
			for i, line := range cut {
				ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u * 1.5).Children(func() {
					ui.Text(c, internal.Commas(first+i)).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
					ui.RichText(c, uiSpans(line, size, k)...)
				})
			}
		})
	}

	// The scroll view is made inside host's Children because a child is
	// whatever was CREATED while that call was running — a scroll built out
	// here would land as a sibling of the block and float outside it.
	host.Children(func() {
		if opts.MaxHeight <= 0 {
			lines()
			return
		}
		scroll := ui.ScrollBoth(c).MaxHeight(opts.MaxHeight).FillWidth()
		scroll.Children(lines)
	})
}

// uiSpans turns the highlighter's tokens into rich-text spans: the one
// translation between the two halves of this package, where a token is a kind
// and some text and a span is a colour and that text.
func uiSpans(tokens []Token, size float32, k theme.Tokens) []ui.Span {
	out := make([]ui.Span, 0, len(tokens))
	for _, t := range tokens {
		fg, _ := TokenColor(t.Kind, k.IsDark())
		out = append(out, ui.Span{Text: t.Text, Size: size, Color: fg})
	}
	return out
}

// splitTokenLines cuts the flat token stream into one list per line.
//
// The scanner emits a single run across the whole source, which is right for a
// copy and wrong for line numbers: without this, the numbers would sit beside
// the first line and every other line would be a number with nothing next to
// it.
func splitTokenLines(tokens []Token) [][]Token {
	var lines [][]Token
	var cur []Token
	flush := func() { lines = append(lines, cur) }
	for _, t := range tokens {
		for {
			i := strings.IndexByte(t.Text, '\n')
			if i < 0 {
				break
			}
			if i > 0 {
				cur = append(cur, Token{Kind: t.Kind, Text: t.Text[:i]})
			}
			flush()
			cur = nil
			t = Token{Kind: t.Kind, Text: t.Text[i+1:]}
		}
		if t.Text != "" {
			cur = append(cur, t)
		}
	}
	// A source ending in a newline has one more line — the empty one — and it
	// is a line the reader can see the end of, so it is kept.
	flush()
	return lines
}

// ── tables ─────────────────────────────────────────────────────────────────

// TableBlockOptions configure a TableBlock.
type TableBlockOptions struct {
	// Head is the head row. A table with no head row is a list of rows with
	// its first row in a different colour, and the caller should use a list.
	Head []string
	// Rows are the body rows.
	Rows [][]string
	// Aligns are the columns' alignments, from the delimiter row. A short
	// slice leaves the rest alone.
	Aligns []Align
	// Caption names the table for assistive technology; empty takes the head.
	Caption string
}

// TableBlock is a markdown table.
//
// It draws itself rather than going through data.DataTable, for one of the
// two reasons that package documents about itself: DataTable needs a Height,
// because it is a viewport that builds only the rows in view. A markdown
// table's height is whatever its text turns out to be, it is never long enough
// to be worth building lazily, and a scroll view inside a bubble is a control
// nobody can hit. What it does take from DataTable is the rule underneath that
// one: the columns keep their width, and a table too narrow for them scrolls
// rather than squeezing names into dashes.
func TableBlock(c *ui.Context, opts TableBlockOptions) *ui.Element {
	if len(opts.Head) == 0 {
		panic("chat: TableBlock needs a Head; a table of rows with no row of names " +
			"is a list")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	caption := opts.Caption
	if caption == "" {
		caption = strings.Join(opts.Head, ", ")
	}

	// The columns are laid out right to left so each one can take what is
	// left over: a row whose first cell took everything would leave the
	// second with nothing, and a markdown table's columns are as wide as the
	// widest thing in them rather than equal shares.
	row := func(cells []string, head bool) {
		if len(cells) == 0 {
			return
		}
		ink, size := k.Text, theme.MetaSize
		if head {
			ink, size = k.TextMuted, theme.CaptionSize
		}
		ui.Row(c).FillWidth().AlignItems(ui.Start).Gap(u*2).
			Padding(u*1.25, u*2).Children(func() {
			for i := len(opts.Head) - 1; i >= 0; i-- {
				text := ""
				if i < len(cells) {
					text = cells[i]
				}
				cell := ui.Box(c).Grow(1).Shrink(0).Justify(tableAlign(opts.Aligns, i))
				cell.Children(func() {
					t := ui.Text(c, text).TextColor(ink).FontSize(core.FontSize(c, size))
					if head {
						t.Bold()
					}
				})
			}
		})
	}

	table := ui.Column(c).FillWidth().Radius(theme.SmallRadius).
		Border(theme.BorderWidth, k.Border).Label(caption)
	table.Children(func() {
		row(opts.Head, true)
		// A rule under the head and between the rows, and none around the
		// table: a frame would say "this is a thing" over what is part of a
		// paragraph.
		ui.Box(c).FillWidth().Height(theme.BorderWidth).Shrink(0).Background(k.Border)
		for _, r := range opts.Rows {
			row(r, false)
		}
	})
	return table
}

// tableAlign is a column's alignment as MyGo spells it.
func tableAlign(aligns []Align, col int) ui.Align {
	if col >= len(aligns) {
		return ui.Start
	}
	switch aligns[col] {
	case AlignCenter:
		return ui.Center
	case AlignRight:
		return ui.End
	}
	return ui.Start
}

// ── images and attachments ─────────────────────────────────────────────────

// ImageGridOptions configure an ImageGrid.
type ImageGridOptions struct {
	// Items are the images. Alt text is required of every one of them: an
	// image with no description is a rectangle, and it is the caller's job to
	// have described it rather than the viewer's excuse for not.
	Items []ImageItem
	// Columns is how many across. Zero takes three, which is what two to four
	// screenshots in an answer want.
	Columns int
	// Size is a tile's side in DIPs. Zero takes a third of the reading width.
	Size float32
	// Gap is the space between tiles.
	Gap float32
	// Label names the group of images; empty takes the first one's alt text.
	Label string
}

// ImageItem is one image: what it shows, and how to draw it.
type ImageItem struct {
	// Alt is what the image is.
	Alt string
	// Draw builds the tile's contents. A function rather than a bitmap so that
	// a caller with a picture hands one in and a caller without can draw the
	// tile that says it has none.
	Draw func()
}

// ImageGrid lays images out in a grid of equal tiles.
//
// It is a grid and not a column because the common case in an answer is two to
// four screenshots side by side, and a column of them would push the words
// around them three screens down.
func ImageGrid(c *ui.Context, opts ImageGridOptions) *ui.Element {
	if len(opts.Items) == 0 {
		panic("chat: ImageGrid needs Items; an empty grid is a box with nothing in it")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	for i, item := range opts.Items {
		if item.Alt == "" {
			panic("chat: ImageGrid item " + internal.Commas(i) + " needs Alt; an image " +
				"with no description is a rectangle")
		}
	}

	size := opts.Size
	if size <= 0 {
		size = ReadingWidth / 3
	}
	gap := opts.Gap
	if gap <= 0 {
		gap = u
	}
	cols := opts.Columns
	if cols <= 0 {
		cols = 3
	}
	name := opts.Label
	if name == "" {
		name = opts.Items[0].Alt
	}

	grid := ui.Column(c).FillWidth().Gap(gap).Label(name)
	grid.Children(func() {
		for start := 0; start < len(opts.Items); start += cols {
			chunk := opts.Items[start:min(start+cols, len(opts.Items))]
			ui.Row(c).FillWidth().Gap(gap).Children(func() {
				for _, item := range chunk {
					tile := ui.Box(c).Size(size, size).Radius(theme.SmallRadius).
						Background(k.Surface).Border(theme.BorderWidth, k.Border).
						Clip().Label(item.Alt)
					if item.Draw != nil {
						tile.Children(item.Draw)
					} else {
						tile.Center().Children(func() {
							ui.Text(c, item.Alt).TextColor(k.TextMuted).
								FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
						})
					}
				}
			})
		}
	})
	return grid
}

// AttachmentChipResult carries an AttachmentChip and what was done to it.
type AttachmentChipResult struct {
	// Element is the chip.
	Element *ui.Element
	removed bool
	opened  bool
}

// Removed reports the chip's remove button being pressed.
func (r AttachmentChipResult) Removed() bool { return r.removed }

// Opened reports the chip being pressed anywhere but on the remove button,
// which is how a chip with a file behind it says "open this".
func (r AttachmentChipResult) Opened() bool { return r.opened }

// AttachmentChipOptions configure an AttachmentChip.
type AttachmentChipOptions struct {
	// Name is the file's name. Required: a chip with no name is a shape.
	Name string
	// Meta is the line under it — a size, a page count, a type.
	Meta string
	// Mark names the picture before the text. Required, for the reason a glyph
	// is required everywhere else in this library.
	Mark string
	// Tone tints the chip's edge, for a file that could not be attached.
	Tone core.Severity
	// Removable draws the remove button, reporting through Removed.
	Removable bool
	// Openable makes the chip itself a press, reporting through Opened.
	Openable bool
	// Label names the chip; empty takes the name.
	Label string
}

// AttachmentChip is one attached file: its name, what is known about it, and
// the two things a person can do with it.
//
// The two actions are separate results because they are separate promises —
// removing an attachment is usually undoable and opening one is not — and a
// caller handed one boolean would have to guess which happened.
func AttachmentChip(c *ui.Context, opts AttachmentChipOptions) AttachmentChipResult {
	if opts.Name == "" {
		panic("chat: AttachmentChip needs a Name")
	}
	if opts.Mark == "" {
		panic("chat: AttachmentChip needs a Mark; a chip whose picture has no name is " +
			"an image to a screen reader")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	label := opts.Label
	if label == "" {
		label = opts.Name
	}

	chip := ui.Box(c).Shrink(0).Radius(theme.ControlRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).Label(label)
	if opts.Tone != core.Neutral {
		// The edge is tinted rather than the fill: a chip the reader has to
		// re-read the words on is worse than one they notice late.
		_, fg := opts.Tone.Pair(k)
		chip = chip.BorderColor(fg.Alpha(0.45))
	}

	var res AttachmentChipResult
	chip.Children(func() {
		if opts.Openable {
			// The whole chip is the press; the cross is not drawn inside it,
			// because a remove button that is also an open button is a control
			// that opens the file you were trying to take off.
			open := ui.ButtonBase(c).AlignItems(ui.Center).
				Radius(theme.PillRadius).Padding(u, u*2).Label(label).
				TextColor(k.Text).Children(func() {
				chipContents(c, opts, false, &res.removed)
			})
			if open.Clicked() {
				res.opened = true
			}
		} else {
			chipContents(c, opts, opts.Removable, &res.removed)
		}
	})
	res.Element = chip
	return res
}

// chipContents is what a chip shows — the mark, the name, the line under it
// and the remove button — and it sets removed when the button is pressed.
//
// It takes the flag rather than returning it because the button is drawn
// inside a closure three frames deep, and a pointer is the only way to carry a
// press back out of a drawing callback without making every chip on the screen
// share one package-level flag.
//
// It is a function rather than written out twice because both arrangements of
// a chip need it: one where the whole thing opens and one where only the
// cross does. Two copies of these lines would be two shapes that drift.
func chipContents(c *ui.Context, opts AttachmentChipOptions, removable bool, removed *bool) {
	k, u := core.Tokens(c), core.Density(c).Unit()

	ui.Row(c).AlignItems(ui.Center).Gap(u * 1.25).Children(func() {
		ui.Box(c).Size(u*3.25, u*4).Shrink(0).Role(ui.RoleNone).
			Label(opts.Mark).Draw(func(p *ui.Painter, r ui.Rect) {
			documentMark(p, r, k.TextMuted)
		})
		if opts.Meta != "" {
			ui.Column(c).AlignItems(ui.Start).Children(func() {
				ui.Text(c, opts.Name).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
				ui.Text(c, opts.Meta).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			})
		} else {
			ui.Text(c, opts.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
		}
		if removable {
			cross := ui.Box(c).Size(u*2.75, u*2.75).Shrink(0).Role(ui.RoleNone).
				Label("Remove " + opts.Name).
				Draw(func(p *ui.Painter, r ui.Rect) { markCross(p, r, k.TextMuted) })
			if cross.Clicked() {
				*removed = true
			}
		}
	})
}

// documentMark is the folded-corner page that says "a file". It is drawn rather
// than taken from the icon set because a chip is a piece of a transcript and
// not a row in a browser: the mark is the same shape whatever the file is,
// because a chip's mark says "file" and its name says which.
func documentMark(p *ui.Painter, r ui.Rect, col ui.Color) {
	fold := min(r.W, r.H) * 0.38
	var path ui.Path
	path.MoveTo(r.X, r.Y).
		LineTo(r.X+r.W-fold, r.Y).
		LineTo(r.X+r.W, r.Y+fold).
		LineTo(r.X+r.W, r.Y+r.H).
		LineTo(r.X, r.Y+r.H).
		Close()
	p.FillPath(&path, col.Alpha(0.12))
	p.StrokePath(&path, 1.2, col)
}

// markCross is the small ✕ a removable chip wears.
func markCross(p *ui.Painter, r ui.Rect, col ui.Color) {
	t := r.W * 0.09
	p.Line(r.X+t, r.Y+t, r.X+r.W-t, r.Y+r.H-t, 1.4, col)
	p.Line(r.X+r.W-t, r.Y+t, r.X+t, r.Y+r.H-t, 1.4, col)
}
