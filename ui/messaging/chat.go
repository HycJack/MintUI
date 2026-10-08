package messaging

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chat"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/overlay"
	"github.com/HycJack/MintUI/ui/theme"
)

// Channel is one conversation in the sidebar: a direct message, a group, a
// thread's parent.
type Channel struct {
	// ID identifies the channel and is required: the caller's list is going to
	// be reordered, filtered and reloaded, and a row the caller cannot
	// recognise afterwards is a row it cannot act on.
	ID string
	// Name is what the reader sees.
	Name string
	// Topic sits under the name — what the channel is for.
	Topic string
	// When is when anything last happened in it, as the caller writes it.
	When string
	// Unread is how many messages have arrived since it was last read.
	Unread int
	// Mentions is how many of those mention the reader. It is separate from
	// Unread because a channel with one mention in it is a different thing
	// from a channel with one message in it, and a badge cannot say both.
	Mentions int
	// Draft is what is half-written in it. A channel with a draft is not
	// quiet, and a reader who has written a message and lost it cannot get it
	// back from anywhere else.
	Draft string
	// Muted stops it making a sound or a badge. It is still in the list: a
	// muted channel that disappeared is a channel the reader has to go
	// looking for.
	Muted bool
	// Pinned keeps it above the others.
	Pinned bool
}

// ChannelListOptions configure a ChannelList.
type ChannelListOptions struct {
	// Channels are the caller's, in the order they should be shown: this
	// component does not sort them, because "pinned first, then by recency"
	// is a rule about the product and not about a list of rows.
	Channels []Channel
	// Selected is the channel the keys move from, as an index into Channels;
	// -1 for none.
	Selected *int
	// State and Scroll keep the list's place and its sideways scroll between
	// frames. Each list keeps its own — MyGo panics if one ListState is given
	// to two lists in one frame.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Height is the list's height. It is required: a list with no height grows
	// to fit its rows, which is every row drawn rather than a list.
	Height float32
	// Query filters the list by name and topic as the reader types, and is the
	// caller's because the composer beside it is. Empty shows everything.
	Query *string
	// Label names the list; empty takes the library's "Channels".
	Label string
}

// ChannelListResult carries a ChannelList.
type ChannelListResult struct {
	// Element is the list.
	Element *ui.Element
	// shown is how many rows survived the query, which is what an empty state
	// has to be able to say.
	shown int
}

// Shown is how many channels passed the query. A list that filtered everything
// away and then drew its own "no channels" state under a heading saying
// "Channels" is telling the reader they have no channels, when what happened
// is that they typed something that matched none of them.
func (r ChannelListResult) Shown() int { return r.shown }

// ChannelList is the column of conversations beside a window.
//
// It is [chat.ConversationList] called through, and its rows are
// [chat.ConversationItem]: the two-line row, the unread pill and the pinned
// mark are the same in a direct message and in a channel, and a second
// implementation would be a second set of decisions about what a row in this
// application looks like.
//
// What this adds on top is the two things a channel has and a conversation
// does not: mentions, which are counted separately from unread because a
// mention is a different kind of new, and a draft marker, which is the only
// record anywhere that somebody half-wrote something and lost it.
func ChannelList(c *ui.Context, opts ChannelListOptions) ChannelListResult {
	if opts.Selected == nil {
		panic("messaging: ChannelList needs a Selected channel to point at; it owns no list " +
			"of its own")
	}
	if opts.Height <= 0 {
		panic("messaging: ChannelList needs a Height; a list with no height grows to fit " +
			"every row rather than scrolling")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	name := opts.Label
	if name == "" {
		name = core.Msg(c, "messaging.channels", core.Def("Channels"))
	}

	shown := make([]int, 0, len(opts.Channels))
	for i, ch := range opts.Channels {
		if matches(ch, opts.Query) {
			shown = append(shown, i)
		}
	}

	res := ChannelListResult{shown: len(shown)}
	// data.List names its rows and nothing else, so the column as a whole is
	// wrapped in a named box. A list whose rows announce their own names and
	// which has no name of its own reads as a set of unrelated rows.
	res.Element = ui.Box(c).FillWidth().Label(name).Role(ui.RoleList).
		Children(func() {
			chat.ConversationList(c, chat.ConversationListOptions{
				Conversations: len(shown),
				Height:        opts.Height,
				Selected:      opts.Selected,
				Key:           func(row int) any { return opts.Channels[shown[row]].ID },
				State:         opts.State,
				Scroll:        opts.Scroll,
				Label:         name,
				Conversation: func(row int) {
					ch := opts.Channels[shown[row]]
					chat.ConversationItem(c, chat.ConversationItemOptions{
						Title:    ch.Name,
						Preview:  preview(ch),
						When:     ch.When,
						Unread:   ch.Unread,
						Pinned:   ch.Pinned,
						Selected: *opts.Selected == shown[row],
						Tinted:   ch.Mentions > 0,
					})
					// The decorations are drawn under the row rather than inside the
					// conversation item's own children, which are not the caller's to
					// reach into: a mentions count and a muted marker are this
					// package's business and the item's two lines are its own.
					layoutRow(c, ch)
				},
				Empty: func() {
					ui.Column(c).FillWidth().Center().Padding(u * 3).
						Label("No channels").Children(func() {
						ui.Text(c, "No channels here").TextColor(k.TextMuted).
							FontSize(core.FontSize(c, theme.BodySize))
					})
				},
			})
		})
	return res
}

// matches is the query rule: a case-insensitive substring of the name or the
// topic. The topic is included because a channel is routinely found by what it
// is for — "who is in the billing channel" is answered by a topic line far
// more often than by a name anybody remembers.
func matches(ch Channel, query *string) bool {
	if query == nil || *query == "" {
		return true
	}
	needle := lower(*query)
	return contains(lower(ch.Name), needle) || contains(lower(ch.Topic), needle)
}

// preview is the row's second line: what the channel most needs to say. A
// draft outranks a topic, because a draft is something the reader was in the
// middle of doing and a topic is something somebody else wrote.
func preview(ch Channel) string {
	if ch.Draft != "" {
		return "Draft: " + ch.Draft
	}
	return ch.Topic
}

// layoutRow draws the decorations under a channel row: a mentions count and
// the muted marker. The row's own two lines are the conversation item's, and
// they stay above these rather than beside them.
func layoutRow(c *ui.Context, ch Channel) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Column(c).FillWidth().Gap(u * 0.5).Children(func() {
		if ch.Mentions > 0 {
			// The count is written out rather than badged: a badge is for a
			// number on a picture, and a badge on a two-line row with a pill
			// already at its right is a second number on the same line.
			ui.Text(c, mentionsWord(ch.Mentions)).TextColor(k.AccentText).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		}
		if ch.Muted {
			ui.Text(c, "Muted").TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		}
	})
}

func mentionsWord(n int) string {
	if n == 1 {
		return "mentions you"
	}
	return "mentions you " + itoa(n) + " times"
}

func lower(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		out = append(out, r)
	}
	return string(out)
}

// has is the other direction of [contains]: whether the slice holds the value.
// It is separate rather than a flipped call because a membership test that
// reads "contains(haystack, disallowed)" is exactly the kind of call that gets
// its arguments the right way round by accident.
func has(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func contains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// ── one message ────────────────────────────────────────────────────────────

// ChatMessageOptions configure a ChatMessage.
type ChatMessageOptions struct {
	// Role is whose turn this is. It maps straight onto [chat.Role], so a
	// channel and a transcript make the same three decisions about ink and
	// tail and there is no fourth.
	Role chat.Role
	// Author and Time are who and when, as the caller's strings.
	Author string
	Time   string
	// Body is the message's content. It is the caller's, because what a turn
	// contains is not the box's business: markdown, an attachment, a table, a
	// candidate's waveform.
	Body func()
	// Quoted is the turn this answers, drawn above the bubble. Setting it is
	// the same as calling [QuotedText] by hand; this field is here so a list
	// of messages can be built from records rather than from closures.
	Quoted string
	// QuotedAuthor is whose words Quoted holds.
	QuotedAuthor string
	// Receipt is the delivery state, drawn under the reader's own messages.
	// It is ignored for anyone else's, because a receipt on somebody else's
	// message is information the reader cannot act on.
	Receipt Receipt
	// ReceiptTime is when it got there.
	ReceiptTime string
	// Selected marks the turn, and is the caller's flag: two views of one
	// channel must agree about which turn is chosen.
	Selected bool
	// Divider draws the "new messages" line above this turn. The caller asks
	// [DividerAt] rather than computing it twice.
	Divider bool
	// DividerCount is how many messages are below the line.
	DividerCount int
	// ReplyTo is the number of replies, shown as a thread link under the
	// turn. Zero draws none.
	Replies int
	// Monospace, Muted and MaxWidth pass through to the bubble unchanged.
	Monospace bool
	Muted     bool
	MaxWidth  float32
	// Label names the turn; empty takes the bubble's own name.
	Label string
}

// ChatMessageResult carries a ChatMessage.
type ChatMessageResult struct {
	// Element is the whole turn: the divider, the header, the bubble, the
	// receipt and the thread link.
	Element *ui.Element
	// clicked reports the bubble being pressed this frame.
	clicked bool
	// threaded reports the thread link being pressed this frame.
	threaded bool
}

// Clicked reports the bubble being pressed, which is how a channel asks to
// choose a turn. Selection is the caller's flag; the press is the only thing
// a person does to a message that is not one of its own buttons.
func (r ChatMessageResult) Clicked() bool { return r.clicked }

// Threaded reports the reply link being pressed, which is how a channel asks
// to open a thread. It is separate from Clicked because the two presses mean
// different things to a caller — one selects, one navigates — and collapsing
// them would make selecting a message that has replies impossible.
func (r ChatMessageResult) Threaded() bool { return r.threaded }

// ChatMessage is one message in a channel, with the three decorations a
// transcript does not need.
//
// The bubble is [chat.MessageBubble]. This function adds the divider above,
// the receipt below and the thread link, and passes the rest straight
// through — which is the whole contract. A message drawn by this package and a
// message drawn by the chat window are the same box, because they are the same
// call; only what surrounds it differs.
func ChatMessage(c *ui.Context, opts ChatMessageOptions) ChatMessageResult {
	u := core.Density(c).Unit()

	quoted := opts.Quoted
	author := opts.QuotedAuthor
	if author == "" {
		author = opts.Author
	}

	var res ChatMessageResult
	res.Element = ui.Column(c).FillWidth().Gap(u * 0.75).Children(func() {
		if opts.Divider {
			UnreadDivider(c, UnreadDividerOptions{Count: opts.DividerCount})
		}

		// The bubble is built here rather than beside it so that it lands
		// inside this column: an element belongs to whatever container is
		// current where it is made, and one made out here would land in the
		// window instead of in the column that is supposed to hold the turn.
		bubble := chat.MessageBubble(c, chat.MessageBubbleOptions{
			Role: opts.Role, Author: opts.Author, Time: opts.Time,
			Quoted: quoted, Selected: opts.Selected,
			Monospace: opts.Monospace, Muted: opts.Muted,
			MaxWidth: opts.MaxWidth, Label: opts.Label,
		}, opts.Body)
		res.clicked = bubble.Clicked()

		// The footer hangs off the bubble's own side rather than sitting
		// under the column, so a receipt under a message on the right is on
		// the right. AlignItems on the column would put it under the widest
		// child, which for a short reply is nowhere near where it was sent.
		if sideFor(opts.Role) == ui.End {
			ui.Column(c).FillWidth().AlignItems(ui.End).Children(func() {
				messageFooter(c, opts, &res)
			})
		} else {
			messageFooter(c, opts, &res)
		}
	})
	return res
}

// footer is the receipt and the thread link, in that order for the reader's own
// messages and the other way round for everybody else's — a reply count under
// somebody else's message is the thing they are waiting for an answer to, and
// the delivery state of their message is the thing they are not.
func messageFooter(c *ui.Context, opts ChatMessageOptions, res *ChatMessageResult) {
	k, u := core.Tokens(c), core.Density(c).Unit()
	ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
		if opts.Role == chat.RoleUser {
			ReadReceipt(c, ReadReceiptOptions{
				State: opts.Receipt, Time: opts.ReceiptTime, Muted: true,
			})
		}
		if opts.Replies > 0 {
			link := ui.Box(c).Role(ui.RoleNone).Label(repliesWord(opts.Replies)).
				Cursor(ui.CursorPointer).Children(func() {
				ui.Text(c, repliesWord(opts.Replies)).TextColor(k.AccentText).
					FontSize(core.FontSize(c, theme.CaptionSize))
			})
			if link.Clicked() {
				res.threaded = true
			}
		}
	})
}

func repliesWord(n int) string {
	if n == 1 {
		return "1 reply"
	}
	return itoa(n) + " replies"
}

// sideFor is which side of the column a role's bubble sits on. It is asked
// here rather than taken from the bubble because chat keeps that rule to
// itself, and a second copy of it is exactly the drift this package exists to
// avoid.
func sideFor(role chat.Role) ui.Align {
	if role == chat.RoleUser {
		return ui.End
	}
	if role == chat.RoleSystem {
		return ui.Center
	}
	return ui.Start
}

// ── a thread ───────────────────────────────────────────────────────────────

// ThreadPanelOptions configure a ThreadPanel.
type ThreadPanelOptions struct {
	// Parent is the message the thread hangs off. It is drawn through
	// [ChatMessage] like any other turn, so a thread's parent and the message
	// it was pressed in the column are the same box.
	Parent ChatMessageOptions
	// Replies is how many replies there are; Reply builds each of them.
	Replies int
	Reply   func(i int)
	// Selected is the reply the keys move from, -1 for none.
	Selected *int
	// State is where the thread is scrolled; give one to keep a reader's
	// place, or to follow the end while somebody is still typing.
	State *ui.ScrollState
	// Height is the panel's height, and is required: a thread with no height
	// grows to fit and is a very long column rather than a thread.
	Height float32
	// Title heads the panel; empty takes the library's "Thread".
	Label string
	// Updated says when the last reply arrived, as the caller's string.
	Updated string
}

// ThreadPanelResult carries a ThreadPanel.
type ThreadPanelResult struct {
	// Element is the panel.
	Element *ui.Element
}

// ThreadPanel is the replies to one message, in their own scroll.
//
// It is [chat.MessageList] for the same reason ChannelList is
// [chat.ConversationList]: a column of turns that scrolls is a column of
// turns that scrolls, and writing one here would be a second answer to what a
// gap between bubbles is.
func ThreadPanel(c *ui.Context, opts ThreadPanelOptions) ThreadPanelResult {
	if opts.Reply == nil {
		panic("messaging: ThreadPanel needs a Reply to build each reply with")
	}
	if opts.Height <= 0 {
		panic("messaging: ThreadPanel needs a Height; a thread with no height grows to fit " +
			"and never scrolls")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	name := opts.Label
	if name == "" {
		name = core.Msg(c, "messaging.thread", core.Def("Thread"))
	}

	return ThreadPanelResult{
		Element: chat.MessageList(c, chat.MessageListOptions{
			Messages: opts.Replies,
			Message:  opts.Reply,
			Height:   opts.Height,
			State:    opts.State,
			Top: func() {
				ChatMessage(c, opts.Parent)
				ui.Box(c).FillWidth().Height(theme.BorderWidth).Background(k.Border)
				ui.Row(c).AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
					ui.Text(c, name).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize))
					if opts.Updated != "" {
						ui.Text(c, opts.Updated).TextColor(k.TextFaint).
							FontSize(core.FontSize(c, theme.CaptionSize))
					}
				})
			},
		}).Element,
	}
}

// ── pinned ─────────────────────────────────────────────────────────────────

// PinnedMessage is one message kept at the top of a channel.
type PinnedMessage struct {
	// ID identifies the message, and is required for the same reason a
	// channel's is: the list is reordered by the caller and the rows have to
	// follow their records.
	ID string
	// Author and Text are what is kept.
	Author string
	Text   string
	// When is when it was pinned, as the caller's string.
	When string
}

// PinnedMessagesOptions configure a PinnedMessages.
type PinnedMessagesOptions struct {
	// Messages are the caller's, in the order they should be shown.
	Messages []PinnedMessage
	// Selected is the message the keys move from, as an index; -1 for none.
	Selected *int
	// State and Scroll keep the list's place between frames.
	State  *ui.ListState
	Scroll *ui.ScrollState
	// Height is the list's height. Zero takes four rows, which is what a
	// popover over a channel wants; anything more belongs in its own panel.
	Height float32
	// Title heads the list; empty takes the library's "Pinned".
	Label string
	// Jumped is the message jumped to this frame, as an index into Messages,
	// or -1. It is reported rather than written, because the jump is the
	// caller's scroll to make and only the caller knows where the message
	// ended up.
	Jumped int
}

// PinnedMessagesResult carries a PinnedMessages.
type PinnedMessagesResult struct {
	// Element is the list.
	Element *ui.Element
	// jumped is the message pressed this frame, or -1.
	jumped int
}

// Jumped is the message the reader pressed, as an index into the messages it
// was given, or -1. A press reports rather than scrolls: the list is
// sometimes in a popover over a channel and sometimes in a panel beside one,
// and only the caller knows which scroll it has to move.
func (r PinnedMessagesResult) Jumped() int { return r.jumped }

// PinnedMessages is the handful of messages a channel keeps at the top: the
// rules, the address, the one everybody scrolls back to.
//
// It is a list of rows rather than a stack of bubbles on purpose. A pinned
// message is not there to be read as a turn — it has already been read, which
// is the only reason it is pinned — and a reader looking for the meeting time
// is looking at a list of titles, not at a conversation.
func PinnedMessages(c *ui.Context, opts PinnedMessagesOptions) PinnedMessagesResult {
	if opts.Selected == nil {
		panic("messaging: PinnedMessages needs a Selected message to point at; it owns no " +
			"list of its own")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	name := opts.Label
	if name == "" {
		name = core.Msg(c, "messaging.pinned", core.Def("Pinned messages"))
	}
	height := opts.Height
	if height <= 0 {
		height = u * 11 * 4
	}

	var res PinnedMessagesResult
	res.jumped = -1
	// Wrapped in a named box for the same reason ChannelList is: data.List
	// names its rows and nothing else.
	res.Element = ui.Box(c).FillWidth().Label(name).Role(ui.RoleList).Children(func() {
		chat.ConversationList(c, chat.ConversationListOptions{
			Conversations: len(opts.Messages),
			Height:        height,
			Selected:      opts.Selected,
			Key:           func(row int) any { return opts.Messages[row].ID },
			State:         opts.State,
			Scroll:        opts.Scroll,
			Label:         name,
			Conversation: func(row int) {
				m := opts.Messages[row]
				row2 := ui.Column(c).FillWidth().Gap(u*0.25).
					Padding(u, u*1.5).Radius(theme.SmallRadius).
					Label(m.Text).Role(ui.RoleNone)
				if row2.Clicked() {
					res.jumped = row
				}
				if *opts.Selected == row {
					row2.Background(k.SurfaceHover)
				}
				row2.Children(func() {
					ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
						ui.Text(c, m.Author).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).Grow(1).SingleLine()
						if m.When != "" {
							ui.Text(c, m.When).TextColor(k.TextFaint).
								FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
						}
					})
					ui.Text(c, m.Text).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				})
			},
			Empty: func() {
				ui.Text(c, "Nothing pinned").TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize))
			},
		})
	})
	return res
}

// ── profile ────────────────────────────────────────────────────────────────

// UserProfileCardOptions configure a UserProfileCard.
type UserProfileCardOptions struct {
	// Anchor is the row or the avatar the card hangs off, and is required: a
	// card with nothing to hang from has no place to be.
	Anchor *ui.Element
	// Open is the *bool the card opens and closes with, and is required for
	// the reason it is on every layer here.
	Open *bool
	// Name is who the card is about.
	Name string
	// Handle is their @name, when they have one.
	Handle string
	// Role is what they do.
	Role string
	// Presence is how reachable they are, and When since.
	Presence Presence
	Since    string
	// LocalTime is what the clock says where they are — the other thing a
	// person opening somebody's card is there for, and the thing that is
	// always wrong if it is computed from the reader's own clock.
	LocalTime string
	// Channels is where you can find them, as one line of names.
	Channels string
}

// UserProfileCard is somebody's card: their role, their presence, their local
// time, and where you can find them.
//
// It is [overlay.HoverCard] because that is exactly what it is — a panel that
// opens because the pointer rested on something — and because a card that
// answered a press differently would be a card whose dismissal a reader has to
// learn.
func UserProfileCard(c *ui.Context, opts UserProfileCardOptions) *ui.Element {
	if opts.Anchor == nil {
		panic("messaging: UserProfileCard needs the Anchor it hangs off")
	}
	if opts.Open == nil {
		panic("messaging: UserProfileCard needs the *bool it opens and closes with")
	}
	if opts.Name == "" {
		panic("messaging: UserProfileCard needs a Name; a card about nobody is a blank panel")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	card := overlay.HoverCard(c, opts.Anchor, opts.Open, overlay.HoverCardOptions{
		Modal:    false,
		Title:    opts.Name,
		Subtitle: firstOf(opts.Role, opts.Handle),
		Width:    u * 44,
		Label:    opts.Name,
		Body: func() {
			OnlineStatus(c, OnlineStatusOptions{
				Presence: opts.Presence, Name: opts.Name,
				Since: opts.Since, WithText: true,
			})
			if opts.LocalTime != "" {
				ui.Text(c, opts.LocalTime).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize))
			}
			if opts.Channels != "" {
				ui.Text(c, opts.Channels).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.CaptionSize))
			}
		},
	})
	return card
}

func firstOf(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// ── status ─────────────────────────────────────────────────────────────────

// StatusSetterOptions configure a StatusSetter.
type StatusSetterOptions struct {
	// Status is the caller's own line — "in a workshop until 4", "back
	// Monday" — and Open the *bool the editor opens with. Both are required:
	// the text is the reader's and the panel's open state is not this
	// component's to remember.
	Status *string
	Open   *bool
	// Limit is how many characters fit. Zero takes the library's own, which is
	// the length of a line the status shows in a sidebar before it is
	// truncated.
	Limit int
	// Placeholder is what an empty field shows.
	Placeholder string
	// Expires says when the status runs out, as the caller's string, and
	// clears the button's need to guess.
	Expires string
	// Title heads the panel.
	Title string
	// Label names the control; empty takes the library's "Set a status".
	Label string
}

// StatusSetterResult carries a StatusSetter.
type StatusSetterResult struct {
	// Element is the control.
	Element *ui.Element
	// cleared reports the clear button being pressed this frame.
	cleared bool
}

// Cleared reports the clear button being pressed, which is a distinct action
// from setting a new status and one a caller has to be able to tell apart: a
// status that is removed and a status that is replaced are different things
// to whatever is downstream of this — an away timer, a notification, a list
// of who is busy.
func (r StatusSetterResult) Cleared() bool { return r.cleared }

// StatusSetter is the "set a status" control: a row showing what the reader has
// told people, and the two ways to change it.
func StatusSetter(c *ui.Context, opts StatusSetterOptions) StatusSetterResult {
	if opts.Status == nil {
		panic("messaging: StatusSetter needs a Status to write into; it keeps no line of its own")
	}
	if opts.Open == nil {
		panic("messaging: StatusSetter needs the *bool its editor opens with")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "messaging.status", core.Def("Set a status"))
	}
	title := opts.Title
	if title == "" {
		title = label
	}

	var res StatusSetterResult
	res.Element = ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).
		Label(label).Children(func() {
		// The status is drawn in a box of its own rather than as the button's
		// label, so the button stays the size of a button and the text can
		// truncate without the control it opens growing with it.
		ui.Box(c).Grow(1).Children(func() {
			if *opts.Status == "" {
				ui.Text(c, statusLabel(label)).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.RowSize))
			} else {
				ui.Text(c, *opts.Status).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).MaxLines(1)
			}
			if opts.Expires != "" && *opts.Status != "" {
				ui.Text(c, opts.Expires).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		})

		if *opts.Status != "" {
			clear := ui.Box(c).Size(u*4, u*4).Radius(u * 2).
				Cursor(ui.CursorPointer).Label("Clear status").Role(ui.RoleNone)
			if clear.Clicked() {
				res.cleared = true
				*opts.Status = ""
			}
			clear.Children(func() {
				ui.Icon(c, glyph("failed")).Size(u*3, u*3).TextColor(k.TextMuted)
			})
		}

		if input.Button(c, label, input.ButtonOptions{}).Clicked() {
			*opts.Open = !*opts.Open
		}
	})

	// The editor is a dialog rather than a popover: it is opened deliberately
	// and asks for a sentence, which is a form, and a form in a popover is a
	// panel that closes when the reader looks for the keyboard.
	overlay.Dialog(c, opts.Open, overlay.DialogOptions{
		Title: title, Rule: true, Width: u * 52,
		Body: func() {
			input.TextArea(c, opts.Status, input.TextAreaOptions{
				Label:       "Status",
				Placeholder: opts.Placeholder,
				Lines:       2,
			})
			ui.Text(c, itoa(len(*opts.Status))+" / "+itoa(limit)+" characters").
				TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		},
		Actions: func() {
			if input.Button(c, "Done", input.ButtonOptions{Primary: true}).Clicked() {
				*opts.Open = false
			}
		},
	})
	return res
}

func statusLabel(label string) string { return "No " + lower(label) }

// statusTime is how a status reads when it has one, for a caller that wants
// to show a countdown. It is a helper rather than part of the control because
// "in 4 minutes" is the caller's sentence about the caller's clock.
func statusTime(until time.Time, now time.Time) string {
	left := until.Sub(now)
	if left <= 0 {
		return "Expired"
	}
	return "in " + statusDuration(left)
}

func statusDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return itoa(int(d.Seconds())) + " seconds"
	case d < time.Hour:
		return itoa(int(d.Minutes())) + " minutes"
	}
	return itoa(int(d.Hours())) + " hours"
}
