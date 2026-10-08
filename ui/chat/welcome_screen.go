package chat

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// WelcomeCard is one card on a WelcomeScreen: a title and the sentence under
// it, and the press that opens whatever the title names.
type WelcomeCard struct {
	// Title is what the card is called — "Start a task", "Import a file".
	// Required: it is the line the card is read by, and it is also what the
	// press is reported by, since the result is an index and the caller
	// looks the title up itself.
	Title string
	// Body is the sentence under the title — what happens on the press.
	// Empty draws a card with no second line.
	Body string
}

// WelcomeScreenOptions configure a WelcomeScreen.
type WelcomeScreenOptions struct {
	// Greeting is the line at the top. Required: the screen exists to be
	// looked at before anything else on it, and a screen that opens on a
	// wall of cards without a word of its own is a catalogue, not a welcome.
	Greeting string
	// Body is the sentence under the greeting. Empty draws no second line.
	Body string
	// Cards are the cards, in the order they are shown: odd positions down
	// the left column, even down the right.
	Cards []WelcomeCard
}

// WelcomeScreenResult carries a WelcomeScreen and the card chosen.
type WelcomeScreenResult struct {
	// Element is the whole screen.
	Element *ui.Element
	// picked is -1 until a card is pressed.
	picked int
}

// Picked is the index of the card pressed this frame, into Cards, -1 for
// none. It is the index and not the title, for the reason Suggested is: the
// caller handed the cards over and already knows what number N refers to.
func (r WelcomeScreenResult) Picked() int { return r.picked }

// WelcomeScreen is the first screen of a chat window: the greeting, the
// sentence under it, and the cards that say what the window is for.
//
// The cards are two columns rather than a wrapped row, because a welcome
// screen is read top to bottom in pairs — the left card, then the one beside
// it — and a wrap row of equal cards reorders that reading as the window
// narrows. Two Grow columns keep the pairs stable at every width, and a
// screen with a single card simply leaves its second column empty, which is
// what a single card on two columns is.
func WelcomeScreen(c *ui.Context, opts WelcomeScreenOptions) WelcomeScreenResult {
	if opts.Greeting == "" {
		panic("chat: WelcomeScreen needs a Greeting; a screen that opens on cards with " +
			"no word of its own is a catalogue, not a welcome")
	}
	for i, card := range opts.Cards {
		if card.Title == "" {
			panic("chat: WelcomeScreen card " + internal.Commas(i) +
				" needs a Title; a card with no title cannot be reported by index")
		}
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	var res WelcomeScreenResult
	res.picked = -1

	// One card per column, split by position: the left column takes the
	// even slots, the right the odd. The split is by index and not by
	// height, because the columns are drawn independently and neither
	// knows how tall the other will be.
	var left, right []WelcomeCard
	for i, card := range opts.Cards {
		if i%2 == 0 {
			left = append(left, card)
		} else {
			right = append(right, card)
		}
	}
	drawCard := func(index int, card WelcomeCard) {
		box := ui.ButtonBase(c).FillWidth().AlignItems(ui.Start).
			Radius(theme.ControlRadius).Background(k.Surface).
			Border(theme.BorderWidth, k.Border).Padding(u * 2).
			Label(card.Title)
		if box.Hovered() {
			box.Background(k.SurfaceHover)
		}
		if box.Clicked() {
			res.picked = index
		}
		box.Children(func() {
			ui.Text(c, card.Title).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.RowSize)).Bold()
			if card.Body != "" {
				ui.Text(c, card.Body).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(3)
			}
		})
	}

	res.Element = ui.Column(c).FillWidth().Grow(1).Gap(u * 2).Role(ui.RoleGroup)
	res.Element.Children(func() {
		ui.Column(c).AlignItems(ui.Center).Gap(u).Children(func() {
			ui.Text(c, opts.Greeting).TextColor(k.Text).
				FontSize(core.FontSize(c, theme.LeadSize)).Bold()
			if opts.Body != "" {
				ui.Text(c, opts.Body).TextColor(k.TextMuted).
					FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(2)
			}
		})
		if len(opts.Cards) > 0 {
			ui.Row(c).FillWidth().Gap(u * 2).Children(func() {
				ui.Column(c).Grow(1).Shrink(1).Gap(u * 1.5).Children(func() {
					for i, card := range left {
						drawCard(2*i, card)
					}
				})
				ui.Column(c).Grow(1).Shrink(1).Gap(u * 1.5).Children(func() {
					for i, card := range right {
						drawCard(2*i+1, card)
					}
				})
			})
		}
	})
	return res
}
