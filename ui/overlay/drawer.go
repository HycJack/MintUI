package overlay

import (
	"github.com/egoist/mygo/ui"

	"github.com/HycJack/MintUI/ui/core"
	"github.com/HycJack/MintUI/ui/theme"
)

// DrawerOptions configure a Drawer.
type DrawerOptions struct {
	// Side is ui.Start for a sheet from the left edge, ui.End for one from
	// the right. ui.Center is a panel down the middle of a narrow window,
	// where a drawer from the side would leave nothing of the page to read.
	Side ui.Align
	// Title heads the sheet; empty draws a panel of body alone.
	Title string
	// Subtitle is the second line under the title: what the sheet is for.
	Subtitle string
	// Body is the content.
	Body func()
	// Actions is the button row under the body.
	Actions func()
	// Width is the sheet's width; zero lets it fill the room the side leaves.
	Width float32
	// Rule draws a hairline under the title, for a form under a heading.
	Rule bool
	// NonModal leaves the window behind live.
	NonModal bool
}

// Drawer is a panel the height of the window, hung from one edge: the place a
// record is edited in without losing the board behind it.
//
// It is modal unless NonModal is set, so a press on the scrim or Escape writes
// false into the caller's *bool — the same two ways a Dialog closes, which is
// why both are one component with a Side.
//
//	core.Use(c, core.Settings{})
//	Drawer(c, &app.editing, DrawerOptions{
//	    Side:  ui.End,
//	    Title: "Log callback",
//	    Body:  func() { fields() },
//	})
func Drawer(c *ui.Context, open *bool, opts DrawerOptions) *ui.Element {
	u := core.Density(c).Unit()

	panel := layer(c, open, !opts.NonModal, opts.Side, func(_, p *ui.Element) {
		panelOpts := PanelOptions{
			Title:     opts.Title,
			Subtitle:  opts.Subtitle,
			Rule:      opts.Rule,
			TitleSize: theme.SheetSize,
		}
		if opts.Side == ui.Start {
			// A sheet hung from the left edge is rounded on the side that
			// faces the window and square against it, or its corner would
			// show the page behind through the gap where the scrim ends.
			// That is the one asymmetry the radius family has to spell out.
			panelOpts.Round = func(r float32) (float32, float32, float32, float32) {
				return 0, r, r, 0
			}
		} else if opts.Side == ui.End {
			panelOpts.Round = func(r float32) (float32, float32, float32, float32) {
				return r, 0, r, 0
			}
		}
		Panel(c, p, panelOpts, func() {
			if opts.Body != nil {
				opts.Body()
			}
			if opts.Actions != nil {
				ui.Row(c).FillWidth().Gap(u*2).Justify(ui.End).Margin(u*2.5, 0, 0, 0).
					Children(opts.Actions)
			}
		})
	})
	if panel == nil {
		return nil
	}
	panel.FillHeight()
	if opts.Width > 0 {
		panel.Width(opts.Width).Shrink(0)
	} else {
		// A sheet with no width of its own takes half the window and no
		// more: wide enough for a form, narrow enough that the board it is
		// over stays recognisable.
		panel.WidthPercent(50).MinWidth(u * 60)
	}
	return panel
}
