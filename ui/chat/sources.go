package chat

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/theme"
)

// Where an answer came from, and what is in the window it came with.

// Source is one thing an answer is drawn from: a document, a file, a passage
// of a web page.
type Source struct {
	// Title is what it is called. Required — a source card with no title is
	// a card with no subject.
	Title string
	// Snippet is the passage that was used, which is what makes a citation
	// worth following rather than merely countable.
	Snippet string
	// Where says where it lives — a path, a domain, a page. Empty omits the
	// line rather than printing a bare separator.
	Where string
	// Score is how well it matched, 0 to 1. Zero draws no score: most
	// retrievers have no such number for most of what they return, and a
	// fabricated 0.0 reads as a judgement.
	Score float32
	// Icon names the mark beside the title.
	Icon string
}

// CitationBadgeOptions configure a CitationBadge.
type CitationBadgeOptions struct {
	// N is which citation this is, 1-based. A citation numbered 0 does not
	// exist — the marks in a transcript start at [1] — so it panics rather
	// than drawing a [0] that points at nothing.
	N int
	// Count is how many there are in all, for "[2 of 5]". Zero shows just
	// the number.
	Count int
	// Selected is the caller's flag: which citation the reader is looking at.
	Selected bool
}

// CitationBadgeResult carries a CitationBadge and its press.
type CitationBadgeResult struct {
	// Element is the badge.
	Element *ui.Element
	pressed bool
}

// Pressed reports the badge being pressed.
func (r CitationBadgeResult) Pressed() bool { return r.pressed }

// CitationBadge is the little numbered mark a claim in an answer hangs on.
//
// It is a button and not a decoration because the number is an index into a
// list the reader cannot see from where it sits: pressing it is how they get
// to the passage, and a mark that is not pressable is one they have to count
// their way through by eye.
func CitationBadge(c *ui.Context, opts CitationBadgeOptions) CitationBadgeResult {
	if opts.N < 1 {
		panic("chat: CitationBadge numbers start at 1; a [0] points at nothing")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	label := "[" + itoa(opts.N) + "]"
	name := label
	if opts.Count > 0 {
		name = label + " of " + itoa(opts.Count)
	}

	bg, fg := k.Surface, k.TextMuted
	if opts.Selected {
		bg, fg = k.AccentBg, k.AccentText
	}

	var res CitationBadgeResult
	badge := ui.ButtonBase(c).Radius(theme.SmallRadius).Padding(u*0.5, u*1.25).
		Background(bg).Label(name).Tooltip(name).TextColor(fg).
		Children(func() {
			ui.Text(c, name).TextColor(fg).FontSize(core.FontSize(c, theme.CaptionSize))
		})
	if badge.Clicked() {
		res.pressed = true
	}
	res.Element = badge
	return res
}

// SourceCardResult carries a SourceCard and the two presses on it.
type SourceCardResult struct {
	// Element is the card.
	Element *ui.Element
	opened  bool
	quoted  bool
}

// Opened reports the card being pressed, which is how a citation opens the
// thing it cites.
func (r SourceCardResult) Opened() bool { return r.opened }

// Quoted reports the quote button being pressed, which asks for the passage to
// go back into the composer as context.
//
// It is a separate result from Opened because the two promises are different:
// one navigates away from the answer and one brings a piece of it back into
// the next question. A caller handed one flag would have to work out which
// happened from what it had already done.
func (r SourceCardResult) Quoted() bool { return r.quoted }

// SourceCardOptions configure a SourceCard.
type SourceCardOptions struct {
	// N is the citation number, 1-based; zero leaves it off the card, which
	// is right for a source the reader reached some other way.
	N int
	// Title is what the source is called. Required.
	Title string
	// Snippet is the passage that was used.
	Snippet string
	// Where says where it lives.
	Where string
	// Icon names the mark beside the title.
	Icon string
	// Quotation is what the quote button says; empty draws no button.
	Quotation string
	// Selected is the caller's flag.
	Selected bool
	// Lines is how many lines of the snippet to show; zero takes three.
	Lines int
}

// SourceCard is one source an answer used, with the passage it used.
//
// It is a card rather than a row because a snippet is three or four lines and
// a row that grows to fit them stops being a row: in a panel of sources every
// card is as tall as its longest snippet, and the ones with long snippets are
// the ones the reader least needs to see whole.
func SourceCard(c *ui.Context, opts SourceCardOptions) SourceCardResult {
	if opts.Title == "" {
		panic("chat: SourceCard needs a Title; a source with no name cannot be cited")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	lines := opts.Lines
	if lines <= 0 {
		lines = 3
	}

	var res SourceCardResult
	card := ui.Box(c).FillWidth().Radius(theme.SmallRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border).
		Label(opts.Title)
	if opts.Selected {
		card = card.BorderColor(k.Accent)
	}

	card.Children(func() {
		open := ui.ButtonBase(c).FillWidth().AlignItems(ui.Start).Gap(u*1.5).
			Padding(u*1.5, u*2).Radius(theme.SmallRadius).
			Label(opts.Title).TextColor(k.Text)
		if open.Clicked() {
			res.opened = true
		}
		open.Children(func() {
			// The citation number is on the card as well as in the panel, so a
			// card that has been pulled out of the panel still says which
			// mark in the text it belongs to.
			if opts.N > 0 {
				ui.Box(c).Padding(u*0.5, u*1.25).Radius(theme.SmallRadius).
					Background(k.AccentBg).Shrink(0).Children(func() {
					ui.Text(c, "["+itoa(opts.N)+"]").TextColor(k.AccentText).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
			ui.Column(c).Grow(1).AlignItems(ui.Start).Gap(u * 0.25).Children(func() {
				ui.Text(c, opts.Title).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
				if opts.Where != "" {
					ui.Text(c, opts.Where).TextColor(k.TextFaint).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				}
				if opts.Snippet != "" {
					ui.Text(c, opts.Snippet).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(lines)
				}
			})
		})
		if opts.Quotation != "" {
			// The button is named after the source it belongs to, and this is
			// not decoration: a panel of sources with a button on each of them
			// all called "Quote" gives a screen reader and a test the same
			// name twice, and both then act on whichever came first rather
			// than on the one the reader was looking at.
			quoted := opts.Quotation + " " + opts.Title
			ui.Row(c).FillWidth().Justify(ui.End).Padding(0, u, u, u).Children(func() {
				btn := ui.ButtonBase(c).Radius(theme.PillRadius).Padding(u*0.5, u*1.75).
					Border(theme.BorderWidth, k.Border).Label(quoted).
					TextColor(k.TextMuted).Children(func() {
					ui.Text(c, opts.Quotation).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
				if btn.Clicked() {
					res.quoted = true
				}
			})
		}
	})
	res.Element = card
	return res
}

// SourcesPanelOptions configure a SourcesPanel.
type SourcesPanelOptions struct {
	// Title heads the panel; empty draws no header.
	Title string
	// Sources are the sources, in the order they were cited.
	Sources []Source
	// Selected is which citation the reader is on, 1-based; zero for none. It
	// is the caller's index because the same number is on the badge in the
	// text, and the two must be the same one.
	Selected int
	// Quotation is what every card's quote button says; empty draws none.
	Quotation string
	// Height caps the panel and scrolls what is over. Zero lets it grow,
	// which is right for a sheet and wrong for a sidebar.
	Height float32
	// Empty draws instead of the cards when there are none.
	Empty func()
}

// SourcesPanelResult carries a SourcesPanel and what the user did with it.
type SourcesPanelResult struct {
	// Element is the panel.
	Element *ui.Element
	opened  int
	quoted  int
}

// Opened is the 1-based index of the source pressed this frame, 0 for none.
func (r SourcesPanelResult) Opened() int { return r.opened }

// Quoted is the 1-based index of the source whose quote button was pressed,
// 0 for none.
func (r SourcesPanelResult) Quoted() int { return r.quoted }

// SourcesPanel is the list of everything an answer was drawn from.
//
// It is a panel rather than an inline block because a citation is a question
// asked after the fact: the reader is looking at a claim in the middle of an
// answer and wants the passage, and shoving the passage in under the claim
// would move the text they were reading.
func SourcesPanel(c *ui.Context, opts SourcesPanelOptions) SourcesPanelResult {
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res SourcesPanelResult
	panel := ui.Column(c).FillWidth().Gap(u * 1.5)
	if opts.Title != "" {
		panel.Label(opts.Title)
	}

	panel.Children(func() {
		if opts.Title != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
				ui.Text(c, opts.Title).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.RowSize))
				// The count is the reason the header is worth having: "Sources"
				// alone does not say whether the answer cited two things or two
				// hundred.
				ui.Box(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
					Background(k.Surface).Shrink(0).Children(func() {
					ui.Text(c, itoa(len(opts.Sources))).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			})
			ui.Box(c).FillWidth().Height(theme.BorderWidth).Shrink(0).Background(k.Border)
		}
		if len(opts.Sources) == 0 {
			if opts.Empty != nil {
				opts.Empty()
			} else {
				ui.Text(c, core.Msg(c, "chat.sources.none", core.Def("No sources"))).
					TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
			}
			return
		}
		cards := func() {
			for i, s := range opts.Sources {
				card := SourceCard(c, SourceCardOptions{
					N: i + 1, Title: s.Title, Snippet: s.Snippet, Where: s.Where,
					Icon: s.Icon, Quotation: opts.Quotation,
					Selected: opts.Selected == i+1,
				})
				if card.Opened() {
					res.opened = i + 1
				}
				if card.Quoted() {
					res.quoted = i + 1
				}
			}
		}
		if opts.Height > 0 {
			// The scroll view is made here rather than around the panel: a
			// panel is often in a column that already scrolls, and a scroll
			// inside a scroll is where a reader's wheel gesture goes to die.
			ui.Scroll(c).MaxHeight(opts.Height).FillWidth().Children(cards)
			return
		}
		cards()
	})
	res.Element = panel
	return res
}

// ContextChip is one piece of context attached to a question: a file, a
// selection, a folder.
type ContextChip struct {
	// Name is what it is called. Required.
	Name string
	// Detail is the line under it — a path, a line range.
	Detail string
	// Tone tints the chip's edge, for something that could not be read.
	Tone core.Severity
	// Selected is the caller's flag, for a chip whose context has been turned
	// off without removing it.
	Selected bool
}

// ContextChipsResult carries a row of context chips and what was done to them.
type ContextChipsResult struct {
	// Element is the row.
	Element *ui.Element
	removed int
	toggled int
}

// Removed is the 1-based index of the chip whose remove button was pressed, 0
// for none.
func (r ContextChipsResult) Removed() int { return r.removed }

// Toggled is the 1-based index of the chip pressed anywhere but its remove
// button, 0 for none.
func (r ContextChipsResult) Toggled() int { return r.toggled }

// ContextChips is the row of what has been attached to the next question.
//
// It wraps rather than scrolls, because a chip that has scrolled out of sight
// is a chip whose presence the reader cannot tell: a chip is a claim that the
// model will see something, and a claim the reader cannot see is worse than
// no claim at all.
func ContextChips(c *ui.Context, chips []ContextChip) ContextChipsResult {
	if len(chips) == 0 {
		panic("chat: ContextChips needs at least one chip; an empty row is a gap")
	}
	u := core.Density(c).Unit()
	var res ContextChipsResult

	row := ui.Row(c).FillWidth().Gap(u).Children(func() {
		for i, chip := range chips {
			if chip.Name == "" {
				panic("chat: ContextChips chip " + itoa(i+1) + " needs a Name")
			}
			if chipAt(c, chip, i+1, &res).Clicked() {
				res.toggled = i + 1
			}
		}
	})
	res.Element = row
	return res
}

// chipAt is one chip: its mark, its name, its detail and its remove button.
func chipAt(c *ui.Context, chip ContextChip, n int, res *ContextChipsResult) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	body := ui.Row(c).AlignItems(ui.Center).Gap(u*1.5).Shrink(0).
		Padding(u, u*1.75).Radius(theme.ControlRadius).
		Background(k.Surface).Border(theme.BorderWidth, k.Border)
	if chip.Tone != core.Neutral {
		_, fg := chip.Tone.Pair(k)
		body = body.BorderColor(fg.Alpha(0.45))
	}
	if chip.Selected {
		body = body.Background(k.SurfaceHover)
	}
	body.Label(chip.Name)
	body.Children(func() {
		ui.Box(c).Width(theme.BorderWidth * 2).Height(u * 3.5).Shrink(0).Background(k.Border)
		if chip.Detail != "" {
			ui.Column(c).AlignItems(ui.Start).Children(func() {
				ui.Text(c, chip.Name).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				ui.Text(c, chip.Detail).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
			})
		} else {
			ui.Text(c, chip.Name).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.CaptionSize))
		}
		cross := ui.Box(c).Size(u*2.5, u*2.5).Shrink(0).Role(ui.RoleNone).
			Label("Remove " + chip.Name).
			Draw(func(p *ui.Painter, r ui.Rect) { markCross(p, r, k.TextMuted) })
		if cross.Clicked() {
			res.removed = n
		}
	})
	return body
}

// ProjectListOptions configure a ProjectList.
type ProjectListOptions struct {
	// Projects are the names, in the order they should be shown.
	Projects []string
	// Selected is the project the reader is on; nil selects none.
	Selected *string
	// Height is the list's height. Required for the same reason every other
	// list in this library requires one: a list with no height is every row
	// drawn rather than a list.
	Height float32
	// Query filters the list by name, ignoring case. Empty shows everything.
	Query string
	// Empty draws instead of the rows when nothing matches.
	Empty func()
	// Label names the list; empty takes the library's "Projects".
	Label string
}

// ProjectListResult carries a ProjectList and the project chosen this frame.
type ProjectListResult struct {
	// Element is the list.
	Element *ui.Element
	chosen  string
}

// Chosen is the project pressed this frame, empty when none was. It is a
// one-frame pulse like every other press in this library.
func (r ProjectListResult) Chosen() string { return r.chosen }

// ProjectList is the list of projects a chat window can be scoped to.
//
// It is a column of rows rather than a select, because a chat window's project
// is not a field in a form: it is a place, and switching to it keeps the
// conversation you were having. A dropdown would put a wall between the reader
// and the thing they are already in.
func ProjectList(c *ui.Context, opts ProjectListOptions) ProjectListResult {
	if opts.Height <= 0 {
		panic("chat: ProjectList needs a Height; a list with no height is every row " +
			"drawn rather than a list")
	}
	if len(opts.Projects) == 0 {
		panic("chat: ProjectList needs at least one project")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.projects", core.Def("Projects"))
	}

	var shown []string
	for _, p := range opts.Projects {
		if q := strings.ToLower(strings.TrimSpace(opts.Query)); q != "" &&
			!strings.Contains(strings.ToLower(p), q) {
			continue
		}
		shown = append(shown, p)
	}

	var res ProjectListResult
	if len(shown) == 0 {
		empty := opts.Empty
		if empty == nil {
			empty = func() {
				ui.Box(c).FillWidth().Height(opts.Height).Center().Children(func() {
					ui.Text(c, core.Msg(c, "chat.projects.none", core.Def("No projects"))).
						TextColor(k.TextFaint).FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
		}
		empty()
		res.Element = ui.Box(c).FillWidth().Label(name)
		return res
	}

	rows := func() {
		for _, p := range shown {
			selected := opts.Selected != nil && *opts.Selected == p
			row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
				Padding(u, u*2).Radius(theme.SmallRadius).
				Label(p).Cursor(ui.CursorPointer)
			if selected {
				// Only the chosen row is filled. A column of pills would read
				// as a stack of tabs rather than as a list with one place in
				// it, which is what a project list is.
				row = row.Background(k.SurfaceHover)
			}
			if row.Clicked() {
				res.chosen = p
			}
			row.Children(func() {
				ink := k.TextMuted
				if selected {
					ink = k.Text
				}
				if strings.Contains(p, "/") {
					// A path gets a folder mark and a bare name does not, so a
					// list of paths reads differently from a list of names
					// without a word of copy saying so.
					display.Icon(c, display.IconInventory, display.IconOptions{
						Name: "project", Muted: !selected, Size: u * 4,
					})
				}
				ui.Text(c, p).TextColor(ink).
					FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
			})
		}
	}

	// The scroll is built last and given the rows, rather than the rows being
	// built first and hung inside it: a child is whatever was CREATED while a
	// Children call was running, so a column made out here would be the
	// scroll's sibling rather than its content.
	scroll := ui.Scroll(c).Height(opts.Height).FillWidth()
	scroll.Children(rows)
	res.Element = scroll.Label(name)
	return res
}
