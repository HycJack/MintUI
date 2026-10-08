package chat

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// ChatEmptyStateOptions configure a ChatEmptyState.
type ChatEmptyStateOptions struct {
	// Greeting is the first line — "Good morning, Ada", "Ask me anything".
	// Required: it is the one line a reader with nothing in front of them
	// actually reads, and an empty state that starts with a sentence of
	// body text is body text, not a greeting.
	Greeting string
	// Body is the sentence under the greeting — what this conversation can
	// do, what to ask first. Empty draws no second line.
	Body string
	// Suggestions are the chips under the body: whole questions a reader can
	// send instead of typing one. Empty draws no chips, which is what a
	// conversation that has nothing particular to offer is.
	Suggestions []string
}

// ChatEmptyStateResult carries a ChatEmptyState and the suggestion chosen.
type ChatEmptyStateResult struct {
	// Element is the whole state.
	Element *ui.Element
	// suggested is -1 until a chip is pressed.
	suggested int
}

// Suggested is the index of the suggestion chip pressed this frame, -1 for
// none. It is the index into Suggestions and not the text, because the
// caller already holds the text: it handed the chips over, and the chip is
// what it wants to send.
func (r ChatEmptyStateResult) Suggested() int { return r.suggested }

// ChatEmptyState is what a conversation shows before it has one: the
// greeting, the sentence under it, and the chips that start it.
//
// It is deliberately not feedback.Empty. That component is the generic "this
// surface has nothing in it yet, here is the one action that would change
// it" — a card with an illustration and a single button — and it is the
// right answer for an inbox or a project with no items. A conversation's
// empty state is different in kind: there is no single action, because the
// first thing that fills it is a question, and the suggestions are the
// questions the system already knows the reader might want to ask. A card
// with a button would have to invent a verb, and the chip row answers with
// nouns instead.
func ChatEmptyState(c *ui.Context, opts ChatEmptyStateOptions) ChatEmptyStateResult {
	if opts.Greeting == "" {
		panic("chat: ChatEmptyState needs a Greeting; an empty state that says nothing " +
			"is still empty")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res ChatEmptyStateResult
	res.suggested = -1

	col := ui.Column(c).FillWidth().Grow(1).AlignItems(ui.Center).
		Justify(ui.Center).Gap(u * 2).Role(ui.RoleGroup).Label(opts.Greeting)
	col.Children(func() {
		ui.Text(c, opts.Greeting).TextColor(k.Text).
			FontSize(core.FontSize(c, theme.LeadSize)).Bold()
		if opts.Body != "" {
			ui.Text(c, opts.Body).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(2)
		}
		if len(opts.Suggestions) > 0 {
			ui.Column(c).AlignItems(ui.Center).Gap(u * 0.75).Children(func() {
				for i, s := range opts.Suggestions {
					if s == "" {
						panic("chat: ChatEmptyState suggestion " +
							internal.Commas(i) + " is empty; a chip with no words " +
							"cannot be pressed")
					}
					index := i
					chip := ui.ButtonBase(c).Radius(theme.PillRadius).
						Padding(u*0.75, u*2).Label(s).Children(func() {
						ui.Text(c, s).TextColor(k.Text).
							FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
					})
					if chip.Hovered() {
						chip.Background(k.Surface)
					}
					if chip.Clicked() {
						res.suggested = index
					}
				}
			})
		}
	})
	res.Element = col
	return res
}
