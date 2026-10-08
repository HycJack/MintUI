// Package data holds the components that present a record: cards, the
// figures that summarise a set of them, and the tables that list them.
package data

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// CardOptions configure a Card.
type CardOptions struct {
	// Title heads the card. Empty draws no heading.
	Title string
	// Meta is the line under the title: a reference, a description.
	Meta string
	// Footer builds the bottom slot; nil draws none.
	Footer func()
	// FooterRule separates the footer from the body with a line. A card whose
	// footer is one line of metadata does not need it; a footer holding
	// actions does.
	FooterRule bool
	// Header builds a slot above the title, for tags and controls that sit
	// on a line of their own.
	Header func()
	// Draggable makes the card a drag source for value, which the caller
	// supplies through Draggable's argument.
	Draggable any
	// Meta2 is the third line, for a card that carries one more fact.
	Meta2 string
}

// Card is a record on a surface: something about it, then a footer.
//
// It adds no behaviour of its own. A card that can be pressed, dragged or
// dismissed is a card plus the caller's handler, so that the press stays
// where the card is drawn.
func Card(c *ui.Context, opts CardOptions, body func()) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()

	label := opts.Title
	if label == "" && opts.Meta != "" {
		label = opts.Meta
	}
	card := ui.Column(c).FillWidth().Radius(theme.CardRadius).Background(k.Background).
		Border(theme.BorderWidth, k.Border).Role(ui.RoleGroup)
	if label != "" {
		card.Label(label)
	}
	if opts.Draggable != nil {
		card = card.Drag(opts.Draggable).Cursor(ui.CursorGrab)
	}

	card.Children(func() {
		ui.Column(c).FillWidth().Padding(u*3.75, u*4, u*3.5, u*4).Gap(u * 0.75).Children(func() {
			if opts.Header != nil {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Padding(0, 0, u*2.75, 0).
					Children(opts.Header)
			}
			if opts.Title != "" {
				ui.Text(c, opts.Title).TextColor(k.Text).FontSize(theme.BodySize).Bold().SingleLine()
			}
			if opts.Meta != "" {
				ui.Text(c, opts.Meta).TextColor(k.TextMuted).FontSize(theme.MetaSize).SingleLine()
			}
			if opts.Meta2 != "" {
				ui.Text(c, opts.Meta2).TextColor(k.TextMuted).FontSize(theme.MetaSize).SingleLine()
			}
			if body != nil {
				ui.Column(c).FillWidth().Children(body)
			}
		})
		if opts.FooterRule {
			ui.Box(c).FillWidth().Height(1).Shrink(0).Background(k.Border)
		}
		if opts.Footer != nil {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Padding(u*2.75, u*4, u*2.75, u*4).
				Children(opts.Footer)
		}
	})
	return card
}

// Stat is one figure in a line of figures.
type Stat struct {
	// Label says what is being counted.
	Label string
	// Value is the figure, already formatted; a Stat formats nothing.
	Value string
	// Muted draws the value in the secondary colour, for a figure that is
	// context rather than the point.
	Muted bool
}

// StatLine draws a run of figures separated by middots, the way a header
// summarises what is on screen. The figures stay in the caller's string form:
// a line of numbers is exactly where a wrong unit becomes invisible.
//
// A figure reads value first and its unit after it — "27 open", not "open 27"
// — and the two are one labelled phrase, so a screen reader says the whole
// thing rather than stopping between the number and its unit.
func StatLine(c *ui.Context, stats ...Stat) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	return ui.Row(c).AlignItems(ui.Center).Wrap().Gap(u * 2.5).Children(func() {
		for i, s := range stats {
			if i > 0 {
				ui.Text(c, "·").TextColor(k.TextFaint).FontSize(theme.RowSize)
			}
			col := k.Text
			if s.Muted {
				col = k.TextMuted
			}
			ui.RichText(c,
				ui.Span{Text: s.Value, Weight: 700, Color: col, Size: theme.RowSize},
				ui.Span{Text: " " + s.Label, Color: k.TextMuted, Size: theme.RowSize},
			).Label(s.Value + " " + s.Label)
		}
	})
}
