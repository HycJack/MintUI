package chat

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ChatContainerOptions configure a ChatContainer.
type ChatContainerOptions struct {
	// Title is what the chat area is called. Required: the header is the
	// reader's anchor for which conversation they are in, and a header with
	// a meta and actions but no name is chrome without a subject.
	Title string
	// Meta is the secondary line of the header, beside the title — a model
	// name, a branch, a member count. Empty draws nothing.
	Meta string
	// Messages builds the transcript, inside the growing middle. nil draws
	// the empty area, which is what a conversation with no turns yet is.
	Messages func()
	// Actions builds the marks at the right of the header — export, settings,
	// delete. nil draws none.
	Actions func()
}

// ChatContainer is the skeleton of a whole chat area: the header, the
// transcript's room, and the composer at the bottom.
//
// It is layout and chrome only. It does not scroll — the transcript scrolls
// inside the room it is given, and that is the caller's MessageList, which
// owns the scroll state and the follow-the-end decision — and it keeps no
// state: the title is the caller's string and the two slots are the
// caller's. A container that kept a copy of either would be a second source
// of the same words, and two sources of a conversation's name disagree within
// a week.
//
// The split it makes is the one every chat window makes: the transcript gets
// the room that grows, and the composer gets the strip it asks for, which
// means a window with ten turns and a window with two hundred put the send
// button in the same place.
func ChatContainer(c *ui.Context, opts ChatContainerOptions, composer func()) *ui.Element {
	if opts.Title == "" {
		panic("chat: ChatContainer needs a Title; a chat area with no name cannot be " +
			"told from another")
	}
	if composer == nil {
		panic("chat: ChatContainer needs the composer to draw; a chat area with no way " +
			"in is a transcript, not a conversation")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	return ui.Column(c).Fill().Gap(u * 1.5).Label(opts.Title).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u*1.5).
			Padding(u*1.5, u*2).Shrink(0).Children(func() {
			ui.Text(c, opts.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.TitleSize)).Bold().SingleLine()
			if opts.Meta != "" {
				ui.Text(c, opts.Meta).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).SingleLine()
			}
			ui.Box(c).Grow(1)
			if opts.Actions != nil {
				ui.Row(c).AlignItems(ui.Center).Gap(u).Shrink(0).Children(opts.Actions)
			}
		})
		layoutDivider(c, k)
		ui.Box(c).Grow(1).Shrink(1).FillWidth().
			Label("Messages").Children(func() {
			if opts.Messages != nil {
				opts.Messages()
			}
		})
		layoutDivider(c, k)
		ui.Box(c).FillWidth().Shrink(0).Padding(u*1.5, u*2).
			Label("Composer").Children(composer)
	})
}

// layoutDivider is the rule between the container's three strips, shared so
// the header's and the composer's rules are the same hairline.
func layoutDivider(c *ui.Context, k theme.Tokens) {
	ui.Box(c).FillWidth().Height(theme.BorderWidth).Shrink(0).Background(k.Border)
}
