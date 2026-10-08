// Package navigation holds the chrome that moves a person around an
// application: the icon rail, filter rows and the groups they collect into.
package navigation

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// RailItem is one destination in a Rail.
type RailItem struct {
	// Name is the destination, and what a screen reader announces.
	Name string
	// Icon is the glyph shown when the item is not selected.
	Icon *ui.SVG
	// Badge marks an item as needing attention, as a dot rather than a count.
	Badge bool
	// Selected shows the item as the current destination.
	Selected bool
}

// Rail is the vertical strip of destinations down a window's left edge: an
// identity at the top, the items, then tools at the bottom.
//
// It adds no behaviour of its own — it draws, and the caller reads what the
// app state says is selected.
func Rail(c *ui.Context, identity string, items []RailItem, tools []RailItem) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	side := u * 11

	column := ui.Column(c).Width(side+u*8).FillHeight().Background(k.Background).
		Padding(u*4.5, 0, u*4, 0).Gap(u * 2.5).AlignItems(ui.Center)

	column.Children(func() {
		// The identity, with a presence dot on its corner.
		ui.Box(c).Children(func() {
			ui.Box(c).Size(side, side).Radius(side / 2).Background(k.Fill).Center().
				Label(identity).Children(func() {
				if in := internalInitials(identity); in != "" {
					ui.Text(c, in).TextColor(k.OnFill).FontSize(17).Bold()
				}
			})
			ui.Box(c).Size(u*2.75, u*2.75).Radius(u*1.4).Background(k.Lively).
				Border(2, k.Background).Attach(ui.AnchorBottomRight, ui.AnchorCenter).
				Right(-u * 0.75)
		})
		ui.Box(c).Height(u)

		for _, it := range items {
			railButton(c, it, side, u)
		}
		ui.Spacer(c)
		for _, it := range tools {
			railButton(c, it, side, u)
		}
	})
	return column
}

func railButton(c *ui.Context, it RailItem, side, u float32) {
	k := core.Tokens(c)
	bg, fg := k.Surface, k.TextMuted
	if it.Selected {
		bg, fg = k.Fill, k.OnFill
	}
	ui.Box(c).Children(func() {
		ui.Box(c).Size(side, side).Radius(theme.ControlRadius).Background(bg).
			Center().Label(it.Name).Children(func() {
			if it.Icon != nil {
				ui.Icon(c, it.Icon).TextColor(fg).Size(theme.IconSize, theme.IconSize)
			}
		})
		if it.Badge {
			ui.Box(c).Size(u*2.25, u*2.25).Radius(u*1.2).Background(k.Warning).
				Border(2, k.Background).Attach(ui.AnchorTopRight, ui.AnchorCenter).
				Right(-u * 0.5)
		}
	})
}
