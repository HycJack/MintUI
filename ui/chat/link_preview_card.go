package chat

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/input"
	"github.com/HycJack/MintUI/ui/theme"
)

// LinkPreviewCardOptions configure a LinkPreviewCard.
type LinkPreviewCardOptions struct {
	// URL is where the link goes. It is what the Open button acts on — the
	// caller opens it, through its own transport, and this card only says
	// that it was asked.
	URL string
	// Title is the page's own title. Required: it is the line a reader
	// decides on, and a card with a host and a description but no title is a
	// card about a page nobody can name.
	Title string
	// Host is the site, drawn under the title. Empty takes the host the URL
	// names, which is the usual case — the caller rarely has the two
	// separately — and a URL with no host at all draws nothing.
	Host string
	// Description is the line under the host, what the page is about.
	// Empty draws no description.
	Description string
}

// LinkPreviewCardResult carries a LinkPreviewCard and its press.
type LinkPreviewCardResult struct {
	// Element is the card.
	Element *ui.Element
	opened  bool
}

// Opened reports the Open button being pressed this frame. The card does not
// open anything itself: opening a link is the caller's transport — a
// browser, an in-app webview — and a component that chose would be a
// component that knew more about the app than the app does.
func (r LinkPreviewCardResult) Opened() bool { return r.opened }

// LinkPreviewCard is a link, shown as the page it points at: the title, the
// site, a line about it, and the button that goes there.
//
// It is a card rather than a link because a link in a conversation is a
// reference and a card is a place: the reader has to decide on the title
// before deciding on the link, and a plain blue URL offers nothing to decide
// on. The favicon is the first letter of the host in a square for the same
// reason an avatar is the first letter of a name — the card is recognisable
// at a glance before it is read, and a letter is the one thing every host has.
func LinkPreviewCard(c *ui.Context, opts LinkPreviewCardOptions) LinkPreviewCardResult {
	if opts.Title == "" {
		panic("chat: LinkPreviewCard needs a Title; a card with no title is a rectangle " +
			"about a page nobody can name")
	}
	k, u := core.Tokens(c), core.Density(c).Unit()

	host := opts.Host
	if host == "" {
		host = linkHost(opts.URL)
	}
	initial := ""
	if host != "" {
		if r := []rune(strings.TrimSpace(host)); len(r) > 0 {
			initial = strings.ToUpper(string(r[:1]))
		}
	}

	open := core.Msg(c, "chat.link.open", core.Def("Open"))
	var res LinkPreviewCardResult
	res.Element = ui.Column(c).FillWidth().Gap(u*1.5).
		Radius(theme.CardRadius).Background(k.Surface).
		Border(theme.BorderWidth, k.Border).Padding(u * 2).Label(opts.Title)
	res.Element.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(u * 1.5).Children(func() {
			// The favicon square is a mark, not a control: it sits where an
			// avatar sits and says the same thing — what this came from.
			ui.Box(c).Size(u*7, u*7).Radius(theme.SmallRadius).
				Background(k.Background).Border(theme.BorderWidth, k.Border).
				Shrink(0).Center().Label("Favicon").Children(func() {
				if initial != "" {
					ui.Text(c, initial).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.BodySize))
				}
			})
			ui.Column(c).Grow(1).Shrink(1).AlignItems(ui.Start).Gap(u * 0.25).Children(func() {
				ui.Text(c, opts.Title).TextColor(k.Text).
					FontSize(core.FontSize(c, theme.RowSize)).SingleLine()
				if host != "" {
					ui.Text(c, host).TextColor(k.TextMuted).
						FontSize(core.FontSize(c, theme.CaptionSize)).SingleLine()
				}
			})
			btn := input.Button(c, open, input.ButtonOptions{Label: open})
			if btn.Clicked() {
				res.opened = true
			}
		})
		if opts.Description != "" {
			ui.Text(c, opts.Description).TextColor(k.TextMuted).
				FontSize(core.FontSize(c, theme.MetaSize)).MaxLines(2)
		}
	})
	return res
}

// linkHost is the host part of a URL: what is left once the scheme is off
// and the path is cut away. It is the smallest possible reader of a URL, not
// a URL parser — a URL with no scheme gives its whole front, which is what a
// caller that typed "docs.mintui.dev" rather than "https://docs.mintui.dev"
// wants back.
func linkHost(url string) string {
	s := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	return s
}
