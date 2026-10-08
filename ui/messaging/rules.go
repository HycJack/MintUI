package messaging

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/chat"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ── the unread divider ─────────────────────────────────────────────────────

// FirstUnread is the index the "new messages" divider goes above, or -1 when
// there is nothing new to divide.
//
// index is where the reader has got to — the index of the last message they
// have read, and -1 when they have read none — and count is how many messages
// there are. It is the whole rule in one function because it is the rule every
// message channel gets wrong differently: some put the divider below the last
// read message, some put it above the first one, some show it on every
// re-render after the read marker moves, and none of those is the same as what
// the reader is used to from every other application.
//
// The three cases, which are the three tests:
//
//	all read        index = count-1  →  -1
//	nothing read    index = -1       →   0
//	empty list      count = 0        →  -1
//
// The divider belongs *above* the first unread message, not below the last
// read one, because the reader is looking for where they stopped and the
// line they stopped at is the one they are looking at.
func FirstUnread(index, count int) int {
	if count <= 0 || index >= count-1 {
		return -1
	}
	if index < -1 {
		return -1
	}
	return index + 1
}

// DividerAt reports whether the divider goes above message i, which is what
// the renderer asks on every row of a transcript.
//
// It is asked rather than compared by the caller because a caller that wrote
// the comparison itself would have to re-derive FirstUnread's two edge cases,
// and one of them — an empty list — is the case that turns up exactly when a
// channel has just been opened.
func DividerAt(i, index, count int) bool { return count > 0 && i == FirstUnread(index, count) }

// UnreadDividerOptions configure an UnreadDivider.
type UnreadDividerOptions struct {
	// Count is how many messages are below the line — the number the reader
	// is about to be shown and has not seen. It is the caller's count, taken
	// from the same place the read marker is: a component that counted the
	// messages itself would be counting a copy of a list it does not own.
	Count int
	// Label is what the line says; empty takes the library's own
	// "New messages".
	Label string
	// Since is when the reader was last here, shown beside the count. Empty
	// leaves it out, for a channel whose reader is never away long enough for
	// the time to mean anything.
	Since time.Duration
}

// UnreadDivider is the line that says "everything below this is new".
//
// It is a rule with a word on it rather than a count in a badge beside the
// channel: the badge says how many arrived, and the divider says where the
// reading stopped. Those are different questions, and a reader who has
// scrolled back through yesterday and is now at the line needs the second one
// answered at the line.
func UnreadDivider(c *ui.Context, opts UnreadDividerOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Label
	if label == "" {
		label = core.Msg(c, "messaging.unread", core.Def("New messages"))
	}
	what := label
	switch {
	case opts.Count <= 0:
		// Nothing below the line: the word alone, with no number claiming
		// there is something to read.
	case opts.Count == 1:
		what = label + " · 1 message"
	default:
		what = label + " · " + itoa(opts.Count) + " messages"
	}

	// RoleNone: the line is decoration and the word beside it is the content.
	// A separator role would take the word out of the accessibility tree and
	// leave a reader with an unnamed line where the one piece of information
	// is "there are messages you have not seen".
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).
		Label(what).Role(ui.RoleNone).Children(func() {
		ui.Box(c).Grow(1).Height(theme.BorderWidth).Background(k.Accent)
		ui.Text(c, what).TextColor(k.AccentText).
			FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		ui.Box(c).Grow(1).Height(theme.BorderWidth).Background(k.Accent)
	})
}

// ── the read receipt ───────────────────────────────────────────────────────

// Receipt is how far a message of the reader's own has got.
type Receipt int

const (
	// ReceiptSending is written but not yet handed to the server. One clock.
	ReceiptSending Receipt = iota
	// ReceiptDelivered is on the other end's device. Two ticks.
	ReceiptDelivered
	// ReceiptRead has been opened. Two filled ticks, so the difference from
	// delivered survives being drawn in a single colour.
	ReceiptRead
	// ReceiptFailed never got there. A warning, and the one state that has to
	// carry a word as well as a mark.
	ReceiptFailed
)

func (r Receipt) String() string {
	switch r {
	case ReceiptDelivered:
		return "Delivered"
	case ReceiptRead:
		return "Read"
	case ReceiptFailed:
		return "Failed to send"
	}
	return "Sending"
}

// ReceiptState is the mark a receipt state draws, and whether it draws one at
// all.
//
// The three states are told apart by *shape* rather than by colour alone,
// because a reader who cannot see the difference between two tints of grey
// cannot tell "they got it" from "they read it", and that is the whole
// information a receipt carries:
//
//	Sending    → "clock"     (one mark, open)
//	Delivered  → "delivered" (two marks, open)
//	Read       → "read"      (two marks, filled)
//
// The bool is false for an unknown state, so a receipt arriving from a newer
// server than the app knows about draws nothing rather than drawing a tick
// that claims delivery.
func ReceiptState(s Receipt) (string, bool) {
	switch s {
	case ReceiptSending:
		return "clock", true
	case ReceiptDelivered:
		return "delivered", true
	case ReceiptRead:
		return "read", true
	case ReceiptFailed:
		return "failed", true
	}
	return "", false
}

// receiptName is a state's own name, and false for one this build does not
// know. It is separate from [Receipt.String], which returns "Sending" for
// anything unrecognised — a method with no way to say "I do not know" has to
// pick a plausible value, and a receipt labelled "Sending" when the server
// said something else is a claim the library cannot support.
func receiptName(s Receipt) (string, bool) {
	switch s {
	case ReceiptSending:
		return "Sending", true
	case ReceiptDelivered:
		return "Delivered", true
	case ReceiptRead:
		return "Read", true
	case ReceiptFailed:
		return "Failed to send", true
	}
	return "", false
}

// receiptTone is what colour a receipt's mark is in. Sent and delivered are
// quiet; read is the accent, because it is the only one the reader is waiting
// for; failed is the danger tone, because it is the only one they must act on.
func receiptTone(s Receipt) core.Severity {
	switch s {
	case ReceiptRead:
		return core.Accent
	case ReceiptFailed:
		return core.Danger
	}
	return core.Neutral
}

// ReadReceiptOptions configure a ReadReceipt.
type ReadReceiptOptions struct {
	// State is how far the message got.
	State Receipt
	// Times is when it got there, shown beside the mark. Empty writes the
	// state as a word instead — which is what a failure wants, since a
	// failure with a time on it reads as "it failed at 14:02" when what it
	// means is "it has not arrived and here is why".
	Time string
	// Label names the receipt; empty is built from the state and the time.
	Label string
	// Muted draws the whole thing in the secondary tone, which is what a
	// reader's own outgoing message wants: the receipt is furniture beside
	// the words, not a second thing to read.
	Muted bool
}

// ReadReceipt is the mark under a message of the reader's own saying where it
// got to.
//
// It sits beside the message's own time rather than under it. The time says
// when the message was sent and the receipt says where it got to; putting
// them on separate lines doubles the height of every outgoing turn in a
// transcript, which is the row type a transcript has the most of.
func ReadReceipt(c *ui.Context, opts ReadReceiptOptions) *ui.Element {
	name, shown := ReceiptState(opts.State)
	k, u := core.Tokens(c), core.Density(c).Unit()

	label := opts.Label
	if label == "" {
		word, known := receiptName(opts.State)
		if !known {
			// Honest rather than plausible: the state is one this build has
			// not heard of, and saying "Sending" would tell the reader their
			// message is on its way when the library simply does not know.
			word = "Delivery status unavailable"
		}
		label = word
		if opts.Time != "" {
			label += " " + opts.Time
		}
	}

	e := ui.Row(c).AlignItems(ui.Center).Gap(u * 0.75).Label(label).
		Role(ui.RoleStatus)
	if !shown {
		// An unknown state draws an empty row rather than a tick. The row keeps
		// its height so a transcript does not jump when a server starts
		// reporting a state the app has not learned yet — and it is returned
		// rather than built with no children, because a Children call with no
		// closure is not a way to say "empty" to MyGo.
		return e
	}

	tone := receiptTone(opts.State)
	ink := k.TextMuted
	if tone != core.Neutral {
		_, ink = tone.Pair(k)
	}
	if opts.Muted {
		ink = k.TextFaint
	}
	size := u * 3.25
	return e.Children(func() {
		ui.Icon(c, glyph(name)).Size(size, size).TextColor(ink).Label(label).Shrink(0)
		if opts.Time != "" {
			ui.Text(c, opts.Time).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
		}
	})
}

// ── quoting ────────────────────────────────────────────────────────────────

// QuotedTextOptions configure a QuotedText.
type QuotedTextOptions struct {
	// Author is whose words these are.
	Author string
	// Text is the words. Required: an empty quote is a rule with nothing
	// beside it.
	Text string
	// Lines is how many lines to show; zero takes one.
	Lines int
	// Side is the column the reply is on, so the rule hangs off the same edge
	// as the bubble under it. Zero, ui.Start.
	Side ui.Align
	// Label names the quote; empty is built from the author.
	Label string
}

// QuotedText is the earlier turn a reply answers, shown above it.
//
// It is [chat.QuoteReply] called through, not a second one. The reasoning is
// the same as for the bubble and the stakes are the same: a quote is the one
// place a transcript shows a message inside a message, and it is exactly the
// place where two implementations diverge into a quote that looks like a
// message and one that does not. What this adds is the channel's own
// decoration — a jump to the quoted message — which the transcript has no
// reason to offer and a channel does, because a channel's messages are in a
// scroll rather than in front of the reader.
func QuotedText(c *ui.Context, opts QuotedTextOptions) *ui.Element {
	if opts.Text == "" {
		panic("messaging: QuotedText needs Text; an empty quote is a rule with nothing " +
			"beside it")
	}
	return chat.QuoteReply(c, chat.QuoteReplyOptions{
		Author: opts.Author,
		Text:   opts.Text,
		Lines:  opts.Lines,
		Side:   opts.Side,
		Label:  opts.Label,
	})
}

// ── presence ───────────────────────────────────────────────────────────────

// Presence is how reachable somebody is.
type Presence int

const (
	// PresenceOffline is signed out. No mark at all: the absence of a dot is
	// the answer, and an "offline" pill on every one of two hundred rows is
	// noise that pushes the two people who are online out of sight.
	PresenceOffline Presence = iota
	// PresenceAway is signed in and not at the keyboard.
	PresenceAway
	// PresenceOnline is at the keyboard.
	PresenceOnline
	// PresenceBusy is at the keyboard and not to be interrupted. It carries
	// its own colour rather than a dial tone on the same green, because "busy"
	// is a different request from "here".
	PresenceBusy
)

func (p Presence) String() string {
	switch p {
	case PresenceAway:
		return "Away"
	case PresenceOnline:
		return "Online"
	case PresenceBusy:
		return "Busy"
	}
	return "Offline"
}

// presenceTone is a presence's own severity, out of the ramp rather than off
// the ramp: a presence is not a status, it is a state, and the palette's
// warning and danger are the two colours a reader already reads as "not now"
// and "do not".
func presenceTone(p Presence) core.Severity {
	switch p {
	case PresenceOnline:
		return core.Success
	case PresenceAway:
		return core.Warning
	case PresenceBusy:
		return core.Danger
	}
	return core.Neutral
}

// OnlineStatusOptions configure an OnlineStatus.
type OnlineStatusOptions struct {
	// Presence is how reachable the person is.
	Presence Presence
	// Name is who it is about, and is required: a dot is the one mark in the
	// interface with no shape of its own to be recognised by.
	Name string
	// Since is when they came online or went away, shown as a word beside the
	// dot. Empty leaves it out.
	Since string
	// WithText draws the state as a word rather than as a dot alone, for a
	// profile card or a call tile where there is room and a reason.
	WithText bool
	// Size scales the dot; zero takes the library's own.
	Size float32
}

// OnlineStatus is the presence dot: a small mark in the colour of the state,
// beside a person's name.
func OnlineStatus(c *ui.Context, opts OnlineStatusOptions) *ui.Element {
	if opts.Name == "" {
		panic("messaging: OnlineStatus needs a Name; a dot is the one mark in the interface " +
			"with no shape to be recognised by")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()
	label := opts.Name + " is " + strings.ToLower(opts.Presence.String())
	if opts.Since != "" {
		label += ", " + opts.Since
	}

	e := ui.Row(c).AlignItems(ui.Center).Gap(u * 0.75).Label(label).Role(ui.RoleStatus)
	if opts.Presence == PresenceOffline && !opts.WithText {
		// Offline with no room for a word draws a slot rather than nothing, so
		// that a list of names does not change width when somebody signs off.
		return e.Children(func() {
			ui.Box(c).Size(u*1.5, u*1.5).Radius(u).Shrink(0).Role(ui.RoleNone)
		})
	}

	side := opts.Size
	if side <= 0 {
		side = u * 2.25
	}
	_, dot := opts.Presence.presencePair(k)
	return e.Children(func() {
		ui.Box(c).Size(side, side).Radius(side / 2).Background(dot).Shrink(0).
			Label(label).Role(ui.RoleNone)
		if opts.WithText {
			ui.Text(c, opts.Presence.String()).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			if opts.Since != "" {
				ui.Text(c, opts.Since).TextColor(k.TextFaint).
					FontSize(core.FontSize(c, theme.CaptionSize)).Shrink(0)
			}
		}
	})
}

// presencePair is a presence's dot and its word, so that a dot and a word are
// never drawn from two different answers about what a state looks like.
func (p Presence) presencePair(k theme.Tokens) (ui.Color, ui.Color) {
	if p == PresenceOffline {
		return k.SurfacePressed, k.TextMuted
	}
	_, fg := presenceTone(p).Pair(k)
	return fg, fg
}

// itoa is a small count as a string. It is here so the divider and the member
// counts write their numbers the same way, without each of them reaching for
// strconv over a one-line function.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
