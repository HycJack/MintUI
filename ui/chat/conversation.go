package chat

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/data"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// The list of conversations beside a chat window, and the small parts that go
// with it: the date between two days, the branch a turn came from, the way out
// to another window.

// ConversationItemOptions configure a ConversationItem.
type ConversationItemOptions struct {
	// Title is what the conversation is called. Required.
	Title string
	// Preview is the first line of the last turn, which is what a reader is
	// actually scanning for in a list of them.
	Preview string
	// When is when it was last said, as the caller writes it — "4m", "Tue",
	// "3 Mar" — because what a reader needs from this is how old it is and
	// only the caller knows how old it is.
	When string
	// Unread is how many turns have arrived since it was last read; zero
	// draws no count.
	Unread int
	// Pinned keeps it above the others, which is the caller's ordering rather
	// than this component's: a pinned conversation is a thing about the
	// conversation, and only the caller knows what it is pinned to.
	Pinned bool
	// Selected is the row the keys move from; -1 for none.
	Selected bool
	// Icon names the mark before the title.
	Icon string
	// Tinted colours the count, for a conversation that has something in it
	// that needs answering.
	Tinted bool
}

// ConversationItem is one row of a conversation list.
//
// It is two lines because that is what a conversation has: what it is called,
// and what was last said in it. A single line would force a choice between
// the name a reader chose and the words a reader is looking for, and the
// words are what distinguishes two conversations with the same name.
func ConversationItem(c *ui.Context, opts ConversationItemOptions) *ui.Element {
	if opts.Title == "" {
		panic("chat: ConversationItem needs a Title; a conversation with no name cannot " +
			"be told from another")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	row := ui.Column(c).FillWidth().Gap(u * 0.25).Label(opts.Title)
	if opts.Selected {
		// Only the chosen row is filled. A column of pills reads as a stack of
		// tabs rather than as a list with one place in it, which is what a
		// conversation list is.
		row = row.Padding(u, u*1.5).Radius(theme.SmallRadius).Background(k.SurfaceHover)
	}
	row.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			if opts.Icon != "" {
				display.Icon(c, display.IconName(opts.Icon), display.IconOptions{
					Name: opts.Icon, Muted: !opts.Selected, Size: u * 4,
				})
			}
			ui.Text(c, opts.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
			if opts.Pinned {
				ui.Text(c, "📌").FontSize(core.FontSize(c, theme.CaptionSize))
			}
			if opts.When != "" {
				ui.Text(c, opts.When).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
			if opts.Unread > 0 {
				bg, fg := k.Fill, k.OnFill
				if opts.Tinted {
					bg, fg = k.Accent, k.OnFill
				}
				ui.Box(c).MinWidth(u*4).Padding(0, u*1.25).Radius(theme.PillRadius).
					Background(bg).Shrink(0).Children(func() {
					ui.Text(c, internal.Commas(opts.Unread)).TextColor(fg).
						FontSize(core.FontSize(c, theme.CaptionSize))
				})
			}
		})
		if opts.Preview != "" {
			ui.Text(c, oneLine(opts.Preview)).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
		}
	})
	return row
}

// ConversationListOptions configure a ConversationList.
type ConversationListOptions struct {
	// Conversations is how many there are; Conversation builds each of them.
	Conversations int
	// Conversation builds one row. It is called only for the rows in view, so
	// a list of ten thousand conversations costs what a list of thirty does.
	Conversation func(row int)
	// Height is the list's height. It is required, and for the reason it is
	// required everywhere else in this library: a list with no height grows to
	// fit its rows, which is every row drawn rather than a list.
	Height float32
	// Selected is the row the keys move from, -1 for none.
	Selected *int
	// Choice is the set of chosen rows, for a list that chooses several.
	Choice *data.Selectable
	// Key identifies a row's record, so a row's place and its choice follow
	// the conversation rather than its number in a list that is being
	// reordered.
	Key func(row int) any
	// State is the list's place between frames; nil keeps it in the list's
	// own.
	State *ui.ListState
	// Scroll is where the list is scrolled, when the caller keeps one.
	Scroll *ui.ScrollState
	// Empty draws instead of the rows when there are none.
	Empty func()
	// Label names the list; empty takes the library's "Conversations".
	Label string
}

// ConversationList is the column of conversations beside a chat window.
//
// It is data.List and nothing of its own. That is deliberate and it is the
// point: a conversation list and a table of callbacks are the same control —
// rows of one height, chosen one at a time or several, scrolling, building
// only what is in view — and a second implementation would be a second set of
// rules about what a Shift-click means, written a year later and wrong.
func ConversationList(c *ui.Context, opts ConversationListOptions) *ui.Element {
	if opts.Conversation == nil {
		panic("chat: ConversationList needs a Conversation to build each row with")
	}
	if opts.Conversations < 0 {
		panic("chat: ConversationList cannot show a negative number of conversations")
	}

	name := opts.Label
	if name == "" {
		name = core.Msg(c, "chat.conversations", core.Def("Conversations"))
	}
	// Without a key the rows follow their own numbers. That is right while the
	// list is in the order it was handed and wrong the moment it is sorted —
	// the chosen row would stay at its index and change which conversation it
	// was — so a caller that ever reorders should pass its own key.
	key := opts.Key
	if key == nil {
		key = func(row int) any { return row }
	}

	return data.List(c, data.ListOptions{
		Rows:     opts.Conversations,
		Height:   opts.Height,
		Selected: opts.Selected,
		Choice:   opts.Choice,
		Key:      key,
		Label:    func(row int) string { return name },
		State:    opts.State,
		Scroll:   opts.Scroll,
		Empty:    opts.Empty,
	}, opts.Conversation)
}

// ConversationSearchOptions configure a ConversationSearch.
type ConversationSearchOptions struct {
	// Query is the caller's search text, written by SearchField.
	Query *string
	// Placeholder says what to search; empty takes the library's.
	Placeholder string
	// Found and Total are how many conversations match and how many there
	// are. Zero of zero draws no count.
	Found, Total int
	// Label names the field; empty takes the placeholder.
	Label string
}

// ConversationSearch is the field that filters a conversation list.
//
// It is ui/input's SearchField with the count beside it, and the count is the
// reason this is a component: "12 of 340" is the difference between a search
// that has not run yet, one that found nothing and one that found everything,
// and a field alone cannot tell the reader which of the three they are looking
// at.
func ConversationSearch(c *ui.Context, opts ConversationSearchOptions) *ui.Element {
	if opts.Query == nil {
		panic("chat: ConversationSearch needs the query to point at")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	placeholder := opts.Placeholder
	if placeholder == "" {
		placeholder = core.Msg(c, "chat.search", core.Def("Search conversations"))
	}
	name := opts.Label
	if name == "" {
		name = placeholder
	}

	e := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Label(name)
	e.Children(func() {
		input.SearchField(c, opts.Query, placeholder)
		if opts.Total > 0 {
			count := internal.Commas(opts.Found) + " of " + internal.Commas(opts.Total)
			ui.Box(c).Padding(u*0.25, u*1.5).Radius(theme.PillRadius).
				Background(k.Surface).Shrink(0).Children(func() {
				ui.Text(c, count).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
	})
	return e
}

// ConversationExportFormat is how a conversation can be written out.
type ConversationExportFormat string

const (
	// ExportMarkdown is the transcript as markdown — the format a person
	// pastes into a document, and the only one that keeps the answer's own
	// structure.
	ExportMarkdown ConversationExportFormat = "markdown"
	// ExportText is the transcript as plain text, for a reader who wants the
	// words and none of the marks.
	ExportText ConversationExportFormat = "text"
	// ExportJSON is the transcript as records, for another program to read.
	ExportJSON ConversationExportFormat = "json"
)

// String is the format's name, as the button on it says it.
func (f ConversationExportFormat) String() string {
	switch f {
	case ExportText:
		return core.Def("Text")
	case ExportJSON:
		return core.Def("JSON")
	}
	return core.Def("Markdown")
}

// Turn is one exported turn: the same four things a bubble is drawn from, plus
// the markdown the answer holds.
type Turn struct {
	Role  Role
	Who   string
	When  string
	Text  string
	Quote string
}

// Transcript is a conversation in a form that can be written out.
type Transcript struct {
	Title    string
	Turns    []Turn
	Exported ConversationExportFormat
}

// Render turns a transcript into the text of one export format.
//
// It is a pure function over the transcript and the format, for the same
// reason Parse and Highlight are: the exported text is what a reader pastes
// into a document, and it should be checkable without rendering a window. A
// test that compared the exported string against an expected string would catch
// a lost turn or a doubled heading; a screenshot could not.
func (t Transcript) Render() string {
	var b strings.Builder
	if t.Title != "" {
		switch t.Exported {
		case ExportText:
			b.WriteString(t.Title)
			b.WriteString("\n")
			b.WriteString(strings.Repeat("=", len(t.Title)))
		case ExportJSON:
			// JSON is handled whole below; nothing is written here.
		default:
			b.WriteString("# " + t.Title)
		}
		b.WriteString("\n\n")
	}

	if t.Exported == ExportJSON {
		b.WriteString("[\n")
		for i, turn := range t.Turns {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString(`  {"role": "` + turn.Role.String() + `"`)
			if turn.Who != "" {
				b.WriteString(`, "who": "` + jsonEscape(turn.Who) + `"`)
			}
			if turn.When != "" {
				b.WriteString(`, "when": "` + jsonEscape(turn.When) + `"`)
			}
			b.WriteString(`, "text": "` + jsonEscape(turn.Text) + `"}`)
		}
		b.WriteString("\n]\n")
		return b.String()
	}

	for _, turn := range t.Turns {
		who := turn.Who
		if who == "" {
			who = turn.Role.String()
		}
		switch t.Exported {
		case ExportText:
			b.WriteString("[" + turn.When + "] " + who + ": " + turn.Text + "\n\n")
		default:
			if turn.Quote != "" {
				b.WriteString("> " + turn.Quote + "\n>\n")
			}
			b.WriteString("**" + who + "**")
			if turn.When != "" {
				b.WriteString(" · " + turn.When)
			}
			b.WriteString("\n\n" + turn.Text + "\n\n")
		}
	}
	return b.String()
}

// jsonEscape is the escaping JSON needs for the characters a transcript
// actually contains. It is not the general algorithm and does not pretend to
// be: it handles the quote, the backslash and the three controls below the
// space, which is everything a title, a name and a timestamp can hold, and a
// transcript's text goes in as its own field for a program to parse properly.
func jsonEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ConversationExportOptions configure a ConversationExport.
type ConversationExportOptions struct {
	// Transcript is what is being written out.
	Transcript Transcript
	// Format is what to write. Empty takes markdown, which is the format that
	// keeps the answer's own structure.
	Format ConversationExportFormat
	// Button names the button; empty takes the format's own name.
	Button string
}

// ConversationExportResult carries a ConversationExport and its press.
type ConversationExportResult struct {
	// Element is the button.
	Element  *ui.Element
	exported bool
	// text is what was put on the clipboard.
	text string
}

// Exported reports the button being pressed this frame.
func (r ConversationExportResult) Exported() bool { return r.exported }

// Text is what was put on the clipboard, and is available on the frame it was
// exported as well as after, so a test and a caller that wants to show the
// reader what they got do not have to catch a pulse.
func (r ConversationExportResult) Text() string { return r.text }

// ConversationExport is the button that copies a whole conversation out.
//
// It is a button rather than a menu because the format is the caller's
// decision and not the reader's: a chat window exports in whatever form the
// thing it is going into wants, and offering three formats to a person who
// only wanted the conversation would be a question they cannot answer.
func ConversationExport(c *ui.Context, opts ConversationExportOptions) ConversationExportResult {
	t := opts.Transcript
	if opts.Format != "" {
		t.Exported = opts.Format
	}
	if len(t.Turns) == 0 {
		panic("chat: ConversationExport needs turns; there is no conversation in an " +
			"empty one")
	}

	label := opts.Button
	if label == "" {
		label = core.Msg(c, "chat.export", core.Def("Export")) + " " + t.Exported.String()
	}

	var res ConversationExportResult
	btn := input.Button(c, label, input.ButtonOptions{Label: label})
	if btn.Clicked() {
		res.text = t.Render()
		c.WriteClipboard(res.text)
		res.exported = true
	}
	res.Element = btn
	return res
}

// ConversationContainerOptions configure a ConversationContainer.
type ConversationContainerOptions struct {
	// SidebarWidth is how wide the list of conversations is. Zero takes the
	// library's sidebar width, which is a fixed number for the reason the
	// board's columns are: a sidebar that narrows with the window squeezes
	// the words in it until they truncate, and the words are what it is for.
	SidebarWidth float32
	// Label names the sidebar; empty takes the library's "Conversations".
	SidebarLabel string
}

// ConversationContainerResult carries a ConversationContainer and its splits.
type ConversationContainerResult struct {
	// Element is the whole container.
	Element *ui.Element
	// Collapsed reports the sidebar's collapse control being pressed.
	collapsed bool
}

// Collapsed reports the sidebar being collapsed.
func (r ConversationContainerResult) Collapsed() bool { return r.collapsed }

// ConversationContainer is a chat window with its list of conversations beside
// it.
//
// It is here rather than left to each window to assemble because the two
// halves have one hard requirement between them: the transcript must be given
// the room the sidebar is not using, and the sidebar must not be allowed to
// take it. A window that assembled the pair by hand with two Grow(1)s gave
// half the width to each, and a narrow window then squeezed a conversation's
// own column rather than the list.
func ConversationContainer(c *ui.Context, collapsed *bool, opts ConversationContainerOptions,
	sidebar, main func(),
) ConversationContainerResult {
	if collapsed == nil {
		panic("chat: ConversationContainer needs the *bool that collapses it")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	width := opts.SidebarWidth
	if width == 0 {
		width = theme.SidebarWidth
	}
	if *collapsed {
		width = theme.RailWidth
	}
	name := opts.SidebarLabel
	if name == "" {
		name = core.Msg(c, "chat.conversations", core.Def("Conversations"))
	}

	var res ConversationContainerResult
	res.Element = ui.Row(c).Fill().Children(func() {
		ui.Box(c).Width(width).FillHeight().Shrink(0).
			Background(k.Surface).Label(name).Children(func() {
			collapse := ui.ButtonBase(c).Size(u*4, u*4).Radius(theme.SmallRadius).
				Label(core.Msg(c, "chat.collapse", core.Def("Collapse conversations"))).
				TextColor(k.TextMuted)
			if collapse.Clicked() {
				res.collapsed = true
			}
			collapse.Children(func() {
				caretArrowLeft(c, *collapsed, k.TextMuted)
			})
			if sidebar != nil {
				sidebar()
			}
		})
		// The transcript takes everything else and shrinks before the sidebar
		// does, which is what BasisPercent on its own would not do.
		ui.Box(c).Grow(1).Shrink(1).FillHeight().Background(k.Background).
			Children(func() {
				if main != nil {
					main()
				}
			})
	})
	return res
}

// caretArrowLeft is the collapse control's arrow, pointing the way the
// sidebar is about to go.
func caretArrowLeft(c *ui.Context, collapsed bool, col ui.Color) {
	side := core.Density(c).Unit() * 4
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			cy := r.Y + r.H/2
			if collapsed {
				// Collapsed: the arrow points the way the sidebar will open.
				path.MoveTo(r.X+r.W*0.38, r.Y+r.H*0.22).
					LineTo(r.X+r.W*0.68, cy).LineTo(r.X+r.W*0.38, r.Y+r.H*0.78)
			} else {
				path.MoveTo(r.X+r.W*0.62, r.Y+r.H*0.22).
					LineTo(r.X+r.W*0.32, cy).LineTo(r.X+r.W*0.62, r.Y+r.H*0.78)
			}
			path.Close()
			p.FillPath(&path, col)
		})
}

// BranchNavigatorResult carries a BranchNavigator and where it moved.
type BranchNavigatorResult struct {
	// Element is the navigator.
	Element *ui.Element
	// moved is -1, 0 or 1: which way, if anywhere, this frame's press went.
	moved int
}

// Moved returns -1, 0 or 1: which branch this frame is showing, relative to
// the one before. It is a direction rather than an index because the caller
// holds the index and clamping it here would mean two places knowing where the
// ends are.
func (r BranchNavigatorResult) Moved() int { return r.moved }

// BranchNavigatorOptions configure a BranchNavigator.
type BranchNavigatorOptions struct {
	// Branch is which of how many is showing, 1-based.
	Branch, Of int
	// Previous and Next name the two buttons; empty takes the library's
	// "Previous" and "Next", which are marked up so a screen reader says what
	// they move.
	Previous, Next string
}

// BranchNavigator steps between the alternative answers to one question.
//
// It is two buttons and a count rather than a drop-down, because the branches
// are the alternative and the current one is on screen: a list that hides the
// thing being compared with the other thing is the wrong control for a choice
// between two answers that differ.
func BranchNavigator(c *ui.Context, opts BranchNavigatorOptions) BranchNavigatorResult {
	if opts.Of <= 0 {
		panic("chat: BranchNavigator needs to know how many branches there are")
	}
	if opts.Branch < 1 || opts.Branch > opts.Of {
		panic("chat: BranchNavigator is showing branch " + itoa(opts.Branch) + " of " +
			itoa(opts.Of) + ", which is not one of them")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	prev := opts.Previous
	if prev == "" {
		prev = core.Msg(c, "chat.branch.previous", core.Def("Previous answer"))
	}
	next := opts.Next
	if next == "" {
		next = core.Msg(c, "chat.branch.next", core.Def("Next answer"))
	}

	var res BranchNavigatorResult
	row := ui.Row(c).AlignItems(ui.Center).Gap(u * 0.5).Shrink(0).Label("Answer branches")
	row.Children(func() {
		// Both buttons are drawn whether or not there is anywhere to go, so
		// that stepping to the end does not shift the count sideways — which
		// is the same rule the board's columns and the list rows follow.
		back := ui.ButtonBase(c).Size(u*6, u*6).Radius(theme.SmallRadius).
			Label(prev).Tooltip(prev).TextColor(k.TextMuted).
			Disabled(opts.Branch <= 1)
		if back.Clicked() {
			res.moved = -1
		}
		back.Children(func() { branchChevron(c, k.TextMuted, false) })

		ui.Text(c, itoa(opts.Branch)+" / "+itoa(opts.Of)).TextColor(k.TextMuted).
			FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)

		forward := ui.ButtonBase(c).Size(u*6, u*6).Radius(theme.SmallRadius).
			Label(next).Tooltip(next).TextColor(k.TextMuted).
			Disabled(opts.Branch >= opts.Of)
		if forward.Clicked() {
			res.moved = 1
		}
		forward.Children(func() { branchChevron(c, k.TextMuted, true) })
	})
	res.Element = row
	return res
}

// branchChevron is the navigator's arrow, drawn rather than typed for the
// reason every mark in this library is drawn.
func branchChevron(c *ui.Context, col ui.Color, right bool) {
	side := core.Density(c).Unit() * 4
	ui.Box(c).Size(side, side).Shrink(0).Role(ui.RoleNone).
		Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			y := r.Y + r.H/2
			if right {
				path.MoveTo(r.X+r.W*0.38, y-r.H*0.2).LineTo(r.X+r.W*0.66, y).LineTo(r.X+r.W*0.38, y+r.H*0.2)
			} else {
				path.MoveTo(r.X+r.W*0.62, y-r.H*0.2).LineTo(r.X+r.W*0.34, y).LineTo(r.X+r.W*0.62, y+r.H*0.2)
			}
			p.StrokePath(&path, 1.5, col)
		})
}

// DateSeparator is the line between two days in a transcript.
//
// It is centred and quiet, and it carries the date the caller gives it rather
// than one of its own: "Today" and "Yesterday" are relative and only the
// caller knows what today is, and a separator that computed it would disagree
// with the rest of the window about which day it is the moment either of them
// crossed midnight.
func DateSeparator(c *ui.Context, label string) *ui.Element {
	if label == "" {
		panic("chat: DateSeparator needs a label; a rule with nothing on it does not " +
			"say which day it divides")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	e := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*2).
		Padding(u, 0).Label(label)
	e.Children(func() {
		// The rule is split in two so the date sits in the middle of the
		// transcript rather than at the end of a line beside it, which is
		// where a hairline with a label at one end puts every reader's eye
		// and then nowhere in particular.
		ui.Box(c).Grow(1).Height(theme.BorderWidth).Shrink(0).Background(k.Border)
		ui.Box(c).Padding(u*0.25, u*1.5).Radius(theme.PillRadius).
			Background(k.Surface).Shrink(0).Children(func() {
			ui.Text(c, label).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
		ui.Box(c).Grow(1).Height(theme.BorderWidth).Shrink(0).Background(k.Border)
	})
	return e
}

// DateSeparatorWhen draws a DateSeparator only when the turn's date differs
// from the one before it.
//
// The comparison is the whole component. A caller that put the test in its own
// loop would write it once correctly and once in a hurry, and the second one
// would put a rule above every turn that had no rule above it — which is the
// most common way a transcript comes to look like it has far more days in it
// than it does.
func DateSeparatorWhen(c *ui.Context, when, last *string) *ui.Element {
	if when == nil || last == nil {
		panic("chat: DateSeparatorWhen needs both dates to point at; the point of it " +
			"is that one is a copy of the other")
	}
	if *last == *when {
		return nil
	}
	*last = *when
	return DateSeparator(c, *when)
}

// RegenerateResult carries a RegenerateMenu and what was chosen from it.
type RegenerateResult struct {
	// Element is the trigger.
	Element *ui.Element
	// choice is the name of the item chosen, empty for none.
	choice string
}

// Chosen is the name of the item chosen this frame, empty when none was.
func (r RegenerateResult) Chosen() string { return r.choice }

// RegenerateOptions configure a RegenerateMenu.
type RegenerateOptions struct {
	// Label names the trigger; empty takes the library's "Try again".
	Label string
	// Items are the ways to ask again, each as "name: what it does". The
	// first is what the trigger itself does, so a trigger with no items is a
	// button with no action and panics rather than opening an empty menu.
	Items []string
}

// RegenerateMenu is the menu of ways to ask the same question again.
//
// It is a menu and not a button because the alternatives are different
// requests, not different words: "retry", "retry with more steps" and "retry
// from the last good point" send different messages and cost different
// amounts, and a reader who cannot see that difference should not be able to
// pick one by accident.
func RegenerateMenu(c *ui.Context, opts RegenerateOptions) RegenerateResult {
	if len(opts.Items) == 0 {
		panic("chat: RegenerateMenu needs at least one way to ask again")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	label := opts.Label
	if label == "" {
		label = core.Msg(c, "chat.regenerate", core.Def("Try again"))
	}

	var res RegenerateResult
	trigger := ui.ButtonBase(c).AlignItems(ui.Center).Gap(u).
		Radius(theme.PillRadius).Padding(u*0.75, u*2.5).
		Background(k.Surface).TextColor(k.TextMuted).Label(label).Tooltip(label)
	trigger.Children(func() {
		display.Icon(c, display.IconRefresh, display.IconOptions{
			Name: label, Muted: true, Size: u * 3.5,
		})
		ui.Text(c, label).TextColor(k.TextMuted).FontSize(core.FontSize(c, theme.RowSize))
	})
	res.Element = trigger.Menu(func(m *ui.Menu) {
		for _, line := range opts.Items {
			name, hint, _ := strings.Cut(line, ":")
			name, hint = strings.TrimSpace(name), strings.TrimSpace(hint)
			if hint == "" {
				hint = name
			}
			if m.Item(name + " — " + hint).Chosen() {
				res.choice = name
			}
		}
	})
	return res
}
