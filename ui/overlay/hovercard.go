package overlay

import (
	"github.com/egoist/mygo/ui"
)

// HoverCardOptions configure a HoverCard.
type HoverCardOptions struct {
	// Title heads the card; empty draws one of body alone.
	Title string
	// Subtitle is the second line under the title.
	Subtitle string
	// Body is the content.
	Body func()
	// Width is the card's width; zero fits its content.
	Width float32
	// MaxWidth caps it; zero leaves it to the window.
	MaxWidth float32
	// Label names the card; empty uses the title.
	Label string
	// Modal puts a scrim over the window and sends outside presses to it.
	// It is off by default, and that default is the point: this layer hangs
	// off something on a page the person is still working in, so a scrim
	// across that page would say the work stopped — when what actually
	// happened is that one of its rows is being explained.
	Modal bool
}

// HoverCard is a panel that opens because the pointer rested on something, so
// that a thing can be read without being clicked into: who owns a card, what
// a branch holds, what a file weighs.
//
// The *bool is still the caller's, and the pointer writes to it — true while
// the pointer is on the anchor or on the card, false once it has left both.
// Nothing else closes it, because a card a stray click could dismiss is worse
// than one that stays up a moment too long.
//
// It does not dim the window unless Modal is set, and a hover card never really wants
// NonModal: the pointer is somewhere on the page, and a scrim appearing under
// it would be the loudest thing on screen for a card nobody asked for.
//
//	core.Use(c, core.Settings{})
//	card := ui.Box(c).Label("Maple Street Bakery").Children(func() { … })
//	HoverCard(c, card, &cardOpen, HoverCardOptions{
//	    Modal: true,
//	    Title:    "Last resolved",
//	    Body:     func() { ui.Text(c, "Riverside Clinic, 4 March") },
//	})
func HoverCard(c *ui.Context, anchor *ui.Element, open *bool, opts HoverCardOptions) *ui.Element {
	if anchor == nil {
		panic("overlay: HoverCard needs the element it hangs on")
	}
	if open == nil {
		panic("overlay: HoverCard needs the *bool it opens and closes")
	}
	// Hovered marks the anchor as taking a pointer, which is the other half
	// of the bookkeeping below: a card cannot know the pointer left unless
	// something asked.
	hovering := anchor.Hovered()

	panel := anchored(c, anchor, open, opts.Modal, func(p *ui.Element) {
		Panel(c, p, PanelOptions{
			Title:    opts.Title,
			Subtitle: opts.Subtitle,
			Compact:  true,
			MaxWidth: opts.MaxWidth,
			Label:    opts.Label,
		}, func() {
			if opts.Body != nil {
				opts.Body()
			}
		})
	})

	switch {
	case hovering:
		*open = true
	case *open && panel != nil && !panel.Hovered():
		// The pointer crossed the gap between the anchor and the card, so
		// it is neither on the one nor on the other.
		*open = false
	}

	if panel != nil && opts.Width > 0 {
		panel.Width(opts.Width)
	}
	return panel
}
