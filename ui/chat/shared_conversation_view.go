package chat

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// SharedMessage is one turn of a conversation somebody shared.
type SharedMessage struct {
	// Sender is who said it. Required: a shared conversation is several
	// people, and a turn with no speaker cannot be attributed — which is the
	// whole reason a reader opened it.
	Sender string
	// Text is what was said. Required: a turn with no words is a gap in the
	// transcript, and a gap in somebody else's conversation is not content.
	Text string
	// Time is when it was said, already formatted by the caller — the view
	// draws whatever it is given, for the reason every date in this
	// component is the caller's.
	Time string
}

// SharedConversationViewOptions configure a SharedConversationView.
type SharedConversationViewOptions struct {
	// Messages are the turns, in the order they were said. Empty is allowed
	// and draws the badge and the read-only line and nothing between: a
	// shared conversation with no turns is still a conversation that was
	// shared, and the badge says so.
	Messages []SharedMessage
}

// SharedConversationView is a conversation somebody else had, shown to be
// read: the Shared badge on top, the turns as read-only rows, and the line
// at the bottom that says so.
//
// It is a view and not a MessageList for the same reason a receipt is not an
// account: the turns are finished, nobody in this window will add to them,
// and the one thing the window must say is that there is no composer. A
// reader who looks for a send box and finds the read-only line instead has
// been told what they could not otherwise know — that the conversation ended
// somewhere before this window. The rows are plain text rows rather than
// MessageBubbles on purpose: a bubble announces that its side is the reader,
// and in a shared conversation the reader is on no side.
func SharedConversationView(c *ui.Context, opts SharedConversationViewOptions) *ui.Element {
	for i, m := range opts.Messages {
		if m.Sender == "" || m.Text == "" {
			panic("chat: SharedConversationView message " + internal.Commas(i) +
				" needs a Sender and Text; a shared turn with no speaker or no words " +
				"is a gap")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	badge := core.Msg(c, "chat.shared.badge", core.Def("Shared"))
	readonly := core.Msg(c, "chat.shared.readonly",
		core.Def("This conversation is read-only"))

	return ui.Column(c).FillWidth().Gap(u * 1.5).Label(badge).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(u).Children(func() {
			bg, fg := core.Accent.Pair(k)
			ui.Box(c).Padding(u*0.5, u*1.5).Radius(theme.PillRadius).
				Background(bg).Shrink(0).Children(func() {
				ui.Text(c, badge).TextColor(fg).
					FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
			})
		})
		ui.Column(c).FillWidth().Gap(u).Children(func() {
			for _, m := range opts.Messages {
				ui.Column(c).FillWidth().Gap(u * 0.25).Label(m.Sender).Children(func() {
					ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
						ui.Text(c, m.Sender).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).Bold().SingleLine()
						if m.Time != "" {
							ui.Text(c, m.Time).TextColor(k.TextFaint).
								FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
						}
					})
					ui.Text(c, m.Text).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.BodySize)).MaxLines(3)
				})
			}
		})
		ui.Row(c).FillWidth().Justify(ui.Center).Padding(u, 0, 0, 0).Children(func() {
			ui.Text(c, readonly).TextColor(k.TextFaint).
				FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}
