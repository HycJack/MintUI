// Package layout holds the shapes a screen is arranged in: page headers,
// columns and the board they form.
package layout

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/internal"
	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// HeaderOptions configure a PageHeader.
type HeaderOptions struct {
	// Crumbs is the path shown above the title, joined with a separator.
	Crumbs []string
	// Title is the page's own heading.
	Title string
	// Meta is the line under the title: the counts that describe the screen.
	Meta string
	// MetaSlot builds that line instead, for a screen that wants StatLine or
	// anything else richer than one string. It wins over Meta.
	MetaSlot func()
	// AvatarSlots are the overlapping faces beside the meta line.
	AvatarSlots []string
	// SlotCount is how many people the cluster is standing for, which is
	// usually more than it can fit.
	SlotCount int
	// Tools are the controls at the right of the header, already built.
	Tools func()
	// TitleSize overrides the display size; 0 uses the library's.
	TitleSize float32
}

// PageHeader is the band across the top of a screen: where you are, what it
// is called, what it holds, and what you can do about it.
//
// Nothing here is interactive except Tools and AvatarSlots, which are the
// caller's own components.
func PageHeader(c *ui.Context, opts HeaderOptions) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	size := opts.TitleSize
	if size == 0 {
		size = theme.DisplaySize
	}

	return ui.Column(c).FillWidth().Padding(u*6, u*7, 0, u*7).Gap(u * 2).Children(func() {
		if len(opts.Crumbs) > 0 {
			ui.Text(c, join(opts.Crumbs, "  /  ")).TextColor(k.TextMuted).FontSize(theme.MetaSize)
		}
		ui.Row(c).FillWidth().AlignItems(ui.End).Gap(u * 6).Children(func() {
			ui.Column(c).Grow(1).Gap(u * 2).Children(func() {
				ui.Text(c, opts.Title).TextColor(k.Text).FontSize(size).Bold()
				if opts.MetaSlot != nil {
					opts.MetaSlot()
				} else if opts.Meta != "" {
					ui.Text(c, opts.Meta).TextColor(k.TextMuted).FontSize(theme.RowSize)
				}
				if opts.SlotCount > 0 {
					AvatarSlots(c, opts.AvatarSlots, opts.SlotCount)
				}
			})
			if opts.Tools != nil {
				ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(opts.Tools)
			}
		})
	})
}

// AvatarSlots draws the overlapping faces of a header. It is a thin wrapper
// over display.AvatarCluster, kept here so a header is one import.
func AvatarSlots(c *ui.Context, names []string, total int) *ui.Element {
	k, u := core.Tokens(c), core.Density(c).Unit()
	if total <= 0 {
		return ui.Box(c)
	}
	side := u * 7.5
	return ui.Row(c).AlignItems(ui.Center).Label("team").Children(func() {
		for i, name := range names {
			if i >= 5 {
				break
			}
			box := ui.Box(c).Size(side, side).Radius(side/2).Background(k.Surface).
				Border(2, k.Background).Center().Label(name)
			if i > 0 {
				box.Margin(0, 0, 0, -side*0.22)
			}
			box.Children(func() {
				if in := internal.Initials(name); in != "" {
					ui.Text(c, in).TextColor(k.Text).FontSize(theme.MonoSize).Bold()
				}
			})
		}
		ui.Text(c, itoa(total)+" technicians").TextColor(k.TextMuted).FontSize(theme.RowSize)
	})
}

func join(parts []string, sep string) string { return strings.Join(parts, sep) }

func itoa(n int) string { return internal.Commas(n) }
