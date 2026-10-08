package display

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// BadgeOptions configure a Badge and a PresenceDot.
type BadgeOptions struct {
	// Tone picks the background and ink from the severity ramp. Neutral is
	// the quiet default: an ordinary count is not an alarm.
	Tone core.Severity
	// Max caps what the number reads as; zero is no cap. Past the cap the
	// badge says "99+", because a four-digit badge is wider than the avatar
	// it is stuck to and would push the row out of shape.
	Max int
	// Solid fills the badge with ink instead of tinting it, for the one
	// badge on a screen that should be the first thing the eye lands on.
	Solid bool
	// Name is what the badge is called out loud. It is required, and for a
	// dot it is the whole content: a bare mark says nothing without one.
	Name string
}

// Badge is a number stuck to the corner of something: unread messages on a
// nav item, retries on a job, comments on a card.
//
// A count of zero draws no badge at all. "0 unread" is a claim about a
// collection that the collection already makes; printing it takes the room
// the badge needs for the times there is something to say, and says it when
// there is nothing.
//
// It returns the badge itself, coloured and sized but not placed. The caller
// parents it into the box of the thing it marks and anchors it, which is what
// ui.AnchorTopRight is for:
//
//	display.Badge(c, 3, display.BadgeOptions{Name: "3 unread"}).
//	    Attach(ui.AnchorTopRight, ui.AnchorCenter)
func Badge(c *ui.Context, count int, opts BadgeOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Name == "" {
		panic("display: Badge needs a Name; a count with no word is a smudge")
	}
	if count <= 0 {
		return ui.Box(c)
	}
	bg, fg := opts.Tone.Pair(k)
	if opts.Solid {
		bg, fg = k.Fill, k.OnFill
	}

	label := internal.Commas(count)
	if opts.Max > 0 && count > opts.Max {
		label = internal.Commas(opts.Max) + "+"
	}
	// A coloured badge gets a ring in the window's colour, so a red badge on
	// a red avatar is two shapes and not one: without the ring the number
	// disappears into whatever it was stuck to.
	ring := opts.Tone != core.Neutral && !opts.Solid
	pill := ui.Box(c).Padding(u*0.25, u*1.5).Radius(theme.PillRadius).
		Background(bg).Center().Label(opts.Name)
	if ring {
		pill.Border(2, k.Background)
	}
	return pill.Children(func() {
		ui.Text(c, label).TextColor(fg).
			FontSize(core.FontSize(c, theme.CaptionSize)).Bold()
	})
}

// PresenceDot is the badge with no count: a mark saying something needs
// looking at, whatever it is.
//
// It is a separate name because it is asked for separately — a nav item with
// alerts, a presence nobody has checked — and reaching for Badge instead would
// mean passing a number that means "some", which is a number about nothing.
func PresenceDot(c *ui.Context, opts BadgeOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if opts.Name == "" {
		panic("display: PresenceDot needs a Name; a bare dot says nothing out loud")
	}
	_, fg := opts.Tone.Pair(k)
	if fg == k.Text {
		// Neutral's foreground is body ink, which would make the dot the
		// darkest thing on the row. A dot that means nothing in particular
		// is the warning amber, as the rail's alert mark is.
		fg = k.Warning
	}
	return ui.Box(c).Size(u*2.25, u*2.25).Radius(u*1.2).Background(fg).
		Border(2, k.Background).Label(opts.Name)
}
