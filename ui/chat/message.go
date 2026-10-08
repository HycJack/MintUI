package chat

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/display"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/layout"
	"github.com/HycJack/MintUI/ui/theme"
)

// A turn in a transcript, and the column of them.

// MessageHeaderOptions configure a MessageHeader.
type MessageHeaderOptions struct {
	// Author is who is speaking. It is required for the roles that have a
	// speaker — user and assistant — and empty for the two that do not, which
	// is why Role is here: the header is what decides whether there is a name
	// to show.
	Author string
	// Time is when it was said, for the roles that are dated.
	Time string
	// Role is the turn's role. It decides whether the header is drawn at all:
	// a tool turn has nobody to name, so its header would be a line of
	// nothing above the bubble.
	Role Role
	// Badge is a short mark after the name — a model, a branch.
	Badge string
}

// MessageHeader is who is speaking and when: the line above a bubble.
//
// It is aligned to the same side as the bubble it belongs to, because a header
// on the left above a bubble on the right reads as a caption for somebody
// else's message. That is the whole reason it is a component rather than two
// Text calls in the caller's column.
func MessageHeader(c *ui.Context, opts MessageHeaderOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	if opts.Author == "" && opts.Time == "" && opts.Badge == "" {
		// A header with nothing in it is a blank line the size of a header.
		// Drawing the box and no content would space every tool turn as if it
		// were a message, which is exactly the rhythm the single bubble style
		// exists to keep.
		return nil
	}

	name := opts.Author
	if name == "" {
		name = opts.Role.String()
	}
	head := ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).
		AlignItems(bubbleSide(opts.Role)).Label(name)
	head.Children(func() {
		if opts.Author != "" {
			ui.Text(c, opts.Author).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).Shrink(0)
		}
		if opts.Badge != "" {
			ui.Box(c).Padding(u*0.25, u*1.25).Radius(theme.PillRadius).
				Background(k.Surface).Shrink(0).Children(func() {
				ui.Text(c, opts.Badge).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
		}
		if opts.Time != "" {
			ui.Text(c, opts.Time).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		}
	})
	return head
}

// QuoteReplyOptions configure a QuoteReply.
type QuoteReplyOptions struct {
	// Author is whose words are being quoted; empty quotes the speaker the
	// turn is replying to, which the caller passes as Author.
	Author string
	// Text is what is being quoted. It is truncated rather than shown whole:
	// a quote of a quote of an answer is not a thing anybody reads, and the
	// full text is one scroll away in the transcript above.
	Text string
	// Lines is how many lines of the quote to show. Zero takes one.
	Lines int
	// Side is the column the reply is on, so the rule hangs off the right
	// edge of a reply on the right. Zero, ui.Start.
	Side ui.Align
	// Label names the quote; empty takes the author and the first words.
	Label string
}

// QuoteReply is the lines of an earlier turn shown above a reply to it.
//
// It is the one place a transcript shows one bubble inside another, and it is
// deliberately not built with bubble(): a quote is a reference, not a turn —
// it has no tail, no padding of its own and no identity — so giving it the
// turn's box would make it look like a message that was sent.
func QuoteReply(c *ui.Context, opts QuoteReplyOptions) *ui.Element {
	if opts.Text == "" {
		panic("chat: QuoteReply needs Text; an empty quote is a rule with nothing " +
			"beside it")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	lines := opts.Lines
	if lines <= 0 {
		lines = 1
	}
	label := opts.Label
	if label == "" {
		label = opts.Author
		if label == "" {
			label = "quoted message"
		} else {
			label += ": " + oneLine(opts.Text)
		}
	}

	// A Row with its own alignment, so the rule is on the edge the reply is
	// on and not always on the left.
	side := opts.Side
	if side != ui.End {
		side = ui.Start
	}
	return ui.Row(c).FillWidth().AlignItems(ui.Stretch).Gap(u * 1.5).Children(func() {
		ui.Box(c).Width(theme.BorderWidth * 3).Shrink(0).
			Background(k.Border).AlignSelf(side)
		ui.Column(c).Grow(1).Gap(u * 0.5).Label(label).Children(func() {
			if opts.Author != "" {
				ui.Text(c, opts.Author).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(1)
			}
			ui.Text(c, opts.Text).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).MaxLines(lines)
		})
	})
}

// oneLine is a string with its newlines turned into spaces, for a label that
// is read out in one breath.
func oneLine(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\n' || r == '\r' {
			r = ' '
		}
		out = append(out, r)
	}
	if len(out) > 60 {
		return string(out[:60]) + "…"
	}
	return string(out)
}

// MessageAction is one thing that can be done to a turn.
type MessageAction struct {
	// Label is both what the mark is called out loud and what comes back in
	// MessageActionsResult.Pressed. It is required: a glyph with no name is a
	// shape, and a caller reading Pressed needs a value it can switch on.
	Label string
	// Icon is the mark.
	Icon display.IconName
	// Active fills the button, for an action that is on — a reaction already
	// given, a branch already shown.
	Active bool
	// Disabled greys the action out: it takes neither clicks nor focus.
	Disabled bool
	// Tone tints the mark, for an action that is destructive.
	Tone core.Severity
}

// MessageActionsResult carries a row of actions and the one that was pressed.
type MessageActionsResult struct {
	// Element is the row.
	Element *ui.Element
	pressed string
}

// Pressed is the label of the action pressed this frame, empty when none was.
// It is the whole result of a row of buttons: one press, reported by name, and
// no callback — a caller that wants four different handlers ends up with a
// switch in one place instead of four closures that fire in a frame nobody
// predicted.
func (r MessageActionsResult) Pressed() string { return r.pressed }

// MessageActions is the row of things you can do to a turn: copy it, retry it,
// like it.
//
// It is hidden until the turn is chosen or hovered over, by the caller, who
// holds the state. A row of five grey marks under every message is a row of
// five grey marks under every message.
func MessageActions(c *ui.Context, actions []MessageAction) MessageActionsResult {
	if len(actions) == 0 {
		panic("chat: MessageActions needs at least one action; an empty row is a gap")
	}
	for i, a := range actions {
		if a.Label == "" {
			panic("chat: MessageActions action " + internal.Commas(i) + " needs a Label; " +
				"a mark with no name is a shape, and Pressed has nothing to return")
		}
	}

	k, u := core.Tokens(c), core.Density(c).Unit()
	var res MessageActionsResult

	row := ui.Row(c).AlignItems(ui.Center).Gap(u * 0.5).Role(ui.RoleGroup)
	row.Children(func() {
		for _, a := range actions {
			side := u * 3.25
			btn := ui.ButtonBase(c).Size(side, side).Radius(theme.PillRadius).
				Label(a.Label).Tooltip(a.Label).Disabled(a.Disabled)
			if a.Active {
				btn = btn.Background(k.Surface)
			}
			if btn.Clicked() {
				res.pressed = a.Label
			}
			// The mark is built inside the button's own Children: an element
			// belongs to whatever was being built when it was made, so a glyph
			// built out here would land as the button's next sibling and sit
			// beside it instead of inside it.
			btn.Children(func() {
				display.Icon(c, a.Icon, display.IconOptions{
					Name: a.Label, Muted: !a.Active, Tone: a.Tone, Size: side * 0.62,
				})
			})
		}
	})
	res.Element = row
	return res
}

// MessageEditorOptions configure a MessageEditor.
type MessageEditorOptions struct {
	// Role is the side the edited turn sits on. It is read from the caller's
	// flag rather than asked: an editor that guessed would put a cancel button
	// on the wrong side of the answer it is about to rewrite.
	Role Role
	// Label names the field; it is required, as for every field.
	Label string
	// Save and Cancel name the two buttons. Empty takes the library's "Save"
	// and "Cancel", which is right for editing a message and wrong for
	// anything that is not.
	Save, Cancel string
	// Lines is how tall the field is. Four is the usual answer.
	Lines int
}

// MessageEditorResult carries a MessageEditor and what the user did to it.
type MessageEditorResult struct {
	// Element is the editor.
	Element *ui.Element
	saved   bool
	cancel  bool
}

// Saved reports the save button being pressed this frame. The caller already
// has the text — it is the string the editor was handed — so this is a
// question, not a handover.
func (r MessageEditorResult) Saved() bool { return r.saved }

// Cancelled reports the cancel button being pressed.
func (r MessageEditorResult) Cancelled() bool { return r.cancel }

// MessageEditor rewrites a turn in place: the bubble's box, with a field in it
// where the text was.
//
// It draws the editor with bubble() rather than a box of its own, so an
// editing turn occupies exactly the space the turn it replaces did. That is
// the reason it is worth a component at all — a composer that grew a new shape
// while editing would move the transcript under the reader's hands.
func MessageEditor(c *ui.Context, text *string, opts MessageEditorOptions) MessageEditorResult {
	if text == nil {
		panic("chat: MessageEditor needs the text to point at; it keeps none of its own")
	}
	if opts.Label == "" {
		panic("chat: MessageEditor needs a Label")
	}
	u := core.Density(c).Unit()

	save := opts.Save
	if save == "" {
		save = core.Msg(c, "chat.editor.save", core.Def("Save"))
	}
	cancel := opts.Cancel
	if cancel == "" {
		cancel = core.Msg(c, "chat.editor.cancel", core.Def("Cancel"))
	}

	var res MessageEditorResult
	wrap := ui.Column(c).FillWidth().AlignItems(bubbleSide(opts.Role)).Gap(u)
	wrap.Children(func() {
		bubble(c, bubbleSkin{Role: opts.Role, Label: opts.Label}, func() {
			input.TextArea(c, text, input.TextAreaOptions{
				Label: opts.Label, Lines: opts.Lines,
			})
		})
		ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
			saveBtn := input.Button(c, save, input.ButtonOptions{Primary: true})
			if saveBtn.Clicked() {
				res.saved = true
			}
			if input.Button(c, cancel, input.ButtonOptions{}).Clicked() {
				res.cancel = true
			}
		})
	})
	res.Element = wrap
	return res
}

// MessageListOptions configure a MessageList.
type MessageListOptions struct {
	// Messages is how many turns there are; Message builds each of them.
	Messages int
	// Message builds one turn. It is called only for the turns the caller
	// gives it, which for a transcript is all of them — a transcript is short
	// by nature, and virtualising it would mean the reader could not scroll
	// to a turn by the keyboard.
	Message func(i int)
	// Height is the viewport's height. It is required, as for any scroll area
	// that cannot scroll: a transcript with no height grows to fit and is a
	// very long column rather than a conversation.
	Height float32
	// Gap is the space between turns. Zero takes the bubbles' own, which is
	// the only correct answer for a column of bubbles and the reason the
	// caller almost never passes one.
	Gap float32
	// State is where the scroll position lives. Give one to keep a place
	// across redraws, to come back to where a reader was, or to follow the
	// end while an answer streams in.
	State *ui.ScrollState
	// Top draws above the turns and Bottom below them, inside the scroll, so
	// that a date separator or a composer pinned to the end scrolls with the
	// transcript rather than floating over it.
	Top, Bottom func()
	// Empty draws instead of the turns when there are none.
	Empty func()
	// Label names the transcript for assistive technology.
	Label string
}

// MessageListResult carries a MessageList and where it is scrolled to.
type MessageListResult struct {
	// Element is the transcript.
	Element *ui.Element
	// State is the scroll position, nil when the caller gave none.
	State *ui.ScrollState
}

// AtEnd reports whether the transcript is scrolled to its bottom, which is how
// a composer decides between following a streaming answer and leaving the
// reader where they are. It is true when there is no state, because a
// transcript nobody is scrolling is being read from the top and following it
// is the safe default.
func (r MessageListResult) AtEnd() bool {
	if r.State == nil {
		return true
	}
	return r.State.Y >= r.State.MaxY
}

// MessageList is the column of turns a conversation is read in.
//
// It is a scroll area and nothing more. It does not decide where the reader is
// — that is the caller's state — and it does not decide when to follow a new
// answer, because that depends on whether the reader was already at the bottom
// and only the caller knows. AtEnd is how it says so.
func MessageList(c *ui.Context, opts MessageListOptions) MessageListResult {
	if opts.Message == nil {
		panic("chat: MessageList needs a Message to build each turn with")
	}
	if opts.Messages < 0 {
		panic("chat: MessageList cannot show a negative number of messages")
	}
	u := core.Density(c).Unit()
	m := bubbleMetricsFor(u)

	gap := opts.Gap
	if gap <= 0 {
		gap = m.Gap
	}
	name := opts.Label
	if name == "" {
		name = "conversation"
	}

	scroll := layout.ScrollArea(c, layout.ScrollAreaOptions{
		Vertical: true, Height: opts.Height, State: opts.State, Pad: m.PadX,
	}, func() {
		if opts.Empty != nil && opts.Messages == 0 {
			opts.Empty()
			return
		}
		ui.Column(c).FillWidth().Gap(gap).Children(func() {
			if opts.Top != nil {
				opts.Top()
			}
			for i := range opts.Messages {
				opts.Message(i)
			}
			if opts.Bottom != nil {
				opts.Bottom()
			}
		})
	})
	scroll.Element.Label(name)
	return MessageListResult{Element: scroll.Element, State: opts.State}
}
